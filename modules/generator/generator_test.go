package generator

import (
	"errors"
	"testing"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// TestErrorDefinitions tests error definitions
func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrNoProfilesAvailable",
			err:  ErrNoProfilesAvailable,
			want: "no profiles available: for generators",
		},
		{
			name: "ErrFailedToGenerateFingerprint",
			err:  ErrFailedToGenerateFingerprint,
			want: "invalid fingerprint format: fingerprint generation failed",
		},
		{
			name: "ErrInvalidRandomSource",
			err:  ErrInvalidRandomSource,
			want: "invalid fingerprint format: invalid random source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("error should not be nil")
			}
			if tt.err.Error() != tt.want {
				t.Errorf("error message = %v, want %v", tt.err.Error(), tt.want)
			}
		})
	}
}

// TestIsNoProfilesAvailable tests no-available-fingerprint error check
func TestIsNoProfilesAvailable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "matching error",
			err:  ErrNoProfilesAvailable,
			want: true,
		},
		{
			name: "wrapped matching error",
			err:  errors.Join(errors.New("context"), ErrNoProfilesAvailable),
			want: true,
		},
		{
			name: "different error",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "ErrFailedToGenerateFingerprint",
			err:  ErrFailedToGenerateFingerprint,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNoProfilesAvailable(tt.err)
			if got != tt.want {
				t.Errorf("IsNoProfilesAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsFailedToGenerateFingerprint tests fingerprint-generation-failure error check
func TestIsFailedToGenerateFingerprint(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "matching error",
			err:  ErrFailedToGenerateFingerprint,
			want: true,
		},
		{
			name: "wrapped matching error",
			err:  errors.Join(errors.New("context"), ErrFailedToGenerateFingerprint),
			want: true,
		},
		{
			name: "different error",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "ErrNoProfilesAvailable",
			err:  ErrNoProfilesAvailable,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFailedToGenerateFingerprint(tt.err)
			if got != tt.want {
				t.Errorf("IsFailedToGenerateFingerprint() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestErrorWrapping tests error wrapping
func TestErrorWrapping(t *testing.T) {
	// Test whether errors can be recognized by errors.Is.
	if !errors.Is(ErrNoProfilesAvailable, ErrNoProfilesAvailable) {
		t.Error("ErrNoProfilesAvailable should be self-identifiable")
	}

	if !errors.Is(ErrFailedToGenerateFingerprint, ErrFailedToGenerateFingerprint) {
		t.Error("ErrFailedToGenerateFingerprint should be self-identifiable")
	}

	// Test that one error does not match other errors.
	if errors.Is(ErrNoProfilesAvailable, ErrFailedToGenerateFingerprint) {
		t.Error("ErrNoProfilesAvailable should not match ErrFailedToGenerateFingerprint")
	}

	if errors.Is(ErrFailedToGenerateFingerprint, ErrNoProfilesAvailable) {
		t.Error("ErrFailedToGenerateFingerprint should not match ErrNoProfilesAvailable")
	}
}

// TestPackageExports tests package exports
func TestPackageExports(t *testing.T) {
	// Ensure all exported variables are not nil.
	if ErrNoProfilesAvailable == nil {
		t.Error("ErrNoProfilesAvailable should not be nil")
	}

	if ErrFailedToGenerateFingerprint == nil {
		t.Error("ErrFailedToGenerateFingerprint should not be nil")
	}

	if ErrInvalidRandomSource == nil {
		t.Error("ErrInvalidRandomSource should not be nil")
	}
}

func TestMutateProfile_DoesNotShareMutableNestedState(t *testing.T) {
	gen := NewSmartGenerator(nil, &SmartGeneratorConfig{CacheSize: 1})
	base := &profiles.ClientProfile{
		ID:          "base",
		BrowserType: core.BrowserChrome,
		Headers: &core.HTTPHeaders{
			UserAgent: "base-ua",
			Custom:    map[string]string{"X-Test": "base"},
		},
		Metadata: map[string]interface{}{"key": "base"},
		HTTP3Settings: &core.HTTP3Settings{
			QUICVersion: 1,
		},
		QUICVersions: []uint32{1},
	}

	mutated := gen.mutateProfile(base, 0)
	mutated.Headers.Custom["X-Test"] = "mutated"
	mutated.Metadata["key"] = "mutated"
	mutated.HTTP3Settings.QUICVersion = 2
	mutated.QUICVersions[0] = 2

	if got := base.Headers.Custom["X-Test"]; got != "base" {
		t.Fatalf("expected base headers to remain unchanged, got %q", got)
	}
	if got := base.Metadata["key"]; got != "base" {
		t.Fatalf("expected base metadata to remain unchanged, got %v", got)
	}
	if got := base.HTTP3Settings.QUICVersion; got != 1 {
		t.Fatalf("expected base HTTP3 settings to remain unchanged, got %d", got)
	}
	if got := base.QUICVersions[0]; got != 1 {
		t.Fatalf("expected base QUIC versions to remain unchanged, got %d", got)
	}
}

func TestShouldRetryDuplicateSource(t *testing.T) {
	usedSources := map[string]bool{"profile-1": true}

	if retry := shouldRetryDuplicateSource(usedSources, "profile-1", 4, 10, 1); retry {
		t.Fatal("expected duplicate source to be accepted once unique sources are exhausted")
	}

	if retry := shouldRetryDuplicateSource(usedSources, "profile-1", 1, 10, 3); !retry {
		t.Fatal("expected duplicate source to retry while more unique sources may still exist")
	}

	if retry := shouldRetryDuplicateSource(usedSources, "profile-2", 1, 10, 3); retry {
		t.Fatal("expected unseen source to be accepted immediately")
	}
}
