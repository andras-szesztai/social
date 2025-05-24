package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type application struct {
	config config
}

type config struct {
	addr string
}

func (app *application) mountRoutes() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.StripSlashes)

	router.Route("/v1", func(r chi.Router) {
		r.Get("/healthcheck", app.healthCheckHandler)
	})

	return router
}

func (app *application) serve(router http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      router,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	log.Printf("starting server on %s", srv.Addr)

	return srv.ListenAndServe()
}
