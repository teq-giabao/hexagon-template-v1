package httpserver

import echoSwagger "github.com/swaggo/echo-swagger"

func (s *Server) RegisterSwaggerRoutes() {
	s.Router.GET("/swagger/*", echoSwagger.WrapHandler)
}
