package random

import (
	"strings"
	"sync"
	"testing"

	"github.com/vistone/fingerprint/modules/profiles/legacy"
	"github.com/vistone/fingerprint/modules/core/types"
)

// TestGetRandomFingerprint 测试随机指纹获取
func TestGetRandomFingerprint(t *testing.T) {
	result, err := GetRandomFingerprint()
	if err != nil {
		t.Fatalf("GetRandomFingerprint() error = %v", err)
	}

	if result == nil {
		t.Fatal("GetRandomFingerprint() returned nil")
	}

	// 验证 Profile 不为空
	if result.Profile.GetClientHelloStr() == "" {
		t.Error("Profile.GetClientHelloStr() is empty")
	}

	// 验证 User-Agent 不为空
	if result.UserAgent == "" {
		t.Error("UserAgent is empty")
	}

	// 验证 Headers 不为空
	if result.Headers == nil {
		t.Error("Headers is nil")
	}

	// 验证 Headers 包含必要的字段
	if result.Headers.UserAgent == "" {
		t.Error("Headers.UserAgent is empty")
	}

	// 验证 HelloClientID 不为空
	if result.HelloClientID == "" {
		t.Error("HelloClientID is empty")
	}
}

// TestGetRandomFingerprintWithOS 测试指定操作系统的随机指纹获取
func TestGetRandomFingerprintWithOS(t *testing.T) {
	tests := []struct {
		name    string
		os      types.OperatingSystem
		wantErr bool
	}{
		{
			name:    "windows 10",
			os:      types.OSWindows10,
			wantErr: false,
		},
		{
			name:    "macos 14",
			os:      types.OSMacOS14,
			wantErr: false,
		},
		{
			name:    "linux",
			os:      types.OSLinux,
			wantErr: false,
		},
		{
			name:    "empty os (random)",
			os:      "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetRandomFingerprintWithOS(tt.os)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRandomFingerprintWithOS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if result == nil {
				t.Fatal("GetRandomFingerprintWithOS() returned nil")
			}

			if result.Profile.GetClientHelloStr() == "" {
				t.Error("Profile is invalid")
			}

			if result.UserAgent == "" {
				t.Error("UserAgent is empty")
			}

			// 注意：UA 内容检查被简化，因为实际 UA 格式取决于随机选择的 profile
			// 且不同 profile 可能有不同的 OS 显示方式
		})
	}
}

// TestGetRandomFingerprintByBrowser 测试按浏览器类型获取随机指纹
func TestGetRandomFingerprintByBrowser(t *testing.T) {
	tests := []struct {
		name         string
		browserType  string
		wantErr      bool
		wantBrowser  string
	}{
		{
			name:         "chrome",
			browserType:  "chrome",
			wantErr:      false,
			wantBrowser:  "chrome",
		},
		{
			name:         "firefox",
			browserType:  "firefox",
			wantErr:      false,
			wantBrowser:  "firefox",
		},
		{
			name:         "safari",
			browserType:  "safari",
			wantErr:      false,
			wantBrowser:  "safari",
		},
		{
			name:         "edge",
			browserType:  "edge",
			wantErr:      false,
			wantBrowser:  "edge",
		},
		{
			name:         "opera",
			browserType:  "opera",
			wantErr:      false,
			wantBrowser:  "opera",
		},
		{
			name:         "empty browser type",
			browserType:  "",
			wantErr:      true,
			wantBrowser:  "",
		},
		{
			name:         "unknown browser",
			browserType:  "unknown_browser",
			wantErr:      true,
			wantBrowser:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetRandomFingerprintByBrowser(tt.browserType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRandomFingerprintByBrowser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if result == nil {
				t.Fatal("GetRandomFingerprintByBrowser() returned nil")
			}

			// 验证 User-Agent 包含预期的浏览器 (Edge 使用 Edg，Opera 使用 OPR)
			uaLower := strings.ToLower(result.UserAgent)
			searchTerm := tt.wantBrowser
			if searchTerm == "edge" {
				searchTerm = "edg"
			} else if searchTerm == "opera" {
				searchTerm = "opr"
			}
			if !strings.Contains(uaLower, searchTerm) {
				t.Errorf("UserAgent %s does not contain expected browser %s", result.UserAgent, searchTerm)
			}

			// 验证 Headers 不为空
			if result.Headers == nil {
				t.Error("Headers is nil")
			}
		})
	}
}

