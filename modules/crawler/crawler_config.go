// Package crawler - Active crawler module for testing anti-crawling capabilities and training detection models
//
// Design goals:
//  1. Simulate real browser behavior (using real browser Profiles)
//  2. Support multiple crawling strategies (concurrency control, request intervals, randomization)
//  3. Built-in profile rotation and proxy pool
//  4. Data feedback to ML training system
//
// Typical use cases:
//   - Testing the detection capabilities of your own anti-crawling system
//   - Collecting real traffic samples for model training
//   - Verifying fingerprint spoofing effectiveness
package crawler

import (
	"sync/atomic"
	"time"

	"github.com/vistone/fingerprint/modules/profiles"
)

// CrawlerConfig defines crawler configuration options
type CrawlerConfig struct {
	// Basic settings
	Name           string   // crawler name/identifier
	TargetURLs     []string // target URL list
	MaxDepth       int      // maximum crawling depth
	MaxPages       int      // maximum pages to crawl (0 = unlimited)
	RequestTimeout time.Duration

	// Concurrency control
	Workers    int           // number of concurrent workers
	RateLimit  time.Duration // base request interval
	RateJitter float64       // interval randomization factor (0-1)
	BurstLimit int           // burst request limit

	// Profile strategy
	ProfileStrategy ProfileStrategyType // profile rotation strategy
	ProfilePool     []string            // specified profile pool (empty = use random)
	RotateInterval  time.Duration       // profile rotation interval

	// Proxy configuration
	ProxyStrategy ProxyStrategyType
	ProxyList     []string // proxy address list

	// Behavior simulation
	HumanLike   bool          // enable human behavior simulation
	ScrollDelay time.Duration // scroll delay
	ClickRandom bool          // random clicking
	FormFill    bool          // auto form filling

	// Data collection
	CollectMode CollectMode // data collection mode
	FeedbackURL string      // data feedback endpoint

	// Anti-detection
	StealthMode    bool // enable advanced stealth mode
	BlockAvoidance bool // enable block avoidance
}

// ProfileStrategyType - Profile rotation strategy type
type ProfileStrategyType string

const (
	ProfileStrategyRandom   ProfileStrategyType = "random"   // Fully random
	ProfileStrategyRotate   ProfileStrategyType = "rotate"   // Sequential rotation
	ProfileStrategySticky   ProfileStrategyType = "sticky"   // Session persistence
	ProfileStrategyAdaptive ProfileStrategyType = "adaptive" // Adaptive selection
)

// ProxyStrategyType - Proxy strategy type
type ProxyStrategyType string

const (
	ProxyStrategyNone    ProxyStrategyType = "none"    // No proxy
	ProxyStrategyRandom  ProxyStrategyType = "random"  // Random proxy
	ProxyStrategyRotate  ProxyStrategyType = "rotate"  // Sequential rotation
	ProxyStrategySession ProxyStrategyType = "session" // Session binding
)

// CollectMode - Data collection mode
type CollectMode string

const (
	CollectModeNone     CollectMode = "none"     // No collection
	CollectModeResponse CollectMode = "response" // Collect response
	CollectModeFull     CollectMode = "full"     // Collect full page
	CollectModeBlocked  CollectMode = "blocked"  // Only collect blocked requests
)

// DefaultCrawlerConfig - Default crawler configuration
var DefaultCrawlerConfig = &CrawlerConfig{
	MaxDepth:        3,
	Workers:         5,
	RateLimit:       2 * time.Second,
	RateJitter:      0.3,
	BurstLimit:      3,
	ProfileStrategy: ProfileStrategyRotate,
	RotateInterval:  30 * time.Second,
	ProxyStrategy:   ProxyStrategyNone,
	HumanLike:       true,
	ScrollDelay:     500 * time.Millisecond,
	CollectMode:     CollectModeBlocked,
	StealthMode:     true,
}

// CrawlResult - Crawl result
type CrawlResult struct {
	URL           string
	Depth         int
	StatusCode    int
	ContentType   string
	ContentLength int64
	Headers       map[string][]string
	Body          []byte
	Fingerprint   *profiles.ClientProfile
	ProxyUsed     string
	Duration      time.Duration
	Timestamp     time.Time
	Blocked       bool
	BlockReason   string
	DetectionInfo map[string]interface{} // Anti-crawling detection info
}

// CrawlStats - Crawl statistics
type CrawlStats struct {
	TotalRequests   atomic.Int64
	SuccessRequests atomic.Int64
	FailedRequests  atomic.Int64
	BlockedRequests atomic.Int64
	TotalBytes      atomic.Int64
	StartTime       time.Time
	EndTime         *time.Time
}

// SuccessRate - Calculate success rate
func (s *CrawlStats) SuccessRate() float64 {
	total := s.TotalRequests.Load()
	if total == 0 {
		return 0
	}
	return float64(s.SuccessRequests.Load()) / float64(total)
}

// BlockRate - Calculate block rate
func (s *CrawlStats) BlockRate() float64 {
	total := s.TotalRequests.Load()
	if total == 0 {
		return 0
	}
	return float64(s.BlockedRequests.Load()) / float64(total)
}
