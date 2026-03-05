// Package profiles 提供全面的浏览器指纹配置
// 目标: 90+ 真实浏览器指纹
package profiles

import (
	"strings"

	"github.com/vistone/fingerprint/modules/core"
)

// Chrome 全面配置 (30 profiles)
func initChromeProfiles() {
	chromeVersions := []struct {
		id      string
		version string
		os      core.OperatingSystem
		osVer   string
	}{
		{"chrome_103", "103.0.5060.134", core.OSWindows10, "10.0.19044"},
		{"chrome_104", "104.0.5112.102", core.OSWindows10, "10.0.19044"},
		{"chrome_105", "105.0.5195.127", core.OSWindows10, "10.0.19044"},
		{"chrome_106", "106.0.5249.119", core.OSWindows10, "10.0.19044"},
		{"chrome_107", "107.0.5304.107", core.OSWindows11, "10.0.22000"},
		{"chrome_108", "108.0.5359.125", core.OSWindows11, "10.0.22000"},
		{"chrome_109", "109.0.5414.120", core.OSWindows11, "10.0.22000"},
		{"chrome_110", "110.0.5481.178", core.OSWindows11, "10.0.22621"},
		{"chrome_111", "111.0.5563.147", core.OSWindows11, "10.0.22621"},
		{"chrome_112", "112.0.5615.138", core.OSWindows11, "10.0.22621"},
		{"chrome_113", "113.0.5672.127", core.OSWindows11, "10.0.22621"},
		{"chrome_114", "114.0.5735.199", core.OSWindows11, "10.0.22621"},
		{"chrome_115", "115.0.5790.171", core.OSWindows11, "10.0.22621"},
		{"chrome_116", "116.0.5845.188", core.OSWindows11, "10.0.22621"},
		{"chrome_117", "117.0.5938.149", core.OSWindows11, "10.0.22631"},
		{"chrome_118", "118.0.5993.117", core.OSWindows11, "10.0.22631"},
		{"chrome_119", "119.0.6045.200", core.OSWindows11, "10.0.22631"},
		{"chrome_121", "121.0.6167.185", core.OSMacOS14, "14.0"},
		{"chrome_122", "122.0.6261.128", core.OSMacOS14, "14.0"},
		{"chrome_125", "125.0.6422.142", core.OSMacOS15, "15.0"},
		{"chrome_126", "126.0.6478.126", core.OSMacOS15, "15.0"},
		{"chrome_127", "127.0.6533.120", core.OSLinuxUbuntu, "22.04"},
		{"chrome_128", "128.0.6613.138", core.OSLinuxUbuntu, "22.04"},
		{"chrome_129", "129.0.6668.101", core.OSLinuxUbuntu, "22.04"},
		// 已存在于其他文件中的版本
		// 120, 124, 130, 131, 132, 133
	}

	for _, v := range chromeVersions {
		p := ClientProfile{
			ID:             v.id,
			Name:           "Chrome " + getMajorVersion(v.version),
			BrowserType:    core.BrowserChrome,
			BrowserVersion: v.version,
			OS:             v.os,
			OSVersion:      v.osVer,
			OSArch:         "x86_64",
			OSBitness:      "64",
			TLSVersion:     0x0303,
			CipherSuites:   []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xcca9},
			Extensions: []core.TLSExtension{
				{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
				{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
				{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
				{Type: 0x002d}, {Type: 0x0033},
			},
			SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256},
			HTTP2Settings: core.HTTP2Settings{
				HeaderTableSize: 65536, EnablePush: 0,
				MaxConcurrentStreams: 1000, InitialWindowSize: 6291456,
			},
			Headers: &core.HTTPHeaders{
				Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				AcceptLanguage:  "en-US,en;q=0.9",
				AcceptEncoding:  "gzip, deflate, br",
				SecFetchSite:    "none",
				SecFetchMode:    "navigate",
				SecFetchDest:    "document",
				SecCHUA:         `"Chromium";v="` + safeSliceVersion(v.version) + `", "Google Chrome";v="` + safeSliceVersion(v.version) + `"`,
				SecCHUAMobile:   "?0",
				SecCHUAPlatform: platformString(v.os),
			},
		}
		Register(p)
	}
}

// Firefox 全面配置 (25 profiles)
func initFirefoxProfiles() {
	firefoxVersions := []struct {
		id      string
		version string
		os      core.OperatingSystem
	}{
		{"firefox_102", "102.0", core.OSWindows10},
		{"firefox_103", "103.0", core.OSWindows10},
		{"firefox_104", "104.0", core.OSWindows10},
		{"firefox_105", "105.0", core.OSWindows10},
		{"firefox_106", "106.0", core.OSWindows11},
		{"firefox_107", "107.0", core.OSWindows11},
		{"firefox_108", "108.0", core.OSWindows11},
		{"firefox_109", "109.0", core.OSWindows11},
		{"firefox_110", "110.0", core.OSWindows11},
		{"firefox_111", "111.0", core.OSWindows11},
		{"firefox_112", "112.0", core.OSWindows11},
		{"firefox_113", "113.0", core.OSWindows11},
		{"firefox_114", "114.0", core.OSMacOS13},
		{"firefox_115", "115.0", core.OSMacOS13},
		{"firefox_116", "116.0", core.OSMacOS13},
		{"firefox_117", "117.0", core.OSMacOS14},
		{"firefox_118", "118.0", core.OSMacOS14},
		{"firefox_119", "119.0", core.OSMacOS14},
		{"firefox_121", "121.0", core.OSMacOS15},
		{"firefox_122", "122.0", core.OSLinux},
		{"firefox_123", "123.0", core.OSLinux},
		{"firefox_124", "124.0", core.OSLinux},
		{"firefox_126", "126.0", core.OSLinux},
		{"firefox_127", "127.0", core.OSLinux},
		{"firefox_128", "128.0", core.OSLinux},
		// 120, 125, 130, 132, 133 已在其他文件中
	}

	for _, v := range firefoxVersions {
		p := ClientProfile{
			ID:             v.id,
			Name:           "Firefox " + getMajorVersion(v.version),
			BrowserType:    core.BrowserFirefox,
			BrowserVersion: v.version,
			OS:             v.os,
			OSVersion:      "10.0",
			TLSVersion:     0x0303,
			CipherSuites: []uint16{
				0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f,
				0xcca9, 0xcca8, 0xc00a, 0xc014,
			},
			HTTP2Settings: core.HTTP2Settings{
				HeaderTableSize: 65536, EnablePush: 0,
				MaxConcurrentStreams: 100, InitialWindowSize: 131072,
			},
			PseudoHeaderOrder: []string{":method", ":path", ":authority", ":scheme"},
			Headers: &core.HTTPHeaders{
				Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				AcceptLanguage:  "en-US,en;q=0.5",
				AcceptEncoding:  "gzip, deflate, br",
				UpgradeInsecureRequests: "1",
			},
		}
		Register(p)
	}
}

// Safari 全面配置 (20 profiles)
func initSafariProfiles() {
	safariVersions := []struct {
		id      string
		version string
		os      core.OperatingSystem
		osVer   string
	}{
		{"safari_14_0", "14.0", core.OSMacOS13, "13.0"},
		{"safari_14_1", "14.1", core.OSMacOS13, "13.1"},
		{"safari_15_0", "15.0", core.OSMacOS13, "13.0"},
		{"safari_15_1", "15.1", core.OSMacOS13, "13.1"},
		{"safari_15_3", "15.3", core.OSMacOS13, "13.2"},
		{"safari_15_4", "15.4", core.OSMacOS13, "13.3"},
		{"safari_15_5", "15.5", core.OSMacOS14, "14.0"},
		{"safari_15_6", "15.6", core.OSMacOS14, "14.0"},
		{"safari_16_1", "16.1", core.OSMacOS14, "14.0"},
		{"safari_16_2", "16.2", core.OSMacOS14, "14.1"},
		{"safari_16_3", "16.3", core.OSMacOS14, "14.2"},
		{"safari_16_4", "16.4", core.OSMacOS14, "14.3"},
		{"safari_16_5", "16.5", core.OSMacOS14, "14.4"},
		{"safari_16_6", "16.6", core.OSMacOS14, "14.5"},
		{"safari_17_1", "17.1", core.OSMacOS15, "15.0"},
		{"safari_17_2", "17.2", core.OSMacOS15, "15.1"},
		{"safari_17_3", "17.3", core.OSMacOS15, "15.2"},
		{"safari_17_4", "17.4", core.OSMacOS15, "15.3"},
		{"safari_17_5", "17.5", core.OSMacOS15, "15.4"},
		{"safari_17_6", "17.6", core.OSMacOS15, "15.5"},
		// 15.0, 16.0, 17.0, 18.0, 18.1 已在其他文件中
	}

	for _, v := range safariVersions {
		p := ClientProfile{
			ID:             v.id,
			Name:           "Safari " + v.version,
			BrowserType:    core.BrowserSafari,
			BrowserVersion: v.version,
			OS:             v.os,
			OSVersion:      v.osVer,
			TLSVersion:     0x0303,
			CipherSuites:   []uint16{0x1301, 0x1302, 0x1303, 0xc02c, 0xc02b, 0xc030, 0xc02f},
			SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384, core.CurveP521},
			HTTP2Settings: core.HTTP2Settings{
				HeaderTableSize: 4096, EnablePush: 1,
				MaxConcurrentStreams: 100, InitialWindowSize: 2097152,
			},
			PseudoHeaderOrder: []string{":method", ":scheme", ":path", ":authority"},
			Headers: &core.HTTPHeaders{
				Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				AcceptLanguage:  "en-US,en;q=0.9",
				AcceptEncoding:  "gzip, deflate, br",
			},
		}
		Register(p)
	}
}

