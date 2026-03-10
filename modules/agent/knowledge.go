package agent

import (
	"sort"
	"sync"

	"github.com/vistone/fingerprint/modules/core"
)

// KnowledgeBase 全球指纹特征知识库
//
// 编码了真实浏览器指纹的"应有形态"——从 TLS 握手到 HTTP/2 帧、
// 从 TCP/IP 栈到 JavaScript API，每个浏览器家族在每种 OS 上
// 应该呈现什么样的特征组合。
//
// Agent 利用这些知识来：
//  1. 判断一组特征是否"合理"（一致性校验）
//  2. 检测矛盾信号（如 Chrome 的 TLS 指纹 + Firefox 的 HTTP/2 设置）
//  3. 识别伪造/反检测工具的蛛丝马迹
//  4. 为指纹生成提供可靠的参考基准
type KnowledgeBase struct {
	// 浏览器指纹蓝图：每个 BrowserType → 版本 → OS → 期望特征
	browsers map[core.BrowserType]*BrowserKnowledge

	// TLS 特征字典
	tls *TLSKnowledge

	// TCP/IP 特征字典
	tcpip *TCPIPKnowledge

	// HTTP/2 特征字典
	http2 *HTTP2Knowledge

	// 全局统计
	stats *GlobalStats

	mu sync.RWMutex
}

// BrowserKnowledge 单个浏览器家族的知识
type BrowserKnowledge struct {
	Family   core.BrowserType
	Versions []VersionKnowledge
	// 全版本通用的 TLS 特征签名
	CommonCipherSuites []uint16
	CommonExtensions   []uint16
	CommonCurves       []core.CurveID
	// 该浏览器的市场份额估计 [0,1]
	MarketShare float64
}

// VersionKnowledge 单个版本的知识
type VersionKnowledge struct {
	Version      string
	VersionMajor int
	// 支持的 OS 列表
	SupportedOS []core.OperatingSystem
	// TLS 指纹
	TLSVersion   uint16
	CipherSuites []uint16
	Extensions   []uint16
	Curves       []core.CurveID
	// HTTP 指纹
	SecCHUAPattern string // Sec-CH-UA 的模式
	AcceptPattern  string // Accept 头模式
	// HTTP/2 指纹
	H2InitialWindowSize    uint32
	H2MaxConcurrentStreams uint32
	H2HeaderTableSize      uint32
	PseudoHeaderOrder      []string
	ConnectionFlow         uint32
	// 发布时间范围（用于判断版本合理性）
	ReleasedYear int
	Deprecated   bool // 是否已过时
}

// TLSKnowledge TLS 全局知识
type TLSKnowledge struct {
	// 已知的合法 TLS 1.3 密码套件
	ValidTLS13Suites []uint16
	// 已知的合法 TLS 1.2 密码套件（按浏览器分组）
	ValidTLS12Suites map[core.BrowserType][]uint16
	// 已知的 TLS 扩展类型
	KnownExtensions map[uint16]string
	// 各浏览器的典型扩展数量范围
	ExtensionCountRange map[core.BrowserType][2]int // [min, max]
	// 各浏览器的典型密码套件数量范围
	CipherCountRange map[core.BrowserType][2]int
	// 已知的 GREASE 值（Google 的 TLS 扩展随机化）
	GREASEValues []uint16
}

// TCPIPKnowledge TCP/IP 栈指纹知识
type TCPIPKnowledge struct {
	// OS → 期望的 TCP/IP 参数
	OSFingerprints map[string]*TCPIPExpected
}

// TCPIPExpected 期望的 TCP/IP 特征
type TCPIPExpected struct {
	TTL         uint8
	WindowSize  uint16
	WindowScale uint8
	MSS         uint16
	DF          bool
	SACKPerm    bool
	Timestamps  bool
}

// HTTP2Knowledge HTTP/2 协议指纹知识
type HTTP2Knowledge struct {
	// 浏览器 → 期望的 HTTP/2 设置
	BrowserSettings map[core.BrowserType]*H2Expected
}

