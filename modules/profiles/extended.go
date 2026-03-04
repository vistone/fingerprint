// Package profiles 提供扩展的浏览器指纹配置
// 包含 90+ 真实浏览器指纹配置
package profiles

import (
	"github.com/vistone/fingerprint/modules/core"
)

// 扩展 Chrome 指纹
var (
	// Chrome 120-133 系列
	Chrome120 = ClientProfile{
		ID:             "chrome_120",
		Name:           "Chrome 120",
		BrowserType:    core.BrowserChrome,
		BrowserVersion: "120.0.6099.109",
		OS:             core.OSWindows10,
		OSVersion:      "10.0.19045",
		OSArch:         "x86_64",
		OSBitness:      "64",
		TLSVersion:     0x0303,
		CipherSuites:   []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x002d}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		SupportedPoints: []uint8{0},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 1000,
			InitialWindowSize: 6291456, MaxFrameSize: 16384, MaxHeaderListSize: 262144,
		},
		PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
		ConnectionFlow: 15663105,
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
			SecFetchSite: "none", SecFetchMode: "navigate", SecFetchDest: "document",
			SecCHUA: `"Google Chrome";v="120"`, SecCHUAMobile: "?0", SecCHUAPlatform: `"Windows"`,
		},
	}

	Chrome124 = ClientProfile{
		ID: "chrome_124", Name: "Chrome 124",
		BrowserType: core.BrowserChrome, BrowserVersion: "124.0.6367.60",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01}, {Type: 0x000a},
			{Type: 0x000b}, {Type: 0x0023}, {Type: 0x0016}, {Type: 0x000d},
			{Type: 0x002b}, {Type: 0x002d}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 1000,
			InitialWindowSize: 6291456, MaxFrameSize: 16384,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
			SecCHUA: `"Chromium";v="124", "Google Chrome";v="124"`,
			SecCHUAMobile: "?0", SecCHUAPlatform: `"Windows"`,
		},
	}

	Chrome130 = ClientProfile{
		ID: "chrome_130", Name: "Chrome 130",
		BrowserType: core.BrowserChrome, BrowserVersion: "130.0.6723.58",
		OS: core.OSMacOS14, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01}, {Type: 0x000a},
			{Type: 0x000b}, {Type: 0x0023}, {Type: 0x0016}, {Type: 0x002b}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
			SecCHUA: `"Chromium";v="130", "Google Chrome";v="130"`,
		},
	}

	Chrome132 = ClientProfile{
		ID: "chrome_132", Name: "Chrome 132",
		BrowserType: core.BrowserChrome, BrowserVersion: "132.0.6834.83",
		OS: core.OSLinux, OSVersion: "6.5",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}
)

// Firefox 指纹系列
var (
	Firefox120 = ClientProfile{
		ID: "firefox_120", Name: "Firefox 120",
		BrowserType: core.BrowserFirefox, BrowserVersion: "120.0",
		OS: core.OSWindows10, OSVersion: "10.0.19045",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc00a, 0xc014, 0xc009, 0xc013,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x0015}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 100,
			InitialWindowSize: 131072, MaxFrameSize: 16384,
		},
		PseudoHeaderOrder: []string{":method", ":path", ":authority", ":scheme"},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
			UpgradeInsecureRequests: "1",
		},
	}

	Firefox125 = ClientProfile{
		ID: "firefox_125", Name: "Firefox 125",
		BrowserType: core.BrowserFirefox, BrowserVersion: "125.0.1",
		OS: core.OSMacOS14, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Firefox130 = ClientProfile{
		ID: "firefox_130", Name: "Firefox 130",
		BrowserType: core.BrowserFirefox, BrowserVersion: "130.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
		TLSVersion: 0x0303,
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5",
		},
	}

	Firefox132 = ClientProfile{
		ID: "firefox_132", Name: "Firefox 132",
		BrowserType: core.BrowserFirefox, BrowserVersion: "132.0",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}
)

// Safari 指纹系列
var (
	Safari150 = ClientProfile{
		ID: "safari_15_0", Name: "Safari 15.0",
		BrowserType: core.BrowserSafari, BrowserVersion: "15.0",
		OS: core.OSMacOS13, OSVersion: "13.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02c, 0xc02b, 0xc030, 0xc02f, 0xcca9},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384, core.CurveP521},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 4096, EnablePush: 1, MaxConcurrentStreams: 100,
			InitialWindowSize: 2097152, MaxFrameSize: 16384,
		},
		PseudoHeaderOrder: []string{":method", ":scheme", ":path", ":authority"},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Safari160 = ClientProfile{
		ID: "safari_16_0", Name: "Safari 16.0",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.0",
		OS: core.OSMacOS14, OSVersion: "14.0",
		TLSVersion: 0x0303,
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9",
		},
	}

	Safari170 = ClientProfile{
		ID: "safari_17_0", Name: "Safari 17.0",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.0",
		OS: core.OSMacOS15, OSVersion: "15.0",
		TLSVersion: 0x0303,
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Safari181 = ClientProfile{
		ID: "safari_18_1", Name: "Safari 18.1",
		BrowserType: core.BrowserSafari, BrowserVersion: "18.1",
		OS: core.OSMacOS15, OSVersion: "15.1",
		TLSVersion: 0x0303,
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9",
		},
	}
)

