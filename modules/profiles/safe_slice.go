// translated comment
package profiles

import "strings"

// translated comment
// translated comment
func safeSliceBefore(s string, sep string) string {
	idx := strings.Index(s, sep)
	if idx == -1 {
		return s
	}
	return s[:idx]
}

// translated comment
func safeSliceBeforeByte(s string, sep byte) string {
	idx := strings.IndexByte(s, sep)
	if idx == -1 {
		return s
	}
	return s[:idx]
}

// translated comment
// translated comment
func safeLeft(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// translated comment
func getMajorVersion(version string) string {
	return safeSliceBeforeByte(version, '.')
}

// translated comment
func getMinorVersion(version string) string {
	return safeLeft(version, 3)
}

// translated comment
// translated comment
func safeSliceVersion(version string) string {
	major := getMajorVersion(version)
	if len(major) <= 3 {
		return major
	}
	return safeLeft(major, 3)
}

// translated comment
func validateVersion(version string) bool {
	if version == "" {
		return false
	}
	// translated comment
	for _, c := range version {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}
