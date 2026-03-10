package websocket

import (
	"fmt"
	"net/http"
	"strings"
)

// AnomalyType anomaly type
type AnomalyType string

const (
	// AnomalyInvalidMethod invalid HTTP method
	AnomalyInvalidMethod AnomalyType = "invalid_method"
	// AnomalyMissingHeaders missing required headers
	AnomalyMissingHeaders AnomalyType = "missing_headers"
	// AnomalySuspiciousKey suspicious WebSocket Key
	AnomalySuspiciousKey AnomalyType = "suspicious_key"
	// AnomalyLowEntropyKey low entropy Key (possibly pseudo-random)
	AnomalyLowEntropyKey AnomalyType = "low_entropy_key"
	// AnomalyAbnormalHeaderOrder abnormal header order
	AnomalyAbnormalHeaderOrder AnomalyType = "abnormal_header_order"
	// AnomalySuspiciousExtensions suspicious extensions
	AnomalySuspiciousExtensions AnomalyType = "suspicious_extensions"
	// AnomalyKnownBotSignature known bot signature
	AnomalyKnownBotSignature AnomalyType = "known_bot_signature"
	// AnomalyFrameAnomaly frame anomaly
	AnomalyFrameAnomaly AnomalyType = "frame_anomaly"
	// AnomalyVersionMismatch version mismatch
	AnomalyVersionMismatch AnomalyType = "version_mismatch"
)

// Anomaly detailed anomaly information from detection result
type Anomaly struct {
	Type        AnomalyType
	Description string
	Severity    Severity
	Evidence    map[string]interface{}
}

// Severity severity level
type Severity string

const (
	// SeverityInfo info level
	SeverityInfo Severity = "info"
	// SeverityLow low level
	SeverityLow Severity = "low"
	// SeverityMedium medium level
	SeverityMedium Severity = "medium"
	// SeverityHigh high level
	SeverityHigh Severity = "high"
	// SeverityCritical critical level
	SeverityCritical Severity = "critical"
)

// Detector WebSocket anomaly detector
type Detector struct {
	// Known bot User-Agent patterns
	botPatterns []string
	// Known suspicious Key patterns
	suspiciousKeyPatterns [][]byte
	// Normal header orders (for comparison)
	normalHeaderOrders map[string][]string
}

// DetectionResult detection result
type DetectionResult struct {
	// Whether anomaly detected
	HasAnomaly bool
	// Anomaly list
	Anomalies []Anomaly
	// Risk score (0-100)
	RiskScore int
	// Detected browser type (if any)
	DetectedBrowser string
	// Whether known automation tool
	IsKnownBot bool
}

// NewDetector creates new detector
func NewDetector() *Detector {
	return &Detector{
		botPatterns: []string{
			"bot", "crawler", "spider", "scraper",
			"automation", "headless", "puppeteer",
			"selenium", "playwright", "cypress",
		},
		suspiciousKeyPatterns: [][]byte{
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},        // All zeros
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},        // All zeros
			{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, // Incremental
			{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, // Incremental
		},
		normalHeaderOrders: map[string][]string{
			"chrome": {
				"Host", "Connection", "Pragma", "Cache-Control",
				"User-Agent", "Upgrade", "Origin",
				"Sec-WebSocket-Version", "Sec-WebSocket-Key",
				"Sec-WebSocket-Extensions", "Sec-WebSocket-Protocol",
			},
			"firefox": {
				"Host", "User-Agent", "Accept", "Accept-Language",
				"Accept-Encoding", "Sec-WebSocket-Version",
				"Origin", "Sec-WebSocket-Extensions",
				"Sec-WebSocket-Key", "Connection", "Upgrade",
			},
			"safari": {
				"Host", "Upgrade", "Connection",
				"Sec-WebSocket-Key", "Sec-WebSocket-Version",
				"Origin", "Sec-WebSocket-Extensions",
				"User-Agent", "Accept", "Accept-Language",
			},
		},
	}
}

