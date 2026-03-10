// Package client provides a complete browser fingerprint simulation client
// Full-stack simulation from TCP/IP to TLS to HTTP layer
package client

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// Timeout constants definition (using core package standard values)
const (
	// TimeoutDialConnect - TCP/IP connection establishment timeout
	TimeoutDialConnect = core.DefaultDialTimeout

	// TimeoutTLS - TLS handshake timeout
	TimeoutTLS = core.DefaultTLSTimeout

	// TimeoutDNS - DNS resolution timeout
	TimeoutDNS = core.DefaultDNSTimeout

	// TimeoutReadHeader - HTTP response header read timeout
	TimeoutReadHeader = core.DefaultReadTimeout

	// TimeoutRequest - Single request timeout
	TimeoutRequest = core.DefaultTimeout

	// TimeoutTotal - Total request timeout (including redirects)
	TimeoutTotal = 60 * time.Second

	// KeepAliveInterval - TCP keep-alive interval
	KeepAliveInterval = 30 * time.Second
)

// BrowserClient is the complete browser fingerprint client
type BrowserClient struct {
	profile   profiles.ClientProfile
	transport *SmartTransport
	client    *fhttp.Client
	tracer    *RequestTracer
}

// ClientOptions defines client configuration options
type ClientOptions struct {
	Timeout         time.Duration
	FollowRedirects bool
	ProxyURL        string
	// StrictFingerprint disallows standard TLS compatibility fallback, ensuring requests always use fingerprint chain
	StrictFingerprint bool
}

// DefaultOptions provides default client options
var DefaultOptions = &ClientOptions{
	Timeout:         TimeoutRequest,
	FollowRedirects: true,
}

