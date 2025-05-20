package main

import (
	"fmt"
	"hexagon/adapters/httpserver"
	"hexagon/adapters/postgrestore"
	"hexagon/pkg/config"
	"hexagon/pkg/logging"
	"hexagon/pkg/sentry"
	"log"
	"net/http"

	sentrygo "github.com/getsentry/sentry-go"
	_ "github.com/lib/pq"
)

func main() {
	logger, err := logging.NewLogger()
	if err != nil {
		log.Fatalf("cannot load config: %v\n", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal(err)
	}

	err = sentrygo.Init(sentrygo.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.AppEnv,
		AttachStacktrace: true,
	})
	if err != nil {
		logger.Fatalf("cannot init sentry: %v", err)
	}
	defer sentrygo.Flush(sentry.FlushTime)

	db, err := postgrestore.NewConnection(postgrestore.ParseFromConfig(cfg))
	if err != nil {
		logger.Fatal(err)
	}

	//db, err := inmemstore.NewConnection()

	server, err := httpserver.New(httpserver.WithConfig(cfg))
	if err != nil {
		logger.Fatal(err)
	}

	server.Logger = logger
	server.BookStore = postgrestore.NewBookStore(db)
	//server.BookStore = inmemstore.NewBookStore(db)

	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("server started!")
	logger.Fatal(http.ListenAndServe(addr, server))
}
