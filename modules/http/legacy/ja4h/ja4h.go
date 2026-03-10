package ja4h

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// JA4HResult JA4H fingerprint result (HTTP request header fingerprint)
type JA4HResult struct {
	// Complete JA4H SHA256 hash
	Hash string

	// JA4H_a: Basic string (request method + path features + protocol + header count, etc.)
	JA4Ha string

	// JA4H_r: Raw version (request header order)
	JA4Hr string

	// Decomposed signature parts
	HeaderOrderSignature    string // Request header order hash
	HeaderValueSignature    string // Specific header value hash
	QueryParameterSignature string // Query parameter order hash

	// Complete raw signature string
	RawSignature string

	// Anomaly score
	RiskScore float64

	// Anomaly flag list
	AnomalyFlags []string

	// Matched known clients/browsers
	MatchedBrowsers []string
}

// HTTP2RequestData HTTP request information
type HTTP2RequestData struct {
	// Request line
	Method   string // GET, POST, etc.
	Path     string // /path/to/resource
	Protocol string // HTTP/1.1, HTTP/2, HTTP/3

	// Request headers (serialized, maintaining original order)
	Headers []struct {
		Name  string
		Value string
	}

	// Query parameters
	QueryParams map[string]string

	// Metadata
	Metadata map[string]string
}

// JA4HAnalyzer JA4H analyzer
type JA4HAnalyzer struct {
	knownBrowserProfiles map[string]*HTTP2BrowserProfile
}

// HTTP2BrowserProfile known browser configuration
type HTTP2BrowserProfile struct {
	Name           string
	BrowserName    string
	BrowserVersion string
	HeaderOrder    []string // Standard request header order
	TypicalHeaders map[string]string
	HeaderStrategy string // "standard", "randomized", "optimized"
	RiskScore      float64
}

// NewJA4HAnalyzer creates analyzer
func NewJA4HAnalyzer() *JA4HAnalyzer {
	return &JA4HAnalyzer{
		knownBrowserProfiles: initKnownBrowserProfiles(),
	}
}

// AnalyzeHTTPRequest analyzes HTTP request headers
func (a *JA4HAnalyzer) AnalyzeHTTPRequest(req HTTP2RequestData) (*JA4HResult, error) {
	if req.Method == "" {
		return nil, fmt.Errorf("method required")
	}

	result := &JA4HResult{
		AnomalyFlags: make([]string, 0, 8), // Pre-allocate capacity
	}

	// 1. Build basic signature string: method, protocol, path features, header count
	pathSignature := generatePathSignature(req.Path)
	headerCount := len(req.Headers)
	acceptedEncodingPresent := hasHeader(req.Headers, "accept-encoding")

	result.JA4Ha = fmt.Sprintf("%s,%s,%s,%d,%v",
		req.Method,
		req.Protocol,
		pathSignature,
		headerCount,
		acceptedEncodingPresent,
	)

	// 2. Extract request header order
	var headerNames []string
	for _, h := range req.Headers {
		headerNames = append(headerNames, strings.ToLower(h.Name))
	}
	headerOrderStr := strings.Join(headerNames, ",")
	result.HeaderOrderSignature = generateHeaderOrderSignature(headerNames)

	// 3. Generate signature for specific request header values (for fingerprinting)
	result.HeaderValueSignature = generateHeaderValueSignature(req.Headers)

	// 4. Query parameter signature
	if len(req.QueryParams) > 0 {
		result.QueryParameterSignature = generateQueryParamSignature(req.QueryParams)
	}

	// 5. Complete signature string
	result.RawSignature = fmt.Sprintf(
		"ja4h|%s|%s|%s|%s|%s",
		result.JA4Ha,
		headerOrderStr,
		result.HeaderOrderSignature,
		result.HeaderValueSignature,
		result.QueryParameterSignature,
	)

	// Calculate SHA256 hash
	hash := sha256.Sum256([]byte(result.RawSignature))
	result.Hash = hex.EncodeToString(hash[:])

	// Anomaly detection
	a.detectJA4HAnomalies(result, req)

	return result, nil
}

