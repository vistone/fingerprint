package profiles

import (
	"testing"

	"github.com/vistone/fingerprint/internal/errors"
)

// TestAllProfilesValid 验证所有指纹配置有效
func TestAllProfilesValid(t *testing.T) {
	if len(MappedTLSClients) == 0 {
		t.Fatal("MappedTLSClients is empty")
	}

	t.Logf("Testing %d profiles", len(MappedTLSClients))

	for name, profile := range MappedTLSClients {
		t.Run(name, func(t *testing.T) {
			// 测试 GetClientHelloStr
			str := profile.GetClientHelloStr()
			if str == "" {
				t.Error("GetClientHelloStr() returned empty string")
			}

			// 测试 GetClientHelloId
			helloID := profile.GetClientHelloId()
			if helloID.Client == "" {
				t.Error("GetClientHelloId().Client is empty")
			}

			// 测试 GetSettings
			settings := profile.GetSettings()
			if len(settings) == 0 {
				t.Error("GetSettings() returned empty map")
			}

			// 测试 GetSettingsOrder
			settingsOrder := profile.GetSettingsOrder()
			if len(settingsOrder) == 0 {
				t.Error("GetSettingsOrder() returned empty slice")
			}

			// 测试 GetPseudoHeaderOrder
			pseudoHeaders := profile.GetPseudoHeaderOrder()
			if len(pseudoHeaders) == 0 {
				t.Error("GetPseudoHeaderOrder() returned empty slice")
			}

			// 测试 GetConnectionFlow
			flow := profile.GetConnectionFlow()
			if flow == 0 {
				t.Error("GetConnectionFlow() returned 0")
			}

			// 测试 GetClientHelloSpec（某些 profile 可能返回错误）
			spec, err := profile.GetClientHelloSpec()
			if err != nil {
				if !errors.IsClientHelloSpecNotImplemented(err) {
					t.Errorf("GetClientHelloSpec() unexpected error: %v", err)
				}
			} else {
				// 验证 spec 的基本有效性
				if len(spec.CipherSuites) == 0 {
					t.Error("ClientHelloSpec.CipherSuites is empty")
				}
				if len(spec.Extensions) == 0 {
					t.Error("ClientHelloSpec.Extensions is empty")
				}
			}
		})
	}
}

// TestDefaultClientProfile 验证默认指纹配置
func TestDefaultClientProfile(t *testing.T) {
	if DefaultClientProfile.GetClientHelloStr() == "" {
		t.Error("DefaultClientProfile is invalid")
	}
}

// TestChromeProfiles 验证 Chrome 系列指纹
func TestChromeProfiles(t *testing.T) {
	chromeProfiles := []string{
		"chrome_103", "chrome_104", "chrome_105", "chrome_106", "chrome_107",
		"chrome_108", "chrome_109", "chrome_110", "chrome_111", "chrome_112",
		"chrome_117", "chrome_120", "chrome_124", "chrome_131", "chrome_133",
	}

	for _, name := range chromeProfiles {
		t.Run(name, func(t *testing.T) {
			profile, ok := MappedTLSClients[name]
			if !ok {
				t.Skipf("Profile %s not found", name)
				return
			}

			helloID := profile.GetClientHelloId()
			if helloID.Client != "Chrome" {
				t.Errorf("Expected Client='Chrome', got '%s'", helloID.Client)
			}
		})
	}
}

// TestFirefoxProfiles 验证 Firefox 系列指纹
func TestFirefoxProfiles(t *testing.T) {
	firefoxProfiles := []string{
		"firefox_102", "firefox_104", "firefox_105", "firefox_106",
		"firefox_108", "firefox_110", "firefox_117", "firefox_120",
		"firefox_132", "firefox_133", "firefox_135",
	}

	for _, name := range firefoxProfiles {
		t.Run(name, func(t *testing.T) {
			profile, ok := MappedTLSClients[name]
			if !ok {
				t.Skipf("Profile %s not found", name)
				return
			}

			helloID := profile.GetClientHelloId()
			if helloID.Client != "Firefox" {
				t.Errorf("Expected Client='Firefox', got '%s'", helloID.Client)
			}
		})
	}
}

