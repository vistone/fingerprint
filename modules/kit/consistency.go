// Fingerprint Kit Utilities
// translated comment
package utils

import (
	"fmt"
	"strings"

	"github.com/vistone/fingerprint/modules/profiles"
)

// translated comment
type ConsistencyValidator struct {
	profile *profiles.ClientProfile
}

// translated comment
func NewConsistencyValidator(profile *profiles.ClientProfile) *ConsistencyValidator {
	return &ConsistencyValidator{
		profile: profile,
	}
}

// translated comment
type ConsistencyReport struct {
	// translated comment
	IsConsistent bool
	Score        float64 // translated comment

	// translated comment
	HTTPLayer   *LayerCheckResult
	ClientHints *LayerCheckResult
	JSLayer     *LayerCheckResult
	TCPIPLayer  *LayerCheckResult

	// translated comment
	Mismatches []string
	Warnings   []string
	Details    map[string]interface{}
}

// translated comment
type LayerCheckResult struct {
	LayerName    string
	IsConsistent bool
	Data         map[string]string
	Issues       []string
}

// translated comment
func (cv *ConsistencyValidator) Validate() *ConsistencyReport {
	report := &ConsistencyReport{
		Details:    make(map[string]interface{}),
		Mismatches: []string{},
		Warnings:   []string{},
	}

	// translated comment
	report.HTTPLayer = cv.validateHTTPLayer()

	// translated comment
	report.ClientHints = cv.validateClientHintsLayer()

	// translated comment
	report.JSLayer = cv.validateJSLayer()

	// translated comment
	report.TCPIPLayer = cv.validateTCPIPLayer()

	// translated comment
	cv.crossLayerValidation(report)

	// translated comment
	cv.calculateScore(report)

	return report
}

// translated comment
func (cv *ConsistencyValidator) validateHTTPLayer() *LayerCheckResult {
	result := &LayerCheckResult{
		LayerName: "HTTP",
		Data:      make(map[string]string),
	}

	if cv.profile.Headers == nil {
		result.Issues = append(result.Issues, "Headers 配置缺失")
		return result
	}

	headers := cv.profile.Headers

	// translated comment
	result.Data["User-Agent"] = headers.UserAgent
	result.Data["Accept-Language"] = headers.AcceptLanguage
	result.Data["Sec-CH-UA"] = headers.SecCHUA
	result.Data["Sec-CH-UA-Mobile"] = headers.SecCHUAMobile
	result.Data["Sec-CH-UA-Platform"] = headers.SecCHUAPlatform

	// translated comment
	if headers.UserAgent == "" {
		result.Issues = append(result.Issues, "User-Agent 为空")
	} else if !cv.isValidUserAgent(headers.UserAgent) {
		result.Issues = append(result.Issues, "User-Agent 格式无效")
	}

	// translated comment
	if headers.AcceptLanguage != "" {
		if !cv.isValidLanguageTag(headers.AcceptLanguage) {
			result.Issues = append(result.Issues, "Accept-Language 格式无效")
		}
	}

	result.IsConsistent = len(result.Issues) == 0

	return result
}

// translated comment
func (cv *ConsistencyValidator) validateClientHintsLayer() *LayerCheckResult {
	result := &LayerCheckResult{
		LayerName: "ClientHints",
		Data:      make(map[string]string),
	}

	if cv.profile.Headers == nil {
		result.Issues = append(result.Issues, "Headers 配置缺失")
		return result
	}

	headers := cv.profile.Headers

	// translated comment
	result.Data["Sec-CH-UA"] = headers.SecCHUA
	if headers.SecCHUA != "" {
		if !cv.isValidSecCHUA(headers.SecCHUA) {
			result.Issues = append(result.Issues, "Sec-CH-UA 格式无效")
		}
	}

	// translated comment
	result.Data["Sec-CH-UA-Mobile"] = headers.SecCHUAMobile
	if headers.SecCHUAMobile != "" && headers.SecCHUAMobile != "true" && headers.SecCHUAMobile != "false" {
		result.Issues = append(result.Issues, "Sec-CH-UA-Mobile 值无效")
	}

	// translated comment
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
		result.Issues = append(result.Issues, fmt.Sprintf("Sec-CH-UA-Platform 无效: %s", headers.SecCHUAPlatform))
	}

	result.IsConsistent = len(result.Issues) == 0

	return result
}