// NewBrowserClient creates a new browser fingerprint client
func NewBrowserClient(profile profiles.ClientProfile, opts ...*ClientOptions) (*BrowserClient, error) {
	opt := DefaultOptions
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	}

	bc := &BrowserClient{
		profile: profile,
	}

	// translated comment
	transport, err := NewSmartTransport(profile)
	if err != nil {
		return nil, fmt.Errorf("create transport failed: %w", err)
	}
	transport.SetStrictFingerprint(opt.StrictFingerprint)
	bc.transport = transport

	// Create HTTP client (using fhttp)
	bc.client = &fhttp.Client{
		Transport: bc.transport,
		Timeout:   opt.Timeout,
		CheckRedirect: func(req *fhttp.Request, via []*fhttp.Request) error {
			if !opt.FollowRedirects {
				return fhttp.ErrUseLastResponse
			}
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	return bc, nil
}

// NewTracedClient creates a new client with request tracing
func NewTracedClient(profile profiles.ClientProfile, url, method string, opts ...*ClientOptions) (*BrowserClient, error) {
	client, err := NewBrowserClient(profile, opts...)
	if err != nil {
		return nil, err
	}

	// Create request tracer
	client.tracer = NewRequestTracer(profile, url, method)

	return client, nil
}

// Do executes an HTTP request
func (bc *BrowserClient) Do(req *fhttp.Request) (*fhttp.Response, error) {
	// Explicitly set Host to prevent some HTTP/2 transport implementations from sending empty authority causing 400
	if req.Host == "" && req.URL != nil {
		req.Host = req.URL.Host
	}

	// Apply fingerprint request headers (override all defaults)
	bc.applyHeaders(req)

	// Execute request
	return bc.client.Do(req)
}

// Get initiates a GET request
func (bc *BrowserClient) Get(url string) (*fhttp.Response, error) {
	req, err := fhttp.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return bc.Do(req)
}

// Post initiates a POST request
func (bc *BrowserClient) Post(url string, contentType string, body io.Reader) (*fhttp.Response, error) {
	req, err := fhttp.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return bc.Do(req)
}

// applyHeaders applies fingerprint request headers
// Only override values explicitly provided by profile to avoid clearing defaults causing 400
func (bc *BrowserClient) applyHeaders(req *fhttp.Request) {
	// If profile has no headers, keep defaults
	if bc.profile.Headers == nil {
		return
	}

	h := bc.profile.Headers

	// Required request headers (ensure applied even if already set)
	if h.UserAgent != "" {
		req.Header.Set("User-Agent", h.UserAgent)
	}
	if h.Accept != "" {
		req.Header.Set("Accept", h.Accept)
	}
	if h.AcceptLanguage != "" {
		req.Header.Set("Accept-Language", h.AcceptLanguage)
	}
	if h.AcceptEncoding != "" {
		req.Header.Set("Accept-Encoding", h.AcceptEncoding)
	}

	// Sec-Fetch headers (set if present)
	if h.SecFetchSite != "" {
		req.Header.Set("Sec-Fetch-Site", h.SecFetchSite)
	}
	if h.SecFetchMode != "" {
		req.Header.Set("Sec-Fetch-Mode", h.SecFetchMode)
	}
	if h.SecFetchDest != "" {
		req.Header.Set("Sec-Fetch-Dest", h.SecFetchDest)
	}
	if h.SecFetchUser != "" {
		req.Header.Set("Sec-Fetch-User", h.SecFetchUser)
	}

	// Chrome-specific headers
	if h.SecCHUA != "" {
		req.Header.Set("Sec-CH-UA", h.SecCHUA)
	}
	if h.SecCHUAMobile != "" {
		req.Header.Set("Sec-CH-UA-Mobile", h.SecCHUAMobile)
	}
	if h.SecCHUAPlatform != "" {
		req.Header.Set("Sec-CH-UA-Platform", h.SecCHUAPlatform)
	}
	if h.UpgradeInsecureRequests != "" {
		req.Header.Set("Upgrade-Insecure-Requests", h.UpgradeInsecureRequests)
	}

	// Custom request headers
	for key, value := range h.Custom {
		if value != "" {
			req.Header.Set(key, value)
		}
	}
}

// Close closes the client
func (bc *BrowserClient) Close() error {
	if bc.transport != nil {
		return bc.transport.Close()
	}
	return nil
}

// GetProfile returns the fingerprint profile in use
func (bc *BrowserClient) GetProfile() profiles.ClientProfile {
	return bc.profile
}

// GetTrace returns request tracing information
func (bc *BrowserClient) GetTrace() *RequestTrace {
	if bc.tracer != nil {
		return bc.tracer.Trace
	}
	return nil
}

// ExecuteProxyRequest executes a proxy request with complete tracing
func ExecuteProxyRequest(profile profiles.ClientProfile, url, method, body string, extraHeaders map[string]string) *ProxyResult {
	start := time.Now()

	// Normalize input to reduce 400 errors due to input format in deployment environments
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = "GET"
	}
	url = strings.TrimSpace(url)
	if url != "" && !strings.HasPrefix(strings.ToLower(url), "http://") && !strings.HasPrefix(strings.ToLower(url), "https://") {
		url = "https://" + url
	}

	// Use fixable copy to prevent requests from being rejected due to missing critical headers in profile
	fixedProfile := profile
	_ = profiles.ValidateAndRepair(&fixedProfile)

	result := &ProxyResult{
		Success:      false,
		ErrorDetails: make(map[string]interface{}),
		ProfileUsed: &ProfileInfo{
			ID:             fixedProfile.ID,
			Name:           fixedProfile.Name,
			BrowserType:    string(fixedProfile.BrowserType),
			BrowserVersion: fixedProfile.BrowserVersion,
			OS:             string(fixedProfile.OS),
			OSVersion:      fixedProfile.OSVersion,
		},
	}

	// Create traced client
	client, err := NewTracedClient(fixedProfile, url, method, &ClientOptions{
		Timeout:         TimeoutTotal,
		FollowRedirects: true,
	})
	if err != nil {
		result.Error = "Failed to create client"
		result.ErrorType = "initialization_error"
		result.ErrorCode = 500
		result.ErrorDetails["cause"] = err.Error()
		result.ErrorDetails["stage"] = "client_creation"
		return result
	}
	defer client.Close()

	// Get request tracing information
	result.RequestTrace = client.GetTrace()

	// Prepare request
	var bodyReader io.Reader
	if body != "" && method != "GET" && method != "HEAD" {
		bodyReader = strings.NewReader(body)
	}

	req, err := fhttp.NewRequest(method, url, bodyReader)
	if err != nil {
		result.Error = "Failed to create request"
		result.ErrorType = "request_creation_error"
		result.ErrorCode = 400
		result.ErrorDetails["cause"] = err.Error()
		result.ErrorDetails["url"] = url
		result.ErrorDetails["method"] = method
		return result
	}

	// Add extra request headers
	for key, value := range extraHeaders {
		req.Header.Set(key, value)
	}

	// Record connection start time
	connStart := time.Now()

	// Execute request
	resp, err := client.Do(req)
	connTime := time.Since(connStart)

	if err != nil {
		result.Error = "Request execution failed"
		result.ErrorType = classifyRequestError(err)
		result.ErrorDetails["cause"] = err.Error()
		result.ErrorDetails["url"] = url
		result.ErrorDetails["elapsed_ms"] = connTime.Milliseconds()

		// Set error code based on error type
		switch result.ErrorType {
		case "timeout":
			result.ErrorCode = 504
		case "network_error":
			result.ErrorCode = 503
		case "protocol_error":
			result.ErrorCode = 502
		case "context_canceled":
			result.ErrorCode = 499
		default:
			result.ErrorCode = 500
		}

		result.RequestTrace.Connection = &ConnectionInfo{
			TotalTime: connTime,
		}
		return result
	}

	// Limited retry for transient error statuses on GET/HEAD: 429/502/503/504
	if method == "GET" || method == "HEAD" {
		const maxRetries = 2
		for attempt := 1; attempt <= maxRetries; attempt++ {
			if resp == nil || !isRetriableStatus(resp.StatusCode) {
				break
			}

			delay := retryDelayForStatus(resp.StatusCode, resp.Header.Get("Retry-After"), attempt)
			if delay <= 0 {
				break
			}

			result.ErrorDetails[fmt.Sprintf("retry_%d_status", attempt)] = resp.StatusCode
			result.ErrorDetails[fmt.Sprintf("retry_%d_delay_ms", attempt)] = delay.Milliseconds()

			_ = resp.Body.Close()
			time.Sleep(delay)

			retryReq, reqErr := fhttp.NewRequest(method, url, nil)
			if reqErr != nil {
				result.ErrorDetails["retry_request_error"] = reqErr.Error()
				break
			}
			for key, value := range extraHeaders {
				retryReq.Header.Set(key, value)
			}

			resp, err = client.Do(retryReq)
			if err != nil {
				result.Error = "Request execution failed"
				result.ErrorType = classifyRequestError(err)
				result.ErrorDetails["cause"] = err.Error()
				result.ErrorDetails["url"] = url
				result.ErrorDetails["elapsed_ms"] = connTime.Milliseconds()
				result.ErrorCode = 500
				return result
			}
		}
	}
	defer resp.Body.Close()

	// Read response
	respStart := time.Now()
	respBody, err := io.ReadAll(resp.Body)
	respTime := time.Since(respStart)

	if err != nil {
		result.Error = "Failed to read response body"
		result.ErrorType = "response_read_error"
		result.ErrorCode = 502
		result.ErrorDetails["cause"] = err.Error()
		result.ErrorDetails["status_code"] = resp.StatusCode
		result.ErrorDetails["read_time_ms"] = respTime.Milliseconds()
		return result
	}

	// Fill response trace information
	result.ResponseTrace = &ResponseTrace{
		StatusCode:   resp.StatusCode,
		Status:       resp.Status,
		Protocol:     resp.Proto,
		Headers:      make(map[string]string),
		BodyLength:   len(respBody),
		ResponseTime: respTime,
	}

	// Response body preview (first 2000 characters)
	if len(respBody) > 2000 {
		result.ResponseTrace.BodyPreview = string(respBody[:2000]) + "\n... (truncated)"
	} else {
		result.ResponseTrace.BodyPreview = string(respBody)
	}

	// Extract response headers
	for key, values := range resp.Header {
		if len(values) > 0 {
			result.ResponseTrace.Headers[key] = values[0]
		}
	}

	// Update connection information
	result.RequestTrace.Connection = &ConnectionInfo{
		RemoteAddr:   url,
		TotalTime:    time.Since(start),
		Protocol:     resp.Proto,
		ProtocolUsed: resp.Proto,
	}

	if resp.StatusCode >= 400 {
		result.Success = false
		result.Error = fmt.Sprintf("Target website returned HTTP %d", resp.StatusCode)
		result.ErrorType = "target_http_error"
		result.ErrorCode = resp.StatusCode
		result.ErrorDetails["status_code"] = resp.StatusCode
		result.ErrorDetails["status"] = resp.Status
		if server, ok := result.ResponseTrace.Headers["Server"]; ok && server != "" {
			result.ErrorDetails["server"] = server
		}
		return result
	}

	result.Success = true
	return result
}

