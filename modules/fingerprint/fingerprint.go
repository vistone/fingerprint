// translated comment
// translated comment
//
// translated comment
//
// translated comment
//
//	import "github.com/vistone/fingerprint"
//
// translated comment
//	result := fingerprint.GetRandom()
//
// translated comment
//	chrome := fingerprint.GetByBrowser(fingerprint.BrowserChrome)
//
// translated comment
//	analyzer := fingerprint.NewAnalyzer()
//	result := analyzer.Analyze(request)
//
// translated comment
//
// translated comment
// translated comment
// translated comment
// translated comment
// translated comment
// translated comment
// translated comment
// translated comment
//
package fingerprint

import (
	"context"
	"fmt"
	"net/http"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/defense"
	"github.com/vistone/fingerprint/modules/frontend"
	"github.com/vistone/fingerprint/modules/gateway"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
	"github.com/vistone/fingerprint/modules/tls"
	httpmod "github.com/vistone/fingerprint/modules/http"
)

// translated comment

// translated comment
type BrowserType = core.BrowserType

const (
	BrowserChrome  = core.BrowserChrome
	BrowserFirefox = core.BrowserFirefox
	BrowserSafari  = core.BrowserSafari
	BrowserOpera   = core.BrowserOpera
	BrowserEdge    = core.BrowserEdge
)

// translated comment
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

// ProtocolType protocoltype
type ProtocolType = core.ProtocolType

const (
	ProtocolTLS   = core.ProtocolTLS
	ProtocolHTTP  = core.ProtocolHTTP
	ProtocolHTTP2 = core.ProtocolHTTP2
	ProtocolQUIC  = core.ProtocolQUIC
	ProtocolHTTP3 = core.ProtocolHTTP3
)

// FeatureType featuretype
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

// translated comment
type RiskLevel = core.RiskLevel

const (
	RiskLevelNone     = core.RiskLevelNone
	RiskLevelLow      = core.RiskLevelLow
	RiskLevelMedium   = core.RiskLevelMedium
	RiskLevelHigh     = core.RiskLevelHigh
	RiskLevelCritical = core.RiskLevelCritical
)

// translated comment

// translated comment
type HTTPHeaders = core.HTTPHeaders

// translated comment
type FingerprintResult = core.FingerprintResult

// translated comment
type ClientHelloSpec = core.ClientHelloSpec

// translated comment
type TLSExtension = core.TLSExtension

// translated comment
type CurveID = core.CurveID

// HTTP2Settings HTTP/2 setting
type HTTP2Settings = core.HTTP2Settings

// translated comment
type HTTP2Priority = core.HTTP2Priority

// translated comment
type FeatureVector = core.FeatureVector

// translated comment
type RiskAssessment = core.RiskAssessment

// translated comment
type RiskFactor = core.RiskFactor

// translated comment

// translated comment
func GetRandom() *profiles.ClientProfile {
	p := profiles.GetRandom()
	return &p
}

// translated comment
func GetByBrowser(browser BrowserType) *profiles.ClientProfile {
	p := profiles.GetRandomByBrowser(browser)
	return &p
}

// translated comment
func GetAllProfiles() []profiles.ClientProfile {
	return profiles.GetAll()
}

// translated comment
func GetProfile(id string) (*profiles.ClientProfile, bool) {
	p, ok := profiles.Get(id)
	if !ok {
		return nil, false
	}
	return &p, true
}

// translated comment

// translated comment
func CalculateJA3(spec ClientHelloSpec) *tls.JA3Result {
	return tls.CalculateJA3(spec)
}

// translated comment
func CalculateJA4(spec ClientHelloSpec) *tls.JA4Result {
	return tls.CalculateJA4(spec)
}

// JA3Result JA3 result
type JA3Result = tls.JA3Result

// JA4Result JA4 result
type JA4Result = tls.JA4Result

// translated comment

// translated comment
func CalculateJA4H(headers *HTTPHeaders, method string) *httpmod.JA4HResult {
	return httpmod.CalculateJA4H(headers, method)
}

// JA4HResult JA4H result
type JA4HResult = httpmod.JA4HResult

// translated comment

// translated comment
type HierarchicalClassifier = ml.HierarchicalClassifier

// translated comment
func NewHierarchicalClassifier() *HierarchicalClassifier {
	hc := ml.NewHierarchicalClassifier()
	hc.Initialize()
	return hc
}

// ClassificationResult classifyresult
type ClassificationResult = ml.ClassificationResult

// translated comment
type FeatureExtractor = ml.FeatureExtractor

// translated comment
func NewFeatureExtractor() *FeatureExtractor {
	return ml.NewFeatureExtractor()
}

// translated comment

// translated comment
type DefenseSystem = defense.DefenseSystem

