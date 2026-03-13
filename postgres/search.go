package postgres

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"hexagon/search"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// SearchHotels is the first step in the hotel search flow.
// It focuses on filtering hotels by criteria like location and rating,
// and performs a rapid capacity check to see if a hotel roughly qualifies
// for the guests without doing an expensive room-by-room combination match.
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
		availability, err := evaluateHotelAvailabilityFromCandidates(
			hotels[i],
			criteria,
			candidatesByHotel[hotels[i].ID],
		)
		if err != nil {
			return nil, err
		}

		if !availability.Strict && !availability.Flexible {
			continue
		}

		results = append(results, search.HotelSearchResult{
			HotelID:            hotels[i].ID,
			Name:               hotels[i].Name,
			City:               hotels[i].City,
			Address:            hotels[i].Address,
			Rating:             hotels[i].Rating,
			PaymentOptions:     toEnabledPaymentOptions(hotels[i].PaymentOptions),
			MinPrice:           availability.MinPrice,
			AvailableRoomCount: availability.AvailableRoomTypes,
			MatchesRequested:   availability.Strict,
			FlexibleMatch:      availability.Flexible,
		})
	}

	return results, nil
}

// SearchHotelRooms acts as the detailed capacity and amenity retrieval for a specific hotel.
// It also computes the StrictMatch flag, checking if there's enough available rooms
// to pack all guests WITHOUT EXCEEDING the exactly requested room count.
func (r *SearchRepository) SearchHotelRooms(ctx context.Context, hotelID string, criteria search.Criteria) (search.HotelRoomSearchResult, error) {
	hotel, candidates, err := r.loadHotelWithCandidates(ctx, hotelID, criteria)
	if err != nil {
		return search.HotelRoomSearchResult{}, err
	}

	searchCandidates := toSearchRoomCandidates(candidates)
	minimumRooms := search.MinimumRequiredRooms(searchCandidates, criteria.RoomCount, criteria.Adults, criteria.ChildrenAges, hotel.DefaultChildMaxAge)

	amenityDetailsByRoomID, err := r.roomAmenityDetailsByRoomIDs(ctx, roomIDsFromCandidates(candidates))
	if err != nil {
		return search.HotelRoomSearchResult{}, err
	}

	rooms := buildHotelRoomItems(candidates, amenityDetailsByRoomID)

	return search.HotelRoomSearchResult{
		HotelID:            hotelID,
		RequestedRoomCount: criteria.RoomCount,
		StrictMatch:        minimumRooms > 0 && minimumRooms == criteria.RoomCount,
		Rooms:              rooms,
	}, nil
}

// SearchHotelRoomCombinations generates valid, multi-room arrays where guests
// can fit appropriately. The limits protect against combinatorial explosion
// of checking every single possible permutation of available rooms.
func (r *SearchRepository) SearchHotelRoomCombinations(
	ctx context.Context,
	hotelID string,
	criteria search.Criteria,
	maxCombinations int,
) (search.HotelRoomCombinationsResult, error) {
	hotel, rawCandidates, err := r.loadHotelWithCandidates(ctx, hotelID, criteria)
	if err != nil {
		return search.HotelRoomCombinationsResult{}, err
	}

	if maxCombinations <= 0 {
		maxCombinations = 5
	}

	combinations := buildRoomCombinations(
		toSearchRoomCandidates(rawCandidates),
		criteria.RoomCount,
		criteria.Adults,
		criteria.ChildrenAges,
		hotel.DefaultChildMaxAge,
	)
	combinations = sortAndLimitCombinations(combinations, maxCombinations)

	return search.HotelRoomCombinationsResult{
		HotelID:            hotelID,
		RequestedRoomCount: criteria.RoomCount,
		Combinations:       combinations,
	}, nil
}

func (r *SearchRepository) loadHotelWithCandidates(ctx context.Context, hotelID string, criteria search.Criteria) (HotelModel, []roomCandidate, error) {
	hotel, err := r.findHotelByID(ctx, hotelID)
	if err != nil {
		return HotelModel{}, nil, err
	}

	candidatesByHotel, err := r.loadRoomCandidatesByHotels(ctx, []string{hotelID}, criteria)
	if err != nil {
		return HotelModel{}, nil, err
	}

	return hotel, candidatesByHotel[hotelID], nil
}