func parseRetryAfter(value string) time.Duration {
	v := strings.TrimSpace(value)
	if v == "" {
		return 0
	}

	if secs, err := strconv.Atoi(v); err == nil {
		if secs <= 0 {
			return 0
		}
		if secs > 5 {
			secs = 5
		}
		return time.Duration(secs) * time.Second
	}

	if t, err := time.Parse(time.RFC1123, v); err == nil {
		d := time.Until(t)
		if d <= 0 {
			return 0
		}
		if d > 5*time.Second {
			return 5 * time.Second
		}
		return d
	}

	return 0
}

func isRetriableStatus(code int) bool {
	return code == 429 || code == 502 || code == 503 || code == 504
}

func retryDelayForStatus(statusCode int, retryAfter string, attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}

	if statusCode == 429 {
		if d := parseRetryAfter(retryAfter); d > 0 {
			return d
		}
		d := time.Duration(attempt*2) * time.Second
		if d > 6*time.Second {
			return 6 * time.Second
		}
		return d
	}

	// 5xx transient failures use short backoff
	d := time.Duration(attempt*600) * time.Millisecond
	if d > 2*time.Second {
		return 2 * time.Second
	}
	return d
}

// classifyRequestError classifies request errors
func classifyRequestError(err error) string {
	if err == nil {
		return "unknown"
	}

	errStr := err.Error()
	errLower := strings.ToLower(errStr)

	// Timeout errors
	if strings.Contains(errLower, "timeout") || strings.Contains(errLower, "deadline exceeded") {
		return "timeout"
	}

	// Context cancellation
	if strings.Contains(errLower, "context canceled") {
		return "context_canceled"
	}

	// DNS errors
	if strings.Contains(errLower, "lookup") || strings.Contains(errLower, "no such host") ||
		strings.Contains(errLower, "name resolution") {
		return "dns_error"
	}

	// TLS errors
	if strings.Contains(errLower, "tls") || strings.Contains(errLower, "certificate") ||
		strings.Contains(errLower, "handshake") {
		return "tls_error"
	}

	// Protocol errors
	if strings.Contains(errLower, "protocol_error") || strings.Contains(errLower, "http2") ||
		strings.Contains(errLower, "stream error") || strings.Contains(errLower, "broken pipe") ||
		strings.Contains(errLower, "reset by peer") {
		return "protocol_error"
	}

	// Network errors
	if strings.Contains(errLower, "connection refused") || strings.Contains(errLower, "connection reset") ||
		strings.Contains(errLower, "connection aborted") || strings.Contains(errLower, "no route to host") ||
		strings.Contains(errLower, "network is unreachable") {
		return "network_error"
	}

	// Other I/O errors
	if strings.Contains(errLower, "eof") || strings.Contains(errLower, "i/o") {
		return "io_error"
	}

	return "unknown"
}
