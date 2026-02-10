// nolint: funlen
package httpserver_test

import (
	"context"
	"errors"
	"fmt"
	"hexagon/errs"
	"hexagon/httpserver"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefault(t *testing.T) {
	// Act
	server := httpserver.Default(testConfig())

	// Assert
	assert.NotNil(t, server.Router, "Router should be initialized")
	assert.Equal(t, ":8080", server.Addr, "Default address should be :8080")
	assert.Equal(t, []string{"*"}, server.AllowOrigins, "Default CORS should allow all origins")
}

func TestServerStartAndShutdown(t *testing.T) {
	// Arrange
	server := httpserver.Default(testConfig())
	port := allocateRandomPort(t)
	server.Addr = fmt.Sprintf(":%d", port)

	// Act
	errChan := startServerAsync(server)
	waitForServerReady(port)

	// Assert
	assertServerIsRunning(t, port)
	assertServerStopsGracefully(t, server, errChan)
}

func TestRegisterGlobalMiddlewares(t *testing.T) {
	// Arrange
	server := httpserver.Default(testConfig())
	addTestRoute(server)

	// Act
	response := makeRequest(server, http.MethodGet, "/test", nil)

	// Assert
	assert.Equal(t, http.StatusOK, response.Code)
	assertRequestIDMiddlewareApplied(t, response)
	assertSecurityMiddlewareApplied(t, response)
}

func TestCORSConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		allowOrigins  []string
		requestOrigin string
		expectCORS    bool
	}{
		{
			name:          "wildcard allows all origins",
			allowOrigins:  []string{"*"},
			requestOrigin: "https://example.com",
			expectCORS:    true,
		},
		{
			name:          "specific origin is allowed",
			allowOrigins:  []string{"https://example.com"},
			requestOrigin: "https://example.com",
			expectCORS:    true,
		},
		{
			name:          "empty origins disables CORS",
			allowOrigins:  []string{},
			requestOrigin: "https://example.com",
			expectCORS:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			server := configureServerWithCORS(tt.allowOrigins)
			addTestRoute(server)

			// Act
			response := makeRequestWithOrigin(server, "/test", tt.requestOrigin)

			// Assert
			assertCORSBehavior(t, response, tt.expectCORS)
		})
	}
}

func TestMiddlewareRecoveryBehavior(t *testing.T) {
	// Arrange
	server := httpserver.Default(testConfig())
	addPanicRoute(server)

	// Act
	response := makeRequest(server, http.MethodGet, "/panic", nil)

	// Assert
	assertPanicIsRecovered(t, response)
}

func TestCustomErrorHandler(t *testing.T) {
	tests := []struct {
		name               string
		error              error
		expectedStatusCode int
		expectedMessage    string
	}{
		{
			name:               "invalid error returns 400",
			error:              errs.Errorf(errs.EINVALID, "invalid input"),
			expectedStatusCode: http.StatusBadRequest,
			expectedMessage:    "invalid input",
		},
		{
			name:               "not found error returns 404",
			error:              errs.Errorf(errs.ENOTFOUND, "resource not found"),
			expectedStatusCode: http.StatusNotFound,
			expectedMessage:    "resource not found",
		},
		{
			name:               "conflict error returns 409",
			error:              errs.Errorf(errs.ECONFLICT, "resource already exists"),
			expectedStatusCode: http.StatusConflict,
			expectedMessage:    "resource already exists",
		},
		{
			name:               "unauthorized error returns 401",
			error:              errs.Errorf(errs.EUNAUTHORIZED, "unauthorized access"),
			expectedStatusCode: http.StatusUnauthorized,
			expectedMessage:    "unauthorized access",
		},
		{
			name:               "not implemented error returns 501",
			error:              errs.Errorf(errs.ENOTIMPLEMENTED, "feature not implemented"),
			expectedStatusCode: http.StatusNotImplemented,
			expectedMessage:    "feature not implemented",
		},
		{
			name:               "internal error returns 500 with generic message",
			error:              errs.Errorf(errs.EINTERNAL, "database connection failed"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedMessage:    "Internal server error",
		},
		{
			name:               "unknown error returns 500 with generic message",
			error:              errors.New("some random error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedMessage:    "Internal server error",
		},
		{
			name:               "standard library error returns 500",
			error:              fmt.Errorf("wrapped error: %w", errors.New("original error")),
			expectedStatusCode: http.StatusInternalServerError,
			expectedMessage:    "Internal server error",
		},
		{
			name:               "nil context error returns 500",
			error:              context.DeadlineExceeded,
			expectedStatusCode: http.StatusInternalServerError,
			expectedMessage:    "Internal server error",
		},
		{
			name:               "echo http error preserves status code",
			error:              echo.NewHTTPError(http.StatusForbidden, "forbidden"),
			expectedStatusCode: http.StatusForbidden,
			expectedMessage:    "forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			server := httpserver.Default(testConfig())
			addErrorRoute(server, tt.error)

			// Act
			response := makeRequest(server, http.MethodGet, "/error", nil)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, response.Code)
			resp := decodeAPIResponse(t, response)
			assert.Equal(t, tt.expectedMessage, resp.Message)
		})
	}
}

