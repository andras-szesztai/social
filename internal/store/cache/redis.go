package cache

import (
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	Client *redis.Client
}

func NewRedisCache(addr string, db int, password string) (*RedisCache, error) {
	opts := redis.Options{
		Addr:     addr,
		DB:       db,
		Password: password,
	}

	client := redis.NewClient(&opts)

	return &RedisCache{Client: client}, nil
}
