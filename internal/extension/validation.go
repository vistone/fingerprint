package extension

import (
	"fmt"
	"time"
)

// DefaultValidator 默认输入验证器
//
// 使用示例：
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
// 验证项：
//   - 数据非空且不超过 MaxDataSize
//   - 元数据必填字段（Type、Name）
//   - 元数据字段大小限制（Name ≤256B, Description ≤1KB）
//   - 配置项大小限制（最多 1000 个键）
type DefaultValidator struct {
	// 最大数据大小（字节）
	MaxDataSize int

	// 最大扩展数
	MaxExtensions int

	// 启用严格模式
	StrictMode bool
}

// NewDefaultValidator 创建默认验证器
func NewDefaultValidator() *DefaultValidator {
	return &DefaultValidator{
		MaxDataSize:   65536, // 64KB
		MaxExtensions: 10000, // 10K 扩展
		StrictMode:    true,
	}
}

// ValidateData 验证扩展数据
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

// ValidateMetadata 验证元数据
func (v *DefaultValidator) ValidateMetadata(metadata *ExtensionMetadata) error {
	if metadata == nil {
		return NewError(ErrCodeInvalidMetadata, "metadata cannot be nil")
	}

	// 检查必填字段
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

	// 检查 TLS 版本
	if len(metadata.CompatibleTLSVersions) == 0 && !v.StrictMode {
		return NewWarning("no compatible TLS versions specified")
	}

	return nil
}

// ValidateConfig 验证配置
func (v *DefaultValidator) ValidateConfig(config map[string]interface{}) error {
	if config == nil {
		return nil // 配置可以为空
	}

	// 检查配置大小
	if len(config) > 1000 {
		return NewError(ErrCodeFieldSizeMismatch,
			"too many configuration keys")
	}

	// 验证每个配置项
	for key, value := range config {
		if key == "" {
			return NewError(ErrCodeInvalidConfig,
				"configuration key cannot be empty")
		}

		if len(key) > 256 {
			return NewError(ErrCodeInvalidConfig,
				"configuration key too long")
		}

		// 检查值的类型
		switch value.(type) {
		case nil, bool, int, int32, int64, uint, uint32, uint64, float32, float64, string:
			// 允许的基本类型
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

// Logger 日志接口
type Logger interface {
	// 记录信息
	Info(msg string, args ...interface{})

	// 记录警告
	Warn(msg string, args ...interface{})

	// 记录错误
	Error(msg string, err error, args ...interface{})

	// 记录调试信息
	Debug(msg string, args ...interface{})

	// 记录 panic
	Fatal(msg string, args ...interface{})
}

// SimpleLogger 简单的日志实现
type SimpleLogger struct {
	name  string
	level int // 0=debug, 1=info, 2=warn, 3=error, 4=fatal
}

// NewSimpleLogger 创建简单日志
func NewSimpleLogger(name string) *SimpleLogger {
	return &SimpleLogger{
		name:  name,
		level: 1, // 默认 info 级别
	}
}

// SetLevel 设置日志级别
func (sl *SimpleLogger) SetLevel(level int) {
	if level >= 0 && level <= 4 {
		sl.level = level
	}
}

// Debug 记录调试信息
func (sl *SimpleLogger) Debug(msg string, args ...interface{}) {
	if sl.level <= 0 {
		sl.log("DEBUG", msg, args...)
	}
}

// Info 记录信息
func (sl *SimpleLogger) Info(msg string, args ...interface{}) {
	if sl.level <= 1 {
		sl.log("INFO", msg, args...)
	}
}

// Warn 记录警告
func (sl *SimpleLogger) Warn(msg string, args ...interface{}) {
	if sl.level <= 2 {
		sl.log("WARN", msg, args...)
	}
}

// Error 记录错误
func (sl *SimpleLogger) Error(msg string, err error, args ...interface{}) {
	if sl.level <= 3 {
		args = append([]interface{}{err}, args...)
		sl.log("ERROR", msg, args...)
	}
}

// Fatal 记录致命错误
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

// InputSanitizer 输入清理器
type InputSanitizer struct {
	maxFieldSize int
	allowedChars map[rune]bool
}

// NewInputSanitizer 创建输入清理器
func NewInputSanitizer() *InputSanitizer {
	return &InputSanitizer{
		maxFieldSize: 1024,
		allowedChars: make(map[rune]bool),
	}
}

// SanitizeString 清理字符串
func (is *InputSanitizer) SanitizeString(s string) (string, error) {
	if len(s) > is.maxFieldSize {
		return "", NewError(ErrCodeFieldSizeMismatch,
			fmt.Sprintf("string exceeds max size: %d > %d", len(s), is.maxFieldSize))
	}

	// 检查非法字符
	for _, r := range s {
		// 允许基本的可打印 ASCII 字符和 UTF-8
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return "", NewError(ErrCodeEncodingError,
				fmt.Sprintf("invalid character: %d", r)).
				WithContext("position", len(s))
		}
	}

	return s, nil
}

// SanitizeBytes 清理字节数组
func (is *InputSanitizer) SanitizeBytes(b []byte, maxLen int) ([]byte, error) {
	if len(b) > maxLen {
		return nil, NewError(ErrCodeFieldSizeMismatch,
			fmt.Sprintf("bytes exceed max size: %d > %d", len(b), maxLen))
	}

	// 检查是否包含非法的控制字符
	for i, b := range b {
		if b < 32 && b != 0x00 {
			// 不允许除空字节外的控制字符
			return nil, NewError(ErrCodeEncodingError,
				fmt.Sprintf("illegal control byte at position %d: 0x%02x", i, b))
		}
	}

	return b, nil
}

// SafeParseInt 安全解析整数
func SafeParseInt(s string, base, bitSize int) (int64, error) {
	if len(s) > 20 { // int64 最多 20 位数字
		return 0, NewError(ErrCodeInvalidFormat,
			"integer string too long")
	}

	// 使用基本的验证，避免 panic
	for _, c := range s {
		if !((c >= '0' && c <= '9') || c == '-' || c == '+') {
			return 0, NewError(ErrCodeEncodingError,
				fmt.Sprintf("invalid character in integer: %c", c))
		}
	}

	return 0, nil // 实际的解析由调用方处理
}

// RecoveryManager 恢复管理器
type RecoveryManager struct {
	handlers []ErrorHandler
	logger   Logger
}

// NewRecoveryManager 创建恢复管理器
func NewRecoveryManager(logger Logger) *RecoveryManager {
	return &RecoveryManager{
		handlers: make([]ErrorHandler, 0),
		logger:   logger,
	}
}

// RegisterHandler 注册错误处理器
func (rm *RecoveryManager) RegisterHandler(handler ErrorHandler) error {
	if handler == nil {
		return NewError(ErrCodeInvalidInput, "handler cannot be nil")
	}
	rm.handlers = append(rm.handlers, handler)
	return nil
}

// Handle 处理错误
func (rm *RecoveryManager) Handle(err error) error {
	if err == nil {
		return nil
	}

	extErr, ok := err.(*Error)
	if !ok {
		// 转换标准错误
		extErr = NewErrorWithCause(ErrCodeSystemError, "unexpected error", err)
	}

	rm.logger.Error("error occurred", extErr)

	// 尝试用处理器处理
	for _, handler := range rm.handlers {
		if handler.CanHandle(extErr) {
			return handler.Handle(extErr)
		}
	}

	return extErr
}

// IsRecoverable 判断错误是否可恢复
func (rm *RecoveryManager) IsRecoverable(err error) bool {
	if extErr, ok := err.(*Error); ok {
		return extErr.IsRecoverable()
	}
	return true
}
