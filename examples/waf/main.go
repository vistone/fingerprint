// WAF Integration Example - Multi-layer Protection Demo
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/vistone/fingerprint/modules/waf"
)

func main() {
	fmt.Println("🛡️  Fingerprint WAF Demo")
	fmt.Println("========================")

	// Configure WAF
	config := &waf.WAFConfig{
		Enabled: true,
		Mode:    waf.WAFModeProtection,

		// Enable all protection layers
		NetworkLayerEnabled:  true,
		TLSLayerEnabled:      true,
		HTTPLayerEnabled:     true,
		BehaviorLayerEnabled: true,
		DeviceLayerEnabled:   true,

		// Threshold configuration
		RiskThreshold:  0.6,
		RateLimitRPS:   10,
		RateLimitBurst: 20,
		BlockDuration:  30 * time.Minute,
		ChallengeTTL:   10 * time.Minute,

		// Whitelist and blacklist
		WhitelistPaths: []string{"/health", "/api/public"},
		BlacklistPaths: []string{"/admin", "/internal"},
		BlacklistIPs:   []string{"192.168.1.100"},

		// JA3/JA4 blacklist (known crawler tool fingerprints)
		BlacklistJA3: []string{
			// curl
			"769,47-53-5-10-49161-49162-49171-49172-50-56-19-4,0-10-11,23-24-25,0",
			// python-requests
			"771,49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24-25-256-257,0",
		},

		// ML and Agent
		MLEnabled:    true,
		AgentEnabled: true,
	}

	// Create WAF
	w := waf.NewWAF(config)
	defer w.Stop()

	// Create business handler
	mux := http.NewServeMux()

	// Public endpoint
	mux.HandleFunc("/api/public", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok", "message": "public endpoint"}`))
	})

	// Protected endpoint
	mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok", "data": "sensitive data"}`))
	})

	// Admin endpoint (blacklisted)
	mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "admin"}`))
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Statistics
	mux.HandleFunc("/waf/stats", func(wr http.ResponseWriter, r *http.Request) {
		stats := w.Stats()
		wr.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(wr, `{
			"total": %d,
			"allowed": %d,
			"blocked": %d,
			"challenged": %d,
			"throttled": %d,
			"monitored": %d
		}`,
			stats.TotalRequests,
			stats.AllowedRequests,
			stats.BlockedRequests,
			stats.ChallengedRequests,
			stats.ThrottledRequests,
			stats.MonitoredRequests,
		)
	})

	// Wrap WAF middleware
	handler := w.Middleware(mux)

	fmt.Println("\n🌐 Server Configuration:")
	fmt.Printf("  Mode: %s\n", config.Mode)
	fmt.Printf("  Risk Threshold: %.2f\n", config.RiskThreshold)
	fmt.Printf("  Rate Limit: %d req/s\n", config.RateLimitRPS)
	fmt.Println("\n📍 Endpoints:")
	fmt.Println("  GET /health       - Health check (whitelisted)")
	fmt.Println("  GET /api/public   - Public API (whitelisted)")
	fmt.Println("  GET /api/data     - Protected API (WAF protected)")
	fmt.Println("  GET /admin        - Admin panel (blacklisted)")
	fmt.Println("  GET /waf/stats    - WAF statistics")
	fmt.Println("\n🚀 Starting server on :8080")
	fmt.Println("Press Ctrl+C to stop")

	log.Fatal(http.ListenAndServe(":8080", handler))
}
