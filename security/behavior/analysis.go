package behavior

import fp "github.com/vistone/fingerprint"

// BehaviorSignal 行为信号分析结果。
type BehaviorSignal = fp.BehaviorSignal

// TemporalPattern 时间模式分析结果。
type TemporalPattern = fp.TemporalPattern

// ProtocolProportion 协议占比分析结果。
type ProtocolProportion = fp.ProtocolProportion

// RequestBehavior 请求行为样本。
type RequestBehavior = fp.RequestBehavior

// BehaviorAnalyzer 行为分析器。
type BehaviorAnalyzer = fp.BehaviorAnalyzer

// BehaviorAnalysisConfig 行为分析配置。
type BehaviorAnalysisConfig = fp.BehaviorAnalysisConfig

// NewAnalyzer 创建行为分析器。
func NewAnalyzer(config *BehaviorAnalysisConfig) *BehaviorAnalyzer {
	return fp.NewBehaviorAnalyzer(config)
}
