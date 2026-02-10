package httpserver

import (
	"hexagon/contact"
	"hexagon/user"
)

type AddContactRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func (r AddContactRequest) ToContact() contact.Contact {
	return contact.Contact{
		Name:  r.Name,
		Phone: r.Phone,
	}
}

type AddUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r AddUserRequest) ToUser() user.User {
	return user.User{
		Username: r.Username,
		Email:    r.Email,
		Password: r.Password,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"` // nolint: tagliatelle
}
