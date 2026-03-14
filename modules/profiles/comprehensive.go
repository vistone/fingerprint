// Package profiles provides comprehensive browser fingerprint profiles
// target: 90+ real browser fingerprint
package profiles

import (
	"strings"

	"github.com/vistone/fingerprint/modules/core"
)

// createTCPIP creates TCP/IP fingerprint based on operating system
func createTCPIP(osType core.OperatingSystem) *TCPIPFingerprint {
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

	osStr := string(osType)

	if strings.Contains(osStr, "Windows") {
		base.TTL = 128
		base.WindowSize = 64240
		base.WindowScale = 8
		base.NoOperation = 2
		base.JA4T = "t13d1715h2_8daaf6152771_9e7c7c2f41aa"
	} else if strings.Contains(osStr, "Macintosh") || strings.Contains(osStr, "Mac OS") {
		base.TTL = 64
		base.WindowSize = 65535
		base.WindowScale = 6
		base.NoOperation = 2
		base.JA4T = "t13d1814h2_8daaf6152771_b0b889a3c9b7"
	} else if strings.Contains(osStr, "Linux") || strings.Contains(osStr, "X11") {
		base.TTL = 64
		base.WindowSize = 64240
		base.WindowScale = 7
		base.NoOperation = 2
		base.JA4T = "t13d1714h2_8daaf6152771_02713a6ec338"
	} else if strings.Contains(osStr, "iPhone") || strings.Contains(osStr, "iPad") {
		// iOS characteristics
		base.TTL = 64
		base.WindowSize = 65535
		base.WindowScale = 6
		base.NoOperation = 2
		base.JA4T = "t13d1814h2_8daaf6152771_b0b889a3c9b7"
	} else if strings.Contains(osStr, "Android") {
		// Android characteristics
		base.TTL = 64
		base.WindowSize = 65535
		base.WindowScale = 6
		base.NoOperation = 2
		base.JA4T = "t13d1814h2_8daaf6152771_b0b889a3c9b7"
	} else {
		base.TTL = 128
		base.WindowSize = 64240
		base.WindowScale = 8
		base.NoOperation = 2
		base.JA4T = "t13d1715h2_8daaf6152771_9e7c7c2f41aa"
	}

	return base
}

// createChromeTCPIP create Chrome browser's TCP/IP fingerprint
func createChromeTCPIP(osType core.OperatingSystem) *TCPIPFingerprint {
	return CreateTCPIP(osType)
}

