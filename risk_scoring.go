package fingerprint

// TODO@Phase-3: 本模块暂未迁移（参见 docs/5-process/modularization/PHASE_3_PLAN.md）
import (
	"fmt"
	"math"

	"github.com/vistone/fingerprint/http/http2"
	"github.com/vistone/fingerprint/http/ja4h"
	"github.com/vistone/fingerprint/network/quic"
	"github.com/vistone/fingerprint/tls/ech"
	"github.com/vistone/fingerprint/tls/ja4s"
)

// RiskScore 综合风险评分结果
type RiskScore struct {
	// 总体风险分数 (0.0-1.0)
	TotalScore float64

	// 威胁等级
	ThreatLevel string // "safe", "low", "medium", "high", "critical"

	// 各维度风险分数
	Dimensions RiskDimensions

	// 风险因素详情
	RiskFactors []RiskFactor

	// 异常指标
	AnomalyCount int

	// 置信度 (0.0-1.0)
	Confidence float64

	// 建议措施
	Recommendations []string
}

// RiskDimensions 风险维度评分
type RiskDimensions struct {
	// TLS 指纹风险 (JA3/JA4)
	TLSFingerprint float64

	// 服务端风险 (JA4S)
	ServerBehavior float64

	// HTTP/2 风险
	HTTP2Signature float64

	// HTTP 请求头风险 (JA4H)
	HTTPHeaders float64

	// QUIC 风险
	QUICSignature float64

	// Client Hints 不一致性
	ClientHints float64

	// ECH 影响
	ECHImpact float64

	// 行为异常
	BehaviorAnomaly float64
}

// RiskFactor 风险因素
type RiskFactor struct {
	// 因素类型
	Type string

	// 严重程度 (0.0-1.0)
	Severity float64

	// 描述
	Description string

	// 证据
	Evidence string

	// 权重
	Weight float64

	// 置信度
	Confidence float64
}

// RiskInput 风险评分输入
type RiskInput struct {
	// JA3/JA4 指纹数据
	JA3Hash string
	JA4Hash string

	// JA4S 结果
	JA4SResult *ja4s.JA4SResult

	// HTTP/2 签名结果
	HTTP2Result *http2.HTTP2SignatureResult

	// JA4H 结果
	JA4HResult *ja4h.JA4HResult

	// QUIC 签名结果
	QUICResult *quic.QUICSignatureResult

	// ECH 分析结果
	ECHResult *ech.ECHAnalysisResult

	// 额外上下文
	Context RiskContext
}

// RiskContext 风险评估上下文
type RiskContext struct {
	// IP 信誉分数 (可选)
	IPReputation float64

	// 地理位置风险 (可选)
	GeoRisk float64

	// 历史行为评分 (可选)
	HistoricalScore float64

	// 请求频率 (可选)
	RequestRate float64

	// 是否已知客户端
	IsKnownClient bool
}

// ScoringConfig 评分配置
type ScoringConfig struct {
	// 各维度权重
	Weights DimensionWeights

	// 威胁等级阈值
	Thresholds ThreatThresholds

	// 是否启用严格模式
	StrictMode bool

	// 最小置信度要求
	MinConfidence float64
}

// DimensionWeights 维度权重
type DimensionWeights struct {
	TLSFingerprint  float64 // 默认 0.20
	ServerBehavior  float64 // 默认 0.15
	HTTP2Signature  float64 // 默认 0.15
	HTTPHeaders     float64 // 默认 0.15
	QUICSignature   float64 // 默认 0.10
	ClientHints     float64 // 默认 0.10
	ECHImpact       float64 // 默认 0.05
	BehaviorAnomaly float64 // 默认 0.10
}

// ThreatThresholds 威胁等级阈值
type ThreatThresholds struct {
	Safe     float64 // <= 0.2
	Low      float64 // <= 0.4
	Medium   float64 // <= 0.6
	High     float64 // <= 0.8
	Critical float64 // > 0.8
}

// RiskScorer 风险评分器
type RiskScorer struct {
	config ScoringConfig
}

// NewRiskScorer 创建风险评分器
func NewRiskScorer(config *ScoringConfig) *RiskScorer {
	if config == nil {
		config = DefaultScoringConfig()
	}
	return &RiskScorer{
		config: *config,
	}
}

// DefaultScoringConfig 默认评分配置
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

