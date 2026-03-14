package crawler

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vistone/fingerprint/modules/profiles"
)

// ProfileManager - Profile manager
type ProfileManager struct {
	profiles   []*profiles.ClientProfile
	current    int
	strategy   ProfileStrategyType
	mu         sync.RWMutex
	sessionMap map[string]*profiles.ClientProfile // URL -> Profile mapping
}

// ProxyManager - Proxy manager
type ProxyManager struct {
	proxies  []*url.URL
	current  int
	strategy ProxyStrategyType
	mu       sync.RWMutex
}

// Crawler - Crawler instance
type Crawler struct {
	config      *CrawlerConfig
	stats       *CrawlStats
	resultsChan chan *CrawlResult
	workQueue   chan *crawlTask
	errChan     chan error

	// Profile management
	profileManager *ProfileManager
	currentProfile *profiles.ClientProfile
	profileMu      sync.RWMutex

	// Proxy management
	proxyManager *ProxyManager

	// Control
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running atomic.Bool

	// Anti-detection
	stealth *StealthEngine

	// Data feedback
	feedback *FeedbackCollector

	logger *slog.Logger
}

// crawlTask - Internal crawl task
type crawlTask struct {
	URL      string
	Depth    int
	Referrer string
	Profile  *profiles.ClientProfile
}

// NewCrawler - Create crawler instance
func NewCrawler(config *CrawlerConfig) *Crawler {
	if config == nil {
		config = DefaultCrawlerConfig
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := &Crawler{
		config:      config,
		stats:       &CrawlStats{StartTime: time.Now()},
		resultsChan: make(chan *CrawlResult, 100),
		workQueue:   make(chan *crawlTask, 1000),
		errChan:     make(chan error, 100),
		ctx:         ctx,
		cancel:      cancel,
		logger:      slog.Default().With("component", "crawler"),
	}

	// Initialize profile manager
	c.profileManager = &ProfileManager{
		strategy:   config.ProfileStrategy,
		sessionMap: make(map[string]*profiles.ClientProfile),
	}
	c.initProfilePool()

	// Initialize proxy manager
	c.proxyManager = &ProxyManager{strategy: config.ProxyStrategy}
	c.initProxyPool()

	// Initialize anti-detection engine
	if config.StealthMode {
		c.stealth = NewStealthEngine(config)
	}

	// Initialize data collector
	if config.CollectMode != CollectModeNone {
		c.feedback = NewFeedbackCollector(config.FeedbackURL)
	}

	return c
}

// Start - Start crawler
func (c *Crawler) Start() error {
	if c.running.Load() {
		return fmt.Errorf("crawler already running")
	}

	c.running.Store(true)
	c.stats.StartTime = time.Now()

	// Start workers
	for i := 0; i < c.config.Workers; i++ {
		c.wg.Add(1)
		go c.worker(i)
	}

	// Start profile rotator
	if c.config.RotateInterval > 0 {
		go c.profileRotator()
	}

	// Add seed URLs
	for _, url := range c.config.TargetURLs {
		c.workQueue <- &crawlTask{
			URL:   url,
			Depth: 0,
		}
	}

	c.logger.Info("crawler started",
		"workers", c.config.Workers,
		"targets", len(c.config.TargetURLs))

	return nil
}

// Stop - Stop crawler
func (c *Crawler) Stop() {
	if !c.running.Load() {
		return
	}

	c.running.Store(false)
	c.cancel()
	c.wg.Wait()

	now := time.Now()
	c.stats.EndTime = &now

	close(c.resultsChan)
	close(c.workQueue)
	close(c.errChan)

	c.logger.Info("crawler stopped",
		"duration", now.Sub(c.stats.StartTime),
		"total", c.stats.TotalRequests.Load(),
		"success_rate", fmt.Sprintf("%.2f%%", c.stats.SuccessRate()*100))
}

// GetStats - Get statistics
func (c *Crawler) GetStats() *CrawlStats {
	return c.stats
}

// GetResults - Get results channel
func (c *Crawler) GetResults() <-chan *CrawlResult {
	return c.resultsChan
}

// GetErrors - Get errors channel
func (c *Crawler) GetErrors() <-chan error {
	return c.errChan
}
