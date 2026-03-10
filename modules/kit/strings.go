package utils

import "strings"

// Contains checks whether a string contains a substring.
// Implemented with the standard library for better performance.
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// HasPrefix checks whether a string starts with the given prefix.
func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// ToLower converts a string to lowercase.
func ToLower(s string) string {
	return strings.ToLower(s)
}

// Index returns the first occurrence index of a substring in a string.
func Index(s, substr string) int {
	return strings.Index(s, substr)
}

// TrimPrefix removes a prefix from a string.
func TrimPrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

// Split splits a string.
func Split(s, sep string) []string {
	return strings.Split(s, sep)
}
