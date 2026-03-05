package utils

import "fmt"

// ExtractChromeVersion 从 User-Agent 中提取 Chrome 版本号
func ExtractChromeVersion(ua string) string {
	start := Index(ua, "Chrome/")
	if start == -1 {
		return "120" // 默认版本
	}
	start += 7 // "Chrome/" 的长度
	end := start
	for end < len(ua) && ua[end] != '.' && ua[end] != ' ' && ua[end] != ';' {
		end++
	}
	if end > start {
		return ua[start:end]
	}
	return "120"
}

// ExtractPlatform 从 User-Agent 中提取平台信息
func ExtractPlatform(ua string) string {
	if Contains(ua, "Windows") {
		return `"Windows"`
	} else if Contains(ua, "Macintosh") {
		return `"macOS"`
	} else if Contains(ua, "Linux") {
		return `"Linux"`
	}
	return `"Windows"` // 默认
}

// FormatUserAgent 格式化 User-Agent 字符串
func FormatUserAgent(template string, args ...interface{}) string {
	return fmt.Sprintf(template, args...)
}
