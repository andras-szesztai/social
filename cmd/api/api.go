package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/andras-szesztai/social/docs"
	"github.com/andras-szesztai/social/internal/auth"
	"github.com/andras-szesztai/social/internal/mailer"
	"github.com/andras-szesztai/social/internal/store"
	"github.com/andras-szesztai/social/internal/store/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type application struct {
	config        config
	store         *store.Store
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	cache         *cache.Storage
	authenticator auth.Authenticator
}

type config struct {
	addr        string
	env         string
	db          dbConfig
	apiURL      string
	mail        mailConfig
	frontendURL string
	auth        authConfig
	redis       redisConfig
}

type mailConfig struct {
	expiry time.Duration
	apiKey string
	from   string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type authConfig struct {
	basic basicAuthConfig
	token tokenConfig
}

type redisConfig struct {
	addr     string
	password string
	db       int
	enabled  bool
}

type basicAuthConfig struct {
	username string
	password string
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	aud    string
	iss    string
}

type contextKey string

func (app *application) mountRoutes() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.StripSlashes)

	router.Route("/v1", func(r chi.Router) {
		// r.Use(app.BasicAuthMiddleware)
		r.Get("/healthcheck", app.healthCheckHandler)

		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Post("/", app.createPostHandler)

			r.Route("/{id}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)
				r.Get("/", app.getPostHandler)
				r.Patch("/", app.checkPostOwnership("moderator", app.updatePostHandler))
				r.Delete("/", app.checkPostOwnership("admin", app.deletePostHandler))

				r.Route("/comments", func(r chi.Router) {
					r.Get("/", app.getCommentsByPostIDHandler)
					r.Post("/", app.createCommentHandler)
				})
			})
		})

		r.Route("/comments", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Route("/{id}", func(r chi.Router) {
				r.Use(app.commentsContextMiddleware)
				r.Get("/", app.getCommentHandler)
				r.Patch("/", app.updateCommentHandler)
				r.Delete("/", app.deleteCommentHandler)
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Get("/feed", app.getUserFeedHandler)
			})

			r.Route("/{id}", func(r chi.Router) {
				r.Use(app.usersContextMiddleware)
				r.Put("/activate/{token}", app.activateUserHandler)

				r.Group(func(r chi.Router) {
					r.Use(app.AuthTokenMiddleware)
					r.Get("/", app.getUserHandler)
					r.Post("/follow", app.followUserHandler)
					r.Post("/unfollow", app.unfollowUserHandler)
					r.Delete("/", app.deleteUserHandler)
				})
			})
		})

		r.Route("/authentication", func(r chi.Router) {
			r.Post("/register", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
		})
	})

	return router
}

func (app *application) serve(router http.Handler) error {
	// Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      router,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	app.logger.Infow("starting server", "addr", srv.Addr, "env", app.config.env, "version", version)

	return srv.ListenAndServe()
}
