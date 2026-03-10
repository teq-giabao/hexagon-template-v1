// nolint: nestif
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"hexagon/user"
)

var (
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrInvalidAccessToken       = errors.New("invalid access token")
	ErrAccountLocked            = errors.New("account temporarily locked")
	ErrInvalidRefreshToken      = errors.New("invalid refresh token")
	ErrInvalidOAuthUser         = errors.New("invalid oauth user")
	ErrInvalidResetToken        = errors.New("invalid reset token")
	ErrMissingState             = errors.New("missing oauth state")
	ErrMissingCode              = errors.New("missing oauth code")
	ErrMissingEmail             = errors.New("missing oauth email")
	ErrUnverifiedEmail          = errors.New("unverified oauth email")
	ErrOAuthNotConfigured       = errors.New("oauth provider not configured")
	ErrMailerNotConfigured      = errors.New("password reset mailer not configured")
	ErrPasswordAuthNotAvailable = errors.New("password authentication is not available for this account")
	ErrEmailRegisteredWithOAuth = errors.New("email is already registered with oauth provider")
)

type Service interface {
	Register(ctx context.Context, name, email, phone, password string) (TokenPair, error)
	Login(ctx context.Context, email, password string) (TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, refreshToken string) (TokenPair, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, resetToken, newPassword string) error
	Me(ctx context.Context, accessToken string) (user.User, error)
	GoogleAuthURL(state string) (string, error)
	LoginWithGoogle(ctx context.Context, code string) (TokenPair, error)
}

type OAuthProvider string

