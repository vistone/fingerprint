// Package client 提供真实的浏览器指纹模拟和请求追踪
package client

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/vistone/fingerprint/modules/profiles"
)

// RequestTrace 请求追踪信息 - 显示实际发送给服务器的指纹
type RequestTrace struct {
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"requestId"`
	TargetURL string    `json:"targetUrl"`
	Method    string    `json:"method"`

	// TCP/IP 层指纹 - 实际使用的 TCP 参数
	TCPIP *TCPIPFingerprint `json:"tcpip"`

	// TLS 层指纹 - 实际发送的 ClientHello
	TLS *TLSFingerprint `json:"tls"`

	// HTTP 层指纹 - 实际发送的请求头
	HTTP *HTTPFingerprint `json:"http"`

	// 连接信息
	Connection *ConnectionInfo `json:"connection"`
}

// TCPIPFingerprint TCP/IP 层指纹详情
type TCPIPFingerprint struct {
	TTL           uint8  `json:"ttl"`           // Time To Live
	WindowSize    uint16 `json:"windowSize"`    // TCP Window Size
	MSS           uint16 `json:"mss"`           // Maximum Segment Size
	WindowScale   uint8  `json:"windowScale"`   // Window Scale Option
	DF            bool   `json:"df"`            // Don't Fragment 标志
	SackPermitted bool   `json:"sackPermitted"` // SACK 允许
	Timestamps    bool   `json:"timestamps"`    // TCP Timestamps
	JA4T          string `json:"ja4t"`          // JA4T 指纹哈希
}

// TLSFingerprint TLS 层指纹详情
type TLSFingerprint struct {
	Version         string   `json:"version"`         // TLS 版本
	JA3             string   `json:"ja3"`             // JA3 指纹
	JA3Hash         string   `json:"ja3Hash"`         // JA3 哈希
	CipherSuites    []string `json:"cipherSuites"`    // 加密套件列表
	Extensions      []string `json:"extensions"`      // 扩展列表
	SupportedGroups []string `json:"supportedGroups"` // 支持的曲线
	ECPointFormats  []string `json:"ecPointFormats"`  // EC 点格式
	ALPNProtocols   []string `json:"alpnProtocols"`   // ALPN 协议
	ClientHelloID   string   `json:"clientHelloId"`   // ClientHello 标识
}

// HTTPFingerprint HTTP 层指纹详情
type HTTPFingerprint struct {
	Protocol       string              `json:"protocol"`       // HTTP/1.1, HTTP/2, HTTP/3
	Headers        map[string]string   `json:"headers"`        // 实际发送的请求头
	HeaderOrder    []string            `json:"headerOrder"`    // 请求头顺序
	PseudoHeaders  []string            `json:"pseudoHeaders"`  // HTTP/2 伪头顺序
	HTTP2Settings  *HTTP2SettingsTrace `json:"http2Settings"`  // HTTP/2 设置帧
	HTTP3Settings  *HTTP3SettingsTrace `json:"http3Settings"`  // HTTP/3 设置
	UserAgent      string              `json:"userAgent"`      // User-Agent
	Accept         string              `json:"accept"`         // Accept
	AcceptLanguage string              `json:"acceptLanguage"` // Accept-Language
	AcceptEncoding string              `json:"acceptEncoding"` // Accept-Encoding
}

// HTTP3SettingsTrace HTTP/3 设置追踪
type HTTP3SettingsTrace struct {
	QUICVersion    uint32 `json:"quicVersion"`
	InitialMaxData uint64 `json:"initialMaxData"`
	MaxStreamsBidi uint64 `json:"maxStreamsBidi"`
	MaxStreamsUni  uint64 `json:"maxStreamsUni"`
	MaxUDPPayload  uint64 `json:"maxUdpPayload"`
	Active         bool   `json:"active"`
}

// HTTP2SettingsTrace HTTP/2 设置追踪
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

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	LocalAddr     string        `json:"localAddr"`
	RemoteAddr    string        `json:"remoteAddr"`
	ConnectTime   time.Duration `json:"connectTime"`
	HandshakeTime time.Duration `json:"handshakeTime"`
	TotalTime     time.Duration `json:"totalTime"`
	Protocol      string        `json:"protocol"`     // h2, http/1.1, h3
	ProtocolUsed  string        `json:"protocolUsed"` // 实际使用的协议
	ALPN          string        `json:"alpn"`         // 协商的 ALPN
	TLSVersion    string        `json:"tlsVersion"`
	CipherSuite   string        `json:"cipherSuite"`
}

// ResponseTrace 响应追踪信息
type ResponseTrace struct {
	StatusCode   int               `json:"statusCode"`
	Status       string            `json:"status"`
	Protocol     string            `json:"protocol"`
	Headers      map[string]string `json:"headers"`
	BodyPreview  string            `json:"bodyPreview"`
	BodyLength   int               `json:"bodyLength"`
	ResponseTime time.Duration     `json:"responseTime"`
}

// ProxyResult 代理请求完整结果
type ProxyResult struct {
	Success       bool                   `json:"success"`
	Error         string                 `json:"error,omitempty"`
	ErrorType     string                 `json:"errorType,omitempty"`    // 错误分类：timeout, network, protocol, etc
	ErrorCode     int                    `json:"errorCode,omitempty"`    // HTTP状态码或错误代码
	ErrorDetails  map[string]interface{} `json:"errorDetails,omitempty"` // 详细错误信息
	RequestTrace  *RequestTrace          `json:"requestTrace"`
	ResponseTrace *ResponseTrace         `json:"responseTrace"`
	ProfileUsed   *ProfileInfo           `json:"profileUsed"`
}

// ProfileInfo 使用的配置信息
type ProfileInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	BrowserType    string `json:"browserType"`
	BrowserVersion string `json:"browserVersion"`
	OS             string `json:"os"`
	OSVersion      string `json:"osVersion"`
}

// TracedClient 带追踪的客户端
type TracedClient struct {
	profile profiles.ClientProfile
	tracer  *RequestTracer
}

// RequestTracer 请求追踪器
type RequestTracer struct {
	Trace *RequestTrace
}

// NewRequestTracer 创建请求追踪器
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

// buildTCPIPFingerprint 构建 TCP/IP 指纹信息
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

// buildTLSFingerprint 构建 TLS 指纹信息
func buildTLSFingerprint(profile profiles.ClientProfile) *TLSFingerprint {
	tls := &TLSFingerprint{
		Version:       formatTLSVersion(profile.TLSVersion),
		CipherSuites:  make([]string, 0, len(profile.CipherSuites)),
		Extensions:    make([]string, 0, len(profile.Extensions)),
		JA3:           calculateJA3(profile),
		JA3Hash:       calculateJA3Hash(profile),
		ClientHelloID: string(profile.BrowserType),
	}

	// 格式化加密套件
	for _, cs := range profile.CipherSuites {
		tls.CipherSuites = append(tls.CipherSuites, formatCipherSuite(cs))
	}

	// 格式化扩展
	for _, ext := range profile.Extensions {
		tls.Extensions = append(tls.Extensions, formatExtension(ext))
	}

	// 格式化支持的曲线
	for _, curve := range profile.SupportedCurves {
		tls.SupportedGroups = append(tls.SupportedGroups, formatCurve(uint16(curve)))
	}

	// ALPN 协议
	tls.ALPNProtocols = []string{"h2", "http/1.1"}

	return tls
}

// buildHTTPFingerprint 构建 HTTP 指纹信息
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

	// HTTP/2 伪头顺序
	http.PseudoHeaders = profile.PseudoHeaderOrder
	if len(http.PseudoHeaders) == 0 {
		http.PseudoHeaders = []string{":method", ":authority", ":scheme", ":path"}
	}

	// HTTP/2 设置
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

// 辅助函数

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
	// JA3 指纹格式: SSLVersion,Cipher,SSLExtension,EllipticCurve,EllipticCurvePointFormat
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
	parts[4] = "0" // 默认未压缩

	return strings.Join(parts, ",")
}

func calculateJA3Hash(profile profiles.ClientProfile) string {
	ja3 := calculateJA3(profile)
	hash := md5.Sum([]byte(ja3))
	return fmt.Sprintf("%x", hash[:])
}

func getHeaderOrder(profile profiles.ClientProfile) []string {
	// 标准的 Chrome 请求头顺序
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