// Detect detects anomalies in WebSocket request
func (d *Detector) Detect(req *http.Request, fp *WebSocketFingerprint) *DetectionResult {
	result := &DetectionResult{
		Anomalies: make([]Anomaly, 0),
	}

	// 1. Validate HTTP method
	d.checkMethod(req, result)

	// 2. Check required headers
	// 2. Check required headers
	d.checkRequiredHeaders(req, result)

	// 3. Analyze WebSocket Key
	// 3. Analyze WebSocket Key
	d.analyzeKey(fp, result)

	// 4. Analyze header order
	// 4. Analyze header order
	d.analyzeHeaderOrder(fp, result)

	// 5. Check extensions
	// 5. Check extensions
	d.checkExtensions(fp, result)

	// 6. Detect known bots
	// 6. Detect known bots
	d.detectBot(req, result)

	// 7. Calculate risk score
	// 7. Calculate risk score
	result.RiskScore = d.calculateRiskScore(result)
	result.HasAnomaly = len(result.Anomalies) > 0

	return result
}

// DetectFrameAnomalies detects frame anomalies
func (d *Detector) DetectFrameAnomalies(frame *Frame) *DetectionResult {
	result := &DetectionResult{
		Anomalies: make([]Anomaly, 0),
	}

	// Check RSV bits (should be 0 unless using extensions)
	if frame.RSV1 || frame.RSV2 || frame.RSV3 {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalyFrameAnomaly,
			Description: "RSV bits are set without proper extension negotiation",
			Severity:    SeverityMedium,
			Evidence: map[string]interface{}{
				"rsv1": frame.RSV1,
				"rsv2": frame.RSV2,
				"rsv3": frame.RSV3,
			},
		})
	}

	// Check if server-sent frame is unmasked (server should not mask)
	if !frame.MASK && frame.Opcode != OpCodeClose {
		// This may be client frame, needs further verification
		// Actual detection requires context
	}

	// Check control frame size (control frame payload should not exceed 125 bytes)
	if frame.Opcode >= 0x8 && frame.PayloadLength > 125 {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalyFrameAnomaly,
			Description: "Control frame payload exceeds 125 bytes",
			Severity:    SeverityHigh,
			Evidence: map[string]interface{}{
				"opcode":         frame.Opcode,
				"payload_length": frame.PayloadLength,
			},
		})
	}

	// Check unusually large payload
	if frame.PayloadLength > 10*1024*1024 { // 10MB
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalyFrameAnomaly,
			Description: "Unusually large payload detected",
			Severity:    SeverityMedium,
			Evidence: map[string]interface{}{
				"payload_length": frame.PayloadLength,
			},
		})
	}

	result.RiskScore = d.calculateRiskScore(result)
	result.HasAnomaly = len(result.Anomalies) > 0

	return result
}

// checkMethod checks HTTP method
func (d *Detector) checkMethod(req *http.Request, result *DetectionResult) {
	if req.Method != "GET" {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalyInvalidMethod,
			Description: fmt.Sprintf("Invalid HTTP method: %s (expected GET)", req.Method),
			Severity:    SeverityHigh,
			Evidence: map[string]interface{}{
				"method": req.Method,
			},
		})
	}
}

