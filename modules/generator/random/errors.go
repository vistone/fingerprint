package random

import (
	"errors"
	"fmt"

	"github.com/vistone/fingerprint/modules/generator"
	errs "github.com/vistone/fingerprint/modules/errors"
)

// ============================================================================
// Random subpackage error definitions
// ============================================================================

var (
	// ErrBrowserTypeNotSupported indicates unsupported browser type
	ErrBrowserTypeNotSupported = fmt.Errorf("%w: random generator", errs.ErrUnsupportedBrowser)

	// ErrNoRandomProfileFound indicates no random fingerprint profile found
	ErrNoRandomProfileFound = fmt.Errorf("%w: random profiles", errs.ErrNoProfilesAvailable)

	// ErrRandomProfileInvalid indicates invalid random fingerprint profile
	ErrRandomProfileInvalid = fmt.Errorf("%w: random profile is invalid", generator.ErrFailedToGenerateFingerprint)
)

// IsBrowserTypeNotSupported checks whether the error indicates unsupported browser type
func IsBrowserTypeNotSupported(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrBrowserTypeNotSupported)
}

// IsNoRandomProfileFound checks whether the error indicates no random fingerprint profile
func IsNoRandomProfileFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrNoRandomProfileFound)
}

// IsRandomProfileInvalid checks whether the error indicates invalid random fingerprint profile
func IsRandomProfileInvalid(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrRandomProfileInvalid)
}
