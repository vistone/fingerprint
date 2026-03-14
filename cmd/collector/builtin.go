package main

import (
	"fmt"
	"os"
	"time"
)

func collectBotFingerprints() []CollectedFingerprint {
	type botEntry struct {
		Name       string
		Tool       string
		Ciphers    []uint16
		Extensions []uint16
		Curves     []uint16
		JA3        string
		UA         string
		Notes      string
	}

	bots := []botEntry{
		// Headless Chrome (Puppeteer)
		{
			Name: "headless_chrome_puppeteer", Tool: "puppeteer",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			JA3:        "b32309a26951912be7dba376398abc3b",
			UA:         "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/120.0.0.0 Safari/537.36",
			Notes:      "Headless mode detection: navigator.webdriver=true, missing plugins",
		},
		// Playwright Chromium
		{
			Name: "playwright_chromium", Tool: "playwright",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			JA3:        "b32309a26951912be7dba376398abc3b",
			UA:         "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			Notes:      "Playwright: CDP protocol, modified navigator properties",
		},
		// Playwright Firefox
		{
			Name: "playwright_firefox", Tool: "playwright",
			Ciphers:    []uint16{0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018, 0x0100},
			JA3:        "e058f2d45805cdf9e8e25af97b4a9a66",
			UA:         "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0",
		},
		// Playwright WebKit
		{
			Name: "playwright_webkit", Tool: "playwright",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02c, 0xc02b, 0xc030, 0xc02f, 0xcca9, 0xcca8},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
		},
		// Selenium ChromeDriver
		{
			Name: "selenium_chromedriver", Tool: "selenium",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			JA3:        "cd08e31494f9531f560d64c695473da9",
			UA:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			Notes:      "ChromeDriver: $cdc_ variable, navigator.webdriver=true",
		},
		// Selenium GeckoDriver
		{
			Name: "selenium_geckodriver", Tool: "selenium",
			Ciphers:    []uint16{0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018, 0x0100},
			UA:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
		},
		// cURL / libcurl
		{
			Name: "curl", Tool: "curl",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc00a, 0xc009, 0xc013, 0xc014},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			JA3:        "456523fc94726331a4d5a2e1d40b2cd7",
			UA:         "curl/8.4.0",
			Notes:      "cURL: no HTTP/2 priorities, no cookie jar, no referer",
		},
		// Go net/http default
		{
			Name: "go_http_client", Tool: "go_net_http",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
			Extensions: []uint16{0x0000, 0x000a, 0x000b, 0x000d, 0x002b, 0x002d, 0x0033, 0xff01},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			JA3:        "3cde22e692061f1f4fbc49407c839893",
			UA:         "Go-http-client/2.0",
			Notes:      "Go net/http: unique extension order, missing 0x0023 (session tickets)",
		},
		// Python requests/urllib3
		{
			Name: "python_requests", Tool: "python_requests",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			JA3:        "6f2227e8c16baa0b4e40ab2e0ab87b2c",
			UA:         "python-requests/2.31.0",
			Notes:      "Python urllib3/OpenSSL: deterministic cipher order",
		},
		// Node.js axios/fetch
		{
			Name: "nodejs_http", Tool: "nodejs",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc00a, 0xc013},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			JA3:        "2d1eb5817ece335c24904642f6b9a210",
			UA:         "node-fetch/1.0",
		},
		// Scrapy
		{
			Name: "scrapy", Tool: "scrapy",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Scrapy/2.11",
		},
		// Cloudflare Workers
		{
			Name: "cloudflare_worker", Tool: "cloudflare",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Cloudflare-Worker",
		},
		// Anti-detect browsers (generic patterns)
		{
			Name: "multilogin", Tool: "antidetect",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0010, 0x0005, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			Notes:      "Multilogin: spoofed canvas/WebGL, custom timezone",
		},
		{
			Name: "gologin", Tool: "antidetect",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			Notes:      "GoLogin: Orbita browser with patched Chromium",
		},
		{
			Name: "dolphin_anty", Tool: "antidetect",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033, 0x001b},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			Notes:      "Dolphin Anty: Chromium-based anti-detect",
		},
	}

	fps := make([]CollectedFingerprint, 0, len(bots))
	for _, b := range bots {
		fp := CollectedFingerprint{
			Source:      "bot_knowledge",
			CollectedAt: time.Now().UTC().Format(time.RFC3339),
			BrowserName: b.Name,
			UserAgent:   b.UA,
			JA3Hash:     b.JA3,
			TLS: &TLSFingerprint{
				Version:         0x0303,
				CipherSuites:    b.Ciphers,
				Extensions:      b.Extensions,
				SupportedGroups: b.Curves,
				SNI:             true,
			},
			Metadata: map[string]any{
				"tool":  b.Tool,
				"notes": b.Notes,
			},
		}
		fps = append(fps, fp)
	}
	return fps
}

