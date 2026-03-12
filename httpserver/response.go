// nolint: unused
package httpserver

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"hexagon/hotel"
	"hexagon/room"
	"hexagon/search"
	"hexagon/user"

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

func (s *Server) respondError(c echo.Context, httpStatus int, message, info string) error {
	if s.Config != nil && s.Config.IsProduction() {
		info = ""
	}

	return c.JSON(httpStatus, APIErrorResponse{
		Code:    defaultHTTPStatusCodeMapper.Code(httpStatus),
		Message: message,
		Info:    info,
	})
}

func (s *Server) respondBadRequest(c echo.Context, message, info string) error {
	return s.respondError(c, http.StatusBadRequest, message, info)
}

func (s *Server) respondUnauthorized(c echo.Context, message, info string) error {
	return s.respondError(c, http.StatusUnauthorized, message, info)
}

func (s *Server) respondNotFound(c echo.Context, message, info string) error {
	return s.respondError(c, http.StatusNotFound, message, info)
}

func (s *Server) respondConflict(c echo.Context, message, info string) error {
	return s.respondError(c, http.StatusConflict, message, info)
}

func (s *Server) respondInternalServerError(c echo.Context, message, info string) error {
	return s.respondError(c, http.StatusInternalServerError, message, info)
}

func (s *Server) respondTooManyRequests(c echo.Context, message, info string) error {
	return s.respondError(c, http.StatusTooManyRequests, message, info)
}

func (s *Server) respondNotImplemented(c echo.Context, message, info string) error {
	return s.respondError(c, http.StatusNotImplemented, message, info)
}

func (s *Server) respondOK(c echo.Context, result interface{}) error {
	return c.JSON(http.StatusOK, APISuccessResponse{
		Code:    "200",
		Message: "OK",
		Result:  result,
	})
}

func (s *Server) respondCreated(c echo.Context, result interface{}) error {
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

type SearchHotelsResponse struct {
	Hotels     []SearchHotelItemResponse `json:"hotels"`
	Pagination SearchPaginationResponse  `json:"pagination"`
}

type SearchPaginationResponse struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Offset     int `json:"offset"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

type SearchHotelItemResponse struct {
	HotelID            string   `json:"hotelId"`
	Name               string   `json:"name"`
	City               string   `json:"city"`
	Address            string   `json:"address"`
	Rating             float64  `json:"rating"`
	PaymentOptions     []string `json:"paymentOptions"`
	MinPrice           float64  `json:"minPrice"`
	AvailableRoomCount int      `json:"availableRoomCount"`
	MatchesRequested   bool     `json:"matchesRequested"`
	FlexibleMatch      bool     `json:"flexibleMatch"`
}

func toSearchHotelsResponse(in []search.HotelSearchResult, page, pageSize, offset, total int) SearchHotelsResponse {
	hotels := make([]SearchHotelItemResponse, len(in))
	for i := range in {
		hotels[i] = SearchHotelItemResponse{
			HotelID:            in[i].HotelID,
			Name:               in[i].Name,
			City:               in[i].City,
			Address:            in[i].Address,
			Rating:             in[i].Rating,
			PaymentOptions:     in[i].PaymentOptions,
			MinPrice:           in[i].MinPrice,
			AvailableRoomCount: in[i].AvailableRoomCount,
			MatchesRequested:   in[i].MatchesRequested,
			FlexibleMatch:      in[i].FlexibleMatch,
		}
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	return SearchHotelsResponse{
		Hotels: hotels,
		Pagination: SearchPaginationResponse{
			Page:       page,
			PageSize:   pageSize,
			Offset:     offset,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

type SearchHotelRoomsResponse struct {
	HotelID            string                        `json:"hotelId"`
	RequestedRoomCount int                           `json:"requestedRoomCount"`
	StrictMatch        bool                          `json:"strictMatch"`
	Rooms              []SearchHotelRoomItemResponse `json:"rooms"`
}

type SearchHotelRoomItemResponse struct {
	RoomID         string   `json:"roomId"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	BasePrice      float64  `json:"basePrice"`
	MaxAdult       int      `json:"maxAdult"`
	MaxChild       int      `json:"maxChild"`
	MaxOccupancy   int      `json:"maxOccupancy"`
	AvailableCount int      `json:"availableCount"`
	AmenityIDs     []string `json:"amenityIds"`
	AmenityCodes   []string `json:"amenityCodes"`
	AmenityNames   []string `json:"amenityNames"`
}

