package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type scannerRequest struct {
	HTMLContent    string `json:"html"`
	URL            string `json:"url,omitempty"`
	FollowRedirect bool   `json:"followRedirects,omitempty"`
	MaxRedirects   int    `json:"maxRedirects,omitempty"`
	ExecuteJS      bool   `json:"executeJs,omitempty"`
	WaitMs         int    `json:"waitMs,omitempty"`
	ScanTimeout    int    `json:"scanTimeout,omitempty"`
}

type scannerFetchResult struct {
	HTMLContent             string
	SourceURL               string
	Redirects               []string
	FetchMode               string
	BrowserError            string
	ExternalScriptStats     *ExternalScriptStats
	ExternalScriptsCaptured int
}

type scannerBrowserOptions struct {
	MaxRedirects int
	WaitMs       int
}

// =====================================================================
// Anti-Detection - HTML Injection Handler
// =====================================================================

// AntiDetectCodeHandler returns anti-detection JavaScript code (standalone endpoint)
func (g *Gateway) AntiDetectCodeHandler(w http.ResponseWriter, r *http.Request) {
	// Rate limit check
	clientIP := g.getClientIP(r)
	if !g.limiter.Allow(clientIP) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Check if enabled
	if g.injector == nil {
		http.Error(w, `{"error": "anti-detection not enabled"}`, http.StatusServiceUnavailable)
		return
	}

	// Get Profile ID parameter (optional)
	profileID := r.URL.Query().Get("profile")
	var code string
	if profileID != "" {
		profile, err := g.profileManager.GetProfile(profileID)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "profile not found: %s"}`, profileID), http.StatusNotFound)
			return
		}
		code = g.injector.GenerateInjectionCodeForProfile(profile)
	}

	// Generate code
	if code == "" {
		code = g.injector.GenerateInjectionCode()
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	w.Write([]byte(code))
}

// ProfileListHandler returns the list of available Profiles
func (g *Gateway) ProfileListHandler(w http.ResponseWriter, r *http.Request) {
	if g.profileManager == nil {
		http.Error(w, `{"error": "profile manager not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	profiles := g.profileManager.ListProfiles()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"profiles":    profiles,
		"default":     g.profileManager.defaultID,
		"total_count": len(profiles),
	})
}

