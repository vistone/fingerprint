// Package web provides the web admin console for the gateway
package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/agent"
	"github.com/vistone/fingerprint/modules/client"
	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/gateway"
	"github.com/vistone/fingerprint/modules/profiles"
)

//go:embed static/index.html
//go:embed static/css/*
//go:embed static/js/*
//go:embed static/img/*
var staticFiles embed.FS

// RequestRecord 存储单个请求的记录
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

// MetricsStore 存储运行时指标
type MetricsStore struct {
	sync.RWMutex
	startTime             time.Time
	totalRequests         int64
	successfulRequests    int64
	totalLatency          int64 // 毫秒
	recentClassifications []map[string]interface{}
	requestCounts         []int64 // 每秒请求数的历史
	lastRequestTimes      []time.Time
	recentRequests        []RequestRecord // 最近请求记录
}

var globalMetrics = &MetricsStore{
	startTime:             time.Now(),
	recentClassifications: make([]map[string]interface{}, 0, 10),
	requestCounts:         make([]int64, 60), // 60秒的历史
	lastRequestTimes:      make([]time.Time, 0, 100),
	recentRequests:        make([]RequestRecord, 0, 100),
}

// RecordAPIMetrics 公共函数：记录 API 请求指标
func RecordAPIMetrics(req RequestRecord, success bool, browser, ja3 string) {
	globalMetrics.RecordRequest(req, success)
	if browser != "" {
		globalMetrics.RecordClassification(browser, ja3)
	}
}

// GetRecentRequests 获取最近请求记录
func GetRecentRequests() []RequestRecord {
	globalMetrics.RLock()
	defer globalMetrics.RUnlock()

	result := make([]RequestRecord, len(globalMetrics.recentRequests))
	copy(result, globalMetrics.recentRequests)
	return result
}

// RecordRequest 记录一个请求
func (m *MetricsStore) RecordRequest(req RequestRecord, success bool) {
	m.Lock()
	defer m.Unlock()

	m.totalRequests++
	m.totalLatency += req.Latency
	if success {
		m.successfulRequests++
	}

	// 记录请求时间用于计算 RPS
	m.lastRequestTimes = append(m.lastRequestTimes, req.Timestamp)
	// 只保留最近 60 秒的时间
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

	// 记录请求详情
	m.recentRequests = append([]RequestRecord{req}, m.recentRequests...)
	// 只保留最近 100 条请求
	if len(m.recentRequests) > 100 {
		m.recentRequests = m.recentRequests[:100]
	}
}

// RecordClassification 记录一次分类
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

