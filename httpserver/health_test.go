package httpserver_test

import (
	"encoding/json"
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

    var body map[string]interface{}
    err := json.Unmarshal(rec.Body.Bytes(), &body)
    assert.NoError(t, err)

    assert.Equal(t, "200", body["code"])
    assert.Equal(t, "OK", body["message"])

    result := body["result"].(map[string]interface{})
    assert.Equal(t, "OK", result["status"])
}
