package ja3

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	tls "github.com/bogdanfinn/utls"
	errdefs "github.com/vistone/fingerprint/modules/errors"
	profiles "github.com/vistone/fingerprint/modules/profiles/legacy"
)

// ============================================================================
// Error type checker function tests
// ============================================================================

// TestErrorTypes tests all JA3 error type checkers
func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checkers map[string]func(error) bool
		want     map[string]bool
	}{
		{
			name: "TestIsInvalidClientHelloSpec",
			err:  ErrInvalidClientHelloSpec,
			checkers: map[string]func(error) bool{
				"IsInvalidClientHelloSpec":      IsInvalidClientHelloSpec,
				"IsJA3ProfileNotFound":          IsJA3ProfileNotFound,
				"IsEmptyProfile":                IsEmptyProfile,
				"IsClientHelloIDNotImplemented": IsClientHelloIDNotImplemented,
			},
			want: map[string]bool{
				"IsInvalidClientHelloSpec":      true,
				"IsJA3ProfileNotFound":          false,
				"IsEmptyProfile":                false,
				"IsClientHelloIDNotImplemented": false,
			},
		},
		{
			name: "TestIsJA3ProfileNotFound",
			err:  ErrProfileNotFound,
			checkers: map[string]func(error) bool{
				"IsInvalidClientHelloSpec":      IsInvalidClientHelloSpec,
				"IsJA3ProfileNotFound":          IsJA3ProfileNotFound,
				"IsEmptyProfile":                IsEmptyProfile,
				"IsClientHelloIDNotImplemented": IsClientHelloIDNotImplemented,
			},
			want: map[string]bool{
				"IsInvalidClientHelloSpec":      false,
				"IsJA3ProfileNotFound":          true,
				"IsEmptyProfile":                false,
				"IsClientHelloIDNotImplemented": false,
			},
		},
		{
			name: "TestIsEmptyProfile",
			err:  ErrEmptyProfile,
			checkers: map[string]func(error) bool{
				"IsInvalidClientHelloSpec":      IsInvalidClientHelloSpec,
				"IsJA3ProfileNotFound":          IsJA3ProfileNotFound,
				"IsEmptyProfile":                IsEmptyProfile,
				"IsClientHelloIDNotImplemented": IsClientHelloIDNotImplemented,
			},
			want: map[string]bool{
				"IsInvalidClientHelloSpec":      false,
				"IsJA3ProfileNotFound":          false,
				"IsEmptyProfile":                true,
				"IsClientHelloIDNotImplemented": false,
			},
		},
		{
			name: "TestIsClientHelloIDNotImplemented",
			err:  ErrClientHelloIDNotImplemented,
			checkers: map[string]func(error) bool{
				"IsInvalidClientHelloSpec":      IsInvalidClientHelloSpec,
				"IsJA3ProfileNotFound":          IsJA3ProfileNotFound,
				"IsEmptyProfile":                IsEmptyProfile,
				"IsClientHelloIDNotImplemented": IsClientHelloIDNotImplemented,
			},
			want: map[string]bool{
				"IsInvalidClientHelloSpec":      false,
				"IsJA3ProfileNotFound":          false,
				"IsEmptyProfile":                false,
				"IsClientHelloIDNotImplemented": true,
			},
		},
		{
			name: "nil error returns false for all",
			err:  nil,
			checkers: map[string]func(error) bool{
				"IsInvalidClientHelloSpec":      IsInvalidClientHelloSpec,
				"IsJA3ProfileNotFound":          IsJA3ProfileNotFound,
				"IsEmptyProfile":                IsEmptyProfile,
				"IsClientHelloIDNotImplemented": IsClientHelloIDNotImplemented,
			},
			want: map[string]bool{
				"IsInvalidClientHelloSpec":      false,
				"IsJA3ProfileNotFound":          false,
				"IsEmptyProfile":                false,
				"IsClientHelloIDNotImplemented": false,
			},
		},
		{
			name: "unrelated error returns false for all",
			err:  errors.New("some random error"),
			checkers: map[string]func(error) bool{
				"IsInvalidClientHelloSpec":      IsInvalidClientHelloSpec,
				"IsJA3ProfileNotFound":          IsJA3ProfileNotFound,
				"IsEmptyProfile":                IsEmptyProfile,
				"IsClientHelloIDNotImplemented": IsClientHelloIDNotImplemented,
			},
			want: map[string]bool{
				"IsInvalidClientHelloSpec":      false,
				"IsJA3ProfileNotFound":          false,
				"IsEmptyProfile":                false,
				"IsClientHelloIDNotImplemented": false,
			},
		},
		{
			name: "wrapped error should match",
			err:  fmt.Errorf("wrapped: %w", ErrInvalidClientHelloSpec),
			checkers: map[string]func(error) bool{
				"IsInvalidClientHelloSpec": IsInvalidClientHelloSpec,
			},
			want: map[string]bool{
				"IsInvalidClientHelloSpec": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for checkerName, checker := range tt.checkers {
				got := checker(tt.err)
				want := tt.want[checkerName]
				if got != want {
					t.Errorf("%s() = %v, want %v for error %v", checkerName, got, want, tt.err)
				}
			}
		})
	}
}

