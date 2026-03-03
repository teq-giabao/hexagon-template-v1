package postgres

import (
	"context"
	"errors"
	"hexagon/auth"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OAuthProviderAccountModel struct {
	ID             string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID         string `gorm:"type:uuid;not null"`
	Provider       string `gorm:"not null"`
	ProviderUserID string `gorm:"not null"`
	ProviderEmail  string
	CreatedAt      time.Time `gorm:"not null;autoCreateTime"`
}

func (OAuthProviderAccountModel) TableName() string {
	return "oauth_provider_accounts"
}

type OAuthProviderAccountRepository struct {
	db *gorm.DB
}

func NewOAuthProviderAccountRepository(db *gorm.DB) *OAuthProviderAccountRepository {
	return &OAuthProviderAccountRepository{db: db}
}

func (r *OAuthProviderAccountRepository) GetUserIDByProvider(
	ctx context.Context,
	provider auth.OAuthProvider,
	providerUserID string,
) (string, error) {
	var model OAuthProviderAccountModel
	err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_user_id = ?", string(provider), providerUserID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("oauth account not found")
		}
		return "", err
	}
	return model.UserID, nil
}

func (r *OAuthProviderAccountRepository) Upsert(
	ctx context.Context,
	userID string,
	provider auth.OAuthProvider,
	providerUserID, providerEmail string,
) error {
	model := OAuthProviderAccountModel{
		UserID:         userID,
		Provider:       string(provider),
		ProviderUserID: providerUserID,
		ProviderEmail:  providerEmail,
	}

	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "provider"}, {Name: "provider_user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"user_id", "provider_email"}),
	}).Create(&model).Error
	if err != nil {
		if isOAuthProviderConflict(err) {
			return errors.New("oauth provider account conflict")
		}
		return err
	}
	return nil
}

func isOAuthProviderConflict(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" && strings.Contains(strings.ToLower(pqErr.Constraint), "oauth_provider_accounts")
	}
	return false
}
