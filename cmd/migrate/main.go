package main

import (
	"hexagon/adapters/postgrestore"
	"hexagon/pkg/config"
	"hexagon/pkg/logging"
	"log"
	"strconv"

	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

func main() {
	logger, err := logging.NewLogger()
	if err != nil {
		log.Fatalf("cannot load config: %v\n", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("cannot load config: %v\n", err)
	}

	db, err := postgrestore.NewConnection(postgrestore.Options{
		DBName:   cfg.DB.Name,
		DBUser:   cfg.DB.User,
		Password: cfg.DB.Pass,
		Host:     cfg.DB.Host,
		Port:     strconv.Itoa(cfg.DB.Port),
		SSLMode:  false,
	})
	if err != nil {
		logger.Fatalf("cannot connecting to db: %v\n", err)
	}

	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}

	total, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		logger.Fatalf("cannot execute migration: %v\n", err)
	}

	logger.Infof("applied %d migrations\n", total)
}
