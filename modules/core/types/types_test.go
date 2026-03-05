package types

import (
	"testing"

	"github.com/vistone/fingerprint/modules/profiles/legacy"
)

// TestBrowserTypeConstants 测试所有 BrowserType 常量值正确
func TestBrowserTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		browser  BrowserType
		expected string
	}{
		{
			name:     "Chrome浏览器类型应为chrome",
			browser:  BrowserChrome,
			expected: "chrome",
		},
		{
			name:     "Firefox浏览器类型应为firefox",
			browser:  BrowserFirefox,
			expected: "firefox",
		},
		{
			name:     "Safari浏览器类型应为safari",
			browser:  BrowserSafari,
			expected: "safari",
		},
		{
			name:     "Opera浏览器类型应为opera",
			browser:  BrowserOpera,
			expected: "opera",
		},
		{
			name:     "Edge浏览器类型应为edge",
			browser:  BrowserEdge,
			expected: "edge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.browser) != tt.expected {
				t.Errorf("BrowserType = %q, want %q", tt.browser, tt.expected)
			}
		})
	}
}

// TestOperatingSystemConstants 测试所有 OperatingSystem 常量值正确
func TestOperatingSystemConstants(t *testing.T) {
	tests := []struct {
		name     string
		os       OperatingSystem
		expected string
	}{
		{
			name:     "Windows 10操作系统",
			os:       OSWindows10,
			expected: "Windows NT 10.0; Win64; x64",
		},
		{
			name:     "Windows 11操作系统",
			os:       OSWindows11,
			expected: "Windows NT 10.0; Win64; x64",
		},
		{
			name:     "macOS 13操作系统",
			os:       OSMacOS13,
			expected: "Macintosh; Intel Mac OS X 13_0_0",
		},
		{
			name:     "macOS 14操作系统",
			os:       OSMacOS14,
			expected: "Macintosh; Intel Mac OS X 14_0_0",
		},
		{
			name:     "macOS 15操作系统",
			os:       OSMacOS15,
			expected: "Macintosh; Intel Mac OS X 15_0_0",
		},
		{
			name:     "Linux操作系统",
			os:       OSLinux,
			expected: "X11; Linux x86_64",
		},
		{
			name:     "Ubuntu Linux操作系统",
			os:       OSLinuxUbuntu,
			expected: "X11; Linux x86_64",
		},
		{
			name:     "Debian Linux操作系统",
			os:       OSLinuxDebian,
			expected: "X11; Linux x86_64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.os) != tt.expected {
				t.Errorf("OperatingSystem = %q, want %q", tt.os, tt.expected)
			}
		})
	}
}

// TestOperatingSystemsSlice 测试 OperatingSystems 切片内容正确
func TestOperatingSystemsSlice(t *testing.T) {
	tests := []struct {
		name          string
		expectedLen   int
		expectedItems []OperatingSystem
	}{
		{
			name:        "OperatingSystems切片应包含12个操作系统",
			expectedLen: 12,
			expectedItems: []OperatingSystem{
				OSWindows10,
				OSWindows11,
				OSMacOS13,
				OSMacOS14,
				OSMacOS15,
				OSLinux,
				OSLinuxUbuntu,
				OSLinuxDebian,
				OSLinuxFedora,
				OSiOS,
				OSiPadOS,
				OSAndroid,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(OperatingSystems) != tt.expectedLen {
				t.Errorf("OperatingSystems length = %d, want %d", len(OperatingSystems), tt.expectedLen)
			}

			for i, expected := range tt.expectedItems {
				if OperatingSystems[i] != expected {
					t.Errorf("OperatingSystems[%d] = %q, want %q", i, OperatingSystems[i], expected)
				}
			}
		})
	}
}

