package clienthints

// Phase 3: This module has completed basic migration, awaiting deep optimization (see docs/5-process/modularization/PHASE_3_PLAN.md)
import (
	"fmt"
	"strings"
)

// NegotiationState Client Hints negotiation state
type NegotiationState int

const (
	// NEGOTIATION_INIT initial state (no negotiation)
	NEGOTIATION_INIT NegotiationState = iota
	// NEGOTIATION_REQUESTED server requested hints
	NEGOTIATION_REQUESTED
	// NEGOTIATION_ACCEPTED client accepted
	NEGOTIATION_ACCEPTED
	// NEGOTIATION_REJECTED client rejected
	NEGOTIATION_REJECTED
	// NEGOTIATION_DELEGATED delegated to other origins
	NEGOTIATION_DELEGATED
)

// ServerPreferences server preferences for Client Hints
type ServerPreferences struct {
	// Requested low-entropy hints
	LowEntropyHints []string

	// Requested high-entropy hints
	HighEntropyHints []string

	// Priority (0-100, higher is more priority)
	Priority int

	// Retention duration (seconds), -1 means persistent
	CacheDuration int

	// Cross-origin delegation configuration
	DelegateToOrigins []string

	// Whether to allow insecure connections
	AllowInsecure bool

	// Additional description
	Description string
}

// ClientCapabilities client Client Hints capability declaration
type ClientCapabilities struct {
	// Supported low-entropy hints
	SupportedLowEntropy []string

	// Supported high-entropy hints
	SupportedHighEntropy []string

	// Browser-specific limitations
	BrowserLimitations []string

	// Privacy protection level (0-100)
	PrivacyLevel int

	// User consent status
	UserConsent bool

	// Device type
	DeviceType string
}

// NegotiationStrategy Client Hints negotiation strategy
type NegotiationStrategy struct {
	// Negotiation state
	State NegotiationState

	// Server preferences
	ServerPrefs *ServerPreferences

	// Client capabilities
	ClientCaps *ClientCapabilities

	// Negotiated hints set
	NegotiatedHints []string

	// Rejected hints
	RejectedHints []string

	// Next negotiation time (Unix seconds)
	NextNegotiationTime int64

	// Negotiation history
	NegotiationHistory []NegotiationRecord

	// Risk score
	RiskScore float64

	// Anomaly flags
	AnomalyFlags []string
}

// NegotiationRecord single negotiation record
type NegotiationRecord struct {
	// Timestamp
	Timestamp int64

	// Requested hints
	RequestedHints []string

	// Provided hints
	ProvidedHints []string

	// Decision reason
	Decision string

	// Risk indicators
	RiskIndicators []string
}

// CHNegotiationAnalyzer Client Hints negotiation analyzer
type CHNegotiationAnalyzer struct {
	strategies map[string]*NegotiationStrategy
}

// NewCHNegotiationAnalyzer creates negotiation analyzer
func NewCHNegotiationAnalyzer() *CHNegotiationAnalyzer {
	return &CHNegotiationAnalyzer{
		strategies: make(map[string]*NegotiationStrategy),
	}
}

// InitializeFromAcceptCH initializes negotiation from Accept-CH response header
func (a *CHNegotiationAnalyzer) InitializeFromAcceptCH(acceptCHValue string, origin string) *NegotiationStrategy {
	prefs := a.parseAcceptCH(acceptCHValue)

	strategy := &NegotiationStrategy{
		State:              NEGOTIATION_REQUESTED,
		ServerPrefs:        prefs,
		ClientCaps:         a.getDefaultCapabilities(),
		NegotiatedHints:    []string{},
		RejectedHints:      []string{},
		NegotiationHistory: []NegotiationRecord{},
		AnomalyFlags:       []string{},
	}

	a.strategies[origin] = strategy

	// Evaluate anomaly flags
	if len(prefs.HighEntropyHints) > 5 {
		strategy.AnomalyFlags = append(strategy.AnomalyFlags, "EXCESSIVE_HIGH_ENTROPY_HINTS")
	}

	a.evaluateNegotiationRisk(strategy)

	return strategy
}

