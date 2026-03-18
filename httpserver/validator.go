package httpserver

import (
	"log/slog"
	"strings"
	"time"

	"hexagon/user"

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
		slog.Error("validation failed", "tag", "notblank", "reason", "field is not string")
		return false
	}

	valid := strings.TrimSpace(value) != ""
	if !valid {
		slog.Error("validation failed", "tag", "notblank", "reason", "blank string")
	}

	return valid
}

func validatePassword(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		slog.Error("validation failed", "tag", "password", "reason", "field is not string")
		return false
	}

	if err := user.ValidatePassword(value); err != nil {
		slog.Error("validation failed", "tag", "password", "reason", err.Error())
		return false
	}

	return true
}

func validateDateNotPast(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		slog.Error("validation failed", "tag", "date_not_past", "reason", "field is not string")
		return false
	}

	t, err := isoDate(value)
	if err != nil {
		slog.Error("validation failed", "tag", "date_not_past", "reason", "invalid date format", "value", value)
		return false
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	valid := !t.Before(today)
	if !valid {
		slog.Error("validation failed", "tag", "date_not_past", "reason", "date is in the past", "value", value)
	}

	return valid
}

func validateDateWithinBookingWindow(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		slog.Error("validation failed", "tag", "date_within_booking_window", "reason", "field is not string")
		return false
	}

	t, err := isoDate(value)
	if err != nil {
		slog.Error("validation failed", "tag", "date_within_booking_window", "reason", "invalid date format", "value", value)
		return false
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	maxDate := today.AddDate(2, 0, 0)

	valid := !t.After(maxDate)
	if !valid {
		slog.Error("validation failed", "tag", "date_within_booking_window", "reason", "date exceeds max booking window", "value", value)
	}

	return valid
}

func validateCheckoutAfterCheckin(fl validator.FieldLevel) bool {
	checkOut, ok := fl.Field().Interface().(string)
	if !ok {
		slog.Error("validation failed", "tag", "checkout_after_checkin", "reason", "checkout field is not string")
		return false
	}

	checkIn := fl.Parent().FieldByName("CheckInAt").String()
	if strings.TrimSpace(checkIn) == "" || strings.TrimSpace(checkOut) == "" {
		slog.Error("validation failed", "tag", "checkout_after_checkin", "reason", "missing checkin/checkout value")
		return false
	}

	checkInDate, err := isoDate(checkIn)
	if err != nil {
		slog.Error("validation failed", "tag", "checkout_after_checkin", "reason", "invalid checkin date", "value", checkIn)
		return false
	}

	checkOutDate, err := isoDate(checkOut)
	if err != nil {
		slog.Error("validation failed", "tag", "checkout_after_checkin", "reason", "invalid checkout date", "value", checkOut)
		return false
	}

	valid := checkOutDate.After(checkInDate)
	if !valid {
		slog.Error("validation failed", "tag", "checkout_after_checkin", "reason", "checkout is not after checkin", "checkInAt", checkIn, "checkOutAt", checkOut)
	}

	return valid
}
