package httpserver

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"hexagon/auth"

	"github.com/labstack/echo/v4"
)

func (s *Server) RegisterAuthRoutes() {
	s.Router.POST("/api/auth/login", s.handleLogin)
	s.Router.POST("/api/auth/refresh", s.handleRefresh)
	s.Router.GET("/api/auth/google/login", s.handleGoogleLogin)
	s.Router.GET("/api/auth/google/callback", s.handleGoogleCallback)
}

// handleLogin godoc
// @Summary User Login
// @Description Authenticate user and return access + refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login Credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/login [post]
func (s *Server) handleLogin(c echo.Context) error {
	var req LoginRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	tokens, err := s.AuthService.Login(
		c.Request().Context(),
		req.Email,
		req.Password,
	)

	if err != nil {
		if errors.Is(err, auth.ErrAccountLocked) {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "account temporarily locked",
			})
		}
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid credentials",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "internal error",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

// handleGoogleLogin godoc
// @Summary Google OAuth Login
// @Description Get Google OAuth2 authorization URL
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/google/login [get]
func (s *Server) handleGoogleLogin(c echo.Context) error {
	state, err := generateOAuthState(32)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "internal error",
		})
	}

	authURL, err := s.AuthService.GoogleAuthURL(state)
	if err != nil {
		if errors.Is(err, auth.ErrOAuthNotConfigured) {
			return c.JSON(http.StatusNotImplemented, map[string]string{
				"error": "oauth not configured",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "internal error",
		})
	}

	c.SetCookie(&http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((5 * time.Minute).Seconds()),
	})

	return c.JSON(http.StatusOK, map[string]string{
		"auth_url": authURL,
	})
}

// handleGoogleCallback godoc
// @Summary Google OAuth Callback
// @Description Exchange Google OAuth2 code for tokens
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/google/callback [get]
func (s *Server) handleGoogleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	if code == "" || state == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "missing code or state",
		})
	}

	stateCookie, err := c.Cookie("oauth_state")
	if err != nil || stateCookie == nil || stateCookie.Value != state {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "invalid oauth state",
		})
	}

	tokens, err := s.AuthService.LoginWithGoogle(c.Request().Context(), code)
	if err != nil {
		if errors.Is(err, auth.ErrOAuthNotConfigured) {
			return c.JSON(http.StatusNotImplemented, map[string]string{
				"error": "oauth not configured",
			})
		}
		if errors.Is(err, auth.ErrInvalidOAuthUser) {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid oauth user",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "internal error",
		})
	}

	c.SetCookie(&http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1,
	})

	return c.JSON(http.StatusOK, map[string]string{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

func generateOAuthState(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("invalid state length")
	}
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// handleRefresh godoc
// @Summary Refresh Access Token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh body RefreshRequest true "Refresh Token"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/refresh [post]
func (s *Server) handleRefresh(c echo.Context) error {
	var req RefreshRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	tokens, err := s.AuthService.Refresh(c.Request().Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidRefreshToken) {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid refresh token",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "internal error",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}
