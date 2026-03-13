// Package gateway provides a high-performance API gateway service.
// Integrates rate limiting, caching, ML classification, and risk assessment.
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
	"github.com/vistone/fingerprint/modules/network/ja4t"
	"github.com/vistone/fingerprint/modules/network/tcp"
	"github.com/vistone/fingerprint/modules/plugin"
	tlsmod "github.com/vistone/fingerprint/modules/tls"
)

const (
	// AnalyzeTimeout controls request analysis timeout in HTTP handler.
	AnalyzeTimeout = 5 * time.Second
)

// Gateway is the fingerprint analysis gateway
type Gateway struct {
	config         *GatewayConfig
	classifier     *ml.HierarchicalClassifier
	extractor      *ml.FeatureExtractor
	riskEngine     *defense.RiskEngine
	cache          *FingerprintCache
	limiter        *RateLimiter
	sdk            *frontend.SDK
	profileManager *ProfileManager // Profile configuration manager
	injector       *HTMLInjector   // HTML response injector
	agent          *agent.Agent    // Autonomous security agent
	mlService      *ml.MLService   // Central ML service (optional)
	pluginManager  *plugin.Manager // Plugin subsystem manager
	mu             sync.RWMutex
}

// GatewayConfig is the gateway configuration
type GatewayConfig struct {
	// Rate limiting configuration
	RateLimitRequests int           // Requests per second
	RateLimitBurst    int           // Burst request count
	RateLimitWindow   time.Duration // Rate limit window

	// Cache configuration
	CacheSize    int           // Cache size
	CacheTTL     time.Duration // Cache TTL
	CacheEnabled bool          // Whether to enable cache

	// ML configuration
	MLClassifierPath string // Classifier model path
	MLTrainingData   string // Training data path

	// Risk assessment configuration
	RiskThreshold float64 // Risk threshold

	// Service endpoint
	Endpoint string
	Port     int

	// P3 anti-detection configuration
	P3Enabled       bool   // Whether to enable P3 anti-detection injection
	P3ProfileID     string // Default Profile ID to use
	P3ConfigDir     string // Profile configuration file directory
	P3ProxyTarget   string // Proxy target URL (optional)
	P3DirectProxy   bool   // Whether to use root path as transparent proxy entry
	P3InjectConsist bool   // Whether to inject consistency validation code

	// Scanner browser execution configuration
	ScannerUseBrowser     bool          // Whether to prefer browser-based fetching
	ScannerBrowserWS      string        // Remote Chrome DevTools WS address
	ScannerBrowserTimeout time.Duration // Browser fetch timeout

	// Security configuration
	TrustedProxies []string // Trusted reverse proxy IP list; empty means proxy headers are not trusted

	// Agent configuration
	AgentEnabled bool               // Whether to enable autonomous security agent
	AgentConfig  *agent.AgentConfig // Detailed agent configuration (nil uses defaults)

	// ML Service configuration
	MLServiceEnabled bool              // Whether to enable central ML service
	MLServiceConfig  *ml.ServiceConfig // ML service configuration (nil uses defaults)

	// Plugin configuration
	PluginConfigPath string // Plugin configuration path; empty disables plugin loading
}

// DefaultGatewayConfig is the default gateway configuration
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
	// P3 default configuration
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

// NewGateway creates a new gateway
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

	// Initialize Profile manager
	g.profileManager = NewProfileManager(&ProfileManagerConfig{
		ConfigDir:  config.P3ConfigDir,
		DefaultID:  config.P3ProfileID,
		AutoReload: false,
	})

	// Load Profile configuration
	if err := g.profileManager.LoadAllProfiles(); err != nil {
		fmt.Printf("Warning: failed to load profiles: %v, using defaults\n", err)
	}

	// Initialize HTML injector
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

	// Initialize autonomous security agent
	if config.AgentEnabled {
		g.agent = agent.NewAgent(config.AgentConfig)
		g.agent.Start()
	}

	// Initialize central ML service
	if config.MLServiceEnabled {
		scfg := config.MLServiceConfig
		if scfg == nil {
			scfg = ml.DefaultServiceConfig
		}
		// Wire MLClassifierPath into model store if provided
		if config.MLClassifierPath != "" {
			scfg.ModelStorePath = config.MLClassifierPath
		}
		svc, err := ml.NewMLService(scfg)
		if err != nil {
			fmt.Printf("Warning: failed to initialize ML service: %v\n", err)
		} else {
			g.mlService = svc
		}
	}

	// Initialize plugin manager
	g.pluginManager = plugin.NewManager()
	if config.PluginConfigPath != "" {
		if err := plugin.LoadPlugins(config.PluginConfigPath); err != nil {
			fmt.Printf("Warning: failed to load plugins from %s: %v\n", config.PluginConfigPath, err)
		}
	}

	return g
}

