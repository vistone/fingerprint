// Package web provides the web admin console for the gateway
package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/gateway"
	"github.com/vistone/fingerprint/modules/profiles"
)

//go:embed static/index.html
//go:embed static/css/*
//go:embed static/js/*
//go:embed static/img/*
var staticFiles embed.FS

// RequestRecord stores a single API request record.
type RequestRecord struct {
	Timestamp      time.Time
	IP             string
	Method         string
	Path           string
	JA3            string
	Classification string
	Latency        int64
	Status         int
}

// MetricsStore keeps in-memory runtime metrics.
type MetricsStore struct {
	sync.RWMutex
	startTime             time.Time
	totalRequests         int64
	successfulRequests    int64
	totalLatency          int64 // milliseconds
	recentClassifications []map[string]interface{}
	requestCounts         []int64 // historical requests-per-second data
	lastRequestTimes      []time.Time
	recentRequests        []RequestRecord // recent request records
}

var globalMetrics = &MetricsStore{
	startTime:             time.Now(),
	recentClassifications: make([]map[string]interface{}, 0, 10),
	requestCounts:         make([]int64, 60), // 60-second history
	lastRequestTimes:      make([]time.Time, 0, 100),
	recentRequests:        make([]RequestRecord, 0, 100),
}

// RecordAPIMetrics records metrics for a public API request.
func RecordAPIMetrics(req RequestRecord, success bool, browser, ja3 string) {
	globalMetrics.RecordRequest(req, success)
	if browser != "" {
		globalMetrics.RecordClassification(browser, ja3)
	}
}

// GetRecentRequests returns recent request records.
func GetRecentRequests() []RequestRecord {
	globalMetrics.RLock()
	defer globalMetrics.RUnlock()

	result := make([]RequestRecord, len(globalMetrics.recentRequests))
	copy(result, globalMetrics.recentRequests)
	return result
}

// RecordRequest records one request and updates aggregates.
func (m *MetricsStore) RecordRequest(req RequestRecord, success bool) {
	m.Lock()
	defer m.Unlock()

	m.totalRequests++
	m.totalLatency += req.Latency
	if success {
		m.successfulRequests++
	}

	// Track request times for RPS calculation.
	m.lastRequestTimes = append(m.lastRequestTimes, req.Timestamp)
	// Keep only request times from the last 60 seconds.
	cutoff := time.Now().Add(-60 * time.Second)
	idx := 0
	for i, t := range m.lastRequestTimes {
		if t.After(cutoff) {
			idx = i
			break
		}
	}
	if idx > 0 {
		m.lastRequestTimes = m.lastRequestTimes[idx:]
	}

	// Keep a reverse-chronological request list.
	m.recentRequests = append([]RequestRecord{req}, m.recentRequests...)
	// Keep at most 100 request records.
	if len(m.recentRequests) > 100 {
		m.recentRequests = m.recentRequests[:100]
	}
}

// RecordClassification records one classification result.
func (m *MetricsStore) RecordClassification(browser, ja3 string) {
	m.Lock()
	defer m.Unlock()

	classification := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"ja3Hash":   ja3,
		"browser":   browser,
		"status":    "success",
	}

	m.recentClassifications = append([]map[string]interface{}{classification}, m.recentClassifications...)
	if len(m.recentClassifications) > 10 {
		m.recentClassifications = m.recentClassifications[:10]
	}
}

