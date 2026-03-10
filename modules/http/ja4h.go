// translated comment
package http

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// translated comment
type JA4HResult struct {
	// translated comment
	Fingerprint string
	// translated comment
	Method string
	// translated comment
	Headers []string
	// translated comment
	CookieNames []string
}

// translated comment
func CalculateJA4H(headers *core.HTTPHeaders, method string) *JA4HResult {
	if headers == nil {
		return nil
	}

	// translated comment
	var headerNames []string
	headerMap := headers.ToMap()
	for name := range headerMap {
		// translated comment
		if strings.ToLower(name) != "cookie" {
			headerNames = append(headerNames, strings.ToLower(name))
		}
	}

	// translated comment
	var cookieNames []string
	if cookie, ok := headerMap["Cookie"]; ok {
		cookieNames = extractCookieNames(cookie)
	}

	// translated comment
	// translated comment
	ja4hString := "ja4h_" +
		strings.ToLower(method) + "_" +
		intToHex(len(headerNames)) + "_" +
		intToHex(len(cookieNames))

	// translated comment
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

// translated comment
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

// translated comment
func intToHex(n int) string {
	if n < 10 {
		return string('0' + byte(n))
	}
	return string('a' + byte(n-10))
}

// translated comment
type HTTP2Fingerprint struct {
	// translated comment
	SettingsHash string
	// translated comment
	PriorityHash string
	// translated comment
	WindowUpdateHash string
	// translated comment
	CombinedHash string
}

// translated comment
func FingerprintHTTP2(settings core.HTTP2Settings, priorities []core.HTTP2Priority, connectionFlow uint32) *HTTP2Fingerprint {
	// translated comment
	settingsStr := settingsToString(settings)
	priorityStr := prioritiesToString(priorities)
	flowStr := uintToString(connectionFlow)

	settingsHash := hashString(settingsStr)
	priorityHash := hashString(priorityStr)
	windowHash := hashString(flowStr)
	combinedHash := hashString(settingsStr + priorityStr + flowStr)

	return &HTTP2Fingerprint{
		SettingsHash:     settingsHash,
		PriorityHash:     priorityHash,
		WindowUpdateHash: windowHash,
		CombinedHash:     combinedHash,
	}
}

// translated comment
func settingsToString(settings core.HTTP2Settings) string {
	return uintToString(settings.HeaderTableSize) + "_" +
		uintToString(settings.EnablePush) + "_" +
		uintToString(settings.MaxConcurrentStreams) + "_" +
		uintToString(settings.InitialWindowSize) + "_" +
		uintToString(settings.MaxFrameSize) + "_" +
		uintToString(settings.MaxHeaderListSize)
}

// translated comment
func prioritiesToString(priorities []core.HTTP2Priority) string {
	var parts []string
	for _, p := range priorities {
		parts = append(parts, uintToString(p.StreamID)+"_"+uintToString(uint32(p.Weight)))
	}
	return strings.Join(parts, ",")
}

// translated comment
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

// translated comment
func hashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:8])
}

// translated comment
type Analyzer struct {
	profile *profiles.ClientProfile
}

// translated comment
func NewAnalyzer(profile *profiles.ClientProfile) *Analyzer {
	return &Analyzer{profile: profile}
}

// translated comment
func (a *Analyzer) AnalyzeJA4H(method string) *JA4HResult {
	if a.profile == nil || a.profile.Headers == nil {
		return nil
	}
	return CalculateJA4H(a.profile.Headers, method)
}

// translated comment
func (a *Analyzer) AnalyzeHTTP2() *HTTP2Fingerprint {
	if a.profile == nil {
		return nil
	}
	return FingerprintHTTP2(a.profile.HTTP2Settings, a.profile.HTTP2Priorities, a.profile.ConnectionFlow)
}

// translated comment
func (a *Analyzer) Fingerprint(method string) map[string]interface{} {
	return map[string]interface{}{
		"ja4h":  a.AnalyzeJA4H(method),
		"http2": a.AnalyzeHTTP2(),
	}
}

// translated comment
func MatchProfile(headers *core.HTTPHeaders) *profiles.ClientProfile {
	if headers == nil {
		return nil
	}

	// translated comment
	ua := headers.UserAgent
	if ua == "" {
		return nil
	}

	// translated comment
	allProfiles := profiles.GetAll()

	// translated comment
	for _, p := range allProfiles {
		if p.Headers != nil && p.Headers.UserAgent == ua {
			return &p
		}
	}

	return nil
}
