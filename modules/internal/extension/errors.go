// Package extension provides core functionality for the extension system: error handling, validation, defense, and logging
//
// # Error Handling
//
// Uses the Error struct and ErrorCode enum for unified error handling:
//
//	err := extension.NewError(extension.ErrCodeInvalidInput, "invalid input")
//	err = err.WithContext("user_id", 123).WithContext("field", "email")
//	if err.IsRecoverable() {
//	    // can retry
//	    return retryOperation()
//	}
//	if err.IsFatal() {
//	    return err
//	}
//
// # Error Code System
//
// 46 error codes divided into 8 categories (1000-8999):
//
//	Registry errors  (1000-1999): NotFound, AlreadyRegistered, InvalidMetadata
//	Validation errors    (2000-2999): InvalidInput, FieldSizeMismatch, EncodingError
//	Parsing errors    (3000-3999): ParseFailed, InvalidFormat, MalformedData
//	Analysis errors    (4000-4999): AnalysisFailed, ResourceExhausted
//	Configuration errors    (5000-5999): InvalidConfig, MissingConfig
//	Plugin errors    (6000-6999): PluginNotFound, PluginLoadFailed
//	System errors    (7000-7999): SystemError, MemoryExhausted, Timeout
//	Security errors    (8000-8999): SecurityViolation, Unauthorized
//
// # Severity Levels
//
// 5 error severity levels:
//
//	SeverityInfo      (0) - Info, recoverable
//	SeverityWarning   (1) - Warning, recoverable
//	SeverityError     (2) - Error, recoverable
//	SeverityCritical  (3) - Critical, non-recoverable
//	SeverityFatal     (4) - Fatal, non-recoverable
//
// # Input Validation
//
// Using DefaultValidator for data validation:
//
//	validator := extension.NewDefaultValidator()
//	if err := validator.ValidateData(data); err != nil {
//	    return err
//	}
//
// # Defense System
//
// Using RequestGuard to protect API endpoints:
//
//	policy := extension.DefaultDefensePolicy()
//	guard := extension.NewRequestGuard(policy)
//	if err := guard.ValidateRequest(request); err != nil {
//	    return err
//	}
//
// # Safe Execution
//
// Catches panics and recovers:
//
//	extension.SafeExecuteWithRecovery(
//	    func() error { return unsafeOp() },
//	    func(r interface{}) { log.Printf("panic: %v", r) },
//	)
package extension

import (
	"fmt"
)

// ErrorCode represents an error code (for system integration and error classification)
type ErrorCode int

const (
	// Registry errors (1000-1999)
	ErrCodeNotFound ErrorCode = 1000 + iota
	ErrCodeAlreadyRegistered
	ErrCodeInvalidMetadata
	ErrCodeRegistryFull
)

const (
	// Validation errors (2000-2999)
	ErrCodeValidationFailed ErrorCode = 2000 + iota
	ErrCodeInvalidInput
	ErrCodeMissingField
	ErrCodeFieldSizeMismatch
	ErrCodeEncodingError
	ErrCodeVersionMismatch
)

const (
	// Parsing errors (3000-3999)
	ErrCodeParseFailed ErrorCode = 3000 + iota
	ErrCodeInvalidFormat
	ErrCodeUnexpectedEOF
	ErrCodeInvalidOffset
	ErrCodeMalformedData
)

const (
	// Analysis errors (4000-4999)
	ErrCodeAnalysisFailed ErrorCode = 4000 + iota
	ErrCodeAnalysisTimeout
	ErrCodeResourceExhausted
	ErrCodeInternalError
)

const (
	// Configuration errors (5000-5999)
	ErrCodeInvalidConfig ErrorCode = 5000 + iota
	ErrCodeMissingConfig
	ErrCodeConfigConflict
)

const (
	// Plugin errors (6000-6999)
	ErrCodePluginNotFound ErrorCode = 6000 + iota
	ErrCodePluginInitFailed
	ErrCodePluginLoadFailed
	ErrCodePluginVersionMismatch
)

const (
	// System errors (7000-7999)
	ErrCodeSystemError ErrorCode = 7000 + iota
	ErrCodeMemoryExhausted
	ErrCodeTimeout
	ErrCodeCancelled
)

const (
	// Security errors (8000-8999)
	ErrCodeSecurityViolation ErrorCode = 8000 + iota
	ErrCodeUnauthorized
	ErrCodeForbidden
)

