package ech

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// ECHAnalysisResult ECH (Encrypted Client Hello) analysis result
type ECHAnalysisResult struct {
	// Whether ECH is present
	ECHPresent bool

	// ECH config type
	ECHType string // "outer", "inner", "grease"

	// ECH version
	ECHVersion uint16

	// ClientHello type
	ClientHelloType string // "outer", "inner"

	// Full ECH feature hash
	Hash string

	// Visible fields signature (alternative fingerprint when ECH is present)
	VisibleFieldsSignature string

	// Anomaly score
	RiskScore float64

	// Anomaly flag list
	AnomalyFlags []string

	// Impact assessment
	Impact ECHImpact

	// Suggested alternative strategies
	AlternativeStrategies []string
}

// ECHImpact Impact of ECH on fingerprinting
type ECHImpact struct {
	// SNI visibility
	SNIVisible bool

	// Affected fingerprinting methods
	AffectedMethods []string

	// Still available fingerprinting methods
	AvailableMethods []string

	// Overall impact level
	ImpactLevel string // "none", "low", "medium", "high"
}

// ClientHelloData ClientHello data (used for ECH analysis)
type ClientHelloData struct {
	// TLS version
	TLSVersion uint16

	// Extension list
	Extensions []ExtensionData

	// Cipher Suites
	CipherSuites []uint16

	// Compression Methods
	CompressionMethods []uint8

	// Whether SNI is present
	HasSNI bool

	// SNI value (if visible)
	SNI string
}

// ExtensionData Extension data
type ExtensionData struct {
	Type uint16
	Data []byte
}

// ECHAnalyzer ECH analyzer
type ECHAnalyzer struct {
	// Known ECH configurations
	knownECHConfigs map[string]*ECHConfig
}

// ECHConfig Known ECH configuration
type ECHConfig struct {
	Name        string
	Version     uint16
	Description string
	RiskScore   float64
}

// NewECHAnalyzer creates an analyzer
func NewECHAnalyzer() *ECHAnalyzer {
	return &ECHAnalyzer{
		knownECHConfigs: initKnownECHConfigs(),
	}
}

// AnalyzeClientHello analyzes ECH in ClientHello
func (a *ECHAnalyzer) AnalyzeClientHello(data ClientHelloData) (*ECHAnalysisResult, error) {
	result := &ECHAnalysisResult{
		AnomalyFlags:          []string{},
		AlternativeStrategies: []string{},
	}

	// 1. Detect whether ECH extension is present
	echExt := a.findECHExtension(data.Extensions)
	result.ECHPresent = echExt != nil

	if !result.ECHPresent {
		// No ECH, use standard fingerprinting methods
		result.Impact.ImpactLevel = "none"
		result.Impact.SNIVisible = data.HasSNI
		result.Impact.AvailableMethods = []string{
			"JA3", "JA4", "JA4S", "JA4H", "HTTP/2", "QUIC",
		}

		// Generate standard fingerprint hash
		result.VisibleFieldsSignature = a.generateVisibleFieldsSignature(data)
		fullSignature := fmt.Sprintf("no_ech_%s", result.VisibleFieldsSignature)
		hash := sha256.Sum256([]byte(fullSignature))
		result.Hash = hex.EncodeToString(hash[:])

		return result, nil
	}

	// 2. Analyze ECH configuration
	a.analyzeECHExtension(echExt, result)

	// 3. Assess impact
	a.assessImpact(data, result)

	// 4. Generate visible fields signature
	result.VisibleFieldsSignature = a.generateVisibleFieldsSignature(data)

	// 5. Compute full hash
	fullSignature := fmt.Sprintf("ech_%s_v%d_%s",
		result.ECHType,
		result.ECHVersion,
		result.VisibleFieldsSignature,
	)
	hash := sha256.Sum256([]byte(fullSignature))
	result.Hash = hex.EncodeToString(hash[:])

	// 6. Anomaly detection
	a.detectAnomalies(data, result)

	// 7. Suggest alternative strategies
	a.suggestAlternativeStrategies(result)

	return result, nil
}