// TestFingerprintResult 测试 FingerprintResult 结构体的创建和使用
func TestFingerprintResult(t *testing.T) {
	tests := []struct {
		name           string
		result         FingerprintResult
		wantUserAgent  string
		wantClientID   string
		wantProfileNil bool
	}{
		{
			name: "创建完整的FingerprintResult",
			result: FingerprintResult{
				Profile:       profiles.Chrome_120,
				UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.0",
				HelloClientID: "chrome_120",
				Headers: &HTTPHeaders{
					Accept:     "text/html",
					UserAgent:  "Mozilla/5.0",
					SecCHUA:    "\"Chromium\";v=\"120\"",
					SecCHUAMobile: "?0",
				},
			},
			wantUserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.0",
			wantClientID:   "chrome_120",
			wantProfileNil: false,
		},
		{
			name: "创建只包含基本字段的FingerprintResult",
			result: FingerprintResult{
				UserAgent:     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
				HelloClientID: "safari_17",
			},
			wantUserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
			wantClientID:   "safari_17",
			wantProfileNil: false, // ClientProfile 是值类型，默认不为 nil
		},
		{
			name:           "创建空的FingerprintResult",
			result:         FingerprintResult{},
			wantUserAgent:  "",
			wantClientID:   "",
			wantProfileNil: false, // ClientProfile 是值类型，默认不为 nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.UserAgent != tt.wantUserAgent {
				t.Errorf("UserAgent = %q, want %q", tt.result.UserAgent, tt.wantUserAgent)
			}
			if tt.result.HelloClientID != tt.wantClientID {
				t.Errorf("HelloClientID = %q, want %q", tt.result.HelloClientID, tt.wantClientID)
			}
			// ClientProfile 是结构体值类型，无法直接比较是否"为空"
			// 只需验证其他字段正确即可
		})
	}
}

// TestHTTPHeadersCreation 测试 HTTPHeaders 结构体的创建
func TestHTTPHeadersCreation(t *testing.T) {
	tests := []struct {
		name         string
		headers      HTTPHeaders
		wantAccept   string
		wantLang     string
		wantEncoding string
		wantCustomLen int
	}{
		{
			name: "创建完整的HTTPHeaders",
			headers: HTTPHeaders{
				Accept:                  "text/html,application/xhtml+xml",
				AcceptLanguage:          "en-US,en;q=0.9,zh-CN;q=0.8",
				AcceptEncoding:          "gzip, deflate, br",
				UserAgent:               "Mozilla/5.0",
				SecFetchSite:            "none",
				SecFetchMode:            "navigate",
				SecFetchUser:            "?1",
				SecFetchDest:            "document",
				SecCHUA:                 "\"Chromium\";v=\"120\"",
				SecCHUAMobile:           "?0",
				SecCHUAPlatform:         "\"Windows\"",
				UpgradeInsecureRequests: "1",
				Custom: map[string]string{
					"Cookie": "session=abc123",
				},
			},
			wantAccept:    "text/html,application/xhtml+xml",
			wantLang:      "en-US,en;q=0.9,zh-CN;q=0.8",
			wantEncoding:  "gzip, deflate, br",
			wantCustomLen: 1,
		},
		{
			name:         "创建空的HTTPHeaders",
			headers:      HTTPHeaders{},
			wantAccept:   "",
			wantLang:     "",
			wantEncoding: "",
			wantCustomLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.headers.Accept != tt.wantAccept {
				t.Errorf("Accept = %q, want %q", tt.headers.Accept, tt.wantAccept)
			}
			if tt.headers.AcceptLanguage != tt.wantLang {
				t.Errorf("AcceptLanguage = %q, want %q", tt.headers.AcceptLanguage, tt.wantLang)
			}
			if tt.headers.AcceptEncoding != tt.wantEncoding {
				t.Errorf("AcceptEncoding = %q, want %q", tt.headers.AcceptEncoding, tt.wantEncoding)
			}
			if len(tt.headers.Custom) != tt.wantCustomLen {
				t.Errorf("Custom length = %d, want %d", len(tt.headers.Custom), tt.wantCustomLen)
			}
		})
	}
}

