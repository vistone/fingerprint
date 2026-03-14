package waf

import (
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	rate   int           // tokens per interval
	burst  int           // maximum burst size
	window time.Duration // time window

	buckets map[string]*bucket
	mu      sync.RWMutex
}

// bucket represents a token bucket for a client
type bucket struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate, burst int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		rate:    rate,
		burst:   burst,
		window:  window,
		buckets: make(map[string]*bucket),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request is allowed for the given client
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	b, exists := rl.buckets[clientID]
	if !exists {
		b = &bucket{
			tokens:     rl.burst - 1, // consume one token
			lastRefill: time.Now(),
		}
		rl.buckets[clientID] = b
		rl.mu.Unlock()
		return true
	}
	rl.mu.Unlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens based on time passed
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)
	tokensToAdd := int(elapsed / rl.window * time.Duration(rl.rate))

	if tokensToAdd > 0 {
		b.tokens = min(rl.burst, b.tokens+tokensToAdd)
		b.lastRefill = now
	}

	// Check if we have tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// cleanup removes stale buckets periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for id, b := range rl.buckets {
			b.mu.Lock()
			if now.Sub(b.lastRefill) > 10*time.Minute {
				delete(rl.buckets, id)
			}
			b.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