// H2Expected 期望的 HTTP/2 参数
type H2Expected struct {
	InitialWindowSize    uint32
	MaxConcurrentStreams uint32
	HeaderTableSize      uint32
	MaxFrameSize         uint32
	MaxHeaderListSize    uint32
	ConnectionFlow       uint32
	PseudoHeaderOrder    []string
}

// GlobalStats 全局统计数据
type GlobalStats struct {
	TotalKnownBrowsers  int
	TotalKnownVersions  int
	TotalKnownProfiles  int
	BrowserMarketShares map[core.BrowserType]float64
	OSMarketShares      map[string]float64
}

// MatchResult 知识匹配结果
type MatchResult struct {
	// 是否找到匹配的已知指纹
	Known bool `json:"known"`
	// 最接近的浏览器家族
	ClosestFamily core.BrowserType `json:"closest_family"`
	// 最接近的版本
	ClosestVersion string `json:"closest_version"`
	// 匹配得分 [0,1]
	MatchScore float64 `json:"match_score"`
	// 检测到的矛盾信号
	Contradictions []Contradiction `json:"contradictions,omitempty"`
	// 可疑度 [0,1]（矛盾越多越高）
	SuspicionScore float64 `json:"suspicion_score"`
}

// Contradiction 矛盾信号
type Contradiction struct {
	Field    string `json:"field"`    // 哪个特征域
	Expected string `json:"expected"` // 已知值
	Actual   string `json:"actual"`   // 实际观测值
	Severity string `json:"severity"` // low / medium / high
}

// NewKnowledgeBase 创建并初始化全球指纹知识库
func NewKnowledgeBase() *KnowledgeBase {
	kb := &KnowledgeBase{
		browsers: make(map[core.BrowserType]*BrowserKnowledge),
		tls:      newTLSKnowledge(),
		tcpip:    newTCPIPKnowledge(),
		http2:    newHTTP2Knowledge(),
		stats:    &GlobalStats{},
	}
	kb.loadBuiltinKnowledge()
	kb.computeStats()
	return kb
}

// ===================================================================
// TLS 知识初始化
// ===================================================================

func newTLSKnowledge() *TLSKnowledge {
	return &TLSKnowledge{
		// TLS 1.3 标准密码套件（3 个，所有现代浏览器通用）
		ValidTLS13Suites: []uint16{
			0x1301, // TLS_AES_128_GCM_SHA256
			0x1302, // TLS_AES_256_GCM_SHA384
			0x1303, // TLS_CHACHA20_POLY1305_SHA256
		},
		ValidTLS12Suites: map[core.BrowserType][]uint16{
			core.BrowserChrome: {
				0xc02b, 0xc02f, 0xc02c, 0xc030, // ECDHE-RSA/ECDSA with AES
				0xcca9, 0xcca8, // ECDHE with ChaCha20
				0xc013, 0xc014, // ECDHE-RSA with AES-CBC (legacy)
				0x002f, 0x0035, 0x000a, // RSA (fallback)
			},
			core.BrowserFirefox: {
				0x1301, 0x1303, 0x1302, // TLS 1.3（Firefox 顺序不同）
				0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030,
				0xc00a, 0xc009, 0xc013, 0xc014,
				0x0033, 0x0039, 0x002f, 0x0035,
			},
			core.BrowserSafari: {
				0xc02c, 0xc02b, 0xc030, 0xc02f, // Safari 优先 ECDSA
				0xcca9, 0xcca8,
				0xc00a, 0xc009, 0xc014, 0xc013,
				0x002f, 0x0035, 0x000a,
			},
			core.BrowserEdge: { // Edge = Chromium 内核，与 Chrome 基本一致
				0xc02b, 0xc02f, 0xc02c, 0xc030,
				0xcca9, 0xcca8,
				0xc013, 0xc014,
				0x002f, 0x0035, 0x000a,
			},
		},
		KnownExtensions: map[uint16]string{
			0x0000: "server_name",
			0x0001: "max_fragment_length",
			0x0005: "status_request",
			0x000a: "supported_groups",
			0x000b: "ec_point_formats",
			0x000d: "signature_algorithms",
			0x0010: "ALPN",
			0x0012: "signed_certificate_timestamp",
			0x0016: "encrypt_then_mac",
			0x0017: "extended_master_secret",
			0x001c: "record_size_limit",
			0x0023: "session_ticket",
			0x002b: "supported_versions",
			0x002d: "psk_key_exchange_modes",
			0x0033: "key_share",
			0x0039: "post_handshake_auth",
			0x4469: "application_settings",
			0xfe0d: "encrypted_client_hello",
			0xff01: "renegotiation_info",
		},
		ExtensionCountRange: map[core.BrowserType][2]int{
			core.BrowserChrome:  {10, 18},
			core.BrowserFirefox: {9, 16},
			core.BrowserSafari:  {8, 14},
			core.BrowserEdge:    {10, 18},
			core.BrowserOpera:   {10, 18},
			core.BrowserBrave:   {10, 18},
		},
		CipherCountRange: map[core.BrowserType][2]int{
			core.BrowserChrome:  {9, 17},
			core.BrowserFirefox: {11, 17},
			core.BrowserSafari:  {9, 16},
			core.BrowserEdge:    {9, 17},
			core.BrowserOpera:   {9, 17},
			core.BrowserBrave:   {9, 17},
		},
		// GREASE (Generate Random Extensions And Sustain Extensibility)
		GREASEValues: []uint16{
			0x0a0a, 0x1a1a, 0x2a2a, 0x3a3a, 0x4a4a,
			0x5a5a, 0x6a6a, 0x7a7a, 0x8a8a, 0x9a9a,
			0xaaaa, 0xbaba, 0xcaca, 0xdada, 0xeaea, 0xfafa,
		},
	}
}

