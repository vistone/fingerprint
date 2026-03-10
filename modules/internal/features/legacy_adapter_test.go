package features

import (
	"math"
	"strings"
	"testing"
)

// translated comment
func TestNewLegacyFeatureAdapter(t *testing.T) {
	tests := []struct {
		name   string
		config *FeatureConfig
		want   func(*LegacyFeatureAdapter) bool
	}{
		{
			name:   "使用nil配置（应该使用默认配置）",
			config: nil,
			want: func(adapter *LegacyFeatureAdapter) bool {
				return adapter != nil &&
					adapter.config != nil &&
					adapter.extractor != nil &&
					adapter.config.EntropyHighThreshold == 7.5 &&
					adapter.config.EntropyLowThreshold == 26
			},
		},
		{
			name: "使用自定义配置",
			config: &FeatureConfig{
				EntropyHighThreshold:  8.0,
				EntropyLowThreshold:   30,
				ToolMarkers:           []string{"custom"},
				HeadlessMarkers:       []string{"test"},
				MobileScreenWidthMax:  2000,
				DesktopScreenWidthMin: 900,
			},
			want: func(adapter *LegacyFeatureAdapter) bool {
				return adapter != nil &&
					adapter.config.EntropyHighThreshold == 8.0 &&
					adapter.config.EntropyLowThreshold == 30 &&
					len(adapter.config.ToolMarkers) == 1 &&
					adapter.config.ToolMarkers[0] == "custom"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewLegacyFeatureAdapter(tt.config)
			if !tt.want(adapter) {
				t.Errorf("NewLegacyFeatureAdapter() 创建的适配器不符合预期")
			}
		})
	}
}

