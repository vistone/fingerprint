package utils

import (
	"strings"
	"testing"
)

// TestStringPool_ToLower 测试字符串池小写转换
func TestStringPool_ToLower(t *testing.T) {
	pool := NewStringPool()

	tests := []struct {
		input    string
		expected string
	}{
		{"HELLO", "hello"},
		{"Hello", "hello"},
		{"hello", "hello"},  // 已经是小写
		{"Test123", "test123"},
		{"UPPER_CASE", "upper_case"},
	}

	for _, tt := range tests {
		result := pool.ToLower(tt.input)
		if result != tt.expected {
			t.Errorf("ToLower(%s) = %s, want %s", tt.input, result, tt.expected)
		}

		// 第二次获取应该返回缓存的值
		result2 := pool.ToLower(tt.input)
		if result2 != tt.expected {
			t.Errorf("ToLower cached(%s) = %s, want %s", tt.input, result2, tt.expected)
		}
	}
}

// TestIsAllLower 测试全小写检查
func TestIsAllLower(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello", true},
		{"HELLO", false},
		{"Hello", false},
		{"test123", true},
		{"Test_123", false},
		{"", true},
		{"123", true},
	}

	for _, tt := range tests {
		result := isAllLower(tt.input)
		if result != tt.expected {
			t.Errorf("isAllLower(%s) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

// TestCaseInsensitiveContains 测试不区分大小写的包含检查
func TestCaseInsensitiveContains(t *testing.T) {
	tests := []struct {
		text     string
		substr   string
		expected bool
	}{
		{"Hello World", "world", true},
		{"Hello World", "WORLD", true},
		{"Hello World", "World", true},
		{"Hello World", "foo", false},
		{"", "test", false},
		{"test", "", true},
		{"Test", "test", true},
	}

	for _, tt := range tests {
		result := CaseInsensitiveContains(tt.text, tt.substr)
		if result != tt.expected {
			t.Errorf("CaseInsensitiveContains(%s, %s) = %v, want %v",
				tt.text, tt.substr, result, tt.expected)
		}
	}
}

// TestFastContains 测试快速包含检查
func TestFastContains(t *testing.T) {
	tests := []struct {
		textLower   string
		substrLower string
		expected    bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"", "test", false},
		{"test", "", true},
	}

	for _, tt := range tests {
		result := FastContains(tt.textLower, tt.substrLower)
		if result != tt.expected {
			t.Errorf("FastContains(%s, %s) = %v, want %v",
				tt.textLower, tt.substrLower, result, tt.expected)
		}
	}
}

// BenchmarkStringPool_ToLower 基准测试字符串池小写转换
func BenchmarkStringPool_ToLower(b *testing.B) {
	pool := NewStringPool()
	testStr := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pool.ToLower(testStr)
	}
}

// BenchmarkStringsToLower 基准测试标准库小写转换
func BenchmarkStringsToLower(b *testing.B) {
	testStr := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.ToLower(testStr)
	}
}

// BenchmarkCaseInsensitiveContains 基准测试不区分大小写的包含检查
func BenchmarkCaseInsensitiveContains(b *testing.B) {
	text := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	substr := "windows"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CaseInsensitiveContains(text, substr)
	}
}

// BenchmarkStringsContainsLower 基准测试标准库包含检查
func BenchmarkStringsContainsLower(b *testing.B) {
	text := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	substr := "windows"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Contains(strings.ToLower(text), strings.ToLower(substr))
	}
}
