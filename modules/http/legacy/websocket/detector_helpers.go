package websocket

import (
	"fmt"
	"net/http"
	"strings"
)

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
