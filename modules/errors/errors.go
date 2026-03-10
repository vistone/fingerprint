package errors

// NOTE: 本包与 core/errors.go 存在职责重叠（双重 ErrorCode / 双重哨兵错误）。
// 新代码应优先使用 core.ErrorCode + core.CoreError 体系。
// 本包保留为向后兼容，计划后续版本统一。

import (
	"errors"
	"fmt"
	"strings"
)

// ============================================================================
// 错误码定义
// ============================================================================

// ErrorCode 错误码类型
type ErrorCode string

const (
	// 系统级错误 (SYS)
	ErrCodeSystem     ErrorCode = "SYS001"
	ErrCodeNotFound   ErrorCode = "SYS002"
	ErrCodeInvalidArg ErrorCode = "SYS003"
	ErrCodeTimeout    ErrorCode = "SYS004"
	ErrCodeCancelled  ErrorCode = "SYS005"
	ErrCodeInternal   ErrorCode = "SYS006"
	ErrCodeNotImpl    ErrorCode = "SYS007"

	// Profile 相关错误 (PRF)
	ErrCodeProfileNotFound      ErrorCode = "PRF001"
	ErrCodeProfileInvalid       ErrorCode = "PRF002"
	ErrCodeProfileExists        ErrorCode = "PRF003"
	ErrCodeProfileLoadFailed    ErrorCode = "PRF004"
	ErrCodeProfileSaveFailed    ErrorCode = "PRF005"
	ErrCodeProfileNoDefault     ErrorCode = "PRF006"
	ErrCodeProfileRemoveDefault ErrorCode = "PRF007"

	// 配置相关错误 (CFG)
	ErrCodeConfigNotLoaded    ErrorCode = "CFG001"
	ErrCodeConfigValidation   ErrorCode = "CFG002"
	ErrCodeConfigPathNotFound ErrorCode = "CFG003"
	ErrCodeConfigExists       ErrorCode = "CFG004"
	ErrCodeConfigVersion      ErrorCode = "CFG005"

	// 网络相关错误 (NET)
	ErrCodeNetInvalidIP   ErrorCode = "NET001"
	ErrCodeNetInvalidPort ErrorCode = "NET002"
	ErrCodeNetConnect     ErrorCode = "NET003"
	ErrCodeNetTimeout     ErrorCode = "NET004"

	// 协议相关错误 (PRT)
	ErrCodeProtoTLSVersion  ErrorCode = "PRT001"
	ErrCodeProtoCipherSuite ErrorCode = "PRT002"
	ErrCodeProtoExtension   ErrorCode = "PRT003"
	ErrCodeProtoInvalidSpec ErrorCode = "PRT004"

	// 输入验证错误 (INP)
	ErrCodeInputInvalid  ErrorCode = "INP001"
	ErrCodeInputEmpty    ErrorCode = "INP002"
	ErrCodeInputTooLarge ErrorCode = "INP003"
	ErrCodeInputRequired ErrorCode = "INP004"

	// 缓存相关错误 (CCH)
	ErrCodeCacheNotFound ErrorCode = "CCH001"
	ErrCodeCacheExpired  ErrorCode = "CCH002"

	// ML 相关错误 (ML)
	ErrCodeMLModelNotFound ErrorCode = "ML001"
	ErrCodeMLPrediction    ErrorCode = "ML002"

	// 插件相关错误 (PLG)
	ErrCodePluginNotFound ErrorCode = "PLG001"
	ErrCodePluginInit     ErrorCode = "PLG002"
)

// ============================================================================
// 哨兵错误定义 - 核心错误
// ============================================================================

