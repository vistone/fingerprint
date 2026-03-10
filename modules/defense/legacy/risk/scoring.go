package risk

// Phase 3: This module has completed basic migration, pending deep optimization (see docs/5-process/modularization/PHASE_3_PLAN.md)
import (
	"fmt"
	"math"

	"github.com/vistone/fingerprint/modules/http/legacy/http2"
	"github.com/vistone/fingerprint/modules/http/legacy/ja4h"
	"github.com/vistone/fingerprint/modules/network/quic"
	"github.com/vistone/fingerprint/modules/tls/legacy/ech"
	"github.com/vistone/fingerprint/modules/tls/legacy/ja4s"
)

// RiskScore represents comprehensive risk scoring result
type RiskScore struct {
	// Overall risk score (0.0-1.0)
	TotalScore float64

	// Threat level
	ThreatLevel string // "safe", "low", "medium", "high", "critical"

	// Risk scores by dimensions
	Dimensions RiskDimensions

	// Risk factor details
	RiskFactors []RiskFactor

	// Anomalous metrics
	AnomalyCount int

	// Confidence (0.0-1.0)
	Confidence float64

	// Recommended actions
	Recommendations []string
}

// RiskDimensions represents risk dimension scores
type RiskDimensions struct {
	// TLS fingerprint risk (JA3/JA4)
	TLSFingerprint float64

	// Server-side risk (JA4S)
	ServerBehavior float64

	// HTTP/2 risk
	HTTP2Signature float64

	// HTTP request header risk (JA4H)
	HTTPHeaders float64

	// QUIC risk
	QUICSignature float64

	// Client Hints inconsistency
	ClientHints float64

	// ECH impact
	ECHImpact float64

	// behaviorexception
	BehaviorAnomaly float64
}

// RiskFactor represents a risk factor
type RiskFactor struct {
	// Factor type
	Type string

	// Severity (0.0-1.0)
	Severity float64

	// Description
	Description string

	// Evidence
	Evidence string

	// Weight
	Weight float64

	// Confidence
	Confidence float64
}

// RiskInput represents risk scoring input
type RiskInput struct {
	// JA3/JA4 fingerprint data
	JA3Hash string
	JA4Hash string

	// JA4S result
	JA4SResult *ja4s.JA4SResult

	// HTTP/2 signature result
	HTTP2Result *http2.HTTP2SignatureResult

	// JA4H result
	JA4HResult *ja4h.JA4HResult

	// QUIC signature result
	QUICResult *quic.QUICSignatureResult

	// ECH analyzeresult
	ECHResult *ech.ECHAnalysisResult

	// Additional context
	Context RiskContext
}

// RiskContext represents risk assessment context
type RiskContext struct {
	// IP reputation score (optional)
	IPReputation float64

	// Geolocation risk (optional)
	GeoRisk float64

	// Historical behavior score (optional)
	HistoricalScore float64

	// Request frequency (optional)
	RequestRate float64

	// Whether it's a known client
	IsKnownClient bool
}

// ScoringConfig represents scoring configuration
type ScoringConfig struct {
	// Dimension weights
	Weights DimensionWeights

	// Threat level thresholds
	Thresholds ThreatThresholds

	// Whether strict mode is enabled
	StrictMode bool

	// Minimum confidence requirement
	MinConfidence float64
}

// DimensionWeights represents dimension weights
type DimensionWeights struct {
	TLSFingerprint  float64 // Default 0.20
	ServerBehavior  float64 // Default 0.15
	HTTP2Signature  float64 // Default 0.15
	HTTPHeaders     float64 // Default 0.15
	QUICSignature   float64 // Default 0.10
	ClientHints     float64 // Default 0.10
	ECHImpact       float64 // Default 0.05
	BehaviorAnomaly float64 // Default 0.10
}

// ThreatThresholds represents threat level thresholds
type ThreatThresholds struct {
	Safe     float64 // <= 0.2
	Low      float64 // <= 0.4
	Medium   float64 // <= 0.6
	High     float64 // <= 0.8
	Critical float64 // > 0.8
}

