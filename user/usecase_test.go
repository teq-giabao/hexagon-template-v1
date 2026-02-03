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

//
// TEST AddUser
//
func TestAddUser(t *testing.T) {
	r := new(MockUserRepository)
	uc := user.NewUsecase(r)

	t.Run("should add new user", func(t *testing.T) {
		u := user.User{
			Username: "john",
			Email:    "john@mail.com",
			Password: "secret",
		}

		r.On("CreateUser", mock.Anything, u).Return(nil).Once()

		err := uc.AddUser(context.Background(), u)

		assert.NoError(t, err, "expected no error when adding user")
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
		r.AssertExpectations(t)
	})
}

//
// TEST ListUsers
//
func TestListUsers(t *testing.T) {
	r := new(MockUserRepository)
	uc := user.NewUsecase(r)

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
