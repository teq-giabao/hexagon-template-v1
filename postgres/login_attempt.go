package postgres

import (
	"context"
	"errors"
	"hexagon/auth"
	"time"

	"gorm.io/gorm"
)

// LoginAttemptModel represents the database model for login attempts.
type LoginAttemptModel struct {
	Email       string     `gorm:"primaryKey"`
	FailedCount int        `gorm:"not null"`
	JailedUntil *time.Time `gorm:""`
}

// TableName specifies the table name for GORM.
func (LoginAttemptModel) TableName() string {
	return "login_attempts"
}

// LoginAttemptRepository implements [auth.LoginAttemptRepository].
type LoginAttemptRepository struct {
	db *gorm.DB
}

// NewLoginAttemptRepository creates a new login attempt repository.
func NewLoginAttemptRepository(db *gorm.DB) *LoginAttemptRepository {
	return &LoginAttemptRepository{db: db}
}

// Get implements [auth.LoginAttemptRepository].
func (r *LoginAttemptRepository) Get(ctx context.Context, email string) (auth.LoginAttempt, error) {
	var model LoginAttemptModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return auth.LoginAttempt{}, nil
		}
		return auth.LoginAttempt{}, err
	}

	var jailedUntil time.Time
	if model.JailedUntil != nil {
		jailedUntil = model.JailedUntil.UTC()
	}

	return auth.LoginAttempt{
		FailedCount: model.FailedCount,
		JailedUntil: jailedUntil,
	}, nil
}

// Save implements [auth.LoginAttemptRepository].
func (r *LoginAttemptRepository) Save(ctx context.Context, email string, attempt auth.LoginAttempt) error {
	var jailedUntil *time.Time
	if !attempt.JailedUntil.IsZero() {
		t := attempt.JailedUntil.UTC()
		jailedUntil = &t
	}

	model := LoginAttemptModel{
		Email:       email,
		FailedCount: attempt.FailedCount,
		JailedUntil: jailedUntil,
	}

	return r.db.WithContext(ctx).Save(&model).Error
}

// Reset implements [auth.LoginAttemptRepository].
func (r *LoginAttemptRepository) Reset(ctx context.Context, email string) error {
	return r.db.WithContext(ctx).Where("email = ?", email).Delete(&LoginAttemptModel{}).Error
}
