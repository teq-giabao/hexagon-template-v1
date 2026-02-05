package httpserver_test

import (
	"encoding/json"
	"hexagon/pkg/config"
	"net/http/httptest"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/require"
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

type apiResponse struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
	Info    string          `json:"info"`
}

func decodeAPIResponse(t require.TestingT, rec *httptest.ResponseRecorder) apiResponse {
	var resp apiResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	return resp
}
