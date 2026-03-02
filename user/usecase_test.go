// nolint: funlen
package user_test

import (
	"context"
	"hexagon/user"
	"testing"
	"time"

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

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (user.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserRepository) UpdateProfile(ctx context.Context, id, name, phone string) (user.User, error) {
	args := m.Called(ctx, id, name, phone)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserRepository) UpdatePasswordHash(ctx context.Context, id, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateStatus(ctx context.Context, id string, status user.UserStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
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
	t.Run("should add new user", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)

		u := user.User{
			Name:     "john",
			Email:    "john@mail.com",
			Password: "Secret123!",
		}
		hashed := "hashed-secret"
		expected := user.User{
			Name:         u.Name,
			Email:        u.Email,
			PasswordHash: hashed,
			Role:         user.UserRoleUser,
			Status:       user.UserStatusActive,
		}

		h.On("Hash", u.Password).Return(hashed, nil).Once()
		r.On("CreateUser", mock.Anything, expected).Return(nil).Once()

		err := uc.AddUser(context.Background(), u)

		assert.NoError(t, err, "expected no error when adding user")
		h.AssertExpectations(t)
		r.AssertExpectations(t)
	})

	t.Run("should fail on empty name", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)

		u := user.User{
			Name:     "",
			Email:    "john@mail.com",
			Password: "Secret123!",
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrNameRequired, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail on empty email", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)

		u := user.User{
			Name:     "john",
			Email:    "",
			Password: "Secret123!",
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrEmailRequired, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail on invalid email format", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)

		u := user.User{
			Name:     "john",
			Email:    "invalid-email",
			Password: "Secret123!",
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrEmailInvalidFormat, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail on empty password", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)

		u := user.User{
			Name:     "john",
			Email:    "john@mail.com",
			Password: "",
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrPasswordRequired, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail on short password", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)

		u := user.User{
			Name:     "john",
			Email:    "john@mail.com",
			Password: "a1b2c3d4",
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrPasswordTooShort, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail when password has no uppercase letter", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)

		u := user.User{
			Name:     "john",
			Email:    "john@mail.com",
			Password: "secret123!",
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrPasswordMustContainUppercase, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail when password has no lowercase letter", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)

		u := user.User{
			Name:     "john",
			Email:    "john@mail.com",
			Password: "SECRET123!",
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrPasswordMustContainLowercase, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail when password has no number", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)

		u := user.User{
			Name:     "john",
			Email:    "john@mail.com",
			Password: "Secretabc!",
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrPasswordMustContainNumber, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail when password has no special character", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)

		u := user.User{
			Name:     "john",
			Email:    "john@mail.com",
			Password: "Secret1234",
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrPasswordMustContainSpecialChar, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail on invalid phone", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)

		u := user.User{
			Name:     "john",
			Email:    "john@mail.com",
			Phone:    "12345",
			Password: "Secret123!",
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrPhoneInvalidFormat, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail on invalid role", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)

		u := user.User{
			Name:     "john",
			Email:    "john@mail.com",
			Password: "Secret123!",
			Role:     user.UserRole("owner"),
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrInvalidRole, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail on invalid status", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)

		u := user.User{
			Name:     "john",
			Email:    "john@mail.com",
			Password: "Secret123!",
			Status:   user.UserStatus("pending"),
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrInvalidStatus, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail on negative counters", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)

		u := user.User{
			Name:                "john",
			Email:               "john@mail.com",
			Password:            "Secret123!",
			FailedLoginAttempts: -1,
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrInvalidCounter, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail when locked status has no lock_until", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)

		u := user.User{
			Name:     "john",
			Email:    "john@mail.com",
			Password: "Secret123!",
			Status:   user.UserStatusLocked,
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrInvalidLockState, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail when lock_until is in the past", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)
		past := time.Now().UTC().Add(-1 * time.Minute)

		u := user.User{
			Name:      "john",
			Email:     "john@mail.com",
			Password:  "Secret123!",
			Status:    user.UserStatusLocked,
			LockUntil: &past,
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrInvalidLockState, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail when non-locked status still has lock_until", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)
		future := time.Now().UTC().Add(10 * time.Minute)

		u := user.User{
			Name:      "john",
			Email:     "john@mail.com",
			Password:  "Secret123!",
			Status:    user.UserStatusActive,
			LockUntil: &future,
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrInvalidLockState, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail when failed attempts exist without last_failed_login_at", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)

		u := user.User{
			Name:                "john",
			Email:               "john@mail.com",
			Password:            "Secret123!",
			FailedLoginAttempts: 2,
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrInvalidFailedLoginState, err)
		h.AssertNotCalled(t, "Hash", mock.Anything)
		r.AssertExpectations(t)
	})

	t.Run("should fail when last_failed_login_at is in the future", func(t *testing.T) {
		r := new(MockUserRepository)
		h := new(MockPasswordHasher)
		uc := user.NewUsecase(r, h)
		future := time.Now().UTC().Add(1 * time.Minute)

		u := user.User{
			Name:              "john",
			Email:             "john@mail.com",
			Password:          "Secret123!",
			LastFailedLoginAt: &future,
		}

		err := uc.AddUser(context.Background(), u)

		assert.Equal(t, user.ErrInvalidFailedLoginState, err)
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
			{Name: "john", Email: "john@mail.com", PasswordHash: "123"},
			{Name: "jane", Email: "jane@mail.com", PasswordHash: "456"},
		}

		r.On("AllUsers", mock.Anything).Return(users, nil).Once()

		result, err := uc.ListUsers(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, users, result)
		r.AssertExpectations(t)
	})
}

func TestGetUserByID(t *testing.T) {
	r := new(MockUserRepository)
	h := new(MockPasswordHasher)
	uc := user.NewUsecase(r, h)

	t.Run("should fail on empty id", func(t *testing.T) {
		_, err := uc.GetUserByID(context.Background(), "")
		assert.Equal(t, user.ErrUserIDRequired, err)
		r.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
	})

	t.Run("should get user by id", func(t *testing.T) {
		u := user.User{ID: "u-1", Name: "john", Email: "john@mail.com"}
		r.On("GetByID", mock.Anything, "u-1").Return(u, nil).Once()
		got, err := uc.GetUserByID(context.Background(), "u-1")
		assert.NoError(t, err)
		assert.Equal(t, u, got)
		r.AssertExpectations(t)
	})
}

func TestGetUserByEmail(t *testing.T) {
	r := new(MockUserRepository)
	h := new(MockPasswordHasher)
	uc := user.NewUsecase(r, h)

	t.Run("should fail on invalid email format", func(t *testing.T) {
		_, err := uc.GetUserByEmail(context.Background(), "bad-email")
		assert.Equal(t, user.ErrEmailInvalidFormat, err)
		r.AssertNotCalled(t, "GetByEmail", mock.Anything, mock.Anything)
	})

	t.Run("should get user by email", func(t *testing.T) {
		u := user.User{ID: "u-1", Name: "john", Email: "john@mail.com"}
		r.On("GetByEmail", mock.Anything, "john@mail.com").Return(u, nil).Once()
		got, err := uc.GetUserByEmail(context.Background(), "john@mail.com")
		assert.NoError(t, err)
		assert.Equal(t, u, got)
		r.AssertExpectations(t)
	})
}

func TestUpdateProfile(t *testing.T) {
	r := new(MockUserRepository)
	h := new(MockPasswordHasher)
	uc := user.NewUsecase(r, h)

	t.Run("should fail on empty id", func(t *testing.T) {
		_, err := uc.UpdateProfile(context.Background(), "", "john", "0123")
		assert.Equal(t, user.ErrUserIDRequired, err)
		r.AssertNotCalled(t, "UpdateProfile", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("should fail on invalid name", func(t *testing.T) {
		_, err := uc.UpdateProfile(context.Background(), "u-1", "", "0123")
		assert.Equal(t, user.ErrNameRequired, err)
		r.AssertNotCalled(t, "UpdateProfile", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("should update profile", func(t *testing.T) {
		updated := user.User{ID: "u-1", Name: "John New", Phone: "0999"}
		r.On("UpdateProfile", mock.Anything, "u-1", "John New", "0999").Return(updated, nil).Once()
		got, err := uc.UpdateProfile(context.Background(), "u-1", "John New", "0999")
		assert.NoError(t, err)
		assert.Equal(t, updated, got)
		r.AssertExpectations(t)
	})
}

func TestChangePassword(t *testing.T) {
	r := new(MockUserRepository)
	h := new(MockPasswordHasher)
	uc := user.NewUsecase(r, h)

	t.Run("should fail on empty id", func(t *testing.T) {
		err := uc.ChangePassword(context.Background(), "", "Current123!", "NewPassword1!")
		assert.Equal(t, user.ErrUserIDRequired, err)
	})

	t.Run("should fail on invalid current password", func(t *testing.T) {
		err := uc.ChangePassword(context.Background(), "u-1", "", "NewPassword1!")
		assert.Equal(t, user.ErrCurrentPasswordInvalid, err)
	})

	t.Run("should fail on invalid new password", func(t *testing.T) {
		err := uc.ChangePassword(context.Background(), "u-1", "Current123!", "short")
		assert.Equal(t, user.ErrPasswordTooShort, err)
	})

	t.Run("should fail when current password does not match", func(t *testing.T) {
		u := user.User{ID: "u-1", PasswordHash: "hashed-current"}
		r.On("GetByID", mock.Anything, "u-1").Return(u, nil).Once()
		h.On("Compare", "hashed-current", "Current123!").Return(assert.AnError).Once()
		err := uc.ChangePassword(context.Background(), "u-1", "Current123!", "NewPassword1!")
		assert.Equal(t, user.ErrCurrentPasswordInvalid, err)
		r.AssertExpectations(t)
		h.AssertExpectations(t)
	})

	t.Run("should change password", func(t *testing.T) {
		u := user.User{ID: "u-1", PasswordHash: "hashed-current"}
		r.On("GetByID", mock.Anything, "u-1").Return(u, nil).Once()
		h.On("Compare", "hashed-current", "Current123!").Return(nil).Once()
		h.On("Hash", "NewPassword1!").Return("hashed-new", nil).Once()
		r.On("UpdatePasswordHash", mock.Anything, "u-1", "hashed-new").Return(nil).Once()

		err := uc.ChangePassword(context.Background(), "u-1", "Current123!", "NewPassword1!")
		assert.NoError(t, err)
		r.AssertExpectations(t)
		h.AssertExpectations(t)
	})
}

func TestDeactivateUser(t *testing.T) {
	r := new(MockUserRepository)
	h := new(MockPasswordHasher)
	uc := user.NewUsecase(r, h)

	t.Run("should fail on empty id", func(t *testing.T) {
		err := uc.DeactivateUser(context.Background(), "")
		assert.Equal(t, user.ErrUserIDRequired, err)
		r.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("should deactivate user", func(t *testing.T) {
		r.On("UpdateStatus", mock.Anything, "u-1", user.UserStatusInactive).Return(nil).Once()
		err := uc.DeactivateUser(context.Background(), "u-1")
		assert.NoError(t, err)
		r.AssertExpectations(t)
	})
}
