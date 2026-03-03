package profiles

import (
	"errors"
	"fmt"

	errs "github.com/vistone/fingerprint/internal/errors"
)

// ============================================================================
// Profiles 子包错误定义
// ============================================================================

var (
	// ErrProfileNotFound 表示指纹配置不存在
	ErrProfileNotFound = fmt.Errorf("%w: profiles", errs.ErrProfileNotFound)

	// ErrInvalidProfile 表示指纹配置无效
	ErrInvalidProfile = fmt.Errorf("%w: profiles configuration", errs.ErrInvalidFingerprint)

	// ErrProfileInitializationFailed 表示指纹初始化失败
	ErrProfileInitializationFailed = fmt.Errorf("%w: profile initialization failed", errs.ErrInvalidFingerprint)

	// ErrClientHelloIDNotSupported 表示 ClientHello ID 不支持
	ErrClientHelloIDNotSupported = fmt.Errorf("%w: profiles client hello id", errs.ErrClientHelloSpecNotImplemented)
)

// IsProfileNotFound 检查错误是否为指纹不存在
func IsProfileNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrProfileNotFound)
}

// IsInvalidProfile 检查错误是否为无效指纹
func IsInvalidProfile(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidProfile)
}

// IsProfileInitializationFailed 检查错误是否为初始化失败
func IsProfileInitializationFailed(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrProfileInitializationFailed)
}

// IsClientHelloIDNotSupported 检查错误是否为 ClientHello ID 不支持
func IsClientHelloIDNotSupported(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrClientHelloIDNotSupported)
}