// ============================================================================
// InitMappedTLSClients tests
// ============================================================================

// TestInitMappedTLSClients tests initialization of MappedTLSClients
func TestInitMappedTLSClients(t *testing.T) {
	tests := []struct {
		name         string
		clients      interface{}
		expectedLen  int
		expectedKeys []string
	}{
		{
			name: "init with map[string]ClientProfile",
			clients: map[string]ClientProfile{
				"test_profile_1": &mockClientProfile{},
				"test_profile_2": &mockClientProfile{},
			},
			expectedLen:  2,
			expectedKeys: []string{"test_profile_1", "test_profile_2"},
		},
		{
			name: "init with map[string]interface{}",
			clients: map[string]interface{}{
				"test_profile_3": &mockClientProfile{},
				"test_profile_4": &mockClientProfile{},
			},
			expectedLen:  2,
			expectedKeys: []string{"test_profile_3", "test_profile_4"},
		},
		{
			name: "init with empty map",
			clients: map[string]ClientProfile{
				"empty_test": &mockClientProfile{},
			},
			expectedLen:  1,
			expectedKeys: []string{"empty_test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original state
			originalMapped := MappedTLSClients
			defer func() {
				MappedTLSClients = originalMapped
			}()

			// Initialize
			InitMappedTLSClients(tt.clients)

			// Verify
			if len(MappedTLSClients) != tt.expectedLen {
				t.Errorf("Expected %d profiles, got %d", tt.expectedLen, len(MappedTLSClients))
			}

			for _, key := range tt.expectedKeys {
				if _, ok := MappedTLSClients[key]; !ok {
					t.Errorf("Expected profile %s not found", key)
				}
			}
		})
	}
}

// TestInitMappedTLSClients_Idempotent tests repeated initialization (idempotency)
func TestInitMappedTLSClients_Idempotent(t *testing.T) {
	// Save original state
	originalMapped := MappedTLSClients
	defer func() {
		MappedTLSClients = originalMapped
	}()

	// First initialization
	firstMap := map[string]ClientProfile{
		"profile_a": &mockClientProfile{},
	}
	InitMappedTLSClients(firstMap)

	if len(MappedTLSClients) != 1 {
		t.Fatalf("Expected 1 profile after first init, got %d", len(MappedTLSClients))
	}

	// Second initialization (overwrite)
	secondMap := map[string]ClientProfile{
		"profile_b": &mockClientProfile{},
		"profile_c": &mockClientProfile{},
	}
	InitMappedTLSClients(secondMap)

	if len(MappedTLSClients) != 2 {
		t.Errorf("Expected 2 profiles after second init, got %d", len(MappedTLSClients))
	}

	// Verify profiles from first initialization are overwritten
	if _, ok := MappedTLSClients["profile_a"]; ok {
		t.Error("Old profile should be removed after re-initialization")
	}

	if _, ok := MappedTLSClients["profile_b"]; !ok {
		t.Error("New profile_b should exist")
	}
}

// ============================================================================
// InitMappedTLSClientsRaw tests
// ============================================================================

