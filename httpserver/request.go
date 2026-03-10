package httpserver

import (
	"encoding/json"
	"time"

	"hexagon/hotel"
	"hexagon/room"
	"hexagon/search"
	"hexagon/user"
)

type AddUserRequest struct {
	Name     string `json:"name" validate:"required,notblank,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Phone    string `json:"phone" validate:"omitempty,numeric,len=10"`
	Password string `json:"password" validate:"required,notblank,password"`
}

func (r AddUserRequest) ToUser() user.User {
	return user.User{
		Name:     r.Name,
		Email:    r.Email,
		Phone:    r.Phone,
		Password: r.Password,
	}
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,notblank,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Phone    string `json:"phone" validate:"omitempty,numeric,len=10"`
	Password string `json:"password" validate:"required,notblank,password"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,notblank,max=72"`
}

type UpdateProfileRequest struct {
	Name  string `json:"name" validate:"required,notblank,min=2,max=100"`
	Phone string `json:"phone" validate:"omitempty,numeric,len=10"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required,notblank,max=72"`
	NewPassword     string `json:"newPassword" validate:"required,notblank,password,nefield=CurrentPassword"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required,notblank"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required,notblank"`
	NewPassword string `json:"newPassword" validate:"required,notblank,password"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required,notblank"`
}

type HotelImageRequest struct {
	URL     string `json:"url" validate:"required,notblank,max=1000"`
	IsCover bool   `json:"isCover"`
}

type HotelPaymentOptionRequest struct {
	PaymentOption string `json:"paymentOption" validate:"required,notblank,oneof=immediate pay_at_hotel deferred"`
	Enabled       bool   `json:"enabled"`
}

type AddHotelRequest struct {
	Name               string                      `json:"name" validate:"required,notblank,max=255"`
	Description        string                      `json:"description" validate:"omitempty,max=2000"`
	Address            string                      `json:"address" validate:"required,notblank,max=500"`
	City               string                      `json:"city" validate:"required,notblank,max=255"`
	CheckInTime        string                      `json:"checkInTime" validate:"required,notblank"`
	CheckOutTime       string                      `json:"checkOutTime" validate:"required,notblank"`
	DefaultChildMaxAge int                         `json:"defaultChildMaxAge" validate:"gte=0,lte=17"`
	Images             []HotelImageRequest         `json:"images"`
	PaymentOptions     []HotelPaymentOptionRequest `json:"paymentOptions"`
}

func (r AddHotelRequest) ToHotel(checkIn, checkOut time.Time) hotel.Hotel {
	return hotel.Hotel{
		Name:               r.Name,
		Description:        r.Description,
		Address:            r.Address,
		City:               r.City,
		CheckInTime:        checkIn,
		CheckOutTime:       checkOut,
		DefaultChildMaxAge: r.DefaultChildMaxAge,
		Images:             toHotelImages(r.Images),
		PaymentOptions:     toHotelPaymentOptions(r.PaymentOptions),
	}
}

func toHotelImages(req []HotelImageRequest) []hotel.HotelImage {
	result := make([]hotel.HotelImage, len(req))
	for i := range req {
		result[i] = hotel.HotelImage{
			URL:     req[i].URL,
			IsCover: req[i].IsCover,
		}
	}

	return result
}

func toHotelPaymentOptions(req []HotelPaymentOptionRequest) []hotel.HotelPaymentOption {
	result := make([]hotel.HotelPaymentOption, len(req))
	for i := range req {
		result[i] = hotel.HotelPaymentOption{
			PaymentOption: hotel.PaymentOption(req[i].PaymentOption),
			Enabled:       req[i].Enabled,
		}
	}

	return result
}

type RoomImageRequest struct {
	URL     string `json:"url" validate:"required,notblank,max=1000" example:"https://cdn.example.com/rooms/deluxe-1.jpg"`
	IsCover bool   `json:"isCover" example:"true"`
}