// iOS Safari (15 profiles)
func initiOSProfiles() {
	iosVersions := []struct {
		id      string
		version string
		osVer   string
	}{
		{"safari_ios_14_0", "14.0", "14.0"},
		{"safari_ios_14_5", "14.5", "14.5"},
		{"safari_ios_14_6", "14.6", "14.6"},
		{"safari_ios_14_7", "14.7", "14.7"},
		{"safari_ios_14_8", "14.8", "14.8"},
		{"safari_ios_15_1", "15.1", "15.1"},
		{"safari_ios_15_2", "15.2", "15.2"},
		{"safari_ios_15_3", "15.3", "15.3"},
		{"safari_ios_15_4", "15.4", "15.4"},
		{"safari_ios_16_1", "16.1", "16.1"},
		{"safari_ios_16_2", "16.2", "16.2"},
		{"safari_ios_16_3", "16.3", "16.3"},
		{"safari_ios_16_4", "16.4", "16.4"},
		{"safari_ios_16_5", "16.5", "16.5"},
		{"safari_ios_16_6", "16.6", "16.6"},
		// 15.5, 15.6, 16.0, 17.0, 18.0, 18.1 已在其他文件中
	}

	for _, v := range iosVersions {
		p := ClientProfile{
			ID:          v.id,
			Name:        "Safari iOS " + v.version,
			BrowserType: core.BrowserSafari,
			BrowserVersion: v.version,
			OS:          core.OSMacOS15, // 使用标准 OS 类型
			OSVersion:   v.osVer,
			Headers: &core.HTTPHeaders{
				Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				AcceptLanguage:  "en-US,en;q=0.9",
				SecCHUA:         "",
				SecCHUAMobile:   "?1",
				SecCHUAPlatform: `"iPhone"`,
			},
		}
		Register(p)
	}
}

