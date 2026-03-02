package fingerprint_test

import (
	"strings"
	"testing"

	"github.com/vistone/fingerprint/generator/noise"
	"github.com/vistone/fingerprint/generator/random"
	"github.com/vistone/fingerprint/internal/errors"
	"github.com/vistone/fingerprint/profiles"
	"github.com/vistone/fingerprint/security/defense"
	"github.com/vistone/fingerprint/types"
)

// TestDefaultProfile 测试默认指纹
func TestDefaultProfile(t *testing.T) {
	if profiles.DefaultClientProfile.GetClientHelloStr() == "" {
		t.Error("默认指纹的 ClientHelloStr 不能为空")
	}
}

// TestMappedTLSClients 测试映射表完整性
func TestMappedTLSClients(t *testing.T) {
	if len(profiles.MappedTLSClients) == 0 {
		t.Error("MappedTLSClients 不能为空")
	}

	// 测试几个关键指纹是否存在
	keyProfiles := []string{
		"chrome_133",
		"chrome_120",
		"firefox_135",
		"safari_16_0",
		"opera_91",
	}

	for _, key := range keyProfiles {
		if _, ok := profiles.MappedTLSClients[key]; !ok {
			t.Errorf("关键指纹 %s 不存在", key)
		}
	}
}

// TestProfileMethods 测试所有 Profile 方法
func TestProfileMethods(t *testing.T) {
	profile := profiles.DefaultClientProfile

	// 测试 GetClientHelloStr
	str := profile.GetClientHelloStr()
	if str == "" {
		t.Error("GetClientHelloStr 返回空字符串")
	}

	// 测试 GetClientHelloSpec
	spec, err := profile.GetClientHelloSpec()
	if err != nil {
		t.Errorf("GetClientHelloSpec 返回错误: %v", err)
	}
	if len(spec.CipherSuites) == 0 {
		t.Error("CipherSuites 不能为空")
	}

	// 测试 GetSettings
	settings := profile.GetSettings()
	if settings == nil {
		t.Error("GetSettings 返回 nil")
	}

	// 测试 GetSettingsOrder
	settingsOrder := profile.GetSettingsOrder()
	if settingsOrder == nil {
		t.Error("GetSettingsOrder 返回 nil")
	}

	// 测试 GetPseudoHeaderOrder
	pseudoOrder := profile.GetPseudoHeaderOrder()
	if pseudoOrder == nil {
		t.Error("GetPseudoHeaderOrder 返回 nil")
	}

	// 测试 GetConnectionFlow
	flow := profile.GetConnectionFlow()
	if flow == 0 {
		t.Error("GetConnectionFlow 返回 0")
	}

	// 测试 GetClientHelloId
	helloId := profile.GetClientHelloId()
	if helloId.Str() == "" {
		t.Error("GetClientHelloId 返回无效的 ID")
	}
}

// TestAllProfilesValid 测试所有 profiles 是否有效
// 注意：某些使用预定义 tls.ClientHelloID 的 profiles 可能无法获取 Spec
// 这是正常的，因为它们依赖于 utls 库的实现
func TestAllProfilesValid(t *testing.T) {
	workingCount := 0
	predefinedCount := 0

	for name, profile := range profiles.MappedTLSClients {
		t.Run(name, func(t *testing.T) {
			// 测试每个 profile 的基本方法
			str := profile.GetClientHelloStr()
			if str == "" {
				t.Errorf("Profile %s 的 ClientHelloStr 为空", name)
			}

			// 检查是否有 SpecFactory
			helloId := profile.GetClientHelloId()
			hasSpecFactory := helloId.SpecFactory != nil

			spec, err := profile.GetClientHelloSpec()
			if err != nil {
				// 预定义 ID 的正常情况：未实现 ClientHelloSpec
				if errors.IsClientHelloSpecNotImplemented(err) {
					predefinedCount++
					t.Logf("Profile %s 使用预定义 ID，无法获取 Spec（这是正常的）", name)
					return
				}
				// 如果使用预定义 ID 且没有 SpecFactory，这也是正常的
				if !hasSpecFactory {
					predefinedCount++
					t.Logf("Profile %s 使用预定义 ID，无法获取 Spec（这是正常的）", name)
					return
				}
				t.Errorf("Profile %s 的 GetClientHelloSpec 返回错误: %v", name, err)
				return
			}

			if len(spec.CipherSuites) == 0 {
				t.Errorf("Profile %s 的 CipherSuites 为空", name)
				return
			}

			workingCount++

			settings := profile.GetSettings()
			if settings == nil {
				t.Errorf("Profile %s 的 Settings 为 nil", name)
			}
		})
	}

	t.Logf("正常工作的 profiles: %d", workingCount)
	t.Logf("使用预定义 ID 的 profiles: %d", predefinedCount)
}

// TestProfileCount 测试指纹数量
func TestProfileCount(t *testing.T) {
	expectedMinCount := 70 // 至少应该有70个指纹（含Edge系列）
	actualCount := len(profiles.MappedTLSClients)
	if actualCount < expectedMinCount {
		t.Errorf("指纹数量 %d 少于预期的最小值 %d", actualCount, expectedMinCount)
	}
	t.Logf("当前指纹数量: %d", actualCount)
}

