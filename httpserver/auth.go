package httpserver

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"hexagon/auth"
	"hexagon/user"

	"github.com/labstack/echo/v4"
)

func (s *Server) RegisterAuthRoutes() {
	s.Router.POST("/api/auth/register", s.handleRegister)
	s.Router.POST("/api/auth/login", s.handleLogin)
	s.Router.POST("/api/auth/refresh", s.handleRefresh)
	s.Router.GET("/api/auth/google/login", s.handleGoogleLogin)
	s.Router.GET("/api/auth/google/callback", s.handleGoogleCallback)
}

// handleRegister godoc
// @Summary User Register
// @Description Register a new user and return access + refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body RegisterRequest true "Register payload"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 409 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/auth/register [post]
func (s *Server) handleRegister(c echo.Context) error {
	var req RegisterRequest

	if err := c.Bind(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}

	tokens, err := s.AuthService.Register(
		c.Request().Context(),
		req.Name,
		req.Email,
		req.Phone,
		req.Password,
	)
	if err != nil {
		if errors.Is(err, user.ErrEmailAlreadyExists) {
			return respondError(c, http.StatusConflict, "email already exists", err.Error())
		}
		return respondError(c, http.StatusInternalServerError, "internal error", err.Error())
	}

	return respondOK(c, map[string]string{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
	})
}

// handleLogin godoc
// @Summary User Login
// @Description Authenticate user and return access + refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login Credentials"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 401 {object} APIErrorResponse
// @Failure 429 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/auth/login [post]
func (s *Server) handleLogin(c echo.Context) error {
	var req LoginRequest

	if err := c.Bind(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}

	tokens, err := s.AuthService.Login(
		c.Request().Context(),
		req.Email,
		req.Password,
	)

	if err != nil {
		if errors.Is(err, auth.ErrAccountLocked) {
			return respondError(c, http.StatusTooManyRequests, "account temporarily locked", err.Error())
		}
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return respondError(c, http.StatusUnauthorized, "invalid credentials", err.Error())
		}
		return respondError(c, http.StatusInternalServerError, "internal error", err.Error())
	}

	return respondOK(c, map[string]string{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
	})
}

// handleGoogleLogin godoc
// @Summary Google OAuth Login
// @Description Get Google OAuth2 authorization URL
// @Tags auth
// @Produce json
// @Success 200 {object} APISuccessResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/auth/google/login [get]
func (s *Server) handleGoogleLogin(c echo.Context) error {
	state, err := generateOAuthState(32)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, "internal error", err.Error())
	}

	authURL, err := s.AuthService.GoogleAuthURL(state)
	if err != nil {
		if errors.Is(err, auth.ErrOAuthNotConfigured) {
			return respondError(c, http.StatusNotImplemented, "oauth not configured", err.Error())
		}
		return respondError(c, http.StatusInternalServerError, "internal error", err.Error())
	}

	c.SetCookie(&http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((5 * time.Minute).Seconds()),
	})

	return respondOK(c, map[string]string{
		"authUrl": authURL,
	})
}

// handleGoogleCallback godoc
// @Summary Google OAuth Callback
// @Description Exchange Google OAuth2 code for tokens
// @Tags auth
// @Produce json
// @Param code query string true "OAuth code"
// @Param state query string true "OAuth state"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 401 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/auth/google/callback [get]
func (s *Server) handleGoogleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	if code == "" || state == "" {
		return respondError(c, http.StatusBadRequest, "missing code or state", "missing query parameter code or state")
	}

	stateCookie, err := c.Cookie("oauth_state")
	if err != nil || stateCookie == nil || stateCookie.Value != state {
		return respondError(c, http.StatusUnauthorized, "invalid oauth state", "oauth state mismatch")
	}

	tokens, err := s.AuthService.LoginWithGoogle(c.Request().Context(), code)
	if err != nil {
		if errors.Is(err, auth.ErrOAuthNotConfigured) {
			return respondError(c, http.StatusNotImplemented, "oauth not configured", err.Error())
		}
		if errors.Is(err, auth.ErrMissingCode) || errors.Is(err, auth.ErrMissingState) {
			return respondError(c, http.StatusBadRequest, "missing oauth parameters", err.Error())
		}
		if errors.Is(err, auth.ErrMissingEmail) || errors.Is(err, auth.ErrUnverifiedEmail) || errors.Is(err, auth.ErrInvalidOAuthUser) {
			return respondError(c, http.StatusUnauthorized, "invalid oauth user", err.Error())
		}
		return respondError(c, http.StatusInternalServerError, "internal error", err.Error())
	}

	c.SetCookie(&http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1,
	})

	return respondOK(c, map[string]string{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
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
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 401 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/auth/refresh [post]
func (s *Server) handleRefresh(c echo.Context) error {
	var req RefreshRequest

	if err := c.Bind(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}

	tokens, err := s.AuthService.Refresh(c.Request().Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidRefreshToken) {
			return respondError(c, http.StatusUnauthorized, "invalid refresh token", err.Error())
		}
		return respondError(c, http.StatusInternalServerError, "internal error", err.Error())
	}

	return respondOK(c, map[string]string{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
	})
}
