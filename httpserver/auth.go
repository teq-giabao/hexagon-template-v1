package httpserver

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"hexagon/auth"
	"hexagon/user"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

func (s *Server) RegisterAuthRoutes() {
	authGroup := s.Router.Group("/api/auth")
	authGroup.POST("/register", s.handleRegister)
	authGroup.GET("/google/login", s.handleGoogleLogin)
	authGroup.GET("/google/callback", s.handleGoogleCallback)

	protectedAuth := s.Router.Group("/api/auth")
	protectedAuth.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey:    []byte(s.JWTSecret),
		SigningMethod: "HS256",
	}))
	protectedAuth.POST("/logout", s.handleLogout)
	protectedAuth.GET("/me", s.handleMe)

	sensitiveAuth := s.Router.Group("/api/auth")
	sensitiveAuth.Use(authSensitiveRateLimiter())
	sensitiveAuth.POST("/login", s.handleLogin)
	sensitiveAuth.POST("/forgot-password", s.handleForgotPassword)
	sensitiveAuth.POST("/reset-password", s.handleResetPassword)
	sensitiveAuth.POST("/refresh", s.handleRefresh)
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
		auth.WithClientInfo(c.Request().Context(), auth.ClientInfo{
			UserAgent: c.Request().UserAgent(),
			IPAddress: c.RealIP(),
		}),
		req.Name,
		req.Email,
		req.Phone,
		req.Password,
	)
	if err != nil {
		if errors.Is(err, auth.ErrEmailRegisteredWithOAuth) {
			return respondError(c, http.StatusConflict, "email already registered with oauth", err.Error())
		}
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
		auth.WithClientInfo(c.Request().Context(), auth.ClientInfo{
			UserAgent: c.Request().UserAgent(),
			IPAddress: c.RealIP(),
		}),
		req.Email,
		req.Password,
	)

	if err != nil {
		if errors.Is(err, auth.ErrPasswordAuthNotAvailable) {
			return respondError(c, http.StatusUnauthorized, "password login is not available for this account", err.Error())
		}
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

// handleLogout godoc
// @Summary User Logout
// @Description Revoke current refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body LogoutRequest true "Logout payload"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 401 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/auth/logout [post]
func (s *Server) handleLogout(c echo.Context) error {
	var req LogoutRequest
	if err := c.Bind(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := s.AuthService.Logout(auth.WithClientInfo(c.Request().Context(), auth.ClientInfo{
		UserAgent: c.Request().UserAgent(),
		IPAddress: c.RealIP(),
	}), req.RefreshToken); err != nil {
		if errors.Is(err, auth.ErrInvalidRefreshToken) {
			return respondError(c, http.StatusUnauthorized, "invalid refresh token", err.Error())
		}
		return respondError(c, http.StatusInternalServerError, "internal error", err.Error())
	}
	return respondOK(c, map[string]any{})
}

// handleForgotPassword godoc
// @Summary Forgot Password
// @Description Send reset password email
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body ForgotPasswordRequest true "Forgot password payload"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/auth/forgot-password [post]
func (s *Server) handleForgotPassword(c echo.Context) error {
	var req ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := s.AuthService.ForgotPassword(auth.WithClientInfo(c.Request().Context(), auth.ClientInfo{
		UserAgent: c.Request().UserAgent(),
		IPAddress: c.RealIP(),
	}), req.Email); err != nil {
		if errors.Is(err, auth.ErrPasswordAuthNotAvailable) {
			return respondError(c, http.StatusBadRequest, "password reset is not available for this account", err.Error())
		}
		if errors.Is(err, auth.ErrMailerNotConfigured) {
			return respondError(c, http.StatusNotImplemented, "password reset mailer not configured", err.Error())
		}
		return respondError(c, http.StatusInternalServerError, "internal error", err.Error())
	}
	return respondOK(c, map[string]any{})
}

// handleResetPassword godoc
// @Summary Reset Password
// @Description Reset password using reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body ResetPasswordRequest true "Reset password payload"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 401 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/auth/reset-password [post]
func (s *Server) handleResetPassword(c echo.Context) error {
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := s.AuthService.ResetPassword(auth.WithClientInfo(c.Request().Context(), auth.ClientInfo{
		UserAgent: c.Request().UserAgent(),
		IPAddress: c.RealIP(),
	}), req.Token, req.NewPassword); err != nil {
		if errors.Is(err, auth.ErrPasswordAuthNotAvailable) {
			return respondError(c, http.StatusBadRequest, "password reset is not available for this account", err.Error())
		}
		if errors.Is(err, auth.ErrInvalidResetToken) {
			return respondError(c, http.StatusUnauthorized, "invalid reset token", err.Error())
		}
		return respondError(c, http.StatusInternalServerError, "internal error", err.Error())
	}
	return respondOK(c, map[string]any{})
}

// handleMe godoc
// @Summary Current User
// @Description Get current authenticated user information
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} APISuccessResponse
// @Failure 401 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/auth/me [get]
func (s *Server) handleMe(c echo.Context) error {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok || token == nil {
		return respondError(c, http.StatusUnauthorized, "invalid access token", "missing jwt context")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return respondError(c, http.StatusUnauthorized, "invalid access token", "invalid jwt claims")
	}
	email, _ := claims["email"].(string)
	email = strings.TrimSpace(email)
	if email == "" {
		return respondError(c, http.StatusUnauthorized, "invalid access token", "missing email claim")
	}

	u, err := s.UserService.GetUserByEmail(c.Request().Context(), email)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, "internal error", err.Error())
	}
	return respondOK(c, toUserResponse(u))
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
// @Description Exchange Google OAuth2 code for tokens. Requires oauth_state cookie set by /api/auth/google/login.
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

	tokens, err := s.AuthService.LoginWithGoogle(auth.WithClientInfo(c.Request().Context(), auth.ClientInfo{
		UserAgent: c.Request().UserAgent(),
		IPAddress: c.RealIP(),
	}), code)
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

func extractBearerToken(authorization string) string {
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authorization, bearerPrefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(authorization, bearerPrefix))
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

	tokens, err := s.AuthService.Refresh(auth.WithClientInfo(c.Request().Context(), auth.ClientInfo{
		UserAgent: c.Request().UserAgent(),
		IPAddress: c.RealIP(),
	}), req.RefreshToken)
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

func authSensitiveRateLimiter() echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(5.0 / 60.0),
			Burst:     5,
			ExpiresIn: 3 * time.Minute,
		}),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.Path() + "|" + c.RealIP(), nil
		},
		DenyHandler: func(c echo.Context, _ string, _ error) error {
			return respondError(c, http.StatusTooManyRequests, "too many requests", "rate limit exceeded")
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return respondError(c, http.StatusInternalServerError, "internal error", err.Error())
		},
	})
}
