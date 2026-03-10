// Package profiles 提供内置浏览器指纹配置
package profiles

import (
	"github.com/vistone/fingerprint/modules/core"
)

// 内置 Chrome 指纹配置
// builtinTCPIP 返回 TCP/IP 指纹
func builtinTCPIP(osType core.OperatingSystem) *TCPIPFingerprint {
	return CreateTCPIP(osType)
}

var (
	// Chrome 133
	Chrome133 = ClientProfile{
		ID:             "chrome_133",
		Name:           "Chrome 133",
		BrowserType:    core.BrowserChrome,
		BrowserVersion: "133.0.6943.98",
		OS:             core.OSWindows10,
		OSVersion:      "10.0.19045",
		OSArch:         "x86_64",
		OSBitness:      "64",
		TLSVersion:     0x0303, // TLS 1.2/1.3
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, // TLS 1.3 cipher suites
			0xc02b, 0xc02f, 0xc02c, 0xc030,
			0xcca9, 0xcca8,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, // server_name
			{Type: 0x0017}, // extended_master_secret
			{Type: 0xff01}, // renegotiation_info
			{Type: 0x000a}, // supported_groups
			{Type: 0x000b}, // ec_point_formats
			{Type: 0x0023}, // session_ticket
			{Type: 0x0016}, // ALPN
			{Type: 0x000d}, // signature_algorithms
			{Type: 0x002b}, // supported_versions
			{Type: 0x002d}, // psk_key_exchange_modes
			{Type: 0x0033}, // key_share
		},
		SupportedCurves: []core.CurveID{
			core.CurveX25519,
			core.CurveP256,
			core.CurveP384,
		},
		SupportedPoints: []uint8{0}, // uncompressed
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize:      65536,
			EnablePush:           0,
			MaxConcurrentStreams: 1000,
			InitialWindowSize:    6291456,
			MaxFrameSize:         16384,
			MaxHeaderListSize:    262144,
		},
		HTTP2Priorities: []core.HTTP2Priority{
			{StreamID: 1, Weight: 255, DependsOn: 0, Exclusive: false},
		},
		PseudoHeaderOrder: []string{
			":method", ":authority", ":scheme", ":path",
		},
		ConnectionFlow: 15663105,
		Headers: &core.HTTPHeaders{
			Accept:                  "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage:          "en-US,en;q=0.9",
			AcceptEncoding:          "gzip, deflate, br",
			SecFetchSite:            "none",
			SecFetchMode:            "navigate",
			SecFetchUser:            "?1",
			SecFetchDest:            "document",
			SecCHUA:                 `"Not(A:Brand";v="99", "Google Chrome";v="133", "Chromium";v="133"`,
			SecCHUAMobile:           "?0",
			SecCHUAPlatform:         `"Windows"`,
			UpgradeInsecureRequests: "1",
		},
		TCPIP: builtinTCPIP(core.OSWindows10),
	}

	// Chrome 131
	Chrome131 = ClientProfile{
		ID:             "chrome_131",
		Name:           "Chrome 131",
		BrowserType:    core.BrowserChrome,
		BrowserVersion: "131.0.6778.86",
		OS:             core.OSWindows10,
		OSVersion:      "10.0.19045",
		OSArch:         "x86_64",
		OSBitness:      "64",
		TLSVersion:     0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303,
			0xc02b, 0xc02f, 0xc02c, 0xc030,
			0xcca9, 0xcca8,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x002d}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{
			core.CurveX25519,
			core.CurveP256,
		},
		SupportedPoints: []uint8{0},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize:      65536,
			EnablePush:           0,
			MaxConcurrentStreams: 1000,
			InitialWindowSize:    6291456,
			MaxFrameSize:         16384,
			MaxHeaderListSize:    262144,
		},
		PseudoHeaderOrder: []string{
			":method", ":authority", ":scheme", ":path",
		},
		ConnectionFlow: 15663105,
		Headers: &core.HTTPHeaders{
			Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage:  "en-US,en;q=0.5",
			AcceptEncoding:  "gzip, deflate, br",
			SecFetchSite:    "none",
			SecFetchMode:    "navigate",
			SecFetchDest:    "document",
			SecCHUA:         `"Google Chrome";v="131", "Chromium";v="131"`,
			SecCHUAMobile:   "?0",
			SecCHUAPlatform: `"Windows"`,
		},
		TCPIP: builtinTCPIP(core.OSWindows10),
	}

	// Firefox 133
	Firefox133 = ClientProfile{
		ID:             "firefox_133",
		Name:           "Firefox 133",
		BrowserType:    core.BrowserFirefox,
		BrowserVersion: "133.0",
		OS:             core.OSWindows10,
		OSVersion:      "10.0.19045",
		OSArch:         "x86_64",
		OSBitness:      "64",
		TLSVersion:     0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303,
			0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc00a, 0xc014, 0xc009, 0xc013,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x0015}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{
			core.CurveX25519,
			core.CurveP256,
			core.CurveP384,
		},
		SupportedPoints: []uint8{0},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize:      65536,
			EnablePush:           0,
			MaxConcurrentStreams: 100,
			InitialWindowSize:    131072,
			MaxFrameSize:         16384,
		},
		PseudoHeaderOrder: []string{
			":method", ":path", ":authority", ":scheme",
		},
		ConnectionFlow: 12517377,
		Headers: &core.HTTPHeaders{
			Accept:                  "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage:          "en-US,en;q=0.5",
			AcceptEncoding:          "gzip, deflate, br",
			UpgradeInsecureRequests: "1",
			SecFetchDest:            "document",
			SecFetchMode:            "navigate",
			SecFetchSite:            "none",
			SecFetchUser:            "?1",
		},
		TCPIP: builtinTCPIP(core.OSWindows10),
	}

	// Safari 18.0
	Safari180 = ClientProfile{
		ID:             "safari_18_0",
		Name:           "Safari 18.0",
		BrowserType:    core.BrowserSafari,
		BrowserVersion: "18.0",
		OS:             core.OSMacOS15,
		OSVersion:      "15.0",
		OSArch:         "x86_64",
		OSBitness:      "64",
		TLSVersion:     0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303,
			0xc02c, 0xc02b, 0xc030, 0xc02f,
			0xcca9, 0xcca8,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{
			core.CurveX25519,
			core.CurveP256,
			core.CurveP384,
			core.CurveP521,
		},
		SupportedPoints: []uint8{0},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize:      4096,
			EnablePush:           1,
			MaxConcurrentStreams: 100,
			InitialWindowSize:    2097152,
			MaxFrameSize:         16384,
		},
		PseudoHeaderOrder: []string{
			":method", ":scheme", ":path", ":authority",
		},
		ConnectionFlow: 10485760,
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9",
			AcceptEncoding: "gzip, deflate, br",
			SecFetchDest:   "document",
			SecFetchMode:   "navigate",
			SecFetchSite:   "none",
		},
		TCPIP: builtinTCPIP(core.OSWindows10),
	}
)

