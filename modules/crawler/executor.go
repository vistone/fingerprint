package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/vistone/fingerprint/modules/profiles"
)

// RequestExecutor - HTTP request executor
type RequestExecutor struct {
	profile *profiles.ClientProfile
	proxy   *url.URL
	config  *CrawlerConfig
	client  *http.Client
}

// Response - Simplified response structure
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	Duration   time.Duration
}

// NewRequestExecutor - Create request executor
func NewRequestExecutor(profile *profiles.ClientProfile, proxy *url.URL, config *CrawlerConfig) *RequestExecutor {
	return &RequestExecutor{
		profile: profile,
		proxy:   proxy,
		config:  config,
	}
}

// Do - Execute request
func (e *RequestExecutor) Do(ctx context.Context, targetURL string) (*Response, error) {
	// Use fhttp + utls to create fingerprinted HTTP client
	// Simplified implementation here, actual implementation needs to configure TLS and HTTP/2 based on profile

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set request headers
	e.setHeaders(req)

	// Create client
	client := &http.Client{
		Timeout: e.config.RequestTimeout,
	}

	if e.proxy != nil {
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(e.proxy),
		}
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       body,
		Duration:   time.Since(start),
	}, nil
}

// setHeaders - Set request headers
func (e *RequestExecutor) setHeaders(req *http.Request) {
	if e.profile == nil || e.profile.Headers == nil {
		return
	}

	h := e.profile.Headers
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
	if h.SecCHUA != "" {
		req.Header.Set("Sec-CH-UA", h.SecCHUA)
	}
	if h.SecCHUAPlatform != "" {
		req.Header.Set("Sec-CH-UA-Platform", h.SecCHUAPlatform)
	}

	// Add custom headers
	for k, v := range h.Custom {
		req.Header.Set(k, v)
	}
}
