// Package profiles - International browser fingerprints
// Contains Yandex, UC Browser, QQ Browser, Vivaldi, Whale, and other
// global browser fingerprint profiles used in China, Russia, India, Korea, etc.
package profiles

import (
	"github.com/vistone/fingerprint/modules/core"
)

// =========================================================================
// Yandex Browser (Russia/CIS) — Chromium-based
// =========================================================================

var (
	Yandex24_4 = ClientProfile{
		ID: "yandex_24_4", Name: "Yandex Browser 24.4",
		BrowserType: core.BrowserChrome, BrowserVersion: "124.0.0.0",
		OS: core.OSWindows10, OSVersion: "10.0.19045",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035,
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
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
			AcceptLanguage: "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 YaBrowser/24.4.0.0 Safari/537.36",
			SecCHUA:        `"Chromium";v="124", "YaBrowser";v="24.4", "Not-A.Brand";v="99"`,
			SecCHUAMobile:  "?0", SecCHUAPlatform: `"Windows"`,
			SecFetchSite: "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Yandex24_7 = ClientProfile{
		ID: "yandex_24_7", Name: "Yandex Browser 24.7",
		BrowserType: core.BrowserChrome, BrowserVersion: "126.0.0.0",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035,
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
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
			AcceptEncoding: "gzip, deflate, br, zstd",
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 YaBrowser/24.7.0.0 Safari/537.36",
			SecCHUA:        `"Chromium";v="126", "YaBrowser";v="24.7", "Not-A.Brand";v="99"`,
			SecCHUAMobile:  "?0", SecCHUAPlatform: `"Windows"`,
			SecFetchSite: "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Yandex24_7_Mac = ClientProfile{
		ID: "yandex_24_7_mac", Name: "Yandex Browser 24.7 macOS",
		BrowserType: core.BrowserChrome, BrowserVersion: "126.0.0.0",
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
			{Type: 0x002d}, {Type: 0x0033},
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
			AcceptLanguage: "ru-RU,ru;q=0.9,en;q=0.7",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_5_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 YaBrowser/24.7.0.0 Safari/537.36",
			SecCHUA:        `"Chromium";v="126", "YaBrowser";v="24.7", "Not-A.Brand";v="99"`,
			SecCHUAMobile:  "?0", SecCHUAPlatform: `"macOS"`,
			SecFetchSite: "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		TCPIP: CreateTCPIP(core.OSMacOS14),
	}

	Yandex24_10 = ClientProfile{
		ID: "yandex_24_10", Name: "Yandex Browser 24.10",
		BrowserType: core.BrowserChrome, BrowserVersion: "128.0.0.0",
		OS: core.OSWindows11, OSVersion: "10.0.26100",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0010}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x002d}, {Type: 0x001c}, {Type: 0x001b},
			{Type: 0x0033}, {Type: 0x0015},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 1000,
			InitialWindowSize: 6291456, MaxFrameSize: 16384, MaxHeaderListSize: 262144,
		},
		PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
		ConnectionFlow:    15663105,
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
			AcceptEncoding: "gzip, deflate, br, zstd",
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 YaBrowser/24.10.0.0 Safari/537.36",
			SecCHUA:        `"Chromium";v="128", "YaBrowser";v="24.10", "Not-A.Brand";v="99"`,
			SecCHUAMobile:  "?0", SecCHUAPlatform: `"Windows"`,
			SecFetchSite: "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}
)

// =========================================================================
// Vivaldi (Europe/Global) — Chromium-based
// =========================================================================