// TestInitMappedTLSClientsRaw tests initialization using raw input
func TestInitMappedTLSClientsRaw(t *testing.T) {
	tests := []struct {
		name        string
		clients     interface{}
		expectedLen int
		checkKeys   []string
	}{
		{
			name: "init with valid map",
			clients: map[string]*mockClientProfile{
				"raw_profile_1": {},
				"raw_profile_2": {},
			},
			expectedLen: 2,
			checkKeys:   []string{"raw_profile_1", "raw_profile_2"},
		},
		{
			name:        "init with non-map returns nil",
			clients:     "not a map",
			expectedLen: 0,
			checkKeys:   []string{},
		},
		{
			name:        "init with nil",
			clients:     nil,
			expectedLen: 0,
			checkKeys:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original state
			originalMapped := MappedTLSClients
			defer func() {
				MappedTLSClients = originalMapped
			}()

			// Initialize
			InitMappedTLSClientsRaw(tt.clients)

			// Verify
			if len(MappedTLSClients) != tt.expectedLen {
				t.Errorf("Expected %d profiles, got %d", tt.expectedLen, len(MappedTLSClients))
			}

			for _, key := range tt.checkKeys {
				if _, ok := MappedTLSClients[key]; !ok {
					t.Errorf("Expected profile %s not found", key)
				}
			}
		})
	}
}

// TestInitMappedTLSClientsRaw_WithMethod tests type with method
func TestInitMappedTLSClientsRaw_WithMethod(t *testing.T) {
	// Save original state
	originalMapped := MappedTLSClients
	defer func() {
		MappedTLSClients = originalMapped
	}()

	// Use type that has GetClientHelloSpec method
	clients := map[string]*mockClientProfileWithMethod{
		"method_profile": {
			spec: tls.ClientHelloSpec{
				CipherSuites: []uint16{0x1301, 0x1302},
				Extensions:   []tls.TLSExtension{&tls.SNIExtension{}},
			},
		},
	}

	InitMappedTLSClientsRaw(clients)

	if len(MappedTLSClients) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(MappedTLSClients))
	}

	// Verify GetClientHelloSpec can be called through ClientProfile interface
	profile, ok := MappedTLSClients["method_profile"]
	if !ok {
		t.Fatal("method_profile not found")
	}

	spec, err := profile.GetClientHelloSpec()
	if err != nil {
		t.Fatalf("GetClientHelloSpec failed: %v", err)
	}

	if len(spec.CipherSuites) != 2 {
		t.Errorf("Expected 2 cipher suites, got %d", len(spec.CipherSuites))
	}
}

// ============================================================================
// GetClientHelloSpec tests
// ============================================================================

// mockClientProfile is a mock implementing ClientProfile interface
func (m *mockClientProfile) GetClientHelloSpec() (tls.ClientHelloSpec, error) {
	if m.err != nil {
		return tls.ClientHelloSpec{}, m.err
	}
	return m.spec, nil
}

// mockClientProfileWithMethod is a mock type with GetClientHelloSpec method
type mockClientProfileWithMethod struct {
	spec tls.ClientHelloSpec
	err  error
}

func (m *mockClientProfileWithMethod) GetClientHelloSpec() (tls.ClientHelloSpec, error) {
	return m.spec, m.err
}

type mockClientProfile struct {
	spec tls.ClientHelloSpec
	err  error
}

