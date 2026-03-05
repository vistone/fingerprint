// Package core 提供错误定义
package core

import (
	"errors"
	"fmt"
)

// ============================================================================
// 错误代码系统
// ============================================================================

// ErrorCode 错误码类型
type ErrorCode string

// 错误码分类前缀：
// VAL - 验证错误 (Validation)
// NTF - 未找到错误 (Not Found)
// SYS - 系统错误 (System)
// NET - 网络错误 (Network)
// SEC - 安全错误 (Security)
// CFG - 配置错误 (Configuration)
const (
	// 验证错误 (VAL001-VAL099)
	ErrCodeInvalidInput     ErrorCode = "VAL001"
	ErrCodeRequiredField    ErrorCode = "VAL002"
	ErrCodeInvalidFormat    ErrorCode = "VAL003"
	ErrCodeOutOfRange       ErrorCode = "VAL004"
	ErrCodeNilPointer       ErrorCode = "VAL005"
	ErrCodeInvalidType      ErrorCode = "VAL006"
	ErrCodeEmptyValue       ErrorCode = "VAL007"
	ErrCodeInvalidLength    ErrorCode = "VAL008"
	ErrCodeInvalidCharset   ErrorCode = "VAL009"
	ErrCodeInvalidPattern   ErrorCode = "VAL010"

	// 未找到错误 (NTF001-NTF099)
	ErrCodeProfileNotFound  ErrorCode = "NTF001"
	ErrCodeResourceNotFound ErrorCode = "NTF002"
	ErrCodeKeyNotFound      ErrorCode = "NTF003"
	ErrCodeFileNotFound     ErrorCode = "NTF004"
	ErrCodeRouteNotFound    ErrorCode = "NTF005"

	// 系统错误 (SYS001-SYS099)
	ErrCodeInternal         ErrorCode = "SYS001"
	ErrCodeNotImplemented   ErrorCode = "SYS002"
	ErrCodeIOError          ErrorCode = "SYS003"
	ErrCodeMemoryError      ErrorCode = "SYS004"
	ErrCodeTimeout          ErrorCode = "SYS005"
	ErrCodeUnavailable      ErrorCode = "SYS006"
	ErrCodeAlreadyExists    ErrorCode = "SYS007"
	ErrCodeConcurrencyError ErrorCode = "SYS008"
	ErrCodeStateError       ErrorCode = "SYS009"

	// 网络错误 (NET001-NET099)
	ErrCodeNetworkError     ErrorCode = "NET001"
	ErrCodeConnectionFailed ErrorCode = "NET002"
	ErrCodeDNSFailed        ErrorCode = "NET003"
	ErrCodeTLSFailed        ErrorCode = "NET004"
	ErrCodeHTTPError        ErrorCode = "NET005"
	ErrCodeProxyError       ErrorCode = "NET006"
	ErrCodeTimeoutNetwork   ErrorCode = "NET007"

	// 安全错误 (SEC001-SEC099)
	ErrCodeSecurityError    ErrorCode = "SEC001"
	ErrCodeAuthFailed       ErrorCode = "SEC002"
	ErrCodeUnauthorized     ErrorCode = "SEC003"
	ErrCodeForbidden        ErrorCode = "SEC004"
	ErrCodeInvalidToken     ErrorCode = "SEC005"
	ErrCodeReplayAttack     ErrorCode = "SEC006"
	ErrCodeFingerprintMismatch ErrorCode = "SEC007"
	ErrCodeInvalidTLSVersion ErrorCode = "SEC008"
	ErrCodeInvalidJA3Hash   ErrorCode = "SEC009"
	ErrCodeInvalidJA4Hash   ErrorCode = "SEC010"

	// 配置错误 (CFG001-CFG099)
	ErrCodeConfigError      ErrorCode = "CFG001"
	ErrCodeInvalidConfig    ErrorCode = "CFG002"
	ErrCodeMissingConfig    ErrorCode = "CFG003"
	ErrCodeConfigParseError ErrorCode = "CFG004"
)

// ErrorCategory 获取错误码分类
func (ec ErrorCode) Category() string {
	if len(ec) >= 3 {
		switch ec[:3] {
		case "VAL":
			return "Validation"
		case "NTF":
			return "NotFound"
		case "SYS":
			return "System"
		case "NET":
			return "Network"
		case "SEC":
			return "Security"
		case "CFG":
			return "Configuration"
		}
	}
	return "Unknown"
}

// HTTPStatus 获取错误码对应的 HTTP 状态码
func (ec ErrorCode) HTTPStatus() int {
	switch ec.Category() {
	case "Validation":
		return 400
	case "NotFound":
		return 404
	case "System":
		return 500
	case "Network":
		return 503
	case "Security":
		return 403
	case "Configuration":
		return 500
	default:
		return 500
	}
}

