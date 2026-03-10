// Package core provides input validation utilities
package core

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Validator input validator
type Validator struct {
	errors []error
}

// NewValidator creates new validator
func NewValidator() *Validator {
	return &Validator{
		errors: make([]error, 0),
	}
}

// HasErrors checks if there are errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors gets all errors
func (v *Validator) Errors() []error {
	return v.errors
}

// Error returns combined error
func (v *Validator) Error() error {
	if !v.HasErrors() {
		return nil
	}
	msgs := make([]string, len(v.errors))
	for i, err := range v.errors {
		msgs[i] = err.Error()
	}
	return NewCodedError(ErrCodeInvalidInput, "Validator", errors.New(strings.Join(msgs, "; ")))
}

// AddError adds error
func (v *Validator) AddError(err error) {
	if err != nil {
		v.errors = append(v.errors, err)
	}
}

// AddErrorf adds formatted error
func (v *Validator) AddErrorf(format string, args ...interface{}) {
	v.errors = append(v.errors, fmt.Errorf(format, args...))
}

// NotNil validates not nil
func (v *Validator) NotNil(val interface{}, name string) *Validator {
	if val == nil {
		v.AddErrorf("%s cannot be nil", name)
	}
	return v
}

// NotEmpty validates string is not empty
func (v *Validator) NotEmpty(s string, name string) *Validator {
	if strings.TrimSpace(s) == "" {
		v.AddErrorf("%s cannot be empty", name)
	}
	return v
}

// NotZero validates integer is not zero
func (v *Validator) NotZero(n int, name string) *Validator {
	if n == 0 {
		v.AddErrorf("%s cannot be zero", name)
	}
	return v
}

// Positive validates positive number
func (v *Validator) Positive(n int, name string) *Validator {
	if n <= 0 {
		v.AddErrorf("%s must be positive, got %d", name, n)
	}
	return v
}

// InRange verifyrange
func (v *Validator) InRange(n, min, max int, name string) *Validator {
	if n < min || n > max {
		v.AddErrorf("%s must be in range [%d, %d], got %d", name, min, max, n)
	}
	return v
}

// MinLength verifyminimumlength
func (v *Validator) MinLength(s string, min int, name string) *Validator {
	if len(s) < min {
		v.AddErrorf("%s must be at least %d characters, got %d", name, min, len(s))
	}
	return v
}

// MaxLength verifymaximumlength
func (v *Validator) MaxLength(s string, max int, name string) *Validator {
	if len(s) > max {
		v.AddErrorf("%s must be at most %d characters, got %d", name, max, len(s))
	}
	return v
}

// Matches validates regex match
func (v *Validator) Matches(s string, pattern *regexp.Regexp, name string) *Validator {
	if !pattern.MatchString(s) {
		v.AddErrorf("%s format is invalid", name)
	}
	return v
}

// ValidBrowserType validates browser type
func (v *Validator) ValidBrowserType(bt BrowserType, name string) *Validator {
	valid := []BrowserType{
		BrowserChrome, BrowserFirefox, BrowserSafari,
		BrowserEdge, BrowserOpera, BrowserBrave,
	}
	for _, b := range valid {
		if bt == b {
			return v
		}
	}
	v.AddErrorf("%s is not a valid browser type: %s", name, bt)
	return v
}

// ValidOS validates operating system
func (v *Validator) ValidOS(os OperatingSystem, name string) *Validator {
	valid := []OperatingSystem{
		OSWindows10, OSWindows11, OSMacOS13, OSMacOS14, OSMacOS15,
		OSLinux, OSLinuxUbuntu, OSLinuxDebian, OSLinuxFedora,
		OSiOS, OSiPadOS, OSAndroid,
	}
	for _, o := range valid {
		if os == o {
			return v
		}
	}
	v.AddErrorf("%s is not a valid OS: %s", name, os)
	return v
}

// ValidateTLSVersion validates TLS version
func ValidateTLSVersion(version uint16) error {
	switch version {
	case 0x0301, 0x0302, 0x0303, 0x0304:
		return nil
	default:
		return NewCodedErrorf(ErrCodeInvalidTLSVersion, "ValidateTLSVersion",
			"invalid TLS version: 0x%04x", version)
	}
}

// ValidateJA3Hash validates JA3 hash
func ValidateJA3Hash(hash string) error {
	if len(hash) != 32 {
		return NewCodedErrorf(ErrCodeInvalidJA3Hash, "ValidateJA3Hash",
			"invalid JA3 hash length: %d, expected 32", len(hash))
	}
	// JA3 is MD5, should be 32-character hexadecimal
	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return NewCodedError(ErrCodeInvalidJA3Hash, "ValidateJA3Hash",
				fmt.Errorf("invalid character in JA3 hash: %c", c))
		}
	}
	return nil
}

// SanitizeString sanitizes string input
func SanitizeString(s string, maxLen int) string {
	// remove control characters
	s = strings.Map(func(r rune) rune {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return -1
		}
		return r
	}, s)
	
	// truncate overly long strings
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	
	return strings.TrimSpace(s)
}

// SafeDereference safely dereferences pointer
func SafeDereference(ptr *int, defaultVal int) int {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// SafeSliceAccess safely accesses slice
func SafeSliceAccess(slice []interface{}, index int) (interface{}, bool) {
	if index < 0 || index >= len(slice) {
		return nil, false
	}
	return slice[index], true
}
