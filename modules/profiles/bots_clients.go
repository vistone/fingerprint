package profiles

import (
	"github.com/vistone/fingerprint/modules/core"
)

var (
	CurlOpenSSL = ClientProfile{
		ID: "curl_openssl", Name: "cURL/libcurl with OpenSSL",
		BrowserType: core.BrowserChrome, BrowserVersion: "0.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0x009f,
			0xcca9, 0xcca8, 0xccaa, 0xc02b, 0xc02f, 0x009e,
			0xc024, 0xc028, 0x006b, 0xc023, 0xc027, 0x0067,
			0xc00a, 0xc014, 0x0039, 0xc009, 0xc013, 0x0033,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x000b}, {Type: 0x000a},
			{Type: 0x000d}, {Type: 0x0023}, {Type: 0xff01},
			{Type: 0x0010}, {Type: 0x002b}, {Type: 0x002d},
			{Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 4096, EnablePush: 0, MaxConcurrentStreams: 100,
			InitialWindowSize: 65535, MaxFrameSize: 16384,
		},
		PseudoHeaderOrder: []string{":method", ":path", ":scheme", ":authority"},
		ConnectionFlow:    65535,
		Headers: &core.HTTPHeaders{
			Accept:         "*/*",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "curl/8.4.0",
		},
		TCPIP: CreateTCPIP(core.OSLinuxUbuntu),
		Metadata: map[string]interface{}{
			"bot_type": "http_client", "tool": "curl", "forgery": "non_browser",
		},
	}

	GoHTTPClient = ClientProfile{
		ID: "go_http_client", Name: "Go net/http default",
		BrowserType: core.BrowserChrome, BrowserVersion: "0.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f,
			0xc02c, 0xc030, 0xc009, 0xc013, 0xc00a, 0xc014,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x000a}, {Type: 0x000b},
			{Type: 0x000d}, {Type: 0x002b}, {Type: 0x002d},
			{Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 4096, EnablePush: 0, MaxConcurrentStreams: 250,
			InitialWindowSize: 1048576, MaxFrameSize: 16384,
		},
		PseudoHeaderOrder: []string{":method", ":path", ":scheme", ":authority"},
		ConnectionFlow:    1073741824,
		Headers: &core.HTTPHeaders{
			Accept:         "*/*",
			AcceptEncoding: "gzip",
			UserAgent:      "Go-http-client/2.0",
		},
		TCPIP: CreateTCPIP(core.OSLinuxUbuntu),
		Metadata: map[string]interface{}{
			"bot_type": "http_client", "tool": "go-net-http", "forgery": "non_browser",
		},
	}

	PythonRequests = ClientProfile{
		ID: "python_requests", Name: "Python requests/urllib3",
		BrowserType: core.BrowserChrome, BrowserVersion: "0.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0x009f,
			0xcca9, 0xcca8, 0xccaa, 0xc02b, 0xc02f, 0x009e,
			0xc024, 0xc028, 0x006b, 0xc023, 0xc027, 0x0067,
			0xc00a, 0xc014, 0x0039, 0xc009, 0xc013, 0x0033,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x000b}, {Type: 0x000a},
			{Type: 0x000d}, {Type: 0x0023}, {Type: 0xff01},
			{Type: 0x0010}, {Type: 0x002b}, {Type: 0x002d},
			{Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 4096, EnablePush: 0, MaxConcurrentStreams: 100,
			InitialWindowSize: 65535, MaxFrameSize: 16384,
		},
		PseudoHeaderOrder: []string{":method", ":path", ":scheme", ":authority"},
		ConnectionFlow:    65535,
		Headers: &core.HTTPHeaders{
			Accept:         "*/*",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "python-requests/2.31.0",
		},
		TCPIP: CreateTCPIP(core.OSLinuxUbuntu),
		Metadata: map[string]interface{}{
			"bot_type": "http_client", "tool": "python-requests", "forgery": "non_browser",
		},
	}

	NodeAxios = ClientProfile{
		ID: "node_axios", Name: "Node.js Axios/fetch",
		BrowserType: core.BrowserChrome, BrowserVersion: "0.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f,
			0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0x000a},
			{Type: 0x000b}, {Type: 0x000d}, {Type: 0xff01},
			{Type: 0x0023}, {Type: 0x0010}, {Type: 0x002b},
			{Type: 0x002d}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 4096, EnablePush: 0, MaxConcurrentStreams: 100,
			InitialWindowSize: 65535, MaxFrameSize: 16384,
		},
		PseudoHeaderOrder: []string{":method", ":path", ":scheme", ":authority"},
		ConnectionFlow:    65535,
		Headers: &core.HTTPHeaders{
			Accept:         "*/*",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "axios/1.6.0",
		},
		TCPIP: CreateTCPIP(core.OSLinuxUbuntu),
		Metadata: map[string]interface{}{
			"bot_type": "http_client", "tool": "node-axios", "forgery": "non_browser",
		},
	}

	Scrapy = ClientProfile{
		ID: "scrapy_bot", Name: "Scrapy Web Crawler",
		BrowserType: core.BrowserChrome, BrowserVersion: "0.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0x009f,
			0xcca9, 0xcca8, 0xccaa, 0xc02b, 0xc02f, 0x009e,
			0xc024, 0xc028, 0xc00a, 0xc014, 0xc009, 0xc013,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x000b}, {Type: 0x000a},
			{Type: 0x000d}, {Type: 0x0023}, {Type: 0xff01},
			{Type: 0x002b}, {Type: 0x002d}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en",
			AcceptEncoding: "gzip, deflate",
			UserAgent:      "Scrapy/2.11.0 (+https://scrapy.org)",
		},
		TCPIP: CreateTCPIP(core.OSLinuxUbuntu),
		Metadata: map[string]interface{}{
			"bot_type": "crawler", "tool": "scrapy", "forgery": "non_browser",
		},
	}
)