// ===================================================================
// TCP/IP 知识初始化
// ===================================================================

func newTCPIPKnowledge() *TCPIPKnowledge {
	return &TCPIPKnowledge{
		OSFingerprints: map[string]*TCPIPExpected{
			"windows": {
				TTL: 128, WindowSize: 64240, WindowScale: 8,
				MSS: 1460, DF: true, SACKPerm: true, Timestamps: true,
			},
			"macos": {
				TTL: 64, WindowSize: 65535, WindowScale: 6,
				MSS: 1460, DF: true, SACKPerm: true, Timestamps: true,
			},
			"linux": {
				TTL: 64, WindowSize: 64240, WindowScale: 7,
				MSS: 1460, DF: true, SACKPerm: true, Timestamps: true,
			},
			"ios": {
				TTL: 64, WindowSize: 65535, WindowScale: 6,
				MSS: 1460, DF: true, SACKPerm: true, Timestamps: true,
			},
			"android": {
				TTL: 64, WindowSize: 64240, WindowScale: 7,
				MSS: 1460, DF: true, SACKPerm: true, Timestamps: false,
			},
		},
	}
}

// ===================================================================
// HTTP/2 知识初始化
// ===================================================================

func newHTTP2Knowledge() *HTTP2Knowledge {
	return &HTTP2Knowledge{
		BrowserSettings: map[core.BrowserType]*H2Expected{
			core.BrowserChrome: {
				InitialWindowSize:    6291456,
				MaxConcurrentStreams: 1000,
				HeaderTableSize:      65536,
				MaxFrameSize:         16384,
				MaxHeaderListSize:    262144,
				ConnectionFlow:       15663105,
				PseudoHeaderOrder:    []string{":method", ":authority", ":scheme", ":path"},
			},
			core.BrowserFirefox: {
				InitialWindowSize:    131072,
				MaxConcurrentStreams: 100,
				HeaderTableSize:      65536,
				MaxFrameSize:         16384,
				MaxHeaderListSize:    65536,
				ConnectionFlow:       12517377,
				PseudoHeaderOrder:    []string{":method", ":path", ":authority", ":scheme"},
			},
			core.BrowserSafari: {
				InitialWindowSize:    4194304,
				MaxConcurrentStreams: 100,
				HeaderTableSize:      4096,
				MaxFrameSize:         16384,
				MaxHeaderListSize:    0, // Safari 不发送此设置
				ConnectionFlow:       10485760,
				PseudoHeaderOrder:    []string{":method", ":scheme", ":path", ":authority"},
			},
			core.BrowserEdge: { // Chromium 内核
				InitialWindowSize:    6291456,
				MaxConcurrentStreams: 1000,
				HeaderTableSize:      65536,
				MaxFrameSize:         16384,
				MaxHeaderListSize:    262144,
				ConnectionFlow:       15663105,
				PseudoHeaderOrder:    []string{":method", ":authority", ":scheme", ":path"},
			},
		},
	}
}

