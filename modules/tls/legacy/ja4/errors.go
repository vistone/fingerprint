package ja4

import (
	"errors"
	"fmt"

	errs "github.com/vistone/fingerprint/modules/errors"
)

// ============================================================================
// translated comment
// ============================================================================

var (
	// translated comment
	ErrInvalidClientHelloSpec = fmt.Errorf("%w: invalid client hello spec", errs.ErrInvalidFingerprint)

	// translated comment
	ErrProfileNotFound = fmt.Errorf("%w: ja4 profile", errs.ErrProfileNotFound)

	// translated comment
	ErrEmptyProfile = fmt.Errorf("%w: empty ja4 profile", errs.ErrInvalidFingerprint)

	// translated comment
	ErrInvalidTLSVersion = fmt.Errorf("%w: invalid tls version", errs.ErrInvalidFingerprint)

	// translated comment
	ErrMissingRequiredField = fmt.Errorf("%w: missing required field", errs.ErrInvalidFingerprint)
)

// translated comment
func IsInvalidClientHelloSpec(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidClientHelloSpec)
}

// translated comment
func IsJA4ProfileNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrProfileNotFound)
}

// translated comment
func IsEmptyProfile(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrEmptyProfile)
}

// translated comment
func IsInvalidTLSVersion(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidTLSVersion)
}
