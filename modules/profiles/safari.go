// Package profiles - Safari browser fingerprint
// contains Safari 16-18 version full fingerprint profiles
package profiles

import (
	"strings"

	"github.com/vistone/fingerprint/modules/core"
)

// safariTCPIP returns Safari browser's TCP/IP fingerprint (macOS)
func safariTCPIP() *TCPIPFingerprint {
	return &TCPIPFingerprint{
		IPVersion:        4,
		TTL:              64,
		DF:               true,
		WindowSize:       65535,
		MSS:              1460,
		WindowScale:      6,
		SAckPermitted:    true,
		Timestamps:       true,
		NoOperation:      2,
		EndOfOptions:     true,
		OptionsSignature: "M,N,W,N,N,S,T,E",
		JA4T:             "t13d1814h2_8daaf6152771_b0b889a3c9b7",
	}
}

// Safari browser fingerprint (16-18 versions)
var (
	// Safari 16.x series - macOS Ventura
	Safari16_0 = ClientProfile{
		ID: "safari_16_0", Name: "Safari 16.0",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.0",
		OS: core.OSMacOS13, OSVersion: "13.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
			0xc024, 0xc028, 0xc02b, 0xc02f,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x0015}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveP256, core.CurveP384, core.CurveP521},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 100,
			InitialWindowSize: 2097152, MaxFrameSize: 16384,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}

	Safari16_1 = ClientProfile{
		ID: "safari_16_1", Name: "Safari 16.1",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.1",
		OS: core.OSMacOS13, OSVersion: "13.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}

	Safari16_2 = ClientProfile{
		ID: "safari_16_2", Name: "Safari 16.2",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.2",
		OS: core.OSMacOS13, OSVersion: "13.1",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}

	Safari16_3 = ClientProfile{
		ID: "safari_16_3", Name: "Safari 16.3",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.3",
		OS: core.OSMacOS13, OSVersion: "13.2",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}

	Safari16_4 = ClientProfile{
		ID: "safari_16_4", Name: "Safari 16.4",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.4",
		OS: core.OSMacOS13, OSVersion: "13.3",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}

	Safari16_5 = ClientProfile{
		ID: "safari_16_5", Name: "Safari 16.5",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.5",
		OS: core.OSMacOS13, OSVersion: "13.4",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}

	Safari16_6 = ClientProfile{
		ID: "safari_16_6", Name: "Safari 16.6",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.6",
		OS: core.OSMacOS13, OSVersion: "13.5",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}

	// Safari 17.x series - macOS Sonoma
	Safari17_0 = ClientProfile{
		ID: "safari_17_0", Name: "Safari 17.0",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.0",
		OS: core.OSMacOS14, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}

	Safari17_1 = ClientProfile{
		ID: "safari_17_1", Name: "Safari 17.1",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.1",
		OS: core.OSMacOS14, OSVersion: "14.1",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}

	Safari17_2 = ClientProfile{
		ID: "safari_17_2", Name: "Safari 17.2",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.2",
		OS: core.OSMacOS14, OSVersion: "14.2",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}

	Safari17_3 = ClientProfile{
		ID: "safari_17_3", Name: "Safari 17.3",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.3",
		OS: core.OSMacOS14, OSVersion: "14.3",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}

	Safari17_4 = ClientProfile{
		ID: "safari_17_4", Name: "Safari 17.4",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.4",
		OS: core.OSMacOS14, OSVersion: "14.4",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}

	Safari17_5 = ClientProfile{
		ID: "safari_17_5", Name: "Safari 17.5",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.5",
		OS: core.OSMacOS14, OSVersion: "14.5",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}

	Safari17_6 = ClientProfile{
		ID: "safari_17_6", Name: "Safari 17.6",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.6",
		OS: core.OSMacOS14, OSVersion: "14.6",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}

	// Safari 18.x series - macOS Sequoia
	Safari18_0 = ClientProfile{
		ID: "safari_18_0", Name: "Safari 18.0",
		BrowserType: core.BrowserSafari, BrowserVersion: "18.0",
		OS: core.OSMacOS15, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}

	Safari18_1 = ClientProfile{
		ID: "safari_18_1", Name: "Safari 18.1",
		BrowserType: core.BrowserSafari, BrowserVersion: "18.1",
		OS: core.OSMacOS15, OSVersion: "15.1",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}

	Safari18_2 = ClientProfile{
		ID: "safari_18_2", Name: "Safari 18.2",
		BrowserType: core.BrowserSafari, BrowserVersion: "18.2",
		OS: core.OSMacOS15, OSVersion: "15.2",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: safariTCPIP(),
	}
)

func init() {
	// registers all Safari fingerprints
	profiles := []ClientProfile{
		Safari16_0, Safari16_1, Safari16_2, Safari16_3, Safari16_4, Safari16_5, Safari16_6,
		Safari17_0, Safari17_1, Safari17_2, Safari17_3, Safari17_4, Safari17_5, Safari17_6,
		Safari18_0, Safari18_1, Safari18_2,
	}

	// for each profile fills in missing HTTP/2 and HTTP/3 profile
	for i := range profiles {
		p := &profiles[i]

		// padding HTTP/2 profile (if missing)
		if p.HTTP2Settings.HeaderTableSize == 0 && p.HTTP2Settings.InitialWindowSize == 0 {
			p.HTTP2Settings = core.HTTP2Settings{
				HeaderTableSize:      4096,
				EnablePush:           1,
				MaxConcurrentStreams: 100,
				InitialWindowSize:    2097152,
				MaxFrameSize:         16384,
				MaxHeaderListSize:    262144,
			}
			p.PseudoHeaderOrder = []string{":method", ":scheme", ":path", ":authority"}
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
			h.UserAgent = buildSafariUserAgent(p.BrowserVersion, p.OS)
		}

		Register(*p)
	}
}

// buildSafariUserAgent build Safari User-Agent
func buildSafariUserAgent(version string, os core.OperatingSystem) string {
	osStr := string(os)
	switch {
	case strings.Contains(osStr, "Mac OS") || strings.Contains(osStr, "Macintosh"):
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/" + version + " Safari/605.1.15"
	default:
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/" + version + " Safari/605.1.15"
	}
}

// AllSafariProfiles returns all Safari fingerprints
func AllSafariProfiles() []ClientProfile {
	return []ClientProfile{
		Safari16_0, Safari16_1, Safari16_2, Safari16_3, Safari16_4, Safari16_5, Safari16_6,
		Safari17_0, Safari17_1, Safari17_2, Safari17_3, Safari17_4, Safari17_5, Safari17_6,
		Safari18_0, Safari18_1, Safari18_2,
	}
}
