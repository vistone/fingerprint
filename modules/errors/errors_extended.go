package errors

import (
	"errors"
	"fmt"
)

// ============================================================================
// domain-specific error creation functions
// ============================================================================

// NewConfigError createconfigurationerror
func NewConfigError(op string, err error) error {
	return Wrapf(err, "config operation '%s' failed", op)
}

// NewValidationError createverifyerror
func NewValidationError(field string, reason string) error {
	return NewCategorizedError(CategoryInput, ErrInvalidInput,
		fmt.Sprintf("field '%s': %s", field, reason))
}

// NewNotFoundError creates not found error
func NewNotFoundError(resource string, identifier string) error {
	return NewCategorizedError(CategoryNotFound, ErrProfileNotFound,
		fmt.Sprintf("%s '%s' not found", resource, identifier))
}

// NewProtocolError createprotocolerror
func NewProtocolError(protocol string, err error) error {
	return NewCategorizedError(CategoryProtocol, err,
		fmt.Sprintf("protocol '%s' error", protocol))
}

// NewNetworkError creates network error
func NewNetworkError(op string, err error) error {
	return NewCategorizedError(CategoryNetwork, err,
		fmt.Sprintf("network operation '%s' failed", op))
}

// ============================================================================
// error type with error code
// ============================================================================

// CodeError error with error code
type CodeError struct {
	Code    ErrorCode
	Message string
	Cause   error
	Details map[string]any
}

func (e *CodeError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *CodeError) Unwrap() error {
	return e.Cause
}

// WithDetail adds detailed information
func (e *CodeError) WithDetail(key string, value any) *CodeError {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}

// NewError creates error with error code
func NewError(code ErrorCode, message string) *CodeError {
	return &CodeError{
		Code:    code,
		Message: message,
		Details: make(map[string]any),
	}
}

// NewErrorWithCause creates error with error code and cause
func NewErrorWithCause(code ErrorCode, message string, cause error) *CodeError {
	return &CodeError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Details: make(map[string]any),
	}
}

// ============================================================================
// convenience error creation functions
// ============================================================================

// ProfileNotFound creates Profile not found error
func ProfileNotFound(id string) *CodeError {
	return NewError(ErrCodeProfileNotFound, "profile not found").
		WithDetail("profile_id", id)
}

// ProfileInvalid creates Profile invalid error
func ProfileInvalid(id string, reason string) *CodeError {
	return NewError(ErrCodeProfileInvalid, "profile is invalid").
		WithDetail("profile_id", id).
		WithDetail("reason", reason)
}

// ProfileRemoveDefault creates error for cannot remove default Profile
func ProfileRemoveDefault(id string) *CodeError {
	return NewError(ErrCodeProfileRemoveDefault, "cannot remove default profile").
		WithDetail("profile_id", id)
}

// ConfigNotLoaded creates configuration not loaded error
func ConfigNotLoaded(path string) *CodeError {
	return NewError(ErrCodeConfigNotLoaded, "config not loaded").
		WithDetail("path", path)
}

// InvalidInput creates invalid input error
func InvalidInput(field string, reason string) *CodeError {
	return NewError(ErrCodeInputInvalid, "invalid input").
		WithDetail("field", field).
		WithDetail("reason", reason)
}

// RequiredField creates required field missing error
func RequiredField(field string) *CodeError {
	return NewError(ErrCodeInputRequired, "required field is missing").
		WithDetail("field", field)
}

// Internal creates internal error
func Internal(op string, cause error) *CodeError {
	return NewErrorWithCause(ErrCodeInternal, "internal error occurred", cause).
		WithDetail("operation", op)
}

// Timeout createtimeouterror
func Timeout(op string, duration any) *CodeError {
	return NewError(ErrCodeTimeout, "operation timeout").
		WithDetail("operation", op).
		WithDetail("duration", duration)
}

// ============================================================================
// error code queries
// ============================================================================

// GetCode gets error code
func GetCode(err error) ErrorCode {
	if err == nil {
		return ""
	}
	var codeErr *CodeError
	if As(err, &codeErr) {
		return codeErr.Code
	}
	return ErrCodeSystem
}

// IsCode checks if error matches specified error code
func IsCode(err error, code ErrorCode) bool {
	return GetCode(err) == code
}

// ============================================================================
// extended error code definitions (merged from core/errors.go)
// ============================================================================

const (
	// Validation errors (VAL001-VAL099)
	ErrCodeInvalidInput   ErrorCode = "VAL001"
	ErrCodeRequiredField  ErrorCode = "VAL002"
	ErrCodeInvalidFormat  ErrorCode = "VAL003"
	ErrCodeOutOfRange     ErrorCode = "VAL004"
	ErrCodeNilPointer     ErrorCode = "VAL005"
	ErrCodeInvalidType    ErrorCode = "VAL006"
	ErrCodeEmptyValue     ErrorCode = "VAL007"
	ErrCodeInvalidLength  ErrorCode = "VAL008"
	ErrCodeInvalidCharset ErrorCode = "VAL009"
	ErrCodeInvalidPattern ErrorCode = "VAL010"

	// Not Found errors (NTF001-NTF099)
	ErrCodeProfileNTF  ErrorCode = "NTF001"
	ErrCodeResourceNTF ErrorCode = "NTF002"
	ErrCodeKeyNTF      ErrorCode = "NTF003"
	ErrCodeFileNTF     ErrorCode = "NTF004"
	ErrCodeRouteNTF    ErrorCode = "NTF005"

	// Security errors (SEC001-SEC099)
	ErrCodeSecurityError       ErrorCode = "SEC001"
	ErrCodeAuthFailed          ErrorCode = "SEC002"
	ErrCodeUnauthorized        ErrorCode = "SEC003"
	ErrCodeForbidden           ErrorCode = "SEC004"
	ErrCodeInvalidToken        ErrorCode = "SEC005"
	ErrCodeReplayAttack        ErrorCode = "SEC006"
	ErrCodeFingerprintMismatch ErrorCode = "SEC007"
	ErrCodeInvalidTLSVersion   ErrorCode = "SEC008"
	ErrCodeInvalidJA3Hash      ErrorCode = "SEC009"
	ErrCodeInvalidJA4Hash      ErrorCode = "SEC010"
)

