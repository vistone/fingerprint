// Package profiles 测试
package profiles

import (
	"testing"

	"github.com/vistone/fingerprint/modules/core"
)

func TestProfileRegistry(t *testing.T) {
	// 使用新的注册表实例进行测试
	reg := NewProfileRegistry()

	profile := ClientProfile{
		ID:          "test_profile",
		Name:        "Test Profile",
		BrowserType: core.BrowserChrome,
	}

	// 注册
	reg.Register(profile)

	// 获取
	got, ok := reg.Get("test_profile")
	if !ok {
		t.Fatal("should find registered profile")
	}
	if got.ID != "test_profile" {
		t.Errorf("ID mismatch: got %s, want test_profile", got.ID)
	}

	// 不存在的配置
	_, ok = reg.Get("nonexistent")
	if ok {
		t.Error("should not find unregistered profile")
	}
}

func TestGetByBrowser(t *testing.T) {
	reg := NewProfileRegistry()

	// 注册多个 Chrome 配置
	reg.Register(ClientProfile{ID: "chrome_1", BrowserType: core.BrowserChrome})
	reg.Register(ClientProfile{ID: "chrome_2", BrowserType: core.BrowserChrome})
	reg.Register(ClientProfile{ID: "firefox_1", BrowserType: core.BrowserFirefox})

	chromeProfiles := reg.GetByBrowser(core.BrowserChrome)
	if len(chromeProfiles) != 2 {
		t.Errorf("should have 2 Chrome profiles, got %d", len(chromeProfiles))
	}

	firefoxProfiles := reg.GetByBrowser(core.BrowserFirefox)
	if len(firefoxProfiles) != 1 {
		t.Errorf("should have 1 Firefox profile, got %d", len(firefoxProfiles))
	}
}

func TestGetByOS(t *testing.T) {
	reg := NewProfileRegistry()

	reg.Register(ClientProfile{ID: "win_1", OS: core.OSWindows10})
	reg.Register(ClientProfile{ID: "mac_1", OS: core.OSMacOS14})
	reg.Register(ClientProfile{ID: "linux_1", OS: core.OSLinux})

	winProfiles := reg.GetByOS(core.OSWindows10)
	// Windows 10 和 Windows 11 常量相同，所以可能有多个
	if len(winProfiles) < 1 {
		t.Errorf("should have at least 1 Windows profile, got %d", len(winProfiles))
	}

	macProfiles := reg.GetByOS(core.OSMacOS14)
	if len(macProfiles) != 1 {
		t.Errorf("should have 1 macOS profile, got %d", len(macProfiles))
	}
}

func TestClientProfileGetters(t *testing.T) {
	profile := ClientProfile{
		ID:          "test_id",
		BrowserType: core.BrowserChrome,
		OS:          core.OSWindows10,
		Headers: &core.HTTPHeaders{
			UserAgent: "Mozilla/5.0",
		},
	}

	if profile.GetID() != "test_id" {
		t.Error("GetID() failed")
	}
	if profile.GetBrowserType() != core.BrowserChrome {
		t.Error("GetBrowserType() failed")
	}
	if profile.GetOS() != core.OSWindows10 {
		t.Error("GetOS() failed")
	}
	if profile.GetUserAgent() != "Mozilla/5.0" {
		t.Error("GetUserAgent() failed")
	}
}

func TestDefaultRegistry(t *testing.T) {
	// 测试默认注册表（已通过 init() 加载）
	count := DefaultRegistry.Count()
	if count == 0 {
		t.Error("Default registry should have profiles loaded")
	}

	t.Logf("Default registry has %d profiles", count)
}

func TestGetRandom(t *testing.T) {
	// 测试多次随机选择
	profiles := make(map[string]bool)
	for i := 0; i < 100; i++ {
		p := GetRandom()
		profiles[p.ID] = true
	}

	// 应该有多种不同的配置被选中
	if len(profiles) < 2 {
		t.Error("GetRandom should return varied results")
	}
}

func TestGetRandomByBrowser(t *testing.T) {
	chrome := GetRandomByBrowser(core.BrowserChrome)
	if chrome.BrowserType != core.BrowserChrome {
		t.Errorf("Expected Chrome, got %s", chrome.BrowserType)
	}

	firefox := GetRandomByBrowser(core.BrowserFirefox)
	if firefox.BrowserType != core.BrowserFirefox {
		t.Errorf("Expected Firefox, got %s", firefox.BrowserType)
	}
}

func TestGetProfileCount(t *testing.T) {
	count := GetProfileCount()
	if count < 100 {
		t.Errorf("Should have at least 100 profiles, got %d", count)
	}
	t.Logf("Total profiles: %d", count)
}

func TestGetProfilesByBrowser(t *testing.T) {
	chromeProfiles := GetProfilesByBrowser(core.BrowserChrome)
	if len(chromeProfiles) < 40 {
		t.Errorf("Should have at least 40 Chrome profiles, got %d", len(chromeProfiles))
	}

	firefoxProfiles := GetProfilesByBrowser(core.BrowserFirefox)
	if len(firefoxProfiles) < 20 {
		t.Errorf("Should have at least 20 Firefox profiles, got %d", len(firefoxProfiles))
	}
}

func BenchmarkProfileRegistryGet(b *testing.B) {
	reg := NewProfileRegistry()
	reg.Register(ClientProfile{ID: "bench_profile", BrowserType: core.BrowserChrome})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reg.Get("bench_profile")
	}
}

func BenchmarkGetRandom(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetRandom()
	}
}

func BenchmarkGetByBrowser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetProfilesByBrowser(core.BrowserChrome)
	}
}
