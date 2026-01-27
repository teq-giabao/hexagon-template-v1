package httpserver

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) RegisterContactRoutes() {
	s.Router.GET("/api/contacts", s.handleListContacts)
	s.Router.POST("/api/contacts", s.handleAddContact)
}

func (s *Server) handleAddContact(c echo.Context) error {
	var req AddContactRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := s.ContactService.AddContact(c.Request().Context(), req.ToContact()); err != nil {
		return err
	}

	return c.NoContent(http.StatusCreated)
}

func (s *Server) handleListContacts(c echo.Context) error {
	contacts, err := s.ContactService.ListContacts(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, contacts)
}
