package clienthints

// Phase 3: This module has completed basic migration, awaiting deep optimization (see docs/5-process/modularization/PHASE_3_PLAN.md)
import (
	"fmt"
	"time"

	"github.com/vistone/fingerprint/modules/http/legacy/policy"
)

// CHPhase Client Hints lifecycle phase
type CHPhase int

const (
	// PHASE_INITIAL_REQUEST initial request
	PHASE_INITIAL_REQUEST CHPhase = iota
	// PHASE_SERVER_RESPONSE server response
	PHASE_SERVER_RESPONSE
	// PHASE_SUBSEQUENT_REQUESTS subsequent requests
	PHASE_SUBSEQUENT_REQUESTS
	// PHASE_CROSS_ORIGIN_SUB_REQUESTS cross-origin sub-resource requests
	PHASE_CROSS_ORIGIN_SUB_REQUESTS
	// PHASE_TERMINATED lifecycle terminated
	PHASE_TERMINATED
)

// CHLifecycleEvent lifecycle event
type CHLifecycleEvent struct {
	// Timestamp
	Timestamp time.Time

	// Event type
	Type CHPhase

	// Origin URL
	OriginURL string

	// Related hints
	Hints []string

	// Event details
	Details map[string]interface{}

	// Risk indicators
	RiskIndicators []string
}

// ClientHintsLifecycle complete Client Hints lifecycle management
type ClientHintsLifecycle struct {
	// Start time
	StartTime time.Time

	// Current phase
	CurrentPhase CHPhase

	// Main page URL
	PrimaryOriginURL string

	// Negotiation policy
	NegotiationStrategy *NegotiationStrategy

	// Permissions policy
	PermissionsPolicy *policy.PermissionsPolicy

	// Event log
	EventLog []CHLifecycleEvent

	// Currently active hints
	ActiveHints []string

	// Discovered origins
	DiscoveredOrigins []string

	// Lifecycle integrity markers
	IntegrityFlags []string

	// Risk score
	RiskScore float64
}

// CHLifecycleManager lifecycle manager
type CHLifecycleManager struct {
	lifecycles map[string]*ClientHintsLifecycle

	negotiationAnalyzer *CHNegotiationAnalyzer
	policyAnalyzer      *policy.PermissionsPolicyAnalyzer
}

// NewCHLifecycleManager creates lifecycle manager
func NewCHLifecycleManager() *CHLifecycleManager {
	return &CHLifecycleManager{
		lifecycles:          make(map[string]*ClientHintsLifecycle),
		negotiationAnalyzer: NewCHNegotiationAnalyzer(),
		policyAnalyzer:      policy.NewPermissionsPolicyAnalyzer(),
	}
}

// StartLifecycle starts new Client Hints lifecycle
func (m *CHLifecycleManager) StartLifecycle(primaryOriginURL string, initialHints []string) *ClientHintsLifecycle {
	lifecycle := &ClientHintsLifecycle{
		StartTime:         time.Now(),
		CurrentPhase:      PHASE_INITIAL_REQUEST,
		PrimaryOriginURL:  primaryOriginURL,
		EventLog:          []CHLifecycleEvent{},
		ActiveHints:       initialHints,
		DiscoveredOrigins: []string{primaryOriginURL},
		IntegrityFlags:    []string{},
	}

	// Record initial event
	lifecycle.EventLog = append(lifecycle.EventLog, CHLifecycleEvent{
		Timestamp: lifecycle.StartTime,
		Type:      PHASE_INITIAL_REQUEST,
		OriginURL: primaryOriginURL,
		Hints:     initialHints,
		Details: map[string]interface{}{
			"hint_count": len(initialHints),
		},
	})

	m.lifecycles[primaryOriginURL] = lifecycle
	return lifecycle
}

// ProcessServerResponse processes server Accept-CH response
func (m *CHLifecycleManager) ProcessServerResponse(primaryOriginURL string, acceptCHValue string, permissionsPolicyValue string) error {
	lifecycle, exists := m.lifecycles[primaryOriginURL]
	if !exists {
		return fmt.Errorf("lifecycle not found for %s", primaryOriginURL)
	}

	// Process Accept-CH
	strategy := m.negotiationAnalyzer.InitializeFromAcceptCH(acceptCHValue, primaryOriginURL)
	lifecycle.NegotiationStrategy = strategy

	// Process Permissions-Policy
	if permissionsPolicyValue != "" {
		policy := m.policyAnalyzer.ParsePermissionsPolicy(permissionsPolicyValue)
		lifecycle.PermissionsPolicy = policy
	}

	lifecycle.CurrentPhase = PHASE_SERVER_RESPONSE

	// Record event
	lifecycle.EventLog = append(lifecycle.EventLog, CHLifecycleEvent{
		Timestamp: time.Now(),
		Type:      PHASE_SERVER_RESPONSE,
		OriginURL: primaryOriginURL,
		Hints:     strategy.NegotiatedHints,
		Details: map[string]interface{}{
			"requested_hints":  len(strategy.ServerPrefs.LowEntropyHints) + len(strategy.ServerPrefs.HighEntropyHints),
			"negotiated_hints": len(strategy.NegotiatedHints),
			"rejected_hints":   len(strategy.RejectedHints),
		},
		RiskIndicators: strategy.AnomalyFlags,
	})

	return nil
}

