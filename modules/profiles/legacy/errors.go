package profiles

import (
	"errors"
	"fmt"

	errs "github.com/vistone/fingerprint/modules/errors"
)

// ============================================================================
// Profiles subpackage error definitions
// ============================================================================

var (
	// ErrProfileNotFound indicates missing fingerprint profile
	ErrProfileNotFound = fmt.Errorf("%w: profiles", errs.ErrProfileNotFound)

	// ErrInvalidProfile indicates invalid fingerprint profile
	ErrInvalidProfile = fmt.Errorf("%w: profiles configuration", errs.ErrInvalidFingerprint)

	// ErrProfileInitializationFailed indicates profile initialization failure
	ErrProfileInitializationFailed = fmt.Errorf("%w: profile initialization failed", errs.ErrInvalidFingerprint)

	// ErrClientHelloIDNotSupported indicates unsupported ClientHello ID
	ErrClientHelloIDNotSupported = fmt.Errorf("%w: profiles client hello id", errs.ErrClientHelloSpecNotImplemented)
)

// IsProfileNotFound checks whether error is profile-not-found
func IsProfileNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrProfileNotFound)
}

// IsInvalidProfile checks whether error is invalid profile
func IsInvalidProfile(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidProfile)
}

// IsProfileInitializationFailed checks whether error is initialization failure
func IsProfileInitializationFailed(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrProfileInitializationFailed)
}

// IsClientHelloIDNotSupported checks whether error is unsupported ClientHello ID
func IsClientHelloIDNotSupported(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrClientHelloIDNotSupported)
}