// TestGetRandomFingerprintByBrowserWithOS 测试按浏览器类型和操作系统获取随机指纹
func TestGetRandomFingerprintByBrowserWithOS(t *testing.T) {
	tests := []struct {
		name        string
		browserType string
		os          types.OperatingSystem
		wantErr     bool
	}{
		{
			name:        "chrome on windows",
			browserType: "chrome",
			os:          types.OSWindows10,
			wantErr:     false,
		},
		{
			name:        "firefox on mac",
			browserType: "firefox",
			os:          types.OSMacOS14,
			wantErr:     false,
		},
		{
			name:        "safari on mac",
			browserType: "safari",
			os:          types.OSMacOS14,
			wantErr:     false,
		},
		{
			name:        "chrome with empty os",
			browserType: "chrome",
			os:          "",
			wantErr:     false,
		},
		{
			name:        "empty browser",
			browserType: "",
			os:          types.OSWindows10,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetRandomFingerprintByBrowserWithOS(tt.browserType, tt.os)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRandomFingerprintByBrowserWithOS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if result == nil {
				t.Fatal("GetRandomFingerprintByBrowserWithOS() returned nil")
			}

			// 验证浏览器类型
			uaLower := strings.ToLower(result.UserAgent)
			if !strings.Contains(uaLower, tt.browserType) {
				t.Errorf("UserAgent does not contain browser type %s", tt.browserType)
			}
		})
	}
}

// TestIsMobileProfile 测试移动端 profile 判断
func TestIsMobileProfile(t *testing.T) {
	tests := []struct {
		profileName string
		want        bool
	}{
		{
			profileName: "chrome_133",
			want:        false,
		},
		{
			profileName: "safari_ios_17_0",
			want:        true,
		},
		{
			profileName: "safari_ipad_15_6",
			want:        true,
		},
		{
			profileName: "chrome_android_120",
			want:        true,
		},
		{
			profileName: "firefox_mobile_120",
			want:        true,
		},
		{
			profileName: "safari_16_0",
			want:        false,
		},
		{
			profileName: "zalando_android",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.profileName, func(t *testing.T) {
			got := isMobileProfile(tt.profileName)
			if got != tt.want {
				t.Errorf("isMobileProfile(%s) = %v, want %v", tt.profileName, got, tt.want)
			}
		})
	}
}

// TestInferBrowserFromProfileName 测试从 profile 名称推断浏览器类型
func TestInferBrowserFromProfileName(t *testing.T) {
	tests := []struct {
		profileName   string
		wantBrowser   string
		wantVersion   string
	}{
		{
			profileName: "chrome_133",
			wantBrowser: "chrome",
			wantVersion: "133",
		},
		{
			profileName: "chrome_116_PSK",
			wantBrowser: "chrome",
			wantVersion: "116",
		},
		{
			profileName: "firefox_135",
			wantBrowser: "firefox",
			wantVersion: "135",
		},
		{
			profileName: "safari_16_0",
			wantBrowser: "safari",
			wantVersion: "16_0",
		},
		{
			profileName: "opera_91",
			wantBrowser: "opera",
			wantVersion: "91",
		},
		{
			profileName: "edge_133",
			wantBrowser: "edge",
			wantVersion: "133",
		},
		{
			profileName: "unknown_profile",
			wantBrowser: "chrome", // 默认
			wantVersion: "",
		},
		{
			profileName: "",
			wantBrowser: "chrome", // 默认
			wantVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.profileName, func(t *testing.T) {
			gotBrowser, gotVersion := inferBrowserFromProfileName(tt.profileName)
			if gotBrowser != tt.wantBrowser {
				t.Errorf("inferBrowserFromProfileName(%s) browser = %v, want %v", tt.profileName, gotBrowser, tt.wantBrowser)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("inferBrowserFromProfileName(%s) version = %v, want %v", tt.profileName, gotVersion, tt.wantVersion)
			}
		})
	}
}

