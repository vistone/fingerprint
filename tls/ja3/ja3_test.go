package ja3

import (
	"fmt"
	"strings"
	"testing"

	tls "github.com/bogdanfinn/utls"
	"github.com/vistone/fingerprint/profiles"
)

// TestComputeJA3FromRealProfiles 使用真实指纹配置测试 JA3 计算
func TestComputeJA3FromRealProfiles(t *testing.T) {
	// 使用真实的 Chrome 133 指纹
	chromeProfile, ok := profiles.MappedTLSClients["chrome_133"]
	if !ok {
		t.Fatal("chrome_133 profile not found")
	}

	spec, err := chromeProfile.GetClientHelloSpec()
	if err != nil {
		t.Fatalf("Failed to get ClientHelloSpec for chrome_133: %v", err)
	}

	result, err := ComputeJA3FromSpec(spec)
	if err != nil {
		t.Fatalf("ComputeJA3FromSpec failed: %v", err)
	}

	if result.Hash == "" {
		t.Error("Expected non-empty hash for chrome_133")
	}

	if result.RawString == "" {
		t.Error("Expected non-empty raw string for chrome_133")
	}

	t.Logf("Chrome 133 JA3 Hash: %s", result.Hash)
	t.Logf("Chrome 133 JA3 Raw: %s", result.RawString)

	// 验证 Hash 是有效的 MD5（32 个十六进制字符）
	if len(result.Hash) != 32 {
		t.Errorf("Expected MD5 hash length 32, got %d", len(result.Hash))
	}
}

// TestComputeJA3FromMultipleProfiles 测试多个真实指纹的 JA3 计算
func TestComputeJA3FromMultipleProfiles(t *testing.T) {
	testProfiles := []string{
		"chrome_133",
		"chrome_120",
		"firefox_135",
		"safari_16_0",
	}

	for _, profileName := range testProfiles {
		t.Run(profileName, func(t *testing.T) {
			profile, ok := profiles.MappedTLSClients[profileName]
			if !ok {
				t.Skipf("Profile %s not found", profileName)
				return
			}

			spec, err := profile.GetClientHelloSpec()
			if err != nil {
				t.Skipf("Profile %s does not support spec export: %v", profileName, err)
				return
			}

			result, err := ComputeJA3FromSpec(spec)
			if err != nil {
				t.Fatalf("ComputeJA3FromSpec failed for %s: %v", profileName, err)
			}

			if result.Hash == "" {
				t.Errorf("Expected non-empty hash for %s", profileName)
			}

			if result.TLSVersion == 0 {
				t.Errorf("Expected non-zero TLS version for %s", profileName)
			}

			t.Logf("%s: JA3=%s, TLSVersion=%d, Ciphers=%d, Extensions=%d",
				profileName, result.Hash, result.TLSVersion,
				len(result.CipherSuites), len(result.Extensions))
		})
	}
}

// TestMatchJA3WithRealHashes 使用真实 JA3 哈希测试匹配功能
func TestMatchJA3WithRealHashes(t *testing.T) {
	// 获取真实指纹的 JA3 哈希
	profile, ok := profiles.MappedTLSClients["chrome_133"]
	if !ok {
		t.Fatal("chrome_133 profile not found")
	}

	spec, err := profile.GetClientHelloSpec()
	if err != nil {
		t.Skipf("chrome_133 does not support spec export: %v", err)
		return
	}

	result, err := ComputeJA3FromSpec(spec)
	if err != nil {
		t.Fatalf("ComputeJA3FromSpec failed: %v", err)
	}

	realHash := result.Hash

	// 测试相同哈希匹配
	if !MatchJA3(realHash, realHash) {
		t.Error("Same hash should match")
	}

	// 测试不同哈希不匹配（修改一个字符）
	modifiedHash := realHash[:31] + "x"
	if MatchJA3(realHash, modifiedHash) {
		t.Error("Different hashes should not match")
	}

	// 测试大小写不敏感匹配
	upperHash := "ABCDEF1234567890ABCDEF1234567890"
	lowerHash := "abcdef1234567890abcdef1234567890"
	if !MatchJA3(upperHash, lowerHash) {
		t.Error("Case-insensitive match should work")
	}
}

