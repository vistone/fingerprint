package features

import (
	"testing"

	"github.com/vistone/fingerprint/modules/profiles/legacy"
)

// TestBaseFeatureExtractor_ExtractFeature_Entropy 测试熵特征提取
func TestBaseFeatureExtractor_ExtractFeature_Entropy(t *testing.T) {
	extractor := NewBaseFeatureExtractor(nil)

	tests := []struct {
		name       string
		data       interface{}
		wantScore  float64
		wantAnomaly bool
	}{
		{
			name:       "normal data",
			data:       []byte("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
			wantScore:  0.0,
			wantAnomaly: false,
		},
		{
			name:       "low entropy data (repeated bytes)",
			data:       []byte("aaaaaaaaaaaaaaaaaaaa"),
			wantScore:  0.95,
			wantAnomaly: true,
		},
		{
			name:       "short data (skip)",
			data:       []byte("short"),
			wantScore:  0.0,
			wantAnomaly: false,
		},
		{
			name:       "empty data",
			data:       []byte{},
			wantScore:  0.0,
			wantAnomaly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, isAnomaly := extractor.ExtractFeature(FeatureEntropy, tt.data, nil)
			if isAnomaly != tt.wantAnomaly {
				t.Errorf("ExtractFeature() isAnomaly = %v, want %v", isAnomaly, tt.wantAnomaly)
			}
			if isAnomaly && score < 0.5 {
				t.Errorf("ExtractFeature() anomaly score = %v, want > 0.5", score)
			}
		})
	}
}

// TestBaseFeatureExtractor_ExtractFeature_ToolMarker 测试工具标记特征
func TestBaseFeatureExtractor_ExtractFeature_ToolMarker(t *testing.T) {
	extractor := NewBaseFeatureExtractor(nil)

	// 使用真实 Chrome 133 指纹的 User-Agent
	chromeProfile, ok := profiles.MappedTLSClients["chrome_133"]
	if !ok {
		t.Skip("chrome_133 profile not found")
	}

	// 获取真实 User-Agent
	spec, err := chromeProfile.GetClientHelloSpec()
	if err != nil {
		t.Skipf("chrome_133 does not support spec export: %v", err)
	}

	// 获取真实指纹的 Headers
	realHeaders := spec.Extensions
	if len(realHeaders) == 0 {
		t.Skip("chrome_133 has no extensions")
	}

	tests := []struct {
		name        string
		data        interface{}
		wantAnomaly bool
	}{
		{
			name:        "normal chrome ua",
			data:        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
			wantAnomaly: false,
		},
		{
			name:        "headless chrome",
			data:        "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/120.0.0.0 Safari/537.36",
			wantAnomaly: true,
		},
		{
			name:        "selenium webdriver",
			data:        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Selenium/4.0.0",
			wantAnomaly: true,
		},
		{
			name:        "phantomjs",
			data:        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/538.1 (KHTML, like Gecko) PhantomJS/2.1.1 Safari/538.1",
			wantAnomaly: true,
		},
		{
			name:        "real chrome spec data",
			data:        realHeaders,
			wantAnomaly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, isAnomaly := extractor.ExtractFeature(FeatureToolMarker, tt.data, nil)
			if isAnomaly != tt.wantAnomaly {
				t.Errorf("ExtractFeature() isAnomaly = %v, want %v", isAnomaly, tt.wantAnomaly)
			}
		})
	}
}