// AnalyzeRequest is the analysis request
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

	// Frontend fingerprint (optional)
	Frontend *ml.FrontendFingerprintData `json:"frontend,omitempty"`

	// TCP/IP network layer data (optional)
	TCPData *TCPRequestData `json:"tcp_data,omitempty"`

	// Structured TCP packets for deep TCP/IP analysis (optional)
	TCPPackets []tcp.TCPPacket `json:"tcp_packets,omitempty"`

	// Client IP
	ClientIP string `json:"client_ip"`
}

// TCPRequestData carries TCP SYN parameters for JA4T and TCP/IP fingerprinting
type TCPRequestData struct {
	WindowSize  uint16 `json:"window_size"`
	MSS         uint16 `json:"mss"`
	WindowScale uint8  `json:"window_scale"`
	TTL         uint8  `json:"ttl"`
	DF          bool   `json:"df"`
	Options     []byte `json:"options,omitempty"`
}

// AnalyzeResponse is the analysis response
type AnalyzeResponse struct {
	// Fingerprint hash
	FingerprintHash string `json:"fingerprint_hash"`

	// Classification result
	Classification *ml.ClassificationResult `json:"classification"`

	// Risk assessment
	RiskAssessment *core.RiskAssessment `json:"risk_assessment"`

	// Risk blocked (true when risk exceeds configured threshold)
	RiskBlocked bool `json:"risk_blocked"`

	// Detection findings
	Findings []defense.Finding `json:"findings,omitempty"`

	// JA3/JA4 fingerprints
	JA3 *JA3Info `json:"ja3,omitempty"`
	JA4 *JA4Info `json:"ja4,omitempty"`

	// JA4H fingerprint
	JA4H *JA4HInfo `json:"ja4h,omitempty"`

	// JA4T transport fingerprint
	JA4T *JA4TInfo `json:"ja4t,omitempty"`

	// TCP/IP network analysis
	NetworkAnalysis *NetworkAnalysisResult `json:"network_analysis,omitempty"`

	// Plugin analysis results
	PluginResults []PluginFinding `json:"plugin_results,omitempty"`

	// Defense suggestions
	DefenseHints []string `json:"defense_hints,omitempty"`

	// Agent decision
	AgentDecision *agent.Decision `json:"agent_decision,omitempty"`

	// ML Service enrichment (when MLService is enabled)
	MLValidation *ml.ValidationResult `json:"ml_validation,omitempty"`

	// Cache info
	Cached    bool      `json:"cached"`
	CacheTime time.Time `json:"cache_time,omitempty"`

	// Processing time
	ProcessingTimeMs int64 `json:"processing_time_ms"`
}

// JA4TInfo contains JA4T transport fingerprint info
type JA4TInfo struct {
	Hash      string   `json:"hash"`
	Raw       string   `json:"raw"`
	Anomalies []string `json:"anomalies,omitempty"`
	GuessedOS string   `json:"guessed_os,omitempty"`
}

// NetworkAnalysisResult contains TCP/IP network analysis results
type NetworkAnalysisResult struct {
	OS             string   `json:"os,omitempty"`
	OSFamily       string   `json:"os_family,omitempty"`
	OSConfidence   float64  `json:"os_confidence,omitempty"`
	IsVPN          bool     `json:"is_vpn"`
	IsProxy        bool     `json:"is_proxy"`
	IsNAT          bool     `json:"is_nat"`
	NetworkRisk    float64  `json:"network_risk"`
	InitialTTL     int      `json:"initial_ttl,omitempty"`
	MSS            int      `json:"mss,omitempty"`
	AnomaliesFound []string `json:"anomalies_found,omitempty"`
}

// PluginFinding contains a single plugin analysis result
type PluginFinding struct {
	PluginName string  `json:"plugin_name"`
	Category   string  `json:"category"`
	Message    string  `json:"message"`
	RiskScore  float64 `json:"risk_score"`
}

