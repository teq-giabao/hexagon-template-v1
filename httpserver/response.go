package httpserver

import (
	"fmt"
	"strconv"

	"hexagon/errs"

	"github.com/labstack/echo/v4"
)

const (
	successMessage   = "OK"
	defaultErrorCode = "100500"
)

type APIResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result,omitempty"`
	Info    string      `json:"info,omitempty"`
}

func writeSuccess(c echo.Context, status int, result interface{}) error {
	return c.JSON(status, APIResponse{
		Code:    strconv.Itoa(status),
		Message: successMessage,
		Result:  result,
	})
}

func writeList(c echo.Context, status int, data interface{}) error {
	return writeSuccess(c, status, map[string]interface{}{
		"data": data,
	})
}

//nolint: unused
func writePagedList(c echo.Context, status int, data interface{}, meta interface{}, page, limit, total int) error {
	result := map[string]interface{}{
		"data":  data,
		"page":  page,
		"limit": limit,
		"total": total,
	}
	if meta != nil {
		result["meta"] = meta
	}
	return writeSuccess(c, status, result)
}

func writeError(c echo.Context, status int, message, info string, err error) error {
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
			return "100010"
		case errs.ENOTFOUND:
			return "100404"
		case errs.ECONFLICT:
			return "100409"
		case errs.EUNAUTHORIZED:
			return "100401"
		case errs.ENOTIMPLEMENTED:
			return "100501"
		case errs.EINTERNAL:
			return defaultErrorCode
		}
	}

	if status != 0 {
		return fmt.Sprintf("100%03d", status)
	}
	return defaultErrorCode
}
