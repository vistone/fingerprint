package waf

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewWAF(t *testing.T) {
	config := DefaultWAFConfig
	w := NewWAF(config)

	if w == nil {
		t.Fatal("NewWAF returned nil")
	}

	if w.config.Mode != WAFModeProtection {
		t.Errorf("Expected mode %s, got %s", WAFModeProtection, w.config.Mode)
	}
}

func TestWAFAllow(t *testing.T) {
	config := &WAFConfig{
		Enabled: true,
		Mode:    WAFModeLearning,
	}

	w := NewWAF(config)
	defer w.Stop()

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	result := w.Analyze(req.Context(), req)

	if result.Action != ActionMonitor {
		t.Errorf("Expected action %s in learning mode, got %s", ActionMonitor, result.Action)
	}
}

func TestWAFBlockBlacklistIP(t *testing.T) {
	config := &WAFConfig{
		Enabled:      true,
		Mode:         WAFModeProtection,
		BlacklistIPs: []string{"192.168.1.100"},
	}

	w := NewWAF(config)
	defer w.Stop()

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	result := w.Analyze(req.Context(), req)

	if result.Action != ActionBlock {
		t.Errorf("Expected block action for blacklisted IP, got %s", result.Action)
	}
}

func TestWAFBlockBlacklistPath(t *testing.T) {
	config := &WAFConfig{
		Enabled:        true,
		Mode:           WAFModeProtection,
		BlacklistPaths: []string{"/admin"},
	}

	w := NewWAF(config)
	defer w.Stop()

	req := httptest.NewRequest("GET", "/admin/users", nil)

	result := w.Analyze(req.Context(), req)

	if result.Action != ActionBlock {
		t.Errorf("Expected block action for blacklisted path, got %s", result.Action)
	}
}

func TestWAFWhitelistPath(t *testing.T) {
	config := &WAFConfig{
		Enabled:        true,
		Mode:           WAFModeProtection,
		WhitelistPaths: []string{"/health"},
	}

	w := NewWAF(config)
	defer w.Stop()

	req := httptest.NewRequest("GET", "/health", nil)

	result := w.Analyze(req.Context(), req)

	if result.Action != ActionAllow {
		t.Errorf("Expected allow action for whitelisted path, got %s", result.Action)
	}
}

func TestNetworkEngine(t *testing.T) {
	engine := NewNetworkEngine()

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "192.168.1.50:12345"

	result := engine.Analyze(req)

	if result == nil {
		t.Fatal("NetworkEngine.Analyze returned nil")
	}

	if result.Score < 0 {
		t.Errorf("Expected non-negative score, got %f", result.Score)
	}
}

func TestTLSEngine(t *testing.T) {
	blacklist := []string{"blacklisted-ja3"}
	engine := NewTLSEngine(blacklist, nil)

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-JA3-Fingerprint", "blacklisted-ja3")

	result := engine.Analyze(req)

	if !result.IsSuspicious {
		t.Error("Expected suspicious result for blacklisted JA3")
	}

	if result.Score < 0.9 {
		t.Errorf("Expected high score for blacklisted JA3, got %f", result.Score)
	}
}

func TestHTTPEngine(t *testing.T) {
	engine := NewHTTPEngine()

	// Test with bot User-Agent
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("User-Agent", "curl/7.68.0")

	result := engine.Analyze(req)

	if !result.IsAutomated {
		t.Error("Expected automated detection for curl UA")
	}
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(2, 5, time.Second)

	// Should allow first 5 requests (burst)
	for i := 0; i < 5; i++ {
		if !rl.Allow("client1") {
			t.Errorf("Request %d should be allowed within burst", i+1)
		}
	}

	// 6th request should be rate limited
	if rl.Allow("client1") {
		t.Error("6th request should be rate limited")
	}

	// Different client should be allowed
	if !rl.Allow("client2") {
		t.Error("Different client should be allowed")
	}
}

func TestWAFGetClientIP_IgnoresForwardedHeadersFromUntrustedPeer(t *testing.T) {
	config := &WAFConfig{
		Enabled:        true,
		Mode:           WAFModeProtection,
		TrustedProxies: []string{"203.0.113.10"},
	}
	w := NewWAF(config)
	defer w.Stop()

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "198.51.100.20:12345"
	req.Header.Set("X-Forwarded-For", "192.0.2.10")

	clientIP := w.getClientIP(req)
	if clientIP != "198.51.100.20" {
		t.Fatalf("expected remote IP from untrusted peer, got %q", clientIP)
	}
}

func TestWAFGetClientIP_UsesForwardedHeadersFromTrustedProxy(t *testing.T) {
	config := &WAFConfig{
		Enabled:        true,
		Mode:           WAFModeProtection,
		TrustedProxies: []string{"203.0.113.10"},
	}
	w := NewWAF(config)
	defer w.Stop()

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "203.0.113.10:12345"
	req.Header.Set("X-Forwarded-For", "192.0.2.10, 203.0.113.10")

	clientIP := w.getClientIP(req)
	if clientIP != "192.0.2.10" {
		t.Fatalf("expected forwarded IP from trusted proxy, got %q", clientIP)
	}
}

func TestBlockList(t *testing.T) {
	bl := NewBlockList(1 * time.Hour)
	defer bl.Stop()

	// Block an IP
	bl.Block("10.0.0.1", "test reason")

	// Should be blocked
	if !bl.IsBlocked("10.0.0.1") {
		t.Error("IP should be blocked")
	}

	// Different IP should not be blocked
	if bl.IsBlocked("10.0.0.2") {
		t.Error("Different IP should not be blocked")
	}

	// Unblock
	bl.Unblock("10.0.0.1")
	if bl.IsBlocked("10.0.0.1") {
		t.Error("IP should be unblocked")
	}
}

func BenchmarkWAFAnalyze(b *testing.B) {
	config := DefaultWAFConfig
	w := NewWAF(config)
	defer w.Stop()

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	ctx := req.Context()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Analyze(ctx, req)
	}
}

func BenchmarkRateLimiter(b *testing.B) {
	rl := NewRateLimiter(1000, 1000, time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow("client1")
	}
}
