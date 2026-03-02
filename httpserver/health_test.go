package httpserver_test

import (
	"hexagon/httpserver"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthcheck(t *testing.T) {
	server := httpserver.Default(testConfig())

	req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	rec := httptest.NewRecorder()

	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"200"`)
	assert.Contains(t, rec.Body.String(), `"message":"OK"`)
	assert.Contains(t, rec.Body.String(), `"status":"OK"`)
}
