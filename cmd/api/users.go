package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/andras-szesztai/social/internal/store"
	"github.com/andras-szesztai/social/internal/utils"
	"github.com/go-chi/chi/v5"
)

// GetUser godoc
//
//	@Summary		Get user
//	@Description	Get user by id
//	@Tags			users
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	userResponse
//	@Failure		400	{object}	errorResponse
//	@Failure		404	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserContext(r)
	if err := app.jsonResponse(w, http.StatusOK, userResponse{Data: *user}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// FollowUser godoc
//
//	@Summary		Follow user
//	@Description	Follow a user by their ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int					true	"User ID"
//	@Param			request	body	followUserRequest	true	"Follow request"
//	@Success		204		"Success"
//	@Failure		400		{object}	errorResponse
//	@Failure		404		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Security		ApiKeyAuth
//	@Router			/users/{id}/follow [post]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := app.getUserContext(r)
	followedId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	ctx := r.Context()
	if err := app.store.Users.Follow(ctx, followedId, followerUser.ID); err != nil {
		if err == sql.ErrNoRows {
			app.notFound(w, r)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

// UnfollowUser godoc
//
//	@Summary		Unfollow user
//	@Description	Unfollow a user by their ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int					true	"User ID"
//	@Param			request	body	followUserRequest	true	"Unfollow request"
//	@Success		204		"Success"
//	@Failure		400		{object}	errorResponse
//	@Failure		404		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Security		ApiKeyAuth
//	@Router			/users/{id}/unfollow [post]
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserContext(r)
	followedId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	ctx := r.Context()
	if err := app.store.Users.Unfollow(ctx, followedId, user.ID); err != nil {
		if err == sql.ErrNoRows {
			app.notFound(w, r)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

// GetUserFeed godoc
//
//	@Summary		Get user feed
//	@Description	Get the feed for a user
//	@Tags			users
//	@Produce		json
//	@Param			limit	query		int			false	"Limit"			default(20)
//	@Param			offset	query		int			false	"Offset"		default(0)
//	@Param			sort	query		string		false	"Sort order"	Enums(asc, desc)	default(desc)
//	@Param			tags	query		[]string	false	"Tags"
//	@Param			search	query		string		false	"Search term"
//	@Success		200		{object}	userFeedResponse
//	@Failure		400		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Security		ApiKeyAuth
//	@Router			/users/feed [get]
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	pagination := utils.FeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
		Tags:   []string{},
		Search: "",
	}

	fq, err := pagination.Parse(r)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}
	if err := Validator.Struct(fq); err != nil {
		app.badRequest(w, r, err)
		return
	}

	ctx := r.Context()

	user := app.getUserContext(r)

	feed, err := app.store.Users.ReadFeed(ctx, user.ID, fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, userFeedResponse{Data: feed}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// ActivateUser godoc
//
//	@Summary		Activate user
//	@Description	Activate a user by their token
//	@Tags			users
//	@Produce		json
//	@Param			id		path	int		true	"User ID"
//	@Param			token	path	string	true	"Token"
//	@Success		204		"User activated"
//	@Failure		400		{object}	errorResponse
//	@Failure		404		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Security		ApiKeyAuth
//	@Router			/users/{id}/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserContext(r)
	token := chi.URLParam(r, "token")

	ctx := r.Context()

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	err := app.store.Users.Activate(ctx, user.ID, hashToken)
	if err != nil {
		if err == store.ErrNotFound {
			app.notFound(w, r)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

// DeleteUser godoc
//
//	@Summary		Delete user
//	@Description	Delete a user by their ID
//	@Tags			users
//	@Produce		json
//	@Param			id	path	int	true	"User ID"
//	@Success		204	"User deleted"
//	@Failure		400	{object}	errorResponse
//	@Failure		404	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [delete]
func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserContext(r)

	ctx := r.Context()

	err := app.store.Users.Delete(ctx, user.ID)
	if err != nil {
		app.internalServerError(w, r, err)
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

		user, err := app.store.Users.ReadByID(r.Context(), idInt)
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
