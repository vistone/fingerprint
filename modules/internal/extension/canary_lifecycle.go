package extension

import (
	"context"
	"fmt"
	"time"
)

type CanaryReport struct {
	Stage     CanaryStage
	Duration  time.Duration
	StartTime time.Time
	EndTime   time.Time

	// Traffic distribution
	TotalRequests     int64
	NewMethodRequests int64
	OldMethodRequests int64
	NewMethodPercent  float64

	// Success rate
	SuccessRate float64
	ErrorRate   float64

	// Latency comparison
	NewMethodAvgLatency time.Duration
	OldMethodAvgLatency time.Duration
	LatencyDifference   float64 // percentage

	// Cache effectiveness
	CacheHitRate float64

	// Result consistency
	Consistency float64 // 0-1

	// Recommendation
	Recommendation string
}

// GenerateCanaryReport generates a canary report
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

	// Calculate the percentage of new method usage
	if metrics.TotalRequests > 0 {
		report.NewMethodPercent = float64(metrics.NewMethodRequests) / float64(metrics.TotalRequests)
	}

	// Generate recommendation
	if report.ErrorRate > 0.03 {
		report.Recommendation = "❌ Rollback immediately (error rate too high)"
	} else if report.ErrorRate > 0.01 {
		report.Recommendation = "⚠️  Recommend delaying canary (error rate slightly high)"
	} else if report.CacheHitRate > 0 && report.CacheHitRate < 0.5 {
		report.Recommendation = "⚠️  Recommend checking cache strategy"
	} else if report.SuccessRate > 0.98 {
		report.Recommendation = "✅ Ready to upgrade to next stage"
	} else {
		report.Recommendation = "✅ Canary progressing well, continue monitoring"
	}

	return report
}

// PrintCanaryReport prints the canary report
func (report *CanaryReport) Print() {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════╗")
	fmt.Printf("║  Canary Report: %s\n", report.Stage)
	fmt.Println("╚════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Printf("📊 Traffic Distribution:\n")
	fmt.Printf("  Total Requests: %d\n", report.TotalRequests)
	fmt.Printf("  New Method:   %d (%.1f%%)\n", report.NewMethodRequests, report.NewMethodPercent*100)
	fmt.Printf("  Old Method:   %d (%.1f%%)\n", report.OldMethodRequests, (1-report.NewMethodPercent)*100)
	fmt.Println()

	fmt.Printf("✅ Success Rate: %.2f%%\n", report.SuccessRate*100)
	fmt.Printf("❌ Error Rate: %.2f%%\n", report.ErrorRate*100)
	fmt.Println()

	if report.CacheHitRate > 0 {
		fmt.Printf("💾 Cache Hit Rate: %.1f%%\n", report.CacheHitRate*100)
		fmt.Println()
	}

	fmt.Printf("⏱️  Canary Duration: %v\n", report.Duration)
	fmt.Println()

	fmt.Printf("📋 Recommendation:\n  %s\n", report.Recommendation)
	fmt.Println()
}

// ========================================================================
// Helper functions
// ========================================================================

// hashRequestID hashes a requestID (for consistent traffic distribution)
func hashRequestID(requestID string) uint32 {
	hash := uint32(0)
	for _, ch := range requestID {
		hash = hash*31 + uint32(ch)
	}
	return hash
}

// WaitForCanaryCompletion waits for canary completion
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
			fmt.Printf("[%s] Requests: %d, Error Rate: %.2f%%, Cache Hit: %.1f%%\n",
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
// Canary lifecycle management
// ========================================================================

// CanaryLifecycle manages the canary release lifecycle
type CanaryLifecycle struct {
	collector   *CanaryMetricsCollector
	healthCheck *CanaryHealthCheck
	stages      []CanaryStage
	currentIdx  int
}

// NewCanaryLifecycle creates a canary lifecycle manager
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

// StartStage starts a new canary stage
func (c *CanaryLifecycle) StartStage(stage CanaryStage, duration time.Duration) {
	percentages := map[CanaryStage]float64{
		CanaryStage5Percent:   0.05,
		CanaryStage25Percent:  0.25,
		CanaryStage50Percent:  0.50,
		CanaryStage100Percent: 1.00,
	}

	percentage := percentages[stage]
	c.collector.EnableCanary(stage, percentage, duration)
	fmt.Printf("🚀 Starting canary: %s (%.0f%% traffic, duration %v)\n", stage, percentage*100, duration)
}

// EndStage ends the current canary stage
func (c *CanaryLifecycle) EndStage() *CanaryReport {
	startTime := c.collector.config.StartTime
	report := c.collector.GenerateCanaryReport(startTime)
	c.collector.DisableCanary()
	return report
}

// CanUpgrade checks whether the canary can be upgraded to the next stage
func (c *CanaryLifecycle) CanUpgrade() bool {
	healthy, alerts, _ := c.healthCheck.CheckHealth(context.Background())

	if !healthy && len(alerts) > 0 {
		fmt.Printf("⚠️  Found %d alerts, continue monitoring\n", len(alerts))
		return false
	}

	return healthy
}

// PrintLifecycleStatus prints the lifecycle status
func (c *CanaryLifecycle) PrintLifecycleStatus() {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("Canary rollout progress")
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
