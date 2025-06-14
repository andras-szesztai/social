package main

import (
	"context"
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

type updatePostRequest struct {
	Title   string   `json:"title" validate:"omitempty,max=255"`
	Content string   `json:"content" validate:"omitempty,max=1000"`
	Tags    []string `json:"tags" validate:"omitempty,max=10"`
}

// CreatePost godoc
//
//	@Summary		Create post
//	@Description	Create a new post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			request	body		createPostRequest	true	"Create post request"
//	@Success		201		{object}	postResponse
//	@Failure		400		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Security		ApiKeyAuth
//	@Router			/posts [post]
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

	user := app.getUserContext(r)

	post := store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  user.ID,
	}

	ctx := r.Context()
	createdPost, err := app.store.Posts.Create(ctx, &post)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, postResponse{Data: *createdPost}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// GetPost godoc
//
//	@Summary		Get post
//	@Description	Get a post by id
//	@Tags			posts
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	postResponse
//	@Failure		400	{object}	errorResponse
//	@Failure		404	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [get]
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := app.getPostContext(r)
	if err := app.jsonResponse(w, http.StatusOK, postResponse{Data: *post}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// UpdatePost godoc
//
//	@Summary		Update post
//	@Description	Update a post by id
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Post ID"
//	@Param			request	body		updatePostRequest	true	"Update post request"
//	@Success		200		{object}	postResponse
//	@Failure		400		{object}	errorResponse
//	@Failure		404		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [put]
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := app.getPostContext(r)

	var payload updatePostRequest
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	if err := Validator.Struct(payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	if payload.Title == "" {
		payload.Title = post.Title
	}
	if payload.Content == "" {
		payload.Content = post.Content
	}
	if len(payload.Tags) == 0 {
		payload.Tags = post.Tags
	}

	postToUpdate := store.Post{
		ID:      post.ID,
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		Version: post.Version,
	}

	ctx := r.Context()
	updatedPost, err := app.store.Posts.Update(ctx, &postToUpdate)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, postResponse{Data: *updatedPost}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// DeletePost godoc
//
//	@Summary		Delete post
//	@Description	Delete a post by id
//	@Tags			posts
//	@Produce		json
//	@Param			id	path	int	true	"Post ID"
//	@Success		204	"Success"
//	@Failure		400	{object}	errorResponse
//	@Failure		404	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [delete]
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	post := app.getPostContext(r)

	ctx := r.Context()
	err := app.store.Posts.Delete(ctx, post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

const postContextKey = contextKey("post")

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		intID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			app.badRequest(w, r, err)
			return
		}

		post, err := app.store.Posts.Read(r.Context(), intID)
		if err != nil {
			if err == sql.ErrNoRows {
				app.notFound(w, r)
				return
			}
			app.internalServerError(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), postContextKey, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) getPostContext(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postContextKey).(*store.Post)
	return post
}
