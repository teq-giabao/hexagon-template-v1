package httpserver

import (
	"strings"

	"github.com/labstack/echo/v4"
)

func (s *Server) RegisterSearchRoutes() {
	s.Router.POST("/api/search/hotels", s.handleSearchHotels)
	s.Router.POST("/api/search/hotels/:hotel_id/rooms", s.handleSearchHotelRooms)
	s.Router.POST("/api/search/hotels/:hotel_id/room-combinations", s.handleSearchHotelRoomCombinations)
}

// handleSearchHotels godoc
// @Summary Search Hotels
// @Description Search hotels by name/location, date range, occupancy, and filters. Returns hotels that can satisfy requested room count.
// @Tags search
// @Accept json
// @Produce json
// @Param payload body SearchHotelsRequest true "Hotel search criteria"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/search/hotels [post]
func (s *Server) handleSearchHotels(c echo.Context) error {
	var req SearchHotelsRequest
	if err := c.Bind(&req); err != nil {
		return s.respondBadRequest(c, "invalid request body", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return s.respondBadRequest(c, "invalid request body", err.Error())
	}

	checkIn, err := isoDate(req.CheckInAt)
	if err != nil {
		return s.respondBadRequest(c, "invalid date range", err.Error())
	}

	checkOut, err := isoDate(req.CheckOutAt)
	if err != nil {
		return s.respondBadRequest(c, "invalid date range", err.Error())
	}

	result, err := s.SearchService.SearchHotels(c.Request().Context(), req.ToCriteria(checkIn, checkOut))
	if err != nil {
		return err
	}

	pageSize := req.EffectivePageSize()
	offset := req.EffectiveOffset()
	total := len(result)
	start := offset

	if start > total {
		start = total
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	page := req.EffectivePageFromOffset(start, pageSize)

	return s.respondOK(c, APIDataResult{Data: toSearchHotelsResponse(result[start:end], page, pageSize, start, total)})
}

// handleSearchHotelRooms godoc
// @Summary Search Hotel Rooms
// @Description Search available room types in a hotel by date range, occupancy, and room amenity filters.
// @Tags search
// @Accept json
// @Produce json
// @Param hotel_id path string true "Hotel ID"
// @Param payload body SearchHotelRoomsRequest true "Hotel room search criteria"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 404 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/search/hotels/{hotel_id}/rooms [post]
func (s *Server) handleSearchHotelRooms(c echo.Context) error {
	hotelID := strings.TrimSpace(c.Param("hotel_id"))
	if hotelID == "" {
		return s.respondBadRequest(c, "invalid request path", "hotel_id is required")
	}

	var req SearchHotelRoomsRequest
	if err := c.Bind(&req); err != nil {
		return s.respondBadRequest(c, "invalid request body", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return s.respondBadRequest(c, "invalid request body", err.Error())
	}

	checkIn, err := isoDate(req.CheckInAt)
	if err != nil {
		return s.respondBadRequest(c, "invalid date range", err.Error())
	}

	checkOut, err := isoDate(req.CheckOutAt)
	if err != nil {
		return s.respondBadRequest(c, "invalid date range", err.Error())
	}

	result, err := s.SearchService.SearchHotelRooms(c.Request().Context(), hotelID, req.ToCriteria(checkIn, checkOut))
	if err != nil {
		return err
	}

	return s.respondOK(c, APIDataResult{Data: toSearchHotelRoomsResponse(result)})
}

// handleSearchHotelRoomCombinations godoc
// @Summary Search Hotel Room Combinations
// @Description Return purchasable room combinations for a hotel under requested occupancy and date constraints.
// @Tags search
// @Accept json
// @Produce json
// @Param hotel_id path string true "Hotel ID"
// @Param payload body SearchHotelRoomCombinationsRequest true "Hotel room combinations criteria"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 404 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/search/hotels/{hotel_id}/room-combinations [post]
func (s *Server) handleSearchHotelRoomCombinations(c echo.Context) error {
	hotelID := strings.TrimSpace(c.Param("hotel_id"))
	if hotelID == "" {
		return s.respondBadRequest(c, "invalid request path", "hotel_id is required")
	}

	var req SearchHotelRoomCombinationsRequest
	if err := c.Bind(&req); err != nil {
		return s.respondBadRequest(c, "invalid request body", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return s.respondBadRequest(c, "invalid request body", err.Error())
	}

	checkIn, err := isoDate(req.CheckInAt)
	if err != nil {
		return s.respondBadRequest(c, "invalid date range", err.Error())
	}

	checkOut, err := isoDate(req.CheckOutAt)
	if err != nil {
		return s.respondBadRequest(c, "invalid date range", err.Error())
	}

	result, err := s.SearchService.SearchHotelRoomCombinations(
		c.Request().Context(),
		hotelID,
		req.ToCriteria(checkIn, checkOut),
		req.MaxCombinations,
	)
	if err != nil {
		return err
	}

	return s.respondOK(c, APIDataResult{Data: toSearchHotelRoomCombinationsResponse(result)})
}
