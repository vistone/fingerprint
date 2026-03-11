// Package core tests
package core

import (
	"testing"
)

func TestHTTPHeadersClone(t *testing.T) {
	h := &HTTPHeaders{
		Accept:    "text/html",
		UserAgent: "Mozilla/5.0",
		Custom:    map[string]string{"Cookie": "session=abc"},
	}

	cloned := h.Clone()

	// verify cloned values are equal
	if cloned.Accept != h.Accept {
		t.Errorf("Accept mismatch: got %s, want %s", cloned.Accept, h.Accept)
	}
	if cloned.UserAgent != h.UserAgent {
		t.Errorf("UserAgent mismatch: got %s, want %s", cloned.UserAgent, h.UserAgent)
	}

	// verify deep copy
	cloned.Custom["Cookie"] = "modified"
	if h.Custom["Cookie"] == "modified" {
		t.Error("Clone should be deep copy")
	}
}

func TestHTTPHeadersToMap(t *testing.T) {
	h := &HTTPHeaders{
		Accept:         "text/html",
		AcceptLanguage: "en-US",
		UserAgent:      "Mozilla/5.0",
		Custom:         map[string]string{"X-Custom": "value"},
	}

	m := h.ToMap()

	if m["Accept"] != "text/html" {
		t.Errorf("Accept: got %s, want text/html", m["Accept"])
	}
	if m["X-Custom"] != "value" {
		t.Errorf("Custom header not included")
	}
}

func TestHTTPHeadersSet(t *testing.T) {
	h := &HTTPHeaders{}
	h.Set("Cookie", "session=xyz")

	if h.Custom == nil {
		t.Fatal("Custom map should be initialized")
	}
	if h.Custom["Cookie"] != "session=xyz" {
		t.Errorf("Cookie not set correctly")
	}
}

func TestFeatureVector(t *testing.T) {
	fv := NewFeatureVector()

	fv.Set(FeatureTLSVersion, 0x0303)
	fv.Set(FeatureCipherSuites, 8.0)

	if fv.Get(FeatureTLSVersion) != 0x0303 {
		t.Errorf("TLSVersion: got %v, want 0x0303", fv.Get(FeatureTLSVersion))
	}
	if fv.Get(FeatureCipherSuites) != 8.0 {
		t.Errorf("CipherSuites: got %v, want 8.0", fv.Get(FeatureCipherSuites))
	}
	if fv.Get(FeatureUserAgent) != 0 {
		t.Errorf("Unset feature should return 0")
	}
}

func TestBrowserTypeConsts(t *testing.T) {
	if BrowserChrome != "chrome" {
		t.Errorf("BrowserChrome = %s, want chrome", BrowserChrome)
	}
	if BrowserFirefox != "firefox" {
		t.Errorf("BrowserFirefox = %s, want firefox", BrowserFirefox)
	}
}

func BenchmarkHTTPHeadersClone(b *testing.B) {
	h := &HTTPHeaders{
		Accept:         "text/html,application/xhtml+xml",
		AcceptLanguage: "en-US,en;q=0.9",
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		Custom:         map[string]string{"Cookie": "session=abc", "X-Token": "token123"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.Clone()
	}
}

func BenchmarkFeatureVectorSet(b *testing.B) {
	fv := NewFeatureVector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fv.Set(FeatureTLSVersion, float64(i))
	}
}
