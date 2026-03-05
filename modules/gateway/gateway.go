// Package gateway 提供高性能 API 网关服务
// 集成限流、缓存、ML 分类和风险评估
package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/defense"
	"github.com/vistone/fingerprint/modules/frontend"
	"github.com/vistone/fingerprint/modules/ml"
)

// Gateway 指纹分析网关
type Gateway struct {
	config     *GatewayConfig
	classifier *ml.HierarchicalClassifier
	extractor  *ml.FeatureExtractor
	riskEngine *defense.RiskEngine
	cache      *FingerprintCache
	limiter    *RateLimiter
	sdk        *frontend.SDK
	mu         sync.RWMutex
}

// GatewayConfig 网关配置
type GatewayConfig struct {
	// 限流配置
	RateLimitRequests int           // 每秒请求数
	RateLimitBurst    int           // 突发请求数
	RateLimitWindow   time.Duration // 限流窗口

	// 缓存配置
	CacheSize       int           // 缓存大小
	CacheTTL        time.Duration // 缓存过期时间
	CacheEnabled    bool          // 是否启用缓存

	// ML 配置
	MLClassifierPath string // 分类器模型路径
	MLTrainingData   string // 训练数据路径

	// 风险评估配置
	RiskThreshold float64 // 风险阈值

	// 服务端点
	Endpoint string
	Port     int
}

// DefaultGatewayConfig 默认网关配置
var DefaultGatewayConfig = &GatewayConfig{
	RateLimitRequests: 1000,
	RateLimitBurst:    2000,
	RateLimitWindow:   time.Second,
	CacheSize:         10000,
	CacheTTL:          5 * time.Minute,
	CacheEnabled:      true,
	RiskThreshold:     0.7,
	Endpoint:          "/api/v1",
	Port:              8080,
}

// NewGateway 创建新的网关
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
	return g
}