// parseAcceptCH parses Accept-CH header
func (a *CHNegotiationAnalyzer) parseAcceptCH(acceptCHValue string) *ServerPreferences {
	prefs := &ServerPreferences{
		LowEntropyHints:   []string{},
		HighEntropyHints:  []string{},
		Priority:          50,
		CacheDuration:     -1,
		DelegateToOrigins: []string{},
	}

	if acceptCHValue == "" {
		return prefs
	}

	// Standard Client Hints
	lowEntropyStandard := map[string]bool{
		"sec-ch-ua":          true,
		"sec-ch-ua-mobile":   true,
		"sec-ch-ua-platform": true,
		"user-agent":         true,
	}

	highEntropyStandard := map[string]bool{
		"sec-ch-ua-full-version":     true,
		"sec-ch-ua-platform-version": true,
		"sec-ch-ua-model":            true,
		"sec-ch-ua-arch":             true,
		"sec-ch-ua-bitness":          true,
		"dpr":                        true,
		"viewport-width":             true,
		"device-memory":              true,
		"downlink":                   true,
		"ect":                        true,
		"rtt":                        true,
		"save-data":                  true,
	}

	// Parse hints list
	parts := strings.Split(acceptCHValue, ",")
	for _, part := range parts {
		hint := strings.TrimSpace(part)
		if hint == "" {
			continue
		}

		// Remove quotes and extra markers
		hint = strings.Trim(hint, `"`)
		hint = strings.ToLower(hint)

		if lowEntropyStandard[hint] {
			prefs.LowEntropyHints = append(prefs.LowEntropyHints, hint)
		} else if highEntropyStandard[hint] {
			prefs.HighEntropyHints = append(prefs.HighEntropyHints, hint)
		}
	}

	// Evaluate priority
	if len(prefs.HighEntropyHints) > 5 {
		prefs.Priority = 80
	}

	return prefs
}

// getDefaultCapabilities gets default client capabilities
func (a *CHNegotiationAnalyzer) getDefaultCapabilities() *ClientCapabilities {
	return &ClientCapabilities{
		SupportedLowEntropy: []string{
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
		},
		SupportedHighEntropy: []string{
			"sec-ch-ua-full-version",
			"sec-ch-ua-platform-version",
			"sec-ch-ua-arch",
			"sec-ch-ua-bitness",
			"sec-ch-ua-model",
		},
		BrowserLimitations: []string{},
		PrivacyLevel:       70,
		UserConsent:        false,
		DeviceType:         "desktop",
	}
}

// DecideHints decides hints to send based on server request and client capabilities
func (a *CHNegotiationAnalyzer) DecideHints(strategy *NegotiationStrategy, userPreference string) []string {
	decided := []string{}

	// First process low-entropy hints (unconditionally allowed)
	for _, hint := range strategy.ServerPrefs.LowEntropyHints {
		if a.isSupportedHint(strategy.ClientCaps, hint) {
			decided = append(decided, hint)
		} else {
			strategy.RejectedHints = append(strategy.RejectedHints, hint)
		}
	}

	// Process high-entropy hints (require user consent)
	for _, hint := range strategy.ServerPrefs.HighEntropyHints {
		if a.isSupportedHint(strategy.ClientCaps, hint) {
			// Check user consent
			if strategy.ClientCaps.UserConsent || userPreference == "allow-all" {
				decided = append(decided, hint)
			} else {
				strategy.RejectedHints = append(strategy.RejectedHints, hint)
				strategy.AnomalyFlags = append(strategy.AnomalyFlags, "HIGH_ENTROPY_WITHOUT_CONSENT")
			}
		}
	}

	strategy.NegotiatedHints = decided
	strategy.State = NEGOTIATION_ACCEPTED

	return decided
}

// isSupportedHint checks if hint is supported
func (a *CHNegotiationAnalyzer) isSupportedHint(caps *ClientCapabilities, hint string) bool {
	for _, h := range caps.SupportedLowEntropy {
		if h == hint {
			return true
		}
	}
	for _, h := range caps.SupportedHighEntropy {
		if h == hint {
			return true
		}
	}
	return false
}