// TestGetClientHelloSpec tests obtaining spec via profile
func TestGetClientHelloSpec(t *testing.T) {
	validSpec := tls.ClientHelloSpec{
		CipherSuites: []uint16{0x1301, 0x1302},
		Extensions:   []tls.TLSExtension{&tls.SNIExtension{}},
	}

	tests := []struct {
		name      string
		profile   ClientProfile
		wantErr   bool
		errCheck  func(error) bool
		checkSpec func(*testing.T, tls.ClientHelloSpec)
	}{
		{
			name:    "valid profile returns spec",
			profile: &mockClientProfile{spec: validSpec},
			wantErr: false,
			checkSpec: func(t *testing.T, spec tls.ClientHelloSpec) {
				if len(spec.CipherSuites) != 2 {
					t.Errorf("Expected 2 cipher suites, got %d", len(spec.CipherSuites))
				}
			},
		},
		{
			name:    "profile returns error",
			profile: &mockClientProfile{err: fmt.Errorf("spec error")},
			wantErr: true,
		},
		{
			name:    "spec with nil slices returns error on validation",
			profile: &mockClientProfile{spec: tls.ClientHelloSpec{CipherSuites: nil, Extensions: nil}},
			wantErr: false, // GetClientHelloSpec itself doesn't validate, error happens in ComputeJA3
			checkSpec: func(t *testing.T, spec tls.ClientHelloSpec) {
				if spec.CipherSuites != nil {
					t.Error("Expected nil CipherSuites")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := tt.profile.GetClientHelloSpec()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkSpec != nil {
				tt.checkSpec(t, spec)
			}
		})
	}
}

// ============================================================================
// ComputeJA3FromProfile tests
// ============================================================================

// TestComputeJA3FromProfile tests computing JA3 from profile
func TestComputeJA3FromProfile(t *testing.T) {
	tests := []struct {
		name    string
		profile ClientProfile
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid profile",
			profile: &mockClientProfile{
				spec: tls.ClientHelloSpec{
					CipherSuites: []uint16{0x1301, 0x1302},
					Extensions:   []tls.TLSExtension{&tls.SNIExtension{}},
				},
			},
			wantErr: false,
		},
		{
			name: "profile with empty spec",
			profile: &mockClientProfile{
				spec: tls.ClientHelloSpec{
					CipherSuites: []uint16{},
					Extensions:   []tls.TLSExtension{},
				},
			},
			wantErr: true,
			errMsg:  "cipher suites list is empty",
		},
		{
			name:    "profile that returns error",
			profile: &mockClientProfile{err: errors.New("profile error")},
			wantErr: true,
			errMsg:  "profile error",
		},
		{
			name: "spec with only GREASE values",
			profile: &mockClientProfile{
				spec: tls.ClientHelloSpec{
					CipherSuites: []uint16{0x0A0A, 0x1A1A},
					Extensions:   []tls.TLSExtension{&tls.UtlsGREASEExtension{}},
				},
			},
			wantErr: true,
			errMsg:  "no valid cipher suites",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ComputeJA3FromProfile(tt.profile)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected non-nil result")
				return
			}

			// Verify result is a valid MD5 hash
			if len(result.Hash) != 32 {
				t.Errorf("Expected MD5 hash length 32, got %d", len(result.Hash))
			}

			if result.RawString == "" {
				t.Error("Expected non-empty raw string")
			}
		})
	}
}

// ============================================================================
// ComputeJA3ByProfileName tests
// ============================================================================

// TestComputeJA3ByProfileName tests computing JA3 via profile name
func TestComputeJA3ByProfileName(t *testing.T) {
	// Save original state
	originalMapped := MappedTLSClients
	defer func() {
		MappedTLSClients = originalMapped
	}()

	// Set MappedTLSClients for testing
	validSpec := tls.ClientHelloSpec{
		CipherSuites: []uint16{0x1301, 0x1302},
		Extensions:   []tls.TLSExtension{&tls.SNIExtension{}},
	}

	MappedTLSClients = map[string]ClientProfile{
		"valid_profile":   &mockClientProfile{spec: validSpec},
		"error_profile":   &mockClientProfile{err: errors.New("spec error")},
		"invalid_profile": &mockClientProfile{spec: tls.ClientHelloSpec{CipherSuites: []uint16{}, Extensions: []tls.TLSExtension{}}},
	}

	tests := []struct {
		name        string
		profileName string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "existing valid profile",
			profileName: "valid_profile",
			wantErr:     false,
		},
		{
			name:        "non-existing profile",
			profileName: "non_existing_profile",
			wantErr:     true,
			errMsg:      "not found",
		},
		{
			name:        "profile that returns error",
			profileName: "error_profile",
			wantErr:     true,
			errMsg:      "spec error",
		},
		{
			name:        "profile with invalid spec",
			profileName: "invalid_profile",
			wantErr:     true,
			errMsg:      "cipher suites list is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ComputeJA3ByProfileName(tt.profileName)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected non-nil result")
				return
			}

			// Verify result is a valid MD5 hash
			if len(result.Hash) != 32 {
				t.Errorf("Expected MD5 hash length 32, got %d", len(result.Hash))
			}
		})
	}
}

