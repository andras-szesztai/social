package cache

import (
	"context"

	"github.com/andras-szesztai/social/internal/store"
)

func NewMockCache() *Storage {
	return &Storage{
		Users: &MockUserCache{},
	}
}

type MockUserCache struct{}

func (m *MockUserCache) Get(ctx context.Context, id int64) (*store.User, error) {
	return nil, nil
}

func (m *MockUserCache) Set(ctx context.Context, user *store.User) error {
	return nil
}

func (m *MockUserCache) Delete(ctx context.Context, id int64) error {
	return nil
}
