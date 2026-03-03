package network

import (
	"errors"
	"fmt"
	"strings"

	errs "github.com/vistone/fingerprint/internal/errors"
)

// ============================================================================
// Network 子包错误定义
// ============================================================================

var (
	// ErrTLSHandshakeFailed 表示 TLS 握手失败
	ErrTLSHandshakeFailed = fmt.Errorf("%w: tls handshake failed", errs.ErrInvalidFingerprint)

	// ErrConnectionFailed 表示连接失败
	ErrConnectionFailed = fmt.Errorf("%w: connection failed", errs.ErrInvalidFingerprint)

	// ErrInvalidNetworkConfig 表示无效的网络配置
	ErrInvalidNetworkConfig = fmt.Errorf("%w: invalid network config", errs.ErrInvalidFingerprint)
)

// IsUTLSPSKLimitationError 检查是否为 uTLS PSK 限制导致的错误
// uTLS 库用在某些情况下无法导出 PSK，需要特殊处理
func IsUTLSPSKLimitationError(err error) bool {
	if err == nil {
		return false
	}
	// 这是 uTLS 库返回的特定错误，需要特殊处理
	return strings.Contains(strings.ToLower(err.Error()), "empty psk detected")
}

// IsTLSHandshakeFailed 检查错误是否为 TLS 握手失败
func IsTLSHandshakeFailed(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrTLSHandshakeFailed)
}

// IsConnectionFailed 检查错误是否为连接失败
func IsConnectionFailed(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrConnectionFailed)
}
