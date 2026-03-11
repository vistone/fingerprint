// Fingerprint Kit Utilities
// consistency.go - Cross-layer fingerprint consistency validation
package utils

import (
	"fmt"
	"strings"

	"github.com/vistone/fingerprint/modules/profiles"
)

// ConsistencyValidator is a cross-layer consistency validator
type ConsistencyValidator struct {
	profile *profiles.ClientProfile
}

// NewConsistencyValidator creates a new consistency validator
func NewConsistencyValidator(profile *profiles.ClientProfile) *ConsistencyValidator {
	return &ConsistencyValidator{
		profile: profile,
	}
}

// ConsistencyReport is a cross-layer consistency validation report
type ConsistencyReport struct {
	// Overall results
	IsConsistent bool
	Score        float64 // 0.0 - 1.0, 1.0 means fully consistent

	// Per-layer check results
	HTTPLayer   *LayerCheckResult
	ClientHints *LayerCheckResult
	JSLayer     *LayerCheckResult
	TCPIPLayer  *LayerCheckResult

	// Details
	Mismatches []string
	Warnings   []string
	Details    map[string]interface{}
}

// LayerCheckResult is the check result for a single layer
type LayerCheckResult struct {
	LayerName    string
	IsConsistent bool
	Data         map[string]string
	Issues       []string
}

// Validate performs cross-layer consistency validation
func (cv *ConsistencyValidator) Validate() *ConsistencyReport {
	report := &ConsistencyReport{
		Details:    make(map[string]interface{}),
		Mismatches: []string{},
		Warnings:   []string{},
	}

	// 1. Validate HTTP layer
	report.HTTPLayer = cv.validateHTTPLayer()

	// 2. Validate Client Hints layer
	report.ClientHints = cv.validateClientHintsLayer()

	// 3. Validate JavaScript layer
	report.JSLayer = cv.validateJSLayer()

	// 4. Validate TCP/IP layer
	report.TCPIPLayer = cv.validateTCPIPLayer()

	// 5. Cross-validate consistency between layers
	cv.crossLayerValidation(report)

	// 6. Calculate overall score
	cv.calculateScore(report)

	return report
}

// validateHTTPLayer validates HTTP layer consistency
func (cv *ConsistencyValidator) validateHTTPLayer() *LayerCheckResult {
	result := &LayerCheckResult{
		LayerName: "HTTP",
		Data:      make(map[string]string),
	}

	if cv.profile.Headers == nil {
		result.Issues = append(result.Issues, "Headers configuration missing")
		return result
	}

	headers := cv.profile.Headers

	// Check required fields
	result.Data["User-Agent"] = headers.UserAgent
	result.Data["Accept-Language"] = headers.AcceptLanguage
	result.Data["Sec-CH-UA"] = headers.SecCHUA
	result.Data["Sec-CH-UA-Mobile"] = headers.SecCHUAMobile
	result.Data["Sec-CH-UA-Platform"] = headers.SecCHUAPlatform

	// Validate User-Agent
	if headers.UserAgent == "" {
		result.Issues = append(result.Issues, "User-Agent is empty")
	} else if !cv.isValidUserAgent(headers.UserAgent) {
		result.Issues = append(result.Issues, "User-Agent format is invalid")
	}

	// Validate Accept-Language
	if headers.AcceptLanguage != "" {
		if !cv.isValidLanguageTag(headers.AcceptLanguage) {
			result.Issues = append(result.Issues, "Accept-Language format is invalid")
		}
	}

	result.IsConsistent = len(result.Issues) == 0

	return result
}

// validateClientHintsLayer validates Client Hints layer consistency
func (cv *ConsistencyValidator) validateClientHintsLayer() *LayerCheckResult {
	result := &LayerCheckResult{
		LayerName: "ClientHints",
		Data:      make(map[string]string),
	}

	if cv.profile.Headers == nil {
		result.Issues = append(result.Issues, "Headers configuration missing")
		return result
	}

	headers := cv.profile.Headers

	// Sec-CH-UA format validation
	result.Data["Sec-CH-UA"] = headers.SecCHUA
	if headers.SecCHUA != "" {
		if !cv.isValidSecCHUA(headers.SecCHUA) {
			result.Issues = append(result.Issues, "Sec-CH-UA format is invalid")
		}
	}

	// Sec-CH-UA-Mobile validation
	result.Data["Sec-CH-UA-Mobile"] = headers.SecCHUAMobile
	if headers.SecCHUAMobile != "" && headers.SecCHUAMobile != "true" && headers.SecCHUAMobile != "false" {
		result.Issues = append(result.Issues, "Sec-CH-UA-Mobile value is invalid")
	}

	// Sec-CH-UA-Platform validation
	result.Data["Sec-CH-UA-Platform"] = headers.SecCHUAPlatform
	validPlatforms := map[string]bool{
		"Windows":   true,
		"macOS":     true,
		"Linux":     true,
		"Android":   true,
		"iOS":       true,
		"Chrome OS": true,
	}
	if headers.SecCHUAPlatform != "" && !validPlatforms[strings.Trim(headers.SecCHUAPlatform, `"`)] {
		result.Issues = append(result.Issues, fmt.Sprintf("Sec-CH-UA-Platform invalid: %s", headers.SecCHUAPlatform))
	}

	result.IsConsistent = len(result.Issues) == 0

	return result
}

