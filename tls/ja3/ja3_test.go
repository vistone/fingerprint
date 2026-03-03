package ja3

import (
	"testing"

	tls "github.com/bogdanfinn/utls"
)

func TestComputeJA3FromSpec(t *testing.T) {
	spec := tls.ClientHelloSpec{
		TLSVersMax: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
		},
		Extensions: []tls.TLSExtension{
			&tls.SNIExtension{},
			&tls.SupportedVersionsExtension{Versions: []uint16{tls.VersionTLS13}},
		},
	}

	result, err := ComputeJA3FromSpec(spec)
	if err != nil {
		t.Fatalf("ComputeJA3FromSpec failed: %v", err)
	}

	if result.Hash == "" {
		t.Error("Expected non-empty hash")
	}

	if result.RawString == "" {
		t.Error("Expected non-empty raw string")
	}
}

func TestMatchJA3(t *testing.T) {
	hash1 := "abc123"
	hash2 := "abc123"
	hash3 := "def456"

	if !MatchJA3(hash1, hash2) {
		t.Error("Expected matching hashes to return true")
	}

	if MatchJA3(hash1, hash3) {
		t.Error("Expected different hashes to return false")
	}
}

func TestFindProfileByJA3(t *testing.T) {
	// Test with empty hash
	profiles := FindProfileByJA3("")
	if len(profiles) != 0 {
		t.Error("Expected no profiles for empty hash")
	}
}
