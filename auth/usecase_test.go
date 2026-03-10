package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"hexagon/auth"
	"hexagon/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserRepo struct {
	getByEmailFn          func(ctx context.Context, email string) (user.User, error)
	getByIDFn             func(ctx context.Context, id string) (user.User, error)
	createUserTxFn        func(ctx context.Context, u user.User, fn func(created user.User) error) (user.User, error)
	updatePasswordHashFn  func(ctx context.Context, id, passwordHash string) error
	updateEmailVerifiedFn func(ctx context.Context, id string, verifiedAt *time.Time) error
	updateAuthStateFn     func(ctx context.Context, id string, failedLoginAttempts int, lockUntil *time.Time, lockEscalationLevel int, lastFailedLoginAt *time.Time, status user.UserStatus) error
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (user.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}

	return user.User{}, user.ErrUserNotFound
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (user.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}

	return user.User{}, user.ErrUserNotFound
}

func (m *mockUserRepo) CreateUser(ctx context.Context, u user.User) error {
	return nil
}

func (m *mockUserRepo) CreateUserTx(ctx context.Context, u user.User, fn func(created user.User) error) (user.User, error) {
	if m.createUserTxFn != nil {
		return m.createUserTxFn(ctx, u, fn)
	}

	return user.User{}, errors.New("not implemented")
}

func (m *mockUserRepo) UpdatePasswordHash(ctx context.Context, id, passwordHash string) error {
	if m.updatePasswordHashFn != nil {
		return m.updatePasswordHashFn(ctx, id, passwordHash)
	}

	return nil
}

func (m *mockUserRepo) UpdateEmailVerifiedAt(ctx context.Context, id string, verifiedAt *time.Time) error {
	if m.updateEmailVerifiedFn != nil {
		return m.updateEmailVerifiedFn(ctx, id, verifiedAt)
	}

	return nil
}

func (m *mockUserRepo) UpdateAuthState(
	ctx context.Context,
	id string,
	failedLoginAttempts int,
	lockUntil *time.Time,
	lockEscalationLevel int,
	lastFailedLoginAt *time.Time,
	status user.UserStatus,
) error {
	if m.updateAuthStateFn != nil {
		return m.updateAuthStateFn(ctx, id, failedLoginAttempts, lockUntil, lockEscalationLevel, lastFailedLoginAt, status)
	}

	return nil
}

type mockRefreshRepo struct {
	saveFn              func(ctx context.Context, token auth.RefreshToken) error
	getActiveByHashFn   func(ctx context.Context, tokenHash string) (auth.RefreshToken, error)
	revokeByHashFn      func(ctx context.Context, tokenHash string, revokedAt time.Time) error
	revokeAllByUserIDFn func(ctx context.Context, userID string, revokedAt time.Time) error
}

func (m *mockRefreshRepo) Save(ctx context.Context, token auth.RefreshToken) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, token)
	}

	return nil
}

func (m *mockRefreshRepo) GetActiveByHash(ctx context.Context, tokenHash string) (auth.RefreshToken, error) {
	if m.getActiveByHashFn != nil {
		return m.getActiveByHashFn(ctx, tokenHash)
	}

	return auth.RefreshToken{}, errors.New("not found")
}

func (m *mockRefreshRepo) RevokeByHash(ctx context.Context, tokenHash string, revokedAt time.Time) error {
	if m.revokeByHashFn != nil {
		return m.revokeByHashFn(ctx, tokenHash, revokedAt)
	}

	return nil
}

func (m *mockRefreshRepo) RevokeAllByUserID(ctx context.Context, userID string, revokedAt time.Time) error {
	if m.revokeAllByUserIDFn != nil {
		return m.revokeAllByUserIDFn(ctx, userID, revokedAt)
	}

	return nil
}

type mockResetRepo struct {
	getActiveByHashFn func(ctx context.Context, tokenHash string) (auth.PasswordResetToken, error)
	markUsedFn        func(ctx context.Context, tokenHash string, usedAt time.Time) error
}

func (m *mockResetRepo) Save(ctx context.Context, token auth.PasswordResetToken) error {
	return nil
}

func (m *mockResetRepo) GetActiveByHash(ctx context.Context, tokenHash string) (auth.PasswordResetToken, error) {
	if m.getActiveByHashFn != nil {
		return m.getActiveByHashFn(ctx, tokenHash)
	}

	return auth.PasswordResetToken{}, errors.New("not found")
}

