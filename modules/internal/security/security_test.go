package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vistone/fingerprint/modules/internal/testhelpers"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	testhelpers.AssertNotNil(t, validator)
}

func TestValidator_ValidateString(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid string",
			input:   "hello world",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: false,
		},
		{
			name:    "string with null byte",
			input:   "hello\x00world",
			wantErr: true,
		},
		{
			name:    "valid utf8",
			input:   "hello world",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateString(tt.input)
			if tt.wantErr {
				testhelpers.AssertError(t, err)
			} else {
				testhelpers.AssertNoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateIP(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{
			name:    "valid IPv4",
			ip:      "192.168.1.1",
			wantErr: false,
		},
		{
			name:    "valid IPv4",
			ip:      "10.0.0.1",
			wantErr: false,
		},
		{
			name:    "invalid IP",
			ip:      "not-an-ip",
			wantErr: true,
		},
		{
			name:    "empty IP",
			ip:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateIP(tt.ip)
			if tt.wantErr {
				testhelpers.AssertError(t, err)
			} else {
				testhelpers.AssertNoError(t, err)
			}
		})
	}
}

func TestValidator_ValidatePort(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{
			name:    "valid port",
			port:    8080,
			wantErr: false,
		},
		{
			name:    "minimum port",
			port:    1,
			wantErr: false,
		},
		{
			name:    "maximum port",
			port:    65535,
			wantErr: false,
		},
		{
			name:    "invalid port 0",
			port:    0,
			wantErr: true,
		},
		{
			name:    "invalid port too high",
			port:    65536,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePort(tt.port)
			if tt.wantErr {
				testhelpers.AssertError(t, err)
			} else {
				testhelpers.AssertNoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateUserAgent(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		ua      string
		wantErr bool
	}{
		{
			name:    "valid UA",
			ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			wantErr: false,
		},
		{
			name:    "UA with script",
			ua:      "<script>alert(1)</script>",
			wantErr: true,
		},
		{
			name:    "UA with javascript",
			ua:      "javascript:alert(1)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateUserAgent(tt.ua)
			if tt.wantErr {
				testhelpers.AssertError(t, err)
			} else {
				testhelpers.AssertNoError(t, err)
			}
		})
	}
}

func TestSanitizer_SanitizeString(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal string",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "with null byte",
			input:    "hello\x00world",
			expected: "helloworld",
		},
		{
			name:     "with control chars",
			input:    "hello\x01\x02world",
			expected: "helloworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			testhelpers.AssertEqual(t, result, tt.expected)
		})
	}
}

