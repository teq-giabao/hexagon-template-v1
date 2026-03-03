package httpserver

import (
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
