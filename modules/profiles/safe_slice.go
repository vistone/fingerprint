// Package profiles 提供安全的字符串切片操作
package profiles

import "strings"

// safeSliceBefore 安全获取字符串中某个字符之前的部分
// 如果找不到分隔符，返回整个字符串
func safeSliceBefore(s string, sep string) string {
	idx := strings.Index(s, sep)
	if idx == -1 {
		return s
	}
	return s[:idx]
}

// safeSliceBeforeByte 安全获取字符串中某个字节之前的部分
func safeSliceBeforeByte(s string, sep byte) string {
	idx := strings.IndexByte(s, sep)
	if idx == -1 {
		return s
	}
	return s[:idx]
}

// safeLeft 安全获取字符串左侧 n 个字符
// 如果字符串长度小于 n，返回整个字符串
func safeLeft(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// getMajorVersion 获取版本号的主版本号（第一个数字）
func getMajorVersion(version string) string {
	return safeSliceBeforeByte(version, '.')
}

// getMinorVersion 获取版本号的前3个字符（如 "120.0.1" -> "120"）
func getMinorVersion(version string) string {
	return safeLeft(version, 3)
}

// safeSliceVersion 安全获取版本号的前缀（用于Sec-CH-UA）
// 优先返回主版本号，如果不存在则返回前3个字符
func safeSliceVersion(version string) string {
	major := getMajorVersion(version)
	if len(major) <= 3 {
		return major
	}
	return safeLeft(major, 3)
}

// validateVersion 验证版本号格式是否有效
func validateVersion(version string) bool {
	if version == "" {
		return false
	}
	// 检查是否至少包含一个数字
	for _, c := range version {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}
