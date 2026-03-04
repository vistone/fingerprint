// Package core 提供错误定义
package core

import "errors"

// 核心错误定义
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
)

// CoreError 核心错误类型
type CoreError struct {
	Op      string // 操作
	Err     error  // 原始错误
	Context string // 上下文信息
}

// Error 实现 error 接口
func (e *CoreError) Error() string {
	if e.Context != "" {
		return e.Op + ": " + e.Context + ": " + e.Err.Error()
	}
	return e.Op + ": " + e.Err.Error()
}

// Unwrap 返回原始错误
func (e *CoreError) Unwrap() error {
	return e.Err
}

// NewError 创建新的核心错误
func NewError(op string, err error) *CoreError {
	return &CoreError{Op: op, Err: err}
}

// NewErrorWithContext 创建带上下文的核心错误
func NewErrorWithContext(op, context string, err error) *CoreError {
	return &CoreError{Op: op, Context: context, Err: err}
}
