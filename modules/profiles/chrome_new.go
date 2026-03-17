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
	profiles := []ClientProfile{
		Chrome115, Chrome116, Chrome117, Chrome118, Chrome119,
		Chrome121, Chrome122, Chrome123, Chrome125, Chrome126, Chrome127, Chrome128, Chrome129,
		Chrome134, Chrome135, Chrome136, Chrome137, Chrome138, Chrome139, Chrome140,
		Chrome141, Chrome142, Chrome143, Chrome144,
	}

	for i := range profiles {
		p := &profiles[i]
		ensureChromiumCommonDefaults(p)
		applyChromeHeaderDefaults(p)

		Register(*p)
	}
}

func applyChromeHeaderDefaults(p *ClientProfile) {
	h := p.Headers
	if h.UserAgent == "" {
		h.UserAgent = buildChromeUserAgent(p.BrowserVersion, p.OS)
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