// TestChromeProfiles 测试 Chrome 系列指纹
func TestChromeProfiles(t *testing.T) {
	chromeVersions := []string{
		"chrome_103", "chrome_104", "chrome_105", "chrome_106",
		"chrome_107", "chrome_108", "chrome_109", "chrome_110",
		"chrome_111", "chrome_112", "chrome_117", "chrome_120",
		"chrome_124", "chrome_131", "chrome_133",
	}

	for _, version := range chromeVersions {
		if _, ok := profiles.MappedTLSClients[version]; !ok {
			t.Errorf("Chrome 指纹 %s 不存在", version)
		}
	}
}

// TestFirefoxProfiles 测试 Firefox 系列指纹
func TestFirefoxProfiles(t *testing.T) {
	firefoxVersions := []string{
		"firefox_102", "firefox_104", "firefox_105", "firefox_106",
		"firefox_108", "firefox_110", "firefox_117", "firefox_120",
		"firefox_123", "firefox_132", "firefox_133", "firefox_135",
	}

	for _, version := range firefoxVersions {
		if _, ok := profiles.MappedTLSClients[version]; !ok {
			t.Errorf("Firefox 指纹 %s 不存在", version)
		}
	}
}

// TestSafariProfiles 测试 Safari 系列指纹
func TestSafariProfiles(t *testing.T) {
	safariVersions := []string{
		"safari_15_6_1", "safari_16_0", "safari_ipad_15_6",
		"safari_ios_15_5", "safari_ios_15_6", "safari_ios_16_0",
		"safari_ios_17_0", "safari_ios_18_0", "safari_ios_18_5",
	}

	for _, version := range safariVersions {
		if _, ok := profiles.MappedTLSClients[version]; !ok {
			t.Errorf("Safari 指纹 %s 不存在", version)
		}
	}
}

// TestMobileProfiles 测试移动端指纹
func TestMobileProfiles(t *testing.T) {
	mobileProfiles := []string{
		"zalando_android_mobile", "zalando_ios_mobile",
		"nike_ios_mobile", "nike_android_mobile",
		"mms_ios", "mms_ios_2", "mms_ios_3",
		"mesh_ios", "mesh_ios_2",
		"mesh_android", "mesh_android_2",
		"confirmed_ios", "confirmed_android", "confirmed_android_2",
		"okhttp4_android_7", "okhttp4_android_8", "okhttp4_android_9",
		"okhttp4_android_10", "okhttp4_android_11", "okhttp4_android_12", "okhttp4_android_13",
		"cloudflare_custom",
	}

	for _, profile := range mobileProfiles {
		if _, ok := profiles.MappedTLSClients[profile]; !ok {
			t.Errorf("移动端指纹 %s 不存在", profile)
		}
	}
}

// TestAndroidProfiles 测试 Android 指纹
func TestAndroidProfiles(t *testing.T) {
	androidVersions := []string{
		"okhttp4_android_7", "okhttp4_android_8", "okhttp4_android_9",
		"okhttp4_android_10", "okhttp4_android_11", "okhttp4_android_12",
		"okhttp4_android_13",
	}

	for _, version := range androidVersions {
		if _, ok := profiles.MappedTLSClients[version]; !ok {
			t.Errorf("Android 指纹 %s 不存在", version)
		}
	}
}

// TestEdgeProfiles 测试 Edge 系列指纹
func TestEdgeProfiles(t *testing.T) {
	edgeVersions := []string{
		"edge_99", "edge_101", "edge_120", "edge_131", "edge_133",
	}

	for _, version := range edgeVersions {
		if _, ok := profiles.MappedTLSClients[version]; !ok {
			t.Errorf("Edge 指纹 %s 不存在", version)
		}
	}
}

// TestGetRandomFingerprintByEdge 测试按 Edge 浏览器获取指纹
func TestGetRandomFingerprintByEdge(t *testing.T) {
	result, err := random.GetRandomFingerprintByBrowser("edge")
	if err != nil {
		t.Fatalf("获取 Edge 指纹失败: %v", err)
	}
	if result.UserAgent == "" {
		t.Error("UserAgent 不能为空")
	}
	if !containsCI(result.UserAgent, "Edg") {
		t.Errorf("Edge UserAgent 应包含 'Edg': %s", result.UserAgent)
	}
	t.Logf("Edge 指纹: %s, UA: %s", result.HelloClientID, result.UserAgent)
}

// containsCI 大小写不敏感的字符串包含检查
func containsCI(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		strings.Contains(strings.ToLower(s), strings.ToLower(substr)))
}

