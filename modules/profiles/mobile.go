// Package profiles - 移动设备浏览器指纹
// 包含iOS Safari、Android Chrome、Firefox Mobile等移动浏览器指纹
package profiles

import (
	"github.com/vistone/fingerprint/modules/core"
)

// iOS Safari指纹

var (
	IOSSafari16 = ClientProfile{
		ID: "ios_safari_16", Name: "iOS Safari 16",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.0",
		OS: core.OSiOS, OSVersion: "16.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
			0xc024, 0xc028, 0xc02b, 0xc02f,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
		},
		SupportedCurves: []core.CurveID{core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 100,
			InitialWindowSize: 2097152, MaxFrameSize: 16384,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
			SecFetchDest: "document", SecFetchMode: "navigate",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	IOSSafari17 = ClientProfile{
		ID: "ios_safari_17", Name: "iOS Safari 17",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.0",
		OS: core.OSiOS, OSVersion: "17.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	IOSSafari18 = ClientProfile{
		ID: "ios_safari_18", Name: "iOS Safari 18",
		BrowserType: core.BrowserSafari, BrowserVersion: "18.0",
		OS: core.OSiOS, OSVersion: "18.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	// iPad Safari
	IPadSafari16 = ClientProfile{
		ID: "ipad_safari_16", Name: "iPad Safari 16",
		BrowserType: core.BrowserSafari, BrowserVersion: "16.0",
		OS: core.OSiPadOS, OSVersion: "16.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	IPadSafari17 = ClientProfile{
		ID: "ipad_safari_17", Name: "iPad Safari 17",
		BrowserType: core.BrowserSafari, BrowserVersion: "17.0",
		OS: core.OSiPadOS, OSVersion: "17.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	IPadSafari18 = ClientProfile{
		ID: "ipad_safari_18", Name: "iPad Safari 18",
		BrowserType: core.BrowserSafari, BrowserVersion: "18.0",
		OS: core.OSiPadOS, OSVersion: "18.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02c, 0xc030, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}
)

// Android Chrome指纹

var (
	AndroidChrome115 = ClientProfile{
		ID: "android_chrome_115", Name: "Android Chrome 115",
		BrowserType: core.BrowserChrome, BrowserVersion: "115.0",
		OS: core.OSAndroid, OSVersion: "13.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
			0xc013, 0xc014, 0x002f, 0x0035,
		},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x0015}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings: core.HTTP2Settings{
			HeaderTableSize: 65536, EnablePush: 0, MaxConcurrentStreams: 1000,
			InitialWindowSize: 6291456, MaxFrameSize: 16384,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	AndroidChrome120 = ClientProfile{
		ID: "android_chrome_120", Name: "Android Chrome 120",
		BrowserType: core.BrowserChrome, BrowserVersion: "120.0",
		OS: core.OSAndroid, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	AndroidChrome125 = ClientProfile{
		ID: "android_chrome_125", Name: "Android Chrome 125",
		BrowserType: core.BrowserChrome, BrowserVersion: "125.0",
		OS: core.OSAndroid, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	AndroidChrome130 = ClientProfile{
		ID: "android_chrome_130", Name: "Android Chrome 130",
		BrowserType: core.BrowserChrome, BrowserVersion: "130.0",
		OS: core.OSAndroid, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	AndroidChrome131 = ClientProfile{
		ID: "android_chrome_131", Name: "Android Chrome 131",
		BrowserType: core.BrowserChrome, BrowserVersion: "131.0",
		OS: core.OSAndroid, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}
)

// Android Firefox指纹

var (
	AndroidFirefox115 = ClientProfile{
		ID: "android_firefox_115", Name: "Android Firefox 115",
		BrowserType: core.BrowserFirefox, BrowserVersion: "115.0",
		OS: core.OSAndroid, OSVersion: "13.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	AndroidFirefox120 = ClientProfile{
		ID: "android_firefox_120", Name: "Android Firefox 120",
		BrowserType: core.BrowserFirefox, BrowserVersion: "120.0",
		OS: core.OSAndroid, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	AndroidFirefox125 = ClientProfile{
		ID: "android_firefox_125", Name: "Android Firefox 125",
		BrowserType: core.BrowserFirefox, BrowserVersion: "125.0",
		OS: core.OSAndroid, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	AndroidFirefox130 = ClientProfile{
		ID: "android_firefox_130", Name: "Android Firefox 130",
		BrowserType: core.BrowserFirefox, BrowserVersion: "130.0",
		OS: core.OSAndroid, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.5", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}
)

// Samsung Internet指纹

var (
	SamsungInternet22 = ClientProfile{
		ID: "samsung_internet_22", Name: "Samsung Internet 22",
		BrowserType: core.BrowserSamsung, BrowserVersion: "22.0",
		OS: core.OSAndroid, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	SamsungInternet23 = ClientProfile{
		ID: "samsung_internet_23", Name: "Samsung Internet 23",
		BrowserType: core.BrowserSamsung, BrowserVersion: "23.0",
		OS: core.OSAndroid, OSVersion: "14.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}

	SamsungInternet24 = ClientProfile{
		ID: "samsung_internet_24", Name: "Samsung Internet 24",
		BrowserType: core.BrowserSamsung, BrowserVersion: "24.0",
		OS: core.OSAndroid, OSVersion: "15.0",
		TLSVersion: 0x0303,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9, 0xcca8,
		},
		Headers: &core.HTTPHeaders{
			Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			AcceptLanguage: "en-US,en;q=0.9", AcceptEncoding: "gzip, deflate, br",
		},
		TCPIP: CreateTCPIP(core.OSAndroid),
	}
)

func init() {
	// 注册所有移动设备指纹
	profiles := []ClientProfile{
		// iOS Safari
		IOSSafari16, IOSSafari17, IOSSafari18,
		// iPad Safari
		IPadSafari16, IPadSafari17, IPadSafari18,
		// Android Chrome
		AndroidChrome115, AndroidChrome120, AndroidChrome125, AndroidChrome130, AndroidChrome131,
		// Android Firefox
		AndroidFirefox115, AndroidFirefox120, AndroidFirefox125, AndroidFirefox130,
		// Samsung Internet
		SamsungInternet22, SamsungInternet23, SamsungInternet24,
	}
	for _, p := range profiles {
		Register(p)
	}
}

// AllMobileProfiles 返回所有移动设备指纹
func AllMobileProfiles() []ClientProfile {
	return []ClientProfile{
		// iOS Safari
		IOSSafari16, IOSSafari17, IOSSafari18,
		// iPad Safari
		IPadSafari16, IPadSafari17, IPadSafari18,
		// Android Chrome
		AndroidChrome115, AndroidChrome120, AndroidChrome125, AndroidChrome130, AndroidChrome131,
		// Android Firefox
		AndroidFirefox115, AndroidFirefox120, AndroidFirefox125, AndroidFirefox130,
		// Samsung Internet
		SamsungInternet22, SamsungInternet23, SamsungInternet24,
	}
}
