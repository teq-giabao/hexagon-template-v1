package httpserver

import (
	"hexagon/contact"
	"hexagon/user"
)

type AddContactRequest struct {
	Name  string `json:"name" validate:"required"`
	Phone string `json:"phone" validate:"required,phone"`
}

func (r AddContactRequest) ToContact() contact.Contact {
	return contact.Contact{
		Name:  r.Name,
		Phone: r.Phone,
	}
}

type AddUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (r AddUserRequest) ToUser() user.User {
	return user.User{
		Username: r.Username,
		Email:    r.Email,
		Password: r.Password,
	}
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"` // nolint: tagliatelle
}