// Android Chrome (15 profiles)
func initAndroidProfiles() {
	androidVersions := []struct {
		id      string
		version string
		android string
		device  string
	}{
		{"chrome_android_103", "103.0.5060.71", "13", "SM-G998B"},
		{"chrome_android_105", "105.0.5195.136", "13", "SM-G996B"},
		{"chrome_android_107", "107.0.5304.105", "13", "SM-G991B"},
		{"chrome_android_109", "109.0.5414.117", "13", "SM-S908B"},
		{"chrome_android_111", "111.0.5563.116", "14", "SM-S918B"},
		{"chrome_android_113", "113.0.5672.162", "14", "SM-S928B"},
		{"chrome_android_115", "115.0.5790.166", "14", "Pixel 7"},
		{"chrome_android_117", "117.0.5938.153", "14", "Pixel 7 Pro"},
		{"chrome_android_119", "119.0.6045.193", "14", "Pixel 8"},
		{"chrome_android_121", "121.0.6167.178", "14", "Pixel 8 Pro"},
		{"chrome_android_123", "123.0.6312.80", "14", "SM-G996B"},
		{"chrome_android_125", "125.0.6422.165", "14", "SM-G991B"},
		{"chrome_android_127", "127.0.6533.103", "14", "Pixel 8a"},
		{"chrome_android_129", "129.0.6668.81", "15", "Pixel 9"},
		{"chrome_android_131", "131.0.6778.135", "15", "Pixel 9 Pro"},
		// 120, 130 已在其他文件中
	}

	for _, v := range androidVersions {
		p := ClientProfile{
			ID:             v.id,
			Name:           "Chrome Android " + v.version,
			BrowserType:    core.BrowserChrome,
			BrowserVersion: v.version,
			OS:             core.OSLinux, // 使用标准 OS 类型
			OSVersion:      v.android,
			Headers: &core.HTTPHeaders{
				Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				AcceptLanguage:  "en-US,en;q=0.9",
				SecCHUA:         `"Android";v="` + v.android + `", "Chrome";v="` + safeSliceVersion(v.version) + `"`,
				SecCHUAMobile:   "?1",
				SecCHUAPlatform: `"Android"`,
			},
		}
		Register(p)
	}
}

