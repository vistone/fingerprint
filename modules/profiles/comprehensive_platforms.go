package profiles

import (
	"strings"

	"github.com/vistone/fingerprint/modules/core"
)

func initSafariProfiles() {
	safariVersions := []struct {
		id      string
		version string
		os      core.OperatingSystem
		osVer   string
	}{
		{"safari_14_0", "14.0", core.OSMacOS13, "13.0"},
		{"safari_14_1", "14.1", core.OSMacOS13, "13.1"},
		{"safari_15_0", "15.0", core.OSMacOS13, "13.0"},
		{"safari_15_1", "15.1", core.OSMacOS13, "13.1"},
		{"safari_15_3", "15.3", core.OSMacOS13, "13.2"},
		{"safari_15_4", "15.4", core.OSMacOS13, "13.3"},
		{"safari_15_5", "15.5", core.OSMacOS14, "14.0"},
		{"safari_15_6", "15.6", core.OSMacOS14, "14.0"},
		{"safari_16_1", "16.1", core.OSMacOS14, "14.0"},
		{"safari_16_2", "16.2", core.OSMacOS14, "14.1"},
		{"safari_16_3", "16.3", core.OSMacOS14, "14.2"},
		{"safari_16_4", "16.4", core.OSMacOS14, "14.3"},
		{"safari_16_5", "16.5", core.OSMacOS14, "14.4"},
		{"safari_16_6", "16.6", core.OSMacOS14, "14.5"},
		{"safari_17_1", "17.1", core.OSMacOS15, "15.0"},
		{"safari_17_2", "17.2", core.OSMacOS15, "15.1"},
		{"safari_17_3", "17.3", core.OSMacOS15, "15.2"},
		{"safari_17_4", "17.4", core.OSMacOS15, "15.3"},
		{"safari_17_5", "17.5", core.OSMacOS15, "15.4"},
		{"safari_17_6", "17.6", core.OSMacOS15, "15.5"},
		// 15.0, 16.0, 17.0, 18.0, 18.1 already in other files
	}

	for _, v := range safariVersions {
		p := ClientProfile{
			ID:              v.id,
			Name:            "Safari " + v.version,
			BrowserType:     core.BrowserSafari,
			BrowserVersion:  v.version,
			OS:              v.os,
			OSVersion:       v.osVer,
			TLSVersion:      0x0303,
			CipherSuites:    []uint16{0x1301, 0x1302, 0x1303, 0xc02c, 0xc02b, 0xc030, 0xc02f},
			SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384, core.CurveP521},
			HTTP2Settings: core.HTTP2Settings{
				HeaderTableSize: 4096, EnablePush: 1,
				MaxConcurrentStreams: 100, InitialWindowSize: 2097152,
			},
			// HTTP/3 (QUIC) profile - Safari already supported HTTP/3
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
			PseudoHeaderOrder: []string{":method", ":scheme", ":path", ":authority"},
			Headers: &core.HTTPHeaders{
				Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				AcceptLanguage: "en-US,en;q=0.9",
				AcceptEncoding: "gzip, deflate, br",
			},
			TCPIP: createTCPIP(v.os),
		}
		Register(p)
	}
}