// MaliciousFingerprintEntry represents malicious fingerprint entry
type MaliciousFingerprintEntry struct {
	// Fingerprint hash (JA3 or JA4)
	Hash string
	// Fingerprint type ("JA3" or "JA4")
	Type string
	// Threat type (e.g., "botnet", "scanner", "exploit_kit")
	ThreatType string
	// Severity (0.0-1.0)
	Severity float64
	// Description
	Description string
	// Last seen time (optional)
	LastSeen string
}

// RiskScorer is risk scorer
type RiskScorer struct {
	config                ScoringConfig
	maliciousFingerprints map[string]MaliciousFingerprintEntry
}

// NewRiskScorer creates a risk scorer
func NewRiskScorer(config *ScoringConfig) *RiskScorer {
	if config == nil {
		config = DefaultScoringConfig()
	}
	return &RiskScorer{
		config:                *config,
		maliciousFingerprints: initMaliciousFingerprintDatabase(),
	}
}

// DefaultScoringConfig returns default scoring configuration
func DefaultScoringConfig() *ScoringConfig {
	return &ScoringConfig{
		Weights: DimensionWeights{
			TLSFingerprint:  0.20,
			ServerBehavior:  0.15,
			HTTP2Signature:  0.15,
			HTTPHeaders:     0.15,
			QUICSignature:   0.10,
			ClientHints:     0.10,
			ECHImpact:       0.05,
			BehaviorAnomaly: 0.10,
		},
		Thresholds: ThreatThresholds{
			Safe:     0.2,
			Low:      0.4,
			Medium:   0.6,
			High:     0.8,
			Critical: 1.0,
		},
		StrictMode:    false,
		MinConfidence: 0.5,
	}
}

// CalculateRisk calculates comprehensive risk score
func (s *RiskScorer) CalculateRisk(input RiskInput) (*RiskScore, error) {
	result := &RiskScore{
		Dimensions:      RiskDimensions{},
		RiskFactors:     []RiskFactor{},
		Recommendations: []string{},
	}

	// 1. Calculate risk by dimensions
	s.calculateTLSRisk(input, result)
	s.calculateServerRisk(input, result)
	s.calculateHTTP2Risk(input, result)
	s.calculateHTTPHeadersRisk(input, result)
	s.calculateQUICRisk(input, result)
	s.calculateClientHintsRisk(input, result)
	s.calculateECHRisk(input, result)
	s.calculateBehaviorRisk(input, result)

	// 2. Calculate weighted total score
	result.TotalScore = s.calculateWeightedScore(result.Dimensions)

	// 3. Apply context adjustment
	result.TotalScore = s.applyContextAdjustment(result.TotalScore, input.Context)

	// 4. Ensure score is in reasonable range
	result.TotalScore = math.Max(0.0, math.Min(1.0, result.TotalScore))

	// 5. Determine threat level
	result.ThreatLevel = s.determineThreatLevel(result.TotalScore)

	// 6. Calculate confidence
	result.Confidence = s.calculateConfidence(input)

	// 7. statisticsexceptionquantity
	result.AnomalyCount = s.countAnomalies(input)

	// 8. Generate recommended actions
	result.Recommendations = s.generateRecommendations(result)

	return result, nil
}

