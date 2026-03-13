package jwt

import (
	"testing"
	"time"

	"hexagon/user"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTProvider_GenerateAndParseAccessToken(t *testing.T) {
	provider := NewJWTProvider("secret", time.Minute, time.Hour)
	u := user.User{ID: "u-1", Email: "u1@example.com", Role: user.UserRoleAdmin}

	token, err := provider.GenerateAccessToken(u)
	require.NoError(t, err)

	parsed, err := provider.ParseAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, u.ID, parsed.ID)
	assert.Equal(t, u.Email, parsed.Email)
	assert.Equal(t, u.Role, parsed.Role)
}

func TestJWTProvider_GenerateAndParseRefreshToken(t *testing.T) {
	provider := NewJWTProvider("secret", time.Minute, time.Hour)
	u := user.User{ID: "u-1", Email: "u1@example.com", Role: user.UserRoleUser}

	token, err := provider.GenerateRefreshToken(u)
	require.NoError(t, err)

	parsed, err := provider.ParseRefreshToken(token)
	require.NoError(t, err)
	assert.Equal(t, u.ID, parsed.ID)
	assert.Equal(t, u.Email, parsed.Email)
	assert.Equal(t, u.Role, parsed.Role)
}

func TestJWTProvider_ParseAccessToken_InvalidType(t *testing.T) {
	provider := NewJWTProvider("secret", time.Minute, time.Hour)

	claims := jwt.MapClaims{
		"iss":     provider.Issuer,
		"aud":     provider.Audience,
		"sub":     "u-1",
		"jti":     "jti",
		"iat":     time.Now().UTC().Unix(),
		"nbf":     time.Now().UTC().Unix(),
		"exp":     time.Now().UTC().Add(time.Minute).Unix(),
		"type":    "refresh",
		"user_id": "u-1",
		"email":   "u1@example.com",
		"role":    string(user.UserRoleUser),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(provider.Secret))
	require.NoError(t, err)

	_, err = provider.ParseAccessToken(signed)
	assert.Error(t, err)
}

func TestJWTProvider_ParseAccessToken_InvalidIssuer(t *testing.T) {
	provider := NewJWTProvider("secret", time.Minute, time.Hour)

	claims := jwt.MapClaims{
		"iss":     "other",
		"aud":     provider.Audience,
		"sub":     "u-1",
		"jti":     "jti",
		"iat":     time.Now().UTC().Unix(),
		"nbf":     time.Now().UTC().Unix(),
		"exp":     time.Now().UTC().Add(time.Minute).Unix(),
		"type":    "access",
		"user_id": "u-1",
		"email":   "u1@example.com",
		"role":    string(user.UserRoleUser),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(provider.Secret))
	require.NoError(t, err)

	_, err = provider.ParseAccessToken(signed)
	assert.Error(t, err)
}

func TestJWTProvider_ParseAccessToken_Expired(t *testing.T) {
	provider := NewJWTProvider("secret", time.Minute, time.Hour)

	claims := jwt.MapClaims{
		"iss":     provider.Issuer,
		"aud":     provider.Audience,
		"sub":     "u-1",
		"jti":     "jti",
		"iat":     time.Now().UTC().Add(-2 * time.Hour).Unix(),
		"nbf":     time.Now().UTC().Add(-2 * time.Hour).Unix(),
		"exp":     time.Now().UTC().Add(-time.Minute).Unix(),
		"type":    "access",
		"user_id": "u-1",
		"email":   "u1@example.com",
		"role":    string(user.UserRoleUser),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(provider.Secret))
	require.NoError(t, err)

	_, err = provider.ParseAccessToken(signed)
	assert.Error(t, err)
}

func TestJWTProvider_UserIDAndEmailValidation(t *testing.T) {
	claims := jwt.MapClaims{}

	_, err := userIDFromClaims(claims)
	assert.Error(t, err)

	_, err = emailFromClaims(claims)
	assert.Error(t, err)
}

func TestGenerateJTI(t *testing.T) {
	_, err := generateJTI(0)
	assert.Error(t, err)

	jti, err := generateJTI(8)
	assert.NoError(t, err)
	assert.NotEmpty(t, jti)
}
