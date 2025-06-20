package main

import (
	"fmt"
	"hexagon/adapters/httpserver"
	"hexagon/adapters/postgrestore"
	"hexagon/pkg/config"
	"hexagon/pkg/sentry"
	"log/slog"
	"net/http"
	"os"

	sentrygo "github.com/getsentry/sentry-go"
	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = sentrygo.Init(sentrygo.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.AppEnv,
		AttachStacktrace: true,
	})
	if err != nil {
		log.Fatalf("cannot init sentry: %v", err)
	}
	defer sentrygo.Flush(sentry.FlushTime)

	db, err := postgrestore.NewConnection(postgrestore.ParseFromConfig(cfg))
	if err != nil {
		log.Fatal(err)
	}

	//db, err := inmemstore.NewConnection()

	server, err := httpserver.New(httpserver.WithConfig(cfg))
	if err != nil {
		log.Fatal(err)
	}

	server.BookStore = postgrestore.NewBookStore(db)
	//server.BookStore = inmemstore.NewBookStore(db)

	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("server started!")
	log.Fatal(http.ListenAndServe(addr, server))
}