// TestBaseFeatureExtractor_ExtractFeature_HeadlessBrowser 测试无头浏览器检测
func TestBaseFeatureExtractor_ExtractFeature_HeadlessBrowser(t *testing.T) {
	extractor := NewBaseFeatureExtractor(nil)

	tests := []struct {
		name        string
		data        interface{}
		wantAnomaly bool
		wantScore   float64
	}{
		{
			name:        "normal chrome",
			data:        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
			wantAnomaly: false,
			wantScore:   0.0,
		},
		{
			name:        "headless chrome",
			data:        "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/120.0.0.0 Safari/537.36",
			wantAnomaly: true,
			wantScore:   0.95,
		},
		{
			name:        "playwright",
			data:        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Playwright/1.40.0",
			wantAnomaly: true,
			wantScore:   0.95,
		},
		{
			name:        "puppeteer",
			data:        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Puppeteer/21.0.0",
			wantAnomaly: true,
			wantScore:   0.95,
		},
		{
			name:        "from map",
			data:        map[string]string{"user_agent": "HeadlessChrome/120.0.0.0"},
			wantAnomaly: true,
			wantScore:   0.95,
		},
		{
			name:        "empty ua",
			data:        "",
			wantAnomaly: false,
			wantScore:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, isAnomaly := extractor.ExtractFeature(FeatureHeadlessBrowser, tt.data, nil)
			if isAnomaly != tt.wantAnomaly {
				t.Errorf("ExtractFeature() isAnomaly = %v, want %v", isAnomaly, tt.wantAnomaly)
			}
			if isAnomaly && score < 0.5 {
				t.Errorf("ExtractFeature() anomaly score = %v, want >= 0.5", score)
			}
		})
	}
}

// TestBaseFeatureExtractor_ExtractFeature_OSPlatformContradiction 测试操作系统平台矛盾
func TestBaseFeatureExtractor_ExtractFeature_OSPlatformContradiction(t *testing.T) {
	extractor := NewBaseFeatureExtractor(nil)

	tests := []struct {
		name        string
		data        interface{}
		wantAnomaly bool
	}{
		{
			name: "consistent windows",
			data: map[string]string{
				"os":       "Windows NT 10.0",
				"platform": "Win32",
			},
			wantAnomaly: false,
		},
		{
			name: "consistent mac",
			data: map[string]string{
				"os":       "Mac OS X 14.0",
				"platform": "MacIntel",
			},
			wantAnomaly: false,
		},
		{
			name: "contradiction windows os mac platform",
			data: map[string]string{
				"os":       "Windows NT 10.0",
				"platform": "MacIntel",
			},
			wantAnomaly: true,
		},
		{
			name: "contradiction mac os windows platform",
			data: map[string]string{
				"os":       "Mac OS X 14.0",
				"platform": "Win32",
			},
			wantAnomaly: true,
		},
		{
			name: "contradiction linux os mac platform",
			data: map[string]string{
				"os":       "Linux x86_64",
				"platform": "MacIntel",
			},
			wantAnomaly: true,
		},
		{
			name: "missing os",
			data: map[string]string{
				"platform": "Win32",
			},
			wantAnomaly: false,
		},
		{
			name: "invalid data type",
			data: "invalid",
			wantAnomaly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, isAnomaly := extractor.ExtractFeature(FeatureOSPlatformContradiction, tt.data, nil)
			if isAnomaly != tt.wantAnomaly {
				t.Errorf("ExtractFeature() isAnomaly = %v, want %v", isAnomaly, tt.wantAnomaly)
			}
		})
	}
}

// TestBaseFeatureExtractor_ExtractFeature_UAOSContradiction 测试 UA/OS 矛盾
func TestBaseFeatureExtractor_ExtractFeature_UAOSContradiction(t *testing.T) {
	extractor := NewBaseFeatureExtractor(nil)

	tests := []struct {
		name        string
		data        interface{}
		wantAnomaly bool
	}{
		{
			name: "consistent windows",
			data: map[string]string{
				"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/133.0.0.0",
				"os":         "Windows NT 10.0",
			},
			wantAnomaly: false,
		},
		{
			name: "consistent mac",
			data: map[string]string{
				"user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0_0) Chrome/133.0.0.0",
				"os":         "Mac OS X 14.0",
			},
			wantAnomaly: false,
		},
		{
			name: "contradiction windows ua mac os",
			data: map[string]string{
				"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/133.0.0.0",
				"os":         "Mac OS X 14.0",
			},
			wantAnomaly: true,
		},
		{
			name: "contradiction mac ua windows os",
			data: map[string]string{
				"user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0_0) Chrome/133.0.0.0",
				"os":         "Windows NT 10.0",
			},
			wantAnomaly: true,
		},
		{
			name: "contradiction linux ua windows os",
			data: map[string]string{
				"user_agent": "Mozilla/5.0 (X11; Linux x86_64) Chrome/133.0.0.0",
				"os":         "Windows NT 10.0",
			},
			wantAnomaly: true,
		},
		{
			name: "missing user_agent",
			data: map[string]string{
				"os": "Windows NT 10.0",
			},
			wantAnomaly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, isAnomaly := extractor.ExtractFeature(FeatureUAOSContradiction, tt.data, nil)
			if isAnomaly != tt.wantAnomaly {
				t.Errorf("ExtractFeature() isAnomaly = %v, want %v", isAnomaly, tt.wantAnomaly)
			}
		})
	}
}

