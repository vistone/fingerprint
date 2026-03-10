// Package gateway 提供高性能 API 网关服务
// 集成限流、cache、ML classifyand风险评估
package gateway

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/agent"
	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/defense"
	"github.com/vistone/fingerprint/modules/frontend"
	"github.com/vistone/fingerprint/modules/ml"
	tlsmod "github.com/vistone/fingerprint/modules/tls"
)

const (
	// AnalyzeTimeout controls request analysis timeout in HTTP handler.
	AnalyzeTimeout = 5 * time.Second
)

// Gateway 指纹analyze网关
type Gateway struct {
	config         *GatewayConfig
	classifier     *ml.HierarchicalClassifier
	extractor      *ml.FeatureExtractor
	riskEngine     *defense.RiskEngine
	cache          *FingerprintCache
	limiter        *RateLimiter
	sdk            *frontend.SDK
	profileManager *ProfileManager // Profile configuration管理器
	injector       *HTMLInjector   // HTML response注入器
	agent          *agent.Agent    // 自主安全智能体
	mu             sync.RWMutex
}

// GatewayConfig 网关configuration
type GatewayConfig struct {
	// 限流configuration
	RateLimitRequests int           // 每秒request数
	RateLimitBurst    int           // 突发request数
	RateLimitWindow   time.Duration // 限流窗口

	// cacheconfiguration
	CacheSize    int           // cachesize
	CacheTTL     time.Duration // cache过期时间
	CacheEnabled bool          // whether启用cache

	// ML configuration
	MLClassifierPath string // classify器模型路径
	MLTrainingData   string // 训练data路径

	// 风险评估configuration
	RiskThreshold float64 // 风险阈值

	// 服务端点
	Endpoint string
	Port     int

	// P3 反检测configuration
	P3Enabled       bool   // whether启用 P3 反检测注入
	P3ProfileID     string // 默认使用的 Profile ID
	P3ConfigDir     string // Profile configurationfiledirectory
	P3ProxyTarget   string // proxy目标 URL（optional）
	P3DirectProxy   bool   // whether将根路径作为透明proxy入口
	P3InjectConsist bool   // whether注入一致性validate代码

	// 扫描器浏览器executeconfiguration
	ScannerUseBrowser     bool          // whether优先使用浏览器execute抓取
	ScannerBrowserWS      string        // 远程 Chrome DevTools WS 地址
	ScannerBrowserTimeout time.Duration // 浏览器抓取timeout

	// 安全configuration
	TrustedProxies []string // 信任的反向proxy IP 列表，为空则不信任proxy头

	// Agent 智能体configuration
	AgentEnabled bool               // whether启用自主安全智能体
	AgentConfig  *agent.AgentConfig // 智能体详细configuration（nil 使用默认）
}

// DefaultGatewayConfig 默认网关configuration
var DefaultGatewayConfig = &GatewayConfig{
	RateLimitRequests: core.DefaultRateLimit,
	RateLimitBurst:    core.DefaultRateLimitBurst,
	RateLimitWindow:   core.DefaultRateLimitWindow,
	CacheSize:         core.DefaultCacheSize,
	CacheTTL:          core.DefaultCacheTTL,
	CacheEnabled:      true,
	RiskThreshold:     core.RiskThresholdHigh,
	Endpoint:          "/api/v1",
	Port:              8080,
	// P3 默认configuration
	P3Enabled:             true,
	P3ProfileID:           "chrome_134_default",
	P3ConfigDir:           "./profiles",
	P3ProxyTarget:         "",
	P3DirectProxy:         false,
	P3InjectConsist:       true,
	ScannerUseBrowser:     false,
	ScannerBrowserWS:      "",
	ScannerBrowserTimeout: 25 * time.Second,
	AgentEnabled:          true,
}

