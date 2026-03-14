package client

import (
	"strconv"
	"strings"
	"time"

	"github.com/vistone/fingerprint/modules/profiles"
)

func ExecuteProxyRequest(profile profiles.ClientProfile, url, method, body string, extraHeaders map[string]string) *ProxyResult {
	start := time.Now()
	method, url = normalizeProxyInput(method, url)
	fixedProfile, result := newProxyResult(profile)

	client, err := NewTracedClient(*fixedProfile, url, method, &ClientOptions{
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

	result.RequestTrace = client.GetTrace()

	req, err := buildHTTPRequest(method, url, body, extraHeaders)
	if err != nil {
		result.Error = "Failed to create request"
		result.ErrorType = "request_creation_error"
		result.ErrorCode = 400
		result.ErrorDetails["cause"] = err.Error()
		result.ErrorDetails["url"] = url
		result.ErrorDetails["method"] = method
		return result
	}

	connStart := time.Now()
	resp, err := client.Do(req)
	connTime := time.Since(connStart)

	if err != nil {
		result.Error = "Request execution failed"
		result.ErrorType = classifyRequestError(err)
		result.ErrorDetails["cause"] = err.Error()
		result.ErrorDetails["url"] = url
		result.ErrorDetails["elapsed_ms"] = connTime.Milliseconds()
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
		result.RequestTrace.Connection = &ConnectionInfo{TotalTime: connTime}
		return result
	}

	resp = retryTransientErrors(client, resp, retryParams{method, url, extraHeaders}, result)
	if resp == nil {
		return result
	}
	defer resp.Body.Close()

	return buildProxyResponse(resp, result, start, url)
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
