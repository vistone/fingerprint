// Package profiles - Firefox浏览器指纹
// 包含Firefox 115-135版本的完整指纹配置，含ESR系列
package profiles

import (
	"github.com/vistone/fingerprint/modules/core"
)

// Firefox浏览器指纹 (115-135版本)
var (
	// Firefox 115-119 ESR系列
	Firefox115 = ClientProfile{
		ID: "firefox_115", Name: "Firefox 115 ESR",
		BrowserType: core.BrowserFirefox, BrowserVersion: "115.0",
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
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
			UpgradeInsecureRequests: "1",
		},
	}

	Firefox116 = ClientProfile{
		ID: "firefox_116", Name: "Firefox 116",
		BrowserType: core.BrowserFirefox, BrowserVersion: "116.0",
		OS: core.OSWindows10, OSVersion: "10.0.19045",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Firefox117 = ClientProfile{
		ID: "firefox_117", Name: "Firefox 117",
		BrowserType: core.BrowserFirefox, BrowserVersion: "117.0",
		OS: core.OSMacOS13, OSVersion: "13.5",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Firefox118 = ClientProfile{
		ID: "firefox_118", Name: "Firefox 118",
		BrowserType: core.BrowserFirefox, BrowserVersion: "118.0",
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

	Firefox119 = ClientProfile{
		ID: "firefox_119", Name: "Firefox 119",
		BrowserType: core.BrowserFirefox, BrowserVersion: "119.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	// Firefox 121-124
	Firefox121 = ClientProfile{
		ID: "firefox_121", Name: "Firefox 121",
		BrowserType: core.BrowserFirefox, BrowserVersion: "121.0",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Firefox122 = ClientProfile{
		ID: "firefox_122", Name: "Firefox 122",
		BrowserType: core.BrowserFirefox, BrowserVersion: "122.0",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Firefox123 = ClientProfile{
		ID: "firefox_123", Name: "Firefox 123",
		BrowserType: core.BrowserFirefox, BrowserVersion: "123.0",
		OS: core.OSLinuxUbuntu, OSVersion: "22.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Firefox124 = ClientProfile{
		ID: "firefox_124", Name: "Firefox 124",
		BrowserType: core.BrowserFirefox, BrowserVersion: "124.0",
		OS: core.OSLinuxUbuntu, OSVersion: "24.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	// Firefox 126-129
	Firefox126 = ClientProfile{
		ID: "firefox_126", Name: "Firefox 126",
		BrowserType: core.BrowserFirefox, BrowserVersion: "126.0",
		OS: core.OSMacOS14, OSVersion: "14.5",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Firefox127 = ClientProfile{
		ID: "firefox_127", Name: "Firefox 127",
		BrowserType: core.BrowserFirefox, BrowserVersion: "127.0",
		OS: core.OSMacOS14, OSVersion: "14.6",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Firefox128 = ClientProfile{
		ID: "firefox_128", Name: "Firefox 128 ESR",
		BrowserType: core.BrowserFirefox, BrowserVersion: "128.0",
		OS: core.OSWindows10, OSVersion: "10.0.19045",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Firefox129 = ClientProfile{
		ID: "firefox_129", Name: "Firefox 129",
		BrowserType: core.BrowserFirefox, BrowserVersion: "129.0",
		OS: core.OSLinuxUbuntu, OSVersion: "24.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	// Firefox 131, 134-135
	Firefox131 = ClientProfile{
		ID: "firefox_131", Name: "Firefox 131",
		BrowserType: core.BrowserFirefox, BrowserVersion: "131.0",
		OS: core.OSMacOS15, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Firefox134 = ClientProfile{
		ID: "firefox_134", Name: "Firefox 134",
		BrowserType: core.BrowserFirefox, BrowserVersion: "134.0",
		OS: core.OSMacOS15, OSVersion: "15.1",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Firefox135 = ClientProfile{
		ID: "firefox_135", Name: "Firefox 135",
		BrowserType: core.BrowserFirefox, BrowserVersion: "135.0",
		OS: core.OSWindows11, OSVersion: "10.0.26100",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Firefox136 = ClientProfile{
		ID: "firefox_136", Name: "Firefox 136",
		BrowserType: core.BrowserFirefox, BrowserVersion: "136.0",
		OS: core.OSMacOS15, OSVersion: "15.2",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Firefox137 = ClientProfile{
		ID: "firefox_137", Name: "Firefox 137",
		BrowserType: core.BrowserFirefox, BrowserVersion: "137.0",
		OS: core.OSLinuxUbuntu, OSVersion: "24.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Firefox138 = ClientProfile{
		ID: "firefox_138", Name: "Firefox 138",
		BrowserType: core.BrowserFirefox, BrowserVersion: "138.0",
		OS: core.OSWindows11, OSVersion: "10.0.26100",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Firefox139 = ClientProfile{
		ID: "firefox_139", Name: "Firefox 139",
		BrowserType: core.BrowserFirefox, BrowserVersion: "139.0",
		OS: core.OSMacOS15, OSVersion: "15.3",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Firefox140 = ClientProfile{
		ID: "firefox_140", Name: "Firefox 140",
		BrowserType: core.BrowserFirefox, BrowserVersion: "140.0",
		OS: core.OSLinuxUbuntu, OSVersion: "25.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
	}
)

func init() {
	// 注册所有Firefox指纹
	profiles := []ClientProfile{
		Firefox115, Firefox116, Firefox117, Firefox118, Firefox119,
		Firefox121, Firefox122, Firefox123, Firefox124,
		Firefox126, Firefox127, Firefox128, Firefox129,
		Firefox131, Firefox134, Firefox135, Firefox136, Firefox137, Firefox138, Firefox139, Firefox140,
	}
	for _, p := range profiles {
		Register(p)
	}
}

// AllFirefoxProfiles 返回所有Firefox指纹
func AllFirefoxProfiles() []ClientProfile {
	return []ClientProfile{
		Firefox115, Firefox116, Firefox117, Firefox118, Firefox119,
		Firefox121, Firefox122, Firefox123, Firefox124,
		Firefox126, Firefox127, Firefox128, Firefox129,
		Firefox131, Firefox134, Firefox135,
	}
}
