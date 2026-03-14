package profiles

import (
	"strings"

	"github.com/vistone/fingerprint/modules/core"
)

var (
	// Chrome 134-144
	Chrome134 = ClientProfile{
		ID: "chrome_134", Name: "Chrome 134",
		BrowserType: core.BrowserChrome, BrowserVersion: "134.0.6998.35",
		OS: core.OSLinuxUbuntu, OSVersion: "24.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSWindows10),
	}

	Chrome135 = ClientProfile{
		ID: "chrome_135", Name: "Chrome 135",
		BrowserType: core.BrowserChrome, BrowserVersion: "135.0.7049.95",
		OS: core.OSWindows11, OSVersion: "10.0.26120",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSWindows10),
	}

	Chrome136 = ClientProfile{
		ID: "chrome_136", Name: "Chrome 136",
		BrowserType: core.BrowserChrome, BrowserVersion: "136.0.7103.92",
		OS: core.OSMacOS15, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSWindows10),
	}

	Chrome137 = ClientProfile{
		ID: "chrome_137", Name: "Chrome 137",
		BrowserType: core.BrowserChrome, BrowserVersion: "137.0.7151.68",
		OS: core.OSLinuxFedora, OSVersion: "41",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSWindows10),
	}

	Chrome138 = ClientProfile{
		ID: "chrome_138", Name: "Chrome 138",
		BrowserType: core.BrowserChrome, BrowserVersion: "138.0.7204.157",
		OS: core.OSWindows11, OSVersion: "10.0.26200",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSWindows10),
	}

	Chrome139 = ClientProfile{
		ID: "chrome_139", Name: "Chrome 139",
		BrowserType: core.BrowserChrome, BrowserVersion: "139.0.7258.1",
		OS: core.OSMacOS15, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSWindows10),
	}

	Chrome140 = ClientProfile{
		ID: "chrome_140", Name: "Chrome 140",
		BrowserType: core.BrowserChrome, BrowserVersion: "140.0.0.0",
		OS: core.OSLinuxUbuntu, OSVersion: "24.10",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSWindows10),
	}

	Chrome141 = ClientProfile{
		ID: "chrome_141", Name: "Chrome 141",
		BrowserType: core.BrowserChrome, BrowserVersion: "141.0.0.0",
		OS: core.OSWindows11, OSVersion: "10.0.26200",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSWindows10),
	}

	Chrome142 = ClientProfile{
		ID: "chrome_142", Name: "Chrome 142",
		BrowserType: core.BrowserChrome, BrowserVersion: "142.0.0.0",
		OS: core.OSMacOS15, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSWindows10),
	}

	Chrome143 = ClientProfile{
		ID: "chrome_143", Name: "Chrome 143",
		BrowserType: core.BrowserChrome, BrowserVersion: "143.0.0.0",
		OS: core.OSLinuxUbuntu, OSVersion: "25.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSWindows10),
	}

	Chrome144 = ClientProfile{
		ID: "chrome_144", Name: "Chrome 144",
		BrowserType: core.BrowserChrome, BrowserVersion: "144.0.0.0",
		OS: core.OSWindows11, OSVersion: "10.0.26200",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSWindows10),
	}
)

func init() {
	// registers all Chrome fingerprints
	profiles := []ClientProfile{
		Chrome115, Chrome116, Chrome117, Chrome118, Chrome119,
		Chrome121, Chrome122, Chrome123, Chrome125, Chrome126, Chrome127, Chrome128, Chrome129,
		Chrome134, Chrome135, Chrome136, Chrome137, Chrome138, Chrome139, Chrome140,
		Chrome141, Chrome142, Chrome143, Chrome144,
	}

	// for each profile fills in missing HTTP/2 and HTTP/3 profile
	for i := range profiles {
		p := &profiles[i]

		// padding HTTP/2 profile (if missing)
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

		// padding ConnectionFlow (if missing)
		if p.ConnectionFlow == 0 {
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
			h.AcceptLanguage = "en-US,en;q=0.9"
		}
		if h.AcceptEncoding == "" {
			h.AcceptEncoding = "gzip, deflate, br"
		}
		if h.UserAgent == "" {
			h.UserAgent = buildChromeUserAgent(p.BrowserVersion, p.OS)
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
			h.SecCHUA = `"Chromium";v="` + safeSliceVersion(p.BrowserVersion) + `", "Google Chrome";v="` + safeSliceVersion(p.BrowserVersion) + `"`
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

// buildChromeUserAgent build Chrome User-Agent
func buildChromeUserAgent(version string, os core.OperatingSystem) string {
	osStr := string(os)
	switch {
	case strings.Contains(osStr, "Windows"):
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + version + " Safari/537.36"
	case strings.Contains(osStr, "Mac OS") || strings.Contains(osStr, "Macintosh"):
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + version + " Safari/537.36"
	case strings.Contains(osStr, "Linux"):
		return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + version + " Safari/537.36"
	default:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + version + " Safari/537.36"
	}
}

// AllChromeProfiles returns all Chrome fingerprints
func AllChromeProfiles() []ClientProfile {
	return []ClientProfile{
		Chrome115, Chrome116, Chrome117, Chrome118, Chrome119,
		Chrome121, Chrome122, Chrome123, Chrome125, Chrome126, Chrome127, Chrome128, Chrome129,
		Chrome134, Chrome135, Chrome136, Chrome137, Chrome138, Chrome139, Chrome140,
	}
}
