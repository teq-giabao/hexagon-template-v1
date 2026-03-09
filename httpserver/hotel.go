package httpserver

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (s *Server) RegisterHotelRoutes() {
	s.Router.GET("/api/hotels", s.handleListHotels)
	s.Router.GET("/api/hotels/:hotel_id", s.handleGetHotelByID)
	s.Router.POST("/api/hotels", s.handleAddHotel)
	s.Router.POST("/api/hotels/upload-images", s.handleUploadHotelImages)
}

// handleListHotels godoc
// @Summary List Hotels
// @Description Get all hotels
// @Tags hotels
// @Produce json
// @Success 200 {object} APISuccessResponse
// @Router /api/hotels [get]
func (s *Server) handleListHotels(c echo.Context) error {
	hotels, err := s.HotelService.ListHotels(c.Request().Context())
	if err != nil {
		return err
	}

	return respondOK(c, APIDataResult{Data: toHotelResponses(hotels)})
}

// handleGetHotelByID godoc
// @Summary Get Hotel By ID
// @Description Get hotel detail by id
// @Tags hotels
// @Produce json
// @Param hotel_id path string true "Hotel ID"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 404 {object} APIErrorResponse
// @Router /api/hotels/{hotel_id} [get]
func (s *Server) handleGetHotelByID(c echo.Context) error {
	h, err := s.HotelService.GetHotelByID(c.Request().Context(), c.Param("hotel_id"))
	if err != nil {
		return err
	}

	return respondOK(c, toHotelResponse(h))
}

// handleAddHotel godoc
// @Summary Create Hotel
// @Description Create a new hotel. `rating` is initialized by system (default 0), and `defaultChildMaxAge` defaults to 11 when omitted.
// @Tags hotels
// @Accept json
// @Produce json
// @Param payload body AddHotelRequest true "Hotel payload"
// @Success 201 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Router /api/hotels [post]
func (s *Server) handleAddHotel(c echo.Context) error {
	var req AddHotelRequest
	if err := c.Bind(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}

	checkIn, checkOut, err := parseHotelTimes(req.CheckInTime, req.CheckOutTime)
	if err != nil {
		return respondError(c, http.StatusBadRequest, "invalid check-in/check-out time", err.Error())
	}

	created, err := s.HotelService.AddHotel(c.Request().Context(), req.ToHotel(checkIn, checkOut))
	if err != nil {
		return err
	}

	return respondCreated(c, toHotelResponse(created))
}

// handleUploadHotelImages godoc
// @Summary Upload Hotel Images
// @Description Upload images to S3 using multipart/form-data field `images`. To upload multiple files, send repeated `images` fields. Swagger 2 UI can only pick one file at a time for this endpoint.
// @Tags hotels
// @Accept multipart/form-data
// @Produce json
// @Param images formData []file true "Image file. Repeat this field to upload multiple files (e.g. -F \"images=@a.jpg\" -F \"images=@b.jpg\")"
// @Success 201 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 501 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/hotels/upload-images [post]
func (s *Server) handleUploadHotelImages(c echo.Context) error {
	return s.handleUploadImages(c, "hotel-images")
}

func parseHotelTimes(checkIn, checkOut string) (time.Time, time.Time, error) {
	checkInTime, err := parseClockTime(checkIn)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	checkOutTime, err := parseClockTime(checkOut)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return checkInTime, checkOutTime, nil
}

func parseClockTime(value string) (time.Time, error) {
	if t, err := time.Parse("15:04:05", value); err == nil {
		return t, nil
	}

	return time.Parse("15:04", value)
}