// ===================================================================
// 浏览器指纹蓝图初始化
// ===================================================================

func (kb *KnowledgeBase) loadBuiltinKnowledge() {
	// Chrome 知识
	kb.browsers[core.BrowserChrome] = &BrowserKnowledge{
		Family:             core.BrowserChrome,
		CommonCipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
		CommonExtensions:   []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0016, 0x000d, 0x002b, 0x002d, 0x0033},
		CommonCurves:       []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		MarketShare:        0.65,
		Versions: []VersionKnowledge{
			{Version: "115", VersionMajor: 115, ReleasedYear: 2023, SupportedOS: desktopOSList(),
				TLSVersion: 0x0303, CipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8, 0xc013, 0xc014, 0x002f, 0x0035, 0x000a},
				Extensions:          []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0016, 0x000d, 0x002b, 0x002d, 0x0033, 0x001c},
				Curves:              []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536,
				PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"}, ConnectionFlow: 15663105,
				SecCHUAPattern: `"Google Chrome";v="%d"`, AcceptPattern: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp"},
			{Version: "119", VersionMajor: 119, ReleasedYear: 2023, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536,
				ConnectionFlow: 15663105, SecCHUAPattern: `"Google Chrome";v="%d"`, AcceptPattern: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp"},
			{Version: "121", VersionMajor: 121, ReleasedYear: 2024, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536,
				ConnectionFlow: 15663105, Curves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384, core.CurveP256Kyber}},
			{Version: "131", VersionMajor: 131, ReleasedYear: 2024, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536, ConnectionFlow: 15663105},
			{Version: "133", VersionMajor: 133, ReleasedYear: 2025, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536, ConnectionFlow: 15663105},
			{Version: "134", VersionMajor: 134, ReleasedYear: 2025, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536, ConnectionFlow: 15663105},
		},
	}

	// Firefox 知识
	kb.browsers[core.BrowserFirefox] = &BrowserKnowledge{
		Family:             core.BrowserFirefox,
		CommonCipherSuites: []uint16{0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030},
		CommonExtensions:   []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
		CommonCurves:       []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		MarketShare:        0.03,
		Versions: []VersionKnowledge{
			{Version: "115", VersionMajor: 115, ReleasedYear: 2023, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc00a, 0xc009, 0xc013, 0xc014, 0x0033, 0x0039, 0x002f, 0x0035},
				H2InitialWindowSize: 131072, H2MaxConcurrentStreams: 100, H2HeaderTableSize: 65536, ConnectionFlow: 12517377,
				PseudoHeaderOrder: []string{":method", ":path", ":authority", ":scheme"}},
			{Version: "120", VersionMajor: 120, ReleasedYear: 2023, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc00a, 0xc009, 0xc013, 0xc014},
				H2InitialWindowSize: 131072, H2MaxConcurrentStreams: 100, H2HeaderTableSize: 65536, ConnectionFlow: 12517377},
			{Version: "124", VersionMajor: 124, ReleasedYear: 2024, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030},
				H2InitialWindowSize: 131072, H2MaxConcurrentStreams: 100, H2HeaderTableSize: 65536, ConnectionFlow: 12517377},
			{Version: "135", VersionMajor: 135, ReleasedYear: 2025, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030},
				H2InitialWindowSize: 131072, H2MaxConcurrentStreams: 100, H2HeaderTableSize: 65536, ConnectionFlow: 12517377},
		},
	}

	// Safari 知识
	kb.browsers[core.BrowserSafari] = &BrowserKnowledge{
		Family:             core.BrowserSafari,
		CommonCipherSuites: []uint16{0xc02c, 0xc02b, 0xc030, 0xc02f, 0xcca9, 0xcca8, 0xc00a, 0xc009, 0xc014, 0xc013},
		CommonExtensions:   []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
		CommonCurves:       []core.CurveID{core.CurveP256, core.CurveP384, core.CurveP521, core.CurveX25519},
		MarketShare:        0.19,
		Versions: []VersionKnowledge{
			{Version: "16", VersionMajor: 16, ReleasedYear: 2022, SupportedOS: appleOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0xc02c, 0xc02b, 0xc030, 0xc02f, 0xcca9, 0xcca8, 0xc00a, 0xc009, 0xc014, 0xc013, 0x002f, 0x0035, 0x000a},
				H2InitialWindowSize: 4194304, H2MaxConcurrentStreams: 100, H2HeaderTableSize: 4096, ConnectionFlow: 10485760,
				PseudoHeaderOrder: []string{":method", ":scheme", ":path", ":authority"}},
			{Version: "17", VersionMajor: 17, ReleasedYear: 2023, SupportedOS: appleOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0xc02c, 0xc02b, 0xc030, 0xc02f, 0xcca9, 0xcca8, 0xc00a, 0xc009, 0xc014, 0xc013},
				H2InitialWindowSize: 4194304, H2MaxConcurrentStreams: 100, H2HeaderTableSize: 4096, ConnectionFlow: 10485760},
			{Version: "18", VersionMajor: 18, ReleasedYear: 2024, SupportedOS: appleOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0xc02c, 0xc02b, 0xc030, 0xc02f, 0xcca9, 0xcca8, 0xc00a, 0xc009},
				H2InitialWindowSize: 4194304, H2MaxConcurrentStreams: 100, H2HeaderTableSize: 4096, ConnectionFlow: 10485760},
		},
	}

	// Edge 知识 (Chromium 内核)
	kb.browsers[core.BrowserEdge] = &BrowserKnowledge{
		Family:             core.BrowserEdge,
		CommonCipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
		CommonExtensions:   []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0016, 0x000d, 0x002b, 0x002d, 0x0033},
		CommonCurves:       []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		MarketShare:        0.05,
		Versions: []VersionKnowledge{
			{Version: "118", VersionMajor: 118, ReleasedYear: 2023, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536, ConnectionFlow: 15663105},
			{Version: "131", VersionMajor: 131, ReleasedYear: 2024, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536, ConnectionFlow: 15663105},
		},
	}

	// Opera 知识 (Chromium 内核)
	kb.browsers[core.BrowserOpera] = &BrowserKnowledge{
		Family:             core.BrowserOpera,
		CommonCipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
		CommonExtensions:   []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0016, 0x000d, 0x002b, 0x002d, 0x0033},
		CommonCurves:       []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		MarketShare:        0.03,
		Versions: []VersionKnowledge{
			{Version: "104", VersionMajor: 104, ReleasedYear: 2023, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536, ConnectionFlow: 15663105},
		},
	}

	// Brave 知识 (Chromium 内核，但有随机化)
	kb.browsers[core.BrowserBrave] = &BrowserKnowledge{
		Family:             core.BrowserBrave,
		CommonCipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
		CommonExtensions:   []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0016, 0x000d, 0x002b, 0x002d, 0x0033},
		CommonCurves:       []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		MarketShare:        0.02,
		Versions: []VersionKnowledge{
			{Version: "1.60", VersionMajor: 1, ReleasedYear: 2024, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536, ConnectionFlow: 15663105},
		},
	}
}

