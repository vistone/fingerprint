// Package profiles - Brave浏览器指纹
// 包含Brave 1.6x-1.7x版本的完整指纹配置
package profiles

import (
	"github.com/vistone/fingerprint/modules/core"
)

// Brave浏览器指纹 (1.6x-1.7x版本)
var (
	Brave1_60 = ClientProfile{
		ID: "brave_1_60", Name: "Brave 1.60",
		BrowserType: core.BrowserBrave, BrowserVersion: "1.60.0",
		OS: core.OSWindows11, OSVersion: "10.0.22631",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc013, 0xc014, 0x002f, 0x0035,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x0015}, {Type: 0x0033}, {Type: 0x002d},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 1000,
			InitialWindowSize: 6291456, MaxFrameSize: 16384,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Brave1_62 = ClientProfile{
		ID: "brave_1_62", Name: "Brave 1.62",
		BrowserType: core.BrowserBrave, BrowserVersion: "1.62.0",
		OS: core.OSMacOS14, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Brave1_64 = ClientProfile{
		ID: "brave_1_64", Name: "Brave 1.64",
		BrowserType: core.BrowserBrave, BrowserVersion: "1.64.0",
		OS: core.OSWindows11, OSVersion: "10.0.26100",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Brave1_66 = ClientProfile{
		ID: "brave_1_66", Name: "Brave 1.66",
		BrowserType: core.BrowserBrave, BrowserVersion: "1.66.0",
		OS: core.OSLinuxUbuntu, OSVersion: "24.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Brave1_68 = ClientProfile{
		ID: "brave_1_68", Name: "Brave 1.68",
		BrowserType: core.BrowserBrave, BrowserVersion: "1.68.0",
		OS: core.OSMacOS15, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Brave1_70 = ClientProfile{
		ID: "brave_1_70", Name: "Brave 1.70",
		BrowserType: core.BrowserBrave, BrowserVersion: "1.70.0",
		OS: core.OSWindows11, OSVersion: "10.0.26100",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Brave1_72 = ClientProfile{
		ID: "brave_1_72", Name: "Brave 1.72",
		BrowserType: core.BrowserBrave, BrowserVersion: "1.72.0",
		OS: core.OSLinuxUbuntu, OSVersion: "24.04",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}
)

func init() {
	// 注册所有Brave指纹
	profiles := []ClientProfile{
		Brave1_60, Brave1_62, Brave1_64, Brave1_66, Brave1_68, Brave1_70, Brave1_72,
	}
	for _, p := range profiles {
		Register(p)
	}
}

// AllBraveProfiles 返回所有Brave指纹
func AllBraveProfiles() []ClientProfile {
	return []ClientProfile{
		Brave1_60, Brave1_62, Brave1_64, Brave1_66, Brave1_68, Brave1_70, Brave1_72,
	}
}
