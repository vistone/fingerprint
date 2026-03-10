package ja3

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
	ErrProfileNotFound = fmt.Errorf("%w: ja3 profile", errs.ErrProfileNotFound)

	// translated comment
	ErrEmptyProfile = fmt.Errorf("%w: empty ja3 profile", errs.ErrInvalidFingerprint)

	// translated comment
	ErrClientHelloIDNotImplemented = fmt.Errorf("%w: ja3 client hello id", errs.ErrClientHelloSpecNotImplemented)
)

// translated comment
func IsInvalidClientHelloSpec(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidClientHelloSpec)
}

// translated comment
func IsJA3ProfileNotFound(err error) bool {
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
func IsClientHelloIDNotImplemented(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrClientHelloIDNotImplemented)
}
