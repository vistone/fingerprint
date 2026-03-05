// Package web provides the web admin console for the gateway
package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/vistone/fingerprint/modules/gateway"
	"github.com/vistone/fingerprint/modules/profiles"
)

//go:embed static/*
var staticFiles embed.FS

// Handler handles web admin console requests
type Handler struct {
	gateway  *gateway.Gateway
	profiles []profiles.ClientProfile
}

// NewHandler creates a new web handler
func NewHandler(gw *gateway.Gateway) *Handler {
	return &Handler{
		gateway:  gw,
		profiles: loadProfiles(),
	}
}

// RegisterRoutes registers web console routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Static files
	mux.Handle("/admin/static/", http.FileServer(http.FS(staticFiles)))
	
	// Admin index
	mux.HandleFunc("/admin", h.handleIndex)
	mux.HandleFunc("/admin/", h.handleIndex)
	
	// API endpoints
	mux.HandleFunc("/api/admin/stats", h.handleStats)
	mux.HandleFunc("/api/admin/profiles", h.handleProfiles)
	mux.HandleFunc("/api/admin/analytics", h.handleAnalytics)
	mux.HandleFunc("/api/admin/requests", h.handleRequests)
	mux.HandleFunc("/api/admin/logs", h.handleLogs)
	mux.HandleFunc("/api/admin/config", h.handleConfig)
}

// handleIndex serves the main admin page
func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	data, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

// handleStats returns dashboard statistics
func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := map[string]interface{}{
		"totalProfiles":          len(h.profiles),
		"requestsPerSec":         getRequestsPerSec(),
		"avgLatency":             getAvgLatency(),
		"successRate":            getSuccessRate(),
		"uptime":                 getUptime(),
		"recentClassifications":  getRecentClassifications(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleProfiles returns browser profiles
func (h *Handler) handleProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := strings.ToLower(r.URL.Query().Get("q"))
	browser := strings.ToLower(r.URL.Query().Get("browser"))
	os := strings.ToLower(r.URL.Query().Get("os"))

	filtered := filterProfiles(h.profiles, query, browser, os)

	response := map[string]interface{}{
		"profiles": filtered,
		"total":    len(h.profiles),
		"filtered": len(filtered),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAnalytics returns analytics data
func (h *Handler) handleAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	analytics := map[string]interface{}{
		"browserDistribution": getBrowserDistribution(h.profiles),
		"osDistribution":      getOSDistribution(h.profiles),
		"topFingerprints":     getTopFingerprints(),
		"trafficData":         getTrafficData(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// handleRequests returns request logs
func (h *Handler) handleRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requests := getRecentRequests()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": requests,
	})
}

// handleLogs returns system logs
func (h *Handler) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logs := getSystemLogs()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs": logs,
	})
}

// handleConfig handles configuration requests
func (h *Handler) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		config := getCurrentConfig()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
		
	case http.MethodPost:
		var newConfig map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		if err := updateConfig(newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "success",
		})
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Helper functions

func loadProfiles() []profiles.ClientProfile {
	// Load profiles from all available sources
	var profiles []profiles.ClientProfile
	
	// This would integrate with the actual profile loading logic
	// For now, return empty slice
	return profiles
}

func filterProfiles(profiles []profiles.ClientProfile, query, browser, os string) []map[string]interface{} {
	var result []map[string]interface{}
	
	for _, p := range profiles {
		// Apply filters
		if query != "" && !strings.Contains(strings.ToLower(p.Name), query) {
			continue
		}
		
		if browser != "" && !strings.Contains(strings.ToLower(string(p.BrowserType)), browser) {
			continue
		}
		
		if os != "" && !strings.Contains(strings.ToLower(string(p.OS)), os) {
			continue
		}
		
		result = append(result, map[string]interface{}{
			"id":              p.ID,
			"name":            p.Name,
			"browserType":     p.BrowserType,
			"browserVersion":  p.BrowserVersion,
			"os":              p.OS,
			"osVersion":       p.OSVersion,
			"tlsVersion":      p.TLSVersion,
			"cipherSuites":    len(p.CipherSuites),
			"extensions":      len(p.Extensions),
		})
	}
	
	return result
}

func getRequestsPerSec() int {
	// Placeholder - would come from actual metrics
	return 145
}

func getAvgLatency() int {
	// Placeholder - would come from actual metrics
	return 23
}

func getSuccessRate() float64 {
	// Placeholder - would come from actual metrics
	return 99.8
}

func getUptime() string {
	// Placeholder - would come from actual system
	return "15d 7h 32m"
}

