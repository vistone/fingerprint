package ja4s

import (
	"testing"

	"github.com/vistone/fingerprint/profiles"
)

// TestComputeJA4SFromRealProfiles 使用真实指纹配置测试 JA4S 计算
func TestComputeJA4SFromRealProfiles(t *testing.T) {
	// 使用真实的 Chrome 133 指纹
	chromeProfile, ok := profiles.MappedTLSClients["chrome_133"]
	if !ok {
		t.Fatal("chrome_133 profile not found")
	}

	// 获取 ClientHelloSpec
	clientSpec, err := chromeProfile.GetClientHelloSpec()
	if err != nil {
		t.Skipf("chrome_133 does not support spec export: %v", err)
		return
	}

	// 构造 ServerHello 数据（基于真实 ClientHello 的响应）
	// 使用与 Chrome 133 兼容的 TLS 1.3 配置
	serverHello := ServerHelloData{
		TLSVersion:     0x0304, // TLS 1.3
		CipherSuite:    0x1301, // TLS_AES_128_GCM_SHA256 (Chrome 首选)
		Extensions:     []uint16{0x002b, 0x0033}, // supported_versions, key_share
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