// init 初始化默认指纹注册表
func init() {
	profiles := []*ClientProfile{
		&Chrome133,
		&Chrome131,
		&Firefox133,
		&Safari180,
	}

	// 为每个 profile 填充缺失的 UserAgent
	for _, p := range profiles {
		if p.Headers == nil {
			p.Headers = &core.HTTPHeaders{}
		}

		// 自动构建 UserAgent（如果未设置）
		if p.Headers.UserAgent == "" {
			p.Headers.UserAgent = buildUserAgent(p)
		}

		// 确保 TCP/IP 指纹存在
		if p.TCPIP == nil {
			p.TCPIP = builtinTCPIP(p.OS)
		}

		Register(*p)
	}
}

// buildUserAgent 根据浏览器类型和版本构建 User-Agent 字符串
func buildUserAgent(p *ClientProfile) string {
	osStr := string(p.OS)

	switch p.BrowserType {
	case core.BrowserChrome:
		// Chrome User-Agent 格式
		return "Mozilla/5.0 (" + osStr + ") AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" +
			p.BrowserVersion + " Safari/537.36"

	case core.BrowserFirefox:
		// Firefox User-Agent 格式
		version := p.BrowserVersion
		return "Mozilla/5.0 (" + osStr + "; rv:" + version + ") Gecko/20100101 Firefox/" + version

	case core.BrowserSafari:
		// Safari User-Agent 格式
		return "Mozilla/5.0 (" + osStr + ") AppleWebKit/605.1.15 (KHTML, like Gecko) Version/" +
			p.BrowserVersion + " Safari/605.1.15"

	case core.BrowserEdge:
		// Edge User-Agent 格式
		return "Mozilla/5.0 (" + osStr + ") AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" +
			p.BrowserVersion + " Safari/537.36 Edg/" + p.BrowserVersion

	case core.BrowserBrave:
		// Brave User-Agent 格式（类似Chrome）
		return "Mozilla/5.0 (" + osStr + ") AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" +
			p.BrowserVersion + " Safari/537.36"

	case core.BrowserOpera:
		// Opera User-Agent 格式
		return "Mozilla/5.0 (" + osStr + ") AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" +
			p.BrowserVersion + " Safari/537.36 OPR/" + p.BrowserVersion

	default:
		// 默认 User-Agent
		return "Mozilla/5.0 (" + osStr + ") AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" +
			p.BrowserVersion + " Safari/537.36"
	}
}
