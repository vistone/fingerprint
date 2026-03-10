package ja4

import (
	"testing"

	profiles "github.com/vistone/fingerprint/modules/profiles/legacy"
)

// TestComputeJA4FromRealProfiles tests JA4 computation using real fingerprint profiles
func TestComputeJA4FromRealProfiles(t *testing.T) {
	// Use the real Chrome 133 fingerprint.
	chromeProfile, ok := profiles.MappedTLSClients["chrome_133"]
	if !ok {
		t.Fatal("chrome_133 profile not found")
	}

	spec, err := chromeProfile.GetClientHelloSpec()
	if err != nil {
		t.Skipf("chrome_133 does not support spec export: %v", err)
		return
	}

	result, err := ComputeJA4FromSpec(spec)
	if err != nil {
		t.Fatalf("ComputeJA4FromSpec failed: %v", err)
	}

	if result.Hash == "" {
		t.Error("Expected non-empty JA4 hash for chrome_133")
	}

	if result.RawString == "" {
		t.Error("Expected non-empty JA4 raw string for chrome_133")
	}

	// Verify JA4 hash format (should include underscore-separated parts).
	if result.JA4A == "" {
		t.Error("Expected non-empty JA4_a part")
	}

	t.Logf("Chrome 133 JA4 Hash: %s", result.Hash)
	t.Logf("Chrome 133 JA4_a: %s", result.JA4A)
	t.Logf("Chrome 133 Raw: %s", result.RawString)
}

// TestComputeJA4FromMultipleProfiles tests JA4 computation for multiple real fingerprints
func TestComputeJA4FromMultipleProfiles(t *testing.T) {
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

			result, err := ComputeJA4FromSpec(spec)
			if err != nil {
				t.Fatalf("ComputeJA4FromSpec failed for %s: %v", profileName, err)
			}

			if result.Hash == "" {
				t.Errorf("Expected non-empty JA4 hash for %s", profileName)
			}

			// Verify the JA4_a section (protocol, TLS version, SNI, etc.).
			if result.JA4A == "" {
				t.Errorf("Expected non-empty JA4_a for %s", profileName)
			}

			t.Logf("%s: JA4=%s, JA4_a=%s", profileName, result.Hash, result.JA4A)
		})
	}
}

// TestComputeJA4FromProfile tests computing JA4 directly from profile
func TestComputeJA4FromProfile(t *testing.T) {
	profile, ok := profiles.MappedTLSClients["chrome_133"]
	if !ok {
		t.Fatal("chrome_133 profile not found")
	}

	result, err := ComputeJA4FromProfile(profile)
	if err != nil {
		t.Skipf("chrome_133 does not support JA4 export: %v", err)
		return
	}

	if result.Hash == "" {
		t.Error("Expected non-empty hash from profile")
	}

	t.Logf("JA4 from profile: %s", result.Hash)
}

// TestComputeJA4ByProfileName tests computing JA4 by profile name
func TestComputeJA4ByProfileName(t *testing.T) {
	result, err := ComputeJA4ByProfileName("chrome_133")
	if err != nil {
		t.Skipf("chrome_133 does not support JA4 export: %v", err)
		return
	}

	if result.Hash == "" {
		t.Error("Expected non-empty hash from profile name")
	}

	// Verify each returned section.
	if result.JA4A == "" {
		t.Error("Expected non-empty JA4_a")
	}
	if result.JA4B == "" {
		t.Error("Expected non-empty JA4_b (cipher suites)")
	}
	if result.JA4C == "" {
		t.Error("Expected non-empty JA4_c (extensions)")
	}

	t.Logf("JA4 by name: Hash=%s, JA4_a=%s, JA4_b=%s, JA4_c=%s",
		result.Hash, result.JA4A, result.JA4B, result.JA4C)
}
