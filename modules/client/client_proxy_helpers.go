package client

import (
	"fmt"
	"io"
	"strings"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/vistone/fingerprint/modules/profiles"
)

// retryParams groups parameters for idempotent request retry.
type retryParams struct {
	method       string
	url          string
	extraHeaders map[string]string
}

func normalizeProxyInput(method, url string) (string, string) {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = "GET"
	}
	url = strings.TrimSpace(url)
	if url != "" && !strings.HasPrefix(strings.ToLower(url), "http://") && !strings.HasPrefix(strings.ToLower(url), "https://") {
		url = "https://" + url
	}
	return method, url
}

func newProxyResult(profile profiles.ClientProfile) (*profiles.ClientProfile, *ProxyResult) {
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
	return &fixedProfile, result
}

func buildHTTPRequest(method, url, body string, headers map[string]string) (*fhttp.Request, error) {
	var bodyReader io.Reader
	if body != "" && method != "GET" && method != "HEAD" {
		bodyReader = strings.NewReader(body)
	}
	req, err := fhttp.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

// retryTransientErrors retries idempotent requests on transient HTTP errors.
// Returns nil if the retry sequence failed (result is populated with error info).
func retryTransientErrors(c *BrowserClient, resp *fhttp.Response, p retryParams, result *ProxyResult) *fhttp.Response {
	if p.method != "GET" && p.method != "HEAD" {
		return resp
	}

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

		retryReq, reqErr := fhttp.NewRequest(p.method, p.url, nil)
		if reqErr != nil {
			result.ErrorDetails["retry_request_error"] = reqErr.Error()
			break
		}
		for key, value := range p.extraHeaders {
			retryReq.Header.Set(key, value)
		}

		var err error
		resp, err = c.Do(retryReq)
		if err != nil {
			result.Error = "Request execution failed"
			result.ErrorType = classifyRequestError(err)
			result.ErrorDetails["cause"] = err.Error()
			result.ErrorDetails["url"] = p.url
			result.ErrorCode = 500
			return nil
		}
	}
	return resp
}

// buildProxyResponse reads the response body and populates the ProxyResult.
func buildProxyResponse(resp *fhttp.Response, result *ProxyResult, start time.Time, url string) *ProxyResult {
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

	result.ResponseTrace = &ResponseTrace{
		StatusCode:   resp.StatusCode,
		Status:       resp.Status,
		Protocol:     resp.Proto,
		Headers:      make(map[string]string),
		BodyLength:   len(respBody),
		ResponseTime: respTime,
	}

	if len(respBody) > 2000 {
		result.ResponseTrace.BodyPreview = string(respBody[:2000]) + "\n... (truncated)"
	} else {
		result.ResponseTrace.BodyPreview = string(respBody)
	}

	for key, values := range resp.Header {
		if len(values) > 0 {
			result.ResponseTrace.Headers[key] = values[0]
		}
	}

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
