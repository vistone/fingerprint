package ja4s

import (
	"fmt"
	"strings"
	"testing"

	"github.com/vistone/fingerprint/modules/profiles/legacy"
)

// TestComputeJA4SFromRealProfiles 使用真实指纹配置测试 JA4S 计算
func TestComputeJA4SFromRealProfiles(t *testing.T) {
	// 使用真实的 Chrome 133 指纹
	chromeProfile, ok := profiles.MappedTLSClients["chrome_133"]
	if !ok {
		t.Fatal("chrome_133 profile not found")
	}

	// 获取 ClientHelloSpec
	_, err := chromeProfile.GetClientHelloSpec()
	if err != nil {
		t.Skipf("chrome_133 does not support spec export: %v", err)
		return
	}

	// 构造 ServerHello 数据（基于真实 ClientHello 的响应）
	// 使用与 Chrome 133 兼容的 TLS 1.3 配置
	serverHello := ServerHelloData{
		TLSVersion:     0x0304,                       // TLS 1.3
		CipherSuite:    0x1301,                       // TLS_AES_128_GCM_SHA256 (Chrome 首选)
		Extensions:     []uint16{0x002b, 0x0033},     // supported_versions, key_share
		Compression:    0,
		ServerName:     "www.google.com",
		SelectedALPN:   "h2",
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

// TestComputeJA4SFromMultipleProfiles 测试多个真实指纹的 JA4S 计算
func TestComputeJA4SFromMultipleProfiles(t *testing.T) {
	testCases := []struct {
		profileName string
		cipherSuite uint16
		extensions  []uint16
	}{
		{"chrome_133", 0x1301, []uint16{0x002b, 0x0033, 0x002d}}, // TLS_AES_128_GCM_SHA256
		{"firefox_135", 0x1302, []uint16{0x002b, 0x0033}},       // TLS_AES_256_GCM_SHA384
		{"safari_16_0", 0x1301, []uint16{0x002b, 0x0033}},       // TLS_AES_128_GCM_SHA256
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

// TestAnalyzeServerHelloWithRealData 使用真实 ServerHello 数据测试分析
func TestAnalyzeServerHelloWithRealData(t *testing.T) {
	analyzer := NewJA4SAnalyzer()

	// 测试 TLS 1.3 ServerHello（模拟真实服务器响应）
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

// TestMatchJA4S 测试 JA4S 哈希匹配
func TestMatchJA4S(t *testing.T) {
	serverHello := ServerHelloData{
		TLSVersion:  0x0304,
		CipherSuite: 0x1301,
		Extensions:  []uint16{0x002b, 0x0033},
		Compression: 0,
	}

	result1, _ := ComputeJA4S(serverHello)
	result2, _ := ComputeJA4S(serverHello)

	// 相同输入应该产生相同哈希
	if !MatchJA4S(result1.Hash, result2.Hash) {
		t.Error("Same ServerHello should produce matching JA4S hashes")
	}

	// 不同输入应该产生不同哈希
	serverHello.CipherSuite = 0x1302 // 改变密码套件
	result3, _ := ComputeJA4S(serverHello)

	if MatchJA4S(result1.Hash, result3.Hash) {
		t.Error("Different ServerHello should produce different JA4S hashes")
	}
}

// TestNewJA4SAnalyzer 测试创建分析器时已知配置库已初始化
func TestNewJA4SAnalyzer(t *testing.T) {
	analyzer := NewJA4SAnalyzer()
	if analyzer == nil {
		t.Fatal("NewJA4SAnalyzer should return a non-nil analyzer")
	}
	if analyzer.knownProfiles == nil {
		t.Error("Known profiles should be initialized")
	}

	// 验证已知配置文件是否存在
	expectedProfiles := []string{"nginx_default", "apache_default", "cloudflare"}
	for _, name := range expectedProfiles {
		if _, ok := analyzer.knownProfiles[name]; !ok {
			t.Errorf("Expected known profile %s not found", name)
		}
	}
}

// TestAnalyzeServerHelloBytes 测试分析 TLS ServerHello 字节数据
func TestAnalyzeServerHelloBytes(t *testing.T) {
	analyzer := NewJA4SAnalyzer()

	tests := []struct {
		name        string
		data        []byte
		expectError bool
		errContains string
	}{
		{
			name:        "数据太短",
			data:        make([]byte, 42),
			expectError: true,
			errContains: "ServerHello too short",
		},
		{
			name:        "数据刚好43字节但结构无效",
			data:        make([]byte, 43),
			expectError: false, // 可以解析但可能结果不正确
		},
		{
			name:        "空数据",
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

	// 测试有效的 ServerHello 字节数据
	t.Run("有效的 ServerHello 字节数据", func(t *testing.T) {
		// 构建一个有效的 TLS ServerHello 消息
		// HandshakeType(1) + Length(3) + Version(2) + Random(32) + SessionIDLen(1) + SessionID(0) + CipherSuite(2) + Compression(1)
		// 总共 42 字节基础，加上扩展长度 2 字节
		data := make([]byte, 64)
		data[0] = 0x02                                           // Server Hello
		data[1], data[2], data[3] = 0x00, 0x00, 0x3d             // Length: 61 bytes
		data[4], data[5] = 0x03, 0x04                            // Version: TLS 1.3
		copy(data[6:38], make([]byte, 32))                       // Random (32 bytes of zeros)
		data[38] = 0x00                                          // Session ID Length: 0
		data[39], data[40] = 0x13, 0x01                          // Cipher Suite: TLS_AES_128_GCM_SHA256
		data[41] = 0x00                                          // Compression: null
		data[42], data[43] = 0x00, 0x12                          // Extensions Length: 18 bytes
		// Extension 1: supported_versions (0x002b)
		data[44], data[45] = 0x00, 0x2b                          // Type
		data[46], data[47] = 0x00, 0x02                          // Length
		data[48], data[49] = 0x03, 0x04                          // TLS 1.3
		// Extension 2: key_share (0x0033)
		data[50], data[51] = 0x00, 0x33                          // Type
		data[52], data[53] = 0x00, 0x06                          // Length
		data[54], data[55] = 0x00, 0x17                          // Named Group: secp256r1
		data[56], data[57] = 0x00, 0x02                          // Length
		data[58], data[59] = 0x00, 0x00                          // Key exchange data

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

	// 测试解析扩展列表
	t.Run("解析扩展列表", func(t *testing.T) {
		// 构建一个包含多个扩展的 ServerHello
		data := make([]byte, 80)
		data[0] = 0x02                                           // Server Hello
		data[1], data[2], data[3] = 0x00, 0x00, 0x49             // Length
		data[4], data[5] = 0x03, 0x03                            // Version: TLS 1.2
		copy(data[6:38], make([]byte, 32))                       // Random
		data[38] = 0x00                                          // Session ID Length: 0
		data[39], data[40] = 0x00, 0x2f                          // Cipher Suite: TLS_RSA_WITH_AES_128_CBC_SHA
		data[41] = 0x00                                          // Compression: null
		data[42], data[43] = 0x00, 0x20                          // Extensions Length: 32 bytes

		// Extension 1: server_name (0x0000)
		data[44], data[45] = 0x00, 0x00                          // Type
		data[46], data[47] = 0x00, 0x00                          // Length: 0

		// Extension 2: supported_groups (0x000a)
		data[48], data[49] = 0x00, 0x0a                          // Type
		data[50], data[51] = 0x00, 0x04                          // Length
		data[52], data[53] = 0x00, 0x02                          // List Length
		data[54], data[55] = 0x00, 0x17                          // secp256r1

		// Extension 3: ec_point_formats (0x000b)
		data[56], data[57] = 0x00, 0x0b                          // Type
		data[58], data[59] = 0x00, 0x02                          // Length
		data[60], data[61] = 0x01, 0x00                          // Length 1, uncompressed

		// Extension 4: session_ticket (0x0023)
		data[62], data[63] = 0x00, 0x23                          // Type
		data[64], data[65] = 0x00, 0x00                          // Length: 0

		// Extension 5: renegotiation_info (0xff01)
		data[66], data[67] = 0xff, 0x01                         // Type
		data[68], data[69] = 0x00, 0x01                          // Length
		data[70] = 0x00                                          // Info

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

// TestGenerateServerHelloSignature 测试生成签名和异常检测
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
			name:              "TLS 1.3 正常配置",
			tlsVersion:        0x0304,
			cipherSuite:       0x1301,
			extensions:        []uint16{0x002b, 0x0033, 0x002d},
			compressionMethod: 0,
			expectHashLen:     64,
			expectRiskScore:   0.0,
			expectAnomalies:   nil,
		},
		{
			name:              "TLS 1.2 正常配置",
			tlsVersion:        0x0303,
			cipherSuite:       0x002f,
			extensions:        []uint16{0x0000, 0x000a, 0x000b},
			compressionMethod: 0,
			expectHashLen:     64,
			expectRiskScore:   0.0,
			expectAnomalies:   nil,
		},
		{
			name:              "TLS 1.0 已弃用版本",
			tlsVersion:        0x0301,
			cipherSuite:       0x1301,
			extensions:        []uint16{0x002b, 0x0033, 0x002d},
			compressionMethod: 0,
			expectHashLen:     64,
			expectRiskScore:   0.2,
			expectAnomalies:   []string{"DEPRECATED_TLS_VERSION"},
		},
		{
			name:              "弱密码套件",
			tlsVersion:        0x0303,
			cipherSuite:       0x000a, // 3DES
			extensions:        []uint16{0x002b, 0x0033, 0x002d},
			compressionMethod: 0,
			expectHashLen:     64,
			expectRiskScore:   0.25,
			expectAnomalies:   []string{"WEAK_CIPHER_SUITE"},
		},
		{
			name:              "扩展太少",
			tlsVersion:        0x0304,
			cipherSuite:       0x1301,
			extensions:        []uint16{0x002b},
			compressionMethod: 0,
			expectHashLen:     64,
			expectRiskScore:   0.2,
			expectAnomalies:   []string{"MINIMAL_EXTENSIONS"},
		},
		{
			name:              "不安全的压缩方法",
			tlsVersion:        0x0304,
			cipherSuite:       0x1301,
			extensions:        []uint16{0x002b, 0x0033, 0x002d},
			compressionMethod: 1,
			expectHashLen:     64,
			expectRiskScore:   0.2,
			expectAnomalies:   []string{"UNSAFE_COMPRESSION"},
		},
		{
			name:              "多个异常",
			tlsVersion:        0x0301, // 已弃用
			cipherSuite:       0x000a, // 弱密码
			extensions:        []uint16{0x002b},
			compressionMethod: 1,      // 不安全压缩
			expectHashLen:     64,
			expectRiskScore:   0.85,   // 0.2 + 0.25 + 0.2 + 0.2
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

// TestFindMatchingProfiles 测试查找匹配的已知服务端配置
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
			name: "低风险分数应该匹配",
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
			name: "高风险分数可能不匹配",
			result: &JA4SResult{
				RiskScore:    0.5,
				AnomalyFlags: []string{"DEPRECATED_TLS_VERSION"},
			},
			maxResults:     10,
			minMatches:     0,
			maxMatchLength: 3,
		},
		{
			name: "maxResults 限制为 1",
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
			name: "maxResults 限制为 2",
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

// TestParseServerHello 测试解析 ServerHello 字节数据
func TestParseServerHello(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
		checkFunc   func(*testing.T, *serverHelloData)
	}{
		{
			name:        "数据太短",
			data:        make([]byte, 42),
			expectError: true,
		},
		{
			name:        "无 Session ID",
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
			name:        "有 Session ID",
			data:        buildServerHelloBytesWithSessionID(0x0303, 0x002f, 16, 0, []uint16{}),
			expectError: false,
			checkFunc: func(t *testing.T, sh *serverHelloData) {
				if sh.Version != 0x0303 {
					t.Errorf("Expected version 0x0303, got 0x%04x", sh.Version)
				}
			},
		},
		{
			name:        "有扩展",
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
			name:        "扩展被截断",
			data:        buildTruncatedExtensionsBytes(),
			expectError: false,
			checkFunc: func(t *testing.T, sh *serverHelloData) {
				// 扩展被截断，应该只解析到有效的部分
				if len(sh.Extensions) > 0 {
					t.Logf("Got %d extensions despite truncation", len(sh.Extensions))
				}
			},
		},
		{
			name:        "密码套件数据太短",
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

// TestFormatTLSVersion 测试 TLS 版本格式化
func TestFormatTLSVersion(t *testing.T) {
	tests := []struct {
		input    uint16
		expected string
	}{
		{0x0303, "773"},    // TLS 1.2
		{0x0304, "774"},    // TLS 1.3
		{0x0301, "769"},    // TLS 1.0
		{0x0302, "770"},    // TLS 1.1
		{0x0300, "768"},    // SSL 3.0
		{0x0000, "0"},      // Unknown
		{0xffff, "65535"},  // Unknown
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

// TestTLSVersionString 测试 TLS 版本字符串
func TestTLSVersionString(t *testing.T) {
	tests := []struct {
		input    uint16
		expected string
	}{
		{0x0301, "1.0"},      // TLS 1.0
		{0x0302, "1.1"},      // TLS 1.1
		{0x0303, "1.2"},      // TLS 1.2
		{0x0304, "1.3"},      // TLS 1.3
		{0x0300, "0x0300"},   // SSL 3.0
		{0x0000, "0x0000"},   // Unknown
		{0xffff, "0xffff"},   // Unknown
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

// TestFormatCipherCode 测试密码套件格式化
func TestFormatCipherCode(t *testing.T) {
	tests := []struct {
		input    uint16
		expected string
	}{
		{0x002f, "1"},       // TLS_RSA_WITH_AES_128_CBC_SHA
		{0x007c, "2"},       // TLS_RSA_WITH_AES_256_CBC_SHA
		{0x1301, "3"},       // TLS_AES_128_GCM_SHA256
		{0x1302, "4"},       // TLS_AES_256_GCM_SHA384
		{0x1303, "4867"},    // Unknown
		{0x0000, "0"},       // TLS_NULL_WITH_NULL_NULL
		{0xffff, "65535"},   // Unknown
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

// TestFormatCompressionCode 测试压缩方法格式化
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

// TestIsSupportedTLSVersion 测试支持的 TLS 版本检查
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

// TestIsDeprecatedTLSVersion 测试已弃用 TLS 版本检查
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

// TestIsWeakCipherSuite 测试弱密码套件检查
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

// TestHasValidExtensionOrder 测试扩展顺序验证
func TestHasValidExtensionOrder(t *testing.T) {
	tests := []struct {
		name     string
		extensions []uint16
		expected bool
	}{
		{"空列表", []uint16{}, true},
		{"单扩展", []uint16{0x002b}, true},
		{"无重复", []uint16{0x002b, 0x0033, 0x002d}, true},
		{"有重复", []uint16{0x002b, 0x0033, 0x002b}, false},
		{"多个重复", []uint16{0x002b, 0x002b, 0x002b}, false},
		{"末尾重复", []uint16{0x002b, 0x0033, 0x0033}, false},
		{"开头重复", []uint16{0x002b, 0x002b, 0x0033}, false},
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

// TestDetectAnomalies_DeprecatedTLS 测试检测已弃用的 TLS 版本
func TestDetectAnomalies_DeprecatedTLS(t *testing.T) {
	analyzer := NewJA4SAnalyzer()

	tests := []struct {
		name           string
		version        uint16
		expectAnomaly  string
		expectedScore  float64
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

// TestDetectAnomalies_WeakCipher 测试检测弱密码套件
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

// TestDetectAnomalies_Extensions 测试检测扩展数量异常和重复扩展
func TestDetectAnomalies_Extensions(t *testing.T) {
	analyzer := NewJA4SAnalyzer()

	tests := []struct {
		name            string
		extensions      []uint16
		expectAnomaly   string
	}{
		{"太少扩展", []uint16{0x002b}, "MINIMAL_EXTENSIONS"},
		{"正常扩展", []uint16{0x002b, 0x0033, 0x002d}, ""},
		{"太多扩展", make([]uint16, 31), "EXCESSIVE_EXTENSIONS"},
		{"重复扩展", []uint16{0x002b, 0x0033, 0x002b}, "DUPLICATE_EXTENSIONS"},
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

// TestDetectAnomalies_Compression 测试检测不安全的压缩方法
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

// TestComputeJA4SFromBytes 测试便捷函数 ComputeJA4SFromBytes
func TestComputeJA4SFromBytes(t *testing.T) {
	// 构建有效的 ServerHello 字节数据
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

	// 测试无效数据
	_, err = ComputeJA4SFromBytes(make([]byte, 42))
	if err == nil {
		t.Error("Expected error for invalid data")
	}
}

// TestComputeJA4SFromProfileData 测试从 Profile 数据计算 JA4S
func TestComputeJA4SFromProfileData(t *testing.T) {
	result, err := ComputeJA4SFromProfileData(
		0x0304,                       // TLS 1.3
		0x1301,                       // TLS_AES_128_GCM_SHA256
		[]uint16{0x002b, 0x0033},     // supported_versions, key_share
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

// TestMatchJA4S_Extended 扩展现有测试
func TestMatchJA4S_Extended(t *testing.T) {
	tests := []struct {
		name     string
		hash1    string
		hash2    string
		expected bool
	}{
		{"空字符串", "", "", false},
		{"不同长度", "abc123", "abc1234", false},
		{"相同哈希", strings.Repeat("a", 64), strings.Repeat("a", 64), true},
		{"不同哈希", strings.Repeat("a", 64), strings.Repeat("b", 64), false},
		{"有效64位哈希1", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", true},
		{"有效64位哈希2", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "d3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", false},
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

// TestServerHelloData_String 测试 String() 方法
func TestServerHelloData_String(t *testing.T) {
	tests := []struct {
		name     string
		sh       *serverHelloData
		expected string
	}{
		{
			name: "TLS 1.3 无扩展",
			sh: &serverHelloData{
				Version:           0x0304,
				CipherSuite:       0x1301,
				CompressionMethod: 0,
				Extensions:        []uint16{},
			},
			expected: "TLS304,Cipher1301,Comp0,Ext[]",
		},
		{
			name: "TLS 1.2 有扩展",
			sh: &serverHelloData{
				Version:           0x0303,
				CipherSuite:       0x002f,
				CompressionMethod: 0,
				Extensions:        []uint16{0x002b, 0x0033},
			},
			expected: "TLS303,Cipher2f,Comp0,Ext[43,51]",
		},
		{
			name: "复杂配置",
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

// ==================== 辅助函数 ====================

// buildServerHelloBytes 构建 ServerHello 字节数据
func buildServerHelloBytes(version, cipher uint16, sessionIDLen int, extensions []uint16) []byte {
	// 计算总长度
	baseLen := 43 + sessionIDLen // 基础长度 + Session ID
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
		extTotalLen := len(extensions) * 8 // 每个扩展 8 字节
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

// buildServerHelloBytesWithSessionID 构建带 Session ID 的 ServerHello 字节数据
func buildServerHelloBytesWithSessionID(version, cipher uint16, sessionIDLen int, compression uint8, extensions []uint16) []byte {
	return buildServerHelloBytes(version, cipher, sessionIDLen, extensions)
}

// buildTruncatedExtensionsBytes 构建扩展被截断的 ServerHello
func buildTruncatedExtensionsBytes() []byte {
	// 基础 43 字节 + Session ID Len(1) + Cipher(2) + Compression(1) + Extensions Length(2)
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

	// Extensions Length = 20 (但数据只有 10)
	data[offset] = 0x00
	data[offset+1] = 0x14
	offset += 2

	// 添加一个不完整的扩展
	data[offset] = 0x00
	data[offset+1] = 0x2b
	offset += 2
	data[offset] = 0x00
	data[offset+1] = 0x10 // Length = 16 but no data follows

	return data
}

// buildShortCipherDataBytes 构建密码套件数据太短的数据
func buildShortCipherDataBytes() []byte {
	// 刚好到 Session ID 结束，但没有 Cipher Suite
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

	// 这里应该有两个字节的 Cipher Suite，但数据已经结束了

	return data
}
