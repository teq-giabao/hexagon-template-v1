package main

import (
	"fmt"
	"hexagon/httpserver"
	"hexagon/pkg/config"
	"hexagon/pkg/sentry"
	"hexagon/postgres"
	"log/slog"
	"os"

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

	_, err = postgres.NewConnection(postgres.Options{
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

	server := httpserver.Default()
	server.Addr = fmt.Sprintf(":%d", cfg.Port)

	slog.Info("server started!")
	if err := server.Start(); err != nil {
		slog.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}
