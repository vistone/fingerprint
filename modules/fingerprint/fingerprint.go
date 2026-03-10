// Package fingerprint 是主 facade 包
// 提供统一的 API interface，整合所有子模块功能
//
// 这是 Go Workspace 重构后的主入口点，参考 Rust 版本的 workspace 架构设计
//
// 基本用法：
//
//	import "github.com/vistone/fingerprint"
//
//	// get随机指纹
//	result := fingerprint.GetRandom()
//
//	// get指定浏览器的指纹
//	chrome := fingerprint.GetByBrowser(fingerprint.BrowserChrome)
//
//	// execute指纹analyze
//	analyzer := fingerprint.NewAnalyzer()
//	result := analyzer.Analyze(request)
//
// 模块结构：
//
//	github.com/vistone/fingerprint/core     - 核心typeandinterface
//	github.com/vistone/fingerprint/profiles - 浏览器指纹configuration
//	github.com/vistone/fingerprint/tls      - TLS 指纹generate
//	github.com/vistone/fingerprint/http     - HTTP 指纹generate
//	github.com/vistone/fingerprint/ml       - ML classify器
//	github.com/vistone/fingerprint/defense  - 安全防护
//	github.com/vistone/fingerprint/frontend - 前端 SDK
//	github.com/vistone/fingerprint/gateway  - API 网关
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

// ==================== type别名 ====================

// BrowserType 浏览器type
type BrowserType = core.BrowserType

const (
	BrowserChrome  = core.BrowserChrome
	BrowserFirefox = core.BrowserFirefox
	BrowserSafari  = core.BrowserSafari
	BrowserOpera   = core.BrowserOpera
	BrowserEdge    = core.BrowserEdge
)

// OperatingSystem 操作系统type
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

// RiskLevel 风险等级
type RiskLevel = core.RiskLevel

const (
	RiskLevelNone     = core.RiskLevelNone
	RiskLevelLow      = core.RiskLevelLow
	RiskLevelMedium   = core.RiskLevelMedium
	RiskLevelHigh     = core.RiskLevelHigh
	RiskLevelCritical = core.RiskLevelCritical
)

// ==================== 核心type别名 ====================

// HTTPHeaders HTTP 头
type HTTPHeaders = core.HTTPHeaders

// FingerprintResult 指纹result
type FingerprintResult = core.FingerprintResult

// ClientHelloSpec ClientHello 规范
type ClientHelloSpec = core.ClientHelloSpec

// TLSExtension TLS 扩展
type TLSExtension = core.TLSExtension

// CurveID 曲线 ID
type CurveID = core.CurveID

// HTTP2Settings HTTP/2 setting
type HTTP2Settings = core.HTTP2Settings

// HTTP2Priority HTTP/2 优先级
type HTTP2Priority = core.HTTP2Priority

// FeatureVector feature向量
type FeatureVector = core.FeatureVector

// RiskAssessment 风险评估
type RiskAssessment = core.RiskAssessment

// RiskFactor 风险因子
type RiskFactor = core.RiskFactor

// ==================== 核心函数 ====================

// GetRandom get随机指纹
func GetRandom() *profiles.ClientProfile {
	p := profiles.GetRandom()
	return &p
}

// GetByBrowser 按浏览器typeget随机指纹
func GetByBrowser(browser BrowserType) *profiles.ClientProfile {
	p := profiles.GetRandomByBrowser(browser)
	return &p
}

// GetAllProfiles get所有指纹configuration
func GetAllProfiles() []profiles.ClientProfile {
	return profiles.GetAll()
}

// GetProfile get指定 ID 的指纹configuration
func GetProfile(id string) (*profiles.ClientProfile, bool) {
	p, ok := profiles.Get(id)
	if !ok {
		return nil, false
	}
	return &p, true
}

// ==================== TLS 指纹 ====================

// CalculateJA3 calculate JA3 指纹
func CalculateJA3(spec ClientHelloSpec) *tls.JA3Result {
	return tls.CalculateJA3(spec)
}

// CalculateJA4 calculate JA4 指纹
func CalculateJA4(spec ClientHelloSpec) *tls.JA4Result {
	return tls.CalculateJA4(spec)
}

// JA3Result JA3 result
type JA3Result = tls.JA3Result

// JA4Result JA4 result
type JA4Result = tls.JA4Result

// ==================== HTTP 指纹 ====================

// CalculateJA4H calculate JA4H 指纹
func CalculateJA4H(headers *HTTPHeaders, method string) *httpmod.JA4HResult {
	return httpmod.CalculateJA4H(headers, method)
}

// JA4HResult JA4H result
type JA4HResult = httpmod.JA4HResult

// ==================== ML classify器 ====================

// HierarchicalClassifier 三层分层classify器
type HierarchicalClassifier = ml.HierarchicalClassifier

// NewHierarchicalClassifier create新的分层classify器
func NewHierarchicalClassifier() *HierarchicalClassifier {
	hc := ml.NewHierarchicalClassifier()
	hc.Initialize()
	return hc
}

