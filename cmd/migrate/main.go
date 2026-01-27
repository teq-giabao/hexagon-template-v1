package main

import (
	"hexagon/pkg/config"
	"hexagon/postgres"
	"log/slog"
	"os"
	"strconv"

	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("cannot load config", "error", err)
		os.Exit(1)
	}

	db, err := postgres.NewConnection(postgres.Options{
		DBName:   cfg.DB.Name,
		DBUser:   cfg.DB.User,
		Password: cfg.DB.Pass,
		Host:     cfg.DB.Host,
		Port:     strconv.Itoa(cfg.DB.Port),
		SSLMode:  false,
	})
	if err != nil {
		logger.Error("cannot connecting to db", "error", err)
		os.Exit(1)
	}

	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("cannot get db instance", "error", err)
		os.Exit(1)
	}

	total, err := migrate.Exec(sqlDB, "postgres", migrations, migrate.Up)
	if err != nil {
		logger.Error("cannot execute migration", "error", err)
		os.Exit(1)
	}

	logger.Info("applied migrations", "total", total)
}
