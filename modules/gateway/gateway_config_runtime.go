package gateway

import (
	"github.com/vistone/fingerprint/modules/agent"
	"github.com/vistone/fingerprint/modules/ml"
)

// GetConfig returns the current gateway configuration (read-only copy).
func (g *Gateway) GetConfig() *GatewayConfig {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.config.Clone()
}

// UpdateConfig hot-updates the gateway configuration (thread-safe).
func (g *Gateway) UpdateConfig(apply func(cfg *GatewayConfig)) {
	g.mu.Lock()
	defer g.mu.Unlock()

	previous := g.config.Clone()
	next := g.config.Clone()
	apply(next)

	g.config = next

	if g.limiter != nil && (previous.RateLimitRequests != next.RateLimitRequests ||
		previous.RateLimitBurst != next.RateLimitBurst ||
		previous.RateLimitWindow != next.RateLimitWindow) {
		g.limiter.Update(next.RateLimitRequests, next.RateLimitBurst, next.RateLimitWindow)
	}

	if g.cache != nil && (previous.CacheSize != next.CacheSize || previous.CacheTTL != next.CacheTTL) {
		g.cache.Reconfigure(next.CacheSize, next.CacheTTL)
	}
}

// Clone returns a deep copy of GatewayConfig.
func (c *GatewayConfig) Clone() *GatewayConfig {
	if c == nil {
		return nil
	}

	clone := *c
	if c.TrustedProxies != nil {
		clone.TrustedProxies = append([]string(nil), c.TrustedProxies...)
	}
	if c.APIKeys != nil {
		clone.APIKeys = append([]string(nil), c.APIKeys...)
	}
	clone.AgentConfig = cloneAgentConfig(c.AgentConfig)
	clone.MLServiceConfig = cloneMLServiceConfig(c.MLServiceConfig)
	clone.ClosedLoopConfig = cloneClosedLoopConfig(c.ClosedLoopConfig)

	return &clone
}

func cloneAgentConfig(cfg *agent.AgentConfig) *agent.AgentConfig {
	if cfg == nil {
		return nil
	}

	clone := *cfg
	return &clone
}

func cloneMLServiceConfig(cfg *ml.ServiceConfig) *ml.ServiceConfig {
	if cfg == nil {
		return nil
	}

	clone := *cfg
	return &clone
}

func cloneClosedLoopConfig(cfg *ClosedLoopConfig) *ClosedLoopConfig {
	if cfg == nil {
		return nil
	}

	clone := *cfg
	return &clone
}