// TestBaseFeatureExtractor_ExtractFeature_MobileScreenContradiction 测试移动设备屏幕矛盾
func TestBaseFeatureExtractor_ExtractFeature_MobileScreenContradiction(t *testing.T) {
	extractor := NewBaseFeatureExtractor(nil)

	tests := []struct {
		name        string
		data        interface{}
		wantAnomaly bool
	}{
		{
			name: "normal mobile",
			data: map[string]string{
				"is_mobile":    "true",
				"screen_width": "390",
			},
			wantAnomaly: false,
		},
		{
			name: "normal desktop",
			data: map[string]string{
				"is_mobile":    "false",
				"screen_width": "1920",
			},
			wantAnomaly: false,
		},
		{
			name: "mobile with desktop resolution",
			data: map[string]string{
				"is_mobile":    "true",
				"screen_width": "2560",
			},
			wantAnomaly: true,
		},
		{
			name: "desktop with mobile resolution",
			data: map[string]string{
				"is_mobile":    "false",
				"screen_width": "375",
			},
			wantAnomaly: true,
		},
		{
			name: "missing is_mobile",
			data: map[string]string{
				"screen_width": "1920",
			},
			wantAnomaly: false,
		},
		{
			name: "invalid screen width",
			data: map[string]string{
				"is_mobile":    "true",
				"screen_width": "invalid",
			},
			wantAnomaly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, isAnomaly := extractor.ExtractFeature(FeatureMobileScreenContradiction, tt.data, nil)
			if isAnomaly != tt.wantAnomaly {
				t.Errorf("ExtractFeature() isAnomaly = %v, want %v", isAnomaly, tt.wantAnomaly)
			}
		})
	}
}

// TestBaseFeatureExtractor_ExtractFeature_UAFeatureContradiction 测试 UA 特性矛盾
func TestBaseFeatureExtractor_ExtractFeature_UAFeatureContradiction(t *testing.T) {
	extractor := NewBaseFeatureExtractor(nil)

	tests := []struct {
		name        string
		data        interface{}
		wantAnomaly bool
	}{
		{
			name: "chrome 60 with webgl2",
			data: map[string]string{
				"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.0.0 Safari/537.36",
				"features":   "WebGL2",
			},
			wantAnomaly: true,
		},
		{
			name: "mobile with desktop feature",
			data: map[string]string{
				"user_agent": "Mozilla/5.0 (Linux; Android 10; Mobile) Chrome/133.0.0.0",
				"features":   "desktop",
			},
			wantAnomaly: true,
		},
		{
			name: "modern chrome with webgl2",
			data: map[string]string{
				"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/133.0.0.0",
				"features":   "WebGL2",
			},
			wantAnomaly: false,
		},
		{
			name: "missing features",
			data: map[string]string{
				"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/133.0.0.0",
			},
			wantAnomaly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, isAnomaly := extractor.ExtractFeature(FeatureUAFeatureContradiction, tt.data, nil)
			if isAnomaly != tt.wantAnomaly {
				t.Errorf("ExtractFeature() isAnomaly = %v, want %v", isAnomaly, tt.wantAnomaly)
			}
		})
	}
}