// collectInternationalBrowsers returns fingerprints for browsers popular outside Western markets.
func collectInternationalBrowsers() []CollectedFingerprint {
	type intlEntry struct {
		Name       string
		Browser    string
		Version    string
		OS         string
		Region     string
		Ciphers    []uint16
		Extensions []uint16
		Curves     []uint16
		UA         string
		H2Window   uint32
		H2Streams  uint32
	}

	// International browsers with distinct fingerprints
	browsers := []intlEntry{
		// Yandex Browser (Russia) — Chromium-based with unique extensions
		{
			Name: "yandex_24_7", Browser: "yandex", Version: "24.7", OS: "Windows", Region: "Russia/CIS",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0010, 0x000d, 0x002b, 0x002d, 0x0033, 0x0015},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 YaBrowser/24.7.0.0 Safari/537.36",
			H2Window:   6291456, H2Streams: 1000,
		},
		{
			Name: "yandex_24_7_mac", Browser: "yandex", Version: "24.7", OS: "macOS", Region: "Russia/CIS",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0010, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 YaBrowser/24.7.0.0 Safari/537.36",
			H2Window:   6291456, H2Streams: 1000,
		},
		// UC Browser (China/India/SEA)
		{
			Name: "uc_browser_16_android", Browser: "uc_browser", Version: "16.5", OS: "Android", Region: "China/India/SEA",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Linux; U; Android 14; en-US; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/100.0.4896.127 UCBrowser/16.5.0.1 Mobile Safari/537.36",
			H2Window:   4194304, H2Streams: 100,
		},
		{
			Name: "uc_browser_16_ios", Browser: "uc_browser", Version: "16.3", OS: "iOS", Region: "China/India/SEA",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02c, 0xc02b, 0xc030, 0xc02f, 0xcca9, 0xcca8},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X; en-US) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1 UCBrowser/16.3.0.0",
			H2Window:   2097152, H2Streams: 100,
		},
		// QQ Browser (China)
		{
			Name: "qq_browser_14", Browser: "qq_browser", Version: "14.0", OS: "Windows", Region: "China",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0010, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36 QQBrowser/14.0.0.0",
			H2Window:   6291456, H2Streams: 1000,
		},
		{
			Name: "qq_browser_14_android", Browser: "qq_browser", Version: "14.5", OS: "Android", Region: "China",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Linux; Android 14; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/122.0.0.0 Mobile MQQBrowser/14.5 QQ/9.0.0 Safari/537.36",
			H2Window:   4194304, H2Streams: 100,
		},
		// Baidu Browser (China)
		{
			Name: "baidu_browser_7", Browser: "baidu", Version: "7.65", OS: "Android", Region: "China",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Linux; Android 14; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Mobile Safari/537.36 baiduboxapp/7.65.0.10",
			H2Window:   4194304, H2Streams: 100,
		},
		// Vivaldi (Europe/Global)
		{
			Name: "vivaldi_6_8", Browser: "vivaldi", Version: "6.8", OS: "Windows", Region: "Europe",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0010, 0x000d, 0x002b, 0x002d, 0x0033, 0x0015},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Vivaldi/6.8.3381.44",
			H2Window:   6291456, H2Streams: 1000,
		},
		{
			Name: "vivaldi_6_8_mac", Browser: "vivaldi", Version: "6.8", OS: "macOS", Region: "Europe",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0010, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Vivaldi/6.8.3381.44",
			H2Window:   6291456, H2Streams: 1000,
		},
		{
			Name: "vivaldi_6_8_linux", Browser: "vivaldi", Version: "6.8", OS: "Linux", Region: "Europe",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0010, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Vivaldi/6.8.3381.44",
			H2Window:   6291456, H2Streams: 1000,
		},
		// Naver Whale (South Korea)
		{
			Name: "whale_3_28", Browser: "whale", Version: "3.28", OS: "Windows", Region: "South_Korea",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0010, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Whale/3.28.266.11 Safari/537.36",
			H2Window:   6291456, H2Streams: 1000,
		},
		// Arc Browser (Global)
		{
			Name: "arc_1_60", Browser: "arc", Version: "1.60", OS: "macOS", Region: "Global",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0010, 0x000d, 0x002b, 0x002d, 0x0033, 0x0015},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
			H2Window:   6291456, H2Streams: 1000,
		},
		// Sogou Browser (China)
		{
			Name: "sogou_12", Browser: "sogou", Version: "12.4", OS: "Windows", Region: "China",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 SogouMobileBrowser/12.4.0",
			H2Window:   6291456, H2Streams: 1000,
		},
		// 360 Safe Browser (China)
		{
			Name: "360_safe_15", Browser: "360_safe", Version: "15.3", OS: "Windows", Region: "China",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014, 0x002f, 0x0035},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36 360SE/15.3.1055.64",
			H2Window:   6291456, H2Streams: 1000,
		},
		// DuckDuckGo Browser (Global privacy)
		{
			Name: "duckduckgo_0_70", Browser: "duckduckgo", Version: "0.70", OS: "macOS", Region: "Global",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02c, 0xc02b, 0xc030, 0xc02f, 0xcca9, 0xcca8},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 DuckDuckGo/7 Safari/605.1.15",
			H2Window:   2097152, H2Streams: 100,
		},
		// Tor Browser (Global privacy)
		{
			Name: "tor_13", Browser: "tor", Version: "13.5", OS: "Linux", Region: "Global",
			Ciphers:    []uint16{0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018, 0x0100},
			UA:         "Mozilla/5.0 (Windows NT 10.0; rv:109.0) Gecko/20100101 Firefox/115.0",
			H2Window:   131072, H2Streams: 100,
		},
		// Mi Browser (Xiaomi, China/India)
		{
			Name: "mi_browser_18", Browser: "mi_browser", Version: "18.5", OS: "Android", Region: "China/India",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Linux; Android 14; 23113RKC6C) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36 MiuiBrowser/18.5.30110",
			H2Window:   4194304, H2Streams: 100,
		},
		// Huawei Browser (China/Global)
		{
			Name: "huawei_browser_15", Browser: "huawei_browser", Version: "15.0", OS: "Android", Region: "China",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Linux; Android 14; NOH-AN00) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.88 HuaweiBrowser/15.0.3.300 Mobile Safari/537.36",
			H2Window:   4194304, H2Streams: 100,
		},
		// OPPO Browser (China/SEA)
		{
			Name: "oppo_browser_8", Browser: "oppo_browser", Version: "8.10", OS: "Android", Region: "China/SEA",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc013, 0xc014},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Linux; Android 14; CPH2569) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36 OppoBrowser/8.10.0.0",
			H2Window:   4194304, H2Streams: 100,
		},
		// Puffin Browser (Global secure)
		{
			Name: "puffin_10", Browser: "puffin", Version: "10.2", OS: "Android", Region: "Global",
			Ciphers:    []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8},
			Extensions: []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x000d, 0x002b, 0x002d, 0x0033},
			Curves:     []uint16{0x001d, 0x0017, 0x0018},
			UA:         "Mozilla/5.0 (Linux; Android 14; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36 Puffin/10.2.0.51618AP",
			H2Window:   4194304, H2Streams: 100,
		},
	}

	fps := make([]CollectedFingerprint, 0, len(browsers))
	for _, b := range browsers {
		fp := CollectedFingerprint{
			Source:      "intl_knowledge",
			CollectedAt: time.Now().UTC().Format(time.RFC3339),
			BrowserName: b.Browser,
			BrowserVer:  b.Version,
			OSName:      b.OS,
			Platform:    "desktop",
			UserAgent:   b.UA,
			TLS: &TLSFingerprint{
				Version:         0x0303,
				CipherSuites:    b.Ciphers,
				Extensions:      b.Extensions,
				SupportedGroups: b.Curves,
				SNI:             true,
			},
			HTTP2: &HTTP2Fingerprint{
				HeaderTableSize:      65536,
				InitialWindowSize:    b.H2Window,
				MaxConcurrentStreams: b.H2Streams,
				MaxFrameSize:         16384,
				PseudoHeaderOrder:    []string{":method", ":authority", ":scheme", ":path"},
				ConnectionFlow:       15663105,
			},
			Metadata: map[string]any{"region": b.Region},
		}
		if b.OS == "Android" || b.OS == "iOS" {
			fp.Platform = "mobile"
		}
		fps = append(fps, fp)
	}
	return fps
}

