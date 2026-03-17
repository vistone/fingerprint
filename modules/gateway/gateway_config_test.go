package gateway

import (
	"testing"
	"time"
)

func TestGatewayGetConfig_ReturnsClone(t *testing.T) {
	g := NewGateway(nil)
	defer g.Close()

	cfg := g.GetConfig()
	cfg.RiskThreshold = 0.99
	cfg.TrustedProxies = append(cfg.TrustedProxies, "203.0.113.10")

	latest := g.GetConfig()
	if latest.RiskThreshold == 0.99 {
		t.Fatal("GetConfig returned mutable internal config")
	}
	if len(latest.TrustedProxies) != 0 {
		t.Fatal("trusted proxies should not be mutated through GetConfig result")
	}
}

func TestGatewayUpdateConfig_ReconfiguresLimiterAndCache(t *testing.T) {
	g := NewGateway(nil)
	defer g.Close()

	g.UpdateConfig(func(cfg *GatewayConfig) {
		cfg.RateLimitRequests = 123
		cfg.RateLimitBurst = 7
		cfg.RateLimitWindow = 2 * time.Second
		cfg.CacheSize = 64
		cfg.CacheTTL = 3 * time.Minute
	})

	if g.limiter.rate != 123 || g.limiter.burst != 7 || g.limiter.window != 2*time.Second {
		t.Fatalf("limiter not reconfigured, got rate=%d burst=%d window=%s", g.limiter.rate, g.limiter.burst, g.limiter.window)
	}

	if g.cache.lru.maxEntries != 64 || g.cache.lru.defaultTTL != 3*time.Minute {
		t.Fatalf("cache not reconfigured, got size=%d ttl=%s", g.cache.lru.maxEntries, g.cache.lru.defaultTTL)
	}
}