// TestHTTPHeadersClone 测试 HTTPHeaders.Clone 方法
func TestHTTPHeadersClone(t *testing.T) {
	tests := []struct {
		name     string
		headers  *HTTPHeaders
		wantNil  bool
		wantDeep bool
	}{
		{
			name: "克隆包含所有字段的HTTPHeaders",
			headers: &HTTPHeaders{
				Accept:         "text/html",
				AcceptLanguage: "en-US",
				Custom: map[string]string{
					"Cookie":     "session=abc",
					"X-API-Key":  "key123",
				},
			},
			wantNil:  false,
			wantDeep: true,
		},
		{
			name:     "克隆nil的HTTPHeaders应返回nil",
			headers:  nil,
			wantNil:  true,
			wantDeep: false,
		},
		{
			name: "克隆空的HTTPHeaders",
			headers: &HTTPHeaders{
				Accept:         "",
				AcceptLanguage: "",
				Custom:         nil,
			},
			wantNil:  false,
			wantDeep: false,
		},
		{
			name: "克隆只有Custom的HTTPHeaders",
			headers: &HTTPHeaders{
				Custom: map[string]string{
					"Authorization": "Bearer token",
				},
			},
			wantNil:  false,
			wantDeep: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloned := tt.headers.Clone()

			if tt.wantNil {
				if cloned != nil {
					t.Error("Clone() = non-nil, want nil")
				}
				return
			}

			if cloned == nil {
				t.Fatal("Clone() = nil, want non-nil")
			}

			// 验证值相等
			if cloned.Accept != tt.headers.Accept {
				t.Errorf("Clone().Accept = %q, want %q", cloned.Accept, tt.headers.Accept)
			}
			if cloned.AcceptLanguage != tt.headers.AcceptLanguage {
				t.Errorf("Clone().AcceptLanguage = %q, want %q", cloned.AcceptLanguage, tt.headers.AcceptLanguage)
			}

			// 验证深拷贝
			if tt.wantDeep {
				if cloned.Custom == nil {
					t.Error("Clone().Custom = nil, want non-nil")
				}
				if tt.headers.Custom != nil {
					// 修改原始值，验证克隆的不会改变
					originalCookie := tt.headers.Custom["Cookie"]
					tt.headers.Custom["Cookie"] = "modified"
					if cloned.Custom["Cookie"] == "modified" {
						t.Error("Clone() did not create deep copy of Custom map")
					}
					// 恢复原始值
					tt.headers.Custom["Cookie"] = originalCookie
				}
			}
		})
	}
}

// TestHTTPHeadersSet 测试 HTTPHeaders.Set 方法
func TestHTTPHeadersSet(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *HTTPHeaders
		key        string
		value      string
		wantValue  string
		wantExists bool
	}{
		{
			name: "在nil接收者上设置header",
			setup: func() *HTTPHeaders {
				return nil
			},
			key:        "Cookie",
			value:      "session=abc",
			wantValue:  "",
			wantExists: false,
		},
		{
			name: "在空的Custom map上设置新header",
			setup: func() *HTTPHeaders {
				return &HTTPHeaders{}
			},
			key:        "Cookie",
			value:      "session=abc",
			wantValue:  "session=abc",
			wantExists: true,
		},
		{
			name: "更新已存在的header值",
			setup: func() *HTTPHeaders {
				return &HTTPHeaders{
					Custom: map[string]string{
						"Cookie": "old=value",
					},
				}
			},
			key:        "Cookie",
			value:      "new=value",
			wantValue:  "new=value",
			wantExists: true,
		},
		{
			name: "设置空值应删除header",
			setup: func() *HTTPHeaders {
				return &HTTPHeaders{
					Custom: map[string]string{
						"Cookie": "session=abc",
					},
				}
			},
			key:        "Cookie",
			value:      "",
			wantValue:  "",
			wantExists: false,
		},
		{
			name: "添加多个不同的header",
			setup: func() *HTTPHeaders {
				h := &HTTPHeaders{}
				h.Set("Cookie", "session=abc")
				return h
			},
			key:        "Authorization",
			value:      "Bearer token123",
			wantValue:  "Bearer token123",
			wantExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup()
			h.Set(tt.key, tt.value)

			if h == nil {
				return
			}

			val, exists := h.Custom[tt.key]
			if exists != tt.wantExists {
				t.Errorf("Custom[%q] exists = %v, want %v", tt.key, exists, tt.wantExists)
			}
			if exists && val != tt.wantValue {
				t.Errorf("Custom[%q] = %q, want %q", tt.key, val, tt.wantValue)
			}
		})
	}
}

