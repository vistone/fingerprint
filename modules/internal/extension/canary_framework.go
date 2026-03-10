package extension

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ========================================================================
// translated comment
// ========================================================================

// translated comment
type CanaryStage string

const (
	CanaryStage5Percent   CanaryStage = "5%"
	CanaryStage25Percent  CanaryStage = "25%"
	CanaryStage50Percent  CanaryStage = "50%"
	CanaryStage100Percent CanaryStage = "100%"
)

// translated comment
type CanaryConfig struct {
	// translated comment
	Stage CanaryStage

	// translated comment
	TargetPercentage float64

	// translated comment
	Enabled bool

	// translated comment
	StartTime time.Time

	// translated comment
	Duration time.Duration
}

// translated comment
type CanaryMetrics struct {
	// translated comment
	TotalRequests      int64
	NewMethodRequests  int64
	OldMethodRequests  int64
	SuccessfulRequests int64
	FailedRequests     int64

	// translated comment
	TotalLatency time.Duration
	MinLatency   time.Duration
	MaxLatency   time.Duration
	AvgLatency   time.Duration
	P50Latency   time.Duration
	P95Latency   time.Duration
	P99Latency   time.Duration

	// translated comment
	CacheHits    int64
	CacheMisses  int64
	CacheHitRate float64

	// translated comment
	ErrorCount int64
	ErrorRate  float64

	// translated comment
	MemoryUsage int64
	GCTime      time.Duration

	// translated comment
	CollectedAt time.Time
}

// translated comment
type CanaryMetricsCollector struct {
	mu sync.RWMutex

	// translated comment
	config *CanaryConfig

	// translated comment
	metrics *CanaryMetrics

	// translated comment
	history []*CanaryMetrics

	// translated comment
	router *CanaryRouter
}

// translated comment
type CanaryRouter struct {
	config *CanaryConfig
}

// translated comment
func (r *CanaryRouter) ShouldUseNewMethod(requestID string) bool {
	if !r.config.Enabled {
		return false
	}

	// translated comment
	hash := hashRequestID(requestID)
	// translated comment
	threshold := uint32(float64(^uint32(0)) * r.config.TargetPercentage)
	return hash < threshold
}

// translated comment
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

// translated comment
func (c *CanaryMetricsCollector) EnableCanary(stage CanaryStage, percentage float64, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.config.Stage = stage
	c.config.TargetPercentage = percentage
	c.config.Enabled = true
	c.config.StartTime = time.Now()
	c.config.Duration = duration

	fmt.Printf("✅ 灰度已启用: %s (%.0f%% 流量)\n", stage, percentage*100)
}

// translated comment
func (c *CanaryMetricsCollector) DisableCanary() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.config.Enabled = false
	fmt.Println("⚠️  灰度已禁用")
}

// translated comment
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

	// translated comment
	c.metrics.TotalLatency += duration
	if duration < c.metrics.MinLatency || c.metrics.MinLatency == 0 {
		c.metrics.MinLatency = duration
	}
	if duration > c.metrics.MaxLatency {
		c.metrics.MaxLatency = duration
	}
}

// translated comment
func (c *CanaryMetricsCollector) RecordCacheHit(hit bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if hit {
		atomic.AddInt64(&c.metrics.CacheHits, 1)
	} else {
		atomic.AddInt64(&c.metrics.CacheMisses, 1)
	}
}

// translated comment
func (c *CanaryMetricsCollector) GetCurrentMetrics() *CanaryMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics := *c.metrics
	metrics.CollectedAt = time.Now()

	// translated comment
	if metrics.TotalRequests > 0 {
		metrics.AvgLatency = metrics.TotalLatency / time.Duration(metrics.TotalRequests)
		metrics.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
	}

	// translated comment
	totalCacheOps := metrics.CacheHits + metrics.CacheMisses
	if totalCacheOps > 0 {
		metrics.CacheHitRate = float64(metrics.CacheHits) / float64(totalCacheOps)
	}

	return &metrics
}

// translated comment
func (c *CanaryMetricsCollector) SnapshotMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// translated comment
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

	// translated comment
	if snapshot.TotalRequests > 0 {
		snapshot.AvgLatency = snapshot.TotalLatency / time.Duration(snapshot.TotalRequests)
		snapshot.ErrorRate = float64(snapshot.FailedRequests) / float64(snapshot.TotalRequests)
	}

	// translated comment
	totalCacheOps := snapshot.CacheHits + snapshot.CacheMisses
	if totalCacheOps > 0 {
		snapshot.CacheHitRate = float64(snapshot.CacheHits) / float64(totalCacheOps)
	}

	c.history = append(c.history, snapshot)

	// translated comment
	if len(c.history) > 1000 {
		c.history = c.history[1:]
	}
}

// translated comment
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

