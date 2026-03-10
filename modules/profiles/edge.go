// Package profiles - Microsoft Edge浏览器指纹
// 包含Edge 115-130版本的完整指纹配置
package profiles

import (
	"strings"
	
	"github.com/vistone/fingerprint/modules/core"
)

// Edge浏览器指纹 (115-130版本)
// edgeTCPIP 返回 Edge 的 TCP/IP 指纹
func edgeTCPIP(osType core.OperatingSystem) *TCPIPFingerprint {
	return CreateTCPIP(osType)
}

var (
	// Edge 115-119
	Edge115 = ClientProfile{
		ID: "edge_115", Name: "Microsoft Edge 115",
		BrowserType: core.BrowserEdge, BrowserVersion: "115.0",
		OS: core.OSWindows10, OSVersion: "10.0.19045",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc013, 0xc014, 0x002f, 0x0035, 0x000a,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x0015}, {Type: 0x0033}, {Type: 0x002d},
			{Type: 0x0029}, {Type: 0x001c},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 1000,
			InitialWindowSize: 6291456, MaxFrameSize: 16384, MaxHeaderListSize: 262144,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
			UpgradeInsecureRequests: "1", SecFetchDest: "document",
			SecFetchMode: "navigate", SecFetchSite: "none",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Edge116 = ClientProfile{
		ID: "edge_116", Name: "Microsoft Edge 116",
		BrowserType: core.BrowserEdge, BrowserVersion: "116.0",
		OS: core.OSWindows11, OSVersion: "10.0.22621",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
			UpgradeInsecureRequests: "1",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Edge117 = ClientProfile{
		ID: "edge_117", Name: "Microsoft Edge 117",
		BrowserType: core.BrowserEdge, BrowserVersion: "117.0",
		OS: core.OSWindows11, OSVersion: "10.0.22621",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Edge118 = ClientProfile{
		ID: "edge_118", Name: "Microsoft Edge 118",
		BrowserType: core.BrowserEdge, BrowserVersion: "118.0",
		OS: core.OSMacOS14, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Edge119 = ClientProfile{
		ID: "edge_119", Name: "Microsoft Edge 119",
		BrowserType: core.BrowserEdge, BrowserVersion: "119.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	// Edge 120-124
	Edge120 = ClientProfile{
		ID: "edge_120", Name: "Microsoft Edge 120",
		BrowserType: core.BrowserEdge, BrowserVersion: "120.0",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Edge121 = ClientProfile{
		ID: "edge_121", Name: "Microsoft Edge 121",
		BrowserType: core.BrowserEdge, BrowserVersion: "121.0",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Edge122 = ClientProfile{
		ID: "edge_122", Name: "Microsoft Edge 122",
		BrowserType: core.BrowserEdge, BrowserVersion: "122.0",
		OS: core.OSMacOS14, OSVersion: "14.3",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Edge123 = ClientProfile{
		ID: "edge_123", Name: "Microsoft Edge 123",
		BrowserType: core.BrowserEdge, BrowserVersion: "123.0",
		OS: core.OSMacOS14, OSVersion: "14.4",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Edge124 = ClientProfile{
		ID: "edge_124", Name: "Microsoft Edge 124",
		BrowserType: core.BrowserEdge, BrowserVersion: "124.0",
		OS: core.OSWindows11, OSVersion: "10.0.26100",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	// Edge 126-130
	Edge126 = ClientProfile{
		ID: "edge_126", Name: "Microsoft Edge 126",
		BrowserType: core.BrowserEdge, BrowserVersion: "126.0",
		OS: core.OSWindows11, OSVersion: "10.0.26100",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Edge128 = ClientProfile{
		ID: "edge_128", Name: "Microsoft Edge 128",
		BrowserType: core.BrowserEdge, BrowserVersion: "128.0",
		OS: core.OSMacOS15, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Edge130 = ClientProfile{
		ID: "edge_130", Name: "Microsoft Edge 130",
		BrowserType: core.BrowserEdge, BrowserVersion: "130.0",
		OS: core.OSLinuxUbuntu, OSVersion: "24.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Edge131 = ClientProfile{
		ID: "edge_131", Name: "Microsoft Edge 131",
		BrowserType: core.BrowserEdge, BrowserVersion: "131.0",
		OS: core.OSWindows11, OSVersion: "10.0.26100",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Edge132 = ClientProfile{
		ID: "edge_132", Name: "Microsoft Edge 132",
		BrowserType: core.BrowserEdge, BrowserVersion: "132.0",
		OS: core.OSMacOS15, OSVersion: "15.2",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Edge133 = ClientProfile{
		ID: "edge_133", Name: "Microsoft Edge 133",
		BrowserType: core.BrowserEdge, BrowserVersion: "133.0",
		OS: core.OSLinuxUbuntu, OSVersion: "24.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Edge134 = ClientProfile{
		ID: "edge_134", Name: "Microsoft Edge 134",
		BrowserType: core.BrowserEdge, BrowserVersion: "134.0",
		OS: core.OSWindows11, OSVersion: "10.0.26200",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}
)

func init() {
	// 注册所有Edge指纹
	profiles := []ClientProfile{
		Edge115, Edge116, Edge117, Edge118, Edge119,
		Edge120, Edge121, Edge122, Edge123, Edge124,
		Edge126, Edge128, Edge130, Edge131, Edge132, Edge133, Edge134,
	}
	
	// 为每个 profile 填充缺失的 HTTP/2 和 HTTP/3 配置
	for i := range profiles {
		p := &profiles[i]
		
		// 填充 HTTP/2 配置（如果缺失）
		if p.HTTP2Settings.HeaderTableSize == 0 && p.HTTP2Settings.InitialWindowSize == 0 {
			p.HTTP2Settings = core.HTTP2Settings{
				HeaderTableSize:      65536,
				EnablePush:           0,
				MaxConcurrentStreams: 1000,
				InitialWindowSize:    6291456,
				MaxFrameSize:         16384,
				MaxHeaderListSize:    262144,
			}
			p.PseudoHeaderOrder = []string{":method", ":authority", ":scheme", ":path"}
		}
		
		// 填充 ConnectionFlow（如果缺失）
		if p.ConnectionFlow == 0 {
			p.ConnectionFlow = 15663105
		}
		
		// 填充 HTTP/3 (QUIC) 配置（如果缺失）
		if p.HTTP3Settings == nil {
			p.HTTP3Settings = &core.HTTP3Settings{
				QUICVersion:            core.QUICVersion1,
				InitialMaxData:         16777216,
				InitialMaxStreamData:   6291456,
				InitialMaxStreamsBidi:  100,
				InitialMaxStreamsUni:   100,
				MaxUDPPayloadSize:      1472,
				AckDelayExponent:       3,
				MaxAckDelay:            25,
				DisableActiveMigration: false,
			}
			p.QUICVersions = []uint32{core.QUICVersion1}
		}
		
		// 填充 Headers（如果缺失）
		if p.Headers == nil {
			p.Headers = &core.HTTPHeaders{}
		}
		h := p.Headers
		if h.Accept == "" {
			h.Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
		}
		if h.AcceptLanguage == "" {
			h.AcceptLanguage = "en-US,en;q=0.9"
		}
		if h.AcceptEncoding == "" {
			h.AcceptEncoding = "gzip, deflate, br"
		}
		if h.UserAgent == "" {
			h.UserAgent = buildEdgeUserAgent(p.BrowserVersion, p.OS)
		}
		if h.SecFetchSite == "" {
			h.SecFetchSite = "none"
		}
		if h.SecFetchMode == "" {
			h.SecFetchMode = "navigate"
		}
		if h.SecFetchDest == "" {
			h.SecFetchDest = "document"
		}
		if h.SecCHUA == "" {
			h.SecCHUA = `"Microsoft Edge";v="` + safeSliceVersion(p.BrowserVersion) + `", "Chromium";v="` + safeSliceVersion(p.BrowserVersion) + `"`
		}
		if h.SecCHUAMobile == "" {
			h.SecCHUAMobile = "?0"
		}
		if h.SecCHUAPlatform == "" {
			h.SecCHUAPlatform = platformString(p.OS)
		}
		if h.UpgradeInsecureRequests == "" {
			h.UpgradeInsecureRequests = "1"
		}
		
		Register(*p)
	}
}

// buildEdgeUserAgent 构建 Edge User-Agent
func buildEdgeUserAgent(version string, os core.OperatingSystem) string {
	osStr := string(os)
	switch {
	case strings.Contains(osStr, "Windows"):
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + version + " Safari/537.36 Edg/" + version
	case strings.Contains(osStr, "Mac OS") || strings.Contains(osStr, "Macintosh"):
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + version + " Safari/537.36 Edg/" + version
	default:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + version + " Safari/537.36 Edg/" + version
	}
}

// AllEdgeProfiles 返回所有Edge指纹
func AllEdgeProfiles() []ClientProfile {
	return []ClientProfile{
		Edge115, Edge116, Edge117, Edge118, Edge119,
		Edge120, Edge121, Edge122, Edge123, Edge124,
		Edge126, Edge128, Edge130,
	}
}
