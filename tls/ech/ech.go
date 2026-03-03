package ech

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// ECHAnalysisResult ECH (Encrypted Client Hello) 分析结果
type ECHAnalysisResult struct {
	// ECH 是否存在
	ECHPresent bool

	// ECH 配置类型
	ECHType string // "outer", "inner", "grease"

	// ECH 版本
	ECHVersion uint16

	// ClientHello 类型
	ClientHelloType string // "outer", "inner"

	// 完整 ECH 特征哈希
	Hash string

	// 可见字段签名（ECH 存在时的替代指纹）
	VisibleFieldsSignature string

	// 异常判定分数
	RiskScore float64

	// 异常标记列表
	AnomalyFlags []string

	// 影响评估
	Impact ECHImpact

	// 建议的替代策略
	AlternativeStrategies []string
}

// ECHImpact ECH 对指纹识别的影响
type ECHImpact struct {
	// SNI 可见性
	SNIVisible bool

	// 受影响的指纹方法
	AffectedMethods []string

	// 仍可用的指纹方法
	AvailableMethods []string

	// 整体影响等级
	ImpactLevel string // "none", "low", "medium", "high"
}

// ClientHelloData ClientHello 数据（用于 ECH 分析）
type ClientHelloData struct {
	// TLS 版本
	TLSVersion uint16

	// 扩展列表
	Extensions []ExtensionData

	// Cipher Suites
	CipherSuites []uint16

	// Compression Methods
	CompressionMethods []uint8

	// 是否包含 SNI
	HasSNI bool

	// SNI 值（如果可见）
	SNI string
}

// ExtensionData 扩展数据
type ExtensionData struct {
	Type uint16
	Data []byte
}

// ECHAnalyzer ECH 分析器
type ECHAnalyzer struct {
	// 已知的 ECH 配置
	knownECHConfigs map[string]*ECHConfig
}

// ECHConfig 已知的 ECH 配置
type ECHConfig struct {
	Name        string
	Version     uint16
	Description string
	RiskScore   float64
}

// NewECHAnalyzer 创建分析器
func NewECHAnalyzer() *ECHAnalyzer {
	return &ECHAnalyzer{
		knownECHConfigs: initKnownECHConfigs(),
	}
}

// AnalyzeClientHello 分析 ClientHello 中的 ECH
func (a *ECHAnalyzer) AnalyzeClientHello(data ClientHelloData) (*ECHAnalysisResult, error) {
	result := &ECHAnalysisResult{
		AnomalyFlags:          []string{},
		AlternativeStrategies: []string{},
	}

	// 1. 检测是否存在 ECH 扩展
	echExt := a.findECHExtension(data.Extensions)
	result.ECHPresent = echExt != nil

	if !result.ECHPresent {
		// 没有 ECH，使用标准指纹方法
		result.Impact.ImpactLevel = "none"
		result.Impact.SNIVisible = data.HasSNI
		result.Impact.AvailableMethods = []string{
			"JA3", "JA4", "JA4S", "JA4H", "HTTP/2", "QUIC",
		}

		// 生成标准指纹哈希
		result.VisibleFieldsSignature = a.generateVisibleFieldsSignature(data)
		fullSignature := fmt.Sprintf("no_ech_%s", result.VisibleFieldsSignature)
		hash := sha256.Sum256([]byte(fullSignature))
		result.Hash = hex.EncodeToString(hash[:])

		return result, nil
	}

	// 2. 分析 ECH 配置
	a.analyzeECHExtension(echExt, result)

	// 3. 评估影响
	a.assessImpact(data, result)

	// 4. 生成可见字段签名
	result.VisibleFieldsSignature = a.generateVisibleFieldsSignature(data)

	// 5. 计算完整哈希
	fullSignature := fmt.Sprintf("ech_%s_v%d_%s",
		result.ECHType,
		result.ECHVersion,
		result.VisibleFieldsSignature,
	)
	hash := sha256.Sum256([]byte(fullSignature))
	result.Hash = hex.EncodeToString(hash[:])

	// 6. 异常检测
	a.detectAnomalies(data, result)

	// 7. 建议替代策略
	a.suggestAlternativeStrategies(result)

	return result, nil
}

// findECHExtension 查找 ECH 扩展
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