// TestSafariProfiles 验证 Safari 系列指纹
func TestSafariProfiles(t *testing.T) {
	safariProfiles := []string{
		"safari_15_6_1", "safari_16_0",
	}

	for _, name := range safariProfiles {
		t.Run(name, func(t *testing.T) {
			profile, ok := MappedTLSClients[name]
			if !ok {
				t.Skipf("Profile %s not found", name)
				return
			}

			helloID := profile.GetClientHelloId()
			if helloID.Client != "Safari" {
				t.Errorf("Expected Client='Safari', got '%s'", helloID.Client)
			}
		})
	}
}

// TestSafariIOSProfiles 验证 Safari iOS 系列指纹
func TestSafariIOSProfiles(t *testing.T) {
	safariIOSProfiles := []string{
		"safari_ios_15_5", "safari_ios_15_6", "safari_ios_16_0",
		"safari_ios_17_0", "safari_ios_18_0", "safari_ios_18_5",
	}

	for _, name := range safariIOSProfiles {
		t.Run(name, func(t *testing.T) {
			profile, ok := MappedTLSClients[name]
			if !ok {
				t.Skipf("Profile %s not found", name)
				return
			}

			helloID := profile.GetClientHelloId()
			if helloID.Client != "iOS" {
				t.Errorf("Expected Client='iOS', got '%s'", helloID.Client)
			}
		})
	}
}

// TestEdgeProfiles 验证 Edge 系列指纹
func TestEdgeProfiles(t *testing.T) {
	edgeProfiles := []string{
		"edge_99", "edge_101", "edge_120", "edge_131", "edge_133",
	}

	for _, name := range edgeProfiles {
		t.Run(name, func(t *testing.T) {
			profile, ok := MappedTLSClients[name]
			if !ok {
				t.Skipf("Profile %s not found", name)
				return
			}

			helloID := profile.GetClientHelloId()
			if helloID.Client != "Edge" {
				t.Errorf("Expected Client='Edge', got '%s'", helloID.Client)
			}
		})
	}
}

// TestProfileConsistency 验证指纹配置一致性
func TestProfileConsistency(t *testing.T) {
	for name, profile := range MappedTLSClients {
		t.Run(name, func(t *testing.T) {
			// settings 和 settingsOrder 应该一致
			settings := profile.GetSettings()
			settingsOrder := profile.GetSettingsOrder()

			if len(settings) != len(settingsOrder) {
				t.Errorf("settings count (%d) != settingsOrder count (%d)",
					len(settings), len(settingsOrder))
			}

			// pseudoHeaderOrder 应该包含 :method, :path
			pseudoHeaders := profile.GetPseudoHeaderOrder()
			hasMethod := false
			hasPath := false
			for _, h := range pseudoHeaders {
				if h == ":method" {
					hasMethod = true
				}
				if h == ":path" {
					hasPath = true
				}
			}
			if !hasMethod {
				t.Error("pseudoHeaderOrder missing ':method'")
			}
			if !hasPath {
				t.Error("pseudoHeaderOrder missing ':path'")
			}
		})
	}
}

// BenchmarkGetClientHelloSpec 基准测试：获取 ClientHelloSpec
func BenchmarkGetClientHelloSpec(b *testing.B) {
	profile := DefaultClientProfile
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := profile.GetClientHelloSpec()
		if err != nil && !errors.IsClientHelloSpecNotImplemented(err) {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetSettings 基准测试：获取 Settings
func BenchmarkGetSettings(b *testing.B) {
	profile := DefaultClientProfile
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = profile.GetSettings()
	}
}

// BenchmarkGetPseudoHeaderOrder 基准测试：获取伪头顺序
func BenchmarkGetPseudoHeaderOrder(b *testing.B) {
	profile := DefaultClientProfile
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = profile.GetPseudoHeaderOrder()
	}
}
