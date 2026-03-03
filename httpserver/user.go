package httpserver

import (
	"github.com/labstack/echo/v4"
)

func (s *Server) RegisterUserRoutes() {
	s.Router.GET("/api/users", s.handleListUsers)
	s.Router.GET("/api/users/:id", s.handleGetUserByID)
	s.Router.GET("/api/users/by-email", s.handleGetUserByEmail)
	s.Router.POST("/api/users", s.handleAddUser)
	s.Router.PATCH("/api/users/:id/profile", s.handleUpdateProfile)
	s.Router.PATCH("/api/users/:id/password", s.handleChangePassword)
	s.Router.PATCH("/api/users/:id/deactivate", s.handleDeactivateUser)
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
		return respondError(c, 400, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return respondError(c, 400, "invalid request body", err.Error())
	}

	if err := s.UserService.AddUser(c.Request().Context(), req.ToUser()); err != nil {
		return err
	}

	return respondCreated(c, map[string]any{})
}

// handleListUsers godoc
// @Summary List Users
// @Description Get all users
// @Tags users
// @Produce json
// @Success 200 {array} UserResponse
// @Router /api/users [get]
func (s *Server) handleListUsers(c echo.Context) error {
	users, err := s.UserService.ListUsers(c.Request().Context())
	if err != nil {
		return err
	}

	return respondOK(c, APIDataResult{Data: toUserResponses(users)})
}

// handleGetUserByID godoc
// @Summary Get User By ID
// @Description Get a user by id
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/users/{id} [get]
func (s *Server) handleGetUserByID(c echo.Context) error {
	id := c.Param("id")
	u, err := s.UserService.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return respondOK(c, toUserResponse(u))
}

// handleGetUserByEmail godoc
// @Summary Get User By Email
// @Description Get a user by email
// @Tags users
// @Produce json
// @Param email query string true "User Email"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/users/by-email [get]
func (s *Server) handleGetUserByEmail(c echo.Context) error {
	email := c.QueryParam("email")
	u, err := s.UserService.GetUserByEmail(c.Request().Context(), email)
	if err != nil {
		return err
	}
	return respondOK(c, toUserResponse(u))
}

// handleUpdateProfile godoc
// @Summary Update User Profile
// @Description Update user's name and phone
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param payload body UpdateProfileRequest true "Profile payload"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/users/{id}/profile [patch]
func (s *Server) handleUpdateProfile(c echo.Context) error {
	id := c.Param("id")
	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return respondError(c, 400, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return respondError(c, 400, "invalid request body", err.Error())
	}
	u, err := s.UserService.UpdateProfile(c.Request().Context(), id, req.Name, req.Phone)
	if err != nil {
		return err
	}
	return respondOK(c, toUserResponse(u))
}

// handleChangePassword godoc
// @Summary Change User Password
// @Description Change user's password with current password verification
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param payload body ChangePasswordRequest true "Password payload"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/users/{id}/password [patch]
func (s *Server) handleChangePassword(c echo.Context) error {
	id := c.Param("id")
	var req ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return respondError(c, 400, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return respondError(c, 400, "invalid request body", err.Error())
	}
	if err := s.UserService.ChangePassword(c.Request().Context(), id, req.CurrentPassword, req.NewPassword); err != nil {
		return err
	}
	return respondOK(c, map[string]any{})
}

// handleDeactivateUser godoc
// @Summary Deactivate User
// @Description Deactivate a user account
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/users/{id}/deactivate [patch]
func (s *Server) handleDeactivateUser(c echo.Context) error {
	id := c.Param("id")
	if err := s.UserService.DeactivateUser(c.Request().Context(), id); err != nil {
		return err
	}
	return respondOK(c, map[string]any{})
}
