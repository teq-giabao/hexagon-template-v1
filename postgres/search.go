package postgres

import (
	"context"
	"sort"
	"strings"
	"time"

	"hexagon/search"

	"gorm.io/gorm"
)

type SearchRepository struct {
	db *gorm.DB
}

const (
	hotelQueryChunkSize = 500
	roomQueryChunkSize  = 1000
)

func NewSearchRepository(db *gorm.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

/*
SearchHotels loads candidate hotels, evaluates availability/allocation,
and returns only hotels that satisfy strict or flexible matching rules.
*/
func (r *SearchRepository) SearchHotels(ctx context.Context, criteria search.Criteria) ([]search.HotelSearchResult, error) {
	hotels, err := r.findHotels(ctx, criteria)
	if err != nil {
		return nil, err
	}

	hotelIDs := make([]string, len(hotels))
	for i := range hotels {
		hotelIDs[i] = hotels[i].ID
	}

	candidatesByHotel, err := r.loadRoomCandidatesByHotels(ctx, hotelIDs, criteria)
	if err != nil {
		return nil, err
	}

	results := make([]search.HotelSearchResult, 0, len(hotels))

	for i := range hotels {
		strict, flexible, minPrice, availableRoomTypes := evaluateHotelAvailabilityFromCandidates(
			hotels[i],
			criteria,
			candidatesByHotel[hotels[i].ID],
		)

		if !strict && !flexible {
			continue
		}

		results = append(results, search.HotelSearchResult{
			HotelID:            hotels[i].ID,
			Name:               hotels[i].Name,
			City:               hotels[i].City,
			Address:            hotels[i].Address,
			Rating:             hotels[i].Rating,
			PaymentOptions:     toEnabledPaymentOptions(hotels[i].PaymentOptions),
			MinPrice:           minPrice,
			AvailableRoomCount: availableRoomTypes,
			MatchesRequested:   strict,
			FlexibleMatch:      flexible,
		})
	}

	return results, nil
}

/*
evaluateHotelAvailabilityFromCandidates computes final match flags for one hotel:
- strict: satisfies requested room count.
- flexible: only satisfiable with more rooms than requested.

It also returns minimum base price and number of available room types.
*/
func evaluateHotelAvailabilityFromCandidates(
	hotel HotelModel,
	criteria search.Criteria,
	candidates []roomCandidate,
) (strict bool, flexible bool, minPrice float64, availableRoomTypes int) {
	if !hotelMatchesFilter(hotel, criteria) {
		return false, false, 0, 0
	}

	if len(candidates) == 0 {
		return false, false, 0, 0
	}

	for i := range candidates {
		if minPrice == 0 || candidates[i].Room.BasePrice < minPrice {
			minPrice = candidates[i].Room.BasePrice
		}
	}

	/*
		Find the minimum number of rooms that can satisfy the party constraints.
		This supports flexible matches when a strict room count is not possible.
	*/
	requiredRoomCount := minimumRequiredRooms(candidates, criteria.RoomCount, criteria.Adults, criteria.ChildrenAges, hotel.DefaultChildMaxAge)
	if requiredRoomCount == 0 {
		return false, false, minPrice, len(candidates)
	}

	strict = requiredRoomCount == criteria.RoomCount
	flexible = requiredRoomCount > criteria.RoomCount

	return strict, flexible, minPrice, len(candidates)
}

/*
findHotels applies coarse hotel-level filters in SQL
(text query, rating, payment option) and relies on SQL ordering for query relevance.
*/
func (r *SearchRepository) findHotels(ctx context.Context, criteria search.Criteria) ([]HotelModel, error) {
	query := r.db.WithContext(ctx).Model(&HotelModel{}).
		Preload("PaymentOptions")

	if q := strings.TrimSpace(criteria.Query); q != "" {
		// TODO: replace ILIKE matching with full-text search (and optionally unaccent/trigram)
		normalized := strings.ToLower(q)
		like := "%" + q + "%"
		query = query.Where("hotels.name ILIKE ? OR hotels.city ILIKE ? OR hotels.address ILIKE ?", like, like, like)
		/*
			Relevance ranking in SQL to avoid in-memory sorting.
			Priority: exact name > name prefix > name contains > city matches > address contains.
		*/
		query = query.Order(gorm.Expr(`CASE
			WHEN LOWER(hotels.name) = ? THEN 1
			WHEN LOWER(hotels.name) LIKE ? THEN 2
			WHEN hotels.name ILIKE ? THEN 3
			WHEN LOWER(hotels.city) = ? THEN 4
			WHEN LOWER(hotels.city) LIKE ? THEN 5
			WHEN hotels.city ILIKE ? THEN 6
			WHEN hotels.address ILIKE ? THEN 7
			ELSE 8
		END`, normalized, normalized+"%", like, normalized, normalized+"%", like, like))
	}

	if criteria.RatingMin > 0 {
		query = query.Where("hotels.rating >= ?", criteria.RatingMin)
	}

	if len(criteria.PaymentOptions) > 0 {
		query = query.Joins("JOIN hotel_payment_options hpo ON hpo.hotel_id = hotels.id AND hpo.enabled = TRUE").
			Where("hpo.payment_option IN ?", criteria.PaymentOptions).
			Group("hotels.id")
	}

	query = query.Order("hotels.rating DESC").Order("hotels.id ASC")

	var hotels []HotelModel
	if err := query.Find(&hotels).Error; err != nil {
		return nil, err
	}

	return hotels, nil
}

/*
hotelMatchesFilter performs defensive in-memory checks for hotel-level filters.
This keeps behavior safe even if upstream query shape changes.
*/
func hotelMatchesFilter(h HotelModel, criteria search.Criteria) bool {
	if criteria.RatingMin > 0 && h.Rating < criteria.RatingMin {
		return false
	}

	if len(criteria.PaymentOptions) == 0 {
		return true
	}

	enabled := toEnabledPaymentOptions(h.PaymentOptions)
	if len(enabled) == 0 {
		return false
	}

	set := make(map[string]struct{}, len(enabled))
	for i := range enabled {
		set[enabled[i]] = struct{}{}
	}

	for i := range criteria.PaymentOptions {
		if _, ok := set[criteria.PaymentOptions[i]]; ok {
			return true
		}
	}

	return false
}

/*
toEnabledPaymentOptions extracts enabled payment options from joined DB rows.
*/
func toEnabledPaymentOptions(options []HotelPaymentOptionModel) []string {
	result := make([]string, 0, len(options))

	for i := range options {
		if options[i].Enabled {
			result = append(result, options[i].PaymentOption)
		}
	}

	return result
}

type roomCandidate struct {
	Room           RoomModel
	AvailableCount int
	AmenityIDs     []string
}

/*
loadRoomCandidatesByHotels loads active rooms for the target hotels,
then enriches each room with date-range availability and amenities.
Result is grouped by hotel for downstream allocation checks.
*/
func (r *SearchRepository) loadRoomCandidatesByHotels(ctx context.Context, hotelIDs []string, criteria search.Criteria) (map[string][]roomCandidate, error) {
	result := make(map[string][]roomCandidate, len(hotelIDs))
	if len(hotelIDs) == 0 {
		return result, nil
	}

	var roomModels []RoomModel

	/*
		Chunked loading prevents very large IN clauses and reduces memory spikes.
	*/
	for _, hotelIDChunk := range chunkStrings(hotelIDs, hotelQueryChunkSize) {
		var chunkRooms []RoomModel

		if err := r.db.WithContext(ctx).
			Where("hotel_id IN ? AND status = ?", hotelIDChunk, "active").
			Find(&chunkRooms).Error; err != nil {
			return nil, err
		}

		roomModels = append(roomModels, chunkRooms...)
	}

	if len(roomModels) == 0 {
		return result, nil
	}

	roomIDs := make([]string, len(roomModels))
	for i := range roomModels {
		roomIDs[i] = roomModels[i].ID
	}

	availMap, err := r.roomAvailabilityByDateRange(ctx, roomIDs, criteria.CheckInDate, criteria.CheckOutDate)
	if err != nil {
		return nil, err
	}

	amenityMap, err := r.roomAmenityIDs(ctx, roomIDs)
	if err != nil {
		return nil, err
	}

	for i := range roomModels {
		available := availMap[roomModels[i].ID]
		if available <= 0 {
			continue
		}

		amenityIDs := amenityMap[roomModels[i].ID]
		if !hasAllAmenities(amenityIDs, criteria.AmenityIDs) {
			continue
		}

		hotelID := roomModels[i].HotelID
		result[hotelID] = append(result[hotelID], roomCandidate{
			Room:           roomModels[i],
			AvailableCount: available,
			AmenityIDs:     amenityIDs,
		})
	}

	return result, nil
}

/*
roomAvailabilityByDateRange returns per-room minimum available inventory
across all requested nights. Rooms missing any night are excluded.
*/
func (r *SearchRepository) roomAvailabilityByDateRange(ctx context.Context, roomIDs []string, checkIn, checkOut time.Time) (map[string]int, error) {
	if len(roomIDs) == 0 {
		return map[string]int{}, nil
	}

	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	if nights <= 0 {
		nights = 1
	}

	type availabilityRow struct {
		RoomID         string
		AvailableCount int
		DayCount       int
	}

	result := make(map[string]int)

	/*
		Availability is aggregated per room across the requested date range.
	*/
	for _, roomIDChunk := range chunkStrings(roomIDs, roomQueryChunkSize) {
		var rows []availabilityRow

		err := r.db.WithContext(ctx).
			Table("room_inventories").
			Select("room_id, MIN(total_inventory - held_inventory - booked_inventory) AS available_count, COUNT(*) AS day_count").
			Where("room_id IN ?", roomIDChunk).
			Where("date >= ? AND date < ?", checkIn, checkOut).
			Group("room_id").
			Having("COUNT(*) = ?", nights).
			Scan(&rows).Error
		if err != nil {
			return nil, err
		}

		for i := range rows {
			result[rows[i].RoomID] = rows[i].AvailableCount
		}
	}

	return result, nil
}

/*
roomAmenityIDs loads amenity ids per room and groups them by room id.
*/
func (r *SearchRepository) roomAmenityIDs(ctx context.Context, roomIDs []string) (map[string][]string, error) {
	if len(roomIDs) == 0 {
		return map[string][]string{}, nil
	}

	type row struct {
		RoomID    string
		AmenityID string
	}

	idMap := make(map[string][]string)

	/*
		Chunk amenity lookup for the same reason as room/inventory queries.
	*/
	for _, roomIDChunk := range chunkStrings(roomIDs, roomQueryChunkSize) {
		var rows []row

		err := r.db.WithContext(ctx).
			Table("room_amenity_maps").
			Select("room_id, amenity_id").
			Where("room_id IN ?", roomIDChunk).
			Scan(&rows).Error
		if err != nil {
			return nil, err
		}

		for i := range rows {
			idMap[rows[i].RoomID] = append(idMap[rows[i].RoomID], rows[i].AmenityID)
		}
	}

	return idMap, nil
}

/*
chunkStrings splits large id lists into fixed-size chunks
to keep SQL IN queries manageable.
*/
func chunkStrings(values []string, chunkSize int) [][]string {
	if len(values) == 0 {
		return nil
	}

	if chunkSize <= 0 || len(values) <= chunkSize {
		return [][]string{values}
	}

	chunks := make([][]string, 0, (len(values)+chunkSize-1)/chunkSize)

	for start := 0; start < len(values); start += chunkSize {
		end := start + chunkSize
		if end > len(values) {
			end = len(values)
		}

		chunks = append(chunks, values[start:end])
	}

	return chunks
}

/*
hasAllAmenities checks if a room candidate fully covers requested amenity filters.
*/
func hasAllAmenities(roomAmenityIDs, filterAmenityIDs []string) bool {
	if len(filterAmenityIDs) == 0 {
		return true
	}

	set := make(map[string]struct{}, len(roomAmenityIDs))
	for i := range roomAmenityIDs {
		set[roomAmenityIDs[i]] = struct{}{}
	}

	for i := range filterAmenityIDs {
		if _, ok := set[filterAmenityIDs[i]]; !ok {
			return false
		}
	}

	return true
}

type guestRequest struct {
	adults       int
	childrenAges []int
}

/*
splitGuestsAcrossRooms builds an initial per-room guest distribution using round-robin.
This is only a starting shape; final feasibility is validated by backtracking.
*/
func splitGuestsAcrossRooms(roomCount, adults int, childrenAges []int) []guestRequest {
	if roomCount <= 0 {
		return nil
	}

	rooms := make([]guestRequest, roomCount)
	for i := 0; i < roomCount; i++ {
		rooms[i] = guestRequest{adults: 0, childrenAges: make([]int, 0)}
	}

	/*
		Initial distribution is round-robin; feasibility is validated later.
	*/
	for i := 0; i < adults; i++ {
		rooms[i%roomCount].adults++
	}

	for i := range childrenAges {
		rooms[i%roomCount].childrenAges = append(rooms[i%roomCount].childrenAges, childrenAges[i])
	}

	return rooms
}

/*
canAllocateRequestedRooms validates whether all room requests can be assigned
to available room types while respecting occupancy constraints and inventory counts.
*/
func canAllocateRequestedRooms(candidates []roomCandidate, requests []guestRequest, childMaxAge int) bool {
	if len(requests) == 0 {
		return true
	}

	if hasAnyUnaccompaniedChildren(requests, childMaxAge) {
		return false
	}

	order, compat, ok := buildCompatibilityOrder(candidates, requests, childMaxAge)
	if !ok {
		return false
	}

	remaining := make([]int, len(candidates))
	for i := range candidates {
		remaining[i] = candidates[i].AvailableCount
	}

	/*
		DFS backtracking over compatible room types with inventory accounting.
	*/
	var dfs func(pos int) bool

	dfs = func(pos int) bool {
		if pos == len(order) {
			return true
		}

		reqIdx := order[pos]
		for _, candidateIdx := range compat[reqIdx] {
			if remaining[candidateIdx] <= 0 {
				continue
			}

			remaining[candidateIdx]--

			if dfs(pos + 1) {
				return true
			}

			remaining[candidateIdx]++
		}

		return false
	}

	return dfs(0)
}

/*
buildCompatibilityOrder precomputes which room candidates can satisfy each request,
then orders requests from hardest to easiest (fewest compatible candidates first).
*/
func buildCompatibilityOrder(candidates []roomCandidate, requests []guestRequest, childMaxAge int) ([]int, [][]int, bool) {
	order := make([]int, len(requests))
	for i := range requests {
		order[i] = i
	}

	compat := make([][]int, len(requests))

	for i := range requests {
		for j := range candidates {
			if candidateCanFit(candidates[j], requests[i], childMaxAge) && candidates[j].AvailableCount > 0 {
				compat[i] = append(compat[i], j)
			}
		}

		if len(compat[i]) == 0 {
			return nil, nil, false
		}
	}

	/*
		Heuristic: place the most constrained request first to prune backtracking faster.
	*/
	sort.Slice(order, func(i, j int) bool { return len(compat[order[i]]) < len(compat[order[j]]) })

	return order, compat, true
}

/*
minimumRequiredRooms searches from requested room count upward
and returns the smallest feasible room count.
Returns 0 when no feasible allocation exists.
*/
func minimumRequiredRooms(
	candidates []roomCandidate,
	startRoomCount int,
	adults int,
	childrenAges []int,
	childMaxAge int,
) int {
	totalAvailable := 0
	for i := range candidates {
		totalAvailable += candidates[i].AvailableCount
	}

	/*
		Try from requested room count upward until a feasible allocation is found.
	*/
	for roomCount := startRoomCount; roomCount <= totalAvailable; roomCount++ {
		requests := splitGuestsAcrossRooms(roomCount, adults, childrenAges)
		if canAllocateRequestedRooms(candidates, requests, childMaxAge) {
			return roomCount
		}
	}

	return 0
}

/*
candidateCanFit checks occupancy limits after applying child-age normalization.
*/
func candidateCanFit(c roomCandidate, req guestRequest, childMaxAge int) bool {
	adults, children := normalizeGuest(req, childMaxAge)
	total := adults + children

	return adults <= c.Room.MaxAdult && children <= c.Room.MaxChild && total <= c.Room.MaxOccupancy
}

/*
normalizeGuest converts over-age children into adults based on hotel child policy.
*/
func normalizeGuest(req guestRequest, childMaxAge int) (int, int) {
	adults := req.adults
	children := 0

	for i := range req.childrenAges {
		if req.childrenAges[i] > childMaxAge {
			adults++
		} else {
			children++
		}
	}

	return adults, children
}

/*
hasUnaccompaniedChildren enforces business rule:
children cannot stay without an adult.
*/
func hasUnaccompaniedChildren(req guestRequest, childMaxAge int) bool {
	adults, children := normalizeGuest(req, childMaxAge)

	return children > 0 && adults == 0
}

/*
hasAnyUnaccompaniedChildren checks the no-unaccompanied-children rule across all rooms.
*/
func hasAnyUnaccompaniedChildren(requests []guestRequest, childMaxAge int) bool {
	for i := range requests {
		if hasUnaccompaniedChildren(requests[i], childMaxAge) {
			return true
		}
	}

	return false
}