// HandleCrossOriginDelegation handles cross-origin delegation
func (a *CHNegotiationAnalyzer) HandleCrossOriginDelegation(strategy *NegotiationStrategy, delegateOrigin string, delegateHints []string) error {
	// Verify if delegation origin is authorized
	isAuthorized := false
	for _, allowed := range strategy.ServerPrefs.DelegateToOrigins {
		if allowed == delegateOrigin || allowed == "*" {
			isAuthorized = true
			break
		}
	}

	if !isAuthorized {
		strategy.AnomalyFlags = append(strategy.AnomalyFlags, fmt.Sprintf("UNAUTHORIZED_DELEGATION_ATTEMPT:%s", delegateOrigin))
		return fmt.Errorf("delegation to %s not authorized", delegateOrigin)
	}

	strategy.State = NEGOTIATION_DELEGATED

	// Verify if delegated hints are within server-allowed scope
	for _, delegatedHint := range delegateHints {
		found := false
		for _, negotiated := range strategy.NegotiatedHints {
			if negotiated == delegatedHint {
				found = true
				break
			}
		}
		if !found {
			strategy.AnomalyFlags = append(strategy.AnomalyFlags, fmt.Sprintf("UNAUTHORIZED_HINT_DELEGATION:%s", delegatedHint))
		}
	}

	return nil
}

// evaluateNegotiationRisk evaluates negotiation risk
func (a *CHNegotiationAnalyzer) evaluateNegotiationRisk(strategy *NegotiationStrategy) {
	risk := 0.0

	// Check anomaly count
	if len(strategy.AnomalyFlags) > 2 {
		risk += 0.2
	}

	// Check excessive high-entropy hints (possible fingerprinting attempt)
	if len(strategy.ServerPrefs.HighEntropyHints) > 8 {
		risk += 0.3
		strategy.AnomalyFlags = append(strategy.AnomalyFlags, "EXCESSIVE_FINGERPRINTING_ATTEMPT")
	}

	// Check delegation to too many origins
	if len(strategy.ServerPrefs.DelegateToOrigins) > 5 {
		risk += 0.2
		strategy.AnomalyFlags = append(strategy.AnomalyFlags, "EXCESSIVE_DOMAIN_DELEGATION")
	}

	// Check high-entropy hints on insecure connections
	if strategy.ServerPrefs.AllowInsecure && len(strategy.ServerPrefs.HighEntropyHints) > 0 {
		risk += 0.4
		strategy.AnomalyFlags = append(strategy.AnomalyFlags, "HIGH_ENTROPY_OVER_INSECURE")
	}

	// Cache duration too long
	if strategy.ServerPrefs.CacheDuration > 31536000 { // Over 1 year
		risk += 0.15
	}

	strategy.RiskScore = risk
}

// GetNegotiationSummary gets negotiation summary
func (a *CHNegotiationAnalyzer) GetNegotiationSummary(strategy *NegotiationStrategy) string {
	stateStr := ""
	switch strategy.State {
	case NEGOTIATION_INIT:
		stateStr = "Initial"
	case NEGOTIATION_REQUESTED:
		stateStr = "Requested"
	case NEGOTIATION_ACCEPTED:
		stateStr = "Accepted"
	case NEGOTIATION_REJECTED:
		stateStr = "Rejected"
	case NEGOTIATION_DELEGATED:
		stateStr = "Delegated"
	}

	return fmt.Sprintf(
		"State: %s | Negotiated Hints: %d | Rejected Hints: %d | Anomaly Flags: %d | Risk Score: %.2f",
		stateStr,
		len(strategy.NegotiatedHints),
		len(strategy.RejectedHints),
		len(strategy.AnomalyFlags),
		strategy.RiskScore,
	)
}

// GetAllStrategies gets negotiation strategies for all origins
func (a *CHNegotiationAnalyzer) GetAllStrategies() map[string]*NegotiationStrategy {
	return a.strategies
}