// ProcessSubsequentRequest processes Client Hints in subsequent requests
func (m *CHLifecycleManager) ProcessSubsequentRequest(primaryOriginURL string, requestOriginURL string, includedHints []string) error {
	lifecycle, exists := m.lifecycles[primaryOriginURL]
	if !exists {
		return fmt.Errorf("lifecycle not found for %s", primaryOriginURL)
	}

	// Determine if cross-origin request
	isCrossOrigin := primaryOriginURL != requestOriginURL
	var phase CHPhase
	if isCrossOrigin {
		phase = PHASE_CROSS_ORIGIN_SUB_REQUESTS
	} else {
		phase = PHASE_SUBSEQUENT_REQUESTS
	}

	lifecycle.CurrentPhase = phase

	// Add to discovered origins list
	found := false
	for _, origin := range lifecycle.DiscoveredOrigins {
		if origin == requestOriginURL {
			found = true
			break
		}
	}
	if !found {
		lifecycle.DiscoveredOrigins = append(lifecycle.DiscoveredOrigins, requestOriginURL)
	}

	// Check hints integrity
	riskIndicators := []string{}
	if isCrossOrigin {
		// Validate cross-origin delegation
		if lifecycle.NegotiationStrategy != nil {
			err := m.negotiationAnalyzer.HandleCrossOriginDelegation(
				lifecycle.NegotiationStrategy,
				requestOriginURL,
				includedHints,
			)
			if err != nil {
				riskIndicators = append(riskIndicators, fmt.Sprintf("CROSS_ORIGIN_ERROR:%s", err.Error()))
			}
		}

		// Check Permissions-Policy compliance
		if lifecycle.PermissionsPolicy != nil {
			riskIndicators = m.checkPermissionsPolicyCompliance(lifecycle.PermissionsPolicy, includedHints)
		}
	}

	// Check if hints match negotiated ones
	if lifecycle.NegotiationStrategy != nil {
		for _, hint := range includedHints {
			found := false
			for _, negotiated := range lifecycle.NegotiationStrategy.NegotiatedHints {
				if negotiated == hint {
					found = true
					break
				}
			}
			if !found {
				riskIndicators = append(riskIndicators, fmt.Sprintf("UNAUTHORIZED_HINT_SENT:%s", hint))
			}
		}
	}

	// Record event
	lifecycle.EventLog = append(lifecycle.EventLog, CHLifecycleEvent{
		Timestamp: time.Now(),
		Type:      phase,
		OriginURL: requestOriginURL,
		Hints:     includedHints,
		Details: map[string]interface{}{
			"is_cross_origin": isCrossOrigin,
			"hint_count":      len(includedHints),
		},
		RiskIndicators: riskIndicators,
	})

	return nil
}

// checkPermissionsPolicyCompliance checks Permissions-Policy compliance
func (m *CHLifecycleManager) checkPermissionsPolicyCompliance(pol *policy.PermissionsPolicy, hints []string) []string {
	riskIndicators := []string{}

	// Client Hints related features
	chFeatures := map[string]bool{
		"ch-device-memory":        true,
		"ch-dpr":                  true,
		"ch-downlink":             true,
		"ch-ect":                  true,
		"ch-prefers-color-scheme": true,
		"ch-rtt":                  true,
		"ch-ua":                   true,
		"ch-ua-arch":              true,
		"ch-ua-bitness":           true,
		"ch-ua-mobile":            true,
		"ch-ua-model":             true,
		"ch-ua-platform":          true,
		"ch-ua-platform-version":  true,
	}

	for _, hint := range hints {
		// Check permission directive for hint
		featureName := "ch-" + hint
		if !chFeatures[featureName] {
			featureName = hint // Possibly non-standard hint
		}

		if directive, exists := pol.Directives[featureName]; exists {
			if directive.HasNone {
				riskIndicators = append(riskIndicators, fmt.Sprintf("POLICY_VIOLATION:%s", hint))
			}
		}
	}

	return riskIndicators
}

// TerminateLifecycle terminates lifecycle
func (m *CHLifecycleManager) TerminateLifecycle(primaryOriginURL string) (*ClientHintsLifecycle, error) {
	lifecycle, exists := m.lifecycles[primaryOriginURL]
	if !exists {
		return nil, fmt.Errorf("lifecycle not found for %s", primaryOriginURL)
	}

	lifecycle.CurrentPhase = PHASE_TERMINATED

	// Calculate final risk score
	m.calculateFinalRiskScore(lifecycle)

	// Record termination event
	lifecycle.EventLog = append(lifecycle.EventLog, CHLifecycleEvent{
		Timestamp: time.Now(),
		Type:      PHASE_TERMINATED,
		OriginURL: primaryOriginURL,
		Details: map[string]interface{}{
			"duration_seconds":   time.Since(lifecycle.StartTime).Seconds(),
			"total_events":       len(lifecycle.EventLog),
			"discovered_origins": len(lifecycle.DiscoveredOrigins),
		},
	})

	return lifecycle, nil
}