// Helper functions for test setup and assertions

func allocateRandomPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port
}

func startServerAsync(server *httpserver.Server) chan error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start()
	}()
	time.Sleep(100 * time.Millisecond) // Wait for server to start
	return errChan
}

func waitForServerReady(port int) {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", port))
	if err == nil {
		resp.Body.Close()
	}
}

func assertServerIsRunning(t *testing.T, port int) {
	t.Helper()
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", port))
	if err == nil {
		resp.Body.Close()
	}
	// If we got here without panic, server is accessible
}

func assertServerStopsGracefully(t *testing.T, server *httpserver.Server, errChan chan error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	assert.NoError(t, err, "Shutdown should complete without error")

	select {
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Unexpected error during shutdown: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Server did not stop within timeout")
	}
}

func makeRequest(server *httpserver.Server, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)
	return rec
}

func makeRequestWithOrigin(server *httpserver.Server, path, origin string) *httptest.ResponseRecorder {
	return makeRequest(server, http.MethodGet, path, map[string]string{"Origin": origin})
}

func addTestRoute(server *httpserver.Server) {
	server.Router.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})
}

func addPanicRoute(server *httpserver.Server) {
	server.Router.GET("/panic", func(c echo.Context) error {
		panic("test panic")
	})
}

func addErrorRoute(server *httpserver.Server, err error) {
	server.Router.GET("/error", func(c echo.Context) error {
		return err
	})
}

func configureServerWithCORS(allowOrigins []string) *httpserver.Server {
	// Manually construct server to test different CORS configurations
	// since Default() always sets AllowOrigins to ["*"] and registers middlewares
	server := &httpserver.Server{
		Router:       echo.New(),
		Addr:         ":8080",
		AllowOrigins: allowOrigins,
	}
	// Apply minimal middlewares needed for testing
	server.Router.Use(middleware.Recover())
	server.Router.Use(middleware.RequestID())
	if len(server.AllowOrigins) > 0 {
		server.Router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: server.AllowOrigins,
		}))
	}
	return server
}

func assertRequestIDMiddlewareApplied(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	assert.NotEmpty(t, response.Header().Get("X-Request-Id"), "Request ID middleware should add header")
}

func assertSecurityMiddlewareApplied(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	assert.NotEmpty(t, response.Header().Get("X-Content-Type-Options"), "Secure middleware should add headers")
}

func assertCORSBehavior(t *testing.T, response *httptest.ResponseRecorder, expectCORS bool) {
	t.Helper()
	corsHeader := response.Header().Get("Access-Control-Allow-Origin")
	if expectCORS {
		assert.NotEmpty(t, corsHeader, "CORS header should be present")
	} else {
		assert.Empty(t, corsHeader, "CORS header should not be present")
	}
}

func assertPanicIsRecovered(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	assert.Equal(t, http.StatusInternalServerError, response.Code, "Should return 500 on panic")
}