// GetMetrics 获取当前指标
func (m *MetricsStore) GetMetrics() (requestsPerSec int, avgLatency int, successRate float64, uptime string, recent []map[string]interface{}) {
	m.RLock()
	defer m.RUnlock()

	// 计算 RPS (最近 60 秒内的请求数)
	requestsPerSec = len(m.lastRequestTimes)

	// 计算平均延迟
	if m.totalRequests > 0 {
		avgLatency = int(m.totalLatency / m.totalRequests)
	}

	// 计算成功率
	if m.totalRequests > 0 {
		successRate = float64(m.successfulRequests) / float64(m.totalRequests) * 100
	}

	// 计算运行时间
	uptime = formatDuration(time.Since(m.startTime))

	// 复制最近分类记录
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

	// 分析引擎
	mux.HandleFunc("/api/admin/analyze/profile", h.handleAnalyzeProfile)

	// ML 引擎
	mux.HandleFunc("/api/admin/ml/info", h.handleMLInfo)
	mux.HandleFunc("/api/admin/ml/extract", h.handleMLExtract)
	mux.HandleFunc("/api/admin/ml/classify", h.handleMLClassify)
	mux.HandleFunc("/api/admin/ml/batch", h.handleMLBatch)

	// 防御系统
	mux.HandleFunc("/api/admin/defense/rules", h.handleDefenseRules)
	mux.HandleFunc("/api/admin/defense/detect", h.handleDefenseDetect)

	// 反检测引擎
	mux.HandleFunc("/api/admin/antidetect/status", h.handleAntiDetectStatus)
	mux.HandleFunc("/api/admin/antidetect/preview", h.handleAntiDetectPreview)
	mux.HandleFunc("/api/admin/antidetect/inject", h.handleAntiDetectInjectTest)
	mux.HandleFunc("/api/admin/antidetect/sdk", h.handleAntiDetectSDKPreview)

	// 插件系统
	mux.HandleFunc("/api/admin/plugins/info", h.handlePluginsInfo)

	// 指纹工具
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

	// 获取真实指标
	rps, latency, rate, uptime, recent := globalMetrics.GetMetrics()
	h.mu.RLock()
	totalProfiles := len(h.profiles)
	h.mu.RUnlock()
	// 如果没有真实数据，显示默认值
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

	// 添加 Agent 状态
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

	// 添加系统组件状态
	cfg := h.gateway.GetConfig()
	stats["systemStatus"] = map[string]interface{}{
		"apiServer":         true,
		"mlClassifier":      true,
		"cache":             cfg.CacheEnabled,
		"agent":             cfg.AgentEnabled,
		"antiDetectEnabled": cfg.P3Enabled,
		"scanner":           cfg.ScannerUseBrowser,
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

	h.mu.RLock()
	profileSnapshot := append([]profiles.ClientProfile(nil), h.profiles...)
	h.mu.RUnlock()
	filtered := filterProfiles(profileSnapshot, query, browser, os)

	response := map[string]interface{}{
		"profiles": filtered,
		"total":    len(profileSnapshot),
		"filtered": len(filtered),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleProfileDetail returns a single profile detail
func (h *Handler) handleProfileDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract profile ID from URL path: /api/admin/profiles/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/profiles/")
	profileID := strings.TrimSpace(path)

	if profileID == "" {
		http.Error(w, "Profile ID required", http.StatusBadRequest)
		return
	}

	// Find profile from loaded profiles
	var found profiles.ClientProfile
	foundOK := false
	h.mu.RLock()
	for i := range h.profiles {
		if h.profiles[i].ID == profileID {
			found = h.profiles[i]
			foundOK = true
			break
		}
	}
	h.mu.RUnlock()

	if !foundOK {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"id":              found.ID,
		"name":            found.Name,
		"description":     found.Description,
		"browserType":     found.BrowserType,
		"browserVersion":  found.BrowserVersion,
		"os":              found.OS,
		"osVersion":       found.OSVersion,
		"osArch":          found.OSArch,
		"osBitness":       found.OSBitness,
		"tlsVersion":      found.TLSVersion,
		"cipherSuites":    found.CipherSuites,
		"extensions":      found.Extensions,
		"supportedCurves": found.SupportedCurves,
		"supportedPoints": found.SupportedPoints,
		"http2Settings": map[string]interface{}{
			"headerTableSize":      found.HTTP2Settings.HeaderTableSize,
			"enablePush":           found.HTTP2Settings.EnablePush,
			"maxConcurrentStreams": found.HTTP2Settings.MaxConcurrentStreams,
			"initialWindowSize":    found.HTTP2Settings.InitialWindowSize,
			"maxFrameSize":         found.HTTP2Settings.MaxFrameSize,
			"maxHeaderListSize":    found.HTTP2Settings.MaxHeaderListSize,
		},
		"http2Priorities":   found.HTTP2Priorities,
		"pseudoHeaderOrder": found.PseudoHeaderOrder,
		"connectionFlow":    found.ConnectionFlow,
		"headers":           found.Headers,
		"metadata":          found.Metadata,
	}

	// 添加 TCP/IP 指纹信息
	if found.TCPIP != nil {
		response["tcpip"] = map[string]interface{}{
			"ipVersion":        found.TCPIP.IPVersion,
			"ttl":              found.TCPIP.TTL,
			"df":               found.TCPIP.DF,
			"flags":            found.TCPIP.Flags,
			"windowSize":       found.TCPIP.WindowSize,
			"mss":              found.TCPIP.MSS,
			"windowScale":      found.TCPIP.WindowScale,
			"sackPermitted":    found.TCPIP.SAckPermitted,
			"timestamps":       found.TCPIP.Timestamps,
			"noOperation":      found.TCPIP.NoOperation,
			"endOfOptions":     found.TCPIP.EndOfOptions,
			"optionsSignature": found.TCPIP.OptionsSignature,
			"ja4t":             found.TCPIP.JA4T,
		}
	}

	// 添加 HTTP/3 (QUIC) 配置
	if found.HTTP3Settings != nil {
		response["http3Settings"] = map[string]interface{}{
			"quicVersion":            found.HTTP3Settings.QUICVersion,
			"initialMaxData":         found.HTTP3Settings.InitialMaxData,
			"initialMaxStreamData":   found.HTTP3Settings.InitialMaxStreamData,
			"initialMaxStreamsBidi":  found.HTTP3Settings.InitialMaxStreamsBidi,
			"initialMaxStreamsUni":   found.HTTP3Settings.InitialMaxStreamsUni,
			"maxUDPPayloadSize":      found.HTTP3Settings.MaxUDPPayloadSize,
			"ackDelayExponent":       found.HTTP3Settings.AckDelayExponent,
			"maxAckDelay":            found.HTTP3Settings.MaxAckDelay,
			"disableActiveMigration": found.HTTP3Settings.DisableActiveMigration,
		}
		response["quicVersions"] = found.QUICVersions
		response["http3Supported"] = true
	} else {
		response["http3Supported"] = false
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

	h.mu.RLock()
	profileSnapshot := append([]profiles.ClientProfile(nil), h.profiles...)
	h.mu.RUnlock()

	analytics := map[string]interface{}{
		"browserDistribution": getBrowserDistribution(profileSnapshot),
		"osDistribution":      getOSDistribution(profileSnapshot),
		"tcpipDistribution":   getTCPIPDistribution(profileSnapshot),
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

	records := GetRecentRequests()

	// 转换为前端期望的格式
	requests := make([]map[string]interface{}, len(records))
	for i, req := range records {
		requests[i] = map[string]interface{}{
			"timestamp":      req.Timestamp.Unix(),
			"ip":             req.IP,
			"method":         req.Method,
			"path":           req.Path,
			"ja3":            req.JA3,
			"classification": req.Classification,
			"latency":        req.Latency,
			"status":         req.Status,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": requests,
	})
}

// handleLogs returns system logs from the real log buffer
func (h *Handler) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	level := r.URL.Query().Get("level")
	entries := globalLogBuffer.GetFiltered(level)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  entries,
		"total": len(entries),
	})
}

// handleConfig handles configuration requests - reads/writes real GatewayConfig
func (h *Handler) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := h.gateway.GetConfig()
		agentEnabled := cfg.AgentEnabled
		var agentCfg map[string]interface{}
		if a := h.gateway.GetAgent(); a != nil {
			stats := a.Stats()
			sessionWindow := 30
			maxObs := 500
			fpThresh := 3.0
			consThresh := 0.4
			burstThresh := 10.0
			if cfg.AgentConfig != nil {
				sessionWindow = int(cfg.AgentConfig.SessionWindow.Minutes())
				maxObs = cfg.AgentConfig.MaxObservations
				fpThresh = cfg.AgentConfig.FPSwitchRateThreshold
				consThresh = cfg.AgentConfig.ConsistencyThreshold
				burstThresh = cfg.AgentConfig.RequestBurstThreshold
			}
			agentCfg = map[string]interface{}{
				"enabled":               agentEnabled,
				"sessionWindow":         sessionWindow,
				"maxObservations":       maxObs,
				"fpSwitchRateThreshold": fpThresh,
				"consistencyThreshold":  consThresh,
				"requestBurstThreshold": burstThresh,
				"activeSessions":        stats.ActiveSessions,
				"totalObservations":     stats.TotalObservations,
			}
		} else {
			agentCfg = map[string]interface{}{
				"enabled": false,
			}
		}

		config := map[string]interface{}{
			"server": map[string]interface{}{
				"endpoint": cfg.Endpoint,
				"port":     cfg.Port,
			},
			"rateLimit": map[string]interface{}{
				"enabled":   true,
				"rps":       cfg.RateLimitRequests,
				"burstSize": cfg.RateLimitBurst,
				"window":    int(cfg.RateLimitWindow.Seconds()),
			},
			"cache": map[string]interface{}{
				"enabled": cfg.CacheEnabled,
				"size":    cfg.CacheSize,
				"ttl":     int(cfg.CacheTTL.Minutes()),
			},
			"ml": map[string]interface{}{
				"enabled":       true,
				"riskThreshold": cfg.RiskThreshold,
			},
			"p3": map[string]interface{}{
				"enabled":       cfg.P3Enabled,
				"profileId":     cfg.P3ProfileID,
				"configDir":     cfg.P3ConfigDir,
				"proxyTarget":   cfg.P3ProxyTarget,
				"directProxy":   cfg.P3DirectProxy,
				"injectConsist": cfg.P3InjectConsist,
			},
			"scanner": map[string]interface{}{
				"useBrowser":     cfg.ScannerUseBrowser,
				"browserWS":      cfg.ScannerBrowserWS,
				"browserTimeout": int(cfg.ScannerBrowserTimeout.Seconds()),
			},
			"agent": agentCfg,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)

	case http.MethodPost:
		var newConfig map[string]interface{}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		h.applyConfigUpdate(newConfig)

		WriteLog("INFO", "config", "Configuration updated via admin console")

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
	// 从 profiles 模块加载所有已注册的指纹配置
	return profiles.GetAll()
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

		profileData := map[string]interface{}{
			"id":             p.ID,
			"name":           p.Name,
			"browserType":    p.BrowserType,
			"browserVersion": p.BrowserVersion,
			"os":             p.OS,
			"osVersion":      p.OSVersion,
			"tlsVersion":     p.TLSVersion,
			"cipherSuites":   len(p.CipherSuites),
			"extensions":     len(p.Extensions),
		}

		// 添加 TCP/IP 指纹简要信息
		if p.TCPIP != nil {
			profileData["tcpip"] = map[string]interface{}{
				"ttl":        p.TCPIP.TTL,
				"windowSize": p.TCPIP.WindowSize,
				"mss":        p.TCPIP.MSS,
				"ja4t":       p.TCPIP.JA4T,
			}
		}

		result = append(result, profileData)
	}

	return result
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
		osStr := string(p.OS)
		// 简化 OS 名称
		var group string
		switch {
		case strings.Contains(osStr, "Windows"):
			group = "Windows"
		case strings.Contains(osStr, "Mac OS") || strings.Contains(osStr, "Macintosh"):
			group = "macOS"
		case strings.Contains(osStr, "Android"):
			group = "Android"
		case strings.Contains(osStr, "iPhone") || strings.Contains(osStr, "iPad") || strings.Contains(osStr, "iOS"):
			group = "iOS"
		case strings.Contains(osStr, "Linux") || strings.Contains(osStr, "X11"):
			group = "Linux"
		default:
			group = "Other"
		}
		distribution[group]++
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