// TestBaseFeatureExtractor_GetFeatureName 测试特征名称获取
func TestBaseFeatureExtractor_GetFeatureName(t *testing.T) {
	extractor := NewBaseFeatureExtractor(nil)

	tests := []struct {
		featureType FeatureType
		wantName    string
	}{
		{FeatureEntropy, "Entropy Anomaly"},
		{FeatureToolMarker, "Tool Marker Detection"},
		{FeatureOSPlatformContradiction, "OS/Platform Contradiction"},
		{FeatureUAOSContradiction, "UA/OS Contradiction"},
		{FeatureMobileScreenContradiction, "Mobile/Screen Contradiction"},
		{FeatureUAFeatureContradiction, "UA/Feature Contradiction"},
		{FeatureHeadlessBrowser, "Headless Browser Detection"},
		{FeatureType("unknown"), "Unknown Feature"},
	}

	for _, tt := range tests {
		t.Run(string(tt.featureType), func(t *testing.T) {
			got := extractor.GetFeatureName(tt.featureType)
			if got != tt.wantName {
				t.Errorf("GetFeatureName() = %v, want %v", got, tt.wantName)
			}
		})
	}
}

// TestBaseFeatureExtractor_ExtractFeatureVector 测试完整特征向量提取
func TestBaseFeatureExtractor_ExtractFeatureVector(t *testing.T) {
	extractor := NewBaseFeatureExtractor(nil)

	data := map[string]interface{}{
		"user_agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/133.0.0.0",
		"os":           "Windows NT 10.0",
		"platform":     "Win32",
		"is_mobile":    "false",
		"screen_width": "1920",
		"features":     "WebGL2",
	}

	vector := extractor.ExtractFeatureVector(data, nil)

	if vector == nil {
		t.Fatal("ExtractFeatureVector() returned nil")
	}

	if len(vector.Scores) == 0 {
		t.Error("ExtractFeatureVector() returned empty scores")
	}

	if vector.Hash == "" {
		t.Error("ExtractFeatureVector() returned empty hash")
	}

	// 正常数据不应该有异常
	if vector.RiskScore > 0.5 {
		t.Errorf("ExtractFeatureVector() normal data has high risk score: %v", vector.RiskScore)
	}
}

// TestBaseFeatureExtractor_ExtractFeatureVector_Anomalous 测试异常数据的特征向量
func TestBaseFeatureExtractor_ExtractFeatureVector_Anomalous(t *testing.T) {
	extractor := NewBaseFeatureExtractor(nil)

	// 异常数据：UA 和 OS 矛盾
	data := map[string]interface{}{
		"user_agent":   "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0_0) Chrome/133.0.0.0", // Mac UA
		"os":           "Windows NT 10.0",                                                // Windows OS - 矛盾！
		"platform":     "Win32",
		"is_mobile":    "true",          // 移动设备
		"screen_width": "2560",          // 移动设备大屏幕 - 矛盾！
		"features":     "WebGL2",
	}

	vector := extractor.ExtractFeatureVector(data, nil)

	if vector == nil {
		t.Fatal("ExtractFeatureVector() returned nil")
	}

	// 应该有异常检测到
	if len(vector.Anomalies) == 0 {
		t.Error("Expected anomalies for contradictory data, got none")
	}

	// 风险评分应该较高
	if vector.RiskScore < 0.5 {
		t.Errorf("Expected high risk score for anomalous data, got %v", vector.RiskScore)
	}
}

// TestDefaultFeatureConfig 测试默认配置
func TestDefaultFeatureConfig(t *testing.T) {
	config := DefaultFeatureConfig()

	if config == nil {
		t.Fatal("DefaultFeatureConfig() returned nil")
	}

	if config.EntropyHighThreshold <= 0 {
		t.Errorf("EntropyHighThreshold = %v, want > 0", config.EntropyHighThreshold)
	}

	if config.EntropyLowThreshold <= 0 {
		t.Errorf("EntropyLowThreshold = %v, want > 0", config.EntropyLowThreshold)
	}

	if len(config.ToolMarkers) == 0 {
		t.Error("ToolMarkers is empty")
	}

	if len(config.HeadlessMarkers) == 0 {
		t.Error("HeadlessMarkers is empty")
	}

	if config.MobileScreenWidthMax <= 0 {
		t.Errorf("MobileScreenWidthMax = %v, want > 0", config.MobileScreenWidthMax)
	}

	if config.DesktopScreenWidthMin <= 0 {
		t.Errorf("DesktopScreenWidthMin = %v, want > 0", config.DesktopScreenWidthMin)
	}
}

