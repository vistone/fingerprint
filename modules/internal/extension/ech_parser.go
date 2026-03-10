package extension

import (
	"context"
	"fmt"
)

// translated comment
type ECHExtensionData struct {
	Type            ExtensionType
	Version         uint16
	ConfigID        uint8
	RawData         []byte
	EncryptedData   []byte
	ClientHelloData map[string]interface{}
}

func (e *ECHExtensionData) GetType() ExtensionType {
	return e.Type
}

func (e *ECHExtensionData) GetRawData() []byte {
	return e.RawData
}

func (e *ECHExtensionData) GetName() string {
	return "Encrypted Client Hello"
}

func (e *ECHExtensionData) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"type":             fmt.Sprintf("0x%04x", e.Type),
		"version":          e.Version,
		"config_id":        e.ConfigID,
		"data_length":      len(e.RawData),
		"encrypted_length": len(e.EncryptedData),
	}
}

// translated comment
type ECHAnalysisResultImpl struct {
	ExtType       ExtensionType
	Present       bool
	ECHType       string // outer, inner, grease
	Version       uint16
	RiskScore     float64
	Anomalies     []string
	VisibleFields map[string]interface{}
	Metadata      map[string]interface{}
}

func (r *ECHAnalysisResultImpl) GetExtensionType() ExtensionType {
	return r.ExtType
}

func (r *ECHAnalysisResultImpl) HasAnomalies() bool {
	return len(r.Anomalies) > 0
}

func (r *ECHAnalysisResultImpl) GetAnomalies() []string {
	result := make([]string, len(r.Anomalies))
	copy(result, r.Anomalies)
	return result
}

func (r *ECHAnalysisResultImpl) GetRiskScore() float64 {
	return r.RiskScore
}

func (r *ECHAnalysisResultImpl) ToMap() map[string]interface{} {
	visibleFields := make(map[string]interface{})
	for k, v := range r.VisibleFields {
		visibleFields[k] = v
	}

	return map[string]interface{}{
		"extension_type": fmt.Sprintf("0x%04x", r.ExtType),
		"present":        r.Present,
		"ech_type":       r.ECHType,
		"version":        r.Version,
		"risk_score":     r.RiskScore,
		"anomalies":      r.Anomalies,
		"visible_fields": visibleFields,
		"metadata":       r.Metadata,
	}
}

// translated comment
type ECHParser struct {
	version string
}

func NewECHParser() *ECHParser {
	return &ECHParser{
		version: "1.0.0",
	}
}

func (p *ECHParser) Parse(data []byte, parentContext context.Context) (ExtensionData, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("insufficient data for ECH extension")
	}

	echData := &ECHExtensionData{
		Type:    ExtensionEncryptedClientHello,
		RawData: make([]byte, len(data)),
	}
	copy(echData.RawData, data)

	// translated comment
	echData.Version = uint16(data[0])<<8 | uint16(data[1])

	// translated comment
	if len(data) > 2 {
		echData.ConfigID = data[2]
	}

	return echData, nil
}

func (p *ECHParser) GetType() ExtensionType {
	return ExtensionEncryptedClientHello
}

func (p *ECHParser) GetVersion() string {
	return p.version
}

// translated comment
type ECHAnalyzerImpl struct {
	version string
}

func NewECHAnalyzerImpl() *ECHAnalyzerImpl {
	return &ECHAnalyzerImpl{
		version: "1.0.0",
	}
}

func (a *ECHAnalyzerImpl) Analyze(data ExtensionData, config map[string]interface{}) (AnalysisResult, error) {
	echData, ok := data.(*ECHExtensionData)
	if !ok {
		return nil, fmt.Errorf("invalid data type for ECH analyzer")
	}

	result := &ECHAnalysisResultImpl{
		ExtType:       ExtensionEncryptedClientHello,
		Present:       true,
		Anomalies:     []string{},
		VisibleFields: make(map[string]interface{}),
		Metadata:      make(map[string]interface{}),
	}

	// translated comment
	result.Version = echData.Version
	a.detectECHType(echData, result)

	// translated comment
	result.RiskScore = a.calculateRiskScore(echData, result)

	// translated comment
	a.detectAnomalies(echData, result)

	return result, nil
}

func (a *ECHAnalyzerImpl) GetType() ExtensionType {
	return ExtensionEncryptedClientHello
}

func (a *ECHAnalyzerImpl) GetVersion() string {
	return a.version
}

func (a *ECHAnalyzerImpl) SupportsConfig() []string {
	return []string{
		"strict_validation",
		"check_sni_visible",
		"analyze_visible_fields",
	}
}

func (a *ECHAnalyzerImpl) detectECHType(data *ECHExtensionData, result *ECHAnalysisResultImpl) {
	if len(data.RawData) < 2 {
		result.ECHType = "unknown"
		return
	}

	// translated comment
	var configID uint8
	if len(data.RawData) > 2 {
		configID = data.RawData[2]
	}

	if configID == 0xff {
		result.ECHType = "grease"
	} else if len(data.RawData) > 3 && data.RawData[3] == 0x01 {
		result.ECHType = "outer"
	} else {
		result.ECHType = "inner"
	}
}

func (a *ECHAnalyzerImpl) calculateRiskScore(data *ECHExtensionData, result *ECHAnalysisResultImpl) float64 {
	var score float64

	// translated comment
	if result.ECHType == "grease" {
		score = 0.2
	} else if result.ECHType == "outer" {
		// translated comment
		score = 0.7
	} else {
		// translated comment
		score = 0.5
	}

	return score
}

func (a *ECHAnalyzerImpl) detectAnomalies(data *ECHExtensionData, result *ECHAnalysisResultImpl) {
	// translated comment
	if result.Version == 0x0000 {
		result.Anomalies = append(result.Anomalies, "invalid_version")
	}

	// translated comment
	if len(data.RawData) < 4 {
		result.Anomalies = append(result.Anomalies, "insufficient_data")
	}

	// translated comment
	if result.ECHType == "outer" && data.ClientHelloData != nil {
		if sni, exists := data.ClientHelloData["sni"]; exists && sni != nil {
			result.Anomalies = append(result.Anomalies, "ech_outer_with_visible_sni")
		}
	}
}

// translated comment
func InitializeECHExtension() error {
	// translated comment

	// translated comment
	err := RegisterParserBuilder(ExtensionEncryptedClientHello, func() (Parser, error) {
		return NewECHParser(), nil
	})
	if err != nil {
		return err
	}

	// translated comment
	err = RegisterAnalyzerBuilder(ExtensionEncryptedClientHello, func() (Analyzer, error) {
		return NewECHAnalyzerImpl(), nil
	})
	if err != nil {
		return err
	}

	return nil
}
