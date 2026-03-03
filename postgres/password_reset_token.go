package postgres

import (
	"context"
	"errors"
	"hexagon/auth"
	"time"

	"gorm.io/gorm"
)

type PasswordResetTokenModel struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string    `gorm:"type:uuid;not null"`
	TokenHash string    `gorm:"not null;unique"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
}

func (PasswordResetTokenModel) TableName() string {
	return "password_reset_tokens"
}

type PasswordResetTokenRepository struct {
	db *gorm.DB
}

func NewPasswordResetTokenRepository(db *gorm.DB) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{db: db}
}

func (r *PasswordResetTokenRepository) Save(ctx context.Context, token auth.PasswordResetToken) error {
	model := PasswordResetTokenModel{
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
		UsedAt:    token.UsedAt,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *PasswordResetTokenRepository) GetActiveByHash(ctx context.Context, tokenHash string) (auth.PasswordResetToken, error) {
	var model PasswordResetTokenModel
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND used_at IS NULL", tokenHash).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return auth.PasswordResetToken{}, errors.New("reset token not found")
		}
		return auth.PasswordResetToken{}, err
	}
	return auth.PasswordResetToken{
		UserID:    model.UserID,
		TokenHash: model.TokenHash,
		ExpiresAt: model.ExpiresAt,
		UsedAt:    model.UsedAt,
	}, nil
}

func (r *PasswordResetTokenRepository) MarkUsedByHash(ctx context.Context, tokenHash string, usedAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&PasswordResetTokenModel{}).
		Where("token_hash = ? AND used_at IS NULL", tokenHash).
		Update("used_at", usedAt)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("reset token not found")
	}
	return nil
}
