package main

import (
	"encoding/json"
	"net/http"

	"github.com/andras-szesztai/social/internal/store"
	"github.com/go-playground/validator/v10"
)

var Validator *validator.Validate

func init() {
	Validator = validator.New(validator.WithRequiredStructEnabled())
}

func writeJSONResponse(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	return dec.Decode(data)
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, message string) error {
	return writeJSONResponse(w, status, &errorResponse{Error: message})
}

func (app *application) jsonResponse(w http.ResponseWriter, status int, data any) error {
	return writeJSONResponse(w, status, data)
}

type userResponse struct {
	Data store.User `json:"data"`
}

type userFeedResponse struct {
	Data []store.UserFeed `json:"data"`
}

type postResponse struct {
	Data store.Post `json:"data"`
}

type commentResponse struct {
	Data store.Comment `json:"data"`
}

type commentsResponse struct {
	Data []store.Comment `json:"data"`
}
