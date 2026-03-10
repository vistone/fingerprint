package gateway

import (
	"testing"
	"time"
)

func TestRateLimiter_TokenBucketBurstAndRefill(t *testing.T) {
	rl := NewRateLimiter(10, 2, time.Second)
	key := "127.0.0.1"

	if !rl.Allow(key) {
		t.Fatal("expected first request to pass")
	}
	if !rl.Allow(key) {
		t.Fatal("expected second request to pass")
	}
	if rl.Allow(key) {
		t.Fatal("expected third request to be rate limited")
	}

	time.Sleep(120 * time.Millisecond)
	if !rl.Allow(key) {
		t.Fatal("expected request to pass after token refill")
	}
}

func TestNewRateLimiter_Defaults(t *testing.T) {
	rl := NewRateLimiter(0, 0, 0)
	if rl == nil {
		t.Fatal("expected non-nil limiter")
	}
	if !rl.Allow("k") {
		t.Fatal("expected first request to pass with default config")
	}
}
