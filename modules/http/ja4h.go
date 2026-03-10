// Package http provides HTTP fingerprint generation
package http

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// JA4HResult represents a JA4H fingerprint result
type JA4HResult struct {
	// Fingerprint is the JA4H fingerprint string
	Fingerprint string
	// Method is the HTTP method
	Method string
	// Headers is the list of HTTP header names
	Headers []string
	// CookieNames is the list of Cookie names
	CookieNames []string
}

// CalculateJA4H computes the JA4H fingerprint
func CalculateJA4H(headers *core.HTTPHeaders, method string) *JA4HResult {
	if headers == nil {
		return nil
	}

	// collect all header names
	var headerNames []string
	headerMap := headers.ToMap()
	for name := range headerMap {
		// skip Cookie header (handled separately)
		if strings.ToLower(name) != "cookie" {
			headerNames = append(headerNames, strings.ToLower(name))
		}
	}

	// extract Cookie names
	var cookieNames []string
	if cookie, ok := headerMap["Cookie"]; ok {
		cookieNames = extractCookieNames(cookie)
	}

	// build the JA4H string
	// format: ja4h_<method>_<header_count>_<cookie_count>_<accept_language_hash>
	ja4hString := "ja4h_" +
		strings.ToLower(method) + "_" +
		intToHex(len(headerNames)) + "_" +
		intToHex(len(cookieNames))

	// compute Accept-Language hash
	if lang := headers.AcceptLanguage; lang != "" {
		hash := sha256.Sum256([]byte(lang))
		ja4hString += "_" + hex.EncodeToString(hash[:6])
	} else {
		ja4hString += "_000000"
	}

	return &JA4HResult{
		Fingerprint: ja4hString,
		Method:      method,
		Headers:     headerNames,
		CookieNames: cookieNames,
	}
}

// extractCookieNames extracts Cookie names from a cookie header value
func extractCookieNames(cookie string) []string {
	var names []string
	parts := strings.Split(cookie, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if idx := strings.Index(part, "="); idx > 0 {
			name := strings.TrimSpace(part[:idx])
			if name != "" {
				names = append(names, name)
			}
		}
	}
	return names
}

// intToHex converts an integer to a hex string
func intToHex(n int) string {
	if n < 10 {
		return string('0' + byte(n))
	}
	return string('a' + byte(n-10))
}

// HTTP2Fingerprint represents an HTTP/2 fingerprint
type HTTP2Fingerprint struct {
	// SettingsHash is the hash of the SETTINGS frame values
	SettingsHash string
	// PriorityHash is the hash of the PRIORITY frames
	PriorityHash string
	// WindowUpdateHash is the hash of the WINDOW_UPDATE frame
	WindowUpdateHash string
	// CombinedHash is the combined fingerprint hash
	CombinedHash string
	// SettingsOrder is the order of SETTINGS frame parameter IDs (Akamai-style)
	SettingsOrder []uint16
	// PseudoHeaderOrder is the order of HTTP/2 pseudo-headers
	PseudoHeaderOrder []string
	// AkamaiFingerprint is the combined Akamai-style HTTP/2 fingerprint string
	AkamaiFingerprint string
}

// defaultSettingsOrder is the standard SETTINGS parameter ID order:
// 1=HEADER_TABLE_SIZE, 2=ENABLE_PUSH, 3=MAX_CONCURRENT_STREAMS,
// 4=INITIAL_WINDOW_SIZE, 5=MAX_FRAME_SIZE, 6=MAX_HEADER_LIST_SIZE
var defaultSettingsOrder = []uint16{1, 2, 3, 4, 5, 6}

// defaultPseudoHeaderOrder is the default HTTP/2 pseudo-header order
var defaultPseudoHeaderOrder = []string{":method", ":authority", ":scheme", ":path"}

// FingerprintHTTP2 generates an HTTP/2 fingerprint
func FingerprintHTTP2(settings core.HTTP2Settings, priorities []core.HTTP2Priority, connectionFlow uint32) *HTTP2Fingerprint {
	return FingerprintHTTP2Extended(settings, priorities, connectionFlow, nil, nil)
}

// FingerprintHTTP2Extended generates an HTTP/2 fingerprint with Akamai-style details.
// settingsOrder specifies the order of SETTINGS parameter IDs as sent by the client;
// if nil, the default order (1,2,3,4,5,6) is used.
// pseudoHeaderOrder specifies the order of pseudo-headers; if nil, the default
// order (:method, :authority, :scheme, :path) is used.
func FingerprintHTTP2Extended(settings core.HTTP2Settings, priorities []core.HTTP2Priority, connectionFlow uint32, settingsOrder []uint16, pseudoHeaderOrder []string) *HTTP2Fingerprint {
	if settingsOrder == nil {
		settingsOrder = defaultSettingsOrder
	}
	if pseudoHeaderOrder == nil {
		pseudoHeaderOrder = defaultPseudoHeaderOrder
	}

	settingsStr := settingsToString(settings)
	priorityStr := prioritiesToString(priorities)
	flowStr := uintToString(connectionFlow)

	settingsHash := hashString(settingsStr)
	priorityHash := hashString(priorityStr)
	windowHash := hashString(flowStr)
	combinedHash := hashString(settingsStr + priorityStr + flowStr)

	akamaiStr := buildAkamaiFingerprint(settings, settingsOrder, priorities, connectionFlow, pseudoHeaderOrder)

	return &HTTP2Fingerprint{
		SettingsHash:      settingsHash,
		PriorityHash:      priorityHash,
		WindowUpdateHash:  windowHash,
		CombinedHash:      combinedHash,
		SettingsOrder:     settingsOrder,
		PseudoHeaderOrder: pseudoHeaderOrder,
		AkamaiFingerprint: akamaiStr,
	}
}

