package extension

import (
	"fmt"
	"time"
)

// DefaultValidator is the default input validator
//
// Usage example:
//
//	validator := extension.NewDefaultValidator()
//	validator.MaxDataSize = 8192  // 8KB
//	validator.StrictMode = true
//
//	if err := validator.ValidateData(data); err != nil {
//	    return err
//	}
//
//	if err := validator.ValidateMetadata(metadata); err != nil {
//	    return err
//	}
//
// Validation items:
//   - Data is non-empty and does not exceed MaxDataSize
//   - Metadata required fields (Type, Name)
//   - Metadata field size limits (Name ≤256B, Description ≤1KB)
//   - Configuration item size limits (max 1000 keys)
type DefaultValidator struct {
	// Max data size (bytes)
	MaxDataSize int

	// Max extensions
	MaxExtensions int

	// Enable strict mode
	StrictMode bool
}

// NewDefaultValidator creates a default validator
func NewDefaultValidator() *DefaultValidator {
	return &DefaultValidator{
		MaxDataSize:   65536, // 64KB
		MaxExtensions: 10000, // 10K extensions
		StrictMode:    true,
	}
}

// ValidateData validates extension data
func (v *DefaultValidator) ValidateData(data []byte) error {
	if data == nil {
		return NewError(ErrCodeInvalidInput, "data cannot be nil")
	}

	if len(data) == 0 {
		return NewError(ErrCodeInvalidInput, "data cannot be empty")
	}

	if len(data) > v.MaxDataSize {
		return NewError(ErrCodeFieldSizeMismatch,
			fmt.Sprintf("data size exceeds limit: %d > %d", len(data), v.MaxDataSize)).
			WithContext("actual_size", len(data)).
			WithContext("max_size", v.MaxDataSize)
	}

	return nil
}

// ValidateMetadata validates metadata
func (v *DefaultValidator) ValidateMetadata(metadata *ExtensionMetadata) error {
	if metadata == nil {
		return NewError(ErrCodeInvalidMetadata, "metadata cannot be nil")
	}

	// Check required fields
	if metadata.Type == 0 {
		return NewError(ErrCodeMissingField, "extension type cannot be 0")
	}

	if metadata.Name == "" {
		return NewError(ErrCodeMissingField, "extension name cannot be empty")
	}

	if len(metadata.Name) > 256 {
		return NewError(ErrCodeFieldSizeMismatch,
			"extension name too long").
			WithContext("length", len(metadata.Name))
	}

	if len(metadata.Description) > 1024 {
		return NewError(ErrCodeFieldSizeMismatch,
			"extension description too long")
	}

	// Check TLS versions
	if len(metadata.CompatibleTLSVersions) == 0 && !v.StrictMode {
		return NewWarning("no compatible TLS versions specified")
	}

	return nil
}

// ValidateConfig validates configuration
func (v *DefaultValidator) ValidateConfig(config map[string]interface{}) error {
	if config == nil {
		return nil // configuration can be empty
	}

	// Check configuration size
	if len(config) > 1000 {
		return NewError(ErrCodeFieldSizeMismatch,
			"too many configuration keys")
	}

	// Validate each configuration item
	for key, value := range config {
		if key == "" {
			return NewError(ErrCodeInvalidConfig,
				"configuration key cannot be empty")
		}

		if len(key) > 256 {
			return NewError(ErrCodeInvalidConfig,
				"configuration key too long")
		}

		// Check value type
		switch value.(type) {
		case nil, bool, int, int32, int64, uint, uint32, uint64, float32, float64, string:
			// Allowed basic types
		case []byte:
			b := value.([]byte)
			if len(b) > v.MaxDataSize {
				return NewError(ErrCodeFieldSizeMismatch,
					"configuration value too large")
			}
		default:
			if v.StrictMode {
				return NewError(ErrCodeInvalidConfig,
					fmt.Sprintf("unsupported value type: %T", value))
			}
		}
	}

	return nil
}

// Logger is the logging interface
type Logger interface {
	// Log info
	Info(msg string, args ...interface{})

	// Log warning
	Warn(msg string, args ...interface{})

	// Log error
	Error(msg string, err error, args ...interface{})

	// Log debug
	Debug(msg string, args ...interface{})

	// Log panic
	Fatal(msg string, args ...interface{})
}

// SimpleLogger is a simple logger implementation
type SimpleLogger struct {
	name  string
	level int // 0=debug, 1=info, 2=warn, 3=error, 4=fatal
}

// NewSimpleLogger creates a simple logger
func NewSimpleLogger(name string) *SimpleLogger {
	return &SimpleLogger{
		name:  name,
		level: 1, // default info level
	}
}

// SetLevel sets the log level
func (sl *SimpleLogger) SetLevel(level int) {
	if level >= 0 && level <= 4 {
		sl.level = level
	}
}