// NewGateway create新的网关
func NewGateway(config *GatewayConfig) *Gateway {
	if config == nil {
		config = DefaultGatewayConfig
	}

	g := &Gateway{
		config:     config,
		classifier: ml.NewHierarchicalClassifier(),
		extractor:  ml.NewFeatureExtractor(),
		riskEngine: defense.NewRiskEngine(),
		cache:      NewFingerprintCache(config.CacheSize, config.CacheTTL),
		limiter:    NewRateLimiter(config.RateLimitRequests, config.RateLimitBurst, config.RateLimitWindow),
		sdk:        frontend.NewSDK(nil),
	}

	g.classifier.Initialize()

	// initialize Profile 管理器
	g.profileManager = NewProfileManager(&ProfileManagerConfig{
		ConfigDir:  config.P3ConfigDir,
		DefaultID:  config.P3ProfileID,
		AutoReload: false,
	})

	// 加载 Profile configuration
	if err := g.profileManager.LoadAllProfiles(); err != nil {
		fmt.Printf("Warning: failed to load profiles: %v, using defaults\n", err)
	}

	// initialize HTML 注入器
	if config.P3Enabled {
		profile, err := g.profileManager.GetDefaultProfile()
		if err != nil {
			fmt.Printf("Warning: failed to get default profile: %v\n", err)
			profile = nil
		}

		injectorConfig := &InjectorConfig{
			Enabled:            true,
			TargetURL:          config.P3ProxyTarget,
			Profile:            profile,
			InjectConsistency:  config.P3InjectConsist,
			RequireHeadTag:     true,
			AddInjectionMarker: false,
		}

		g.injector, err = NewHTMLInjector(injectorConfig)
		if err != nil {
			fmt.Printf("Warning: failed to create HTML injector: %v\n", err)
			g.injector = nil
		}
	}

	// initialize自主安全智能体
	if config.AgentEnabled {
		g.agent = agent.NewAgent(config.AgentConfig)
		g.agent.Start()
	}

	return g
}

// AnalyzeRequest analyzerequest
type AnalyzeRequest struct {
	// TLS data
	TLSVersion      uint16              `json:"tls_version"`
	CipherSuites    []uint16            `json:"cipher_suites"`
	Extensions      []core.TLSExtension `json:"extensions"`
	SupportedCurves []core.CurveID      `json:"supported_curves"`

	// HTTP data
	Headers *core.HTTPHeaders `json:"headers"`
	Method  string            `json:"method"`
	Path    string            `json:"path"`

	// 前端指纹（optional）
	Frontend *ml.FrontendFingerprintData `json:"frontend,omitempty"`

	// 客户端 IP
	ClientIP string `json:"client_ip"`
}

// AnalyzeResponse analyzeresponse
type AnalyzeResponse struct {
	// 指纹哈希
	FingerprintHash string `json:"fingerprint_hash"`

	// classifyresult
	Classification *ml.ClassificationResult `json:"classification"`

	// 风险评估
	RiskAssessment *core.RiskAssessment `json:"risk_assessment"`

	// 检测发现
	Findings []defense.Finding `json:"findings,omitempty"`

	// JA3/JA4 指纹
	JA3 *JA3Info `json:"ja3,omitempty"`
	JA4 *JA4Info `json:"ja4,omitempty"`

	// JA4H 指纹
	JA4H *JA4HInfo `json:"ja4h,omitempty"`

	// 防护建议
	DefenseHints []string `json:"defense_hints,omitempty"`

	// Agent 智能体决策
	AgentDecision *agent.Decision `json:"agent_decision,omitempty"`

	// cacheinfo
	Cached    bool      `json:"cached"`
	CacheTime time.Time `json:"cache_time,omitempty"`

	// process时间
	ProcessingTimeMs int64 `json:"processing_time_ms"`
}

// JA3Info JA3 指纹info
type JA3Info struct {
	Hash string `json:"hash"`
	Raw  string `json:"raw"`
}

// JA4Info JA4 指纹info
type JA4Info struct {
	Fingerprint string `json:"fingerprint"`
}

// JA4HInfo JA4H 指纹info
type JA4HInfo struct {
	Fingerprint string   `json:"fingerprint"`
	Headers     []string `json:"headers"`
}

// Close 优雅close网关，释放后台资源
func (g *Gateway) Close() {
	if g.limiter != nil {
		g.limiter.Close()
	}
	if g.agent != nil {
		g.agent.Stop()
	}
}

// GetAgent return自主安全智能体实例（供 Web 管理台query）
func (g *Gateway) GetAgent() *agent.Agent {
	return g.agent
}

// GetClassifier return ML 分层classify器
func (g *Gateway) GetClassifier() *ml.HierarchicalClassifier {
	return g.classifier
}

// GetExtractor returnfeatureextract器
func (g *Gateway) GetExtractor() *ml.FeatureExtractor {
	return g.extractor
}

