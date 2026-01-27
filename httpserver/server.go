package httpserver

import (
	"context"
	"hexagon/contact"
	"hexagon/errs"
	"net/http"

	sentryecho "github.com/getsentry/sentry-go/echo"
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

	// Application services (usecases)
	ContactService contact.Service
}

func Default() *Server {
	s := Server{
		Router:       echo.New(),
		Addr:         ":8080",
		AllowOrigins: []string{"*"},
	}

	s.Router.HTTPErrorHandler = customHTTPErrorHandler
	s.RegisterGlobalMiddlewares()
	s.RegisterContactRoutes()
	s.RegisterHealthRoutes()

	return &s
}

func (s *Server) RegisterGlobalMiddlewares() {
	s.Router.Use(middleware.Recover())
	s.Router.Use(middleware.Secure())
	s.Router.Use(middleware.RequestID())
	s.Router.Use(middleware.Gzip())
	s.Router.Use(sentryecho.New(sentryecho.Options{Repanic: true}))

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
