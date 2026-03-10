package profiles

import (
	"sync"
	"testing"

	"github.com/vistone/fingerprint/modules/errors"
)

// translated comment
func TestAllProfilesValid(t *testing.T) {
	if len(MappedTLSClients) == 0 {
		t.Fatal("MappedTLSClients is empty")
	}

	t.Logf("Testing %d profiles", len(MappedTLSClients))

	for name, profile := range MappedTLSClients {
		t.Run(name, func(t *testing.T) {
			// translated comment
			str := profile.GetClientHelloStr()
			if str == "" {
				t.Error("GetClientHelloStr() returned empty string")
			}

			// translated comment
			helloID := profile.GetClientHelloId()
			if helloID.Client == "" {
				t.Error("GetClientHelloId().Client is empty")
			}

			// translated comment
			settings := profile.GetSettings()
			if len(settings) == 0 {
				t.Error("GetSettings() returned empty map")
			}

			// translated comment
			settingsOrder := profile.GetSettingsOrder()
			if len(settingsOrder) == 0 {
				t.Error("GetSettingsOrder() returned empty slice")
			}

			// translated comment
			pseudoHeaders := profile.GetPseudoHeaderOrder()
			if len(pseudoHeaders) == 0 {
				t.Error("GetPseudoHeaderOrder() returned empty slice")
			}

			// translated comment
			flow := profile.GetConnectionFlow()
			if flow == 0 {
				t.Error("GetConnectionFlow() returned 0")
			}

			// translated comment
			spec, err := profile.GetClientHelloSpec()
			if err != nil {
				if !errors.IsClientHelloSpecNotImplemented(err) {
					t.Errorf("GetClientHelloSpec() unexpected error: %v", err)
				}
			} else {
				// translated comment
				if len(spec.CipherSuites) == 0 {
					t.Error("ClientHelloSpec.CipherSuites is empty")
				}
				if len(spec.Extensions) == 0 {
					t.Error("ClientHelloSpec.Extensions is empty")
				}
			}
		})
	}
}

// translated comment
func TestDefaultClientProfile(t *testing.T) {
	if DefaultClientProfile.GetClientHelloStr() == "" {
		t.Error("DefaultClientProfile is invalid")
	}
}

// translated comment
func TestChromeProfiles(t *testing.T) {
	chromeProfiles := []string{
		"chrome_103", "chrome_104", "chrome_105", "chrome_106", "chrome_107",
		"chrome_108", "chrome_109", "chrome_110", "chrome_111", "chrome_112",
		"chrome_117", "chrome_120", "chrome_124", "chrome_131", "chrome_133",
	}

	for _, name := range chromeProfiles {
		t.Run(name, func(t *testing.T) {
			profile, ok := MappedTLSClients[name]
			if !ok {
				t.Skipf("Profile %s not found", name)
				return
			}

			helloID := profile.GetClientHelloId()
			if helloID.Client != "Chrome" {
				t.Errorf("Expected Client='Chrome', got '%s'", helloID.Client)
			}
		})
	}
}

// translated comment
func TestFirefoxProfiles(t *testing.T) {
	firefoxProfiles := []string{
		"firefox_102", "firefox_104", "firefox_105", "firefox_106",
		"firefox_108", "firefox_110", "firefox_117", "firefox_120",
		"firefox_132", "firefox_133", "firefox_135",
	}

	for _, name := range firefoxProfiles {
		t.Run(name, func(t *testing.T) {
			profile, ok := MappedTLSClients[name]
			if !ok {
				t.Skipf("Profile %s not found", name)
				return
			}

			helloID := profile.GetClientHelloId()
			if helloID.Client != "Firefox" {
				t.Errorf("Expected Client='Firefox', got '%s'", helloID.Client)
			}
		})
	}
}

// translated comment
func TestSafariProfiles(t *testing.T) {
	safariProfiles := []string{
		"safari_15_6_1", "safari_16_0",
	}

	for _, name := range safariProfiles {
		t.Run(name, func(t *testing.T) {
			profile, ok := MappedTLSClients[name]
			if !ok {
				t.Skipf("Profile %s not found", name)
				return
			}

			helloID := profile.GetClientHelloId()
			if helloID.Client != "Safari" {
				t.Errorf("Expected Client='Safari', got '%s'", helloID.Client)
			}
		})
	}
}

// translated comment
func TestSafariIOSProfiles(t *testing.T) {
	safariIOSProfiles := []string{
		"safari_ios_15_5", "safari_ios_15_6", "safari_ios_16_0",
		"safari_ios_17_0", "safari_ios_18_0", "safari_ios_18_5",
	}

	for _, name := range safariIOSProfiles {
		t.Run(name, func(t *testing.T) {
			profile, ok := MappedTLSClients[name]
			if !ok {
				t.Skipf("Profile %s not found", name)
				return
			}

			helloID := profile.GetClientHelloId()
			if helloID.Client != "iOS" {
				t.Errorf("Expected Client='iOS', got '%s'", helloID.Client)
			}
		})
	}
}

