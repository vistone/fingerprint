// Package core provides error definitions
package core

import (
	"errors"
	"fmt"
)

// ============================================================================
// error code system
// ============================================================================

// ErrorCode error code type
type ErrorCode string

// error code category prefixes:
// VAL - verifyerror (Validation)
// NTF - Not Found errors
// SYS - System errors
// NET - Network errors
// SEC - Security errors
// CFG - configurationerror (Configuration)
const (
	// verifyerror (VAL001-VAL099)
	ErrCodeInvalidInput     ErrorCode = "VAL001"
	ErrCodeRequiredField    ErrorCode = "VAL002"
	ErrCodeInvalidFormat    ErrorCode = "VAL003"
	ErrCodeOutOfRange       ErrorCode = "VAL004"
	ErrCodeNilPointer       ErrorCode = "VAL005"
	ErrCodeInvalidType      ErrorCode = "VAL006"
	ErrCodeEmptyValue       ErrorCode = "VAL007"
	ErrCodeInvalidLength    ErrorCode = "VAL008"
	ErrCodeInvalidCharset   ErrorCode = "VAL009"
	ErrCodeInvalidPattern   ErrorCode = "VAL010"

	// Not Found errors (NTF001-NTF099)
	ErrCodeProfileNotFound  ErrorCode = "NTF001"
	ErrCodeResourceNotFound ErrorCode = "NTF002"
	ErrCodeKeyNotFound      ErrorCode = "NTF003"
	ErrCodeFileNotFound     ErrorCode = "NTF004"
	ErrCodeRouteNotFound    ErrorCode = "NTF005"

	// System errors (SYS001-SYS099)
	ErrCodeInternal         ErrorCode = "SYS001"
	ErrCodeNotImplemented   ErrorCode = "SYS002"
	ErrCodeIOError          ErrorCode = "SYS003"
	ErrCodeMemoryError      ErrorCode = "SYS004"
	ErrCodeTimeout          ErrorCode = "SYS005"
	ErrCodeUnavailable      ErrorCode = "SYS006"
	ErrCodeAlreadyExists    ErrorCode = "SYS007"
	ErrCodeConcurrencyError ErrorCode = "SYS008"
	ErrCodeStateError       ErrorCode = "SYS009"

	// Network errors (NET001-NET099)
	ErrCodeNetworkError     ErrorCode = "NET001"
	ErrCodeConnectionFailed ErrorCode = "NET002"
	ErrCodeDNSFailed        ErrorCode = "NET003"
	ErrCodeTLSFailed        ErrorCode = "NET004"
	ErrCodeHTTPError        ErrorCode = "NET005"
	ErrCodeProxyError       ErrorCode = "NET006"
	ErrCodeTimeoutNetwork   ErrorCode = "NET007"

	// Security errors (SEC001-SEC099)
	ErrCodeSecurityError    ErrorCode = "SEC001"
	ErrCodeAuthFailed       ErrorCode = "SEC002"
	ErrCodeUnauthorized     ErrorCode = "SEC003"
	ErrCodeForbidden        ErrorCode = "SEC004"
	ErrCodeInvalidToken     ErrorCode = "SEC005"
	ErrCodeReplayAttack     ErrorCode = "SEC006"
	ErrCodeFingerprintMismatch ErrorCode = "SEC007"
	ErrCodeInvalidTLSVersion ErrorCode = "SEC008"
	ErrCodeInvalidJA3Hash   ErrorCode = "SEC009"
	ErrCodeInvalidJA4Hash   ErrorCode = "SEC010"

	// configurationerror (CFG001-CFG099)
	ErrCodeConfigError      ErrorCode = "CFG001"
	ErrCodeInvalidConfig    ErrorCode = "CFG002"
	ErrCodeMissingConfig    ErrorCode = "CFG003"
	ErrCodeConfigParseError ErrorCode = "CFG004"
)

// ErrorCategory gets error code category
func (ec ErrorCode) Category() string {
	if len(ec) >= 3 {
		switch ec[:3] {
		case "VAL":
			return "Validation"
		case "NTF":
			return "NotFound"
		case "SYS":
			return "System"
		case "NET":
			return "Network"
		case "SEC":
			return "Security"
		case "CFG":
			return "Configuration"
		}
	}
	return "Unknown"
}

// HTTPStatus gets HTTP status code for error code
func (ec ErrorCode) HTTPStatus() int {
	switch ec.Category() {
	case "Validation":
		return 400
	case "NotFound":
		return 404
	case "System":
		return 500
	case "Network":
		return 503
	case "Security":
		return 403
	case "Configuration":
		return 500
	default:
		return 500
	}
}

