package postgres

import (
	"context"
	"errors"
	"hexagon/user"

	"gorm.io/gorm"
)

// UserModel represents the database model for users
type UserModel struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"not null;unique"`
	Email    string `gorm:"not null;unique"`
	Password string `gorm:"not null"`
}

// TableName specifies the table name for GORM
func (UserModel) TableName() string {
	return "users"
}

// UserRepository implements user.Repository interface
type UserRepository struct {
	db *gorm.DB
}

// GetByEmail implements [auth.UserRepository].
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	var model UserModel

	err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user.User{}, user.ErrInvalidEmail // or return custom ErrNotFound
		}
		return user.User{}, err
	}

	return user.User{
		ID:       int64(model.ID),
		Username: model.Username,
		Email:    model.Email,
		Password: model.Password,
	}, nil
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(ctx context.Context, u user.User) error {
	model := UserModel{
		Username: u.Username,
		Email:    u.Email,
		Password: u.Password,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

// CreateUserTx creates a new user and runs fn inside the same transaction.
// If fn returns an error, the transaction is rolled back.
func (r *UserRepository) CreateUserTx(ctx context.Context, u user.User, fn func(created user.User) error) (user.User, error) {
	var created user.User
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := UserModel{
			Username: u.Username,
			Email:    u.Email,
			Password: u.Password,
		}
		if err := tx.Create(&model).Error; err != nil {
			return err
		}
		created = user.User{
			ID:       int64(model.ID),
			Username: model.Username,
			Email:    model.Email,
			Password: model.Password,
		}
		if fn != nil {
			if err := fn(created); err != nil {
				return err
			}
		}
		return nil
	})
	return created, err
}

// AllUsers fetches all users from the database
func (r *UserRepository) AllUsers(ctx context.Context) ([]user.User, error) {
	var models []UserModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}

	users := make([]user.User, len(models))
	for i, model := range models {
		users[i] = user.User{
			ID:       int64(model.ID),
			Username: model.Username,
			Email:    model.Email,
			Password: model.Password,
		}
	}
	return users, nil
}
