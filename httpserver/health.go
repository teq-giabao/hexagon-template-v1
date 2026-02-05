package httpserver

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) RegisterHealthRoutes() {
	s.Router.GET("/healthcheck", s.healthCheck)
}

// healthCheck godoc
// @Summary Health Check
// @Description Check if server is alive
// @Tags health
// @Success 200 {object} map[string]string
// @Router /healthcheck [get]
func (s *Server) healthCheck(c echo.Context) error {
	return writeSuccess(c, http.StatusOK, map[string]string{
		"status": "OK",
	})
}
