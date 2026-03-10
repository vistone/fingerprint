// Package profiles 提供指纹验证器
package profiles

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
)

// ProfileValidationResult 指纹验证结果
type ProfileValidationResult struct {
	Valid         bool
	MissingFields []string
	Warnings      []string
	Errors        []string
}

// ProfileValidator 指纹验证器
type ProfileValidator struct {
	strictMode bool
}

// NewProfileValidator 创建新的验证器
func NewProfileValidator() *ProfileValidator {
	return &ProfileValidator{
		strictMode: false,
	}
}

// SetStrictMode 设置严格模式（缺失字段作为错误）
func (pv *ProfileValidator) SetStrictMode(strict bool) {
	pv.strictMode = strict
}

// Validate 验证指纹配置
func (pv *ProfileValidator) Validate(profile ClientProfile) ProfileValidationResult {
	result := ProfileValidationResult{
		Valid:         true,
		MissingFields: []string{},
		Warnings:      []string{},
		Errors:        []string{},
	}

	// 1. 必需字段检查
	if profile.ID == "" {
		result.Errors = append(result.Errors, "ID is required")
		result.Valid = false
	}
	if profile.Name == "" {
		result.Errors = append(result.Errors, "Name is required")
		result.Valid = false
	}

	// 2. 浏览器类型检查
	if profile.BrowserType == "" {
		result.MissingFields = append(result.MissingFields, "BrowserType")
		if pv.strictMode {
			result.Errors = append(result.Errors, "BrowserType is required")
			result.Valid = false
		}
	}

	// 3. 操作系统检查
	if profile.OS == "" {
		result.MissingFields = append(result.MissingFields, "OS")
		if pv.strictMode {
			result.Errors = append(result.Errors, "OS is required")
			result.Valid = false
		}
	}

	// 4. TLS 配置检查
	if profile.TLSVersion == 0 {
		result.MissingFields = append(result.MissingFields, "TLSVersion")
		result.Warnings = append(result.Warnings, "TLSVersion not specified, will use default TLS 1.2")
	}

	if len(profile.CipherSuites) == 0 {
		result.MissingFields = append(result.MissingFields, "CipherSuites")
		result.Warnings = append(result.Warnings, "CipherSuites is empty, will use system default")
	}

	if len(profile.Extensions) == 0 {
		result.MissingFields = append(result.MissingFields, "Extensions")
		result.Warnings = append(result.Warnings, "TLS Extensions not configured")
	}

	// 5. HTTP Headers 检查
	if profile.Headers == nil {
		result.MissingFields = append(result.MissingFields, "Headers")
		result.Warnings = append(result.Warnings, "HTTP Headers missing, request will use defaults")
	} else {
		headerValidator := ValidateHeaders(profile.Headers)
		if len(headerValidator.Missing) > 0 {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Missing HTTP headers: %s", strings.Join(headerValidator.Missing, ", ")))
		}
	}

	// 6. TCP/IP 配置检查
	if profile.TCPIP == nil {
		result.Warnings = append(result.Warnings, "TCPIP configuration missing, TCP fingerprint will not be applied")
	} else {
		if err := ValidateTCPIP(profile.TCPIP); err != "" {
			result.Warnings = append(result.Warnings, fmt.Sprintf("TCPIP issue: %s", err))
		}
	}

	// 7. HTTP/2 配置检查
	if profile.HTTP2Settings.HeaderTableSize == 0 {
		result.Warnings = append(result.Warnings, "HTTP/2 Settings not configured")
	}

	// 8. 版本信息完整性
	if profile.BrowserVersion == "" {
		result.Warnings = append(result.Warnings, "BrowserVersion not specified")
	}
	if profile.OSVersion == "" {
		result.Warnings = append(result.Warnings, "OSVersion not specified")
	}

	return result
}

// HeaderValidationResult 头部验证结果
type HeaderValidationResult struct {
	Missing []string
	Empty   []string
}

