// Package profiles - Safari浏览器指纹
// 包含Safari 16-18版本的完整指纹配置
package profiles

import (
	"github.com/vistone/fingerprint/modules/core"
)

// Safari浏览器指纹 (16-18版本)
var (
	// Safari 16.x 系列 - macOS Ventura
	Safari16_0 = ClientProfile{
		ID: "safari_16_0", Name: "Safari 16.0",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.0",
		OS: core.OSMacOS13, OSVersion: "13.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
			0xc024, 0xc028, 0xc02b, 0xc02f,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x0015}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveP256, core.CurveP384, core.CurveP521},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 100,
			InitialWindowSize: 2097152, MaxFrameSize: 16384,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Safari16_1 = ClientProfile{
		ID: "safari_16_1", Name: "Safari 16.1",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.1",
		OS: core.OSMacOS13, OSVersion: "13.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Safari16_2 = ClientProfile{
		ID: "safari_16_2", Name: "Safari 16.2",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.2",
		OS: core.OSMacOS13, OSVersion: "13.1",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Safari16_3 = ClientProfile{
		ID: "safari_16_3", Name: "Safari 16.3",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.3",
		OS: core.OSMacOS13, OSVersion: "13.2",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Safari16_4 = ClientProfile{
		ID: "safari_16_4", Name: "Safari 16.4",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.4",
		OS: core.OSMacOS13, OSVersion: "13.3",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Safari16_5 = ClientProfile{
		ID: "safari_16_5", Name: "Safari 16.5",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.5",
		OS: core.OSMacOS13, OSVersion: "13.4",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Safari16_6 = ClientProfile{
		ID: "safari_16_6", Name: "Safari 16.6",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.6",
		OS: core.OSMacOS13, OSVersion: "13.5",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	// Safari 17.x 系列 - macOS Sonoma
	Safari17_0 = ClientProfile{
		ID: "safari_17_0", Name: "Safari 17.0",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.0",
		OS: core.OSMacOS14, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Safari17_1 = ClientProfile{
		ID: "safari_17_1", Name: "Safari 17.1",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.1",
		OS: core.OSMacOS14, OSVersion: "14.1",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Safari17_2 = ClientProfile{
		ID: "safari_17_2", Name: "Safari 17.2",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.2",
		OS: core.OSMacOS14, OSVersion: "14.2",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Safari17_3 = ClientProfile{
		ID: "safari_17_3", Name: "Safari 17.3",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.3",
		OS: core.OSMacOS14, OSVersion: "14.3",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Safari17_4 = ClientProfile{
		ID: "safari_17_4", Name: "Safari 17.4",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.4",
		OS: core.OSMacOS14, OSVersion: "14.4",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Safari17_5 = ClientProfile{
		ID: "safari_17_5", Name: "Safari 17.5",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.5",
		OS: core.OSMacOS14, OSVersion: "14.5",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Safari17_6 = ClientProfile{
		ID: "safari_17_6", Name: "Safari 17.6",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.6",
		OS: core.OSMacOS14, OSVersion: "14.6",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	// Safari 18.x 系列 - macOS Sequoia
	Safari18_0 = ClientProfile{
		ID: "safari_18_0", Name: "Safari 18.0",
		BrowserType: core.BrowserSafari, BrowserVersion: "18.0",
		OS: core.OSMacOS15, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Safari18_1 = ClientProfile{
		ID: "safari_18_1", Name: "Safari 18.1",
		BrowserType: core.BrowserSafari, BrowserVersion: "18.1",
		OS: core.OSMacOS15, OSVersion: "15.1",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}

	Safari18_2 = ClientProfile{
		ID: "safari_18_2", Name: "Safari 18.2",
		BrowserType: core.BrowserSafari, BrowserVersion: "18.2",
		OS: core.OSMacOS15, OSVersion: "15.2",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
	}
)

func init() {
	// 注册所有Safari指纹
	profiles := []ClientProfile{
		Safari16_0, Safari16_1, Safari16_2, Safari16_3, Safari16_4, Safari16_5, Safari16_6,
		Safari17_0, Safari17_1, Safari17_2, Safari17_3, Safari17_4, Safari17_5, Safari17_6,
		Safari18_0, Safari18_1, Safari18_2,
	}
	for _, p := range profiles {
		Register(p)
	}
}

// AllSafariProfiles 返回所有Safari指纹
func AllSafariProfiles() []ClientProfile {
	return []ClientProfile{
		Safari16_0, Safari16_1, Safari16_2, Safari16_3, Safari16_4, Safari16_5, Safari16_6,
		Safari17_0, Safari17_1, Safari17_2, Safari17_3, Safari17_4, Safari17_5, Safari17_6,
		Safari18_0, Safari18_1, Safari18_2,
	}
}
