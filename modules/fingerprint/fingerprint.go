// Package fingerprint is the main facade package.
// It provides a unified API interface, integrating all submodule functionality.
//
// This is the main entry point after the Go Workspace refactoring,
// inspired by the Rust workspace architecture design.
//
// Basic usage:
//
//	import "github.com/vistone/fingerprint"
//
//	// Get a random fingerprint
//	result := fingerprint.GetRandom()
//
//	// Get a fingerprint for a specific browser
//	chrome := fingerprint.GetByBrowser(fingerprint.BrowserChrome)
//
//	// Execute fingerprint analysis
//	analyzer := fingerprint.NewAnalyzer()
//	result := analyzer.Analyze(request)
//
// Module structure:
//
//	github.com/vistone/fingerprint/core     - Core types and interfaces
//	github.com/vistone/fingerprint/profiles - Browser fingerprint profiles
//	github.com/vistone/fingerprint/tls      - TLS fingerprint generation
//	github.com/vistone/fingerprint/http     - HTTP fingerprint generation
//	github.com/vistone/fingerprint/ml       - ML classifier
//	github.com/vistone/fingerprint/defense  - Security defense
//	github.com/vistone/fingerprint/frontend - Frontend SDK
//	github.com/vistone/fingerprint/gateway  - API gateway
package fingerprint

import (
	"context"
	"fmt"
	"net/http"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/defense"
	"github.com/vistone/fingerprint/modules/frontend"
	"github.com/vistone/fingerprint/modules/gateway"
	httpmod "github.com/vistone/fingerprint/modules/http"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
	"github.com/vistone/fingerprint/modules/tls"
)

// Version is the current version of the fingerprint library.
const Version = "2.1.0"

// ==================== Type aliases ====================

// BrowserType represents the browser type
type BrowserType = core.BrowserType

const (
	BrowserChrome  = core.BrowserChrome
	BrowserFirefox = core.BrowserFirefox
	BrowserSafari  = core.BrowserSafari
	BrowserOpera   = core.BrowserOpera
	BrowserEdge    = core.BrowserEdge
)

// OperatingSystem represents the operating system type
type OperatingSystem = core.OperatingSystem

const (
	OSWindows10   = core.OSWindows10
	OSWindows11   = core.OSWindows11
	OSMacOS13     = core.OSMacOS13
	OSMacOS14     = core.OSMacOS14
	OSMacOS15     = core.OSMacOS15
	OSLinux       = core.OSLinux
	OSLinuxUbuntu = core.OSLinuxUbuntu
)

// ProtocolType represents the protocol type
type ProtocolType = core.ProtocolType

const (
	ProtocolTLS   = core.ProtocolTLS
	ProtocolHTTP  = core.ProtocolHTTP
	ProtocolHTTP2 = core.ProtocolHTTP2
	ProtocolQUIC  = core.ProtocolQUIC
	ProtocolHTTP3 = core.ProtocolHTTP3
)

// FeatureType represents the feature type
type FeatureType = core.FeatureType

const (
	FeatureTLSVersion      = core.FeatureTLSVersion
	FeatureCipherSuites    = core.FeatureCipherSuites
	FeatureExtensions      = core.FeatureExtensions
	FeatureHTTP2Settings   = core.FeatureHTTP2Settings
	FeatureHTTPHeaders     = core.FeatureHTTPHeaders
	FeatureUserAgent       = core.FeatureUserAgent
	FeatureCanvas          = core.FeatureCanvas
	FeatureWebGL           = core.FeatureWebGL
	FeatureAudio           = core.FeatureAudio
	FeatureFonts           = core.FeatureFonts
	FeatureStorage         = core.FeatureStorage
	FeatureWebRTC          = core.FeatureWebRTC
	FeatureHardware        = core.FeatureHardware
	FeatureTiming          = core.FeatureTiming
	FeatureHeadlessBrowser = core.FeatureHeadlessBrowser
	FeatureEntropy         = core.FeatureEntropy
)

// RiskLevel represents the risk level
type RiskLevel = core.RiskLevel

const (
	RiskLevelNone     = core.RiskLevelNone
	RiskLevelLow      = core.RiskLevelLow
	RiskLevelMedium   = core.RiskLevelMedium
	RiskLevelHigh     = core.RiskLevelHigh
	RiskLevelCritical = core.RiskLevelCritical
)

// ==================== Core type aliases ====================

// HTTPHeaders represents HTTP headers
type HTTPHeaders = core.HTTPHeaders

// FingerprintResult represents a fingerprint result
type FingerprintResult = core.FingerprintResult

// ClientHelloSpec represents a ClientHello specification
type ClientHelloSpec = core.ClientHelloSpec

// TLSExtension represents a TLS extension
type TLSExtension = core.TLSExtension

// CurveID represents a curve ID
type CurveID = core.CurveID

// HTTP2Settings represents HTTP/2 settings
type HTTP2Settings = core.HTTP2Settings

// HTTP2Priority represents HTTP/2 priority
type HTTP2Priority = core.HTTP2Priority

