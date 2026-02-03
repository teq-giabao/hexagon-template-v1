package httpserver

import (
	"errors"
	"net/http"

	"hexagon/auth"

	"github.com/labstack/echo/v4"
)

func (s *Server) RegisterAuthRoutes() {
	s.Router.POST("/api/auth/login", s.handleLogin)
	s.Router.POST("/api/auth/refresh", s.handleRefresh)
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
