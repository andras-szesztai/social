package main

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/andras-szesztai/social/internal/store"
	"github.com/go-chi/chi/v5"
)

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserContext(r)
	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

type followUserRequest struct {
	FollowerID int64 `json:"follower_id" validate:"required"`
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserContext(r)

	// TODO: use auth middleware to get the user id
	var payload followUserRequest
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequest(w, r, err)
		return
	}
	if err := Validator.Struct(payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	ctx := r.Context()
	if err := app.store.Users.Follow(ctx, user.ID, payload.FollowerID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserContext(r)

	// TODO: use auth middleware to get the user id
	var payload followUserRequest
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequest(w, r, err)
		return
	}
	if err := Validator.Struct(payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	ctx := r.Context()
	if err := app.store.Users.Unfollow(ctx, user.ID, payload.FollowerID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

const userContextKey = contextKey("user")

func (app *application) usersContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		user, err := app.store.Users.Get(r.Context(), idInt)
		if err != nil {
			if err == sql.ErrNoRows {
				app.notFound(w, r)
				return
			}
			app.internalServerError(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) getUserContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userContextKey).(*store.User)
	return user
}
