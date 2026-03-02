package fingerprint

import (
	sb "github.com/vistone/fingerprint/security/behavior"
)

// BehaviorSignal 行为信号分析（兼容别名）。
type BehaviorSignal = sb.BehaviorSignal

// TemporalPattern 时序模式（兼容别名）。
type TemporalPattern = sb.TemporalPattern

// ProtocolProportion 协议比例（兼容别名）。
type ProtocolProportion = sb.ProtocolProportion

// RequestBehavior 请求行为（兼容别名）。
type RequestBehavior = sb.RequestBehavior

// BehaviorAnalyzer 行为分析器（兼容别名）。
type BehaviorAnalyzer = sb.BehaviorAnalyzer

// BehaviorAnalysisConfig 行为分析配置（兼容别名）。
type BehaviorAnalysisConfig = sb.BehaviorAnalysisConfig

// NewBehaviorAnalyzer 创建新的行为分析器（兼容入口）。
func NewBehaviorAnalyzer(config *BehaviorAnalysisConfig) *BehaviorAnalyzer {
	return sb.NewBehaviorAnalyzer(config)
}
