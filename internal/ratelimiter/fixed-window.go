package ratelimiter

import (
	"sync"
	"time"
)

type FixedWindowLimiter struct {
	sync.RWMutex
	clients map[string]int
	limit   int
	window  time.Duration
}

func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		clients: make(map[string]int),
		limit:   limit,
		window:  window,
	}
}
