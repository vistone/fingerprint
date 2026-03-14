package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

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
					headlessFetchOptions{
						TargetURL:    request.URL,
						BrowserWS:    g.config.ScannerBrowserWS,
						MaxRedirects: maxRedirects,
						WaitMs:       waitMs,
						Timeout:      browserTimeout,
					},
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

// InjectProxyHandler provides HTML proxy and auto-injection (for proxy mode)
func (g *Gateway) InjectProxyHandler() http.Handler {
	if g.injector == nil || g.config.AntiDetectProxyTarget == "" {
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