var (
	// 配置相关错误
	ErrProfileNotFound               = errors.New("profile not found")
	ErrInvalidFingerprint            = errors.New("invalid fingerprint format")
	ErrClientHelloSpecNotImplemented = errors.New("client hello spec not implemented")
	ErrInvalidUserAgent              = errors.New("invalid user agent")
	ErrNoProfilesAvailable           = errors.New("no profiles available")
	ErrUnsupportedBrowser            = errors.New("unsupported browser")

	// 配置中心错误
	ErrConfigNotLoaded     = errors.New("config not loaded")
	ErrConfigValidation    = errors.New("config validation failed")
	ErrConfigPathNotFound  = errors.New("config path not found")
	ErrConfigAlreadyExists = errors.New("config already exists")
	ErrVersionNotFound     = errors.New("version not found")

	// 网络相关错误
	ErrInvalidIP        = errors.New("invalid IP address")
	ErrInvalidPort      = errors.New("invalid port")
	ErrConnectionFailed = errors.New("connection failed")

	// 协议相关错误
	ErrInvalidTLSVersion  = errors.New("invalid TLS version")
	ErrInvalidCipherSuite = errors.New("invalid cipher suite")
	ErrInvalidExtension   = errors.New("invalid extension")

	// 输入验证错误
	ErrInvalidInput  = errors.New("invalid input")
	ErrEmptyInput    = errors.New("empty input")
	ErrInputTooLarge = errors.New("input too large")
	ErrRequiredField = errors.New("required field is missing")

	// 内部错误
	ErrInternal       = errors.New("internal error")
	ErrNotImplemented = errors.New("not implemented")
	ErrTimeout        = errors.New("operation timeout")
	ErrCancelled      = errors.New("operation cancelled")
)

// ============================================================================
// 错误类型分类
// ============================================================================

// ErrorCategory 错误分类
type ErrorCategory string

const (
	CategoryConfig   ErrorCategory = "config"
	CategoryNetwork  ErrorCategory = "network"
	CategoryProtocol ErrorCategory = "protocol"
	CategoryInput    ErrorCategory = "input"
	CategoryInternal ErrorCategory = "internal"
	CategoryNotFound ErrorCategory = "not_found"
)

// CategorizedError 带分类的错误
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

// NewCategorizedError 创建带分类的错误
func NewCategorizedError(category ErrorCategory, err error, details string) *CategorizedError {
	return &CategorizedError{
		Category: category,
		Err:      err,
		Details:  details,
	}
}

// ============================================================================
// 错误包装辅助函数
// ============================================================================

// Wrap 包装错误并添加上下文
func Wrap(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// Wrapf 使用格式化字符串包装错误
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	context := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", context, err)
}

// New 创建新错误（兼容标准库）
func New(text string) error {
	return errors.New(text)
}

// Newf 使用格式化字符串创建错误
func Newf(format string, args ...interface{}) error {
	return errors.New(fmt.Sprintf(format, args...))
}

// Is 检查错误是否匹配（兼容标准库）
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As 类型断言（兼容标准库）
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// ============================================================================
// 错误检查辅助函数
// ============================================================================

const clientHelloSpecNotImplementedMsg = "please implement this method"

// IsClientHelloSpecNotImplemented 判断错误是否表示 profile 未实现 ClientHelloSpec。
func IsClientHelloSpecNotImplemented(err error) bool {
	if err == nil {
		return false
	}
	// 优先检查哨兵错误
	if errors.Is(err, ErrClientHelloSpecNotImplemented) {
		return true
	}
	// 向后兼容：检查错误消息字符串
	return strings.Contains(strings.ToLower(err.Error()), clientHelloSpecNotImplementedMsg)
}

// IsNotFound 检查是否为"未找到"类错误
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrProfileNotFound) ||
		errors.Is(err, ErrConfigPathNotFound) ||
		errors.Is(err, ErrVersionNotFound)
}

// IsInvalidInput 检查是否为"无效输入"类错误
func IsInvalidInput(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidInput) ||
		errors.Is(err, ErrInvalidFingerprint) ||
		errors.Is(err, ErrInvalidUserAgent)
}

// IsConfigError 检查是否为配置类错误
func IsConfigError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrConfigNotLoaded) ||
		errors.Is(err, ErrConfigValidation) ||
		errors.Is(err, ErrConfigPathNotFound)
}

