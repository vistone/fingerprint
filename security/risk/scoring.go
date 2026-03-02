package risk

import fp "github.com/vistone/fingerprint"

// RiskScore 综合风险评分结果。
type RiskScore = fp.RiskScore

// RiskDimensions 风险维度评分。
type RiskDimensions = fp.RiskDimensions

// RiskFactor 风险因素详情。
type RiskFactor = fp.RiskFactor

// RiskInput 风险评分输入数据。
type RiskInput = fp.RiskInput

// RiskContext 风险上下文。
type RiskContext = fp.RiskContext

// ScoringConfig 评分配置。
type ScoringConfig = fp.ScoringConfig

// DimensionWeights 维度权重。
type DimensionWeights = fp.DimensionWeights

// ThreatThresholds 威胁等级阈值。
type ThreatThresholds = fp.ThreatThresholds

// RiskScorer 风险评分器。
type RiskScorer = fp.RiskScorer

// NewScorer 创建风险评分器。
func NewScorer(config *ScoringConfig) *RiskScorer {
	return fp.NewRiskScorer(config)
}

// DefaultConfig 获取默认评分配置。
func DefaultConfig() *ScoringConfig {
	return fp.DefaultScoringConfig()
}

// Calculate 便捷函数：计算综合风险评分。
func Calculate(input RiskInput) (*RiskScore, error) {
	return fp.CalculateRisk(input)
}
