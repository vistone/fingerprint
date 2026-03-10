package gateway

import (
	"testing"
	"time"
)

func TestFingerprintCache_ReturnsClonedResponses(t *testing.T) {
	cache := NewFingerprintCache(16, time.Minute)
	original := &AnalyzeResponse{
		FingerprintHash: "hash-1",
		DefenseHints:    []string{"hint-a"},
		JA4H: &JA4HInfo{
			Fingerprint: "ja4h-a",
			Headers:     []string{"x-test"},
		},
	}

	cache.Set("k", original)
	original.DefenseHints[0] = "mutated-outside"
	original.JA4H.Headers[0] = "mutated-header"

	first, ok := cache.Get("k")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if first.DefenseHints[0] != "hint-a" {
		t.Fatalf("unexpected cached defense hint: %q", first.DefenseHints[0])
	}
	if first.JA4H.Headers[0] != "x-test" {
		t.Fatalf("unexpected cached JA4H header: %q", first.JA4H.Headers[0])
	}

	first.DefenseHints[0] = "mutated-on-read"
	first.JA4H.Headers[0] = "mutated-on-read-header"

	second, ok := cache.Get("k")
	if !ok {
		t.Fatal("expected cache hit on second read")
	}
	if second.DefenseHints[0] != "hint-a" {
		t.Fatalf("cache read should be immutable, got defense hint: %q", second.DefenseHints[0])
	}
	if second.JA4H.Headers[0] != "x-test" {
		t.Fatalf("cache read should be immutable, got JA4H header: %q", second.JA4H.Headers[0])
	}
}
