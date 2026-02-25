package httpserver

import (
	"context"
	"hexagon/auth"
	"hexagon/errs"
	"hexagon/pkg/config"
	"hexagon/user"
	"net/http"

	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	// Router is the Echo router instance
	Router *echo.Echo

	// Addr represents the address the server will listen on
	Addr string

	// Allowed origins for CORS
	AllowOrigins []string

	UserService user.Service

	AuthService auth.Service

	JWTSecret string 
}

func Default(cfg *config.Config) *Server {
	s := Server{
		Router:       echo.New(),
		Addr:         ":8080",
		AllowOrigins: []string{"*"},
	}

	s.Router.HTTPErrorHandler = customHTTPErrorHandler
	s.RegisterGlobalMiddlewares()
	api := s.Router.Group("/api")

	// PUBLIC
	public := api.Group("")
	s.RegisterPublicRoutes(public)

	// PRIVATE
	private := api.Group("")
	private.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(cfg.Auth.JWTSecret),
		SigningMethod: "HS256",
	}))
	s.RegisterPrivateRoutes(private)
	s.RegisterHealthRoutes()
	s.RegisterSwaggerRoutes()
	s.RegisterUserRoutes()
	s.RegisterAuthRoutes()
	return &s
}

func (s *Server) RegisterGlobalMiddlewares() {
	s.Router.Use(middleware.Recover())
	s.Router.Use(middleware.Secure())
	s.Router.Use(middleware.RequestID())
	s.Router.Use(middleware.Gzip())
	s.Router.Use(sentryecho.New(sentryecho.Options{Repanic: true}))
	s.Router.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	// CORS
	if len(s.AllowOrigins) > 0 {
		s.Router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: s.AllowOrigins,
		}))
	}
}

func (s *Server) Start() error {
	return s.Router.Start(s.Addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.Router.Shutdown(ctx)
}

// customHTTPErrorHandler maps application errors to appropriate HTTP status codes
func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal server error"

	// Check if it's an Echo HTTPError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message.(string)
	} else {
		// Map application error codes to HTTP status codes
		switch errs.ErrorCode(err) {
		case errs.EINVALID:
			code = http.StatusBadRequest
			message = errs.ErrorMessage(err)
		case errs.ENOTFOUND:
			code = http.StatusNotFound
			message = errs.ErrorMessage(err)
		case errs.ECONFLICT:
			code = http.StatusConflict
			message = errs.ErrorMessage(err)
		case errs.EUNAUTHORIZED:
			code = http.StatusUnauthorized
			message = errs.ErrorMessage(err)
		case errs.ENOTIMPLEMENTED:
			code = http.StatusNotImplemented
			message = errs.ErrorMessage(err)
		case errs.EINTERNAL:
			code = http.StatusInternalServerError
			message = "Internal server error"
		}
	}

	// Don't write response if already committed
	if !c.Response().Committed {
		err = c.JSON(code, map[string]string{"error": message})
		if err != nil {
			c.Logger().Error(err)
		}
	}
}

func (s *Server) RegisterPublicRoutes(g *echo.Group) {
	// public part of contacts

	// other public modules
	// s.RegisterHealthRoutes()
	// s.RegisterSwaggerRoutes()
	// s.RegisterAuthRoutes()
	// s.RegisterUserRoutes()
}

func (s *Server) RegisterPrivateRoutes(g *echo.Group) {
	// s.RegisterPrivateContactRoutes(g)
}