// TestFindProfileByJA3WithRealHashes 使用真实 JA3 哈希测试查找功能
func TestFindProfileByJA3WithRealHashes(t *testing.T) {
	// 为几个真实指纹计算 JA3
	testProfiles := []string{"chrome_133", "firefox_135"}

	for _, profileName := range testProfiles {
		profile, ok := profiles.MappedTLSClients[profileName]
		if !ok {
			continue
		}

		spec, err := profile.GetClientHelloSpec()
		if err != nil {
			continue
		}

		result, err := ComputeJA3FromSpec(spec)
		if err != nil {
			continue
		}

		// 使用真实哈希查找 profile
		foundProfiles := FindProfileByJA3(result.Hash)
		t.Logf("Profile %s has JA3 hash %s, found profiles: %v",
			profileName, result.Hash, foundProfiles)
	}

	// 测试空哈希
	emptyResult := FindProfileByJA3("")
	if len(emptyResult) != 0 {
		t.Error("Expected no profiles for empty hash")
	}
}

// TestValidateClientHelloSpec 测试 ClientHello 规范验证
func TestValidateClientHelloSpec(t *testing.T) {
	tests := []struct {
		name    string
		spec    tls.ClientHelloSpec
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid spec",
			spec: tls.ClientHelloSpec{
				CipherSuites: []uint16{0x1301, 0x1302},
				Extensions:   []tls.TLSExtension{&tls.SNIExtension{}},
			},
			wantErr: false,
		},
		{
			name: "empty cipher suites",
			spec: tls.ClientHelloSpec{
				CipherSuites: []uint16{},
				Extensions:   []tls.TLSExtension{&tls.SNIExtension{}},
			},
			wantErr: true,
			errMsg:  "cipher suites list is empty",
		},
		{
			name: "too many cipher suites",
			spec: tls.ClientHelloSpec{
				CipherSuites: make([]uint16, 300),
				Extensions:   []tls.TLSExtension{&tls.SNIExtension{}},
			},
			wantErr: true,
			errMsg:  "cipher suites list too long",
		},
		{
			name: "empty extensions",
			spec: tls.ClientHelloSpec{
				CipherSuites: []uint16{0x1301},
				Extensions:   []tls.TLSExtension{},
			},
			wantErr: true,
			errMsg:  "extensions list is empty",
		},
		{
			name: "too many extensions",
			spec: tls.ClientHelloSpec{
				CipherSuites: []uint16{0x1301},
				Extensions:   make([]tls.TLSExtension, 300),
			},
			wantErr: true,
			errMsg:  "extensions list too long",
		},
		{
			name: "zero cipher suite value",
			spec: tls.ClientHelloSpec{
				CipherSuites: []uint16{0x1301, 0x0000, 0x1302},
				Extensions:   []tls.TLSExtension{&tls.SNIExtension{}},
			},
			wantErr: true,
			errMsg:  "cipher suite at index 1 is zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateClientHelloSpec(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateClientHelloSpec() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateClientHelloSpec() error = %v, want to contain %q", err, tt.errMsg)
			}
		})
	}
}

// TestComputeJA3FromSpecWithInvalidInput 测试无效输入处理
func TestComputeJA3FromSpecWithInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		spec    tls.ClientHelloSpec
		wantErr bool
	}{
		{
			name: "nil extensions",
			spec: tls.ClientHelloSpec{
				CipherSuites: []uint16{0x1301},
				Extensions:   nil,
			},
			wantErr: true,
		},
		{
			name: "empty cipher suites",
			spec: tls.ClientHelloSpec{
				CipherSuites: []uint16{},
				Extensions:   []tls.TLSExtension{&tls.SNIExtension{}},
			},
			wantErr: true,
		},
		{
			name: "extension with nil data",
			spec: tls.ClientHelloSpec{
				CipherSuites: []uint16{0x1301},
				Extensions:   []tls.TLSExtension{nil},
			},
			wantErr: false, // 应该跳过 nil 扩展
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ComputeJA3FromSpec(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComputeJA3FromSpec() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result == nil {
				t.Error("Expected non-nil result for valid input")
			}
		})
	}
}

