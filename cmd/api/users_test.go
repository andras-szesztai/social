package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andras-szesztai/social/internal/store"
	"github.com/andras-szesztai/social/internal/store/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUser(t *testing.T) {
	app := newTestApplication(t)
	mux := app.mountRoutes()

	testToken, err := app.authenticator.GenerateToken(nil)
	assert.NoError(t, err)

	t.Run("should not allow unauthenticated users to get a user", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		assert.NoError(t, err)

		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("should allow authenticated users to get a user", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		assert.NoError(t, err)

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testToken))

		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("should hit cache first and if not found, hit the database", func(t *testing.T) {
		mockCache := app.cache.Users.(*cache.MockUserCache)
		mockCache.On("Get", 1).Return(nil, nil)
		mockCache.On("Get", 1).Return(&store.User{ID: 1}, nil)
		mockCache.On("Set", mock.Anything, mock.Anything).Return(nil)

		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		assert.NoError(t, err)

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testToken))

		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		mockCache.AssertExpectations(t)
		mockCache.AssertNumberOfCalls(t, "Get", 2)
		mockCache.AssertNumberOfCalls(t, "Set", 1)
	})

}