// calculateFinalRiskScore calculates final risk score
func (m *CHLifecycleManager) calculateFinalRiskScore(lifecycle *ClientHintsLifecycle) {
	risk := 0.0

	// Negotiation policy risk
	if lifecycle.NegotiationStrategy != nil {
		risk += lifecycle.NegotiationStrategy.RiskScore * 0.3
	}

	// Permissions policy risk
	if lifecycle.PermissionsPolicy != nil {
		risk += lifecycle.PermissionsPolicy.RiskScore * 0.3
	}

	// Risk indicators in events
	anomalyCount := 0
	for _, event := range lifecycle.EventLog {
		anomalyCount += len(event.RiskIndicators)
	}
	if anomalyCount > 5 {
		risk += 0.2
		lifecycle.IntegrityFlags = append(lifecycle.IntegrityFlags, "EXCESSIVE_ANOMALIES_DETECTED")
	}

	// Too many cross-origin discoveries
	if len(lifecycle.DiscoveredOrigins) > 10 {
		risk += 0.15
		lifecycle.IntegrityFlags = append(lifecycle.IntegrityFlags, "EXCESSIVE_CROSS_ORIGIN_DISCOVERY")
	}

	// Lifecycle too long
	duration := time.Since(lifecycle.StartTime)
	if duration < time.Second {
		// Too fast may indicate automation
		risk += 0.1
		lifecycle.IntegrityFlags = append(lifecycle.IntegrityFlags, "UNUSUALLY_SHORT_LIFECYCLE")
	} else if duration > 24*time.Hour {
		// Sessions over 24 hours may be problematic
		risk += 0.05
		lifecycle.IntegrityFlags = append(lifecycle.IntegrityFlags, "UNUSUALLY_LONG_LIFECYCLE")
	}

	lifecycle.RiskScore = risk
}

// GetLifecycleReport gets lifecycle report
func (m *CHLifecycleManager) GetLifecycleReport(primaryOriginURL string) (*ClientHintsLifecycle, error) {
	lifecycle, exists := m.lifecycles[primaryOriginURL]
	if !exists {
		return nil, fmt.Errorf("lifecycle not found for %s", primaryOriginURL)
	}

	return lifecycle, nil
}

// GetLifecycleMetrics gets lifecycle metrics
func (m *CHLifecycleManager) GetLifecycleMetrics(lifecycle *ClientHintsLifecycle) map[string]interface{} {
	metrics := make(map[string]interface{})

	metrics["primary_origin"] = lifecycle.PrimaryOriginURL
	metrics["current_phase"] = lifecycle.CurrentPhase
	metrics["total_events"] = len(lifecycle.EventLog)
	metrics["discovered_origins"] = len(lifecycle.DiscoveredOrigins)
	metrics["active_hints"] = len(lifecycle.ActiveHints)
	metrics["risk_score"] = lifecycle.RiskScore
	metrics["duration_seconds"] = time.Since(lifecycle.StartTime).Seconds()
	metrics["integrity_flags"] = lifecycle.IntegrityFlags

	if lifecycle.NegotiationStrategy != nil {
		metrics["negotiated_hints"] = len(lifecycle.NegotiationStrategy.NegotiatedHints)
		metrics["rejected_hints"] = len(lifecycle.NegotiationStrategy.RejectedHints)
	}

	if lifecycle.PermissionsPolicy != nil {
		metrics["policy_features"] = len(lifecycle.PermissionsPolicy.Directives)
		metrics["policy_anomalies"] = len(lifecycle.PermissionsPolicy.AnomalyFlags)
	}

	return metrics
}

// GetSummary gets summary
func (m *CHLifecycleManager) GetSummary(lifecycle *ClientHintsLifecycle) string {
	phaseStr := ""
	switch lifecycle.CurrentPhase {
	case PHASE_INITIAL_REQUEST:
		phaseStr = "Initial Request"
	case PHASE_SERVER_RESPONSE:
		phaseStr = "Server Response"
	case PHASE_SUBSEQUENT_REQUESTS:
		phaseStr = "Subsequent Requests"
	case PHASE_CROSS_ORIGIN_SUB_REQUESTS:
		phaseStr = "Cross-Origin Sub-Resources"
	case PHASE_TERMINATED:
		phaseStr = "Terminated"
	}

	return fmt.Sprintf(
		"Origin: %s | Phase: %s | Discovered Origins: %d | Events: %d | Risk Score: %.2f",
		lifecycle.PrimaryOriginURL,
		phaseStr,
		len(lifecycle.DiscoveredOrigins),
		len(lifecycle.EventLog),
		lifecycle.RiskScore,
	)
}
