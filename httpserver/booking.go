package httpserver

import (
	"hexagon/booking"
	"hexagon/hotel"

	"github.com/labstack/echo/v4"
)

func (s *Server) RegisterBookingRoutes() {
	s.Router.POST("/api/bookings", s.handleCreateBooking)
	s.Router.GET("/api/bookings/:booking_id", s.handleGetBookingByID)
	s.Router.POST("/api/bookings/:booking_id/payment-option", s.handleSelectBookingPaymentOption)
	s.Router.POST("/api/bookings/:booking_id/confirm-payment", s.handleConfirmBookingPayment)
	s.Router.POST("/api/bookings/:booking_id/cancel", s.handleCancelBooking)
}

// handleCreateBooking godoc
// @Summary Create Booking
// @Description Create a booking, hold inventory, and return available payment options for the hotel.
// @Tags bookings
// @Accept json
// @Produce json
// @Param payload body CreateBookingRequest true "Booking payload"
// @Success 201 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 409 {object} APIErrorResponse
// @Router /api/bookings [post]
func (s *Server) handleCreateBooking(c echo.Context) error {
	var req CreateBookingRequest
	if err := c.Bind(&req); err != nil {
		return s.respondBadRequest(c, "invalid request body", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return s.respondBadRequest(c, "invalid request body", err.Error())
	}

	checkIn, err := isoDate(req.CheckInAt)
	if err != nil {
		return s.respondBadRequest(c, "invalid check-in date", err.Error())
	}

	checkOut, err := isoDate(req.CheckOutAt)
	if err != nil {
		return s.respondBadRequest(c, "invalid check-out date", err.Error())
	}

	checkout, err := s.BookingService.CreateBooking(c.Request().Context(), booking.CreateRequest{
		RoomID:       req.RoomID,
		CheckInDate:  checkIn,
		CheckOutDate: checkOut,
		RoomCount:    req.RoomCount,
		GuestCount:   req.GuestCount,
	})
	if err != nil {
		return err
	}

	return s.respondCreated(c, toBookingCheckoutResponse(checkout))
}

// handleGetBookingByID godoc
// @Summary Get Booking By ID
// @Description Get booking details.
// @Tags bookings
// @Produce json
// @Param booking_id path string true "Booking ID"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 404 {object} APIErrorResponse
// @Router /api/bookings/{booking_id} [get]
func (s *Server) handleGetBookingByID(c echo.Context) error {
	bookingID := c.Param("booking_id")
	if bookingID == "" {
		return s.respondBadRequest(c, "invalid booking id", "booking_id is required")
	}

	result, err := s.BookingService.GetBookingByID(c.Request().Context(), bookingID)
	if err != nil {
		return err
	}

	return s.respondOK(c, APIDataResult{Data: toBookingResponse(result)})
}

// handleSelectBookingPaymentOption godoc
// @Summary Select Booking Payment Option
// @Description Select payment option for a pending booking.
// @Tags bookings
// @Accept json
// @Produce json
// @Param booking_id path string true "Booking ID"
// @Param payload body SelectBookingPaymentOptionRequest true "Payment option"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 409 {object} APIErrorResponse
// @Router /api/bookings/{booking_id}/payment-option [post]
func (s *Server) handleSelectBookingPaymentOption(c echo.Context) error {
	bookingID := c.Param("booking_id")
	if bookingID == "" {
		return s.respondBadRequest(c, "invalid booking id", "booking_id is required")
	}

	var req SelectBookingPaymentOptionRequest
	if err := c.Bind(&req); err != nil {
		return s.respondBadRequest(c, "invalid request body", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return s.respondBadRequest(c, "invalid request body", err.Error())
	}

	updated, err := s.BookingService.SelectPaymentOption(
		c.Request().Context(),
		bookingID,
		hotel.PaymentOption(req.PaymentOption),
	)
	if err != nil {
		return err
	}

	return s.respondOK(c, APIDataResult{Data: toBookingResponse(updated)})
}

// handleConfirmBookingPayment godoc
// @Summary Confirm Booking Payment
// @Description Mark booking payment as successful and confirm the booking.
// @Tags bookings
// @Produce json
// @Param booking_id path string true "Booking ID"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 409 {object} APIErrorResponse
// @Router /api/bookings/{booking_id}/confirm-payment [post]
func (s *Server) handleConfirmBookingPayment(c echo.Context) error {
	bookingID := c.Param("booking_id")
	if bookingID == "" {
		return s.respondBadRequest(c, "invalid booking id", "booking_id is required")
	}

	updated, err := s.BookingService.MarkPaymentPaid(c.Request().Context(), bookingID)
	if err != nil {
		return err
	}

	return s.respondOK(c, APIDataResult{Data: toBookingResponse(updated)})
}

// handleCancelBooking godoc
// @Summary Cancel Booking
// @Description Cancel booking if status allows; release inventory and optionally calculate refund.
// @Tags bookings
// @Accept json
// @Produce json
// @Param booking_id path string true "Booking ID"
// @Param payload body CancelBookingRequest true "Cancellation data"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 409 {object} APIErrorResponse
// @Router /api/bookings/{booking_id}/cancel [post]
func (s *Server) handleCancelBooking(c echo.Context) error {
	bookingID := c.Param("booking_id")
	if bookingID == "" {
		return s.respondBadRequest(c, "invalid booking id", "booking_id is required")
	}

	var req CancelBookingRequest
	if err := c.Bind(&req); err != nil {
		return s.respondBadRequest(c, "invalid request body", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return s.respondBadRequest(c, "invalid request body", err.Error())
	}

	updated, err := s.BookingService.CancelBooking(c.Request().Context(), bookingID, req.CancellationFee)
	if err != nil {
		return err
	}

	return s.respondOK(c, APIDataResult{Data: toBookingResponse(updated)})
}
