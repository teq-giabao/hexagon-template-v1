package httpserver

import (
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

type RequestValidator struct {
	validator *validator.Validate
}

func NewRequestValidator() *RequestValidator {
	v := validator.New()
	_ = v.RegisterValidation("notblank", validateNotBlank)
	_ = v.RegisterValidation("password", validatePassword)

	return &RequestValidator{
		validator: v,
	}
}

func (v *RequestValidator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

func validateNotBlank(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	return strings.TrimSpace(value) != ""
}

func validatePassword(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, r := range value {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsNumber(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}
