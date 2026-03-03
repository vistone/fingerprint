package ja3

import (
	"errors"
	"fmt"

	errs "github.com/vistone/fingerprint/internal/errors"
)

// ============================================================================
// JA3 子包错误定义
// ============================================================================

var (
	// ErrInvalidClientHelloSpec 表示无效的 ClientHello 规范
	ErrInvalidClientHelloSpec = fmt.Errorf("%w: invalid client hello spec", errs.ErrInvalidFingerprint)

	// ErrProfileNotFound 表示 JA3 指纹配置不存在
	ErrProfileNotFound = fmt.Errorf("%w: ja3 profile", errs.ErrProfileNotFound)

	// ErrEmptyProfile 表示 JA3 指纹为空
	ErrEmptyProfile = fmt.Errorf("%w: empty ja3 profile", errs.ErrInvalidFingerprint)

	// ErrClientHelloIDNotImplemented 表示 ClientHello ID 不支持 Spec 导出
	ErrClientHelloIDNotImplemented = fmt.Errorf("%w: ja3 client hello id", errs.ErrClientHelloSpecNotImplemented)
)

// IsInvalidClientHelloSpec 检查错误是否为无效的 ClientHello 规范
func IsInvalidClientHelloSpec(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidClientHelloSpec)
}

// IsJA3ProfileNotFound 检查错误是否为 JA3 指纹不存在
func IsJA3ProfileNotFound(err error) bool {
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

// IsClientHelloIDNotImplemented 检查错误是否为 ClientHello ID 不支持
func IsClientHelloIDNotImplemented(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrClientHelloIDNotImplemented)
}
