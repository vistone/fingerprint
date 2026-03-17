// Package profiles - Opera browser fingerprint
// contains Opera 100-110 version full fingerprint profiles
package profiles

import (
	"github.com/vistone/fingerprint/modules/core"
)

// Opera browser fingerprint (100-110 versions)
// operaTCPIP returns Opera TCP/IP fingerprint
func operaTCPIP(osType core.OperatingSystem) *TCPIPFingerprint {
	return CreateTCPIP(osType)
}

var (
	// Opera One series (Chromium 100+)
	Opera100 = ClientProfile{
		ID: "opera_100", Name: "Opera 100",
		BrowserType: core.BrowserOpera, BrowserVersion: "100.0",
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
		TCPIP: CreateTCPIP(core.OSWindows11),
	}

	Opera102 = ClientProfile{
		ID: "opera_102", Name: "Opera 102",
		BrowserType: core.BrowserOpera, BrowserVersion: "102.0",
		OS: core.OSMacOS14, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows11),
	}

	Opera104 = ClientProfile{
		ID: "opera_104", Name: "Opera 104",
		BrowserType: core.BrowserOpera, BrowserVersion: "104.0",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows11),
	}

	Opera106 = ClientProfile{
		ID: "opera_106", Name: "Opera 106",
		BrowserType: core.BrowserOpera, BrowserVersion: "106.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows11),
	}

	Opera108 = ClientProfile{
		ID: "opera_108", Name: "Opera 108",
		BrowserType: core.BrowserOpera, BrowserVersion: "108.0",
		OS: core.OSWindows11, OSVersion: "10.0.26100",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows11),
	}

	Opera110 = ClientProfile{
		ID: "opera_110", Name: "Opera 110",
		BrowserType: core.BrowserOpera, BrowserVersion: "110.0",
		OS: core.OSMacOS15, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSWindows11),
	}
)

func init() {
	profiles := []ClientProfile{
		Opera100, Opera102, Opera104, Opera106, Opera108, Opera110,
	}

	for i := range profiles {
		p := &profiles[i]
		ensureChromiumCommonDefaults(p)
		applyOperaHeaderDefaults(p)

		Register(*p)
	}
}

func applyOperaHeaderDefaults(p *ClientProfile) {
	h := p.Headers
	if h.UserAgent == "" {
		h.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + p.BrowserVersion + " Safari/537.36 OPR/" + p.BrowserVersion
	}
	if h.SecCHUA == "" {
		h.SecCHUA = `"Opera";v="` + safeSliceVersion(p.BrowserVersion) + `"`
	}
	if h.SecCHUAMobile == "" {
		h.SecCHUAMobile = "?0"
	}
	if h.SecCHUAPlatform == "" {
		h.SecCHUAPlatform = `"Windows"`
	}
}

// AllOperaProfiles returns all Opera fingerprints
func AllOperaProfiles() []ClientProfile {
	return []ClientProfile{
		Opera100, Opera102, Opera104, Opera106, Opera108, Opera110,
	}
}
