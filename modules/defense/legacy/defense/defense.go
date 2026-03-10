package defense

// Phase 3: This module has completed basic migration, pending deep optimization (see docs/5-process/modularization/PHASE_3_PLAN.md)
import (
	"math"
	"strings"

	"github.com/vistone/fingerprint/modules/core/types"
)

// AnomalyDetector is anomaly fingerprint detector
// Used to analyze suspicious patterns in fingerprint data and determine whether it's a bot or forged fingerprint
type AnomalyDetector struct{}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector() *AnomalyDetector {
	return &AnomalyDetector{}
}

// DetectAnomalies detects anomalies in fingerprint data
// data: fingerprint raw byte data
// return true indicates anomaly detected
func (d *AnomalyDetector) DetectAnomalies(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// check1: entropy too low (repetitive pattern)
	if d.hasLowEntropy(data) {
		return true
	}

	// check2: entropy too high (completely random)
	if d.hasExcessiveEntropy(data) {
		return true
	}

	// check3: known automation tool features
	if d.containsSpoofingMarkers(data) {
		return true
	}

	return false
}

// DetectHeadlessBrowser detects whether User-Agent is headless browser
func (d *AnomalyDetector) DetectHeadlessBrowser(userAgent string) bool {
	uaLower := strings.ToLower(userAgent)
	headlessMarkers := []string{
		"headlesschrome",
		"phantomjs",
		"selenium",
		"webdriver",
		"puppeteer",
		"playwright",
		"cypress",
		"jsdom",
		"zombie",
		"htmlunit",
	}
	for _, marker := range headlessMarkers {
		if strings.Contains(uaLower, marker) {
			return true
		}
	}
	return false
}

// hasLowEntropy checks whether data entropy is too low (suspicious uniform data)
func (d *AnomalyDetector) hasLowEntropy(data []byte) bool {
	if len(data) < 10 {
		return false
	}
	var byteCounts [256]int
	for _, b := range data {
		byteCounts[b]++
	}
	uniqueBytes := 0
	for _, count := range byteCounts {
		if count > 0 {
			uniqueBytes++
		}
	}
	// Less than 26/256 ≈ 10% different byte values is considered suspicious
	return uniqueBytes < 26
}

// hasExcessiveEntropy checks whether data entropy is too high (too random)
func (d *AnomalyDetector) hasExcessiveEntropy(data []byte) bool {
	if len(data) < 20 {
		return false
	}
	var byteCounts [256]int
	for _, b := range data {
		byteCounts[b]++
	}
	// Calculate Shannon entropy
	n := float64(len(data))
	entropy := 0.0
	for _, count := range byteCounts {
		if count > 0 {
			p := float64(count) / n
			entropy -= p * math.Log2(p)
		}
	}
	// Over 7.5 bits is considered too random
	return entropy > 7.5
}

// containsSpoofingMarkers checks whether it contains known automation tool features
func (d *AnomalyDetector) containsSpoofingMarkers(data []byte) bool {
	patterns := [][]byte{
		[]byte("HeadlessChrome"),
		[]byte("PhantomJS"),
		[]byte("webdriver"),
		[]byte("selenium"),
		[]byte("puppeteer"),
	}
	for _, pattern := range patterns {
		if len(pattern) <= len(data) {
			for i := 0; i <= len(data)-len(pattern); i++ {
				match := true
				for j := 0; j < len(pattern); j++ {
					if data[i+j] != pattern[j] {
						match = false
						break
					}
				}
				if match {
					return true
				}
			}
		}
	}
	return false
}

// ContradictionDetector is contradiction fingerprint detector
// Detects logical inconsistencies between fingerprint attributes
type ContradictionDetector struct{}

// NewContradictionDetector creates a new contradiction detector
func NewContradictionDetector() *ContradictionDetector {
	return &ContradictionDetector{}
}

