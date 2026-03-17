// Package profiles provides fingerprint validator
package profiles

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
)

// ProfileValidationResult fingerprint validation result
type ProfileValidationResult struct {
	Valid         bool
	MissingFields []string
	Warnings      []string
	Errors        []string
}

// ProfileValidator fingerprint validator
type ProfileValidator struct {
	strictMode bool
}

// NewProfileValidator creates a new validator
func NewProfileValidator() *ProfileValidator {
	return &ProfileValidator{
		strictMode: false,
	}
}

// SetStrictMode sets strict mode (missing fields treated as errors)
func (pv *ProfileValidator) SetStrictMode(strict bool) {
	pv.strictMode = strict
}

// Validate validates fingerprint profile
func (pv *ProfileValidator) Validate(profile ClientProfile) ProfileValidationResult {
	result := newValidationResult()
	pv.validateRequiredFields(profile, &result)
	pv.validateCoreProfile(profile, &result)
	pv.validateTLSProfile(profile, &result)
	pv.validateHTTPHeaders(profile, &result)
	pv.validateTCPIPProfile(profile, &result)
	pv.validateVersionMetadata(profile, &result)
	return result
}

func newValidationResult() ProfileValidationResult {
	return ProfileValidationResult{
		Valid:         true,
		MissingFields: []string{},
		Warnings:      []string{},
		Errors:        []string{},
	}
}

func (pv *ProfileValidator) validateRequiredFields(profile ClientProfile, result *ProfileValidationResult) {
	if profile.ID == "" {
		result.Errors = append(result.Errors, "ID is required")
		result.Valid = false
	}
	if profile.Name == "" {
		result.Errors = append(result.Errors, "Name is required")
		result.Valid = false
	}
}

func (pv *ProfileValidator) validateCoreProfile(profile ClientProfile, result *ProfileValidationResult) {
	if profile.BrowserType == "" {
		result.MissingFields = append(result.MissingFields, "BrowserType")
		if pv.strictMode {
			result.Errors = append(result.Errors, "BrowserType is required")
			result.Valid = false
		}
	}
	if profile.OS == "" {
		result.MissingFields = append(result.MissingFields, "OS")
		if pv.strictMode {
			result.Errors = append(result.Errors, "OS is required")
			result.Valid = false
		}
	}
}

func (pv *ProfileValidator) validateTLSProfile(profile ClientProfile, result *ProfileValidationResult) {
	if profile.TLSVersion == 0 {
		result.MissingFields = append(result.MissingFields, "TLSVersion")
		result.Warnings = append(result.Warnings, "TLSVersion not specified, will use default TLS 1.2")
	}
	if len(profile.CipherSuites) == 0 {
		result.MissingFields = append(result.MissingFields, "CipherSuites")
		result.Warnings = append(result.Warnings, "CipherSuites is empty, will use system default")
	}
	if len(profile.Extensions) == 0 {
		result.MissingFields = append(result.MissingFields, "Extensions")
		result.Warnings = append(result.Warnings, "TLS Extensions not configured")
	}
	if profile.HTTP2Settings.HeaderTableSize == 0 {
		result.Warnings = append(result.Warnings, "HTTP/2 Settings not configured")
	}
}

func (pv *ProfileValidator) validateHTTPHeaders(profile ClientProfile, result *ProfileValidationResult) {
	if profile.Headers == nil {
		result.MissingFields = append(result.MissingFields, "Headers")
		result.Warnings = append(result.Warnings, "HTTP Headers missing, request will use defaults")
		return
	}

	headerValidator := ValidateHeaders(profile.Headers)
	if len(headerValidator.Missing) > 0 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Missing HTTP headers: %s", strings.Join(headerValidator.Missing, ", ")))
	}
}

func (pv *ProfileValidator) validateTCPIPProfile(profile ClientProfile, result *ProfileValidationResult) {
	if profile.TCPIP == nil {
		result.Warnings = append(result.Warnings, "TCPIP configuration missing, TCP fingerprint will not be applied")
		return
	}
	if err := ValidateTCPIP(profile.TCPIP); err != "" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("TCPIP issue: %s", err))
	}
}

func (pv *ProfileValidator) validateVersionMetadata(profile ClientProfile, result *ProfileValidationResult) {
	if profile.BrowserVersion == "" {
		result.Warnings = append(result.Warnings, "BrowserVersion not specified")
	}
	if profile.OSVersion == "" {
		result.Warnings = append(result.Warnings, "OSVersion not specified")
	}
}

// HeaderValidationResult header validation result
type HeaderValidationResult struct {
	Missing []string
	Empty   []string
}

