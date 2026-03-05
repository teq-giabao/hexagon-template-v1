package httpserver

import (
	"encoding/json"
	"hexagon/hotel"
	"hexagon/room"
	"hexagon/user"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type APIErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Info    string `json:"info,omitempty"`
}

type APISuccessResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

type APIDataResult struct {
	Data interface{} `json:"data"`
}

type HTTPStatusCodeMapper struct {
	codes map[int]string
}

var defaultHTTPStatusCodeMapper = HTTPStatusCodeMapper{
	codes: map[int]string{
		http.StatusBadRequest:          "100400",
		http.StatusUnauthorized:        "100401",
		http.StatusNotFound:            "100404",
		http.StatusConflict:            "100409",
		http.StatusNotImplemented:      "100501",
		http.StatusInternalServerError: "100500",
	},
}

func respondError(c echo.Context, httpStatus int, message, info string) error {
	return c.JSON(httpStatus, APIErrorResponse{
		Code:    defaultHTTPStatusCodeMapper.Code(httpStatus),
		Message: message,
		Info:    info,
	})
}

func respondOK(c echo.Context, result interface{}) error {
	return c.JSON(http.StatusOK, APISuccessResponse{
		Code:    "200",
		Message: "OK",
		Result:  result,
	})
}

func respondCreated(c echo.Context, result interface{}) error {
	return c.JSON(http.StatusCreated, APISuccessResponse{
		Code:    "201",
		Message: "Created",
		Result:  result,
	})
}

func (m HTTPStatusCodeMapper) Code(httpStatus int) string {
	if code, ok := m.codes[httpStatus]; ok {
		return code
	}
	return strconv.Itoa(httpStatus)
}

