package user

import (
	"context"
	"strings"
)

type Service interface {
	AddUser(ctx context.Context, u User) error
	ListUsers(ctx context.Context) ([]User, error)
	GetUserByID(ctx context.Context, id string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	UpdateProfile(ctx context.Context, id, name, phone string) (User, error)
	ChangePassword(ctx context.Context, id, currentPassword, newPassword string) error
	DeactivateUser(ctx context.Context, id string) error
}

type Repository interface {
	CreateUser(ctx context.Context, u User) error
	AllUsers(ctx context.Context) ([]User, error)
	GetByID(ctx context.Context, id string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	UpdateProfile(ctx context.Context, id, name, phone string) (User, error)
	UpdatePasswordHash(ctx context.Context, id, passwordHash string) error
	UpdateStatus(ctx context.Context, id string, status UserStatus) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashed, plain string) error
}

type Usecase struct {
	r      Repository
	hasher PasswordHasher
}

func NewUsecase(r Repository, h PasswordHasher) *Usecase {
	return &Usecase{
		r:      r,
		hasher: h,
	}
}

func (uc *Usecase) AddUser(ctx context.Context, u User) error {
	if u.Role == "" {
		u.Role = UserRoleUser
	}
	if u.Status == "" {
		u.Status = UserStatusActive
	}
	if err := u.Validate(); err != nil {
		return err
	}
	hashed, err := uc.hasher.Hash(u.Password)
	if err != nil {
		return err
	}
	u.Password = ""
	u.PasswordHash = hashed
	return uc.r.CreateUser(ctx, u)
}

func (uc *Usecase) ListUsers(ctx context.Context) ([]User, error) {
	return uc.r.AllUsers(ctx)
}

func (uc *Usecase) GetUserByID(ctx context.Context, id string) (User, error) {
	if strings.TrimSpace(id) == "" {
		return User{}, ErrUserIDRequired
	}
	return uc.r.GetByID(ctx, id)
}

func (uc *Usecase) GetUserByEmail(ctx context.Context, email string) (User, error) {
	email = strings.TrimSpace(email)
	if err := validateEmail(email); err != nil {
		return User{}, err
	}
	return uc.r.GetByEmail(ctx, email)
}

func (uc *Usecase) UpdateProfile(ctx context.Context, id, name, phone string) (User, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	if id == "" {
		return User{}, ErrUserIDRequired
	}
	if err := validateName(name); err != nil {
		return User{}, err
	}
	return uc.r.UpdateProfile(ctx, id, name, strings.TrimSpace(phone))
}

func (uc *Usecase) ChangePassword(ctx context.Context, id, currentPassword, newPassword string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrUserIDRequired
	}
	currentPassword = strings.TrimSpace(currentPassword)
	if currentPassword == "" {
		return ErrCurrentPasswordInvalid
	}
	newPassword = strings.TrimSpace(newPassword)
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	existing, err := uc.r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := uc.hasher.Compare(existing.PasswordHash, currentPassword); err != nil {
		return ErrCurrentPasswordInvalid
	}

	hashed, err := uc.hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	return uc.r.UpdatePasswordHash(ctx, id, hashed)
}

func (uc *Usecase) DeactivateUser(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrUserIDRequired
	}
	return uc.r.UpdateStatus(ctx, id, UserStatusInactive)
}
