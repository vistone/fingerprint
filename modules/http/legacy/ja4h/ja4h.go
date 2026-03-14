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
