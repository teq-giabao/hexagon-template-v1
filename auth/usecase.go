// nolint: nestif
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"hexagon/user"
	"strings"
	"time"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrAccountLocked       = errors.New("account temporarily locked")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidOAuthUser    = errors.New("invalid oauth user")
	ErrMissingState        = errors.New("missing oauth state")
	ErrMissingCode         = errors.New("missing oauth code")
	ErrMissingEmail        = errors.New("missing oauth email")
	ErrUnverifiedEmail     = errors.New("unverified oauth email")
	ErrOAuthNotConfigured  = errors.New("oauth provider not configured")
)

type Service interface {
	Register(ctx context.Context, name, email, phone, password string) (TokenPair, error)
	Login(ctx context.Context, email, password string) (TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (TokenPair, error)
	GoogleAuthURL(state string) (string, error)
	LoginWithGoogle(ctx context.Context, code string) (TokenPair, error)
}

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (user.User, error)
	CreateUser(ctx context.Context, u user.User) error
	CreateUserTx(ctx context.Context, u user.User, fn func(created user.User) error) (user.User, error)
	UpdateAuthState(
		ctx context.Context,
		id string,
		failedLoginAttempts int,
		lockUntil *time.Time,
		lockEscalationLevel int,
		lastFailedLoginAt *time.Time,
		status user.UserStatus,
	) error
}

type PasswordHasher interface {
	Compare(hashed, plain string) error
	Hash(password string) (string, error)
}

type TokenProvider interface {
	GenerateAccessToken(u user.User) (string, error)
	GenerateRefreshToken(u user.User) (string, error)
	ParseRefreshToken(refreshToken string) (user.User, error)
}

type OAuthUser struct {
	Email         string
	Name          string
	EmailVerified bool
}

type GoogleOAuthProvider interface {
	AuthCodeURL(state string) string
	Exchange(ctx context.Context, code string) (OAuthUser, error)
}