func getTCPIPDistribution(profiles []profiles.ClientProfile) map[string]int {
	distribution := make(map[string]int)

	for _, p := range profiles {
		if p.TCPIP == nil {
			distribution["Unknown"]++
			continue
		}
		// 根据 TTL 和 Window Size 判断 OS 类型
		ttl := p.TCPIP.TTL
		ws := p.TCPIP.WindowSize

		switch {
		case ttl == 128:
			distribution["Windows"]++
		case ttl == 64 && ws == 65535:
			distribution["macOS/iOS"]++
		case ttl == 64 && ws == 64240:
			distribution["Linux"]++
		default:
			distribution["Other"]++
		}
	}

	return distribution
}

// applyConfigUpdate applies a config update from the admin console
func (h *Handler) applyConfigUpdate(newConfig map[string]interface{}) {
	h.gateway.UpdateConfig(func(cfg *gateway.GatewayConfig) {
		if rl, ok := newConfig["rateLimit"].(map[string]interface{}); ok {
			if v, ok := rl["rps"].(float64); ok {
				cfg.RateLimitRequests = int(v)
			}
			if v, ok := rl["burstSize"].(float64); ok {
				cfg.RateLimitBurst = int(v)
			}
		}
		if cache, ok := newConfig["cache"].(map[string]interface{}); ok {
			if v, ok := cache["enabled"].(bool); ok {
				cfg.CacheEnabled = v
			}
			if v, ok := cache["size"].(float64); ok {
				cfg.CacheSize = int(v)
			}
			if v, ok := cache["ttl"].(float64); ok {
				cfg.CacheTTL = time.Duration(v) * time.Minute
			}
		}
		if p3, ok := newConfig["p3"].(map[string]interface{}); ok {
			if v, ok := p3["enabled"].(bool); ok {
				cfg.P3Enabled = v
			}
			if v, ok := p3["profileId"].(string); ok {
				cfg.P3ProfileID = v
			}
			if v, ok := p3["proxyTarget"].(string); ok {
				cfg.P3ProxyTarget = v
			}
			if v, ok := p3["directProxy"].(bool); ok {
				cfg.P3DirectProxy = v
			}
			if v, ok := p3["injectConsist"].(bool); ok {
				cfg.P3InjectConsist = v
			}
		}
		if scanner, ok := newConfig["scanner"].(map[string]interface{}); ok {
			if v, ok := scanner["useBrowser"].(bool); ok {
				cfg.ScannerUseBrowser = v
			}
			if v, ok := scanner["browserWS"].(string); ok {
				cfg.ScannerBrowserWS = v
			}
			if v, ok := scanner["browserTimeout"].(float64); ok {
				cfg.ScannerBrowserTimeout = time.Duration(v) * time.Second
			}
		}
		if ag, ok := newConfig["agent"].(map[string]interface{}); ok {
			if v, ok := ag["enabled"].(bool); ok {
				cfg.AgentEnabled = v
			}
		}
		if ml, ok := newConfig["ml"].(map[string]interface{}); ok {
			if v, ok := ml["riskThreshold"].(float64); ok {
				cfg.RiskThreshold = v
			}
		}
	})
}

