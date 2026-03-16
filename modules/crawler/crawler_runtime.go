package crawler

import (
	"fmt"
	"strings"
)

// Running reports whether the crawler worker loop is active.
func (c *Crawler) Running() bool {
	return c.running.Load()
}

// ConfigSnapshot returns a copy of the crawler configuration.
func (c *Crawler) ConfigSnapshot() CrawlerConfig {
	if c == nil || c.config == nil {
		return CrawlerConfig{}
	}
	return *c.config
}

// CrawlOnce triggers one synchronous crawl request using the crawler's
// current profile/proxy strategy and returns the raw crawl result.
func (c *Crawler) CrawlOnce(targetURL string) (*CrawlResult, error) {
	if c == nil {
		return nil, fmt.Errorf("crawler is nil")
	}
	if !c.running.Load() {
		return nil, fmt.Errorf("crawler is not running")
	}

	targetURL = strings.TrimSpace(targetURL)
	if targetURL == "" {
		return nil, fmt.Errorf("target URL is required")
	}

	task := &crawlTask{
		URL:   targetURL,
		Depth: 0,
	}
	c.stats.TotalRequests.Add(1)
	c.applyRateLimit()

	profile := c.getProfileForTask(task)
	if profile == nil {
		c.stats.FailedRequests.Add(1)
		return nil, fmt.Errorf("no profile available")
	}

	result := c.executeRequest(task, profile, c.getProxy())
	if result == nil {
		c.stats.FailedRequests.Add(1)
		return nil, fmt.Errorf("crawl execution failed")
	}

	if result.Blocked {
		c.stats.BlockedRequests.Add(1)
	} else if result.StatusCode >= 200 && result.StatusCode < 300 {
		c.stats.SuccessRequests.Add(1)
		c.stats.TotalBytes.Add(result.ContentLength)
	} else {
		c.stats.FailedRequests.Add(1)
	}

	if c.feedback != nil {
		c.feedback.Collect(result)
	}
	if c.mlAdapter != nil {
		c.mlAdapter.RecordResult(result)
	}

	select {
	case c.resultsChan <- result:
	default:
		// Keep manual crawl non-blocking when result channel is full.
	}

	return result, nil
}
