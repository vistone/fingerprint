package ja4s

import (
	"fmt"
	"strings"
	"testing"

	"github.com/vistone/fingerprint/modules/profiles/legacy"
)

// TestComputeJA4SFromRealProfiles tests JA4S computation using real fingerprint profiles
func TestComputeJA4SFromRealProfiles(t *testing.T) {
	// Use real Chrome 133 fingerprint
	chromeProfile, ok := profiles.MappedTLSClients["chrome_133"]
	if !ok {
		t.Fatal("chrome_133 profile not found")
	}

	// Get ClientHelloSpec
	_, err := chromeProfile.GetClientHelloSpec()
	if err != nil {
		t.Skipf("chrome_133 does not support spec export: %v", err)
		return
	}

	// Construct ServerHello data (based on real ClientHello response)
	// Use TLS 1.3 configuration compatible with Chrome 133
	serverHello := ServerHelloData{
		TLSVersion:   0x0304,                   // TLS 1.3
		CipherSuite:  0x1301,                   // TLS_AES_128_GCM_SHA256 (Chrome preferred)
		Extensions:   []uint16{0x002b, 0x0033}, // supported_versions, key_share
		Compression:  0,
		ServerName:   "www.google.com",
		SelectedALPN: "h2",
	}

	result, err := ComputeJA4S(serverHello)
	if err != nil {
		t.Fatalf("ComputeJA4S failed: %v", err)
	}

	if result.Hash == "" {
		t.Error("Expected non-empty JA4S hash")
	}

	if result.RawString == "" {
		t.Error("Expected non-empty JA4S raw string")
	}

	t.Logf("Chrome 133 -> Server JA4S Hash: %s", result.Hash)
	t.Logf("Chrome 133 -> Server JA4S Raw: %s", result.RawString)
}

// TestComputeJA4SFromMultipleProfiles tests JA4S computation for multiple real fingerprints
func TestComputeJA4SFromMultipleProfiles(t *testing.T) {
	testCases := []struct {
		profileName string
		cipherSuite uint16
		extensions  []uint16
	}{
		{"chrome_133", 0x1301, []uint16{0x002b, 0x0033, 0x002d}}, // TLS_AES_128_GCM_SHA256
		{"firefox_135", 0x1302, []uint16{0x002b, 0x0033}},        // TLS_AES_256_GCM_SHA384
		{"safari_16_0", 0x1301, []uint16{0x002b, 0x0033}},        // TLS_AES_128_GCM_SHA256
	}

	for _, tc := range testCases {
		t.Run(tc.profileName, func(t *testing.T) {
			profile, ok := profiles.MappedTLSClients[tc.profileName]
			if !ok {
				t.Skipf("Profile %s not found", tc.profileName)
				return
			}

			_, err := profile.GetClientHelloSpec()
			if err != nil {
				t.Skipf("Profile %s does not support spec export: %v", tc.profileName, err)
				return
			}

			serverHello := ServerHelloData{
				TLSVersion:   0x0304,
				CipherSuite:  tc.cipherSuite,
				Extensions:   tc.extensions,
				Compression:  0,
				ServerName:   "example.com",
				SelectedALPN: "h2",
			}

			result, err := ComputeJA4S(serverHello)
			if err != nil {
				t.Fatalf("ComputeJA4S failed for %s: %v", tc.profileName, err)
			}

			if result.Hash == "" {
				t.Errorf("Expected non-empty hash for %s", tc.profileName)
			}

			if result.TLSVersion != "1.3" {
				t.Errorf("Expected TLS 1.3, got %s", result.TLSVersion)
			}

			t.Logf("%s: JA4S=%s, Cipher=0x%04x", tc.profileName, result.Hash, tc.cipherSuite)
		})
	}
}

// TestAnalyzeServerHelloWithRealData tests analysis using real ServerHello data
func TestAnalyzeServerHelloWithRealData(t *testing.T) {
	analyzer := NewJA4SAnalyzer()

	// Test TLS 1.3 ServerHello (simulating real server response)
	serverHello := ServerHelloData{
		TLSVersion:  0x0304,
		CipherSuite: 0x1301, // TLS_AES_128_GCM_SHA256
		Extensions: []uint16{
			0x002b, // supported_versions
			0x0033, // key_share
			0x002d, // psk_key_exchange_modes
		},
		Compression:  0,
		ServerName:   "cloudflare.com",
		SelectedALPN: "h2",
	}

	result, err := analyzer.AnalyzeServerHello(serverHello)
	if err != nil {
		t.Fatalf("AnalyzeServerHello failed: %v", err)
	}

	if result.Hash == "" {
		t.Error("Expected non-empty hash")
	}

	if result.RiskScore < 0 {
		t.Error("Risk score should be non-negative")
	}

	t.Logf("ServerHello Analysis: Hash=%s, RiskScore=%.2f", result.Hash, result.RiskScore)
}

