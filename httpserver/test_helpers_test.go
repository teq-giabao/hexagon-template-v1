package httpserver_test

import (
	"encoding/json"
	"hexagon/httpserver"
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

type apiResponse = httpserver.APIResponse

func decodeAPIResponse(t require.TestingT, rec *httptest.ResponseRecorder) apiResponse {
	var resp apiResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	return resp
}

func decodeAPIResult(t require.TestingT, result interface{}, target interface{}) {
	data, err := json.Marshal(result)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, target))
}
