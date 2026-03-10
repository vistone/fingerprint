// translated comment
// translated comment
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

// translated comment
type Gateway struct {
	config         *GatewayConfig
	classifier     *ml.HierarchicalClassifier
	extractor      *ml.FeatureExtractor
	riskEngine     *defense.RiskEngine
	cache          *FingerprintCache
	limiter        *RateLimiter
	sdk            *frontend.SDK
	profileManager *ProfileManager // translated comment
	injector       *HTMLInjector   // translated comment
	agent          *agent.Agent    // translated comment
	mu             sync.RWMutex
}

// translated comment
type GatewayConfig struct {
	// translated comment
	RateLimitRequests int           // translated comment
	RateLimitBurst    int           // translated comment
	RateLimitWindow   time.Duration // translated comment

	// cacheconfiguration
	CacheSize    int           // cachesize
	CacheTTL     time.Duration // translated comment
	CacheEnabled bool          // translated comment

	// ML configuration
	MLClassifierPath string // translated comment
	MLTrainingData   string // translated comment

	// translated comment
	RiskThreshold float64 // translated comment

	// translated comment
	Endpoint string
	Port     int

	// translated comment
	P3Enabled       bool   // translated comment
	P3ProfileID     string // translated comment
	P3ConfigDir     string // Profile configurationfiledirectory
	P3ProxyTarget   string // translated comment
	P3DirectProxy   bool   // translated comment
	P3InjectConsist bool   // translated comment

	// translated comment
	ScannerUseBrowser     bool          // translated comment
	ScannerBrowserWS      string        // translated comment
	ScannerBrowserTimeout time.Duration // translated comment

	// translated comment
	TrustedProxies []string // translated comment

	// translated comment
	AgentEnabled bool               // translated comment
	AgentConfig  *agent.AgentConfig // translated comment
}

// translated comment
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
	// translated comment
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

// translated comment
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

	// translated comment
	g.profileManager = NewProfileManager(&ProfileManagerConfig{
		ConfigDir:  config.P3ConfigDir,
		DefaultID:  config.P3ProfileID,
		AutoReload: false,
	})

	// translated comment
	if err := g.profileManager.LoadAllProfiles(); err != nil {
		fmt.Printf("Warning: failed to load profiles: %v, using defaults\n", err)
	}

	// translated comment
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

	// translated comment
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

	// translated comment
	Frontend *ml.FrontendFingerprintData `json:"frontend,omitempty"`

	// translated comment
	ClientIP string `json:"client_ip"`
}

// AnalyzeResponse analyzeresponse
type AnalyzeResponse struct {
	// translated comment
	FingerprintHash string `json:"fingerprint_hash"`

	// classifyresult
	Classification *ml.ClassificationResult `json:"classification"`

	// translated comment
	RiskAssessment *core.RiskAssessment `json:"risk_assessment"`

	// translated comment
	Findings []defense.Finding `json:"findings,omitempty"`

	// translated comment
	JA3 *JA3Info `json:"ja3,omitempty"`
	JA4 *JA4Info `json:"ja4,omitempty"`

	// translated comment
	JA4H *JA4HInfo `json:"ja4h,omitempty"`

	// translated comment
	DefenseHints []string `json:"defense_hints,omitempty"`

	// translated comment
	AgentDecision *agent.Decision `json:"agent_decision,omitempty"`

	// cacheinfo
	Cached    bool      `json:"cached"`
	CacheTime time.Time `json:"cache_time,omitempty"`

	// translated comment
	ProcessingTimeMs int64 `json:"processing_time_ms"`
}

// translated comment
type JA3Info struct {
	Hash string `json:"hash"`
	Raw  string `json:"raw"`
}

// translated comment
type JA4Info struct {
	Fingerprint string `json:"fingerprint"`
}

// translated comment
type JA4HInfo struct {
	Fingerprint string   `json:"fingerprint"`
	Headers     []string `json:"headers"`
}

// translated comment
func (g *Gateway) Close() {
	if g.limiter != nil {
		g.limiter.Close()
	}
	if g.agent != nil {
		g.agent.Stop()
	}
}

// translated comment
func (g *Gateway) GetAgent() *agent.Agent {
	return g.agent
}

// translated comment
func (g *Gateway) GetClassifier() *ml.HierarchicalClassifier {
	return g.classifier
}

// translated comment
func (g *Gateway) GetExtractor() *ml.FeatureExtractor {
	return g.extractor
}

// translated comment
func (g *Gateway) GetRiskEngine() *defense.RiskEngine {
	return g.riskEngine
}

// translated comment
func (g *Gateway) GetSDK() *frontend.SDK {
	return g.sdk
}

// translated comment
func (g *Gateway) GetInjector() *HTMLInjector {
	return g.injector
}

// translated comment
func (g *Gateway) GetProfileManager() *ProfileManager {
	return g.profileManager
}

// translated comment
func (g *Gateway) GetConfig() *GatewayConfig {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.config
}