// settingIDValue returns the SETTINGS value for a given parameter ID
func settingIDValue(settings core.HTTP2Settings, id uint16) uint32 {
	switch id {
	case 1:
		return settings.HeaderTableSize
	case 2:
		return settings.EnablePush
	case 3:
		return settings.MaxConcurrentStreams
	case 4:
		return settings.InitialWindowSize
	case 5:
		return settings.MaxFrameSize
	case 6:
		return settings.MaxHeaderListSize
	default:
		return 0
	}
}

// buildAkamaiFingerprint constructs an Akamai-style HTTP/2 fingerprint string.
// Format: S[id:val;...]|WU[window_update]|P[stream:depends:exclusive:weight,...]|PS[h1,h2,...]
func buildAkamaiFingerprint(settings core.HTTP2Settings, settingsOrder []uint16, priorities []core.HTTP2Priority, connectionFlow uint32, pseudoHeaderOrder []string) string {
	// Settings section
	var settingsParts []string
	for _, id := range settingsOrder {
		val := settingIDValue(settings, id)
		settingsParts = append(settingsParts, uintToString(uint32(id))+":"+uintToString(val))
	}
	sSection := "S[" + strings.Join(settingsParts, ";") + "]"

	// Window Update section
	wuSection := "WU[" + uintToString(connectionFlow) + "]"

	// Priority section
	var priorityParts []string
	for _, p := range priorities {
		exc := "0"
		if p.Exclusive {
			exc = "1"
		}
		priorityParts = append(priorityParts,
			uintToString(p.StreamID)+":"+
				uintToString(p.DependsOn)+":"+
				exc+":"+
				uintToString(uint32(p.Weight)))
	}
	pSection := "P[" + strings.Join(priorityParts, ",") + "]"

	// Pseudo-header order section
	psSection := "PS[" + strings.Join(pseudoHeaderOrder, ",") + "]"

	return sSection + "|" + wuSection + "|" + pSection + "|" + psSection
}

// settingsToString converts HTTP/2 SETTINGS values to a string
func settingsToString(settings core.HTTP2Settings) string {
	return uintToString(settings.HeaderTableSize) + "_" +
		uintToString(settings.EnablePush) + "_" +
		uintToString(settings.MaxConcurrentStreams) + "_" +
		uintToString(settings.InitialWindowSize) + "_" +
		uintToString(settings.MaxFrameSize) + "_" +
		uintToString(settings.MaxHeaderListSize)
}

// prioritiesToString converts HTTP/2 PRIORITY frames to a string
func prioritiesToString(priorities []core.HTTP2Priority) string {
	var parts []string
	for _, p := range priorities {
		parts = append(parts, uintToString(p.StreamID)+"_"+uintToString(uint32(p.Weight)))
	}
	return strings.Join(parts, ",")
}

// uintToString converts a uint32 to its decimal string representation
func uintToString(n uint32) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	return string(result)
}

// hashString computes a truncated SHA-256 hash of a string
func hashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:8])
}

// Analyzer is an HTTP fingerprint analyzer
type Analyzer struct {
	profile *profiles.ClientProfile
}

// NewAnalyzer creates a new HTTP analyzer
func NewAnalyzer(profile *profiles.ClientProfile) *Analyzer {
	return &Analyzer{profile: profile}
}

// AnalyzeJA4H analyzes the JA4H fingerprint
func (a *Analyzer) AnalyzeJA4H(method string) *JA4HResult {
	if a.profile == nil || a.profile.Headers == nil {
		return nil
	}
	return CalculateJA4H(a.profile.Headers, method)
}

// AnalyzeHTTP2 analyzes the HTTP/2 fingerprint
func (a *Analyzer) AnalyzeHTTP2() *HTTP2Fingerprint {
	if a.profile == nil {
		return nil
	}
	return FingerprintHTTP2Extended(
		a.profile.HTTP2Settings,
		a.profile.HTTP2Priorities,
		a.profile.ConnectionFlow,
		nil,
		a.profile.PseudoHeaderOrder,
	)
}

// Fingerprint generates a complete HTTP fingerprint
func (a *Analyzer) Fingerprint(method string) map[string]interface{} {
	return map[string]interface{}{
		"ja4h":  a.AnalyzeJA4H(method),
		"http2": a.AnalyzeHTTP2(),
	}
}

// MatchProfile matches a browser profile from HTTP headers
func MatchProfile(headers *core.HTTPHeaders) *profiles.ClientProfile {
	if headers == nil {
		return nil
	}

	// match based on User-Agent
	ua := headers.UserAgent
	if ua == "" {
		return nil
	}

	// retrieve all profiles
	allProfiles := profiles.GetAll()

	// simple matching logic (a more sophisticated algorithm should be used in production)
	for _, p := range allProfiles {
		if p.Headers != nil && p.Headers.UserAgent == ua {
			return &p
		}
	}

	return nil
}