// ============================================================================
// core error definitions
// ============================================================================

var (
	// ErrInvalidProfile invalid fingerprint configuration
	ErrInvalidProfile = errors.New("invalid fingerprint profile")

	// ErrProfileNotFound fingerprint configuration not found
	ErrProfileNotFound = errors.New("fingerprint profile not found")

	// ErrUnsupportedBrowser unsupported browser type
	ErrUnsupportedBrowser = errors.New("unsupported browser type")

	// ErrInvalidTLSVersion invalid TLS version
	ErrInvalidTLSVersion = errors.New("invalid TLS version")

	// ErrInvalidJA3Hash invalid JA3 hash
	ErrInvalidJA3Hash = errors.New("invalid JA3 hash")

	// ErrInvalidJA4Hash invalid JA4 hash
	ErrInvalidJA4Hash = errors.New("invalid JA4 hash")

	// ErrFeatureExtractionFailed featureextractfailed
	ErrFeatureExtractionFailed = errors.New("feature extraction failed")

	// ErrClassificationFailed classifyfailed
	ErrClassificationFailed = errors.New("classification failed")

	// ErrRiskAssessmentFailed risk assessment failed
	ErrRiskAssessmentFailed = errors.New("risk assessment failed")

	// ErrNilPointer nil pointer error
	ErrNilPointer = errors.New("nil pointer dereference")

	// ErrInvalidInput invalid input
	ErrInvalidInput = errors.New("invalid input")

	// ErrOutOfRange out of range
	ErrOutOfRange = errors.New("value out of range")

	// ErrEmptyValue empty value error
	ErrEmptyValue = errors.New("empty value")

	// ErrAlreadyExists already exists
	ErrAlreadyExists = errors.New("already exists")
)

// ============================================================================
// errortype
// ============================================================================

// CoreError core error type (with error code)
type CoreError struct {
	Code    ErrorCode // error code
	Op      string    // operation
	Err     error     // original error
	Context string    // context info
}

// Error implement error interface
func (e *CoreError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("[%s] %s: %s: %v", e.Code, e.Op, e.Context, e.Err)
	}
	return fmt.Sprintf("[%s] %s: %v", e.Code, e.Op, e.Err)
}

// Unwrap returns original error
func (e *CoreError) Unwrap() error {
	return e.Err
}

// Is implement errors.Is
func (e *CoreError) Is(target error) bool {
	if t, ok := target.(*CoreError); ok {
		return e.Code == t.Code
	}
	return errors.Is(e.Err, target)
}

// Category geterrorclassify
func (e *CoreError) Category() string {
	return e.Code.Category()
}

// HTTPStatus gets HTTP status code
func (e *CoreError) HTTPStatus() int {
	return e.Code.HTTPStatus()
}

// ============================================================================
// error creation functions
// ============================================================================

// NewError creates new core error
func NewError(op string, err error) *CoreError {
	return &CoreError{Op: op, Err: err}
}

// NewErrorWithContext creates core error with context
func NewErrorWithContext(op, context string, err error) *CoreError {
	return &CoreError{Op: op, Context: context, Err: err}
}

// NewCodedError creates core error with error code
func NewCodedError(code ErrorCode, op string, err error) *CoreError {
	return &CoreError{Code: code, Op: op, Err: err}
}

// NewCodedErrorf creates error with formatted context
func NewCodedErrorf(code ErrorCode, op, format string, args ...interface{}) *CoreError {
	return &CoreError{
		Code:    code,
		Op:      op,
		Context: fmt.Sprintf(format, args...),
		Err:     errors.New("validation failed"),
	}
}

// WrapError wraps error
func WrapError(code ErrorCode, op string, err error, context string) *CoreError {
	return &CoreError{Code: code, Op: op, Err: err, Context: context}
}

// WrapErrorf wraps error with formatted string
func WrapErrorf(code ErrorCode, op string, err error, format string, args ...interface{}) *CoreError {
	return &CoreError{
		Code:    code,
		Op:      op,
		Err:     err,
		Context: fmt.Sprintf(format, args...),
	}
}

// IsCoreError checks if it is a CoreError
func IsCoreError(err error) bool {
	var ce *CoreError
	return errors.As(err, &ce)
}

// GetErrorCode gets error code (returns empty string if not CoreError)
func GetErrorCode(err error) ErrorCode {
	var ce *CoreError
	if errors.As(err, &ce) {
		return ce.Code
	}
	return ""
}

// IsErrorCode checks if error matches specified error code
func IsErrorCode(err error, code ErrorCode) bool {
	return GetErrorCode(err) == code
}