const (
	OAuthProviderGoogle OAuthProvider = "google"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (user.User, error)
	GetByID(ctx context.Context, id string) (user.User, error)
	CreateUser(ctx context.Context, u user.User) error
	CreateUserTx(ctx context.Context, u user.User, fn func(created user.User) error) (user.User, error)
	UpdatePasswordHash(ctx context.Context, id, passwordHash string) error
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

type OAuthProviderAccountRepository interface {
	GetUserIDByProvider(ctx context.Context, provider OAuthProvider, providerUserID string) (string, error)
	Upsert(ctx context.Context, userID string, provider OAuthProvider, providerUserID, providerEmail string) error
}

type RefreshTokenRepository interface {
	Save(ctx context.Context, token RefreshToken) error
	GetActiveByHash(ctx context.Context, tokenHash string) (RefreshToken, error)
	RevokeByHash(ctx context.Context, tokenHash string, revokedAt time.Time) error
	RevokeAllByUserID(ctx context.Context, userID string, revokedAt time.Time) error
}

type PasswordResetTokenRepository interface {
	Save(ctx context.Context, token PasswordResetToken) error
	GetActiveByHash(ctx context.Context, tokenHash string) (PasswordResetToken, error)
	MarkUsedByHash(ctx context.Context, tokenHash string, usedAt time.Time) error
}

type RefreshToken struct {
	UserID    string
	TokenHash string
	UserAgent string
	IPAddress string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type PasswordResetToken struct {
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
}

type PasswordHasher interface {
	Compare(hashed, plain string) error
	Hash(password string) (string, error)
}

type TokenProvider interface {
	GenerateAccessToken(u user.User) (string, error)
	GenerateRefreshToken(u user.User) (string, error)
	ParseAccessToken(accessToken string) (user.User, error)
	ParseRefreshToken(refreshToken string) (user.User, error)
}

type OAuthUser struct {
	ProviderUserID string
	Email          string
	Name           string
	EmailVerified  bool
}

type GoogleOAuthProvider interface {
	AuthCodeURL(state string) string
	Exchange(ctx context.Context, code string) (OAuthUser, error)
}

type PasswordResetMailer interface {
	SendResetPasswordEmail(ctx context.Context, toEmail, toName, resetURL string) error
}

type Usecase struct {
	userRepo       UserRepository
	oauthRepo      OAuthProviderAccountRepository
	refreshRepo    RefreshTokenRepository
	resetTokenRepo PasswordResetTokenRepository
	passwordHasher PasswordHasher
	tokenProvider  TokenProvider
	googleProvider GoogleOAuthProvider
	resetMailer    PasswordResetMailer
	resetBaseURL   string
	maxRetries     int
	jailDuration   time.Duration
	resetTTL       time.Duration
	now            func() time.Time
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// NowForTest overrides clock in tests.
func (uc *Usecase) NowForTest(now time.Time) {
	uc.now = func() time.Time { return now }
}

func NewUsecase(
	userRepo UserRepository,
	oauthRepo OAuthProviderAccountRepository,
	refreshRepo RefreshTokenRepository,
	resetTokenRepo PasswordResetTokenRepository,
	passwordHasher PasswordHasher,
	tokenProvider TokenProvider,
	googleProvider GoogleOAuthProvider,
	resetMailer PasswordResetMailer,
	resetBaseURL string,
) *Usecase {
	return &Usecase{
		userRepo:       userRepo,
		oauthRepo:      oauthRepo,
		refreshRepo:    refreshRepo,
		resetTokenRepo: resetTokenRepo,
		passwordHasher: passwordHasher,
		tokenProvider:  tokenProvider,
		googleProvider: googleProvider,
		resetMailer:    resetMailer,
		resetBaseURL:   strings.TrimSpace(resetBaseURL),
		maxRetries:     5,
		jailDuration:   15 * time.Minute,
		resetTTL:       30 * time.Minute,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// nolint: funlen
func (uc *Usecase) Register(ctx context.Context, name, email, phone, password string) (TokenPair, error) {
	existing, err := uc.userRepo.GetByEmail(ctx, strings.TrimSpace(email))
	if err == nil && strings.TrimSpace(existing.PasswordHash) == "" {
		return TokenPair{}, ErrEmailRegisteredWithOAuth
	}

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

	var (
		tokens  TokenPair
		created user.User
	)

	created, err = uc.userRepo.CreateUserTx(ctx, newUser, func(created user.User) error {
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

	if err := uc.refreshRepo.Save(ctx, RefreshToken{
		UserID:    created.ID,
		TokenHash: hashToken(tokens.RefreshToken),
		UserAgent: clientInfoFromContext(ctx).UserAgent,
		IPAddress: clientInfoFromContext(ctx).IPAddress,
		ExpiresAt: uc.now().Add(uc.tokenProviderRefreshTTL()),
	}); err != nil {
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

	if strings.TrimSpace(u.PasswordHash) == "" {
		return TokenPair{}, ErrPasswordAuthNotAvailable
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

	return uc.issueTokens(ctx, u)
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

func (uc *Usecase) issueTokens(ctx context.Context, u user.User) (TokenPair, error) {
	accessToken, err := uc.tokenProvider.GenerateAccessToken(u)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, err := uc.tokenProvider.GenerateRefreshToken(u)
	if err != nil {
		return TokenPair{}, err
	}

	refreshExp := uc.now().Add(uc.tokenProviderRefreshTTL())
	if err := uc.refreshRepo.Save(ctx, RefreshToken{
		UserID:    u.ID,
		TokenHash: hashToken(refreshToken),
		UserAgent: clientInfoFromContext(ctx).UserAgent,
		IPAddress: clientInfoFromContext(ctx).IPAddress,
		ExpiresAt: refreshExp,
	}); err != nil {
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

	tokenHash := hashToken(refreshToken)

	stored, err := uc.refreshRepo.GetActiveByHash(ctx, tokenHash)
	if err != nil {
		return TokenPair{}, ErrInvalidRefreshToken
	}

	if stored.ExpiresAt.Before(uc.now()) {
		return TokenPair{}, ErrInvalidRefreshToken
	}

	if !sameClientSession(stored, clientInfoFromContext(ctx)) {
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

	now := uc.now()
	if err := uc.refreshRepo.RevokeByHash(ctx, tokenHash, now); err != nil {
		return TokenPair{}, err
	}

	if err := uc.refreshRepo.Save(ctx, RefreshToken{
		UserID:    u.ID,
		TokenHash: hashToken(newRefreshToken),
		UserAgent: clientInfoFromContext(ctx).UserAgent,
		IPAddress: clientInfoFromContext(ctx).IPAddress,
		ExpiresAt: now.Add(uc.tokenProviderRefreshTTL()),
	}); err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (uc *Usecase) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return ErrInvalidRefreshToken
	}

	if _, err := uc.tokenProvider.ParseRefreshToken(refreshToken); err != nil {
		return ErrInvalidRefreshToken
	}

	return uc.refreshRepo.RevokeByHash(ctx, hashToken(refreshToken), uc.now())
}

func (uc *Usecase) ForgotPassword(ctx context.Context, email string) error {
	email = strings.TrimSpace(email)
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrMissingEmail
	}

	u, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Avoid user-enumeration: respond success even when user does not exist.
		return nil
	}

	if strings.TrimSpace(u.PasswordHash) == "" {
		return ErrPasswordAuthNotAvailable
	}

	resetToken, err := generateRandomPassword(24)
	if err != nil {
		return err
	}

	if err := uc.resetTokenRepo.Save(ctx, PasswordResetToken{
		UserID:    u.ID,
		TokenHash: hashToken(resetToken),
		ExpiresAt: uc.now().Add(uc.resetTTL),
	}); err != nil {
		return err
	}

	if uc.resetMailer == nil || uc.resetBaseURL == "" {
		return ErrMailerNotConfigured
	}

	resetURL, err := composeResetPasswordURL(uc.resetBaseURL, resetToken)
	if err != nil {
		return err
	}

	return uc.resetMailer.SendResetPasswordEmail(ctx, u.Email, u.Name, resetURL)
}

func (uc *Usecase) ResetPassword(ctx context.Context, resetToken, newPassword string) error {
	resetToken = strings.TrimSpace(resetToken)
	if resetToken == "" {
		return ErrInvalidResetToken
	}

	// Reuse domain password policy.
	passCheck := user.User{
		Name:     "temp",
		Email:    "temp@example.com",
		Password: strings.TrimSpace(newPassword),
		Role:     user.UserRoleUser,
		Status:   user.UserStatusActive,
	}
	if err := passCheck.Validate(); err != nil {
		return err
	}

	tokenHash := hashToken(resetToken)

	entry, err := uc.resetTokenRepo.GetActiveByHash(ctx, tokenHash)
	if err != nil || entry.ExpiresAt.Before(uc.now()) {
		return ErrInvalidResetToken
	}

	u, err := uc.userRepo.GetByID(ctx, entry.UserID)
	if err != nil {
		return ErrInvalidResetToken
	}

	if strings.TrimSpace(u.PasswordHash) == "" {
		return ErrPasswordAuthNotAvailable
	}

	hashed, err := uc.passwordHasher.Hash(passCheck.Password)
	if err != nil {
		return err
	}

	if err := uc.userRepo.UpdatePasswordHash(ctx, u.ID, hashed); err != nil {
		return err
	}

	now := uc.now()
	if err := uc.refreshRepo.RevokeAllByUserID(ctx, u.ID, now); err != nil {
		return err
	}

	return uc.resetTokenRepo.MarkUsedByHash(ctx, tokenHash, now)
}

func (uc *Usecase) Me(ctx context.Context, accessToken string) (user.User, error) {
	if strings.TrimSpace(accessToken) == "" {
		return user.User{}, ErrInvalidAccessToken
	}

	tokenUser, err := uc.tokenProvider.ParseAccessToken(accessToken)
	if err != nil {
		return user.User{}, ErrInvalidAccessToken
	}

	u, err := uc.userRepo.GetByEmail(ctx, tokenUser.Email)
	if err != nil {
		return user.User{}, ErrInvalidAccessToken
	}

	return u, nil
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

	return uc.issueTokens(ctx, u)
}

func (uc *Usecase) getOrCreateUserFromOAuth(ctx context.Context, oauthUser OAuthUser) (user.User, TokenPair, bool, error) {
	if found, ok, err := uc.findOAuthUserByProvider(ctx, oauthUser); err != nil {
		return user.User{}, TokenPair{}, false, err
	} else if ok {
		return found, TokenPair{}, false, nil
	}

	if found, ok, err := uc.findOrLinkOAuthUserByEmail(ctx, oauthUser); err != nil {
		return user.User{}, TokenPair{}, false, err
	} else if ok {
		return found, TokenPair{}, false, nil
	}

	created, tokens, err := uc.createOAuthUser(ctx, oauthUser)
	if err != nil {
		return user.User{}, TokenPair{}, false, err
	}

	return created, tokens, true, nil
}

func (uc *Usecase) findOAuthUserByProvider(ctx context.Context, oauthUser OAuthUser) (user.User, bool, error) {
	providerUserID := strings.TrimSpace(oauthUser.ProviderUserID)
	if providerUserID == "" {
		return user.User{}, false, nil
	}

	userID, err := uc.oauthRepo.GetUserIDByProvider(ctx, OAuthProviderGoogle, providerUserID)
	if err != nil {
		return user.User{}, false, nil
	}

	u, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return user.User{}, false, nil
	}

	return u, true, nil
}

func (uc *Usecase) findOrLinkOAuthUserByEmail(ctx context.Context, oauthUser OAuthUser) (user.User, bool, error) {
	u, err := uc.userRepo.GetByEmail(ctx, oauthUser.Email)
	if err == nil {
		if err := uc.linkOAuthProvider(ctx, u.ID, oauthUser); err != nil {
			return user.User{}, false, err
		}

		return u, true, nil
	}

	if !errors.Is(err, user.ErrUserNotFound) {
		return user.User{}, false, err
	}

	return user.User{}, false, nil
}

func (uc *Usecase) createOAuthUser(ctx context.Context, oauthUser OAuthUser) (user.User, TokenPair, error) {
	name, err := uc.resolveOAuthName(oauthUser.Name)
	if err != nil {
		return user.User{}, TokenPair{}, err
	}

	newUser := user.User{
		Name:   name,
		Email:  oauthUser.Email,
		Role:   user.UserRoleUser,
		Status: user.UserStatusActive,
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
		return user.User{}, TokenPair{}, err
	}

	if err := uc.linkOAuthProvider(ctx, created.ID, oauthUser); err != nil {
		return user.User{}, TokenPair{}, err
	}

	info := clientInfoFromContext(ctx)
	if err := uc.refreshRepo.Save(ctx, RefreshToken{
		UserID:    created.ID,
		TokenHash: hashToken(tokens.RefreshToken),
		UserAgent: info.UserAgent,
		IPAddress: info.IPAddress,
		ExpiresAt: uc.now().Add(uc.tokenProviderRefreshTTL()),
	}); err != nil {
		return user.User{}, TokenPair{}, err
	}

	return created, tokens, nil
}

func (uc *Usecase) resolveOAuthName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed != "" {
		return trimmed, nil
	}

	return generateRandomName(16)
}

func (uc *Usecase) linkOAuthProvider(ctx context.Context, userID string, oauthUser OAuthUser) error {
	providerUserID := strings.TrimSpace(oauthUser.ProviderUserID)
	if providerUserID == "" {
		return nil
	}

	return uc.oauthRepo.Upsert(ctx, userID, OAuthProviderGoogle, providerUserID, oauthUser.Email)
}

func (uc *Usecase) recordFailure(ctx context.Context, u user.User, now time.Time) error {
	failedCount := u.FailedLoginAttempts + 1
	lockUntil := u.LockUntil
	lockEscalationLevel := u.LockEscalationLevel
	status := u.Status
	lastFailedLoginAt := now

	if failedCount >= uc.maxRetries {
		failedCount = 0

		// Tính toán thời gian khóa tăng dần theo level
		// Level 0 (lần đầu bị phạt): jailDuration * 2^0 (Ví dụ: 15p * 1 = 15p)
		// Level 1: jailDuration * 2^1 (Ví dụ: 15p * 2 = 30p)
		// Level 2: jailDuration * 2^2 (Ví dụ: 15p * 4 = 1h)
		// ...
		multiplier := time.Duration(1 << lockEscalationLevel)
		t := now.Add(uc.jailDuration * multiplier)

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

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (uc *Usecase) tokenProviderRefreshTTL() time.Duration {
	if provider, ok := uc.tokenProvider.(interface{ GetRefreshTTL() time.Duration }); ok {
		return provider.GetRefreshTTL()
	}

	return 7 * 24 * time.Hour
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

func composeResetPasswordURL(baseURL, token string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", err
	}

	query := u.Query()
	query.Set("token", token)
	u.RawQuery = query.Encode()

	return u.String(), nil
}

func sameClientSession(stored RefreshToken, info ClientInfo) bool {
	normalized := normalizeClientInfo(info)
	if stored.UserAgent != "" && stored.UserAgent != normalized.UserAgent {
		return false
	}

	if stored.IPAddress != "" && stored.IPAddress != normalized.IPAddress {
		return false
	}

	return true
}
