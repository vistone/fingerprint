// Fingerprint Kit Utilities
// consistency.go - 跨层指纹一致性校验
package utils

import (
	"fmt"
	"strings"

	"github.com/vistone/fingerprint/modules/profiles"
)

// ConsistencyValidator 跨层一致性校验器
type ConsistencyValidator struct {
	profile *profiles.ClientProfile
}

// NewConsistencyValidator 创建新的一致性校验器
func NewConsistencyValidator(profile *profiles.ClientProfile) *ConsistencyValidator {
	return &ConsistencyValidator{
		profile: profile,
	}
}

// ConsistencyReport 跨层一致性校验报告
type ConsistencyReport struct {
	// 总体结果
	IsConsistent bool
	Score        float64 // 0.0 - 1.0, 1.0 表示完全一致

	// 各层检查结果
	HTTPLayer   *LayerCheckResult
	ClientHints *LayerCheckResult
	JSLayer     *LayerCheckResult
	TCPIPLayer  *LayerCheckResult

	// 详细信息
	Mismatches []string
	Warnings   []string
	Details    map[string]interface{}
}

// LayerCheckResult 单个层的检查结果
type LayerCheckResult struct {
	LayerName    string
	IsConsistent bool
	Data         map[string]string
	Issues       []string
}

// Validate 执行跨层一致性校验
func (cv *ConsistencyValidator) Validate() *ConsistencyReport {
	report := &ConsistencyReport{
		Details:    make(map[string]interface{}),
		Mismatches: []string{},
		Warnings:   []string{},
	}

	// 1. 验证 HTTP 层
	report.HTTPLayer = cv.validateHTTPLayer()

	// 2. 验证 Client Hints 层
	report.ClientHints = cv.validateClientHintsLayer()

	// 3. 验证 JavaScript 层
	report.JSLayer = cv.validateJSLayer()

	// 4. 验证 TCP/IP 层
	report.TCPIPLayer = cv.validateTCPIPLayer()

	// 5. 交叉验证各层之间的一致性
	cv.crossLayerValidation(report)

	// 6. 计算总体得分
	cv.calculateScore(report)

	return report
}

// validateHTTPLayer 验证 HTTP 层一致性
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

	// 检查必要字段
	result.Data["User-Agent"] = headers.UserAgent
	result.Data["Accept-Language"] = headers.AcceptLanguage
	result.Data["Sec-CH-UA"] = headers.SecCHUA
	result.Data["Sec-CH-UA-Mobile"] = headers.SecCHUAMobile
	result.Data["Sec-CH-UA-Platform"] = headers.SecCHUAPlatform

	// 验证 User-Agent 有效性
	if headers.UserAgent == "" {
		result.Issues = append(result.Issues, "User-Agent 为空")
	} else if !cv.isValidUserAgent(headers.UserAgent) {
		result.Issues = append(result.Issues, "User-Agent 格式无效")
	}

	// 验证 Accept-Language 有效性
	if headers.AcceptLanguage != "" {
		if !cv.isValidLanguageTag(headers.AcceptLanguage) {
			result.Issues = append(result.Issues, "Accept-Language 格式无效")
		}
	}

	result.IsConsistent = len(result.Issues) == 0

	return result
}

// validateClientHintsLayer 验证 Client Hints 层一致性
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

	// Sec-CH-UA 格式验证
	result.Data["Sec-CH-UA"] = headers.SecCHUA
	if headers.SecCHUA != "" {
		if !cv.isValidSecCHUA(headers.SecCHUA) {
			result.Issues = append(result.Issues, "Sec-CH-UA 格式无效")
		}
	}

	// Sec-CH-UA-Mobile 验证
	result.Data["Sec-CH-UA-Mobile"] = headers.SecCHUAMobile
	if headers.SecCHUAMobile != "" && headers.SecCHUAMobile != "true" && headers.SecCHUAMobile != "false" {
		result.Issues = append(result.Issues, "Sec-CH-UA-Mobile 值无效")
	}

	// Sec-CH-UA-Platform 验证
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