// CalculateRisk 计算综合风险评分
func (s *RiskScorer) CalculateRisk(input RiskInput) (*RiskScore, error) {
	result := &RiskScore{
		Dimensions:      RiskDimensions{},
		RiskFactors:     []RiskFactor{},
		Recommendations: []string{},
	}

	// 1. 计算各维度风险
	s.calculateTLSRisk(input, result)
	s.calculateServerRisk(input, result)
	s.calculateHTTP2Risk(input, result)
	s.calculateHTTPHeadersRisk(input, result)
	s.calculateQUICRisk(input, result)
	s.calculateClientHintsRisk(input, result)
	s.calculateECHRisk(input, result)
	s.calculateBehaviorRisk(input, result)

	// 2. 计算加权总分
	result.TotalScore = s.calculateWeightedScore(result.Dimensions)

	// 3. 应用上下文调整
	result.TotalScore = s.applyContextAdjustment(result.TotalScore, input.Context)

	// 4. 确保分数在合理范围
	result.TotalScore = math.Max(0.0, math.Min(1.0, result.TotalScore))

	// 5. 判定威胁等级
	result.ThreatLevel = s.determineThreatLevel(result.TotalScore)

	// 6. 计算置信度
	result.Confidence = s.calculateConfidence(input)

	// 7. 统计异常数量
	result.AnomalyCount = s.countAnomalies(input)

	// 8. 生成建议措施
	result.Recommendations = s.generateRecommendations(result)

	return result, nil
}

// calculateTLSRisk 计算 TLS 指纹风险
func (s *RiskScorer) calculateTLSRisk(input RiskInput, result *RiskScore) {
	score := 0.0

	// 基于 JA3/JA4 哈希的已知度判断
	if input.JA3Hash == "" && input.JA4Hash == "" {
		score = 0.3 // 缺少 TLS 指纹
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Type:        "TLS_FINGERPRINT_MISSING",
			Severity:    0.3,
			Description: "缺少 TLS 指纹数据",
			Weight:      s.config.Weights.TLSFingerprint,
			Confidence:  0.9,
		})
	}

	// TODO: 可以集成已知恶意指纹数据库进行匹配
	result.Dimensions.TLSFingerprint = score
}

// calculateServerRisk 计算服务端行为风险
func (s *RiskScorer) calculateServerRisk(input RiskInput, result *RiskScore) {
	if input.JA4SResult == nil {
		result.Dimensions.ServerBehavior = 0.0
		return
	}

	score := input.JA4SResult.RiskScore

	// 添加风险因素
	for _, anomaly := range input.JA4SResult.AnomalyFlags {
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Type:        fmt.Sprintf("JA4S_%s", anomaly),
			Severity:    0.5,
			Description: fmt.Sprintf("服务端异常: %s", anomaly),
			Evidence:    input.JA4SResult.Hash,
			Weight:      s.config.Weights.ServerBehavior,
			Confidence:  0.8,
		})
	}

	result.Dimensions.ServerBehavior = score
}

// calculateHTTP2Risk 计算 HTTP/2 风险
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
			Description: fmt.Sprintf("HTTP/2 异常: %s", anomaly),
			Weight:      s.config.Weights.HTTP2Signature,
			Confidence:  0.7,
		})
	}

	result.Dimensions.HTTP2Signature = score
}

// calculateHTTPHeadersRisk 计算 HTTP 请求头风险
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
			Description: fmt.Sprintf("HTTP 请求头异常: %s", anomaly),
			Weight:      s.config.Weights.HTTPHeaders,
			Confidence:  0.85,
		})
	}

	result.Dimensions.HTTPHeaders = score
}

// calculateQUICRisk 计算 QUIC 风险
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
			Description: fmt.Sprintf("QUIC 异常: %s", anomaly),
			Weight:      s.config.Weights.QUICSignature,
			Confidence:  0.75,
		})
	}

	result.Dimensions.QUICSignature = score
}

// calculateClientHintsRisk 计算 Client Hints 风险
func (s *RiskScorer) calculateClientHintsRisk(input RiskInput, result *RiskScore) {
	if input.JA4HResult == nil {
		result.Dimensions.ClientHints = 0.0
		return
	}

	// Client Hints 不一致性检测在 JA4H 中完成
	// 通过异常标记识别 Client Hints 相关问题
	score := 0.0

	// 如果有 Client Hints 相关的异常标记
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

// calculateECHRisk 计算 ECH 影响
func (s *RiskScorer) calculateECHRisk(input RiskInput, result *RiskScore) {
	if input.ECHResult == nil {
		result.Dimensions.ECHImpact = 0.0
		return
	}

	score := input.ECHResult.RiskScore

	// ECH 本身不是风险，但会影响检测能力
	if input.ECHResult.ECHPresent && input.ECHResult.Impact.ImpactLevel == "high" {
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Type:        "ECH_HIGH_IMPACT",
			Severity:    0.3,
			Description: "ECH 高影响：SNI 不可见，需使用替代检测方法",
			Evidence:    input.ECHResult.VisibleFieldsSignature,
			Weight:      s.config.Weights.ECHImpact,
			Confidence:  0.9,
		})
	}

	// ECH 配置错误是真实风险
	for _, anomaly := range input.ECHResult.AnomalyFlags {
		if anomaly != "GREASE_ECH" {
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Type:        fmt.Sprintf("ECH_%s", anomaly),
				Severity:    0.6,
				Description: fmt.Sprintf("ECH 异常: %s", anomaly),
				Weight:      s.config.Weights.ECHImpact,
				Confidence:  0.85,
			})
		}
	}

	result.Dimensions.ECHImpact = score
}