// translated comment
func NewDefenseSystem() *DefenseSystem {
	return defense.NewDefenseSystem()
}

// translated comment
type DetectionResult = defense.DetectionResult

// translated comment
type RiskEngine = defense.RiskEngine

// translated comment
func NewRiskEngine() *RiskEngine {
	return defense.NewRiskEngine()
}

// translated comment
type ProtectionConfig = defense.ProtectionConfig

// translated comment
var DefaultProtectionConfig = defense.DefaultProtectionConfig

// translated comment

// translated comment
type FrontendSDK = frontend.SDK

// translated comment
func NewFrontendSDK(config *frontend.SDKConfig) *FrontendSDK {
	return frontend.NewSDK(config)
}

// translated comment
var DefaultSDKConfig = frontend.DefaultSDKConfig

// translated comment
type FrontendFingerprintData = ml.FrontendFingerprintData

// translated comment

// translated comment
type Gateway = gateway.Gateway

// translated comment
type GatewayConfig = gateway.GatewayConfig

// translated comment
func NewGateway(config *GatewayConfig) *Gateway {
	return gateway.NewGateway(config)
}

// translated comment
var DefaultGatewayConfig = gateway.DefaultGatewayConfig

// AnalyzeRequest analyzerequest
type AnalyzeRequest = gateway.AnalyzeRequest

// AnalyzeResponse analyzeresponse
type AnalyzeResponse = gateway.AnalyzeResponse

// translated comment
func StartGateway(config *GatewayConfig) error {
	gw := NewGateway(config)
	return gw.Start()
}

// translated comment

// translated comment
type Analyzer struct {
	classifier *ml.HierarchicalClassifier
	extractor  *ml.FeatureExtractor
	riskEngine *defense.RiskEngine
	defense    *defense.DefenseSystem
}

// translated comment
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

// translated comment
func (a *Analyzer) Analyze(req *gateway.AnalyzeRequest) *AnalyzeResponse {
	ctx := context.Background()
	gw := gateway.NewGateway(nil)
	resp, _ := gw.Analyze(ctx, req)
	return resp
}

// Classify executeclassify
func (a *Analyzer) Classify(features *core.FeatureVector) *ml.ClassificationResult {
	return a.classifier.Classify(features)
}

// ExtractFeatures fromconfigurationextractfeature
func (a *Analyzer) ExtractFeatures(profile *profiles.ClientProfile) *core.FeatureVector {
	return a.extractor.ExtractFromProfile(profile)
}

// translated comment
func (a *Analyzer) EvaluateRisk(features *core.FeatureVector, classification *ml.ClassificationResult) *core.RiskAssessment {
	return a.riskEngine.Evaluate(features, classification)
}

// translated comment
func (a *Analyzer) GetDefenseAdvice(features *core.FeatureVector, classification *ml.ClassificationResult) *defense.DefenseAdvice {
	return a.defense.Analyze(features, classification)
}

// translated comment

// translated comment
func QuickAnalyze(headers *HTTPHeaders, method string) *QuickAnalyzeResult {
	// extractfeature
	extractor := ml.NewFeatureExtractor()
	features := extractor.ExtractFromHTTPHeaders(headers)

	// classify
	hc := ml.NewHierarchicalClassifier()
	hc.Initialize()
	classification := hc.Classify(features)

	// translated comment
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

// translated comment
type QuickAnalyzeResult struct {
	Classification *ml.ClassificationResult
	JA4H           *httpmod.JA4HResult
	RiskLevel      RiskLevel
}

// translated comment
func MatchProfile(headers *HTTPHeaders) (*profiles.ClientProfile, float64) {
	profiles := profiles.GetAll()
	if len(profiles) == 0 {
		return nil, 0
	}

	// translated comment
	ua := headers.UserAgent
	if ua == "" {
		return nil, 0
	}

	for _, p := range profiles {
		if p.Headers != nil && p.Headers.UserAgent == ua {
			return &p, 1.0
		}
	}

	// translated comment
	return &profiles[0], 0.1
}

// GenerateUserAgent generate User-Agent
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

// translated comment
func IsHeadless(features *core.FeatureVector) bool {
	// translated comment
	if features.Get(core.FeatureHeadlessBrowser) > 0.5 {
		return true
	}
	// translated comment
	return false
}

// translated comment
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

// translated comment

// translated comment
func HTTPHandler() http.Handler {
	gw := gateway.NewGateway(nil)
	
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/analyze", gw.HTTPHandler)
	mux.HandleFunc("/api/v1/sdk.js", gw.SDKHandler)
	mux.HandleFunc("/api/v1/collect", gw.CollectHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","version":"2.1.0"}`))
	})
	
	return mux
}


