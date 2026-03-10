// translated comment
//
// translated comment
//
// translated comment
//
// translated comment
//	err = err.WithContext("user_id", 123).WithContext("field", "email")
//	if err.IsRecoverable() {
// translated comment
//	    return retryOperation()
//	}
//	if err.IsFatal() {
//	    return err
//	}
//
// translated comment
//
// translated comment
//
// translated comment
// translated comment
// translated comment
// translated comment
// translated comment
// translated comment
// translated comment
// translated comment
//
// translated comment
//
// translated comment
//
// translated comment
// translated comment
// translated comment
// translated comment
// translated comment
//
// translated comment
//
// translated comment
//
//	validator := extension.NewDefaultValidator()
//	if err := validator.ValidateData(data); err != nil {
//	    return err
//	}
//
// translated comment
//
// translated comment
//
//	policy := extension.DefaultDefensePolicy()
//	guard := extension.NewRequestGuard(policy)
//	if err := guard.ValidateRequest(request); err != nil {
//	    return err
//	}
//
// translated comment
//
// translated comment
//
//	extension.SafeExecuteWithRecovery(
//	    func() error { return unsafeOp() },
//	    func(r interface{}) { log.Printf("panic: %v", r) },
//	)
package extension

import (
	"fmt"
)

// translated comment
type ErrorCode int

const (
	// translated comment
	ErrCodeNotFound ErrorCode = 1000 + iota
	ErrCodeAlreadyRegistered
	ErrCodeInvalidMetadata
	ErrCodeRegistryFull
)

const (
	// translated comment
	ErrCodeValidationFailed ErrorCode = 2000 + iota
	ErrCodeInvalidInput
	ErrCodeMissingField
	ErrCodeFieldSizeMismatch
	ErrCodeEncodingError
	ErrCodeVersionMismatch
)

const (
	// translated comment
	ErrCodeParseFailed ErrorCode = 3000 + iota
	ErrCodeInvalidFormat
	ErrCodeUnexpectedEOF
	ErrCodeInvalidOffset
	ErrCodeMalformedData
)

const (
	// translated comment
	ErrCodeAnalysisFailed ErrorCode = 4000 + iota
	ErrCodeAnalysisTimeout
	ErrCodeResourceExhausted
	ErrCodeInternalError
)

const (
	// translated comment
	ErrCodeInvalidConfig ErrorCode = 5000 + iota
	ErrCodeMissingConfig
	ErrCodeConfigConflict
)

const (
	// translated comment
	ErrCodePluginNotFound ErrorCode = 6000 + iota
	ErrCodePluginInitFailed
	ErrCodePluginLoadFailed
	ErrCodePluginVersionMismatch
)

const (
	// translated comment
	ErrCodeSystemError ErrorCode = 7000 + iota
	ErrCodeMemoryExhausted
	ErrCodeTimeout
	ErrCodeCancelled
)

const (
	// translated comment
	ErrCodeSecurityViolation ErrorCode = 8000 + iota
	ErrCodeUnauthorized
	ErrCodeForbidden
)

// translated comment
//
// translated comment
//
// translated comment
// translated comment
// translated comment
// translated comment
// translated comment
// translated comment
//
// translated comment
//
// translated comment
//	err := extension.NewError(
//	    extension.ErrCodeInvalidInput,
// translated comment
//	)
//
// translated comment
//	err = err.WithContext("user_id", 12345)
//	err = err.WithContext("field", "email")
//	err = err.WithContext("value", userEmail)
//
// translated comment
//	if err.IsRecoverable() {
// translated comment
//	    return retryOperation()
//	}
//
//	if err.IsFatal() {
// translated comment
//	    return err
//	}
//
// translated comment
//
// translated comment
// translated comment
// translated comment
// translated comment
//
// translated comment
//
//	err := processData(data)
//	if err != nil {
//	    return extension.NewErrorWithCause(
//	        extension.ErrCodeAnalysisFailed,
// translated comment
// translated comment
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

// translated comment
type ErrorSeverity int

const (
	SeverityInfo ErrorSeverity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
	SeverityFatal
)

// translated comment
func (e *Error) Error() string {
	msg := fmt.Sprintf("[%s] %s (code: %d)", e.Severity.String(), e.Message, e.Code)
	if e.Cause != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Cause)
	}
	return msg
}

// translated comment
func (e *Error) Unwrap() error {
	return e.Cause
}

// translated comment
func (e *Error) WithContext(key string, value interface{}) *Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// translated comment
func (e *Error) IsRecoverable() bool {
	switch e.Severity {
	case SeverityFatal, SeverityCritical:
		return false
	default:
		return true
	}
}

// translated comment
func (e *Error) IsFatal() bool {
	return e.Severity == SeverityFatal
}

// translated comment
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

// translated comment
func NewError(code ErrorCode, message string) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		Context:  make(map[string]interface{}),
		Severity: SeverityError,
	}
}

// translated comment
func NewErrorWithCause(code ErrorCode, message string, cause error) *Error {
	err := NewError(code, message)
	err.Cause = cause
	return err
}

// translated comment
func NewFatalError(code ErrorCode, message string) *Error {
	err := NewError(code, message)
	err.Severity = SeverityFatal
	return err
}

// translated comment
func NewWarning(message string) *Error {
	err := &Error{
		Code:     ErrCodeInvalidInput,
		Message:  message,
		Context:  make(map[string]interface{}),
		Severity: SeverityWarning,
	}
	return err
}

// translated comment
func ToError(code ErrorCode, err error) *Error {
	if extErr, ok := err.(*Error); ok {
		return extErr
	}
	return NewErrorWithCause(code, "operation failed", err)
}

// translated comment
type ErrorHandler interface {
	// translated comment
	Handle(err *Error) error

	// translated comment
	CanHandle(err *Error) bool

	// translated comment
	GetName() string
}

// translated comment
type PanicHandler struct {
	name string
}

func NewPanicHandler() *PanicHandler {
	return &PanicHandler{name: "panic_handler"}
}

func (ph *PanicHandler) Handle(err *Error) error {
	if err.Severity == SeverityFatal {
		// translated comment
		return fmt.Errorf("fatal error occurred: %w", err)
	}
	return err
}

func (ph *PanicHandler) CanHandle(err *Error) bool {
	return true // translated comment
}

func (ph *PanicHandler) GetName() string {
	return ph.name
}

// translated comment
// translated comment
func SafeExecute(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// translated comment
			err = NewErrorWithCause(ErrCodeSystemError,
				fmt.Sprintf("panic recovered: %v", r), nil)
		}
	}()

	return fn()
}

// translated comment
func SafeExecuteWithRecovery(fn func() error, handler func(interface{})) error {
	defer func() {
		if r := recover(); r != nil {
			handler(r)
		}
	}()

	return fn()
}
