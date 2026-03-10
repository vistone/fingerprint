package utils

import (
	"strings"
	"testing"
)

// translated comment
func TestStringPool_ToLower(t *testing.T) {
	pool := NewStringPool()

	tests := []struct {
		input    string
		expected string
	}{
		{"HELLO", "hello"},
		{"Hello", "hello"},
		{"hello", "hello"},  // translated comment
		{"Test123", "test123"},
		{"UPPER_CASE", "upper_case"},
	}

	for _, tt := range tests {
		result := pool.ToLower(tt.input)
		if result != tt.expected {
			t.Errorf("ToLower(%s) = %s, want %s", tt.input, result, tt.expected)
		}

		// translated comment
		result2 := pool.ToLower(tt.input)
		if result2 != tt.expected {
			t.Errorf("ToLower cached(%s) = %s, want %s", tt.input, result2, tt.expected)
		}
	}
}

// translated comment
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

// translated comment
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

// translated comment
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

// translated comment
func BenchmarkStringPool_ToLower(b *testing.B) {
	pool := NewStringPool()
	testStr := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pool.ToLower(testStr)
	}
}

// translated comment
func BenchmarkStringsToLower(b *testing.B) {
	testStr := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.ToLower(testStr)
	}
}

// translated comment
func BenchmarkCaseInsensitiveContains(b *testing.B) {
	text := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	substr := "windows"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CaseInsensitiveContains(text, substr)
	}
}

// translated comment
func BenchmarkStringsContainsLower(b *testing.B) {
	text := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	substr := "windows"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Contains(strings.ToLower(text), strings.ToLower(substr))
	}
}
