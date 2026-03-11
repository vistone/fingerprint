// Package errors provides the unified error system for the fingerprint project.
// This is the canonical error package — all error codes, sentinel errors,
// error types, and helper functions live here.
// modules/core/errors.go re-exports a subset for backward compatibility.
package errors

import (
	"errors"
	"fmt"
	"strings"
)

// ============================================================================
// error code definitions
// ============================================================================

// ErrorCode error code type
type ErrorCode string

const (
	// system level errors (SYS)
	ErrCodeSystem     ErrorCode = "SYS001"
	ErrCodeNotFound   ErrorCode = "SYS002"
	ErrCodeInvalidArg ErrorCode = "SYS003"
	ErrCodeTimeout    ErrorCode = "SYS004"
	ErrCodeCancelled  ErrorCode = "SYS005"
	ErrCodeInternal   ErrorCode = "SYS006"
	ErrCodeNotImpl    ErrorCode = "SYS007"

	// Profile related errors (PRF)
	ErrCodeProfileNotFound      ErrorCode = "PRF001"
	ErrCodeProfileInvalid       ErrorCode = "PRF002"
	ErrCodeProfileExists        ErrorCode = "PRF003"
	ErrCodeProfileLoadFailed    ErrorCode = "PRF004"
	ErrCodeProfileSaveFailed    ErrorCode = "PRF005"
	ErrCodeProfileNoDefault     ErrorCode = "PRF006"
	ErrCodeProfileRemoveDefault ErrorCode = "PRF007"

	// configuration related errors (CFG)
	ErrCodeConfigNotLoaded    ErrorCode = "CFG001"
	ErrCodeConfigValidation   ErrorCode = "CFG002"
	ErrCodeConfigPathNotFound ErrorCode = "CFG003"
	ErrCodeConfigExists       ErrorCode = "CFG004"
	ErrCodeConfigVersion      ErrorCode = "CFG005"

	// network related errors (NET)
	ErrCodeNetInvalidIP   ErrorCode = "NET001"
	ErrCodeNetInvalidPort ErrorCode = "NET002"
	ErrCodeNetConnect     ErrorCode = "NET003"
	ErrCodeNetTimeout     ErrorCode = "NET004"

	// protocol related errors (PRT)
	ErrCodeProtoTLSVersion  ErrorCode = "PRT001"
	ErrCodeProtoCipherSuite ErrorCode = "PRT002"
	ErrCodeProtoExtension   ErrorCode = "PRT003"
	ErrCodeProtoInvalidSpec ErrorCode = "PRT004"

	// input validation errors (INP)
	ErrCodeInputInvalid  ErrorCode = "INP001"
	ErrCodeInputEmpty    ErrorCode = "INP002"
	ErrCodeInputTooLarge ErrorCode = "INP003"
	ErrCodeInputRequired ErrorCode = "INP004"

	// cache related errors (CCH)
	ErrCodeCacheNotFound ErrorCode = "CCH001"
	ErrCodeCacheExpired  ErrorCode = "CCH002"

	// ML related errors (ML)
	ErrCodeMLModelNotFound ErrorCode = "ML001"
	ErrCodeMLPrediction    ErrorCode = "ML002"

	// plugin related errors (PLG)
	ErrCodePluginNotFound ErrorCode = "PLG001"
	ErrCodePluginInit     ErrorCode = "PLG002"
)

// ============================================================================
// sentinel error definitions - core errors
// ============================================================================