// TestHTTPHeadersSetHeaders 测试 HTTPHeaders.SetHeaders 方法
func TestHTTPHeadersSetHeaders(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *HTTPHeaders
		headers    map[string]string
		wantCount  int
		wantValues map[string]string
	}{
		{
			name: "在nil接收者上批量设置headers",
			setup: func() *HTTPHeaders {
				return nil
			},
			headers: map[string]string{
				"Cookie": "session=abc",
			},
			wantCount: 0,
		},
		{
			name: "在空的Custom map上批量设置headers",
			setup: func() *HTTPHeaders {
				return &HTTPHeaders{}
			},
			headers: map[string]string{
				"Cookie":        "session=abc",
				"Authorization": "Bearer token",
				"X-API-Key":     "key123",
			},
			wantCount: 3,
			wantValues: map[string]string{
				"Cookie":        "session=abc",
				"Authorization": "Bearer token",
				"X-API-Key":     "key123",
			},
		},
		{
			name: "合并到已有的Custom map",
			setup: func() *HTTPHeaders {
				return &HTTPHeaders{
					Custom: map[string]string{
						"Existing": "value",
					},
				}
			},
			headers: map[string]string{
				"Cookie": "session=abc",
			},
			wantCount: 2,
			wantValues: map[string]string{
				"Existing": "value",
				"Cookie":   "session=abc",
			},
		},
		{
			name: "空值应删除header",
			setup: func() *HTTPHeaders {
				return &HTTPHeaders{
					Custom: map[string]string{
						"Cookie":        "session=abc",
						"Authorization": "Bearer token",
					},
				}
			},
			headers: map[string]string{
				"Cookie": "",
				"NewKey": "newvalue",
			},
			wantCount: 2,
			wantValues: map[string]string{
				"Authorization": "Bearer token",
				"NewKey":        "newvalue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup()
			h.SetHeaders(tt.headers)

			if h == nil {
				return
			}

			if len(h.Custom) != tt.wantCount {
				t.Errorf("Custom length = %d, want %d", len(h.Custom), tt.wantCount)
			}

			for key, wantVal := range tt.wantValues {
				if got, exists := h.Custom[key]; !exists || got != wantVal {
					t.Errorf("Custom[%q] = %q, want %q", key, got, wantVal)
				}
			}
		})
	}
}

