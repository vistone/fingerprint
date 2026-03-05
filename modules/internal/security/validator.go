// Package security provides security utilities for the fingerprint gateway
package security

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Validator provides input validation utilities
type Validator struct {
	maxStringLength int
	allowedPatterns map[string]*regexp.Regexp
}

// NewValidator creates a new validator with default settings
func NewValidator() *Validator {
	return &Validator{
		maxStringLength: 4096,
		allowedPatterns: map[string]*regexp.Regexp{
			"alphanumeric": regexp.MustCompile(`^[a-zA-Z0-9]+$`),
			"hex":          regexp.MustCompile(`^[a-fA-F0-9]+$`),
			"base64":       regexp.MustCompile(`^[a-zA-Z0-9+/]*={0,2}$`),
			"ip":           regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`),
			"uuid":         regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`),
		},
	}
}

// SetMaxStringLength sets the maximum allowed string length
func (v *Validator) SetMaxStringLength(length int) {
	v.maxStringLength = length
}

// ValidateString checks if a string is safe
func (v *Validator) ValidateString(s string) error {
	if len(s) > v.maxStringLength {
		return fmt.Errorf("string exceeds maximum length of %d", v.maxStringLength)
	}

	if !utf8.ValidString(s) {
		return fmt.Errorf("string contains invalid UTF-8")
	}

	// Check for null bytes
	if strings.Contains(s, "\x00") {
		return fmt.Errorf("string contains null bytes")
	}

	return nil
}

// ValidateIP validates an IP address
func (v *Validator) ValidateIP(ip string) error {
	if err := v.ValidateString(ip); err != nil {
		return fmt.Errorf("validate IP: %w", err)
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP address format")
	}

	return nil
}

// ValidatePort validates a port number
func (v *Validator) ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}

// ValidateHex validates a hexadecimal string
func (v *Validator) ValidateHex(s string) error {
	if err := v.ValidateString(s); err != nil {
		return fmt.Errorf("validate hex: %w", err)
	}

	if !v.allowedPatterns["hex"].MatchString(s) {
		return fmt.Errorf("invalid hexadecimal format")
	}

	return nil
}

// ValidateBase64 validates a base64 string
func (v *Validator) ValidateBase64(s string) error {
	if err := v.ValidateString(s); err != nil {
		return fmt.Errorf("validate base64: %w", err)
	}

	if !v.allowedPatterns["base64"].MatchString(s) {
		return fmt.Errorf("invalid base64 format")
	}

	return nil
}

// ValidateUserAgent validates a User-Agent string
func (v *Validator) ValidateUserAgent(ua string) error {
	if err := v.ValidateString(ua); err != nil {
		return fmt.Errorf("validate user agent: %w", err)
	}

	// Check for common injection patterns
	dangerous := []string{"<script", "javascript:", "onerror=", "onload="}
	lowerUA := strings.ToLower(ua)
	for _, pattern := range dangerous {
		if strings.Contains(lowerUA, pattern) {
			return fmt.Errorf("user agent contains potentially dangerous content")
		}
	}

	return nil
}

// ValidateHeaderName validates an HTTP header name
func (v *Validator) ValidateHeaderName(name string) error {
	if err := v.ValidateString(name); err != nil {
		return fmt.Errorf("validate header name: %w", err)
	}

	// Header names should only contain printable ASCII characters except colon
	for _, r := range name {
		if r < 33 || r > 126 || r == ':' {
			return fmt.Errorf("invalid character in header name")
		}
	}

	return nil
}

// ValidateHeaderValue validates an HTTP header value
func (v *Validator) ValidateHeaderValue(value string) error {
	if err := v.ValidateString(value); err != nil {
		return fmt.Errorf("validate header value: %w", err)
	}

	// Check for CRLF injection
	if strings.Contains(value, "\r") || strings.Contains(value, "\n") {
		return fmt.Errorf("header value contains line breaks")
	}

	return nil
}

// Sanitizer provides input sanitization utilities
type Sanitizer struct {
	maxLength int
}

// NewSanitizer creates a new sanitizer
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		maxLength: 4096,
	}
}

// SetMaxLength sets the maximum output length
func (s *Sanitizer) SetMaxLength(length int) {
	s.maxLength = length
}

// SanitizeString removes potentially dangerous characters
func (s *Sanitizer) SanitizeString(input string) string {
	if len(input) > s.maxLength {
		input = input[:s.maxLength]
	}

	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except common whitespace
	var result strings.Builder
	for _, r := range input {
		if r == '\t' || r == '\n' || r == '\r' || r >= 32 {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// SanitizeHTML escapes HTML special characters
func (s *Sanitizer) SanitizeHTML(input string) string {
	input = s.SanitizeString(input)

	replacements := map[string]string{
		"&":  "&amp;",
		"<":  "&lt;",
		">":  "&gt;",
		"\"": "&quot;",
		"'":  "&#x27;",
		"/":  "&#x2F;",
	}

	var result strings.Builder
	for _, r := range input {
		if replacement, ok := replacements[string(r)]; ok {
			result.WriteString(replacement)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// SanitizeLogValue sanitizes values for logging
func (s *Sanitizer) SanitizeLogValue(input string) string {
	// Truncate if too long
	if len(input) > s.maxLength {
		input = input[:s.maxLength] + "..."
	}

	// Remove newlines to prevent log injection
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "\r", " ")

	return input
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	MaxRequestSize      int64
	MaxHeaderSize       int
	MaxHeaders          int
	RateLimitEnabled    bool
	RateLimitRequests   int
	RateLimitWindow     int // seconds
	RequireTLS          bool
	AllowedHosts        []string
	BlockedIPs          []string
	SanitizeInput       bool
	ValidateInput       bool
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		MaxRequestSize:    10 * 1024 * 1024, // 10MB
		MaxHeaderSize:     1 * 1024 * 1024,  // 1MB
		MaxHeaders:        100,
		RateLimitEnabled:  true,
		RateLimitRequests: 1000,
		RateLimitWindow:   60,
		RequireTLS:        false,
		AllowedHosts:      []string{},
		BlockedIPs:        []string{},
		SanitizeInput:     true,
		ValidateInput:     true,
	}
}

// IsHostAllowed checks if a host is in the allowed list
func (c *SecurityConfig) IsHostAllowed(host string) bool {
	if len(c.AllowedHosts) == 0 {
		return true // Allow all if no restrictions
	}

	for _, allowed := range c.AllowedHosts {
		if strings.EqualFold(allowed, host) {
			return true
		}
	}

	return false
}

// IsIPBlocked checks if an IP is blocked
func (c *SecurityConfig) IsIPBlocked(ip string) bool {
	for _, blocked := range c.BlockedIPs {
		if blocked == ip {
			return true
		}
	}
	return false
}
