package client

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/vistone/fingerprint/modules/profiles"
)

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