// JA3Info contains JA3 fingerprint info
type JA3Info struct {
	Hash string `json:"hash"`
	Raw  string `json:"raw"`
}

// JA4Info contains JA4 fingerprint info
type JA4Info struct {
	Fingerprint string `json:"fingerprint"`
}

// JA4HInfo contains JA4H fingerprint info
type JA4HInfo struct {
	Fingerprint string   `json:"fingerprint"`
	Headers     []string `json:"headers"`
}

// Close gracefully shuts down the gateway and releases background resources
func (g *Gateway) Close() {
	if g.limiter != nil {
		g.limiter.Close()
	}
	if g.agent != nil {
		g.agent.Stop()
	}
}

// GetAgent returns the autonomous security agent instance (for web admin console queries)
func (g *Gateway) GetAgent() *agent.Agent {
	return g.agent
}

// GetClassifier returns the ML hierarchical classifier
func (g *Gateway) GetClassifier() *ml.HierarchicalClassifier {
	return g.classifier
}

// GetExtractor returns the feature extractor
func (g *Gateway) GetExtractor() *ml.FeatureExtractor {
	return g.extractor
}

// GetRiskEngine returns the risk assessment engine
func (g *Gateway) GetRiskEngine() *defense.RiskEngine {
	return g.riskEngine
}

// GetSDK returns the frontend SDK
func (g *Gateway) GetSDK() *frontend.SDK {
	return g.sdk
}

// GetInjector returns the HTML injector
func (g *Gateway) GetInjector() *HTMLInjector {
	return g.injector
}

// GetProfileManager returns the Profile manager
func (g *Gateway) GetProfileManager() *ProfileManager {
	return g.profileManager
}

// GetMLService returns the central ML service (nil if not enabled)
func (g *Gateway) GetMLService() *ml.MLService {
	return g.mlService
}

// GetPluginManager returns the plugin manager
func (g *Gateway) GetPluginManager() *plugin.Manager {
	return g.pluginManager
}

// GetConfig returns the current gateway configuration (read-only copy)
func (g *Gateway) GetConfig() *GatewayConfig {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.config
}

// UpdateConfig hot-updates the gateway configuration (thread-safe)
func (g *Gateway) UpdateConfig(apply func(cfg *GatewayConfig)) {
	g.mu.Lock()
	defer g.mu.Unlock()
	apply(g.config)
}

