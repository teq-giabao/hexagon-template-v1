package httpserver

import (
	"hexagon/hotel"
	"hexagon/user"
	"time"
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