// checkRequiredHeaders checks required headers
func (d *Detector) checkRequiredHeaders(req *http.Request, result *DetectionResult) {
	requiredHeaders := map[string]string{
		"Upgrade":               "websocket",
		"Connection":            "Upgrade",
		"Sec-Websocket-Key":     "",
		"Sec-Websocket-Version": "13",
	}

	missingHeaders := []string{}
	invalidHeaders := []string{}

	for header, expectedValue := range requiredHeaders {
		value := req.Header.Get(header)
		if value == "" {
			// Try lowercase version
			value = req.Header.Get(strings.ToLower(header))
		}

		if value == "" {
			missingHeaders = append(missingHeaders, header)
		} else if expectedValue != "" && !strings.EqualFold(value, expectedValue) {
			invalidHeaders = append(invalidHeaders, fmt.Sprintf("%s=%s", header, value))
		}
	}

	if len(missingHeaders) > 0 {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalyMissingHeaders,
			Description: fmt.Sprintf("Missing required headers: %v", missingHeaders),
			Severity:    SeverityHigh,
			Evidence: map[string]interface{}{
				"missing_headers": missingHeaders,
			},
		})
	}

	if len(invalidHeaders) > 0 {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalyMissingHeaders,
			Description: fmt.Sprintf("Invalid header values: %v", invalidHeaders),
			Severity:    SeverityMedium,
			Evidence: map[string]interface{}{
				"invalid_headers": invalidHeaders,
			},
		})
	}
}

// analyzeKey analyzes WebSocket Key
func (d *Detector) analyzeKey(fp *WebSocketFingerprint, result *DetectionResult) {
	keyChar := fp.Handshake.SecWebSocketKeyCharacteristics

	// Check Key length
	if keyChar.Length != 24 {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalySuspiciousKey,
			Description: fmt.Sprintf("Invalid Sec-WebSocket-Key length: %d (expected 24)", keyChar.Length),
			Severity:    SeverityHigh,
			Evidence: map[string]interface{}{
				"key_length": keyChar.Length,
			},
		})
	}

	// Check if standard Base64
	if !keyChar.IsStandardBase64 {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalySuspiciousKey,
			Description: "Sec-WebSocket-Key is not valid Base64 encoded 16-byte value",
			Severity:    SeverityMedium,
			Evidence: map[string]interface{}{
				"is_standard_base64": keyChar.IsStandardBase64,
			},
		})
	}

	// Check entropy value
	if keyChar.Entropy < 3.0 {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalyLowEntropyKey,
			Description: fmt.Sprintf("Low entropy in Sec-WebSocket-Key: %.2f (suspicious)", keyChar.Entropy),
			Severity:    SeverityMedium,
			Evidence: map[string]interface{}{
				"entropy":      keyChar.Entropy,
				"pattern_type": keyChar.PatternType,
				"has_pattern":  keyChar.HasPattern,
			},
		})
	}

	// Check known patterns
	if keyChar.HasPattern && keyChar.PatternType != "" {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalySuspiciousKey,
			Description: fmt.Sprintf("Detected pattern in Sec-WebSocket-Key: %s", keyChar.PatternType),
			Severity:    SeverityHigh,
			Evidence: map[string]interface{}{
				"pattern_type": keyChar.PatternType,
			},
		})
	}
}

// analyzeHeaderOrder analyzes header order
func (d *Detector) analyzeHeaderOrder(fp *WebSocketFingerprint, result *DetectionResult) {
	if len(fp.Handshake.HeaderOrder) == 0 {
		return
	}

	// Check User-Agent to identify browser
	browser := d.identifyBrowserFromUA(fp.Handshake.UserAgent)
	if browser == "" {
		return
	}

	normalOrder, exists := d.normalHeaderOrders[browser]
	if !exists {
		return
	}

	// Calculate order match score
	matchScore := d.calculateHeaderOrderMatch(fp.Handshake.HeaderOrder, normalOrder)

	// If match score too low, possibly abnormal client
	if matchScore < 0.3 {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalyAbnormalHeaderOrder,
			Description: fmt.Sprintf("Abnormal header order for %s (match score: %.2f)", browser, matchScore),
			Severity:    SeverityLow,
			Evidence: map[string]interface{}{
				"browser":      browser,
				"match_score":  matchScore,
				"actual_order": fp.Handshake.HeaderOrder,
				"normal_order": normalOrder,
			},
		})
	}
}