// ValidateHeaders 验证 HTTP 头
func ValidateHeaders(headers *core.HTTPHeaders) HeaderValidationResult {
	result := HeaderValidationResult{
		Missing: []string{},
		Empty:   []string{},
	}

	if headers == nil {
		return result
	}

	// 必需的 headers
	requiredHeaders := map[string]string{
		"User-Agent":      headers.UserAgent,
		"Accept":          headers.Accept,
		"Accept-Language": headers.AcceptLanguage,
		"Accept-Encoding": headers.AcceptEncoding,
	}

	for name, value := range requiredHeaders {
		if value == "" {
			result.Empty = append(result.Empty, name)
		}
	}

	// Chrome 特有 headers
	chromeHeaders := map[string]string{
		"Sec-CH-UA":          headers.SecCHUA,
		"Sec-CH-UA-Mobile":   headers.SecCHUAMobile,
		"Sec-CH-UA-Platform": headers.SecCHUAPlatform,
	}

	for name, value := range chromeHeaders {
		if value == "" {
			result.Missing = append(result.Missing, name)
		}
	}

	return result
}

// ValidateTCPIP 验证 TCP/IP 配置
func ValidateTCPIP(tcpip *TCPIPFingerprint) string {
	if tcpip == nil {
		return "TCPIP is nil"
	}

	if tcpip.TTL == 0 {
		return "TTL value is 0"
	}

	if tcpip.WindowSize == 0 {
		return "WindowSize is 0"
	}

	if tcpip.JA4T == "" {
		return "JA4T fingerprint not set"
	}

	return ""
}

// ValidateAndRepair 验证并修复 profile
func ValidateAndRepair(profile *ClientProfile) ProfileValidationResult {
	validator := NewProfileValidator()
	result := validator.Validate(*profile)

	// 自动修复逻辑
	if profile.TLSVersion == 0 {
		profile.TLSVersion = 0x0303 // TLS 1.2
	}

	if profile.Headers == nil {
		profile.Headers = &core.HTTPHeaders{}
	}

	// 确保关键字段不为空
	if profile.Headers.UserAgent == "" {
		profile.Headers.UserAgent = generateDefaultUserAgent(profile.BrowserType)
	}
	if profile.Headers.Accept == "" {
		profile.Headers.Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"
	}
	if profile.Headers.AcceptLanguage == "" {
		profile.Headers.AcceptLanguage = "en-US,en;q=0.9"
	}
	if profile.Headers.AcceptEncoding == "" {
		profile.Headers.AcceptEncoding = "gzip, deflate, br"
	}

	// 修复 Brave 历史 profile 中的异常 UA/CH-UA（如 Chrome/1.xx, v="1"），
	// 这些值会被部分站点当作异常流量并返回 403。
	normalizeBraveHeaders(profile)

	return result
}

func normalizeBraveHeaders(profile *ClientProfile) {
	if profile == nil || profile.Headers == nil {
		return
	}
	if profile.BrowserType != core.BrowserBrave {
		return
	}

	uaLower := strings.ToLower(profile.Headers.UserAgent)
	chua := profile.Headers.SecCHUA

	if strings.Contains(uaLower, "chrome/1.") || strings.Contains(chua, `v="1"`) {
		major := deriveBraveChromiumMajor(profile.BrowserVersion)
		if major <= 0 {
			major = 120
		}

		profile.Headers.UserAgent = fmt.Sprintf(
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%d.0.0.0 Safari/537.36",
			major,
		)
		// 对部分严格站点，Brave 品牌标识会触发风控；兼容场景使用 Chrome 品牌组合。
		profile.Headers.SecCHUA = fmt.Sprintf(`"Chromium";v="%d", "Google Chrome";v="%d"`, major, major)
		if profile.Headers.SecCHUAMobile == "" {
			profile.Headers.SecCHUAMobile = "?0"
		}
		if profile.Headers.SecCHUAPlatform == "" {
			profile.Headers.SecCHUAPlatform = `"Windows"`
		}
	}
}

func deriveBraveChromiumMajor(braveVersion string) int {
	parts := strings.Split(braveVersion, ".")
	if len(parts) < 2 {
		return 0
	}

	// Brave 1.xx 通常对应 Chromium (xx + 59)，例如 1.60 -> 119。
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}
	if minor < 40 {
		return 0
	}
	return minor + 59
}

// generateDefaultUserAgent 生成默认 User-Agent
func generateDefaultUserAgent(browserType core.BrowserType) string {
	switch browserType {
	case core.BrowserChrome:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	case core.BrowserFirefox:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0"
	case core.BrowserSafari:
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_2) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15"
	case core.BrowserEdge:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0"
	default:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	}
}

// ValidateProfileList 验证整个 profile 列表
func ValidateProfileList(profiles []ClientProfile) map[string]ProfileValidationResult {
	validator := NewProfileValidator()
	results := make(map[string]ProfileValidationResult)

	for _, p := range profiles {
		results[p.ID] = validator.Validate(p)
	}

	return results
}