// FeatureVector represents a feature vector
type FeatureVector = core.FeatureVector

// RiskAssessment represents a risk assessment
type RiskAssessment = core.RiskAssessment

// RiskFactor represents a risk factor
type RiskFactor = core.RiskFactor

// ==================== Core functions ====================

// GetRandom returns a random fingerprint
func GetRandom() *profiles.ClientProfile {
	p := profiles.GetRandom()
	return &p
}

// GetByBrowser returns a random fingerprint by browser type
func GetByBrowser(browser BrowserType) *profiles.ClientProfile {
	p := profiles.GetRandomByBrowser(browser)
	return &p
}

// GetAllProfiles returns all fingerprint profiles
func GetAllProfiles() []profiles.ClientProfile {
	return profiles.GetAll()
}

// GetProfile returns a fingerprint profile by ID
func GetProfile(id string) (*profiles.ClientProfile, bool) {
	p, ok := profiles.Get(id)
	if !ok {
		return nil, false
	}
	return &p, true
}

// ==================== TLS fingerprint ====================

// CalculateJA3 calculates the JA3 fingerprint
func CalculateJA3(spec ClientHelloSpec) *tls.JA3Result {
	return tls.CalculateJA3(spec)
}

// CalculateJA4 calculates the JA4 fingerprint
func CalculateJA4(spec ClientHelloSpec) *tls.JA4Result {
	return tls.CalculateJA4(spec)
}

// JA3Result JA3 result
type JA3Result = tls.JA3Result

// JA4Result JA4 result
type JA4Result = tls.JA4Result

// ==================== HTTP fingerprint ====================

// CalculateJA4H calculates the JA4H fingerprint
func CalculateJA4H(headers *HTTPHeaders, method string) *httpmod.JA4HResult {
	return httpmod.CalculateJA4H(headers, method)
}

// JA4HResult JA4H result
type JA4HResult = httpmod.JA4HResult

// ==================== ML classifier ====================

// HierarchicalClassifier is a three-layer hierarchical classifier
type HierarchicalClassifier = ml.HierarchicalClassifier

// NewHierarchicalClassifier creates a new hierarchical classifier
func NewHierarchicalClassifier() *HierarchicalClassifier {
	hc := ml.NewHierarchicalClassifier()
	hc.Initialize()
	return hc
}

// ClassificationResult represents a classification result
type ClassificationResult = ml.ClassificationResult

// FeatureExtractor is a feature extractor
type FeatureExtractor = ml.FeatureExtractor

// NewFeatureExtractor creates a new feature extractor
func NewFeatureExtractor() *FeatureExtractor {
	return ml.NewFeatureExtractor()
}

// ==================== Security defense ====================

// DefenseSystem represents the defense system
type DefenseSystem = defense.DefenseSystem

// NewDefenseSystem creates a new defense system
func NewDefenseSystem() *DefenseSystem {
	return defense.NewDefenseSystem()
}

// DetectionResult represents a detection result
type DetectionResult = defense.DetectionResult

// RiskEngine represents the risk engine
type RiskEngine = defense.RiskEngine

// NewRiskEngine creates a new risk engine
func NewRiskEngine() *RiskEngine {
	return defense.NewRiskEngine()
}

// ProtectionConfig represents the protection configuration
type ProtectionConfig = defense.ProtectionConfig

// DefaultProtectionConfig is the default protection configuration
var DefaultProtectionConfig = defense.DefaultProtectionConfig

// ==================== Frontend SDK ====================

// FrontendSDK represents the frontend SDK
type FrontendSDK = frontend.SDK

// NewFrontendSDK creates a new frontend SDK
func NewFrontendSDK(config *frontend.SDKConfig) *FrontendSDK {
	return frontend.NewSDK(config)
}

// DefaultSDKConfig is the default SDK configuration
var DefaultSDKConfig = frontend.DefaultSDKConfig

// FrontendFingerprintData represents frontend fingerprint data
type FrontendFingerprintData = ml.FrontendFingerprintData

// ==================== Gateway ====================

// Gateway represents the API gateway
type Gateway = gateway.Gateway

// GatewayConfig represents the gateway configuration
type GatewayConfig = gateway.GatewayConfig

// NewGateway creates a new gateway
func NewGateway(config *GatewayConfig) *Gateway {
	return gateway.NewGateway(config)
}

// DefaultGatewayConfig is the default gateway configuration
var DefaultGatewayConfig = gateway.DefaultGatewayConfig

// AnalyzeRequest represents an analysis request
type AnalyzeRequest = gateway.AnalyzeRequest

// AnalyzeResponse represents an analysis response
type AnalyzeResponse = gateway.AnalyzeResponse

// StartGateway starts the gateway service
func StartGateway(config *GatewayConfig) error {
	gw := NewGateway(config)
	return gw.Start()
}

// ==================== Analyzer ====================