func roomIDsFromCandidates(candidates []roomCandidate) []string {
	roomIDs := make([]string, len(candidates))
	for i := range candidates {
		roomIDs[i] = candidates[i].Room.ID
	}

	return roomIDs
}

func buildHotelRoomItems(candidates []roomCandidate, amenityDetailsByRoomID map[string][]roomAmenityDetailRow) []search.HotelRoomSearchItem {
	rooms := make([]search.HotelRoomSearchItem, len(candidates))

	for i := range candidates {
		amenities := amenityDetailsByRoomID[candidates[i].Room.ID]
		amenityIDs := make([]string, len(amenities))
		amenityCodes := make([]string, len(amenities))
		amenityNames := make([]string, len(amenities))

		for j := range amenities {
			amenityIDs[j] = amenities[j].AmenityID
			amenityCodes[j] = amenities[j].AmenityCode
			amenityNames[j] = amenities[j].AmenityName
		}

		rooms[i] = search.HotelRoomSearchItem{
			RoomID:         candidates[i].Room.ID,
			Name:           candidates[i].Room.Name,
			Description:    candidates[i].Room.Description,
			BasePrice:      candidates[i].Room.BasePrice,
			MaxAdult:       candidates[i].Room.MaxAdult,
			MaxChild:       candidates[i].Room.MaxChild,
			MaxOccupancy:   candidates[i].Room.MaxOccupancy,
			AvailableCount: candidates[i].AvailableCount,
			AmenityIDs:     amenityIDs,
			AmenityCodes:   amenityCodes,
			AmenityNames:   amenityNames,
		}
	}

	sort.Slice(rooms, func(i, j int) bool {
		if rooms[i].BasePrice == rooms[j].BasePrice {
			return rooms[i].RoomID < rooms[j].RoomID
		}

		return rooms[i].BasePrice < rooms[j].BasePrice
	})

	return rooms
}

func buildRoomCombinations(candidates []search.RoomCandidate, roomCount, adults int, childrenAges []int, childMaxAge int) []search.RoomCombination {
	combinations := make([]search.RoomCombination, 0, len(candidates))

	for i := range candidates {
		quantity, ok := search.RequiredSingleRoomTypeQuantity(candidates[i], roomCount, adults, childrenAges, childMaxAge)
		if !ok {
			continue
		}

		subtotal := float64(quantity) * candidates[i].BasePrice
		combinations = append(combinations, search.RoomCombination{
			Items: []search.RoomCombinationItem{{
				RoomID:    candidates[i].RoomID,
				RoomName:  candidates[i].Name,
				Quantity:  quantity,
				UnitPrice: candidates[i].BasePrice,
				Subtotal:  subtotal,
			}},
			TotalPrice:     subtotal,
			TotalRooms:     quantity,
			TotalMaxAdult:  quantity * candidates[i].MaxAdult,
			TotalMaxChild:  quantity * candidates[i].MaxChild,
			TotalOccupancy: quantity * candidates[i].MaxOccupancy,
		})
	}

	return combinations
}

func sortAndLimitCombinations(combinations []search.RoomCombination, maxCombinations int) []search.RoomCombination {
	sort.Slice(combinations, func(i, j int) bool {
		if combinations[i].TotalPrice == combinations[j].TotalPrice {
			if combinations[i].TotalRooms == combinations[j].TotalRooms {
				return combinations[i].Items[0].RoomID < combinations[j].Items[0].RoomID
			}

			return combinations[i].TotalRooms < combinations[j].TotalRooms
		}

		return combinations[i].TotalPrice < combinations[j].TotalPrice
	})

	if len(combinations) > maxCombinations {
		return combinations[:maxCombinations]
	}

	return combinations
}

