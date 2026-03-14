package ja4h

import "strings"

func checkMissingCommonHeaders(headers []string) []string {
	commonHeaders := []string{"host", "user-agent", "accept", "accept-language"}
	var missing []string

	for _, common := range commonHeaders {
		if !contains(headers, common) {
			missing = append(missing, common)
		}
	}

	return missing
}

func isSuspiciousHeaderValue(name, value string) bool {
	name = strings.ToLower(name)

	// Check known abnormal values
	switch name {
	case "user-agent":
		// Check known bot keywords
		botKeywords := []string{"bot", "crawler", "spider", "scraper", "headless"}
		for _, kw := range botKeywords {
			if strings.Contains(strings.ToLower(value), kw) {
				return false // This may not necessarily be abnormal, depending on context
			}
		}

	case "accept":
		// Check if missing reasonable MIME type
		if !strings.Contains(value, "text/html") && !strings.Contains(value, "*/*") {
			return true
		}

	case "accept-encoding":
		// Check if empty or contains abnormal encoding
		if value == "" || strings.Contains(value, "identity") {
			return true
		}
	}

	return false
}

func isValidUserAgent(ua string, method string, path string) bool {
	// Simple check: User-Agent should not be empty or too short
	if ua == "" || len(ua) < 10 {
		return false
	}

	// If GET request to certain paths, User-Agent should look like a browser
	if method == "GET" && !strings.HasPrefix(path, "/api") {
		if !strings.Contains(strings.ToLower(ua), "mozilla") &&
			!strings.Contains(strings.ToLower(ua), "opera") &&
			!strings.Contains(strings.ToLower(ua), "curl") {
			return false
		}
	}

	return true
}

func isMethodPathConsistent(method string, path string) bool {
	// GET should not be used for /api/post or /api/delete, etc.
	if method == "GET" && strings.Contains(strings.ToLower(path), "/delete") {
		return false
	}

	// POST/PUT typically used for submitting data
	if (method == "POST" || method == "PUT") && strings.HasSuffix(path, ".ico") {
		return false
	}

	return true
}

func isSuspiciousQueryParams(params map[string]string) bool {
	// Check SQL injection signs
	sqlKeywords := []string{"union", "select", "insert", "delete", "drop", "--", "/*", "*/"}

	for _, v := range params {
		for _, kw := range sqlKeywords {
			if strings.Contains(strings.ToLower(v), kw) {
				return true
			}
		}
	}

	return false
}

// detectClientHintsInconsistencies detects Client Hints inconsistencies
func detectClientHintsInconsistencies(headers []struct {
	Name  string
	Value string
}) []string {
	var inconsistencies []string

	// Extract Client Hints
	secCHUA := getHeaderValue(headers, "sec-ch-ua")
	secCHUAMobile := getHeaderValue(headers, "sec-ch-ua-mobile")
	secCHUAPlatform := getHeaderValue(headers, "sec-ch-ua-platform")
	secCHUAArch := getHeaderValue(headers, "sec-ch-ua-arch")
	secCHUABitness := getHeaderValue(headers, "sec-ch-ua-bitness")
	secCHUAPlatformVersion := getHeaderValue(headers, "sec-ch-ua-platform-version")
	userAgent := getHeaderValue(headers, "user-agent")

	// 1. Sec-CH-UA present but User-Agent doesn't look like Chromium
	if secCHUA != "" && userAgent != "" {
		if !strings.Contains(userAgent, "Chrome") && !strings.Contains(userAgent, "Chromium") && !strings.Contains(userAgent, "Edge") {
			inconsistencies = append(inconsistencies, "UA_CH_MISMATCH")
		}
	}

	// 2. Claims mobile device but platform is not mobile
	if secCHUAMobile == "?1" {
		platform := strings.ToLower(secCHUAPlatform)
		if platform != "" && platform != `"android"` && platform != `"ios"` {
			inconsistencies = append(inconsistencies, "MOBILE_PLATFORM_MISMATCH")
		}
	}

	// 3. High-entropy hints present but low-entropy hints missing (abnormal)
	hasHighEntropyHints := secCHUAArch != "" || secCHUABitness != "" || secCHUAPlatformVersion != ""
	hasLowEntropyHints := secCHUA != "" || secCHUAMobile != "" || secCHUAPlatform != ""
	if hasHighEntropyHints && !hasLowEntropyHints {
		inconsistencies = append(inconsistencies, "MISSING_LOW_ENTROPY_HINTS")
	}

	// 4. Platform architecture inconsistent with User-Agent
	if secCHUAPlatform != "" && userAgent != "" {
		platform := strings.ToLower(secCHUAPlatform)
		ua := strings.ToLower(userAgent)

		if strings.Contains(platform, "windows") && !strings.Contains(ua, "windows") {
			inconsistencies = append(inconsistencies, "PLATFORM_UA_MISMATCH")
		}
		if strings.Contains(platform, "macos") && !strings.Contains(ua, "macintosh") {
			inconsistencies = append(inconsistencies, "PLATFORM_UA_MISMATCH")
		}
		if strings.Contains(platform, "linux") && !strings.Contains(ua, "linux") {
			inconsistencies = append(inconsistencies, "PLATFORM_UA_MISMATCH")
		}
	}

	// 5. Architecture bitness inconsistent
	if secCHUABitness != "" && userAgent != "" {
		bitness := strings.Trim(secCHUABitness, `"`)
		ua := strings.ToLower(userAgent)

		if bitness == "64" && !strings.Contains(ua, "64") && !strings.Contains(ua, "x86_64") && !strings.Contains(ua, "win64") {
			inconsistencies = append(inconsistencies, "BITNESS_MISMATCH")
		}
	}

	return inconsistencies
}

