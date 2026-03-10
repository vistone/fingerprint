// translated comment
// translated comment
package profiles

import (
	"strings"

	"github.com/vistone/fingerprint/modules/core"
)

// translated comment
func getChromeTCPIP(osType core.OperatingSystem) *TCPIPFingerprint {
	base := &TCPIPFingerprint{
		IPVersion:        4,
		DF:               true,
		SYN:              true,
		ACK:              false,
		MSS:              1460,
		SAckPermitted:    true,
		Timestamps:       true,
		EndOfOptions:     true,
		OptionsSignature: "M,N,W,N,N,S,T,E",
	}

	// translated comment
	osStr := string(osType)
	
	if strings.Contains(osStr, "Windows") {
		// translated comment
		base.TTL = 128
		base.WindowSize = 64240
		base.WindowScale = 8
		base.NoOperation = 2
		base.JA4T = "t13d1715h2_8daaf6152771_9e7c7c2f41aa"
	} else if strings.Contains(osStr, "Macintosh") || strings.Contains(osStr, "Mac OS") {
		// translated comment
		base.TTL = 64
		base.WindowSize = 65535
		base.WindowScale = 6
		base.NoOperation = 2
		base.JA4T = "t13d1814h2_8daaf6152771_b0b889a3c9b7"
	} else if strings.Contains(osStr, "Linux") || strings.Contains(osStr, "X11") {
		// translated comment
		base.TTL = 64
		base.WindowSize = 64240
		base.WindowScale = 7
		base.NoOperation = 2
		base.JA4T = "t13d1714h2_8daaf6152771_02713a6ec338"
	} else {
		// translated comment
		base.TTL = 128
		base.WindowSize = 64240
		base.WindowScale = 8
		base.NoOperation = 2
		base.JA4T = "t13d1715h2_8daaf6152771_9e7c7c2f41aa"
	}

	return base
}

