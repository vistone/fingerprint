package ech

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// translated comment
type ECHAnalysisResult struct {
	// translated comment
	ECHPresent bool

	// translated comment
	ECHType string // "outer", "inner", "grease"

	// translated comment
	ECHVersion uint16

	// translated comment
	ClientHelloType string // "outer", "inner"

	// translated comment
	Hash string

	// translated comment
	VisibleFieldsSignature string

	// translated comment
	RiskScore float64

	// translated comment
	AnomalyFlags []string

	// translated comment
	Impact ECHImpact

	// translated comment
	AlternativeStrategies []string
}

// translated comment
type ECHImpact struct {
	// translated comment
	SNIVisible bool

	// translated comment
	AffectedMethods []string

	// translated comment
	AvailableMethods []string

	// translated comment
	ImpactLevel string // "none", "low", "medium", "high"
}

// translated comment
type ClientHelloData struct {
	// translated comment
	TLSVersion uint16

	// translated comment
	Extensions []ExtensionData

	// Cipher Suites
	CipherSuites []uint16

	// Compression Methods
	CompressionMethods []uint8

	// translated comment
	HasSNI bool

	// translated comment
	SNI string
}

// translated comment
type ExtensionData struct {
	Type uint16
	Data []byte
}

// translated comment
type ECHAnalyzer struct {
	// translated comment
	knownECHConfigs map[string]*ECHConfig
}

// translated comment
type ECHConfig struct {
	Name        string
	Version     uint16
	Description string
	RiskScore   float64
}

// translated comment
func NewECHAnalyzer() *ECHAnalyzer {
	return &ECHAnalyzer{
		knownECHConfigs: initKnownECHConfigs(),
	}
}

// translated comment
func (a *ECHAnalyzer) AnalyzeClientHello(data ClientHelloData) (*ECHAnalysisResult, error) {
	result := &ECHAnalysisResult{
		AnomalyFlags:          []string{},
		AlternativeStrategies: []string{},
	}

	// translated comment
	echExt := a.findECHExtension(data.Extensions)
	result.ECHPresent = echExt != nil

	if !result.ECHPresent {
		// translated comment
		result.Impact.ImpactLevel = "none"
		result.Impact.SNIVisible = data.HasSNI
		result.Impact.AvailableMethods = []string{
			"JA3", "JA4", "JA4S", "JA4H", "HTTP/2", "QUIC",
		}

		// translated comment
		result.VisibleFieldsSignature = a.generateVisibleFieldsSignature(data)
		fullSignature := fmt.Sprintf("no_ech_%s", result.VisibleFieldsSignature)
		hash := sha256.Sum256([]byte(fullSignature))
		result.Hash = hex.EncodeToString(hash[:])

		return result, nil
	}

	// translated comment
	a.analyzeECHExtension(echExt, result)

	// translated comment
	a.assessImpact(data, result)

	// translated comment
	result.VisibleFieldsSignature = a.generateVisibleFieldsSignature(data)

	// translated comment
	fullSignature := fmt.Sprintf("ech_%s_v%d_%s",
		result.ECHType,
		result.ECHVersion,
		result.VisibleFieldsSignature,
	)
	hash := sha256.Sum256([]byte(fullSignature))
	result.Hash = hex.EncodeToString(hash[:])

	// translated comment
	a.detectAnomalies(data, result)

	// translated comment
	a.suggestAlternativeStrategies(result)

	return result, nil
}

// translated comment
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

