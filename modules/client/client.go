// Package client 提供完整的浏览器指纹模拟客户端
// 从 TCP/IP 层到 TLS 层到 HTTP 层的全栈模拟
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

// 超时常量定义（使用 core 包标准值）
const (
	// TimeoutDialConnect - 建立 TCP/IP 连接超时
	TimeoutDialConnect = core.DefaultDialTimeout

	// TimeoutTLS - TLS 握手超时
	TimeoutTLS = core.DefaultTLSTimeout

	// TimeoutDNS - DNS 解析超时
	TimeoutDNS = core.DefaultDNSTimeout

	// TimeoutReadHeader - 读取 HTTP 响应头超时
	TimeoutReadHeader = core.DefaultReadTimeout

	// TimeoutRequest - 单个请求超时
	TimeoutRequest = core.DefaultTimeout

	// TimeoutTotal - 总请求超时（包括重定向）
	TimeoutTotal = 60 * time.Second

	// KeepAliveInterval - TCP 保活间隔
	KeepAliveInterval = 30 * time.Second
)

// BrowserClient 完整的浏览器指纹客户端
type BrowserClient struct {
	profile   profiles.ClientProfile
	transport *SmartTransport
	client    *fhttp.Client
	tracer    *RequestTracer
}

// ClientOptions 客户端选项
type ClientOptions struct {
	Timeout         time.Duration
	FollowRedirects bool
	ProxyURL        string
	// StrictFingerprint 禁止使用标准 TLS 兼容回退路径，确保请求始终走指纹链路。
	StrictFingerprint bool
}

// DefaultOptions 默认选项
var DefaultOptions = &ClientOptions{
	Timeout:         TimeoutRequest,
	FollowRedirects: true,
}

// NewBrowserClient 创建浏览器指纹客户端
func NewBrowserClient(profile profiles.ClientProfile, opts ...*ClientOptions) (*BrowserClient, error) {
	opt := DefaultOptions
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	}

	bc := &BrowserClient{
		profile: profile,
	}

	// 创建智能传输层（支持 HTTP/2 → HTTP/1.1 回退）
	transport, err := NewSmartTransport(profile)
	if err != nil {
		return nil, fmt.Errorf("create transport failed: %w", err)
	}
	transport.SetStrictFingerprint(opt.StrictFingerprint)
	bc.transport = transport

	// 创建 HTTP 客户端（使用 fhttp）
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

// NewTracedClient 创建带追踪的客户端
func NewTracedClient(profile profiles.ClientProfile, url, method string, opts ...*ClientOptions) (*BrowserClient, error) {
	client, err := NewBrowserClient(profile, opts...)
	if err != nil {
		return nil, err
	}

	// 创建请求追踪器
	client.tracer = NewRequestTracer(profile, url, method)

	return client, nil
}

// Do 执行 HTTP 请求
func (bc *BrowserClient) Do(req *fhttp.Request) (*fhttp.Response, error) {
	// 显式设置 Host，避免部分传输实现在 HTTP/2 下发送空 authority 导致 400。
	if req.Host == "" && req.URL != nil {
		req.Host = req.URL.Host
	}

	// 应用指纹的请求头（覆盖所有默认头）
	bc.applyHeaders(req)

	// 执行请求
	return bc.client.Do(req)
}

// Get 发起 GET 请求
func (bc *BrowserClient) Get(url string) (*fhttp.Response, error) {
	req, err := fhttp.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return bc.Do(req)
}

// Post 发起 POST 请求
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

// applyHeaders 应用指纹的请求头。
// 只覆盖 profile 明确提供的值，避免把默认头清空导致目标站点 400。
func (bc *BrowserClient) applyHeaders(req *fhttp.Request) {
	// 如果 profile 没有 headers，保留默认头
	if bc.profile.Headers == nil {
		return
	}

	h := bc.profile.Headers

	// 必需请求头（确保应用，即使之前有值）
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

	// Sec-Fetch 头（如果存在则设置）
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

	// Chrome 特有头
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

	// 自定义请求头
	for key, value := range h.Custom {
		if value != "" {
			req.Header.Set(key, value)
		}
	}
}

// Close 关闭客户端
func (bc *BrowserClient) Close() error {
	if bc.transport != nil {
		return bc.transport.Close()
	}
	return nil
}

// GetProfile 获取使用的指纹
func (bc *BrowserClient) GetProfile() profiles.ClientProfile {
	return bc.profile
}