// Edge 更多版本 (8 profiles)
func initEdgeProfiles() {
	edgeVersions := []struct {
		id      string
		version string
		channel string
	}{
		{"edge_110", "110.0.1587.63", "stable"},
		{"edge_111", "111.0.1661.54", "stable"},
		{"edge_112", "112.0.1722.58", "stable"},
		{"edge_113", "113.0.1774.57", "stable"},
		{"edge_114", "114.0.1823.67", "stable"},
		{"edge_115", "115.0.1901.203", "stable"},
		{"edge_125", "125.0.2535.92", "stable"},
		{"edge_128", "128.0.2739.79", "stable"},
		// 120, 130 已在其他文件中
	}

	for _, v := range edgeVersions {
		p := ClientProfile{
			ID:             v.id,
			Name:           "Edge " + v.version,
			BrowserType:    core.BrowserEdge,
			BrowserVersion: v.version,
			OS:             core.OSWindows11,
			Headers: &core.HTTPHeaders{
				Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				SecCHUA: `"Microsoft Edge";v="` + safeSliceVersion(v.version) + `", "Chromium";v="` + safeSliceVersion(v.version) + `"`,
			},
		}
		Register(p)
	}
}

// Opera 更多版本 (5 profiles)
func initOperaProfiles() {
	operaVersions := []struct {
		id      string
		version string
	}{
		{"opera_95", "95.0.4635.90"},
		{"opera_96", "96.0.4693.80"},
		{"opera_97", "97.0.4719.63"},
		{"opera_98", "98.0.4759.39"},
		{"opera_99", "99.0.4788.77"},
		// 100, 105 已在其他文件中
	}

	for _, v := range operaVersions {
		p := ClientProfile{
			ID:             v.id,
			Name:           "Opera " + v.version,
			BrowserType:    core.BrowserOpera,
			BrowserVersion: v.version,
			OS:             core.OSWindows10,
			Headers: &core.HTTPHeaders{
				Accept:  "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				SecCHUA: `"Opera";v="` + safeSliceVersion(v.version) + `"`,
			},
		}
		Register(p)
	}
}

// helper 函数
func platformString(os core.OperatingSystem) string {
	osStr := string(os)
	if contains(osStr, "Windows") {
		return `"Windows"`
	}
	if contains(osStr, "Mac OS X") || contains(osStr, "Macintosh") {
		return `"macOS"`
	}
	if contains(osStr, "Linux") {
		return `"Linux"`
	}
	return `"Windows"`
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// 初始化所有扩展配置
func init() {
	initChromeProfiles()
	initFirefoxProfiles()
	initSafariProfiles()
	initiOSProfiles()
	initAndroidProfiles()
	initEdgeProfiles()
	initOperaProfiles()
}