// Chrome comprehensive profile (30 profiles)
func initChromeProfiles() {
	chromeVersions := []struct {
		id      string
		version string
		os      core.OperatingSystem
		osVer   string
	}{
		{"chrome_103", "103.0.5060.134", core.OSWindows10, "10.0.19044"},
		{"chrome_104", "104.0.5112.102", core.OSWindows10, "10.0.19044"},
		{"chrome_105", "105.0.5195.127", core.OSWindows10, "10.0.19044"},
		{"chrome_106", "106.0.5249.119", core.OSWindows10, "10.0.19044"},
		{"chrome_107", "107.0.5304.107", core.OSWindows11, "10.0.22000"},
		{"chrome_108", "108.0.5359.125", core.OSWindows11, "10.0.22000"},
		{"chrome_109", "109.0.5414.120", core.OSWindows11, "10.0.22000"},
		{"chrome_110", "110.0.5481.178", core.OSWindows11, "10.0.22621"},
		{"chrome_111", "111.0.5563.147", core.OSWindows11, "10.0.22621"},
		{"chrome_112", "112.0.5615.138", core.OSWindows11, "10.0.22621"},
		{"chrome_113", "113.0.5672.127", core.OSWindows11, "10.0.22621"},
		{"chrome_114", "114.0.5735.199", core.OSWindows11, "10.0.22621"},
		{"chrome_115", "115.0.5790.171", core.OSWindows11, "10.0.22621"},
		{"chrome_116", "116.0.5845.188", core.OSWindows11, "10.0.22621"},
		{"chrome_117", "117.0.5938.149", core.OSWindows11, "10.0.22631"},
		{"chrome_118", "118.0.5993.117", core.OSWindows11, "10.0.22631"},
		{"chrome_119", "119.0.6045.200", core.OSWindows11, "10.0.22631"},
		{"chrome_121", "121.0.6167.185", core.OSMacOS14, "14.0"},
		{"chrome_122", "122.0.6261.128", core.OSMacOS14, "14.0"},
		{"chrome_125", "125.0.6422.142", core.OSMacOS15, "15.0"},
		{"chrome_126", "126.0.6478.126", core.OSMacOS15, "15.0"},
		{"chrome_127", "127.0.6533.120", core.OSLinuxUbuntu, "22.04"},
		{"chrome_128", "128.0.6613.138", core.OSLinuxUbuntu, "22.04"},
		{"chrome_129", "129.0.6668.101", core.OSLinuxUbuntu, "22.04"},
		// versions already present in other files
		// 120, 124, 130, 131, 132, 133
	}

	for _, v := range chromeVersions {
		p := ClientProfile{
			ID:             v.id,
			Name:           "Chrome " + getMajorVersion(v.version),
			BrowserType:    core.BrowserChrome,
			BrowserVersion: v.version,
			OS:             v.os,
			OSVersion:      v.osVer,
			OSArch:         "x86_64",
			OSBitness:      "64",
			TLSVersion:     0x0303,
			CipherSuites:   []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9},
			Extensions: []core.TLSExtension{
				{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
				{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
				{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
				{Type: 0x002d}, {Type: 0x0033},
			},
			SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256},
			HTTP2Settings: core.HTTP2Settings{
				HeaderTableSize: 65536, EnablePush: 0,
				MaxConcurrentStreams: 1000, InitialWindowSize: 6291456,
			},
			// HTTP/3 (QUIC) profile - Chrome already supported HTTP/3
			HTTP3Settings: &core.HTTP3Settings{
				QUICVersion:            core.QUICVersion1,
				InitialMaxData:         16777216, // 16MB
				InitialMaxStreamData:   6291456,  // 6MB
				InitialMaxStreamsBidi:  100,
				InitialMaxStreamsUni:   100,
				MaxUDPPayloadSize:      1472,
				AckDelayExponent:       3,
				MaxAckDelay:            25,
				DisableActiveMigration: false,
			},
			QUICVersions: []uint32{core.QUICVersion1},
			Headers: &core.HTTPHeaders{
				Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				AcceptLanguage:  "en-US,en;q=0.9",
				AcceptEncoding:  "gzip, deflate, br",
				SecFetchSite:    "none",
				SecFetchMode:    "navigate",
				SecFetchDest:    "document",
				SecCHUA:         `"Chromium";v="` + safeSliceVersion(v.version) + `", "Google Chrome";v="` + safeSliceVersion(v.version) + `"`,
				SecCHUAMobile:   "?0",
				SecCHUAPlatform: platformString(v.os),
			},
			TCPIP: createChromeTCPIP(v.os),
		}
		Register(p)
	}
}

// Firefox comprehensive profile (25 profiles)
func initFirefoxProfiles() {
	firefoxVersions := []struct {
		id      string
		version string
		os      core.OperatingSystem
	}{
		{"firefox_102", "102.0", core.OSWindows10},
		{"firefox_103", "103.0", core.OSWindows10},
		{"firefox_104", "104.0", core.OSWindows10},
		{"firefox_105", "105.0", core.OSWindows10},
		{"firefox_106", "106.0", core.OSWindows11},
		{"firefox_107", "107.0", core.OSWindows11},
		{"firefox_108", "108.0", core.OSWindows11},
		{"firefox_109", "109.0", core.OSWindows11},
		{"firefox_110", "110.0", core.OSWindows11},
		{"firefox_111", "111.0", core.OSWindows11},
		{"firefox_112", "112.0", core.OSWindows11},
		{"firefox_113", "113.0", core.OSWindows11},
		{"firefox_114", "114.0", core.OSMacOS13},
		{"firefox_115", "115.0", core.OSMacOS13},
		{"firefox_116", "116.0", core.OSMacOS13},
		{"firefox_117", "117.0", core.OSMacOS14},
		{"firefox_118", "118.0", core.OSMacOS14},
		{"firefox_119", "119.0", core.OSMacOS14},
		{"firefox_121", "121.0", core.OSMacOS15},
		{"firefox_122", "122.0", core.OSLinux},
		{"firefox_123", "123.0", core.OSLinux},
		{"firefox_124", "124.0", core.OSLinux},
		{"firefox_126", "126.0", core.OSLinux},
		{"firefox_127", "127.0", core.OSLinux},
		{"firefox_128", "128.0", core.OSLinux},
		// 120, 125, 130, 132, 133 already in other files
	}

	for _, v := range firefoxVersions {
		p := ClientProfile{
			ID:             v.id,
			Name:           "Firefox " + getMajorVersion(v.version),
			BrowserType:    core.BrowserFirefox,
			BrowserVersion: v.version,
			OS:             v.os,
			OSVersion:      "10.0",
			TLSVersion:     0x0303,
			CipherSuites: []uint16{
				0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f,
				0xcca9, 0xcca8, 0xc00a, 0xc014,
			},
			HTTP2Settings: core.HTTP2Settings{
				HeaderTableSize: 65536, EnablePush: 0,
				MaxConcurrentStreams: 100, InitialWindowSize: 131072,
			},
			// HTTP/3 (QUIC) profile - Firefox already supported HTTP/3
			HTTP3Settings: &core.HTTP3Settings{
				QUICVersion:            core.QUICVersion1,
				InitialMaxData:         16777216, // 16MB
				InitialMaxStreamData:   6291456,  // 6MB
				InitialMaxStreamsBidi:  100,
				InitialMaxStreamsUni:   100,
				MaxUDPPayloadSize:      1472,
				AckDelayExponent:       3,
				MaxAckDelay:            25,
				DisableActiveMigration: false,
			},
			QUICVersions:      []uint32{core.QUICVersion1},
			PseudoHeaderOrder: []string{":method", ":path", ":authority", ":scheme"},
			Headers: &core.HTTPHeaders{
				Accept:                  "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				AcceptLanguage:          "en-US,en;q=0.5",
				AcceptEncoding:          "gzip, deflate, br",
				UpgradeInsecureRequests: "1",
			},
			TCPIP: createTCPIP(v.os),
		}
		Register(p)
	}
}

// Safari comprehensive profile (20 profiles)
