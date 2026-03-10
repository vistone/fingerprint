package types

import (
	"testing"

	profiles "github.com/vistone/fingerprint/modules/profiles/legacy"
)

// TestBrowserTypeConstants tests all BrowserType constant values are correct
func TestBrowserTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		browser  BrowserType
		expected string
	}{
		{
			name:     "Chrome browser type should be chrome",
			browser:  BrowserChrome,
			expected: "chrome",
		},
		{
			name:     "Firefox browser type should be firefox",
			browser:  BrowserFirefox,
			expected: "firefox",
		},
		{
			name:     "Safari browser type should be safari",
			browser:  BrowserSafari,
			expected: "safari",
		},
		{
			name:     "Opera browser type should be opera",
			browser:  BrowserOpera,
			expected: "opera",
		},
		{
			name:     "Edge browser type should be edge",
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

// TestOperatingSystemConstants tests all OperatingSystem constant values are correct
func TestOperatingSystemConstants(t *testing.T) {
	tests := []struct {
		name     string
		os       OperatingSystem
		expected string
	}{
		{
			name:     "Windows 10 operating system",
			os:       OSWindows10,
			expected: "Windows NT 10.0; Win64; x64",
		},
		{
			name:     "Windows 11 operating system",
			os:       OSWindows11,
			expected: "Windows NT 10.0; Win64; x64",
		},
		{
			name:     "macOS 13 operating system",
			os:       OSMacOS13,
			expected: "Macintosh; Intel Mac OS X 13_0_0",
		},
		{
			name:     "macOS 14 operating system",
			os:       OSMacOS14,
			expected: "Macintosh; Intel Mac OS X 14_0_0",
		},
		{
			name:     "macOS 15 operating system",
			os:       OSMacOS15,
			expected: "Macintosh; Intel Mac OS X 15_0_0",
		},
		{
			name:     "Linux operating system",
			os:       OSLinux,
			expected: "X11; Linux x86_64",
		},
		{
			name:     "Ubuntu Linux operating system",
			os:       OSLinuxUbuntu,
			expected: "X11; Linux x86_64",
		},
		{
			name:     "Debian Linux operating system",
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

// TestOperatingSystemsSlice tests OperatingSystems slice content is correct
func TestOperatingSystemsSlice(t *testing.T) {
	tests := []struct {
		name          string
		expectedLen   int
		expectedItems []OperatingSystem
	}{
		{
			name:        "OperatingSystems slice should contain 8 unique operating systems",
			expectedLen: 8,
			expectedItems: []OperatingSystem{
				OSWindows10,
				OSMacOS13,
				OSMacOS14,
				OSMacOS15,
				OSLinux,
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

// TestFingerprintResult tests FingerprintResult struct creation and usage
func TestFingerprintResult(t *testing.T) {
	tests := []struct {
		name           string
		result         FingerprintResult
		wantUserAgent  string
		wantClientID   string
		wantProfileNil bool
	}{
		{
			name: "create complete FingerprintResult",
			result: FingerprintResult{
				Profile:       profiles.Chrome_120,
				UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.0",
				HelloClientID: "chrome_120",
				Headers: &HTTPHeaders{
					Accept:        "text/html",
					UserAgent:     "Mozilla/5.0",
					SecCHUA:       "\"Chromium\";v=\"120\"",
					SecCHUAMobile: "?0",
				},
			},
			wantUserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.0",
			wantClientID:   "chrome_120",
			wantProfileNil: false,
		},
		{
			name: "create FingerprintResult with basic fields only",
			result: FingerprintResult{
				UserAgent:     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
				HelloClientID: "safari_17",
			},
			wantUserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
			wantClientID:   "safari_17",
			wantProfileNil: false, // ClientProfile is value type, not nil by default
		},
		{
			name:           "create empty FingerprintResult",
			result:         FingerprintResult{},
			wantUserAgent:  "",
			wantClientID:   "",
			wantProfileNil: false, // ClientProfile is value type, not nil by default
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
			// ClientProfile is struct value type, cannot directly compare if "empty"
			// only need to verify other fields are correct
		})
	}
}

// TestHTTPHeadersCreation tests HTTPHeaders struct creation
func TestHTTPHeadersCreation(t *testing.T) {
	tests := []struct {
		name          string
		headers       HTTPHeaders
		wantAccept    string
		wantLang      string
		wantEncoding  string
		wantCustomLen int
	}{
		{
			name: "create complete HTTPHeaders",
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
			name:          "create empty HTTPHeaders",
			headers:       HTTPHeaders{},
			wantAccept:    "",
			wantLang:      "",
			wantEncoding:  "",
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

// TestHTTPHeadersClone tests HTTPHeaders.Clone method
func TestHTTPHeadersClone(t *testing.T) {
	tests := []struct {
		name     string
		headers  *HTTPHeaders
		wantNil  bool
		wantDeep bool
	}{
		{
			name: "clone HTTPHeaders containing all fields",
			headers: &HTTPHeaders{
				Accept:         "text/html",
				AcceptLanguage: "en-US",
				Custom: map[string]string{
					"Cookie":    "session=abc",
					"X-API-Key": "key123",
				},
			},
			wantNil:  false,
			wantDeep: true,
		},
		{
			name:     "cloning nil HTTPHeaders should return nil",
			headers:  nil,
			wantNil:  true,
			wantDeep: false,
		},
		{
			name: "clone empty HTTPHeaders",
			headers: &HTTPHeaders{
				Accept:         "",
				AcceptLanguage: "",
				Custom:         nil,
			},
			wantNil:  false,
			wantDeep: false,
		},
		{
			name: "clone HTTPHeaders with Custom only",
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

			// verify values are equal
			if cloned.Accept != tt.headers.Accept {
				t.Errorf("Clone().Accept = %q, want %q", cloned.Accept, tt.headers.Accept)
			}
			if cloned.AcceptLanguage != tt.headers.AcceptLanguage {
				t.Errorf("Clone().AcceptLanguage = %q, want %q", cloned.AcceptLanguage, tt.headers.AcceptLanguage)
			}

			// verify deep copy
			if tt.wantDeep {
				if cloned.Custom == nil {
					t.Error("Clone().Custom = nil, want non-nil")
				}
				if tt.headers.Custom != nil {
					// modify original value, verify clone won't change
					originalCookie := tt.headers.Custom["Cookie"]
					tt.headers.Custom["Cookie"] = "modified"
					if cloned.Custom["Cookie"] == "modified" {
						t.Error("Clone() did not create deep copy of Custom map")
					}
					// restore original value
					tt.headers.Custom["Cookie"] = originalCookie
				}
			}
		})
	}
}

// TestHTTPHeadersSet tests HTTPHeaders.Set method
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
			name: "set header on nil receiver",
			setup: func() *HTTPHeaders {
				return nil
			},
			key:        "Cookie",
			value:      "session=abc",
			wantValue:  "",
			wantExists: false,
		},
		{
			name: "set new header on empty Custom map",
			setup: func() *HTTPHeaders {
				return &HTTPHeaders{}
			},
			key:        "Cookie",
			value:      "session=abc",
			wantValue:  "session=abc",
			wantExists: true,
		},
		{
			name: "update existing header value",
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
			name: "setting empty value should delete header",
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
			name: "add multiple different headers",
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

// TestHTTPHeadersSetHeaders tests HTTPHeaders.SetHeaders method
func TestHTTPHeadersSetHeaders(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *HTTPHeaders
		headers    map[string]string
		wantCount  int
		wantValues map[string]string
	}{
		{
			name: "batch set headers on nil receiver",
			setup: func() *HTTPHeaders {
				return nil
			},
			headers: map[string]string{
				"Cookie": "session=abc",
			},
			wantCount: 0,
		},
		{
			name: "batch set headers on empty Custom map",
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
			name: "merge to existing Custom map",
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
			name: "empty value should delete header",
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

// TestHTTPHeadersMerge tests HTTPHeaders.Merge method
func TestHTTPHeadersMerge(t *testing.T) {
	tests := []struct {
		name          string
		headers       *HTTPHeaders
		customHeaders map[string]string
		wantNil       bool
		checkStandard map[string]string
		checkCustom   map[string]string
	}{
		{
			name:          "merge nil receiver should return nil",
			headers:       nil,
			customHeaders: map[string]string{"Cookie": "session=abc"},
			wantNil:       true,
		},
		{
			name: "merge standard headers",
			headers: &HTTPHeaders{
				Accept:         "text/html",
				AcceptLanguage: "en-US",
			},
			customHeaders: map[string]string{
				"Accept":          "application/json",
				"Accept-Language": "zh-CN",
				"Cookie":          "session=abc",
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
			name: "merge empty customHeaders",
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
			name: "merge all standard headers fields",
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
				"Custom-Header":             "custom-value",
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
			name: "empty values should be skipped",
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

			// check standard fields
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

			// check custom fields
			for key, want := range tt.checkCustom {
				if got, exists := merged.Custom[key]; !exists || got != want {
					t.Errorf("Merged Custom[%q] = %q, want %q", key, got, want)
				}
			}
		})
	}
}

// TestHTTPHeadersToMap tests HTTPHeaders.ToMap method
func TestHTTPHeadersToMap(t *testing.T) {
	tests := []struct {
		name       string
		headers    *HTTPHeaders
		wantCount  int
		wantValues map[string]string
	}{
		{
			name: "convert complete HTTPHeaders",
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
			wantCount: 13, // 12 standard fields + 1 Custom field
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
			name:       "nil headers should return empty map",
			headers:    nil,
			wantCount:  0,
			wantValues: map[string]string{},
		},
		{
			name: "empty value fields should not be included in map",
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
			name: "empty values in Custom should not be included in map",
			headers: &HTTPHeaders{
				Accept: "text/html",
				Custom: map[string]string{
					"Cookie":      "session=abc",
					"EmptyHeader": "",
				},
			},
			wantCount: 2,
			wantValues: map[string]string{
				"Accept": "text/html",
				"Cookie": "session=abc",
			},
		},
		{
			name: "headers with Custom only",
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

// TestHTTPHeadersToMapWithCustom tests HTTPHeaders.ToMapWithCustom method
func TestHTTPHeadersToMapWithCustom(t *testing.T) {
	tests := []struct {
		name          string
		headers       *HTTPHeaders
		customHeaders map[string]string
		wantCount     int
		wantValues    map[string]string
	}{
		{
			name:          "mergenil headersandcustom headers",
			headers:       nil,
			customHeaders: map[string]string{"Cookie": "session=abc"},
			wantCount:     1,
			wantValues:    map[string]string{"Cookie": "session=abc"},
		},
		{
			name: "merge headers and override standard fields",
			headers: &HTTPHeaders{
				Accept:         "text/html",
				AcceptLanguage: "en-US",
			},
			customHeaders: map[string]string{
				"Accept": "application/json",
				"Cookie": "session=abc",
			},
			wantCount: 3,
			wantValues: map[string]string{
				"Accept":          "application/json",
				"Accept-Language": "en-US",
				"Cookie":          "session=abc",
			},
		},
		{
			name: "custom headers take priority over Custom field",
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
			name: "empty values should be filtered",
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
			name: "merge empty customHeaders",
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

// TestUserAgentTemplate tests UserAgentTemplate struct creation and usage
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
			name: "create Chrome desktop UA template",
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
			name: "create Safari mobile UA template",
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
			name:           "create empty UserAgentTemplate",
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

// TestIntegration tests integration usage between types
func TestIntegration(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "complete fingerprint result creation and usage flow",
			test: func(t *testing.T) {
				// create HTTPHeaders
				headers := &HTTPHeaders{
					Accept:         "text/html,application/xhtml+xml",
					AcceptLanguage: "en-US,en;q=0.9",
					AcceptEncoding: "gzip, deflate, br",
					UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0",
					SecCHUA:        "\"Google Chrome\";v=\"120\"",
				}

				// set custom headers
				headers.Set("Cookie", "session=abc123")
				headers.SetHeaders(map[string]string{
					"Authorization": "Bearer token456",
				})

				// create FingerprintResult
				result := FingerprintResult{
					Profile:       profiles.Chrome_120,
					UserAgent:     headers.UserAgent,
					HelloClientID: "chrome_120",
					Headers:       headers,
				}

				// verify result - Profile contains incomparable fields, just verify Headers is not empty
				if result.Headers == nil {
					t.Error("Headers should not be nil")
				}

				// convert to map
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
			name: "create UA template with different OS",
			test: func(t *testing.T) {
				uaTemplate := UserAgentTemplate{
					Browser:    BrowserFirefox,
					Version:    "121",
					Template:   "Mozilla/5.0 (%s; rv:121.0) Gecko/20100101 Firefox/121.0",
					OSRequired: true,
				}

				// use different operating systems
				for _, os := range OperatingSystems {
					ua := uaTemplate.Template
					if uaTemplate.OSRequired {
						_ = ua + string(os) // simulate formatting
					}
					if os == "" {
						t.Error("Operating system should not be empty")
					}
				}
			},
		},
		{
			name: "Merge does not modify original object",
			test: func(t *testing.T) {
				original := &HTTPHeaders{
					Accept: "text/html",
					Custom: map[string]string{
						"Cookie": "original",
					},
				}

				// clone for verification
				cloned := original.Clone()

				// executemerge
				merged := original.Merge(map[string]string{
					"Accept": "application/json",
					"Cookie": "modified",
				})

				// verify original object is not modified
				if original.Accept != cloned.Accept {
					t.Error("Original Accept was modified")
				}
				if original.Custom["Cookie"] != cloned.Custom["Cookie"] {
					t.Error("Original Cookie was modified")
				}

				// verifymergeresult
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
