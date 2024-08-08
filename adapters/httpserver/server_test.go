package httpserver_test

import (
	"hexagon/adapters/httpserver"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	server, err := httpserver.New()
	assert.NoError(t, err)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	server.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), "{\"message\":\"Service is up and running\",\"status\":\"OK\"}")
}
