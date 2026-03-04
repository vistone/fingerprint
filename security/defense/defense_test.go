package defense

import (
	"testing"

	"github.com/vistone/fingerprint/profiles"
)

// TestNewAnomalyDetector 测试异常检测器创建
func TestNewAnomalyDetector(t *testing.T) {
	detector := NewAnomalyDetector()
	if detector == nil {
		t.Fatal("NewAnomalyDetector() returned nil")
	}
}

// TestAnomalyDetector_DetectAnomalies 测试异常检测
func TestAnomalyDetector_DetectAnomalies(t *testing.T) {
	detector := NewAnomalyDetector()

	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "empty data",
			data:     []byte{},
			expected: false,
		},
		{
			name:     "nil data",
			data:     nil,
			expected: false,
		},
		{
			name:     "normal data",
			data:     []byte("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
			expected: false,
		},
		{
			name:     "low entropy data (repeated bytes)",
			data:     []byte("aaaaaaaaaaaaaaaaaaaa"),
			expected: true,
		},
		{
			name:     "high entropy data (random-like)",
			data:     []byte{0x00, 0xFF, 0x42, 0x13, 0xA7, 0x8B, 0xC9, 0xD1, 0xE2, 0xF3, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA},
			expected: true,
		},
		{
			name:     "contains HeadlessChrome",
			data:     []byte("Mozilla/5.0 HeadlessChrome/120.0.0.0"),
			expected: true,
		},
		{
			name:     "contains webdriver",
			data:     []byte("normal data with webdriver marker"),
			expected: true,
		},
		{
			name:     "short data",
			data:     []byte("short"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectAnomalies(tt.data)
			if result != tt.expected {
				t.Errorf("DetectAnomalies() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestAnomalyDetector_DetectHeadlessBrowser 测试无头浏览器检测
func TestAnomalyDetector_DetectHeadlessBrowser(t *testing.T) {
	detector := NewAnomalyDetector()

	// 首先获取真实指纹数据
	chromeProfile, ok := profiles.MappedTLSClients["chrome_133"]
	if !ok {
		t.Skip("chrome_133 profile not found")
	}

	_, err := chromeProfile.GetClientHelloSpec()
	if err != nil {
		t.Skipf("chrome_133 does not support spec export: %v", err)
	}

	tests := []struct {
		name     string
		ua       string
		expected bool
	}{
		{
			name:     "normal chrome",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
			expected: false,
		},
		{
			name:     "normal firefox",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
			expected: false,
		},
		{
			name:     "normal safari",
			ua:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0_0) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
			expected: false,
		},
		{
			name:     "headless chrome",
			ua:       "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/120.0.0.0 Safari/537.36",
			expected: true,
		},
		{
			name:     "phantomjs",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/538.1 (KHTML, like Gecko) PhantomJS/2.1.1 Safari/538.1",
			expected: true,
		},
		{
			name:     "selenium",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Selenium/4.0.0",
			expected: true,
		},
		{
			name:     "webdriver",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 WebDriver/4.0",
			expected: true,
		},
		{
			name:     "puppeteer",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Puppeteer/21.0.0",
			expected: true,
		},
		{
			name:     "playwright",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Playwright/1.40.0",
			expected: true,
		},
		{
			name:     "cypress",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Cypress/13.0.0",
			expected: true,
		},
		{
			name:     "jsdom",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 JSDOM/24.0.0",
			expected: true,
		},
		{
			name:     "empty ua",
			ua:       "",
			expected: false,
		},
		{
			name:     "case insensitive headless",
			ua:       "HEADLESSCHROME/120.0.0.0",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectHeadlessBrowser(tt.ua)
			if result != tt.expected {
				t.Errorf("DetectHeadlessBrowser() = %v, want %v for UA: %s", result, tt.expected, tt.ua)
			}
		})
	}
}

// TestNewContradictionDetector 测试矛盾检测器创建
func TestNewContradictionDetector(t *testing.T) {
	detector := NewContradictionDetector()
	if detector == nil {
		t.Fatal("NewContradictionDetector() returned nil")
	}
}

// TestContradictionDetector_CheckContradictions 测试矛盾检测
func TestContradictionDetector_CheckContradictions(t *testing.T) {
	detector := NewContradictionDetector()

	tests := []struct {
		name     string
		attrs    map[string]string
		expected bool
	}{
		{
			name:     "empty attributes",
			attrs:    map[string]string{},
			expected: false,
		},
		{
			name:     "nil attributes",
			attrs:    nil,
			expected: false,
		},
		{
			name: "consistent windows",
			attrs: map[string]string{
				"os":       "Windows NT 10.0",
				"platform": "Win32",
			},
			expected: false,
		},
		{
			name: "consistent mac",
			attrs: map[string]string{
				"os":       "Mac OS X 14.0",
				"platform": "MacIntel",
			},
			expected: false,
		},
		{
			name: "consistent linux",
			attrs: map[string]string{
				"os":       "Linux x86_64",
				"platform": "Linux x86_64",
			},
			expected: false,
		},
		{
			name: "os platform contradiction - windows os mac platform",
			attrs: map[string]string{
				"os":       "Windows NT 10.0",
				"platform": "MacIntel",
			},
			expected: true,
		},
		{
			name: "os platform contradiction - mac os windows platform",
			attrs: map[string]string{
				"os":       "Mac OS X 14.0",
				"platform": "Win32",
			},
			expected: true,
		},
		{
			name: "os platform contradiction - linux os mac platform",
			attrs: map[string]string{
				"os":       "Linux x86_64",
				"platform": "MacIntel",
			},
			expected: true,
		},
		{
			name: "ua os contradiction - windows ua mac os",
			attrs: map[string]string{
				"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/133.0.0.0",
				"os":         "Mac OS X 14.0",
			},
			expected: true,
		},
		{
			name: "ua os contradiction - mac ua windows os",
			attrs: map[string]string{
				"user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0_0) Chrome/133.0.0.0",
				"os":         "Windows NT 10.0",
			},
			expected: true,
		},
		{
			name: "ua os contradiction - linux ua windows os",
			attrs: map[string]string{
				"user_agent": "Mozilla/5.0 (X11; Linux x86_64) Chrome/133.0.0.0",
				"os":         "Windows NT 10.0",
			},
			expected: true,
		},
		{
			name: "mobile screen contradiction - mobile with desktop resolution",
			attrs: map[string]string{
				"is_mobile":    "true",
				"screen_width": "2560",
			},
			expected: true,
		},
		{
			name: "mobile screen contradiction - desktop with mobile resolution",
			attrs: map[string]string{
				"is_mobile":    "false",
				"screen_width": "375",
			},
			expected: true,
		},
		{
			name: "consistent mobile",
			attrs: map[string]string{
				"is_mobile":    "true",
				"screen_width": "390",
			},
			expected: false,
		},
		{
			name: "consistent desktop",
			attrs: map[string]string{
				"is_mobile":    "false",
				"screen_width": "1920",
			},
			expected: false,
		},
		{
			name: "ua feature contradiction - old chrome with webgl2",
			attrs: map[string]string{
				"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/60.0.0.0",
				"features":   "WebGL2",
			},
			expected: true,
		},
		{
			name: "ua feature contradiction - mobile with desktop",
			attrs: map[string]string{
				"user_agent": "Mozilla/5.0 (Linux; Android 10; Mobile) Chrome/133.0.0.0",
				"features":   "desktop",
			},
			expected: true,
		},
		{
			name: "modern chrome with webgl2 (not contradiction)",
			attrs: map[string]string{
				"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/133.0.0.0",
				"features":   "WebGL2",
			},
			expected: false,
		},
		{
			name: "multiple contradictions",
			attrs: map[string]string{
				"user_agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/133.0.0.0",
				"os":           "Mac OS X 14.0",
				"platform":     "MacIntel",
				"is_mobile":    "true",
				"screen_width": "2560",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.CheckContradictions(tt.attrs)
			if result != tt.expected {
				t.Errorf("CheckContradictions() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestNewPassiveRecognizer 测试被动识别器创建
func TestNewPassiveRecognizer(t *testing.T) {
	recognizer := NewPassiveRecognizer()
	if recognizer == nil {
		t.Fatal("NewPassiveRecognizer() returned nil")
	}
}

// TestPassiveRecognizer_RecognizeFromHeaders 测试从 Headers 识别
func TestPassiveRecognizer_RecognizeFromHeaders(t *testing.T) {
	recognizer := NewPassiveRecognizer()

	tests := []struct {
		name           string
		headers        map[string]string
		wantBrowser    string
		wantIsMobile   bool
		wantIsBot      bool
		wantConfidence float64
	}{
		{
			name:           "empty headers",
			headers:        map[string]string{},
			wantBrowser:    "",
			wantIsMobile:   false,
			wantIsBot:      false,
			wantConfidence: 0.0,
		},
		{
			name: "chrome windows",
			headers: map[string]string{
				"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
				"Accept":          "text/html,application/xhtml+xml",
				"Accept-Language": "en-US,en;q=0.9",
				"Sec-CH-UA":       `"Google Chrome";v="133"`,
			},
			wantBrowser:    "chrome",
			wantIsMobile:   false,
			wantIsBot:      false,
			wantConfidence: 0.9,
		},
		{
			name: "firefox mac",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 14.0; rv:135.0) Gecko/20100101 Firefox/135.0",
			},
			wantBrowser:    "firefox",
			wantIsMobile:   false,
			wantIsBot:      false,
			wantConfidence: 0.7,
		},
		{
			name: "safari ios mobile",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			},
			wantBrowser:    "safari",
			wantIsMobile:   true,
			wantIsBot:      false,
			wantConfidence: 0.7,
		},
		{
			name: "edge windows",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/133.0.0.0",
			},
			wantBrowser:    "edge",
			wantIsMobile:   false,
			wantIsBot:      false,
			wantConfidence: 0.7,
		},
		{
			name: "opera",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/91.0.0.0",
			},
			wantBrowser:    "opera",
			wantIsMobile:   false,
			wantIsBot:      false,
			wantConfidence: 0.7,
		},
		{
			name: "headless chrome bot",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/120.0.0.0 Safari/537.36",
			},
			wantBrowser:    "",
			wantIsMobile:   false,
			wantIsBot:      true,
			wantConfidence: 0.9,
		},
		{
			name: "android mobile",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Mobile Safari/537.36",
			},
			wantBrowser:    "chrome",
			wantIsMobile:   true,
			wantIsBot:      false,
			wantConfidence: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := recognizer.RecognizeFromHeaders(tt.headers)

			if string(result.Browser) != tt.wantBrowser {
				t.Errorf("Browser = %v, want %v", result.Browser, tt.wantBrowser)
			}

			if result.IsMobile != tt.wantIsMobile {
				t.Errorf("IsMobile = %v, want %v", result.IsMobile, tt.wantIsMobile)
			}

			if result.IsBot != tt.wantIsBot {
				t.Errorf("IsBot = %v, want %v", result.IsBot, tt.wantIsBot)
			}

			if result.Confidence < tt.wantConfidence - 0.01 {
				t.Errorf("Confidence = %v, want >= %v", result.Confidence, tt.wantConfidence)
			}
		})
	}
}

// TestRecognizeFromUserAgent 测试从 User-Agent 识别
func TestRecognizeFromUserAgent(t *testing.T) {
	tests := []struct {
		name         string
		ua           string
		wantBrowser  string
		wantIsMobile bool
	}{
		{
			name:         "chrome",
			ua:           "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
			wantBrowser:  "chrome",
			wantIsMobile: false,
		},
		{
			name:         "firefox",
			ua:           "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
			wantBrowser:  "firefox",
			wantIsMobile: false,
		},
		{
			name:         "safari",
			ua:           "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0_0) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
			wantBrowser:  "safari",
			wantIsMobile: false,
		},
		{
			name:         "edge",
			ua:           "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/133.0.0.0",
			wantBrowser:  "edge",
			wantIsMobile: false,
		},
		{
			name:         "opera",
			ua:           "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/91.0.0.0",
			wantBrowser:  "opera",
			wantIsMobile: false,
		},
		{
			name:         "mobile chrome",
			ua:           "Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Mobile Safari/537.36",
			wantBrowser:  "chrome",
			wantIsMobile: true,
		},
		{
			name:         "iphone safari",
			ua:           "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			wantBrowser:  "safari",
			wantIsMobile: true,
		},
		{
			name:         "ipad",
			ua:           "Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			wantBrowser:  "safari",
			wantIsMobile: true,
		},
		{
			name:         "empty ua",
			ua:           "",
			wantBrowser:  "", // 空UA返回空
			wantIsMobile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RecognizeFromUserAgent(tt.ua)

			if string(result.Browser) != tt.wantBrowser {
				t.Errorf("Browser = %v, want %v", result.Browser, tt.wantBrowser)
			}

			if result.IsMobile != tt.wantIsMobile {
				t.Errorf("IsMobile = %v, want %v", result.IsMobile, tt.wantIsMobile)
			}
		})
	}
}

