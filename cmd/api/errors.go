package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal server error", "method", r.Method, "url", r.URL.Path, "error", err.Error())
	err = writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem and could not process your request")
	if err != nil {
		app.logger.Errorw("failed to write JSON error", "error", err.Error())
	}
}

func (app *application) badRequest(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("bad request", "method", r.Method, "url", r.URL.Path, "error", err.Error())
	err = writeJSONError(w, http.StatusBadRequest, err.Error())
	if err != nil {
		app.logger.Errorw("failed to write JSON error", "error", err.Error())
	}
}

func (app *application) notFound(w http.ResponseWriter, r *http.Request) {
	app.logger.Warnw("not found", "method", r.Method, "url", r.URL.Path)
	err := writeJSONError(w, http.StatusNotFound, "not found")
	if err != nil {
		app.logger.Errorw("failed to write JSON error", "error", err.Error())
	}
}

func (app *application) unauthorized(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("unauthorized", "method", r.Method, "url", r.URL.Path, "error", err.Error())
	err = writeJSONError(w, http.StatusUnauthorized, "unauthorized")
	if err != nil {
		app.logger.Errorw("failed to write JSON error", "error", err.Error())
	}
}

func (app *application) forbidden(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("forbidden", "method", r.Method, "url", r.URL.Path, "error", err.Error())
	err = writeJSONError(w, http.StatusForbidden, "forbidden")
	if err != nil {
		app.logger.Errorw("failed to write JSON error", "error", err.Error())
	}
}

func (app *application) unauthorizedBasic(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("unauthorized", "method", r.Method, "url", r.URL.Path, "error", err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted", charset="UTF-8"`)

	err = writeJSONError(w, http.StatusUnauthorized, "unauthorized")
	if err != nil {
		app.logger.Errorw("failed to write JSON error", "error", err.Error())
	}
}

func (app *application) tooManyRequests(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("too many requests", "method", r.Method, "url", r.URL.Path, "error", err.Error())
	err = writeJSONError(w, http.StatusTooManyRequests, "too many requests")
	if err != nil {
		app.logger.Errorw("failed to write JSON error", "error", err.Error())
	}
}
