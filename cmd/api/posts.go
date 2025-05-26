package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/andras-szesztai/social/internal/store"
	"github.com/go-chi/chi/v5"
)

type createPostRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload createPostRequest
	if err := readJSON(w, r, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
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
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, *createdPost)
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	post, err := app.store.Posts.Get(r.Context(), intID)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSONError(w, http.StatusNotFound, "post not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, post)
}
