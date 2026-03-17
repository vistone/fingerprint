package gateway

import (
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/defense"
	tlsmod "github.com/vistone/fingerprint/modules/tls"
)

type RateLimiter struct {
	rate     int
	burst    int
	window   time.Duration
	visitors map[string]*Visitor
	mu       sync.Mutex
	stopCh   chan struct{}
}

// Visitor represents a visitor
type Visitor struct {
	tokens     float64
	lastRefill time.Time
	lastSeen   time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate, burst int, window time.Duration) *RateLimiter {
	if rate <= 0 {
		rate = 1000
	}
	if burst <= 0 {
		burst = rate
	}
	if window <= 0 {
		window = time.Second
	}

	rl := &RateLimiter{
		rate:     rate,
		burst:    burst,
		window:   window,
		visitors: make(map[string]*Visitor),
		stopCh:   make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

// Allow checks whether the request is allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	fillRatePerSec := float64(rl.rate) / rl.window.Seconds()
	if fillRatePerSec <= 0 {
		fillRatePerSec = float64(rl.rate)
	}
	capacity := float64(rl.burst)

	v, ok := rl.visitors[key]
	if !ok {
		rl.visitors[key] = &Visitor{
			tokens:     capacity - 1,
			lastRefill: now,
			lastSeen:   now,
		}
		return true
	}

	// Refill tokens based on elapsed time
	if now.After(v.lastRefill) {
		elapsed := now.Sub(v.lastRefill).Seconds()
		v.tokens += elapsed * fillRatePerSec
		if v.tokens > capacity {
			v.tokens = capacity
		}
		v.lastRefill = now
	}

	if v.tokens < 1 {
		v.lastSeen = now
		return false
	}

	v.tokens -= 1
	v.lastSeen = now
	return true
}

// Update applies a new rate-limit configuration without replacing the limiter instance.
func (rl *RateLimiter) Update(rate, burst int, window time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rate <= 0 {
		rate = 1000
	}
	if burst <= 0 {
		burst = rate
	}
	if window <= 0 {
		window = time.Second
	}

	rl.rate = rate
	rl.burst = burst
	rl.window = window
}

// cleanup removes expired visitors
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for key, v := range rl.visitors {
				if time.Since(v.lastSeen) > 5*time.Minute {
					delete(rl.visitors, key)
				}
			}
			rl.mu.Unlock()
		case <-rl.stopCh:
			return
		}
	}
}

// Close stops the rate limiter's background goroutine
func (rl *RateLimiter) Close() {
	close(rl.stopCh)
}

// FingerprintCache is a fingerprint cache (based on LRUCache implementation)
type FingerprintCache struct {
	lru *LRUCache
}

// NewFingerprintCache creates a new fingerprint cache
func NewFingerprintCache(size int, ttl time.Duration) *FingerprintCache {
	return &FingerprintCache{
		lru: NewLRUCache(size, ttl),
	}
}

// Get retrieves from cache
func (c *FingerprintCache) Get(key string) (*AnalyzeResponse, bool) {
	val, found := c.lru.Get(key)
	if !found {
		return nil, false
	}
	return cloneAnalyzeResponse(val.(*AnalyzeResponse)), true
}

// Set stores in cache
func (c *FingerprintCache) Set(key string, response *AnalyzeResponse) {
	c.lru.Set(key, cloneAnalyzeResponse(response), 0) // Use LRUCache's default TTL
}

// Reconfigure updates cache capacity and TTL without replacing the cache instance.
func (c *FingerprintCache) Reconfigure(size int, ttl time.Duration) {
	if c == nil || c.lru == nil {
		return
	}
	c.lru.Reconfigure(size, ttl)
}

func cloneAnalyzeResponse(resp *AnalyzeResponse) *AnalyzeResponse {
	if resp == nil {
		return nil
	}
	clone := *resp
	if resp.Findings != nil {
		clone.Findings = append([]defense.Finding(nil), resp.Findings...)
	}
	if resp.DefenseHints != nil {
		clone.DefenseHints = append([]string(nil), resp.DefenseHints...)
	}
	if resp.JA3 != nil {
		ja3 := *resp.JA3
		clone.JA3 = &ja3
	}
	if resp.JA4 != nil {
		ja4 := *resp.JA4
		clone.JA4 = &ja4
	}
	if resp.JA4H != nil {
		ja4h := *resp.JA4H
		if resp.JA4H.Headers != nil {
			ja4h.Headers = append([]string(nil), resp.JA4H.Headers...)
		}
		clone.JA4H = &ja4h
	}
	return &clone
}

// calculateJA3 calculates JA3 fingerprint (using tls package implementation)
func calculateJA3(spec core.ClientHelloSpec) *tlsmod.JA3Result {
	return tlsmod.CalculateJA3(spec)
}

// calculateJA4 calculates JA4 fingerprint (using tls package implementation)
func calculateJA4(spec core.ClientHelloSpec) *tlsmod.JA4Result {
	return tlsmod.CalculateJA4(spec)
}
