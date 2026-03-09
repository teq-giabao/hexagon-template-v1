package hotel

import (
	"strings"
	"time"

	"hexagon/errs"
)

var (
	ErrHotelIDRequired = errs.Errorf(errs.EINVALID, "hotel: id is required")
	ErrHotelNotFound   = errs.Errorf(errs.ENOTFOUND, "hotel: not found")
	ErrNameRequired    = errs.Errorf(errs.EINVALID, "hotel: name is required")
	ErrAddressRequired = errs.Errorf(errs.EINVALID, "hotel: address is required")
	ErrCityRequired    = errs.Errorf(errs.EINVALID, "hotel: city is required")
	ErrRatingInvalid   = errs.Errorf(errs.EINVALID, "hotel: rating must be between 0 and 5")
	ErrPaymentOption   = errs.Errorf(errs.EINVALID, "hotel: invalid payment option")
)

type PaymentOption string

const (
	PaymentOptionImmediate  PaymentOption = "immediate"
	PaymentOptionPayAtHotel PaymentOption = "pay_at_hotel"
	PaymentOptionDeferred   PaymentOption = "deferred"
)

type Hotel struct {
	ID                 string
	Name               string
	Description        string
	Address            string
	City               string
	Rating             float64
	CheckInTime        time.Time
	CheckOutTime       time.Time
	DefaultChildMaxAge int
	Images             []HotelImage
	PaymentOptions     []HotelPaymentOption
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type HotelImage struct {
	ID      string
	HotelID string
	URL     string
	IsCover bool
}

type HotelPaymentOption struct {
	ID            string
	HotelID       string
	PaymentOption PaymentOption
	Enabled       bool
}

func ValidateID(id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrHotelIDRequired
	}

	return nil
}

func (h Hotel) ValidateForCreate() error {
	if strings.TrimSpace(h.Name) == "" {
		return ErrNameRequired
	}

	if strings.TrimSpace(h.Address) == "" {
		return ErrAddressRequired
	}

	if strings.TrimSpace(h.City) == "" {
		return ErrCityRequired
	}

	if h.Rating < 0 || h.Rating > 5 {
		return ErrRatingInvalid
	}

	for i := range h.PaymentOptions {
		if !h.PaymentOptions[i].PaymentOption.IsValid() {
			return ErrPaymentOption
		}
	}

	return nil
}

func (p PaymentOption) IsValid() bool {
	return p == PaymentOptionImmediate || p == PaymentOptionPayAtHotel || p == PaymentOptionDeferred
}
