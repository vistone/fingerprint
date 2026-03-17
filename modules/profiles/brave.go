// Package profiles - Brave browser fingerprint
// contains Brave 1.6x-1.7x version full fingerprint profiles
package profiles

import (
	"github.com/vistone/fingerprint/modules/core"
)

// Brave browser fingerprint (1.6x-1.7x versions)
// braveTCPIP returns Brave TCP/IP fingerprint
func braveTCPIP(osType core.OperatingSystem) *TCPIPFingerprint {
	return CreateTCPIP(osType)
}

var (
	Brave1_60 = ClientProfile{
		ID: "brave_1_60", Name: "Brave 1.60",
		BrowserType: core.BrowserBrave, BrowserVersion: "1.60.0",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc013, 0xc014, 0x002f, 0x0035,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x0015}, {Type: 0x0033}, {Type: 0x002d},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 1000,
			InitialWindowSize: 6291456, MaxFrameSize: 16384,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: braveTCPIP(core.OSWindows10),
	}

	Brave1_62 = ClientProfile{
		ID: "brave_1_62", Name: "Brave 1.62",
		BrowserType: core.BrowserBrave, BrowserVersion: "1.62.0",
		OS: core.OSMacOS14, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: braveTCPIP(core.OSWindows10),
	}

	Brave1_64 = ClientProfile{
		ID: "brave_1_64", Name: "Brave 1.64",
		BrowserType: core.BrowserBrave, BrowserVersion: "1.64.0",
		OS: core.OSWindows11, OSVersion: "10.0.26100",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: braveTCPIP(core.OSWindows10),
	}

	Brave1_66 = ClientProfile{
		ID: "brave_1_66", Name: "Brave 1.66",
		BrowserType: core.BrowserBrave, BrowserVersion: "1.66.0",
		OS: core.OSLinuxUbuntu, OSVersion: "24.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: braveTCPIP(core.OSWindows10),
	}

	Brave1_68 = ClientProfile{
		ID: "brave_1_68", Name: "Brave 1.68",
		BrowserType: core.BrowserBrave, BrowserVersion: "1.68.0",
		OS: core.OSMacOS15, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: braveTCPIP(core.OSWindows10),
	}

	Brave1_70 = ClientProfile{
		ID: "brave_1_70", Name: "Brave 1.70",
		BrowserType: core.BrowserBrave, BrowserVersion: "1.70.0",
		OS: core.OSWindows11, OSVersion: "10.0.26100",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: braveTCPIP(core.OSWindows10),
	}

	Brave1_72 = ClientProfile{
		ID: "brave_1_72", Name: "Brave 1.72",
		BrowserType: core.BrowserBrave, BrowserVersion: "1.72.0",
		OS: core.OSLinuxUbuntu, OSVersion: "24.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: braveTCPIP(core.OSWindows10),
	}
)

func init() {
	profiles := []ClientProfile{
		Brave1_60, Brave1_62, Brave1_64, Brave1_66, Brave1_68, Brave1_70, Brave1_72,
	}

	for i := range profiles {
		p := &profiles[i]
		ensureChromiumCommonDefaults(p)
		applyBraveHeaderDefaults(p)

		Register(*p)
	}
}

func applyBraveHeaderDefaults(p *ClientProfile) {
	h := p.Headers
	if h.UserAgent == "" {
		h.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + p.BrowserVersion + " Safari/537.36"
	}
	if h.SecCHUA == "" {
		h.SecCHUA = `"Brave";v="` + safeSliceVersion(p.BrowserVersion) + `", "Chromium";v="` + safeSliceVersion(p.BrowserVersion) + `"`
	}
	if h.SecCHUAMobile == "" {
		h.SecCHUAMobile = "?0"
	}
	if h.SecCHUAPlatform == "" {
		h.SecCHUAPlatform = `"Windows"`
	}
}

// AllBraveProfiles returns all Brave fingerprints
func AllBraveProfiles() []ClientProfile {
	return []ClientProfile{
		Brave1_60, Brave1_62, Brave1_64, Brave1_66, Brave1_68, Brave1_70, Brave1_72,
	}
}
