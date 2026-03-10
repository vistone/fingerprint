// Package profiles provides safe string slice operations
package profiles

import "strings"

// safeSliceBefore safely gets the part of a string before a given character
// if the separator is not found, return the entire string
func safeSliceBefore(s string, sep string) string {
	idx := strings.Index(s, sep)
	if idx == -1 {
		return s
	}
	return s[:idx]
}

// safeSliceBeforeByte safely gets the part of a string before a given byte
func safeSliceBeforeByte(s string, sep byte) string {
	idx := strings.IndexByte(s, sep)
	if idx == -1 {
		return s
	}
	return s[:idx]
}

// safeLeft safely gets the first n characters of a string
// if the string length is less than n, returns the entire string
func safeLeft(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// getMajorVersion gets the major version number (first digit) from a version string
func getMajorVersion(version string) string {
	return safeSliceBeforeByte(version, '.')
}

// getMinorVersion gets the version number prefix of 3 characters (e.g. "120.0.1" -> "120")
func getMinorVersion(version string) string {
	return safeLeft(version, 3)
}

// safeSliceVersion safely gets the version number prefix (for Sec-CH-UA)
// preferably returns the major version; if not found, returns the first 3 characters
func safeSliceVersion(version string) string {
	major := getMajorVersion(version)
	if len(major) <= 3 {
		return major
	}
	return safeLeft(major, 3)
}

// validateVersion validates whether the version number format is valid
func validateVersion(version string) bool {
	if version == "" {
		return false
	}
	// checks whether it contains at least one digit
	for _, c := range version {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}
