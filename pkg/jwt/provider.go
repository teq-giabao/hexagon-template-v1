package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"hexagon/user"

	"time"

	"github.com/golang-jwt/jwt"
)

type JWTProvider struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Issuer     string
	Audience   string
}

func NewJWTProvider(secret string, accessTTL, refreshTTL time.Duration) *JWTProvider {
	return &JWTProvider{
		Secret:     secret,
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
		Issuer:     "hexagon-api",
		Audience:   "hexagon-clients",
	}
}

func (p *JWTProvider) GenerateAccessToken(u user.User) (string, error) {
	now := time.Now().UTC()
	jti, err := generateJTI(24)
	if err != nil {
		return "", err
	}
	claims := jwt.MapClaims{
		"iss":     p.Issuer,
		"aud":     p.Audience,
		"sub":     u.ID,
		"jti":     jti,
		"iat":     now.Unix(),
		"nbf":     now.Unix(),
		"exp":     now.Add(p.AccessTTL).Unix(),
		"type":    "access",
		"user_id": u.ID,
		"email":   u.Email,
		"role":    string(u.Role),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(p.Secret))
}

func (p *JWTProvider) GenerateRefreshToken(u user.User) (string, error) {
	now := time.Now().UTC()
	jti, err := generateJTI(24)
	if err != nil {
		return "", err
	}
	claims := jwt.MapClaims{
		"iss":     p.Issuer,
		"aud":     p.Audience,
		"sub":     u.ID,
		"jti":     jti,
		"iat":     now.Unix(),
		"nbf":     now.Unix(),
		"exp":     now.Add(p.RefreshTTL).Unix(),
		"type":    "refresh",
		"user_id": u.ID,
		"email":   u.Email,
		"role":    string(u.Role),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(p.Secret))
}

func (p *JWTProvider) ParseRefreshToken(refreshToken string) (user.User, error) {
	claims, err := p.parseTokenClaims(refreshToken)
	if err != nil {
		return user.User{}, err
	}
	if err := p.validateRefreshClaims(claims); err != nil {
		return user.User{}, err
	}

	userID, err := userIDFromClaims(claims)
	if err != nil {
		return user.User{}, err
	}
	email, err := emailFromClaims(claims)
	if err != nil {
		return user.User{}, err
	}

	return user.User{
		ID:    userID,
		Email: email,
		Role:  user.UserRole(roleFromClaims(claims)),
	}, nil
}

func (p *JWTProvider) parseTokenClaims(refreshToken string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(p.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	if err := claims.Valid(); err != nil {
		return nil, errors.New("token expired")
	}
	return claims, nil
}

func (p *JWTProvider) validateRefreshClaims(claims jwt.MapClaims) error {
	if claimType, ok := claims["type"].(string); !ok || claimType != "refresh" {
		return errors.New("invalid token type")
	}
	if iss, ok := claims["iss"].(string); !ok || iss != p.Issuer {
		return errors.New("invalid token issuer")
	}
	if aud, ok := claims["aud"].(string); !ok || aud != p.Audience {
		return errors.New("invalid token audience")
	}
	if jti, ok := claims["jti"].(string); !ok || jti == "" {
		return errors.New("invalid token id")
	}
	return nil
}

func userIDFromClaims(claims jwt.MapClaims) (string, error) {
	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		if v, fallbackOK := claims["user_id"].(string); fallbackOK && v != "" {
			userID = v
			ok = true
		}
	}
	if !ok || userID == "" {
		return "", errors.New("invalid user id")
	}
	return userID, nil
}

func emailFromClaims(claims jwt.MapClaims) (string, error) {
	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return "", errors.New("invalid email")
	}
	return email, nil
}

func roleFromClaims(claims jwt.MapClaims) string {
	role, ok := claims["role"].(string)
	if !ok || role == "" {
		return string(user.UserRoleUser)
	}
	return role
}

func generateJTI(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("invalid jti length")
	}
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
