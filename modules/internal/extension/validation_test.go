package extension

import (
	"bytes"
	"strings"
	"testing"
)

// TestDefaultValidator_ValidateData tests data validation
func TestDefaultValidator_ValidateData(t *testing.T) {
	validator := NewDefaultValidator()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		errCode ErrorCode
	}{
		{
			name:    "valid data passes validation",
			data:    []byte("valid data"),
			wantErr: false,
		},
		{
			name:    "nil data returns error",
			data:    nil,
			wantErr: true,
			errCode: ErrCodeInvalidInput,
		},
		{
			name:    "empty data returns error",
			data:    []byte{},
			wantErr: true,
			errCode: ErrCodeInvalidInput,
		},
		{
			name:    "exceeding MaxDataSize returns error",
			data:    bytes.Repeat([]byte("a"), validator.MaxDataSize+1),
			wantErr: true,
			errCode: ErrCodeFieldSizeMismatch,
		},
		{
			name:    "boundary - exactly equal to MaxDataSize",
			data:    bytes.Repeat([]byte("b"), validator.MaxDataSize),
			wantErr: false,
		},
		{
			name:    "boundary - single byte",
			data:    []byte("x"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateData(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateData() error = nil, wantErr %v", tt.wantErr)
					return
				}
				extErr, ok := err.(*Error)
				if !ok {
					t.Errorf("ValidateData() error is not *Error type")
					return
				}
				if extErr.Code != tt.errCode {
					t.Errorf("ValidateData() error code = %d, want %d", extErr.Code, tt.errCode)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateData() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

// TestDefaultValidator_ValidateMetadata tests metadata validation
func TestDefaultValidator_ValidateMetadata(t *testing.T) {
	validator := NewDefaultValidator()
	validatorStrict := NewDefaultValidator()
	validatorStrict.StrictMode = true

	tests := []struct {
		name       string
		validator  *DefaultValidator
		metadata   *ExtensionMetadata
		wantErr    bool
		isWarning  bool
		errCode    ErrorCode
		errContain string
	}{
		{
			name:      "valid metadata passes validation",
			validator: validator,
			metadata: &ExtensionMetadata{
				Type:                  1,
				Name:                  "test-extension",
				Description:           "A test extension",
				CompatibleTLSVersions: []uint16{0x0301},
			},
			wantErr: false,
		},
		{
			name:      "nil metadata returns error",
			validator: validator,
			metadata:  nil,
			wantErr:   true,
			errCode:   ErrCodeInvalidMetadata,
		},
		{
			name:      "Type=0 returns error",
			validator: validator,
			metadata: &ExtensionMetadata{
				Type: 0,
				Name: "test",
			},
			wantErr: true,
			errCode: ErrCodeMissingField,
		},
		{
			name:      "empty Name returns error",
			validator: validator,
			metadata: &ExtensionMetadata{
				Type: 1,
				Name: "",
			},
			wantErr: true,
			errCode: ErrCodeMissingField,
		},
		{
			name:      "Name exceeding 256 chars returns error",
			validator: validator,
			metadata: &ExtensionMetadata{
				Type: 1,
				Name: strings.Repeat("a", 257),
			},
			wantErr: true,
			errCode: ErrCodeFieldSizeMismatch,
		},
		{
			name:      "Name exactly 256 chars passes",
			validator: validator,
			metadata: &ExtensionMetadata{
				Type: 1,
				Name: strings.Repeat("b", 256),
			},
			wantErr: false,
		},
		{
			name:      "Description exceeding 1024 chars returns error",
			validator: validator,
			metadata: &ExtensionMetadata{
				Type:        1,
				Name:        "test",
				Description: strings.Repeat("c", 1025),
			},
			wantErr: true,
			errCode: ErrCodeFieldSizeMismatch,
		},
		{
			name:      "Description exactly 1024 chars passes",
			validator: validator,
			metadata: &ExtensionMetadata{
				Type:        1,
				Name:        "test",
				Description: strings.Repeat("d", 1024),
			},
			wantErr: false,
		},
		{
			name:      "strict mode without CompatibleTLSVersions passes",
			validator: validatorStrict,
			metadata: &ExtensionMetadata{
				Type:                  1,
				Name:                  "test",
				CompatibleTLSVersions: []uint16{},
			},
			wantErr: false,
		},
		{
			name:      "default validator without CompatibleTLSVersions passes",
			validator: validator,
			metadata: &ExtensionMetadata{
				Type:                  1,
				Name:                  "test",
				CompatibleTLSVersions: []uint16{},
			},
			wantErr: false, // Default StrictMode=true, will not return warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator.ValidateMetadata(tt.metadata)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateMetadata() error = nil, wantErr %v", tt.wantErr)
					return
				}
				extErr, ok := err.(*Error)
				if !ok {
					t.Errorf("ValidateMetadata() error is not *Error type")
					return
				}
				if extErr.Code != tt.errCode {
					t.Errorf("ValidateMetadata() error code = %d, want %d", extErr.Code, tt.errCode)
				}
				if tt.isWarning && extErr.Severity != SeverityWarning {
					t.Errorf("ValidateMetadata() expected warning severity, got %v", extErr.Severity)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateMetadata() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

// TestDefaultValidator_ValidateConfig tests configuration validation
func TestDefaultValidator_ValidateConfig(t *testing.T) {
	validator := NewDefaultValidator()
	strictValidator := NewDefaultValidator()
	strictValidator.StrictMode = true
	nonStrictValidator := NewDefaultValidator()
	nonStrictValidator.StrictMode = false

	tests := []struct {
		name      string
		validator *DefaultValidator
		config    map[string]interface{}
		wantErr   bool
		errCode   ErrorCode
	}{
		{
			name:      "nil configuration passes",
			validator: validator,
			config:    nil,
			wantErr:   false,
		},
		{
			name:      "empty configuration passes",
			validator: validator,
			config:    map[string]interface{}{},
			wantErr:   false,
		},
		{
			name:      "exceeding 1000 keys returns error",
			validator: validator,
			config: func() map[string]interface{} {
				config := make(map[string]interface{})
				for i := 0; i < 1001; i++ {
					config[strings.Repeat("k", i%10)+string(rune(i))] = i
				}
				return config
			}(),
			wantErr: true,
			errCode: ErrCodeFieldSizeMismatch,
		},
		{
			name:      "exactly 1000 keys passes",
			validator: validator,
			config: func() map[string]interface{} {
				config := make(map[string]interface{})
				for i := 0; i < 1000; i++ {
					config[string(rune('a'+i%26))+string(rune(i))] = i
				}
				return config
			}(),
			wantErr: false,
		},
		{
			name:      "empty key returns error",
			validator: validator,
			config: map[string]interface{}{
				"": "value",
			},
			wantErr: true,
			errCode: ErrCodeInvalidConfig,
		},
		{
			name:      "long key returns error",
			validator: validator,
			config: map[string]interface{}{
				strings.Repeat("k", 257): "value",
			},
			wantErr: true,
			errCode: ErrCodeInvalidConfig,
		},
		{
			name:      "exactly 256-char key passes",
			validator: validator,
			config: map[string]interface{}{
				strings.Repeat("k", 256): "value",
			},
			wantErr: false,
		},
		{
			name:      "nil value allowed",
			validator: validator,
			config: map[string]interface{}{
				"nil_key": nil,
			},
			wantErr: false,
		},
		{
			name:      "bool value allowed",
			validator: validator,
			config: map[string]interface{}{
				"bool_true":  true,
				"bool_false": false,
			},
			wantErr: false,
		},
		{
			name:      "int value allowed",
			validator: validator,
			config: map[string]interface{}{
				"int": 42,
			},
			wantErr: false,
		},
		{
			name:      "int32 value allowed",
			validator: validator,
			config: map[string]interface{}{
				"int32": int32(42),
			},
			wantErr: false,
		},
		{
			name:      "int64 value allowed",
			validator: validator,
			config: map[string]interface{}{
				"int64": int64(42),
			},
			wantErr: false,
		},
		{
			name:      "uint value allowed",
			validator: validator,
			config: map[string]interface{}{
				"uint": uint(42),
			},
			wantErr: false,
		},
		{
			name:      "uint32 value allowed",
			validator: validator,
			config: map[string]interface{}{
				"uint32": uint32(42),
			},
			wantErr: false,
		},
		{
			name:      "uint64 value allowed",
			validator: validator,
			config: map[string]interface{}{
				"uint64": uint64(42),
			},
			wantErr: false,
		},
		{
			name:      "float32 value allowed",
			validator: validator,
			config: map[string]interface{}{
				"float32": float32(3.14),
			},
			wantErr: false,
		},
		{
			name:      "float64 value allowed",
			validator: validator,
			config: map[string]interface{}{
				"float64": float64(3.14),
			},
			wantErr: false,
		},
		{
			name:      "string value allowed",
			validator: validator,
			config: map[string]interface{}{
				"string": "hello world",
			},
			wantErr: false,
		},
		{
			name:      "[]byte value with valid length",
			validator: validator,
			config: map[string]interface{}{
				"bytes": []byte("normal byte array"),
			},
			wantErr: false,
		},
		{
			name:      "[]byte value exceeding MaxDataSize",
			validator: validator,
			config: map[string]interface{}{
				"large_bytes": bytes.Repeat([]byte("x"), validator.MaxDataSize+1),
			},
			wantErr: true,
			errCode: ErrCodeFieldSizeMismatch,
		},
		{
			name:      "unsupported type returns error in strict mode",
			validator: strictValidator,
			config: map[string]interface{}{
				"slice": []string{"a", "b"},
			},
			wantErr: true,
			errCode: ErrCodeInvalidConfig,
		},
		{
			name:      "unsupported type passes in non-strict mode",
			validator: nonStrictValidator,
			config: map[string]interface{}{
				"slice": []string{"a", "b"},
			},
			wantErr: false,
		},
		{
			name:      "map type returns error in strict mode",
			validator: strictValidator,
			config: map[string]interface{}{
				"nested_map": map[string]interface{}{},
			},
			wantErr: true,
			errCode: ErrCodeInvalidConfig,
		},
		{
			name:      "map type passes in non-strict mode",
			validator: nonStrictValidator,
			config: map[string]interface{}{
				"nested_map": map[string]interface{}{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator.ValidateConfig(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateConfig() error = nil, wantErr %v", tt.wantErr)
					return
				}
				extErr, ok := err.(*Error)
				if !ok {
					t.Errorf("ValidateConfig() error is not *Error type")
					return
				}
				if extErr.Code != tt.errCode {
					t.Errorf("ValidateConfig() error code = %d, want %d", extErr.Code, tt.errCode)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

// TestSimpleLogger_AllLevels tests all log levels
func TestSimpleLogger_AllLevels(t *testing.T) {
	logger := NewSimpleLogger("test")

	// Test that each level can be called without panicking
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warning message")
	logger.Error("error message", nil)
	logger.Fatal("fatal message")
}

// TestSimpleLogger_SetLevel tests setting log level
func TestSimpleLogger_SetLevel(t *testing.T) {
	tests := []struct {
		name       string
		setLevel   int
		wantLevel  int
		expectFail bool
	}{
		{
			name:      "set valid level - Debug",
			setLevel:  0,
			wantLevel: 0,
		},
		{
			name:      "set valid level - Info",
			setLevel:  1,
			wantLevel: 1,
		},
		{
			name:      "set valid level - Warn",
			setLevel:  2,
			wantLevel: 2,
		},
		{
			name:      "set valid level - Error",
			setLevel:  3,
			wantLevel: 3,
		},
		{
			name:      "set valid level - Fatal",
			setLevel:  4,
			wantLevel: 4,
		},
		{
			name:       "set invalid level - negative",
			setLevel:   -1,
			wantLevel:  1, // default is 1, will not change
			expectFail: true,
		},
		{
			name:       "set invalid level - greater than 4",
			setLevel:   5,
			wantLevel:  1, // default is 1, will not change
			expectFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewSimpleLogger("test")
			logger.SetLevel(tt.setLevel)

			if tt.expectFail {
				// Invalid values should not change the original level (default is 1)
				if logger.level != 1 {
					t.Errorf("SetLevel(%d) changed level to %d, expected 1 (unchanged)", tt.setLevel, logger.level)
				}
			} else {
				if logger.level != tt.wantLevel {
					t.Errorf("SetLevel(%d) = %d, want %d", tt.setLevel, logger.level, tt.wantLevel)
				}
			}
		})
	}
}

// TestSimpleLogger_LevelFiltering tests log level filtering
func TestSimpleLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		logFunc  func(*SimpleLogger)
		expected string // expected output level prefix
	}{
		{
			name:  "Debug level prints Debug log",
			level: 0,
			logFunc: func(l *SimpleLogger) {
				l.Debug("test debug")
			},
			expected: "DEBUG",
		},
		{
			name:  "Debug level prints Info log",
			level: 0,
			logFunc: func(l *SimpleLogger) {
				l.Info("test info")
			},
			expected: "INFO",
		},
		{
			name:  "Info level does not print Debug log",
			level: 1,
			logFunc: func(l *SimpleLogger) {
				l.Debug("test debug")
			},
			expected: "", // no output
		},
		{
			name:  "Info level prints Info log",
			level: 1,
			logFunc: func(l *SimpleLogger) {
				l.Info("test info")
			},
			expected: "INFO",
		},
		{
			name:  "Warn level does not print Info log",
			level: 2,
			logFunc: func(l *SimpleLogger) {
				l.Info("test info")
			},
			expected: "", // no output
		},
		{
			name:  "Warn level prints Warn log",
			level: 2,
			logFunc: func(l *SimpleLogger) {
				l.Warn("test warn")
			},
			expected: "WARN",
		},
		{
			name:  "Error level does not print Warn log",
			level: 3,
			logFunc: func(l *SimpleLogger) {
				l.Warn("test warn")
			},
			expected: "", // no output
		},
		{
			name:  "Error level prints Error log",
			level: 3,
			logFunc: func(l *SimpleLogger) {
				l.Error("test error", nil)
			},
			expected: "ERROR",
		},
		{
			name:  "Fatal level prints Fatal log",
			level: 4,
			logFunc: func(l *SimpleLogger) {
				l.Fatal("test fatal")
			},
			expected: "FATAL",
		},
		{
			name:  "Fatal level does not print Error log",
			level: 4,
			logFunc: func(l *SimpleLogger) {
				l.Error("test error", nil)
			},
			expected: "", // no output
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewSimpleLogger("test")
			logger.SetLevel(tt.level)

			// Just verify calls don't panic
			// Actual output can't be captured since it prints directly to stdout
			tt.logFunc(logger)
		})
	}
}

// TestInputSanitizer_SanitizeString tests string sanitization
func TestInputSanitizer_SanitizeString(t *testing.T) {
	sanitizer := NewInputSanitizer()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errCode ErrorCode
	}{
		{
			name:    "normal string",
			input:   "Hello World",
			want:    "Hello World",
			wantErr: false,
		},
		{
			name:    "string with newline",
			input:   "Line1\nLine2",
			want:    "Line1\nLine2",
			wantErr: false,
		},
		{
			name:    "string with carriage return",
			input:   "Line1\rLine2",
			want:    "Line1\rLine2",
			wantErr: false,
		},
		{
			name:    "string with tab",
			input:   "Col1\tCol2",
			want:    "Col1\tCol2",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "string too long",
			input:   strings.Repeat("a", 1025),
			want:    "",
			wantErr: true,
			errCode: ErrCodeFieldSizeMismatch,
		},
		{
			name:    "string at max length",
			input:   strings.Repeat("b", 1024),
			want:    strings.Repeat("b", 1024),
			wantErr: false,
		},
		{
			name:    "contains illegal control char - NUL",
			input:   "Hello\x00World",
			want:    "",
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "contains illegal control char - BEL",
			input:   "Hello\x07World",
			want:    "",
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "contains illegal control char - ESC",
			input:   "Hello\x1bWorld",
			want:    "",
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "Unicode chars OK",
			input:   "Hello 世界 🌍",
			want:    "Hello 世界 🌍",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitizer.SanitizeString(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("SanitizeString() error = nil, wantErr %v", tt.wantErr)
					return
				}
				extErr, ok := err.(*Error)
				if !ok {
					t.Errorf("SanitizeString() error is not *Error type")
					return
				}
				if extErr.Code != tt.errCode {
					t.Errorf("SanitizeString() error code = %d, want %d", extErr.Code, tt.errCode)
				}
			} else {
				if err != nil {
					t.Errorf("SanitizeString() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("SanitizeString() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// TestInputSanitizer_SanitizeBytes tests byte slice sanitization
func TestInputSanitizer_SanitizeBytes(t *testing.T) {
	sanitizer := NewInputSanitizer()

	tests := []struct {
		name    string
		input   []byte
		maxLen  int
		want    []byte
		wantErr bool
		errCode ErrorCode
	}{
		{
			name:    "normal byte slice",
			input:   []byte("Hello World"),
			maxLen:  100,
			want:    []byte("Hello World"),
			wantErr: false,
		},
		{
			name:    "empty byte slice",
			input:   []byte{},
			maxLen:  100,
			want:    []byte{},
			wantErr: false,
		},
		{
			name:    "nil byte slice",
			input:   nil,
			maxLen:  100,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "byte slice too long",
			input:   bytes.Repeat([]byte("a"), 101),
			maxLen:  100,
			want:    nil,
			wantErr: true,
			errCode: ErrCodeFieldSizeMismatch,
		},
		{
			name:    "byte slice at max length",
			input:   bytes.Repeat([]byte("b"), 100),
			maxLen:  100,
			want:    bytes.Repeat([]byte("b"), 100),
			wantErr: false,
		},
		{
			name:    "contains null byte",
			input:   []byte("Hello\x00World"),
			maxLen:  100,
			want:    []byte("Hello\x00World"),
			wantErr: false, // null bytes are allowed
		},
		{
			name:    "contains illegal control char - BEL",
			input:   []byte("Hello\x07World"),
			maxLen:  100,
			want:    nil,
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "contains illegal control char - ESC",
			input:   []byte("Hello\x1bWorld"),
			maxLen:  100,
			want:    nil,
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "contains illegal control char - at start",
			input:   []byte("\x01Hello"),
			maxLen:  100,
			want:    nil,
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "contains illegal control char - at end",
			input:   []byte("Hello\x1f"),
			maxLen:  100,
			want:    nil,
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitizer.SanitizeBytes(tt.input, tt.maxLen)
			if tt.wantErr {
				if err == nil {
					t.Errorf("SanitizeBytes() error = nil, wantErr %v", tt.wantErr)
					return
				}
				extErr, ok := err.(*Error)
				if !ok {
					t.Errorf("SanitizeBytes() error is not *Error type")
					return
				}
				if extErr.Code != tt.errCode {
					t.Errorf("SanitizeBytes() error code = %d, want %d", extErr.Code, tt.errCode)
				}
			} else {
				if err != nil {
					t.Errorf("SanitizeBytes() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !bytes.Equal(got, tt.want) {
					t.Errorf("SanitizeBytes() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// TestSafeParseInt tests safe integer parsing
func TestSafeParseInt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		base    int
		bitSize int
		wantErr bool
		errCode ErrorCode
	}{
		{
			name:    "normal number string",
			input:   "12345",
			base:    10,
			bitSize: 64,
			wantErr: false,
		},
		{
			name:    "negative number string",
			input:   "-12345",
			base:    10,
			bitSize: 64,
			wantErr: false,
		},
		{
			name:    "string with plus sign",
			input:   "+12345",
			base:    10,
			bitSize: 64,
			wantErr: false,
		},
		{
			name:    "zero",
			input:   "0",
			base:    10,
			bitSize: 64,
			wantErr: false,
		},
		{
			name:    "exactly 20 digits",
			input:   "12345678901234567890",
			base:    10,
			bitSize: 64,
			wantErr: false,
		},
		{
			name:    "more than 20 digits",
			input:   "123456789012345678901",
			base:    10,
			bitSize: 64,
			wantErr: true,
			errCode: ErrCodeInvalidFormat,
		},
		{
			name:    "contains illegal char - letter",
			input:   "123abc",
			base:    10,
			bitSize: 64,
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "contains illegal char - space",
			input:   "123 456",
			base:    10,
			bitSize: 64,
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "contains illegal char - symbol",
			input:   "123@456",
			base:    10,
			bitSize: 64,
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "empty string",
			input:   "",
			base:    10,
			bitSize: 64,
			wantErr: false, // empty string passes basic validation
		},
		{
			name:    "sign only - minus",
			input:   "-",
			base:    10,
			bitSize: 64,
			wantErr: false, // sign only passes basic validation
		},
		{
			name:    "sign only - plus",
			input:   "+",
			base:    10,
			bitSize: 64,
			wantErr: false, // sign only passes basic validation
		},
		{
			name:    "decimal point",
			input:   "123.456",
			base:    10,
			bitSize: 64,
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SafeParseInt(tt.input, tt.base, tt.bitSize)
			if tt.wantErr {
				if err == nil {
					t.Errorf("SafeParseInt() error = nil, wantErr %v", tt.wantErr)
					return
				}
				extErr, ok := err.(*Error)
				if !ok {
					t.Errorf("SafeParseInt() error is not *Error type")
					return
				}
				if extErr.Code != tt.errCode {
					t.Errorf("SafeParseInt() error code = %d, want %d", extErr.Code, tt.errCode)
				}
			} else {
				if err != nil {
					t.Errorf("SafeParseInt() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

// TestRecoveryManager_RegisterHandler tests registering handlers
func TestRecoveryManager_RegisterHandler(t *testing.T) {
	logger := NewSimpleLogger("test")

	tests := []struct {
		name    string
		handler ErrorHandler
		wantErr bool
		errCode ErrorCode
	}{
		{
			name:    "register valid handler",
			handler: NewPanicHandler(),
			wantErr: false,
		},
		{
			name:    "register nil handler returns error",
			handler: nil,
			wantErr: true,
			errCode: ErrCodeInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new RecoveryManager for each test
			testRm := NewRecoveryManager(logger)

			err := testRm.RegisterHandler(tt.handler)
			if tt.wantErr {
				if err == nil {
					t.Errorf("RegisterHandler() error = nil, wantErr %v", tt.wantErr)
					return
				}
				extErr, ok := err.(*Error)
				if !ok {
					t.Errorf("RegisterHandler() error is not *Error type")
					return
				}
				if extErr.Code != tt.errCode {
					t.Errorf("RegisterHandler() error code = %d, want %d", extErr.Code, tt.errCode)
				}
			} else {
				if err != nil {
					t.Errorf("RegisterHandler() error = %v, wantErr %v", err, tt.wantErr)
				}
				// Verify handler is registered
				if len(testRm.handlers) != 1 {
					t.Errorf("RegisterHandler() registered %d handlers, want 1", len(testRm.handlers))
				}
			}
		})
	}

	// Test registering multiple handlers
	t.Run("register multiple handlers", func(t *testing.T) {
		multiRm := NewRecoveryManager(logger)
		handler1 := NewPanicHandler()
		handler2 := NewPanicHandler()

		if err := multiRm.RegisterHandler(handler1); err != nil {
			t.Errorf("RegisterHandler() first handler error = %v", err)
		}
		if err := multiRm.RegisterHandler(handler2); err != nil {
			t.Errorf("RegisterHandler() second handler error = %v", err)
		}

		if len(multiRm.handlers) != 2 {
			t.Errorf("RegisterHandler() registered %d handlers, want 2", len(multiRm.handlers))
		}
	})
}

// TestRecoveryManager_Handle tests error handling
func TestRecoveryManager_Handle(t *testing.T) {
	logger := NewSimpleLogger("test")

	tests := []struct {
		name    string
		setup   func(*RecoveryManager)
		err     error
		wantErr bool
		checkFn func(error) bool
	}{
		{
			name:    "handle nil error",
			setup:   func(rm *RecoveryManager) {},
			err:     nil,
			wantErr: false,
		},
		{
			name:    "handle normal error",
			setup:   func(rm *RecoveryManager) {},
			err:     NewError(ErrCodeSystemError, "test error"),
			wantErr: true,
			checkFn: func(err error) bool {
				extErr, ok := err.(*Error)
				return ok && extErr.Code == ErrCodeSystemError
			},
		},
		{
			name:    "handle standard library error",
			setup:   func(rm *RecoveryManager) {},
			err:     &testError{msg: "standard error"},
			wantErr: true,
			checkFn: func(err error) bool {
				extErr, ok := err.(*Error)
				return ok && extErr.Code == ErrCodeSystemError
			},
		},
		{
			name: "handle with handler",
			setup: func(rm *RecoveryManager) {
				// Register a handler that can handle all errors
				rm.RegisterHandler(&testErrorHandler{
					name:    "test_handler",
					handles: true,
				})
			},
			err:     NewError(ErrCodeInvalidInput, "test error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := NewRecoveryManager(logger)
			tt.setup(rm)

			err := rm.Handle(tt.err)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Handle() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.checkFn != nil && !tt.checkFn(err) {
					t.Errorf("Handle() error check failed")
				}
			} else {
				if err != nil {
					t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

// Standard error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// Error handler for testing
type testErrorHandler struct {
	name    string
	handles bool
}

func (h *testErrorHandler) Handle(err *Error) error {
	return err
}

func (h *testErrorHandler) CanHandle(err *Error) bool {
	return h.handles
}

func (h *testErrorHandler) GetName() string {
	return h.name
}

// TestRecoveryManager_IsRecoverable tests error recoverability
func TestRecoveryManager_IsRecoverable(t *testing.T) {
	logger := NewSimpleLogger("test")
	rm := NewRecoveryManager(logger)

	tests := []struct {
		name      string
		err       error
		want      bool
		wantPanic bool
	}{
		{
			name: "SeverityInfo is recoverable",
			err: &Error{
				Code:     ErrCodeInvalidInput,
				Message:  "info",
				Severity: SeverityInfo,
			},
			want: true,
		},
		{
			name: "SeverityWarning is recoverable",
			err: &Error{
				Code:     ErrCodeInvalidInput,
				Message:  "warning",
				Severity: SeverityWarning,
			},
			want: true,
		},
		{
			name: "SeverityError is recoverable",
			err: &Error{
				Code:     ErrCodeInvalidInput,
				Message:  "error",
				Severity: SeverityError,
			},
			want: true,
		},
		{
			name: "SeverityCritical is not recoverable",
			err: &Error{
				Code:     ErrCodeSystemError,
				Message:  "critical",
				Severity: SeverityCritical,
			},
			want: false,
		},
		{
			name: "SeverityFatal is not recoverable",
			err: &Error{
				Code:     ErrCodeSystemError,
				Message:  "fatal",
				Severity: SeverityFatal,
			},
			want: false,
		},
		{
			name: "non-*Error type defaults to recoverable",
			err: &testError{
				msg: "standard error",
			},
			want: true,
		},
		{
			name: "nil error defaults to recoverable",
			err:  nil,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rm.IsRecoverable(tt.err)
			if got != tt.want {
				t.Errorf("IsRecoverable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestErrorSeverity_String tests error severity string output
func TestErrorSeverity_String(t *testing.T) {
	tests := []struct {
		name string
		es   ErrorSeverity
		want string
	}{
		{
			name: "SeverityInfo",
			es:   SeverityInfo,
			want: "INFO",
		},
		{
			name: "SeverityWarning",
			es:   SeverityWarning,
			want: "WARNING",
		},
		{
			name: "SeverityError",
			es:   SeverityError,
			want: "ERROR",
		},
		{
			name: "SeverityCritical",
			es:   SeverityCritical,
			want: "CRITICAL",
		},
		{
			name: "SeverityFatal",
			es:   SeverityFatal,
			want: "FATAL",
		},
		{
			name: "unknown severity level",
			es:   ErrorSeverity(999),
			want: "UNKNOWN",
		},
		{
			name: "negative severity level",
			es:   ErrorSeverity(-1),
			want: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.es.String()
			if got != tt.want {
				t.Errorf("ErrorSeverity.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
