package httpserver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type validatorPayload struct {
	Name      string `validate:"required,notblank"`
	Password  string `validate:"required,password"`
	CheckInAt string `validate:"required,datetime=2006-01-02,date_not_past,date_within_booking_window"`
	CheckOutAt string `validate:"required,datetime=2006-01-02,checkout_after_checkin"`
}

func TestRequestValidator_Success(t *testing.T) {
	v := NewRequestValidator()
	today := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	tomorrow := time.Now().Add(48 * time.Hour).Format("2006-01-02")

	payload := validatorPayload{
		Name:      "John",
		Password:  "Password123!",
		CheckInAt: today,
		CheckOutAt: tomorrow,
	}

	err := v.Validate(payload)
	assert.NoError(t, err)
}

func TestRequestValidator_InvalidPassword(t *testing.T) {
	v := NewRequestValidator()
	payload := validatorPayload{
		Name:      "John",
		Password:  "short",
		CheckInAt: time.Now().Add(24 * time.Hour).Format("2006-01-02"),
		CheckOutAt: time.Now().Add(48 * time.Hour).Format("2006-01-02"),
	}

	err := v.Validate(payload)
	assert.Error(t, err)
}

func TestRequestValidator_DateRules(t *testing.T) {
	v := NewRequestValidator()
	past := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	farFuture := time.Now().AddDate(3, 0, 0).Format("2006-01-02")

	payload := validatorPayload{
		Name:      "John",
		Password:  "Password123!",
		CheckInAt: past,
		CheckOutAt: farFuture,
	}

	err := v.Validate(payload)
	assert.Error(t, err)
}

func TestRequestValidator_CheckoutAfterCheckin(t *testing.T) {
	v := NewRequestValidator()
	checkIn := time.Now().Add(48 * time.Hour).Format("2006-01-02")
	checkOut := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	payload := validatorPayload{
		Name:      "John",
		Password:  "Password123!",
		CheckInAt: checkIn,
		CheckOutAt: checkOut,
	}

	err := v.Validate(payload)
	assert.Error(t, err)
}