// TestMatchJA4S tests JA4S hash matching
func TestMatchJA4S(t *testing.T) {
	serverHello := ServerHelloData{
		TLSVersion:  0x0304,
		CipherSuite: 0x1301,
		Extensions:  []uint16{0x002b, 0x0033},
		Compression: 0,
	}

	result1, _ := ComputeJA4S(serverHello)
	result2, _ := ComputeJA4S(serverHello)

	// Same input should produce same hash
	if !MatchJA4S(result1.Hash, result2.Hash) {
		t.Error("Same ServerHello should produce matching JA4S hashes")
	}

	// Different input should produce different hash
	serverHello.CipherSuite = 0x1302 // Change cipher suite
	result3, _ := ComputeJA4S(serverHello)

	if MatchJA4S(result1.Hash, result3.Hash) {
		t.Error("Different ServerHello should produce different JA4S hashes")
	}
}

// TestNewJA4SAnalyzer tests that known config database is initialized when creating analyzer
func TestNewJA4SAnalyzer(t *testing.T) {
	analyzer := NewJA4SAnalyzer()
	if analyzer == nil {
		t.Fatal("NewJA4SAnalyzer should return a non-nil analyzer")
	}
	if analyzer.knownProfiles == nil {
		t.Error("Known profiles should be initialized")
	}

	// Verify that known config profiles exist
	expectedProfiles := []string{"nginx_default", "apache_default", "cloudflare"}
	for _, name := range expectedProfiles {
		if _, ok := analyzer.knownProfiles[name]; !ok {
			t.Errorf("Expected known profile %s not found", name)
		}
	}
}

