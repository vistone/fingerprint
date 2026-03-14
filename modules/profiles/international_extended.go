package profiles

import (
	"github.com/vistone/fingerprint/modules/core"
)

// =========================================================================
// Chinese OEM mobile browsers (Xiaomi, Huawei, OPPO)
// =========================================================================

var (
	MiBrowser18 = ClientProfile{
		ID: "mi_browser_18", Name: "Xiaomi Mi Browser 18.5",
		BrowserType: core.BrowserChrome, BrowserVersion: "120.0.0.0",
		OS: core.OSAndroid, OSVersion: "14",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc02c, 0xc030, 0xc013, 0xc014,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x000d}, {Type: 0x002b}, {Type: 0x002d},
			{Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 100,
			InitialWindowSize: 4194304, MaxFrameSize: 16384,
		},
		PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
		ConnectionFlow:    15663105,
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "zh-CN,zh;q=0.9,en;q=0.8",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Linux; Android 14; 23113RKC6C) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36 MiuiBrowser/18.5.30110",
			SecFetchSite:   "none", SecFetchMode: "navigate",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	HuaweiBrowser15 = ClientProfile{
		ID: "huawei_browser_15", Name: "Huawei Browser 15.0",
		BrowserType: core.BrowserChrome, BrowserVersion: "99.0.4844.88",
		OS: core.OSAndroid, OSVersion: "14",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc02c, 0xc030, 0xc013, 0xc014,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x000d}, {Type: 0x002b}, {Type: 0x002d},
			{Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 100,
			InitialWindowSize: 4194304, MaxFrameSize: 16384,
		},
		PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
		ConnectionFlow:    15663105,
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "zh-CN,zh;q=0.9",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Linux; Android 14; NOH-AN00) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.88 HuaweiBrowser/15.0.3.300 Mobile Safari/537.36",
			SecFetchSite:   "none", SecFetchMode: "navigate",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	OPPOBrowser8 = ClientProfile{
		ID: "oppo_browser_8", Name: "OPPO Browser 8.10",
		BrowserType: core.BrowserChrome, BrowserVersion: "120.0.0.0",
		OS: core.OSAndroid, OSVersion: "14",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc02c, 0xc030, 0xc013, 0xc014,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x000d}, {Type: 0x002b}, {Type: 0x002d},
			{Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 100,
			InitialWindowSize: 4194304, MaxFrameSize: 16384,
		},
		PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
		ConnectionFlow:    15663105,
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "zh-CN,zh;q=0.9",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Linux; Android 14; CPH2569) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36 OppoBrowser/8.10.0.0",
			SecFetchSite:   "none", SecFetchMode: "navigate",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}
)

// =========================================================================
// Privacy browsers (Tor, DuckDuckGo)
// =========================================================================

var (
	TorBrowser13 = ClientProfile{
		ID: "tor_browser_13", Name: "Tor Browser 13.5",
		BrowserType: core.BrowserFirefox, BrowserVersion: "115.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc02c, 0xc030,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x000d}, {Type: 0x002b}, {Type: 0x002d},
			{Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 100,
			InitialWindowSize: 131072, MaxFrameSize: 16384,
		},
		PseudoHeaderOrder: []string{":method", ":path", ":authority", ":scheme"},
		ConnectionFlow:    12517377,
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; rv:109.0) Gecko/20100101 Firefox/115.0",
			SecFetchSite:   "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		TCPIP: CreateTCPIP(core.OSLinuxUbuntu),
	}

	DuckDuckGo7_Mac = ClientProfile{
		ID: "duckduckgo_7_mac", Name: "DuckDuckGo 7 macOS",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.0",
		OS: core.OSMacOS14, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc02b, 0xc030, 0xc02f,
			0xcca9, 0xcca8,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x000d}, {Type: 0x002b}, {Type: 0x002d},
			{Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 4096, EnablePush: 1, MaxConcurrentStreams: 100,
			InitialWindowSize: 2097152, MaxFrameSize: 16384,
		},
		PseudoHeaderOrder: []string{":method", ":scheme", ":path", ":authority"},
		ConnectionFlow:    10485760,
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 DuckDuckGo/7 Safari/605.1.15",
			SecFetchSite:   "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		TCPIP: CreateTCPIP(core.OSMacOS14),
	}
)

// =========================================================================
// Chinese desktop browsers (360 Safe, Sogou, Baidu)
// =========================================================================

