package ratelimiter

import "time"

type Limiter interface {
	Allow(ip string) (bool, time.Duration)
}

type Config struct {
	RequestPerTimeFrame int
	TimeFrame           time.Duration
	Enabled             bool
}

func (rl *FixedWindowLimiter) Allow(ip string) (bool, time.Duration) {
	rl.Lock()
	defer rl.Unlock()

	count, exists := rl.clients[ip]
	if !exists || count < rl.limit {
		rl.Lock()
		if !exists {
			go rl.resetCount(ip)
		}
		rl.clients[ip] = count + 1
		rl.Unlock()
		return true, 0
	}

	return false, rl.window
}

func (rl *FixedWindowLimiter) resetCount(ip string) {
	time.Sleep(rl.window)
	rl.Lock()
	defer rl.Unlock()
	delete(rl.clients, ip)
}