// ProfileDetailHandler returns detailed info for the specified Profile
func (g *Gateway) ProfileDetailHandler(w http.ResponseWriter, r *http.Request) {
	if g.profileManager == nil {
		http.Error(w, `{"error": "profile manager not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	profileID := r.URL.Query().Get("id")
	if profileID == "" {
		http.Error(w, `{"error": "profile id required"}`, http.StatusBadRequest)
		return
	}

	profile, err := g.profileManager.GetProfile(profileID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "profile not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// writeJSONError safely writes a JSON error response without leaking internal info
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp, _ := json.Marshal(map[string]string{"error": msg})
	w.Write(resp)
}

// V8ScannerHandler scans JavaScript for fingerprint detection code
func (g *Gateway) V8ScannerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeScannerJSONError(w, http.StatusMethodNotAllowed, "POST method required")
		return
	}

	req, ok := parseScannerRequest(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), resolveScanTimeout(req.ScanTimeout))
	defer cancel()

	fetchResult, statusCode, err := g.resolveScannerFetch(ctx, req)
	if err != nil {
		writeScannerJSONError(w, statusCode, err.Error())
		return
	}

	result, statusCode, err := runScanner(ctx, fetchResult.HTMLContent)
	if err != nil {
		writeScannerJSONError(w, statusCode, err.Error())
		return
	}

	writeScannerSuccess(w, req, fetchResult, result)
}

func parseScannerRequest(w http.ResponseWriter, r *http.Request) (*scannerRequest, bool) {
	// Limit request body size (maximum 5MB).
	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)

	var request scannerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeScannerJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %s", err.Error()))
		return nil, false
	}

	if request.HTMLContent == "" && strings.TrimSpace(request.URL) == "" {
		writeScannerJSONError(w, http.StatusBadRequest, "html or url is required")
		return nil, false
	}

	return &request, true
}

func resolveScanTimeout(scanTimeoutSeconds int) time.Duration {
	scanTimeout := 20 * time.Second
	if scanTimeoutSeconds <= 0 {
		return scanTimeout
	}

	scanTimeout = time.Duration(scanTimeoutSeconds) * time.Second
	if scanTimeout < 10*time.Second {
		scanTimeout = 10 * time.Second
	}
	if scanTimeout > 120*time.Second {
		scanTimeout = 120 * time.Second
	}

	return scanTimeout
}

func (g *Gateway) resolveScannerFetch(ctx context.Context, req *scannerRequest) (*scannerFetchResult, int, error) {
	result := &scannerFetchResult{
		HTMLContent: req.HTMLContent,
		SourceURL:   req.URL,
		Redirects:   []string{},
		FetchMode:   "inline-html",
	}

	if strings.TrimSpace(req.URL) == "" {
		return result, http.StatusOK, nil
	}

	config := g.GetConfig()
	result.FetchMode = "http"
	followRedirect := true
	maxRedirects := req.MaxRedirects
	if maxRedirects <= 0 {
		maxRedirects = 10
	}

	if req.ExecuteJS || config.ScannerUseBrowser {
		g.tryScannerBrowserFetch(ctx, req, scannerBrowserOptions{
			MaxRedirects: maxRedirects,
			WaitMs:       req.WaitMs,
		}, config, result)
	}

	if result.FetchMode == "http" {
		if err := g.tryScannerHTTPFetch(ctx, req, followRedirect, maxRedirects, result); err != nil {
			fetchErr, ok := err.(*scannerFetchError)
			if ok {
				return nil, fetchErr.statusCode, fetchErr
			}
			return nil, http.StatusBadGateway, err
		}
	}

	if result.FetchMode == "headless-browser" && result.ExternalScriptStats != nil {
		result.ExternalScriptsCaptured = result.ExternalScriptStats.Count
	}

	return result, http.StatusOK, nil
}

type scannerFetchError struct {
	statusCode int
	message    string
}

func (e *scannerFetchError) Error() string {
	return e.message
}

func (g *Gateway) tryScannerBrowserFetch(
	ctx context.Context,
	req *scannerRequest,
	options scannerBrowserOptions,
	config *GatewayConfig,
	result *scannerFetchResult,
) {
	waitMs := options.WaitMs
	if waitMs <= 0 {
		waitMs = 1200
	}
	if waitMs > 3500 {
		waitMs = 3500
	}
	options.WaitMs = waitMs

	if strings.TrimSpace(config.ScannerBrowserWS) != "" {
		g.tryScannerHeadlessFetch(ctx, req, options, config, result)
	}
	if result.FetchMode != "headless-browser" {
		g.tryScannerRedirectEmulationFetch(ctx, req.URL, options.MaxRedirects, options.WaitMs, result)
	}
}

func (g *Gateway) tryScannerHeadlessFetch(
	ctx context.Context,
	req *scannerRequest,
	options scannerBrowserOptions,
	config *GatewayConfig,
	result *scannerFetchResult,
) {
	browserTimeout := config.ScannerBrowserTimeout
	if browserTimeout <= 0 {
		browserTimeout = 25 * time.Second
	}
	if deadline, ok := ctx.Deadline(); ok {
		left := time.Until(deadline) - 3*time.Second
		if left > browserTimeout {
			browserTimeout = left
		}
	}
	if browserTimeout > 25*time.Second {
		browserTimeout = 25 * time.Second
	}

	html, finalURL, chain, stats, err := fetchHTMLWithHeadlessBrowser(ctx, headlessFetchOptions{
		TargetURL:    req.URL,
		BrowserWS:    config.ScannerBrowserWS,
		MaxRedirects: options.MaxRedirects,
		WaitMs:       options.WaitMs,
		Timeout:      browserTimeout,
	})
	if err != nil {
		result.BrowserError = err.Error()
		return
	}
	if html == "" {
		return
	}

	result.HTMLContent = html
	result.SourceURL = finalURL
	result.Redirects = chain
	result.FetchMode = "headless-browser"
	result.ExternalScriptStats = stats
}

func (g *Gateway) tryScannerRedirectEmulationFetch(
	ctx context.Context,
	targetURL string,
	maxRedirects int,
	waitMs int,
	result *scannerFetchResult,
) {
	html, finalURL, chain, err := fetchHTMLWithClientSideRedirects(ctx, targetURL, maxRedirects, waitMs)
	if err != nil {
		if result.BrowserError == "" {
			result.BrowserError = err.Error()
		}
		return
	}
	if html == "" {
		return
	}

	result.HTMLContent = html
	result.SourceURL = finalURL
	result.Redirects = chain
	result.FetchMode = "js-redirect-emulation"
}

func (g *Gateway) tryScannerHTTPFetch(
	ctx context.Context,
	req *scannerRequest,
	followRedirect bool,
	maxRedirects int,
	result *scannerFetchResult,
) error {
	remainingBudget := 12 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		left := time.Until(deadline) - 500*time.Millisecond
		if left <= 2*time.Second {
			return &scannerFetchError{statusCode: http.StatusGatewayTimeout, message: "scan timeout"}
		}
		if left < remainingBudget {
			remainingBudget = left
		}
	}

	html, finalURL, chain, err := fetchHTMLWithRedirects(ctx, req.URL, followRedirect, maxRedirects, remainingBudget)
	if err != nil && result.HTMLContent == "" {
		return &scannerFetchError{
			statusCode: http.StatusBadGateway,
			message:    fmt.Sprintf("fetch url failed: %s", err.Error()),
		}
	}
	if html == "" {
		return nil
	}

	result.HTMLContent = html
	result.SourceURL = finalURL
	result.Redirects = chain
	return nil
}

func runScanner(ctx context.Context, htmlContent string) (*JSDetectionResult, int, error) {
	type scanResult struct {
		result *JSDetectionResult
		err    error
	}

	resultChan := make(chan scanResult, 1)
	go func() {
		result, err := ScanJavaScriptWithV8(ctx, htmlContent)
		resultChan <- scanResult{result: result, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, http.StatusGatewayTimeout, fmt.Errorf("scan timeout")
	case res := <-resultChan:
		if res.err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("scan failed: %s", res.err.Error())
		}
		return res.result, http.StatusOK, nil
	}
}

func writeScannerSuccess(w http.ResponseWriter, req *scannerRequest, fetchResult *scannerFetchResult, result *JSDetectionResult) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Scan-URL", fetchResult.SourceURL)

	responseData := map[string]interface{}{
		"url":                     req.URL,
		"finalUrl":                fetchResult.SourceURL,
		"redirectChain":           fetchResult.Redirects,
		"fetchMode":               fetchResult.FetchMode,
		"browserError":            fetchResult.BrowserError,
		"externalScriptsCaptured": fetchResult.ExternalScriptsCaptured,
		"result":                  result,
	}
	if fetchResult.ExternalScriptStats != nil && fetchResult.FetchMode == "headless-browser" {
		responseData["externalScriptDetails"] = fetchResult.ExternalScriptStats
	}

	_ = json.NewEncoder(w).Encode(responseData)
}

func writeScannerJSONError(w http.ResponseWriter, statusCode int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// InjectProxyHandler provides HTML proxy and auto-injection (for proxy mode)
func (g *Gateway) InjectProxyHandler() http.Handler {
	config := g.GetConfig()
	if g.injector == nil || config.AntiDetectProxyTarget == "" {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "proxy mode not configured", http.StatusServiceUnavailable)
		})
	}
	return g.injector
}

// GetInjectorMiddleware returns the injector middleware (for wrapping existing routes)
func (g *Gateway) GetInjectorMiddleware() func(http.Handler) http.Handler {
	if g.injector == nil {
		// Return transparent middleware
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return g.injector.InjectorMiddleware
}