func toSearchHotelRoomsResponse(in search.HotelRoomSearchResult) SearchHotelRoomsResponse {
	rooms := make([]SearchHotelRoomItemResponse, len(in.Rooms))

	for i := range in.Rooms {
		rooms[i] = SearchHotelRoomItemResponse{
			RoomID:         in.Rooms[i].RoomID,
			Name:           in.Rooms[i].Name,
			Description:    in.Rooms[i].Description,
			BasePrice:      in.Rooms[i].BasePrice,
			MaxAdult:       in.Rooms[i].MaxAdult,
			MaxChild:       in.Rooms[i].MaxChild,
			MaxOccupancy:   in.Rooms[i].MaxOccupancy,
			AvailableCount: in.Rooms[i].AvailableCount,
			AmenityIDs:     in.Rooms[i].AmenityIDs,
			AmenityCodes:   in.Rooms[i].AmenityCodes,
			AmenityNames:   in.Rooms[i].AmenityNames,
		}
	}

	return SearchHotelRoomsResponse{
		HotelID:            in.HotelID,
		RequestedRoomCount: in.RequestedRoomCount,
		StrictMatch:        in.StrictMatch,
		Rooms:              rooms,
	}
}

type SearchHotelRoomCombinationsResponse struct {
	HotelID            string                               `json:"hotelId"`
	RequestedRoomCount int                                  `json:"requestedRoomCount"`
	Combinations       []SearchHotelRoomCombinationResponse `json:"combinations"`
}

type SearchHotelRoomCombinationResponse struct {
	Items          []SearchHotelRoomCombinationItemResponse `json:"items"`
	TotalPrice     float64                                  `json:"totalPrice"`
	TotalRooms     int                                      `json:"totalRooms"`
	TotalMaxAdult  int                                      `json:"totalMaxAdult"`
	TotalMaxChild  int                                      `json:"totalMaxChild"`
	TotalOccupancy int                                      `json:"totalOccupancy"`
}

type SearchHotelRoomCombinationItemResponse struct {
	RoomID    string  `json:"roomId"`
	RoomName  string  `json:"roomName"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
	Subtotal  float64 `json:"subtotal"`
}

func toSearchHotelRoomCombinationsResponse(in search.HotelRoomCombinationsResult) SearchHotelRoomCombinationsResponse {
	combinations := make([]SearchHotelRoomCombinationResponse, len(in.Combinations))

	for i := range in.Combinations {
		items := make([]SearchHotelRoomCombinationItemResponse, len(in.Combinations[i].Items))

		for j := range in.Combinations[i].Items {
			items[j] = SearchHotelRoomCombinationItemResponse{
				RoomID:    in.Combinations[i].Items[j].RoomID,
				RoomName:  in.Combinations[i].Items[j].RoomName,
				Quantity:  in.Combinations[i].Items[j].Quantity,
				UnitPrice: in.Combinations[i].Items[j].UnitPrice,
				Subtotal:  in.Combinations[i].Items[j].Subtotal,
			}
		}

		combinations[i] = SearchHotelRoomCombinationResponse{
			Items:          items,
			TotalPrice:     in.Combinations[i].TotalPrice,
			TotalRooms:     in.Combinations[i].TotalRooms,
			TotalMaxAdult:  in.Combinations[i].TotalMaxAdult,
			TotalMaxChild:  in.Combinations[i].TotalMaxChild,
			TotalOccupancy: in.Combinations[i].TotalOccupancy,
		}
	}

	return SearchHotelRoomCombinationsResponse{
		HotelID:            in.HotelID,
		RequestedRoomCount: in.RequestedRoomCount,
		Combinations:       combinations,
	}
}
