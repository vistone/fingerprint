// Package core provides core interface definitions
package core

// FingerprintSpec fingerprint specification interface
type FingerprintSpec interface {
	// GetID returns unique fingerprint identifier
	GetID() string
	// GetBrowserType returns browser type
	GetBrowserType() BrowserType
	// GetOS returns operating system
	GetOS() OperatingSystem
}

// TLSClient defines TLS client capabilities
type TLSClient interface {
	// GetClientHelloSpec returns ClientHello specification
	GetClientHelloSpec() (ClientHelloSpec, error)
	// GetJA3 returns JA3 fingerprint string
	GetJA3() string
	// GetJA4 returns JA4 fingerprint string
	GetJA4() string
}

// ClientHelloSpec TLS ClientHello specification
type ClientHelloSpec struct {
	CipherSuites       []uint16
	Extensions         []TLSExtension
	SupportedCurves    []CurveID
	SupportedPoints    []uint8
	TLSVersion         uint16
	CompressionMethods []uint8
}

// TLSExtension TLS extension definition
type TLSExtension struct {
	Type uint16
	Data []byte
}

// CurveID elliptic curve ID
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
	QUICVersion            uint32 // QUIC version (e.g., 0x00000001 for RFC 9000)
	InitialMaxData         uint64 // initial maximum data
	InitialMaxStreamData   uint64 // initial maximum stream data
	InitialMaxStreamsBidi  uint64 // initial maximum bidirectional streams
	InitialMaxStreamsUni   uint64 // initial maximum unidirectional streams
	MaxUDPPayloadSize      uint64 // maximum UDP payload size
	AckDelayExponent       uint8  // ACK delay exponent
	MaxAckDelay            uint16 // maximum ACK delay (ms)
	DisableActiveMigration bool   // disable active connection migration
}

// QUICVersion defines known QUIC versions
const (
	QUICVersion1       uint32 = 0x00000001 // RFC 9000
	QUICVersionDraft29 uint32 = 0xff00001d // Draft 29
	QUICVersionDraft30 uint32 = 0xff00001e // Draft 30
	QUICVersionDraft31 uint32 = 0xff00001f // Draft 31
	QUICVersionDraft32 uint32 = 0xff000020 // Draft 32
)

// FingerprintResult fingerprint result interface
type FingerprintResult interface {
	// GetUserAgent return User-Agent
	GetUserAgent() string
	// GetHeaders return HTTP Headers
	GetHeaders() *HTTPHeaders
	// GetSpec returns fingerprint specification
	GetSpec() FingerprintSpec
}

// ProtocolType protocoltype
type ProtocolType string

const (
	ProtocolTLS   ProtocolType = "tls"
	ProtocolHTTP  ProtocolType = "http"
	ProtocolHTTP2 ProtocolType = "http2"
	ProtocolQUIC  ProtocolType = "quic"
	ProtocolHTTP3 ProtocolType = "http3"
)

// FeatureType featuretype
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

// FeatureVector feature vector
type FeatureVector struct {
	Features map[FeatureType]float64
	Metadata map[string]interface{}
}

// NewFeatureVector creates new feature vector
func NewFeatureVector() *FeatureVector {
	return &FeatureVector{
		Features: make(map[FeatureType]float64),
		Metadata: make(map[string]interface{}),
	}
}

// Set sets feature value
func (fv *FeatureVector) Set(feature FeatureType, value float64) {
	fv.Features[feature] = value
}

// Get gets feature value
func (fv *FeatureVector) Get(feature FeatureType) float64 {
	return fv.Features[feature]
}

// ClassificationResult classifyresult
type ClassificationResult struct {
	Protocol   ProtocolType
	Family     BrowserType
	Version    string
	Confidence float64
	Labels     map[string]string
}

// RiskLevel risk level
type RiskLevel int

const (
	RiskLevelNone     RiskLevel = 0
	RiskLevelLow      RiskLevel = 1
	RiskLevelMedium   RiskLevel = 2
	RiskLevelHigh     RiskLevel = 3
	RiskLevelCritical RiskLevel = 4
)

// String returns string representation of risk level
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

// RiskAssessment risk assessment result
type RiskAssessment struct {
	Score       float64
	Level       RiskLevel
	Factors     []RiskFactor
	Suggestions []string
}

// RiskFactor risk factor
type RiskFactor struct {
	Name        string
	Weight      float64
	Description string
}

// RiskLevelFromScore calculates risk level from risk score
// score range: 0.0-1.0
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
