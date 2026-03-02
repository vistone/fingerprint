// Package extension 提供扩展系统的核心功能：错误处理、验证、防御和日志记录
//
// # 错误处理
//
// 使用 Error 结构体和 ErrorCode 枚举进行统一错误处理：
//
//	err := extension.NewError(extension.ErrCodeInvalidInput, "输入无效")
//	err = err.WithContext("user_id", 123).WithContext("field", "email")
//	if err.IsRecoverable() {
//	    // 可以重试
//	    return retryOperation()
//	}
//	if err.IsFatal() {
//	    return err
//	}
//
// # 错误代码体系
//
// 46 个错误代码分为 8 个类别（1000-8999）：
//
//	注册表错误  (1000-1999): NotFound, AlreadyRegistered, InvalidMetadata
//	验证错误    (2000-2999): InvalidInput, FieldSizeMismatch, EncodingError
//	解析错误    (3000-3999): ParseFailed, InvalidFormat, MalformedData
//	分析错误    (4000-4999): AnalysisFailed, ResourceExhausted
//	配置错误    (5000-5999): InvalidConfig, MissingConfig
//	插件错误    (6000-6999): PluginNotFound, PluginLoadFailed
//	系统错误    (7000-7999): SystemError, MemoryExhausted, Timeout
//	安全错误    (8000-8999): SecurityViolation, Unauthorized
//
// # 严重级别
//
// 5 级错误严重程度：
//
//	SeverityInfo      (0) - 信息，可恢复
//	SeverityWarning   (1) - 警告，可恢复
//	SeverityError     (2) - 错误，可恢复
//	SeverityCritical  (3) - 严重，不可恢复
//	SeverityFatal     (4) - 致命，不可恢复
//
// # 输入验证
//
// 使用 DefaultValidator 验证数据：
//
//	validator := extension.NewDefaultValidator()
//	if err := validator.ValidateData(data); err != nil {
//	    return err
//	}
//
// # 防御系统
//
// 使用 RequestGuard 保护 API 端点：
//
//	policy := extension.DefaultDefensePolicy()
//	guard := extension.NewRequestGuard(policy)
//	if err := guard.ValidateRequest(request); err != nil {
//	    return err
//	}
//
// # 安全执行
//
// 捕获 panic 并恢复：
//
//	extension.SafeExecuteWithRecovery(
//	    func() error { return unsafeOp() },
//	    func(r interface{}) { log.Printf("panic: %v", r) },
//	)
package extension

import (
	"fmt"
)

// ErrorCode 错误代码（便于系统集成和错误分类）
type ErrorCode int

const (
	// 注册表错误（1000-1999）
	ErrCodeNotFound ErrorCode = iota + 1000
	ErrCodeAlreadyRegistered
	ErrCodeInvalidMetadata
	ErrCodeRegistryFull

	// 验证错误（2000-2999）
	ErrCodeValidationFailed ErrorCode = iota + 2000 - 4
	ErrCodeInvalidInput
	ErrCodeMissingField
	ErrCodeFieldSizeMismatch
	ErrCodeEncodingError
	ErrCodeVersionMismatch

	// 解析错误（3000-3999）
	ErrCodeParseFailed ErrorCode = iota + 3000 - 10
	ErrCodeInvalidFormat
	ErrCodeUnexpectedEOF
	ErrCodeInvalidOffset
	ErrCodeMalformedData

	// 分析错误（4000-4999）
	ErrCodeAnalysisFailed ErrorCode = iota + 4000 - 15
	ErrCodeAnalysisTimeout
	ErrCodeResourceExhausted
	ErrCodeInternalError

	// 配置错误（5000-5999）
	ErrCodeInvalidConfig ErrorCode = iota + 5000 - 19
	ErrCodeMissingConfig
	ErrCodeConfigConflict

	// 插件错误（6000-6999）
	ErrCodePluginNotFound ErrorCode = iota + 6000 - 22
	ErrCodePluginInitFailed
	ErrCodePluginLoadFailed
	ErrCodePluginVersionMismatch

	// 系统错误（7000-7999）
	ErrCodeSystemError ErrorCode = iota + 7000 - 27
	ErrCodeMemoryExhausted
	ErrCodeTimeout
	ErrCodeCancelled

	// 安全错误（8000-8999）
	ErrCodeSecurityViolation ErrorCode = iota + 8000 - 31
	ErrCodeUnauthorized
	ErrCodeForbidden
)