// GetRiskEngine return风险评估引擎
func (g *Gateway) GetRiskEngine() *defense.RiskEngine {
	return g.riskEngine
}

// GetSDK return前端 SDK
func (g *Gateway) GetSDK() *frontend.SDK {
	return g.sdk
}

// GetInjector return HTML 注入器
func (g *Gateway) GetInjector() *HTMLInjector {
	return g.injector
}

// GetProfileManager return Profile 管理器
func (g *Gateway) GetProfileManager() *ProfileManager {
	return g.profileManager
}

// GetConfig return当前网关configuration（只读副本）
func (g *Gateway) GetConfig() *GatewayConfig {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.config
}

// UpdateConfig 热update网关configuration（线程安全）
func (g *Gateway) UpdateConfig(apply func(cfg *GatewayConfig)) {
	g.mu.Lock()
	defer g.mu.Unlock()
	apply(g.config)
}

// Analyze execute完整的指纹analyze
func (g *Gateway) Analyze(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error) {
	start := time.Now()

	// generate指纹哈希
	fingerprintHash := g.generateFingerprintHash(req)

	// checkcache
	if g.config.CacheEnabled {
		if cached, ok := g.cache.Get(fingerprintHash); ok {
			resp := cloneAnalyzeResponse(cached)
			resp.Cached = true
			resp.CacheTime = time.Now()
			resp.ProcessingTimeMs = time.Since(start).Milliseconds()
			return resp, nil
		}
	}

	// extractfeature
	features := g.extractFeatures(req)

	// ML classify
	classification := g.classifier.Classify(features)

	// 风险评估
	risk := g.riskEngine.Evaluate(features, classification)

	// 检测
	detector := defense.NewDetector()
	detection := detector.Detect(features)

	// generate JA3/JA4/JA4H
	ja3, ja4 := g.calculateTLSFingerprints(req)
	ja4h := g.calculateHTTPFingerprint(req)

	response := &AnalyzeResponse{
		FingerprintHash:  fingerprintHash,
		Classification:   classification,
		RiskAssessment:   risk,
		Findings:         detection.Findings,
		JA3:              ja3,
		JA4:              ja4,
		JA4H:             ja4h,
		DefenseHints:     risk.Suggestions,
		Cached:           false,
		ProcessingTimeMs: time.Since(start).Milliseconds(),
	}

	// Agent 智能体process
	if g.agent != nil {
		obs := &agent.Observation{
			ID:              fingerprintHash,
			ClientID:        req.ClientIP,
			Timestamp:       time.Now(),
			Features:        features,
			Classification:  classification,
			Detection:       detection,
			RiskAssessment:  risk,
			FingerprintHash: fingerprintHash,
		}
		response.AgentDecision = g.agent.Process(ctx, obs)
	}

	// 存入cache
	if g.config.CacheEnabled {
		g.cache.Set(fingerprintHash, response)
	}

	return response, nil
}

// extractFeatures extractfeature
func (g *Gateway) extractFeatures(req *AnalyzeRequest) *core.FeatureVector {
	fv := core.NewFeatureVector()

	// TLS feature
	fv.Set(core.FeatureTLSVersion, float64(req.TLSVersion))
	fv.Set(core.FeatureCipherSuites, float64(len(req.CipherSuites)))
	fv.Set(core.FeatureExtensions, float64(len(req.Extensions)))

	// HTTP feature
	if req.Headers != nil {
		httpFV := g.extractor.ExtractFromHTTPHeaders(req.Headers)
		for ft, v := range httpFV.Features {
			fv.Set(ft, v)
		}
	}

	// 前端feature
	if req.Frontend != nil {
		frontendFV := g.extractor.ExtractFromFrontend(*req.Frontend)
		for ft, v := range frontendFV.Features {
			fv.Set(ft, v)
		}
	}

	return fv
}

// calculateTLSFingerprints calculate TLS 指纹
func (g *Gateway) calculateTLSFingerprints(req *AnalyzeRequest) (*JA3Info, *JA4Info) {
	spec := core.ClientHelloSpec{
		TLSVersion:      req.TLSVersion,
		CipherSuites:    req.CipherSuites,
		Extensions:      req.Extensions,
		SupportedCurves: req.SupportedCurves,
	}

	// JA3
	ja3Result := calculateJA3(spec)

	// JA4
	ja4Result := calculateJA4(spec)

	return &JA3Info{
			Hash: ja3Result.Hash,
			Raw:  ja3Result.RawString,
		}, &JA4Info{
			Fingerprint: ja4Result.Fingerprint,
		}
}

