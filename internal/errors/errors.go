package errors

import (
	"errors"
	"strings"
)

// ============================================================================
// 哨兵错误定义
// ============================================================================

var (
	// ErrProfileNotFound 表示请求的指纹配置不存在
	ErrProfileNotFound = errors.New("profile not found")

	// ErrInvalidFingerprint 表示指纹格式无效
	ErrInvalidFingerprint = errors.New("invalid fingerprint format")

	// ErrClientHelloSpecNotImplemented 表示 ClientHelloSpec 未实现
	ErrClientHelloSpecNotImplemented = errors.New("client hello spec not implemented")

	// ErrInvalidUserAgent 表示无效的 User-Agent
	ErrInvalidUserAgent = errors.New("invalid user agent")

	// ErrNoProfilesAvailable 表示没有可用的指纹配置
	ErrNoProfilesAvailable = errors.New("no profiles available")

	// ErrUnsupportedBrowser 表示不支持的浏览器类型
	ErrUnsupportedBrowser = errors.New("unsupported browser")
)

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
