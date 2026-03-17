package fingerprint_test

import (
	"strings"
	"testing"

	"github.com/vistone/fingerprint/modules/core/types"
	"github.com/vistone/fingerprint/modules/defense/legacy/defense"
	"github.com/vistone/fingerprint/modules/errors"
	"github.com/vistone/fingerprint/modules/generator/noise"
	"github.com/vistone/fingerprint/modules/generator/random"
	"github.com/vistone/fingerprint/modules/profiles/legacy"
)

// Translated comment
func TestDefaultProfile(t *testing.T) {
	if profiles.DefaultClientProfile.GetClientHelloStr() == "" {
		t.Error("默认指纹的 ClientHelloStr 不能为空")
	}
}

// Translated comment
func TestMappedTLSClients(t *testing.T) {
	if len(profiles.MappedTLSClients) == 0 {
		t.Error("MappedTLSClients 不能为空")
	}

	// Translated comment
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

// Translated comment
func TestProfileMethods(t *testing.T) {
	profile := profiles.DefaultClientProfile

	// Translated comment
	str := profile.GetClientHelloStr()
	if str == "" {
		t.Error("GetClientHelloStr return空字符串")
	}

	// Translated comment
	spec, err := profile.GetClientHelloSpec()
	if err != nil {
		t.Errorf("GetClientHelloSpec returnerror: %v", err)
	}
	if len(spec.CipherSuites) == 0 {
		t.Error("CipherSuites 不能为空")
	}

	// Translated comment
	settings := profile.GetSettings()
	if settings == nil {
		t.Error("GetSettings return nil")
	}

	// Translated comment
	settingsOrder := profile.GetSettingsOrder()
	if settingsOrder == nil {
		t.Error("GetSettingsOrder return nil")
	}

	// Translated comment
	pseudoOrder := profile.GetPseudoHeaderOrder()
	if pseudoOrder == nil {
		t.Error("GetPseudoHeaderOrder return nil")
	}

	// Translated comment
	flow := profile.GetConnectionFlow()
	if flow == 0 {
		t.Error("GetConnectionFlow return 0")
	}

	// Translated comment
	helloId := profile.GetClientHelloId()
	if helloId.Str() == "" {
		t.Error("GetClientHelloId return无效的 ID")
	}
}

// Translated comment
// Translated comment
// Translated comment
func TestAllProfilesValid(t *testing.T) {
	workingCount := 0
	predefinedCount := 0

	for name, profile := range profiles.MappedTLSClients {
		t.Run(name, func(t *testing.T) {
			// Translated comment
			str := profile.GetClientHelloStr()
			if str == "" {
				t.Errorf("Profile %s 的 ClientHelloStr 为空", name)
			}

			// Translated comment
			helloId := profile.GetClientHelloId()
			hasSpecFactory := helloId.SpecFactory != nil

			spec, err := profile.GetClientHelloSpec()
			if err != nil {
				// Translated comment
				if errors.IsClientHelloSpecNotImplemented(err) {
					predefinedCount++
					t.Logf("Profile %s 使用预定义 ID，无法get Spec（这是正常的）", name)
					return
				}
				// Translated comment
				if !hasSpecFactory {
					predefinedCount++
					t.Logf("Profile %s 使用预定义 ID，无法get Spec（这是正常的）", name)
					return
				}
				t.Errorf("Profile %s 的 GetClientHelloSpec returnerror: %v", name, err)
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

// Translated comment
func TestProfileCount(t *testing.T) {
	expectedMinCount := 70 // Translated comment
	actualCount := len(profiles.MappedTLSClients)
	if actualCount < expectedMinCount {
		t.Errorf("指纹quantity %d 少于预期的minimum值 %d", actualCount, expectedMinCount)
	}
	t.Logf("当前指纹quantity: %d", actualCount)
}

// Translated comment
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

// Translated comment
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

// Translated comment
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

// Translated comment
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

// Translated comment
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

// Translated comment
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

// Translated comment
func TestGetRandomFingerprintByEdge(t *testing.T) {
	result, err := random.GetRandomFingerprintByBrowser("edge")
	if err != nil {
		t.Fatalf("get Edge 指纹failed: %v", err)
	}
	if result.UserAgent == "" {
		t.Error("UserAgent 不能为空")
	}
	if !containsCI(result.UserAgent, "Edg") {
		t.Errorf("Edge UserAgent 应包含 'Edg': %s", result.UserAgent)
	}
	t.Logf("Edge 指纹: %s, UA: %s", result.HelloClientID, result.UserAgent)
}

// Translated comment
func containsCI(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		strings.Contains(strings.ToLower(s), strings.ToLower(substr)))
}

// Translated comment
func TestAnomalyDetector(t *testing.T) {
	detector := defense.NewAnomalyDetector()

	// Translated comment
	result, err := random.GetRandomFingerprintByBrowser("chrome")
	if err != nil {
		t.Fatalf("get Chrome 指纹failed: %v", err)
	}
	normalUA := result.UserAgent

	// Translated comment
	if detector.DetectHeadlessBrowser(normalUA) {
		t.Error("真实 Chrome UA 不应被检测为无头浏览器")
	}

	// Translated comment
	headlessUAs := []string{
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 HeadlessChrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/118.0.0.0 Safari/537.36",
	}
	for _, headlessUA := range headlessUAs {
		if !detector.DetectHeadlessBrowser(headlessUA) {
			t.Errorf("Headless UA 应被检测为无头浏览器: %s", headlessUA)
		}
	}

	// Translated comment
	highEntropyData := make([]byte, 1000)
	for i := range highEntropyData {
		highEntropyData[i] = byte(i % 256)
	}
	// Translated comment
	detector.DetectAnomalies(highEntropyData)
}

// Translated comment
func TestContradictionDetector(t *testing.T) {
	detector := defense.NewContradictionDetector()

	// Translated comment
	result, err := random.GetRandomFingerprintByBrowserWithOS("chrome", types.OSWindows10)
	if err != nil {
		t.Fatalf("get指纹failed: %v", err)
	}

	// Translated comment
	consistentAttrs := map[string]string{
		"os":            "Windows NT 10.0",
		"platform":      "Win32",
		"user_agent":    result.UserAgent,
		"is_mobile":     "false",
		"screen_width":  "1920",
		"screen_height": "1080",
	}
	if detector.CheckContradictions(consistentAttrs) {
		t.Error("一致的attribute不应被检测到矛盾")
	}

	// Translated comment
	contradictAttrs := map[string]string{
		"os":       "Windows NT 10.0",
		"platform": "MacIntel",
	}
	if !detector.CheckContradictions(contradictAttrs) {
		t.Error("Windows OS 与 Mac Platform 应被检测为矛盾")
	}

	// Translated comment
	uaOSContradict := map[string]string{
		"user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/133.0.0.0",
		"os":         "Windows NT 10.0",
	}
	if !detector.CheckContradictions(uaOSContradict) {
		t.Error("Mac UA 配 Windows OS 应被检测为矛盾")
	}

	// Translated comment
	mobileScreenContradict := map[string]string{
		"is_mobile":     "true",
		"screen_width":  "3840", // Translated comment
		"screen_height": "2160",
	}
	if !detector.CheckContradictions(mobileScreenContradict) {
		t.Error("移动设备配 4K 屏幕应被检测为矛盾")
	}
}

// Translated comment
func TestPassiveRecognizer(t *testing.T) {
	recognizer := defense.NewPassiveRecognizer()

	// Translated comment
	result, err := random.GetRandomFingerprintByBrowser("chrome")
	if err != nil {
		t.Fatalf("get Chrome 指纹failed: %v", err)
	}

	// Translated comment
	headers := result.Headers.ToMap()

	recognitionResult := recognizer.RecognizeFromHeaders(headers)
	if recognitionResult.Browser != types.BrowserChrome {
		t.Errorf("应识别为 Chrome，实际为 %s", recognitionResult.Browser)
	}
	if recognitionResult.IsBot {
		t.Error("真实 Chrome 指纹不应被识别为机器人")
	}
	if recognitionResult.Confidence < 0.5 {
		t.Errorf("置信度应大于 0.5，实际为 %f", recognitionResult.Confidence)
	}
	t.Logf("识别result: 浏览器=%s, 版本=%s, OS=%s, 置信度=%.2f",
		recognitionResult.Browser, recognitionResult.BrowserVersion,
		recognitionResult.OS, recognitionResult.Confidence)

	// Translated comment
	firefoxResult, err := random.GetRandomFingerprintByBrowser("firefox")
	if err != nil {
		t.Fatalf("get Firefox 指纹failed: %v", err)
	}

	firefoxHeaders := firefoxResult.Headers.ToMap()
	firefoxRecognition := recognizer.RecognizeFromHeaders(firefoxHeaders)
	if firefoxRecognition.Browser != types.BrowserFirefox {
		t.Errorf("应识别为 Firefox，实际为 %s", firefoxRecognition.Browser)
	}
	t.Logf("Firefox 识别result: 浏览器=%s", firefoxRecognition.Browser)
}

// Translated comment
func TestNilHTTPHeadersToMap(t *testing.T) {
	var headers *types.HTTPHeaders

	result := headers.ToMap()
	if result == nil {
		t.Fatal("nil HTTPHeaders 调用 ToMap 应return非 nil map")
	}
	if len(result) != 0 {
		t.Fatalf("nil HTTPHeaders 的 ToMap result应为空，实际length: %d", len(result))
	}
}

// Translated comment
func TestNilHTTPHeadersToMapWithCustom(t *testing.T) {
	var headers *types.HTTPHeaders

	result := headers.ToMapWithCustom(map[string]string{
		"X-Test":     "ok",
		"X-Empty":    "",
		"User-Agent": "custom-ua",
	})

	if result == nil {
		t.Fatal("nil HTTPHeaders 调用 ToMapWithCustom 应return非 nil map")
	}
	if len(result) != 2 {
		t.Fatalf("ToMapWithCustom resultlength应为 2，实际: %d", len(result))
	}
	if result["X-Test"] != "ok" {
		t.Fatal("应包含 X-Test=ok")
	}
	if result["User-Agent"] != "custom-ua" {
		t.Fatal("应包含 User-Agent=custom-ua")
	}
}

// Translated comment
func TestNoiseInjector(t *testing.T) {
	config := noise.DefaultNoiseConfig
	injector := noise.NewNoiseInjector(config)

	// Translated comment
	canvasNoise := injector.GenerateCanvasNoise()
	if canvasNoise == nil {
		t.Fatal("Canvas 噪声不能为 nil")
	}

	// Translated comment
	audioNoise := injector.GenerateAudioNoise()
	if audioNoise == nil {
		t.Fatal("Audio 噪声不能为 nil")
	}
	if audioNoise.NoiseLevel < 0 || audioNoise.NoiseLevel > 0.02 {
		t.Errorf("Audio 噪声级别超出range: %f", audioNoise.NoiseLevel)
	}

	// Translated comment
	webglNoise := injector.GenerateWebGLNoise()
	if webglNoise == nil {
		t.Fatal("WebGL 噪声不能为 nil")
	}

	// Translated comment
	profile := injector.GenerateFullProfile()
	if profile == nil {
		t.Fatal("完整噪声configuration不能为 nil")
	}
	if profile.Canvas == nil || profile.Audio == nil || profile.WebGL == nil {
		t.Error("完整噪声configured各部分不能为 nil")
	}

	t.Logf("Canvas 噪声: R=%d, G=%d, B=%d", canvasNoise.PixelOffsetR, canvasNoise.PixelOffsetG, canvasNoise.PixelOffsetB)
	t.Logf("Audio 噪声: Level=%.4f", audioNoise.NoiseLevel)
}