// translated comment
var (
	// translated comment
	Chrome115 = ClientProfile{
		ID: "chrome_115", Name: "Chrome 115",
		BrowserType: core.BrowserChrome, BrowserVersion: "115.0.5790.170",
		OS: core.OSWindows10, OSVersion: "10.0.19045",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030,
			0xcca9, 0xcca8, 0xc013, 0xc014, 0x002f, 0x0035, 0x000a,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x002d}, {Type: 0x0033}, {Type: 0x001c},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 1000,
			InitialWindowSize: 6291456, MaxFrameSize: 16384, MaxHeaderListSize: 262144,
		},
		PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
		ConnectionFlow: 15663105,
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
			SecFetchDest: "document", SecFetchMode: "navigate", SecFetchSite: "none",
			SecCHUA: `"Google Chrome";v="115", "Not=A?Brand";v="24", "Chromium";v="115"`,
			SecCHUAMobile: "?0", SecCHUAPlatform: `"Windows"`,
			UpgradeInsecureRequests: "1",
		},
		TCPIP: getChromeTCPIP(core.OSWindows10),
	}

	Chrome116 = ClientProfile{
		ID: "chrome_116", Name: "Chrome 116",
		BrowserType: core.BrowserChrome, BrowserVersion: "116.0.5845.188",
		OS: core.OSWindows10, OSVersion: "10.0.19045",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
			SecCHUA: `"Chromium";v="116", "Not)A;Brand";v="24", "Google Chrome";v="116"`,
		},
		TCPIP: getChromeTCPIP(core.OSWindows10),
	}

	Chrome117 = ClientProfile{
		ID: "chrome_117", Name: "Chrome 117",
		BrowserType: core.BrowserChrome, BrowserVersion: "117.0.5938.149",
		OS: core.OSWindows11, OSVersion: "10.0.22621",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
			SecCHUA: `"Google Chrome";v="117", "Not;A=Brand";v="8", "Chromium";v="117"`,
		},
		TCPIP: getChromeTCPIP(core.OSWindows11),
	}

	Chrome118 = ClientProfile{
		ID: "chrome_118", Name: "Chrome 118",
		BrowserType: core.BrowserChrome, BrowserVersion: "118.0.5993.117",
		OS: core.OSMacOS14, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
			SecCHUA: `"Chromium";v="118", "Google Chrome";v="118", "Not=A?Brand";v="99"`,
		},
		TCPIP: getChromeTCPIP(core.OSMacOS14),
	}

	Chrome119 = ClientProfile{
		ID: "chrome_119", Name: "Chrome 119",
		BrowserType: core.BrowserChrome, BrowserVersion: "119.0.6045.200",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSLinuxUbuntu),
	}

	// Chrome 121-129
	Chrome121 = ClientProfile{
		ID: "chrome_121", Name: "Chrome 121",
		BrowserType: core.BrowserChrome, BrowserVersion: "121.0.6167.184",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSLinuxUbuntu),
	}

	Chrome122 = ClientProfile{
		ID: "chrome_122", Name: "Chrome 122",
		BrowserType: core.BrowserChrome, BrowserVersion: "122.0.6261.112",
		OS: core.OSMacOS14, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSLinuxUbuntu),
	}

	Chrome123 = ClientProfile{
		ID: "chrome_123", Name: "Chrome 123",
		BrowserType: core.BrowserChrome, BrowserVersion: "123.0.6312.122",
		OS: core.OSLinuxDebian, OSVersion: "12",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSLinuxDebian),
	}

	Chrome125 = ClientProfile{
		ID: "chrome_125", Name: "Chrome 125",
		BrowserType: core.BrowserChrome, BrowserVersion: "125.0.6422.141",
		OS: core.OSWindows11, OSVersion: "10.0.26100",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSLinuxFedora),
	}

	Chrome126 = ClientProfile{
		ID: "chrome_126", Name: "Chrome 126",
		BrowserType: core.BrowserChrome, BrowserVersion: "126.0.6478.126",
		OS: core.OSMacOS14, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSLinuxFedora),
	}

	Chrome127 = ClientProfile{
		ID: "chrome_127", Name: "Chrome 127",
		BrowserType: core.BrowserChrome, BrowserVersion: "127.0.6533.119",
		OS: core.OSWindows11, OSVersion: "10.0.26100",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSLinuxUbuntu),
	}

	Chrome128 = ClientProfile{
		ID: "chrome_128", Name: "Chrome 128",
		BrowserType: core.BrowserChrome, BrowserVersion: "128.0.6613.138",
		OS: core.OSMacOS15, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSLinuxUbuntu),
	}

	Chrome129 = ClientProfile{
		ID: "chrome_129", Name: "Chrome 129",
		BrowserType: core.BrowserChrome, BrowserVersion: "129.0.6668.101",
		OS: core.OSLinuxFedora, OSVersion: "40",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSLinuxUbuntu),
	}

	// Chrome 134-140
	Chrome134 = ClientProfile{
		ID: "chrome_134", Name: "Chrome 134",
		BrowserType: core.BrowserChrome, BrowserVersion: "134.0.6998.35",
		OS: core.OSLinuxUbuntu, OSVersion: "24.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
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
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: getChromeTCPIP(core.OSWindows10),
	}
)

func init() {
	// translated comment
	profiles := []ClientProfile{
		Chrome115, Chrome116, Chrome117, Chrome118, Chrome119,
		Chrome121, Chrome122, Chrome123, Chrome125, Chrome126, Chrome127, Chrome128, Chrome129,
		Chrome134, Chrome135, Chrome136, Chrome137, Chrome138, Chrome139, Chrome140,
		Chrome141, Chrome142, Chrome143, Chrome144,
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

// translated comment
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

// translated comment
func AllChromeProfiles() []ClientProfile {
	return []ClientProfile{
		Chrome115, Chrome116, Chrome117, Chrome118, Chrome119,
		Chrome121, Chrome122, Chrome123, Chrome125, Chrome126, Chrome127, Chrome128, Chrome129,
		Chrome134, Chrome135, Chrome136, Chrome137, Chrome138, Chrome139, Chrome140,
	}
}
