package httpserver

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"hexagon/auth"
	"hexagon/errs"

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
		return writeError(c, http.StatusBadRequest, "invalid request body", err.Error(), err)
	}

	tokens, err := s.AuthService.Login(
		c.Request().Context(),
		req.Email,
		req.Password,
	)

	if err != nil {
		if errors.Is(err, auth.ErrAccountLocked) {
			return writeError(c, http.StatusTooManyRequests, "account temporarily locked", err.Error(), err)
		}
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return writeError(c, http.StatusUnauthorized, "invalid credentials", err.Error(), err)
		}
		return writeError(c, http.StatusInternalServerError, "internal error", err.Error(), err)
	}

	return writeSuccess(c, http.StatusOK, map[string]string{
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
		return writeError(c, http.StatusInternalServerError, "internal error", err.Error(), err)
	}

	authURL, err := s.AuthService.GoogleAuthURL(state)
	if err != nil {
		if errors.Is(err, auth.ErrOAuthNotConfigured) {
			return writeError(c, http.StatusNotImplemented, "oauth not configured", err.Error(), err)
		}
		return writeError(c, http.StatusInternalServerError, "internal error", err.Error(), err)
	}

	c.SetCookie(&http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((5 * time.Minute).Seconds()),
	})

	return writeSuccess(c, http.StatusOK, map[string]string{
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
		return writeError(c, http.StatusBadRequest, "missing code or state", "missing code or state", errs.Errorf(errs.EINVALID, "missing code or state"))
	}

	stateCookie, err := c.Cookie("oauth_state")
	if err != nil || stateCookie == nil || stateCookie.Value != state {
		return writeError(c, http.StatusUnauthorized, "invalid oauth state", "invalid oauth state", errs.Errorf(errs.EUNAUTHORIZED, "invalid oauth state"))
	}

	tokens, err := s.AuthService.LoginWithGoogle(c.Request().Context(), code)
	if err != nil {
		if errors.Is(err, auth.ErrOAuthNotConfigured) {
			return writeError(c, http.StatusNotImplemented, "oauth not configured", err.Error(), err)
		}
		if errors.Is(err, auth.ErrMissingCode) || errors.Is(err, auth.ErrMissingState) {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "missing oauth parameters",
			})
		}
		if errors.Is(err, auth.ErrMissingEmail) || errors.Is(err, auth.ErrUnverifiedEmail) || errors.Is(err, auth.ErrInvalidOAuthUser) {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid oauth user",
			})
		}
		return writeError(c, http.StatusInternalServerError, "internal error", err.Error(), err)
	}

	c.SetCookie(&http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1,
	})

	return writeSuccess(c, http.StatusOK, map[string]string{
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
		return writeError(c, http.StatusBadRequest, "invalid request body", err.Error(), err)
	}

	tokens, err := s.AuthService.Refresh(c.Request().Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidRefreshToken) {
			return writeError(c, http.StatusUnauthorized, "invalid refresh token", err.Error(), err)
		}
		return writeError(c, http.StatusInternalServerError, "internal error", err.Error(), err)
	}

	return writeSuccess(c, http.StatusOK, map[string]string{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}
