package ja4t

import (
	"testing"
)

func TestComputeJA4T(t *testing.T) {
	tests := []struct {
		name       string
		data       TCPSYNData
		wantEmpty  bool
		checkFunc  func(t *testing.T, result *JA4TResult)
	}{
		{
			name: "windows_typical",
			data: TCPSYNData{
				WindowSize:  65535,
				Options:     []uint8{2, 1, 3, 1, 1, 4},
				MSS:         1460,
				WindowScale: 8,
				TTL:         128,
				DF:          true,
			},
			checkFunc: func(t *testing.T, result *JA4TResult) {
				if result.RawFingerprint != "65535_2-1-3-1-1-4_1460_8" {
					t.Errorf("RawFingerprint = %s, want 65535_2-1-3-1-1-4_1460_8", result.RawFingerprint)
				}
				if len(result.Hash) != 12 {
					t.Errorf("Hash length = %d, want 12", len(result.Hash))
				}
				if result.ProbableOS != "Windows" {
					t.Errorf("ProbableOS = %s, want Windows", result.ProbableOS)
				}
				if result.WindowSize != 65535 {
					t.Errorf("WindowSize = %d, want 65535", result.WindowSize)
				}
			},
		},
		{
			name: "linux_typical",
			data: TCPSYNData{
				WindowSize:  65535,
				Options:     []uint8{2, 4, 8, 1, 3},
				MSS:         1460,
				WindowScale: 7,
				TTL:         64,
				DF:          true,
			},
			checkFunc: func(t *testing.T, result *JA4TResult) {
				if result.RawFingerprint != "65535_2-4-8-1-3_1460_7" {
					t.Errorf("RawFingerprint = %s, want 65535_2-4-8-1-3_1460_7", result.RawFingerprint)
				}
				if len(result.Hash) != 12 {
					t.Errorf("Hash length = %d, want 12", len(result.Hash))
				}
				if result.ProbableOS != "Linux" {
					t.Errorf("ProbableOS = %s, want Linux", result.ProbableOS)
				}
			},
		},
		{
			name: "macos_typical",
			data: TCPSYNData{
				WindowSize:  65535,
				Options:     []uint8{2, 1, 1, 4, 8, 3},
				MSS:         1460,
				WindowScale: 6,
				TTL:         64,
				DF:          true,
			},
			checkFunc: func(t *testing.T, result *JA4TResult) {
				if result.ProbableOS != "macOS" {
					t.Errorf("ProbableOS = %s, want macOS", result.ProbableOS)
				}
			},
		},
		{
			name: "anomalous_zero_window",
			data: TCPSYNData{
				WindowSize:  0,
				Options:     []uint8{2, 1, 3},
				MSS:         1460,
				WindowScale: 7,
				TTL:         64,
			},
			checkFunc: func(t *testing.T, result *JA4TResult) {
				found := false
				for _, flag := range result.AnomalyFlags {
					if flag == "ZERO_WINDOW" {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected ZERO_WINDOW anomaly flag")
				}
				if result.RiskScore == 0 {
					t.Error("Expected non-zero risk score")
				}
			},
		},
		{
			name: "anomalous_no_options",
			data: TCPSYNData{
				WindowSize:  65535,
				Options:     []uint8{},
				MSS:         0,
				WindowScale: 0,
				TTL:         64,
			},
			checkFunc: func(t *testing.T, result *JA4TResult) {
				foundNoOpts := false
				foundNoMSS := false
				for _, flag := range result.AnomalyFlags {
					if flag == "NO_OPTIONS" {
						foundNoOpts = true
					}
					if flag == "NO_MSS" {
						foundNoMSS = true
					}
				}
				if !foundNoOpts {
					t.Error("Expected NO_OPTIONS anomaly flag")
				}
				if !foundNoMSS {
					t.Error("Expected NO_MSS anomaly flag")
				}
			},
		},
		{
			name: "low_ttl",
			data: TCPSYNData{
				WindowSize:  65535,
				Options:     []uint8{2, 1, 3},
				MSS:         1460,
				WindowScale: 7,
				TTL:         20,
			},
			checkFunc: func(t *testing.T, result *JA4TResult) {
				found := false
				for _, flag := range result.AnomalyFlags {
					if flag == "LOW_TTL" {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected LOW_TTL anomaly flag")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeJA4T(tt.data)
			if result == nil {
				t.Fatal("ComputeJA4T returned nil")
			}
			if result.RawFingerprint == "" {
				t.Error("RawFingerprint should not be empty")
			}
			if result.Hash == "" {
				t.Error("Hash should not be empty")
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestComputeJA4TS(t *testing.T) {
	data := TCPSYNData{
		WindowSize:  28960,
		Options:     []uint8{2, 1, 3, 1, 1, 8, 4},
		MSS:         1460,
		WindowScale: 7,
		TTL:         64,
	}

	result := ComputeJA4TS(data)
	if result == nil {
		t.Fatal("ComputeJA4TS returned nil")
	}
	if result.RawFingerprint != "28960_2-1-3-1-1-8-4_1460_7" {
		t.Errorf("RawFingerprint = %s, want 28960_2-1-3-1-1-8-4_1460_7", result.RawFingerprint)
	}
	if len(result.Hash) != 12 {
		t.Errorf("Hash length = %d, want 12", len(result.Hash))
	}
}

func TestMatchJA4T(t *testing.T) {
	data1 := TCPSYNData{
		WindowSize: 65535, Options: []uint8{2, 1, 3}, MSS: 1460, WindowScale: 7,
	}
	data2 := TCPSYNData{
		WindowSize: 65535, Options: []uint8{2, 1, 3}, MSS: 1460, WindowScale: 7,
	}
	data3 := TCPSYNData{
		WindowSize: 32768, Options: []uint8{2, 1, 3}, MSS: 1460, WindowScale: 7,
	}

	r1 := ComputeJA4T(data1)
	r2 := ComputeJA4T(data2)
	r3 := ComputeJA4T(data3)

	if !MatchJA4T(r1.Hash, r2.Hash) {
		t.Error("Identical data should produce matching hashes")
	}
	if MatchJA4T(r1.Hash, r3.Hash) {
		t.Error("Different data should produce different hashes")
	}
	if MatchJA4T("short", r1.Hash) {
		t.Error("Short hash should not match")
	}
}

func TestGetOptionNames(t *testing.T) {
	options := []uint8{2, 1, 3, 4, 8}
	names := GetOptionNames(options)

	expected := []string{"MSS", "NOP", "WS", "SACK", "TS"}
	if len(names) != len(expected) {
		t.Fatalf("len(names) = %d, want %d", len(names), len(expected))
	}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("names[%d] = %s, want %s", i, name, expected[i])
		}
	}
}

func TestKnownOSProfiles(t *testing.T) {
	profiles := KnownOSProfiles()
	if len(profiles) == 0 {
		t.Fatal("KnownOSProfiles should return profiles")
	}

	for _, p := range profiles {
		if p.Name == "" {
			t.Error("Profile name should not be empty")
		}
		synData := p.ToSYNData()
		result := ComputeJA4T(synData)
		if result.Hash == "" {
			t.Errorf("Profile %s should produce a hash", p.Name)
		}
	}
}

func TestDeterministicHash(t *testing.T) {
	data := TCPSYNData{
		WindowSize:  65535,
		Options:     []uint8{2, 1, 3, 1, 1, 4},
		MSS:         1460,
		WindowScale: 8,
		TTL:         128,
	}

	r1 := ComputeJA4T(data)
	r2 := ComputeJA4T(data)

	if r1.Hash != r2.Hash {
		t.Error("Same input should produce same hash")
	}
	if r1.RawFingerprint != r2.RawFingerprint {
		t.Error("Same input should produce same raw fingerprint")
	}
}

func TestAnomalyScoreCapped(t *testing.T) {
	// Create data with multiple anomalies to verify score is capped at 1.0
	data := TCPSYNData{
		WindowSize:  0,
		Options:     []uint8{},
		MSS:         100,
		WindowScale: 15,
		TTL:         10,
	}

	result := ComputeJA4T(data)
	if result.RiskScore > 1.0 {
		t.Errorf("RiskScore = %f, should not exceed 1.0", result.RiskScore)
	}
}

// BenchmarkComputeJA4T 基准测试
func BenchmarkComputeJA4T(b *testing.B) {
	data := TCPSYNData{
		WindowSize:  65535,
		Options:     []uint8{2, 1, 3, 1, 1, 4},
		MSS:         1460,
		WindowScale: 8,
		TTL:         128,
		DF:          true,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ComputeJA4T(data)
	}
}