type hotelAvailability struct {
	Strict             bool
	Flexible           bool
	MinPrice           float64
	AvailableRoomTypes int
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
) (hotelAvailability, error) {
	if !hotelMatchesFilter(hotel, criteria) {
		return hotelAvailability{}, nil
	}

	if len(candidates) == 0 {
		return hotelAvailability{}, nil
	}

	minPrice := 0.0
	for i := range candidates {
		if minPrice == 0 || candidates[i].Room.BasePrice < minPrice {
			minPrice = candidates[i].Room.BasePrice
		}
	}

	/*
		Find the minimum number of rooms that can satisfy the party constraints.
		This supports flexible matches when a strict room count is not possible.
	*/
	requiredRoomCount := search.MinimumRequiredRooms(toSearchRoomCandidates(candidates), criteria.RoomCount, criteria.Adults, criteria.ChildrenAges, hotel.DefaultChildMaxAge)
	if requiredRoomCount == 0 {
		return hotelAvailability{
			MinPrice:           minPrice,
			AvailableRoomTypes: len(candidates),
		}, nil
	}

	return hotelAvailability{
		Strict:             requiredRoomCount == criteria.RoomCount,
		Flexible:           requiredRoomCount > criteria.RoomCount,
		MinPrice:           minPrice,
		AvailableRoomTypes: len(candidates),
	}, nil
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
		like := "%" + normalized + "%"
		query = query.Where(clause.Or(
			clause.Like{Column: clause.Expr{SQL: "LOWER(hotels.name)"}, Value: like},
			clause.Like{Column: clause.Expr{SQL: "LOWER(hotels.city)"}, Value: like},
			clause.Like{Column: clause.Expr{SQL: "LOWER(hotels.address)"}, Value: like},
		))
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

func (r *SearchRepository) findHotelByID(ctx context.Context, hotelID string) (HotelModel, error) {
	var hotel HotelModel

	err := r.db.WithContext(ctx).
		Where("id = ?", hotelID).
		First(&hotel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return HotelModel{}, search.ErrHotelNotFound
		}

		return HotelModel{}, err
	}

	return hotel, nil
}

type roomCandidate struct {
	Room           RoomModel
	AvailableCount int
	AmenityIDs     []string
}

func toSearchRoomCandidates(in []roomCandidate) []search.RoomCandidate {
	out := make([]search.RoomCandidate, len(in))
	for i := range in {
		out[i] = search.RoomCandidate{
			RoomID:         in[i].Room.ID,
			Name:           in[i].Room.Name,
			Description:    in[i].Room.Description,
			BasePrice:      in[i].Room.BasePrice,
			MaxAdult:       in[i].Room.MaxAdult,
			MaxChild:       in[i].Room.MaxChild,
			MaxOccupancy:   in[i].Room.MaxOccupancy,
			AvailableCount: in[i].AvailableCount,
		}
	}

	return out
}

// loadRoomCandidatesByHotels finds all active rooms for given hotels, enriches each room
// with date-range availability/amenities, and batches GORM queries to avoid
// PostgreSQL's array parameter binding limits (65535 parameters max).
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

type roomAmenityDetailRow struct {
	RoomID      string
	AmenityID   string
	AmenityCode string
	AmenityName string
}

func (r *SearchRepository) roomAmenityDetailsByRoomIDs(ctx context.Context, roomIDs []string) (map[string][]roomAmenityDetailRow, error) {
	if len(roomIDs) == 0 {
		return map[string][]roomAmenityDetailRow{}, nil
	}

	result := make(map[string][]roomAmenityDetailRow)

	for _, roomIDChunk := range chunkStrings(roomIDs, roomQueryChunkSize) {
		var rows []roomAmenityDetailRow

		err := r.db.WithContext(ctx).
			Table("room_amenity_maps AS ram").
			Select("ram.room_id, ra.id AS amenity_id, ra.code AS amenity_code, ra.name AS amenity_name").
			Joins("JOIN room_amenities AS ra ON ra.id = ram.amenity_id").
			Where("ram.room_id IN ?", roomIDChunk).
			Order("ram.room_id ASC, ra.code ASC, ra.id ASC").
			Scan(&rows).Error
		if err != nil {
			return nil, err
		}

		for i := range rows {
			result[rows[i].RoomID] = append(result[rows[i].RoomID], rows[i])
		}
	}

	return result, nil
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
