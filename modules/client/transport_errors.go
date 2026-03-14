package client

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"

	fhttp "github.com/bogdanfinn/fhttp"
	tls "github.com/bogdanfinn/utls"
	"github.com/vistone/fingerprint/modules/profiles"
	legacyprofiles "github.com/vistone/fingerprint/modules/profiles/legacy"
)

func (st *SmartTransport) Close() error {
	return nil
}

// shouldFallbackToHTTP1Compat determine whether to attempt HTTP/1.1 compatibility path with standard TLS.
func shouldFallbackToHTTP1Compat(err error) bool {
	if err == nil {
		return false
	}

	errLower := strings.ToLower(err.Error())

	patterns := []string{
		"http/1.x transport connection broken",
		"malformed http response",
		"http2_handshake_failed",
		"first record does not look like a tls handshake",
		"tls handshake",
		"alpn",
	}

	for _, p := range patterns {
		if strings.Contains(errLower, p) {
			return true
		}
	}

	return false
}

// ErrorType errortype
type ErrorType int

const (
	ErrorTypeUnknown ErrorType = iota
	ErrorTypeProtocol
	ErrorTypeNetwork
	ErrorTypeTLS
	ErrorTypeTimeout
	ErrorTypeCanceled
	ErrorTypeDNS
)

// classifyError classify error
func classifyError(err error) ErrorType {
	if err == nil {
		return ErrorTypeUnknown
	}

	errStr := err.Error()

	// determine based on context error
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorTypeTimeout
	}
	if errors.Is(err, context.Canceled) {
		return ErrorTypeCanceled
	}

	// determine based on error information
	timeoutPatterns := []string{"timeout", "i/o timeout", "deadline exceeded"}
	for _, pattern := range timeoutPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return ErrorTypeTimeout
		}
	}

	dnsPatterns := []string{"no such host", "lookup", "nameserver"}
	for _, pattern := range dnsPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return ErrorTypeDNS
		}
	}

	protocolPatterns := []string{
		"PROTOCOL_ERROR",
		"http2:",
		"stream error",
		"broken pipe",
		"reset by peer",
		"connection reset",
	}
	for _, pattern := range protocolPatterns {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(pattern)) {
			return ErrorTypeProtocol
		}
	}

	tlsPatterns := []string{"tls", "certificate", "handshake"}
	for _, pattern := range tlsPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return ErrorTypeTLS
		}
	}

	networkPatterns := []string{
		"connection refused",
		"connection aborted",
		"no route to host",
		"network is unreachable",
		"EOF",
	}
	for _, pattern := range networkPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return ErrorTypeNetwork
		}
	}

	return ErrorTypeUnknown
}

// isProtocolError check if it is protocol error (deprecated, use classifyError)
func isProtocolError(err error) bool {
	if err == nil {
		return false
	}
	return classifyError(err) == ErrorTypeProtocol
}

func shouldRetryWithHTTP1(resp *fhttp.Response) bool {
	if resp == nil {
		return false
	}
	return resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusMisdirectedRequest
}

// getProfileClientHelloSpec fine-grained ClientHello specification based on profile resolution that can be reused.
// when legacy map does not exist return nil, caller will fall back to current Auto ClientHello behavior.
func (st *SmartTransport) getProfileClientHelloSpec() (*tls.ClientHelloSpec, error) {
	legacyProfileID := resolveLegacyProfileID(st.profile)
	if legacyProfileID == "" {
		return nil, nil
	}

	legacyProfile, ok := legacyprofiles.GetClientProfile(legacyProfileID)
	if !ok {
		return nil, nil
	}

	spec, err := legacyProfile.GetClientHelloSpec()
	if err != nil {
		// some legacy profiles have not yet implemented ToSpec, when encountered maintain compatibility fallback to Auto ClientHello.
		return nil, nil
	}

	return &spec, nil
}

// resolveLegacyProfileID return legacy profile ID usable for uTLS ApplyPreset.
// rule: prioritize precise ID, then match nearby by browser type and major version number, finally fall back to latest available version for that browser.
func resolveLegacyProfileID(profile profiles.ClientProfile) string {
	if profile.ID != "" {
		if _, ok := legacyprofiles.GetClientProfile(profile.ID); ok {
			return profile.ID
		}
	}

	browser := strings.ToLower(string(profile.BrowserType))
	if browser == "" {
		return ""
	}
	prefix := browser + "_"

	hasTargetMajor := false
	targetMajor := 0
	if profile.BrowserVersion != "" {
		if major, ok := parseLeadingMajor(profile.BrowserVersion); ok {
			targetMajor = major
			hasTargetMajor = true
		}
	}

	bestUnderOrEqualID := ""
	bestUnderOrEqualMajor := -1
	latestID := ""
	latestMajor := -1

	legacyIDs := legacyprofiles.GetAllProfiles()
	sort.Strings(legacyIDs)
	for _, id := range legacyIDs {
		idLower := strings.ToLower(id)
		if !strings.HasPrefix(idLower, prefix) {
			continue
		}
		if browser == "safari" && (strings.HasPrefix(idLower, "safari_ios_") || strings.HasPrefix(idLower, "safari_ipad_")) {
			continue
		}

		remainder := strings.TrimPrefix(idLower, prefix)
		major, ok := parseLeadingMajor(remainder)
		if !ok {
			continue
		}

		if major > latestMajor {
			latestMajor = major
			latestID = id
		}

		if hasTargetMajor && major <= targetMajor && major > bestUnderOrEqualMajor {
			bestUnderOrEqualMajor = major
			bestUnderOrEqualID = id
		}
	}

	if bestUnderOrEqualID != "" {
		return bestUnderOrEqualID
	}
	return latestID
}

func parseLeadingMajor(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}

	idx := 0
	for idx < len(value) {
		c := value[idx]
		if c < '0' || c > '9' {
			break
		}
		idx++
	}
	if idx == 0 {
		return 0, false
	}

	major, err := strconv.Atoi(value[:idx])
	if err != nil {
		return 0, false
	}
	return major, true
}

// getClientHelloID getbrowser ClientHello ID
func getClientHelloID(browserType string) tls.ClientHelloID {
	switch browserType {
	case "chrome":
		return tls.HelloChrome_Auto
	case "firefox":
		return tls.HelloFirefox_Auto
	case "safari":
		return tls.HelloSafari_Auto
	case "edge":
		return tls.HelloChrome_Auto
	case "opera":
		return tls.HelloChrome_Auto
	default:
		return tls.HelloChrome_Auto
	}
}

func convertTLSVersion(version uint16) uint16 {
	switch version {
	case 0x0301:
		return tls.VersionTLS10
	case 0x0302:
		return tls.VersionTLS11
	case 0x0303:
		return tls.VersionTLS12
	case 0x0304:
		return tls.VersionTLS13
	default:
		return tls.VersionTLS12
	}
}

func convertCipherSuites(suites []uint16) []uint16 {
	result := make([]uint16, 0, len(suites))
	for _, suite := range suites {
		switch suite {
		case 0x1301:
			result = append(result, tls.TLS_AES_128_GCM_SHA256)
		case 0x1302:
			result = append(result, tls.TLS_AES_256_GCM_SHA384)
		case 0x1303:
			result = append(result, tls.TLS_CHACHA20_POLY1305_SHA256)
		case 0xc02b:
			result = append(result, tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256)
		case 0xc02f:
			result = append(result, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)
		case 0xc02c:
			result = append(result, tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384)
		case 0xc030:
			result = append(result, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384)
		default:
			result = append(result, suite)
		}
	}
	return result
}
