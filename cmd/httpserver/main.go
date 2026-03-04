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
	"context"
	"fmt"
	"hexagon/auth"
	"hexagon/hotel"
	"hexagon/httpserver"
	"hexagon/pkg/config"
	"hexagon/pkg/hashing"
	"hexagon/pkg/jwt"
	resendmailer "hexagon/pkg/mailer/resend"
	oauthgoogle "hexagon/pkg/oauth/google"
	"hexagon/pkg/sentry"
	s3storage "hexagon/pkg/storage/s3"
	"hexagon/postgres"
	"hexagon/upload"
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
	hotelRepo := postgres.NewHotelRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)
	userService := user.NewUsecaseWithSession(
		userRepo,
		hashing.NewBcryptHasher(),
		refreshTokenRepo,
	)
	hotelService := hotel.NewUsecase(hotelRepo)
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
	server.HotelService = hotelService
	server.UploadService = createUploadService(cfg)
	server.Addr = fmt.Sprintf(":%d", cfg.Port)

	slog.Info("server started!")
	if err := server.Start(); err != nil {
		slog.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}

func createUploadService(cfg *config.Config) upload.Service {
	return upload.NewUsecase(createImageUploader(cfg))
}

func createImageUploader(cfg *config.Config) upload.Uploader {
	if cfg == nil || cfg.Storage.S3Bucket == "" {
		slog.Warn("s3 uploader is disabled because S3_BUCKET is empty")
		return nil
	}

	uploader, err := s3storage.NewUploader(context.Background(), s3storage.Config{
		Region:          cfg.Storage.S3Region,
		Bucket:          cfg.Storage.S3Bucket,
		BaseURL:         cfg.Storage.S3BaseURL,
		Prefix:          cfg.Storage.S3Prefix,
		Endpoint:        cfg.Storage.S3Endpoint,
		AccessKeyID:     cfg.Storage.S3AccessKeyID,
		SecretAccessKey: cfg.Storage.S3SecretAccessKey,
		SessionToken:    cfg.Storage.S3SessionToken,
	})
	if err != nil {
		slog.Error("cannot initialize s3 uploader", "error", err)
		return nil
	}
	return uploader
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