// TestAnalyzeServerHelloBytes tests analyzing TLS ServerHello byte data
func TestAnalyzeServerHelloBytes(t *testing.T) {
	analyzer := NewJA4SAnalyzer()

	tests := []struct {
		name        string
		data        []byte
		expectError bool
		errContains string
	}{
		{
			name:        "data too short",
			data:        make([]byte, 42),
			expectError: true,
			errContains: "ServerHello too short",
		},
		{
			name:        "data exactly 43 bytes but invalid structure",
			data:        make([]byte, 43),
			expectError: false, // Can be parsed but result may be incorrect
		},
		{
			name:        "empty data",
			data:        []byte{},
			expectError: true,
			errContains: "ServerHello too short",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeServerHelloBytes(tc.data)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none, result: %v", result)
				} else if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("Expected error containing %q, got %q", tc.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}

	// Test valid ServerHello byte data
	t.Run("valid ServerHello byte data", func(t *testing.T) {
		// Build a valid TLS ServerHello message
		// HandshakeType(1) + Length(3) + Version(2) + Random(32) + SessionIDLen(1) + SessionID(0) + CipherSuite(2) + Compression(1)
		// Total 42 bytes base, plus 2 bytes extension length
		data := make([]byte, 64)
		data[0] = 0x02                               // Server Hello
		data[1], data[2], data[3] = 0x00, 0x00, 0x3d // Length: 61 bytes
		data[4], data[5] = 0x03, 0x04                // Version: TLS 1.3
		copy(data[6:38], make([]byte, 32))           // Random (32 bytes of zeros)
		data[38] = 0x00                              // Session ID Length: 0
		data[39], data[40] = 0x13, 0x01              // Cipher Suite: TLS_AES_128_GCM_SHA256
		data[41] = 0x00                              // Compression: null
		data[42], data[43] = 0x00, 0x12              // Extensions Length: 18 bytes
		// Extension 1: supported_versions (0x002b)
		data[44], data[45] = 0x00, 0x2b // Type
		data[46], data[47] = 0x00, 0x02 // Length
		data[48], data[49] = 0x03, 0x04 // TLS 1.3
		// Extension 2: key_share (0x0033)
		data[50], data[51] = 0x00, 0x33 // Type
		data[52], data[53] = 0x00, 0x06 // Length
		data[54], data[55] = 0x00, 0x17 // Named Group: secp256r1
		data[56], data[57] = 0x00, 0x02 // Length
		data[58], data[59] = 0x00, 0x00 // Key exchange data

		result, err := analyzer.AnalyzeServerHelloBytes(data)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.Hash == "" {
			t.Error("Expected non-empty hash")
		}
		if result.TLSVersion != "1.3" {
			t.Errorf("Expected TLS version 1.3, got %s", result.TLSVersion)
		}
	})

	// Test parsing extension list
	t.Run("parse extension list", func(t *testing.T) {
		// Build a ServerHello with multiple extensions
		data := make([]byte, 80)
		data[0] = 0x02                               // Server Hello
		data[1], data[2], data[3] = 0x00, 0x00, 0x49 // Length
		data[4], data[5] = 0x03, 0x03                // Version: TLS 1.2
		copy(data[6:38], make([]byte, 32))           // Random
		data[38] = 0x00                              // Session ID Length: 0
		data[39], data[40] = 0x00, 0x2f              // Cipher Suite: TLS_RSA_WITH_AES_128_CBC_SHA
		data[41] = 0x00                              // Compression: null
		data[42], data[43] = 0x00, 0x20              // Extensions Length: 32 bytes

		// Extension 1: server_name (0x0000)
		data[44], data[45] = 0x00, 0x00 // Type
		data[46], data[47] = 0x00, 0x00 // Length: 0

		// Extension 2: supported_groups (0x000a)
		data[48], data[49] = 0x00, 0x0a // Type
		data[50], data[51] = 0x00, 0x04 // Length
		data[52], data[53] = 0x00, 0x02 // List Length
		data[54], data[55] = 0x00, 0x17 // secp256r1

		// Extension 3: ec_point_formats (0x000b)
		data[56], data[57] = 0x00, 0x0b // Type
		data[58], data[59] = 0x00, 0x02 // Length
		data[60], data[61] = 0x01, 0x00 // Length 1, uncompressed

		// Extension 4: session_ticket (0x0023)
		data[62], data[63] = 0x00, 0x23 // Type
		data[64], data[65] = 0x00, 0x00 // Length: 0

		// Extension 5: renegotiation_info (0xff01)
		data[66], data[67] = 0xff, 0x01 // Type
		data[68], data[69] = 0x00, 0x01 // Length
		data[70] = 0x00                 // Info

		result, err := analyzer.AnalyzeServerHelloBytes(data)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(result.JA4Sr) == 0 {
			t.Error("Expected JA4Sr to be set")
		}
		t.Logf("Extensions result: %s", result.JA4Sr)
	})
}

// TestGenerateServerHelloSignature tests signature generation and anomaly detection
func TestGenerateServerHelloSignature(t *testing.T) {
	analyzer := NewJA4SAnalyzer()

	tests := []struct {
		name              string
		tlsVersion        uint16
		cipherSuite       uint16
		extensions        []uint16
		compressionMethod uint8
		expectHashLen     int
		expectRiskScore   float64
		expectAnomalies   []string
	}{
		{
			name:              "TLS 1.3 normal configuration",
			tlsVersion:        0x0304,
			cipherSuite:       0x1301,
			extensions:        []uint16{0x002b, 0x0033, 0x002d},
			compressionMethod: 0,
			expectHashLen:     64,
			expectRiskScore:   0.0,
			expectAnomalies:   nil,
		},
		{
			name:              "TLS 1.2 normal configuration",
			tlsVersion:        0x0303,
			cipherSuite:       0x002f,
			extensions:        []uint16{0x0000, 0x000a, 0x000b},
			compressionMethod: 0,
			expectHashLen:     64,
			expectRiskScore:   0.0,
			expectAnomalies:   nil,
		},
		{
			name:              "TLS 1.0 deprecated version",
			tlsVersion:        0x0301,
			cipherSuite:       0x1301,
			extensions:        []uint16{0x002b, 0x0033, 0x002d},
			compressionMethod: 0,
			expectHashLen:     64,
			expectRiskScore:   0.2,
			expectAnomalies:   []string{"DEPRECATED_TLS_VERSION"},
		},
		{
			name:              "weak cipher suite",
			tlsVersion:        0x0303,
			cipherSuite:       0x000a, // 3DES
			extensions:        []uint16{0x002b, 0x0033, 0x002d},
			compressionMethod: 0,
			expectHashLen:     64,
			expectRiskScore:   0.25,
			expectAnomalies:   []string{"WEAK_CIPHER_SUITE"},
		},
		{
			name:              "too few extensions",
			tlsVersion:        0x0304,
			cipherSuite:       0x1301,
			extensions:        []uint16{0x002b},
			compressionMethod: 0,
			expectHashLen:     64,
			expectRiskScore:   0.2,
			expectAnomalies:   []string{"MINIMAL_EXTENSIONS"},
		},
		{
			name:              "unsafe compression method",
			tlsVersion:        0x0304,
			cipherSuite:       0x1301,
			extensions:        []uint16{0x002b, 0x0033, 0x002d},
			compressionMethod: 1,
			expectHashLen:     64,
			expectRiskScore:   0.2,
			expectAnomalies:   []string{"UNSAFE_COMPRESSION"},
		},
		{
			name:              "multiple anomalies",
			tlsVersion:        0x0301, // Deprecated
			cipherSuite:       0x000a, // Weak cipher
			extensions:        []uint16{0x002b},
			compressionMethod: 1, // Unsafe compression
			expectHashLen:     64,
			expectRiskScore:   0.85, // 0.2 + 0.25 + 0.2 + 0.2
			expectAnomalies:   []string{"DEPRECATED_TLS_VERSION", "WEAK_CIPHER_SUITE", "MINIMAL_EXTENSIONS", "UNSAFE_COMPRESSION"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := analyzer.GenerateServerHelloSignature(tc.tlsVersion, tc.cipherSuite, tc.extensions, tc.compressionMethod)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result.Hash) != tc.expectHashLen {
				t.Errorf("Expected hash length %d, got %d", tc.expectHashLen, len(result.Hash))
			}

			if result.RiskScore < tc.expectRiskScore-0.001 || result.RiskScore > tc.expectRiskScore+0.001 {
				t.Errorf("Expected risk score %.2f, got %.2f", tc.expectRiskScore, result.RiskScore)
			}

			if tc.expectAnomalies != nil {
				for _, anomaly := range tc.expectAnomalies {
					found := false
					for _, flag := range result.AnomalyFlags {
						if flag == anomaly {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected anomaly flag %s not found, got %v", anomaly, result.AnomalyFlags)
					}
				}
			} else if len(result.AnomalyFlags) > 0 {
				t.Errorf("Expected no anomalies, got %v", result.AnomalyFlags)
			}
		})
	}
}

// TestFindMatchingProfiles tests finding matching known server configurations
func TestFindMatchingProfiles(t *testing.T) {
	analyzer := NewJA4SAnalyzer()

	tests := []struct {
		name           string
		result         *JA4SResult
		maxResults     int
		minMatches     int
		maxMatchLength int
	}{
		{
			name: "low risk score should match",
			result: &JA4SResult{
				RiskScore:       0.05,
				AnomalyFlags:    []string{},
				MatchedProfiles: []string{},
			},
			maxResults:     10,
			minMatches:     0,
			maxMatchLength: 3,
		},
		{
			name: "high risk score may not match",
			result: &JA4SResult{
				RiskScore:    0.5,
				AnomalyFlags: []string{"DEPRECATED_TLS_VERSION"},
			},
			maxResults:     10,
			minMatches:     0,
			maxMatchLength: 3,
		},
		{
			name: "maxResults limited to 1",
			result: &JA4SResult{
				RiskScore:       0.0,
				AnomalyFlags:    []string{},
				MatchedProfiles: []string{},
			},
			maxResults:     1,
			minMatches:     0,
			maxMatchLength: 1,
		},
		{
			name: "maxResults limited to 2",
			result: &JA4SResult{
				RiskScore:       0.0,
				AnomalyFlags:    []string{},
				MatchedProfiles: []string{},
			},
			maxResults:     2,
			minMatches:     0,
			maxMatchLength: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matches := analyzer.FindMatchingProfiles(tc.result, tc.maxResults)

			if len(matches) < tc.minMatches {
				t.Errorf("Expected at least %d matches, got %d", tc.minMatches, len(matches))
			}

			if len(matches) > tc.maxMatchLength {
				t.Errorf("Expected at most %d matches, got %d", tc.maxMatchLength, len(matches))
			}

			// MatchedProfiles should be set by FindMatchingProfiles or pre-initialized
		})
	}
}

