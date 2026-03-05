package utils

import "strings"

// Contains 检查字符串是否包含子字符串
// 使用标准库实现，性能优于自定义实现
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// HasPrefix 检查字符串是否以指定前缀开始
func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// ToLower 将字符串转换为小写
func ToLower(s string) string {
	return strings.ToLower(s)
}

// Index 返回子字符串在字符串中首次出现的位置
func Index(s, substr string) int {
	return strings.Index(s, substr)
}

// TrimPrefix 移除字符串的前缀
func TrimPrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

// Split 分割字符串
func Split(s, sep string) []string {
	return strings.Split(s, sep)
}
