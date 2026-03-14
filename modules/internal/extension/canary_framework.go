package extension

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ========================================================================
// Canary release framework: ProcessWithPipeline production canary management
// ========================================================================

// CanaryStage represents a canary release stage
type CanaryStage string

const (
	CanaryStage5Percent   CanaryStage = "5%"
	CanaryStage25Percent  CanaryStage = "25%"
	CanaryStage50Percent  CanaryStage = "50%"
	CanaryStage100Percent CanaryStage = "100%"
)

// CanaryConfig holds canary release configuration
type CanaryConfig struct {
	// Canary stage
	Stage CanaryStage

	// Target traffic percentage
	TargetPercentage float64

	// Enable canary
	Enabled bool

	// Canary start time
	StartTime time.Time

	// Canary duration (0 means no limit)
	Duration time.Duration
}

// CanaryMetrics holds canary release metrics
type CanaryMetrics struct {
	// Basic metrics
	TotalRequests      int64
	NewMethodRequests  int64
	OldMethodRequests  int64
	SuccessfulRequests int64
	FailedRequests     int64

	// Latency metrics
	TotalLatency time.Duration
	MinLatency   time.Duration
	MaxLatency   time.Duration
	AvgLatency   time.Duration
	P50Latency   time.Duration
	P95Latency   time.Duration
	P99Latency   time.Duration

	// Cache metrics
	CacheHits    int64
	CacheMisses  int64
	CacheHitRate float64

	// Error metrics
	ErrorCount int64
	ErrorRate  float64

	// Resource metrics
	MemoryUsage int64
	GCTime      time.Duration

	// Timestamps
	CollectedAt time.Time
}

// CanaryMetricsCollector collects canary release metrics
type CanaryMetricsCollector struct {
	mu sync.RWMutex

	// Current configuration
	config *CanaryConfig

	// Metrics data
	metrics *CanaryMetrics

	// Historical data (for trend analysis)
	history []*CanaryMetrics

	// Traffic router
	router *CanaryRouter
}

// CanaryRouter routes canary traffic
type CanaryRouter struct {
	config *CanaryConfig
}

// ShouldUseNewMethod determines whether to use the new method
func (r *CanaryRouter) ShouldUseNewMethod(requestID string) bool {
	if !r.config.Enabled {
		return false
	}

	// Consistent assignment based on requestID hash
	hash := hashRequestID(requestID)
	// Calculate threshold: percentage * max uint32
	threshold := uint32(float64(^uint32(0)) * r.config.TargetPercentage)
	return hash < threshold
}

// NewCanaryMetricsCollector creates a canary metrics collector
func NewCanaryMetricsCollector(config *CanaryConfig) *CanaryMetricsCollector {
	if config == nil {
		config = &CanaryConfig{
			Stage:            CanaryStage5Percent,
			TargetPercentage: 0.05,
			Enabled:          false,
		}
	}

	return &CanaryMetricsCollector{
		config: config,
		metrics: &CanaryMetrics{
			CollectedAt: time.Now(),
		},
		history: make([]*CanaryMetrics, 0, 1000),
		router:  &CanaryRouter{config: config},
	}
}

// EnableCanary enables canary release
func (c *CanaryMetricsCollector) EnableCanary(stage CanaryStage, percentage float64, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.config.Stage = stage
	c.config.TargetPercentage = percentage
	c.config.Enabled = true
	c.config.StartTime = time.Now()
	c.config.Duration = duration

	fmt.Printf("✅ Canary enabled: %s (%.0f%% traffic)\n", stage, percentage*100)
}

// DisableCanary disables canary release
func (c *CanaryMetricsCollector) DisableCanary() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.config.Enabled = false
	fmt.Println("⚠️  Canary disabled")
}

// RecordRequest records a request
func (c *CanaryMetricsCollector) RecordRequest(requestID string, useNewMethod bool, duration time.Duration, success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	atomic.AddInt64(&c.metrics.TotalRequests, 1)

	if useNewMethod {
		atomic.AddInt64(&c.metrics.NewMethodRequests, 1)
	} else {
		atomic.AddInt64(&c.metrics.OldMethodRequests, 1)
	}

	if success {
		atomic.AddInt64(&c.metrics.SuccessfulRequests, 1)
	} else {
		atomic.AddInt64(&c.metrics.FailedRequests, 1)
	}

	// Update latency metrics
	c.metrics.TotalLatency += duration
	if duration < c.metrics.MinLatency || c.metrics.MinLatency == 0 {
		c.metrics.MinLatency = duration
	}
	if duration > c.metrics.MaxLatency {
		c.metrics.MaxLatency = duration
	}
}

// RecordCacheHit records a cache hit
func (c *CanaryMetricsCollector) RecordCacheHit(hit bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if hit {
		atomic.AddInt64(&c.metrics.CacheHits, 1)
	} else {
		atomic.AddInt64(&c.metrics.CacheMisses, 1)
	}
}

// GetCurrentMetrics returns the current metrics
func (c *CanaryMetricsCollector) GetCurrentMetrics() *CanaryMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics := *c.metrics
	metrics.CollectedAt = time.Now()

	// Calculate average latency
	if metrics.TotalRequests > 0 {
		metrics.AvgLatency = metrics.TotalLatency / time.Duration(metrics.TotalRequests)
		metrics.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
	}

	// Calculate cache hit rate
	totalCacheOps := metrics.CacheHits + metrics.CacheMisses
	if totalCacheOps > 0 {
		metrics.CacheHitRate = float64(metrics.CacheHits) / float64(totalCacheOps)
	}

	return &metrics
}