// iOS Safari (15 profiles)
func initiOSProfiles() {
	iosVersions := []struct {
		id      string
		version string
		osVer   string
	}{
		{"safari_ios_14_0", "14.0", "14.0"},
		{"safari_ios_14_5", "14.5", "14.5"},
		{"safari_ios_14_6", "14.6", "14.6"},
		{"safari_ios_14_7", "14.7", "14.7"},
		{"safari_ios_14_8", "14.8", "14.8"},
		{"safari_ios_15_1", "15.1", "15.1"},
		{"safari_ios_15_2", "15.2", "15.2"},
		{"safari_ios_15_3", "15.3", "15.3"},
		{"safari_ios_15_4", "15.4", "15.4"},
		{"safari_ios_16_1", "16.1", "16.1"},
		{"safari_ios_16_2", "16.2", "16.2"},
		{"safari_ios_16_3", "16.3", "16.3"},
		{"safari_ios_16_4", "16.4", "16.4"},
		{"safari_ios_16_5", "16.5", "16.5"},
		{"safari_ios_16_6", "16.6", "16.6"},
		// 15.5, 15.6, 16.0, 17.0, 18.0, 18.1 already in other files
	}

	for _, v := range iosVersions {
		p := ClientProfile{
			ID:             v.id,
			Name:           "Safari iOS " + v.version,
			BrowserType:    core.BrowserSafari,
			BrowserVersion: v.version,
			OS:             core.OSMacOS15,
			OSVersion:      v.osVer,
			TLSVersion:     0x0303, // TLS 1.2
			CipherSuites: []uint16{
				0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030,
				0xcca9, 0xcca8, 0xc013, 0xc014,
			},
			Extensions: []core.TLSExtension{
				{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
				{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
				{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
				{Type: 0x002d}, {Type: 0x0033},
			},
			SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
			HTTP2Settings: core.HTTP2Settings{
				HeaderTableSize:      65536,
				EnablePush:           0,
				MaxConcurrentStreams: 1000,
				InitialWindowSize:    6291456,
				MaxFrameSize:         16384,
				MaxHeaderListSize:    262144,
			},
			// HTTP/3 (QUIC) profile - iOS Safari already supported HTTP/3
			HTTP3Settings: &core.HTTP3Settings{
				QUICVersion:            core.QUICVersion1,
				InitialMaxData:         16777216,
				InitialMaxStreamData:   6291456,
				InitialMaxStreamsBidi:  100,
				InitialMaxStreamsUni:   100,
				MaxUDPPayloadSize:      1472,
				AckDelayExponent:       3,
				MaxAckDelay:            25,
				DisableActiveMigration: false,
			},
			QUICVersions:      []uint32{core.QUICVersion1},
			PseudoHeaderOrder: []string{":method", ":scheme", ":authority", ":path"},
			ConnectionFlow:    15663105,
			Headers: &core.HTTPHeaders{
				Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				AcceptLanguage:  "en-US,en;q=0.9",
				AcceptEncoding:  "gzip, deflate, br",
				SecCHUA:         "",
				SecCHUAMobile:   "?1",
				SecCHUAPlatform: `"iPhone"`,
			},
			TCPIP: createTCPIP(core.OSiOS),
		}
		Register(p)
	}
}

// Android Chrome (15 profiles)
func initAndroidProfiles() {
	for _, v := range androidChromeVersions {
		Register(buildAndroidChromeProfile(v))
	}
}

type androidChromeVersion struct {
	id      string
	version string
	android string
	device  string
}

var androidChromeVersions = []androidChromeVersion{
	{"chrome_android_103", "103.0.5060.71", "13", "SM-G998B"},
	{"chrome_android_105", "105.0.5195.136", "13", "SM-G996B"},
	{"chrome_android_107", "107.0.5304.105", "13", "SM-G991B"},
	{"chrome_android_109", "109.0.5414.117", "13", "SM-S908B"},
	{"chrome_android_111", "111.0.5563.116", "14", "SM-S918B"},
	{"chrome_android_113", "113.0.5672.162", "14", "SM-S928B"},
	{"chrome_android_115", "115.0.5790.166", "14", "Pixel 7"},
	{"chrome_android_117", "117.0.5938.153", "14", "Pixel 7 Pro"},
	{"chrome_android_119", "119.0.6045.193", "14", "Pixel 8"},
	{"chrome_android_121", "121.0.6167.178", "14", "Pixel 8 Pro"},
	{"chrome_android_123", "123.0.6312.80", "14", "SM-G996B"},
	{"chrome_android_125", "125.0.6422.165", "14", "SM-G991B"},
	{"chrome_android_127", "127.0.6533.103", "14", "Pixel 8a"},
	{"chrome_android_129", "129.0.6668.81", "15", "Pixel 9"},
	{"chrome_android_131", "131.0.6778.135", "15", "Pixel 9 Pro"},
}

func buildAndroidChromeProfile(v androidChromeVersion) ClientProfile {
	return ClientProfile{
		ID:             v.id,
		Name:           "Chrome Android " + v.version,
		BrowserType:    core.BrowserChrome,
		BrowserVersion: v.version,
		OS:             core.OSLinux,
		OSVersion:      v.android,
		TLSVersion:     0x0303,
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
			HeaderTableSize:      65536,
			EnablePush:           0,
			MaxConcurrentStreams: 1000,
			InitialWindowSize:    6291456,
			MaxFrameSize:         16384,
			MaxHeaderListSize:    262144,
		},
		HTTP3Settings: &core.HTTP3Settings{
			QUICVersion:            core.QUICVersion1,
			InitialMaxData:         16777216,
			InitialMaxStreamData:   6291456,
			InitialMaxStreamsBidi:  100,
			InitialMaxStreamsUni:   100,
			MaxUDPPayloadSize:      1472,
			AckDelayExponent:       3,
			MaxAckDelay:            25,
			DisableActiveMigration: false,
		},
		QUICVersions:      []uint32{core.QUICVersion1},
		PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
		ConnectionFlow:    15663105,
		Headers: &core.HTTPHeaders{
			Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage:  "en-US,en;q=0.9",
			AcceptEncoding:  "gzip, deflate, br",
			SecCHUA:         `"Android";v="` + v.android + `", "Chrome";v="` + safeSliceVersion(v.version) + `"`,
			SecCHUAMobile:   "?1",
			SecCHUAPlatform: `"Android"`,
		},
		TCPIP: createTCPIP(core.OSAndroid),
	}
}

// Edge more versions (8 profiles)
func initEdgeProfiles() {
	edgeVersions := []struct {
		id      string
		version string
		channel string
	}{
		{"edge_110", "110.0.1587.63", "stable"},
		{"edge_111", "111.0.1661.54", "stable"},
		{"edge_112", "112.0.1722.58", "stable"},
		{"edge_113", "113.0.1774.57", "stable"},
		{"edge_114", "114.0.1823.67", "stable"},
		{"edge_115", "115.0.1901.203", "stable"},
		{"edge_125", "125.0.2535.92", "stable"},
		{"edge_128", "128.0.2739.79", "stable"},
		// 120, 130 already in other files
	}

	for _, v := range edgeVersions {
		p := ClientProfile{
			ID:             v.id,
			Name:           "Edge " + v.version,
			BrowserType:    core.BrowserEdge,
			BrowserVersion: v.version,
			OS:             core.OSWindows11,
			OSVersion:      "10.0.22621",
			TLSVersion:     0x0303, // TLS 1.2
			CipherSuites: []uint16{
				0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030,
				0xcca9, 0xcca8, 0xc013, 0xc014, 0x002f, 0x0035, 0x000a,
			},
			Extensions: []core.TLSExtension{
				{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
				{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
				{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
				{Type: 0x002d}, {Type: 0x0033},
			},
			SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
			HTTP2Settings: core.HTTP2Settings{
				HeaderTableSize:      65536,
				EnablePush:           0,
				MaxConcurrentStreams: 1000,
				InitialWindowSize:    6291456,
				MaxFrameSize:         16384,
				MaxHeaderListSize:    262144,
			},
			// HTTP/3 (QUIC) profile - Edge already supported HTTP/3
			HTTP3Settings: &core.HTTP3Settings{
				QUICVersion:            core.QUICVersion1,
				InitialMaxData:         16777216,
				InitialMaxStreamData:   6291456,
				InitialMaxStreamsBidi:  100,
				InitialMaxStreamsUni:   100,
				MaxUDPPayloadSize:      1472,
				AckDelayExponent:       3,
				MaxAckDelay:            25,
				DisableActiveMigration: false,
			},
			QUICVersions:      []uint32{core.QUICVersion1},
			PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
			ConnectionFlow:    15663105,
			Headers: &core.HTTPHeaders{
				Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				AcceptLanguage:  "en-US,en;q=0.9",
				AcceptEncoding:  "gzip, deflate, br",
				SecCHUA:         `"Microsoft Edge";v="` + safeSliceVersion(v.version) + `", "Chromium";v="` + safeSliceVersion(v.version) + `"`,
				SecCHUAMobile:   "?0",
				SecCHUAPlatform: `"Windows"`,
			},
			TCPIP: createTCPIP(core.OSWindows11),
		}
		Register(p)
	}
}

// Opera more versions (5 profiles)
func initOperaProfiles() {
	operaVersions := []struct {
		id      string
		version string
	}{
		{"opera_95", "95.0.4635.90"},
		{"opera_96", "96.0.4693.80"},
		{"opera_97", "97.0.4719.63"},
		{"opera_98", "98.0.4759.39"},
		{"opera_99", "99.0.4788.77"},
		// 100, 105 already in other files
	}

	for _, v := range operaVersions {
		p := ClientProfile{
			ID:             v.id,
			Name:           "Opera " + v.version,
			BrowserType:    core.BrowserOpera,
			BrowserVersion: v.version,
			OS:             core.OSWindows10,
			OSVersion:      "10.0.19045",
			TLSVersion:     0x0303, // TLS 1.2
			CipherSuites: []uint16{
				0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030,
				0xcca9, 0xcca8, 0xc013, 0xc014,
			},
			Extensions: []core.TLSExtension{
				{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
				{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
				{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
				{Type: 0x002d}, {Type: 0x0033},
			},
			SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
			HTTP2Settings: core.HTTP2Settings{
				HeaderTableSize:      65536,
				EnablePush:           0,
				MaxConcurrentStreams: 1000,
				InitialWindowSize:    6291456,
				MaxFrameSize:         16384,
				MaxHeaderListSize:    262144,
			},
			// HTTP/3 (QUIC) profile - Opera already supported HTTP/3
			HTTP3Settings: &core.HTTP3Settings{
				QUICVersion:            core.QUICVersion1,
				InitialMaxData:         16777216,
				InitialMaxStreamData:   6291456,
				InitialMaxStreamsBidi:  100,
				InitialMaxStreamsUni:   100,
				MaxUDPPayloadSize:      1472,
				AckDelayExponent:       3,
				MaxAckDelay:            25,
				DisableActiveMigration: false,
			},
			QUICVersions:      []uint32{core.QUICVersion1},
			PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
			ConnectionFlow:    15663105,
			Headers: &core.HTTPHeaders{
				Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				AcceptLanguage:  "en-US,en;q=0.9",
				AcceptEncoding:  "gzip, deflate, br",
				SecCHUA:         `"Opera";v="` + safeSliceVersion(v.version) + `"`,
				SecCHUAMobile:   "?0",
				SecCHUAPlatform: `"Windows"`,
			},
			TCPIP: createTCPIP(core.OSWindows10),
		}
		Register(p)
	}
}

// helper function
func platformString(os core.OperatingSystem) string {
	osStr := string(os)
	if contains(osStr, "Windows") {
		return `"Windows"`
	}
	if contains(osStr, "Mac OS X") || contains(osStr, "Macintosh") {
		return `"macOS"`
	}
	if contains(osStr, "Linux") {
		return `"Linux"`
	}
	return `"Windows"`
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// initializes all extended profiles
func init() {
	initChromeProfiles()
	initFirefoxProfiles()
	initSafariProfiles()
	initiOSProfiles()
	initAndroidProfiles()
	initEdgeProfiles()
	initOperaProfiles()
}