// generateGoProfiles generates Go code from collected fingerprints.
func generateGoProfiles(result *CollectionResult, outPath string) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "// Code generated by collector. DO NOT EDIT.\n")
	fmt.Fprintf(f, "// Source: %d fingerprints from %v\n", result.TotalCount, result.Sources)
	fmt.Fprintf(f, "// Generated: %s\n\n", result.CollectedAt)
	fmt.Fprintf(f, "package collected\n\n")
	fmt.Fprintf(f, "// CollectedFP represents a collected fingerprint for training.\n")
	fmt.Fprintf(f, "type CollectedFP struct {\n")
	fmt.Fprintf(f, "\tName       string\n")
	fmt.Fprintf(f, "\tBrowser    string\n")
	fmt.Fprintf(f, "\tVersion    string\n")
	fmt.Fprintf(f, "\tOS         string\n")
	fmt.Fprintf(f, "\tCipherSuites []uint16\n")
	fmt.Fprintf(f, "\tExtensions   []uint16\n")
	fmt.Fprintf(f, "\tCurves       []uint16\n")
	fmt.Fprintf(f, "\tUA         string\n")
	fmt.Fprintf(f, "\tJA3Hash    string\n")
	fmt.Fprintf(f, "\tSource     string\n")
	fmt.Fprintf(f, "}\n\n")
	fmt.Fprintf(f, "// AllFingerprints contains all collected fingerprints.\n")
	fmt.Fprintf(f, "var AllFingerprints = []CollectedFP{\n")

	for _, fp := range result.Fingerprints {
		fmt.Fprintf(f, "\t{\n")
		fmt.Fprintf(f, "\t\tName: %q, Browser: %q, Version: %q, OS: %q,\n",
			fp.BrowserName, fp.BrowserName, fp.BrowserVer, fp.OSName)
		if fp.TLS != nil && len(fp.TLS.CipherSuites) > 0 {
			fmt.Fprintf(f, "\t\tCipherSuites: []uint16{")
			for i, c := range fp.TLS.CipherSuites {
				if i > 0 {
					fmt.Fprintf(f, ", ")
				}
				fmt.Fprintf(f, "0x%04x", c)
			}
			fmt.Fprintf(f, "},\n")
		}
		if fp.UserAgent != "" {
			fmt.Fprintf(f, "\t\tUA: %q,\n", fp.UserAgent)
		}
		if fp.JA3Hash != "" {
			fmt.Fprintf(f, "\t\tJA3Hash: %q,\n", fp.JA3Hash)
		}
		fmt.Fprintf(f, "\t\tSource: %q,\n", fp.Source)
		fmt.Fprintf(f, "\t},\n")
	}

	fmt.Fprintf(f, "}\n")
	return nil
}