// AnalyzeRequest 分析请求
type AnalyzeRequest struct {
	// TLS 数据
	TLSVersion      uint16           `json:"tls_version"`
	CipherSuites    []uint16         `json:"cipher_suites"`
	Extensions      []core.TLSExtension `json:"extensions"`
	SupportedCurves []core.CurveID   `json:"supported_curves"`
	
	// HTTP 数据
	Headers     *core.HTTPHeaders `json:"headers"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	
	// 前端指纹（可选）
	Frontend *ml.FrontendFingerprintData `json:"frontend,omitempty"`
	
	// 客户端 IP
	ClientIP string `json:"client_ip"`
}

// AnalyzeResponse 分析响应
type AnalyzeResponse struct {
	// 指纹哈希
	FingerprintHash string `json:"fingerprint_hash"`
	
	// 分类结果
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
	
	// 缓存信息
	Cached    bool      `json:"cached"`
	CacheTime time.Time `json:"cache_time,omitempty"`
	
	// 处理时间
	ProcessingTimeMs int64 `json:"processing_time_ms"`
}

// JA3Info JA3 指纹信息
type JA3Info struct {
	Hash string `json:"hash"`
	Raw  string `json:"raw"`
}

// JA4Info JA4 指纹信息
type JA4Info struct {
	Fingerprint string `json:"fingerprint"`
}

// JA4HInfo JA4H 指纹信息
type JA4HInfo struct {
	Fingerprint string   `json:"fingerprint"`
	Headers     []string `json:"headers"`
}

// Analyze 执行完整的指纹分析
func (g *Gateway) Analyze(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error) {
	start := time.Now()

	// 生成指纹哈希
	fingerprintHash := g.generateFingerprintHash(req)

	// 检查缓存
	if g.config.CacheEnabled {
		if cached, ok := g.cache.Get(fingerprintHash); ok {
			cached.Cached = true
			cached.CacheTime = time.Now()
			cached.ProcessingTimeMs = time.Since(start).Milliseconds()
			return cached, nil
		}
	}

	// 提取特征
	features := g.extractFeatures(req)

	// ML 分类
	classification := g.classifier.Classify(features)

	// 风险评估
	risk := g.riskEngine.Evaluate(features, classification)

	// 检测
	detector := defense.NewDetector()
	detection := detector.Detect(features)

	// 生成 JA3/JA4/JA4H
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

	// 存入缓存
	if g.config.CacheEnabled {
		g.cache.Set(fingerprintHash, response)
	}

	return response, nil
}

// extractFeatures 提取特征
func (g *Gateway) extractFeatures(req *AnalyzeRequest) *core.FeatureVector {
	fv := core.NewFeatureVector()

	// TLS 特征
	fv.Set(core.FeatureTLSVersion, float64(req.TLSVersion))
	fv.Set(core.FeatureCipherSuites, float64(len(req.CipherSuites)))
	fv.Set(core.FeatureExtensions, float64(len(req.Extensions)))

	// HTTP 特征
	if req.Headers != nil {
		httpFV := g.extractor.ExtractFromHTTPHeaders(req.Headers)
		for ft, v := range httpFV.Features {
			fv.Set(ft, v)
		}
	}

	// 前端特征
	if req.Frontend != nil {
		frontendFV := g.extractor.ExtractFromFrontend(*req.Frontend)
		for ft, v := range frontendFV.Features {
			fv.Set(ft, v)
		}
	}

	return fv
}

// calculateTLSFingerprints 计算 TLS 指纹
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

// calculateHTTPFingerprint 计算 HTTP 指纹
func (g *Gateway) calculateHTTPFingerprint(req *AnalyzeRequest) *JA4HInfo {
	if req.Headers == nil {
		return nil
	}

	// 简化的 JA4H 计算
	headers := []string{}
	headerMap := req.Headers.ToMap()
	for name := range headerMap {
		headers = append(headers, name)
	}

	return &JA4HInfo{
		Fingerprint: fmt.Sprintf("ja4h_%s_%d", req.Method, len(headers)),
		Headers:     headers,
	}
}

// generateFingerprintHash 生成指纹哈希
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

// HTTPHandler HTTP 处理函数
func (g *Gateway) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	// 限流检查
	clientIP := g.getClientIP(r)
	if !g.limiter.Allow(clientIP) {
		http.Error(w, `{"error": "rate limit exceeded"}`, http.StatusTooManyRequests)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "invalid json: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	// 设置客户端 IP
	req.ClientIP = clientIP

	// 执行分析
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	response, err := g.Analyze(ctx, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SDKHandler SDK 脚本处理函数
func (g *Gateway) SDKHandler(w http.ResponseWriter, r *http.Request) {
	// 限流检查
	clientIP := g.getClientIP(r)
	if !g.limiter.Allow(clientIP) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// 生成 SDK 代码
	js := g.sdk.GenerateJSInjector(g.config.Endpoint + "/collect")

	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte(js))
}

// CollectHandler 前端数据收集处理函数
func (g *Gateway) CollectHandler(w http.ResponseWriter, r *http.Request) {
	// 使用 SDK 的处理函数
	g.sdk.HandleCollect(w, r)
}

// getClientIP 获取客户端 IP
func (g *Gateway) getClientIP(r *http.Request) string {
	// 从 X-Forwarded-For 获取
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// 从 X-Real-IP 获取
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// 从 RemoteAddr 获取
	return r.RemoteAddr
}

// Start 启动网关服务
func (g *Gateway) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc(g.config.Endpoint+"/analyze", g.HTTPHandler)
	mux.HandleFunc(g.config.Endpoint+"/sdk.js", g.SDKHandler)
	mux.HandleFunc(g.config.Endpoint+"/collect", g.CollectHandler)
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
}

// Visitor 访问者
type Visitor struct {
	requests  int
	lastSeen  time.Time
	windowEnd time.Time
}

// NewRateLimiter 创建新的限流器
func NewRateLimiter(rate, burst int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		rate:     rate,
		burst:    burst,
		window:   window,
		visitors: make(map[string]*Visitor),
	}
	go rl.cleanup()
	return rl
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, ok := rl.visitors[key]
	if !ok {
		rl.visitors[key] = &Visitor{
			requests:  1,
			lastSeen:  now,
			windowEnd: now.Add(rl.window),
		}
		return true
	}

	// 检查窗口是否重置
	if now.After(v.windowEnd) {
		v.requests = 1
		v.windowEnd = now.Add(rl.window)
		v.lastSeen = now
		return true
	}

	// 检查是否超过突发限制
	if v.requests >= rl.burst {
		return false
	}

	v.requests++
	v.lastSeen = now
	return true
}

// cleanup 清理过期访问者
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for key, v := range rl.visitors {
			if time.Since(v.lastSeen) > 5*time.Minute {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// FingerprintCache 指纹缓存
type FingerprintCache struct {
	cache  map[string]*fingerprintCacheEntry
	size   int
	ttl    time.Duration
	mu     sync.RWMutex
}

// cacheEntry 缓存项
type fingerprintCacheEntry struct {
	response *AnalyzeResponse
	expires  time.Time
}

// NewFingerprintCache 创建新的指纹缓存
func NewFingerprintCache(size int, ttl time.Duration) *FingerprintCache {
	return &FingerprintCache{
		cache: make(map[string]*fingerprintCacheEntry),
		size:  size,
		ttl:   ttl,
	}
}

// Get 获取缓存
func (c *FingerprintCache) Get(key string) (*AnalyzeResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fentry, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(fentry.expires) {
		return nil, false
	}

	return fentry.response, true
}

// Set 设置缓存
func (c *FingerprintCache) Set(key string, response *AnalyzeResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果缓存满了，删除最旧的项
	if len(c.cache) >= c.size {
		var oldestKey string
		var oldestTime time.Time
		first := true
		for k, v := range c.cache {
			if first || v.expires.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.expires
				first = false
			}
		}
		delete(c.cache, oldestKey)
	}

	c.cache[key] = &fingerprintCacheEntry{
		response: response,
		expires:  time.Now().Add(c.ttl),
	}
}

// 简化的 JA3 计算
func calculateJA3(spec core.ClientHelloSpec) struct {
	Hash      string
	RawString string
} {
	// 简化实现
	raw := fmt.Sprintf("%d-%d-%d", spec.TLSVersion, len(spec.CipherSuites), len(spec.Extensions))
	hash := sha256.Sum256([]byte(raw))
	return struct {
		Hash      string
		RawString string
	}{
		Hash:      hex.EncodeToString(hash[:16]),
		RawString: raw,
	}
}

// 简化的 JA4 计算
func calculateJA4(spec core.ClientHelloSpec) struct {
	Fingerprint string
} {
	version := "13"
	if spec.TLSVersion < 0x0303 {
		version = "12"
	}
	return struct {
		Fingerprint string
	}{
		Fingerprint: fmt.Sprintf("t%sd%d%d", version, len(spec.CipherSuites), len(spec.Extensions)),
	}
}