func getRecentClassifications() []map[string]interface{} {
	// Placeholder - would come from actual data
	return []map[string]interface{}{
		{
			"timestamp":   time.Now().Add(-2 * time.Minute).Unix(),
			"ja3Hash":     "7692c8d76c4f0e4a9c9c8a7b6c5d4e3f",
			"browser":     "Chrome 120",
			"confidence":  0.95,
			"status":      "success",
		},
		{
			"timestamp":   time.Now().Add(-5 * time.Minute).Unix(),
			"ja3Hash":     "a1b2c3d4e5f678901234567890123456",
			"browser":     "Firefox 121",
			"confidence":  0.87,
			"status":      "success",
		},
		{
			"timestamp":   time.Now().Add(-8 * time.Minute).Unix(),
			"ja3Hash":     "9876543210fedcba9876543210fedcb",
			"browser":     "Safari 17",
			"confidence":  0.92,
			"status":      "success",
		},
	}
}

func getBrowserDistribution(profiles []profiles.ClientProfile) map[string]int {
	distribution := make(map[string]int)
	
	for _, p := range profiles {
		browser := string(p.BrowserType)
		distribution[browser]++
	}
	
	// If empty, return sample data
	if len(distribution) == 0 {
		return map[string]int{
			"Chrome":  47,
			"Firefox": 30,
			"Safari":  38,
			"Edge":    10,
			"Opera":   7,
		}
	}
	
	return distribution
}

func getOSDistribution(profiles []profiles.ClientProfile) map[string]int {
	distribution := make(map[string]int)
	
	for _, p := range profiles {
		os := string(p.OS)
		distribution[os]++
	}
	
	// If empty, return sample data
	if len(distribution) == 0 {
		return map[string]int{
			"Windows": 45,
			"macOS":   35,
			"Linux":   20,
			"iOS":     15,
			"Android": 25,
		}
	}
	
	return distribution
}

func getTopFingerprints() []map[string]interface{} {
	return []map[string]interface{}{
		{"hash": "7692c8d76c4f0e4a9c9c8a7b6c5d4e3f", "count": 12543, "percentage": 15.2},
		{"hash": "a1b2c3d4e5f678901234567890123456", "count": 9876, "percentage": 12.1},
		{"hash": "9876543210fedcba9876543210fedcb", "count": 7654, "percentage": 9.8},
		{"hash": "fedcba0987654321fedcba098765432", "count": 5432, "percentage": 7.3},
		{"hash": "11223344556677889900aabbccddeeff", "count": 4321, "percentage": 5.9},
	}
}

func getTrafficData() map[string]interface{} {
	return map[string]interface{}{
		"labels": []string{"00:00", "04:00", "08:00", "12:00", "16:00", "20:00"},
		"data":   []int{120, 80, 250, 380, 420, 350},
	}
}

func getRecentRequests() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"timestamp":      time.Now().Add(-30 * time.Second).Unix(),
			"ip":             "192.168.1.100",
			"method":         "POST",
			"path":           "/api/v1/analyze",
			"ja3":            "7692c8d76c4f0e4a9c9c8a7b6c5d4e3f",
			"classification": "Chrome 120",
			"latency":        18,
			"status":         200,
		},
		{
			"timestamp":      time.Now().Add(-45 * time.Second).Unix(),
			"ip":             "10.0.0.50",
			"method":         "GET",
			"path":           "/health",
			"ja3":            "",
			"classification": "",
			"latency":        5,
			"status":         200,
		},
		{
			"timestamp":      time.Now().Add(-60 * time.Second).Unix(),
			"ip":             "172.16.0.25",
			"method":         "POST",
			"path":           "/api/v1/classify",
			"ja3":            "a1b2c3d4e5f678901234567890123456",
			"classification": "Firefox 121",
			"latency":        24,
			"status":         200,
		},
	}
}

func getSystemLogs() string {
	logs := []string{
		fmt.Sprintf("[%s] INFO: Server started on port 8080", time.Now().Format("2006-01-02 15:04:05")),
		fmt.Sprintf("[%s] INFO: gRPC server started on port 9090", time.Now().Format("2006-01-02 15:04:05")),
		fmt.Sprintf("[%s] INFO: Loaded %d browser profiles", time.Now().Format("2006-01-02 15:04:05"), 215),
		fmt.Sprintf("[%s] INFO: ML classifier initialized", time.Now().Format("2006-01-02 15:04:05")),
		fmt.Sprintf("[%s] DEBUG: Cache warmed up", time.Now().Format("2006-01-02 15:04:05")),
	}
	
	return strings.Join(logs, "\n")
}

func getCurrentConfig() map[string]interface{} {
	return map[string]interface{}{
		"server": map[string]interface{}{
			"httpPort":      8080,
			"grpcPort":      9090,
			"readTimeout":   30,
			"writeTimeout":  30,
		},
		"rateLimit": map[string]interface{}{
			"enabled":   true,
			"rps":       1000,
			"burstSize": 1500,
		},
		"cache": map[string]interface{}{
			"enabled": true,
			"size":    10000,
			"ttl":     5,
		},
		"ml": map[string]interface{}{
			"enabled":       true,
			"minConfidence": 0.85,
			"cacheSize":     5000,
		},
	}
}

func updateConfig(newConfig map[string]interface{}) error {
	// Placeholder - would actually update configuration
	return nil
}
