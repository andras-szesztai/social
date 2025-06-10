package main

import (
	"testing"

	"github.com/andras-szesztai/social/internal/auth"
	"github.com/andras-szesztai/social/internal/store"
	"github.com/andras-szesztai/social/internal/store/cache"
	"go.uber.org/zap"
)

func newTestApplication(t *testing.T) *application {
	t.Helper()

	return &application{
		logger:        zap.NewNop().Sugar(),
		store:         store.NewMockStore(),
		cache:         cache.NewMockCache(),
		authenticator: auth.NewMockAuth(),
		config: config{
			redis: redisConfig{
				enabled: true,
			},
		},
	}
}
