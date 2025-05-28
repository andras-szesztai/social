package main

import (
	"github.com/andras-szesztai/social/internal/db"
	"github.com/andras-szesztai/social/internal/env"
	"github.com/andras-szesztai/social/internal/store"
	_ "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			Social API
//	@description	API for the Social application

//	@BasePath					/v1
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
//	@scheme						bearer
//	@type						http
//	@name						Authorization

func main() {
	cfg := config{
		addr:   env.GetString("ADDR", ":8080"),
		env:    env.GetString("ENV", "development"),
		apiURL: env.GetString("API_URL", "localhost:8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 25),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 25),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
	}

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	db, err := db.NewDB(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	store := store.NewStore(db)

	app := application{
		config: cfg,
		store:  store,
		logger: logger,
	}

	err = app.serve(app.mountRoutes())
	if err != nil {
		logger.Fatal(err)
	}

}
