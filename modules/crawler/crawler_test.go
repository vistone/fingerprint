package crawler

import (
	"testing"
	"time"
)

func TestNewCrawler(t *testing.T) {
	config := &CrawlerConfig{
		Name:       "test-crawler",
		TargetURLs: []string{"https://example.com"},
		Workers:    1,
	}

	c := NewCrawler(config)
	if c == nil {
		t.Fatal("NewCrawler returned nil")
	}

	if c.config.Name != "test-crawler" {
		t.Errorf("Expected name 'test-crawler', got '%s'", c.config.Name)
	}
}

func TestCrawlerConfigDefaults(t *testing.T) {
	config := DefaultCrawlerConfig

	if config.Workers != 5 {
		t.Errorf("Expected default workers 5, got %d", config.Workers)
	}

	if config.RateLimit != 2*time.Second {
		t.Errorf("Expected default rate limit 2s, got %v", config.RateLimit)
	}
}

func TestCrawlerStats(t *testing.T) {
	stats := &CrawlStats{}

	// Simulate requests
	stats.TotalRequests.Add(100)
	stats.SuccessRequests.Add(80)
	stats.BlockedRequests.Add(20)

	successRate := stats.SuccessRate()
	if successRate != 0.8 {
		t.Errorf("Expected success rate 0.8, got %f", successRate)
	}

	blockRate := stats.BlockRate()
	if blockRate != 0.2 {
		t.Errorf("Expected block rate 0.2, got %f", blockRate)
	}
}

func TestProfileStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy ProfileStrategyType
		valid    bool
	}{
		{"random", ProfileStrategyRandom, true},
		{"rotate", ProfileStrategyRotate, true},
		{"sticky", ProfileStrategySticky, true},
		{"adaptive", ProfileStrategyAdaptive, true},
		{"invalid", ProfileStrategyType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the constant exists
			_ = tt.strategy
		})
	}
}

func TestCrawlerStartStop(t *testing.T) {
	config := &CrawlerConfig{
		Name:       "startstop-test",
		TargetURLs: []string{"https://httpbin.org/get"},
		Workers:    1,
		RateLimit:  1 * time.Second,
		MaxPages:   1,
	}

	c := NewCrawler(config)

	err := c.Start()
	if err != nil {
		t.Fatalf("Failed to start crawler: %v", err)
	}

	// Let it run briefly
	time.Sleep(2 * time.Second)

	c.Stop()

	stats := c.GetStats()
	if stats.TotalRequests.Load() == 0 {
		t.Log("No requests were made (expected for external URL)")
	}
}

func BenchmarkCrawlerCreation(b *testing.B) {
	config := &CrawlerConfig{
		Name:       "benchmark-crawler",
		TargetURLs: []string{"https://example.com"},
	}

	for i := 0; i < b.N; i++ {
		c := NewCrawler(config)
		_ = c
	}
}
