package main

import (
	"fmt"
	sentrygo "github.com/getsentry/sentry-go"
	_ "github.com/lib/pq"
	"hexagon/adapters/httpserver"
	"hexagon/adapters/postgrestore"
	"hexagon/pkg/config"
	"hexagon/pkg/logger"
	"hexagon/pkg/sentry"
	"log"
	"net/http"
)

func main() {
	applog, err := logger.NewAppLogger()
	if err != nil {
		log.Fatalf("cannot load config: %v\n", err)
	}
	defer logger.Sync(applog)

	cfg, err := config.LoadConfig()
	if err != nil {
		applog.Fatal(err)
	}

	err = sentrygo.Init(sentrygo.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.AppEnv,
		AttachStacktrace: true,
	})
	if err != nil {
		applog.Fatalf("cannot init sentry: %v", err)
	}
	defer sentrygo.Flush(sentry.FlushTime)

	db, err := postgrestore.NewConnection(postgrestore.ParseFromConfig(cfg))
	if err != nil {
		applog.Fatal(err)
	}

	//db, err := inmemstore.NewConnection()

	server, err := httpserver.New()
	if err != nil {
		applog.Fatal(err)
	}

	server.Logger = applog
	server.Config = cfg
	server.BookStore = postgrestore.NewBookStore(db)
	//server.BookStore = inmemstore.NewBookStore(db)

	addr := fmt.Sprintf(":%d", cfg.Port)
	applog.Info("server started!")
	applog.Fatal(http.ListenAndServe(addr, server))
}
