package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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

}