// SnapshotMetrics saves a metrics snapshot (for historical analysis)
func (c *CanaryMetricsCollector) SnapshotMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Copy current metrics directly without calling other functions that require locks
	snapshot := &CanaryMetrics{
		TotalRequests:      atomic.LoadInt64(&c.metrics.TotalRequests),
		NewMethodRequests:  atomic.LoadInt64(&c.metrics.NewMethodRequests),
		OldMethodRequests:  atomic.LoadInt64(&c.metrics.OldMethodRequests),
		SuccessfulRequests: atomic.LoadInt64(&c.metrics.SuccessfulRequests),
		FailedRequests:     atomic.LoadInt64(&c.metrics.FailedRequests),
		TotalLatency:       c.metrics.TotalLatency,
		MinLatency:         c.metrics.MinLatency,
		MaxLatency:         c.metrics.MaxLatency,
		CacheHits:          atomic.LoadInt64(&c.metrics.CacheHits),
		CacheMisses:        atomic.LoadInt64(&c.metrics.CacheMisses),
		CollectedAt:        time.Now(),
	}

	// Calculate average latency
	if snapshot.TotalRequests > 0 {
		snapshot.AvgLatency = snapshot.TotalLatency / time.Duration(snapshot.TotalRequests)
		snapshot.ErrorRate = float64(snapshot.FailedRequests) / float64(snapshot.TotalRequests)
	}

	// Calculate cache hit rate
	totalCacheOps := snapshot.CacheHits + snapshot.CacheMisses
	if totalCacheOps > 0 {
		snapshot.CacheHitRate = float64(snapshot.CacheHits) / float64(totalCacheOps)
	}

	c.history = append(c.history, snapshot)

	// Keep only the most recent 1000 snapshots
	if len(c.history) > 1000 {
		c.history = c.history[1:]
	}
}

// GetMetricsHistory returns the metrics history
func (c *CanaryMetricsCollector) GetMetricsHistory(hours int) []*CanaryMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
	result := make([]*CanaryMetrics, 0)

	for _, m := range c.history {
		if m.CollectedAt.After(cutoff) {
			result = append(result, m)
		}
	}

	return result
}

// CanaryHealthCheck performs canary health checks
type CanaryHealthCheck struct {
	collector  *CanaryMetricsCollector
	thresholds *CanaryThresholds
}

// CanaryThresholds defines canary thresholds
type CanaryThresholds struct {
	// Alert thresholds
	MaxErrorRate    float64
	MaxP99Latency   time.Duration
	MinCacheHitRate float64
	MaxMemoryGrowth int64 // bytes/minute

	// Rollback trigger thresholds
	CriticalErrorRate  float64
	CriticalP99Latency time.Duration
}

// DefaultCanaryThresholds returns default canary thresholds
func DefaultCanaryThresholds() *CanaryThresholds {
	return &CanaryThresholds{
		// Alert thresholds (strict)
		MaxErrorRate:    0.01, // 1%
		MaxP99Latency:   150 * time.Millisecond,
		MinCacheHitRate: 0.5,    // 50%
		MaxMemoryGrowth: 100000, // 100KB

		// Rollback trigger thresholds (strict)
		CriticalErrorRate:  0.03, // 3%
		CriticalP99Latency: 500 * time.Millisecond,
	}
}

// NewCanaryHealthCheck creates a canary health check
func NewCanaryHealthCheck(collector *CanaryMetricsCollector) *CanaryHealthCheck {
	return &CanaryHealthCheck{
		collector:  collector,
		thresholds: DefaultCanaryThresholds(),
	}
}

// CheckHealth checks canary health status
func (c *CanaryHealthCheck) CheckHealth(ctx context.Context) (healthy bool, alerts []string, critique string) {
	metrics := c.collector.GetCurrentMetrics()

	if metrics.TotalRequests == 0 {
		return true, []string{}, "insufficient data"
	}

	alerts = []string{}
	healthy = true

	// Check error rate
	if metrics.ErrorRate > c.thresholds.CriticalErrorRate {
		critique = fmt.Sprintf("❌ Critical error rate: %.2f%% (> %.2f%%)",
			metrics.ErrorRate*100, c.thresholds.CriticalErrorRate*100)
		return false, alerts, critique
	}

	if metrics.ErrorRate > c.thresholds.MaxErrorRate {
		alerts = append(alerts, fmt.Sprintf("⚠️  High error rate: %.2f%%", metrics.ErrorRate*100))
		healthy = false
	}

	// Check latency
	if metrics.P99Latency > c.thresholds.CriticalP99Latency && metrics.P99Latency > 0 {
		critique = fmt.Sprintf("❌ Critical latency increase: P99=%.0fms (> %.0fms)",
			metrics.P99Latency.Seconds()*1000, c.thresholds.CriticalP99Latency.Seconds()*1000)
		return false, alerts, critique
	}

	if metrics.P99Latency > c.thresholds.MaxP99Latency && metrics.P99Latency > 0 {
		alerts = append(alerts, fmt.Sprintf("⚠️  P99 latency increase: %.0fms", metrics.P99Latency.Seconds()*1000))
		healthy = false
	}

	// Check cache hit rate
	if metrics.CacheHitRate > 0 && metrics.CacheHitRate < c.thresholds.MinCacheHitRate {
		alerts = append(alerts, fmt.Sprintf("⚠️  Cache hit rate too low: %.1f%%", metrics.CacheHitRate*100))
		healthy = false
	}

	if healthy {
		critique = "✅ Canary healthy"
	}

	return healthy, alerts, critique
}

// CanaryReport represents a canary report