// Category returns the error code category string
func (ec ErrorCode) Category() string {
	if len(ec) >= 3 {
		switch string(ec[:3]) {
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
		case "PRF":
			return "Profile"
		case "PRT":
			return "Protocol"
		case "INP":
			return "Input"
		case "CCH":
			return "Cache"
		case "ML":
			return "MachineLearning"
		case "PLG":
			return "Plugin"
		}
	}
	return "Unknown"
}

// HTTPStatus returns the HTTP status code for the error code category
func (ec ErrorCode) HTTPStatus() int {
	switch ec.Category() {
	case "Validation", "Input":
		return 400
	case "NotFound", "Profile", "Cache":
		return 404
	case "System", "Configuration", "MachineLearning", "Plugin":
		return 500
	case "Network", "Protocol":
		return 503
	case "Security":
		return 403
	default:
		return 500
	}
}

// ============================================================================
// extended sentinel errors (merged from core/errors.go)
// ============================================================================

var (
	// ErrInvalidProfile invalid fingerprint profile
	ErrInvalidProfile = errors.New("invalid fingerprint profile")

	// ErrInvalidJA3Hash invalid JA3 hash
	ErrInvalidJA3Hash = errors.New("invalid JA3 hash")

	// ErrInvalidJA4Hash invalid JA4 hash
	ErrInvalidJA4Hash = errors.New("invalid JA4 hash")

	// ErrFeatureExtractionFailed feature extraction failed
	ErrFeatureExtractionFailed = errors.New("feature extraction failed")

	// ErrClassificationFailed classification failed
	ErrClassificationFailed = errors.New("classification failed")

	// ErrRiskAssessmentFailed risk assessment failed
	ErrRiskAssessmentFailed = errors.New("risk assessment failed")

	// ErrNilPointer nil pointer dereference
	ErrNilPointer = errors.New("nil pointer dereference")

	// ErrOutOfRange value out of range
	ErrOutOfRange = errors.New("value out of range")

	// ErrEmptyValue empty value
	ErrEmptyValue = errors.New("empty value")

	// ErrAlreadyExists resource already exists
	ErrAlreadyExists = errors.New("already exists")
)

// ============================================================================
// CoreError type (merged from core/errors.go)
// ============================================================================

// CoreError is a structured error type with error code, operation, and context
type CoreError struct {
	Code    ErrorCode
	Op      string
	Err     error
	Context string
}

// Error implements the error interface
func (e *CoreError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("[%s] %s: %s: %v", e.Code, e.Op, e.Context, e.Err)
	}
	return fmt.Sprintf("[%s] %s: %v", e.Code, e.Op, e.Err)
}

// Unwrap returns the underlying error
func (e *CoreError) Unwrap() error {
	return e.Err
}

// Is implements errors.Is matching by error code
func (e *CoreError) Is(target error) bool {
	if t, ok := target.(*CoreError); ok {
		return e.Code == t.Code
	}
	return errors.Is(e.Err, target)
}

// CoreCategory returns the error category string
func (e *CoreError) CoreCategory() string {
	return e.Code.Category()
}

// CoreHTTPStatus returns the HTTP status code for this error
func (e *CoreError) CoreHTTPStatus() int {
	return e.Code.HTTPStatus()
}

// NewCoreError creates a new CoreError with operation and underlying error
func NewCoreError(op string, err error) *CoreError {
	return &CoreError{Op: op, Err: err}
}

// NewCoreErrorWithContext creates a CoreError with context
func NewCoreErrorWithContext(op, context string, err error) *CoreError {
	return &CoreError{Op: op, Context: context, Err: err}
}

// NewCodedError creates a CoreError with error code
func NewCodedError(code ErrorCode, op string, err error) *CoreError {
	return &CoreError{Code: code, Op: op, Err: err}
}

// NewCodedErrorf creates a CoreError with formatted context
func NewCodedErrorf(code ErrorCode, op, format string, args ...interface{}) *CoreError {
	return &CoreError{
		Code:    code,
		Op:      op,
		Context: fmt.Sprintf(format, args...),
		Err:     errors.New("validation failed"),
	}
}

// WrapError wraps an error with error code, operation, and context
func WrapError(code ErrorCode, op string, err error, context string) *CoreError {
	return &CoreError{Code: code, Op: op, Err: err, Context: context}
}

// WrapErrorf wraps an error with formatted context
func WrapErrorf(code ErrorCode, op string, err error, format string, args ...interface{}) *CoreError {
	return &CoreError{
		Code:    code,
		Op:      op,
		Err:     err,
		Context: fmt.Sprintf(format, args...),
	}
}

// IsCoreError checks if the error is a CoreError
func IsCoreError(err error) bool {
	var ce *CoreError
	return errors.As(err, &ce)
}

// GetErrorCode returns the ErrorCode from a CoreError, or empty string
func GetErrorCode(err error) ErrorCode {
	var ce *CoreError
	if errors.As(err, &ce) {
		return ce.Code
	}
	return ""
}

// IsErrorCode checks if the error has the specified error code
func IsErrorCode(err error, code ErrorCode) bool {
	return GetErrorCode(err) == code
}