func toUserResponse(u user.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Phone:     u.Phone,
		Role:      string(u.Role),
		Status:    string(u.Status),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func toUserResponses(users []user.User) []UserResponse {
	resp := make([]UserResponse, len(users))
	for i, u := range users {
		resp[i] = toUserResponse(u)
	}
	return resp
}

type HotelResponse struct {
	ID                 string                       `json:"id"`
	Name               string                       `json:"name"`
	Description        string                       `json:"description"`
	Address            string                       `json:"address"`
	City               string                       `json:"city"`
	Rating             float64                      `json:"rating"`
	CheckInTime        time.Time                    `json:"checkInTime"`
	CheckOutTime       time.Time                    `json:"checkOutTime"`
	DefaultChildMaxAge int                          `json:"defaultChildMaxAge"`
	Images             []HotelImageResponse         `json:"images"`
	PaymentOptions     []HotelPaymentOptionResponse `json:"paymentOptions"`
	CreatedAt          time.Time                    `json:"createdAt"`
	UpdatedAt          time.Time                    `json:"updatedAt"`
}

type HotelImageResponse struct {
	ID      string `json:"id"`
	HotelID string `json:"hotelId"`
	URL     string `json:"url"`
	IsCover bool   `json:"isCover"`
}

type HotelPaymentOptionResponse struct {
	ID            string `json:"id"`
	HotelID       string `json:"hotelId"`
	PaymentOption string `json:"paymentOption"`
	Enabled       bool   `json:"enabled"`
}

type UploadImagesResponse struct {
	Files []UploadedImageResponse `json:"files"`
}

type UploadedImageResponse struct {
	FileName    string `json:"fileName"`
	URL         string `json:"url"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
}

func toHotelResponse(h hotel.Hotel) HotelResponse {
	images := make([]HotelImageResponse, len(h.Images))
	for i := range h.Images {
		images[i] = HotelImageResponse{
			ID:      h.Images[i].ID,
			HotelID: h.Images[i].HotelID,
			URL:     h.Images[i].URL,
			IsCover: h.Images[i].IsCover,
		}
	}

	paymentOptions := make([]HotelPaymentOptionResponse, len(h.PaymentOptions))
	for i := range h.PaymentOptions {
		paymentOptions[i] = HotelPaymentOptionResponse{
			ID:            h.PaymentOptions[i].ID,
			HotelID:       h.PaymentOptions[i].HotelID,
			PaymentOption: string(h.PaymentOptions[i].PaymentOption),
			Enabled:       h.PaymentOptions[i].Enabled,
		}
	}

	return HotelResponse{
		ID:                 h.ID,
		Name:               h.Name,
		Description:        h.Description,
		Address:            h.Address,
		City:               h.City,
		Rating:             h.Rating,
		CheckInTime:        h.CheckInTime,
		CheckOutTime:       h.CheckOutTime,
		DefaultChildMaxAge: h.DefaultChildMaxAge,
		Images:             images,
		PaymentOptions:     paymentOptions,
		CreatedAt:          h.CreatedAt,
		UpdatedAt:          h.UpdatedAt,
	}
}

func toHotelResponses(hotels []hotel.Hotel) []HotelResponse {
	resp := make([]HotelResponse, len(hotels))
	for i := range hotels {
		resp[i] = toHotelResponse(hotels[i])
	}
	return resp
}

type RoomResponse struct {
	ID           string                `json:"id"`
	HotelID      string                `json:"hotelId"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	BasePrice    float64               `json:"basePrice"`
	MaxAdult     int                   `json:"maxAdult"`
	MaxChild     int                   `json:"maxChild"`
	MaxOccupancy int                   `json:"maxOccupancy"`
	BedOptions   any                   `json:"bedOptions"`
	SizeSqm      int                   `json:"sizeSqm"`
	Status       string                `json:"status"`
	Images       []RoomImageResponse   `json:"images"`
	Amenities    []RoomAmenityResponse `json:"amenities"`
	CreatedAt    time.Time             `json:"createdAt"`
	UpdatedAt    time.Time             `json:"updatedAt"`
}

type RoomImageResponse struct {
	ID      string `json:"id"`
	RoomID  string `json:"roomId"`
	URL     string `json:"url"`
	IsCover bool   `json:"isCover"`
}

type RoomAmenityResponse struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type RoomInventoryResponse struct {
	ID              string    `json:"id"`
	RoomID          string    `json:"roomId"`
	Date            time.Time `json:"date"`
	TotalInventory  int       `json:"totalInventory"`
	HeldInventory   int       `json:"heldInventory"`
	BookedInventory int       `json:"bookedInventory"`
}

func toRoomResponse(r room.Room) RoomResponse {
	images := make([]RoomImageResponse, len(r.Images))
	for i := range r.Images {
		images[i] = RoomImageResponse{
			ID:      r.Images[i].ID,
			RoomID:  r.Images[i].RoomID,
			URL:     r.Images[i].URL,
			IsCover: r.Images[i].IsCover,
		}
	}
	amenities := make([]RoomAmenityResponse, len(r.Amenities))
	for i := range r.Amenities {
		amenities[i] = toRoomAmenityResponse(r.Amenities[i])
	}
	return RoomResponse{
		ID:           r.ID,
		HotelID:      r.HotelID,
		Name:         r.Name,
		Description:  r.Description,
		BasePrice:    r.BasePrice,
		MaxAdult:     r.MaxAdult,
		MaxChild:     r.MaxChild,
		MaxOccupancy: r.MaxOccupancy,
		BedOptions:   jsonValueOrEmptyArray(r.BedOptions),
		SizeSqm:      r.SizeSqm,
		Status:       string(r.Status),
		Images:       images,
		Amenities:    amenities,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

func toRoomAmenityResponse(a room.RoomAmenity) RoomAmenityResponse {
	return RoomAmenityResponse{
		ID:          a.ID,
		Code:        a.Code,
		Name:        a.Name,
		Description: a.Description,
		Icon:        a.Icon,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

func toRoomInventoryResponse(i room.RoomInventory) RoomInventoryResponse {
	return RoomInventoryResponse{
		ID:              i.ID,
		RoomID:          i.RoomID,
		Date:            i.Date,
		TotalInventory:  i.TotalInventory,
		HeldInventory:   i.HeldInventory,
		BookedInventory: i.BookedInventory,
	}
}

func jsonValueOrEmptyArray(raw json.RawMessage) any {
	if len(raw) == 0 {
		return []any{}
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return []any{}
	}
	return decoded
}
