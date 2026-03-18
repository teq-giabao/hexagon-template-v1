//nolint:unused
package httpserver_test

import (
	"time"

	"hexagon/pkg/config"

	"github.com/golang-jwt/jwt"
)

const testJWTSecret = "test-jwt-secret"

func testConfig() *config.Config {
	cfg := &config.Config{}
	cfg.Auth.JWTSecret = testJWTSecret

	return cfg
}

func signTestToken() (string, error) {
	claims := jwt.MapClaims{
		"sub":            "u-1",
		"user_id":        "u-1",
		"email":          "john@mail.com",
		"email_verified": true,
		"exp":            time.Now().Add(1 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return t.SignedString([]byte(testJWTSecret))
}
