package postgres

import (
	"context"
	"errors"
	"time"

	"hexagon/auth"

	"gorm.io/gorm"
)

type EmailVerificationTokenModel struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string    `gorm:"type:uuid;not null"`
	TokenHash string    `gorm:"not null;unique"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time 
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
}

func (EmailVerificationTokenModel) TableName() string {
	return "email_verification_tokens"
}

type EmailVerificationTokenRepository struct {
	db *gorm.DB
}

func NewEmailVerificationTokenRepository(db *gorm.DB) *EmailVerificationTokenRepository {
	return &EmailVerificationTokenRepository{db: db}
}

func (r *EmailVerificationTokenRepository) Save(ctx context.Context, token auth.EmailVerificationToken) error {
	model := EmailVerificationTokenModel{
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
		UsedAt:    token.UsedAt,
	}

	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *EmailVerificationTokenRepository) GetActiveByHash(ctx context.Context, tokenHash string) (auth.EmailVerificationToken, error) {
	var model EmailVerificationTokenModel

	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND used_at IS NULL", tokenHash).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return auth.EmailVerificationToken{}, errors.New("verify token not found")
		}

		return auth.EmailVerificationToken{}, err
	}

	return auth.EmailVerificationToken{
		UserID:    model.UserID,
		TokenHash: model.TokenHash,
		ExpiresAt: model.ExpiresAt,
		UsedAt:    model.UsedAt,
	}, nil
}

func (r *EmailVerificationTokenRepository) MarkUsedByHash(ctx context.Context, tokenHash string, usedAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&EmailVerificationTokenModel{}).
		Where("token_hash = ? AND used_at IS NULL", tokenHash).
		Update("used_at", usedAt)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("verify token not found")
	}

	return nil
}