func desktopOSList() []core.OperatingSystem {
	return []core.OperatingSystem{core.OSWindows10, core.OSMacOS13, core.OSMacOS14, core.OSMacOS15, core.OSLinux}
}

func appleOSList() []core.OperatingSystem {
	return []core.OperatingSystem{core.OSMacOS13, core.OSMacOS14, core.OSMacOS15, core.OSiOS, core.OSiPadOS}
}

func (kb *KnowledgeBase) computeStats() {
	kb.stats.BrowserMarketShares = make(map[core.BrowserType]float64)
	kb.stats.OSMarketShares = map[string]float64{
		"windows": 0.72, "macos": 0.16, "linux": 0.04,
		"ios": 0.04, "android": 0.03, "other": 0.01,
	}
	totalVersions := 0
	for bt, bk := range kb.browsers {
		kb.stats.BrowserMarketShares[bt] = bk.MarketShare
		totalVersions += len(bk.Versions)
	}
	kb.stats.TotalKnownBrowsers = len(kb.browsers)
	kb.stats.TotalKnownVersions = totalVersions
	kb.stats.TotalKnownProfiles = totalVersions * 3 // 估算×OS 组合数
}

// ===================================================================
// 查询接口
// ===================================================================

// GetBrowserKnowledge 获取指定浏览器家族的知识
func (kb *KnowledgeBase) GetBrowserKnowledge(family core.BrowserType) *BrowserKnowledge {
	kb.mu.RLock()
	defer kb.mu.RUnlock()
	return kb.browsers[family]
}

