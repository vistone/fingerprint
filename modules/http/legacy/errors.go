package http

import (
	"errors"
	"fmt"

	errs "github.com/vistone/fingerprint/modules/errors"
)

// ============================================================================
// HTTP 子包错误定义
// ============================================================================

var (
	// ErrInvalidUserAgent 表示无效的 User-Agent
	ErrInvalidUserAgent = fmt.Errorf("%w: invalid user agent", errs.ErrInvalidUserAgent)

	// ErrHeaderBuildingFailed 表示构建 HTTP 头失败
	ErrHeaderBuildingFailed = fmt.Errorf("%w: failed to build header", errs.ErrInvalidFingerprint)

	// ErrMissingRequiredHeader 表示缺少必需的 HTTP 头
	ErrMissingRequiredHeader = fmt.Errorf("%w: missing required header", errs.ErrInvalidFingerprint)

	// ErrInvalidHeaderValue 表示无效的 HTTP 头值
	ErrInvalidHeaderValue = fmt.Errorf("%w: invalid header value", errs.ErrInvalidFingerprint)
)

// IsInvalidUserAgent 检查错误是否为无效的 User-Agent
func IsInvalidUserAgent(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidUserAgent)
}

// IsHeaderBuildingFailed 检查错误是否为头部构建失败
func IsHeaderBuildingFailed(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrHeaderBuildingFailed)
}

// IsMissingRequiredHeader 检查错误是否为缺少头部
func IsMissingRequiredHeader(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrMissingRequiredHeader)
}
