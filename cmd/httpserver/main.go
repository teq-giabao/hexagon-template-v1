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
	"hexagon/contact"
	"hexagon/dynamodb"
	"hexagon/grpcserver"
	"hexagon/httpserver"
	"hexagon/pkg/config"
	"hexagon/pkg/hashing"
	"hexagon/pkg/jwt"
	oauthgoogle "hexagon/pkg/oauth/google"
	"hexagon/pkg/sentry"
	"hexagon/postgres"
	"hexagon/user"
	"log/slog"
	"os"
	"strings"
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

	storageDriver := strings.ToLower(strings.TrimSpace(cfg.DB.Driver))
	if storageDriver == "" {
		storageDriver = "postgres"
	}

	var (
		contactService contact.Service
		userService    user.Service
		authService    auth.Service
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

	switch storageDriver {
	case "dynamodb":
		client, err := dynamodb.NewClient(context.Background(), dynamodb.Options{
			Region:       cfg.DynamoDB.Region,
			Endpoint:     cfg.DynamoDB.Endpoint,
			AccessKey:    cfg.DynamoDB.AccessKey,
			SecretKey:    cfg.DynamoDB.SecretKey,
			SessionToken: cfg.DynamoDB.SessionToken,
		})
		if err != nil {
			slog.Error("Cannot init dynamodb client", "error", err)
			os.Exit(1)
		}

		userRepo := dynamodb.NewUserRepository(client, cfg.DynamoDB.UsersTable)
		attemptRepo := dynamodb.NewLoginAttemptRepository(client, cfg.DynamoDB.LoginAttemptsTable)

		contactService = contact.NewUsecase(
			dynamodb.NewContactRepository(client, cfg.DynamoDB.ContactsTable),
		)
		userService = user.NewUsecase(
			userRepo,
			hashing.NewBcryptHasher(),
		)
		authService = auth.NewUsecase(
			userRepo,
			attemptRepo,
			hashing.NewBcryptHasher(),
			jwt.NewJWTProvider(
				cfg.Auth.JWTSecret,
				time.Duration(cfg.Auth.TokenTTL)*time.Second,
				time.Duration(cfg.Auth.RefreshTTL)*time.Second,
			),
			googleProvider,
		)
	default:
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
		attemptRepo := postgres.NewLoginAttemptRepository(db)

		contactService = contact.NewUsecase(postgres.NewContactRepository(db))
		userService = user.NewUsecase(
			userRepo,
			hashing.NewBcryptHasher(),
		)
		authService = auth.NewUsecase(
			userRepo,
			attemptRepo,
			hashing.NewBcryptHasher(),
			jwt.NewJWTProvider(
				cfg.Auth.JWTSecret,
				time.Duration(cfg.Auth.TokenTTL)*time.Second,
				time.Duration(cfg.Auth.RefreshTTL)*time.Second,
			),
			googleProvider,
		)
	}
	server := httpserver.Default(cfg)
	server.JWTSecret = cfg.Auth.JWTSecret
	server.ContactService = contactService
	server.UserService = userService
	server.AuthService = authService
	server.Addr = fmt.Sprintf(":%d", cfg.Port)

	grpcPort := cfg.GRPCPort
	if grpcPort == 0 {
		grpcPort = cfg.Port + 1
	}
	grpcAddr := fmt.Sprintf(":%d", grpcPort)
	grpcServer := grpcserver.New(grpcAddr)
	go func() {
		slog.Info("grpc server started", "addr", grpcAddr)
		if err := grpcServer.Start(); err != nil {
			slog.Error("grpc server stopped with error", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("server started!")
	if err := server.Start(); err != nil {
		slog.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}
