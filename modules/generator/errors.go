package generator

import (
	"errors"
	"fmt"

	errs "github.com/vistone/fingerprint/modules/errors"
)

// ============================================================================
// Generator subpackage error definitions
// ============================================================================

var (
	// ErrNoProfilesAvailable indicates no available fingerprint profiles
	ErrNoProfilesAvailable = fmt.Errorf("%w: for generators", errs.ErrNoProfilesAvailable)

	// ErrFailedToGenerateFingerprint indicates fingerprint generation failure
	ErrFailedToGenerateFingerprint = fmt.Errorf("%w: fingerprint generation failed", errs.ErrInvalidFingerprint)

	// ErrInvalidRandomSource indicates invalid random source
	ErrInvalidRandomSource = fmt.Errorf("%w: invalid random source", errs.ErrInvalidFingerprint)
)

// IsNoProfilesAvailable checks whether the error indicates no available fingerprints
func IsNoProfilesAvailable(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrNoProfilesAvailable)
}

// IsFailedToGenerateFingerprint checks whether the error indicates fingerprint generation failure
func IsFailedToGenerateFingerprint(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrFailedToGenerateFingerprint)
}