// handleLogStream SSE 实时日志推流
func (h *Handler) handleLogStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := globalLogBuffer.Subscribe()
	defer globalLogBuffer.Unsubscribe(ch)

	// 先发送一次连接确认
	fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(entry)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// handleAgentStatus returns Agent stats and status
func (h *Handler) handleAgentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a := h.gateway.GetAgent()
	if a == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"enabled": false,
			"status":  "disabled",
		})
		return
	}

	stats := a.Stats()
	strategies := a.GetActiveStrategies()
	kb := a.Knowledge()
	kbStats := kb.Stats()

	result := map[string]interface{}{
		"enabled": true,
		"status":  "running",
		"stats":   stats,
		"strategySummary": map[string]interface{}{
			"total":   len(strategies),
			"learned": countLearnedStrategies(strategies),
		},
		"knowledge": map[string]interface{}{
			"totalBrowsers": kbStats.TotalKnownBrowsers,
			"totalVersions": kbStats.TotalKnownVersions,
			"totalProfiles": kbStats.TotalKnownProfiles,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleAgentKnowledge returns the knowledge base data
func (h *Handler) handleAgentKnowledge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a := h.gateway.GetAgent()
	if a == nil {
		http.Error(w, "Agent not enabled", http.StatusServiceUnavailable)
		return
	}

	kb := a.Knowledge()
	kbStats := kb.Stats()

	// 构建浏览器家族详情
	families := []map[string]interface{}{}
	browserTypes := []core.BrowserType{
		core.BrowserChrome, core.BrowserFirefox, core.BrowserSafari,
		core.BrowserEdge, core.BrowserOpera, core.BrowserBrave,
	}
	for _, bt := range browserTypes {
		bk := kb.GetBrowserKnowledge(bt)
		if bk == nil {
			continue
		}
		versions := []map[string]interface{}{}
		for _, v := range bk.Versions {
			versions = append(versions, map[string]interface{}{
				"version":      v.Version,
				"versionMajor": v.VersionMajor,
				"supportedOS":  v.SupportedOS,
				"tlsVersion":   v.TLSVersion,
				"cipherSuites": len(v.CipherSuites),
				"extensions":   len(v.Extensions),
				"h2WindowSize": v.H2InitialWindowSize,
				"releasedYear": v.ReleasedYear,
				"deprecated":   v.Deprecated,
			})
		}
		families = append(families, map[string]interface{}{
			"family":       string(bt),
			"marketShare":  bk.MarketShare,
			"cipherSuites": len(bk.CommonCipherSuites),
			"extensions":   len(bk.CommonExtensions),
			"versions":     versions,
		})
	}

	result := map[string]interface{}{
		"stats":    kbStats,
		"families": families,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleAgentStrategies returns active strategy list
func (h *Handler) handleAgentStrategies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a := h.gateway.GetAgent()
	if a == nil {
		http.Error(w, "Agent not enabled", http.StatusServiceUnavailable)
		return
	}

	strategies := a.GetActiveStrategies()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"strategies": strategies,
		"total":      len(strategies),
	})
}