// Analyze executes a complete fingerprint analysis
func (g *Gateway) Analyze(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error) {
	start := time.Now()

	// Generate fingerprint hash
	fingerprintHash := g.generateFingerprintHash(req)

	// Check cache
	if g.config.CacheEnabled {
		if cached, ok := g.cache.Get(fingerprintHash); ok {
			resp := cloneAnalyzeResponse(cached)
			resp.Cached = true
			resp.CacheTime = time.Now()
			resp.ProcessingTimeMs = time.Since(start).Milliseconds()
			return resp, nil
		}
	}

	// Extract features
	features := g.extractFeatures(req)

	// ML classification
	classification := g.classifier.Classify(features)

	// Risk assessment
	risk := g.riskEngine.Evaluate(features, classification)

	// Detection
	detector := defense.NewDetector()
	detection := detector.Detect(features)

	// Generate JA3/JA4/JA4H
	ja3, ja4 := g.calculateTLSFingerprints(req)
	ja4h := g.calculateHTTPFingerprint(req)

	// Network layer: JA4T transport fingerprint + TCP/IP analysis
	var ja4tInfo *JA4TInfo
	var netAnalysis *NetworkAnalysisResult
	if req.TCPData != nil {
		// Compute JA4T fingerprint from TCP SYN data
		synData := ja4t.TCPSYNData{
			WindowSize:  req.TCPData.WindowSize,
			Options:     req.TCPData.Options,
			MSS:         req.TCPData.MSS,
			WindowScale: req.TCPData.WindowScale,
			TTL:         req.TCPData.TTL,
			DF:          req.TCPData.DF,
		}
		ja4tResult := ja4t.ComputeJA4T(synData)
		if ja4tResult != nil {
			ja4tInfo = &JA4TInfo{
				Hash:      ja4tResult.Hash,
				Raw:       ja4tResult.RawFingerprint,
				Anomalies: ja4tResult.AnomalyFlags,
				GuessedOS: ja4tResult.ProbableOS,
			}
			// Merge TCP features into the feature vector
			features.Set(core.FeatureType("tcp_window_size"), float64(synData.WindowSize))
			features.Set(core.FeatureType("tcp_mss"), float64(synData.MSS))
			features.Set(core.FeatureType("tcp_ttl"), float64(synData.TTL))
			if synData.DF {
				features.Set(core.FeatureType("tcp_df"), 1.0)
			}
		}
	}

	// TCP/IP packet-level analysis (if structured packets provided)
	if len(req.TCPPackets) > 0 {
		analyzer := tcp.NewTCPIPAnalyzer()
		for _, pkt := range req.TCPPackets {
			analyzer.AddPacket(pkt)
		}
		tcpResult, err := analyzer.AnalyzeStream()
		if err == nil {
			netAnalysis = &NetworkAnalysisResult{
				OS:             tcpResult.OS,
				OSFamily:       tcpResult.OSFamily,
				OSConfidence:   tcpResult.Confidence,
				IsVPN:          tcpResult.IsVPN,
				IsProxy:        tcpResult.IsProxy,
				IsNAT:          tcpResult.IsNAT,
				NetworkRisk:    tcpResult.RiskScore,
				InitialTTL:     tcpResult.InitialTTL,
				MSS:            tcpResult.MSS,
				AnomaliesFound: tcpResult.AnomaliesFound,
			}
			features.Set(core.FeatureType("network_risk"), tcpResult.RiskScore)
			if tcpResult.IsVPN {
				features.Set(core.FeatureType("is_vpn"), 1.0)
			}
			if tcpResult.IsProxy {
				features.Set(core.FeatureType("is_proxy"), 1.0)
			}
		}
	}

	// Plugin analysis pipeline
	var pluginFindings []PluginFinding
	if g.pluginManager != nil {
		pluginData := map[string]interface{}{
			"fingerprint_hash": fingerprintHash,
			"client_ip":        req.ClientIP,
			"tls_version":      req.TLSVersion,
			"cipher_suites":    req.CipherSuites,
			"classification":   classification,
			"risk":             risk,
		}
		if results, err := g.pluginManager.ExecuteAnalyzers(ctx, pluginData); err == nil {
			for _, r := range results {
				msg := fmt.Sprintf("score=%.2f confidence=%.2f", r.Score, r.Confidence)
				category := ""
				if cat, ok := r.Labels["category"]; ok {
					category = cat
				}
				pluginName := ""
				if name, ok := r.Labels["plugin"]; ok {
					pluginName = name
				}
				pluginFindings = append(pluginFindings, PluginFinding{
					PluginName: pluginName,
					Category:   category,
					Message:    msg,
					RiskScore:  r.Score,
				})
			}
		}
		// Run validators
		if vResult, err := g.pluginManager.ExecuteValidators(ctx, pluginData); err == nil && vResult != nil && !vResult.Valid {
			for _, issue := range vResult.Errors {
				pluginFindings = append(pluginFindings, PluginFinding{
					PluginName: "validator",
					Category:   "validation",
					Message:    issue,
				})
			}
			for _, warn := range vResult.Warnings {
				pluginFindings = append(pluginFindings, PluginFinding{
					PluginName: "validator",
					Category:   "warning",
					Message:    warn,
				})
			}
		}
	}

	// Check risk threshold
	riskBlocked := false
	if g.config.RiskThreshold > 0 && risk != nil && risk.Score >= g.config.RiskThreshold {
		riskBlocked = true
	}

	response := &AnalyzeResponse{
		FingerprintHash:  fingerprintHash,
		Classification:   classification,
		RiskAssessment:   risk,
		RiskBlocked:      riskBlocked,
		Findings:         detection.Findings,
		JA3:              ja3,
		JA4:              ja4,
		JA4H:             ja4h,
		JA4T:             ja4tInfo,
		NetworkAnalysis:  netAnalysis,
		PluginResults:    pluginFindings,
		DefenseHints:     risk.Suggestions,
		Cached:           false,
		ProcessingTimeMs: time.Since(start).Milliseconds(),
	}

	// ML Service validation (when enabled)
	if g.mlService != nil && g.mlService.IsReady() {
		result := g.mlService.InferFromFeatures(features, nil)
		if result != nil {
			vr := &ml.ValidationResult{
				Valid:            !result.Forgery.IsForgery,
				ForgeryProb:      result.Forgery.ForgeryProb,
				ForgeryType:      result.Forgery.ForgeryType,
				ConsistencyScore: 1.0 - result.Forgery.ForgeryProb,
				BrowserFamily:    string(result.Browser.Family),
				Confidence:       result.Browser.Confidence,
			}
			response.MLValidation = vr
			if !vr.Valid {
				response.DefenseHints = append(response.DefenseHints,
					fmt.Sprintf("ML: forgery detected (type=%s, prob=%.2f)", vr.ForgeryType, vr.ForgeryProb))
			}
		}
	}

	// Agent processing
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

	// Store in cache
	if g.config.CacheEnabled {
		g.cache.Set(fingerprintHash, response)
	}

	return response, nil
}