// translated comment
type CanaryHealthCheck struct {
	collector  *CanaryMetricsCollector
	thresholds *CanaryThresholds
}

// translated comment
type CanaryThresholds struct {
	// translated comment
	MaxErrorRate    float64
	MaxP99Latency   time.Duration
	MinCacheHitRate float64
	MaxMemoryGrowth int64 // bytes/minute

	// translated comment
	CriticalErrorRate  float64
	CriticalP99Latency time.Duration
}

// translated comment
func DefaultCanaryThresholds() *CanaryThresholds {
	return &CanaryThresholds{
		// translated comment
		MaxErrorRate:    0.01, // 1%
		MaxP99Latency:   150 * time.Millisecond,
		MinCacheHitRate: 0.5,    // 50%
		MaxMemoryGrowth: 100000, // 100KB

		// translated comment
		CriticalErrorRate:  0.03, // 3%
		CriticalP99Latency: 500 * time.Millisecond,
	}
}

// translated comment
func NewCanaryHealthCheck(collector *CanaryMetricsCollector) *CanaryHealthCheck {
	return &CanaryHealthCheck{
		collector:  collector,
		thresholds: DefaultCanaryThresholds(),
	}
}

// translated comment
func (c *CanaryHealthCheck) CheckHealth(ctx context.Context) (healthy bool, alerts []string, critique string) {
	metrics := c.collector.GetCurrentMetrics()

	if metrics.TotalRequests == 0 {
		return true, []string{}, "无足够的数据"
	}

	alerts = []string{}
	healthy = true

	// translated comment
	if metrics.ErrorRate > c.thresholds.CriticalErrorRate {
		critique = fmt.Sprintf("❌ 严重错误率: %.2f%% (> %.2f%%)",
			metrics.ErrorRate*100, c.thresholds.CriticalErrorRate*100)
		return false, alerts, critique
	}

	if metrics.ErrorRate > c.thresholds.MaxErrorRate {
		alerts = append(alerts, fmt.Sprintf("⚠️  高错误率: %.2f%%", metrics.ErrorRate*100))
		healthy = false
	}

	// translated comment
	if metrics.P99Latency > c.thresholds.CriticalP99Latency && metrics.P99Latency > 0 {
		critique = fmt.Sprintf("❌ 严重延迟升高: P99=%.0fms (> %.0fms)",
			metrics.P99Latency.Seconds()*1000, c.thresholds.CriticalP99Latency.Seconds()*1000)
		return false, alerts, critique
	}

	if metrics.P99Latency > c.thresholds.MaxP99Latency && metrics.P99Latency > 0 {
		alerts = append(alerts, fmt.Sprintf("⚠️  P99 延迟升高: %.0fms", metrics.P99Latency.Seconds()*1000))
		healthy = false
	}

	// translated comment
	if metrics.CacheHitRate > 0 && metrics.CacheHitRate < c.thresholds.MinCacheHitRate {
		alerts = append(alerts, fmt.Sprintf("⚠️  缓存命中率过低: %.1f%%", metrics.CacheHitRate*100))
		healthy = false
	}

	if healthy {
		critique = "✅ 灰度健康"
	}

	return healthy, alerts, critique
}

// translated comment
type CanaryReport struct {
	Stage     CanaryStage
	Duration  time.Duration
	StartTime time.Time
	EndTime   time.Time

	// translated comment
	TotalRequests     int64
	NewMethodRequests int64
	OldMethodRequests int64
	NewMethodPercent  float64

	// translated comment
	SuccessRate float64
	ErrorRate   float64

	// translated comment
	NewMethodAvgLatency time.Duration
	OldMethodAvgLatency time.Duration
	LatencyDifference   float64 // percentage

	// translated comment
	CacheHitRate float64

	// translated comment
	Consistency float64 // 0-1

	// translated comment
	Recommendation string
}

// translated comment
func (c *CanaryMetricsCollector) GenerateCanaryReport(startTime time.Time) *CanaryReport {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics := c.metrics
	duration := time.Since(startTime)

	report := &CanaryReport{
		Stage:             c.config.Stage,
		Duration:          duration,
		StartTime:         startTime,
		EndTime:           time.Now(),
		TotalRequests:     metrics.TotalRequests,
		NewMethodRequests: metrics.NewMethodRequests,
		OldMethodRequests: metrics.OldMethodRequests,
		SuccessRate:       float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests),
		ErrorRate:         metrics.ErrorRate,
		CacheHitRate:      metrics.CacheHitRate,
	}

	// translated comment
	if metrics.TotalRequests > 0 {
		report.NewMethodPercent = float64(metrics.NewMethodRequests) / float64(metrics.TotalRequests)
	}

	// translated comment
	if report.ErrorRate > 0.03 {
		report.Recommendation = "❌ 立即回滚 (错误率过高)"
	} else if report.ErrorRate > 0.01 {
		report.Recommendation = "⚠️  建议延缓灰度 (错误率略高)"
	} else if report.CacheHitRate > 0 && report.CacheHitRate < 0.5 {
		report.Recommendation = "⚠️  建议检查缓存策略"
	} else if report.SuccessRate > 0.98 {
		report.Recommendation = "✅ 可以升级到下一阶段"
	} else {
		report.Recommendation = "✅ 灰度进行良好，继续监控"
	}

	return report
}