// TestBaseFeatureExtractor_Concurrency 测试并发安全性
func TestBaseFeatureExtractor_Concurrency(t *testing.T) {
	extractor := NewBaseFeatureExtractor(nil)

	// 并发读取测试
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			data := map[string]interface{}{
				"user_agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/133.0.0.0",
				"os":           "Windows NT 10.0",
				"platform":     "Win32",
				"is_mobile":    "false",
				"screen_width": "1920",
			}
			_ = extractor.ExtractFeatureVector(data, nil)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestCalculateRiskScore 测试风险评分计算
func TestCalculateRiskScore(t *testing.T) {
	tests := []struct {
		name   string
		scores map[FeatureType]float64
		want   float64
	}{
		{
			name:   "empty scores",
			scores: map[FeatureType]float64{},
			want:   0.0,
		},
		{
			name: "single low score",
			scores: map[FeatureType]float64{
				FeatureEntropy: 0.2,
			},
			want: 0.18, // 0.2 * (1 + (0-1)*0.1) = 0.18 (没有异常时 anomalyCount=0)
		},
		{
			name: "single high score",
			scores: map[FeatureType]float64{
				FeatureEntropy: 0.9,
			},
			want: 0.9,
		},
		{
			name: "multiple anomalies",
			scores: map[FeatureType]float64{
				FeatureEntropy:                   0.9,
				FeatureHeadlessBrowser:           0.95,
				FeatureOSPlatformContradiction:   0.8,
			},
			want: 1.0, // 应该被限制在 1.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateRiskScore(tt.scores)
			// 使用范围比较处理浮点数精度问题
			if got < tt.want - 0.001 || got > tt.want + 0.001 {
				t.Errorf("calculateRiskScore() = %v, want ~%v", got, tt.want)
			}
		})
	}
}

// TestExtractFeatureVector_HashConsistency 测试特征向量哈希一致性
func TestExtractFeatureVector_HashConsistency(t *testing.T) {
	extractor := NewBaseFeatureExtractor(nil)

	data := map[string]interface{}{
		"user_agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/133.0.0.0",
		"os":           "Windows NT 10.0",
		"platform":     "Win32",
		"is_mobile":    "false",
		"screen_width": "1920",
		"features":     "WebGL2",
	}

	// 多次提取应该产生相同的哈希
	vector1 := extractor.ExtractFeatureVector(data, nil)
	vector2 := extractor.ExtractFeatureVector(data, nil)

	if vector1.Hash != vector2.Hash {
		t.Errorf("Hash inconsistency: %s vs %s", vector1.Hash, vector2.Hash)
	}

	// 不同数据应该产生不同哈希 - 使用明显异常的数据
	data2 := map[string]interface{}{
		"user_agent":   "HeadlessChrome/120.0.0.0", // 无头浏览器
		"os":           "Windows NT 10.0",
		"platform":     "MacIntel", // 矛盾：Windows OS 但 Mac platform
		"is_mobile":    "true",
		"screen_width": "2560", // 移动设备大屏幕
		"features":     "WebGL2",
	}

	vector3 := extractor.ExtractFeatureVector(data2, nil)

	if vector1.Hash == vector3.Hash {
		t.Error("Different data produced same hash")
	}
}

// BenchmarkExtractFeature 基准测试特征提取
func BenchmarkExtractFeature(b *testing.B) {
	extractor := NewBaseFeatureExtractor(nil)
	data := map[string]interface{}{
		"user_agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/133.0.0.0",
		"os":           "Windows NT 10.0",
		"platform":     "Win32",
		"is_mobile":    "false",
		"screen_width": "1920",
		"features":     "WebGL2",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractor.ExtractFeatureVector(data, nil)
	}
}

// BenchmarkExtractFeature_Headless 基准测试无头浏览器检测
func BenchmarkExtractFeature_Headless(b *testing.B) {
	extractor := NewBaseFeatureExtractor(nil)
	ua := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 HeadlessChrome/120.0.0.0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractor.ExtractFeature(FeatureHeadlessBrowser, ua, nil)
	}
}