// calculateTLSRisk calculates TLS fingerprint risk
func (s *RiskScorer) calculateTLSRisk(input RiskInput, result *RiskScore) {
	score := 0.0

	// Based on known JA3/JA4 hash assessment
	if input.JA3Hash == "" && input.JA4Hash == "" {
		score = 0.3 // Missing TLS fingerprint
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Type:        "TLS_FINGERPRINT_MISSING",
			Severity:    0.3,
			Description: "Missing TLS fingerprint data",
			Weight:      s.config.Weights.TLSFingerprint,
			Confidence:  0.9,
		})
	} else {
		// Check whether match known malicious fingerprint database
		if entry, found := s.checkMaliciousFingerprint(input.JA3Hash, "JA3"); found {
			score = entry.Severity
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Type:        "MALICIOUS_TLS_FINGERPRINT",
				Severity:    entry.Severity,
				Description: fmt.Sprintf("Malicious fingerprint detected: %s (%s)", entry.ThreatType, entry.Description),
				Weight:      s.config.Weights.TLSFingerprint,
				Confidence:  0.95,
			})
		} else if entry, found := s.checkMaliciousFingerprint(input.JA4Hash, "JA4"); found {
			score = entry.Severity
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Type:        "MALICIOUS_TLS_FINGERPRINT",
				Severity:    entry.Severity,
				Description: fmt.Sprintf("Malicious fingerprint detected: %s (%s)", entry.ThreatType, entry.Description),
				Weight:      s.config.Weights.TLSFingerprint,
				Confidence:  0.95,
			})
		}
	}

	result.Dimensions.TLSFingerprint = score
}

// calculateServerRisk calculates server-side behavior risk
func (s *RiskScorer) calculateServerRisk(input RiskInput, result *RiskScore) {
	if input.JA4SResult == nil {
		result.Dimensions.ServerBehavior = 0.0
		return
	}

	score := input.JA4SResult.RiskScore

	// Add risk factor
	for _, anomaly := range input.JA4SResult.AnomalyFlags {
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Type:        fmt.Sprintf("JA4S_%s", anomaly),
			Severity:    0.5,
			Description: fmt.Sprintf("Server-side anomaly: %s", anomaly),
			Evidence:    input.JA4SResult.Hash,
			Weight:      s.config.Weights.ServerBehavior,
			Confidence:  0.8,
		})
	}

	result.Dimensions.ServerBehavior = score
}

// calculateHTTP2Risk calculates HTTP/2 risk
func (s *RiskScorer) calculateHTTP2Risk(input RiskInput, result *RiskScore) {
	if input.HTTP2Result == nil {
		result.Dimensions.HTTP2Signature = 0.0
		return
	}

	score := input.HTTP2Result.RiskScore

	for _, anomaly := range input.HTTP2Result.AnomalyFlags {
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Type:        fmt.Sprintf("HTTP2_%s", anomaly),
			Severity:    0.4,
			Description: fmt.Sprintf("HTTP/2 exception: %s", anomaly),
			Weight:      s.config.Weights.HTTP2Signature,
			Confidence:  0.7,
		})
	}

	result.Dimensions.HTTP2Signature = score
}

// calculateHTTPHeadersRisk calculates HTTP request header risk
func (s *RiskScorer) calculateHTTPHeadersRisk(input RiskInput, result *RiskScore) {
	if input.JA4HResult == nil {
		result.Dimensions.HTTPHeaders = 0.0
		return
	}

	score := input.JA4HResult.RiskScore

	for _, anomaly := range input.JA4HResult.AnomalyFlags {
		severity := 0.5
		if anomaly == "CRITICAL_UA_CH_MISMATCH" {
			severity = 0.8
		}

		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Type:        fmt.Sprintf("JA4H_%s", anomaly),
			Severity:    severity,
			Description: fmt.Sprintf("HTTP header anomaly: %s", anomaly),
			Weight:      s.config.Weights.HTTPHeaders,
			Confidence:  0.85,
		})
	}

	result.Dimensions.HTTPHeaders = score
}

// calculateQUICRisk calculates QUIC risk
func (s *RiskScorer) calculateQUICRisk(input RiskInput, result *RiskScore) {
	if input.QUICResult == nil {
		result.Dimensions.QUICSignature = 0.0
		return
	}

	score := input.QUICResult.RiskScore

	for _, anomaly := range input.QUICResult.AnomalyFlags {
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Type:        fmt.Sprintf("QUIC_%s", anomaly),
			Severity:    0.4,
			Description: fmt.Sprintf("QUIC exception: %s", anomaly),
			Weight:      s.config.Weights.QUICSignature,
			Confidence:  0.75,
		})
	}

	result.Dimensions.QUICSignature = score
}

