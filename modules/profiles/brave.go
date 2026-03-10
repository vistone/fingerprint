// translated comment
// translated comment
package profiles

import (
	"github.com/vistone/fingerprint/modules/core"
)

// translated comment
// translated comment
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: braveTCPIP(core.OSWindows10),
	}
)

func init() {
	// translated comment
	profiles := []ClientProfile{
		Brave1_60, Brave1_62, Brave1_64, Brave1_66, Brave1_68, Brave1_70, Brave1_72,
	}
	
	// translated comment
	for i := range profiles {
		p := &profiles[i]
		
		// translated comment
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
		
		// translated comment
		if p.ConnectionFlow == 0 {
			p.ConnectionFlow = 15663105
		}
		
		// translated comment
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
		
		// translated comment
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
			h.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + p.BrowserVersion + " Safari/537.36"
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
			h.SecCHUA = `"Brave";v="` + safeSliceVersion(p.BrowserVersion) + `", "Chromium";v="` + safeSliceVersion(p.BrowserVersion) + `"`
		}
		if h.SecCHUAMobile == "" {
			h.SecCHUAMobile = "?0"
		}
		if h.SecCHUAPlatform == "" {
			h.SecCHUAPlatform = `"Windows"`
		}
		if h.UpgradeInsecureRequests == "" {
			h.UpgradeInsecureRequests = "1"
		}
		
		Register(*p)
	}
}

// translated comment
func AllBraveProfiles() []ClientProfile {
	return []ClientProfile{
		Brave1_60, Brave1_62, Brave1_64, Brave1_66, Brave1_68, Brave1_70, Brave1_72,
	}
}
