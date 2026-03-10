package profiles

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
	ErrProfileNotFound = fmt.Errorf("%w: profiles", errs.ErrProfileNotFound)

	// translated comment
	ErrInvalidProfile = fmt.Errorf("%w: profiles configuration", errs.ErrInvalidFingerprint)

	// translated comment
	ErrProfileInitializationFailed = fmt.Errorf("%w: profile initialization failed", errs.ErrInvalidFingerprint)

	// translated comment
	ErrClientHelloIDNotSupported = fmt.Errorf("%w: profiles client hello id", errs.ErrClientHelloSpecNotImplemented)
)

// translated comment
func IsProfileNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrProfileNotFound)
}

// translated comment
func IsInvalidProfile(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidProfile)
}

// translated comment
func IsProfileInitializationFailed(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrProfileInitializationFailed)
}

// translated comment
func IsClientHelloIDNotSupported(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrClientHelloIDNotSupported)
}
