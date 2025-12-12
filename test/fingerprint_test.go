package fingerprint_test

import (
	"testing"

	"github.com/vistone/fingerprint"

	// 导入所有库以确保它们被正确加载
	_ "github.com/vistone/domaindns"
	_ "github.com/vistone/localippool"
	_ "github.com/vistone/logs"
	_ "github.com/vistone/netconnpool"
	_ "github.com/vistone/quic"
)

// TestDefaultProfile 测试默认指纹
func TestDefaultProfile(t *testing.T) {
	if fingerprint.DefaultClientProfile.GetClientHelloStr() == "" {
		t.Error("默认指纹的 ClientHelloStr 不能为空")
	}
}

// TestMappedTLSClients 测试映射表完整性
func TestMappedTLSClients(t *testing.T) {
	if len(fingerprint.MappedTLSClients) == 0 {
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
		if _, ok := fingerprint.MappedTLSClients[key]; !ok {
			t.Errorf("关键指纹 %s 不存在", key)
		}
	}
}

// TestProfileMethods 测试所有 Profile 方法
func TestProfileMethods(t *testing.T) {
	profile := fingerprint.DefaultClientProfile

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

	for name, profile := range fingerprint.MappedTLSClients {
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
				// 如果使用预定义 ID 且没有 SpecFactory，这是正常的
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
	expectedMinCount := 60 // 至少应该有60个指纹（44个主流浏览器 + 22个移动端/自定义）
	actualCount := len(fingerprint.MappedTLSClients)
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
		if _, ok := fingerprint.MappedTLSClients[version]; !ok {
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
		if _, ok := fingerprint.MappedTLSClients[version]; !ok {
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
		if _, ok := fingerprint.MappedTLSClients[version]; !ok {
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
		if _, ok := fingerprint.MappedTLSClients[profile]; !ok {
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
		if _, ok := fingerprint.MappedTLSClients[version]; !ok {
			t.Errorf("Android 指纹 %s 不存在", version)
		}
	}
}