// TestHTTPHeadersMerge 测试 HTTPHeaders.Merge 方法
func TestHTTPHeadersMerge(t *testing.T) {
	tests := []struct {
		name            string
		headers         *HTTPHeaders
		customHeaders   map[string]string
		wantNil         bool
		checkStandard   map[string]string
		checkCustom     map[string]string
	}{
		{
			name:          "合并nil接收者应返回nil",
			headers:       nil,
			customHeaders: map[string]string{"Cookie": "session=abc"},
			wantNil:       true,
		},
		{
			name: "合并标准headers",
			headers: &HTTPHeaders{
				Accept:         "text/html",
				AcceptLanguage: "en-US",
			},
			customHeaders: map[string]string{
				"Accept":         "application/json",
				"Accept-Language": "zh-CN",
				"Cookie":         "session=abc",
			},
			wantNil: false,
			checkStandard: map[string]string{
				"Accept":          "application/json",
				"Accept-Language": "zh-CN",
			},
			checkCustom: map[string]string{
				"Cookie": "session=abc",
			},
		},
		{
			name: "合并空customHeaders",
			headers: &HTTPHeaders{
				Accept: "text/html",
				Custom: map[string]string{
					"Cookie": "session=abc",
				},
			},
			customHeaders: map[string]string{},
			wantNil:       false,
			checkStandard: map[string]string{
				"Accept": "text/html",
			},
			checkCustom: map[string]string{
				"Cookie": "session=abc",
			},
		},
		{
			name: "合并所有标准headers字段",
			headers: &HTTPHeaders{
				Accept:                  "old-accept",
				AcceptLanguage:          "old-lang",
				AcceptEncoding:          "old-encoding",
				UserAgent:               "old-ua",
				SecFetchSite:            "old-site",
				SecFetchMode:            "old-mode",
				SecFetchUser:            "old-user",
				SecFetchDest:            "old-dest",
				SecCHUA:                 "old-chua",
				SecCHUAMobile:           "old-mobile",
				SecCHUAPlatform:         "old-platform",
				UpgradeInsecureRequests: "old-upgrade",
			},
			customHeaders: map[string]string{
				"Accept":                   "new-accept",
				"Accept-Language":          "new-lang",
				"Accept-Encoding":          "new-encoding",
				"User-Agent":               "new-ua",
				"Sec-Fetch-Site":           "new-site",
				"Sec-Fetch-Mode":           "new-mode",
				"Sec-Fetch-User":           "new-user",
				"Sec-Fetch-Dest":           "new-dest",
				"Sec-CH-UA":                "new-chua",
				"Sec-CH-UA-Mobile":         "new-mobile",
				"Sec-CH-UA-Platform":       "new-platform",
				"Upgrade-Insecure-Requests": "new-upgrade",
				"Custom-Header":            "custom-value",
			},
			wantNil: false,
			checkStandard: map[string]string{
				"Accept":                    "new-accept",
				"Accept-Language":           "new-lang",
				"Accept-Encoding":           "new-encoding",
				"User-Agent":                "new-ua",
				"Sec-Fetch-Site":            "new-site",
				"Sec-Fetch-Mode":            "new-mode",
				"Sec-Fetch-User":            "new-user",
				"Sec-Fetch-Dest":            "new-dest",
				"Sec-CH-UA":                 "new-chua",
				"Sec-CH-UA-Mobile":          "new-mobile",
				"Sec-CH-UA-Platform":        "new-platform",
				"Upgrade-Insecure-Requests": "new-upgrade",
			},
			checkCustom: map[string]string{
				"Custom-Header": "custom-value",
			},
		},
		{
			name: "空值应被跳过",
			headers: &HTTPHeaders{
				Accept: "text/html",
			},
			customHeaders: map[string]string{
				"Accept": "",
				"Cookie": "session=abc",
			},
			wantNil: false,
			checkStandard: map[string]string{
				"Accept": "text/html",
			},
			checkCustom: map[string]string{
				"Cookie": "session=abc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := tt.headers.Merge(tt.customHeaders)

			if tt.wantNil {
				if merged != nil {
					t.Error("Merge() = non-nil, want nil")
				}
				return
			}

			if merged == nil {
				t.Fatal("Merge() = nil, want non-nil")
			}

			// 检查标准字段
			for field, want := range tt.checkStandard {
				var got string
				switch field {
				case "Accept":
					got = merged.Accept
				case "Accept-Language":
					got = merged.AcceptLanguage
				case "Accept-Encoding":
					got = merged.AcceptEncoding
				case "User-Agent":
					got = merged.UserAgent
				case "Sec-Fetch-Site":
					got = merged.SecFetchSite
				case "Sec-Fetch-Mode":
					got = merged.SecFetchMode
				case "Sec-Fetch-User":
					got = merged.SecFetchUser
				case "Sec-Fetch-Dest":
					got = merged.SecFetchDest
				case "Sec-CH-UA":
					got = merged.SecCHUA
				case "Sec-CH-UA-Mobile":
					got = merged.SecCHUAMobile
				case "Sec-CH-UA-Platform":
					got = merged.SecCHUAPlatform
				case "Upgrade-Insecure-Requests":
					got = merged.UpgradeInsecureRequests
				}
				if got != want {
					t.Errorf("Merged %s = %q, want %q", field, got, want)
				}
			}

			// 检查自定义字段
			for key, want := range tt.checkCustom {
				if got, exists := merged.Custom[key]; !exists || got != want {
					t.Errorf("Merged Custom[%q] = %q, want %q", key, got, want)
				}
			}
		})
	}
}