// ============================================================================
// 核心错误定义
// ============================================================================

var (
	// ErrInvalidProfile 无效的指纹配置
	ErrInvalidProfile = errors.New("invalid fingerprint profile")

	// ErrProfileNotFound 未找到指纹配置
	ErrProfileNotFound = errors.New("fingerprint profile not found")

	// ErrUnsupportedBrowser 不支持的浏览器类型
	ErrUnsupportedBrowser = errors.New("unsupported browser type")

	// ErrInvalidTLSVersion 无效的 TLS 版本
	ErrInvalidTLSVersion = errors.New("invalid TLS version")

	// ErrInvalidJA3Hash 无效的 JA3 哈希
	ErrInvalidJA3Hash = errors.New("invalid JA3 hash")

	// ErrInvalidJA4Hash 无效的 JA4 哈希
	ErrInvalidJA4Hash = errors.New("invalid JA4 hash")

	// ErrFeatureExtractionFailed 特征提取失败
	ErrFeatureExtractionFailed = errors.New("feature extraction failed")

	// ErrClassificationFailed 分类失败
	ErrClassificationFailed = errors.New("classification failed")

	// ErrRiskAssessmentFailed 风险评估失败
	ErrRiskAssessmentFailed = errors.New("risk assessment failed")

	// ErrNilPointer nil 指针错误
	ErrNilPointer = errors.New("nil pointer dereference")

	// ErrInvalidInput 无效输入
	ErrInvalidInput = errors.New("invalid input")

	// ErrOutOfRange 超出范围
	ErrOutOfRange = errors.New("value out of range")

	// ErrEmptyValue 空值错误
	ErrEmptyValue = errors.New("empty value")

	// ErrAlreadyExists 已存在
	ErrAlreadyExists = errors.New("already exists")
)

// ============================================================================
// 错误类型
// ============================================================================

// CoreError 核心错误类型（带错误码）
type CoreError struct {
	Code    ErrorCode // 错误码
	Op      string    // 操作
	Err     error     // 原始错误
	Context string    // 上下文信息
}

// Error 实现 error 接口
func (e *CoreError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("[%s] %s: %s: %v", e.Code, e.Op, e.Context, e.Err)
	}
	return fmt.Sprintf("[%s] %s: %v", e.Code, e.Op, e.Err)
}

// Unwrap 返回原始错误
func (e *CoreError) Unwrap() error {
	return e.Err
}

// Is 实现 errors.Is
func (e *CoreError) Is(target error) bool {
	if t, ok := target.(*CoreError); ok {
		return e.Code == t.Code
	}
	return errors.Is(e.Err, target)
}

// Category 获取错误分类
func (e *CoreError) Category() string {
	return e.Code.Category()
}

// HTTPStatus 获取 HTTP 状态码
func (e *CoreError) HTTPStatus() int {
	return e.Code.HTTPStatus()
}

// ============================================================================
// 错误创建函数
// ============================================================================

// NewError 创建新的核心错误
func NewError(op string, err error) *CoreError {
	return &CoreError{Op: op, Err: err}
}

// NewErrorWithContext 创建带上下文的核心错误
func NewErrorWithContext(op, context string, err error) *CoreError {
	return &CoreError{Op: op, Context: context, Err: err}
}

// NewCodedError 创建带错误码的核心错误
func NewCodedError(code ErrorCode, op string, err error) *CoreError {
	return &CoreError{Code: code, Op: op, Err: err}
}

// NewCodedErrorf 创建带格式化上下文的错误
func NewCodedErrorf(code ErrorCode, op, format string, args ...interface{}) *CoreError {
	return &CoreError{
		Code:    code,
		Op:      op,
		Context: fmt.Sprintf(format, args...),
		Err:     errors.New("validation failed"),
	}
}

// WrapError 包装错误
func WrapError(code ErrorCode, op string, err error, context string) *CoreError {
	return &CoreError{Code: code, Op: op, Err: err, Context: context}
}

// WrapErrorf 使用格式化字符串包装错误
func WrapErrorf(code ErrorCode, op string, err error, format string, args ...interface{}) *CoreError {
	return &CoreError{
		Code:    code,
		Op:      op,
		Err:     err,
		Context: fmt.Sprintf(format, args...),
	}
}

// IsCoreError 检查是否为 CoreError
func IsCoreError(err error) bool {
	var ce *CoreError
	return errors.As(err, &ce)
}

// GetErrorCode 获取错误码（如果不是 CoreError 返回空字符串）
func GetErrorCode(err error) ErrorCode {
	var ce *CoreError
	if errors.As(err, &ce) {
		return ce.Code
	}
	return ""
}

// IsErrorCode 检查错误是否匹配指定错误码
func IsErrorCode(err error, code ErrorCode) bool {
	return GetErrorCode(err) == code
}