// extractFeatures extracts features
func (g *Gateway) extractFeatures(req *AnalyzeRequest) *core.FeatureVector {
	fv := core.NewFeatureVector()

	// TLS features
	fv.Set(core.FeatureTLSVersion, float64(req.TLSVersion))
	fv.Set(core.FeatureCipherSuites, float64(len(req.CipherSuites)))
	fv.Set(core.FeatureExtensions, float64(len(req.Extensions)))

	// HTTP features
	if req.Headers != nil {
		httpFV := g.extractor.ExtractFromHTTPHeaders(req.Headers)
		for ft, v := range httpFV.Features {
			fv.Set(ft, v)
		}
	}

	// Frontend features
	if req.Frontend != nil {
		frontendFV := g.extractor.ExtractFromFrontend(*req.Frontend)
		for ft, v := range frontendFV.Features {
			fv.Set(ft, v)
		}
	}

	return fv
}

// calculateTLSFingerprints calculates TLS fingerprints
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

// calculateHTTPFingerprint calculates HTTP fingerprint
func (g *Gateway) calculateHTTPFingerprint(req *AnalyzeRequest) *JA4HInfo {
	if req.Headers == nil {
		return nil
	}

	// Simplified JA4H calculation
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

// generateFingerprintHash generates a fingerprint hash
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

// HTTPHandler is the HTTP handler function
func (g *Gateway) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	// Rate limit check
	clientIP := g.getClientIP(r)
	if !g.limiter.Allow(clientIP) {
		writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Limit request body size to prevent DoS
	r.Body = http.MaxBytesReader(w, r.Body, core.MaxRequestBodySize)

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// Set client IP
	req.ClientIP = clientIP

	// Execute analysis
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

// SDKHandler is the SDK script handler function
func (g *Gateway) SDKHandler(w http.ResponseWriter, r *http.Request) {
	// Rate limit check
	clientIP := g.getClientIP(r)
	if !g.limiter.Allow(clientIP) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Generate SDK code
	js := g.sdk.GenerateJSInjector(g.config.Endpoint + "/collect")

	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte(js))
}

// CollectHandler is the frontend data collection handler function
func (g *Gateway) CollectHandler(w http.ResponseWriter, r *http.Request) {
	// Use SDK's handler function
	g.sdk.HandleCollect(w, r)
}

// getClientIP gets the client IP
// Only trust proxy headers when RemoteAddr is in the TrustedProxies list
func (g *Gateway) getClientIP(r *http.Request) string {
	remoteIP := r.RemoteAddr
	if host, _, err := net.SplitHostPort(remoteIP); err == nil && host != "" {
		remoteIP = host
	}

	// Only read proxy headers from trusted proxy requests
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

// isTrustedProxy checks if IP is in the trusted proxy list
func (g *Gateway) isTrustedProxy(ip string) bool {
	for _, trusted := range g.config.TrustedProxies {
		if trusted == ip {
			return true
		}
	}
	return false
}

// Start starts the gateway service
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

// RateLimiter is the rate limiter
type RateLimiter struct {
	rate     int
	burst    int
	window   time.Duration
	visitors map[string]*Visitor
	mu       sync.Mutex
	stopCh   chan struct{}
}

// Visitor represents a visitor
type Visitor struct {
	tokens     float64
	lastRefill time.Time
	lastSeen   time.Time
}

// NewRateLimiter creates a new rate limiter
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

// Allow checks whether the request is allowed
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

	// Refill tokens based on elapsed time
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

// cleanup removes expired visitors
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

// Close stops the rate limiter's background goroutine
func (rl *RateLimiter) Close() {
	close(rl.stopCh)
}

// FingerprintCache is a fingerprint cache (based on LRUCache implementation)
type FingerprintCache struct {
	lru *LRUCache
}

// NewFingerprintCache creates a new fingerprint cache
func NewFingerprintCache(size int, ttl time.Duration) *FingerprintCache {
	return &FingerprintCache{
		lru: NewLRUCache(size, ttl),
	}
}

// Get retrieves from cache
func (c *FingerprintCache) Get(key string) (*AnalyzeResponse, bool) {
	val, found := c.lru.Get(key)
	if !found {
		return nil, false
	}
	return cloneAnalyzeResponse(val.(*AnalyzeResponse)), true
}

// Set stores in cache
func (c *FingerprintCache) Set(key string, response *AnalyzeResponse) {
	c.lru.Set(key, cloneAnalyzeResponse(response), 0) // Use LRUCache's default TTL
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

// calculateJA3 calculates JA3 fingerprint (using tls package implementation)
func calculateJA3(spec core.ClientHelloSpec) *tlsmod.JA3Result {
	return tlsmod.CalculateJA3(spec)
}

// calculateJA4 calculates JA4 fingerprint (using tls package implementation)
func calculateJA4(spec core.ClientHelloSpec) *tlsmod.JA4Result {
	return tlsmod.CalculateJA4(spec)
}

// =====================================================================
// P3 Anti-Detection - HTML Injection Handler
// =====================================================================

// AntiDetectCodeHandler returns P3 anti-detection JavaScript code (standalone endpoint)
func (g *Gateway) AntiDetectCodeHandler(w http.ResponseWriter, r *http.Request) {
	// Rate limit check
	clientIP := g.getClientIP(r)
	if !g.limiter.Allow(clientIP) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Check if enabled
	if g.injector == nil {
		http.Error(w, `{"error": "P3 anti-detection not enabled"}`, http.StatusServiceUnavailable)
		return
	}

	// Get Profile ID parameter (optional)
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

	// Generate code
	if code == "" {
		code = g.injector.GenerateInjectionCode()
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	w.Write([]byte(code))
}

// ProfileListHandler returns the list of available Profiles
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

// ProfileDetailHandler returns detailed info for the specified Profile
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

// writeJSONError safely writes a JSON error response without leaking internal info
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp, _ := json.Marshal(map[string]string{"error": msg})
	w.Write(resp)
}

// V8ScannerHandler scans JavaScript for fingerprint detection code
func (g *Gateway) V8ScannerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeScannerJSONError(w, http.StatusMethodNotAllowed, "POST method required")
		return
	}

	// Limit request body size (maximum 5MB)
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

	// Call scanner (with configurable timeout)
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
			// Enable redirect following by default
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
				// Give headless fetch more time budget, but still reserve time for scan and encoding.
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

	// Execute scan in goroutine so it can be interrupted by ctx.Done()
	type scanResult struct {
		result *JSDetectionResult
		err    error
	}

	resultChan := make(chan scanResult, 1)
	go func() {
		result, err := ScanJavaScriptWithV8(ctx, htmlContent)
		resultChan <- scanResult{result, err}
	}()

	// Wait for scan to complete or timeout
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

// fetchHTMLWithRedirects fetches URL content and returns the final URL and redirect chain
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

// fetchHTMLWithClientSideRedirects fetches a page and emulates common client-side redirects.
// Supports meta refresh, window.location, location.href, location.replace redirect patterns.
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

		// Simple wait, simulating initial page script execution window
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

// InjectProxyHandler provides HTML proxy and auto-injection (for proxy mode)
func (g *Gateway) InjectProxyHandler() http.Handler {
	if g.injector == nil || g.config.P3ProxyTarget == "" {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "proxy mode not configured", http.StatusServiceUnavailable)
		})
	}
	return g.injector
}

// GetInjectorMiddleware returns the injector middleware (for wrapping existing routes)
func (g *Gateway) GetInjectorMiddleware() func(http.Handler) http.Handler {
	if g.injector == nil {
		// Return transparent middleware
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return g.injector.InjectorMiddleware
}
