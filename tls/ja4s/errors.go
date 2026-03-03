package ja4s

import (
	"errors"
	"fmt"

	errs "github.com/vistone/fingerprint/internal/errors"
)

// ============================================================================
// JA4S 子包错误定义
// ============================================================================

var (
	// ErrInvalidServerHello 表示无效的 ServerHello 消息
	ErrInvalidServerHello = fmt.Errorf("%w: invalid server hello spec", errs.ErrInvalidFingerprint)

	// ErrMissingServerHello 表示缺少 ServerHello
	ErrMissingServerHello = fmt.Errorf("%w: missing server hello", errs.ErrInvalidFingerprint)

	// ErrInvalidCipherSuite 表示无效的密码套件
	ErrInvalidCipherSuite = fmt.Errorf("%w: invalid cipher suite", errs.ErrInvalidFingerprint)

	// ErrUnsupportedExtension 表示不支持的扩展
	ErrUnsupportedExtension = fmt.Errorf("%w: unsupported extension", errs.ErrInvalidFingerprint)
)

// IsInvalidServerHello 检查错误是否为无效的 ServerHello
func IsInvalidServerHello(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidServerHello)
}

// IsMissingServerHello 检查错误是否为缺少 ServerHello
func IsMissingServerHello(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrMissingServerHello)
}

// IsInvalidCipherSuite 检查错误是否为无效的密码套件
func IsInvalidCipherSuite(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidCipherSuite)
}