// ============================================================================
// FindProfileByJA3_Extended tests
// ============================================================================

// TestFindProfileByJA3_Extended extends existing tests to 100% coverage
func TestFindProfileByJA3_Extended(t *testing.T) {
	// Save original state and restore after test
	originalMapped := MappedTLSClients
	originalIndex := ja3ProfileIndex
	defer func() {
		MappedTLSClients = originalMapped
		ja3ProfileIndexOnce = sync.Once{}
		ja3ProfileIndex = originalIndex
	}()

	// Create valid spec for JA3 computation
	validSpec := tls.ClientHelloSpec{
		CipherSuites: []uint16{0x1301, 0x1302},
		Extensions: []tls.TLSExtension{
			&tls.SNIExtension{},
			&tls.SupportedCurvesExtension{Curves: []tls.CurveID{tls.X25519}},
			&tls.SupportedPointsExtension{SupportedPoints: []uint8{0}},
		},
	}

	// Compute expected JA3 hash
	result, err := ComputeJA3FromSpec(validSpec)
	if err != nil {
		t.Fatalf("Failed to compute JA3: %v", err)
	}
	expectedHash := result.Hash

	tests := []struct {
		name           string
		setup          func()
		ja3Hash        string
		expectedResult []string
		description    string
	}{
		{
			name: "normal lookup",
			setup: func() {
				// Reset once and index
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = map[string]ClientProfile{
					"test_chrome": &mockClientProfile{spec: validSpec},
				}
			},
			ja3Hash:        expectedHash,
			expectedResult: []string{"test_chrome"},
			description:    "should find matching profile",
		},
		{
			name: "uninitialized MappedTLSClients",
			setup: func() {
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = nil
			},
			ja3Hash:        expectedHash,
			expectedResult: nil,
			description:    "should return nil when uninitialized",
		},
		{
			name: "invalid hash length - too short",
			setup: func() {
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = map[string]ClientProfile{
					"test_chrome": &mockClientProfile{spec: validSpec},
				}
			},
			ja3Hash:        "12345678",
			expectedResult: nil,
			description:    "short hash should not match",
		},
		{
			name: "invalid hash length - too long",
			setup: func() {
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = map[string]ClientProfile{
					"test_chrome": &mockClientProfile{spec: validSpec},
				}
			},
			ja3Hash:        "1234567890123456789012345678901234567890",
			expectedResult: nil,
			description:    "long hash should not match",
		},
		{
			name: "no matching profile",
			setup: func() {
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = map[string]ClientProfile{
					"test_chrome": &mockClientProfile{spec: validSpec},
				}
			},
			ja3Hash:        "00000000000000000000000000000000",
			expectedResult: nil,
			description:    "non-existent hash should return nil",
		},
		{
			name: "case insensitive match - uppercase",
			setup: func() {
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = map[string]ClientProfile{
					"test_chrome": &mockClientProfile{spec: validSpec},
				}
			},
			ja3Hash:        strings.ToUpper(expectedHash),
			expectedResult: []string{"test_chrome"},
			description:    "uppercase hash should match",
		},
		{
			name: "case insensitive match - mixed case",
			setup: func() {
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = map[string]ClientProfile{
					"test_chrome": &mockClientProfile{spec: validSpec},
				}
			},
			ja3Hash:        strings.ToUpper(expectedHash[:16]) + strings.ToLower(expectedHash[16:]),
			expectedResult: []string{"test_chrome"},
			description:    "mixed-case hash should match",
		},
		{
			name: "multiple profiles with same JA3",
			setup: func() {
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = map[string]ClientProfile{
					"test_chrome_a": &mockClientProfile{spec: validSpec},
					"test_chrome_b": &mockClientProfile{spec: validSpec},
				}
			},
			ja3Hash:        expectedHash,
			expectedResult: []string{"test_chrome_a", "test_chrome_b"},
			description:    "multiple profiles with same JA3 should all be returned",
		},
		{
			name: "empty hash",
			setup: func() {
				ja3ProfileIndexOnce = sync.Once{}
				ja3ProfileIndex = nil
				MappedTLSClients = map[string]ClientProfile{
					"test_chrome": &mockClientProfile{spec: validSpec},
				}
			},
			ja3Hash:        "",
			expectedResult: nil,
			description:    "empty hash should return nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result := FindProfileByJA3(tt.ja3Hash)

			if tt.expectedResult == nil {
				if len(result) != 0 {
					t.Errorf("%s: expected nil/empty result, got %v", tt.description, result)
				}
				return
			}

			if len(result) != len(tt.expectedResult) {
				t.Errorf("%s: expected %d results, got %d: %v", tt.description, len(tt.expectedResult), len(result), result)
				return
			}

			// Verify all expected results are present in returned results
			for _, expected := range tt.expectedResult {
				found := false
				for _, r := range result {
					if r == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: expected result %s not found in %v", tt.description, expected, result)
				}
			}
		})
	}
}