// Debug logs debug information
func (sl *SimpleLogger) Debug(msg string, args ...interface{}) {
	if sl.level <= 0 {
		sl.log("DEBUG", msg, args...)
	}
}

// Info logs information
func (sl *SimpleLogger) Info(msg string, args ...interface{}) {
	if sl.level <= 1 {
		sl.log("INFO", msg, args...)
	}
}

// Warn logs a warning
func (sl *SimpleLogger) Warn(msg string, args ...interface{}) {
	if sl.level <= 2 {
		sl.log("WARN", msg, args...)
	}
}

// Error logs an error
func (sl *SimpleLogger) Error(msg string, err error, args ...interface{}) {
	if sl.level <= 3 {
		args = append([]interface{}{err}, args...)
		sl.log("ERROR", msg, args...)
	}
}

// Fatal logs a fatal error
func (sl *SimpleLogger) Fatal(msg string, args ...interface{}) {
	sl.log("FATAL", msg, args...)
}

func (sl *SimpleLogger) log(level, msg string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05.000")
	prefix := fmt.Sprintf("[%s] [%s] [%s]", timestamp, sl.name, level)
	if len(args) > 0 {
		fmt.Printf("%s %s %v\n", prefix, msg, args)
	} else {
		fmt.Printf("%s %s\n", prefix, msg)
	}
}

// InputSanitizer is the input sanitizer
type InputSanitizer struct {
	maxFieldSize int
	allowedChars map[rune]bool
}

// NewInputSanitizer creates an input sanitizer
func NewInputSanitizer() *InputSanitizer {
	return &InputSanitizer{
		maxFieldSize: 1024,
		allowedChars: make(map[rune]bool),
	}
}

// SanitizeString sanitizes a string
func (is *InputSanitizer) SanitizeString(s string) (string, error) {
	if len(s) > is.maxFieldSize {
		return "", NewError(ErrCodeFieldSizeMismatch,
			fmt.Sprintf("string exceeds max size: %d > %d", len(s), is.maxFieldSize))
	}

	// Check for illegal characters
	for _, r := range s {
		// Allow basic printable ASCII characters and UTF-8
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return "", NewError(ErrCodeEncodingError,
				fmt.Sprintf("invalid character: %d", r)).
				WithContext("position", len(s))
		}
	}

	return s, nil
}

// SanitizeBytes sanitizes a byte slice
func (is *InputSanitizer) SanitizeBytes(b []byte, maxLen int) ([]byte, error) {
	if len(b) > maxLen {
		return nil, NewError(ErrCodeFieldSizeMismatch,
			fmt.Sprintf("bytes exceed max size: %d > %d", len(b), maxLen))
	}

	// Check for illegal control characters
	for i, b := range b {
		if b < 32 && b != 0x00 {
			// Disallow control characters other than null bytes
			return nil, NewError(ErrCodeEncodingError,
				fmt.Sprintf("illegal control byte at position %d: 0x%02x", i, b))
		}
	}

	return b, nil
}

// SafeParseInt safely parses an integer
func SafeParseInt(s string, base, bitSize int) (int64, error) {
	if len(s) > 20 { // int64 has at most 20 digits
		return 0, NewError(ErrCodeInvalidFormat,
			"integer string too long")
	}

	// Use basic validation to avoid panics
	for _, c := range s {
		if !((c >= '0' && c <= '9') || c == '-' || c == '+') {
			return 0, NewError(ErrCodeEncodingError,
				fmt.Sprintf("invalid character in integer: %c", c))
		}
	}

	return 0, nil // actual parsing is handled by the caller
}

// RecoveryManager is the recovery manager
type RecoveryManager struct {
	handlers []ErrorHandler
	logger   Logger
}

// NewRecoveryManager creates a recovery manager
func NewRecoveryManager(logger Logger) *RecoveryManager {
	return &RecoveryManager{
		handlers: make([]ErrorHandler, 0),
		logger:   logger,
	}
}

// RegisterHandler registers an error handler
func (rm *RecoveryManager) RegisterHandler(handler ErrorHandler) error {
	if handler == nil {
		return NewError(ErrCodeInvalidInput, "handler cannot be nil")
	}
	rm.handlers = append(rm.handlers, handler)
	return nil
}

// Handle handles an error
func (rm *RecoveryManager) Handle(err error) error {
	if err == nil {
		return nil
	}

	extErr, ok := err.(*Error)
	if !ok {
		// Convert standard error
		extErr = NewErrorWithCause(ErrCodeSystemError, "unexpected error", err)
	}

	rm.logger.Error("error occurred", extErr)

	// Try to handle with handlers
	for _, handler := range rm.handlers {
		if handler.CanHandle(extErr) {
			return handler.Handle(extErr)
		}
	}

	return extErr
}

// IsRecoverable determines whether an error is recoverable
func (rm *RecoveryManager) IsRecoverable(err error) bool {
	if extErr, ok := err.(*Error); ok {
		return extErr.IsRecoverable()
	}
	return true
}