// translated comment
func (cv *ConsistencyValidator) validateJSLayer() *LayerCheckResult {
	result := &LayerCheckResult{
		LayerName: "JavaScript",
		Data:      make(map[string]string),
	}

	// translated comment
	if cv.profile.JSAntiDetection == nil {
		result.Issues = append(result.Issues, "JSAntiDetection 配置缺失")
		return result
	}

	antiDetect := cv.profile.JSAntiDetection

	// translated comment
	if antiDetect.WebGPU != nil {
		result.Data["WebGPU.Available"] = fmt.Sprintf("%v", antiDetect.WebGPU.Available)
		if antiDetect.WebGPU.Available && antiDetect.WebGPU.AdapterName == "" {
			result.Issues = append(result.Issues, "WebGPU 启用但 AdapterName 为空")
		}
	}

	if antiDetect.MediaDevices != nil {
		result.Data["MediaDevices.Devices"] = fmt.Sprintf("%d", len(antiDetect.MediaDevices.VideoInputs)+len(antiDetect.MediaDevices.AudioInputs))
		if (len(antiDetect.MediaDevices.VideoInputs) > 0 || len(antiDetect.MediaDevices.AudioInputs) > 0) &&
			len(antiDetect.MediaDevices.VideoInputs) == 0 && len(antiDetect.MediaDevices.AudioInputs) == 0 {
			result.Issues = append(result.Issues, "MediaDevices 配置不完整")
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

// translated comment
func (cv *ConsistencyValidator) validateTCPIPLayer() *LayerCheckResult {
	result := &LayerCheckResult{
		LayerName: "TCP/IP",
		Data:      make(map[string]string),
	}

	if cv.profile.TCPIP == nil {
		result.Issues = append(result.Issues, "TCPIP 配置缺失")
		return result
	}

	tcpip := cv.profile.TCPIP

	// translated comment
	result.Data["TTL"] = fmt.Sprintf("%d", tcpip.TTL)
	result.Data["WindowSize"] = fmt.Sprintf("%d", tcpip.WindowSize)
	result.Data["MSS"] = fmt.Sprintf("%d", tcpip.MSS)
	result.Data["OS"] = string(cv.profile.OS)

	// translated comment
	if tcpip.TTL == 0 {
		result.Issues = append(result.Issues, "TTL 值为 0")
	}

	// translated comment
	if strings.Contains(string(cv.profile.OS), "Windows") && tcpip.TTL != 128 {
		result.Issues = append(result.Issues, fmt.Sprintf("Windows 期望 TTL 128，实际 %d", tcpip.TTL))
	}

	// translated comment
	if (strings.Contains(string(cv.profile.OS), "Linux") || strings.Contains(string(cv.profile.OS), "Mac OS")) &&
		tcpip.TTL != 64 {
		result.Issues = append(result.Issues, fmt.Sprintf("Unix 类系统期望 TTL 64，实际 %d", tcpip.TTL))
	}

	// translated comment
	if tcpip.JA4T == "" {
		result.Issues = append(result.Issues, "JA4T 指纹为空")
	}

	result.IsConsistent = len(result.Issues) == 0

	return result
}

// translated comment
func (cv *ConsistencyValidator) crossLayerValidation(report *ConsistencyReport) {
	if cv.profile.Headers == nil {
		return
	}

	headers := cv.profile.Headers
	ua := headers.UserAgent

	// translated comment
	if !cv.isUAConsistentWithBrowser(ua) {
		report.Mismatches = append(report.Mismatches, "User-Agent 与 BrowserType 不一致")
	}

	// translated comment
	if !cv.isUAConsistentWithOS(ua) {
		report.Mismatches = append(report.Mismatches, "User-Agent 与 OS 信息不一致")
	}

	// translated comment
	if !cv.isSecCHUAConsistentWithUA(ua, headers.SecCHUA) {
		report.Mismatches = append(report.Mismatches, "Sec-CH-UA 与 User-Agent 不一致")
	}

	// translated comment
	if !cv.isLanguageConsistent() {
		report.Warnings = append(report.Warnings, "语言信息可能不一致")
	}

	// translated comment
	if cv.profile.TCPIP != nil {
		if !cv.isTCPIPConsistentWithOS() {
			report.Mismatches = append(report.Mismatches, "TCP/IP 配置与 OS 不一致")
		}
	}
}

// translated comment
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

	// translated comment
	baseiScore := float64(consistentLayers) / float64(totalLayers)

	// translated comment
	mismatchPenalty := float64(len(report.Mismatches)) * 0.1
	warningPenalty := float64(len(report.Warnings)) * 0.05

	report.Score = baseiScore - mismatchPenalty - warningPenalty
	if report.Score < 0 {
		report.Score = 0
	}

	// translated comment
	report.IsConsistent = report.Score >= 0.8 && len(report.Mismatches) == 0
}

// translated comment

func (cv *ConsistencyValidator) isValidUserAgent(ua string) bool {
	// translated comment
	browsers := []string{"Chrome", "Firefox", "Safari", "Edge", "Opera", "Brave"}
	for _, b := range browsers {
		if strings.Contains(ua, b) {
			return true
		}
	}
	return false
}

func (cv *ConsistencyValidator) isValidLanguageTag(lang string) bool {
	// translated comment
	parts := strings.Split(lang, "-")
	if len(parts) >= 1 && len(parts[0]) == 2 {
		return true
	}
	return false
}

func (cv *ConsistencyValidator) isValidSecCHUA(secCHUA string) bool {
	// translated comment
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
	// translated comment
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

	// translated comment
	// translated comment
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
