package main

import (
	"log"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error: %s %s %s", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem and could not process your request")
}

func (app *application) badRequest(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request: %s %s %s", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) notFound(w http.ResponseWriter, r *http.Request) {
	log.Printf("not found: %s %s", r.Method, r.URL.Path)
	writeJSONError(w, http.StatusNotFound, "not found")
}
