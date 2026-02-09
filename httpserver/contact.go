package httpserver

import (
	"hexagon/errs"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) RegisterPublicContactRoutes(g *echo.Group) {
	g.GET("/contacts", s.handleListContacts)
}

func (s *Server) RegisterPrivateContactRoutes(g *echo.Group) {
	g.POST("/contacts", s.handleAddContact)
}

// handleAddContact godoc
// @Summary Create Contact
// @Description Add a new contact (requires authentication)
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contact body AddContactRequest true "Contact Data"
// @Success 201 {string} string "Created"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/contacts [post]
func (s *Server) handleAddContact(c echo.Context) error {
	var req AddContactRequest
	if err := c.Bind(&req); err != nil {
		return errs.Errorf(errs.EINVALID, "invalid request body")
	}

	if err := s.ContactService.AddContact(c.Request().Context(), req.ToContact()); err != nil {
		return err
	}

	return RespondSuccess(c, http.StatusCreated, map[string]string{
		"status": "created",
	})
}

// handleListContacts godoc
// @Summary List Contacts
// @Description Get all contacts
// @Tags contacts
// @Produce json
// @Success 200 {array} contact.Contact
// @Failure 500 {object} map[string]string
// @Router /api/contacts [get]
func (s *Server) handleListContacts(c echo.Context) error {
	contacts, err := s.ContactService.ListContacts(c.Request().Context())
	if err != nil {
		return err
	}

	return RespondList(c, http.StatusOK, contacts)
}