var (
	// configuration related errors
	ErrProfileNotFound               = errors.New("profile not found")
	ErrInvalidFingerprint            = errors.New("invalid fingerprint format")
	ErrClientHelloSpecNotImplemented = errors.New("client hello spec not implemented")
	ErrInvalidUserAgent              = errors.New("invalid user agent")
	ErrNoProfilesAvailable           = errors.New("no profiles available")
	ErrUnsupportedBrowser            = errors.New("unsupported browser")

	// configuration center errors
	ErrConfigNotLoaded     = errors.New("config not loaded")
	ErrConfigValidation    = errors.New("config validation failed")
	ErrConfigPathNotFound  = errors.New("config path not found")
	ErrConfigAlreadyExists = errors.New("config already exists")
	ErrVersionNotFound     = errors.New("version not found")

	// network related errors
	ErrInvalidIP        = errors.New("invalid IP address")
	ErrInvalidPort      = errors.New("invalid port")
	ErrConnectionFailed = errors.New("connection failed")

	// protocol related errors
	ErrInvalidTLSVersion  = errors.New("invalid TLS version")
	ErrInvalidCipherSuite = errors.New("invalid cipher suite")
	ErrInvalidExtension   = errors.New("invalid extension")

	// input validation errors
	ErrInvalidInput  = errors.New("invalid input")
	ErrEmptyInput    = errors.New("empty input")
	ErrInputTooLarge = errors.New("input too large")
	ErrRequiredField = errors.New("required field is missing")

	// internal errors
	ErrInternal       = errors.New("internal error")
	ErrNotImplemented = errors.New("not implemented")
	ErrTimeout        = errors.New("operation timeout")
	ErrCancelled      = errors.New("operation cancelled")
)

// ============================================================================
// errortypeclassify
// ============================================================================

// ErrorCategory errorclassify
type ErrorCategory string

const (
	CategoryConfig   ErrorCategory = "config"
	CategoryNetwork  ErrorCategory = "network"
	CategoryProtocol ErrorCategory = "protocol"
	CategoryInput    ErrorCategory = "input"
	CategoryInternal ErrorCategory = "internal"
	CategoryNotFound ErrorCategory = "not_found"
)

// CategorizedError categorized error
type CategorizedError struct {
	Category ErrorCategory
	Err      error
	Details  string
}

func (e *CategorizedError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Category, e.Err.Error(), e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Category, e.Err.Error())
}

func (e *CategorizedError) Unwrap() error {
	return e.Err
}

// NewCategorizedError creates categorized error
func NewCategorizedError(category ErrorCategory, err error, details string) *CategorizedError {
	return &CategorizedError{
		Category: category,
		Err:      err,
		Details:  details,
	}
}

// ============================================================================
// error wrapping helper functions
// ============================================================================

// Wrap wraps error and adds context
func Wrap(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// Wrapf wraps error using formatted string
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	context := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", context, err)
}

// New creates new error (compatible with standard library)
func New(text string) error {
	return errors.New(text)
}

// Newf creates error using formatted string
func Newf(format string, args ...interface{}) error {
	return errors.New(fmt.Sprintf(format, args...))
}

// Is checks if error matches (compatible with standard library)
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As type assertion (compatible with standard library)
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// ============================================================================
// error checking helper functions
// ============================================================================

const clientHelloSpecNotImplementedMsg = "please implement this method"

// IsClientHelloSpecNotImplemented checks if error indicates profile does not implement ClientHelloSpec.
func IsClientHelloSpecNotImplemented(err error) bool {
	if err == nil {
		return false
	}
	// check sentinel errors first
	if errors.Is(err, ErrClientHelloSpecNotImplemented) {
		return true
	}
	// backward compatibility: check error message string
	return strings.Contains(strings.ToLower(err.Error()), clientHelloSpecNotImplementedMsg)
}

// IsNotFound checks if it is a "not found" error
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrProfileNotFound) ||
		errors.Is(err, ErrConfigPathNotFound) ||
		errors.Is(err, ErrVersionNotFound)
}

// IsInvalidInput checks if it is an "invalid input" error
func IsInvalidInput(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidInput) ||
		errors.Is(err, ErrInvalidFingerprint) ||
		errors.Is(err, ErrInvalidUserAgent)
}

// IsConfigError checks if it is a configuration error
func IsConfigError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrConfigNotLoaded) ||
		errors.Is(err, ErrConfigValidation) ||
		errors.Is(err, ErrConfigPathNotFound)
}

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