// validateJSLayer validates JavaScript layer consistency
func (cv *ConsistencyValidator) validateJSLayer() *LayerCheckResult {
	result := &LayerCheckResult{
		LayerName: "JavaScript",
		Data:      make(map[string]string),
	}

	// Check JSAntiDetection configuration
	if cv.profile.JSAntiDetection == nil {
		result.Issues = append(result.Issues, "JSAntiDetection configuration missing")
		return result
	}

	antiDetect := cv.profile.JSAntiDetection

	// Validate each anti-detection point configuration
	if antiDetect.WebGPU != nil {
		result.Data["WebGPU.Available"] = fmt.Sprintf("%v", antiDetect.WebGPU.Available)
		if antiDetect.WebGPU.Available && antiDetect.WebGPU.AdapterName == "" {
			result.Issues = append(result.Issues, "WebGPU enabled but AdapterName is empty")
		}
	}

	if antiDetect.MediaDevices != nil {
		result.Data["MediaDevices.Devices"] = fmt.Sprintf("%d", len(antiDetect.MediaDevices.VideoInputs)+len(antiDetect.MediaDevices.AudioInputs))
		if (len(antiDetect.MediaDevices.VideoInputs) > 0 || len(antiDetect.MediaDevices.AudioInputs) > 0) &&
			len(antiDetect.MediaDevices.VideoInputs) == 0 && len(antiDetect.MediaDevices.AudioInputs) == 0 {
			result.Issues = append(result.Issues, "MediaDevices configuration incomplete")
		}
	}

	if antiDetect.Permissions != nil {
		result.Data["Permissions.States"] = fmt.Sprintf("%d", len(antiDetect.Permissions.PermissionState))
	}

	if antiDetect.Automation != nil {
		result.Data["Automation.WebDriver"] = fmt.Sprintf("%v", antiDetect.Automation.WebDriver)
		result.Data["Automation.Headless"] = fmt.Sprintf("%v", antiDetect.Automation.Headless)
	}

	result.IsConsistent = len(result.Issues) == 0

	return result
}

// validateTCPIPLayer validates TCP/IP layer consistency
func (cv *ConsistencyValidator) validateTCPIPLayer() *LayerCheckResult {
	result := &LayerCheckResult{
		LayerName: "TCP/IP",
		Data:      make(map[string]string),
	}

	if cv.profile.TCPIP == nil {
		result.Issues = append(result.Issues, "TCPIP configuration missing")
		return result
	}

	tcpip := cv.profile.TCPIP

	// Validate TCP/IP parameter validity
	result.Data["TTL"] = fmt.Sprintf("%d", tcpip.TTL)
	result.Data["WindowSize"] = fmt.Sprintf("%d", tcpip.WindowSize)
	result.Data["MSS"] = fmt.Sprintf("%d", tcpip.MSS)
	result.Data["OS"] = string(cv.profile.OS)

	// TTL value validation
	if tcpip.TTL == 0 {
		result.Issues = append(result.Issues, "TTL value is 0")
	}

	// Windows should have TTL 128
	if strings.Contains(string(cv.profile.OS), "Windows") && tcpip.TTL != 128 {
		result.Issues = append(result.Issues, fmt.Sprintf("Windows expects TTL 128, got %d", tcpip.TTL))
	}

	// Linux/Mac should have TTL 64
	if (strings.Contains(string(cv.profile.OS), "Linux") || strings.Contains(string(cv.profile.OS), "Mac OS")) &&
		tcpip.TTL != 64 {
		result.Issues = append(result.Issues, fmt.Sprintf("Unix-like system expects TTL 64, got %d", tcpip.TTL))
	}

	// JA4T fingerprint validity validation
	if tcpip.JA4T == "" {
		result.Issues = append(result.Issues, "JA4T fingerprint is empty")
	}

	result.IsConsistent = len(result.Issues) == 0

	return result
}

// crossLayerValidation performs cross-layer cross-validation
func (cv *ConsistencyValidator) crossLayerValidation(report *ConsistencyReport) {
	if cv.profile.Headers == nil {
		return
	}

	headers := cv.profile.Headers
	ua := headers.UserAgent

	// 1. Browser info in UA vs BrowserType
	if !cv.isUAConsistentWithBrowser(ua) {
		report.Mismatches = append(report.Mismatches, "User-Agent inconsistent with BrowserType")
	}

	// 2. OS info in UA vs OS field
	if !cv.isUAConsistentWithOS(ua) {
		report.Mismatches = append(report.Mismatches, "User-Agent inconsistent with OS info")
	}

	// 3. Sec-CH-UA consistency with UA
	if !cv.isSecCHUAConsistentWithUA(ua, headers.SecCHUA) {
		report.Mismatches = append(report.Mismatches, "Sec-CH-UA inconsistent with User-Agent")
	}

	// 4. Language information consistency
	if !cv.isLanguageConsistent() {
		report.Warnings = append(report.Warnings, "Language information may be inconsistent")
	}

	// 5. TCP/IP consistency with OS
	if cv.profile.TCPIP != nil {
		if !cv.isTCPIPConsistentWithOS() {
			report.Mismatches = append(report.Mismatches, "TCP/IP configuration inconsistent with OS")
		}
	}
}

