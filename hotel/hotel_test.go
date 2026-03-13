package hotel

import (
	"testing"

	"hexagon/errs"

	"github.com/stretchr/testify/assert"
)

func TestValidateID(t *testing.T) {
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(ValidateID("")))
	assert.NoError(t, ValidateID("h-1"))
}

func TestHotelValidateForCreate(t *testing.T) {
	base := Hotel{
		Name:    "Hotel",
		Address: "Address",
		City:    "City",
		Rating:  4.2,
		PaymentOptions: []HotelPaymentOption{
			{PaymentOption: PaymentOptionImmediate},
		},
	}

	assert.NoError(t, base.ValidateForCreate())

	bad := base
	bad.Name = ""
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = base
	bad.Address = ""
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = base
	bad.City = ""
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = base
	bad.Rating = 6
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = base
	bad.PaymentOptions = []HotelPaymentOption{{PaymentOption: "unknown"}}
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))
}

func TestPaymentOption_IsValid(t *testing.T) {
	assert.True(t, PaymentOptionImmediate.IsValid())
	assert.True(t, PaymentOptionPayAtHotel.IsValid())
	assert.True(t, PaymentOptionDeferred.IsValid())
	assert.False(t, PaymentOption("invalid").IsValid())
}