// calculateHTTPFingerprint calculate HTTP 指纹
func (g *Gateway) calculateHTTPFingerprint(req *AnalyzeRequest) *JA4HInfo {
	if req.Headers == nil {
		return nil
	}

	// 简化的 JA4H calculate
	headerMap := req.Headers.ToMap()
	headers := make([]string, 0, len(headerMap))
	for name := range headerMap {
		headers = append(headers, name)
	}

	return &JA4HInfo{
		Fingerprint: fmt.Sprintf("ja4h_%s_%d", req.Method, len(headers)),
		Headers:     headers,
	}
}

// generateFingerprintHash generate指纹哈希
func (g *Gateway) generateFingerprintHash(req *AnalyzeRequest) string {
	data := fmt.Sprintf("%d:%v:%v:%s:%s",
		req.TLSVersion,
		req.CipherSuites,
		req.Extensions,
		req.ClientIP,
		req.Method,
	)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

// HTTPHandler HTTP process函数
func (g *Gateway) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	// 限流check
	clientIP := g.getClientIP(r)
	if !g.limiter.Allow(clientIP) {
		writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// limitrequest体size，防止 DoS
	r.Body = http.MaxBytesReader(w, r.Body, core.MaxRequestBodySize)

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// setting客户端 IP
	req.ClientIP = clientIP

	// executeanalyze
	ctx, cancel := context.WithTimeout(r.Context(), AnalyzeTimeout)
	defer cancel()

	response, err := g.Analyze(ctx, &req)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SDKHandler SDK 脚本process函数
func (g *Gateway) SDKHandler(w http.ResponseWriter, r *http.Request) {
	// 限流check
	clientIP := g.getClientIP(r)
	if !g.limiter.Allow(clientIP) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// generate SDK 代码
	js := g.sdk.GenerateJSInjector(g.config.Endpoint + "/collect")

	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte(js))
}

// CollectHandler 前端data收集process函数
func (g *Gateway) CollectHandler(w http.ResponseWriter, r *http.Request) {
	// 使用 SDK 的process函数
	g.sdk.HandleCollect(w, r)
}

// getClientIP get客户端 IP
// 仅当 RemoteAddr 在 TrustedProxies 列表中时才信任proxy头
func (g *Gateway) getClientIP(r *http.Request) string {
	remoteIP := r.RemoteAddr
	if host, _, err := net.SplitHostPort(remoteIP); err == nil && host != "" {
		remoteIP = host
	}

	// 只有来自受信任proxyrequest才读取proxy头
	if g.isTrustedProxy(remoteIP) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if idx := strings.Index(xff, ","); idx != -1 {
				return strings.TrimSpace(xff[:idx])
			}
			return strings.TrimSpace(xff)
		}
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return strings.TrimSpace(xri)
		}
	}

	return remoteIP
}

// isTrustedProxy check IP whether在受信任proxy列表中
func (g *Gateway) isTrustedProxy(ip string) bool {
	for _, trusted := range g.config.TrustedProxies {
		if trusted == ip {
			return true
		}
	}
	return false
}

