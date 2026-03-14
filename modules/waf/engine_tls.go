package waf

import (
	"net/http"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
)

// TLSEngine analyzes TLS layer characteristics
type TLSEngine struct {
	blacklistJA3 []string
	blacklistJA4 []string
}

// TLSResult contains TLS analysis results
type TLSResult struct {
	Score        float64
	Factors      []core.RiskFactor
	JA3          string
	JA4          string
	IsSuspicious bool
}

// NewTLSEngine creates a new TLS analysis engine
func NewTLSEngine(blacklistJA3, blacklistJA4 []string) *TLSEngine {
	return &TLSEngine{
		blacklistJA3: blacklistJA3,
		blacklistJA4: blacklistJA4,
	}
}

// Analyze performs TLS layer analysis
func (e *TLSEngine) Analyze(req *http.Request) *TLSResult {
	result := &TLSResult{
		Score:   0,
		Factors: make([]core.RiskFactor, 0),
	}

	// Extract JA3 fingerprint from headers
	ja3 := req.Header.Get("X-JA3-Fingerprint")
	if ja3 == "" {
		ja3 = req.Header.Get("X-TLS-JA3")
	}
	result.JA3 = ja3

	// Extract JA4 fingerprint from headers
	ja4 := req.Header.Get("X-JA4-Fingerprint")
	if ja4 == "" {
		ja4 = req.Header.Get("X-TLS-JA4")
	}
	result.JA4 = ja4

	// Check JA3 blacklist
	if ja3 != "" && e.isBlacklistedJA3(ja3) {
		result.Score += 0.9
		result.IsSuspicious = true
		result.Factors = append(result.Factors, core.RiskFactor{
			Name:        "blacklisted_ja3",
			Weight:      0.9,
			Description: "JA3 fingerprint is blacklisted (known bot/tool)",
		})
	}

	// Check JA4 blacklist
	if ja4 != "" && e.isBlacklistedJA4(ja4) {
		result.Score += 0.9
		result.IsSuspicious = true
		result.Factors = append(result.Factors, core.RiskFactor{
			Name:        "blacklisted_ja4",
			Weight:      0.9,
			Description: "JA4 fingerprint is blacklisted (known bot/tool)",
		})
	}

	// Check for missing TLS info (might indicate direct HTTP)
	if ja3 == "" && ja4 == "" && req.TLS == nil {
		// Could be a direct HTTP request or missing TLS termination info
		result.Factors = append(result.Factors, core.RiskFactor{
			Name:        "missing_tls_info",
			Weight:      0.2,
			Description: "TLS fingerprint information not available",
		})
	}

	// Check TLS version
	if req.TLS != nil {
		version := req.TLS.Version
		if version < 0x0303 { // TLS 1.2 = 0x0303
			// Old TLS version
			result.Score += 0.3
			result.Factors = append(result.Factors, core.RiskFactor{
				Name:        "old_tls_version",
				Weight:      0.3,
				Description: "Outdated TLS version",
			})
		}
	}

	return result
}

func (e *TLSEngine) isBlacklistedJA3(ja3 string) bool {
	for _, b := range e.blacklistJA3 {
		if strings.Contains(ja3, b) || b == ja3 {
			return true
		}
	}
	return false
}

func (e *TLSEngine) isBlacklistedJA4(ja4 string) bool {
	for _, b := range e.blacklistJA4 {
		if strings.Contains(ja4, b) || b == ja4 {
			return true
		}
	}
	return false
}
