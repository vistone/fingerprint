package ja4s

import (
	"errors"
	"fmt"

	errs "github.com/vistone/fingerprint/modules/errors"
)

// ============================================================================
// JA4S subpackage error definitions
// ============================================================================

var (
	// ErrInvalidServerHello indicates an invalid ServerHello message.
	ErrInvalidServerHello = fmt.Errorf("%w: invalid server hello spec", errs.ErrInvalidFingerprint)

	// ErrMissingServerHello indicates a missing ServerHello.
	ErrMissingServerHello = fmt.Errorf("%w: missing server hello", errs.ErrInvalidFingerprint)

	// ErrInvalidCipherSuite indicates an invalid cipher suite.
	ErrInvalidCipherSuite = fmt.Errorf("%w: invalid cipher suite", errs.ErrInvalidFingerprint)

	// ErrUnsupportedExtension indicates an unsupported extension.
	ErrUnsupportedExtension = fmt.Errorf("%w: unsupported extension", errs.ErrInvalidFingerprint)
)

// IsInvalidServerHello checks whether the error is invalid ServerHello
func IsInvalidServerHello(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidServerHello)
}

// IsMissingServerHello checks whether the error is missing ServerHello
func IsMissingServerHello(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrMissingServerHello)
}

// IsInvalidCipherSuite checks whether the error is invalid cipher suite
func IsInvalidCipherSuite(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidCipherSuite)
}