// calculateClientHintsRisk calculates Client Hints risk
func (s *RiskScorer) calculateClientHintsRisk(input RiskInput, result *RiskScore) {
	if input.JA4HResult == nil {
		result.Dimensions.ClientHints = 0.0
		return
	}

	// Client Hints inconsistency detection is completed in JA4H
	// Identify Client Hints related issues through anomaly markers
	score := 0.0

	// If there are Client Hints related anomaly markers
	for _, anomaly := range input.JA4HResult.AnomalyFlags {
		if anomaly == "CLIENT_HINTS_UA_MISMATCH" ||
			anomaly == "CLIENT_HINTS_MOBILE_MISMATCH" ||
			anomaly == "CLIENT_HINTS_PLATFORM_MISMATCH" ||
			anomaly == "UA_CH_MISMATCH" {
			score += 0.15
		}
	}

	result.Dimensions.ClientHints = math.Min(score, 1.0)
}

// calculateECHRisk calculates ECH impact
func (s *RiskScorer) calculateECHRisk(input RiskInput, result *RiskScore) {
	if input.ECHResult == nil {
		result.Dimensions.ECHImpact = 0.0
		return
	}

	score := input.ECHResult.RiskScore

	// ECH itself is not a risk, but affects detection capability
	if input.ECHResult.ECHPresent && input.ECHResult.Impact.ImpactLevel == "high" {
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Type:        "ECH_HIGH_IMPACT",
			Severity:    0.3,
			Description: "High ECH impact: SNI invisible, need to use alternative detection methods",
			Evidence:    input.ECHResult.VisibleFieldsSignature,
			Weight:      s.config.Weights.ECHImpact,
			Confidence:  0.9,
		})
	}

	// ECH configuration error is a real risk
	for _, anomaly := range input.ECHResult.AnomalyFlags {
		if anomaly != "GREASE_ECH" {
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Type:        fmt.Sprintf("ECH_%s", anomaly),
				Severity:    0.6,
				Description: fmt.Sprintf("ECH exception: %s", anomaly),
				Weight:      s.config.Weights.ECHImpact,
				Confidence:  0.85,
			})
		}
	}

	result.Dimensions.ECHImpact = score
}

// calculateBehaviorRisk calculates behavior anomaly risk
func (s *RiskScorer) calculateBehaviorRisk(input RiskInput, result *RiskScore) {
	score := 0.0

	// Behavior analysis based on context
	if input.Context.RequestRate > 100 {
		score += 0.3
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Type:        "HIGH_REQUEST_RATE",
			Severity:    0.3,
			Description: "Anomalously high request frequency",
			Evidence:    fmt.Sprintf("%.0f req/min", input.Context.RequestRate),
			Weight:      s.config.Weights.BehaviorAnomaly,
			Confidence:  0.7,
		})
	}

	result.Dimensions.BehaviorAnomaly = math.Min(score, 1.0)
}

// calculateWeightedScore calculates weighted total score
func (s *RiskScorer) calculateWeightedScore(dims RiskDimensions) float64 {
	w := s.config.Weights

	score := dims.TLSFingerprint*w.TLSFingerprint +
		dims.ServerBehavior*w.ServerBehavior +
		dims.HTTP2Signature*w.HTTP2Signature +
		dims.HTTPHeaders*w.HTTPHeaders +
		dims.QUICSignature*w.QUICSignature +
		dims.ClientHints*w.ClientHints +
		dims.ECHImpact*w.ECHImpact +
		dims.BehaviorAnomaly*w.BehaviorAnomaly

	return score
}

// applyContextAdjustment applies context adjustment
func (s *RiskScorer) applyContextAdjustment(score float64, ctx RiskContext) float64 {
	adjustment := 0.0

	// IP reputation adjustment
	if ctx.IPReputation > 0 {
		adjustment += (1.0 - ctx.IPReputation) * 0.1
	}

	// Geolocation risk
	if ctx.GeoRisk > 0 {
		adjustment += ctx.GeoRisk * 0.05
	}

	// Historical behavior
	if ctx.HistoricalScore > 0 {
		// Good history can reduce risk
		adjustment -= (1.0 - ctx.HistoricalScore) * 0.1
	}

	// Known client reduces risk
	if ctx.IsKnownClient {
		adjustment -= 0.1
	}

	return score + adjustment
}

