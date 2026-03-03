// @title Hexagon API
// @version 1.0
// @description API Documentation for Hexagon project.
// @host localhost:8088
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
// nolint: funlen
package main

import (
	"fmt"
	"hexagon/auth"
	"hexagon/httpserver"
	"hexagon/pkg/config"
	"hexagon/pkg/hashing"
	"hexagon/pkg/jwt"
	resendmailer "hexagon/pkg/mailer/resend"
	oauthgoogle "hexagon/pkg/oauth/google"
	"hexagon/pkg/sentry"
	"hexagon/postgres"
	"hexagon/user"
	"log/slog"
	"os"
	"time"

	_ "hexagon/docs"

	sentrygo "github.com/getsentry/sentry-go"
	_ "github.com/lib/pq"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Cannot load config", "error", err)
		os.Exit(1)
	}

	err = sentrygo.Init(sentrygo.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.AppEnv,
		AttachStacktrace: true,
	})
	if err != nil {
		slog.Error("Cannot init sentry", "error", err)
		os.Exit(1)
	}
	defer sentrygo.Flush(sentry.FlushTime)

	db, err := postgres.NewConnection(postgres.Options{
		DBName:   cfg.DB.Name,
		DBUser:   cfg.DB.User,
		Password: cfg.DB.Pass,
		Host:     cfg.DB.Host,
		Port:     fmt.Sprintf("%d", cfg.DB.Port),
		SSLMode:  false,
	})
	if err != nil {
		slog.Error("Cannot open postgres connection", "error", err)
		os.Exit(1)
	}

	userRepo := postgres.NewUserRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)
	userService := user.NewUsecaseWithSession(
		userRepo,
		hashing.NewBcryptHasher(),
		refreshTokenRepo,
	)
	googleProvider, err := oauthgoogle.NewProvider(
		cfg.Auth.GoogleClientID,
		cfg.Auth.GoogleClientSecret,
		cfg.Auth.GoogleRedirectURL,
	)
	if err != nil {
		slog.Error("Cannot init google oauth provider", "error", err)
		os.Exit(1)
	}
	authService := auth.NewUsecase(
		userRepo,
		postgres.NewOAuthProviderAccountRepository(db),
		refreshTokenRepo,
		postgres.NewPasswordResetTokenRepository(db),
		hashing.NewBcryptHasher(),
		jwt.NewJWTProvider(
			cfg.Auth.JWTSecret,
			time.Duration(cfg.Auth.TokenTTL)*time.Second,
			time.Duration(cfg.Auth.RefreshTTL)*time.Second,
		),
		googleProvider,
		createResetMailer(cfg),
		cfg.Auth.ResetPasswordURL,
	)
	server := httpserver.Default(cfg)
	server.JWTSecret = cfg.Auth.JWTSecret
	server.UserService = userService
	server.AuthService = authService
	server.Addr = fmt.Sprintf(":%d", cfg.Port)

	slog.Info("server started!")
	if err := server.Start(); err != nil {
		slog.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}

func createResetMailer(cfg *config.Config) auth.PasswordResetMailer {
	if cfg == nil {
		return nil
	}
	mailer, err := resendmailer.NewProvider(
		cfg.Auth.ResendAPIKey,
		cfg.Auth.ResendFromEmail,
		cfg.Auth.ResendFromName,
	)
	if err != nil {
		slog.Warn("password reset mailer is not configured", "error", err)
		return nil
	}
	return mailer
}