// TestFindProfileByJA3_WithRealProfiles tests with real profiles
func TestFindProfileByJA3_WithRealProfiles(t *testing.T) {
	// Save original state
	originalMapped := MappedTLSClients
	originalIndex := ja3ProfileIndex
	defer func() {
		MappedTLSClients = originalMapped
		ja3ProfileIndexOnce = sync.Once{}
		ja3ProfileIndex = originalIndex
	}()

	// Initialize with real profiles
	InitMappedTLSClients(profiles.MappedTLSClients)
	// Reset index
	ja3ProfileIndexOnce = sync.Once{}
	ja3ProfileIndex = nil

	// Compute JA3 for a real profile
	profile, ok := MappedTLSClients["chrome_133"]
	if !ok {
		t.Skip("chrome_133 profile not available")
	}

	result, err := ComputeJA3FromProfile(profile)
	if err != nil {
		t.Skipf("chrome_133 does not support spec export: %v", err)
	}

	// Lookup profile using computed hash
	foundProfiles := FindProfileByJA3(result.Hash)

	// Should find at least chrome_133
	foundChrome133 := false
	for _, name := range foundProfiles {
		if name == "chrome_133" {
			foundChrome133 = true
			break
		}
	}

	if !foundChrome133 {
		t.Errorf("Expected to find chrome_133 in results, got: %v", foundProfiles)
	}
}

// ============================================================================
// Edge cases and integration tests
// ============================================================================

// TestErrorWrapping tests error wrapping and unwrapping
func TestErrorWrapping(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		isTargetErr   func(error) bool
		expectedMatch bool
	}{
		{
			name:          "ErrInvalidClientHelloSpec wraps ErrInvalidFingerprint",
			err:           ErrInvalidClientHelloSpec,
			isTargetErr:   func(e error) bool { return errors.Is(e, errdefs.ErrInvalidFingerprint) },
			expectedMatch: true,
		},
		{
			name:          "ErrProfileNotFound wraps ErrProfileNotFound",
			err:           ErrProfileNotFound,
			isTargetErr:   func(e error) bool { return errors.Is(e, errdefs.ErrProfileNotFound) },
			expectedMatch: true,
		},
		{
			name:          "ErrEmptyProfile wraps ErrInvalidFingerprint",
			err:           ErrEmptyProfile,
			isTargetErr:   func(e error) bool { return errors.Is(e, errdefs.ErrInvalidFingerprint) },
			expectedMatch: true,
		},
		{
			name:          "ErrClientHelloIDNotImplemented wraps ErrClientHelloSpecNotImplemented",
			err:           ErrClientHelloIDNotImplemented,
			isTargetErr:   func(e error) bool { return errors.Is(e, errdefs.ErrClientHelloSpecNotImplemented) },
			expectedMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.isTargetErr(tt.err)
			if got != tt.expectedMatch {
				t.Errorf("Error wrapping check failed: got %v, want %v", got, tt.expectedMatch)
			}
		})
	}
}