// Analyzer is a comprehensive fingerprint analyzer
type Analyzer struct {
	classifier *ml.HierarchicalClassifier
	extractor  *ml.FeatureExtractor
	riskEngine *defense.RiskEngine
	defense    *defense.DefenseSystem
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer() *Analyzer {
	hc := ml.NewHierarchicalClassifier()
	hc.Initialize()

	return &Analyzer{
		classifier: hc,
		extractor:  ml.NewFeatureExtractor(),
		riskEngine: defense.NewRiskEngine(),
		defense:    defense.NewDefenseSystem(),
	}
}

// Analyze executes a comprehensive analysis
func (a *Analyzer) Analyze(req *gateway.AnalyzeRequest) *AnalyzeResponse {
	ctx := context.Background()
	gw := gateway.NewGateway(nil)
	resp, _ := gw.Analyze(ctx, req)
	return resp
}

// Classify executes classification
func (a *Analyzer) Classify(features *core.FeatureVector) *ml.ClassificationResult {
	return a.classifier.Classify(features)
}

// ExtractFeatures extracts features from a profile
func (a *Analyzer) ExtractFeatures(profile *profiles.ClientProfile) *core.FeatureVector {
	return a.extractor.ExtractFromProfile(profile)
}

// EvaluateRisk evaluates the risk
func (a *Analyzer) EvaluateRisk(features *core.FeatureVector, classification *ml.ClassificationResult) *core.RiskAssessment {
	return a.riskEngine.Evaluate(features, classification)
}

// GetDefenseAdvice returns defense recommendations
func (a *Analyzer) GetDefenseAdvice(features *core.FeatureVector, classification *ml.ClassificationResult) *defense.DefenseAdvice {
	return a.defense.Analyze(features, classification)
}

// ==================== Convenience functions ====================

// QuickAnalyze performs a quick analysis
func QuickAnalyze(headers *HTTPHeaders, method string) *QuickAnalyzeResult {
	// Extract features
	extractor := ml.NewFeatureExtractor()
	features := extractor.ExtractFromHTTPHeaders(headers)

	// Classify
	hc := ml.NewHierarchicalClassifier()
	hc.Initialize()
	classification := hc.Classify(features)

	// Calculate fingerprint
	ja4h := httpmod.CalculateJA4H(headers, method)

	riskLevel := RiskLevelMedium
	if classification.Family != "" {
		riskLevel = RiskLevelLow
	}
	return &QuickAnalyzeResult{
		Classification: classification,
		JA4H:           ja4h,
		RiskLevel:      riskLevel,
	}
}

// QuickAnalyzeResult represents a quick analysis result
type QuickAnalyzeResult struct {
	Classification *ml.ClassificationResult
	JA4H           *httpmod.JA4HResult
	RiskLevel      RiskLevel
}

// MatchProfile matches a fingerprint profile
func MatchProfile(headers *HTTPHeaders) (*profiles.ClientProfile, float64) {
	allProfiles := profiles.GetAll()
	if len(allProfiles) == 0 {
		return nil, 0
	}

	// Simple matching logic: compare User-Agent
	ua := headers.UserAgent
	if ua == "" {
		return nil, 0
	}

	for _, p := range allProfiles {
		if p.Headers != nil && p.Headers.UserAgent == ua {
			return &p, 1.0
		}
	}

	// Return the first one as the default
	return &allProfiles[0], 0.1
}

// GenerateUserAgent generates a User-Agent string
func GenerateUserAgent(browser BrowserType, version string, os OperatingSystem) string {
	switch browser {
	case BrowserChrome:
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36",
			os, version)
	case BrowserFirefox:
		return fmt.Sprintf("Mozilla/5.0 (%s; rv:%s) Gecko/20100101 Firefox/%s",
			os, version, version)
	case BrowserSafari:
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/%s Safari/605.1.15",
			os, version)
	default:
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36", os)
	}
}

// IsHeadless detects whether the browser is headless
func IsHeadless(features *core.FeatureVector) bool {
	// Determine based on features
	if features.Get(core.FeatureHeadlessBrowser) > 0.5 {
		return true
	}
	// Other detection logic...
	return false
}

// CalculateSimilarity calculates the similarity between two fingerprints
func CalculateSimilarity(a, b *core.FeatureVector) float64 {
	if a == nil || b == nil {
		return 0
	}

	var totalWeight, matchWeight float64
	weights := map[core.FeatureType]float64{
		core.FeatureTLSVersion:    1.0,
		core.FeatureCipherSuites:  1.0,
		core.FeatureExtensions:    0.8,
		core.FeatureHTTP2Settings: 0.6,
		core.FeatureHTTPHeaders:   0.7,
		core.FeatureUserAgent:     0.9,
	}

	for ft, weight := range weights {
		totalWeight += weight
		if a.Get(ft) == b.Get(ft) {
			matchWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0
	}
	return matchWeight / totalWeight
}

// ==================== HTTP handler ====================

// HTTPHandler creates an HTTP handler
func HTTPHandler() http.Handler {
	gw := gateway.NewGateway(nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/analyze", gw.HTTPHandler)
	mux.HandleFunc("/api/v1/sdk.js", gw.SDKHandler)
	mux.HandleFunc("/api/v1/collect", gw.CollectHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","version":"` + Version + `"}`))
	})

	return mux
}