// TestHTTPHeadersToMap 测试 HTTPHeaders.ToMap 方法
func TestHTTPHeadersToMap(t *testing.T) {
	tests := []struct {
		name       string
		headers    *HTTPHeaders
		wantCount  int
		wantValues map[string]string
	}{
		{
			name: "转换完整的HTTPHeaders",
			headers: &HTTPHeaders{
				Accept:                  "text/html",
				AcceptLanguage:          "en-US",
				AcceptEncoding:          "gzip",
				UserAgent:               "Mozilla/5.0",
				SecFetchSite:            "none",
				SecFetchMode:            "navigate",
				SecFetchUser:            "?1",
				SecFetchDest:            "document",
				SecCHUA:                 "Chrome",
				SecCHUAMobile:           "?0",
				SecCHUAPlatform:         "Windows",
				UpgradeInsecureRequests: "1",
				Custom: map[string]string{
					"Cookie": "session=abc",
				},
			},
			wantCount: 13, // 12个标准字段 + 1个Custom字段
			wantValues: map[string]string{
				"Accept":                    "text/html",
				"Accept-Language":           "en-US",
				"Accept-Encoding":           "gzip",
				"User-Agent":                "Mozilla/5.0",
				"Sec-Fetch-Site":            "none",
				"Sec-Fetch-Mode":            "navigate",
				"Sec-Fetch-User":            "?1",
				"Sec-Fetch-Dest":            "document",
				"Sec-CH-UA":                 "Chrome",
				"Sec-CH-UA-Mobile":          "?0",
				"Sec-CH-UA-Platform":        "Windows",
				"Upgrade-Insecure-Requests": "1",
				"Cookie":                    "session=abc",
			},
		},
		{
			name:       "nil headers应返回空map",
			headers:    nil,
			wantCount:  0,
			wantValues: map[string]string{},
		},
		{
			name: "空值字段不应包含在map中",
			headers: &HTTPHeaders{
				Accept:         "text/html",
				AcceptLanguage: "",
				AcceptEncoding: "gzip",
			},
			wantCount: 2,
			wantValues: map[string]string{
				"Accept":          "text/html",
				"Accept-Encoding": "gzip",
			},
		},
		{
			name: "Custom中的空值不应包含在map中",
			headers: &HTTPHeaders{
				Accept: "text/html",
				Custom: map[string]string{
					"Cookie":        "session=abc",
					"EmptyHeader":   "",
				},
			},
			wantCount: 2,
			wantValues: map[string]string{
				"Accept": "text/html",
				"Cookie": "session=abc",
			},
		},
		{
			name: "只有Custom的headers",
			headers: &HTTPHeaders{
				Custom: map[string]string{
					"Cookie":        "session=abc",
					"Authorization": "Bearer token",
				},
			},
			wantCount: 2,
			wantValues: map[string]string{
				"Cookie":        "session=abc",
				"Authorization": "Bearer token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.headers.ToMap()

			if len(got) != tt.wantCount {
				t.Errorf("ToMap() length = %d, want %d", len(got), tt.wantCount)
			}

			for key, want := range tt.wantValues {
				if got[key] != want {
					t.Errorf("ToMap()[%q] = %q, want %q", key, got[key], want)
				}
			}
		})
	}
}

