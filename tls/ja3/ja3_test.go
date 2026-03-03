package ja3

import (
	"testing"

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
