package room

import (
	"testing"
	"time"

	"hexagon/errs"

	"github.com/stretchr/testify/assert"
)

func TestRoomValidateID(t *testing.T) {
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(ValidateID("")))
	assert.NoError(t, ValidateID("r-1"))
}

func TestRoomStatus_IsValid(t *testing.T) {
	assert.True(t, RoomStatusActive.IsValid())
	assert.True(t, RoomStatusInactive.IsValid())
	assert.False(t, RoomStatus("unknown").IsValid())
}

func TestRoomValidateForCreate(t *testing.T) {
	base := Room{
		HotelID:      "h-1",
		Name:         "Room",
		BasePrice:    100,
		MaxAdult:     2,
		MaxChild:     1,
		MaxOccupancy: 3,
		Status:       RoomStatusActive,
		Images:       []RoomImage{{URL: "https://image"}},
	}

	assert.NoError(t, base.ValidateForCreate())

	bad := base
	bad.HotelID = ""
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = base
	bad.Name = ""
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = base
	bad.BasePrice = 0
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = base
	bad.MaxAdult = 0
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = base
	bad.MaxChild = -1
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = base
	bad.MaxOccupancy = 0
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = base
	bad.MaxOccupancy = 2
	bad.MaxAdult = 2
	bad.MaxChild = 1
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = base
	bad.Status = "invalid"
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = base
	bad.Images = nil
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = base
	bad.Images = []RoomImage{{URL: ""}}
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = base
	bad.AmenityIDs = []string{""}
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))
}

func TestRoomImageValidate(t *testing.T) {
	img := RoomImage{RoomID: "r-1", URL: "https://image"}
	assert.NoError(t, img.ValidateForCreate())

	img = RoomImage{RoomID: "", URL: "https://image"}
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(img.ValidateForCreate()))

	img = RoomImage{RoomID: "r-1", URL: ""}
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(img.ValidateForCreate()))
}

func TestRoomInventoryValidate(t *testing.T) {
	inv := RoomInventory{RoomID: "r-1", Date: time.Now().UTC(), TotalInventory: 10}
	assert.NoError(t, inv.ValidateForCreate())

	bad := inv
	bad.RoomID = ""
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = inv
	bad.Date = time.Time{}
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = inv
	bad.TotalInventory = -1
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = inv
	bad.HeldInventory = -1
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = inv
	bad.BookedInventory = -1
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))

	bad = inv
	bad.HeldInventory = 6
	bad.BookedInventory = 5
	bad.TotalInventory = 10
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.ValidateForCreate()))
}

func TestRoomAmenityValidate(t *testing.T) {
	amenity := RoomAmenity{Code: "wifi", Name: "WiFi"}
	assert.NoError(t, amenity.ValidateForCreate())

	amenity = RoomAmenity{Name: "WiFi"}
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(amenity.ValidateForCreate()))

	amenity = RoomAmenity{Code: "wifi"}
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(amenity.ValidateForCreate()))
}

func TestRoomAmenityMapValidate(t *testing.T) {
	m := RoomAmenityMap{RoomID: "r-1", AmenityID: "a-1"}
	assert.NoError(t, m.ValidateForCreate())

	m = RoomAmenityMap{AmenityID: "a-1"}
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(m.ValidateForCreate()))

	m = RoomAmenityMap{RoomID: "r-1"}
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(m.ValidateForCreate()))
}
