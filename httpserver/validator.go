package httpserver

import (
	"hexagon/errs"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var phonePattern = regexp.MustCompile(`^[0-9]{10}$`)

type AppValidator struct {
	validate *validator.Validate
}

func NewAppValidator() *AppValidator {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		if name == "" {
			return fld.Name
		}
		return name
	})
	_ = v.RegisterValidation("phone", isPhoneNumberVN10Digits)
	_ = v.RegisterValidation("password", isPasswordMin12CharsWithUpperLowerNumberSpecial)
	return &AppValidator{validate: v}
}

func (cv *AppValidator) Validate(i interface{}) error {
	if err := cv.validate.Struct(i); err != nil {
		return errs.Errorf(errs.EINVALID, "%s", formatValidationError(err))
	}
	return nil
}

// isPhoneNumberVN10Digits validates Vietnamese phone number format.
// Requirements:
// - Must be exactly 10 digits
// - Contains only numeric characters (0-9)
func isPhoneNumberVN10Digits(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.String {
		return false
	}
	value := strings.TrimSpace(fl.Field().String())
	return phonePattern.MatchString(value)
}

// isPasswordMin12CharsWithUpperLowerNumberSpecial validates password strength.
// Requirements:
// - Minimum 12 characters in length
// - Must contain at least one lowercase letter (a-z)
// - Must contain at least one uppercase letter (A-Z)
// - Must contain at least one number (0-9)
// - Must contain at least one special character
func isPasswordMin12CharsWithUpperLowerNumberSpecial(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.String {
		return false
	}
	value := strings.TrimSpace(fl.Field().String())
	if len(value) < 12 {
		return false
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= '0' && r <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	return hasLower && hasUpper && hasDigit && hasSpecial
}

func formatValidationError(err error) string {
	if errs, ok := err.(validator.ValidationErrors); ok {
		parts := make([]string, 0, len(errs))
		for _, fe := range errs {
			field := fe.Field()
			if field == "" {
				field = fe.StructField()
			}
			parts = append(parts, field+" failed on "+fe.Tag())
		}
		return "validation error: " + strings.Join(parts, "; ")
	}
	return "validation error"
}
