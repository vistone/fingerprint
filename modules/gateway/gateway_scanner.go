package gateway

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// =====================================================================
// P3 Anti-Detection - HTML Injection Handler
// =====================================================================

// AntiDetectCodeHandler returns P3 anti-detection JavaScript code (standalone endpoint)
func (g *Gateway) AntiDetectCodeHandler(w http.ResponseWriter, r *http.Request) {
	// Rate limit check
	clientIP := g.getClientIP(r)
	if !g.limiter.Allow(clientIP) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Check if enabled
	if g.injector == nil {
		http.Error(w, `{"error": "P3 anti-detection not enabled"}`, http.StatusServiceUnavailable)
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

	// Limit request body size (maximum 5MB)
	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)

	var request struct {
		HTMLContent    string `json:"html"`
		URL            string `json:"url,omitempty"`
		FollowRedirect bool   `json:"followRedirects,omitempty"`
		MaxRedirects   int    `json:"maxRedirects,omitempty"`
		ExecuteJS      bool   `json:"executeJs,omitempty"`
		WaitMs         int    `json:"waitMs,omitempty"`
		ScanTimeout    int    `json:"scanTimeout,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeScannerJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %s", err.Error()))
		return
	}

	if request.HTMLContent == "" && strings.TrimSpace(request.URL) == "" {
		writeScannerJSONError(w, http.StatusBadRequest, "html or url is required")
		return
	}

	// Call scanner (with configurable timeout)
	scanTimeout := 20 * time.Second
	if request.ScanTimeout > 0 {
		scanTimeout = time.Duration(request.ScanTimeout) * time.Second
		if scanTimeout < 10*time.Second {
			scanTimeout = 10 * time.Second
		}
		if scanTimeout > 120*time.Second {
			scanTimeout = 120 * time.Second
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), scanTimeout)
	defer cancel()

	htmlContent := request.HTMLContent
	sourceURL := request.URL
	redirects := []string{}
	fetchMode := "inline-html"
	browserError := ""
	usedHeadless := false
	externalScriptsCaptured := 0
	var externalScriptStats *ExternalScriptStats

	if strings.TrimSpace(request.URL) != "" {
		fetchMode = "http"
		followRedirect := request.FollowRedirect
		if !request.FollowRedirect {
			// Enable redirect following by default
			followRedirect = true
		}

		maxRedirects := request.MaxRedirects
		if maxRedirects <= 0 {
			maxRedirects = 10
		}

		wantBrowser := request.ExecuteJS || g.config.ScannerUseBrowser
		if wantBrowser {
			waitMs := request.WaitMs
			if waitMs <= 0 {
				waitMs = 1200
			}
			if waitMs > 3500 {
				waitMs = 3500
			}

			if strings.TrimSpace(g.config.ScannerBrowserWS) != "" {
				browserTimeout := g.config.ScannerBrowserTimeout
				if browserTimeout <= 0 {
					browserTimeout = 25 * time.Second
				}
				// Give headless fetch more time budget, but still reserve time for scan and encoding.
				if scanTimeout > 8*time.Second {
					candidate := scanTimeout - 3*time.Second
					if candidate > browserTimeout {
						browserTimeout = candidate
					}
				}
				if browserTimeout > 25*time.Second {
					browserTimeout = 25 * time.Second
				}

				browserHTML, browserFinalURL, browserChain, stats, err := fetchHTMLWithHeadlessBrowser(
					ctx,
					request.URL,
					g.config.ScannerBrowserWS,
					maxRedirects,
					waitMs,
					browserTimeout,
				)
				if err == nil && browserHTML != "" {
					htmlContent = browserHTML
					sourceURL = browserFinalURL
					redirects = browserChain
					fetchMode = "headless-browser"
					usedHeadless = true
					externalScriptStats = stats
				} else if err != nil {
					browserError = err.Error()
				}
			}

			if !usedHeadless {
				execHTML, execFinalURL, execChain, err := fetchHTMLWithClientSideRedirects(
					ctx,
					request.URL,
					maxRedirects,
					waitMs,
				)
				if err == nil && execHTML != "" {
					htmlContent = execHTML
					sourceURL = execFinalURL
					redirects = execChain
					fetchMode = "js-redirect-emulation"
				} else if err != nil && browserError == "" {
					browserError = err.Error()
				}
			}
		}

		if fetchMode == "http" {
			remainingBudget := 12 * time.Second
			if deadline, ok := ctx.Deadline(); ok {
				left := time.Until(deadline) - 500*time.Millisecond
				if left <= 2*time.Second {
					writeScannerJSONError(w, http.StatusGatewayTimeout, "scan timeout")
					return
				}
				if left < remainingBudget {
					remainingBudget = left
				}
			}

			fetchedHTML, finalURL, chain, err := fetchHTMLWithRedirects(ctx, request.URL, followRedirect, maxRedirects, remainingBudget)
			if err == nil && fetchedHTML != "" {
				htmlContent = fetchedHTML
				sourceURL = finalURL
				redirects = chain
			} else if htmlContent == "" {
				writeScannerJSONError(w, http.StatusBadGateway, fmt.Sprintf("fetch url failed: %s", err.Error()))
				return
			}
		}
	}

	if fetchMode == "headless-browser" && externalScriptStats != nil {
		externalScriptsCaptured = externalScriptStats.Count
	}

	// Execute scan in goroutine so it can be interrupted by ctx.Done()
	type scanResult struct {
		result *JSDetectionResult
		err    error
	}

	resultChan := make(chan scanResult, 1)
	go func() {
		result, err := ScanJavaScriptWithV8(ctx, htmlContent)
		resultChan <- scanResult{result, err}
	}()

	// Wait for scan to complete or timeout
	select {
	case <-ctx.Done():
		writeScannerJSONError(w, http.StatusGatewayTimeout, "scan timeout")
		return
	case res := <-resultChan:
		if res.err != nil {
			writeScannerJSONError(w, http.StatusInternalServerError, fmt.Sprintf("scan failed: %s", res.err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Scan-URL", sourceURL)
		responseData := map[string]interface{}{
			"url":                     request.URL,
			"finalUrl":                sourceURL,
			"redirectChain":           redirects,
			"fetchMode":               fetchMode,
			"browserError":            browserError,
			"externalScriptsCaptured": externalScriptsCaptured,
			"result":                  res.result,
		}
		if externalScriptStats != nil && fetchMode == "headless-browser" {
			responseData["externalScriptDetails"] = externalScriptStats
		}
		json.NewEncoder(w).Encode(responseData)
	}
}

func writeScannerJSONError(w http.ResponseWriter, statusCode int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// fetchHTMLWithRedirects fetches URL content and returns the final URL and redirect chain
func fetchHTMLWithRedirects(ctx context.Context, rawURL string, followRedirect bool, maxRedirects int, requestTimeout time.Duration) (string, string, []string, error) {
	redirectChain := []string{}
	trimmedURL := strings.TrimSpace(rawURL)
	if trimmedURL == "" {
		return "", "", redirectChain, fmt.Errorf("empty url")
	}
	if requestTimeout <= 0 {
		requestTimeout = 12 * time.Second
	}
	if requestTimeout < 3*time.Second {
		requestTimeout = 3 * time.Second
	}
	if requestTimeout > 20*time.Second {
		requestTimeout = 20 * time.Second
	}

	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt < 2; attempt++ {
		attemptChain := []string{}
		client := &http.Client{
			Timeout: requestTimeout,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   requestTimeout,
				ResponseHeaderTimeout: 8 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				IdleConnTimeout:       30 * time.Second,
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
		}
		if followRedirect {
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				attemptChain = append(attemptChain, req.URL.String())
				if len(via) >= maxRedirects {
					return fmt.Errorf("too many redirects: %d", maxRedirects)
				}
				return nil
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, trimmedURL, nil)
		if err != nil {
			return "", "", redirectChain, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

		resp, err = client.Do(req)
		if err == nil {
			redirectChain = attemptChain
			break
		}

		lastErr = err
		if !strings.Contains(strings.ToLower(err.Error()), "tls handshake timeout") || attempt == 1 {
			return "", "", redirectChain, err
		}
		// Retry once for transient TLS handshake stalls.
		time.Sleep(250 * time.Millisecond)
	}

	if resp == nil {
		if lastErr != nil {
			return "", "", redirectChain, lastErr
		}
		return "", "", redirectChain, fmt.Errorf("fetch failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 6*1024*1024))
	if err != nil {
		return "", "", redirectChain, err
	}

	return string(body), resp.Request.URL.String(), redirectChain, nil
}

// fetchHTMLWithClientSideRedirects fetches a page and emulates common client-side redirects.
// Supports meta refresh, window.location, location.href, location.replace redirect patterns.
func fetchHTMLWithClientSideRedirects(ctx context.Context, startURL string, maxHops int, waitMs int) (string, string, []string, error) {
	if maxHops <= 0 {
		maxHops = 5
	}
	if maxHops > 10 {
		maxHops = 10
	}

	currentURL := strings.TrimSpace(startURL)
	if currentURL == "" {
		return "", "", nil, fmt.Errorf("empty start url")
	}

	chain := []string{}
	visited := map[string]bool{}

	for hop := 0; hop < maxHops; hop++ {
		if visited[currentURL] {
			return "", "", chain, fmt.Errorf("redirect loop detected")
		}
		visited[currentURL] = true

		html, finalURL, httpChain, err := fetchHTMLWithRedirects(ctx, currentURL, true, maxHops, 8*time.Second)
		if err != nil {
			return "", "", chain, err
		}
		chain = append(chain, httpChain...)
		chain = append(chain, finalURL)

		// Simple wait, simulating initial page script execution window
		if waitMs > 0 {
			time.Sleep(time.Duration(waitMs) * time.Millisecond)
		}

		target := extractClientSideRedirectTarget(html)
		if target == "" {
			return html, finalURL, chain, nil
		}

		resolved, err := resolveRedirectURL(finalURL, target)
		if err != nil {
			return html, finalURL, chain, nil
		}

		currentURL = resolved
	}

	return "", currentURL, chain, fmt.Errorf("max client-side redirect hops reached")
}

func extractClientSideRedirectTarget(html string) string {
	if strings.TrimSpace(html) == "" {
		return ""
	}

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)<meta[^>]+http-equiv\s*=\s*['"]?refresh['"]?[^>]+content\s*=\s*['"][^;"']*;\s*url=([^"'>\s]+)`),
		regexp.MustCompile(`(?is)window\.location\.href\s*=\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?is)window\.location\s*=\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?is)location\.href\s*=\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?is)location\.replace\(\s*['"]([^'"]+)['"]\s*\)`),
	}

	for _, p := range patterns {
		m := p.FindStringSubmatch(html)
		if len(m) > 1 {
			return strings.TrimSpace(m[1])
		}
	}

	return ""
}

func resolveRedirectURL(baseURL, target string) (string, error) {
	baseParsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", err
	}
	targetParsed, err := url.Parse(strings.TrimSpace(target))
	if err != nil {
		return "", err
	}
	return baseParsed.ResolveReference(targetParsed).String(), nil
}

// InjectProxyHandler provides HTML proxy and auto-injection (for proxy mode)
func (g *Gateway) InjectProxyHandler() http.Handler {
	if g.injector == nil || g.config.P3ProxyTarget == "" {
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