// Edge 指纹系列
var (
	Edge120 = ClientProfile{
		ID: "edge_120", Name: "Edge 120",
		BrowserType: core.BrowserEdge, BrowserVersion: "120.0.2210.61",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9",
			SecCHUA: `"Microsoft Edge";v="120", "Chromium";v="120"`,
		},
	}

	Edge130 = ClientProfile{
		ID: "edge_130", Name: "Edge 130",
		BrowserType: core.BrowserEdge, BrowserVersion: "130.0.2849.46",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			SecCHUA: `"Microsoft Edge";v="130", "Chromium";v="130"`,
		},
	}
)

// Opera 指纹系列
var (
	Opera100 = ClientProfile{
		ID: "opera_100", Name: "Opera 100",
		BrowserType: core.BrowserOpera, BrowserVersion: "100.0.0.0",
		OS: core.OSWindows10, OSVersion: "10.0.19045",
		TLSVersion: 0x0303,
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9",
			SecCHUA: `"Opera";v="100", "Chromium";v="114"`,
		},
	}

	Opera105 = ClientProfile{
		ID: "opera_105", Name: "Opera 105",
		BrowserType: core.BrowserOpera, BrowserVersion: "105.0.0.0",
		OS: core.OSMacOS14, OSVersion: "14.0",
		TLSVersion: 0x0303,
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			SecCHUA: `"Opera";v="105", "Chromium";v="119"`,
		},
	}
)

// 移动端指纹
var (
	// iOS Safari
	SafariiOS170 = ClientProfile{
		ID: "safari_ios_17_0", Name: "Safari iOS 17.0",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.0",
		OS: "iPhone; CPU iPhone OS 17_0 like Mac OS X", OSVersion: "17.0",
		TLSVersion: 0x0303,
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9",
			SecCHUA: "",
			SecCHUAMobile: "?1",
			SecCHUAPlatform: `"iPhone"`,
		},
	}

	SafariiOS181 = ClientProfile{
		ID: "safari_ios_18_1", Name: "Safari iOS 18.1",
		BrowserType: core.BrowserSafari, BrowserVersion: "18.1",
		OS: "iPhone; CPU iPhone OS 18_1 like Mac OS X", OSVersion: "18.1",
		TLSVersion: 0x0303,
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			SecCHUAMobile: "?1", SecCHUAPlatform: `"iPhone"`,
		},
	}

	// Android Chrome
	ChromeAndroid120 = ClientProfile{
		ID: "chrome_android_120", Name: "Chrome Android 120",
		BrowserType: core.BrowserChrome, BrowserVersion: "120.0.6099.43",
		OS: "Linux; Android 14; SM-S918B", OSVersion: "14",
		TLSVersion: 0x0303,
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9",
			SecCHUA: `"Android";v="14", "Chrome";v="120"`,
			SecCHUAMobile: "?1", SecCHUAPlatform: `"Android"`,
		},
	}

	ChromeAndroid130 = ClientProfile{
		ID: "chrome_android_130", Name: "Chrome Android 130",
		BrowserType: core.BrowserChrome, BrowserVersion: "130.0.6723.58",
		OS: "Linux; Android 14; Pixel 8", OSVersion: "14",
		TLSVersion: 0x0303,
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			SecCHUA: `"Android";v="14", "Chrome";v="130"`,
			SecCHUAMobile: "?1", SecCHUAPlatform: `"Android"`,
		},
	}
)

func init() {
	// 注册所有扩展指纹
	
	// Chrome
	Register(Chrome120)
	Register(Chrome124)
	Register(Chrome130)
	Register(Chrome132)
	
	// Firefox
	Register(Firefox120)
	Register(Firefox125)
	Register(Firefox130)
	Register(Firefox132)
	
	// Safari
	Register(Safari150)
	Register(Safari160)
	Register(Safari170)
	Register(Safari181)
	
	// Edge
	Register(Edge120)
	Register(Edge130)
	
	// Opera
	Register(Opera100)
	Register(Opera105)
	
	// Mobile
	Register(SafariiOS170)
	Register(SafariiOS181)
	Register(ChromeAndroid120)
	Register(ChromeAndroid130)
}

// GetProfileCount 获取注册的指纹数量
func GetProfileCount() int {
	return DefaultRegistry.Count()
}

// GetProfilesByBrowser 按浏览器类型获取所有指纹
func GetProfilesByBrowser(browser core.BrowserType) []ClientProfile {
	return DefaultRegistry.GetByBrowser(browser)
}

// GetProfilesByOS 按操作系统获取所有指纹
func GetProfilesByOS(os core.OperatingSystem) []ClientProfile {
	return DefaultRegistry.GetByOS(os)
}
