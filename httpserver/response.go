package httpserver

import (
	"fmt"
	"net/http"
	"strconv"

	"hexagon/errs"

	"github.com/labstack/echo/v4"
)

const (
	successMessage   = "OK"
	defaultErrorCode = "100500"

	ErrCodeInvalid        = "100010"
	ErrCodeNotFound       = "100404"
	ErrCodeUnauthorized   = "100401"
	ErrCodeConflict       = "100409"
	ErrCodeNotImplemented = "100501"
	ErrCodeInternal       = "100500"
)

type APIResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Result  any    `json:"result,omitempty"`
	Info    string `json:"info,omitempty"`
}

func RespondSuccess(c echo.Context, status int, result any) error {
	return c.JSON(status, APIResponse{
		Code:    strconv.Itoa(status),
		Message: successMessage,
		Result:  result,
	})
}

func RespondList(c echo.Context, status int, data any) error {
	return RespondSuccess(c, status, map[string]any{
		"data": data,
	})
}

// nolint: unused
func RespondPagedList(c echo.Context, status int, data any, meta any, page, limit, total int) error {
	result := map[string]any{
		"data":  data,
		"page":  page,
		"limit": limit,
		"total": total,
	}
	if meta != nil {
		result["meta"] = meta
	}
	return RespondSuccess(c, status, result)
}

func RespondError(c echo.Context, status int, message, info string, err error) error {
	if message == "" && status != 0 {
		message = http.StatusText(status)
	}
	if message == "" {
		message = "Error"
	}
	if status >= 500 {
		info = ""
	}
	return c.JSON(status, APIResponse{
		Code:    errorCode(err, status),
		Message: message,
		Info:    info,
	})
}

func errorCode(err error, status int) string {
	if _, ok := err.(*errs.Error); ok {
		switch errs.ErrorCode(err) {
		case errs.EINVALID:
			return ErrCodeInvalid
		case errs.ENOTFOUND:
			return ErrCodeNotFound
		case errs.ECONFLICT:
			return ErrCodeConflict
		case errs.EUNAUTHORIZED:
			return ErrCodeUnauthorized
		case errs.ENOTIMPLEMENTED:
			return ErrCodeNotImplemented
		case errs.EINTERNAL:
			return ErrCodeInternal
		}
	}

	if status != 0 {
		return fmt.Sprintf("100%03d", status)
	}
	return defaultErrorCode
}