// CheckContradictions checks whether fingerprint attributes have contradictions
// attributes: attribute key-value pair list, e.g. [("os", "Windows"), ("platform", "Linux")]
// return true indicates contradiction found
func (c *ContradictionDetector) CheckContradictions(attributes map[string]string) bool {
	if len(attributes) == 0 {
		return false
	}

	// Check OS and platform contradiction
	if os, ok := attributes["os"]; ok {
		if platform, ok := attributes["platform"]; ok {
			if c.hasOSPlatformContradiction(os, platform) {
				return true
			}
		}
	}

	// Check User-Agent and feature contradiction
	if ua, ok := attributes["user_agent"]; ok {
		if features, ok := attributes["features"]; ok {
			if c.hasUserAgentFeatureContradiction(ua, features) {
				return true
			}
		}
	}

	// Check mobile device and screen resolution contradiction
	if isMobile, ok := attributes["is_mobile"]; ok {
		if screenWidth, ok := attributes["screen_width"]; ok {
			if c.hasMobileScreenContradiction(isMobile, screenWidth) {
				return true
			}
		}
	}

	// Check User-Agent and OS contradiction
	if ua, ok := attributes["user_agent"]; ok {
		if os, ok := attributes["os"]; ok {
			if c.hasUAOSContradiction(ua, os) {
				return true
			}
		}
	}

	return false
}

// hasOSPlatformContradiction checks OS and platform contradiction
func (c *ContradictionDetector) hasOSPlatformContradiction(os, platform string) bool {
	if strings.Contains(os, "Windows") && !strings.Contains(platform, "Win") {
		return true
	}
	if strings.Contains(os, "Mac") && !strings.Contains(platform, "Mac") {
		return true
	}
	if strings.Contains(os, "Linux") && !strings.Contains(platform, "Linux") && !strings.Contains(platform, "X11") {
		return true
	}
	return false
}

// hasUserAgentFeatureContradiction checks User-Agent and feature contradiction
func (c *ContradictionDetector) hasUserAgentFeatureContradiction(userAgent, features string) bool {
	// Old browsers should not support modern features
	if strings.Contains(userAgent, "Chrome/60") && strings.Contains(features, "WebGL2") {
		return true
	}
	// Mobile browsers should not declare desktop features
	if strings.Contains(userAgent, "Mobile") && strings.Contains(features, "desktop") {
		return true
	}
	return false
}

// hasMobileScreenContradiction checks mobile device and screen size contradiction
func (c *ContradictionDetector) hasMobileScreenContradiction(isMobile, screenWidth string) bool {
	width := 0
	_, err := parseIntFallback(screenWidth, &width)
	if err != nil {
		return false
	}
	// Mobile device using desktop resolution is suspicious
	if isMobile == "true" && width > 1920 {
		return true
	}
	// Desktop device using extremely small resolution is suspicious
	if isMobile == "false" && width < 800 {
		return true
	}
	return false
}

// hasUAOSContradiction checks User-Agent and OS contradiction
func (c *ContradictionDetector) hasUAOSContradiction(userAgent, os string) bool {
	uaLower := strings.ToLower(userAgent)
	osLower := strings.ToLower(os)

	// Windows UA claims Mac system
	if strings.Contains(uaLower, "windows") && strings.Contains(osLower, "mac") {
		return true
	}
	// Mac UA claims Windows system
	if strings.Contains(uaLower, "macintosh") && strings.Contains(osLower, "windows") {
		return true
	}
	// Linux UA claims Windows system
	if strings.Contains(uaLower, "x11; linux") && strings.Contains(osLower, "windows") {
		return true
	}
	return false
}

// parseIntFallback simple integer parsing without strconv dependency
func parseIntFallback(s string, result *int) (int, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return 0, &parseError{s: s}
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, &parseError{s: s}
		}
		n = n*10 + int(c-'0')
	}
	*result = n
	return n, nil
}

type parseError struct{ s string }

func (e *parseError) Error() string { return "parse error: " + e.s }

// PassiveRecognizer is passive fingerprint recognizer
// Infers browser and operating system by analyzing HTTP request headers
type PassiveRecognizer struct{}

// NewPassiveRecognizer creates a new passive recognizer
func NewPassiveRecognizer() *PassiveRecognizer {
	return &PassiveRecognizer{}
}

// RecognitionResult represents passive recognition result
type RecognitionResult struct {
	// Detected browser type
	Browser types.BrowserType
	// Detected operating system
	OS types.OperatingSystem
	// Detected browser version
	BrowserVersion string
	// Confidence (0.0-1.0)
	Confidence float64
	// Whether it's a mobile device
	IsMobile bool
	// Whether it's suspected to be a bot
	IsBot bool
}

