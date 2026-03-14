// Crawler and WAF Complete Integration Example - Testing Loop
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/crawler"
	"github.com/vistone/fingerprint/modules/waf"
)

// TrainingData structure for training data
type TrainingData struct {
	Timestamp   time.Time              `json:"timestamp"`
	Features    map[string]float64     `json:"features"`
	Label       int                    `json:"label"` // 1=crawler, 0=normal
	Fingerprint string                 `json:"fingerprint"`
	Blocked     bool                   `json:"blocked"`
	Detection   map[string]interface{} `json:"detection"`
}

// IntegrationDemo integration demonstration
type IntegrationDemo struct {
	waf          *waf.WAF
	crawler      *crawler.Crawler
	trainingData []TrainingData
	dataMu       sync.RWMutex
	blockCount   int
	allowCount   int
}

func main() {
	fmt.Println("🔗 Crawler + WAF Integration Demo")
	fmt.Println("==================================")

	demo := &IntegrationDemo{
		trainingData: make([]TrainingData, 0),
	}

	// 1. Start WAF
	demo.setupWAF()

	// 2. Start data collection HTTP service
	go demo.startDataCollector()

	// 3. Wait for WAF to be ready
	time.Sleep(2 * time.Second)

	// 4. Run crawler test
	demo.runCrawlerTest()

	// 5. Output results
	demo.printResults()
}

// setupWAF configures and starts WAF
func (d *IntegrationDemo) setupWAF() {
	config := &waf.WAFConfig{
		Enabled: true,
		Mode:    waf.WAFModeLearning, // Learning mode, collect data

		NetworkLayerEnabled:  true,
		TLSLayerEnabled:      true,
		HTTPLayerEnabled:     true,
		BehaviorLayerEnabled: true,

		RiskThreshold:  0.5,
		RateLimitRPS:   100,
		RateLimitBurst: 150,

		MLEnabled:    false,
		AgentEnabled: true,
	}

	d.waf = waf.NewWAF(config)

	// Start HTTP service
	mux := http.NewServeMux()

	// Protected endpoint
	mux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "success", "message": "data retrieved"}`))
	})

	// Simulate anti-crawl response
	mux.HandleFunc("/api/sensitive", func(w http.ResponseWriter, r *http.Request) {
		// Simulate partial blocking
		if time.Now().Unix()%3 == 0 {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "blocked"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "success", "data": "sensitive"}`))
	})

	handler := d.waf.Middleware(mux)

	go func() {
		log.Println("WAF server starting on :8080")
		http.ListenAndServe(":8080", handler)
	}()
}

// startDataCollector starts the data collector
func (d *IntegrationDemo) startDataCollector() {
	http.HandleFunc("/api/feedback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var entries []crawler.FeedbackEntry
		if err := json.NewDecoder(r.Body).Decode(&entries); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		d.dataMu.Lock()
		for _, entry := range entries {
			// Extract features
			features := extractFeatures(entry)

			// Assign label
			label := 0
			if entry.Blocked {
				label = 1
				d.blockCount++
			} else {
				d.allowCount++
			}

			d.trainingData = append(d.trainingData, TrainingData{
				Timestamp:   entry.Timestamp,
				Features:    features,
				Label:       label,
				Fingerprint: entry.Fingerprint,
				Blocked:     entry.Blocked,
				Detection:   entry.Detection,
			})
		}
		d.dataMu.Unlock()

		w.WriteHeader(http.StatusOK)
	})

	log.Println("Feedback collector starting on :9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}

// runCrawlerTest runs the crawler test
func (d *IntegrationDemo) runCrawlerTest() {
	config := &crawler.CrawlerConfig{
		Name: "integration_test",
		TargetURLs: []string{
			"http://localhost:8080/api/test",
			"http://localhost:8080/api/sensitive",
		},
		Workers:         3,
		RateLimit:       1 * time.Second,
		ProfileStrategy: crawler.ProfileStrategyRotate,
		ProfilePool: []string{
			"chrome_133_windows",
			"firefox_135_macos",
			"safari_17_macos",
		},
		CollectMode: crawler.CollectModeFull,
		FeedbackURL: "http://localhost:9090/api/feedback",
		StealthMode: true,
		HumanLike:   true,
	}

	d.crawler = crawler.NewCrawler(config)

	// Start crawler
	if err := d.crawler.Start(); err != nil {
		log.Printf("Crawler start error: %v", err)
		return
	}

	// Process results
	go func() {
		for result := range d.crawler.GetResults() {
			status := "✅"
			if result.Blocked {
				status = "🚫"
			}
			fmt.Printf("%s %s | Status: %d | Profile: %s\n",
				status,
				result.URL,
				result.StatusCode,
				result.Fingerprint.ID,
			)
		}
	}()

	// Run for a while
	fmt.Println("\n🕷️  Running crawler for 20 seconds...")
	time.Sleep(20 * time.Second)

	d.crawler.Stop()
}

// printResults outputs the results
func (d *IntegrationDemo) printResults() {
	fmt.Println("\n📊 Results Summary")
	fmt.Println("==================")

	// Crawler statistics
	stats := d.crawler.GetStats()
	fmt.Printf("\nCrawler Stats:\n")
	fmt.Printf("  Total Requests: %d\n", stats.TotalRequests.Load())
	fmt.Printf("  Success: %d (%.2f%%)\n",
		stats.SuccessRequests.Load(),
		stats.SuccessRate()*100)
	fmt.Printf("  Blocked: %d (%.2f%%)\n",
		stats.BlockedRequests.Load(),
		stats.BlockRate()*100)
	fmt.Printf("  Failed: %d\n", stats.FailedRequests.Load())

	// Training data statistics
	d.dataMu.RLock()
	fmt.Printf("\nTraining Data Collected:\n")
	fmt.Printf("  Total Samples: %d\n", len(d.trainingData))
	fmt.Printf("  Positive (Blocked): %d\n", d.blockCount)
	fmt.Printf("  Negative (Clean): %d\n", d.allowCount)
	d.dataMu.RUnlock()

	// WAF statistics
	wafStats := d.waf.Stats()
	fmt.Printf("\nWAF Stats:\n")
	fmt.Printf("  Total Requests: %d\n", wafStats.TotalRequests)
	fmt.Printf("  Allowed: %d\n", wafStats.AllowedRequests)
	fmt.Printf("  Blocked: %d\n", wafStats.BlockedRequests)
	fmt.Printf("  Challenged: %d\n", wafStats.ChallengedRequests)

	fmt.Println("\n✨ Integration demo completed!")
	fmt.Println("   The collected data can now be used to train the ML model.")
}

// extractFeatures extracts features from feedback entry
func extractFeatures(entry crawler.FeedbackEntry) map[string]float64 {
	features := make(map[string]float64)

	// Extract features from detection results
	if entry.Detection != nil {
		if v, ok := entry.Detection["anti_bot_detected"]; ok {
			features["anti_bot"] = 1.0
			_ = v
		}
	}

	// Time features
	hour := float64(entry.Timestamp.Hour())
	features["hour_of_day"] = hour / 24.0

	// Response type features
	if entry.ContentType != "" {
		if entry.ContentType == "application/json" {
			features["content_type_json"] = 1.0
		}
	}

	return features
}
