package main

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/andras-szesztai/social/internal/store"
	"github.com/go-chi/chi/v5"
)

type createCommentRequest struct {
	Content string `json:"content" validate:"required,max=1000"`
}

// CreateComment godoc
//
//	@Summary		Create comment
//	@Description	Create a new comment
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			request	body		createCommentRequest	true	"Create comment request"
//	@Success		201		{object}	commentResponse
//	@Failure		400		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Security		ApiKeyAuth
//	@Router			/posts/{id}/comments [post]
func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	post := app.getPostContext(r)

	var payload createCommentRequest
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	if err := Validator.Struct(payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	user := app.getUserContext(r)

	comment := store.Comment{
		PostID:  post.ID,
		UserID:  user.ID,
		Content: payload.Content,
	}

	ctx := r.Context()
	createdComment, err := app.store.Comments.Create(ctx, &comment)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, commentResponse{Data: *createdComment}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// GetComment godoc
//
//	@Summary		Get comment
//	@Description	Get a comment by id
//	@Tags			comments
//	@Produce		json
//	@Param			id	path		int	true	"Comment ID"
//	@Success		200	{object}	commentResponse
//	@Failure		400	{object}	errorResponse
//	@Failure		404	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Security		ApiKeyAuth
//	@Router			/comments/{id} [get]
func (app *application) getCommentHandler(w http.ResponseWriter, r *http.Request) {
	comment := app.getCommentContext(r)

	ctx := r.Context()
	comment, err := app.store.Comments.Read(ctx, comment.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, commentResponse{Data: *comment}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// GetCommentsByPostID godoc
//
//	@Summary		Get comments by post id
//	@Description	Get all comments for a post
//	@Tags			posts
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	commentsResponse
//	@Failure		400	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Security		ApiKeyAuth
//	@Router			/posts/{id}/comments [get]
func (app *application) getCommentsByPostIDHandler(w http.ResponseWriter, r *http.Request) {
	post := app.getPostContext(r)

	ctx := r.Context()
	comments, err := app.store.Comments.ReadByPostID(ctx, post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, commentsResponse{Data: comments}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// UpdateComment godoc
//
//	@Summary		Update comment
//	@Description	Update a comment by id
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Comment ID"
//	@Param			request	body		createCommentRequest	true	"Update comment request"
//	@Success		200		{object}	commentResponse
//	@Failure		400		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Security		ApiKeyAuth
//	@Router			/comments/{id} [put]
func (app *application) updateCommentHandler(w http.ResponseWriter, r *http.Request) {
	comment := app.getCommentContext(r)

	var payload createCommentRequest
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	if err := Validator.Struct(payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	commentToUpdate := store.Comment{
		ID:      comment.ID,
		Content: payload.Content,
	}

	ctx := r.Context()
	updatedComment, err := app.store.Comments.Update(ctx, &commentToUpdate)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, commentResponse{Data: *updatedComment}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// DeleteComment godoc
//
//	@Summary		Delete comment
//	@Description	Delete a comment by id
//	@Tags			comments
//	@Produce		json
//	@Param			id	path	int	true	"Comment ID"
//	@Success		204	"Success"
//	@Failure		400	{object}	errorResponse
//	@Failure		404	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Security		ApiKeyAuth
//	@Router			/comments/{id} [delete]
func (app *application) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	comment := app.getCommentContext(r)

	ctx := r.Context()
	err := app.store.Comments.Delete(ctx, comment.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

const commentContextKey = contextKey("comment")

func (app *application) commentsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		intID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			app.badRequest(w, r, err)
			return
		}

		comment, err := app.store.Comments.Read(r.Context(), intID)
		if err != nil {
			if err == sql.ErrNoRows {
				app.notFound(w, r)
				return
			}
			app.internalServerError(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), commentContextKey, comment)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) getCommentContext(r *http.Request) *store.Comment {
	comment, _ := r.Context().Value(commentContextKey).(*store.Comment)
	return comment
}
