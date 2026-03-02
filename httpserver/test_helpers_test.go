//nolint:unused
package httpserver_test

import (
	"hexagon/pkg/config"
	"time"

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
		"user_id": 1,
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(testJWTSecret))
}