// TestDetectOSFromUA 测试从 UA 检测操作系统
func TestDetectOSFromUA(t *testing.T) {
	tests := []struct {
		name    string
		ua      string
		wantOS  string
	}{
		{
			name:   "windows 10",
			ua:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/133.0.0.0",
			wantOS: "Windows NT 10.0; Win64; x64",
		},
		{
			name:   "macos 15",
			ua:     "Mozilla/5.0 (Macintosh; Intel Mac OS X 15_0_0) Chrome/133.0.0.0",
			wantOS: "Macintosh; Intel Mac OS X 15_0_0",
		},
		{
			name:   "macos 14",
			ua:     "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0_0) Chrome/133.0.0.0",
			wantOS: "Macintosh; Intel Mac OS X 14_0_0",
		},
		{
			name:   "macos 13",
			ua:     "Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0_0) Chrome/133.0.0.0",
			wantOS: "Macintosh; Intel Mac OS X 13_0_0",
		},
		{
			name:   "linux",
			ua:     "Mozilla/5.0 (X11; Linux x86_64) Chrome/133.0.0.0",
			wantOS: "X11; Linux x86_64",
		},
		{
			name:   "unknown os (default to windows)",
			ua:     "Mozilla/5.0 (Unknown OS) Chrome/133.0.0.0",
			wantOS: "Windows NT 10.0; Win64; x64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectOSFromUA(tt.ua)
			if string(got) != tt.wantOS {
				t.Errorf("detectOSFromUA() = %v, want %v", got, tt.wantOS)
			}
		})
	}
}