// IsKnownCipherSuite 检查密码套件是否已知
func (kb *KnowledgeBase) IsKnownCipherSuite(suite uint16) bool {
	for _, s := range kb.tls.ValidTLS13Suites {
		if s == suite {
			return true
		}
	}
	for _, suites := range kb.tls.ValidTLS12Suites {
		for _, s := range suites {
			if s == suite {
				return true
			}
		}
	}
	return false
}

// IsGREASE 检查是否为 GREASE 值
func (kb *KnowledgeBase) IsGREASE(value uint16) bool {
	for _, g := range kb.tls.GREASEValues {
		if value == g {
			return true
		}
	}
	return false
}

// GetExpectedTCPIP 获取指定 OS 的期望 TCP/IP 参数
func (kb *KnowledgeBase) GetExpectedTCPIP(osFamily string) *TCPIPExpected {
	return kb.tcpip.OSFingerprints[osFamily]
}

// GetExpectedH2 获取指定浏览器的期望 HTTP/2 参数
func (kb *KnowledgeBase) GetExpectedH2(family core.BrowserType) *H2Expected {
	return kb.http2.BrowserSettings[family]
}

// Stats 返回全局统计
func (kb *KnowledgeBase) Stats() *GlobalStats {
	return kb.stats
}

// FindClosestVersion 在指定浏览器中找到最匹配的版本
func (kb *KnowledgeBase) FindClosestVersion(family core.BrowserType, cipherSuites []uint16) *VersionKnowledge {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	bk, ok := kb.browsers[family]
	if !ok || len(bk.Versions) == 0 {
		return nil
	}

	bestScore := -1.0
	var bestVersion *VersionKnowledge

	for i := range bk.Versions {
		v := &bk.Versions[i]
		if len(v.CipherSuites) == 0 {
			continue
		}
		score := jaccardSimilarity(cipherSuites, v.CipherSuites)
		if score > bestScore {
			bestScore = score
			bestVersion = v
		}
	}

	return bestVersion
}

// jaccardSimilarity 计算两个 uint16 集合的 Jaccard 相似度
func jaccardSimilarity(a, b []uint16) float64 {
	setA := make(map[uint16]struct{}, len(a))
	for _, v := range a {
		setA[v] = struct{}{}
	}
	setB := make(map[uint16]struct{}, len(b))
	for _, v := range b {
		setB[v] = struct{}{}
	}

	intersection := 0
	for v := range setA {
		if _, ok := setB[v]; ok {
			intersection++
		}
	}

	union := len(setA)
	for v := range setB {
		if _, ok := setA[v]; !ok {
			union++
		}
	}

	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// sortedUint16 返回排序后的副本
func sortedUint16(in []uint16) []uint16 {
	out := make([]uint16, len(in))
	copy(out, in)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