func TestSanitizer_SanitizeHTML(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal string",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "with script tag",
			input:    "<script>alert(1)</script>",
			expected: "&lt;script&gt;alert(1)&lt;&#x2F;script&gt;",
		},
		{
			name:     "with quotes",
			input:    `"hello"`,
			expected: `&quot;hello&quot;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeHTML(tt.input)
			testhelpers.AssertEqual(t, result, tt.expected)
		})
	}
}

func TestMasker_Mask(t *testing.T) {
	masker := NewMasker()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "long string",
			input:    "password123",
			expected: "*******d123",
		},
		{
			name:     "short string",
			input:    "abc",
			expected: "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := masker.Mask(tt.input)
			testhelpers.AssertEqual(t, result, tt.expected)
		})
	}
}

func TestMasker_MaskEmail(t *testing.T) {
	masker := NewMasker()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal email",
			input:    "user@example.com",
			expected: "us**@example.com",
		},
		{
			name:     "short local part",
			input:    "a@example.com",
			expected: "*@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := masker.MaskEmail(tt.input)
			testhelpers.AssertEqual(t, result, tt.expected)
		})
	}
}

func TestRedactor_Redact(t *testing.T) {
	redactor := NewRedactor()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with password",
			input:    "password=secret123",
			expected: "[REDACTED_PASSWORD]",
		},
		{
			name:     "with token",
			input:    "token=abc123xyz",
			expected: "[REDACTED_TOKEN]",
		},
		{
			name:     "normal text",
			input:    "hello world",
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := redactor.Redact(tt.input)
			testhelpers.AssertContains(t, result, tt.expected)
		})
	}
}

func TestDataClassifier_Classify(t *testing.T) {
	classifier := NewDataClassifier()

	tests := []struct {
		name     string
		input    string
		expected SensitivityLevel
	}{
		{
			name:     "public data",
			input:    "hello world",
			expected: Public,
		},
		{
			name:     "internal data - email",
			input:    "user@example.com",
			expected: Internal,
		},
		{
			name:     "internal data - IP",
			input:    "192.168.1.1",
			expected: Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.input)
			testhelpers.AssertEqual(t, result, tt.expected)
		})
	}
}

func TestNewRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(10, 60)
	testhelpers.AssertNotNil(t, limiter)
	testhelpers.AssertEqual(t, limiter.maxReq, 10)
}

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(2, 1)

	t.Run("allow within limit", func(t *testing.T) {
		testhelpers.AssertEqual(t, limiter.Allow("client1"), true)
		testhelpers.AssertEqual(t, limiter.Allow("client1"), true)
	})

	t.Run("block over limit", func(t *testing.T) {
		testhelpers.AssertEqual(t, limiter.Allow("client1"), false)
	})

	t.Run("different clients", func(t *testing.T) {
		testhelpers.AssertEqual(t, limiter.Allow("client2"), true)
	})

	t.Run("reset after window", func(t *testing.T) {
		time.Sleep(1100 * time.Millisecond)
		testhelpers.AssertEqual(t, limiter.Allow("client1"), true)
	})
}

func TestNewMiddleware(t *testing.T) {
	config := DefaultSecurityConfig()
	middleware := NewMiddleware(config)
	testhelpers.AssertNotNil(t, middleware)
}

func TestMiddleware_SecureHeaders(t *testing.T) {
	middleware := NewMiddleware(nil)

	handler := middleware.SecureHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	testhelpers.AssertEqual(t, rr.Code, http.StatusOK)
	testhelpers.AssertEqual(t, rr.Header().Get("X-Content-Type-Options"), "nosniff")
	testhelpers.AssertEqual(t, rr.Header().Get("X-Frame-Options"), "DENY")
}

func TestMiddleware_RateLimit(t *testing.T) {
	config := &SecurityConfig{
		RateLimitEnabled:  true,
		RateLimitRequests: 2,
		RateLimitWindow:   60,
	}
	middleware := NewMiddleware(config)

	handler := middleware.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("allow requests within limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		testhelpers.AssertEqual(t, rr.Code, http.StatusOK)
	})

	t.Run("block requests over limit", func(t *testing.T) {
		// Make multiple requests from same IP
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
		}
		// Last request should be rate limited
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		testhelpers.AssertEqual(t, rr.Code, http.StatusTooManyRequests)
	})
}

func TestAuditEvent(t *testing.T) {
	event := &AuditEvent{
		Timestamp: time.Now(),
		Level:     "INFO",
		Category:  "test",
		Action:    "test_action",
		Resource:  "test_resource",
		Result:    "success",
	}

	testhelpers.AssertEqual(t, event.Level, "INFO")
	testhelpers.AssertEqual(t, event.Category, "test")
}

func TestAuditLevel_String(t *testing.T) {
	tests := []struct {
		level    AuditLevel
		expected string
	}{
		{Debug, "DEBUG"},
		{Info, "INFO"},
		{Warning, "WARNING"},
		{Error, "ERROR"},
		{Critical, "CRITICAL"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			testhelpers.AssertEqual(t, result, tt.expected)
		})
	}
}

func TestSecurityConfig(t *testing.T) {
	config := DefaultSecurityConfig()

	t.Run("default config", func(t *testing.T) {
		testhelpers.AssertEqual(t, config.MaxRequestSize, int64(10*1024*1024))
		testhelpers.AssertEqual(t, config.RateLimitEnabled, true)
		testhelpers.AssertEqual(t, config.SanitizeInput, true)
	})

	t.Run("is host allowed - no restrictions", func(t *testing.T) {
		testhelpers.AssertEqual(t, config.IsHostAllowed("any-host"), true)
	})

	t.Run("is host allowed - with restrictions", func(t *testing.T) {
		config.AllowedHosts = []string{"example.com", "test.com"}
		testhelpers.AssertEqual(t, config.IsHostAllowed("example.com"), true)
		testhelpers.AssertEqual(t, config.IsHostAllowed("notallowed.com"), false)
	})

	t.Run("is IP blocked", func(t *testing.T) {
		config.BlockedIPs = []string{"192.168.1.1"}
		testhelpers.AssertEqual(t, config.IsIPBlocked("192.168.1.1"), true)
		testhelpers.AssertEqual(t, config.IsIPBlocked("192.168.1.2"), false)
	})
}

func BenchmarkSanitizer_SanitizeString(b *testing.B) {
	sanitizer := NewSanitizer()
	input := "hello\x00world\x01test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sanitizer.SanitizeString(input)
	}
}

func BenchmarkValidator_ValidateString(b *testing.B) {
	validator := NewValidator()
	input := "hello world test string"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateString(input)
	}
}

func BenchmarkDataClassifier_Classify(b *testing.B) {
	classifier := NewDataClassifier()
	input := "user@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		classifier.Classify(input)
	}
}