// GetTrace 获取请求追踪信息
func (bc *BrowserClient) GetTrace() *RequestTrace {
	if bc.tracer != nil {
		return bc.tracer.Trace
	}
	return nil
}

// ExecuteProxyRequest 执行带完整追踪的代理请求
func ExecuteProxyRequest(profile profiles.ClientProfile, url, method, body string, extraHeaders map[string]string) *ProxyResult {
	start := time.Now()

	// 规范化输入，降低部署环境下因输入格式导致的 400。
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = "GET"
	}
	url = strings.TrimSpace(url)
	if url != "" && !strings.HasPrefix(strings.ToLower(url), "http://") && !strings.HasPrefix(strings.ToLower(url), "https://") {
		url = "https://" + url
	}

	// 使用可修复副本，避免 profile 缺失关键头导致请求被站点拒绝。
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

	// 创建带追踪的客户端
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

	// 获取请求追踪信息
	result.RequestTrace = client.GetTrace()

	// 准备请求
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

	// 添加额外请求头
	for key, value := range extraHeaders {
		req.Header.Set(key, value)
	}

	// 记录连接开始时间
	connStart := time.Now()

	// 执行请求
	resp, err := client.Do(req)
	connTime := time.Since(connStart)

	if err != nil {
		result.Error = "Request execution failed"
		result.ErrorType = classifyRequestError(err)
		result.ErrorDetails["cause"] = err.Error()
		result.ErrorDetails["url"] = url
		result.ErrorDetails["elapsed_ms"] = connTime.Milliseconds()

		// 根据错误类型设置错误代码
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

	// 对 GET/HEAD 的瞬时错误状态进行有限重试：429/502/503/504。
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

	// 读取响应
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

	// 填充响应追踪信息
	result.ResponseTrace = &ResponseTrace{
		StatusCode:   resp.StatusCode,
		Status:       resp.Status,
		Protocol:     resp.Proto,
		Headers:      make(map[string]string),
		BodyLength:   len(respBody),
		ResponseTime: respTime,
	}

	// 响应体预览（前 2000 字符）
	if len(respBody) > 2000 {
		result.ResponseTrace.BodyPreview = string(respBody[:2000]) + "\n... (truncated)"
	} else {
		result.ResponseTrace.BodyPreview = string(respBody)
	}

	// 提取响应头
	for key, values := range resp.Header {
		if len(values) > 0 {
			result.ResponseTrace.Headers[key] = values[0]
		}
	}

	// 更新连接信息
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

	// 5xx 瞬时故障使用短退避。
	d := time.Duration(attempt*600) * time.Millisecond
	if d > 2*time.Second {
		return 2 * time.Second
	}
	return d
}

// classifyRequestError 分类请求错误
func classifyRequestError(err error) string {
	if err == nil {
		return "unknown"
	}

	errStr := err.Error()
	errLower := strings.ToLower(errStr)

	// 超时错误
	if strings.Contains(errLower, "timeout") || strings.Contains(errLower, "deadline exceeded") {
		return "timeout"
	}

	// 上下文取消
	if strings.Contains(errLower, "context canceled") {
		return "context_canceled"
	}

	// DNS 错误
	if strings.Contains(errLower, "lookup") || strings.Contains(errLower, "no such host") ||
		strings.Contains(errLower, "name resolution") {
		return "dns_error"
	}

	// TLS 错误
	if strings.Contains(errLower, "tls") || strings.Contains(errLower, "certificate") ||
		strings.Contains(errLower, "handshake") {
		return "tls_error"
	}

	// 协议错误
	if strings.Contains(errLower, "protocol_error") || strings.Contains(errLower, "http2") ||
		strings.Contains(errLower, "stream error") || strings.Contains(errLower, "broken pipe") ||
		strings.Contains(errLower, "reset by peer") {
		return "protocol_error"
	}

	// 网络错误
	if strings.Contains(errLower, "connection refused") || strings.Contains(errLower, "connection reset") ||
		strings.Contains(errLower, "connection aborted") || strings.Contains(errLower, "no route to host") ||
		strings.Contains(errLower, "network is unreachable") {
		return "network_error"
	}

	// 其他 I/O 错误
	if strings.Contains(errLower, "eof") || strings.Contains(errLower, "i/o") {
		return "io_error"
	}

	return "unknown"
}