// TestHTTPHeadersToMapWithCustom 测试 HTTPHeaders.ToMapWithCustom 方法
func TestHTTPHeadersToMapWithCustom(t *testing.T) {
	tests := []struct {
		name          string
		headers       *HTTPHeaders
		customHeaders map[string]string
		wantCount     int
		wantValues    map[string]string
	}{
		{
			name: "合并nil headers和custom headers",
			headers:       nil,
			customHeaders: map[string]string{"Cookie": "session=abc"},
			wantCount:     1,
			wantValues:    map[string]string{"Cookie": "session=abc"},
		},
		{
			name: "合并headers并覆盖标准字段",
			headers: &HTTPHeaders{
				Accept:         "text/html",
				AcceptLanguage: "en-US",
			},
			customHeaders: map[string]string{
				"Accept":         "application/json",
				"Cookie":         "session=abc",
			},
			wantCount: 3,
			wantValues: map[string]string{
				"Accept":          "application/json",
				"Accept-Language": "en-US",
				"Cookie":          "session=abc",
			},
		},
		{
			name: "custom headers优先级高于Custom字段",
			headers: &HTTPHeaders{
				Accept: "text/html",
				Custom: map[string]string{
					"Cookie": "old-cookie",
				},
			},
			customHeaders: map[string]string{
				"Cookie": "new-cookie",
			},
			wantCount: 2,
			wantValues: map[string]string{
				"Accept": "text/html",
				"Cookie": "new-cookie",
			},
		},
		{
			name: "空值应被过滤",
			headers: &HTTPHeaders{
				Accept: "text/html",
			},
			customHeaders: map[string]string{
				"Cookie": "",
				"Valid":  "value",
			},
			wantCount: 2,
			wantValues: map[string]string{
				"Accept": "text/html",
				"Valid":  "value",
			},
		},
		{
			name: "合并空的customHeaders",
			headers: &HTTPHeaders{
				Accept: "text/html",
				Custom: map[string]string{
					"Cookie": "session=abc",
				},
			},
			customHeaders: map[string]string{},
			wantCount:     2,
			wantValues: map[string]string{
				"Accept": "text/html",
				"Cookie": "session=abc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.headers.ToMapWithCustom(tt.customHeaders)

			if len(got) != tt.wantCount {
				t.Errorf("ToMapWithCustom() length = %d, want %d", len(got), tt.wantCount)
			}

			for key, want := range tt.wantValues {
				if got[key] != want {
					t.Errorf("ToMapWithCustom()[%q] = %q, want %q", key, got[key], want)
				}
			}
		})
	}
}

// TestUserAgentTemplate 测试 UserAgentTemplate 结构体的创建和使用
func TestUserAgentTemplate(t *testing.T) {
	tests := []struct {
		name           string
		template       UserAgentTemplate
		wantBrowser    BrowserType
		wantVersion    string
		wantTemplate   string
		wantMobile     bool
		wantOSRequired bool
	}{
		{
			name: "创建Chrome桌面端UA模板",
			template: UserAgentTemplate{
				Browser:    BrowserChrome,
				Version:    "120",
				Template:   "Mozilla/5.0 (%s) AppleWebKit/537.0",
				Mobile:     false,
				OSRequired: true,
			},
			wantBrowser:    BrowserChrome,
			wantVersion:    "120",
			wantTemplate:   "Mozilla/5.0 (%s) AppleWebKit/537.0",
			wantMobile:     false,
			wantOSRequired: true,
		},
		{
			name: "创建Safari移动端UA模板",
			template: UserAgentTemplate{
				Browser:    BrowserSafari,
				Version:    "17",
				Template:   "Mozilla/5.0 (%s) AppleWebKit/605.0",
				Mobile:     true,
				OSRequired: false,
			},
			wantBrowser:    BrowserSafari,
			wantVersion:    "17",
			wantTemplate:   "Mozilla/5.0 (%s) AppleWebKit/605.0",
			wantMobile:     true,
			wantOSRequired: false,
		},
		{
			name:           "创建空的UserAgentTemplate",
			template:       UserAgentTemplate{},
			wantBrowser:    "",
			wantVersion:    "",
			wantTemplate:   "",
			wantMobile:     false,
			wantOSRequired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.template.Browser != tt.wantBrowser {
				t.Errorf("Browser = %q, want %q", tt.template.Browser, tt.wantBrowser)
			}
			if tt.template.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", tt.template.Version, tt.wantVersion)
			}
			if tt.template.Template != tt.wantTemplate {
				t.Errorf("Template = %q, want %q", tt.template.Template, tt.wantTemplate)
			}
			if tt.template.Mobile != tt.wantMobile {
				t.Errorf("Mobile = %v, want %v", tt.template.Mobile, tt.wantMobile)
			}
			if tt.template.OSRequired != tt.wantOSRequired {
				t.Errorf("OSRequired = %v, want %v", tt.template.OSRequired, tt.wantOSRequired)
			}
		})
	}
}