// Start start网关服务
func (g *Gateway) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc(g.config.Endpoint+"/analyze", g.HTTPHandler)
	mux.HandleFunc(g.config.Endpoint+"/sdk.js", g.SDKHandler)
	mux.HandleFunc(g.config.Endpoint+"/collect", g.CollectHandler)
	mux.HandleFunc(g.config.Endpoint+"/scan", g.V8ScannerHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "ok"}`))
	})

	addr := fmt.Sprintf(":%d", g.config.Port)
	fmt.Printf("Fingerprint Gateway starting on %s\n", addr)
	return http.ListenAndServe(addr, mux)
}

// RateLimiter 限流器
type RateLimiter struct {
	rate     int
	burst    int
	window   time.Duration
	visitors map[string]*Visitor
	mu       sync.Mutex
	stopCh   chan struct{}
}

// Visitor visit者
type Visitor struct {
	tokens     float64
	lastRefill time.Time
	lastSeen   time.Time
}

// NewRateLimiter create新的限流器
func NewRateLimiter(rate, burst int, window time.Duration) *RateLimiter {
	if rate <= 0 {
		rate = 1000
	}
	if burst <= 0 {
		burst = rate
	}
	if window <= 0 {
		window = time.Second
	}

	rl := &RateLimiter{
		rate:     rate,
		burst:    burst,
		window:   window,
		visitors: make(map[string]*Visitor),
		stopCh:   make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

// Allow checkwhetherallowrequest
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	fillRatePerSec := float64(rl.rate) / rl.window.Seconds()
	if fillRatePerSec <= 0 {
		fillRatePerSec = float64(rl.rate)
	}
	capacity := float64(rl.burst)

	v, ok := rl.visitors[key]
	if !ok {
		rl.visitors[key] = &Visitor{
			tokens:     capacity - 1,
			lastRefill: now,
			lastSeen:   now,
		}
		return true
	}

	// 按时间回填令牌
	if now.After(v.lastRefill) {
		elapsed := now.Sub(v.lastRefill).Seconds()
		v.tokens += elapsed * fillRatePerSec
		if v.tokens > capacity {
			v.tokens = capacity
		}
		v.lastRefill = now
	}

	if v.tokens < 1 {
		v.lastSeen = now
		return false
	}

	v.tokens -= 1
	v.lastSeen = now
	return true
}

// cleanup cleanup过期visit者
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for key, v := range rl.visitors {
				if time.Since(v.lastSeen) > 5*time.Minute {
					delete(rl.visitors, key)
				}
			}
			rl.mu.Unlock()
		case <-rl.stopCh:
			return
		}
	}
}

// Close stop限流器的后台 goroutine
func (rl *RateLimiter) Close() {
	close(rl.stopCh)
}

// FingerprintCache 指纹cache（基于 LRUCache implement）
type FingerprintCache struct {
	lru *LRUCache
}

// NewFingerprintCache create新的指纹cache
func NewFingerprintCache(size int, ttl time.Duration) *FingerprintCache {
	return &FingerprintCache{
		lru: NewLRUCache(size, ttl),
	}
}

// Get getcache
func (c *FingerprintCache) Get(key string) (*AnalyzeResponse, bool) {
	val, found := c.lru.Get(key)
	if !found {
		return nil, false
	}
	return cloneAnalyzeResponse(val.(*AnalyzeResponse)), true
}

// Set settingcache
func (c *FingerprintCache) Set(key string, response *AnalyzeResponse) {
	c.lru.Set(key, cloneAnalyzeResponse(response), 0) // 使用 LRUCache 的默认 TTL
}

func cloneAnalyzeResponse(resp *AnalyzeResponse) *AnalyzeResponse {
	if resp == nil {
		return nil
	}
	clone := *resp
	if resp.Findings != nil {
		clone.Findings = append([]defense.Finding(nil), resp.Findings...)
	}
	if resp.DefenseHints != nil {
		clone.DefenseHints = append([]string(nil), resp.DefenseHints...)
	}
	if resp.JA3 != nil {
		ja3 := *resp.JA3
		clone.JA3 = &ja3
	}
	if resp.JA4 != nil {
		ja4 := *resp.JA4
		clone.JA4 = &ja4
	}
	if resp.JA4H != nil {
		ja4h := *resp.JA4H
		if resp.JA4H.Headers != nil {
			ja4h.Headers = append([]string(nil), resp.JA4H.Headers...)
		}
		clone.JA4H = &ja4h
	}
	return &clone
}

// calculateJA3 calculate JA3 指纹（使用 tls 包implement）
func calculateJA3(spec core.ClientHelloSpec) *tlsmod.JA3Result {
	return tlsmod.CalculateJA3(spec)
}

// calculateJA4 calculate JA4 指纹（使用 tls 包implement）
func calculateJA4(spec core.ClientHelloSpec) *tlsmod.JA4Result {
	return tlsmod.CalculateJA4(spec)
}

// =====================================================================
// P3 反检测功能 - HTML 注入 Handler
// =====================================================================

// AntiDetectCodeHandler return P3 反检测 JavaScript 代码（单独interface）
func (g *Gateway) AntiDetectCodeHandler(w http.ResponseWriter, r *http.Request) {
	// 限流check
	clientIP := g.getClientIP(r)
	if !g.limiter.Allow(clientIP) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// checkwhether启用
	if g.injector == nil {
		http.Error(w, `{"error": "P3 anti-detection not enabled"}`, http.StatusServiceUnavailable)
		return
	}

	// get Profile ID parameter（optional）
	profileID := r.URL.Query().Get("profile")
	var code string
	if profileID != "" {
		profile, err := g.profileManager.GetProfile(profileID)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "profile not found: %s"}`, profileID), http.StatusNotFound)
			return
		}
		code = g.injector.GenerateInjectionCodeForProfile(profile)
	}

	// generate代码
	if code == "" {
		code = g.injector.GenerateInjectionCode()
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600") // cache 1 小时
	w.Write([]byte(code))
}