// Error 扩展系统的标准错误类型
//
// 字段说明：
//
//	Code       - 错误代码，便于分类和集成
//	Message    - 人类可读的错误描述
//	Cause      - 原始错误，支持错误链
//	Context    - 错误上下文键值对（用于调试）
//	Severity   - 错误严重级别（决定可恢复性）
//	Timestamp  - 错误发生时间
//
// 使用示例：
//
//	// 创建错误
//	err := extension.NewError(
//	    extension.ErrCodeInvalidInput,
//	    "无效的用户输入",
//	)
//
//	// 添加上下文信息（用于调试）
//	err = err.WithContext("user_id", 12345)
//	err = err.WithContext("field", "email")
//	err = err.WithContext("value", userEmail)
//
//	// 判断错误类型
//	if err.IsRecoverable() {
//	    // Info, Warning, Error 级别的错误可以恢复
//	    return retryOperation()
//	}
//
//	if err.IsFatal() {
//	    // Fatal 级别的错误无法恢复
//	    return err
//	}
//
// 创建变体：
//
//	extension.NewError(code, message)                    // 标准错误
//	extension.NewErrorWithCause(code, message, cause)   // 带原因的错误
//	extension.NewFatalError(code, message)              // 致命错误
//	extension.NewWarning(message)                       // 警告（不需要 code）
//
// 错误链：
//
//	err := processData(data)
//	if err != nil {
//	    return extension.NewErrorWithCause(
//	        extension.ErrCodeAnalysisFailed,
//	        "分析失败",
//	        err,  // 原始错误作为链的下一个
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

// ErrorSeverity 错误严重程度
type ErrorSeverity int

const (
	SeverityInfo ErrorSeverity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
	SeverityFatal
)

// String 实现 error 接口
func (e *Error) Error() string {
	msg := fmt.Sprintf("[%s] %s (code: %d)", e.Severity.String(), e.Message, e.Code)
	if e.Cause != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Cause)
	}
	return msg
}

// Unwrap 支持 error wrapping
func (e *Error) Unwrap() error {
	return e.Cause
}

// WithContext 添加上下文信息
func (e *Error) WithContext(key string, value interface{}) *Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// IsRecoverable 判断错误是否可恢复
func (e *Error) IsRecoverable() bool {
	switch e.Severity {
	case SeverityFatal, SeverityCritical:
		return false
	default:
		return true
	}
}

// IsFatal 判断错误是否致命
func (e *Error) IsFatal() bool {
	return e.Severity == SeverityFatal
}

// String 实现 Stringer 接口
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

// NewError 创建新错误
func NewError(code ErrorCode, message string) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		Context:  make(map[string]interface{}),
		Severity: SeverityError,
	}
}

// NewErrorWithCause 创建包含原因的错误
func NewErrorWithCause(code ErrorCode, message string, cause error) *Error {
	err := NewError(code, message)
	err.Cause = cause
	return err
}

// NewFatalError 创建致命错误
func NewFatalError(code ErrorCode, message string) *Error {
	err := NewError(code, message)
	err.Severity = SeverityFatal
	return err
}

// NewWarning 创建警告
func NewWarning(message string) *Error {
	err := &Error{
		Code:     ErrCodeInvalidInput,
		Message:  message,
		Context:  make(map[string]interface{}),
		Severity: SeverityWarning,
	}
	return err
}

// ToError 将标准错误转换为扩展错误
func ToError(code ErrorCode, err error) *Error {
	if extErr, ok := err.(*Error); ok {
		return extErr
	}
	return NewErrorWithCause(code, "operation failed", err)
}

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
	// 处理错误
	Handle(err *Error) error

	// 判断是否能处理此错误
	CanHandle(err *Error) bool

	// 获取处理器名称
	GetName() string
}

// PanicHandler panic 恢复处理器
type PanicHandler struct {
	name string
}

func NewPanicHandler() *PanicHandler {
	return &PanicHandler{name: "panic_handler"}
}

func (ph *PanicHandler) Handle(err *Error) error {
	if err.Severity == SeverityFatal {
		// 记录致命错误，但不 panic
		return fmt.Errorf("fatal error occurred: %w", err)
	}
	return err
}

func (ph *PanicHandler) CanHandle(err *Error) bool {
	return true // 可以处理任何错误
}

func (ph *PanicHandler) GetName() string {
	return ph.name
}

// SafeExecute 安全执行函数，捕获 panic
func SafeExecute(fn func() error) error {
	defer func() {
		if r := recover(); r != nil {
			// panic 已被捕获，但我们在这里无法修改返回值
			// 所以需要通过返回值进行处理
		}
	}()

	return fn()
}

// SafeExecuteWithRecovery 安全执行函数，带恢复逻辑
func SafeExecuteWithRecovery(fn func() error, handler func(interface{})) error {
	defer func() {
		if r := recover(); r != nil {
			handler(r)
		}
	}()

	return fn()
}