// TestParseServerHello tests parsing ServerHello byte data
func TestParseServerHello(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
		checkFunc   func(*testing.T, *serverHelloData)
	}{
		{
			name:        "data too short",
			data:        make([]byte, 42),
			expectError: true,
		},
		{
			name:        "no Session ID",
			data:        buildServerHelloBytes(0x0303, 0x002f, 0, []uint16{}),
			expectError: false,
			checkFunc: func(t *testing.T, sh *serverHelloData) {
				if sh.Version != 0x0303 {
					t.Errorf("Expected version 0x0303, got 0x%04x", sh.Version)
				}
				if sh.CipherSuite != 0x002f {
					t.Errorf("Expected cipher 0x002f, got 0x%04x", sh.CipherSuite)
				}
				if len(sh.Extensions) != 0 {
					t.Errorf("Expected 0 extensions, got %d", len(sh.Extensions))
				}
			},
		},
		{
			name:        "with Session ID",
			data:        buildServerHelloBytesWithSessionID(0x0303, 0x002f, 16, 0, []uint16{}),
			expectError: false,
			checkFunc: func(t *testing.T, sh *serverHelloData) {
				if sh.Version != 0x0303 {
					t.Errorf("Expected version 0x0303, got 0x%04x", sh.Version)
				}
			},
		},
		{
			name:        "with extensions",
			data:        buildServerHelloBytes(0x0304, 0x1301, 0, []uint16{0x002b, 0x0033}),
			expectError: false,
			checkFunc: func(t *testing.T, sh *serverHelloData) {
				if len(sh.Extensions) != 2 {
					t.Errorf("Expected 2 extensions, got %d", len(sh.Extensions))
				}
				if sh.Extensions[0] != 0x002b || sh.Extensions[1] != 0x0033 {
					t.Errorf("Unexpected extension values: %v", sh.Extensions)
				}
			},
		},
		{
			name:        "truncated extensions",
			data:        buildTruncatedExtensionsBytes(),
			expectError: false,
			checkFunc: func(t *testing.T, sh *serverHelloData) {
				// Extensions are truncated, should only parse up to valid portion
				if len(sh.Extensions) > 0 {
					t.Logf("Got %d extensions despite truncation", len(sh.Extensions))
				}
			},
		},
		{
			name:        "cipher suite data too short",
			data:        buildShortCipherDataBytes(),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sh, err := parseServerHello(tc.data)
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if tc.checkFunc != nil {
					tc.checkFunc(t, sh)
				}
			}
		})
	}
}

