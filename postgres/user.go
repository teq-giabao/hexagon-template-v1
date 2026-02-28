package postgres

import (
	"context"
	"errors"
	"hexagon/user"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// UserModel represents the database model for users
type UserModel struct {
	ID                  string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name                string `gorm:"not null"`
	Email               string `gorm:"not null;unique"`
	Phone               string
	PasswordHash        string
	Role                string `gorm:"not null;default:user"`
	Status              string `gorm:"not null;default:active"`
	FailedLoginAttempts int    `gorm:"not null;default:0"`
	LockUntil           *time.Time
	LockEscalationLevel int `gorm:"not null;default:0"`
	LastFailedLoginAt   *time.Time
	CreatedAt           time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt           time.Time `gorm:"not null;autoUpdateTime"`
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
			return user.User{}, user.ErrUserNotFound
		}
		return user.User{}, err
	}

	return toDomainUser(model), nil
}

// GetByID fetches a user by id.
func (r *UserRepository) GetByID(ctx context.Context, id string) (user.User, error) {
	var model UserModel

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user.User{}, user.ErrUserNotFound
		}
		return user.User{}, err
	}

	return toDomainUser(model), nil
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(ctx context.Context, u user.User) error {
	model := toModelUser(u)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		if isDuplicateEmailError(err) {
			return user.ErrEmailAlreadyExists
		}
		return err
	}
	return nil
}

// CreateUserTx creates a new user and runs fn inside the same transaction.
// If fn returns an error, the transaction is rolled back.
func (r *UserRepository) CreateUserTx(ctx context.Context, u user.User, fn func(created user.User) error) (user.User, error) {
	var created user.User
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := toModelUser(u)
		if err := tx.Create(&model).Error; err != nil {
			if isDuplicateEmailError(err) {
				return user.ErrEmailAlreadyExists
			}
			return err
		}
		created = toDomainUser(model)
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
		users[i] = toDomainUser(model)
	}
	return users, nil
}

// UpdateProfile updates mutable profile fields and returns updated user.
func (r *UserRepository) UpdateProfile(ctx context.Context, id, name, phone string) (user.User, error) {
	result := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":       name,
		"phone":      phone,
		"updated_at": time.Now().UTC(),
	})
	if result.Error != nil {
		return user.User{}, result.Error
	}
	if result.RowsAffected == 0 {
		return user.User{}, user.ErrUserNotFound
	}
	return r.GetByID(ctx, id)
}

// UpdatePasswordHash updates user's password hash.
func (r *UserRepository) UpdatePasswordHash(ctx context.Context, id, passwordHash string) error {
	result := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"password_hash": passwordHash,
		"updated_at":    time.Now().UTC(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return user.ErrUserNotFound
	}
	return nil
}

// UpdateStatus updates user's status.
func (r *UserRepository) UpdateStatus(ctx context.Context, id string, status user.UserStatus) error {
	result := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     string(status),
		"updated_at": time.Now().UTC(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return user.ErrUserNotFound
	}
	return nil
}

func toDomainUser(model UserModel) user.User {
	return user.User{
		ID:                  model.ID,
		Name:                model.Name,
		Email:               model.Email,
		Phone:               model.Phone,
		PasswordHash:        model.PasswordHash,
		Role:                user.UserRole(model.Role),
		Status:              user.UserStatus(model.Status),
		FailedLoginAttempts: model.FailedLoginAttempts,
		LockUntil:           model.LockUntil,
		LockEscalationLevel: model.LockEscalationLevel,
		LastFailedLoginAt:   model.LastFailedLoginAt,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}
}

func toModelUser(u user.User) UserModel {
	return UserModel{
		ID:                  u.ID,
		Name:                u.Name,
		Email:               u.Email,
		Phone:               u.Phone,
		PasswordHash:        u.PasswordHash,
		Role:                string(u.Role),
		Status:              string(u.Status),
		FailedLoginAttempts: u.FailedLoginAttempts,
		LockUntil:           u.LockUntil,
		LockEscalationLevel: u.LockEscalationLevel,
		LastFailedLoginAt:   u.LastFailedLoginAt,
	}
}

func isDuplicateEmailError(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" && strings.Contains(strings.ToLower(pqErr.Constraint), "email")
	}
	return false
}
