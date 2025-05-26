package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/andras-szesztai/social/internal/store"
	"github.com/go-chi/chi/v5"
)

type createPostRequest struct {
	Title   string   `json:"title" validate:"required,max=255"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags" validate:"required,max=10"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload createPostRequest
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	if err := Validator.Struct(payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	post := store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  1,
	}

	ctx := r.Context()
	createdPost, err := app.store.Posts.Create(ctx, &post)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, *createdPost)
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	post, err := app.store.Posts.Get(r.Context(), intID)
	if err != nil {
		if err == sql.ErrNoRows {
			app.notFound(w, r)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, post)
}
