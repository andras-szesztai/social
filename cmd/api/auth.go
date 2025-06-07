package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/andras-szesztai/social/internal/mailer"
	"github.com/andras-szesztai/social/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// RegisterUser godoc
//
//	@Summary		Register user
//	@Description	Register a new user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body	RegisterUserPayload	true	"Register user payload"
//	@Success		201		"User created"
//	@Failure		400		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/authentication/register [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	if err := Validator.Struct(payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
		Role: &store.Role{
			Name: "user",
		},
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	token := uuid.New().String()
	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.expiry)
	if err != nil {
		switch err {
		case store.ErrEmailAlreadyExists:
			app.badRequest(w, r, err)
		case store.ErrUsernameAlreadyExists:
			app.badRequest(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	err = app.mailer.Send(mailer.UserInvitationTemplate, user.Username, user.Email, map[string]any{
		"Username":      user.Username,
		"ActivationURL": fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, token),
	}, app.config.env == "production")
	if err != nil {
		app.logger.Error("failed to send user invitation email", "error", err)
		// rollback the user creation
		if err := app.store.Users.Delete(ctx, user.ID); err != nil {
			app.logger.Error("failed to rollback user creation", "error", err)
		}
		app.internalServerError(w, r, err)
	}

	if err := app.jsonResponse(w, http.StatusCreated, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

type CreateTokenPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// CreateToken godoc
//
//	@Summary		Create token
//	@Description	Create a new token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body	CreateTokenPayload	true	"Create token payload"
//	@Success		200		"Token created"
//	@Failure		400		{object}	errorResponse
//	@Failure		401		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/authentication/token [post]
func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateTokenPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	if err := Validator.Struct(payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	user, err := app.store.Users.ReadByEmail(r.Context(), payload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			app.unauthorized(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	if err := user.Password.Compare(payload.Password); err != nil {
		app.unauthorized(w, r, err)
		return
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"nbf": time.Now().Unix(),
		"iat": time.Now().Unix(),
		"iss": app.config.auth.token.iss,
		"aud": app.config.auth.token.aud,
	}

	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, token); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
