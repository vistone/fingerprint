package network

import (
	"errors"
	"fmt"
	"strings"

	errs "github.com/vistone/fingerprint/modules/errors"
)

// ============================================================================
// Network subpackage error definitions
// ============================================================================

var (
	// ErrTLSHandshakeFailed indicates TLS handshake failure
	ErrTLSHandshakeFailed = fmt.Errorf("%w: tls handshake failed", errs.ErrInvalidFingerprint)

	// ErrConnectionFailed indicates connection failure
	ErrConnectionFailed = fmt.Errorf("%w: connection failed", errs.ErrInvalidFingerprint)

	// ErrInvalidNetworkConfig indicates invalid network configuration
	ErrInvalidNetworkConfig = fmt.Errorf("%w: invalid network config", errs.ErrInvalidFingerprint)
)

// IsUTLSPSKLimitationError checks whether the error is caused by uTLS PSK limitation
// In some cases, the uTLS library cannot export PSK and requires special handling
func IsUTLSPSKLimitationError(err error) bool {
	if err == nil {
		return false
	}
	// This is a specific error returned by the uTLS library and requires special handling.
	return strings.Contains(strings.ToLower(err.Error()), "empty psk detected")
}

// IsTLSHandshakeFailed checks whether the error is a TLS handshake failure
func IsTLSHandshakeFailed(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrTLSHandshakeFailed)
}

// IsConnectionFailed checks whether the error is a connection failure
func IsConnectionFailed(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrConnectionFailed)
}
