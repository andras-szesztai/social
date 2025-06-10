package cache

import (
	"context"

	"github.com/andras-szesztai/social/internal/store"
	"github.com/stretchr/testify/mock"
)

func NewMockCache() *Storage {
	return &Storage{
		Users: &MockUserCache{},
	}
}

type MockUserCache struct {
	mock.Mock
}

func (m *MockUserCache) Get(ctx context.Context, id int64) (*store.User, error) {
	args := m.Called(id)
	return args.Get(0).(*store.User), args.Error(1)
}

func (m *MockUserCache) Set(ctx context.Context, user *store.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserCache) Delete(ctx context.Context, id int64) error {
	args := m.Called(id)
	return args.Error(0)
}
