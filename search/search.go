package search

import (
	"strings"
	"time"

	"hexagon/errs"
)

var (
	ErrCheckInRequired   = errs.Errorf(errs.EINVALID, "search: check-in date is required")
	ErrCheckOutRequired  = errs.Errorf(errs.EINVALID, "search: check-out date is required")
	ErrInvalidDateRange  = errs.Errorf(errs.EINVALID, "search: check-out date must be after check-in date")
	ErrAdultCountInvalid = errs.Errorf(errs.EINVALID, "search: adults must be greater than 0")
	ErrChildAgeInvalid   = errs.Errorf(errs.EINVALID, "search: child age must be between 0 and 17")
	ErrRoomCountInvalid  = errs.Errorf(errs.EINVALID, "search: room count must be greater than 0")
	ErrHotelNotFound     = errs.Errorf(errs.ENOTFOUND, "search: hotel not found")
)

type Criteria struct {
	Query          string
	CheckInDate    time.Time
	CheckOutDate   time.Time
	Adults         int
	ChildrenAges   []int
	RoomCount      int
	RatingMin      float64
	AmenityIDs     []string
	PaymentOptions []string
}

type HotelSearchResult struct {
	HotelID            string
	Name               string
	City               string
	Address            string
	Rating             float64
	PaymentOptions     []string
	MinPrice           float64
	AvailableRoomCount int
	MatchesRequested   bool
	FlexibleMatch      bool
}

type HotelRoomSearchResult struct {
	HotelID            string
	RequestedRoomCount int
	StrictMatch        bool
	Rooms              []HotelRoomSearchItem
}

type HotelRoomSearchItem struct {
	RoomID         string
	Name           string
	Description    string
	BasePrice      float64
	MaxAdult       int
	MaxChild       int
	MaxOccupancy   int
	AvailableCount int
	AmenityIDs     []string
	AmenityCodes   []string
	AmenityNames   []string
}

type HotelRoomCombinationsResult struct {
	HotelID            string
	RequestedRoomCount int
	Combinations       []RoomCombination
}

type RoomCombination struct {
	Items          []RoomCombinationItem
	TotalPrice     float64
	TotalRooms     int
	TotalMaxAdult  int
	TotalMaxChild  int
	TotalOccupancy int
}

type RoomCombinationItem struct {
	RoomID    string
	RoomName  string
	Quantity  int
	UnitPrice float64
	Subtotal  float64
}

func (c Criteria) Validate() error {
	if c.CheckInDate.IsZero() {
		return ErrCheckInRequired
	}

	if c.CheckOutDate.IsZero() {
		return ErrCheckOutRequired
	}

	if !c.CheckOutDate.After(c.CheckInDate) {
		return ErrInvalidDateRange
	}

	if c.Adults <= 0 {
		return ErrAdultCountInvalid
	}

	for i := range c.ChildrenAges {
		age := c.ChildrenAges[i]
		if age < 0 || age > 17 {
			return ErrChildAgeInvalid
		}
	}

	if c.RoomCount <= 0 {
		return ErrRoomCountInvalid
	}

	for i := range c.AmenityIDs {
		if strings.TrimSpace(c.AmenityIDs[i]) == "" {
			return errs.Errorf(errs.EINVALID, "search: amenity id is required")
		}
	}

	for i := range c.PaymentOptions {
		if strings.TrimSpace(c.PaymentOptions[i]) == "" {
			return errs.Errorf(errs.EINVALID, "search: payment option is required")
		}
	}

	if c.RatingMin < 0 || c.RatingMin > 5 {
		return errs.Errorf(errs.EINVALID, "search: ratingMin must be between 0 and 5")
	}

	return nil
}