// calculateScore calculates the consistency score
func (cv *ConsistencyValidator) calculateScore(report *ConsistencyReport) {
	totalLayers := 4
	consistentLayers := 0

	if report.HTTPLayer.IsConsistent {
		consistentLayers++
	}
	if report.ClientHints.IsConsistent {
		consistentLayers++
	}
	if report.JSLayer.IsConsistent {
		consistentLayers++
	}
	if report.TCPIPLayer.IsConsistent {
		consistentLayers++
	}

	// Base score
	baseiScore := float64(consistentLayers) / float64(totalLayers)

	// Adjust based on mismatch count
	mismatchPenalty := float64(len(report.Mismatches)) * 0.1
	warningPenalty := float64(len(report.Warnings)) * 0.05

	report.Score = baseiScore - mismatchPenalty - warningPenalty
	if report.Score < 0 {
		report.Score = 0
	}

	// Determine overall consistency
	report.IsConsistent = report.Score >= 0.8 && len(report.Mismatches) == 0
}

// Helper methods

func (cv *ConsistencyValidator) isValidUserAgent(ua string) bool {
	// User-Agent should contain browser identifier
	browsers := []string{"Chrome", "Firefox", "Safari", "Edge", "Opera", "Brave"}
	for _, b := range browsers {
		if strings.Contains(ua, b) {
			return true
		}
	}
	return false
}

func (cv *ConsistencyValidator) isValidLanguageTag(lang string) bool {
	// Basic language tag validation (e.g. "en-US", "zh-CN")
	parts := strings.Split(lang, "-")
	if len(parts) >= 1 && len(parts[0]) == 2 {
		return true
	}
	return false
}

func (cv *ConsistencyValidator) isValidSecCHUA(secCHUA string) bool {
	// Sec-CH-UA should follow format: "brand";v="version", ...
	return strings.Contains(secCHUA, "v=") || strings.Contains(secCHUA, "Not")
}

func (cv *ConsistencyValidator) isUAConsistentWithBrowser(ua string) bool {
	browser := string(cv.profile.BrowserType)

	switch browser {
	case "Chrome":
		return strings.Contains(ua, "Chrome") && !strings.Contains(ua, "Chromium")
	case "Firefox":
		return strings.Contains(ua, "Firefox")
	case "Safari":
		return strings.Contains(ua, "Safari") && !strings.Contains(ua, "Chrome")
	case "Edge":
		return strings.Contains(ua, "Edg")
	default:
		return true
	}
}

func (cv *ConsistencyValidator) isUAConsistentWithOS(ua string) bool {
	os := string(cv.profile.OS)

	switch {
	case strings.Contains(os, "Windows"):
		return strings.Contains(ua, "Windows")
	case strings.Contains(os, "Mac OS") || strings.Contains(os, "Macintosh"):
		return strings.Contains(ua, "Macintosh") || strings.Contains(ua, "Mac OS")
	case strings.Contains(os, "Linux"):
		return strings.Contains(ua, "Linux") || strings.Contains(ua, "X11")
	case strings.Contains(os, "Android"):
		return strings.Contains(ua, "Android")
	case strings.Contains(os, "iPhone") || strings.Contains(os, "iPad"):
		return strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad")
	default:
		return true
	}
}

func (cv *ConsistencyValidator) isSecCHUAConsistentWithUA(ua, secCHUA string) bool {
	// Simple consistency check: if UA contains a brand, Sec-CH-UA should too
	if strings.Contains(ua, "Chrome") {
		return strings.Contains(secCHUA, "Chrome") || strings.Contains(secCHUA, "Chromium")
	}
	if strings.Contains(ua, "Firefox") {
		return strings.Contains(secCHUA, "Firefox")
	}
	if strings.Contains(ua, "Safari") {
		return strings.Contains(secCHUA, "Safari")
	}
	return true
}

func (cv *ConsistencyValidator) isLanguageConsistent() bool {
	if cv.profile.Headers == nil {
		return true
	}

	lang := cv.profile.Headers.AcceptLanguage
	if lang == "" {
		return true
	}

	// Language should be consistent with OS or other clues in UA
	// Only doing basic non-empty check here
	return true
}

func (cv *ConsistencyValidator) isTCPIPConsistentWithOS() bool {
	if cv.profile.TCPIP == nil {
		return true
	}

	os := string(cv.profile.OS)
	ttl := cv.profile.TCPIP.TTL

	if strings.Contains(os, "Windows") && ttl != 128 && ttl != 127 {
		return false
	}
	if (strings.Contains(os, "Linux") || strings.Contains(os, "Mac OS")) && ttl != 64 && ttl != 63 {
		return false
	}

	return true
}