// ProfileListHandler return可用的 Profile 列表
func (g *Gateway) ProfileListHandler(w http.ResponseWriter, r *http.Request) {
	if g.profileManager == nil {
		http.Error(w, `{"error": "profile manager not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	profiles := g.profileManager.ListProfiles()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"profiles":    profiles,
		"default":     g.profileManager.defaultID,
		"total_count": len(profiles),
	})
}

// ProfileDetailHandler return指定 Profile 的详细info
func (g *Gateway) ProfileDetailHandler(w http.ResponseWriter, r *http.Request) {
	if g.profileManager == nil {
		http.Error(w, `{"error": "profile manager not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	profileID := r.URL.Query().Get("id")
	if profileID == "" {
		http.Error(w, `{"error": "profile id required"}`, http.StatusBadRequest)
		return
	}

	profile, err := g.profileManager.GetProfile(profileID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "profile not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// writeJSONError 安全地写入 JSON errorresponse，不泄露内部info
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp, _ := json.Marshal(map[string]string{"error": msg})
	w.Write(resp)
}

// V8ScannerHandler 扫描 JavaScript 中的指纹检测代码
func (g *Gateway) V8ScannerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeScannerJSONError(w, http.StatusMethodNotAllowed, "POST method required")
		return
	}

	// limitrequest体size（maximum5MB）
	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)

	var request struct {
		HTMLContent    string `json:"html"`
		URL            string `json:"url,omitempty"`
		FollowRedirect bool   `json:"followRedirects,omitempty"`
		MaxRedirects   int    `json:"maxRedirects,omitempty"`
		ExecuteJS      bool   `json:"executeJs,omitempty"`
		WaitMs         int    `json:"waitMs,omitempty"`
		ScanTimeout    int    `json:"scanTimeout,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeScannerJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %s", err.Error()))
		return
	}

	if request.HTMLContent == "" && strings.TrimSpace(request.URL) == "" {
		writeScannerJSONError(w, http.StatusBadRequest, "html or url is required")
		return
	}

	// 调用扫描器（带可configurationtimeout）
	scanTimeout := 20 * time.Second
	if request.ScanTimeout > 0 {
		scanTimeout = time.Duration(request.ScanTimeout) * time.Second
		if scanTimeout < 10*time.Second {
			scanTimeout = 10 * time.Second
		}
		if scanTimeout > 120*time.Second {
			scanTimeout = 120 * time.Second
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), scanTimeout)
	defer cancel()

	htmlContent := request.HTMLContent
	sourceURL := request.URL
	redirects := []string{}
	fetchMode := "inline-html"
	browserError := ""
	usedHeadless := false
	externalScriptsCaptured := 0
	var externalScriptStats *ExternalScriptStats

	if strings.TrimSpace(request.URL) != "" {
		fetchMode = "http"
		followRedirect := request.FollowRedirect
		if !request.FollowRedirect {
			// 默认启用重定向跟随
			followRedirect = true
		}

		maxRedirects := request.MaxRedirects
		if maxRedirects <= 0 {
			maxRedirects = 10
		}

		wantBrowser := request.ExecuteJS || g.config.ScannerUseBrowser
		if wantBrowser {
			waitMs := request.WaitMs
			if waitMs <= 0 {
				waitMs = 1200
			}
			if waitMs > 3500 {
				waitMs = 3500
			}

			if strings.TrimSpace(g.config.ScannerBrowserWS) != "" {
				browserTimeout := g.config.ScannerBrowserTimeout
				if browserTimeout <= 0 {
					browserTimeout = 25 * time.Second
				}
				// 给 headless 抓取更多预算，但仍保留收尾时间给扫描与encodingreturn。
				if scanTimeout > 8*time.Second {
					candidate := scanTimeout - 3*time.Second
					if candidate > browserTimeout {
						browserTimeout = candidate
					}
				}
				if browserTimeout > 25*time.Second {
					browserTimeout = 25 * time.Second
				}

				browserHTML, browserFinalURL, browserChain, stats, err := fetchHTMLWithHeadlessBrowser(
					ctx,
					request.URL,
					g.config.ScannerBrowserWS,
					maxRedirects,
					waitMs,
					browserTimeout,
				)
				if err == nil && browserHTML != "" {
					htmlContent = browserHTML
					sourceURL = browserFinalURL
					redirects = browserChain
					fetchMode = "headless-browser"
					usedHeadless = true
					externalScriptStats = stats
				} else if err != nil {
					browserError = err.Error()
				}
			}

			if !usedHeadless {
				execHTML, execFinalURL, execChain, err := fetchHTMLWithClientSideRedirects(
					ctx,
					request.URL,
					maxRedirects,
					waitMs,
				)
				if err == nil && execHTML != "" {
					htmlContent = execHTML
					sourceURL = execFinalURL
					redirects = execChain
					fetchMode = "js-redirect-emulation"
				} else if err != nil && browserError == "" {
					browserError = err.Error()
				}
			}
		}

		if fetchMode == "http" {
			remainingBudget := 12 * time.Second
			if deadline, ok := ctx.Deadline(); ok {
				left := time.Until(deadline) - 500*time.Millisecond
				if left <= 2*time.Second {
					writeScannerJSONError(w, http.StatusGatewayTimeout, "scan timeout")
					return
				}
				if left < remainingBudget {
					remainingBudget = left
				}
			}

			fetchedHTML, finalURL, chain, err := fetchHTMLWithRedirects(ctx, request.URL, followRedirect, maxRedirects, remainingBudget)
			if err == nil && fetchedHTML != "" {
				htmlContent = fetchedHTML
				sourceURL = finalURL
				redirects = chain
			} else if htmlContent == "" {
				writeScannerJSONError(w, http.StatusBadGateway, fmt.Sprintf("fetch url failed: %s", err.Error()))
				return
			}
		}
	}

	if fetchMode == "headless-browser" && externalScriptStats != nil {
		externalScriptsCaptured = externalScriptStats.Count
	}

	// 在goroutine中execute扫描，以便可以被ctx.Done()中断
	type scanResult struct {
		result *JSDetectionResult
		err    error
	}

	resultChan := make(chan scanResult, 1)
	go func() {
		result, err := ScanJavaScriptWithV8(ctx, htmlContent)
		resultChan <- scanResult{result, err}
	}()

	// wait扫描completeortimeout
	select {
	case <-ctx.Done():
		writeScannerJSONError(w, http.StatusGatewayTimeout, "scan timeout")
		return
	case res := <-resultChan:
		if res.err != nil {
			writeScannerJSONError(w, http.StatusInternalServerError, fmt.Sprintf("scan failed: %s", res.err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Scan-URL", sourceURL)
		responseData := map[string]interface{}{
			"url":                     request.URL,
			"finalUrl":                sourceURL,
			"redirectChain":           redirects,
			"fetchMode":               fetchMode,
			"browserError":            browserError,
			"externalScriptsCaptured": externalScriptsCaptured,
			"result":                  res.result,
		}
		if externalScriptStats != nil && fetchMode == "headless-browser" {
			responseData["externalScriptDetails"] = externalScriptStats
		}
		json.NewEncoder(w).Encode(responseData)
	}
}

func writeScannerJSONError(w http.ResponseWriter, statusCode int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// fetchHTMLWithRedirects 抓取 URL 内容并return最终 URL 与重定向链
func fetchHTMLWithRedirects(ctx context.Context, rawURL string, followRedirect bool, maxRedirects int, requestTimeout time.Duration) (string, string, []string, error) {
	redirectChain := []string{}
	trimmedURL := strings.TrimSpace(rawURL)
	if trimmedURL == "" {
		return "", "", redirectChain, fmt.Errorf("empty url")
	}
	if requestTimeout <= 0 {
		requestTimeout = 12 * time.Second
	}
	if requestTimeout < 3*time.Second {
		requestTimeout = 3 * time.Second
	}
	if requestTimeout > 20*time.Second {
		requestTimeout = 20 * time.Second
	}

	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt < 2; attempt++ {
		attemptChain := []string{}
		client := &http.Client{
			Timeout: requestTimeout,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   requestTimeout,
				ResponseHeaderTimeout: 8 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				IdleConnTimeout:       30 * time.Second,
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
		}
		if followRedirect {
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				attemptChain = append(attemptChain, req.URL.String())
				if len(via) >= maxRedirects {
					return fmt.Errorf("too many redirects: %d", maxRedirects)
				}
				return nil
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, trimmedURL, nil)
		if err != nil {
			return "", "", redirectChain, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

		resp, err = client.Do(req)
		if err == nil {
			redirectChain = attemptChain
			break
		}

		lastErr = err
		if !strings.Contains(strings.ToLower(err.Error()), "tls handshake timeout") || attempt == 1 {
			return "", "", redirectChain, err
		}
		// Retry once for transient TLS handshake stalls.
		time.Sleep(250 * time.Millisecond)
	}

	if resp == nil {
		if lastErr != nil {
			return "", "", redirectChain, lastErr
		}
		return "", "", redirectChain, fmt.Errorf("fetch failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 6*1024*1024))
	if err != nil {
		return "", "", redirectChain, err
	}

	return string(body), resp.Request.URL.String(), redirectChain, nil
}

// fetchHTMLWithClientSideRedirects 抓取页面并模拟execute常见客户端跳转。
// 支持 meta refresh、window.location、location.href、location.replace 等跳转模式。
func fetchHTMLWithClientSideRedirects(ctx context.Context, startURL string, maxHops int, waitMs int) (string, string, []string, error) {
	if maxHops <= 0 {
		maxHops = 5
	}
	if maxHops > 10 {
		maxHops = 10
	}

	currentURL := strings.TrimSpace(startURL)
	if currentURL == "" {
		return "", "", nil, fmt.Errorf("empty start url")
	}

	chain := []string{}
	visited := map[string]bool{}

	for hop := 0; hop < maxHops; hop++ {
		if visited[currentURL] {
			return "", "", chain, fmt.Errorf("redirect loop detected")
		}
		visited[currentURL] = true

		html, finalURL, httpChain, err := fetchHTMLWithRedirects(ctx, currentURL, true, maxHops, 8*time.Second)
		if err != nil {
			return "", "", chain, err
		}
		chain = append(chain, httpChain...)
		chain = append(chain, finalURL)

		// 简单wait，模拟页面初始脚本execute窗口
		if waitMs > 0 {
			time.Sleep(time.Duration(waitMs) * time.Millisecond)
		}

		target := extractClientSideRedirectTarget(html)
		if target == "" {
			return html, finalURL, chain, nil
		}

		resolved, err := resolveRedirectURL(finalURL, target)
		if err != nil {
			return html, finalURL, chain, nil
		}

		currentURL = resolved
	}

	return "", currentURL, chain, fmt.Errorf("max client-side redirect hops reached")
}

func extractClientSideRedirectTarget(html string) string {
	if strings.TrimSpace(html) == "" {
		return ""
	}

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)<meta[^>]+http-equiv\s*=\s*['"]?refresh['"]?[^>]+content\s*=\s*['"][^;"']*;\s*url=([^"'>\s]+)`),
		regexp.MustCompile(`(?is)window\.location\.href\s*=\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?is)window\.location\s*=\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?is)location\.href\s*=\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?is)location\.replace\(\s*['"]([^'"]+)['"]\s*\)`),
	}

	for _, p := range patterns {
		m := p.FindStringSubmatch(html)
		if len(m) > 1 {
			return strings.TrimSpace(m[1])
		}
	}

	return ""
}

func resolveRedirectURL(baseURL, target string) (string, error) {
	baseParsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", err
	}
	targetParsed, err := url.Parse(strings.TrimSpace(target))
	if err != nil {
		return "", err
	}
	return baseParsed.ResolveReference(targetParsed).String(), nil
}

// InjectProxyHandler 提供 HTML proxyand自动注入功能（用于proxy模式）
func (g *Gateway) InjectProxyHandler() http.Handler {
	if g.injector == nil || g.config.P3ProxyTarget == "" {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "proxy mode not configured", http.StatusServiceUnavailable)
		})
	}
	return g.injector
}

// GetInjectorMiddleware return注入器中间件（用于包装现有路由）
func (g *Gateway) GetInjectorMiddleware() func(http.Handler) http.Handler {
	if g.injector == nil {
		// return透明中间件
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return g.injector.InjectorMiddleware
}
