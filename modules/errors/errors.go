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