// TestExtractVersionFromUA 测试从 UA 提取版本
func TestExtractVersionFromUA(t *testing.T) {
	tests := []struct {
		name   string
		ua     string
		prefix string
		want   string
	}{
		{
			name:   "chrome version",
			ua:     "Mozilla/5.0 Chrome/133.0.0.0 Safari/537.36",
			prefix: "Chrome/",
			want:   "133",
		},
		{
			name:   "firefox version",
			ua:     "Mozilla/5.0 Firefox/135.0",
			prefix: "Firefox/",
			want:   "135",
		},
		{
			name:   "safari version",
			ua:     "Mozilla/5.0 Version/17.0 Safari/605.1",
			prefix: "Version/",
			want:   "17",
		},
		{
			name:   "edge version",
			ua:     "Mozilla/5.0 Edg/133.0.0.0",
			prefix: "Edg/",
			want:   "133",
		},
		{
			name:   "opera version",
			ua:     "Mozilla/5.0 OPR/91.0.0.0",
			prefix: "OPR/",
			want:   "91",
		},
		{
			name:   "prefix not found",
			ua:     "Mozilla/5.0 Chrome/133.0.0.0",
			prefix: "Firefox/",
			want:   "",
		},
		{
			name:   "empty ua",
			ua:     "",
			prefix: "Chrome/",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractVersionFromUA(tt.ua, tt.prefix)
			if got != tt.want {
				t.Errorf("extractVersionFromUA() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCalculateConfidence 测试置信度计算
func TestCalculateConfidence(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		result         *RecognitionResult
		wantConfidence float64
	}{
		{
			name:           "empty headers",
			headers:        map[string]string{},
			result:         &RecognitionResult{},
			wantConfidence: 0.5,
		},
		{
			name: "with user-agent",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 Chrome/133.0.0.0",
			},
			result:         &RecognitionResult{},
			wantConfidence: 0.7,
		},
		{
			name: "with all headers",
			headers: map[string]string{
				"User-Agent":      "Mozilla/5.0 Chrome/133.0.0.0",
				"Accept":          "text/html",
				"Accept-Language": "en-US",
				"Sec-CH-UA":       `"Chrome";v="133"`,
			},
			result:         &RecognitionResult{},
			wantConfidence: 0.99, // 使用 >= 比较而不是 ==
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateConfidence(tt.headers, tt.result)
			// 使用范围比较处理浮点数精度问题
			if got < tt.wantConfidence - 0.01 || got > tt.wantConfidence + 0.01 {
				t.Errorf("calculateConfidence() = %v, want ~%v", got, tt.wantConfidence)
			}
		})
	}
}

// TestParseIntFallback 测试整数解析回退
func TestParseIntFallback(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    int
		wantErr bool
	}{
		{
			name:    "valid number",
			s:       "1920",
			want:    1920,
			wantErr: false,
		},
		{
			name:    "zero",
			s:       "0",
			want:    0,
			wantErr: false,
		},
		{
			name:    "empty string",
			s:       "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "with spaces",
			s:       "  1920  ",
			want:    1920, // 实际会解析成功
			wantErr: false,
		},
		{
			name:    "non-numeric",
			s:       "abc",
			want:    0,
			wantErr: true,
		},
		{
			name:    "mixed",
			s:       "1920abc",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result int
			_, err := parseIntFallback(tt.s, &result)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIntFallback() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.want {
				t.Errorf("parseIntFallback() = %v, want %v", result, tt.want)
			}
		})
	}
}

// BenchmarkDetectHeadlessBrowser 基准测试无头浏览器检测
func BenchmarkDetectHeadlessBrowser(b *testing.B) {
	detector := NewAnomalyDetector()
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.DetectHeadlessBrowser(ua)
	}
}

// BenchmarkCheckContradictions 基准测试矛盾检测
func BenchmarkCheckContradictions(b *testing.B) {
	detector := NewContradictionDetector()
	attrs := map[string]string{
		"user_agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/133.0.0.0",
		"os":           "Windows NT 10.0",
		"platform":     "Win32",
		"is_mobile":    "false",
		"screen_width": "1920",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.CheckContradictions(attrs)
	}
}

// BenchmarkRecognizeFromHeaders 基准测试 Headers 识别
func BenchmarkRecognizeFromHeaders(b *testing.B) {
	recognizer := NewPassiveRecognizer()
	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Accept":          "text/html,application/xhtml+xml",
		"Accept-Language": "en-US,en;q=0.9",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recognizer.RecognizeFromHeaders(headers)
	}
}