// translated comment
func TestDetectAnomalies(t *testing.T) {
	adapter := NewLegacyFeatureAdapter(nil)

	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "正常数据（无异常）",
			data:     []byte("This is a normal text with good entropy and variety of characters! 1234567890 ABCDEFGHIJKLMNOPQRSTUVWXYZ"),
			expected: false,
		},
		{
			name:     "包含工具标记的数据（有异常）",
			data:     []byte("This contains selenium and webdriver markers"),
			expected: true,
		},
		{
			name:     "包含HeadlessChrome标记",
			data:     []byte("User-Agent: HeadlessChrome/90.0"),
			expected: true,
		},
		{
			name:     "低熵数据（重复字符）",
			data:     []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
			expected: true,
		},
		{
			name:     "高熵数据",
			data:     []byte("a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0"),
			expected: false,
		},
		{
			name:     "空数据",
			data:     []byte(""),
			expected: false,
		},
		{
			name:     "短数据",
			data:     []byte("short"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapter.DetectAnomalies(tt.data)
			if got != tt.expected {
				t.Errorf("DetectAnomalies() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// translated comment
func TestHasLowEntropy(t *testing.T) {
	adapter := NewLegacyFeatureAdapter(nil)

	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "低熵数据返回true（重复字符）",
			data:     []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
			expected: true,
		},
		{
			name:     "低熵数据返回true（少量不同字符）",
			data:     []byte("abcabcabcabcabcabcabcabcabcabc"),
			expected: true,
		},
		{
			name:     "正常熵值返回false（丰富字符）",
			data:     []byte("The quick brown fox jumps over the lazy dog! 1234567890"),
			expected: false,
		},
		{
			name:     "正常熵值返回false",
			data:     []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"),
			expected: false,
		},
		{
			name:     "空数据",
			data:     []byte(""),
			expected: false,
		},
		{
			name:     "短数据",
			data:     []byte("short"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapter.HasLowEntropy(tt.data)
			if got != tt.expected {
				t.Errorf("HasLowEntropy() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// translated comment
func TestHasExcessiveEntropy(t *testing.T) {
	adapter := NewLegacyFeatureAdapter(nil)

	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "极高熵值二进制数据",
			data:     []byte("\xff\xfe\xfd\xfc\xfb\xfa\xf9\xf8\xf7\xf6\xf5\xf4\xf3\xf2\xf1\xf0\xef\xee\xed\xec\xeb\xea\xe9\xe8\xe7\xe6\xe5\xe4\xe3\xe2\xe1\xe0\xdf\xde\xdd\xdc\xdb\xda\xd9\xd8\xd7\xd6\xd5\xd4\xd3\xd2\xd1\xd0"),
			expected: false, // translated comment
		},
		{
			name:     "正常熵值返回false",
			data:     []byte("This is a normal text with reasonable entropy levels."),
			expected: false,
		},
		{
			name:     "低熵数据返回false",
			data:     []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
			expected: false,
		},
		{
			name:     "空数据",
			data:     []byte(""),
			expected: false,
		},
		{
			name:     "短数据",
			data:     []byte("short"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapter.HasExcessiveEntropy(tt.data)
			if got != tt.expected {
				t.Errorf("HasExcessiveEntropy() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// translated comment
func TestContainsSpoofingMarkers(t *testing.T) {
	adapter := NewLegacyFeatureAdapter(nil)

	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "包含selenium标记返回true",
			data:     []byte("User-Agent contains selenium driver"),
			expected: true,
		},
		{
			name:     "包含webdriver标记返回true",
			data:     []byte("This is a webdriver instance"),
			expected: true,
		},
		{
			name:     "包含puppeteer标记返回true",
			data:     []byte("Running with puppeteer automation"),
			expected: true,
		},
		{
			name:     "包含PhantomJS标记返回true",
			data:     []byte("PhantomJS browser detected"),
			expected: true,
		},
		{
			name:     "包含HeadlessChrome标记返回true",
			data:     []byte("HeadlessChrome/90.0"),
			expected: true,
		},
		{
			name:     "无标记返回false",
			data:     []byte("Normal browser user agent string"),
			expected: false,
		},
		{
			name:     "空数据",
			data:     []byte(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapter.ContainsSpoofingMarkers(tt.data)
			if got != tt.expected {
				t.Errorf("ContainsSpoofingMarkers() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// translated comment
func TestDetectHeadlessBrowser(t *testing.T) {
	adapter := NewLegacyFeatureAdapter(nil)

	tests := []struct {
		name       string
		userAgent  string
		expected   bool
	}{
		{
			name:       "检测headlesschrome无头浏览器",
			userAgent:  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/90.0.4430.93 Safari/537.36",
			expected:   true,
		},
		{
			name:       "检测phantomjs无头浏览器",
			userAgent:  "Mozilla/5.0 (Unknown; Linux x86_64) AppleWebKit/538.1 (KHTML, like Gecko) PhantomJS/2.1.1 Safari/538.1",
			expected:   true,
		},
		{
			name:       "检测selenium无头浏览器",
			userAgent:  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36 Selenium/3.141.59",
			expected:   true,
		},
		{
			name:       "检测playwright无头浏览器",
			userAgent:  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36 Playwright/1.12.0",
			expected:   true,
		},
		{
			name:       "正常Chrome浏览器返回false",
			userAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36",
			expected:   false,
		},
		{
			name:       "正常Firefox浏览器返回false",
			userAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:88.0) Gecko/20100101 Firefox/88.0",
			expected:   false,
		},
		{
			name:       "空User-Agent返回false",
			userAgent:  "",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapter.DetectHeadlessBrowser(tt.userAgent)
			if got != tt.expected {
				t.Errorf("DetectHeadlessBrowser() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// translated comment
func TestCheckContradictions(t *testing.T) {
	adapter := NewLegacyFeatureAdapter(nil)

	tests := []struct {
		name       string
		attributes map[string]string
		expected   bool
	}{
		{
			name:       "空属性返回false",
			attributes: map[string]string{},
			expected:   false,
		},
		{
			name:       "nil属性",
			attributes: nil,
			expected:   false,
		},
		{
			name: "OS与Platform矛盾（Windows配Mac平台）",
			attributes: map[string]string{
				"os":       "Windows 10",
				"platform": "MacIntel",
			},
			expected: true,
		},
		{
			name: "OS与Platform矛盾（Mac配Win平台）",
			attributes: map[string]string{
				"os":       "MacOS 15",
				"platform": "Win32",
			},
			expected: true,
		},
		{
			name: "OS与Platform矛盾（Linux配Mac平台）",
			attributes: map[string]string{
				"os":       "Linux",
				"platform": "MacIntel",
			},
			expected: true,
		},
		{
			name: "OS与Platform无矛盾（Windows配Win32）",
			attributes: map[string]string{
				"os":       "Windows 10",
				"platform": "Win32",
			},
			expected: false,
		},
		{
			name: "UA与OS矛盾（Windows UA配Mac OS）",
			attributes: map[string]string{
				"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
				"os":         "MacOS 15",
			},
			expected: true,
		},
		{
			name: "UA与OS矛盾（Mac UA配Windows OS）",
			attributes: map[string]string{
				"user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 15_0)",
				"os":         "Windows 10",
			},
			expected: true,
		},
		{
			name: "UA与OS矛盾（Linux UA配Windows OS）",
			attributes: map[string]string{
				"user_agent": "Mozilla/5.0 (X11; Linux x86_64)",
				"os":         "Windows 10",
			},
			expected: true,
		},
		{
			name: "UA与OS无矛盾",
			attributes: map[string]string{
				"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
				"os":         "Windows 10",
			},
			expected: false,
		},
		{
			name: "UA与Features矛盾（Chrome 60配WebGL2）",
			attributes: map[string]string{
				"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/60.0.3112.113",
				"features":   "webgl2,canvas,webrtc",
			},
			expected: true,
		},
		{
			name: "UA与Features矛盾（Mobile配Desktop）",
			attributes: map[string]string{
				"user_agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) Mobile/15E148",
				"features":   "desktop",
			},
			expected: true,
		},
		{
			name: "UA与Features无矛盾",
			attributes: map[string]string{
				"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/90.0.4430.93",
				"features":   "webgl2,canvas",
			},
			expected: false,
		},
		{
			name: "移动设备与屏幕分辨率矛盾（移动设备超大分辨率）",
			attributes: map[string]string{
				"is_mobile":    "true",
				"screen_width": "2560",
			},
			expected: true,
		},
		{
			name: "桌面设备与屏幕分辨率矛盾（桌面设备超小分辨率）",
			attributes: map[string]string{
				"is_mobile":    "false",
				"screen_width": "400",
			},
			expected: true,
		},
		{
			name: "移动设备与屏幕分辨率无矛盾",
			attributes: map[string]string{
				"is_mobile":    "true",
				"screen_width": "375",
			},
			expected: false,
		},
		{
			name: "无矛盾返回false",
			attributes: map[string]string{
				"os":           "Windows 10",
				"platform":     "Win32",
				"user_agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
				"features":     "webgl2",
				"is_mobile":    "false",
				"screen_width": "1920",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapter.CheckContradictions(tt.attributes)
			if got != tt.expected {
				t.Errorf("CheckContradictions() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// translated comment
func TestRecognizeFromHeaders(t *testing.T) {
	adapter := NewLegacyFeatureAdapter(nil)

	tests := []struct {
		name           string
		headers        map[string]string
		wantBrowser    string
		wantOS         string
		wantIsMobile   bool
		wantIsBot      bool
		wantConfidence float64
	}{
		{
			name:           "空User-Agent",
			headers:        map[string]string{},
			wantBrowser:    "",
			wantOS:         "",
			wantIsMobile:   false,
			wantIsBot:      false,
			wantConfidence: 0.0,
		},
		{
			name: "无头浏览器识别为Bot",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 HeadlessChrome/90.0.4430.93",
			},
			wantBrowser:    "",
			wantOS:         "",
			wantIsMobile:   false,
			wantIsBot:      true,
			wantConfidence: 0.9,
		},
		{
			name: "识别Chrome浏览器",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36",
				"Accept":     "text/html,application/xhtml+xml",
			},
			wantBrowser:    "Chrome",
			wantOS:         "Windows 10",
			wantIsMobile:   false,
			wantIsBot:      false,
			wantConfidence: 0.8,
		},
		{
			name: "识别Firefox浏览器",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:88.0) Gecko/20100101 Firefox/88.0",
				"Accept":     "text/html,application/xhtml+xml",
			},
			wantBrowser:    "Firefox",
			wantOS:         "Windows 10",
			wantIsMobile:   false,
			wantIsBot:      false,
			wantConfidence: 0.8,
		},
		{
			name: "识别Safari浏览器",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1 Safari/605.1.15",
				"Accept":     "text/html,application/xhtml+xml",
			},
			wantBrowser:    "Safari",
			wantOS:         "Windows 10",
			wantIsMobile:   false,
			wantIsBot:      false,
			wantConfidence: 0.8,
		},
		{
			name: "识别Edge浏览器",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36 Edg/90.0.818.49",
				"Accept":     "text/html,application/xhtml+xml",
			},
			wantBrowser:    "Edge",
			wantOS:         "Windows 10",
			wantIsMobile:   false,
			wantIsBot:      false,
			wantConfidence: 0.8,
		},
		{
			name: "识别Opera浏览器",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36 OPR/76.0.4017.123",
				"Accept":     "text/html,application/xhtml+xml",
			},
			wantBrowser:    "Opera",
			wantOS:         "Windows 10",
			wantIsMobile:   false,
			wantIsBot:      false,
			wantConfidence: 0.8,
		},
		{
			name: "识别移动设备（Android）",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Mobile Safari/537.36",
				"Accept":     "text/html",
			},
			wantBrowser:    "Chrome",
			wantOS:         "Windows 10",
			wantIsMobile:   true,
			wantIsBot:      false,
			wantConfidence: 0.8,
		},
		{
			name: "识别移动设备（iPhone）",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
				"Accept":     "text/html",
			},
			wantBrowser:    "Safari",
			wantOS:         "Windows 10",
			wantIsMobile:   true,
			wantIsBot:      false,
			wantConfidence: 0.8,
		},
		{
			name: "识别iPad",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (iPad; CPU OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
				"Accept":     "text/html",
			},
			wantBrowser:    "Safari",
			wantOS:         "Windows 10",
			wantIsMobile:   true,
			wantIsBot:      false,
			wantConfidence: 0.8,
		},
		{
			name: "识别MacOS",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36",
				"Accept":     "text/html",
			},
			wantBrowser:    "Chrome",
			wantOS:         "MacOS 13",
			wantIsMobile:   false,
			wantIsBot:      false,
			wantConfidence: 0.8,
		},
		{
			name: "识别Linux",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36",
				"Accept":     "text/html",
			},
			wantBrowser:    "Chrome",
			wantOS:         "Linux",
			wantIsMobile:   false,
			wantIsBot:      false,
			wantConfidence: 0.8,
		},
		{
			name: "高置信度（完整头部）",
			headers: map[string]string{
				"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/90.0.4430.93",
				"Accept":          "text/html,application/xhtml+xml",
				"Accept-Language": "en-US,en;q=0.9",
				"Sec-CH-UA":       "\"Google Chrome\";v=\"90\"",
			},
			wantBrowser:    "Chrome",
			wantOS:         "Windows 10",
			wantIsMobile:   false,
			wantIsBot:      false,
			wantConfidence: 1.0,
		},
		{
			name: "低置信度（仅User-Agent）",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/90.0.4430.93",
			},
			wantBrowser:    "Chrome",
			wantOS:         "Windows 10",
			wantIsMobile:   false,
			wantIsBot:      false,
			wantConfidence: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapter.RecognizeFromHeaders(tt.headers)

			if got.Browser != tt.wantBrowser {
				t.Errorf("Browser = %v, want %v", got.Browser, tt.wantBrowser)
			}
			if got.OS != tt.wantOS {
				t.Errorf("OS = %v, want %v", got.OS, tt.wantOS)
			}
			if got.IsMobile != tt.wantIsMobile {
				t.Errorf("IsMobile = %v, want %v", got.IsMobile, tt.wantIsMobile)
			}
			if got.IsBot != tt.wantIsBot {
				t.Errorf("IsBot = %v, want %v", got.IsBot, tt.wantIsBot)
			}
			if math.Abs(got.Confidence-tt.wantConfidence) > 0.0001 {
				t.Errorf("Confidence = %v, want %v (diff > 0.0001)", got.Confidence, tt.wantConfidence)
			}
		})
	}
}

// translated comment
func TestDetectBrowserFromUALegacy(t *testing.T) {
	tests := []struct {
		name            string
		ua              string
		wantBrowser     string
		wantVersion     string
	}{
		{
			name:            "检测Chrome",
			ua:              "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36",
			wantBrowser:     "Chrome",
			wantVersion:     "90",
		},
		{
			name:            "检测Firefox",
			ua:              "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:88.0) Gecko/20100101 Firefox/88.0",
			wantBrowser:     "Firefox",
			wantVersion:     "88",
		},
		{
			name:            "检测Safari",
			ua:              "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1 Safari/605.1.15",
			wantBrowser:     "Safari",
			wantVersion:     "14",
		},
		{
			name:            "检测Edge（Edg格式）",
			ua:              "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36 Edg/90.0.818.49",
			wantBrowser:     "Edge",
			wantVersion:     "90",
		},
		{
			name:            "检测Opera",
			ua:              "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36 OPR/76.0.4017.123",
			wantBrowser:     "Opera",
			wantVersion:     "76",
		},
		{
			name:            "Edge在Chrome之前检测",
			ua:              "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36 Edg/90.0.818.49",
			wantBrowser:     "Edge",
			wantVersion:     "90",
		},
		{
			name:            "Opera在Chrome之前检测",
			ua:              "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36 OPR/76.0.4017.123",
			wantBrowser:     "Opera",
			wantVersion:     "76",
		},
		{
			name:            "Chromium包含Chrome字符串",
			ua:              "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chromium/90.0.4430.93 Chrome/90.0.4430.93 Safari/537.36",
			wantBrowser:     "Safari", // translated comment
			wantVersion:     "",
		},
		{
			name:            "默认返回Chrome",
			ua:              "Unknown browser agent",
			wantBrowser:     "Chrome",
			wantVersion:     "",
		},
		{
			name:            "空User-Agent默认Chrome",
			ua:              "",
			wantBrowser:     "Chrome",
			wantVersion:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBrowser, gotVersion := detectBrowserFromUALegacy(tt.ua)
			if gotBrowser != tt.wantBrowser {
				t.Errorf("detectBrowserFromUALegacy() browser = %v, want %v", gotBrowser, tt.wantBrowser)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("detectBrowserFromUALegacy() version = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}

// translated comment
func TestExtractVersionFromUALegacy(t *testing.T) {
	tests := []struct {
		name    string
		ua      string
		prefix  string
		want    string
	}{
		{
			name:    "提取Chrome版本",
			ua:      "Mozilla/5.0 Chrome/90.0.4430.93 Safari/537.36",
			prefix:  "Chrome/",
			want:    "90",
		},
		{
			name:    "提取Firefox版本",
			ua:      "Mozilla/5.0 Firefox/88.0",
			prefix:  "Firefox/",
			want:    "88",
		},
		{
			name:    "提取Safari版本",
			ua:      "Mozilla/5.0 Version/14.1 Safari/605.1.15",
			prefix:  "Version/",
			want:    "14",
		},
		{
			name:    "提取Edge版本",
			ua:      "Mozilla/5.0 Edg/90.0.818.49",
			prefix:  "Edg/",
			want:    "90",
		},
		{
			name:    "提取Opera版本",
			ua:      "Mozilla/5.0 OPR/76.0.4017.123",
			prefix:  "OPR/",
			want:    "76",
		},
		{
			name:    "前缀不存在",
			ua:      "Mozilla/5.0 Chrome/90.0.4430.93",
			prefix:  "Firefox/",
			want:    "",
		},
		{
			name:    "空User-Agent",
			ua:      "",
			prefix:  "Chrome/",
			want:    "",
		},
		{
			name:    "版本号在末尾",
			ua:      "Mozilla/5.0 Chrome/90",
			prefix:  "Chrome/",
			want:    "90",
		},
		{
			name:    "版本号后跟空格",
			ua:      "Mozilla/5.0 Chrome/90 Safari/537.36",
			prefix:  "Chrome/",
			want:    "90",
		},
		{
			name:    "版本号后跟分号",
			ua:      "Mozilla/5.0 Chrome/90; Safari/537.36",
			prefix:  "Chrome/",
			want:    "90",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractVersionFromUALegacy(tt.ua, tt.prefix)
			if got != tt.want {
				t.Errorf("extractVersionFromUALegacy() = %v, want %v", got, tt.want)
			}
		})
	}
}

// translated comment
func TestDetectOSFromUALegacy(t *testing.T) {
	tests := []struct {
		name    string
		ua      string
		want    string
	}{
		{
			name:    "检测Windows 10",
			ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/90.0.4430.93",
			want:    "Windows 10",
		},
		{
			name:    "检测MacOS 15",
			ua:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 15_0) AppleWebKit/537.36",
			want:    "MacOS 15",
		},
		{
			name:    "检测MacOS 14",
			ua:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4) AppleWebKit/537.36",
			want:    "MacOS 14",
		},
		{
			name:    "检测MacOS 13",
			ua:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0) AppleWebKit/537.36",
			want:    "MacOS 13",
		},
		{
			name:    "检测Linux",
			ua:      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
			want:    "Linux",
		},
		{
			name:    "默认返回Windows 10",
			ua:      "Unknown OS",
			want:    "Windows 10",
		},
		{
			name:    "空User-Agent默认Windows 10",
			ua:      "",
			want:    "Windows 10",
		},
		{
			name:    "检测MacOS 15（完整格式）",
			ua:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 15_0_0) AppleWebKit/537.36",
			want:    "MacOS 15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectOSFromUALegacy(tt.ua)
			if got != tt.want {
				t.Errorf("detectOSFromUALegacy() = %v, want %v", got, tt.want)
			}
		})
	}
}

// translated comment
func TestCalculateConfidenceLegacy(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    float64
	}{
		{
			name:    "空头部",
			headers: map[string]string{},
			want:    0.5,
		},
		{
			name: "仅User-Agent",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0",
			},
			want: 0.7,
		},
		{
			name: "User-Agent和Accept",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0",
				"Accept":     "text/html",
			},
			want: 0.8,
		},
		{
			name: "User-Agent、Accept和Accept-Language",
			headers: map[string]string{
				"User-Agent":      "Mozilla/5.0",
				"Accept":          "text/html",
				"Accept-Language": "en-US",
			},
			want: 0.9,
		},
		{
			name: "完整头部（所有字段）",
			headers: map[string]string{
				"User-Agent":      "Mozilla/5.0",
				"Accept":          "text/html",
				"Accept-Language": "en-US",
				"Sec-CH-UA":       "\"Google Chrome\";v=\"90\"",
			},
			want: 1.0,
		},
		{
			name: "置信度上限1.0",
			headers: map[string]string{
				"User-Agent":      "Mozilla/5.0",
				"Accept":          "text/html",
				"Accept-Language": "en-US",
				"Sec-CH-UA":       "Chrome/90",
				"Extra-Header":    "value",
			},
			want: 1.0,
		},
		{
			name: "只有Accept",
			headers: map[string]string{
				"Accept": "text/html",
			},
			want: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateConfidenceLegacy(tt.headers)
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("calculateConfidenceLegacy() = %v, want %v (diff > 0.0001)", got, tt.want)
			}
		})
	}
}

// translated comment
func TestLegacyFeatureAdapter_Integration(t *testing.T) {
	// translated comment
	adapter := NewLegacyFeatureAdapter(nil)

	// translated comment
	t.Run("检测正常浏览器", func(t *testing.T) {
		headers := map[string]string{
			"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36",
			"Accept":          "text/html,application/xhtml+xml",
			"Accept-Language": "en-US,en;q=0.9",
		}

		result := adapter.RecognizeFromHeaders(headers)

		if result.IsBot {
			t.Error("正常浏览器不应被识别为Bot")
		}
		if result.Browser != "Chrome" {
			t.Errorf("浏览器识别错误: got %v, want Chrome", result.Browser)
		}
		if result.OS != "Windows 10" {
			t.Errorf("操作系统识别错误: got %v, want Windows 10", result.OS)
		}
	})

	// translated comment
	t.Run("检测无头浏览器", func(t *testing.T) {
		headers := map[string]string{
			"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 HeadlessChrome/90.0.4430.93",
			"Accept":     "text/html",
		}

		result := adapter.RecognizeFromHeaders(headers)

		if !result.IsBot {
			t.Error("无头浏览器应被识别为Bot")
		}
		if result.Confidence != 0.9 {
			t.Errorf("Bot置信度应为0.9: got %v", result.Confidence)
		}
	})

	// translated comment
	t.Run("检测欺骗标记数据", func(t *testing.T) {
		data := []byte("User-Agent: selenium webdriver phantomjs puppeteer")

		if !adapter.DetectAnomalies(data) {
			t.Error("包含欺骗标记的数据应被检测为异常")
		}
		if !adapter.ContainsSpoofingMarkers(data) {
			t.Error("应检测到欺骗标记")
		}
	})

	// translated comment
	t.Run("检测属性矛盾", func(t *testing.T) {
		attrs := map[string]string{
			"os":       "Windows 10",
			"platform": "MacIntel",
		}

		if !adapter.CheckContradictions(attrs) {
			t.Error("OS与Platform矛盾应被检测到")
		}
	})

	// translated comment
	t.Run("使用自定义配置", func(t *testing.T) {
		customConfig := &FeatureConfig{
			EntropyHighThreshold:  8.0,
			EntropyLowThreshold:   30,
			ToolMarkers:           []string{"custom_tool"},
			HeadlessMarkers:       []string{"custom_headless"},
			MobileScreenWidthMax:  2000,
			DesktopScreenWidthMin: 900,
		}

		customAdapter := NewLegacyFeatureAdapter(customConfig)

		// translated comment
		data := []byte("This contains custom_tool marker")
		if !customAdapter.ContainsSpoofingMarkers(data) {
			t.Error("应检测到自定义工具标记")
		}

		// translated comment
		data2 := []byte("This contains selenium marker")
		if customAdapter.ContainsSpoofingMarkers(data2) {
			t.Error("不应检测到默认工具标记（配置已更改）")
		}
	})
}

// translated comment
func TestEdgeCases(t *testing.T) {
	adapter := NewLegacyFeatureAdapter(nil)

	t.Run("长文本工具标记检测", func(t *testing.T) {
		// translated comment
		longText := strings.Repeat("Normal text content. ", 100) + "selenium" + strings.Repeat(" More text. ", 100)
		if !adapter.ContainsSpoofingMarkers([]byte(longText)) {
			t.Error("应在长文本中检测到工具标记")
		}
	})

	t.Run("特殊字符处理", func(t *testing.T) {
		specialChars := []byte("!@#$%^&*()_+-=[]{}|;':\",./<>?")
		if adapter.DetectAnomalies(specialChars) {
			t.Error("特殊字符不应被检测为异常（太短）")
		}
	})

	t.Run("Unicode内容处理", func(t *testing.T) {
		unicode := []byte("Hello 世界 🌍 Привет мир")
		// translated comment
		longUnicode := append(unicode, []byte(strings.Repeat(" additional text ", 10))...)
		_ = adapter.DetectAnomalies(longUnicode) // translated comment
	})

	t.Run("矛盾检查-缺失字段", func(t *testing.T) {
		// translated comment
		attrs := map[string]string{
			"os": "Windows 10",
			// translated comment
		}
		if adapter.CheckContradictions(attrs) {
			t.Error("缺少必要字段不应触发矛盾检测")
		}
	})

	t.Run("矛盾检查-无效屏幕宽度", func(t *testing.T) {
		attrs := map[string]string{
			"is_mobile":    "true",
			"screen_width": "invalid",
		}
		if adapter.CheckContradictions(attrs) {
			t.Error("无效屏幕宽度不应触发矛盾")
		}
	})

	t.Run("浏览器检测-旧版Edge", func(t *testing.T) {
		ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.79 Safari/537.36 Edge/14.14393"
		browser, version := detectBrowserFromUALegacy(ua)
		if browser != "Edge" {
			t.Errorf("应检测为Edge: got %v", browser)
		}
		if version != "14" {
			t.Errorf("版本应为14: got %v", version)
		}
	})
}