// calculateBehaviorRisk 计算行为异常风险
func (s *RiskScorer) calculateBehaviorRisk(input RiskInput, result *RiskScore) {
	score := 0.0

	// 基于上下文的行为分析
	if input.Context.RequestRate > 100 {
		score += 0.3
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Type:        "HIGH_REQUEST_RATE",
			Severity:    0.3,
			Description: "异常高的请求频率",
			Evidence:    fmt.Sprintf("%.0f req/min", input.Context.RequestRate),
			Weight:      s.config.Weights.BehaviorAnomaly,
			Confidence:  0.7,
		})
	}

	result.Dimensions.BehaviorAnomaly = math.Min(score, 1.0)
}

// calculateWeightedScore 计算加权总分
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

// applyContextAdjustment 应用上下文调整
func (s *RiskScorer) applyContextAdjustment(score float64, ctx RiskContext) float64 {
	adjustment := 0.0

	// IP 信誉调整
	if ctx.IPReputation > 0 {
		adjustment += (1.0 - ctx.IPReputation) * 0.1
	}

	// 地理位置风险
	if ctx.GeoRisk > 0 {
		adjustment += ctx.GeoRisk * 0.05
	}

	// 历史行为
	if ctx.HistoricalScore > 0 {
		// 历史良好可以降低风险
		adjustment -= (1.0 - ctx.HistoricalScore) * 0.1
	}

	// 已知客户端降低风险
	if ctx.IsKnownClient {
		adjustment -= 0.1
	}

	return score + adjustment
}

// determineThreatLevel 判定威胁等级
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

// calculateConfidence 计算置信度
func (s *RiskScorer) calculateConfidence(input RiskInput) float64 {
	confidence := 0.0
	count := 0

	// 根据可用数据量计算置信度
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

	// 上下文数据增加置信度
	if input.Context.IPReputation > 0 {
		confidence += 0.10
	}
	if input.Context.HistoricalScore > 0 {
		confidence += 0.10
	}

	return math.Min(confidence, 1.0)
}

// countAnomalies 统计异常数量
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

// generateRecommendations 生成建议措施
func (s *RiskScorer) generateRecommendations(result *RiskScore) []string {
	recommendations := []string{}

	switch result.ThreatLevel {
	case "safe":
		recommendations = append(recommendations, "无需特殊处理，继续正常监控")

	case "low":
		recommendations = append(recommendations, "启用基础监控，记录请求特征")
		if result.Dimensions.TLSFingerprint > 0.2 {
			recommendations = append(recommendations, "关注 TLS 指纹变化")
		}

	case "medium":
		recommendations = append(recommendations, "增强监控频率，分析行为模式")
		recommendations = append(recommendations, "考虑启用额外验证（如 CAPTCHA）")
		if result.AnomalyCount > 3 {
			recommendations = append(recommendations, "多个异常指标，建议详细检查")
		}

	case "high":
		recommendations = append(recommendations, "限制访问频率，启用严格验证")
		recommendations = append(recommendations, "记录完整请求上下文用于分析")
		recommendations = append(recommendations, "考虑临时阻断，人工审核")
		if result.Dimensions.HTTPHeaders > 0.6 {
			recommendations = append(recommendations, "HTTP 请求头高度可疑，可能存在伪造")
		}

	case "critical":
		recommendations = append(recommendations, "立即阻断该请求")
		recommendations = append(recommendations, "记录所有相关特征用于威胁情报")
		recommendations = append(recommendations, "触发安全告警，通知管理员")
		recommendations = append(recommendations, "检查是否存在关联攻击")
	}

	// 针对 ECH 高影响的建议
	if result.Dimensions.ECHImpact > 0.2 {
		recommendations = append(recommendations, "检测到 ECH，建议使用可见字段指纹和行为分析")
	}

	// 针对置信度的建议
	if result.Confidence < s.config.MinConfidence {
		recommendations = append(recommendations, "置信度不足，建议收集更多指纹数据")
	}

	return recommendations
}

// GetSummary 获取风险摘要
func (r *RiskScore) GetSummary() string {
	return fmt.Sprintf("威胁等级: %s, 风险分数: %.2f, 置信度: %.2f, 异常数: %d",
		r.ThreatLevel,
		r.TotalScore,
		r.Confidence,
		r.AnomalyCount,
	)
}

// CalculateRisk 便捷函数：使用默认配置计算风险
func CalculateRisk(input RiskInput) (*RiskScore, error) {
	scorer := NewRiskScorer(nil)
	return scorer.CalculateRisk(input)
}
