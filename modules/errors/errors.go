package errors

import (
	"errors"
	"fmt"
	"strings"
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
	ErrInvalidIP      = errors.New("invalid IP address")
	ErrInvalidPort    = errors.New("invalid port")
	ErrConnectionFailed = errors.New("connection failed")

	// 协议相关错误
	ErrInvalidTLSVersion   = errors.New("invalid TLS version")
	ErrInvalidCipherSuite  = errors.New("invalid cipher suite")
	ErrInvalidExtension    = errors.New("invalid extension")

	// 输入验证错误
	ErrInvalidInput   = errors.New("invalid input")
	ErrEmptyInput     = errors.New("empty input")
	ErrInputTooLarge  = errors.New("input too large")
	ErrRequiredField  = errors.New("required field is missing")

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
	CategoryConfig    ErrorCategory = "config"
	CategoryNetwork   ErrorCategory = "network"
	CategoryProtocol  ErrorCategory = "protocol"
	CategoryInput     ErrorCategory = "input"
	CategoryInternal  ErrorCategory = "internal"
	CategoryNotFound  ErrorCategory = "not_found"
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
