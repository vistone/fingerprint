package utils

import "fmt"

// ExtractChromeVersion extracts the Chrome version from User-Agent.
func ExtractChromeVersion(ua string) string {
	start := Index(ua, "Chrome/")
	if start == -1 {
		return "120" // Default version
	}
	start += 7 // Length of "Chrome/"
	end := start
	for end < len(ua) && ua[end] != '.' && ua[end] != ' ' && ua[end] != ';' {
		end++
	}
	if end > start {
		return ua[start:end]
	}
	return "120"
}

// ExtractPlatform extracts platform info from User-Agent.
func ExtractPlatform(ua string) string {
	if Contains(ua, "Windows") {
		return `"Windows"`
	} else if Contains(ua, "Macintosh") {
		return `"macOS"`
	} else if Contains(ua, "Linux") {
		return `"Linux"`
	}
	return `"Windows"` // Default
}

// FormatUserAgent formats the User-Agent string.
func FormatUserAgent(template string, args ...interface{}) string {
	return fmt.Sprintf(template, args...)
}