// translated comment
func (g *Gateway) UpdateConfig(apply func(cfg *GatewayConfig)) {
	g.mu.Lock()
	defer g.mu.Unlock()
	apply(g.config)
}

// translated comment
func (g *Gateway) Analyze(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error) {
	start := time.Now()

	// translated comment
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

	// translated comment
	risk := g.riskEngine.Evaluate(features, classification)

	// translated comment
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

	// translated comment
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

	// translated comment
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

	// translated comment
	if req.Frontend != nil {
		frontendFV := g.extractor.ExtractFromFrontend(*req.Frontend)
		for ft, v := range frontendFV.Features {
			fv.Set(ft, v)
		}
	}

	return fv
}

// translated comment
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

// translated comment
func (g *Gateway) calculateHTTPFingerprint(req *AnalyzeRequest) *JA4HInfo {
	if req.Headers == nil {
		return nil
	}

	// translated comment
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

// translated comment
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

// translated comment
func (g *Gateway) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	// translated comment
	clientIP := g.getClientIP(r)
	if !g.limiter.Allow(clientIP) {
		writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// translated comment
	r.Body = http.MaxBytesReader(w, r.Body, core.MaxRequestBodySize)

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// translated comment
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

// translated comment
func (g *Gateway) SDKHandler(w http.ResponseWriter, r *http.Request) {
	// translated comment
	clientIP := g.getClientIP(r)
	if !g.limiter.Allow(clientIP) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// translated comment
	js := g.sdk.GenerateJSInjector(g.config.Endpoint + "/collect")

	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte(js))
}

// translated comment
func (g *Gateway) CollectHandler(w http.ResponseWriter, r *http.Request) {
	// translated comment
	g.sdk.HandleCollect(w, r)
}

// translated comment
// translated comment
func (g *Gateway) getClientIP(r *http.Request) string {
	remoteIP := r.RemoteAddr
	if host, _, err := net.SplitHostPort(remoteIP); err == nil && host != "" {
		remoteIP = host
	}

	// translated comment
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

// translated comment
func (g *Gateway) isTrustedProxy(ip string) bool {
	for _, trusted := range g.config.TrustedProxies {
		if trusted == ip {
			return true
		}
	}
	return false
}

// translated comment
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

// translated comment
type RateLimiter struct {
	rate     int
	burst    int
	window   time.Duration
	visitors map[string]*Visitor
	mu       sync.Mutex
	stopCh   chan struct{}
}

// translated comment
type Visitor struct {
	tokens     float64
	lastRefill time.Time
	lastSeen   time.Time
}

// translated comment
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

	// translated comment
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

// translated comment
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

// translated comment
func (rl *RateLimiter) Close() {
	close(rl.stopCh)
}

// translated comment
type FingerprintCache struct {
	lru *LRUCache
}

// translated comment
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
	c.lru.Set(key, cloneAnalyzeResponse(response), 0) // translated comment
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

// translated comment
func calculateJA3(spec core.ClientHelloSpec) *tlsmod.JA3Result {
	return tlsmod.CalculateJA3(spec)
}

// translated comment
func calculateJA4(spec core.ClientHelloSpec) *tlsmod.JA4Result {
	return tlsmod.CalculateJA4(spec)
}

// =====================================================================
// translated comment
// =====================================================================

// translated comment
func (g *Gateway) AntiDetectCodeHandler(w http.ResponseWriter, r *http.Request) {
	// translated comment
	clientIP := g.getClientIP(r)
	if !g.limiter.Allow(clientIP) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// translated comment
	if g.injector == nil {
		http.Error(w, `{"error": "P3 anti-detection not enabled"}`, http.StatusServiceUnavailable)
		return
	}

	// translated comment
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

	// translated comment
	if code == "" {
		code = g.injector.GenerateInjectionCode()
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600") // translated comment
	w.Write([]byte(code))
}

// translated comment
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

// translated comment
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

// translated comment
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp, _ := json.Marshal(map[string]string{"error": msg})
	w.Write(resp)
}

// translated comment
func (g *Gateway) V8ScannerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeScannerJSONError(w, http.StatusMethodNotAllowed, "POST method required")
		return
	}

	// translated comment
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

	// translated comment
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
			// translated comment
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
				// translated comment
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

	// translated comment
	type scanResult struct {
		result *JSDetectionResult
		err    error
	}

	resultChan := make(chan scanResult, 1)
	go func() {
		result, err := ScanJavaScriptWithV8(ctx, htmlContent)
		resultChan <- scanResult{result, err}
	}()

	// translated comment
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

// translated comment
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

// translated comment
// translated comment
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

		// translated comment
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

// translated comment
func (g *Gateway) InjectProxyHandler() http.Handler {
	if g.injector == nil || g.config.P3ProxyTarget == "" {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "proxy mode not configured", http.StatusServiceUnavailable)
		})
	}
	return g.injector
}

// translated comment
func (g *Gateway) GetInjectorMiddleware() func(http.Handler) http.Handler {
	if g.injector == nil {
		// translated comment
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return g.injector.InjectorMiddleware
}
