package cache

import (
	"context"

	"github.com/andras-szesztai/social/internal/store"
)

type Storage struct {
	Users interface {
		Get(ctx context.Context, id int64) (*store.User, error)
		Set(ctx context.Context, user *store.User) error
		Delete(ctx context.Context, id int64) error
	}
}

func NewRedisStorage(redis *RedisCache) *Storage {
	return &Storage{
		Users: NewUserStorage(redis),
	}
}