// ClassificationResult classifyresult
type ClassificationResult = ml.ClassificationResult

// FeatureExtractor featureextract器
type FeatureExtractor = ml.FeatureExtractor

// NewFeatureExtractor create新的featureextract器
func NewFeatureExtractor() *FeatureExtractor {
	return ml.NewFeatureExtractor()
}

// ==================== 安全防护 ====================

// DefenseSystem 防护系统
type DefenseSystem = defense.DefenseSystem

// NewDefenseSystem create新的防护系统
func NewDefenseSystem() *DefenseSystem {
	return defense.NewDefenseSystem()
}

// DetectionResult 检测result
type DetectionResult = defense.DetectionResult

// RiskEngine 风险引擎
type RiskEngine = defense.RiskEngine

// NewRiskEngine create新的风险引擎
func NewRiskEngine() *RiskEngine {
	return defense.NewRiskEngine()
}

// ProtectionConfig 防护configuration
type ProtectionConfig = defense.ProtectionConfig

// DefaultProtectionConfig 默认防护configuration
var DefaultProtectionConfig = defense.DefaultProtectionConfig

// ==================== 前端 SDK ====================

// FrontendSDK 前端 SDK
type FrontendSDK = frontend.SDK

// NewFrontendSDK create新的前端 SDK
func NewFrontendSDK(config *frontend.SDKConfig) *FrontendSDK {
	return frontend.NewSDK(config)
}

// DefaultSDKConfig 默认 SDK configuration
var DefaultSDKConfig = frontend.DefaultSDKConfig

// FrontendFingerprintData 前端指纹data
type FrontendFingerprintData = ml.FrontendFingerprintData

// ==================== 网关 ====================

// Gateway 网关
type Gateway = gateway.Gateway

// GatewayConfig 网关configuration
type GatewayConfig = gateway.GatewayConfig

// NewGateway create新的网关
func NewGateway(config *GatewayConfig) *Gateway {
	return gateway.NewGateway(config)
}

// DefaultGatewayConfig 默认网关configuration
var DefaultGatewayConfig = gateway.DefaultGatewayConfig

// AnalyzeRequest analyzerequest
type AnalyzeRequest = gateway.AnalyzeRequest

// AnalyzeResponse analyzeresponse
type AnalyzeResponse = gateway.AnalyzeResponse

// StartGateway start网关服务
func StartGateway(config *GatewayConfig) error {
	gw := NewGateway(config)
	return gw.Start()
}

// ==================== analyze器 ====================

// Analyzer 综合analyze器
type Analyzer struct {
	classifier *ml.HierarchicalClassifier
	extractor  *ml.FeatureExtractor
	riskEngine *defense.RiskEngine
	defense    *defense.DefenseSystem
}

// NewAnalyzer create新的analyze器
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

// Analyze execute综合analyze
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

// EvaluateRisk 评估风险
func (a *Analyzer) EvaluateRisk(features *core.FeatureVector, classification *ml.ClassificationResult) *core.RiskAssessment {
	return a.riskEngine.Evaluate(features, classification)
}

// GetDefenseAdvice get防护建议
func (a *Analyzer) GetDefenseAdvice(features *core.FeatureVector, classification *ml.ClassificationResult) *defense.DefenseAdvice {
	return a.defense.Analyze(features, classification)
}

// ==================== 便捷函数 ====================

// QuickAnalyze 快速analyze
func QuickAnalyze(headers *HTTPHeaders, method string) *QuickAnalyzeResult {
	// extractfeature
	extractor := ml.NewFeatureExtractor()
	features := extractor.ExtractFromHTTPHeaders(headers)

	// classify
	hc := ml.NewHierarchicalClassifier()
	hc.Initialize()
	classification := hc.Classify(features)

	// calculate指纹
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

// QuickAnalyzeResult 快速analyzeresult
type QuickAnalyzeResult struct {
	Classification *ml.ClassificationResult
	JA4H           *httpmod.JA4HResult
	RiskLevel      RiskLevel
}

// MatchProfile match指纹configuration
func MatchProfile(headers *HTTPHeaders) (*profiles.ClientProfile, float64) {
	profiles := profiles.GetAll()
	if len(profiles) == 0 {
		return nil, 0
	}

	// 简单的match逻辑：compare User-Agent
	ua := headers.UserAgent
	if ua == "" {
		return nil, 0
	}

	for _, p := range profiles {
		if p.Headers != nil && p.Headers.UserAgent == ua {
			return &p, 1.0
		}
	}

	// return第一个作为默认值
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

// IsHeadless 检测whether为无头浏览器
func IsHeadless(features *core.FeatureVector) bool {
	// 基于feature判断
	if features.Get(core.FeatureHeadlessBrowser) > 0.5 {
		return true
	}
	// 其他检测逻辑...
	return false
}

// CalculateSimilarity calculate两个指纹的相似度
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

// ==================== HTTP process器 ====================

// HTTPHandler create HTTP process器
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


