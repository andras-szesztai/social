package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/andras-szesztai/social/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
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

		user, err := app.getUser(r.Context(), userId)
		if err != nil {
			app.unauthorized(w, r, fmt.Errorf("invalid token"))
			return
		}

		ctx := context.WithValue(r.Context(), contextKey("user"), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.getUserContext(r)
		post := app.getPostContext(r)

		if user.ID == post.UserID {
			next.ServeHTTP(w, r)
			return
		}

		allowed, err := app.checkRolePrecedence(r.Context(), user.Role.Level, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbidden(w, r, fmt.Errorf("forbidden"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) checkRolePrecedence(ctx context.Context, userRoleID int64, requiredRole string) (bool, error) {
	role, err := app.store.Roles.ReadByName(ctx, requiredRole)
	if err != nil {
		return false, err
	}

	return userRoleID >= role.Level, nil
}

func (app *application) getUser(ctx context.Context, id int64) (*store.User, error) {
	if !app.config.redis.enabled {
		return app.store.Users.ReadByID(ctx, id)
	}

	user, err := app.cache.Users.Get(ctx, id)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	if user != nil {
		app.logger.Infow("user found in cache", "user", user)
		return user, nil
	}

	user, err = app.store.Users.ReadByID(ctx, id)
	if err != nil {
		return nil, err
	}

	app.logger.Infow("user not found in cache, read from database", "user", user)

	if err := app.cache.Users.Set(ctx, user); err != nil {
		app.logger.Errorw("failed to set user in cache", "error", err)
		return user, nil
	}

	return user, nil
}

func (app *application) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.rateLimiter.Enabled {
			ip := r.RemoteAddr
			allowed, _ := app.rateLimiter.Allow(ip)
			if !allowed {
				app.tooManyRequests(w, r, fmt.Errorf("too many requests"))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