type Usecase struct {
	userRepo       UserRepository
	passwordHasher PasswordHasher
	tokenProvider  TokenProvider
	googleProvider GoogleOAuthProvider
	maxRetries     int
	jailDuration   time.Duration
	now            func() time.Time
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func NewUsecase(
	userRepo UserRepository,
	passwordHasher PasswordHasher,
	tokenProvider TokenProvider,
	googleProvider GoogleOAuthProvider,
) *Usecase {
	return &Usecase{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		tokenProvider:  tokenProvider,
		googleProvider: googleProvider,
		maxRetries:     5,
		jailDuration:   15 * time.Minute,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (uc *Usecase) Register(ctx context.Context, name, email, phone, password string) (TokenPair, error) {
	newUser := user.User{
		Name:     strings.TrimSpace(name),
		Email:    strings.TrimSpace(email),
		Phone:    strings.TrimSpace(phone),
		Password: strings.TrimSpace(password),
		Role:     user.UserRoleUser,
		Status:   user.UserStatusActive,
	}
	if err := newUser.Validate(); err != nil {
		return TokenPair{}, err
	}

	hashed, err := uc.passwordHasher.Hash(newUser.Password)
	if err != nil {
		return TokenPair{}, err
	}
	newUser.Password = ""
	newUser.PasswordHash = hashed

	var tokens TokenPair
	_, err = uc.userRepo.CreateUserTx(ctx, newUser, func(created user.User) error {
		accessToken, err := uc.tokenProvider.GenerateAccessToken(created)
		if err != nil {
			return err
		}
		refreshToken, err := uc.tokenProvider.GenerateRefreshToken(created)
		if err != nil {
			return err
		}
		tokens = TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}
		return nil
	})
	if err != nil {
		return TokenPair{}, err
	}

	return tokens, nil
}

func (uc *Usecase) Login(ctx context.Context, email, password string) (TokenPair, error) {
	now := uc.now()

	u, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return TokenPair{}, ErrInvalidCredentials
	}

	u, err = uc.handleLockState(ctx, u, now)
	if err != nil {
		return TokenPair{}, err
	}

	// Compare password
	if err := uc.passwordHasher.Compare(u.PasswordHash, password); err != nil {
		if err := uc.recordFailure(ctx, u, now); err != nil {
			return TokenPair{}, err
		}
		return TokenPair{}, ErrInvalidCredentials
	}

	if err := uc.resetAuthState(ctx, u); err != nil {
		return TokenPair{}, err
	}

	return uc.issueTokens(u)
}

func (uc *Usecase) handleLockState(ctx context.Context, u user.User, now time.Time) (user.User, error) {
	if u.Status != user.UserStatusLocked || u.LockUntil == nil {
		return u, nil
	}
	if u.LockUntil.After(now) {
		return user.User{}, ErrAccountLocked
	}

	if err := uc.userRepo.UpdateAuthState(
		ctx,
		u.ID,
		0,
		nil,
		u.LockEscalationLevel,
		nil,
		user.UserStatusActive,
	); err != nil {
		return user.User{}, err
	}
	u.Status = user.UserStatusActive
	u.LockUntil = nil
	u.FailedLoginAttempts = 0
	u.LastFailedLoginAt = nil
	return u, nil
}

func (uc *Usecase) resetAuthState(ctx context.Context, u user.User) error {
	return uc.userRepo.UpdateAuthState(
		ctx,
		u.ID,
		0,
		nil,
		u.LockEscalationLevel,
		nil,
		u.Status,
	)
}

func (uc *Usecase) issueTokens(u user.User) (TokenPair, error) {
	accessToken, err := uc.tokenProvider.GenerateAccessToken(u)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, err := uc.tokenProvider.GenerateRefreshToken(u)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (uc *Usecase) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	u, err := uc.tokenProvider.ParseRefreshToken(refreshToken)
	if err != nil {
		return TokenPair{}, ErrInvalidRefreshToken
	}

	accessToken, err := uc.tokenProvider.GenerateAccessToken(u)
	if err != nil {
		return TokenPair{}, err
	}

	newRefreshToken, err := uc.tokenProvider.GenerateRefreshToken(u)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (uc *Usecase) GoogleAuthURL(state string) (string, error) {
	if uc.googleProvider == nil {
		return "", ErrOAuthNotConfigured
	}
	if strings.TrimSpace(state) == "" {
		return "", ErrMissingState
	}
	return uc.googleProvider.AuthCodeURL(state), nil
}

func (uc *Usecase) LoginWithGoogle(ctx context.Context, code string) (TokenPair, error) {
	if uc.googleProvider == nil {
		return TokenPair{}, ErrOAuthNotConfigured
	}
	if strings.TrimSpace(code) == "" {
		return TokenPair{}, ErrMissingCode
	}

	oauthUser, err := uc.googleProvider.Exchange(ctx, code)
	if err != nil {
		return TokenPair{}, err
	}
	if strings.TrimSpace(oauthUser.Email) == "" {
		return TokenPair{}, ErrMissingEmail
	}
	if !oauthUser.EmailVerified {
		return TokenPair{}, ErrUnverifiedEmail
	}

	u, tokens, created, err := uc.getOrCreateUserFromOAuth(ctx, oauthUser)
	if err != nil {
		return TokenPair{}, err
	}
	if created {
		return tokens, nil
	}

	accessToken, err := uc.tokenProvider.GenerateAccessToken(u)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, err := uc.tokenProvider.GenerateRefreshToken(u)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (uc *Usecase) getOrCreateUserFromOAuth(ctx context.Context, oauthUser OAuthUser) (user.User, TokenPair, bool, error) {
	u, err := uc.userRepo.GetByEmail(ctx, oauthUser.Email)
	if err == nil {
		return u, TokenPair{}, false, nil
	}
	if !errors.Is(err, user.ErrUserNotFound) {
		return user.User{}, TokenPair{}, false, err
	}

	password, err := generateRandomPassword(32)
	if err != nil {
		return user.User{}, TokenPair{}, false, err
	}
	hashed, err := uc.passwordHasher.Hash(password)
	if err != nil {
		return user.User{}, TokenPair{}, false, err
	}
	name := strings.TrimSpace(oauthUser.Name)
	if name == "" {
		name, err = generateRandomName(16)
		if err != nil {
			return user.User{}, TokenPair{}, false, err
		}
	}
	newUser := user.User{
		Name:         name,
		Email:        oauthUser.Email,
		PasswordHash: hashed,
		Role:         user.UserRoleUser,
		Status:       user.UserStatusActive,
	}

	var tokens TokenPair
	created, err := uc.userRepo.CreateUserTx(ctx, newUser, func(created user.User) error {
		accessToken, err := uc.tokenProvider.GenerateAccessToken(created)
		if err != nil {
			return err
		}
		refreshToken, err := uc.tokenProvider.GenerateRefreshToken(created)
		if err != nil {
			return err
		}
		tokens = TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}
		return nil
	})
	if err != nil {
		return user.User{}, TokenPair{}, false, err
	}

	return created, tokens, true, nil
}

func (uc *Usecase) recordFailure(ctx context.Context, u user.User, now time.Time) error {
	failedCount := u.FailedLoginAttempts + 1
	lockUntil := u.LockUntil
	lockEscalationLevel := u.LockEscalationLevel
	status := u.Status
	lastFailedLoginAt := now

	if failedCount >= uc.maxRetries {
		failedCount = 0
		t := now.Add(uc.jailDuration)
		lockUntil = &t
		lockEscalationLevel++
		status = user.UserStatusLocked
	}

	return uc.userRepo.UpdateAuthState(
		ctx,
		u.ID,
		failedCount,
		lockUntil,
		lockEscalationLevel,
		&lastFailedLoginAt,
		status,
	)
}

func generateRandomPassword(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("invalid password length")
	}
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func generateRandomName(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("invalid name length")
	}
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "user_" + base64.RawURLEncoding.EncodeToString(buf), nil
}