// determineThreatLevel determines threat level
func (s *RiskScorer) determineThreatLevel(score float64) string {
	t := s.config.Thresholds

	if score <= t.Safe {
		return "safe"
	} else if score <= t.Low {
		return "low"
	} else if score <= t.Medium {
		return "medium"
	} else if score <= t.High {
		return "high"
	}
	return "critical"
}

// calculateConfidence calculates confidence
func (s *RiskScorer) calculateConfidence(input RiskInput) float64 {
	confidence := 0.0
	count := 0

	// Calculate confidence based on available data
	if input.JA3Hash != "" || input.JA4Hash != "" {
		confidence += 0.15
		count++
	}
	if input.JA4SResult != nil {
		confidence += 0.15
		count++
	}
	if input.HTTP2Result != nil {
		confidence += 0.15
		count++
	}
	if input.JA4HResult != nil {
		confidence += 0.15
		count++
	}
	if input.QUICResult != nil {
		confidence += 0.10
		count++
	}
	if input.ECHResult != nil {
		confidence += 0.10
		count++
	}

	// Context data increases confidence
	if input.Context.IPReputation > 0 {
		confidence += 0.10
	}
	if input.Context.HistoricalScore > 0 {
		confidence += 0.10
	}

	return math.Min(confidence, 1.0)
}

// countAnomalies statisticsexceptionquantity
func (s *RiskScorer) countAnomalies(input RiskInput) int {
	count := 0

	if input.JA4SResult != nil {
		count += len(input.JA4SResult.AnomalyFlags)
	}
	if input.HTTP2Result != nil {
		count += len(input.HTTP2Result.AnomalyFlags)
	}
	if input.JA4HResult != nil {
		count += len(input.JA4HResult.AnomalyFlags)
	}
	if input.QUICResult != nil {
		count += len(input.QUICResult.AnomalyFlags)
	}
	if input.ECHResult != nil {
		count += len(input.ECHResult.AnomalyFlags)
	}

	return count
}

// generateRecommendations generates recommended actions
func (s *RiskScorer) generateRecommendations(result *RiskScore) []string {
	recommendations := []string{}

	switch result.ThreatLevel {
	case "safe":
		recommendations = append(recommendations, "No special handling needed, continue normal monitoring")

	case "low":
		recommendations = append(recommendations, "Enable basic monitoring, log request features")
		if result.Dimensions.TLSFingerprint > 0.2 {
			recommendations = append(recommendations, "Monitor TLS fingerprint changes")
		}

	case "medium":
		recommendations = append(recommendations, "Increase monitoring frequency, analyze behavior patterns")
		recommendations = append(recommendations, "Consider enabling additional verification (e.g., CAPTCHA)")
		if result.AnomalyCount > 3 {
			recommendations = append(recommendations, "Multiple anomalous metrics, detailed inspection recommended")
		}

	case "high":
		recommendations = append(recommendations, "Limit access frequency, enable strict verification")
		recommendations = append(recommendations, "Log complete request context for analysis")
		recommendations = append(recommendations, "Consider temporary blocking, manual review")
		if result.Dimensions.HTTPHeaders > 0.6 {
			recommendations = append(recommendations, "HTTP headers highly suspicious, possible forgery")
		}

	case "critical":
		recommendations = append(recommendations, "Block this request immediately")
		recommendations = append(recommendations, "Log all related features for threat intelligence")
		recommendations = append(recommendations, "Trigger security alert, notify administrator")
		recommendations = append(recommendations, "Check for correlated attacks")
	}

	// Recommendations for high ECH impact
	if result.Dimensions.ECHImpact > 0.2 {
		recommendations = append(recommendations, "ECH detected, recommend using visible field fingerprinting and behavior analysis")
	}

	// Recommendations for confidence
	if result.Confidence < s.config.MinConfidence {
		recommendations = append(recommendations, "Insufficient confidence, recommend collecting more fingerprint data")
	}

	return recommendations
}

