package httpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"hexagon/httpserver"
	"hexagon/user"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) AddUser(ctx context.Context, u user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserService) ListUsers(ctx context.Context) ([]user.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]user.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id string) (user.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (user.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserService) UpdateProfile(ctx context.Context, id, name, phone string) (user.User, error) {
	args := m.Called(ctx, id, name, phone)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserService) ChangePassword(ctx context.Context, id, currentPassword, newPassword string) error {
	args := m.Called(ctx, id, currentPassword, newPassword)
	return args.Error(0)
}

func (m *MockUserService) DeactivateUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestUserRoutes_ListUsers_HidesPasswordHash(t *testing.T) {
	svc := new(MockUserService)
	server := httpserver.Default(testConfig())
	server.UserService = svc

	users := []user.User{
		{ID: "u-1", Name: "John", Email: "john@mail.com", PasswordHash: "hashed", Role: user.UserRoleUser, Status: user.UserStatusActive},
	}
	svc.On("ListUsers", mock.Anything).Return(users, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"200"`)
	assert.Contains(t, rec.Body.String(), `"message":"OK"`)
	assert.NotContains(t, rec.Body.String(), "password_hash")
	assert.Contains(t, rec.Body.String(), "\"email\":\"john@mail.com\"")
	svc.AssertExpectations(t)
}

func TestUserRoutes_GetByID(t *testing.T) {
	svc := new(MockUserService)
	server := httpserver.Default(testConfig())
	server.UserService = svc

	u := user.User{ID: "u-1", Name: "John", Email: "john@mail.com", PasswordHash: "hashed"}
	svc.On("GetUserByID", mock.Anything, "u-1").Return(u, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/users/u-1", nil)
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"200"`)
	assert.NotContains(t, rec.Body.String(), "password_hash")
	assert.Contains(t, rec.Body.String(), "\"id\":\"u-1\"")
	svc.AssertExpectations(t)
}

func TestUserRoutes_GetByEmail(t *testing.T) {
	svc := new(MockUserService)
	server := httpserver.Default(testConfig())
	server.UserService = svc

	u := user.User{ID: "u-1", Name: "John", Email: "john@mail.com"}
	svc.On("GetUserByEmail", mock.Anything, "john@mail.com").Return(u, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/users/by-email?email=john@mail.com", nil)
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"200"`)
	assert.Contains(t, rec.Body.String(), "\"email\":\"john@mail.com\"")
	svc.AssertExpectations(t)
}

func TestUserRoutes_UpdateProfile(t *testing.T) {
	svc := new(MockUserService)
	server := httpserver.Default(testConfig())
	server.UserService = svc

	payload := map[string]string{
		"name":  "John Updated",
		"phone": "0987654321",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	updated := user.User{ID: "u-1", Name: "John Updated", Email: "john@mail.com", Phone: "0987654321"}
	svc.On("UpdateProfile", mock.Anything, "u-1", "John Updated", "0987654321").Return(updated, nil).Once()

	req := httptest.NewRequest(http.MethodPatch, "/api/users/u-1/profile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"200"`)
	assert.Contains(t, rec.Body.String(), "\"name\":\"John Updated\"")
	assert.NotContains(t, rec.Body.String(), "password_hash")
	svc.AssertExpectations(t)
}

func TestUserRoutes_ChangePassword(t *testing.T) {
	svc := new(MockUserService)
	server := httpserver.Default(testConfig())
	server.UserService = svc

	payload := map[string]string{
		"currentPassword": "Current123!",
		"newPassword":     "NewPassword1!",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	svc.On("ChangePassword", mock.Anything, "u-1", "Current123!", "NewPassword1!").Return(nil).Once()

	req := httptest.NewRequest(http.MethodPatch, "/api/users/u-1/password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"200"`)
	svc.AssertExpectations(t)
}

func TestUserRoutes_Deactivate(t *testing.T) {
	svc := new(MockUserService)
	server := httpserver.Default(testConfig())
	server.UserService = svc

	svc.On("DeactivateUser", mock.Anything, "u-1").Return(nil).Once()

	req := httptest.NewRequest(http.MethodPatch, "/api/users/u-1/deactivate", nil)
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"200"`)
	svc.AssertExpectations(t)
}
