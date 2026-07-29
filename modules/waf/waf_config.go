// Package waf Advanced Web Application Firewall - Multi-layer protection against crawlers and automation tools
//
// Protection Layers:
//
//	L1: Network Layer - TCP/IP fingerprinting, rate limiting, IP reputation
//	L2: Transport Layer - TLS fingerprint verification, JA3/JA4 blacklist
//	L3: Application Layer - HTTP behavior analysis, request fingerprinting
//	L4: Business Layer - Behavior sequence detection, bot detection
//	L5: Persistence Layer - Device fingerprinting, long-term tracking
//
// Response Strategies:
//   - Allow: Direct access granted
//   - Monitor: Monitoring mode, logs but does not block
//   - Challenge: JS challenge, CAPTCHA
//   - Throttle: Rate limiting
//   - Block: Direct blocking
package waf

import (
	"time"

	"github.com/vistone/fingerprint/modules/core"
)

// WAFConfig WAF configuration
type WAFConfig struct {
	// Basic configuration
	Enabled  bool
	Mode     WAFMode  // Operating mode
	LogLevel LogLevel // Log level

	// Layer protection switches
	NetworkLayerEnabled  bool // L1: Network Layer
	TLSLayerEnabled      bool // L2: Transport Layer
	HTTPLayerEnabled     bool // L3: Application Layer
	BehaviorLayerEnabled bool // L4: Behavior Layer
	DeviceLayerEnabled   bool // L5: Device Layer

	// Threshold configuration
	RiskThreshold  float64       // Risk threshold
	RateLimitRPS   int           // Requests per second limit
	RateLimitBurst int           // Burst limit
	BlockDuration  time.Duration // Block duration
	ChallengeTTL   time.Duration // Challenge TTL

	// Machine Learning
	MLClassifierPath string
	MLEnabled        bool

	// Autonomous Agent
	AgentEnabled bool

	// Response templates
	BlockResponse []byte // Block response content
	ChallengeHTML []byte // JS challenge page

	// Whitelist/Blacklist
	WhitelistIPs   []string
	BlacklistIPs   []string
	WhitelistPaths []string // URL whitelist
	BlacklistPaths []string // URL blacklist
	BlacklistJA3   []string // JA3 blacklist
	BlacklistJA4   []string // JA4 blacklist
	TrustedProxies []string // Trusted reverse proxy IPs allowed to supply forwarding headers
}

// WAFMode operating mode
type WAFMode string

const (
	WAFModeLearning   WAFMode = "learning"   // Learning mode: log only, no blocking
	WAFModeDetection  WAFMode = "detection"  // Detection mode: log and alert
	WAFModeProtection WAFMode = "protection" // Protection mode: active blocking
	WAFModeAggressive WAFMode = "aggressive" // Aggressive mode: strict blocking
)

// LogLevel log level
type LogLevel string

const (
	LogLevelError LogLevel = "error"
	LogLevelWarn  LogLevel = "warn"
	LogLevelInfo  LogLevel = "info"
	LogLevelDebug LogLevel = "debug"
)

// DefaultWAFConfig default configuration
var DefaultWAFConfig = &WAFConfig{
	Enabled:              true,
	Mode:                 WAFModeProtection,
	LogLevel:             LogLevelInfo,
	NetworkLayerEnabled:  true,
	TLSLayerEnabled:      true,
	HTTPLayerEnabled:     true,
	BehaviorLayerEnabled: true,
	DeviceLayerEnabled:   true,
	RiskThreshold:        0.7,
	RateLimitRPS:         100,
	RateLimitBurst:       150,
	BlockDuration:        1 * time.Hour,
	ChallengeTTL:         10 * time.Minute,
	MLEnabled:            true,
	AgentEnabled:         true,
}

// WAFStats statistics
type WAFStats struct {
	TotalRequests      int64
	AllowedRequests    int64
	BlockedRequests    int64
	ChallengedRequests int64
	ThrottledRequests  int64
	MonitoredRequests  int64
}

// WAFDecision is a compact runtime record of one WAF decision.
type WAFDecision struct {
	Timestamp       time.Time `json:"timestamp"`
	Action          WAFAction `json:"action"`
	Reason          string    `json:"reason"`
	RiskScore       float64   `json:"riskScore"`
	ClientIP        string    `json:"clientIp"`
	Method          string    `json:"method"`
	Path            string    `json:"path"`
	DetectionLayers []string  `json:"detectionLayers,omitempty"`
}

// WAFResult detection result
type WAFResult struct {
	Action          WAFAction
	Reason          string
	RiskScore       float64
	RiskLevel       core.RiskLevel
	DetectionLayers []string
	FingerprintInfo *FingerprintInfo
	ChallengeToken  string
	BlockDuration   time.Duration
	RiskFactors     []core.RiskFactor
}

// WAFAction response action
type WAFAction string

const (
	ActionAllow     WAFAction = "allow"
	ActionMonitor   WAFAction = "monitor"
	ActionChallenge WAFAction = "challenge"
	ActionThrottle  WAFAction = "throttle"
	ActionBlock     WAFAction = "block"
)

// FingerprintInfo fingerprint information
type FingerprintInfo struct {
	JA3         string
	JA4         string
	JA4H        string
	JA4T        string
	DeviceID    string
	SessionID   string
	RiskFactors []core.RiskFactor
}
