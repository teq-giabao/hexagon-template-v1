package user

import (
	"hexagon/errs"
	"strings"
)

var (
	ErrInvalidUsername = errs.Errorf(errs.EINVALID, "user: invalid username")
	ErrInvalidEmail    = errs.Errorf(errs.EINVALID, "user: invalid email")
	ErrInvalidPassword = errs.Errorf(errs.EINVALID, "user: invalid password")
)

type User struct {
	ID       int64
	Username string
	Email    string
	Password string
}

func (u User) Validate() error {
	username := strings.TrimSpace(u.Username)
	email := strings.TrimSpace(u.Email)
	password := strings.TrimSpace(u.Password)

	if username == "" {
		return ErrInvalidUsername
	}

	if email == "" {
		return ErrInvalidEmail
	}

	if password == "" {
		return ErrInvalidPassword
	}

	return nil
}
