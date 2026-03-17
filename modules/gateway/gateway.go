// Package gateway provides a high-performance API gateway service.
// Integrates rate limiting, caching, ML classification, and risk assessment.
package gateway

import (
	"log/slog"
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/agent"
	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/crawler"
	"github.com/vistone/fingerprint/modules/defense"
	"github.com/vistone/fingerprint/modules/frontend"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/network/tcp"
	"github.com/vistone/fingerprint/modules/plugin"
	"github.com/vistone/fingerprint/modules/waf"
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
	profileManager *ProfileManager       // Profile configuration manager
	injector       *HTMLInjector         // HTML response injector
	agent          *agent.Agent          // Autonomous security agent
	mlService      *ml.MLService         // Central ML service (optional)
	pluginManager  *plugin.Manager       // Plugin subsystem manager
	crawler        *crawler.Crawler      // Integrated crawler runtime
	waf            *waf.WAF              // Integrated WAF runtime
	closedLoop     *ClosedLoopController // Adversarial closed-loop controller
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

	// Anti-detection configuration
	AntiDetectEnabled       bool   // Whether to enable anti-detection injection
	AntiDetectProfileID     string // Default Profile ID to use
	AntiDetectConfigDir     string // Profile configuration file directory
	AntiDetectProxyTarget   string // Proxy target URL (optional)
	AntiDetectDirectProxy   bool   // Whether to use root path as transparent proxy entry
	AntiDetectInjectConsist bool   // Whether to inject consistency validation code

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

	// Closed-loop configuration
	ClosedLoopEnabled bool              // Whether to enable adversarial closed-loop training
	ClosedLoopConfig  *ClosedLoopConfig // Closed-loop configuration (nil uses defaults)
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
	// Anti-detection default configuration
	AntiDetectEnabled:       true,
	AntiDetectProfileID:     "chrome_134_default",
	AntiDetectConfigDir:     "./profiles",
	AntiDetectProxyTarget:   "",
	AntiDetectDirectProxy:   false,
	AntiDetectInjectConsist: true,
	ScannerUseBrowser:       false,
	ScannerBrowserWS:        "",
	ScannerBrowserTimeout:   25 * time.Second,
	AgentEnabled:            true,
}

// NewGateway creates a new gateway
func NewGateway(config *GatewayConfig) *Gateway {
	if config == nil {
		config = DefaultGatewayConfig.Clone()
	} else {
		config = config.Clone()
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
	g.initProfileAndInjector(config)
	g.initAgent(config)
	g.initMLService(config)
	g.initPlugins(config)
	g.initClosedLoop(config)

	return g
}

// initProfileAndInjector initializes the profile manager and HTML injector.
func (g *Gateway) initProfileAndInjector(config *GatewayConfig) {
	g.profileManager = NewProfileManager(&ProfileManagerConfig{
		ConfigDir:  config.AntiDetectConfigDir,
		DefaultID:  config.AntiDetectProfileID,
		AutoReload: false,
	})

	if err := g.profileManager.LoadAllProfiles(); err != nil {
		slog.Warn("failed to load profiles, using defaults", "error", err)
	}
	if !config.AntiDetectEnabled {
		return
	}

	profile, err := g.profileManager.GetDefaultProfile()
	if err != nil {
		slog.Warn("failed to get default profile", "error", err)
		profile = nil
	}

	g.injector, err = NewHTMLInjector(&InjectorConfig{
		Enabled:            true,
		TargetURL:          config.AntiDetectProxyTarget,
		Profile:            profile,
		InjectConsistency:  config.AntiDetectInjectConsist,
		RequireHeadTag:     true,
		AddInjectionMarker: false,
	})
	if err != nil {
		slog.Warn("failed to create HTML injector", "error", err)
		g.injector = nil
	}
}

// initAgent initializes the autonomous security agent.
func (g *Gateway) initAgent(config *GatewayConfig) {
	if !config.AgentEnabled {
		return
	}
	g.agent = agent.NewAgent(config.AgentConfig)
	g.agent.Start()
}

// initMLService initializes the central ML service.
func (g *Gateway) initMLService(config *GatewayConfig) {
	if !config.MLServiceEnabled {
		return
	}
	scfg := cloneMLServiceConfig(config.MLServiceConfig)
	if scfg == nil {
		scfg = cloneMLServiceConfig(ml.DefaultServiceConfig)
	}
	if config.MLClassifierPath != "" {
		scfg.ModelStorePath = config.MLClassifierPath
	}
	svc, err := ml.NewMLService(scfg)
	if err != nil {
		slog.Warn("failed to initialize ML service", "error", err)
		return
	}
	g.mlService = svc
}

// initPlugins initializes the plugin manager.
func (g *Gateway) initPlugins(config *GatewayConfig) {
	g.pluginManager = plugin.NewManager()
	if config.PluginConfigPath == "" {
		return
	}
	if err := plugin.LoadPlugins(config.PluginConfigPath); err != nil {
		slog.Warn("failed to load plugins", "path", config.PluginConfigPath, "error", err)
	}
}

// initClosedLoop initializes the adversarial closed-loop controller.
func (g *Gateway) initClosedLoop(config *GatewayConfig) {
	if !config.ClosedLoopEnabled || g.mlService == nil {
		return
	}
	clCfg := cloneClosedLoopConfig(config.ClosedLoopConfig)
	if clCfg == nil {
		clCfg = cloneClosedLoopConfig(DefaultClosedLoopConfig)
	}
	clCfg.Enabled = true
	g.closedLoop = NewClosedLoopController(clCfg, g.mlService)
	g.closedLoop.Start()
}

// SetCrawler injects a crawler instance into the gateway's closed-loop
// controller, enabling the crawler → ML feedback pipeline.
func (g *Gateway) SetCrawler(cr *crawler.Crawler) {
	g.crawler = cr
	if g.closedLoop != nil {
		g.closedLoop.SetCrawler(cr)
	}
	// Wire gateway's ML service into the crawler for adaptive profiles
	if g.mlService != nil && cr != nil {
		cr.SetMLService(g.mlService)
	}
}

// SetWAF injects a WAF instance into the gateway's closed-loop controller,
// enabling the WAF → ML feedback pipeline.
func (g *Gateway) SetWAF(w *waf.WAF) {
	g.waf = w
	if g.mlService != nil && w != nil {
		w.SetMLService(g.mlService)
	}
	if g.closedLoop != nil {
		g.closedLoop.SetWAF(w)
	}
}

// ClosedLoopStats returns adversarial closed-loop statistics (nil if disabled).
func (g *Gateway) ClosedLoopStats() *ClosedLoopStats {
	if g.closedLoop == nil {
		return nil
	}
	return g.closedLoop.Stats()
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

// Analyze executes a complete fingerprint analysis