var (
	Vivaldi6_6 = ClientProfile{
		ID: "vivaldi_6_6", Name: "Vivaldi 6.6",
		BrowserType: core.BrowserChrome, BrowserVersion: "122.0.0.0",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0010}, {Type: 0x0005}, {Type: 0x000d},
			{Type: 0x002b}, {Type: 0x002d}, {Type: 0x0033},
			{Type: 0x0015},
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
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
			SecFetchSite:   "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Vivaldi6_8 = ClientProfile{
		ID: "vivaldi_6_8", Name: "Vivaldi 6.8",
		BrowserType: core.BrowserChrome, BrowserVersion: "124.0.0.0",
		OS: core.OSLinuxUbuntu, OSVersion: "24.04",
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
			UserAgent:      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
			SecFetchSite:   "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		TCPIP: CreateTCPIP(core.OSLinuxUbuntu),
	}

	Vivaldi6_9 = ClientProfile{
		ID: "vivaldi_6_9", Name: "Vivaldi 6.9",
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
			{Type: 0x002d}, {Type: 0x0033},
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

// =========================================================================
// QQ Browser (China) — Chromium-based
// =========================================================================

var (
	QQBrowser14 = ClientProfile{
		ID: "qq_browser_14", Name: "QQ Browser 14.0",
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
			{Type: 0x0010}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x002d}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 1000,
			InitialWindowSize: 6291456, MaxFrameSize: 16384, MaxHeaderListSize: 262144,
		},
		PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
		ConnectionFlow:    15663105,
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "zh-CN,zh;q=0.9,en;q=0.8",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36 QQBrowser/14.0.0.0",
			SecFetchSite:   "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	QQBrowser14_5_Android = ClientProfile{
		ID: "qq_browser_14_5_android", Name: "QQ Browser 14.5 Android",
		BrowserType: core.BrowserChrome, BrowserVersion: "122.0.0.0",
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
			UserAgent:      "Mozilla/5.0 (Linux; Android 14; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/122.0.0.0 Mobile MQQBrowser/14.5 QQ/9.0.0 Safari/537.36",
			SecFetchSite:   "none", SecFetchMode: "navigate",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}
)

// =========================================================================
// UC Browser (China/India/SEA) — Chromium-based
// =========================================================================

var (
	UCBrowser16_Android = ClientProfile{
		ID: "uc_browser_16_android", Name: "UC Browser 16.5 Android",
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
			AcceptLanguage: "en-US,en;q=0.9,zh-CN;q=0.8",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Linux; U; Android 14; en-US; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/100.0.4896.127 UCBrowser/16.5.0.1 Mobile Safari/537.36",
			SecFetchSite:   "none", SecFetchMode: "navigate",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	UCBrowser16_iOS = ClientProfile{
		ID: "uc_browser_16_ios", Name: "UC Browser 16.3 iOS",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.3",
		OS: core.OSiOS, OSVersion: "17.0",
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
			AcceptLanguage: "zh-CN,zh-Hans;q=0.9",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1 UCBrowser/16.3.0.0",
			SecFetchSite:   "none", SecFetchMode: "navigate",
		},
		TCPIP: CreateTCPIP(core.OSiOS),
	}
)

// =========================================================================
// Naver Whale (South Korea) — Chromium-based
// =========================================================================

var (
	Whale3_25 = ClientProfile{
		ID: "whale_3_25", Name: "Naver Whale 3.25",
		BrowserType: core.BrowserChrome, BrowserVersion: "120.0.0.0",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0010}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x002d}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 1000,
			InitialWindowSize: 6291456, MaxFrameSize: 16384, MaxHeaderListSize: 262144,
		},
		PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
		ConnectionFlow:    15663105,
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Whale/3.25.232.19 Safari/537.36",
			SecCHUA:        `"Chromium";v="120", "Whale";v="3.25", "Not-A.Brand";v="99"`,
			SecCHUAMobile:  "?0", SecCHUAPlatform: `"Windows"`,
			SecFetchSite: "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}

	Whale3_28 = ClientProfile{
		ID: "whale_3_28", Name: "Naver Whale 3.28",
		BrowserType: core.BrowserChrome, BrowserVersion: "122.0.0.0",
		OS: core.OSWindows11, OSVersion: "10.0.26100",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0010}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x002d}, {Type: 0x0033},
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
			AcceptLanguage: "ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7",
			AcceptEncoding: "gzip, deflate, br, zstd",
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Whale/3.28.266.11 Safari/537.36",
			SecCHUA:        `"Chromium";v="122", "Whale";v="3.28", "Not-A.Brand";v="99"`,
			SecCHUAMobile:  "?0", SecCHUAPlatform: `"Windows"`,
			SecFetchSite: "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
	}
)