// TestAnomalyDetector 测试异常检测器
func TestAnomalyDetector(t *testing.T) {
	detector := defense.NewAnomalyDetector()

	// 测试正常数据
	normalData := []byte("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	if detector.DetectAnomalies(normalData) {
		t.Error("正常数据不应被检测为异常")
	}

	// 测试无头浏览器
	headlessUA := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 HeadlessChrome/120"
	if !detector.DetectHeadlessBrowser(headlessUA) {
		t.Error("HeadlessChrome 应被检测为无头浏览器")
	}

	// 测试正常 UA
	normalUA := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/133"
	if detector.DetectHeadlessBrowser(normalUA) {
		t.Error("正常 UA 不应被检测为无头浏览器")
	}
}

// TestContradictionDetector 测试矛盾检测器
func TestContradictionDetector(t *testing.T) {
	detector := defense.NewContradictionDetector()

	// 测试无矛盾
	attrs := map[string]string{
		"os":       "Windows NT 10.0",
		"platform": "Win32",
	}
	if detector.CheckContradictions(attrs) {
		t.Error("无矛盾的属性不应被检测到矛盾")
	}

	// 测试 OS/Platform 矛盾
	contradictAttrs := map[string]string{
		"os":       "Windows NT 10.0",
		"platform": "MacIntel",
	}
	if !detector.CheckContradictions(contradictAttrs) {
		t.Error("Windows OS 与 Mac Platform 应被检测为矛盾")
	}
}

// TestPassiveRecognizer 测试被动识别器
func TestPassiveRecognizer(t *testing.T) {
	recognizer := defense.NewPassiveRecognizer()

	// 测试 Chrome UA 识别
	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
		"Sec-CH-UA":       `"Google Chrome";v="133", "Chromium";v="133"`,
	}
	result := recognizer.RecognizeFromHeaders(headers)
	if result.Browser != types.BrowserChrome {
		t.Errorf("应识别为 Chrome，实际为 %s", result.Browser)
	}
	if result.IsBot {
		t.Error("正常 Chrome UA 不应被识别为机器人")
	}
	if result.Confidence < 0.5 {
		t.Errorf("置信度应大于 0.5，实际为 %f", result.Confidence)
	}
	t.Logf("识别结果: 浏览器=%s, 版本=%s, OS=%s, 置信度=%.2f", result.Browser, result.BrowserVersion, result.OS, result.Confidence)
}

// TestNilHTTPHeadersToMap 测试 nil HTTPHeaders 的 ToMap 安全性
func TestNilHTTPHeadersToMap(t *testing.T) {
	var headers *types.HTTPHeaders

	result := headers.ToMap()
	if result == nil {
		t.Fatal("nil HTTPHeaders 调用 ToMap 应返回非 nil map")
	}
	if len(result) != 0 {
		t.Fatalf("nil HTTPHeaders 的 ToMap 结果应为空，实际长度: %d", len(result))
	}
}

// TestNilHTTPHeadersToMapWithCustom 测试 nil HTTPHeaders 的 ToMapWithCustom 合并行为
func TestNilHTTPHeadersToMapWithCustom(t *testing.T) {
	var headers *types.HTTPHeaders

	result := headers.ToMapWithCustom(map[string]string{
		"X-Test":     "ok",
		"X-Empty":    "",
		"User-Agent": "custom-ua",
	})

	if result == nil {
		t.Fatal("nil HTTPHeaders 调用 ToMapWithCustom 应返回非 nil map")
	}
	if len(result) != 2 {
		t.Fatalf("ToMapWithCustom 结果长度应为 2，实际: %d", len(result))
	}
	if result["X-Test"] != "ok" {
		t.Fatal("应包含 X-Test=ok")
	}
	if result["User-Agent"] != "custom-ua" {
		t.Fatal("应包含 User-Agent=custom-ua")
	}
}

// TestNoiseInjector 测试噪声注入器
func TestNoiseInjector(t *testing.T) {
	config := noise.DefaultNoiseConfig
	injector := noise.NewNoiseInjector(config)

	// 测试 Canvas 噪声
	canvasNoise := injector.GenerateCanvasNoise()
	if canvasNoise == nil {
		t.Fatal("Canvas 噪声不能为 nil")
	}

	// 测试 Audio 噪声
	audioNoise := injector.GenerateAudioNoise()
	if audioNoise == nil {
		t.Fatal("Audio 噪声不能为 nil")
	}
	if audioNoise.NoiseLevel < 0 || audioNoise.NoiseLevel > 0.02 {
		t.Errorf("Audio 噪声级别超出范围: %f", audioNoise.NoiseLevel)
	}

	// 测试 WebGL 噪声
	webglNoise := injector.GenerateWebGLNoise()
	if webglNoise == nil {
		t.Fatal("WebGL 噪声不能为 nil")
	}

	// 测试完整配置
	profile := injector.GenerateFullProfile()
	if profile == nil {
		t.Fatal("完整噪声配置不能为 nil")
	}
	if profile.Canvas == nil || profile.Audio == nil || profile.WebGL == nil {
		t.Error("完整噪声配置的各部分不能为 nil")
	}

	t.Logf("Canvas 噪声: R=%d, G=%d, B=%d", canvasNoise.PixelOffsetR, canvasNoise.PixelOffsetG, canvasNoise.PixelOffsetB)
	t.Logf("Audio 噪声: Level=%.4f", audioNoise.NoiseLevel)
}
