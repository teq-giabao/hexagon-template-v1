package httpserver

import (
	"hexagon/errs"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var phonePattern = regexp.MustCompile(`^\+?[0-9]{8,15}$`)

type CustomValidator struct {
	validate *validator.Validate
}

func NewValidator() *CustomValidator {
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
	_ = v.RegisterValidation("phone", validatePhone)
	return &CustomValidator{validate: v}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validate.Struct(i); err != nil {
		return errs.Errorf(errs.EINVALID, formatValidationError(err))
	}
	return nil
}

func validatePhone(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.String {
		return false
	}
	value := strings.TrimSpace(fl.Field().String())
	return phonePattern.MatchString(value)
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
