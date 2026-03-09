package room

import (
	"encoding/json"
	"strings"
	"time"

	"hexagon/errs"
)

var (
	ErrRoomIDRequired               = errs.Errorf(errs.EINVALID, "room: id is required")
	ErrRoomNotFound                 = errs.Errorf(errs.ENOTFOUND, "room: not found")
	ErrHotelIDRequired              = errs.Errorf(errs.EINVALID, "room: hotel id is required")
	ErrNameRequired                 = errs.Errorf(errs.EINVALID, "room: name is required")
	ErrRoomImagesRequired           = errs.Errorf(errs.EINVALID, "room: at least one image is required")
	ErrBasePriceInvalid             = errs.Errorf(errs.EINVALID, "room: base price must be greater than 0")
	ErrMaxAdultInvalid              = errs.Errorf(errs.EINVALID, "room: max adult must be greater than 0")
	ErrMaxChildInvalid              = errs.Errorf(errs.EINVALID, "room: max child must be greater or equal to 0")
	ErrMaxOccupancyInvalid          = errs.Errorf(errs.EINVALID, "room: max occupancy must be greater than 0")
	ErrMaxOccupancyLessThanCapacity = errs.Errorf(errs.EINVALID, "room: max occupancy must be >= max adult + max child")
	ErrStatusInvalid                = errs.Errorf(errs.EINVALID, "room: invalid room status")
	ErrRoomImageIDRequired          = errs.Errorf(errs.EINVALID, "room: image id is required")
	ErrInventoryIDRequired          = errs.Errorf(errs.EINVALID, "room: inventory id is required")
	ErrDateRequired                 = errs.Errorf(errs.EINVALID, "room: date is required")
	ErrInventoryTotalInvalid        = errs.Errorf(errs.EINVALID, "room: total inventory must be >= 0")
	ErrInventoryHeldInvalid         = errs.Errorf(errs.EINVALID, "room: held inventory must be >= 0")
	ErrInventoryBookedInvalid       = errs.Errorf(errs.EINVALID, "room: booked inventory must be >= 0")
	ErrInventoryOverTotal           = errs.Errorf(errs.EINVALID, "room: held + booked inventory must be <= total inventory")
	ErrImageURLRequired             = errs.Errorf(errs.EINVALID, "room: image url is required")
	ErrAmenityIDRequired            = errs.Errorf(errs.EINVALID, "room: amenity id is required")
	ErrAmenityCodeRequired          = errs.Errorf(errs.EINVALID, "room: amenity code is required")
	ErrAmenityNameRequired          = errs.Errorf(errs.EINVALID, "room: amenity name is required")
)

type RoomStatus string

const (
	RoomStatusActive   RoomStatus = "active"
	RoomStatusInactive RoomStatus = "inactive"
)

type Room struct {
	ID           string
	HotelID      string
	Name         string
	Description  string
	BasePrice    float64
	MaxAdult     int
	MaxChild     int
	MaxOccupancy int
	BedOptions   json.RawMessage
	SizeSqm      int
	Status       RoomStatus
	Images       []RoomImage
	AmenityIDs   []string
	Amenities    []RoomAmenity
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RoomImage struct {
	ID      string
	RoomID  string
	URL     string
	IsCover bool
}

type RoomInventory struct {
	ID              string
	RoomID          string
	Date            time.Time
	TotalInventory  int
	HeldInventory   int
	BookedInventory int
}

type RoomAmenity struct {
	ID          string
	Code        string
	Name        string
	Description string
	Icon        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RoomAmenityMap struct {
	ID        string
	RoomID    string
	AmenityID string
	CreatedAt time.Time
}

func ValidateID(id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrRoomIDRequired
	}

	return nil
}

func (s RoomStatus) IsValid() bool {
	return s == RoomStatusActive || s == RoomStatusInactive
}

func (r Room) ValidateForCreate() error {
	if strings.TrimSpace(r.HotelID) == "" {
		return ErrHotelIDRequired
	}

	if strings.TrimSpace(r.Name) == "" {
		return ErrNameRequired
	}

	if r.BasePrice <= 0 {
		return ErrBasePriceInvalid
	}

	if r.MaxAdult <= 0 {
		return ErrMaxAdultInvalid
	}

	if r.MaxChild < 0 {
		return ErrMaxChildInvalid
	}

	if r.MaxOccupancy <= 0 {
		return ErrMaxOccupancyInvalid
	}

	if r.MaxOccupancy < r.MaxAdult+r.MaxChild {
		return ErrMaxOccupancyLessThanCapacity
	}

	if !r.Status.IsValid() {
		return ErrStatusInvalid
	}

	if len(r.Images) == 0 {
		return ErrRoomImagesRequired
	}

	for i := range r.Images {
		if err := r.Images[i].ValidateForRoomCreate(); err != nil {
			return err
		}
	}

	for i := range r.AmenityIDs {
		if strings.TrimSpace(r.AmenityIDs[i]) == "" {
			return ErrAmenityIDRequired
		}
	}

	return nil
}

func (img RoomImage) ValidateForCreate() error {
	if strings.TrimSpace(img.RoomID) == "" {
		return ErrRoomIDRequired
	}

	return img.ValidateForRoomCreate()
}

func (img RoomImage) ValidateForRoomCreate() error {
	if strings.TrimSpace(img.URL) == "" {
		return ErrImageURLRequired
	}

	return nil
}

func (inv RoomInventory) ValidateForCreate() error {
	if strings.TrimSpace(inv.RoomID) == "" {
		return ErrRoomIDRequired
	}

	if inv.Date.IsZero() {
		return ErrDateRequired
	}

	if inv.TotalInventory < 0 {
		return ErrInventoryTotalInvalid
	}

	if inv.HeldInventory < 0 {
		return ErrInventoryHeldInvalid
	}

	if inv.BookedInventory < 0 {
		return ErrInventoryBookedInvalid
	}

	if inv.HeldInventory+inv.BookedInventory > inv.TotalInventory {
		return ErrInventoryOverTotal
	}

	return nil
}

func (a RoomAmenity) ValidateForCreate() error {
	if strings.TrimSpace(a.Code) == "" {
		return ErrAmenityCodeRequired
	}

	if strings.TrimSpace(a.Name) == "" {
		return ErrAmenityNameRequired
	}

	return nil
}

func (m RoomAmenityMap) ValidateForCreate() error {
	if strings.TrimSpace(m.RoomID) == "" {
		return ErrRoomIDRequired
	}

	if strings.TrimSpace(m.AmenityID) == "" {
		return ErrAmenityIDRequired
	}

	return nil
}
