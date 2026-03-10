// Package tls tests
package tls

import (
	"testing"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

func TestCalculateJA3(t *testing.T) {
	spec := core.ClientHelloSpec{
		TLSVersion:   0x0303,
		CipherSuites: []uint16{0x1301, 0x1302, 0x1303},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256},
		SupportedPoints: []uint8{0},
	}

	result := CalculateJA3(spec)

	if result == nil {
		t.Fatal("CalculateJA3 returned nil")
	}

	if result.Hash == "" {
		t.Error("JA3 hash should not be empty")
	}

	if result.RawString == "" {
		t.Error("JA3 raw string should not be empty")
	}

	if result.TLSVersion != 0x0303 {
		t.Errorf("TLSVersion = 0x%04x, want 0x0303", result.TLSVersion)
	}

	// Verify MD5 hash format (32 hexadecimal characters).
	if len(result.Hash) != 32 {
		t.Errorf("JA3 hash should be 32 chars, got %d", len(result.Hash))
	}
}

func TestCalculateJA4(t *testing.T) {
	spec := core.ClientHelloSpec{
		TLSVersion:   0x0303,
		CipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0x000a},
		},
	}

	result := CalculateJA4(spec)

	if result == nil {
		t.Fatal("CalculateJA4 returned nil")
	}

	if result.Fingerprint == "" {
		t.Error("JA4 fingerprint should not be empty")
	}

	// Verify JA4 format (simplified check).
	if result.CipherSuitesCount == 0 {
		t.Error("CipherSuitesCount should not be 0")
	}
	if result.ExtensionsCount == 0 {
		t.Error("ExtensionsCount should not be 0")
	}
}

func TestIsGREASEUint16(t *testing.T) {
	tests := []struct {
		value    uint16
		expected bool
	}{
		{0x0A0A, true},  // GREASE
		{0x1A1A, true},  // GREASE
		{0x2A2A, true},  // GREASE
		{0xFAFA, true},  // GREASE
		{0x1301, false}, // TLS_AES_128_GCM_SHA256
		{0x1302, false}, // TLS_AES_256_GCM_SHA384
		{0x002B, false}, // supported_versions
		{0x0000, false}, // server_name
	}

	for _, tt := range tests {
		got := IsGREASEUint16(tt.value)
		if got != tt.expected {
			t.Errorf("IsGREASEUint16(0x%04x) = %v, want %v", tt.value, got, tt.expected)
		}
	}
}

func TestFilterGREASEUint16(t *testing.T) {
	input := []uint16{0x1301, 0x0A0A, 0x1302, 0x1A1A}
	expected := []uint16{0x1301, 0x1302}

	result := filterGREASEUint16(input)

	if len(result) != len(expected) {
		t.Errorf("filtered length = %d, want %d", len(result), len(expected))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("result[%d] = 0x%04x, want 0x%04x", i, v, expected[i])
		}
	}
}

func TestFilterGREASEExtensions(t *testing.T) {
	input := []core.TLSExtension{
		{Type: 0x0000}, {Type: 0x0A0A}, {Type: 0x0017}, {Type: 0x1A1A},
	}

	result := filterGREASEExtensions(input)

	// Should filter out 2 GREASE extensions.
	if len(result) != 2 {
		t.Errorf("filtered length = %d, want 2", len(result))
	}

	// Verify the remaining extensions are non-GREASE.
	for _, ext := range result {
		if IsGREASEUint16(ext.Type) {
			t.Errorf("Extension 0x%04x should be filtered", ext.Type)
		}
	}
}

func TestJoinUint16(t *testing.T) {
	input := []uint16{0x1301, 0x1302, 0x1303}
	result := joinUint16(input)

	expected := "4865-4866-4867" // Decimal values
	if result != expected {
		t.Errorf("joinUint16 = %s, want %s", result, expected)
	}
}

func TestJoinUint8(t *testing.T) {
	input := []uint8{0, 1, 2}
	result := joinUint8(input)

	expected := "0-1-2"
	if result != expected {
		t.Errorf("joinUint8 = %s, want %s", result, expected)
	}
}

func TestJoinCurves(t *testing.T) {
	input := []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384}
	result := joinCurves(input)

	expected := "29-23-24" // Decimal values
	if result != expected {
		t.Errorf("joinCurves = %s, want %s", result, expected)
	}
}

func TestAnalyzer(t *testing.T) {
	profile := &profiles.ClientProfile{
		TLSVersion:   0x0303,
		CipherSuites: []uint16{0x1301, 0x1302},
		Extensions:   []core.TLSExtension{{Type: 0x0000}},
	}

	analyzer := NewAnalyzer(profile)

	ja3 := analyzer.AnalyzeJA3()
	if ja3 == nil {
		t.Error("AnalyzeJA3 should return result")
	}

	ja4 := analyzer.AnalyzeJA4()
	if ja4 == nil {
		t.Error("AnalyzeJA4 should return result")
	}

	fingerprint := analyzer.Fingerprint()
	if fingerprint["ja3"] == nil {
		t.Error("Fingerprint should include ja3")
	}
	if fingerprint["ja4"] == nil {
		t.Error("Fingerprint should include ja4")
	}
}

func TestAnalyzerNilProfile(t *testing.T) {
	analyzer := NewAnalyzer(nil)

	ja3 := analyzer.AnalyzeJA3()
	if ja3 != nil {
		t.Error("AnalyzeJA3 with nil profile should return nil")
	}

	ja4 := analyzer.AnalyzeJA4()
	if ja4 != nil {
		t.Error("AnalyzeJA4 with nil profile should return nil")
	}
}

func BenchmarkCalculateJA3(b *testing.B) {
	spec := core.ClientHelloSpec{
		TLSVersion:   0x0303,
		CipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
			{Type: 0x0016}, {Type: 0x000d}, {Type: 0x002b},
			{Type: 0x002d}, {Type: 0x0033},
		},
		SupportedCurves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		SupportedPoints: []uint8{0},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateJA3(spec)
	}
}

func BenchmarkCalculateJA4(b *testing.B) {
	spec := core.ClientHelloSpec{
		TLSVersion:   0x0303,
		CipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, {Type: 0x0017}, {Type: 0xff01},
			{Type: 0x000a}, {Type: 0x000b}, {Type: 0x0023},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateJA4(spec)
	}
}

func BenchmarkFilterGREASE(b *testing.B) {
	input := []uint16{
		0x1301, 0x1302, 0x0A0A, 0x1303, 0x1A1A,
		0xc02b, 0xc02f, 0x2A2A, 0xcca9, 0x3A3A,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filterGREASEUint16(input)
	}
}
