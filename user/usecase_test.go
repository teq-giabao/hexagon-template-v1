// nolint: funlen
package user_test

import (
	"context"
	"hexagon/user"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock User Repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, u user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) AllUsers(ctx context.Context) ([]user.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]user.User), args.Error(1)
}

type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) Compare(hashed, plain string) error {
	args := m.Called(hashed, plain)
	return args.Error(0)
}

// TEST AddUser
func TestAddUser(t *testing.T) {
	r := new(MockUserRepository)
	h := new(MockPasswordHasher)
	uc := user.NewUsecase(r, h)

	t.Run("should add new user", func(t *testing.T) {
		u := user.User{
			Username: "john",
			Email:    "john@mail.com",
			Password: "secret",
		}
		hashed := "hashed-secret"
		expected := user.User{
			Username: u.Username,
			Email:    u.Email,
			Password: hashed,
		}

		h.On("Hash", u.Password).Return(hashed, nil).Once()
		r.On("CreateUser", mock.Anything, expected).Return(nil).Once()

		err := uc.AddUser(context.Background(), u)

		assert.NoError(t, err, "expected no error when adding user")
		h.AssertExpectations(t)
		r.AssertExpectations(t)
	})

	t.Run("should fail on empty username", func(t *testing.T) {
		u := user.User{
			Username: "",
			Email:    "john@mail.com",
			Password: "secret",
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrInvalidUsername, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail on empty email", func(t *testing.T) {
		u := user.User{
			Username: "john",
			Email:    "",
			Password: "secret",
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrInvalidEmail, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail on empty password", func(t *testing.T) {
		u := user.User{
			Username: "john",
			Email:    "john@mail.com",
			Password: "",
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrInvalidPassword, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})
}

// TEST ListUsers
func TestListUsers(t *testing.T) {
	r := new(MockUserRepository)
	h := new(MockPasswordHasher)
	uc := user.NewUsecase(r, h)

	t.Run("should return list of users", func(t *testing.T) {
		users := []user.User{
			{Username: "john", Email: "john@mail.com", Password: "123"},
			{Username: "jane", Email: "jane@mail.com", Password: "456"},
		}

		r.On("AllUsers", mock.Anything).Return(users, nil).Once()

		result, err := uc.ListUsers(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, users, result)
		r.AssertExpectations(t)
	})
}