// analyzeECHExtension 分析 ECH 扩展
func (a *ECHAnalyzer) analyzeECHExtension(ext *ExtensionData, result *ECHAnalysisResult) {
	if ext == nil {
		return
	}

	// 解析 ECH 类型和版本
	if len(ext.Data) < 2 {
		result.ECHType = "unknown"
		result.ECHVersion = 0
		return
	}

	// ECH 版本在数据的前两个字节
	result.ECHVersion = uint16(ext.Data[0])<<8 | uint16(ext.Data[1])

	// GREASE 检测：版本号为 0
	if result.ECHVersion == 0x0000 {
		result.ECHType = "grease"
		return
	}

	// 判断 ClientHello 类型（版本后的第一个字节）
	if len(ext.Data) > 2 {
		clientHelloType := ext.Data[2]
		switch clientHelloType {
		case 0x00:
			result.ECHType = "inner"
			result.ClientHelloType = "inner"
		case 0x01:
			result.ECHType = "outer"
			result.ClientHelloType = "outer"
		default:
			result.ECHType = "unknown"
		}
	} else {
		result.ECHType = "unknown"
	}
}

// assessImpact 评估 ECH 影响
func (a *ECHAnalyzer) assessImpact(data ClientHelloData, result *ECHAnalysisResult) {
	result.Impact.SNIVisible = false // ECH 加密了 SNI

	// 受影响的方法
	result.Impact.AffectedMethods = []string{
		"SNI-based routing",
		"SNI filtering",
		"Domain-specific policies",
	}

	// 仍可用的方法
	result.Impact.AvailableMethods = []string{
		"JA3/JA4 fingerprinting",
		"Cipher suite analysis",
		"Extension order analysis",
		"HTTP/2 frame analysis",
		"QUIC signature analysis",
		"Application layer patterns",
		"Behavioral analysis",
	}

	// 影响等级评估
	if result.ECHType == "grease" {
		result.Impact.ImpactLevel = "low" // GREASE 不真正加密
	} else if result.ClientHelloType == "outer" {
		result.Impact.ImpactLevel = "high" // Outer ClientHello，完整 ECH
	} else {
		result.Impact.ImpactLevel = "medium"
	}
}

// generateVisibleFieldsSignature 生成可见字段签名
func (a *ECHAnalyzer) generateVisibleFieldsSignature(data ClientHelloData) string {
	// 基于仍然可见的字段生成签名
	var parts []string

	// TLS 版本
	parts = append(parts, fmt.Sprintf("tls_%04x", data.TLSVersion))

	// Cipher Suites（前5个）
	cipherPart := "cs_"
	for i, cs := range data.CipherSuites {
		if i >= 5 {
			break
		}
		cipherPart += fmt.Sprintf("%04x", cs)
	}
	parts = append(parts, cipherPart)

	// 扩展类型（排除 ECH）
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

// detectAnomalies 检测异常
func (a *ECHAnalyzer) detectAnomalies(data ClientHelloData, result *ECHAnalysisResult) {
	baseScore := 0.0

	// 异常 1: GREASE ECH（可能是测试或探测）
	if result.ECHType == "grease" {
		result.AnomalyFlags = append(result.AnomalyFlags, "GREASE_ECH")
		baseScore += 0.1
	}

	// 异常 2: 未知 ECH 版本
	if result.ECHVersion > 0xfe0d {
		result.AnomalyFlags = append(result.AnomalyFlags, "UNKNOWN_ECH_VERSION")
		baseScore += 0.15
	}

	// 异常 3: ECH 但仍有可见 SNI（配置错误）
	if result.ECHPresent && data.HasSNI && result.ECHType != "grease" {
		result.AnomalyFlags = append(result.AnomalyFlags, "ECH_WITH_VISIBLE_SNI")
		baseScore += 0.3
	}

	// 异常 4: Outer ClientHello 缺少必要扩展
	if result.ClientHelloType == "outer" && len(data.Extensions) < 5 {
		result.AnomalyFlags = append(result.AnomalyFlags, "INCOMPLETE_OUTER_HELLO")
		baseScore += 0.2
	}

	// 异常 5: TLS 版本过低但使用 ECH（ECH 需要 TLS 1.3）
	if data.TLSVersion < 0x0304 && result.ECHPresent {
		result.AnomalyFlags = append(result.AnomalyFlags, "ECH_WITH_OLD_TLS")
		baseScore += 0.35
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}
	result.RiskScore = baseScore
}

// suggestAlternativeStrategies 建议替代策略
func (a *ECHAnalyzer) suggestAlternativeStrategies(result *ECHAnalysisResult) {
	if !result.ECHPresent {
		return
	}

	// 基于 ECH 类型建议策略
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

	// 通用建议
	if result.Impact.ImpactLevel == "high" || result.Impact.ImpactLevel == "medium" {
		result.AlternativeStrategies = append(result.AlternativeStrategies,
			"实施多层防御策略，不依赖单一指纹方法",
		)
	}
}

// GetImpactSummary 获取影响摘要
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

// ============ 已知配置库 ============

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

// AnalyzeECH 便捷函数：分析 ECH
func AnalyzeECH(data ClientHelloData) (*ECHAnalysisResult, error) {
	analyzer := NewECHAnalyzer()
	return analyzer.AnalyzeClientHello(data)
}
