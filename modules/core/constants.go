package core

import "time"

// ============================================================================
// time constants
// ============================================================================

const (
	// DefaultTimeout default timeout duration
	DefaultTimeout = 30 * time.Second
	// DefaultDialTimeout default connection timeout
	DefaultDialTimeout = 10 * time.Second
	// DefaultTLSTimeout default TLS handshake timeout
	DefaultTLSTimeout = 15 * time.Second
	// DefaultDNSTimeout default DNS resolution timeout
	DefaultDNSTimeout = 5 * time.Second
	// DefaultReadTimeout default read timeout
	DefaultReadTimeout = 10 * time.Second
)

// ============================================================================
// cache constants
// ============================================================================

const (
	// DefaultCacheSize default cache size
	DefaultCacheSize = 10000
	// DefaultCacheTTL default cache expiration time
	DefaultCacheTTL = 5 * time.Minute
	// MaxCacheKeySize maximum cache key length
	MaxCacheKeySize = 1024
)

// ============================================================================
// rate limiting constants
// ============================================================================

const (
	// DefaultRateLimit default requests per second limit
	DefaultRateLimit = 1000
	// DefaultRateLimitBurst default burst request count
	DefaultRateLimitBurst = 2000
	// DefaultRateLimitWindow default rate limit window
	DefaultRateLimitWindow = time.Second
	// DefaultVisitorTTL default visitor log expiration time
	DefaultVisitorTTL = 5 * time.Minute
)

// ============================================================================
// requestlimit
// ============================================================================

const (
	// MaxRequestBodySize maximum request body size (5MB)
	MaxRequestBodySize = 5 * 1024 * 1024
	// MaxRedirects maximum redirect count
	MaxRedirects = 10
	// MaxHeaderSize maximum header size (1MB)
	MaxHeaderSize = 1 * 1024 * 1024
)

// ============================================================================
// risk score thresholds
// ============================================================================

const (
	// RiskThresholdLow low risk threshold
	RiskThresholdLow = 0.1
	// RiskThresholdMedium medium risk threshold
	RiskThresholdMedium = 0.4
	// RiskThresholdHigh high risk threshold
	RiskThresholdHigh = 0.7
	// RiskThresholdCritical critical risk threshold
	RiskThresholdCritical = 0.9
)

// ============================================================================
// TLS constants
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
// HTTP client constants
// ============================================================================

const (
	// DefaultHTTPTimeout default HTTP timeout
	DefaultHTTPTimeout = 30 * time.Second
	// DefaultHTTPReadTimeout HTTP read timeout
	DefaultHTTPReadTimeout = 10 * time.Second
	// DefaultHTTPWriteTimeout HTTP write timeout
	DefaultHTTPWriteTimeout = 10 * time.Second
	// DefaultIdleConnTimeout idle connection timeout
	DefaultIdleConnTimeout = 30 * time.Second
	// DefaultResponseHeaderTimeout response header timeout
	DefaultResponseHeaderTimeout = 8 * time.Second
	// DefaultExpectContinueTimeout ExpectContinue timeout
	DefaultExpectContinueTimeout = 1 * time.Second
	// DefaultKeepAliveInterval TCP keep-alive interval
	DefaultKeepAliveInterval = 30 * time.Second
)

// ============================================================================
// scanner constants
// ============================================================================

const (
	// DefaultScanTimeout default scan timeout
	DefaultScanTimeout = 20 * time.Second
	// MinScanTimeout minimum scan timeout
	MinScanTimeout = 10 * time.Second
	// MaxScanTimeout maximum scan timeout
	MaxScanTimeout = 120 * time.Second
	// DefaultBrowserTimeout default browser scraping timeout
	DefaultBrowserTimeout = 25 * time.Second
	// DefaultBrowserWaitMs browser wait milliseconds
	DefaultBrowserWaitMs = 1200
	// MaxBrowserWaitMs maximum browser wait milliseconds
	MaxBrowserWaitMs = 3500
	// DefaultRemainingBudget default remaining time budget
	DefaultRemainingBudget = 12 * time.Second
	// MinRequestTimeout minimumrequesttimeout
	MinRequestTimeout = 3 * time.Second
	// MaxRequestTimeout maximumrequesttimeout
	MaxRequestTimeout = 20 * time.Second
)

// ============================================================================
// session constants
// ============================================================================

const (
	// DefaultSessionTimeout default session timeout
	DefaultSessionTimeout = 30 * time.Minute
	// DefaultPoolIdleTime connection pool idle time
	DefaultPoolIdleTime = 10 * time.Minute
)

// ============================================================================
// data size constants
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
	// MaxCapturedScriptBytes maximum captured script bytes
	MaxCapturedScriptBytes = 1 * 1024 * 1024
	// MaxConfigDataSize maximumconfigurationdatasize
	MaxConfigDataSize = 1 * 1024 * 1024
)
