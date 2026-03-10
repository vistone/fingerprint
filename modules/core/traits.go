// Package core 提供核心接口定义
package core

// FingerprintSpec 指纹规范接口
type FingerprintSpec interface {
	// GetID 返回指纹唯一标识
	GetID() string
	// GetBrowserType 返回浏览器类型
	GetBrowserType() BrowserType
	// GetOS 返回操作系统
	GetOS() OperatingSystem
}

// TLSClient 定义 TLS 客户端能力
type TLSClient interface {
	// GetClientHelloSpec 返回 ClientHello 规范
	GetClientHelloSpec() (ClientHelloSpec, error)
	// GetJA3 返回 JA3 指纹字符串
	GetJA3() string
	// GetJA4 返回 JA4 指纹字符串
	GetJA4() string
}

// ClientHelloSpec TLS ClientHello 规范
type ClientHelloSpec struct {
	CipherSuites       []uint16
	Extensions         []TLSExtension
	SupportedCurves    []CurveID
	SupportedPoints    []uint8
	TLSVersion         uint16
	CompressionMethods []uint8
}

// TLSExtension TLS 扩展定义
type TLSExtension struct {
	Type uint16
	Data []byte
}

// CurveID 椭圆曲线 ID
type CurveID uint16

const (
	CurveX25519    CurveID = 0x001d
	CurveP256      CurveID = 0x0017
	CurveP384      CurveID = 0x0018
	CurveP521      CurveID = 0x0019
	CurveP256Kyber CurveID = 0x6399
)

// HTTP2Settings HTTP/2 Settings
type HTTP2Settings struct {
	HeaderTableSize      uint32
	EnablePush           uint32
	MaxConcurrentStreams uint32
	InitialWindowSize    uint32
	MaxFrameSize         uint32
	MaxHeaderListSize    uint32
}

// HTTP2Priority HTTP/2 Priority
type HTTP2Priority struct {
	StreamID  uint32
	Weight    uint8
	DependsOn uint32
	Exclusive bool
}

// HTTP3Settings HTTP/3 (QUIC) Settings
type HTTP3Settings struct {
	QUICVersion            uint32 // QUIC 版本 (e.g., 0x00000001 for RFC 9000)
	InitialMaxData         uint64 // 初始最大数据量
	InitialMaxStreamData   uint64 // 初始最大流数据量
	InitialMaxStreamsBidi  uint64 // 初始最大双向流数
	InitialMaxStreamsUni   uint64 // 初始最大单向流数
	MaxUDPPayloadSize      uint64 // 最大 UDP 负载大小
	AckDelayExponent       uint8  // ACK 延迟指数
	MaxAckDelay            uint16 // 最大 ACK 延迟 (ms)
	DisableActiveMigration bool   // 禁用连接迁移
}

// QUICVersion 定义已知的 QUIC 版本
const (
	QUICVersion1       uint32 = 0x00000001 // RFC 9000
	QUICVersionDraft29 uint32 = 0xff00001d // Draft 29
	QUICVersionDraft30 uint32 = 0xff00001e // Draft 30
	QUICVersionDraft31 uint32 = 0xff00001f // Draft 31
	QUICVersionDraft32 uint32 = 0xff000020 // Draft 32
)

// FingerprintResult 指纹结果接口
type FingerprintResult interface {
	// GetUserAgent 返回 User-Agent
	GetUserAgent() string
	// GetHeaders 返回 HTTP Headers
	GetHeaders() *HTTPHeaders
	// GetSpec 返回指纹规范
	GetSpec() FingerprintSpec
}

// ProtocolType 协议类型
type ProtocolType string

const (
	ProtocolTLS   ProtocolType = "tls"
	ProtocolHTTP  ProtocolType = "http"
	ProtocolHTTP2 ProtocolType = "http2"
	ProtocolQUIC  ProtocolType = "quic"
	ProtocolHTTP3 ProtocolType = "http3"
)

// FeatureType 特征类型
type FeatureType string

const (
	FeatureTLSVersion      FeatureType = "tls_version"
	FeatureCipherSuites    FeatureType = "cipher_suites"
	FeatureExtensions      FeatureType = "extensions"
	FeatureHTTP2Settings   FeatureType = "http2_settings"
	FeatureHTTPHeaders     FeatureType = "http_headers"
	FeatureUserAgent       FeatureType = "user_agent"
	FeatureCanvas          FeatureType = "canvas"
	FeatureWebGL           FeatureType = "webgl"
	FeatureAudio           FeatureType = "audio"
	FeatureFonts           FeatureType = "fonts"
	FeatureStorage         FeatureType = "storage"
	FeatureWebRTC          FeatureType = "webrtc"
	FeatureHardware        FeatureType = "hardware"
	FeatureTiming          FeatureType = "timing"
	FeatureHeadlessBrowser FeatureType = "headless_browser"
	FeatureEntropy         FeatureType = "entropy"
	FeatureToolMarker      FeatureType = "tool_marker"
	FeatureBehaviorPattern FeatureType = "behavior_pattern"
)

// FeatureVector 特征向量
type FeatureVector struct {
	Features map[FeatureType]float64
	Metadata map[string]interface{}
}

// NewFeatureVector 创建新的特征向量
func NewFeatureVector() *FeatureVector {
	return &FeatureVector{
		Features: make(map[FeatureType]float64),
		Metadata: make(map[string]interface{}),
	}
}

// Set 设置特征值
func (fv *FeatureVector) Set(feature FeatureType, value float64) {
	fv.Features[feature] = value
}

// Get 获取特征值
func (fv *FeatureVector) Get(feature FeatureType) float64 {
	return fv.Features[feature]
}

// ClassificationResult 分类结果
type ClassificationResult struct {
	Protocol   ProtocolType
	Family     BrowserType
	Version    string
	Confidence float64
	Labels     map[string]string
}

// RiskLevel 风险等级
type RiskLevel int

const (
	RiskLevelNone     RiskLevel = 0
	RiskLevelLow      RiskLevel = 1
	RiskLevelMedium   RiskLevel = 2
	RiskLevelHigh     RiskLevel = 3
	RiskLevelCritical RiskLevel = 4
)

// String 返回风险等级的字符串表示
func (r RiskLevel) String() string {
	switch r {
	case RiskLevelNone:
		return "none"
	case RiskLevelLow:
		return "low"
	case RiskLevelMedium:
		return "medium"
	case RiskLevelHigh:
		return "high"
	case RiskLevelCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// RiskAssessment 风险评估结果
type RiskAssessment struct {
	Score       float64
	Level       RiskLevel
	Factors     []RiskFactor
	Suggestions []string
}

// RiskFactor 风险因子
type RiskFactor struct {
	Name        string
	Weight      float64
	Description string
}

// RiskLevelFromScore 根据风险分数计算风险等级
// 分数范围: 0.0-1.0
func RiskLevelFromScore(score float64) RiskLevel {
	switch {
	case score >= 0.9:
		return RiskLevelCritical
	case score >= 0.7:
		return RiskLevelHigh
	case score >= 0.4:
		return RiskLevelMedium
	case score >= 0.1:
		return RiskLevelLow
	default:
		return RiskLevelNone
	}
}
