package extension

import (
	"fmt"
	"time"
)

// translated comment
//
// translated comment
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
// translated comment
// translated comment
// translated comment
// translated comment
// translated comment
type DefaultValidator struct {
	// translated comment
	MaxDataSize int

	// translated comment
	MaxExtensions int

	// translated comment
	StrictMode bool
}

// translated comment
func NewDefaultValidator() *DefaultValidator {
	return &DefaultValidator{
		MaxDataSize:   65536, // 64KB
		MaxExtensions: 10000, // translated comment
		StrictMode:    true,
	}
}

// translated comment
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

// translated comment
func (v *DefaultValidator) ValidateMetadata(metadata *ExtensionMetadata) error {
	if metadata == nil {
		return NewError(ErrCodeInvalidMetadata, "metadata cannot be nil")
	}

	// translated comment
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

	// translated comment
	if len(metadata.CompatibleTLSVersions) == 0 && !v.StrictMode {
		return NewWarning("no compatible TLS versions specified")
	}

	return nil
}

// translated comment
func (v *DefaultValidator) ValidateConfig(config map[string]interface{}) error {
	if config == nil {
		return nil // translated comment
	}

	// translated comment
	if len(config) > 1000 {
		return NewError(ErrCodeFieldSizeMismatch,
			"too many configuration keys")
	}

	// translated comment
	for key, value := range config {
		if key == "" {
			return NewError(ErrCodeInvalidConfig,
				"configuration key cannot be empty")
		}

		if len(key) > 256 {
			return NewError(ErrCodeInvalidConfig,
				"configuration key too long")
		}

		// translated comment
		switch value.(type) {
		case nil, bool, int, int32, int64, uint, uint32, uint64, float32, float64, string:
			// translated comment
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

// translated comment
type Logger interface {
	// translated comment
	Info(msg string, args ...interface{})

	// translated comment
	Warn(msg string, args ...interface{})

	// translated comment
	Error(msg string, err error, args ...interface{})

	// translated comment
	Debug(msg string, args ...interface{})

	// translated comment
	Fatal(msg string, args ...interface{})
}

// translated comment
type SimpleLogger struct {
	name  string
	level int // 0=debug, 1=info, 2=warn, 3=error, 4=fatal
}

// translated comment
func NewSimpleLogger(name string) *SimpleLogger {
	return &SimpleLogger{
		name:  name,
		level: 1, // translated comment
	}
}

// translated comment
func (sl *SimpleLogger) SetLevel(level int) {
	if level >= 0 && level <= 4 {
		sl.level = level
	}
}

// translated comment
func (sl *SimpleLogger) Debug(msg string, args ...interface{}) {
	if sl.level <= 0 {
		sl.log("DEBUG", msg, args...)
	}
}

// translated comment
func (sl *SimpleLogger) Info(msg string, args ...interface{}) {
	if sl.level <= 1 {
		sl.log("INFO", msg, args...)
	}
}

// translated comment
func (sl *SimpleLogger) Warn(msg string, args ...interface{}) {
	if sl.level <= 2 {
		sl.log("WARN", msg, args...)
	}
}

// translated comment
func (sl *SimpleLogger) Error(msg string, err error, args ...interface{}) {
	if sl.level <= 3 {
		args = append([]interface{}{err}, args...)
		sl.log("ERROR", msg, args...)
	}
}

// translated comment
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

// translated comment
type InputSanitizer struct {
	maxFieldSize int
	allowedChars map[rune]bool
}

// translated comment
func NewInputSanitizer() *InputSanitizer {
	return &InputSanitizer{
		maxFieldSize: 1024,
		allowedChars: make(map[rune]bool),
	}
}

// translated comment
func (is *InputSanitizer) SanitizeString(s string) (string, error) {
	if len(s) > is.maxFieldSize {
		return "", NewError(ErrCodeFieldSizeMismatch,
			fmt.Sprintf("string exceeds max size: %d > %d", len(s), is.maxFieldSize))
	}

	// translated comment
	for _, r := range s {
		// translated comment
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return "", NewError(ErrCodeEncodingError,
				fmt.Sprintf("invalid character: %d", r)).
				WithContext("position", len(s))
		}
	}

	return s, nil
}

// translated comment
func (is *InputSanitizer) SanitizeBytes(b []byte, maxLen int) ([]byte, error) {
	if len(b) > maxLen {
		return nil, NewError(ErrCodeFieldSizeMismatch,
			fmt.Sprintf("bytes exceed max size: %d > %d", len(b), maxLen))
	}

	// translated comment
	for i, b := range b {
		if b < 32 && b != 0x00 {
			// translated comment
			return nil, NewError(ErrCodeEncodingError,
				fmt.Sprintf("illegal control byte at position %d: 0x%02x", i, b))
		}
	}

	return b, nil
}

// translated comment
func SafeParseInt(s string, base, bitSize int) (int64, error) {
	if len(s) > 20 { // translated comment
		return 0, NewError(ErrCodeInvalidFormat,
			"integer string too long")
	}

	// translated comment
	for _, c := range s {
		if !((c >= '0' && c <= '9') || c == '-' || c == '+') {
			return 0, NewError(ErrCodeEncodingError,
				fmt.Sprintf("invalid character in integer: %c", c))
		}
	}

	return 0, nil // translated comment
}

// translated comment
type RecoveryManager struct {
	handlers []ErrorHandler
	logger   Logger
}

// translated comment
func NewRecoveryManager(logger Logger) *RecoveryManager {
	return &RecoveryManager{
		handlers: make([]ErrorHandler, 0),
		logger:   logger,
	}
}

// translated comment
func (rm *RecoveryManager) RegisterHandler(handler ErrorHandler) error {
	if handler == nil {
		return NewError(ErrCodeInvalidInput, "handler cannot be nil")
	}
	rm.handlers = append(rm.handlers, handler)
	return nil
}

// translated comment
func (rm *RecoveryManager) Handle(err error) error {
	if err == nil {
		return nil
	}

	extErr, ok := err.(*Error)
	if !ok {
		// translated comment
		extErr = NewErrorWithCause(ErrCodeSystemError, "unexpected error", err)
	}

	rm.logger.Error("error occurred", extErr)

	// translated comment
	for _, handler := range rm.handlers {
		if handler.CanHandle(extErr) {
			return handler.Handle(extErr)
		}
	}

	return extErr
}

// translated comment
func (rm *RecoveryManager) IsRecoverable(err error) bool {
	if extErr, ok := err.(*Error); ok {
		return extErr.IsRecoverable()
	}
	return true
}
