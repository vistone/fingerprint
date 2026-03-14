package risk

// Phase 3: This module has completed basic migration, pending deep optimization (see docs/5-process/modularization/PHASE_3_PLAN.md)
import (
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
