package extension

import (
	"context"
	"fmt"
)

// ECHExtensionData ECH 扩展数据结构
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

// ECHAnalysisResultImpl ECH 分析结果实现
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

// ECHParser 新式 ECH 解析器实现
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

	// 解析版本（前两字节）
	echData.Version = uint16(data[0])<<8 | uint16(data[1])

	// 判断 ECH 类型
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

// ECHAnalyzerImpl 新式 ECH 分析器实现
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

	// 分析版本
	result.Version = echData.Version
	a.detectECHType(echData, result)

	// 计算风险评分
	result.RiskScore = a.calculateRiskScore(echData, result)

	// 检测异常
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

	// 根据配置 ID 判断类型
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

	// GREASE 类型：低影响
	if result.ECHType == "grease" {
		score = 0.2
	} else if result.ECHType == "outer" {
		// Outer ClientHello：高影响
		score = 0.7
	} else {
		// Inner ClientHello：中等影响
		score = 0.5
	}

	return score
}

func (a *ECHAnalyzerImpl) detectAnomalies(data *ECHExtensionData, result *ECHAnalysisResultImpl) {
	// 检查版本
	if result.Version == 0x0000 {
		result.Anomalies = append(result.Anomalies, "invalid_version")
	}

	// 检查数据长度
	if len(data.RawData) < 4 {
		result.Anomalies = append(result.Anomalies, "insufficient_data")
	}

	// 检查与 SNI 的一致性
	if result.ECHType == "outer" && data.ClientHelloData != nil {
		if sni, exists := data.ClientHelloData["sni"]; exists && sni != nil {
			result.Anomalies = append(result.Anomalies, "ech_outer_with_visible_sni")
		}
	}
}

// InitializeECHExtension 初始化 ECH 扩展支持
func InitializeECHExtension() error {
	// 注册元数据已在 initStandardExtensions 中完成

	// 注册解析器
	err := RegisterParserBuilder(ExtensionEncryptedClientHello, func() (Parser, error) {
		return NewECHParser(), nil
	})
	if err != nil {
		return err
	}

	// 注册分析器
	err = RegisterAnalyzerBuilder(ExtensionEncryptedClientHello, func() (Analyzer, error) {
		return NewECHAnalyzerImpl(), nil
	})
	if err != nil {
		return err
	}

	return nil
}