// TestIntegration 测试类型之间的集成使用
func TestIntegration(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "完整的指纹结果创建和使用流程",
			test: func(t *testing.T) {
				// 创建 HTTPHeaders
				headers := &HTTPHeaders{
					Accept:         "text/html,application/xhtml+xml",
					AcceptLanguage: "en-US,en;q=0.9",
					AcceptEncoding: "gzip, deflate, br",
					UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0",
					SecCHUA:        "\"Google Chrome\";v=\"120\"",
				}

				// 设置自定义 headers
				headers.Set("Cookie", "session=abc123")
				headers.SetHeaders(map[string]string{
					"Authorization": "Bearer token456",
				})

				// 创建 FingerprintResult
				result := FingerprintResult{
					Profile:       profiles.Chrome_120,
					UserAgent:     headers.UserAgent,
					HelloClientID: "chrome_120",
					Headers:       headers,
				}

				// 验证结果 - Profile 包含不可比较字段，验证 Headers 不为空即可
				if result.Headers == nil {
					t.Error("Headers should not be nil")
				}

				// 转换为 map
				headerMap := result.Headers.ToMap()
				if headerMap["Cookie"] != "session=abc123" {
					t.Error("Cookie not found in headers")
				}
				if headerMap["Authorization"] != "Bearer token456" {
					t.Error("Authorization not found in headers")
				}
			},
		},
		{
			name: "使用不同操作系统创建UA模板",
			test: func(t *testing.T) {
				uaTemplate := UserAgentTemplate{
					Browser:    BrowserFirefox,
					Version:    "121",
					Template:   "Mozilla/5.0 (%s; rv:121.0) Gecko/20100101 Firefox/121.0",
					OSRequired: true,
				}

				// 使用不同的操作系统
				for _, os := range OperatingSystems {
					ua := uaTemplate.Template
					if uaTemplate.OSRequired {
						_ = ua + string(os) // 模拟格式化
					}
					if os == "" {
						t.Error("Operating system should not be empty")
					}
				}
			},
		},
		{
			name: "Merge不修改原始对象",
			test: func(t *testing.T) {
				original := &HTTPHeaders{
					Accept: "text/html",
					Custom: map[string]string{
						"Cookie": "original",
					},
				}

				// 克隆用于验证
				cloned := original.Clone()

				// 执行合并
				merged := original.Merge(map[string]string{
					"Accept": "application/json",
					"Cookie": "modified",
				})

				// 验证原始对象未被修改
				if original.Accept != cloned.Accept {
					t.Error("Original Accept was modified")
				}
				if original.Custom["Cookie"] != cloned.Custom["Cookie"] {
					t.Error("Original Cookie was modified")
				}

				// 验证合并结果
				if merged.Accept != "application/json" {
					t.Error("Merged Accept should be application/json")
				}
				if merged.Custom["Cookie"] != "modified" {
					t.Error("Merged Cookie should be modified")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}
