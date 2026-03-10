// Package client provides real browser fingerprint simulation and request tracing
package client

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/vistone/fingerprint/modules/profiles"
)

// RequestTrace contains request tracing information - showing the actual fingerprint sent to server
type RequestTrace struct {
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"requestId"`
	TargetURL string    `json:"targetUrl"`
	Method    string    `json:"method"`

	// TCP/IP layer fingerprint - actual TCP parameters used
	TCPIP *TCPIPFingerprint `json:"tcpip"`

	// TLS layer fingerprint - actual ClientHello sent
	TLS *TLSFingerprint `json:"tls"`

	// HTTP layer fingerprint - actual request headers sent
	HTTP *HTTPFingerprint `json:"http"`

	// Connection information
	Connection *ConnectionInfo `json:"connection"`
}

// TCPIPFingerprint contains TCP/IP layer fingerprint details
type TCPIPFingerprint struct {
	TTL           uint8  `json:"ttl"`           // Time To Live
	WindowSize    uint16 `json:"windowSize"`    // TCP Window Size
	MSS           uint16 `json:"mss"`           // Maximum Segment Size
	WindowScale   uint8  `json:"windowScale"`   // Window Scale Option
	DF            bool   `json:"df"`            // Don't Fragment flag
	SackPermitted bool   `json:"sackPermitted"` // SACK permitted
	Timestamps    bool   `json:"timestamps"`    // TCP Timestamps
	JA4T          string `json:"ja4t"`          // JA4T fingerprint hash
}

// TLSFingerprint contains TLS layer fingerprint details
type TLSFingerprint struct {
	Version         string   `json:"version"`         // TLS version
	JA3             string   `json:"ja3"`             // JA3 fingerprint
	JA3Hash         string   `json:"ja3Hash"`         // JA3 hash
	CipherSuites    []string `json:"cipherSuites"`    // Cipher suites list
	Extensions      []string `json:"extensions"`      // Extensions list
	SupportedGroups []string `json:"supportedGroups"` // Supported curves
	ECPointFormats  []string `json:"ecPointFormats"`  // EC point formats
	ALPNProtocols   []string `json:"alpnProtocols"`   // ALPN protocols
	ClientHelloID   string   `json:"clientHelloId"`   // ClientHello identifier
}

// HTTPFingerprint contains HTTP layer fingerprint details
type HTTPFingerprint struct {
	Protocol       string              `json:"protocol"`       // HTTP/1.1, HTTP/2, HTTP/3
	Headers        map[string]string   `json:"headers"`        // Actual request headers sent
	HeaderOrder    []string            `json:"headerOrder"`    // Request header order
	PseudoHeaders  []string            `json:"pseudoHeaders"`  // HTTP/2 pseudo header order
	HTTP2Settings  *HTTP2SettingsTrace `json:"http2Settings"`  // HTTP/2 settings frame
	HTTP3Settings  *HTTP3SettingsTrace `json:"http3Settings"`  // HTTP/3 settings
	UserAgent      string              `json:"userAgent"`      // User-Agent
	Accept         string              `json:"accept"`         // Accept
	AcceptLanguage string              `json:"acceptLanguage"` // Accept-Language
	AcceptEncoding string              `json:"acceptEncoding"` // Accept-Encoding
}

// HTTP3SettingsTrace contains HTTP/3 settings tracing
type HTTP3SettingsTrace struct {
	QUICVersion    uint32 `json:"quicVersion"`
	InitialMaxData uint64 `json:"initialMaxData"`
	MaxStreamsBidi uint64 `json:"maxStreamsBidi"`
	MaxStreamsUni  uint64 `json:"maxStreamsUni"`
	MaxUDPPayload  uint64 `json:"maxUdpPayload"`
	Active         bool   `json:"active"`
}

// HTTP2SettingsTrace contains HTTP/2 settings tracing
type HTTP2SettingsTrace struct {
	HeaderTableSize      uint32 `json:"headerTableSize"`
	EnablePush           uint32 `json:"enablePush"`
	MaxConcurrentStreams uint32 `json:"maxConcurrentStreams"`
	InitialWindowSize    uint32 `json:"initialWindowSize"`
	MaxFrameSize         uint32 `json:"maxFrameSize"`
	MaxHeaderListSize    uint32 `json:"maxHeaderListSize"`
	ConnectionFlow       uint32 `json:"connectionFlow"`
	PriorityFrames       int    `json:"priorityFrames"`
}

// ConnectionInfo contains connection information
type ConnectionInfo struct {
	LocalAddr     string        `json:"localAddr"`
	RemoteAddr    string        `json:"remoteAddr"`
	ConnectTime   time.Duration `json:"connectTime"`
	HandshakeTime time.Duration `json:"handshakeTime"`
	TotalTime     time.Duration `json:"totalTime"`
	Protocol      string        `json:"protocol"`     // h2, http/1.1, h3
	ProtocolUsed  string        `json:"protocolUsed"` // Actual protocol used
	ALPN          string        `json:"alpn"`         // Negotiated ALPN
	TLSVersion    string        `json:"tlsVersion"`
	CipherSuite   string        `json:"cipherSuite"`
}

// ResponseTrace contains response tracing information
type ResponseTrace struct {
	StatusCode   int               `json:"statusCode"`
	Status       string            `json:"status"`
	Protocol     string            `json:"protocol"`
	Headers      map[string]string `json:"headers"`
	BodyPreview  string            `json:"bodyPreview"`
	BodyLength   int               `json:"bodyLength"`
	ResponseTime time.Duration     `json:"responseTime"`
}

// ProxyResult contains complete proxy request result
type ProxyResult struct {
	Success       bool                   `json:"success"`
	Error         string                 `json:"error,omitempty"`
	ErrorType     string                 `json:"errorType,omitempty"`    // Error classification: timeout, network, protocol, etc
	ErrorCode     int                    `json:"errorCode,omitempty"`    // HTTP status code or error code
	ErrorDetails  map[string]interface{} `json:"errorDetails,omitempty"` // Detailed error information
	RequestTrace  *RequestTrace          `json:"requestTrace"`
	ResponseTrace *ResponseTrace         `json:"responseTrace"`
	ProfileUsed   *ProfileInfo           `json:"profileUsed"`
}

// ProfileInfo contains information about the profile used
type ProfileInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	BrowserType    string `json:"browserType"`
	BrowserVersion string `json:"browserVersion"`
	OS             string `json:"os"`
	OSVersion      string `json:"osVersion"`
}

// TracedClient is a client with request tracing capability
type TracedClient struct {
	profile profiles.ClientProfile
	tracer  *RequestTracer
}

// RequestTracer is the request tracer
type RequestTracer struct {
	Trace *RequestTrace
}

// NewRequestTracer creates a new request tracer
func NewRequestTracer(profile profiles.ClientProfile, url, method string) *RequestTracer {
	return &RequestTracer{
		Trace: &RequestTrace{
			Timestamp: time.Now(),
			RequestID: generateRequestID(),
			TargetURL: url,
			Method:    method,
			TCPIP:     buildTCPIPFingerprint(profile),
			TLS:       buildTLSFingerprint(profile),
			HTTP:      buildHTTPFingerprint(profile),
		},
	}
}

// buildTCPIPFingerprint builds TCP/IP fingerprint information
func buildTCPIPFingerprint(profile profiles.ClientProfile) *TCPIPFingerprint {
	if profile.TCPIP == nil {
		return &TCPIPFingerprint{
			TTL:         128,
			WindowSize:  64240,
			MSS:         1460,
			WindowScale: 8,
			DF:          true,
		}
	}

	tcpip := profile.TCPIP
	return &TCPIPFingerprint{
		TTL:           tcpip.TTL,
		WindowSize:    tcpip.WindowSize,
		MSS:           tcpip.MSS,
		WindowScale:   tcpip.WindowScale,
		DF:            tcpip.DF,
		SackPermitted: tcpip.SAckPermitted,
		Timestamps:    tcpip.Timestamps,
		JA4T:          tcpip.JA4T,
	}
}

// buildTLSFingerprint builds TLS fingerprint information
func buildTLSFingerprint(profile profiles.ClientProfile) *TLSFingerprint {
	tls := &TLSFingerprint{
		Version:       formatTLSVersion(profile.TLSVersion),
		CipherSuites:  make([]string, 0, len(profile.CipherSuites)),
		Extensions:    make([]string, 0, len(profile.Extensions)),
		JA3:           calculateJA3(profile),
		JA3Hash:       calculateJA3Hash(profile),
		ClientHelloID: string(profile.BrowserType),
	}

	// Format cipher suites
	for _, cs := range profile.CipherSuites {
		tls.CipherSuites = append(tls.CipherSuites, formatCipherSuite(cs))
	}

	// Format extensions
	for _, ext := range profile.Extensions {
		tls.Extensions = append(tls.Extensions, formatExtension(ext))
	}

	// Format supported curves
	for _, curve := range profile.SupportedCurves {
		tls.SupportedGroups = append(tls.SupportedGroups, formatCurve(uint16(curve)))
	}

	// ALPN protocols
	tls.ALPNProtocols = []string{"h2", "http/1.1"}

	return tls
}

// buildHTTPFingerprint builds HTTP fingerprint information
func buildHTTPFingerprint(profile profiles.ClientProfile) *HTTPFingerprint {
	http := &HTTPFingerprint{
		Protocol:    "HTTP/2",
		Headers:     make(map[string]string),
		HeaderOrder: getHeaderOrder(profile),
	}

	if profile.Headers != nil {
		h := profile.Headers
		http.UserAgent = h.UserAgent
		http.Accept = h.Accept
		http.AcceptLanguage = h.AcceptLanguage
		http.AcceptEncoding = h.AcceptEncoding

		http.Headers["User-Agent"] = h.UserAgent
		http.Headers["Accept"] = h.Accept
		http.Headers["Accept-Language"] = h.AcceptLanguage
		http.Headers["Accept-Encoding"] = h.AcceptEncoding
		http.Headers["Sec-Fetch-Site"] = h.SecFetchSite
		http.Headers["Sec-Fetch-Mode"] = h.SecFetchMode
		http.Headers["Sec-Fetch-Dest"] = h.SecFetchDest

		if h.SecCHUA != "" {
			http.Headers["Sec-CH-UA"] = h.SecCHUA
		}
		if h.SecCHUAMobile != "" {
			http.Headers["Sec-CH-UA-Mobile"] = h.SecCHUAMobile
		}
		if h.SecCHUAPlatform != "" {
			http.Headers["Sec-CH-UA-Platform"] = h.SecCHUAPlatform
		}
	}

	// HTTP/2 pseudo header order
	http.PseudoHeaders = profile.PseudoHeaderOrder
	if len(http.PseudoHeaders) == 0 {
		http.PseudoHeaders = []string{":method", ":authority", ":scheme", ":path"}
	}

	// HTTP/2 settings
	http.HTTP2Settings = &HTTP2SettingsTrace{
		HeaderTableSize:      profile.HTTP2Settings.HeaderTableSize,
		EnablePush:           profile.HTTP2Settings.EnablePush,
		MaxConcurrentStreams: profile.HTTP2Settings.MaxConcurrentStreams,
		InitialWindowSize:    profile.HTTP2Settings.InitialWindowSize,
		MaxFrameSize:         profile.HTTP2Settings.MaxFrameSize,
		MaxHeaderListSize:    profile.HTTP2Settings.MaxHeaderListSize,
		ConnectionFlow:       profile.ConnectionFlow,
		PriorityFrames:       len(profile.HTTP2Priorities),
	}

	return http
}

// Helper functions

func generateRequestID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("req-%x", timestamp)
}

func formatTLSVersion(version uint16) string {
	switch version {
	case 0x0301:
		return "TLS 1.0"
	case 0x0302:
		return "TLS 1.1"
	case 0x0303:
		return "TLS 1.2"
	case 0x0304:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("0x%04X", version)
	}
}

func formatCipherSuite(cs uint16) string {
	switch cs {
	case 0x1301:
		return "TLS_AES_128_GCM_SHA256"
	case 0x1302:
		return "TLS_AES_256_GCM_SHA384"
	case 0x1303:
		return "TLS_CHACHA20_POLY1305_SHA256"
	case 0x002f:
		return "TLS_RSA_WITH_AES_128_CBC_SHA"
	case 0x0035:
		return "TLS_RSA_WITH_AES_256_CBC_SHA"
	case 0xc02b:
		return "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
	case 0xc02f:
		return "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	case 0xc02c:
		return "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
	case 0xc030:
		return "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
	case 0xcca8:
		return "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"
	case 0xcca9:
		return "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256"
	default:
		return fmt.Sprintf("0x%04X", cs)
	}
}

func formatExtension(ext interface{}) string {
	switch e := ext.(type) {
	case string:
		return e
	default:
		return fmt.Sprintf("%v", e)
	}
}

func formatCurve(curve uint16) string {
	switch curve {
	case 0x0017:
		return "secp256r1 (P-256)"
	case 0x0018:
		return "secp384r1 (P-384)"
	case 0x0019:
		return "secp521r1 (P-521)"
	case 0x001d:
		return "X25519"
	case 0x001e:
		return "X448"
	default:
		return fmt.Sprintf("0x%04X", curve)
	}
}

func calculateJA3(profile profiles.ClientProfile) string {
	// JA3 fingerprint format: SSLVersion,Cipher,SSLExtension,EllipticCurve,EllipticCurvePointFormat
	parts := make([]string, 5)

	// SSL Version
	parts[0] = fmt.Sprintf("%d", profile.TLSVersion)

	// Cipher Suites
	csStrs := make([]string, len(profile.CipherSuites))
	for i, cs := range profile.CipherSuites {
		csStrs[i] = fmt.Sprintf("%d", cs)
	}
	parts[1] = strings.Join(csStrs, "-")

	// Extensions
	extStrs := make([]string, len(profile.Extensions))
	for i, ext := range profile.Extensions {
		extStrs[i] = fmt.Sprintf("%v", ext)
	}
	parts[2] = strings.Join(extStrs, "-")

	// Elliptic Curves
	curveStrs := make([]string, len(profile.SupportedCurves))
	for i, curve := range profile.SupportedCurves {
		curveStrs[i] = fmt.Sprintf("%d", curve)
	}
	parts[3] = strings.Join(curveStrs, "-")

	// EC Point Formats
	parts[4] = "0" // Default uncompressed

	return strings.Join(parts, ",")
}

func calculateJA3Hash(profile profiles.ClientProfile) string {
	ja3 := calculateJA3(profile)
	hash := md5.Sum([]byte(ja3))
	return fmt.Sprintf("%x", hash[:])
}

func getHeaderOrder(profile profiles.ClientProfile) []string {
	// Standard Chrome request header order
	return []string{
		":method",
		":authority",
		":scheme",
		":path",
		"sec-ch-ua",
		"sec-ch-ua-mobile",
		"sec-ch-ua-platform",
		"upgrade-insecure-requests",
		"user-agent",
		"accept",
		"sec-fetch-site",
		"sec-fetch-mode",
		"sec-fetch-user",
		"sec-fetch-dest",
		"accept-encoding",
		"accept-language",
		"cookie",
	}
}
