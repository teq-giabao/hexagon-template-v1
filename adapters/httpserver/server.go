package httpserver

import (
	"hexagon/domain/book"
	"hexagon/pkg/config"
	"hexagon/pkg/logger"
	"hexagon/pkg/sentry"
	"net/http"
	"strings"

	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type Server struct {
	Router *echo.Echo
	Config *config.Config
	Logger *zap.SugaredLogger

	// storage adapters
	BookStore book.Storage
}

func New(options ...Options) (*Server, error) {
	s := Server{
		Router: echo.New(),
		Config: config.Empty,
		Logger: logger.NOOPLogger,
	}

	for _, fn := range options {
		if err := fn(&s); err != nil {
			return nil, err
		}
	}

	s.RegisterGlobalMiddlewares()

	s.RegisterHealthCheck(s.Router.Group(""))
	s.RegisterBookRoutes(s.Router.Group("/api/books"))

	return &s, nil
}

func (s *Server) RegisterGlobalMiddlewares() {
	s.Router.Use(middleware.Recover())
	s.Router.Use(middleware.Secure())
	s.Router.Use(middleware.RequestID())
	s.Router.Use(middleware.Gzip())
	s.Router.Use(sentryecho.New(sentryecho.Options{Repanic: true}))

	// CORS
	if s.Config.AllowOrigins != "" {
		aos := strings.Split(s.Config.AllowOrigins, ",")
		s.Router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: aos,
		}))
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

func (s *Server) RegisterHealthCheck(router *echo.Group) {
	router.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  http.StatusText(http.StatusOK),
			"message": "Service is up and running",
		})
	})
}

func (s *Server) handleError(c echo.Context, err error, status int) error {
	s.Logger.Errorw(
		err.Error(),
		zap.String("request_id", s.requestID(c)),
	)

	if status >= http.StatusInternalServerError {
		sentry.WithContext(c).Error(err)
	}

	return c.JSON(status, map[string]string{
		"message": http.StatusText(status),
	})
}

func (s *Server) requestID(c echo.Context) string {
	return c.Response().Header().Get(echo.HeaderXRequestID)
}