var (
	Browser360Safe15 = ClientProfile{
		ID: "360_safe_15", Name: "360 Safe Browser 15.3",
		BrowserType: core.BrowserChrome, BrowserVersion: "122.0.0.0",
		OS: core.OSWindows10, OSVersion: "10.0.19045",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x000d}, {Type: 0x002b}, {Type: 0x002d},
			{Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 1000,
			InitialWindowSize: 6291456, MaxFrameSize: 16384, MaxHeaderListSize: 262144,
		},
		PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
		ConnectionFlow:    15663105,
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "zh-CN,zh;q=0.9",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36 360SE/15.3.1055.64",
			SecFetchSite:   "none", SecFetchMode: "navigate",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	SogouBrowser12 = ClientProfile{
		ID: "sogou_browser_12", Name: "Sogou Browser 12.4",
		BrowserType: core.BrowserChrome, BrowserVersion: "108.0.0.0",
		OS: core.OSWindows10, OSVersion: "10.0.19045",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x000d}, {Type: 0x002b}, {Type: 0x002d},
			{Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 1000,
			InitialWindowSize: 6291456, MaxFrameSize: 16384, MaxHeaderListSize: 262144,
		},
		PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
		ConnectionFlow:    15663105,
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "zh-CN,zh;q=0.9",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 SogouMobileBrowser/12.4.0",
			SecFetchSite:   "none", SecFetchMode: "navigate",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	BaiduBrowser7 = ClientProfile{
		ID: "baidu_browser_7", Name: "Baidu Browser 7.65",
		BrowserType: core.BrowserChrome, BrowserVersion: "100.0.4896.127",
		OS: core.OSAndroid, OSVersion: "14",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc02c, 0xc030, 0xc013, 0xc014,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x000d}, {Type: 0x002b}, {Type: 0x002d},
			{Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 100,
			InitialWindowSize: 4194304, MaxFrameSize: 16384,
		},
		PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
		ConnectionFlow:    15663105,
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "zh-CN,zh;q=0.9",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Linux; Android 14; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Mobile Safari/537.36 baiduboxapp/7.65.0.10",
			SecFetchSite:   "none", SecFetchMode: "navigate",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}
)

// =========================================================================
// Arc Browser (Global) — Chromium-based
// =========================================================================

var (
	Arc1_50 = ClientProfile{
		ID: "arc_1_50", Name: "Arc 1.50",
		BrowserType: core.BrowserChrome, BrowserVersion: "124.0.0.0",
		OS: core.OSMacOS14, OSVersion: "14.5",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc02c, 0xc030, 0xc013, 0xc014,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0010}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x002d}, {Type: 0x0033}, {Type: 0x0015},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 1000,
			InitialWindowSize: 6291456, MaxFrameSize: 16384, MaxHeaderListSize: 262144,
		},
		PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
		ConnectionFlow:    15663105,
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_5_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
			SecFetchSite:   "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		TCPIP: CreateTCPIP(core.OSMacOS14),
	}

	Arc1_60 = ClientProfile{
		ID: "arc_1_60", Name: "Arc 1.60",
		BrowserType: core.BrowserChrome, BrowserVersion: "128.0.0.0",
		OS: core.OSMacOS15, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc02c, 0xc030, 0xc013, 0xc014,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0010}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x002d}, {Type: 0x0033}, {Type: 0x0015},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 1000,
			InitialWindowSize: 6291456, MaxFrameSize: 16384, MaxHeaderListSize: 262144,
		},
		PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
		ConnectionFlow:    15663105,
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9",
			AcceptEncoding: "gzip, deflate, br, zstd",
			UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 15_0_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
			SecFetchSite:   "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		TCPIP: CreateTCPIP(core.OSMacOS15),
	}
)

func init() {
	all := []ClientProfile{
		// Yandex (Russia/CIS)
		Yandex24_4, Yandex24_7, Yandex24_7_Mac, Yandex24_10,
		// Vivaldi (Europe)
		Vivaldi6_6, Vivaldi6_8, Vivaldi6_9,
		// QQ Browser (China)
		QQBrowser14, QQBrowser14_5_Android,
		// UC Browser (China/India/SEA)
		UCBrowser16_Android, UCBrowser16_iOS,
		// Naver Whale (Korea)
		Whale3_25, Whale3_28,
		// Chinese OEM mobile
		MiBrowser18, HuaweiBrowser15, OPPOBrowser8,
		// Privacy browsers
		TorBrowser13, DuckDuckGo7_Mac,
		// Chinese desktop
		Browser360Safe15, SogouBrowser12, BaiduBrowser7,
		// Arc
		Arc1_50, Arc1_60,
	}

	for i := range all {
		p := &all[i]

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

		if p.TCPIP == nil {
			p.TCPIP = CreateTCPIP(p.OS)
		}

		Register(*p)
	}
}
