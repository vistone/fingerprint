package ja4

import (
	"errors"
	"fmt"

	errs "github.com/vistone/fingerprint/modules/errors"
)

// ============================================================================
// JA4 子包错误定义
// ============================================================================

var (
	// ErrInvalidClientHelloSpec 表示无效的 ClientHello 规范
	ErrInvalidClientHelloSpec = fmt.Errorf("%w: invalid client hello spec", errs.ErrInvalidFingerprint)

	// ErrProfileNotFound 表示 JA4 指纹配置不存在
	ErrProfileNotFound = fmt.Errorf("%w: ja4 profile", errs.ErrProfileNotFound)

	// ErrEmptyProfile 表示 JA4 指纹为空
	ErrEmptyProfile = fmt.Errorf("%w: empty ja4 profile", errs.ErrInvalidFingerprint)

	// ErrInvalidTLSVersion 表示无效的 TLS 版本
	ErrInvalidTLSVersion = fmt.Errorf("%w: invalid tls version", errs.ErrInvalidFingerprint)

	// ErrMissingRequiredField 表示缺少必需字段
	ErrMissingRequiredField = fmt.Errorf("%w: missing required field", errs.ErrInvalidFingerprint)
)

// IsInvalidClientHelloSpec 检查错误是否为无效的 ClientHello 规范
func IsInvalidClientHelloSpec(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidClientHelloSpec)
}

// IsJA4ProfileNotFound 检查错误是否为 JA4 指纹不存在
func IsJA4ProfileNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrProfileNotFound)
}

// IsEmptyProfile 检查错误是否为空指纹
func IsEmptyProfile(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrEmptyProfile)
}

// IsInvalidTLSVersion 检查错误是否为无效的 TLS 版本
func IsInvalidTLSVersion(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidTLSVersion)
}
