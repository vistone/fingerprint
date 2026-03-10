// Package profiles - Firefox browser fingerprint
// contains Firefox 115-135 version full fingerprint profiles, including ESR series
package profiles

import (
	"strings"
	
	"github.com/vistone/fingerprint/modules/core"
)

// Firefox browser fingerprint (115-135 versions)
// firefoxTCPIP returns Firefox TCP/IP fingerprint
func firefoxTCPIP(osType core.OperatingSystem) *TCPIPFingerprint {
	return CreateTCPIP(osType)
}

var (
	// Firefox 115-119 ESR series
	Firefox115 = ClientProfile{
		ID: "firefox_115", Name: "Firefox 115 ESR",
		BrowserType: core.BrowserFirefox, BrowserVersion: "115.0",
		OS: core.OSWindows10, OSVersion: "10.0.19045",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc00a, 0xc014, 0xc009, 0xc013,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x0015}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 100,
			InitialWindowSize: 131072, MaxFrameSize: 16384,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
			UpgradeInsecureRequests: "1",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox116 = ClientProfile{
		ID: "firefox_116", Name: "Firefox 116",
		BrowserType: core.BrowserFirefox, BrowserVersion: "116.0",
		OS: core.OSWindows10, OSVersion: "10.0.19045",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox117 = ClientProfile{
		ID: "firefox_117", Name: "Firefox 117",
		BrowserType: core.BrowserFirefox, BrowserVersion: "117.0",
		OS: core.OSMacOS13, OSVersion: "13.5",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox118 = ClientProfile{
		ID: "firefox_118", Name: "Firefox 118",
		BrowserType: core.BrowserFirefox, BrowserVersion: "118.0",
		OS: core.OSMacOS14, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox119 = ClientProfile{
		ID: "firefox_119", Name: "Firefox 119",
		BrowserType: core.BrowserFirefox, BrowserVersion: "119.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	// Firefox 121-124
	Firefox121 = ClientProfile{
		ID: "firefox_121", Name: "Firefox 121",
		BrowserType: core.BrowserFirefox, BrowserVersion: "121.0",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox122 = ClientProfile{
		ID: "firefox_122", Name: "Firefox 122",
		BrowserType: core.BrowserFirefox, BrowserVersion: "122.0",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox123 = ClientProfile{
		ID: "firefox_123", Name: "Firefox 123",
		BrowserType: core.BrowserFirefox, BrowserVersion: "123.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox124 = ClientProfile{
		ID: "firefox_124", Name: "Firefox 124",
		BrowserType: core.BrowserFirefox, BrowserVersion: "124.0",
		OS: core.OSLinuxUbuntu, OSVersion: "24.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	// Firefox 126-129
	Firefox126 = ClientProfile{
		ID: "firefox_126", Name: "Firefox 126",
		BrowserType: core.BrowserFirefox, BrowserVersion: "126.0",
		OS: core.OSMacOS14, OSVersion: "14.5",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox127 = ClientProfile{
		ID: "firefox_127", Name: "Firefox 127",
		BrowserType: core.BrowserFirefox, BrowserVersion: "127.0",
		OS: core.OSMacOS14, OSVersion: "14.6",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox128 = ClientProfile{
		ID: "firefox_128", Name: "Firefox 128 ESR",
		BrowserType: core.BrowserFirefox, BrowserVersion: "128.0",
		OS: core.OSWindows10, OSVersion: "10.0.19045",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox129 = ClientProfile{
		ID: "firefox_129", Name: "Firefox 129",
		BrowserType: core.BrowserFirefox, BrowserVersion: "129.0",
		OS: core.OSLinuxUbuntu, OSVersion: "24.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	// Firefox 131, 134-135
	Firefox131 = ClientProfile{
		ID: "firefox_131", Name: "Firefox 131",
		BrowserType: core.BrowserFirefox, BrowserVersion: "131.0",
		OS: core.OSMacOS15, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox134 = ClientProfile{
		ID: "firefox_134", Name: "Firefox 134",
		BrowserType: core.BrowserFirefox, BrowserVersion: "134.0",
		OS: core.OSMacOS15, OSVersion: "15.1",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox135 = ClientProfile{
		ID: "firefox_135", Name: "Firefox 135",
		BrowserType: core.BrowserFirefox, BrowserVersion: "135.0",
		OS: core.OSWindows11, OSVersion: "10.0.26100",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox136 = ClientProfile{
		ID: "firefox_136", Name: "Firefox 136",
		BrowserType: core.BrowserFirefox, BrowserVersion: "136.0",
		OS: core.OSMacOS15, OSVersion: "15.2",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox137 = ClientProfile{
		ID: "firefox_137", Name: "Firefox 137",
		BrowserType: core.BrowserFirefox, BrowserVersion: "137.0",
		OS: core.OSLinuxUbuntu, OSVersion: "24.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox138 = ClientProfile{
		ID: "firefox_138", Name: "Firefox 138",
		BrowserType: core.BrowserFirefox, BrowserVersion: "138.0",
		OS: core.OSWindows11, OSVersion: "10.0.26100",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox139 = ClientProfile{
		ID: "firefox_139", Name: "Firefox 139",
		BrowserType: core.BrowserFirefox, BrowserVersion: "139.0",
		OS: core.OSMacOS15, OSVersion: "15.3",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Firefox140 = ClientProfile{
		ID: "firefox_140", Name: "Firefox 140",
		BrowserType: core.BrowserFirefox, BrowserVersion: "140.0",
		OS: core.OSLinuxUbuntu, OSVersion: "25.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}
)

func init() {
	// registers all Firefox fingerprints
	profiles := []ClientProfile{
		Firefox115, Firefox116, Firefox117, Firefox118, Firefox119,
		Firefox121, Firefox122, Firefox123, Firefox124,
		Firefox126, Firefox127, Firefox128, Firefox129,
		Firefox131, Firefox134, Firefox135, Firefox136, Firefox137, Firefox138, Firefox139, Firefox140,
	}
	
	// for each profile fills in missing HTTP/2 and HTTP/3 profile
	for i := range profiles {
		p := &profiles[i]
		
		// padding HTTP/2 profile (if missing)
		if p.HTTP2Settings.HeaderTableSize == 0 && p.HTTP2Settings.InitialWindowSize == 0 {
			p.HTTP2Settings = core.HTTP2Settings{
				HeaderTableSize:      65536,
				EnablePush:           0,
				MaxConcurrentStreams: 100,
				InitialWindowSize:    131072,
				MaxFrameSize:         16384,
				MaxHeaderListSize:    262144,
			}
			p.PseudoHeaderOrder = []string{":method", ":path", ":authority", ":scheme"}
			p.ConnectionFlow = 15663105
		}
		
		// padding HTTP/3 (QUIC) profile (if missing)
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
		
		// padding Headers (if missing)
		if p.Headers == nil {
			p.Headers = &core.HTTPHeaders{}
		}
		h := p.Headers
		if h.Accept == "" {
			h.Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
		}
		if h.AcceptLanguage == "" {
			h.AcceptLanguage = "en-US,en;q=0.5"
		}
		if h.AcceptEncoding == "" {
			h.AcceptEncoding = "gzip, deflate, br"
		}
		if h.UserAgent == "" {
			h.UserAgent = buildFirefoxUserAgent(p.BrowserVersion, p.OS)
		}
		if h.UpgradeInsecureRequests == "" {
			h.UpgradeInsecureRequests = "1"
		}
		
		Register(*p)
	}
}

// buildFirefoxUserAgent build Firefox User-Agent
func buildFirefoxUserAgent(version string, os core.OperatingSystem) string {
	osStr := string(os)
	switch {
	case strings.Contains(osStr, "Windows"):
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:" + version + ") Gecko/20100101 Firefox/" + version
	case strings.Contains(osStr, "Mac OS") || strings.Contains(osStr, "Macintosh"):
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:" + version + ") Gecko/20100101 Firefox/" + version
	case strings.Contains(osStr, "Linux"):
		return "Mozilla/5.0 (X11; Linux x86_64; rv:" + version + ") Gecko/20100101 Firefox/" + version
	default:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:" + version + ") Gecko/20100101 Firefox/" + version
	}
}

// AllFirefoxProfiles returns all Firefox fingerprints
func AllFirefoxProfiles() []ClientProfile {
	return []ClientProfile{
		Firefox115, Firefox116, Firefox117, Firefox118, Firefox119,
		Firefox121, Firefox122, Firefox123, Firefox124,
		Firefox126, Firefox127, Firefox128, Firefox129,
		Firefox131, Firefox134, Firefox135,
	}
}
