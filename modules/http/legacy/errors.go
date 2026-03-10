package http

import (
	"errors"
	"fmt"

	errs "github.com/vistone/fingerprint/modules/errors"
)

// ============================================================================
// HTTP subpackage error definitions
// ============================================================================

var (
	// ErrInvalidUserAgent indicates an invalid User-Agent
	ErrInvalidUserAgent = fmt.Errorf("%w: invalid user agent", errs.ErrInvalidUserAgent)

	// ErrHeaderBuildingFailed indicates HTTP header building failure
	ErrHeaderBuildingFailed = fmt.Errorf("%w: failed to build header", errs.ErrInvalidFingerprint)

	// ErrMissingRequiredHeader indicates missing required HTTP header
	ErrMissingRequiredHeader = fmt.Errorf("%w: missing required header", errs.ErrInvalidFingerprint)

	// ErrInvalidHeaderValue indicates invalid HTTP header value
	ErrInvalidHeaderValue = fmt.Errorf("%w: invalid header value", errs.ErrInvalidFingerprint)
)

// IsInvalidUserAgent checks if the error is for invalid User-Agent
func IsInvalidUserAgent(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidUserAgent)
}

// IsHeaderBuildingFailed checks if the error is for header building failure
func IsHeaderBuildingFailed(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrHeaderBuildingFailed)
}

// IsMissingRequiredHeader checks if the error is for missing header
func IsMissingRequiredHeader(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrMissingRequiredHeader)
}
