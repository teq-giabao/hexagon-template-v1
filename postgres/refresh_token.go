package postgres

import (
	"context"
	"errors"
	"time"

	"hexagon/auth"

	"gorm.io/gorm"
)

type RefreshTokenModel struct {
	ID        string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string `gorm:"type:uuid;not null"`
	TokenHash string `gorm:"not null;unique"`
	UserAgent string
	IPAddress string
	ExpiresAt time.Time `gorm:"not null"`
	RevokedAt *time.Time
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
}

func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Save(ctx context.Context, token auth.RefreshToken) error {
	model := RefreshTokenModel{
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		UserAgent: token.UserAgent,
		IPAddress: token.IPAddress,
		ExpiresAt: token.ExpiresAt,
		RevokedAt: token.RevokedAt,
	}

	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *RefreshTokenRepository) GetActiveByHash(ctx context.Context, tokenHash string) (auth.RefreshToken, error) {
	var model RefreshTokenModel

	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return auth.RefreshToken{}, errors.New("refresh token not found")
		}

		return auth.RefreshToken{}, err
	}

	return auth.RefreshToken{
		UserID:    model.UserID,
		TokenHash: model.TokenHash,
		UserAgent: model.UserAgent,
		IPAddress: model.IPAddress,
		ExpiresAt: model.ExpiresAt,
		RevokedAt: model.RevokedAt,
	}, nil
}

func (r *RefreshTokenRepository) RevokeByHash(ctx context.Context, tokenHash string, revokedAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&RefreshTokenModel{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", revokedAt)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("refresh token not found")
	}

	return nil
}

func (r *RefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID string, revokedAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&RefreshTokenModel{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", revokedAt)

	return result.Error
}
