package httpserver

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) RegisterSearchRoutes() {
	s.Router.POST("/api/search/hotels", s.handleSearchHotels)
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
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}

	checkIn, checkOut, err := parseISODateRange(req.CheckInAt, req.CheckOutAt)
	if err != nil {
		return respondError(c, http.StatusBadRequest, "invalid date range", err.Error())
	}

	result, err := s.SearchService.SearchHotels(c.Request().Context(), req.ToCriteria(checkIn, checkOut))
	if err != nil {
		return err
	}
	return respondOK(c, APIDataResult{Data: toSearchHotelsResponse(result)})
}