// ============================================================================
// 特定领域的错误创建函数
// ============================================================================

// NewConfigError 创建配置错误
func NewConfigError(op string, err error) error {
	return Wrapf(err, "config operation '%s' failed", op)
}

// NewValidationError 创建验证错误
func NewValidationError(field string, reason string) error {
	return NewCategorizedError(CategoryInput, ErrInvalidInput,
		fmt.Sprintf("field '%s': %s", field, reason))
}

// NewNotFoundError 创建未找到错误
func NewNotFoundError(resource string, identifier string) error {
	return NewCategorizedError(CategoryNotFound, ErrProfileNotFound,
		fmt.Sprintf("%s '%s' not found", resource, identifier))
}

// NewProtocolError 创建协议错误
func NewProtocolError(protocol string, err error) error {
	return NewCategorizedError(CategoryProtocol, err,
		fmt.Sprintf("protocol '%s' error", protocol))
}

// NewNetworkError 创建网络错误
func NewNetworkError(op string, err error) error {
	return NewCategorizedError(CategoryNetwork, err,
		fmt.Sprintf("network operation '%s' failed", op))
}

// ============================================================================
// 带错误码的错误类型
// ============================================================================

// CodeError 带错误码的错误
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

// WithDetail 添加详细信息
func (e *CodeError) WithDetail(key string, value any) *CodeError {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}

// NewError 创建带错误码的错误
func NewError(code ErrorCode, message string) *CodeError {
	return &CodeError{
		Code:    code,
		Message: message,
		Details: make(map[string]any),
	}
}

// NewErrorWithCause 创建带错误码和原因的错误
func NewErrorWithCause(code ErrorCode, message string, cause error) *CodeError {
	return &CodeError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Details: make(map[string]any),
	}
}

// ============================================================================
// 便捷错误创建函数
// ============================================================================

// ProfileNotFound 创建 Profile 不存在错误
func ProfileNotFound(id string) *CodeError {
	return NewError(ErrCodeProfileNotFound, "profile not found").
		WithDetail("profile_id", id)
}

// ProfileInvalid 创建 Profile 无效错误
func ProfileInvalid(id string, reason string) *CodeError {
	return NewError(ErrCodeProfileInvalid, "profile is invalid").
		WithDetail("profile_id", id).
		WithDetail("reason", reason)
}

// ProfileRemoveDefault 创建不能删除默认 Profile 错误
func ProfileRemoveDefault(id string) *CodeError {
	return NewError(ErrCodeProfileRemoveDefault, "cannot remove default profile").
		WithDetail("profile_id", id)
}

// ConfigNotLoaded 创建配置未加载错误
func ConfigNotLoaded(path string) *CodeError {
	return NewError(ErrCodeConfigNotLoaded, "config not loaded").
		WithDetail("path", path)
}

// InvalidInput 创建输入无效错误
func InvalidInput(field string, reason string) *CodeError {
	return NewError(ErrCodeInputInvalid, "invalid input").
		WithDetail("field", field).
		WithDetail("reason", reason)
}

// RequiredField 创建必填字段缺失错误
func RequiredField(field string) *CodeError {
	return NewError(ErrCodeInputRequired, "required field is missing").
		WithDetail("field", field)
}

// Internal 创建内部错误
func Internal(op string, cause error) *CodeError {
	return NewErrorWithCause(ErrCodeInternal, "internal error occurred", cause).
		WithDetail("operation", op)
}

// Timeout 创建超时错误
func Timeout(op string, duration any) *CodeError {
	return NewError(ErrCodeTimeout, "operation timeout").
		WithDetail("operation", op).
		WithDetail("duration", duration)
}

// ============================================================================
// 错误码查询
// ============================================================================

// GetCode 获取错误码
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

// IsCode 检查错误是否匹配指定错误码
func IsCode(err error, code ErrorCode) bool {
	return GetCode(err) == code
}