// TestFormatTLSVersion tests TLS version formatting
func TestFormatTLSVersion(t *testing.T) {
	tests := []struct {
		input    uint16
		expected string
	}{
		{0x0303, "773"},   // TLS 1.2
		{0x0304, "774"},   // TLS 1.3
		{0x0301, "769"},   // TLS 1.0
		{0x0302, "770"},   // TLS 1.1
		{0x0300, "768"},   // SSL 3.0
		{0x0000, "0"},     // Unknown
		{0xffff, "65535"}, // Unknown
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := formatTLSVersion(tc.input)
			if result != tc.expected {
				t.Errorf("formatTLSVersion(0x%04x) = %s, expected %s", tc.input, result, tc.expected)
			}
		})
	}
}

// TestTLSVersionString tests TLS version string
func TestTLSVersionString(t *testing.T) {
	tests := []struct {
		input    uint16
		expected string
	}{
		{0x0301, "1.0"},    // TLS 1.0
		{0x0302, "1.1"},    // TLS 1.1
		{0x0303, "1.2"},    // TLS 1.2
		{0x0304, "1.3"},    // TLS 1.3
		{0x0300, "0x0300"}, // SSL 3.0
		{0x0000, "0x0000"}, // Unknown
		{0xffff, "0xffff"}, // Unknown
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := tlsVersionString(tc.input)
			if result != tc.expected {
				t.Errorf("tlsVersionString(0x%04x) = %s, expected %s", tc.input, result, tc.expected)
			}
		})
	}
}

// TestFormatCipherCode tests cipher suite formatting
func TestFormatCipherCode(t *testing.T) {
	tests := []struct {
		input    uint16
		expected string
	}{
		{0x002f, "1"},     // TLS_RSA_WITH_AES_128_CBC_SHA
		{0x007c, "2"},     // TLS_RSA_WITH_AES_256_CBC_SHA
		{0x1301, "3"},     // TLS_AES_128_GCM_SHA256
		{0x1302, "4"},     // TLS_AES_256_GCM_SHA384
		{0x1303, "4867"},  // Unknown
		{0x0000, "0"},     // TLS_NULL_WITH_NULL_NULL
		{0xffff, "65535"}, // Unknown
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := formatCipherCode(tc.input)
			if result != tc.expected {
				t.Errorf("formatCipherCode(0x%04x) = %s, expected %s", tc.input, result, tc.expected)
			}
		})
	}
}

// TestFormatCompressionCode tests compression method formatting
func TestFormatCompressionCode(t *testing.T) {
	tests := []struct {
		input    uint8
		expected string
	}{
		{0, "0"},     // null compression
		{1, "1"},     // DEFLATE
		{2, "2"},     // LZS
		{255, "255"}, // Unknown
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := formatCompressionCode(tc.input)
			if result != tc.expected {
				t.Errorf("formatCompressionCode(%d) = %s, expected %s", tc.input, result, tc.expected)
			}
		})
	}
}

