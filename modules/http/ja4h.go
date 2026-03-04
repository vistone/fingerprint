// Package http 提供 HTTP 指纹生成功能
package http

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// JA4HResult JA4H 指纹结果
type JA4HResult struct {
	// Fingerprint JA4H 指纹
	Fingerprint string
	// Method HTTP 方法
	Method string
	// Headers HTTP 头列表
	Headers []string
	// CookieNames Cookie 名称列表
	CookieNames []string
}

// CalculateJA4H 计算 JA4H 指纹
func CalculateJA4H(headers *core.HTTPHeaders, method string) *JA4HResult {
	if headers == nil {
		return nil
	}

	// 收集所有 header 名称
	var headerNames []string
	headerMap := headers.ToMap()
	for name := range headerMap {
		// 忽略 Cookie 头（单独处理）
		if strings.ToLower(name) != "cookie" {
			headerNames = append(headerNames, strings.ToLower(name))
		}
	}

	// 提取 Cookie 名称
	var cookieNames []string
	if cookie, ok := headerMap["Cookie"]; ok {
		cookieNames = extractCookieNames(cookie)
	}

	// 构建 JA4H 字符串
	// 格式: ja4h_<method>_<header_count>_<cookie_count>_<accept_language_hash>
	ja4hString := "ja4h_" +
		strings.ToLower(method) + "_" +
		intToHex(len(headerNames)) + "_" +
		intToHex(len(cookieNames))

	// 计算 Accept-Language 哈希
	if lang := headers.AcceptLanguage; lang != "" {
		hash := sha256.Sum256([]byte(lang))
		ja4hString += "_" + hex.EncodeToString(hash[:6])
	} else {
		ja4hString += "_000000"
	}

	return &JA4HResult{
		Fingerprint: ja4hString,
		Method:      method,
		Headers:     headerNames,
		CookieNames: cookieNames,
	}
}

// extractCookieNames 提取 Cookie 名称
func extractCookieNames(cookie string) []string {
	var names []string
	parts := strings.Split(cookie, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if idx := strings.Index(part, "="); idx > 0 {
			name := strings.TrimSpace(part[:idx])
			if name != "" {
				names = append(names, name)
			}
		}
	}
	return names
}

// intToHex 将整数转换为十六进制字符串
func intToHex(n int) string {
	if n < 10 {
		return string('0' + byte(n))
	}
	return string('a' + byte(n-10))
}

// HTTP2Fingerprint HTTP/2 指纹
type HTTP2Fingerprint struct {
	// SettingsHash Settings 帧哈希
	SettingsHash string
	// PriorityHash Priority 帧哈希
	PriorityHash string
	// WindowUpdateHash WINDOW_UPDATE 帧哈希
	WindowUpdateHash string
	// CombinedHash 综合哈希
	CombinedHash string
}

// FingerprintHTTP2 生成 HTTP/2 指纹
func FingerprintHTTP2(settings core.HTTP2Settings, priorities []core.HTTP2Priority, connectionFlow uint32) *HTTP2Fingerprint {
	// 简化实现
	settingsStr := settingsToString(settings)
	priorityStr := prioritiesToString(priorities)
	flowStr := uintToString(connectionFlow)

	settingsHash := hashString(settingsStr)
	priorityHash := hashString(priorityStr)
	windowHash := hashString(flowStr)
	combinedHash := hashString(settingsStr + priorityStr + flowStr)

	return &HTTP2Fingerprint{
		SettingsHash:     settingsHash,
		PriorityHash:     priorityHash,
		WindowUpdateHash: windowHash,
		CombinedHash:     combinedHash,
	}
}

// settingsToString 将 HTTP/2 Settings 转换为字符串
func settingsToString(settings core.HTTP2Settings) string {
	return uintToString(settings.HeaderTableSize) + "_" +
		uintToString(settings.EnablePush) + "_" +
		uintToString(settings.MaxConcurrentStreams) + "_" +
		uintToString(settings.InitialWindowSize) + "_" +
		uintToString(settings.MaxFrameSize) + "_" +
		uintToString(settings.MaxHeaderListSize)
}

// prioritiesToString 将 HTTP/2 Priorities 转换为字符串
func prioritiesToString(priorities []core.HTTP2Priority) string {
	var parts []string
	for _, p := range priorities {
		parts = append(parts, uintToString(p.StreamID)+"_"+uintToString(uint32(p.Weight)))
	}
	return strings.Join(parts, ",")
}

// uintToString 将 uint32 转换为字符串
func uintToString(n uint32) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	return string(result)
}

// hashString 计算字符串哈希（简化版）
func hashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:8])
}

// Analyzer HTTP 分析器
type Analyzer struct {
	profile *profiles.ClientProfile
}

// NewAnalyzer 创建新的 HTTP 分析器
func NewAnalyzer(profile *profiles.ClientProfile) *Analyzer {
	return &Analyzer{profile: profile}
}

// AnalyzeJA4H 分析 JA4H 指纹
func (a *Analyzer) AnalyzeJA4H(method string) *JA4HResult {
	if a.profile == nil || a.profile.Headers == nil {
		return nil
	}
	return CalculateJA4H(a.profile.Headers, method)
}

// AnalyzeHTTP2 分析 HTTP/2 指纹
func (a *Analyzer) AnalyzeHTTP2() *HTTP2Fingerprint {
	if a.profile == nil {
		return nil
	}
	return FingerprintHTTP2(a.profile.HTTP2Settings, a.profile.HTTP2Priorities, a.profile.ConnectionFlow)
}

// Fingerprint 生成完整的 HTTP 指纹
func (a *Analyzer) Fingerprint(method string) map[string]interface{} {
	return map[string]interface{}{
		"ja4h":  a.AnalyzeJA4H(method),
		"http2": a.AnalyzeHTTP2(),
	}
}

// MatchProfile 从 HTTP 头匹配浏览器配置
func MatchProfile(headers *core.HTTPHeaders) *profiles.ClientProfile {
	if headers == nil {
		return nil
	}

	// 基于 User-Agent 匹配
	ua := headers.UserAgent
	if ua == "" {
		return nil
	}

	// 获取所有配置
	allProfiles := profiles.GetAll()

	// 简单的匹配逻辑（实际应该使用更复杂的算法）
	for _, p := range allProfiles {
		if p.Headers != nil && p.Headers.UserAgent == ua {
			return &p
		}
	}

	return nil
}