// RecognizeFromHeaders recognizes browser fingerprint from HTTP request headers
func (r *PassiveRecognizer) RecognizeFromHeaders(headers map[string]string) *RecognitionResult {
	result := &RecognitionResult{}

	ua := headers["User-Agent"]
	if ua == "" {
		result.Confidence = 0.0
		return result
	}

	// Detect bot
	anomalyDetector := NewAnomalyDetector()
	if anomalyDetector.DetectHeadlessBrowser(ua) {
		result.IsBot = true
		result.Confidence = 0.9
		return result
	}

	uaLower := strings.ToLower(ua)

	// Detect mobile device
	result.IsMobile = strings.Contains(uaLower, "mobile") ||
		strings.Contains(uaLower, "android") ||
		strings.Contains(uaLower, "iphone") ||
		strings.Contains(uaLower, "ipad")

	// Recognize browser
	result.Browser, result.BrowserVersion = detectBrowserFromUA(ua)

	// Recognize operating system
	result.OS = detectOSFromUA(ua)

	// Calculate confidence
	result.Confidence = calculateConfidence(headers, result)

	return result
}

// detectBrowserFromUA detects browser type and version from User-Agent
func detectBrowserFromUA(ua string) (types.BrowserType, string) {
	uaLower := strings.ToLower(ua)

	// Edge must be detected before Chrome (Edge UA also contains Chrome)
	if strings.Contains(uaLower, "edg/") || strings.Contains(uaLower, "edge/") {
		version := extractVersionFromUA(ua, "Edg/")
		if version == "" {
			version = extractVersionFromUA(ua, "Edge/")
		}
		return types.BrowserEdge, version
	}

	// Opera must be detected before Chrome
	if strings.Contains(uaLower, "opr/") {
		version := extractVersionFromUA(ua, "OPR/")
		return types.BrowserOpera, version
	}

	// Chrome
	if strings.Contains(uaLower, "chrome/") && !strings.Contains(uaLower, "chromium") {
		version := extractVersionFromUA(ua, "Chrome/")
		return types.BrowserChrome, version
	}

	// Firefox
	if strings.Contains(uaLower, "firefox/") {
		version := extractVersionFromUA(ua, "Firefox/")
		return types.BrowserFirefox, version
	}

	// Safari
	if strings.Contains(uaLower, "safari/") {
		version := extractVersionFromUA(ua, "Version/")
		return types.BrowserSafari, version
	}

	return types.BrowserChrome, ""
}

// extractVersionFromUA extracts version number after specific identifier from User-Agent
func extractVersionFromUA(ua, prefix string) string {
	idx := strings.Index(ua, prefix)
	if idx == -1 {
		return ""
	}
	start := idx + len(prefix)
	end := start
	for end < len(ua) && ua[end] != '.' && ua[end] != ' ' && ua[end] != ';' {
		end++
	}
	if end > start {
		return ua[start:end]
	}
	return ""
}

// detectOSFromUA detects operating system from User-Agent
func detectOSFromUA(ua string) types.OperatingSystem {
	if strings.Contains(ua, "Windows NT 10.0") {
		return types.OSWindows10
	}
	if strings.Contains(ua, "Macintosh; Intel Mac OS X 15") {
		return types.OSMacOS15
	}
	if strings.Contains(ua, "Macintosh; Intel Mac OS X 14") {
		return types.OSMacOS14
	}
	if strings.Contains(ua, "Macintosh; Intel Mac OS X 13") {
		return types.OSMacOS13
	}
	if strings.Contains(ua, "X11; Linux") {
		return types.OSLinux
	}
	return types.OSWindows10 // Default
}

// calculateConfidence calculates recognition confidence
func calculateConfidence(headers map[string]string, result *RecognitionResult) float64 {
	score := 0.5

	// Has User-Agent, add score
	if headers["User-Agent"] != "" {
		score += 0.2
	}

	// Has Accept header, add score
	if headers["Accept"] != "" {
		score += 0.1
	}

	// Has Accept-Language, add score
	if headers["Accept-Language"] != "" {
		score += 0.1
	}

	// Chrome-like browsers have Sec-CH-UA, add score
	if headers["Sec-CH-UA"] != "" {
		score += 0.1
	}

	return math.Min(score, 1.0)
}

// RecognizeFromUserAgent recognizes browser fingerprint only from User-Agent string
func RecognizeFromUserAgent(userAgent string) *RecognitionResult {
	recognizer := NewPassiveRecognizer()
	return recognizer.RecognizeFromHeaders(map[string]string{
		"User-Agent": userAgent,
	})
}
