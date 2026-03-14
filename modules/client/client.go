// Package client provides a complete browser fingerprint simulation client
// Full-stack simulation from TCP/IP to TLS to HTTP layer
package client

import (
	"fmt"
	"io"
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

	// Create smart transport layer (supports HTTP/2 → HTTP/1.1 fallback)
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
