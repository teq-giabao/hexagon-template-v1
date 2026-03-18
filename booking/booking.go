package booking

import (
	"strings"
	"time"

	"hexagon/errs"
	"hexagon/hotel"
)

var (
	ErrBookingIDRequired     = errs.Errorf(errs.EINVALID, "booking: id is required")
	ErrBookingNotFound       = errs.Errorf(errs.ENOTFOUND, "booking: not found")
	ErrUserIDRequired        = errs.Errorf(errs.EINVALID, "booking: user id is required")
	ErrRoomIDRequired        = errs.Errorf(errs.EINVALID, "booking: room id is required")
	ErrRoomNotFound          = errs.Errorf(errs.ENOTFOUND, "booking: room not found")
	ErrCheckInRequired       = errs.Errorf(errs.EINVALID, "booking: check-in date is required")
	ErrCheckOutRequired      = errs.Errorf(errs.EINVALID, "booking: check-out date is required")
	ErrCheckOutBeforeCheckIn = errs.Errorf(errs.EINVALID, "booking: check-out must be after check-in")
	ErrRoomCountInvalid      = errs.Errorf(errs.EINVALID, "booking: room count must be greater than 0")
	ErrGuestCountInvalid     = errs.Errorf(errs.EINVALID, "booking: guest count must be greater than 0")
	ErrRoomUnavailable       = errs.Errorf(errs.ECONFLICT, "booking: room is not available")
	ErrBookingExpired        = errs.Errorf(errs.ECONFLICT, "booking: booking has expired")
	ErrBookingCancelled      = errs.Errorf(errs.ECONFLICT, "booking: booking has been cancelled")
	ErrBookingNotPending     = errs.Errorf(errs.ECONFLICT, "booking: booking is not pending")
	ErrBookingNotConfirmed   = errs.Errorf(errs.ECONFLICT, "booking: booking is not confirmed")
	ErrPaymentOptionInvalid  = errs.Errorf(errs.EINVALID, "booking: invalid payment option")
	ErrPaymentStatusInvalid  = errs.Errorf(errs.EINVALID, "booking: invalid payment status")
	ErrCancellationFee       = errs.Errorf(errs.EINVALID, "booking: cancellation fee is invalid")
)

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusExpired   BookingStatus = "expired"
)

type PaymentStatus string

const (
	PaymentStatusUnpaid   PaymentStatus = "unpaid"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

type CreateRequest struct {
	UserID       string
	RoomID       string
	CheckInDate  time.Time
	CheckOutDate time.Time
	RoomCount    int
	GuestCount   int
}

func (r CreateRequest) Validate() error {
	if strings.TrimSpace(r.UserID) == "" {
		return ErrUserIDRequired
	}

	if strings.TrimSpace(r.RoomID) == "" {
		return ErrRoomIDRequired
	}

	if r.CheckInDate.IsZero() {
		return ErrCheckInRequired
	}

	if r.CheckOutDate.IsZero() {
		return ErrCheckOutRequired
	}

	if !r.CheckOutDate.After(r.CheckInDate) {
		return ErrCheckOutBeforeCheckIn
	}

	if r.RoomCount <= 0 {
		return ErrRoomCountInvalid
	}

	if r.GuestCount <= 0 {
		return ErrGuestCountInvalid
	}

	return nil
}

type Booking struct {
	ID              string
	UserID          string
	HotelID         string
	RoomID          string
	CheckInDate     time.Time
	CheckOutDate    time.Time
	Nights          int
	RoomCount       int
	GuestCount      int
	NightlyPrice    float64
	TotalPrice      float64
	Status          BookingStatus
	PaymentOption   hotel.PaymentOption
	PaymentStatus   PaymentStatus
	HoldExpiresAt   *time.Time
	PaymentDeadline *time.Time
	CancelledAt     *time.Time
	CancellationFee float64
	RefundAmount    float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (s BookingStatus) IsValid() bool {
	return s == BookingStatusPending || s == BookingStatusConfirmed || s == BookingStatusCancelled || s == BookingStatusExpired
}

func (s PaymentStatus) IsValid() bool {
	return s == PaymentStatusUnpaid || s == PaymentStatusPaid || s == PaymentStatusRefunded
}

func ValidateID(id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrBookingIDRequired
	}

	return nil
}

func ValidatePaymentOption(option hotel.PaymentOption) error {
	if !option.IsValid() {
		return ErrPaymentOptionInvalid
	}

	return nil
}
