package httpserver

import (
	"hexagon/user"
)

type AddUserRequest struct {
	Name     string `json:"name" validate:"required,notblank,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Phone    string `json:"phone" validate:"omitempty,numeric,len=10"`
	Password string `json:"password" validate:"required,notblank,min=9,max=72,password"`
}

func (r AddUserRequest) ToUser() user.User {
	return user.User{
		Name:     r.Name,
		Email:    r.Email,
		Phone:    r.Phone,
		Password: r.Password,
	}
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,notblank,max=72"`
}

type UpdateProfileRequest struct {
	Name  string `json:"name" validate:"required,notblank,min=2,max=100"`
	Phone string `json:"phone" validate:"omitempty,numeric,len=10"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required,notblank,max=72"`
	NewPassword     string `json:"newPassword" validate:"required,notblank,min=9,max=72,password,nefield=CurrentPassword"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required,notblank"`
}
