package httpserver

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) RegisterUserRoutes() {
	s.Router.GET("/api/users", s.handleListUsers)
	s.Router.POST("/api/users", s.handleAddUser)
}

// handleAddUser godoc
// @Summary Create User
// @Description Add a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body AddUserRequest true "User Data"
// @Success 201
// @Failure 400 {object} map[string]string
// @Router /api/users [post]
func (s *Server) handleAddUser(c echo.Context) error {
	var req AddUserRequest

	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := s.UserService.AddUser(c.Request().Context(), req.ToUser()); err != nil {
		return err
	}

	return RespondSuccess(c, http.StatusCreated, map[string]string{
		"status": "created",
	})
}

// handleListUsers godoc
// @Summary List Users
// @Description Get all users
// @Tags users
// @Produce json
// @Success 200 {array} user.User
// @Router /api/users [get]
func (s *Server) handleListUsers(c echo.Context) error {
	users, err := s.UserService.ListUsers(c.Request().Context())
	if err != nil {
		return err
	}

	return RespondList(c, http.StatusOK, users)
}
