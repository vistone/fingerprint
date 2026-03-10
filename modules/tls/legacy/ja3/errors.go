package ja3

import (
	"errors"
	"fmt"

	errs "github.com/vistone/fingerprint/modules/errors"
)

// ============================================================================
// JA3 subpackage error definitions
// ============================================================================

var (
	// ErrInvalidClientHelloSpec indicates an invalid ClientHello spec
	ErrInvalidClientHelloSpec = fmt.Errorf("%w: invalid client hello spec", errs.ErrInvalidFingerprint)

	// ErrProfileNotFound indicates the JA3 fingerprint profile was not found
	ErrProfileNotFound = fmt.Errorf("%w: ja3 profile", errs.ErrProfileNotFound)

	// ErrEmptyProfile indicates the JA3 fingerprint is empty
	ErrEmptyProfile = fmt.Errorf("%w: empty ja3 profile", errs.ErrInvalidFingerprint)

	// ErrClientHelloIDNotImplemented indicates the ClientHello ID does not support Spec export
	ErrClientHelloIDNotImplemented = fmt.Errorf("%w: ja3 client hello id", errs.ErrClientHelloSpecNotImplemented)
)

// IsInvalidClientHelloSpec checks if the error is an invalid ClientHello spec
func IsInvalidClientHelloSpec(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidClientHelloSpec)
}

// IsJA3ProfileNotFound checks if the error is a JA3 fingerprint not found
func IsJA3ProfileNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrProfileNotFound)
}

// IsEmptyProfile checks if the error is an empty fingerprint
func IsEmptyProfile(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrEmptyProfile)
}

// IsClientHelloIDNotImplemented checks if the error is a ClientHello ID not implemented
func IsClientHelloIDNotImplemented(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrClientHelloIDNotImplemented)
}
