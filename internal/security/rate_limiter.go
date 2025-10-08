package security

import (
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	requests := rl.requests[identifier]
	var validRequests []time.Time

	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	if len(validRequests) >= rl.limit {
		return false
	}

	validRequests = append(validRequests, now)
	rl.requests[identifier] = validRequests

	return true
}

func (rl *RateLimiter) RetryAfter(identifier string) time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	requests := rl.requests[identifier]
	if len(requests) == 0 {
		return 0
	}

	oldest := requests[0]
	windowEnd := oldest.Add(rl.window)
	now := time.Now()

	if windowEnd.After(now) {
		return windowEnd.Sub(now)
	}

	return 0
}

func (rl *RateLimiter) GetRemaining(identifier string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	requests := rl.requests[identifier]
	var validRequests []time.Time

	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	return rl.limit - len(validRequests)
}

func (rl *RateLimiter) Reset(identifier string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.requests, identifier)
}

func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	for identifier, requests := range rl.requests {
		var validRequests []time.Time

		for _, reqTime := range requests {
			if reqTime.After(windowStart) {
				validRequests = append(validRequests, reqTime)
			}
		}

		if len(validRequests) == 0 {
			delete(rl.requests, identifier)
		} else {
			rl.requests[identifier] = validRequests
		}
	}
}

func (rl *RateLimiter) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		rl.Cleanup()
	}
}