// ValidateHeaders validates HTTP headers
func ValidateHeaders(headers *core.HTTPHeaders) HeaderValidationResult {
	result := HeaderValidationResult{
		Missing: []string{},
		Empty:   []string{},
	}

	if headers == nil {
		return result
	}

	// required headers
	requiredHeaders := map[string]string{
		"User-Agent":      headers.UserAgent,
		"Accept":          headers.Accept,
		"Accept-Language": headers.AcceptLanguage,
		"Accept-Encoding": headers.AcceptEncoding,
	}

	for name, value := range requiredHeaders {
		if value == "" {
			result.Empty = append(result.Empty, name)
		}
	}

	// Chrome specific headers
	chromeHeaders := map[string]string{
		"Sec-CH-UA":          headers.SecCHUA,
		"Sec-CH-UA-Mobile":   headers.SecCHUAMobile,
		"Sec-CH-UA-Platform": headers.SecCHUAPlatform,
	}

	for name, value := range chromeHeaders {
		if value == "" {
			result.Missing = append(result.Missing, name)
		}
	}

	return result
}

// ValidateTCPIP validate TCP/IP profile
func ValidateTCPIP(tcpip *TCPIPFingerprint) string {
	if tcpip == nil {
		return "TCPIP is nil"
	}

	if tcpip.TTL == 0 {
		return "TTL value is 0"
	}

	if tcpip.WindowSize == 0 {
		return "WindowSize is 0"
	}

	if tcpip.JA4T == "" {
		return "JA4T fingerprint not set"
	}

	return ""
}

// ValidateAndRepair validates and repairs profile
func ValidateAndRepair(profile *ClientProfile) ProfileValidationResult {
	validator := NewProfileValidator()
	result := validator.Validate(*profile)

	// auto-repair logic
	if profile.TLSVersion == 0 {
		profile.TLSVersion = 0x0303 // TLS 1.2
	}

	if profile.Headers == nil {
		profile.Headers = &core.HTTPHeaders{}
	}

	// ensures critical fields are not empty
	if profile.Headers.UserAgent == "" {
		profile.Headers.UserAgent = generateDefaultUserAgent(profile.BrowserType)
	}
	if profile.Headers.Accept == "" {
		profile.Headers.Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"
	}
	if profile.Headers.AcceptLanguage == "" {
		profile.Headers.AcceptLanguage = "en-US,en;q=0.9"
	}
	if profile.Headers.AcceptEncoding == "" {
		profile.Headers.AcceptEncoding = "gzip, deflate, br"
	}

	// repair Brave history profile anomaly in UA/CH-UA (e.g. Chrome/1.xx, v="1"),
	// these values may be flagged as anomalous traffic by some sites and return 403.
	normalizeBraveHeaders(profile)

	return result
}

func normalizeBraveHeaders(profile *ClientProfile) {
	if profile == nil || profile.Headers == nil {
		return
	}
	if profile.BrowserType != core.BrowserBrave {
		return
	}

	uaLower := strings.ToLower(profile.Headers.UserAgent)
	chua := profile.Headers.SecCHUA

	if strings.Contains(uaLower, "chrome/1.") || strings.Contains(chua, `v="1"`) {
		major := deriveBraveChromiumMajor(profile.BrowserVersion)
		if major <= 0 {
			major = 120
		}

		profile.Headers.UserAgent = fmt.Sprintf(
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%d.0.0.0 Safari/537.36",
			major,
		)
		// for some strict sites, Brave brand identifiers may trigger risk controls; use Chrome brand combination for compatibility.
		profile.Headers.SecCHUA = fmt.Sprintf(`"Chromium";v="%d", "Google Chrome";v="%d"`, major, major)
		if profile.Headers.SecCHUAMobile == "" {
			profile.Headers.SecCHUAMobile = "?0"
		}
		if profile.Headers.SecCHUAPlatform == "" {
			profile.Headers.SecCHUAPlatform = `"Windows"`
		}
	}
}

func deriveBraveChromiumMajor(braveVersion string) int {
	parts := strings.Split(braveVersion, ".")
	if len(parts) < 2 {
		return 0
	}

	// Brave 1.xx typically corresponds to Chromium (xx + 59), e.g. 1.60 -> 119.
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}
	if minor < 40 {
		return 0
	}
	return minor + 59
}

// generateDefaultUserAgent generates default User-Agent
func generateDefaultUserAgent(browserType core.BrowserType) string {
	switch browserType {
	case core.BrowserChrome:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	case core.BrowserFirefox:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0"
	case core.BrowserSafari:
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_2) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15"
	case core.BrowserEdge:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0"
	default:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	}
}

// ValidateProfileList validates the entire profile list
func ValidateProfileList(profiles []ClientProfile) map[string]ProfileValidationResult {
	validator := NewProfileValidator()
	results := make(map[string]ProfileValidationResult)

	for _, p := range profiles {
		results[p.ID] = validator.Validate(p)
	}

	return results
}