// Error is the standard error type for the extension system
//
// Field descriptions:
//
//	Code       - Error code for classification and integration
//	Message    - Human-readable error description
//	Cause      - Original error, supports error chaining
//	Context    - Error context key-value pairs (for debugging)
//	Severity   - Error severity level (determines recoverability)
//	Timestamp  - Time when the error occurred
//
// Usage example:
//
//	// Create error
//	err := extension.NewError(
//	    extension.ErrCodeInvalidInput,
//	    "invalid user input",
//	)
//
//	// Add context information (for debugging)
//	err = err.WithContext("user_id", 12345)
//	err = err.WithContext("field", "email")
//	err = err.WithContext("value", userEmail)
//
//	// Determine error type
//	if err.IsRecoverable() {
//	    // Info, Warning, Error level errors are recoverable
//	    return retryOperation()
//	}
//
//	if err.IsFatal() {
//	    // Fatal level errors are non-recoverable
//	    return err
//	}
//
// Creation variants:
//
//	extension.NewError(code, message)                    // Standard error
//	extension.NewErrorWithCause(code, message, cause)   // Error with cause
//	extension.NewFatalError(code, message)              // Fatal error
//	extension.NewWarning(message)                       // Warning (no code needed)
//
// Error chaining:
//
//	err := processData(data)
//	if err != nil {
//	    return extension.NewErrorWithCause(
//	        extension.ErrCodeAnalysisFailed,
//	        "analysis failed",
//	        err,  // original error as the next in the chain
//	    )
//	}
type Error struct {
	Code      ErrorCode
	Message   string
	Cause     error
	Context   map[string]interface{}
	Severity  ErrorSeverity
	Timestamp int64
}

// ErrorSeverity represents error severity level
type ErrorSeverity int

const (
	SeverityInfo ErrorSeverity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
	SeverityFatal
)

// String implements the error interface
func (e *Error) Error() string {
	msg := fmt.Sprintf("[%s] %s (code: %d)", e.Severity.String(), e.Message, e.Code)
	if e.Cause != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Cause)
	}
	return msg
}

// Unwrap supports error wrapping
func (e *Error) Unwrap() error {
	return e.Cause
}

// WithContext adds context information
func (e *Error) WithContext(key string, value interface{}) *Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// IsRecoverable determines whether the error is recoverable
func (e *Error) IsRecoverable() bool {
	switch e.Severity {
	case SeverityFatal, SeverityCritical:
		return false
	default:
		return true
	}
}

// IsFatal determines whether the error is fatal
func (e *Error) IsFatal() bool {
	return e.Severity == SeverityFatal
}

// String implements the Stringer interface
func (es ErrorSeverity) String() string {
	switch es {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRITICAL"
	case SeverityFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// NewError creates a new error
func NewError(code ErrorCode, message string) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		Context:  make(map[string]interface{}),
		Severity: SeverityError,
	}
}

// NewErrorWithCause creates an error with a cause
func NewErrorWithCause(code ErrorCode, message string, cause error) *Error {
	err := NewError(code, message)
	err.Cause = cause
	return err
}

// NewFatalError creates a fatal error
func NewFatalError(code ErrorCode, message string) *Error {
	err := NewError(code, message)
	err.Severity = SeverityFatal
	return err
}

// NewWarning creates a warning
func NewWarning(message string) *Error {
	err := &Error{
		Code:     ErrCodeInvalidInput,
		Message:  message,
		Context:  make(map[string]interface{}),
		Severity: SeverityWarning,
	}
	return err
}

// ToError converts a standard error to an extension error
func ToError(code ErrorCode, err error) *Error {
	if extErr, ok := err.(*Error); ok {
		return extErr
	}
	return NewErrorWithCause(code, "operation failed", err)
}

// ErrorHandler is the error handler interface
type ErrorHandler interface {
	// Handle the error
	Handle(err *Error) error

	// Determine whether this error can be handled
	CanHandle(err *Error) bool

	// Get handler name
	GetName() string
}

// PanicHandler is the panic recovery handler
type PanicHandler struct {
	name string
}

func NewPanicHandler() *PanicHandler {
	return &PanicHandler{name: "panic_handler"}
}

func (ph *PanicHandler) Handle(err *Error) error {
	if err.Severity == SeverityFatal {
		// Record fatal error, but do not panic
		return fmt.Errorf("fatal error occurred: %w", err)
	}
	return err
}

func (ph *PanicHandler) CanHandle(err *Error) bool {
	return true // can handle any error
}

func (ph *PanicHandler) GetName() string {
	return ph.name
}

// SafeExecute safely executes a function, catching panics
// If a panic occurs, it returns an error containing the panic information
func SafeExecute(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// panic has been caught, convert panic to error and return
			err = NewErrorWithCause(ErrCodeSystemError,
				fmt.Sprintf("panic recovered: %v", r), nil)
		}
	}()

	return fn()
}

// SafeExecuteWithRecovery safely executes a function with recovery logic
func SafeExecuteWithRecovery(fn func() error, handler func(interface{})) error {
	defer func() {
		if r := recover(); r != nil {
			handler(r)
		}
	}()

	return fn()
}