func (m *mockResetRepo) MarkUsedByHash(ctx context.Context, tokenHash string, usedAt time.Time) error {
	if m.markUsedFn != nil {
		return m.markUsedFn(ctx, tokenHash, usedAt)
	}

	return nil
}

type mockVerifyRepo struct {
	getActiveByHashFn func(ctx context.Context, tokenHash string) (auth.EmailVerificationToken, error)
	markUsedFn        func(ctx context.Context, tokenHash string, usedAt time.Time) error
}

func (m *mockVerifyRepo) Save(ctx context.Context, token auth.EmailVerificationToken) error {
	return nil
}

func (m *mockVerifyRepo) GetActiveByHash(ctx context.Context, tokenHash string) (auth.EmailVerificationToken, error) {
	if m.getActiveByHashFn != nil {
		return m.getActiveByHashFn(ctx, tokenHash)
	}

	return auth.EmailVerificationToken{}, errors.New("not found")
}

func (m *mockVerifyRepo) MarkUsedByHash(ctx context.Context, tokenHash string, usedAt time.Time) error {
	if m.markUsedFn != nil {
		return m.markUsedFn(ctx, tokenHash, usedAt)
	}

	return nil
}

type mockOAuthRepo struct{}

func (m *mockOAuthRepo) GetUserIDByProvider(ctx context.Context, provider auth.OAuthProvider, providerUserID string) (string, error) {
	return "", errors.New("not found")
}

func (m *mockOAuthRepo) Upsert(ctx context.Context, userID string, provider auth.OAuthProvider, providerUserID, providerEmail string) error {
	return nil
}

type mockHasher struct {
	hashFn    func(password string) (string, error)
	compareFn func(hashed, plain string) error
}

func (m *mockHasher) Compare(hashed, plain string) error {
	if m.compareFn != nil {
		return m.compareFn(hashed, plain)
	}

	return nil
}

func (m *mockHasher) Hash(password string) (string, error) {
	if m.hashFn != nil {
		return m.hashFn(password)
	}

	return "hashed", nil
}

type mockTokenProvider struct {
	parseRefreshFn    func(refreshToken string) (user.User, error)
	generateAccessFn  func(u user.User) (string, error)
	generateRefreshFn func(u user.User) (string, error)
}

func (m *mockTokenProvider) GenerateAccessToken(u user.User) (string, error) {
	if m.generateAccessFn != nil {
		return m.generateAccessFn(u)
	}

	return "access-token", nil
}

func (m *mockTokenProvider) GenerateRefreshToken(u user.User) (string, error) {
	if m.generateRefreshFn != nil {
		return m.generateRefreshFn(u)
	}

	return "refresh-token-next", nil
}

func (m *mockTokenProvider) ParseAccessToken(accessToken string) (user.User, error) {
	return user.User{}, errors.New("not implemented")
}

func (m *mockTokenProvider) ParseRefreshToken(refreshToken string) (user.User, error) {
	if m.parseRefreshFn != nil {
		return m.parseRefreshFn(refreshToken)
	}

	return user.User{}, errors.New("invalid")
}

func newUsecaseForTest(u *mockUserRepo, r *mockRefreshRepo, p *mockResetRepo, h *mockHasher, t *mockTokenProvider) *auth.Usecase {
	return auth.NewUsecase(u, &mockOAuthRepo{}, r, p, &mockVerifyRepo{}, h, t, nil, nil, "", "")
}

func TestRegister_ReturnsConflictWhenEmailBelongsToOAuthOnly(t *testing.T) {
	repo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (user.User, error) {
			return user.User{ID: "u1", Email: email, PasswordHash: ""}, nil
		},
	}
	uc := newUsecaseForTest(repo, &mockRefreshRepo{}, &mockResetRepo{}, &mockHasher{}, &mockTokenProvider{})

	err := uc.Register(context.Background(), "John", "john@example.com", "0123456789", "Password@123")

	require.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrEmailRegisteredWithOAuth)
}