// findECHExtension finds the ECH extension
func (a *ECHAnalyzer) findECHExtension(extensions []ExtensionData) *ExtensionData {
	const (
		ExtensionEncryptedClientHello = 0xfe0d // ECH extension type
		ExtensionECHOuterExtensions   = 0xfd00 // ECH outer extensions
	)

	for i := range extensions {
		if extensions[i].Type == ExtensionEncryptedClientHello ||
			extensions[i].Type == ExtensionECHOuterExtensions {
			return &extensions[i]
		}
	}
	return nil
}

// analyzeECHExtension analyzes the ECH extension
func (a *ECHAnalyzer) analyzeECHExtension(ext *ExtensionData, result *ECHAnalysisResult) {
	if ext == nil {
		return
	}

	// Parse ECH type and version
	if len(ext.Data) < 2 {
		result.ECHType = "unknown"
		result.ECHVersion = 0
		return
	}

	// ECH version is in the first two bytes of data
	result.ECHVersion = uint16(ext.Data[0])<<8 | uint16(ext.Data[1])

	// GREASE detection: version number is 0
	if result.ECHVersion == 0x0000 {
		result.ECHType = "grease"
		return
	}

	// Determine ClientHello type (first byte after version)
	if len(ext.Data) > 2 {
		clientHelloType := ext.Data[2]
		switch clientHelloType {
		case 0x00:
			result.ECHType = "outer"
			result.ClientHelloType = "outer"
		case 0x01:
			result.ECHType = "inner"
			result.ClientHelloType = "inner"
		default:
			result.ECHType = "unknown"
		}
	} else {
		result.ECHType = "unknown"
	}
}

// assessImpact assesses ECH impact
func (a *ECHAnalyzer) assessImpact(data ClientHelloData, result *ECHAnalysisResult) {
	result.Impact.SNIVisible = false // ECH encrypted the SNI

	// Affected methods
	result.Impact.AffectedMethods = []string{
		"SNI-based routing",
		"SNI filtering",
		"Domain-specific policies",
	}

	// Still available methods
	result.Impact.AvailableMethods = []string{
		"JA3/JA4 fingerprinting",
		"Cipher suite analysis",
		"Extension order analysis",
		"HTTP/2 frame analysis",
		"QUIC signature analysis",
		"Application layer patterns",
		"Behavioral analysis",
	}

	// Impact level assessment
	if result.ECHType == "grease" {
		result.Impact.ImpactLevel = "low" // GREASE does not truly encrypt
	} else if result.ClientHelloType == "outer" {
		result.Impact.ImpactLevel = "high" // Outer ClientHello, full ECH
	} else {
		result.Impact.ImpactLevel = "medium"
	}
}

// generateVisibleFieldsSignature generates a visible fields signature
func (a *ECHAnalyzer) generateVisibleFieldsSignature(data ClientHelloData) string {
	// Generate signature based on still-visible fields
	var parts []string

	// TLS version
	parts = append(parts, fmt.Sprintf("tls_%04x", data.TLSVersion))

	// Cipher Suites (first 5)
	cipherPart := "cs_"
	for i, cs := range data.CipherSuites {
		if i >= 5 {
			break
		}
		cipherPart += fmt.Sprintf("%04x", cs)
	}
	parts = append(parts, cipherPart)

	// Extension types (excluding ECH)
	extPart := "ext_"
	for _, ext := range data.Extensions {
		if ext.Type != 0xfe0d && ext.Type != 0xfd00 {
			extPart += fmt.Sprintf("%04x", ext.Type)
		}
	}
	parts = append(parts, extPart)

	signature := ""
	for _, part := range parts {
		signature += part + "_"
	}

	hash := sha256.Sum256([]byte(signature))
	return hex.EncodeToString(hash[:8])
}

