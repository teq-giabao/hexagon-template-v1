package httpserver

import (
	"strings"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
)

type RequestValidator struct {
	validator *validator.Validate
}

func NewRequestValidator() *RequestValidator {
	v := validator.New()
	_ = v.RegisterValidation("notblank", validateNotBlank)
	_ = v.RegisterValidation("password", validatePassword)
	_ = v.RegisterValidation("date_not_past", validateDateNotPast)
	_ = v.RegisterValidation("date_within_booking_window", validateDateWithinBookingWindow)
	_ = v.RegisterValidation("checkout_after_checkin", validateCheckoutAfterCheckin)

	return &RequestValidator{
		validator: v,
	}
}

func (v *RequestValidator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

func validateNotBlank(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	return strings.TrimSpace(value) != ""
}

func validatePassword(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	value = strings.TrimSpace(value)
	if len(value) < 9 || len(value) > 72 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, r := range value {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsNumber(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func validateDateNotPast(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	t, err := parseISODate(value)
	if err != nil {
		return false
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return !t.Before(today)
}

func validateDateWithinBookingWindow(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	t, err := parseISODate(value)
	if err != nil {
		return false
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	maxDate := today.AddDate(2, 0, 0)
	return !t.After(maxDate)
}

func validateCheckoutAfterCheckin(fl validator.FieldLevel) bool {
	checkOut, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	checkIn := fl.Parent().FieldByName("CheckInAt").String()
	if strings.TrimSpace(checkIn) == "" || strings.TrimSpace(checkOut) == "" {
		return false
	}
	checkInDate, err := parseISODate(checkIn)
	if err != nil {
		return false
	}
	checkOutDate, err := parseISODate(checkOut)
	if err != nil {
		return false
	}
	return checkOutDate.After(checkInDate)
}
