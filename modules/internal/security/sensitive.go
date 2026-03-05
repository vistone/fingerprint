package security

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// DataClassifier classifies data sensitivity levels
type DataClassifier struct {
	patterns map[string]*regexp.Regexp
}

// SensitivityLevel represents data sensitivity
type SensitivityLevel int

const (
	// Public data can be freely shared
	Public SensitivityLevel = iota
	// Internal data for internal use only
	Internal
	// Confidential sensitive data
	Confidential
	// Restricted highly sensitive data
	Restricted
)

// String returns the string representation of sensitivity level
func (s SensitivityLevel) String() string {
	switch s {
	case Public:
		return "public"
	case Internal:
		return "internal"
	case Confidential:
		return "confidential"
	case Restricted:
		return "restricted"
	default:
		return "unknown"
	}
}

// NewDataClassifier creates a new data classifier
func NewDataClassifier() *DataClassifier {
	return &DataClassifier{
		patterns: map[string]*regexp.Regexp{
			"email":     regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
			"ipv4":      regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),
			"ipv6":      regexp.MustCompile(`\b(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\b`),
			"phone":     regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`),
			"credit_card": regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`),
			"ssn":       regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			"api_key":   regexp.MustCompile(`(?i)(api[_-]?key|apikey)[\s]*[=:][\s]*[a-zA-Z0-9]{16,}`),
			"token":     regexp.MustCompile(`(?i)(token|bearer)[\s]+[a-zA-Z0-9_\-\.]+`),
		},
	}
}

// Classify classifies data sensitivity
func (c *DataClassifier) Classify(data string) SensitivityLevel {
	// Check for restricted patterns
	if c.patterns["ssn"].MatchString(data) ||
		c.patterns["credit_card"].MatchString(data) ||
		c.patterns["api_key"].MatchString(data) {
		return Restricted
	}

	// Check for confidential patterns
	if c.patterns["token"].MatchString(data) ||
		c.patterns["phone"].MatchString(data) {
		return Confidential
	}

	// Check for internal patterns
	if c.patterns["email"].MatchString(data) ||
		c.patterns["ipv4"].MatchString(data) ||
		c.patterns["ipv6"].MatchString(data) {
		return Internal
	}

	return Public
}

// Masker provides data masking utilities
type Masker struct {
	preserveLength int
	maskChar       rune
}

// NewMasker creates a new data masker
func NewMasker() *Masker {
	return &Masker{
		preserveLength: 4,
		maskChar:       '*',
	}
}

// SetPreserveLength sets how many characters to preserve at the end
func (m *Masker) SetPreserveLength(length int) {
	m.preserveLength = length
}

// Mask masks sensitive data
func (m *Masker) Mask(input string) string {
	if len(input) <= m.preserveLength {
		return strings.Repeat(string(m.maskChar), len(input))
	}

	visible := input[len(input)-m.preserveLength:]
	masked := strings.Repeat(string(m.maskChar), len(input)-m.preserveLength)
	return masked + visible
}

// MaskEmail masks an email address
func (m *Masker) MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return m.Mask(email)
	}

	local := parts[0]
	domain := parts[1]

	// Mask most of local part
	if len(local) > 2 {
		local = local[:2] + strings.Repeat(string(m.maskChar), len(local)-2)
	}

	return local + "@" + domain
}

// MaskIP masks an IP address
func (m *Masker) MaskIP(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return m.Mask(ip)
	}

	// Keep last octet, mask others
	return "***.***.***." + parts[3]
}

