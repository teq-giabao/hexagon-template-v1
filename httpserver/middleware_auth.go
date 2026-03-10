package httpserver

import (
	"strings"

	"hexagon/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func (s *Server) requireVerifiedEmail() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, ok := c.Get("user").(*jwt.Token)
			if !ok || token == nil {
				return s.respondUnauthorized(c, "invalid access token", "missing jwt context")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return s.respondUnauthorized(c, "invalid access token", "invalid jwt claims")
			}

			email, _ := claims["email"].(string)
			email = strings.TrimSpace(email)

			if email == "" {
				return s.respondUnauthorized(c, "invalid access token", "missing email claim")
			}

			u, err := s.UserService.GetUserByEmail(c.Request().Context(), email)
			if err != nil {
				return s.respondUnauthorized(c, "invalid access token", "user not found")
			}

			if u.EmailVerifiedAt == nil {
				return s.respondUnauthorized(c, "email is not verified", auth.ErrEmailNotVerified.Error())
			}

			return next(c)
		}
	}
}
