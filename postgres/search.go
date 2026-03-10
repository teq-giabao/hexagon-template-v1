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

func NewSearchRepository(db *gorm.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

func (r *SearchRepository) SearchHotels(ctx context.Context, criteria search.Criteria) ([]search.HotelSearchResult, error) {
	hotels, err := r.findHotels(ctx, criteria)
	if err != nil {
		return nil, err
	}

	results := make([]search.HotelSearchResult, 0, len(hotels))

	for i := range hotels {
		strict, flexible, minPrice, availableRoomTypes, err := r.evaluateHotelAvailability(ctx, hotels[i], criteria)
		if err != nil {
			return nil, err
		}

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

	sort.Slice(results, func(i, j int) bool {
		if results[i].MinPrice == results[j].MinPrice {
			return results[i].Rating > results[j].Rating
		}

		return results[i].MinPrice < results[j].MinPrice
	})

	return results, nil
}

func (r *SearchRepository) findHotels(ctx context.Context, criteria search.Criteria) ([]HotelModel, error) {
	query := r.db.WithContext(ctx).Model(&HotelModel{}).
		Preload("PaymentOptions")

	if q := strings.TrimSpace(criteria.Query); q != "" {
		// TODO: replace ILIKE matching with full-text search (and optionally unaccent/trigram)
		like := "%" + q + "%"
		query = query.Where("hotels.name ILIKE ? OR hotels.city ILIKE ? OR hotels.address ILIKE ?", like, like, like)
	}

	if criteria.RatingMin > 0 {
		query = query.Where("hotels.rating >= ?", criteria.RatingMin)
	}

	if len(criteria.PaymentOptions) > 0 {
		query = query.Joins("JOIN hotel_payment_options hpo ON hpo.hotel_id = hotels.id AND hpo.enabled = TRUE").
			Where("hpo.payment_option IN ?", criteria.PaymentOptions).
			Group("hotels.id")
	}

	var hotels []HotelModel
	if err := query.Find(&hotels).Error; err != nil {
		return nil, err
	}

	return hotels, nil
}

func (r *SearchRepository) evaluateHotelAvailability(
	ctx context.Context,
	hotel HotelModel,
	criteria search.Criteria,
) (strict bool, flexible bool, minPrice float64, availableRoomTypes int, err error) {
	if !hotelMatchesFilter(hotel, criteria) {
		return false, false, 0, 0, nil
	}

	candidates, err := r.loadRoomCandidates(ctx, hotel, criteria)
	if err != nil {
		return false, false, 0, 0, err
	}

	if len(candidates) == 0 {
		return false, false, 0, 0, nil
	}

	for i := range candidates {
		if minPrice == 0 || candidates[i].Room.BasePrice < minPrice {
			minPrice = candidates[i].Room.BasePrice
		}
	}

	requiredRoomCount := minimumRequiredRooms(candidates, criteria.RoomCount, criteria.Adults, criteria.ChildrenAges, hotel.DefaultChildMaxAge)
	if requiredRoomCount == 0 {
		return false, false, minPrice, len(candidates), nil
	}

	strict = requiredRoomCount == criteria.RoomCount
	flexible = requiredRoomCount > criteria.RoomCount

	return strict, flexible, minPrice, len(candidates), nil
}

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

func (r *SearchRepository) loadRoomCandidates(ctx context.Context, hotel HotelModel, criteria search.Criteria) ([]roomCandidate, error) {
	var roomModels []RoomModel
	if err := r.db.WithContext(ctx).
		Where("hotel_id = ? AND status = ?", hotel.ID, "active").
		Find(&roomModels).Error; err != nil {
		return nil, err
	}

	if len(roomModels) == 0 {
		return nil, nil
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

	candidates := make([]roomCandidate, 0, len(roomModels))

	for i := range roomModels {
		available := availMap[roomModels[i].ID]
		if available <= 0 {
			continue
		}

		amenityIDs := amenityMap[roomModels[i].ID]
		if !hasAllAmenities(amenityIDs, criteria.AmenityIDs) {
			continue
		}

		candidates = append(candidates, roomCandidate{
			Room:           roomModels[i],
			AvailableCount: available,
			AmenityIDs:     amenityIDs,
		})
	}

	return candidates, nil
}

func (r *SearchRepository) roomAvailabilityByDateRange(ctx context.Context, roomIDs []string, checkIn, checkOut time.Time) (map[string]int, error) {
	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	if nights <= 0 {
		nights = 1
	}

	type availabilityRow struct {
		RoomID         string
		AvailableCount int
		DayCount       int
	}

	var rows []availabilityRow

	err := r.db.WithContext(ctx).
		Table("room_inventories").
		Select("room_id, MIN(total_inventory - held_inventory - booked_inventory) AS available_count, COUNT(*) AS day_count").
		Where("room_id IN ?", roomIDs).
		Where("date >= ? AND date < ?", checkIn, checkOut).
		Group("room_id").
		Having("COUNT(*) = ?", nights).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]int, len(rows))
	for i := range rows {
		result[rows[i].RoomID] = rows[i].AvailableCount
	}

	return result, nil
}

func (r *SearchRepository) roomAmenityIDs(ctx context.Context, roomIDs []string) (map[string][]string, error) {
	type row struct {
		RoomID    string
		AmenityID string
	}

	var rows []row

	err := r.db.WithContext(ctx).
		Table("room_amenity_maps").
		Select("room_id, amenity_id").
		Where("room_id IN ?", roomIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	idMap := make(map[string][]string)
	for i := range rows {
		idMap[rows[i].RoomID] = append(idMap[rows[i].RoomID], rows[i].AmenityID)
	}

	return idMap, nil
}

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

func splitGuestsAcrossRooms(roomCount, adults int, childrenAges []int) []guestRequest {
	if roomCount <= 0 {
		return nil
	}

	rooms := make([]guestRequest, roomCount)
	for i := 0; i < roomCount; i++ {
		rooms[i] = guestRequest{adults: 0, childrenAges: make([]int, 0)}
	}

	for i := 0; i < adults; i++ {
		rooms[i%roomCount].adults++
	}

	for i := range childrenAges {
		rooms[i%roomCount].childrenAges = append(rooms[i%roomCount].childrenAges, childrenAges[i])
	}

	return rooms
}

func canAllocateRequestedRooms(candidates []roomCandidate, requests []guestRequest, childMaxAge int) bool {
	if len(requests) == 0 {
		return true
	}

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
			return false
		}
	}

	sort.Slice(order, func(i, j int) bool { return len(compat[order[i]]) < len(compat[order[j]]) })

	remaining := make([]int, len(candidates))
	for i := range candidates {
		remaining[i] = candidates[i].AvailableCount
	}

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

	for roomCount := startRoomCount; roomCount <= totalAvailable; roomCount++ {
		requests := splitGuestsAcrossRooms(roomCount, adults, childrenAges)
		if canAllocateRequestedRooms(candidates, requests, childMaxAge) {
			return roomCount
		}
	}

	return 0
}

func candidateCanFit(c roomCandidate, req guestRequest, childMaxAge int) bool {
	adults, children := normalizeGuest(req, childMaxAge)
	total := adults + children

	return adults <= c.Room.MaxAdult && children <= c.Room.MaxChild && total <= c.Room.MaxOccupancy
}

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