// validateJSLayer 验证 JavaScript 层一致性
func (cv *ConsistencyValidator) validateJSLayer() *LayerCheckResult {
	result := &LayerCheckResult{
		LayerName: "JavaScript",
		Data:      make(map[string]string),
	}

	// 检查 JSAntiDetection 配置
	if cv.profile.JSAntiDetection == nil {
		result.Issues = append(result.Issues, "JSAntiDetection 配置缺失")
		return result
	}

	antiDetect := cv.profile.JSAntiDetection

	// 验证各对抗点配置
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

// validateTCPIPLayer 验证 TCP/IP 层一致性
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

	// 验证 TCP/IP 参数有效性
	result.Data["TTL"] = fmt.Sprintf("%d", tcpip.TTL)
	result.Data["WindowSize"] = fmt.Sprintf("%d", tcpip.WindowSize)
	result.Data["MSS"] = fmt.Sprintf("%d", tcpip.MSS)
	result.Data["OS"] = string(cv.profile.OS)

	// TTL 值验证
	if tcpip.TTL == 0 {
		result.Issues = append(result.Issues, "TTL 值为 0")
	}

	// Windows 应该有 TTL 128
	if strings.Contains(string(cv.profile.OS), "Windows") && tcpip.TTL != 128 {
		result.Issues = append(result.Issues, fmt.Sprintf("Windows 期望 TTL 128，实际 %d", tcpip.TTL))
	}

	// Linux/Mac 应该有 TTL 64
	if (strings.Contains(string(cv.profile.OS), "Linux") || strings.Contains(string(cv.profile.OS), "Mac OS")) &&
		tcpip.TTL != 64 {
		result.Issues = append(result.Issues, fmt.Sprintf("Unix 类系统期望 TTL 64，实际 %d", tcpip.TTL))
	}

	// JA4T 指纹有效性验证
	if tcpip.JA4T == "" {
		result.Issues = append(result.Issues, "JA4T 指纹为空")
	}

	result.IsConsistent = len(result.Issues) == 0

	return result
}

// crossLayerValidation 执行跨层交叉验证
func (cv *ConsistencyValidator) crossLayerValidation(report *ConsistencyReport) {
	if cv.profile.Headers == nil {
		return
	}

	headers := cv.profile.Headers
	ua := headers.UserAgent

	// 1. UA 中的浏览器信息 vs BrowserType
	if !cv.isUAConsistentWithBrowser(ua) {
		report.Mismatches = append(report.Mismatches, "User-Agent 与 BrowserType 不一致")
	}

	// 2. UA 中的 OS 信息 vs OS 字段
	if !cv.isUAConsistentWithOS(ua) {
		report.Mismatches = append(report.Mismatches, "User-Agent 与 OS 信息不一致")
	}

	// 3. Sec-CH-UA 与 UA 一致性
	if !cv.isSecCHUAConsistentWithUA(ua, headers.SecCHUA) {
		report.Mismatches = append(report.Mismatches, "Sec-CH-UA 与 User-Agent 不一致")
	}

	// 4. 语言信息一致性
	if !cv.isLanguageConsistent() {
		report.Warnings = append(report.Warnings, "语言信息可能不一致")
	}

	// 5. TCP/IP 与 OS 一致性
	if cv.profile.TCPIP != nil {
		if !cv.isTCPIPConsistentWithOS() {
			report.Mismatches = append(report.Mismatches, "TCP/IP 配置与 OS 不一致")
		}
	}
}

// calculateScore 计算一致性得分
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

	// 基础分
	baseiScore := float64(consistentLayers) / float64(totalLayers)

	// 根据不匹配数调整
	mismatchPenalty := float64(len(report.Mismatches)) * 0.1
	warningPenalty := float64(len(report.Warnings)) * 0.05

	report.Score = baseiScore - mismatchPenalty - warningPenalty
	if report.Score < 0 {
		report.Score = 0
	}

	// 确定总体一致性
	report.IsConsistent = report.Score >= 0.8 && len(report.Mismatches) == 0
}

// 辅助方法

func (cv *ConsistencyValidator) isValidUserAgent(ua string) bool {
	// User-Agent 应该包含浏览器标识
	browsers := []string{"Chrome", "Firefox", "Safari", "Edge", "Opera", "Brave"}
	for _, b := range browsers {
		if strings.Contains(ua, b) {
			return true
		}
	}
	return false
}

func (cv *ConsistencyValidator) isValidLanguageTag(lang string) bool {
	// 基本的语言标签验证 (如 "en-US", "zh-CN")
	parts := strings.Split(lang, "-")
	if len(parts) >= 1 && len(parts[0]) == 2 {
		return true
	}
	return false
}

func (cv *ConsistencyValidator) isValidSecCHUA(secCHUA string) bool {
	// Sec-CH-UA 应该遵循格式: "brand";v="version", ...
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
	// 简单的一致性检查：如果 UA 中有品牌，Sec-CH-UA 也应该有
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

	// 语言应该与 UA 中的 OS 或其他线索一致
	// 这里只做基本的非空检查
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
