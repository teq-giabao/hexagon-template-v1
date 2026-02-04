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
	ErrOAuthNotConfigured  = errors.New("oauth provider not configured")
)

type Service interface {
	Login(ctx context.Context, email, password string) (TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (TokenPair, error)
	GoogleAuthURL(state string) (string, error)
	LoginWithGoogle(ctx context.Context, code string) (TokenPair, error)
}

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (user.User, error)
	CreateUser(ctx context.Context, u user.User) error
}

type LoginAttempt struct {
	FailedCount int
	JailedUntil time.Time
}

type LoginAttemptRepository interface {
	Get(ctx context.Context, email string) (LoginAttempt, error)
	Save(ctx context.Context, email string, attempt LoginAttempt) error
	Reset(ctx context.Context, email string) error
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
	attemptsRepo   LoginAttemptRepository
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
	attemptsRepo LoginAttemptRepository,
	passwordHasher PasswordHasher,
	tokenProvider TokenProvider,
	googleProvider GoogleOAuthProvider,
) *Usecase {
	return &Usecase{
		userRepo:       userRepo,
		attemptsRepo:   attemptsRepo,
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

func (uc *Usecase) Login(ctx context.Context, email, password string) (TokenPair, error) {
	attempt, err := uc.attemptsRepo.Get(ctx, email)
	if err != nil {
		return TokenPair{}, err
	}

	if !attempt.JailedUntil.IsZero() {
		if attempt.JailedUntil.After(uc.now()) {
			return TokenPair{}, ErrAccountLocked
		}
		attempt.JailedUntil = time.Time{}
		attempt.FailedCount = 0
		if err := uc.attemptsRepo.Save(ctx, email, attempt); err != nil {
			return TokenPair{}, err
		}
	}

	u, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err := uc.recordFailure(ctx, email, attempt); err != nil {
			return TokenPair{}, err
		}
		return TokenPair{}, ErrInvalidCredentials
	}

	// Compare password
	if err := uc.passwordHasher.Compare(u.Password, password); err != nil {
		if err := uc.recordFailure(ctx, email, attempt); err != nil {
			return TokenPair{}, err
		}
		return TokenPair{}, ErrInvalidCredentials
	}

	if err := uc.attemptsRepo.Reset(ctx, email); err != nil {
		return TokenPair{}, err
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
		return "", ErrInvalidOAuthUser
	}
	return uc.googleProvider.AuthCodeURL(state), nil
}

func (uc *Usecase) LoginWithGoogle(ctx context.Context, code string) (TokenPair, error) {
	if uc.googleProvider == nil {
		return TokenPair{}, ErrOAuthNotConfigured
	}
	if strings.TrimSpace(code) == "" {
		return TokenPair{}, ErrInvalidOAuthUser
	}

	oauthUser, err := uc.googleProvider.Exchange(ctx, code)
	if err != nil {
		return TokenPair{}, err
	}
	if !oauthUser.EmailVerified || strings.TrimSpace(oauthUser.Email) == "" {
		return TokenPair{}, ErrInvalidOAuthUser
	}

	u, err := uc.userRepo.GetByEmail(ctx, oauthUser.Email)
	if err != nil {
		if errors.Is(err, user.ErrInvalidEmail) {
			password, err := generateRandomPassword(32)
			if err != nil {
				return TokenPair{}, err
			}
			hashed, err := uc.passwordHasher.Hash(password)
			if err != nil {
				return TokenPair{}, err
			}
			u = user.User{
				Username: oauthUser.Email,
				Email:    oauthUser.Email,
				Password: hashed,
			}
			if err := uc.userRepo.CreateUser(ctx, u); err != nil {
				return TokenPair{}, err
			}
			created, err := uc.userRepo.GetByEmail(ctx, oauthUser.Email)
			if err != nil {
				return TokenPair{}, err
			}
			u = created
		} else {
			return TokenPair{}, err
		}
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

func (uc *Usecase) recordFailure(ctx context.Context, email string, attempt LoginAttempt) error {
	attempt.FailedCount++
	if attempt.FailedCount >= uc.maxRetries {
		attempt.FailedCount = 0
		attempt.JailedUntil = uc.now().Add(uc.jailDuration)
	}
	return uc.attemptsRepo.Save(ctx, email, attempt)
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
