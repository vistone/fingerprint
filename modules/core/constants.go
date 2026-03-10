package core

import "time"

// ============================================================================
// 时间常量
// ============================================================================

const (
	// DefaultTimeout 默认超时时间
	DefaultTimeout = 30 * time.Second
	// DefaultDialTimeout 默认连接超时
	DefaultDialTimeout = 10 * time.Second
	// DefaultTLSTimeout 默认 TLS 握手超时
	DefaultTLSTimeout = 15 * time.Second
	// DefaultDNSTimeout 默认 DNS 解析超时
	DefaultDNSTimeout = 5 * time.Second
	// DefaultReadTimeout 默认读取超时
	DefaultReadTimeout = 10 * time.Second
)

// ============================================================================
// 缓存常量
// ============================================================================

const (
	// DefaultCacheSize 默认缓存大小
	DefaultCacheSize = 10000
	// DefaultCacheTTL 默认缓存过期时间
	DefaultCacheTTL = 5 * time.Minute
	// MaxCacheKeySize 最大缓存键长度
	MaxCacheKeySize = 1024
)

// ============================================================================
// 限流常量
// ============================================================================

const (
	// DefaultRateLimit 默认每秒请求数限制
	DefaultRateLimit = 1000
	// DefaultRateLimitBurst 默认突发请求数
	DefaultRateLimitBurst = 2000
	// DefaultRateLimitWindow 默认限流窗口
	DefaultRateLimitWindow = time.Second
	// DefaultVisitorTTL 默认访问者记录过期时间
	DefaultVisitorTTL = 5 * time.Minute
)

// ============================================================================
// 请求限制
// ============================================================================

const (
	// MaxRequestBodySize 最大请求体大小 (5MB)
	MaxRequestBodySize = 5 * 1024 * 1024
	// MaxRedirects 最大重定向次数
	MaxRedirects = 10
	// MaxHeaderSize 最大头大小 (1MB)
	MaxHeaderSize = 1 * 1024 * 1024
)

// ============================================================================
// 风险评分阈值
// ============================================================================

const (
	// RiskThresholdLow 低风险阈值
	RiskThresholdLow = 0.1
	// RiskThresholdMedium 中风险阈值
	RiskThresholdMedium = 0.4
	// RiskThresholdHigh 高风险阈值
	RiskThresholdHigh = 0.7
	// RiskThresholdCritical 严重风险阈值
	RiskThresholdCritical = 0.9
)

// ============================================================================
// TLS 常量
// ============================================================================

const (
	// TLSVersion13 TLS 1.3
	TLSVersion13 = 0x0304
	// TLSVersion12 TLS 1.2
	TLSVersion12 = 0x0303
	// TLSVersion11 TLS 1.1
	TLSVersion11 = 0x0302
	// TLSVersion10 TLS 1.0
	TLSVersion10 = 0x0301
)

// ============================================================================
// HTTP 客户端常量
// ============================================================================

const (
	// DefaultHTTPTimeout HTTP 默认超时
	DefaultHTTPTimeout = 30 * time.Second
	// DefaultHTTPReadTimeout HTTP 读取超时
	DefaultHTTPReadTimeout = 10 * time.Second
	// DefaultHTTPWriteTimeout HTTP 写入超时
	DefaultHTTPWriteTimeout = 10 * time.Second
	// DefaultIdleConnTimeout 空闲连接超时
	DefaultIdleConnTimeout = 30 * time.Second
	// DefaultResponseHeaderTimeout 响应头超时
	DefaultResponseHeaderTimeout = 8 * time.Second
	// DefaultExpectContinueTimeout ExpectContinue 超时
	DefaultExpectContinueTimeout = 1 * time.Second
	// DefaultKeepAliveInterval TCP 保活间隔
	DefaultKeepAliveInterval = 30 * time.Second
)

// ============================================================================
// 扫描器常量
// ============================================================================

const (
	// DefaultScanTimeout 默认扫描超时
	DefaultScanTimeout = 20 * time.Second
	// MinScanTimeout 最小扫描超时
	MinScanTimeout = 10 * time.Second
	// MaxScanTimeout 最大扫描超时
	MaxScanTimeout = 120 * time.Second
	// DefaultBrowserTimeout 浏览器抓取默认超时
	DefaultBrowserTimeout = 25 * time.Second
	// DefaultBrowserWaitMs 浏览器等待毫秒数
	DefaultBrowserWaitMs = 1200
	// MaxBrowserWaitMs 最大浏览器等待毫秒数
	MaxBrowserWaitMs = 3500
	// DefaultRemainingBudget 默认剩余时间预算
	DefaultRemainingBudget = 12 * time.Second
	// MinRequestTimeout 最小请求超时
	MinRequestTimeout = 3 * time.Second
	// MaxRequestTimeout 最大请求超时
	MaxRequestTimeout = 20 * time.Second
)

// ============================================================================
// 会话常量
// ============================================================================

const (
	// DefaultSessionTimeout 默认会话超时
	DefaultSessionTimeout = 30 * time.Minute
	// DefaultPoolIdleTime 连接池空闲时间
	DefaultPoolIdleTime = 10 * time.Minute
)

// ============================================================================
// 数据大小常量
// ============================================================================

const (
	// Size1KB 1 KB
	Size1KB = 1024
	// Size1MB 1 MB
	Size1MB = 1024 * 1024
	// Size5MB 5 MB
	Size5MB = 5 * 1024 * 1024
	// Size10MB 10 MB
	Size10MB = 10 * 1024 * 1024
	// MaxCapturedScriptBytes 最大捕获脚本字节数
	MaxCapturedScriptBytes = 1 * 1024 * 1024
	// MaxConfigDataSize 最大配置数据大小
	MaxConfigDataSize = 1 * 1024 * 1024
)
