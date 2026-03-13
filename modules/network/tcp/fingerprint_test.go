package tcp

import (
	"testing"
)

func ptrUint16(v uint16) *uint16 { return &v }
func ptrUint8(v uint8) *uint8    { return &v }

func TestComputeSignature(t *testing.T) {
	analyzer := NewTCPIPAnalyzer()

	packet := TCPPacket{
		IPHeader: IPHeader{
			TTL:   64,
			Flags: 0x02, // DF set
		},
		WindowSize: 65535,
		Flags:      TCPFlags{SYN: true},
		Options: TCPOptions{
			MSS:           ptrUint16(1460),
			WindowScale:   ptrUint8(7),
			SAckPermitted: true,
			Timestamps:    true,
		},
	}

	sig, err := analyzer.ComputeSignature(packet)
	if err != nil {
		t.Fatalf("ComputeSignature: %v", err)
	}
	if sig == "" {
		t.Error("Signature should not be empty")
	}

	// Verify format: TTL:WindowSize:DF:Options:MSS:WindowScale
	parts := splitSignature(sig)
	if len(parts) != 6 {
		t.Errorf("Expected 6 parts in signature, got %d: %s", len(parts), sig)
	}
}

func splitSignature(sig string) []string {
	var parts []string
	current := ""
	for _, c := range sig {
		if c == ':' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	parts = append(parts, current)
	return parts
}

func TestAnalyzePacket(t *testing.T) {
	analyzer := NewTCPIPAnalyzer()

	t.Run("windows_packet", func(t *testing.T) {
		packet := TCPPacket{
			IPHeader: IPHeader{
				TTL: 128,
			},
			WindowSize: 65535,
			Flags:      TCPFlags{SYN: true},
			Options: TCPOptions{
				MSS:         ptrUint16(1460),
				WindowScale: ptrUint8(8),
			},
		}

		sig, err := analyzer.AnalyzePacket(packet)
		if err != nil {
			t.Fatalf("AnalyzePacket: %v", err)
		}

		if sig.Hash == "" {
			t.Error("Hash should not be empty")
		}
		if sig.RawSignature == "" {
			t.Error("RawSignature should not be empty")
		}
		if sig.OS != "Windows" {
			t.Errorf("OS = %s, want Windows", sig.OS)
		}
		if sig.TTLValue != 128 {
			t.Errorf("TTLValue = %d, want 128", sig.TTLValue)
		}
		if sig.MSS != 1460 {
			t.Errorf("MSS = %d, want 1460", sig.MSS)
		}
	})

	t.Run("linux_packet", func(t *testing.T) {
		packet := TCPPacket{
			IPHeader: IPHeader{
				TTL: 64,
			},
			WindowSize: 65535,
			Flags:      TCPFlags{SYN: true},
			Options: TCPOptions{
				MSS:           ptrUint16(1460),
				WindowScale:   ptrUint8(7),
				SAckPermitted: true,
				Timestamps:    true,
			},
		}

		sig, err := analyzer.AnalyzePacket(packet)
		if err != nil {
			t.Fatalf("AnalyzePacket: %v", err)
		}

		if sig.OS == "" {
			t.Error("OS should not be empty")
		}
	})
}

func TestAnalyzeStream(t *testing.T) {
	analyzer := NewTCPIPAnalyzer()

	// Add multiple packets
	for i := 0; i < 5; i++ {
		analyzer.AddPacket(TCPPacket{
			IPHeader: IPHeader{
				TTL: 64,
			},
			WindowSize: 65535,
			Flags:      TCPFlags{SYN: i == 0},
			Options: TCPOptions{
				MSS:           ptrUint16(1460),
				WindowScale:   ptrUint8(7),
				SAckPermitted: true,
				Timestamps:    true,
			},
			RoundTripMs: 50,
		})
	}

	result, err := analyzer.AnalyzeStream()
	if err != nil {
		t.Fatalf("AnalyzeStream: %v", err)
	}

	if result.AverageWindowSize != 65535 {
		t.Errorf("AverageWindowSize = %d, want 65535", result.AverageWindowSize)
	}
	if result.NetworkLatency != 50 {
		t.Errorf("NetworkLatency = %d, want 50", result.NetworkLatency)
	}
	if result.InitialTTL != 64 {
		t.Errorf("InitialTTL = %d, want 64", result.InitialTTL)
	}
	if result.MSS != 1460 {
		t.Errorf("MSS = %d, want 1460", result.MSS)
	}
}

func TestAnalyzeStreamEmpty(t *testing.T) {
	analyzer := NewTCPIPAnalyzer()

	result, err := analyzer.AnalyzeStream()
	if err != nil {
		t.Fatalf("AnalyzeStream: %v", err)
	}
	if result.AverageWindowSize != 0 {
		t.Errorf("Empty stream should have 0 average window size")
	}
}

func TestDetectAnomalies(t *testing.T) {
	t.Run("ttl_inconsistency", func(t *testing.T) {
		analyzer := NewTCPIPAnalyzer()
		ttls := []uint8{64, 128, 63, 127, 255}
		for _, ttl := range ttls {
			analyzer.AddPacket(TCPPacket{
				IPHeader: IPHeader{TTL: ttl},
			})
		}

		anomalies := analyzer.DetectAnomalies()
		found := false
		for _, a := range anomalies {
			if a == "TTL_INCONSISTENCY" {
				found = true
			}
		}
		if !found {
			t.Error("Expected TTL_INCONSISTENCY anomaly")
		}
	})

	t.Run("excessive_rst", func(t *testing.T) {
		analyzer := NewTCPIPAnalyzer()
		for i := 0; i < 6; i++ {
			analyzer.AddPacket(TCPPacket{
				IPHeader: IPHeader{TTL: 64},
				Flags:    TCPFlags{RST: i < 4}, // 4 out of 6 RST
			})
		}

		anomalies := analyzer.DetectAnomalies()
		found := false
		for _, a := range anomalies {
			if a == "EXCESSIVE_RST" {
				found = true
			}
		}
		if !found {
			t.Error("Expected EXCESSIVE_RST anomaly")
		}
	})

	t.Run("zero_window", func(t *testing.T) {
		analyzer := NewTCPIPAnalyzer()
		for i := 0; i < 5; i++ {
			analyzer.AddPacket(TCPPacket{
				IPHeader:   IPHeader{TTL: 64},
				WindowSize: 0,
			})
		}

		anomalies := analyzer.DetectAnomalies()
		found := false
		for _, a := range anomalies {
			if a == "ZERO_WINDOW_PROBES" {
				found = true
			}
		}
		if !found {
			t.Error("Expected ZERO_WINDOW_PROBES anomaly")
		}
	})
}

func TestDetectVPN(t *testing.T) {
	t.Run("low_mss", func(t *testing.T) {
		analyzer := NewTCPIPAnalyzer()
		analyzer.AddPacket(TCPPacket{
			IPHeader: IPHeader{TTL: 64},
			Options: TCPOptions{
				MSS: ptrUint16(1380), // Lower due to VPN overhead
			},
		})

		if !analyzer.DetectVPN() {
			t.Error("Low MSS should indicate VPN")
		}
	})

	t.Run("normal_mss", func(t *testing.T) {
		analyzer := NewTCPIPAnalyzer()
		analyzer.AddPacket(TCPPacket{
			IPHeader: IPHeader{TTL: 64},
			Options: TCPOptions{
				MSS: ptrUint16(1460),
			},
		})

		if analyzer.DetectVPN() {
			t.Error("Normal MSS should not indicate VPN")
		}
	})
}

func TestDetectNAT(t *testing.T) {
	t.Run("with_nat", func(t *testing.T) {
		analyzer := NewTCPIPAnalyzer()
		// Large gaps in IP ID indicate multiple hosts behind NAT
		ids := []uint16{100, 5000, 200, 8000, 300, 9000}
		for _, id := range ids {
			analyzer.AddPacket(TCPPacket{
				IPHeader: IPHeader{TTL: 64, ID: id},
			})
		}

		if !analyzer.DetectNAT() {
			t.Error("Large IP ID gaps should indicate NAT")
		}
	})

	t.Run("without_nat", func(t *testing.T) {
		analyzer := NewTCPIPAnalyzer()
		for i := uint16(100); i < 110; i++ {
			analyzer.AddPacket(TCPPacket{
				IPHeader: IPHeader{TTL: 64, ID: i},
			})
		}

		if analyzer.DetectNAT() {
			t.Error("Sequential IP IDs should not indicate NAT")
		}
	})
}

func TestGetRiskScore(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		analyzer := NewTCPIPAnalyzer()
		if analyzer.GetRiskScore() != 0.0 {
			t.Error("Empty analyzer should have 0 risk score")
		}
	})

	t.Run("anomalous_packets", func(t *testing.T) {
		analyzer := NewTCPIPAnalyzer()
		analyzer.AddPacket(TCPPacket{
			IPHeader:   IPHeader{TTL: 10}, // Very low TTL
			WindowSize: 0,                 // Zero window
			Flags:      TCPFlags{RST: true},
		})

		score := analyzer.GetRiskScore()
		if score == 0.0 {
			t.Error("Anomalous packets should produce non-zero risk score")
		}
		if score > 1.0 {
			t.Error("Risk score should not exceed 1.0")
		}
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("nearestDefaultTTL", func(t *testing.T) {
		tests := []struct {
			ttl  uint8
			want int
		}{
			{20, 32},
			{32, 32},
			{60, 64},
			{64, 64},
			{100, 128},
			{128, 128},
			{200, 255},
		}
		for _, tt := range tests {
			got := nearestDefaultTTL(tt.ttl)
			if got != tt.want {
				t.Errorf("nearestDefaultTTL(%d) = %d, want %d", tt.ttl, got, tt.want)
			}
		}
	})

	t.Run("windowSizeFamily", func(t *testing.T) {
		if windowSizeFamily(65535) != "Linux/macOS/Windows" {
			t.Error("65535 should be Linux/macOS/Windows")
		}
		if windowSizeFamily(8192) != "Windows" {
			t.Error("8192 should be Windows")
		}
	})

	t.Run("formatTCPOptions", func(t *testing.T) {
		opts := TCPOptions{
			MSS:           ptrUint16(1460),
			WindowScale:   ptrUint8(7),
			SAckPermitted: true,
			Timestamps:    true,
		}
		result := formatTCPOptions(opts)
		if result == "" || result == "none" {
			t.Error("Options should produce non-empty result")
		}
	})

	t.Run("formatTCPOptions_empty", func(t *testing.T) {
		opts := TCPOptions{}
		result := formatTCPOptions(opts)
		if result != "none" {
			t.Errorf("Empty options should produce 'none', got %s", result)
		}
	})
}

func TestNewAnalyzer(t *testing.T) {
	a := NewAnalyzer()
	if a == nil {
		t.Fatal("NewAnalyzer returned nil")
	}
}

func TestSetOSDatabase(t *testing.T) {
	analyzer := NewTCPIPAnalyzer()
	db := []OSFingerprint{
		{Name: "TestOS", Family: "Test", DefaultTTL: 64},
	}
	analyzer.SetOSDatabase(db)
	got := analyzer.GetOSFingerprints()
	if len(got) != 1 || got[0].Name != "TestOS" {
		t.Error("SetOSDatabase/GetOSFingerprints mismatch")
	}
}

// BenchmarkAnalyzePacket benchmark
func BenchmarkAnalyzePacket(b *testing.B) {
	analyzer := NewTCPIPAnalyzer()
	packet := TCPPacket{
		IPHeader:   IPHeader{TTL: 64, Flags: 0x02},
		WindowSize: 65535,
		Flags:      TCPFlags{SYN: true},
		Options: TCPOptions{
			MSS:           ptrUint16(1460),
			WindowScale:   ptrUint8(7),
			SAckPermitted: true,
			Timestamps:    true,
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = analyzer.AnalyzePacket(packet)
	}
}
