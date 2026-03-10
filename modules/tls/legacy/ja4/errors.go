package ja4

import (
	"errors"
	"fmt"

	errs "github.com/vistone/fingerprint/modules/errors"
)

// ============================================================================
// JA4 subpackage error definitions
// ============================================================================

var (
	// ErrInvalidClientHelloSpec indicates an invalid ClientHello spec
	ErrInvalidClientHelloSpec = fmt.Errorf("%w: invalid client hello spec", errs.ErrInvalidFingerprint)

	// ErrProfileNotFound indicates the JA4 fingerprint profile was not found
	ErrProfileNotFound = fmt.Errorf("%w: ja4 profile", errs.ErrProfileNotFound)

	// ErrEmptyProfile indicates the JA4 fingerprint is empty
	ErrEmptyProfile = fmt.Errorf("%w: empty ja4 profile", errs.ErrInvalidFingerprint)

	// ErrInvalidTLSVersion indicates an invalid TLS version
	ErrInvalidTLSVersion = fmt.Errorf("%w: invalid tls version", errs.ErrInvalidFingerprint)

	// ErrMissingRequiredField indicates a missing required field
	ErrMissingRequiredField = fmt.Errorf("%w: missing required field", errs.ErrInvalidFingerprint)
)

// IsInvalidClientHelloSpec checks if the error is an invalid ClientHello spec
func IsInvalidClientHelloSpec(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidClientHelloSpec)
}

// IsJA4ProfileNotFound checks if the error is a JA4 fingerprint not found
func IsJA4ProfileNotFound(err error) bool {
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

// IsInvalidTLSVersion checks if the error is an invalid TLS version
func IsInvalidTLSVersion(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidTLSVersion)
}