// translated comment
func TestEdgeProfiles(t *testing.T) {
	edgeProfiles := []string{
		"edge_99", "edge_101", "edge_120", "edge_131", "edge_133",
	}

	for _, name := range edgeProfiles {
		t.Run(name, func(t *testing.T) {
			profile, ok := MappedTLSClients[name]
			if !ok {
				t.Skipf("Profile %s not found", name)
				return
			}

			helloID := profile.GetClientHelloId()
			if helloID.Client != "Edge" {
				t.Errorf("Expected Client='Edge', got '%s'", helloID.Client)
			}
		})
	}
}

// translated comment
func TestProfileConsistency(t *testing.T) {
	for name, profile := range MappedTLSClients {
		t.Run(name, func(t *testing.T) {
			// translated comment
			settings := profile.GetSettings()
			settingsOrder := profile.GetSettingsOrder()

			if len(settings) != len(settingsOrder) {
				t.Errorf("settings count (%d) != settingsOrder count (%d)",
					len(settings), len(settingsOrder))
			}

			// translated comment
			pseudoHeaders := profile.GetPseudoHeaderOrder()
			hasMethod := false
			hasPath := false
			for _, h := range pseudoHeaders {
				if h == ":method" {
					hasMethod = true
				}
				if h == ":path" {
					hasPath = true
				}
			}
			if !hasMethod {
				t.Error("pseudoHeaderOrder missing ':method'")
			}
			if !hasPath {
				t.Error("pseudoHeaderOrder missing ':path'")
			}
		})
	}
}

// translated comment
func BenchmarkGetClientHelloSpec(b *testing.B) {
	profile := DefaultClientProfile
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := profile.GetClientHelloSpec()
		if err != nil && !errors.IsClientHelloSpecNotImplemented(err) {
			b.Fatal(err)
		}
	}
}

// translated comment
func BenchmarkGetSettings(b *testing.B) {
	profile := DefaultClientProfile
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = profile.GetSettings()
	}
}

// translated comment
func BenchmarkGetPseudoHeaderOrder(b *testing.B) {
	profile := DefaultClientProfile
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = profile.GetPseudoHeaderOrder()
	}
}

// translated comment
func TestGetClientProfile_Concurrent(t *testing.T) {
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// translated comment
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// translated comment
			profileName := "chrome_133"
			if id%2 == 0 {
				profileName = "firefox_135"
			}
			
			profile, ok := GetClientProfile(profileName)
			if !ok {
				return
			}
			
			// translated comment
			_, err := profile.GetClientHelloSpec()
			if err != nil && err.Error() != "please implement this method" {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// translated comment
	errCount := 0
	for err := range errors {
		if err != nil {
			errCount++
			t.Errorf("Unexpected error: %v", err)
		}
	}

	if errCount > 0 {
		t.Fatalf("%d goroutines encountered errors", errCount)
	}
}

// translated comment
func TestGetAllProfiles_Concurrent(t *testing.T) {
	var wg sync.WaitGroup
	results := make(chan []string, 50)

	// translated comment
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			names := GetAllProfiles()
			results <- names
		}()
	}

	wg.Wait()
	close(results)

	// translated comment
	var expectedLen int
	first := true
	for names := range results {
		if first {
			expectedLen = len(names)
			first = false
		} else if len(names) != expectedLen {
			t.Errorf("Inconsistent profile count: got %d, want %d", len(names), expectedLen)
		}
	}
}

// translated comment
func TestHasProfile(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"chrome_133", true},
		{"firefox_135", true},
		{"nonexistent_profile", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasProfile(tt.name); got != tt.want {
				t.Errorf("HasProfile(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// translated comment
func TestGetClientProfile(t *testing.T) {
	tests := []struct {
		name    string
		wantOK  bool
		wantErr bool
	}{
		{"chrome_133", true, false},
		{"firefox_135", true, false},
		{"nonexistent", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, ok := GetClientProfile(tt.name)
			if ok != tt.wantOK {
				t.Errorf("GetClientProfile(%q) ok = %v, want %v", tt.name, ok, tt.wantOK)
			}
			
			if ok {
				// translated comment
				_, err := profile.GetClientHelloSpec()
				hasErr := (err != nil)
				if hasErr != tt.wantErr {
					t.Errorf("GetClientHelloSpec() err = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

// translated comment
func BenchmarkGetClientProfile(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = GetClientProfile("chrome_133")
	}
}

// translated comment
func BenchmarkGetClientProfile_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = GetClientProfile("chrome_133")
		}
	})
}
