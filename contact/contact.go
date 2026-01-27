package contact

import (
	"hexagon/errs"
)

var (
	ErrInvalidName  = errs.Errorf(errs.EINVALID, "invalid name")
	ErrInvalidPhone = errs.Errorf(errs.EINVALID, "invalid phone")
)

type Contact struct {
	Name  string
	Phone string
}

func (c Contact) Validate() error {
	if c.Name == "" {
		return ErrInvalidName
	}

	if c.Phone == "" {
		return ErrInvalidPhone
	}

	return nil
}
