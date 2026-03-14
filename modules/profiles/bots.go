// Package profiles - Bot/headless browser and anti-detect tool fingerprints
// These profiles serve as training data for the forgery detection model,
// representing automated tools (Puppeteer, Playwright, Selenium), HTTP
// clients (cURL, Go, Python, Node.js), and anti-detect browsers
// (Multilogin, GoLogin, Dolphin Anty).
package profiles

import (
	"github.com/vistone/fingerprint/modules/core"
)

// =========================================================================
// Headless Chrome / Puppeteer
// =========================================================================

var (
	PuppeteerHeadless120 = ClientProfile{
		ID: "puppeteer_headless_120", Name: "Puppeteer Headless Chrome 120",
		BrowserType: core.BrowserChrome, BrowserVersion: "120.0.0.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
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
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/120.0.0.0 Safari/537.36",
			SecCHUA:        `"Chromium";v="120", "HeadlessChrome";v="120", "Not-A.Brand";v="99"`,
			SecCHUAMobile:  "?0", SecCHUAPlatform: `"Linux"`,
			SecFetchSite: "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		JSAntiDetection: &JSAntiDetection{
			Automation: &AutomationAntiDetect{
				WebDriver: true, Headless: true, Puppeteer: true,
			},
		},
		TCPIP: CreateTCPIP(core.OSLinuxUbuntu),
		Metadata: map[string]interface{}{
			"bot_type": "headless", "tool": "puppeteer", "forgery": "headless",
		},
	}

	PuppeteerHeadless124 = ClientProfile{
		ID: "puppeteer_headless_124", Name: "Puppeteer Headless Chrome 124",
		BrowserType: core.BrowserChrome, BrowserVersion: "124.0.0.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
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
			{Type: 0x0015}, {Type: 0x001b},
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
			UserAgent:      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/124.0.0.0 Safari/537.36",
			SecCHUA:        `"Chromium";v="124", "HeadlessChrome";v="124", "Not-A.Brand";v="99"`,
			SecCHUAMobile:  "?0", SecCHUAPlatform: `"Linux"`,
			SecFetchSite: "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		JSAntiDetection: &JSAntiDetection{
			Automation: &AutomationAntiDetect{
				WebDriver: true, Headless: true, Puppeteer: true,
			},
		},
		TCPIP: CreateTCPIP(core.OSLinuxUbuntu),
		Metadata: map[string]interface{}{
			"bot_type": "headless", "tool": "puppeteer", "forgery": "headless",
		},
	}

	PuppeteerStealth120 = ClientProfile{
		ID: "puppeteer_stealth_120", Name: "Puppeteer Stealth Chrome 120",
		BrowserType: core.BrowserChrome, BrowserVersion: "120.0.0.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
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
			UserAgent:      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			SecCHUA:        `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`,
			SecCHUAMobile:  "?0", SecCHUAPlatform: `"Linux"`,
			SecFetchSite: "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		JSAntiDetection: &JSAntiDetection{
			Automation: &AutomationAntiDetect{
				WebDriver: false, Headless: true, Puppeteer: true,
				PluginsOverride: true, LanguageOverride: true, VendorOverride: true,
				RuntimeOverride: true,
			},
		},
		TCPIP: CreateTCPIP(core.OSLinuxUbuntu),
		Metadata: map[string]interface{}{
			"bot_type": "stealth", "tool": "puppeteer-extra-stealth", "forgery": "antidetect",
		},
	}
)

// =========================================================================
// Playwright
// =========================================================================

var (
	PlaywrightChromium120 = ClientProfile{
		ID: "playwright_chromium_120", Name: "Playwright Chromium 120",
		BrowserType: core.BrowserChrome, BrowserVersion: "120.0.0.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
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
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			SecCHUA:        `"Chromium";v="120", "Not-A.Brand";v="99"`,
			SecCHUAMobile:  "?0", SecCHUAPlatform: `"Linux"`,
			SecFetchSite: "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		JSAntiDetection: &JSAntiDetection{
			Automation: &AutomationAntiDetect{
				WebDriver: true, Headless: true, Playwright: true,
			},
		},
		TCPIP: CreateTCPIP(core.OSLinuxUbuntu),
		Metadata: map[string]interface{}{
			"bot_type": "headless", "tool": "playwright", "forgery": "headless",
		},
	}

	PlaywrightFirefox121 = ClientProfile{
		ID: "playwright_firefox_121", Name: "Playwright Firefox 121",
		BrowserType: core.BrowserFirefox, BrowserVersion: "121.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
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
			InitialWindowSize: 131072, MaxFrameSize: 16384,
		},
		PseudoHeaderOrder: []string{":method", ":path", ":authority", ":scheme"},
		ConnectionFlow:    12517377,
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0",
			SecFetchSite:   "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		JSAntiDetection: &JSAntiDetection{
			Automation: &AutomationAntiDetect{
				WebDriver: true, Headless: true, Playwright: true,
			},
		},
		TCPIP: CreateTCPIP(core.OSLinuxUbuntu),
		Metadata: map[string]interface{}{
			"bot_type": "headless", "tool": "playwright", "forgery": "headless",
		},
	}

	PlaywrightWebkit17 = ClientProfile{
		ID: "playwright_webkit_17", Name: "Playwright WebKit 17.4",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.4",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
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
			UserAgent:      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Safari/605.1.15",
			SecFetchSite:   "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		JSAntiDetection: &JSAntiDetection{
			Automation: &AutomationAntiDetect{
				WebDriver: true, Headless: true, Playwright: true,
			},
		},
		TCPIP: CreateTCPIP(core.OSLinuxUbuntu),
		Metadata: map[string]interface{}{
			"bot_type": "headless", "tool": "playwright", "forgery": "headless",
		},
	}
)

// =========================================================================
// Selenium
// =========================================================================

var (
	SeleniumChromeDriver120 = ClientProfile{
		ID: "selenium_chromedriver_120", Name: "Selenium ChromeDriver 120",
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
				WebDriver: true, Selenium: true, ChromeDebugPort: true,
			},
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
		Metadata: map[string]interface{}{
			"bot_type": "selenium", "tool": "selenium-chromedriver", "forgery": "headless",
		},
	}

	SeleniumGeckoDriver121 = ClientProfile{
		ID: "selenium_geckodriver_121", Name: "Selenium GeckoDriver/Firefox 121",
		BrowserType: core.BrowserFirefox, BrowserVersion: "121.0",
		OS: core.OSWindows10, OSVersion: "10.0.19045",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
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
			InitialWindowSize: 131072, MaxFrameSize: 16384,
		},
		PseudoHeaderOrder: []string{":method", ":path", ":authority", ":scheme"},
		ConnectionFlow:    12517377,
		Headers: &core.HTTPHeaders{
			Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5",
			AcceptEncoding: "gzip, deflate, br",
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
			SecFetchSite:   "none", SecFetchMode: "navigate", SecFetchDest: "document",
		},
		JSAntiDetection: &JSAntiDetection{
			Automation: &AutomationAntiDetect{
				WebDriver: true, Selenium: true,
			},
		},
		TCPIP: CreateTCPIP(core.OSWindows10),
		Metadata: map[string]interface{}{
			"bot_type": "selenium", "tool": "selenium-geckodriver", "forgery": "headless",
		},
	}
)

// =========================================================================
// HTTP Client Libraries (non-browser TLS fingerprints)
// =========================================================================