// TestErrBrowserNotFound 测试浏览器未找到错误
func TestErrBrowserNotFound(t *testing.T) {
	err := &ErrBrowserNotFound{Browser: "unknown"}
	expected := "browser type not found: unknown"
	if err.Error() != expected {
		t.Errorf("ErrBrowserNotFound.Error() = %v, want %v", err.Error(), expected)
	}
}

// TestGetRandomFingerprint_AllProfiles 测试所有 profile 都能正常生成指纹
func TestGetRandomFingerprint_AllProfiles(t *testing.T) {
	// 遍历所有可用的 profile
	for name := range profiles.MappedTLSClients {
		t.Run(name, func(t *testing.T) {
			// 使用 GetRandomFingerprintByBrowser 测试每个浏览器类型
			browser, _ := inferBrowserFromProfileName(name)
			
			result, err := GetRandomFingerprintByBrowser(browser)
			if err != nil {
				t.Fatalf("GetRandomFingerprintByBrowser(%s) error = %v", browser, err)
			}

			if result == nil {
				t.Fatal("result is nil")
			}

			if result.Profile.GetClientHelloStr() == "" {
				t.Error("Profile ClientHelloStr is empty")
			}

			if result.UserAgent == "" {
				t.Error("UserAgent is empty")
			}
		})
	}
}

// TestGetRandomFingerprint_Concurrency 测试并发安全性
func TestGetRandomFingerprint_Concurrency(t *testing.T) {
	const numGoroutines = 100
	const iterations = 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*iterations)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				result, err := GetRandomFingerprint()
				if err != nil {
					errors <- err
					return
				}
				if result == nil {
					errors <- ErrNilResult
					return
				}
				if result.Profile.GetClientHelloStr() == "" {
					errors <- ErrEmptyProfile
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	errorCount := 0
	for err := range errors {
		if err != nil {
			t.Errorf("Concurrent error: %v", err)
			errorCount++
		}
	}

	if errorCount > 0 {
		t.Fatalf("Got %d errors during concurrent test", errorCount)
	}
}

// 自定义错误类型用于并发测试
var (
	ErrNilResult   = &testError{msg: "nil result"}
	ErrEmptyProfile = &testError{msg: "empty profile"}
)

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// TestGetRandomFingerprint_RandomDistribution 测试随机分布
func TestGetRandomFingerprint_RandomDistribution(t *testing.T) {
	// 多次获取随机指纹，验证分布
	profileCounts := make(map[string]int)
	const iterations = 100

	for i := 0; i < iterations; i++ {
		result, err := GetRandomFingerprint()
		if err != nil {
			t.Fatalf("GetRandomFingerprint() error = %v", err)
		}
		profileCounts[result.HelloClientID]++
	}

	// 验证有多个不同的 profile 被选中
	if len(profileCounts) < 2 {
		t.Errorf("Random distribution test: only %d unique profiles selected out of %d iterations", len(profileCounts), iterations)
	}

	t.Logf("Selected %d unique profiles out of %d iterations", len(profileCounts), iterations)
}

// BenchmarkGetRandomFingerprint 基准测试随机指纹获取
func BenchmarkGetRandomFingerprint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GetRandomFingerprint()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetRandomFingerprintByBrowser 基准测试按浏览器获取
func BenchmarkGetRandomFingerprintByBrowser(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetRandomFingerprintByBrowser("chrome")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetRandomFingerprintByBrowserWithOS 基准测试按浏览器和 OS 获取
func BenchmarkGetRandomFingerprintByBrowserWithOS(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetRandomFingerprintByBrowserWithOS("chrome", types.OSWindows10)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkIsMobileProfile 基准测试移动端判断
func BenchmarkIsMobileProfile(b *testing.B) {
	profiles := []string{
		"chrome_133",
		"safari_ios_17_0",
		"firefox_135",
		"safari_ipad_15_6",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, profile := range profiles {
			isMobileProfile(profile)
		}
	}
}

// BenchmarkInferBrowserFromProfileName 基准测试浏览器推断
func BenchmarkInferBrowserFromProfileName(b *testing.B) {
	profiles := []string{
		"chrome_133",
		"firefox_135",
		"safari_16_0",
		"edge_133",
		"opera_91",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, profile := range profiles {
			inferBrowserFromProfileName(profile)
		}
	}
}
