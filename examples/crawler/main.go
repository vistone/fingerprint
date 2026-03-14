// Crawler Test Example - Active crawling and anti-detection testing
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/vistone/fingerprint/modules/crawler"
)

func main() {
	fmt.Println("🕷️  Fingerprint Crawler Test Suite")
	fmt.Println("====================================")

	// Scenario 1: Basic Test - Single profile single target
	fmt.Println("\n📋 Scenario 1: Basic Test")
	runBasicTest()

	// Scenario 2: Profile Rotation Test
	fmt.Println("\n📋 Scenario 2: Profile Rotation Test")
	runProfileRotationTest()

	// Scenario 3: Proxy Pool Test
	fmt.Println("\n📋 Scenario 3: Proxy Pool Test")
	runProxyTest()

	// Scenario 4: Feedback Loop Test
	fmt.Println("\n📋 Scenario 4: Feedback Loop Test")
	runFeedbackTest()
}

// Basic Test
func runBasicTest() {
	config := &crawler.CrawlerConfig{
		Name:            "basic_test",
		TargetURLs:      []string{"https://httpbin.org/get"},
		Workers:         1,
		RateLimit:       3 * time.Second,
		ProfileStrategy: crawler.ProfileStrategyRandom,
		ProfilePool: []string{
			"chrome_133_windows",
		},
		CollectMode: crawler.CollectModeFull,
		StealthMode: true,
	}

	c := crawler.NewCrawler(config)

	if err := c.Start(); err != nil {
		log.Printf("Start error: %v", err)
		return
	}

	// Wait for some results
	time.Sleep(10 * time.Second)
	c.Stop()

	stats := c.GetStats()
	fmt.Printf("  Total: %d | Success: %d | Blocked: %d\n",
		stats.TotalRequests.Load(),
		stats.SuccessRequests.Load(),
		stats.BlockedRequests.Load())
}

// Profile Rotation Test
func runProfileRotationTest() {
	config := &crawler.CrawlerConfig{
		Name: "rotation_test",
		TargetURLs: []string{
			"https://httpbin.org/user-agent",
		},
		Workers:         3,
		RateLimit:       2 * time.Second,
		ProfileStrategy: crawler.ProfileStrategyRotate,
		ProfilePool: []string{
			"chrome_133_windows",
			"firefox_135_macos",
			"safari_17_ios",
			"edge_120_windows",
		},
		RotateInterval: 5 * time.Second,
		HumanLike:      true,
		StealthMode:    true,
	}

	c := crawler.NewCrawler(config)

	if err := c.Start(); err != nil {
		log.Printf("Start error: %v", err)
		return
	}

	// Process results
	go func() {
		for result := range c.GetResults() {
			if result.Fingerprint != nil {
				fmt.Printf("  Request: %s | Profile: %s | Status: %d\n",
					result.URL,
					result.Fingerprint.ID,
					result.StatusCode)
			}
		}
	}()

	time.Sleep(15 * time.Second)
	c.Stop()
}

// Proxy Pool Test
func runProxyTest() {
	config := &crawler.CrawlerConfig{
		Name:            "proxy_test",
		TargetURLs:      []string{"https://httpbin.org/ip"},
		Workers:         2,
		RateLimit:       3 * time.Second,
		ProfileStrategy: crawler.ProfileStrategySticky,
		ProxyStrategy:   crawler.ProxyStrategyRotate,
		// Use example proxies here, actual valid proxies need to be configured
		ProxyList: []string{
			// "http://proxy1:8080",
			// "http://proxy2:8080",
		},
		CollectMode: crawler.CollectModeResponse,
	}

	c := crawler.NewCrawler(config)

	if err := c.Start(); err != nil {
		log.Printf("Start error: %v", err)
		return
	}

	time.Sleep(8 * time.Second)
	c.Stop()

	stats := c.GetStats()
	fmt.Printf("  Total: %d | Block Rate: %.2f%%\n",
		stats.TotalRequests.Load(),
		stats.BlockRate()*100)
}

// Feedback Loop Test
func runFeedbackTest() {
	config := &crawler.CrawlerConfig{
		Name: "feedback_test",
		TargetURLs: []string{
			"https://httpbin.org/get",
			"https://httpbin.org/headers",
		},
		Workers:         2,
		RateLimit:       2 * time.Second,
		ProfileStrategy: crawler.ProfileStrategyRandom,
		CollectMode:     crawler.CollectModeBlocked,
		// Data feedback endpoint
		FeedbackURL: "http://localhost:8080/api/feedback",
	}

	c := crawler.NewCrawler(config)

	if err := c.Start(); err != nil {
		log.Printf("Start error: %v", err)
		return
	}

	// Collect blocked data
	blockedCount := 0
	go func() {
		for result := range c.GetResults() {
			if result.Blocked {
				blockedCount++
				fmt.Printf("  🚫 Blocked: %s - %s\n", result.URL, result.BlockReason)
			}
		}
	}()

	time.Sleep(12 * time.Second)
	c.Stop()

	fmt.Printf("  Total blocked collected: %d\n", blockedCount)
}
