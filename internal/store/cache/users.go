package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/andras-szesztai/social/internal/store"
)

type UserStorage struct {
	redis *RedisCache
}

func NewUserStorage(redis *RedisCache) *UserStorage {
	return &UserStorage{redis: redis}
}

func (s *UserStorage) Get(ctx context.Context, id int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user:%d", id)
	user, err := s.redis.Client.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}
	if user == "" {
		return nil, nil
	}

	userData := &store.User{}
	err = json.Unmarshal([]byte(user), userData)
	if err != nil {
		return nil, err
	}

	return userData, nil
}

func (s *UserStorage) Set(ctx context.Context, user *store.User) error {
	if user.ID <= 0 {
		return fmt.Errorf("user ID is required")
	}

	cacheKey := fmt.Sprintf("user:%d", user.ID)
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.redis.Client.SetEx(ctx, cacheKey, userJSON, 1*time.Hour).Err()
}

func (s *UserStorage) Delete(ctx context.Context, id int64) error {
	cacheKey := fmt.Sprintf("user:%d", id)
	return s.redis.Client.Del(ctx, cacheKey).Err()
}