type AddRoomRequest struct {
	HotelID      string             `json:"hotelId" validate:"required,notblank" example:"6d3f8c67-f3f4-4e8f-8c89-c5ff3f2c1244"`
	Name         string             `json:"name" validate:"required,notblank,max=255" example:"Deluxe Twin Room"`
	Description  string             `json:"description" validate:"omitempty,max=2000" example:"Spacious room with city view and balcony"`
	BasePrice    float64            `json:"basePrice" validate:"required,gt=0" example:"1200000"`
	MaxAdult     int                `json:"maxAdult" validate:"required,gt=0" example:"2"`
	MaxChild     int                `json:"maxChild" validate:"gte=0" example:"1"`
	MaxOccupancy int                `json:"maxOccupancy" validate:"required,gt=0" example:"3"`
	BedOptions   json.RawMessage    `json:"bedOptions" swaggertype:"string" example:"{\"beds\":[{\"type\":\"queen\",\"count\":2}]}"`
	SizeSqm      int                `json:"sizeSqm" validate:"gte=0" example:"35"`
	Status       room.RoomStatus    `json:"status" validate:"omitempty,oneof=active inactive" example:"active"`
	Images       []RoomImageRequest `json:"images" validate:"required,min=1,dive"`
	AmenityIDs   []string           `json:"amenityIds" validate:"omitempty,dive,required,notblank" example:"550e8400-e29b-41d4-a716-446655440000,660e8400-e29b-41d4-a716-446655440001"`
}

func (r AddRoomRequest) ToRoom() room.Room {
	return room.Room{
		HotelID:      r.HotelID,
		Name:         r.Name,
		Description:  r.Description,
		BasePrice:    r.BasePrice,
		MaxAdult:     r.MaxAdult,
		MaxChild:     r.MaxChild,
		MaxOccupancy: r.MaxOccupancy,
		BedOptions:   r.BedOptions,
		SizeSqm:      r.SizeSqm,
		Status:       r.Status,
		Images:       toRoomImages(r.Images),
		AmenityIDs:   r.AmenityIDs,
	}
}

func toRoomImages(req []RoomImageRequest) []room.RoomImage {
	result := make([]room.RoomImage, len(req))
	for i := range req {
		result[i] = room.RoomImage{
			URL:     req[i].URL,
			IsCover: req[i].IsCover,
		}
	}

	return result
}

type AddRoomAmenityRequest struct {
	Code        string `json:"code" validate:"required,notblank,max=100"`
	Name        string `json:"name" validate:"required,notblank,max=255"`
	Description string `json:"description" validate:"omitempty,max=2000"`
	Icon        string `json:"icon" validate:"omitempty,max=255"`
}

func (r AddRoomAmenityRequest) ToRoomAmenity() room.RoomAmenity {
	return room.RoomAmenity{
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		Icon:        r.Icon,
	}
}

type AddRoomInventoryRequest struct {
	Date            string `json:"date" validate:"required,notblank"`
	TotalInventory  int    `json:"totalInventory" validate:"gte=0"`
	HeldInventory   int    `json:"heldInventory" validate:"gte=0"`
	BookedInventory int    `json:"bookedInventory" validate:"gte=0"`
}

type SearchHotelsRequest struct {
	Query          string   `json:"query" validate:"required,notblank" example:"ha noi"`
	CheckInAt      string   `json:"checkInAt" validate:"required,notblank,datetime=2006-01-02,date_not_past,date_within_booking_window" example:"2026-04-01"`
	CheckOutAt     string   `json:"checkOutAt" validate:"required,notblank,datetime=2006-01-02,date_within_booking_window,checkout_after_checkin" example:"2026-04-03"`
	RoomCount      int      `json:"roomCount" validate:"required,gt=0" example:"2"`
	AdultCount     int      `json:"adultCount" validate:"required,gt=0" example:"3"`
	ChildrenAges   []int    `json:"childrenAges" validate:"omitempty,dive,gte=0,lte=17" example:"5"`
	RatingMin      float64  `json:"ratingMin" validate:"gte=0,lte=5" example:"4"`
	AmenityIDs     []string `json:"amenityIds" validate:"omitempty,dive,required,notblank"`
	PaymentOptions []string `json:"paymentOptions" validate:"omitempty,dive,oneof=immediate pay_at_hotel deferred" example:"immediate,pay_at_hotel"`
}

func (r SearchHotelsRequest) ToCriteria(checkInAt, checkOutAt time.Time) search.Criteria {
	return search.Criteria{
		Query:          r.Query,
		CheckInDate:    checkInAt,
		CheckOutDate:   checkOutAt,
		Adults:         r.AdultCount,
		ChildrenAges:   r.ChildrenAges,
		RoomCount:      r.RoomCount,
		RatingMin:      r.RatingMin,
		AmenityIDs:     r.AmenityIDs,
		PaymentOptions: r.PaymentOptions,
	}
}