// translated comment
func (report *CanaryReport) Print() {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════╗")
	fmt.Printf("║  灰度报告: %s\n", report.Stage)
	fmt.Println("╚════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Printf("📊 流量分布:\n")
	fmt.Printf("  总请求数: %d\n", report.TotalRequests)
	fmt.Printf("  新方式:   %d (%.1f%%)\n", report.NewMethodRequests, report.NewMethodPercent*100)
	fmt.Printf("  旧方式:   %d (%.1f%%)\n", report.OldMethodRequests, (1-report.NewMethodPercent)*100)
	fmt.Println()

	fmt.Printf("✅ 成功率: %.2f%%\n", report.SuccessRate*100)
	fmt.Printf("❌ 错误率: %.2f%%\n", report.ErrorRate*100)
	fmt.Println()

	if report.CacheHitRate > 0 {
		fmt.Printf("💾 缓存命中率: %.1f%%\n", report.CacheHitRate*100)
		fmt.Println()
	}

	fmt.Printf("⏱️  灰度时长: %v\n", report.Duration)
	fmt.Println()

	fmt.Printf("📋 建议:\n  %s\n", report.Recommendation)
	fmt.Println()
}

// ========================================================================
// translated comment
// ========================================================================

// translated comment
func hashRequestID(requestID string) uint32 {
	hash := uint32(0)
	for _, ch := range requestID {
		hash = hash*31 + uint32(ch)
	}
	return hash
}

// translated comment
func WaitForCanaryCompletion(ctx context.Context, collector *CanaryMetricsCollector, duration time.Duration) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	deadline := time.Now().Add(duration)

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			metrics := collector.GetCurrentMetrics()
			fmt.Printf("[%s] 请求数: %d, 错误率: %.2f%%, 缓存命中: %.1f%%\n",
				time.Now().Format("15:04"), metrics.TotalRequests, metrics.ErrorRate*100, metrics.CacheHitRate*100)

			if time.Now().After(deadline) {
				return
			}

		case <-time.After(duration):
			return
		}
	}
}

// ========================================================================
// translated comment
// ========================================================================

// translated comment
type CanaryLifecycle struct {
	collector   *CanaryMetricsCollector
	healthCheck *CanaryHealthCheck
	stages      []CanaryStage
	currentIdx  int
}

// translated comment
func NewCanaryLifecycle() *CanaryLifecycle {
	config := &CanaryConfig{}
	collector := NewCanaryMetricsCollector(config)

	return &CanaryLifecycle{
		collector:   collector,
		healthCheck: NewCanaryHealthCheck(collector),
		stages: []CanaryStage{
			CanaryStage5Percent,
			CanaryStage25Percent,
			CanaryStage50Percent,
			CanaryStage100Percent,
		},
		currentIdx: 0,
	}
}

// translated comment
func (c *CanaryLifecycle) StartStage(stage CanaryStage, duration time.Duration) {
	percentages := map[CanaryStage]float64{
		CanaryStage5Percent:   0.05,
		CanaryStage25Percent:  0.25,
		CanaryStage50Percent:  0.50,
		CanaryStage100Percent: 1.00,
	}

	percentage := percentages[stage]
	c.collector.EnableCanary(stage, percentage, duration)
	fmt.Printf("🚀 开始灰度: %s (%.0f%% 流量, 持续 %v)\n", stage, percentage*100, duration)
}

// translated comment
func (c *CanaryLifecycle) EndStage() *CanaryReport {
	startTime := c.collector.config.StartTime
	report := c.collector.GenerateCanaryReport(startTime)
	c.collector.DisableCanary()
	return report
}

// translated comment
func (c *CanaryLifecycle) CanUpgrade() bool {
	healthy, alerts, _ := c.healthCheck.CheckHealth(context.Background())

	if !healthy && len(alerts) > 0 {
		fmt.Printf("⚠️  发现 %d 个警告，继续观察\n", len(alerts))
		return false
	}

	return healthy
}

// translated comment
func (c *CanaryLifecycle) PrintLifecycleStatus() {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("灰度推出进度")
	fmt.Println("═══════════════════════════════════════════════════")

	for i, stage := range c.stages {
		status := "⏳"
		if i < c.currentIdx {
			status = "✅"
		} else if i == c.currentIdx {
			status = "🔄"
		}

		fmt.Printf("%s %s\n", status, stage)
	}

	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println()
}
