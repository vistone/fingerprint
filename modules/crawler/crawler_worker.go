package crawler

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net/url"
	"time"

	"github.com/vistone/fingerprint/modules/profiles"
)

// worker - Worker goroutine
func (c *Crawler) worker(id int) {
	defer c.wg.Done()

	logger := c.logger.With("worker", id)

	for {
		select {
		case <-c.ctx.Done():
			return
		case task, ok := <-c.workQueue:
			if !ok {
				return
			}
			c.processTask(task, logger)
		}
	}
}

// processTask - Process single task
func (c *Crawler) processTask(task *crawlTask, logger *slog.Logger) {
	c.stats.TotalRequests.Add(1)

	// Apply rate limiting
	c.applyRateLimit()

	// Get profile
	profile := c.getProfileForTask(task)
	if profile == nil {
		c.stats.FailedRequests.Add(1)
		return
	}

	// Get proxy
	proxy := c.getProxy()

	// Execute request
	result := c.executeRequest(task, profile, proxy)

	// Process result
	if result.Blocked {
		c.stats.BlockedRequests.Add(1)
		logger.Warn("request blocked",
			"url", task.URL,
			"reason", result.BlockReason)
	} else if result.StatusCode >= 200 && result.StatusCode < 300 {
		c.stats.SuccessRequests.Add(1)
		c.stats.TotalBytes.Add(result.ContentLength)
	} else {
		c.stats.FailedRequests.Add(1)
	}

	// Send result
	select {
	case c.resultsChan <- result:
	case <-c.ctx.Done():
	}

	// Data feedback
	if c.feedback != nil {
		c.feedback.Collect(result)
	}

	// ML online learning feedback
	if c.mlAdapter != nil {
		c.mlAdapter.RecordResult(result)
	}

	// Deep crawling
	if !result.Blocked && task.Depth < c.config.MaxDepth {
		c.discoverLinks(result, task.Depth+1)
	}
}

// applyRateLimit - Apply rate limiting
func (c *Crawler) applyRateLimit() {
	baseDelay := c.config.RateLimit
	if c.config.RateJitter > 0 {
		jitter := time.Duration(float64(baseDelay) * c.config.RateJitter * (rand.Float64()*2 - 1))
		baseDelay += jitter
	}
	time.Sleep(baseDelay)
}

// executeRequest - Execute HTTP request
func (c *Crawler) executeRequest(task *crawlTask, profile *profiles.ClientProfile, proxy *url.URL) *CrawlResult {
	result := &CrawlResult{
		URL:         task.URL,
		Depth:       task.Depth,
		Timestamp:   time.Now(),
		Fingerprint: profile,
	}

	if proxy != nil {
		result.ProxyUsed = proxy.String()
	}

	// Create request executor
	executor := NewRequestExecutor(profile, proxy, c.config)

	// Execute request
	resp, err := executor.Do(c.ctx, task.URL)
	if err != nil {
		result.Blocked = true
		result.BlockReason = err.Error()
		return result
	}

	// Parse response
	result.StatusCode = resp.StatusCode
	result.Headers = resp.Header
	result.ContentType = resp.Header.Get("Content-Type")
	result.Body = resp.Body
	result.ContentLength = int64(len(resp.Body))
	result.Duration = resp.Duration

	// Detect if blocked
	result.DetectionInfo = c.detectBlocking(result)
	if result.DetectionInfo["blocked"] == true {
		result.Blocked = true
		result.BlockReason = result.DetectionInfo["reason"].(string)
	}

	return result
}

// detectBlocking - Detect if blocked
func (c *Crawler) detectBlocking(result *CrawlResult) map[string]interface{} {
	info := make(map[string]interface{})
	info["blocked"] = false

	// Status code detection
	if result.StatusCode == 403 || result.StatusCode == 429 || result.StatusCode == 503 {
		info["blocked"] = true
		info["reason"] = fmt.Sprintf("HTTP %d", result.StatusCode)
		return info
	}

	// Response content detection
	body := string(result.Body)

	// Common anti-crawling page signatures
	blockSignals := []string{
		"captcha",
		"verification",
		"access denied",
		"blocked",
		"rate limit",
		"too many requests",
		"security check",
		"challenge",
	}

	for _, signal := range blockSignals {
		if containsIgnoreCase(body, signal) {
			info["blocked"] = true
			info["reason"] = "block_page_detected"
			info["signal"] = signal
			return info
		}
	}

	// Detect anti-crawling JS code
	antiBotJS := []string{
		"_cf_bm", // Cloudflare
		"__cf_bm",
		"turnstile",  // Cloudflare Turnstile
		"recaptcha",  // Google reCAPTCHA
		"hcaptcha",   // hCaptcha
		"datadome",   // DataDome
		"perimeterx", // PerimeterX
		"akamai",     // Akamai
		"imperva",    // Imperva
	}

	for _, signal := range antiBotJS {
		if containsIgnoreCase(body, signal) {
			info["anti_bot_detected"] = signal
			break
		}
	}

	return info
}

// discoverLinks - Discover links and add to queue
func (c *Crawler) discoverLinks(result *CrawlResult, depth int) {
	// TODO: Implement HTML parsing and link discovery
	// Can use libraries like goquery
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(containsLower(s, substr) || containsLower(s, substr))
}

func containsLower(s, substr string) bool {
	// Simplified implementation, should use strings.Contains(strings.ToLower(s), strings.ToLower(substr))
	return true
}