// GetSummary gets risk summary
func (r *RiskScore) GetSummary() string {
	return fmt.Sprintf("Threat level: %s, Risk score: %.2f, Confidence: %.2f, Anomalies: %d",
		r.ThreatLevel,
		r.TotalScore,
		r.Confidence,
		r.AnomalyCount,
	)
}

// CalculateRisk convenience function: calculate risk using default configuration
func CalculateRisk(input RiskInput) (*RiskScore, error) {
	scorer := NewRiskScorer(nil)
	return scorer.CalculateRisk(input)
}

// checkMaliciousFingerprint checks whether fingerprint is in malicious fingerprint database
func (s *RiskScorer) checkMaliciousFingerprint(hash, fingerprintType string) (MaliciousFingerprintEntry, bool) {
	if hash == "" {
		return MaliciousFingerprintEntry{}, false
	}

	key := fingerprintType + ":" + hash
	entry, found := s.maliciousFingerprints[key]
	return entry, found
}

// initMaliciousFingerprintDatabase initializes malicious fingerprint database
// This contains some known malicious fingerprint examples, in actual use should load from external data sources
func initMaliciousFingerprintDatabase() map[string]MaliciousFingerprintEntry {
	db := make(map[string]MaliciousFingerprintEntry)

	// Example: Known malicious fingerprint entries
	// Note: These are example data, actual deployment should use real threat intelligence data

	// Mirai botnet
	db["JA3:6734f37431670b3ab4292b8f60f29984"] = MaliciousFingerprintEntry{
		Hash:        "6734f37431670b3ab4292b8f60f29984",
		Type:        "JA3",
		ThreatType:  "botnet",
		Severity:    0.9,
		Description: "Mirai botnet variant",
		LastSeen:    "2024-01",
	}

	// Metasploit Framework
	db["JA3:a0e9f5d64349fb13191bc781f81f42e1"] = MaliciousFingerprintEntry{
		Hash:        "a0e9f5d64349fb13191bc781f81f42e1",
		Type:        "JA3",
		ThreatType:  "exploit_framework",
		Severity:    0.85,
		Description: "Metasploit default configuration",
		LastSeen:    "2024-02",
	}

	// Nmap Scanner
	db["JA3:c4d5c8a8e5d91e9e9f8d5b5c5d5e5f5a"] = MaliciousFingerprintEntry{
		Hash:        "c4d5c8a8e5d91e9e9f8d5b5c5d5e5f5a",
		Type:        "JA3",
		ThreatType:  "scanner",
		Severity:    0.7,
		Description: "Nmap network scanner",
		LastSeen:    "2024-03",
	}

	// SQLMap Injection Tool
	db["JA3:b32309a26951912be7dba376398abc3b"] = MaliciousFingerprintEntry{
		Hash:        "b32309a26951912be7dba376398abc3b",
		Type:        "JA3",
		ThreatType:  "injection_tool",
		Severity:    0.95,
		Description: "SQLMap SQL injection tool",
		LastSeen:    "2024-02",
	}

	// ZGrab Scanner
	db["JA3:f436b9416f37d134cadd04886327d3e8"] = MaliciousFingerprintEntry{
		Hash:        "f436b9416f37d134cadd04886327d3e8",
		Type:        "JA3",
		ThreatType:  "scanner",
		Severity:    0.65,
		Description: "ZGrab security scanner",
		LastSeen:    "2024-01",
	}

	// JA4 example: Malicious crawler
	db["JA4:t13d1516h2_8daaf6152771_b0da82dd1658"] = MaliciousFingerprintEntry{
		Hash:        "t13d1516h2_8daaf6152771_b0da82dd1658",
		Type:        "JA4",
		ThreatType:  "malicious_crawler",
		Severity:    0.75,
		Description: "Malicious data collection crawler",
		LastSeen:    "2024-03",
	}

	return db
}
