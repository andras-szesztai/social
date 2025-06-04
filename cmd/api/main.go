package main

import (
	"time"

	"github.com/andras-szesztai/social/internal/auth"
	"github.com/andras-szesztai/social/internal/db"
	"github.com/andras-szesztai/social/internal/env"
	"github.com/andras-szesztai/social/internal/mailer"
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
		addr:        env.GetString("ADDR", ":8080"),
		env:         env.GetString("ENV", "development"),
		apiURL:      env.GetString("API_URL", "localhost:8080"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:3000"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 25),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 25),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		mail: mailConfig{
			expiry: 24 * time.Hour,
			apiKey: env.GetString("MAIL_API_KEY", ""),
			from:   env.GetString("MAIL_FROM", ""),
		},
		auth: authConfig{
			basic: basicAuthConfig{
				username: env.GetString("BASIC_AUTH_USERNAME", "admin"),
				password: env.GetString("BASIC_AUTH_PASSWORD", "admin"),
			},
			token: tokenConfig{
				secret: env.GetString("TOKEN_SECRET", ""),
				exp:    env.GetDuration("TOKEN_EXP", 3*24*time.Hour),
				aud:    env.GetString("TOKEN_AUD", ""),
				iss:    env.GetString("TOKEN_ISS", ""),
			},
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

	mailer := mailer.NewSendGridMailer(cfg.mail.from, cfg.mail.apiKey)

	authenticator := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.aud, cfg.auth.token.iss)

	app := application{
		config:        cfg,
		store:         store,
		logger:        logger,
		mailer:        mailer,
		authenticator: authenticator,
	}

	err = app.serve(app.mountRoutes())
	if err != nil {
		logger.Fatal(err)
	}

}