// detectJA4HAnomalies detects JA4H anomalies
func (a *JA4HAnalyzer) detectJA4HAnomalies(result *JA4HResult, req HTTP2RequestData) {
	baseScore := 0.0

	// Anomaly 1: Abnormal request header order
	headerNames := make([]string, len(req.Headers))
	for i, h := range req.Headers {
		headerNames[i] = strings.ToLower(h.Name)
	}

	if !isStandardHeaderOrderJA4H(headerNames) {
		result.AnomalyFlags = append(result.AnomalyFlags, "UNUSUAL_HEADER_ORDER")
		baseScore += 0.2
	}

	// Anomaly 2: Missing common request headers
	missingHeaders := checkMissingCommonHeaders(headerNames)
	if len(missingHeaders) > 2 {
		result.AnomalyFlags = append(result.AnomalyFlags, fmt.Sprintf("MISSING_HEADERS_%d", len(missingHeaders)))
		baseScore += 0.15
	}

	// Anomaly 3: Abnormal request header values
	for _, h := range req.Headers {
		if isSuspiciousHeaderValue(h.Name, h.Value) {
			result.AnomalyFlags = append(result.AnomalyFlags, fmt.Sprintf("SUSPICIOUS_%s", h.Name))
			baseScore += 0.1
		}
	}

	// Anomaly 4: User-Agent mismatch
	if ua := getHeaderValue(req.Headers, "user-agent"); ua != "" {
		if !isValidUserAgent(ua, req.Method, req.Path) {
			result.AnomalyFlags = append(result.AnomalyFlags, "INVALID_USER_AGENT")
			baseScore += 0.2
		}
	}

	// Anomaly 5: Request method and path mismatch
	if !isMethodPathConsistent(req.Method, req.Path) {
		result.AnomalyFlags = append(result.AnomalyFlags, "METHOD_PATH_MISMATCH")
		baseScore += 0.15
	}

	// Anomaly 6: Abnormal query parameters (SQL injection, etc.)
	if len(req.QueryParams) > 0 && isSuspiciousQueryParams(req.QueryParams) {
		result.AnomalyFlags = append(result.AnomalyFlags, "SUSPICIOUS_QUERY_PARAMS")
		baseScore += 0.25
	}

	// Anomaly 7: Client Hints inconsistencies
	chInconsistencies := detectClientHintsInconsistencies(req.Headers)
	for _, inconsistency := range chInconsistencies {
		result.AnomalyFlags = append(result.AnomalyFlags, fmt.Sprintf("CH_%s", inconsistency))
		baseScore += 0.1
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}
	result.RiskScore = baseScore
}

// FindMatchingBrowsers finds matching known browsers
func (a *JA4HAnalyzer) FindMatchingBrowsers(
	result *JA4HResult,
	maxResults int,
) []string {
	var matches []string

	for name, profile := range a.knownBrowserProfiles {
		// Based on risk score and feature similarity
		if profile.RiskScore < result.RiskScore+0.15 {
			matches = append(matches, name)
		}

		if len(matches) >= maxResults {
			break
		}
	}

	result.MatchedBrowsers = matches
	return matches
}

// ============ Helper functions ============

func generatePathSignature(path string) string {
	// Based on path depth and features
	if path == "" || path == "/" {
		return "root"
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	depth := len(parts)

	// Detect common path patterns
	if contains(parts, "api") {
		return fmt.Sprintf("api_d%d", depth)
	}
	if contains(parts, "static") {
		return fmt.Sprintf("static_d%d", depth)
	}
	if contains(parts, "v1") || contains(parts, "v2") {
		return fmt.Sprintf("versioned_d%d", depth)
	}

	return fmt.Sprintf("custom_d%d", depth)
}

func generateHeaderOrderSignature(headers []string) string {
	// Normalize header order and calculate hash
	orderStr := strings.Join(headers, ",")
	hash := sha256.Sum256([]byte(orderStr))
	return hex.EncodeToString(hash[:8])
}

func generateHeaderValueSignature(headers []struct {
	Name  string
	Value string
}) string {
	// Generate signature based on key header values
	var keyValues []string

	keyHeaders := []string{"user-agent", "accept", "accept-language", "accept-encoding"}

	for _, h := range headers {
		if contains(keyHeaders, strings.ToLower(h.Name)) {
			// Calculate simple hash of value
			hash := sha256.Sum256([]byte(h.Value))
			keyValues = append(keyValues, fmt.Sprintf("%s:%s", h.Name, hex.EncodeToString(hash[:4])))
		}
	}

	if len(keyValues) == 0 {
		return "no_key_headers"
	}

	combinedSig := strings.Join(keyValues, "|")
	hash := sha256.Sum256([]byte(combinedSig))
	return hex.EncodeToString(hash[:8])
}

func generateQueryParamSignature(params map[string]string) string {
	// Hash of parameter order and names
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	paramStr := strings.Join(keys, ",")
	hash := sha256.Sum256([]byte(paramStr))
	return hex.EncodeToString(hash[:8])
}

// ============ Validation functions ============

func isStandardHeaderOrderJA4H(headers []string) bool {
	// Standard browser request header order typically:
	// Host, User-Agent, Accept, Accept-Language, Accept-Encoding, ...
	if len(headers) == 0 {
		return false
	}

	// Check if starts with Host (very common)
	if headers[0] != "host" && headers[0] != "connection" {
		// Not necessarily abnormal, some clients may differ
		return true
	}

	// Check User-Agent relative position (usually in first 3)
	uaIdx := -1
	for i, h := range headers {
		if h == "user-agent" {
			uaIdx = i
			break
		}
	}

	if uaIdx > 5 {
		return false // Abnormal: UA is too far back
	}

	return true
}

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
