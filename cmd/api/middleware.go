package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (app *application) BasicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedBasic(w, r, fmt.Errorf("authorization header is required"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			app.unauthorized(w, r, fmt.Errorf("invalid authorization header"))
			return
		}

		token := parts[1]
		if token == "" {
			app.unauthorized(w, r, fmt.Errorf("token is required"))
			return
		}

		decodedToken, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			app.unauthorized(w, r, fmt.Errorf("invalid token"))
			return
		}

		username := app.config.auth.basic.username
		password := app.config.auth.basic.password

		parts = strings.Split(string(decodedToken), ":")
		if len(parts) != 2 || parts[0] != username || parts[1] != password {
			app.unauthorized(w, r, fmt.Errorf("invalid credentials"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorized(w, r, fmt.Errorf("authorization header is required"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			app.unauthorized(w, r, fmt.Errorf("invalid authorization header"))
			return
		}

		token := parts[1]
		if token == "" {
			app.unauthorized(w, r, fmt.Errorf("token is required"))
			return
		}

		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorized(w, r, fmt.Errorf("invalid token"))
			return
		}

		claims := jwtToken.Claims.(jwt.MapClaims)
		userId, err := strconv.ParseInt(fmt.Sprintf("%.0f", claims["sub"].(float64)), 10, 64)
		if err != nil {
			app.unauthorized(w, r, fmt.Errorf("invalid token"))
			return
		}

		user, err := app.store.Users.ReadByID(r.Context(), userId)
		if err != nil {
			app.unauthorized(w, r, fmt.Errorf("invalid token"))
			return
		}

		ctx := context.WithValue(r.Context(), contextKey("user"), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