func TestLogin_ReturnsPasswordAuthNotAvailableForOAuthOnly(t *testing.T) {
	repo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (user.User, error) {
			return user.User{ID: "u1", Email: email, PasswordHash: ""}, nil
		},
	}
	uc := newUsecaseForTest(repo, &mockRefreshRepo{}, &mockResetRepo{}, &mockHasher{}, &mockTokenProvider{})

	_, err := uc.Login(context.Background(), "oauth@example.com", "Password@123")

	require.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrPasswordAuthNotAvailable)
}

func TestLogin_ReturnsEmailNotVerifiedWhenEmailNotVerified(t *testing.T) {
	repo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (user.User, error) {
			return user.User{
				ID:           "u1",
				Email:        email,
				PasswordHash: "hashed-password",
				Status:       user.UserStatusActive,
			}, nil
		},
	}

	hasher := &mockHasher{
		compareFn: func(hashed, plain string) error {
			return nil
		},
	}

	uc := newUsecaseForTest(repo, &mockRefreshRepo{}, &mockResetRepo{}, hasher, &mockTokenProvider{})

	_, err := uc.Login(context.Background(), "unverified@example.com", "Password@123")

	require.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrEmailNotVerified)
}

func TestRefresh_FailsWhenSessionClientInfoMismatches(t *testing.T) {
	now := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)
	uc := auth.NewUsecase(&mockUserRepo{}, &mockOAuthRepo{}, &mockRefreshRepo{
		getActiveByHashFn: func(ctx context.Context, tokenHash string) (auth.RefreshToken, error) {
			return auth.RefreshToken{
				UserID:    "u1",
				TokenHash: tokenHash,
				UserAgent: "agent-a",
				IPAddress: "1.1.1.1",
				ExpiresAt: now.Add(5 * time.Minute),
			}, nil
		},
	}, &mockResetRepo{}, &mockVerifyRepo{}, &mockHasher{}, &mockTokenProvider{
		parseRefreshFn: func(refreshToken string) (user.User, error) {
			return user.User{ID: "u1", Email: "u1@example.com", Role: user.UserRoleUser}, nil
		},
	}, nil, nil, "", "")
	uc.NowForTest(now)

	ctx := auth.WithClientInfo(context.Background(), auth.ClientInfo{
		UserAgent: "agent-b",
		IPAddress: "1.1.1.1",
	})
	_, err := uc.Refresh(ctx, "refresh-token")

	require.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrInvalidRefreshToken)
}

func TestResetPassword_RevokesAllSessionsOnSuccess(t *testing.T) {
	now := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)

	var revokeAllCalled bool

	var markUsedCalled bool

	repo := &mockUserRepo{
		getByIDFn: func(ctx context.Context, id string) (user.User, error) {
			return user.User{ID: id, Email: "john@example.com", PasswordHash: "old-hash"}, nil
		},
		updatePasswordHashFn: func(ctx context.Context, id, passwordHash string) error {
			assert.Equal(t, "u1", id)
			assert.Equal(t, "new-hash", passwordHash)

			return nil
		},
	}
	refreshRepo := &mockRefreshRepo{
		revokeAllByUserIDFn: func(ctx context.Context, userID string, revokedAt time.Time) error {
			revokeAllCalled = true

			assert.Equal(t, "u1", userID)
			assert.Equal(t, now, revokedAt)

			return nil
		},
	}
	resetRepo := &mockResetRepo{
		getActiveByHashFn: func(ctx context.Context, tokenHash string) (auth.PasswordResetToken, error) {
			return auth.PasswordResetToken{
				UserID:    "u1",
				TokenHash: tokenHash,
				ExpiresAt: now.Add(10 * time.Minute),
			}, nil
		},
		markUsedFn: func(ctx context.Context, tokenHash string, usedAt time.Time) error {
			markUsedCalled = true

			assert.Equal(t, now, usedAt)

			return nil
		},
	}
	hasher := &mockHasher{
		hashFn: func(password string) (string, error) {
			assert.Equal(t, "NewPassword@123", password)
			return "new-hash", nil
		},
	}
	uc := auth.NewUsecase(repo, &mockOAuthRepo{}, refreshRepo, resetRepo, &mockVerifyRepo{}, hasher, &mockTokenProvider{}, nil, nil, "", "")
	uc.NowForTest(now)

	err := uc.ResetPassword(context.Background(), "raw-reset-token", "NewPassword@123")

	require.NoError(t, err)
	assert.True(t, revokeAllCalled)
	assert.True(t, markUsedCalled)
}