// translated comment
func (a *ECHAnalyzer) analyzeECHExtension(ext *ExtensionData, result *ECHAnalysisResult) {
	if ext == nil {
		return
	}

	// translated comment
	if len(ext.Data) < 2 {
		result.ECHType = "unknown"
		result.ECHVersion = 0
		return
	}

	// translated comment
	result.ECHVersion = uint16(ext.Data[0])<<8 | uint16(ext.Data[1])

	// translated comment
	if result.ECHVersion == 0x0000 {
		result.ECHType = "grease"
		return
	}

	// translated comment
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

// translated comment
func (a *ECHAnalyzer) assessImpact(data ClientHelloData, result *ECHAnalysisResult) {
	result.Impact.SNIVisible = false // translated comment

	// translated comment
	result.Impact.AffectedMethods = []string{
		"SNI-based routing",
		"SNI filtering",
		"Domain-specific policies",
	}

	// translated comment
	result.Impact.AvailableMethods = []string{
		"JA3/JA4 fingerprinting",
		"Cipher suite analysis",
		"Extension order analysis",
		"HTTP/2 frame analysis",
		"QUIC signature analysis",
		"Application layer patterns",
		"Behavioral analysis",
	}

	// translated comment
	if result.ECHType == "grease" {
		result.Impact.ImpactLevel = "low" // translated comment
	} else if result.ClientHelloType == "outer" {
		result.Impact.ImpactLevel = "high" // translated comment
	} else {
		result.Impact.ImpactLevel = "medium"
	}
}

// translated comment
func (a *ECHAnalyzer) generateVisibleFieldsSignature(data ClientHelloData) string {
	// translated comment
	var parts []string

	// translated comment
	parts = append(parts, fmt.Sprintf("tls_%04x", data.TLSVersion))

	// translated comment
	cipherPart := "cs_"
	for i, cs := range data.CipherSuites {
		if i >= 5 {
			break
		}
		cipherPart += fmt.Sprintf("%04x", cs)
	}
	parts = append(parts, cipherPart)

	// translated comment
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

// translated comment
func (a *ECHAnalyzer) detectAnomalies(data ClientHelloData, result *ECHAnalysisResult) {
	baseScore := 0.0

	// translated comment
	if result.ECHType == "grease" {
		result.AnomalyFlags = append(result.AnomalyFlags, "GREASE_ECH")
		baseScore += 0.1
	}

	// translated comment
	if result.ECHVersion > 0xfe0d {
		result.AnomalyFlags = append(result.AnomalyFlags, "UNKNOWN_ECH_VERSION")
		baseScore += 0.15
	}

	// translated comment
	if result.ECHPresent && data.HasSNI && result.ECHType != "grease" {
		result.AnomalyFlags = append(result.AnomalyFlags, "ECH_WITH_VISIBLE_SNI")
		baseScore += 0.3
	}

	// translated comment
	if result.ClientHelloType == "outer" && len(data.Extensions) < 5 {
		result.AnomalyFlags = append(result.AnomalyFlags, "INCOMPLETE_OUTER_HELLO")
		baseScore += 0.2
	}

	// translated comment
	if data.TLSVersion < 0x0304 && result.ECHPresent {
		result.AnomalyFlags = append(result.AnomalyFlags, "ECH_WITH_OLD_TLS")
		baseScore += 0.35
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}
	result.RiskScore = baseScore
}

// translated comment
func (a *ECHAnalyzer) suggestAlternativeStrategies(result *ECHAnalysisResult) {
	if !result.ECHPresent {
		return
	}

	// translated comment
	switch result.ECHType {
	case "grease":
		result.AlternativeStrategies = append(result.AlternativeStrategies,
			"标准指纹方法仍完全可用（GREASE ECH 不加密）",
		)
	case "outer":
		result.AlternativeStrategies = append(result.AlternativeStrategies,
			"使用 JA3/JA4 基于可见字段的指纹",
			"分析 Cipher Suite 和扩展顺序",
			"结合 HTTP/2 帧签名和 QUIC 特征",
			"应用层行为分析（请求模式、时序）",
			"IP 信誉和地理位置分析",
		)
	case "inner":
		result.AlternativeStrategies = append(result.AlternativeStrategies,
			"使用可见字段签名",
			"跨请求行为关联",
			"传输层特征分析",
		)
	}

	// translated comment
	if result.Impact.ImpactLevel == "high" || result.Impact.ImpactLevel == "medium" {
		result.AlternativeStrategies = append(result.AlternativeStrategies,
			"实施多层防御策略，不依赖单一指纹方法",
		)
	}
}

// translated comment
func (r *ECHAnalysisResult) GetImpactSummary() string {
	if !r.ECHPresent {
		return "无 ECH，标准指纹方法完全可用"
	}

	return fmt.Sprintf("ECH 类型: %s, 影响等级: %s, SNI 可见: %v",
		r.ECHType,
		r.Impact.ImpactLevel,
		r.Impact.SNIVisible,
	)
}

// translated comment

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

// translated comment
func AnalyzeECH(data ClientHelloData) (*ECHAnalysisResult, error) {
	analyzer := NewECHAnalyzer()
	return analyzer.AnalyzeClientHello(data)
}
