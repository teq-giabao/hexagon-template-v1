package jwt

import (
	"errors"
	"hexagon/user"

	"time"

	"github.com/golang-jwt/jwt"
)

type JWTProvider struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func NewJWTProvider(secret string, accessTTL, refreshTTL time.Duration) *JWTProvider {
	return &JWTProvider{
		Secret:     secret,
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
	}
}

func (p *JWTProvider) GenerateAccessToken(u user.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": u.ID,
		"email":   u.Email,
		"type":    "access",
		"exp":     time.Now().Add(p.AccessTTL).Unix(),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(p.Secret))
}

func (p *JWTProvider) GenerateRefreshToken(u user.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": u.ID,
		"email":   u.Email,
		"type":    "refresh",
		"exp":     time.Now().Add(p.RefreshTTL).Unix(),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(p.Secret))
}

func (p *JWTProvider) ParseRefreshToken(refreshToken string) (user.User, error) {
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(p.Secret), nil
	})
	if err != nil || !token.Valid {
		return user.User{}, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return user.User{}, errors.New("invalid token claims")
	}
	if err := claims.Valid(); err != nil {
		return user.User{}, errors.New("token expired")
	}

	if claimType, ok := claims["type"].(string); !ok || claimType != "refresh" {
		return user.User{}, errors.New("invalid token type")
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return user.User{}, errors.New("invalid user id")
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return user.User{}, errors.New("invalid email")
	}

	return user.User{
		ID:    userID,
		Email: email,
	}, nil
}