// TestComputeJA3FromSpecEdgeCases 测试边界情况
func TestComputeJA3FromSpecEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		spec     tls.ClientHelloSpec
		validate func(*testing.T, *JA3Result, error)
	}{
		{
			name: "single cipher suite",
			spec: tls.ClientHelloSpec{
				CipherSuites: []uint16{0x1301},
				Extensions:   []tls.TLSExtension{&tls.SNIExtension{}},
			},
			validate: func(t *testing.T, result *JA3Result, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if len(result.CipherSuites) != 1 {
					t.Errorf("Expected 1 cipher suite, got %d", len(result.CipherSuites))
				}
			},
		},
		{
			name: "all grease values",
			spec: tls.ClientHelloSpec{
				CipherSuites: []uint16{0x0A0A, 0x1A1A}, // GREASE values
				Extensions:   []tls.TLSExtension{&tls.UtlsGREASEExtension{}},
			},
			validate: func(t *testing.T, result *JA3Result, err error) {
				if err == nil {
					t.Error("Expected error when all cipher suites are GREASE")
				}
			},
		},
		{
			name: "multiple extensions of same type",
			spec: tls.ClientHelloSpec{
				CipherSuites: []uint16{0x1301, 0x1302},
				Extensions: []tls.TLSExtension{
					&tls.SNIExtension{},
					&tls.SNIExtension{},
					&tls.SNIExtension{},
				},
			},
			validate: func(t *testing.T, result *JA3Result, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				// 应该正确处理重复扩展
				if len(result.Extensions) < 1 {
					t.Error("Expected at least one extension")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ComputeJA3FromSpec(tt.spec)
			tt.validate(t, result, err)
		})
	}
}

// TestIsGREASEValue 测试 GREASE 值检测
func TestIsGREASEValue(t *testing.T) {
	tests := []struct {
		value uint16
		want  bool
	}{
		{0x0A0A, true},
		{0x1A1A, true},
		{0x2A2A, true},
		{0x3A3A, true},
		{0x4A4A, true},
		{0x5A5A, true},
		{0x6A6A, true},
		{0x7A7A, true},
		{0x8A8A, true},
		{0x9A9A, true},
		{0xAAAA, true},
		{0xBABA, true},
		{0xCACA, true},
		{0xDADA, true},
		{0xEAEA, true},
		{0xFAFA, true},
		{0x1301, false}, // 正常密码套件
		{0x0000, false},
		{0xFFFF, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("0x%04X", tt.value), func(t *testing.T) {
			if got := isGREASEValue(tt.value); got != tt.want {
				t.Errorf("isGREASEValue(0x%04X) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

// TestFilterGREASEFunctions 测试 GREASE 过滤函数
func TestFilterGREASEFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint16
		expected []uint16
	}{
		{
			name:     "no grease",
			input:    []uint16{0x1301, 0x1302, 0x1303},
			expected: []uint16{0x1301, 0x1302, 0x1303},
		},
		{
			name:     "all grease",
			input:    []uint16{0x0A0A, 0x1A1A, 0x2A2A},
			expected: []uint16{},
		},
		{
			name:     "mixed values",
			input:    []uint16{0x1301, 0x0A0A, 0x1302, 0x2A2A},
			expected: []uint16{0x1301, 0x1302},
		},
		{
			name:     "empty input",
			input:    []uint16{},
			expected: []uint16{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterGREASEUint16(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("filterGREASEUint16() length = %d, want %d", len(result), len(tt.expected))
			}
			for i, v := range result {
				if i < len(tt.expected) && v != tt.expected[i] {
					t.Errorf("filterGREASEUint16()[%d] = 0x%04X, want 0x%04X", i, v, tt.expected[i])
				}
			}
		})
	}
}

// BenchmarkComputeJA3FromSpec 基准测试：JA3 计算性能
func BenchmarkComputeJA3FromSpec(b *testing.B) {
	spec := tls.ClientHelloSpec{
		CipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xC02C, 0xC02B},
		Extensions: []tls.TLSExtension{
			&tls.SNIExtension{},
			&tls.SupportedVersionsExtension{Versions: []uint16{0x0304, 0x0303}},
			&tls.SupportedCurvesExtension{Curves: []tls.CurveID{tls.X25519, tls.CurveP256}},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := ComputeJA3FromSpec(spec)
		if err != nil {
			b.Fatal(err)
		}
	}
}