// TestIsSupportedTLSVersion tests supported TLS version check
func TestIsSupportedTLSVersion(t *testing.T) {
	tests := []struct {
		input    uint16
		expected bool
	}{
		{0x0303, true},  // TLS 1.2
		{0x0304, true},  // TLS 1.3
		{0x0300, false}, // SSL 3.0
		{0x0301, false}, // TLS 1.0
		{0x0302, false}, // TLS 1.1
		{0x0000, false}, // Unknown
		{0xffff, false}, // Unknown
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("0x%04x_%v", tc.input, tc.expected), func(t *testing.T) {
			result := isSupportedTLSVersion(tc.input)
			if result != tc.expected {
				t.Errorf("isSupportedTLSVersion(0x%04x) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

// TestIsDeprecatedTLSVersion tests deprecated TLS version check
func TestIsDeprecatedTLSVersion(t *testing.T) {
	tests := []struct {
		input    uint16
		expected bool
	}{
		{0x0300, true},  // SSL 3.0
		{0x0301, true},  // TLS 1.0
		{0x0302, true},  // TLS 1.1
		{0x0303, false}, // TLS 1.2
		{0x0304, false}, // TLS 1.3
		{0x0000, false}, // Unknown
		{0xffff, false}, // Unknown
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("0x%04x_%v", tc.input, tc.expected), func(t *testing.T) {
			result := isDeprecatedTLSVersion(tc.input)
			if result != tc.expected {
				t.Errorf("isDeprecatedTLSVersion(0x%04x) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

// TestIsWeakCipherSuite tests weak cipher suite check
func TestIsWeakCipherSuite(t *testing.T) {
	weakCiphers := []uint16{
		0x0000, // TLS_NULL_WITH_NULL_NULL
		0x0001, // TLS_RSA_WITH_NULL_MD5
		0x0002, // TLS_RSA_WITH_NULL_SHA
		0x0003, // TLS_RSA_EXPORT_WITH_RC4_40_MD5
		0x0004, // TLS_RSA_WITH_RC4_128_MD5
		0x0005, // TLS_RSA_WITH_RC4_128_SHA
		0x000a, // TLS_RSA_WITH_3DES_EDE_CBC_SHA (SWEET32)
	}

	strongCiphers := []uint16{
		0x002f, // TLS_RSA_WITH_AES_128_CBC_SHA
		0x007c, // TLS_RSA_WITH_AES_256_CBC_SHA
		0x1301, // TLS_AES_128_GCM_SHA256
		0x1302, // TLS_AES_256_GCM_SHA384
		0x00ff, // Unknown
	}

	for _, cipher := range weakCiphers {
		t.Run("weak_cipher", func(t *testing.T) {
			if !isWeakCipherSuite(cipher) {
				t.Errorf("Expected cipher 0x%04x to be weak", cipher)
			}
		})
	}

	for _, cipher := range strongCiphers {
		t.Run("strong_cipher", func(t *testing.T) {
			if isWeakCipherSuite(cipher) {
				t.Errorf("Expected cipher 0x%04x to be strong", cipher)
			}
		})
	}
}

// TestHasValidExtensionOrder tests extension order validation
func TestHasValidExtensionOrder(t *testing.T) {
	tests := []struct {
		name       string
		extensions []uint16
		expected   bool
	}{
		{"empty list", []uint16{}, true},
		{"single extension", []uint16{0x002b}, true},
		{"no duplicates", []uint16{0x002b, 0x0033, 0x002d}, true},
		{"has duplicates", []uint16{0x002b, 0x0033, 0x002b}, false},
		{"multiple duplicates", []uint16{0x002b, 0x002b, 0x002b}, false},
		{"trailing duplicates", []uint16{0x002b, 0x0033, 0x0033}, false},
		{"leading duplicates", []uint16{0x002b, 0x002b, 0x0033}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := hasValidExtensionOrder(tc.extensions)
			if result != tc.expected {
				t.Errorf("hasValidExtensionOrder(%v) = %v, expected %v", tc.extensions, result, tc.expected)
			}
		})
	}
}

// TestDetectAnomalies_DeprecatedTLS tests detecting deprecated TLS versions
func TestDetectAnomalies_DeprecatedTLS(t *testing.T) {
	analyzer := NewJA4SAnalyzer()

	tests := []struct {
		name          string
		version       uint16
		expectAnomaly string
		expectedScore float64
	}{
		{"SSL 3.0", 0x0300, "DEPRECATED_TLS_VERSION", 0.2},
		{"TLS 1.0", 0x0301, "DEPRECATED_TLS_VERSION", 0.2},
		{"TLS 1.1", 0x0302, "DEPRECATED_TLS_VERSION", 0.2},
		{"TLS 1.2", 0x0303, "", 0.0},
		{"TLS 1.3", 0x0304, "", 0.0},
		{"Unsupported", 0x0200, "UNSUPPORTED_TLS_VERSION", 0.3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := &JA4SResult{
				AnomalyFlags: make([]string, 0, 8),
			}
			sh := &serverHelloData{
				Version:           tc.version,
				CipherSuite:       0x1301,
				CompressionMethod: 0,
				Extensions:        []uint16{0x002b, 0x0033, 0x002d},
			}

			analyzer.detectAnomalies(result, sh)

			if tc.expectAnomaly != "" {
				found := false
				for _, flag := range result.AnomalyFlags {
					if flag == tc.expectAnomaly {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected anomaly %s, got %v", tc.expectAnomaly, result.AnomalyFlags)
				}
			}

			if result.RiskScore != tc.expectedScore {
				t.Errorf("Expected risk score %.2f, got %.2f", tc.expectedScore, result.RiskScore)
			}
		})
	}
}

// TestDetectAnomalies_WeakCipher tests detecting weak cipher suites
func TestDetectAnomalies_WeakCipher(t *testing.T) {
	analyzer := NewJA4SAnalyzer()

	tests := []struct {
		cipher        uint16
		expectAnomaly bool
	}{
		{0x0000, true},  // TLS_NULL_WITH_NULL_NULL
		{0x000a, true},  // TLS_RSA_WITH_3DES_EDE_CBC_SHA
		{0x1301, false}, // TLS_AES_128_GCM_SHA256
		{0x1302, false}, // TLS_AES_256_GCM_SHA384
	}

	for _, tc := range tests {
		t.Run(formatCipherCode(tc.cipher), func(t *testing.T) {
			result := &JA4SResult{
				AnomalyFlags: make([]string, 0, 8),
			}
			sh := &serverHelloData{
				Version:           0x0304,
				CipherSuite:       tc.cipher,
				CompressionMethod: 0,
				Extensions:        []uint16{0x002b, 0x0033, 0x002d},
			}

			analyzer.detectAnomalies(result, sh)

			found := false
			for _, flag := range result.AnomalyFlags {
				if flag == "WEAK_CIPHER_SUITE" {
					found = true
					break
				}
			}

			if tc.expectAnomaly && !found {
				t.Error("Expected WEAK_CIPHER_SUITE anomaly but not found")
			}
			if !tc.expectAnomaly && found {
				t.Error("Unexpected WEAK_CIPHER_SUITE anomaly")
			}
		})
	}
}

// TestDetectAnomalies_Extensions tests detecting extension count anomalies and duplicate extensions
func TestDetectAnomalies_Extensions(t *testing.T) {
	analyzer := NewJA4SAnalyzer()

	tests := []struct {
		name          string
		extensions    []uint16
		expectAnomaly string
	}{
		{"too few extensions", []uint16{0x002b}, "MINIMAL_EXTENSIONS"},
		{"normal extensions", []uint16{0x002b, 0x0033, 0x002d}, ""},
		{"too many extensions", make([]uint16, 31), "EXCESSIVE_EXTENSIONS"},
		{"duplicate extensions", []uint16{0x002b, 0x0033, 0x002b}, "DUPLICATE_EXTENSIONS"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := &JA4SResult{
				AnomalyFlags: make([]string, 0, 8),
			}
			sh := &serverHelloData{
				Version:           0x0304,
				CipherSuite:       0x1301,
				CompressionMethod: 0,
				Extensions:        tc.extensions,
			}

			analyzer.detectAnomalies(result, sh)

			if tc.expectAnomaly != "" {
				found := false
				for _, flag := range result.AnomalyFlags {
					if flag == tc.expectAnomaly {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected anomaly %s, got %v", tc.expectAnomaly, result.AnomalyFlags)
				}
			}
		})
	}
}

// TestDetectAnomalies_Compression tests detecting unsafe compression methods
func TestDetectAnomalies_Compression(t *testing.T) {
	analyzer := NewJA4SAnalyzer()

	tests := []struct {
		compression   uint8
		expectAnomaly bool
	}{
		{0, false}, // null compression
		{1, true},  // DEFLATE
		{2, true},  // LZS
	}

	for _, tc := range tests {
		t.Run(formatCompressionCode(tc.compression), func(t *testing.T) {
			result := &JA4SResult{
				AnomalyFlags: make([]string, 0, 8),
			}
			sh := &serverHelloData{
				Version:           0x0304,
				CipherSuite:       0x1301,
				CompressionMethod: tc.compression,
				Extensions:        []uint16{0x002b, 0x0033, 0x002d},
			}

			analyzer.detectAnomalies(result, sh)

			found := false
			for _, flag := range result.AnomalyFlags {
				if flag == "UNSAFE_COMPRESSION" {
					found = true
					break
				}
			}

			if tc.expectAnomaly && !found {
				t.Error("Expected UNSAFE_COMPRESSION anomaly but not found")
			}
			if !tc.expectAnomaly && found {
				t.Error("Unexpected UNSAFE_COMPRESSION anomaly")
			}
		})
	}
}

// TestComputeJA4SFromBytes tests convenience function ComputeJA4SFromBytes
func TestComputeJA4SFromBytes(t *testing.T) {
	// Build valid ServerHello byte data
	data := buildServerHelloBytes(0x0304, 0x1301, 0, []uint16{0x002b, 0x0033})

	result, err := ComputeJA4SFromBytes(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Hash == "" {
		t.Error("Expected non-empty hash")
	}
	if result.TLSVersion != "1.3" {
		t.Errorf("Expected TLS version 1.3, got %s", result.TLSVersion)
	}

	// Test invalid data
	_, err = ComputeJA4SFromBytes(make([]byte, 42))
	if err == nil {
		t.Error("Expected error for invalid data")
	}
}

// TestComputeJA4SFromProfileData tests computing JA4S from profile data
func TestComputeJA4SFromProfileData(t *testing.T) {
	result, err := ComputeJA4SFromProfileData(
		0x0304,                   // TLS 1.3
		0x1301,                   // TLS_AES_128_GCM_SHA256
		[]uint16{0x002b, 0x0033}, // supported_versions, key_share
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Hash == "" {
		t.Error("Expected non-empty hash")
	}
	if result.JA4Sa == "" {
		t.Error("Expected non-empty JA4Sa")
	}
	if result.JA4Sr == "" {
		t.Error("Expected non-empty JA4Sr")
	}
}

// TestMatchJA4S_Extended extends existing tests
func TestMatchJA4S_Extended(t *testing.T) {
	tests := []struct {
		name     string
		hash1    string
		hash2    string
		expected bool
	}{
		{"empty string", "", "", false},
		{"different lengths", "abc123", "abc1234", false},
		{"same hash", strings.Repeat("a", 64), strings.Repeat("a", 64), true},
		{"different hash", strings.Repeat("a", 64), strings.Repeat("b", 64), false},
		{"valid 64-char hash 1", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", true},
		{"valid 64-char hash 2", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "d3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := MatchJA4S(tc.hash1, tc.hash2)
			if result != tc.expected {
				t.Errorf("MatchJA4S(%q, %q) = %v, expected %v", tc.hash1, tc.hash2, result, tc.expected)
			}
		})
	}
}

// TestServerHelloData_String tests String() method
func TestServerHelloData_String(t *testing.T) {
	tests := []struct {
		name     string
		sh       *serverHelloData
		expected string
	}{
		{
			name: "TLS 1.3 no extensions",
			sh: &serverHelloData{
				Version:           0x0304,
				CipherSuite:       0x1301,
				CompressionMethod: 0,
				Extensions:        []uint16{},
			},
			expected: "TLS304,Cipher1301,Comp0,Ext[]",
		},
		{
			name: "TLS 1.2 with extensions",
			sh: &serverHelloData{
				Version:           0x0303,
				CipherSuite:       0x002f,
				CompressionMethod: 0,
				Extensions:        []uint16{0x002b, 0x0033},
			},
			expected: "TLS303,Cipher2f,Comp0,Ext[43,51]",
		},
		{
			name: "complex configuration",
			sh: &serverHelloData{
				Version:           0x0303,
				CipherSuite:       0x007c,
				CompressionMethod: 1,
				Extensions:        []uint16{0x0000, 0x000a, 0x000b, 0xff01},
			},
			expected: "TLS303,Cipher7c,Comp1,Ext[0,10,11,65281]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.sh.String()
			if result != tc.expected {
				t.Errorf("String() = %q, expected %q", result, tc.expected)
			}
		})
	}
}

// ==================== Helper Functions ====================

// buildServerHelloBytes builds ServerHello byte data
func buildServerHelloBytes(version, cipher uint16, sessionIDLen int, extensions []uint16) []byte {
	// Calculate total length
	baseLen := 43 + sessionIDLen // Base length + Session ID
	extLen := 0
	for range extensions {
		extLen += 4 + 4 // Type(2) + Len(2) + DummyData(4) for simplicity
	}
	if len(extensions) > 0 {
		extLen += 2 // Extensions Length field
	}

	data := make([]byte, baseLen+extLen)
	offset := 0

	data[offset] = 0x02
	offset++
	// Length (3 bytes)
	data[offset] = 0x00
	data[offset+1] = byte((baseLen + extLen - 4) >> 8)
	data[offset+2] = byte((baseLen + extLen - 4) & 0xff)
	offset += 3

	// Version
	data[offset] = byte(version >> 8)
	data[offset+1] = byte(version & 0xff)
	offset += 2

	// Random (32 bytes)
	offset += 32

	// Session ID Length
	data[offset] = byte(sessionIDLen)
	offset++

	// Session ID
	offset += sessionIDLen

	// Cipher Suite
	data[offset] = byte(cipher >> 8)
	data[offset+1] = byte(cipher & 0xff)
	offset += 2

	// Compression Method
	data[offset] = 0x00
	offset++

	// Extensions
	if len(extensions) > 0 {
		// Extensions Length
		extTotalLen := len(extensions) * 8 // 8 bytes per extension
		data[offset] = byte(extTotalLen >> 8)
		data[offset+1] = byte(extTotalLen & 0xff)
		offset += 2

		for _, ext := range extensions {
			// Type
			data[offset] = byte(ext >> 8)
			data[offset+1] = byte(ext & 0xff)
			offset += 2
			// Length (4 bytes dummy data)
			data[offset] = 0x00
			data[offset+1] = 0x04
			offset += 2
			// Dummy data
			data[offset] = 0x00
			data[offset+1] = 0x00
			data[offset+2] = 0x00
			data[offset+3] = 0x00
			offset += 4
		}
	}

	return data
}

// buildServerHelloBytesWithSessionID builds ServerHello byte data with Session ID
func buildServerHelloBytesWithSessionID(version, cipher uint16, sessionIDLen int, compression uint8, extensions []uint16) []byte {
	return buildServerHelloBytes(version, cipher, sessionIDLen, extensions)
}

// buildTruncatedExtensionsBytes builds a ServerHello with truncated extensions
func buildTruncatedExtensionsBytes() []byte {
	// Base 43 bytes + Session ID Len(1) + Cipher(2) + Compression(1) + Extensions Length(2)
	data := make([]byte, 49)
	offset := 0

	data[offset] = 0x02
	offset++
	data[offset] = 0x00
	data[offset+1] = 0x00
	data[offset+2] = 0x2d // Length = 45
	offset += 3

	// Version TLS 1.2
	data[offset] = 0x03
	data[offset+1] = 0x03
	offset += 2

	// Random
	offset += 32

	// Session ID Length = 0
	data[offset] = 0x00
	offset++

	// Cipher
	data[offset] = 0x00
	data[offset+1] = 0x2f
	offset += 2

	// Compression
	data[offset] = 0x00
	offset++

	// Extensions Length = 20 (but data only has 10)
	data[offset] = 0x00
	data[offset+1] = 0x14
	offset += 2

	// Add an incomplete extension
	data[offset] = 0x00
	data[offset+1] = 0x2b
	offset += 2
	data[offset] = 0x00
	data[offset+1] = 0x10 // Length = 16 but no data follows

	return data
}

// buildShortCipherDataBytes builds data with cipher suite data too short
func buildShortCipherDataBytes() []byte {
	// Just up to end of Session ID, but no Cipher Suite
	data := make([]byte, 42)
	offset := 0

	data[offset] = 0x02
	offset++
	data[offset] = 0x00
	data[offset+1] = 0x00
	data[offset+2] = 0x26 // Length
	offset += 3

	// Version
	data[offset] = 0x03
	data[offset+1] = 0x03
	offset += 2

	// Random
	offset += 32

	// Session ID Length = 0
	data[offset] = 0x00
	offset++

	// There should be two bytes of Cipher Suite here, but the data has ended

	return data
}