// checkExtensions checks extensions
func (d *Detector) checkExtensions(fp *WebSocketFingerprint, result *DetectionResult) {
	// Check if there are known suspicious extensions
	suspiciousExts := []string{"x-unknown-extension", "malicious-ext"}

	for _, ext := range fp.Extensions {
		for _, suspicious := range suspiciousExts {
			if strings.EqualFold(ext, suspicious) {
				result.Anomalies = append(result.Anomalies, Anomaly{
					Type:        AnomalySuspiciousExtensions,
					Description: fmt.Sprintf("Suspicious extension detected: %s", ext),
					Severity:    SeverityHigh,
					Evidence: map[string]interface{}{
						"extension": ext,
					},
				})
			}
		}
	}
}

// detectBot detects bots
func (d *Detector) detectBot(req *http.Request, result *DetectionResult) {
	ua := req.Header.Get("User-Agent")
	if ua == "" {
		return
	}

	uaLower := strings.ToLower(ua)

	for _, pattern := range d.botPatterns {
		if strings.Contains(uaLower, pattern) {
			result.IsKnownBot = true
			result.Anomalies = append(result.Anomalies, Anomaly{
				Type:        AnomalyKnownBotSignature,
				Description: fmt.Sprintf("Detected known bot/automation pattern: %s", pattern),
				Severity:    SeverityHigh,
				Evidence: map[string]interface{}{
					"pattern":    pattern,
					"user_agent": ua,
				},
			})
			break
		}
	}
}

// calculateRiskScore calculates risk score
func (d *Detector) calculateRiskScore(result *DetectionResult) int {
	if len(result.Anomalies) == 0 {
		return 0
	}

	score := 0
	for _, anomaly := range result.Anomalies {
		switch anomaly.Severity {
		case SeverityInfo:
			score += 5
		case SeverityLow:
			score += 15
		case SeverityMedium:
			score += 30
		case SeverityHigh:
			score += 50
		case SeverityCritical:
			score += 100
		}
	}

	// If known bot, increase risk score
	if result.IsKnownBot {
		score += 50
	}

	// Limit to 0-100
	// Limit to 0-100
	if score > 100 {
		score = 100
	}

	return score
}

// identifyBrowserFromUA identifies browser from User-Agent
func (d *Detector) identifyBrowserFromUA(ua string) string {
	uaLower := strings.ToLower(ua)

	if strings.Contains(uaLower, "chrome") && !strings.Contains(uaLower, "edg") {
		return "chrome"
	}
	if strings.Contains(uaLower, "firefox") {
		return "firefox"
	}
	if strings.Contains(uaLower, "safari") && !strings.Contains(uaLower, "chrome") {
		return "safari"
	}
	if strings.Contains(uaLower, "edge") || strings.Contains(uaLower, "edg") {
		return "edge"
	}

	return ""
}

// calculateHeaderOrderMatch calculates header order match score
func (d *Detector) calculateHeaderOrderMatch(actual, normal []string) float64 {
	if len(actual) == 0 || len(normal) == 0 {
		return 0.0
	}

	// Create position mapping, use -1 for non-existent
	normalPos := make(map[string]int)
	for i, h := range normal {
		normalPos[strings.ToLower(h)] = i
	}

	// Calculate matches for first few headers
	matchCount := 0
	checkCount := min(len(actual), len(normal), 5) // Check first 5

	for i := 0; i < checkCount; i++ {
		if i < len(actual) {
			actualLower := strings.ToLower(actual[i])
			// Only count when header exists in normal and position matches
			if pos, exists := normalPos[actualLower]; exists && pos == i {
				matchCount++
			}
		}
	}

	return float64(matchCount) / float64(checkCount)
}

// GetSeverityWeight gets severity weight
func GetSeverityWeight(s Severity) int {
	switch s {
	case SeverityInfo:
		return 1
	case SeverityLow:
		return 2
	case SeverityMedium:
		return 3
	case SeverityHigh:
		return 4
	case SeverityCritical:
		return 5
	default:
		return 0
	}
}

// min returns minimum value
func min(values ...int) int {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values[1:] {
		if v < m {
			m = v
		}
	}
	return m
}
