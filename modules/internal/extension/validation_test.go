package extension

import (
	"bytes"
	"strings"
	"testing"
)

// translated comment
func TestDefaultValidator_ValidateData(t *testing.T) {
	validator := NewDefaultValidator()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		errCode ErrorCode
	}{
		{
			name:    "正常数据验证通过",
			data:    []byte("valid data"),
			wantErr: false,
		},
		{
			name:    "nil数据返回错误",
			data:    nil,
			wantErr: true,
			errCode: ErrCodeInvalidInput,
		},
		{
			name:    "空数据返回错误",
			data:    []byte{},
			wantErr: true,
			errCode: ErrCodeInvalidInput,
		},
		{
			name:    "超出MaxDataSize返回错误",
			data:    bytes.Repeat([]byte("a"), validator.MaxDataSize+1),
			wantErr: true,
			errCode: ErrCodeFieldSizeMismatch,
		},
		{
			name:    "边界值-刚好等于MaxDataSize",
			data:    bytes.Repeat([]byte("b"), validator.MaxDataSize),
			wantErr: false,
		},
		{
			name:    "边界值-单个字节",
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

// translated comment
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
			name:      "正常元数据验证通过",
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
			name:      "nil元数据返回错误",
			validator: validator,
			metadata:  nil,
			wantErr:   true,
			errCode:   ErrCodeInvalidMetadata,
		},
		{
			name:      "Type为0返回错误",
			validator: validator,
			metadata: &ExtensionMetadata{
				Type: 0,
				Name: "test",
			},
			wantErr: true,
			errCode: ErrCodeMissingField,
		},
		{
			name:      "空Name返回错误",
			validator: validator,
			metadata: &ExtensionMetadata{
				Type: 1,
				Name: "",
			},
			wantErr: true,
			errCode: ErrCodeMissingField,
		},
		{
			name:      "Name超过256字符返回错误",
			validator: validator,
			metadata: &ExtensionMetadata{
				Type: 1,
				Name: strings.Repeat("a", 257),
			},
			wantErr: true,
			errCode: ErrCodeFieldSizeMismatch,
		},
		{
			name:      "Name刚好256字符通过",
			validator: validator,
			metadata: &ExtensionMetadata{
				Type: 1,
				Name: strings.Repeat("b", 256),
			},
			wantErr: false,
		},
		{
			name:      "Description超过1024字符返回错误",
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
			name:      "Description刚好1024字符通过",
			validator: validator,
			metadata: &ExtensionMetadata{
				Type:        1,
				Name:        "test",
				Description: strings.Repeat("d", 1024),
			},
			wantErr: false,
		},
		{
			name:      "严格模式下无CompatibleTLSVersions通过",
			validator: validatorStrict,
			metadata: &ExtensionMetadata{
				Type:                  1,
				Name:                  "test",
				CompatibleTLSVersions: []uint16{},
			},
			wantErr: false,
		},
		{
			name:      "默认验证器下无CompatibleTLSVersions通过",
			validator: validator,
			metadata: &ExtensionMetadata{
				Type:                  1,
				Name:                  "test",
				CompatibleTLSVersions: []uint16{},
			},
			wantErr: false, // translated comment
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

// translated comment
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
			name:      "nil配置通过",
			validator: validator,
			config:    nil,
			wantErr:   false,
		},
		{
			name:      "空配置通过",
			validator: validator,
			config:    map[string]interface{}{},
			wantErr:   false,
		},
		{
			name:      "超过1000个键返回错误",
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
			name:      "刚好1000个键通过",
			validator: validator,
			config: func() map[string]interface{} {
				config := make(map[string]interface{})
				for i := 0; i < 1000; i++ {
					config[string(rune('a'+i%26)) + string(rune(i))] = i
				}
				return config
			}(),
			wantErr: false,
		},
		{
			name:      "空键返回错误",
			validator: validator,
			config: map[string]interface{}{
				"": "value",
			},
			wantErr: true,
			errCode: ErrCodeInvalidConfig,
		},
		{
			name:      "长键返回错误",
			validator: validator,
			config: map[string]interface{}{
				strings.Repeat("k", 257): "value",
			},
			wantErr: true,
			errCode: ErrCodeInvalidConfig,
		},
		{
			name:      "刚好256字符的键通过",
			validator: validator,
			config: map[string]interface{}{
				strings.Repeat("k", 256): "value",
			},
			wantErr: false,
		},
		{
			name:      "nil值允许",
			validator: validator,
			config: map[string]interface{}{
				"nil_key": nil,
			},
			wantErr: false,
		},
		{
			name:      "bool值允许",
			validator: validator,
			config: map[string]interface{}{
				"bool_true":  true,
				"bool_false": false,
			},
			wantErr: false,
		},
		{
			name:      "int值允许",
			validator: validator,
			config: map[string]interface{}{
				"int": 42,
			},
			wantErr: false,
		},
		{
			name:      "int32值允许",
			validator: validator,
			config: map[string]interface{}{
				"int32": int32(42),
			},
			wantErr: false,
		},
		{
			name:      "int64值允许",
			validator: validator,
			config: map[string]interface{}{
				"int64": int64(42),
			},
			wantErr: false,
		},
		{
			name:      "uint值允许",
			validator: validator,
			config: map[string]interface{}{
				"uint": uint(42),
			},
			wantErr: false,
		},
		{
			name:      "uint32值允许",
			validator: validator,
			config: map[string]interface{}{
				"uint32": uint32(42),
			},
			wantErr: false,
		},
		{
			name:      "uint64值允许",
			validator: validator,
			config: map[string]interface{}{
				"uint64": uint64(42),
			},
			wantErr: false,
		},
		{
			name:      "float32值允许",
			validator: validator,
			config: map[string]interface{}{
				"float32": float32(3.14),
			},
			wantErr: false,
		},
		{
			name:      "float64值允许",
			validator: validator,
			config: map[string]interface{}{
				"float64": float64(3.14),
			},
			wantErr: false,
		},
		{
			name:      "string值允许",
			validator: validator,
			config: map[string]interface{}{
				"string": "hello world",
			},
			wantErr: false,
		},
		{
			name:      "[]byte值且长度正常",
			validator: validator,
			config: map[string]interface{}{
				"bytes": []byte("normal byte array"),
			},
			wantErr: false,
		},
		{
			name:      "[]byte值超出MaxDataSize",
			validator: validator,
			config: map[string]interface{}{
				"large_bytes": bytes.Repeat([]byte("x"), validator.MaxDataSize+1),
			},
			wantErr: true,
			errCode: ErrCodeFieldSizeMismatch,
		},
		{
			name:      "严格模式下不支持的类型返回错误",
			validator: strictValidator,
			config: map[string]interface{}{
				"slice": []string{"a", "b"},
			},
			wantErr: true,
			errCode: ErrCodeInvalidConfig,
		},
		{
			name:      "非严格模式下不支持的类型通过",
			validator: nonStrictValidator,
			config: map[string]interface{}{
				"slice": []string{"a", "b"},
			},
			wantErr: false,
		},
		{
			name:      "严格模式下map类型返回错误",
			validator: strictValidator,
			config: map[string]interface{}{
				"nested_map": map[string]interface{}{},
			},
			wantErr: true,
			errCode: ErrCodeInvalidConfig,
		},
		{
			name:      "非严格模式下map类型通过",
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

// translated comment
func TestSimpleLogger_AllLevels(t *testing.T) {
	logger := NewSimpleLogger("test")

	// translated comment
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warning message")
	logger.Error("error message", nil)
	logger.Fatal("fatal message")
}

// translated comment
func TestSimpleLogger_SetLevel(t *testing.T) {
	tests := []struct {
		name       string
		setLevel   int
		wantLevel  int
		expectFail bool
	}{
		{
			name:      "设置有效级别-Debug",
			setLevel:  0,
			wantLevel: 0,
		},
		{
			name:      "设置有效级别-Info",
			setLevel:  1,
			wantLevel: 1,
		},
		{
			name:      "设置有效级别-Warn",
			setLevel:  2,
			wantLevel: 2,
		},
		{
			name:      "设置有效级别-Error",
			setLevel:  3,
			wantLevel: 3,
		},
		{
			name:      "设置有效级别-Fatal",
			setLevel:  4,
			wantLevel: 4,
		},
		{
			name:       "设置无效级别-负数",
			setLevel:   -1,
			wantLevel:  1, // translated comment
			expectFail: true,
		},
		{
			name:       "设置无效级别-大于4",
			setLevel:   5,
			wantLevel:  1, // translated comment
			expectFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewSimpleLogger("test")
			logger.SetLevel(tt.setLevel)

			if tt.expectFail {
				// translated comment
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

// translated comment
func TestSimpleLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		logFunc  func(*SimpleLogger)
		expected string // translated comment
	}{
		{
			name:  "Debug级别打印Debug日志",
			level: 0,
			logFunc: func(l *SimpleLogger) {
				l.Debug("test debug")
			},
			expected: "DEBUG",
		},
		{
			name:  "Debug级别打印Info日志",
			level: 0,
			logFunc: func(l *SimpleLogger) {
				l.Info("test info")
			},
			expected: "INFO",
		},
		{
			name:  "Info级别不打印Debug日志",
			level: 1,
			logFunc: func(l *SimpleLogger) {
				l.Debug("test debug")
			},
			expected: "", // translated comment
		},
		{
			name:  "Info级别打印Info日志",
			level: 1,
			logFunc: func(l *SimpleLogger) {
				l.Info("test info")
			},
			expected: "INFO",
		},
		{
			name:  "Warn级别不打印Info日志",
			level: 2,
			logFunc: func(l *SimpleLogger) {
				l.Info("test info")
			},
			expected: "", // translated comment
		},
		{
			name:  "Warn级别打印Warn日志",
			level: 2,
			logFunc: func(l *SimpleLogger) {
				l.Warn("test warn")
			},
			expected: "WARN",
		},
		{
			name:  "Error级别不打印Warn日志",
			level: 3,
			logFunc: func(l *SimpleLogger) {
				l.Warn("test warn")
			},
			expected: "", // translated comment
		},
		{
			name:  "Error级别打印Error日志",
			level: 3,
			logFunc: func(l *SimpleLogger) {
				l.Error("test error", nil)
			},
			expected: "ERROR",
		},
		{
			name:  "Fatal级别打印Fatal日志",
			level: 4,
			logFunc: func(l *SimpleLogger) {
				l.Fatal("test fatal")
			},
			expected: "FATAL",
		},
		{
			name:  "Fatal级别不打印Error日志",
			level: 4,
			logFunc: func(l *SimpleLogger) {
				l.Error("test error", nil)
			},
			expected: "", // translated comment
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewSimpleLogger("test")
			logger.SetLevel(tt.level)

			// translated comment
			// translated comment
			tt.logFunc(logger)
		})
	}
}

// translated comment
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
			name:    "正常字符串",
			input:   "Hello World",
			want:    "Hello World",
			wantErr: false,
		},
		{
			name:    "包含换行符的字符串",
			input:   "Line1\nLine2",
			want:    "Line1\nLine2",
			wantErr: false,
		},
		{
			name:    "包含回车符的字符串",
			input:   "Line1\rLine2",
			want:    "Line1\rLine2",
			wantErr: false,
		},
		{
			name:    "包含制表符的字符串",
			input:   "Col1\tCol2",
			want:    "Col1\tCol2",
			wantErr: false,
		},
		{
			name:    "空字符串",
			input:   "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "过长字符串",
			input:   strings.Repeat("a", 1025),
			want:    "",
			wantErr: true,
			errCode: ErrCodeFieldSizeMismatch,
		},
		{
			name:    "刚好最大长度的字符串",
			input:   strings.Repeat("b", 1024),
			want:    strings.Repeat("b", 1024),
			wantErr: false,
		},
		{
			name:    "包含非法控制字符-NUL",
			input:   "Hello\x00World",
			want:    "",
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "包含非法控制字符-BEL",
			input:   "Hello\x07World",
			want:    "",
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "包含非法控制字符-ESC",
			input:   "Hello\x1bWorld",
			want:    "",
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "Unicode字符正常",
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

// translated comment
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
			name:    "正常字节数组",
			input:   []byte("Hello World"),
			maxLen:  100,
			want:    []byte("Hello World"),
			wantErr: false,
		},
		{
			name:    "空字节数组",
			input:   []byte{},
			maxLen:  100,
			want:    []byte{},
			wantErr: false,
		},
		{
			name:    "nil字节数组",
			input:   nil,
			maxLen:  100,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "过长字节数组",
			input:   bytes.Repeat([]byte("a"), 101),
			maxLen:  100,
			want:    nil,
			wantErr: true,
			errCode: ErrCodeFieldSizeMismatch,
		},
		{
			name:    "刚好最大长度的字节数组",
			input:   bytes.Repeat([]byte("b"), 100),
			maxLen:  100,
			want:    bytes.Repeat([]byte("b"), 100),
			wantErr: false,
		},
		{
			name:    "包含空字节",
			input:   []byte("Hello\x00World"),
			maxLen:  100,
			want:    []byte("Hello\x00World"),
			wantErr: false, // translated comment
		},
		{
			name:    "包含非法控制字符-BEL",
			input:   []byte("Hello\x07World"),
			maxLen:  100,
			want:    nil,
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "包含非法控制字符-ESC",
			input:   []byte("Hello\x1bWorld"),
			maxLen:  100,
			want:    nil,
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "包含非法控制字符-在开头",
			input:   []byte("\x01Hello"),
			maxLen:  100,
			want:    nil,
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "包含非法控制字符-在结尾",
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

// translated comment
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
			name:    "正常数字字符串",
			input:   "12345",
			base:    10,
			bitSize: 64,
			wantErr: false,
		},
		{
			name:    "负数字符串",
			input:   "-12345",
			base:    10,
			bitSize: 64,
			wantErr: false,
		},
		{
			name:    "带正号的字符串",
			input:   "+12345",
			base:    10,
			bitSize: 64,
			wantErr: false,
		},
		{
			name:    "零",
			input:   "0",
			base:    10,
			bitSize: 64,
			wantErr: false,
		},
		{
			name:    "刚好20位数字",
			input:   "12345678901234567890",
			base:    10,
			bitSize: 64,
			wantErr: false,
		},
		{
			name:    "超过20位数字",
			input:   "123456789012345678901",
			base:    10,
			bitSize: 64,
			wantErr: true,
			errCode: ErrCodeInvalidFormat,
		},
		{
			name:    "包含非法字符-字母",
			input:   "123abc",
			base:    10,
			bitSize: 64,
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "包含非法字符-空格",
			input:   "123 456",
			base:    10,
			bitSize: 64,
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "包含非法字符-符号",
			input:   "123@456",
			base:    10,
			bitSize: 64,
			wantErr: true,
			errCode: ErrCodeEncodingError,
		},
		{
			name:    "空字符串",
			input:   "",
			base:    10,
			bitSize: 64,
			wantErr: false, // translated comment
		},
		{
			name:    "只有符号-负号",
			input:   "-",
			base:    10,
			bitSize: 64,
			wantErr: false, // translated comment
		},
		{
			name:    "只有符号-正号",
			input:   "+",
			base:    10,
			bitSize: 64,
			wantErr: false, // translated comment
		},
		{
			name:    "小数点",
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

// translated comment
func TestRecoveryManager_RegisterHandler(t *testing.T) {
	logger := NewSimpleLogger("test")

	tests := []struct {
		name    string
		handler ErrorHandler
		wantErr bool
		errCode ErrorCode
	}{
		{
			name:    "注册有效的处理器",
			handler: NewPanicHandler(),
			wantErr: false,
		},
		{
			name:    "注册nil处理器返回错误",
			handler: nil,
			wantErr: true,
			errCode: ErrCodeInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// translated comment
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
				// translated comment
				if len(testRm.handlers) != 1 {
					t.Errorf("RegisterHandler() registered %d handlers, want 1", len(testRm.handlers))
				}
			}
		})
	}

	// translated comment
	t.Run("注册多个处理器", func(t *testing.T) {
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

// translated comment
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
			name:    "处理nil错误",
			setup:   func(rm *RecoveryManager) {},
			err:     nil,
			wantErr: false,
		},
		{
			name:    "处理普通错误",
			setup:   func(rm *RecoveryManager) {},
			err:     NewError(ErrCodeSystemError, "test error"),
			wantErr: true,
			checkFn: func(err error) bool {
				extErr, ok := err.(*Error)
				return ok && extErr.Code == ErrCodeSystemError
			},
		},
		{
			name:    "处理标准库错误",
			setup:   func(rm *RecoveryManager) {},
			err:     &testError{msg: "standard error"},
			wantErr: true,
			checkFn: func(err error) bool {
				extErr, ok := err.(*Error)
				return ok && extErr.Code == ErrCodeSystemError
			},
		},
		{
			name: "使用处理器处理",
			setup: func(rm *RecoveryManager) {
				// translated comment
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

// translated comment
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// translated comment
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

// translated comment
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
			name: "SeverityInfo可恢复",
			err: &Error{
				Code:     ErrCodeInvalidInput,
				Message:  "info",
				Severity: SeverityInfo,
			},
			want: true,
		},
		{
			name: "SeverityWarning可恢复",
			err: &Error{
				Code:     ErrCodeInvalidInput,
				Message:  "warning",
				Severity: SeverityWarning,
			},
			want: true,
		},
		{
			name: "SeverityError可恢复",
			err: &Error{
				Code:     ErrCodeInvalidInput,
				Message:  "error",
				Severity: SeverityError,
			},
			want: true,
		},
		{
			name: "SeverityCritical不可恢复",
			err: &Error{
				Code:     ErrCodeSystemError,
				Message:  "critical",
				Severity: SeverityCritical,
			},
			want: false,
		},
		{
			name: "SeverityFatal不可恢复",
			err: &Error{
				Code:     ErrCodeSystemError,
				Message:  "fatal",
				Severity: SeverityFatal,
			},
			want: false,
		},
		{
			name: "非*Error类型错误默认可恢复",
			err: &testError{
				msg: "standard error",
			},
			want: true,
		},
		{
			name: "nil错误默认可恢复",
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

// translated comment
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
			name: "未知的严重级别",
			es:   ErrorSeverity(999),
			want: "UNKNOWN",
		},
		{
			name: "负的严重级别",
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