// detectAnomalies detects anomalies
func (a *ECHAnalyzer) detectAnomalies(data ClientHelloData, result *ECHAnalysisResult) {
	baseScore := 0.0

	// Anomaly 1: GREASE ECH (possibly testing or probing)
	if result.ECHType == "grease" {
		result.AnomalyFlags = append(result.AnomalyFlags, "GREASE_ECH")
		baseScore += 0.1
	}

	// Anomaly 2: Unknown ECH version
	if result.ECHVersion > 0xfe0d {
		result.AnomalyFlags = append(result.AnomalyFlags, "UNKNOWN_ECH_VERSION")
		baseScore += 0.15
	}

	// Anomaly 3: ECH but SNI still visible (misconfiguration)
	if result.ECHPresent && data.HasSNI && result.ECHType != "grease" {
		result.AnomalyFlags = append(result.AnomalyFlags, "ECH_WITH_VISIBLE_SNI")
		baseScore += 0.3
	}

	// Anomaly 4: Outer ClientHello missing required extensions
	if result.ClientHelloType == "outer" && len(data.Extensions) < 5 {
		result.AnomalyFlags = append(result.AnomalyFlags, "INCOMPLETE_OUTER_HELLO")
		baseScore += 0.2
	}

	// Anomaly 5: TLS version too old but using ECH (ECH requires TLS 1.3)
	if data.TLSVersion < 0x0304 && result.ECHPresent {
		result.AnomalyFlags = append(result.AnomalyFlags, "ECH_WITH_OLD_TLS")
		baseScore += 0.35
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}
	result.RiskScore = baseScore
}

// suggestAlternativeStrategies suggests alternative strategies
func (a *ECHAnalyzer) suggestAlternativeStrategies(result *ECHAnalysisResult) {
	if !result.ECHPresent {
		return
	}

	// Suggest strategies based on ECH type
	switch result.ECHType {
	case "grease":
		result.AlternativeStrategies = append(result.AlternativeStrategies,
			"Standard fingerprinting methods still fully available (GREASE ECH does not encrypt)",
		)
	case "outer":
		result.AlternativeStrategies = append(result.AlternativeStrategies,
			"Use JA3/JA4 fingerprinting based on visible fields",
			"Analyze cipher suite and extension order",
			"Combine HTTP/2 frame signatures and QUIC characteristics",
			"Application layer behavior analysis (request patterns, timing)",
			"IP reputation and geolocation analysis",
		)
	case "inner":
		result.AlternativeStrategies = append(result.AlternativeStrategies,
			"Use visible field signatures",
			"Cross-request behavior correlation",
			"Transport layer characteristic analysis",
		)
	}

	// General recommendations
	if result.Impact.ImpactLevel == "high" || result.Impact.ImpactLevel == "medium" {
		result.AlternativeStrategies = append(result.AlternativeStrategies,
			"Implement multi-layered defense strategy, do not rely on a single fingerprinting method",
		)
	}
}

// GetImpactSummary gets the impact summary
func (r *ECHAnalysisResult) GetImpactSummary() string {
	if !r.ECHPresent {
		return "No ECH, standard fingerprinting methods fully available"
	}

	return fmt.Sprintf("ECH type: %s, impact level: %s, SNI visible: %v",
		r.ECHType,
		r.Impact.ImpactLevel,
		r.Impact.SNIVisible,
	)
}

// ============ Known configuration database ============

func initKnownECHConfigs() map[string]*ECHConfig {
	return map[string]*ECHConfig{
		"draft_13": {
			Name:        "ECH Draft 13",
			Version:     0xfe0d,
			Description: "TLS Encrypted Client Hello draft-13",
			RiskScore:   0.0,
		},
		"grease": {
			Name:        "ECH GREASE",
			Version:     0x0000,
			Description: "GREASE ECH for compatibility testing",
			RiskScore:   0.1,
		},
	}
}

// AnalyzeECH convenience function: analyzes ECH
func AnalyzeECH(data ClientHelloData) (*ECHAnalysisResult, error) {
	analyzer := NewECHAnalyzer()
	return analyzer.AnalyzeClientHello(data)
}
