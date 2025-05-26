package main

import (
	"net/http"
	"strconv"

	"github.com/andras-szesztai/social/internal/store"
	"github.com/go-chi/chi/v5"
)

type createCommentRequest struct {
	Content string `json:"content" validate:"required,max=1000"`
}

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "postID")
	intPostID, err := strconv.ParseInt(postID, 10, 64)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	userID := chi.URLParam(r, "userID")
	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	var payload createCommentRequest
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	if err := Validator.Struct(payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	comment := store.Comment{
		PostID:  intPostID,
		UserID:  intUserID,
		Content: payload.Content,
	}

	ctx := r.Context()
	createdComment, err := app.store.Comments.Create(ctx, &comment)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, *createdComment)
}

func (app *application) getCommentsByPostIDHandler(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	intPostID, err := strconv.ParseInt(postID, 10, 64)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	ctx := r.Context()
	comments, err := app.store.Comments.GetByPostID(ctx, intPostID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, comments)
}