// =========================================================================
// Anti-Detect Browsers (Multilogin, GoLogin, Dolphin Anty)
// =========================================================================

var (
	MultiloginMimic10 = ClientProfile{
		ID: "multilogin_mimic_10", Name: "Multilogin Mimic 10 (Chrome 120 spoof)",
		BrowserType: core.BrowserChrome, BrowserVersion: "120.0.0.0",
		OS: core.OSWindows10, OSVersion: "10.0.19045",
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
			AcceptLanguage: "en-US,en;q=0.9",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			SecCHUA:        `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`,
			SecCHUAMobile:  "?0", SecCHUAPlatform: `"Windows"`,
			SecFetchSite: "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		JSAntiDetection: &JSAntiDetection{
			Automation: &AutomationAntiDetect{
				WebDriver: false, Headless: false,
				PluginsOverride: true, LanguageOverride: true,
				VendorOverride: true, RuntimeOverride: true,
			},
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
		Metadata: map[string]interface{}{
			"bot_type": "antidetect", "tool": "multilogin", "forgery": "antidetect",
		},
	}

	GoLoginChrome124 = ClientProfile{
		ID: "gologin_chrome_124", Name: "GoLogin Chrome 124 Profile",
		BrowserType: core.BrowserChrome, BrowserVersion: "124.0.0.0",
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
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9",
			AcceptEncoding: "gzip, deflate, br, zstd",
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
			SecCHUA:        `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`,
			SecCHUAMobile:  "?0", SecCHUAPlatform: `"Windows"`,
			SecFetchSite: "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		JSAntiDetection: &JSAntiDetection{
			Automation: &AutomationAntiDetect{
				WebDriver: false, Headless: false,
				PluginsOverride: true, LanguageOverride: true,
			},
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
		Metadata: map[string]interface{}{
			"bot_type": "antidetect", "tool": "gologin", "forgery": "antidetect",
		},
	}

	DolphinAntyChrome = ClientProfile{
		ID: "dolphin_anty_chrome", Name: "Dolphin Anty Chrome Profile",
		BrowserType: core.BrowserChrome, BrowserVersion: "120.0.0.0",
		OS: core.OSWindows10, OSVersion: "10.0.19045",
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
			AcceptLanguage: "en-US,en;q=0.9,ru;q=0.8",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			SecCHUA:        `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`,
			SecCHUAMobile:  "?0", SecCHUAPlatform: `"Windows"`,
			SecFetchSite: "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		JSAntiDetection: &JSAntiDetection{
			Automation: &AutomationAntiDetect{
				WebDriver: false, Headless: false,
				PluginsOverride: true, VendorOverride: true, RuntimeOverride: true,
			},
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
		Metadata: map[string]interface{}{
			"bot_type": "antidetect", "tool": "dolphin-anty", "forgery": "antidetect",
		},
	}

	CloudflareWorker = ClientProfile{
		ID: "cloudflare_worker", Name: "Cloudflare Worker fetch()",
		BrowserType: core.BrowserChrome, BrowserVersion: "0.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f,
			0xcca9, 0xcca8, 0xc02c, 0xc030,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0x000a},
			{Type: 0x000b}, {Type: 0x000d}, {Type: 0xff01},
			{Type: 0x002b}, {Type: 0x002d}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 4096, EnablePush: 0, MaxConcurrentStreams: 100,
			InitialWindowSize: 65535, MaxFrameSize: 16384,
		},
		PseudoHeaderOrder: []string{":method", ":path", ":scheme", ":authority"},
		ConnectionFlow:    65535,
		Headers: &core.HTTPHeaders{
			Accept:         "*/*",
			AcceptEncoding: "gzip",
			UserAgent:      "Cloudflare-Workers",
		},
		TCPIP: CreateTCPIP(core.OSLinuxUbuntu),
		Metadata: map[string]interface{}{
			"bot_type": "serverless", "tool": "cloudflare-workers", "forgery": "non_browser",
		},
	}
)

func init() {
	all := []ClientProfile{
		PuppeteerHeadless120, PuppeteerHeadless124, PuppeteerStealth120,
		PlaywrightChromium120, PlaywrightFirefox121, PlaywrightWebkit17,
		SeleniumChromeDriver120, SeleniumGeckoDriver121,
		CurlOpenSSL, GoHTTPClient, PythonRequests, NodeAxios, Scrapy,
		MultiloginMimic10, GoLoginChrome124, DolphinAntyChrome,
		CloudflareWorker,
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
