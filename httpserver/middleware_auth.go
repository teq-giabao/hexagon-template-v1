package httpserver

import (
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

			verified, ok := claimBool(claims["email_verified"])
			if !ok {
				return s.respondUnauthorized(c, "invalid access token", "missing email_verified claim")
			}

			if !verified {
				return s.respondUnauthorized(c, "email is not verified", auth.ErrEmailNotVerified.Error())
			}

			return next(c)
		}
	}
}

func claimBool(value any) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		if v == "" {
			return false, false
		}

		switch v {
		case "true", "1", "yes":
			return true, true
		case "false", "0", "no":
			return false, true
		default:
			return false, false
		}
	case float64:
		return v != 0, true
	default:
		return false, false
	}
}