// GetMetrics returns current aggregated metrics.
func (m *MetricsStore) GetMetrics() (requestsPerSec int, avgLatency int, successRate float64, uptime string, recent []map[string]interface{}) {
	m.RLock()
	defer m.RUnlock()

	// RPS is computed from request count in the last 60 seconds.
	requestsPerSec = len(m.lastRequestTimes)

	// Compute average latency.
	if m.totalRequests > 0 {
		avgLatency = int(m.totalLatency / m.totalRequests)
	}

	// Compute success rate.
	if m.totalRequests > 0 {
		successRate = float64(m.successfulRequests) / float64(m.totalRequests) * 100
	}

	// Compute process uptime.
	uptime = formatDuration(time.Since(m.startTime))

	// Copy recent classifications for safe read access.
	recent = make([]map[string]interface{}, len(m.recentClassifications))
	copy(recent, m.recentClassifications)

	return
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

// Handler handles web admin console requests
type Handler struct {
	gateway  *gateway.Gateway
	profiles []profiles.ClientProfile
	mu       sync.RWMutex

	// async training state
	trainingMu     sync.Mutex
	trainingActive bool
	trainingPhase  string // "train" or "evolve"
	trainingStart  time.Time
	trainingErr    string
	trainingDone   bool
	trainingResult map[string]interface{}
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
	// Static files - extract "static" subdirectory from embed.FS
	staticSubFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Printf("admin static resources unavailable: %v", err)
		mux.HandleFunc("/admin/static/", func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		})
	} else {
		mux.Handle("/admin/static/", http.StripPrefix("/admin/static/", http.FileServer(http.FS(staticSubFS))))
	}

	// Admin index
	mux.HandleFunc("/admin", h.handleIndex)
	mux.HandleFunc("/admin/", h.handleIndex)

	// API endpoints
	mux.HandleFunc("/api/admin/stats", h.handleStats)
	mux.HandleFunc("/api/admin/profiles", h.handleProfiles)
	mux.HandleFunc("/api/admin/profiles/", h.handleProfileDetail)
	mux.HandleFunc("/api/admin/analytics", h.handleAnalytics)
	mux.HandleFunc("/api/admin/requests", h.handleRequests)
	mux.HandleFunc("/api/admin/logs", h.handleLogs)
	mux.HandleFunc("/api/admin/logs/stream", h.handleLogStream)
	mux.HandleFunc("/api/admin/config", h.handleConfig)
	mux.HandleFunc("/api/admin/client/test", h.handleClientTest)

	// Agent API endpoints
	mux.HandleFunc("/api/admin/agent/status", h.handleAgentStatus)
	mux.HandleFunc("/api/admin/agent/knowledge", h.handleAgentKnowledge)
	mux.HandleFunc("/api/admin/agent/strategies", h.handleAgentStrategies)
	mux.HandleFunc("/api/admin/crawler/status", h.handleCrawlerStatus)
	mux.HandleFunc("/api/admin/waf/status", h.handleWAFStatus)

	// Analysis engine endpoints
	mux.HandleFunc("/api/admin/analyze/profile", h.handleAnalyzeProfile)

	// ML engine endpoints
	mux.HandleFunc("/api/admin/ml/info", h.handleMLInfo)
	mux.HandleFunc("/api/admin/ml/extract", h.handleMLExtract)
	mux.HandleFunc("/api/admin/ml/classify", h.handleMLClassify)
	mux.HandleFunc("/api/admin/ml/batch", h.handleMLBatch)

	// MLService endpoints
	mux.HandleFunc("/api/admin/ml/service/stats", h.handleMLServiceStats)
	mux.HandleFunc("/api/admin/ml/service/health", h.handleMLServiceHealth)
	mux.HandleFunc("/api/admin/ml/service/infer", h.handleMLServiceInfer)
	mux.HandleFunc("/api/admin/ml/service/validate", h.handleMLServiceValidate)
	mux.HandleFunc("/api/admin/ml/service/generate", h.handleMLServiceGenerate)
	mux.HandleFunc("/api/admin/ml/service/evolve", h.handleMLServiceEvolve)
	mux.HandleFunc("/api/admin/ml/service/train", h.handleMLServiceTrain)
	mux.HandleFunc("/api/admin/ml/service/training-status", h.handleMLServiceTrainingStatus)
	mux.HandleFunc("/api/admin/ml/service/feedback", h.handleMLServiceFeedback)

	// Defense endpoints
	mux.HandleFunc("/api/admin/defense/rules", h.handleDefenseRules)
	mux.HandleFunc("/api/admin/defense/detect", h.handleDefenseDetect)

	// Anti-detection endpoints
	mux.HandleFunc("/api/admin/antidetect/status", h.handleAntiDetectStatus)
	mux.HandleFunc("/api/admin/antidetect/preview", h.handleAntiDetectPreview)
	mux.HandleFunc("/api/admin/antidetect/inject", h.handleAntiDetectInjectTest)
	mux.HandleFunc("/api/admin/antidetect/sdk", h.handleAntiDetectSDKPreview)

	// Plugin endpoints
	mux.HandleFunc("/api/admin/plugins/info", h.handlePluginsInfo)

	// Fingerprint tool endpoints
	mux.HandleFunc("/api/admin/tools/ja3", h.handleToolsJA3)
	mux.HandleFunc("/api/admin/tools/validate", h.handleToolsValidate)
	mux.HandleFunc("/api/admin/tools/compare", h.handleToolsCompare)
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

	// Read live metrics.
	rps, latency, rate, uptime, recent := globalMetrics.GetMetrics()
	h.mu.RLock()
	totalProfiles := len(h.profiles)
	h.mu.RUnlock()
	// Keep stable defaults when no live data is available.
	if rps == 0 {
		rps = 0
	}
	if latency == 0 {
		latency = 0
	}
	if rate == 0 {
		rate = 100.0
	}
	if len(recent) == 0 {
		recent = []map[string]interface{}{}
	}

	stats := map[string]interface{}{
		"totalProfiles":         totalProfiles,
		"requestsPerSec":        rps,
		"avgLatency":            latency,
		"successRate":           rate,
		"uptime":                uptime,
		"recentClassifications": recent,
	}

	// Attach agent status.
	if a := h.gateway.GetAgent(); a != nil {
		agentStats := a.Stats()
		stats["agent"] = map[string]interface{}{
			"enabled":           true,
			"activeSessions":    agentStats.ActiveSessions,
			"totalObservations": agentStats.TotalObservations,
			"activeStrategies":  agentStats.ActiveStrategies,
			"learnedPatterns":   agentStats.LearnedPatterns,
		}
	} else {
		stats["agent"] = map[string]interface{}{"enabled": false}
	}

	// Attach ML service status.
	if svc := h.gateway.GetMLService(); svc != nil {
		svcStats := svc.Stats()
		mlSvc := map[string]interface{}{
			"enabled":       true,
			"ready":         svc.IsReady(),
			"inferCount":    svcStats.InferCount,
			"feedbackCount": svcStats.FeedbackCount,
			"evolveCount":   svcStats.EvolveCount,
			"modelVersions": svcStats.ModelVersions,
		}
		if svcStats.LearnerStats != nil {
			mlSvc["learner"] = map[string]interface{}{
				"totalSamples":    svcStats.LearnerStats.TotalSamples,
				"bufferFilled":    svcStats.LearnerStats.BufferFilled,
				"peakAccuracy":    svcStats.LearnerStats.PeakAccuracy,
				"recentAccuracy":  svcStats.LearnerStats.RecentAccuracy,
				"driftDetected":   svcStats.LearnerStats.DriftDetected,
				"driftEventCount": svcStats.LearnerStats.DriftEventCount,
			}
		}
		stats["mlService"] = mlSvc
	} else {
		stats["mlService"] = map[string]interface{}{"enabled": false, "ready": false}
	}

	// Attach runtime component status.
	cfg := h.gateway.GetConfig()
	mlServiceReady := false
	if svc := h.gateway.GetMLService(); svc != nil {
		mlServiceReady = svc.IsReady()
	}
	stats["systemStatus"] = map[string]interface{}{
		"apiServer":         true,
		"mlClassifier":      true,
		"cache":             cfg.CacheEnabled,
		"agent":             cfg.AgentEnabled,
		"antiDetectEnabled": cfg.AntiDetectEnabled,
		"scanner":           cfg.ScannerUseBrowser,
		"mlServiceEnabled":  cfg.MLServiceEnabled,
		"mlServiceReady":    mlServiceReady,
	}
	h.appendRuntimeStats(stats)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleProfiles returns browser profiles
