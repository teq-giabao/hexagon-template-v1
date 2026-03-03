package user

import (
	"hexagon/errs"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var (
	ErrNameRequired                   = errs.Errorf(errs.EINVALID, "user: name is required")
	ErrNameTooLong                    = errs.Errorf(errs.EINVALID, "user: name is too long")
	ErrEmailRequired                  = errs.Errorf(errs.EINVALID, "user: email is required")
	ErrEmailInvalidFormat             = errs.Errorf(errs.EINVALID, "user: email format is invalid")
	ErrEmailAlreadyExists             = errs.Errorf(errs.ECONFLICT, "user: email already exists")
	ErrPhoneInvalidFormat             = errs.Errorf(errs.EINVALID, "user: phone format is invalid")
	ErrPasswordRequired               = errs.Errorf(errs.EINVALID, "user: password is required")
	ErrPasswordTooShort               = errs.Errorf(errs.EINVALID, "user: password must be at least 9 characters")
	ErrPasswordTooLong                = errs.Errorf(errs.EINVALID, "user: password must be at most 72 characters")
	ErrPasswordMustContainUppercase   = errs.Errorf(errs.EINVALID, "user: password must contain at least one uppercase letter")
	ErrPasswordMustContainLowercase   = errs.Errorf(errs.EINVALID, "user: password must contain at least one lowercase letter")
	ErrPasswordMustContainNumber      = errs.Errorf(errs.EINVALID, "user: password must contain at least one number")
	ErrPasswordMustContainSpecialChar = errs.Errorf(errs.EINVALID, "user: password must contain at least one special character")
	ErrUserIDRequired                 = errs.Errorf(errs.EINVALID, "user: id is required")
	ErrUserNotFound                   = errs.Errorf(errs.ENOTFOUND, "user: not found")
	ErrCurrentPasswordInvalid         = errs.Errorf(errs.EUNAUTHORIZED, "user: current password is invalid")
	ErrInvalidRole                    = errs.Errorf(errs.EINVALID, "user: invalid role")
	ErrInvalidStatus                  = errs.Errorf(errs.EINVALID, "user: invalid status")
	ErrInvalidCounter                 = errs.Errorf(errs.EINVALID, "user: invalid counter")
	ErrInvalidLockState               = errs.Errorf(errs.EINVALID, "user: invalid lock state")
	ErrInvalidFailedLoginState        = errs.Errorf(errs.EINVALID, "user: invalid failed login state")
	ErrInvalidName                    = ErrNameRequired
	ErrInvalidEmail                   = ErrEmailRequired
	ErrInvalidPassword                = ErrPasswordRequired
	maxNameLength                     = 100
	minPasswordLengthGreaterThanEight = 9
	maxPasswordLength                 = 72
	phoneRegex                        = regexp.MustCompile(`^\d{10}$`)
)

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusLocked   UserStatus = "locked"
)

type User struct {
	ID                  string
	Name                string
	Email               string
	Phone               string
	Password            string
	PasswordHash        string
	Role                UserRole
	Status              UserStatus
	FailedLoginAttempts int
	LockUntil           *time.Time
	LockEscalationLevel int
	LastFailedLoginAt   *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (u User) Validate() error {
	name := strings.TrimSpace(u.Name)
	email := strings.TrimSpace(u.Email)
	phone := strings.TrimSpace(u.Phone)
	password := strings.TrimSpace(u.Password)

	if err := validateName(name); err != nil {
		return err
	}
	if err := validateEmail(email); err != nil {
		return err
	}
	if err := validatePhone(phone); err != nil {
		return err
	}
	if err := validatePassword(password); err != nil {
		return err
	}
	if err := u.validateRoleAndStatus(); err != nil {
		return err
	}
	if err := u.validateCounters(); err != nil {
		return err
	}
	if err := u.validateLockAndFailedLoginState(time.Now().UTC()); err != nil {
		return err
	}

	return nil
}

func (u User) validateRoleAndStatus() error {
	if u.Role != "" && !u.Role.IsValid() {
		return ErrInvalidRole
	}
	if u.Status != "" && !u.Status.IsValid() {
		return ErrInvalidStatus
	}
	return nil
}

func (u User) validateCounters() error {
	if u.FailedLoginAttempts < 0 || u.LockEscalationLevel < 0 {
		return ErrInvalidCounter
	}
	return nil
}

func (u User) validateLockAndFailedLoginState(now time.Time) error {
	if u.Status == UserStatusLocked && (u.LockUntil == nil || !u.LockUntil.After(now)) {
		return ErrInvalidLockState
	}
	if u.Status != UserStatusLocked && u.LockUntil != nil {
		return ErrInvalidLockState
	}
	if u.LastFailedLoginAt != nil && u.LastFailedLoginAt.After(now) {
		return ErrInvalidFailedLoginState
	}
	if u.FailedLoginAttempts > 0 && u.LastFailedLoginAt == nil {
		return ErrInvalidFailedLoginState
	}
	return nil
}

func (r UserRole) IsValid() bool {
	return r == UserRoleUser || r == UserRoleAdmin
}

func (s UserStatus) IsValid() bool {
	return s == UserStatusActive || s == UserStatusInactive || s == UserStatusLocked
}

func containsUppercase(value string) bool {
	for _, r := range value {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func containsLowercase(value string) bool {
	for _, r := range value {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

func containsNumber(value string) bool {
	for _, r := range value {
		if unicode.IsNumber(r) {
			return true
		}
	}
	return false
}

func containsSpecialChar(value string) bool {
	for _, r := range value {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return true
		}
	}
	return false
}

func validateName(name string) error {
	if name == "" {
		return ErrNameRequired
	}
	if len(name) > maxNameLength {
		return ErrNameTooLong
	}
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return ErrEmailRequired
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrEmailInvalidFormat
	}
	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return ErrPasswordRequired
	}
	if len(password) < minPasswordLengthGreaterThanEight {
		return ErrPasswordTooShort
	}
	if len(password) > maxPasswordLength {
		return ErrPasswordTooLong
	}
	if !containsUppercase(password) {
		return ErrPasswordMustContainUppercase
	}
	if !containsLowercase(password) {
		return ErrPasswordMustContainLowercase
	}
	if !containsNumber(password) {
		return ErrPasswordMustContainNumber
	}
	if !containsSpecialChar(password) {
		return ErrPasswordMustContainSpecialChar
	}
	return nil
}

func validatePhone(phone string) error {
	if phone == "" {
		return nil
	}
	if !phoneRegex.MatchString(phone) {
		return ErrPhoneInvalidFormat
	}
	return nil
}