// Hash hashes sensitive data for comparison without storing plaintext
func (m *Masker) Hash(input string) string {
	h := sha256.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

// Redactor redacts sensitive information
type Redactor struct {
	patterns map[string]*regexp.Regexp
}

// NewRedactor creates a new redactor
func NewRedactor() *Redactor {
	return &Redactor{
		patterns: map[string]*regexp.Regexp{
			"password":  regexp.MustCompile(`(?i)(password|passwd|pwd)[\s]*[=:][\s]*[^\s]+`),
			"secret":    regexp.MustCompile(`(?i)(secret|private_key)[\s]*[=:][\s]*[^\s]+`),
			"token":     regexp.MustCompile(`(?i)(token|access_token)[\s]*[=:][\s]*[^\s]+`),
			"api_key":   regexp.MustCompile(`(?i)(api_key|apikey)[\s]*[=:][\s]*[^\s]+`),
			"auth":      regexp.MustCompile(`(?i)(authorization|auth)[\s]*[=:][\s]*[^\s]+`),
			"cookie":    regexp.MustCompile(`(?i)(cookie|session)[\s]*[=:][\s]*[^\s]+`),
		},
	}
}

// Redact redacts sensitive patterns from text
func (r *Redactor) Redact(input string) string {
	result := input

	replacements := map[string]string{
		"password": "[REDACTED_PASSWORD]",
		"secret":   "[REDACTED_SECRET]",
		"token":    "[REDACTED_TOKEN]",
		"api_key":  "[REDACTED_API_KEY]",
		"auth":     "[REDACTED_AUTH]",
		"cookie":   "[REDACTED_COOKIE]",
	}

	for name, pattern := range r.patterns {
		result = pattern.ReplaceAllString(result, replacements[name])
	}

	return result
}

// PartialRedact partially redacts sensitive data (keep first/last few chars)
func (r *Redactor) PartialRedact(input string, visibleChars int) string {
	if len(input) <= visibleChars*2 {
		return strings.Repeat("*", len(input))
	}

	prefix := input[:visibleChars]
	suffix := input[len(input)-visibleChars:]
	return prefix + strings.Repeat("*", len(input)-visibleChars*2) + suffix
}

// SensitiveDataHandler handles sensitive data processing
type SensitiveDataHandler struct {
	classifier *DataClassifier
	masker     *Masker
	redactor   *Redactor
	minLevel   SensitivityLevel
}

// NewSensitiveDataHandler creates a new handler
func NewSensitiveDataHandler() *SensitiveDataHandler {
	return &SensitiveDataHandler{
		classifier: NewDataClassifier(),
		masker:     NewMasker(),
		redactor:   NewRedactor(),
		minLevel:   Internal,
	}
}

// SetMinLevel sets the minimum sensitivity level to process
func (h *SensitiveDataHandler) SetMinLevel(level SensitivityLevel) {
	h.minLevel = level
}

// Process processes sensitive data based on its classification
func (h *SensitiveDataHandler) Process(data string, action string) string {
	level := h.classifier.Classify(data)

	if level < h.minLevel {
		return data // No processing needed
	}

	switch action {
	case "mask":
		return h.masker.Mask(data)
	case "hash":
		return h.masker.Hash(data)
	case "redact":
		return "[REDACTED]"
	case "email":
		return h.masker.MaskEmail(data)
	case "ip":
		return h.masker.MaskIP(data)
	default:
		return h.masker.Mask(data)
	}
}

// ProcessMap processes a map of data
func (h *SensitiveDataHandler) ProcessMap(data map[string]string, sensitiveFields []string) map[string]string {
	result := make(map[string]string)

	for key, value := range data {
		if h.isSensitiveField(key, sensitiveFields) {
			result[key] = h.Process(value, "mask")
		} else {
			result[key] = value
		}
	}

	return result
}

// isSensitiveField checks if a field is in the sensitive list
func (h *SensitiveDataHandler) isSensitiveField(field string, sensitiveFields []string) bool {
	lowerField := strings.ToLower(field)
	for _, sensitive := range sensitiveFields {
		if strings.Contains(lowerField, strings.ToLower(sensitive)) {
			return true
		}
	}
	return false
}

// SanitizeForLogging sanitizes data for logging
func (h *SensitiveDataHandler) SanitizeForLogging(data string) string {
	// First redact known sensitive patterns
	data = h.redactor.Redact(data)
	
	// Then classify and mask any remaining sensitive data
	level := h.classifier.Classify(data)
	if level >= Confidential {
		return h.masker.Mask(data)
	}

	return data
}

// Common sensitive field names
var CommonSensitiveFields = []string{
	"password",
	"secret",
	"token",
	"key",
	"auth",
	"credential",
	"private",
	"ssn",
	"social",
	"credit",
	"card",
	"cvv",
	"pin",
	"passphrase",
}