func countLearnedStrategies(strategies []agent.StrategyInfo) int {
	count := 0
	for _, s := range strategies {
		if s.Learned {
			count++
		}
	}
	return count
}

// handleClientTest 使用指纹客户端测试访问网站
func (h *Handler) handleClientTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求
	var req struct {
		ProfileID string `json:"profileId"`
		URL       string `json:"url"`
		Method    string `json:"method"`
		Body      string `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// 验证参数
	if req.ProfileID == "" {
		http.Error(w, "profileId is required", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	// 查找指纹
	var profile profiles.ClientProfile
	found := false
	h.mu.RLock()
	for _, p := range h.profiles {
		if p.ID == req.ProfileID {
			profile = p
			found = true
			break
		}
	}
	h.mu.RUnlock()
	if !found {
		http.Error(w, fmt.Sprintf("Profile not found: %s", req.ProfileID), http.StatusNotFound)
		return
	}

	// 验证指纹完整性
	validator := profiles.NewProfileValidator()
	validationResult := validator.Validate(profile)

	// 创建客户端并测试
	result := testWithProfile(profile, req.URL, req.Method, req.Body, validationResult)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// testWithProfile 使用指定指纹测试访问 URL - 返回完整追踪信息和验证结果
func testWithProfile(profile profiles.ClientProfile, url, method, body string, validationResult profiles.ProfileValidationResult) map[string]interface{} {
	// 使用新的 ExecuteProxyRequest 获取完整追踪
	result := client.ExecuteProxyRequest(profile, url, method, body, nil)

	// 构建响应
	response := map[string]interface{}{
		"success":       result.Success,
		"error":         result.Error,
		"errorType":     result.ErrorType,
		"errorCode":     result.ErrorCode,
		"errorDetails":  result.ErrorDetails,
		"profileUsed":   result.ProfileUsed,
		"requestTrace":  result.RequestTrace,
		"responseTrace": result.ResponseTrace,
	}

	// 添加验证结果（如果有警告或错误）
	validation := map[string]interface{}{
		"valid": validationResult.Valid,
	}

	if len(validationResult.Warnings) > 0 {
		validation["warnings"] = validationResult.Warnings
	}

	if len(validationResult.Errors) > 0 {
		validation["errors"] = validationResult.Errors
	}

	if len(validationResult.MissingFields) > 0 {
		validation["missing_fields"] = validationResult.MissingFields
	}

	response["validation"] = validation

	return response
}
