package user

import "context"

type Service interface {
	AddUser(ctx context.Context, u User) error
	ListUsers(ctx context.Context) ([]User, error)
}

type Repository interface {
	CreateUser(ctx context.Context, u User) error
	AllUsers(ctx context.Context) ([]User, error)
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
	if err := u.Validate(); err != nil {
		return err
	}
	hashed, err := uc.hasher.Hash(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashed
	return uc.r.CreateUser(ctx, u)
}

func (uc *Usecase) ListUsers(ctx context.Context) ([]User, error) {
	return uc.r.AllUsers(ctx)
}