// AnalyzeClientHints analyzes Client Hints in request
func (a *JA4HAnalyzer) AnalyzeClientHints(req HTTP2RequestData) *ClientHintsAnalysis {
	analysis := &ClientHintsAnalysis{
		LowEntropyHints:  make(map[string]string),
		HighEntropyHints: make(map[string]string),
		Inconsistencies:  []string{},
	}

	// Extract low-entropy hints
	for _, h := range req.Headers {
		name := strings.ToLower(h.Name)
		switch name {
		case "sec-ch-ua":
			analysis.LowEntropyHints["Sec-CH-UA"] = h.Value
			analysis.HasLowEntropyHints = true
		case "sec-ch-ua-mobile":
			analysis.LowEntropyHints["Sec-CH-UA-Mobile"] = h.Value
			analysis.HasLowEntropyHints = true
		case "sec-ch-ua-platform":
			analysis.LowEntropyHints["Sec-CH-UA-Platform"] = h.Value
			analysis.HasLowEntropyHints = true
		}
	}

	// Extract high-entropy hints
	for _, h := range req.Headers {
		name := strings.ToLower(h.Name)
		switch name {
		case "sec-ch-ua-arch":
			analysis.HighEntropyHints["Sec-CH-UA-Arch"] = h.Value
			analysis.HasHighEntropyHints = true
		case "sec-ch-ua-bitness":
			analysis.HighEntropyHints["Sec-CH-UA-Bitness"] = h.Value
			analysis.HasHighEntropyHints = true
		case "sec-ch-ua-full-version-list":
			analysis.HighEntropyHints["Sec-CH-UA-Full-Version-List"] = h.Value
			analysis.HasHighEntropyHints = true
		case "sec-ch-ua-platform-version":
			analysis.HighEntropyHints["Sec-CH-UA-Platform-Version"] = h.Value
			analysis.HasHighEntropyHints = true
		case "sec-ch-ua-model":
			analysis.HighEntropyHints["Sec-CH-UA-Model"] = h.Value
			analysis.HasHighEntropyHints = true
		case "sec-ch-ua-wow64":
			analysis.HighEntropyHints["Sec-CH-UA-WoW64"] = h.Value
			analysis.HasHighEntropyHints = true
		}
	}

	// Detect inconsistencies
	analysis.Inconsistencies = detectClientHintsInconsistencies(req.Headers)

	// Identify browser type
	if secCHUA, ok := analysis.LowEntropyHints["Sec-CH-UA"]; ok {
		ua := strings.ToLower(secCHUA)
		if strings.Contains(ua, "chrome") {
			analysis.BrowserType = "Chrome"
		} else if strings.Contains(ua, "edge") {
			analysis.BrowserType = "Edge"
		} else if strings.Contains(ua, "chromium") {
			analysis.BrowserType = "Chromium"
		}
	}

	return analysis
}

// ClientHintsAnalysis Client Hints analysis result
type ClientHintsAnalysis struct {
	HasLowEntropyHints  bool
	HasHighEntropyHints bool
	LowEntropyHints     map[string]string
	HighEntropyHints    map[string]string
	Inconsistencies     []string
	BrowserType         string
}

// ============ Helper functions ============

func hasHeader(headers []struct {
	Name  string
	Value string
}, name string) bool {
	name = strings.ToLower(name)
	for _, h := range headers {
		if strings.ToLower(h.Name) == name {
			return true
		}
	}
	return false
}

func getHeaderValue(headers []struct {
	Name  string
	Value string
}, name string) string {
	name = strings.ToLower(name)
	for _, h := range headers {
		if strings.ToLower(h.Name) == name {
			return h.Value
		}
	}
	return ""
}

func contains(slice []string, item string) bool {
	item = strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == item {
			return true
		}
	}
	return false
}

// ============ Known configuration library ============

func initKnownBrowserProfiles() map[string]*HTTP2BrowserProfile {
	return map[string]*HTTP2BrowserProfile{
		"chrome_windows": {
			Name:           "Chrome on Windows",
			BrowserName:    "Chrome",
			BrowserVersion: "120+",
			HeaderOrder: []string{
				"host", "user-agent", "accept", "accept-language",
				"accept-encoding", "connection", "sec-fetch-dest",
				"sec-fetch-mode", "sec-fetch-site", "upgrade-insecure-requests",
			},
			TypicalHeaders: map[string]string{
				"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				"accept-language": "en-US,en;q=0.5",
				"accept-encoding": "gzip, deflate",
			},
			HeaderStrategy: "standard",
			RiskScore:      0.05,
		},
		"firefox_windows": {
			Name:           "Firefox on Windows",
			BrowserName:    "Firefox",
			BrowserVersion: "121+",
			HeaderOrder: []string{
				"host", "user-agent", "accept", "accept-language",
				"accept-encoding", "connection",
			},
			TypicalHeaders: map[string]string{
				"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				"accept-language": "en-US,en;q=0.5",
				"accept-encoding": "gzip, deflate",
			},
			HeaderStrategy: "standard",
			RiskScore:      0.08,
		},
		"safari_macos": {
			Name:           "Safari on macOS",
			BrowserName:    "Safari",
			BrowserVersion: "17+",
			HeaderOrder: []string{
				"host", "user-agent", "accept", "accept-language",
				"accept-encoding", "connection",
			},
			TypicalHeaders: map[string]string{
				"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				"accept-language": "en-US,en;q=0.5",
				"accept-encoding": "gzip, deflate, br",
			},
			HeaderStrategy: "standard",
			RiskScore:      0.06,
		},
	}
}

// ComputeJA4H convenience function: calculates JA4H
func ComputeJA4H(req HTTP2RequestData) (*JA4HResult, error) {
	analyzer := NewJA4HAnalyzer()
	return analyzer.AnalyzeHTTPRequest(req)
}
