package agent

import (
	"sync"

	"github.com/vistone/fingerprint/modules/core"
)

// KnowledgeBase global fingerprint feature knowledge base
//
// Encodes the "canonical form" of real browser fingerprints — from TLS handshake to HTTP/2 frames,
// from TCP/IP stack to JavaScript API, what feature combinations each browser family should
// exhibit on each OS.
//
// Agent uses this knowledge to:
//  1. Determine if a set of features is "reasonable" (consistency validation)
//  2. Detect contradictory signals (e.g., Chrome TLS fingerprint + Firefox HTTP/2 settings)
//  3. Identify traces of forgery/anti-detection tools
//  4. Provide reliable reference baselines for fingerprint generation
type KnowledgeBase struct {
	// Browser fingerprint blueprints: each BrowserType → version → OS → expected features
	browsers map[core.BrowserType]*BrowserKnowledge

	// TLS feature dictionary
	tls *TLSKnowledge

	// TCP/IP feature dictionary
	tcpip *TCPIPKnowledge

	// HTTP/2 feature dictionary
	http2 *HTTP2Knowledge

	// Global statistics
	stats *GlobalStats

	mu sync.RWMutex
}

// BrowserKnowledge knowledge of a single browser family
type BrowserKnowledge struct {
	Family   core.BrowserType
	Versions []VersionKnowledge
	// TLS feature signatures common to all versions
	CommonCipherSuites []uint16
	CommonExtensions   []uint16
	CommonCurves       []core.CurveID
	// Market share estimate of this browser [0,1]
	MarketShare float64
}

// VersionKnowledge knowledge of a single version
type VersionKnowledge struct {
	Version      string
	VersionMajor int
	// List of supported OS
	SupportedOS []core.OperatingSystem
	// TLS fingerprint
	TLSVersion   uint16
	CipherSuites []uint16
	Extensions   []uint16
	Curves       []core.CurveID
	// HTTP fingerprint
	SecCHUAPattern string // Sec-CH-UA pattern
	AcceptPattern  string // Accept header pattern
	// HTTP/2 fingerprint
	H2InitialWindowSize    uint32
	H2MaxConcurrentStreams uint32
	H2HeaderTableSize      uint32
	PseudoHeaderOrder      []string
	ConnectionFlow         uint32
	// Release time range (for determining version plausibility)
	ReleasedYear int
	Deprecated   bool // whether deprecated
}

// TLSKnowledge global TLS knowledge
type TLSKnowledge struct {
	// Known valid TLS 1.3 cipher suites
	ValidTLS13Suites []uint16
	// Known valid TLS 1.2 cipher suites (grouped by browser)
	ValidTLS12Suites map[core.BrowserType][]uint16
	// Known TLS extension types
	KnownExtensions map[uint16]string
	// Typical extension count range for each browser
	ExtensionCountRange map[core.BrowserType][2]int // [min, max]
	// Typical cipher suite count range for each browser
	CipherCountRange map[core.BrowserType][2]int
	// Known GREASE values (Google's TLS extension randomization)
	GREASEValues []uint16
}

// TCPIPKnowledge TCP/IP stack fingerprint knowledge
type TCPIPKnowledge struct {
	// OS → expected TCP/IP parameters
	OSFingerprints map[string]*TCPIPExpected
}

// TCPIPExpected expected TCP/IP features
type TCPIPExpected struct {
	TTL         uint8
	WindowSize  uint16
	WindowScale uint8
	MSS         uint16
	DF          bool
	SACKPerm    bool
	Timestamps  bool
}

// HTTP2Knowledge HTTP/2 protocol fingerprint knowledge
type HTTP2Knowledge struct {
	// Browser → expected HTTP/2 settings
	BrowserSettings map[core.BrowserType]*H2Expected
}

// H2Expected expected HTTP/2 parameters
type H2Expected struct {
	InitialWindowSize    uint32
	MaxConcurrentStreams uint32
	HeaderTableSize      uint32
	MaxFrameSize         uint32
	MaxHeaderListSize    uint32
	ConnectionFlow       uint32
	PseudoHeaderOrder    []string
}

// GlobalStats global statistics data
type GlobalStats struct {
	TotalKnownBrowsers  int
	TotalKnownVersions  int
	TotalKnownProfiles  int
	BrowserMarketShares map[core.BrowserType]float64
	OSMarketShares      map[string]float64
}

// MatchResult knowledge match result
type MatchResult struct {
	// Whether a matching known fingerprint was found
	Known bool `json:"known"`
	// Closest browser family
	ClosestFamily core.BrowserType `json:"closest_family"`
	// Closest version
	ClosestVersion string `json:"closest_version"`
	// Match score [0,1]
	MatchScore float64 `json:"match_score"`
	// Detected contradictory signals
	Contradictions []Contradiction `json:"contradictions,omitempty"`
	// Suspicion score [0,1] (higher with more contradictions)
	SuspicionScore float64 `json:"suspicion_score"`
}

// Contradiction contradictory signal
type Contradiction struct {
	Field    string `json:"field"`    // which feature domain
	Expected string `json:"expected"` // known value
	Actual   string `json:"actual"`   // actually observed value
	Severity string `json:"severity"` // low / medium / high
}

// NewKnowledgeBase create and initialize global fingerprint knowledge base
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
// TLS knowledge initialization
// ===================================================================

func newTLSKnowledge() *TLSKnowledge {
	return &TLSKnowledge{
		// TLS 1.3 standard cipher suites (3, common to all modern browsers)
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
				0x1301, 0x1303, 0x1302, // TLS 1.3 (Firefox order is different)
				0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030,
				0xc00a, 0xc009, 0xc013, 0xc014,
				0x0033, 0x0039, 0x002f, 0x0035,
			},
			core.BrowserSafari: {
				0xc02c, 0xc02b, 0xc030, 0xc02f, // Safari prioritizes ECDSA
				0xcca9, 0xcca8,
				0xc00a, 0xc009, 0xc014, 0xc013,
				0x002f, 0x0035, 0x000a,
			},
			core.BrowserEdge: { // Edge = Chromium kernel, basically identical to Chrome
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
// TCP/IP knowledge initialization
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
// HTTP/2 knowledge initialization
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
				MaxHeaderListSize:    0, // Safari doesn't send this setting
				ConnectionFlow:       10485760,
				PseudoHeaderOrder:    []string{":method", ":scheme", ":path", ":authority"},
			},
			core.BrowserEdge: { // Chromium kernel
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
// Browser fingerprint blueprint initialization
// ===================================================================
