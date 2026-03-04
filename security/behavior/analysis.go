package behavior

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/vistone/fingerprint/internal/metrics"
)

// BehaviorSignal 行为信号分析
type BehaviorSignal struct {
	// 信号类型
	SignalType string

	// 信号值 (0-1 之间)
	Score float64

	// 信号描述
	Description string

	// 检测时间
	Timestamp time.Time

	// 风险等级
	RiskLevel string // "safe", "low", "medium", "high", "critical"
}

// TemporalPattern 时序模式
type TemporalPattern struct {
	// 请求间隔（毫秒）
	Intervals []int64

	// 平均间隔
	MeanInterval float64

	// 标准差
	StdDev float64

	// 最小/最大间隔
	MinInterval int64
	MaxInterval int64

	// 规律性指数 (0-1，越高越规律)
	RegularityIndex float64

	// 异常间隔数
	AnomalousIntervals int
}

// ProtocolProportion 协议比例
type ProtocolProportion struct {
	// TLS 版本分布
	TLSVersions map[string]float64 // "1.2" -> 0.6, "1.3" -> 0.4

	// Cipher Suite 分布 (top 5)
	TopCipherSuites map[string]float64

	// Extension 分布 (top 10)
	TopExtensions map[string]float64

	// HTTP 版本分布
	HTTPVersions map[string]float64 // "1.1" -> 0.3, "2" -> 0.7

	// ALPN 协议分布
	ALPNProtocols map[string]float64

	// 熵值 (用于检测不自然的分布)
	EntropyScore float64

	// 是否存在异常分布
	IsAnomalous bool
}

// RequestBehavior 单个请求行为
type RequestBehavior struct {
	// 时间戳
	Timestamp time.Time

	// TLS 版本
	TLSVersion string

	// Cipher Suite
	CipherSuite string

	// 使用的 Extensions
	Extensions []string

	// HTTP 版本
	HTTPVersion string

	// 请求大小
	RequestSize int

	// 响应大小
	ResponseSize int

	// 延迟
	Latency time.Duration

	// 是否使用了连接复用
	ReusingConnection bool

	// 源地址
	SourceIP string

	// 目标地址
	DestinationIP string

	// 目标端口
	DestinationPort int

	// SNI
	SNI string
}

// BehaviorAnalyzer 行为信号分析器
type BehaviorAnalyzer struct {
	// 请求历史
	requestHistory []RequestBehavior

	// 时序模式
	temporalPatterns map[string]*TemporalPattern

	// 协议比例
	protocolProportions map[string]*ProtocolProportion

	// 检测到的信号
	signals []BehaviorSignal

	// 分析参数
	config *BehaviorAnalysisConfig
}

// BehaviorAnalysisConfig 分析配置
type BehaviorAnalysisConfig struct {
	// 时间窗口大小 (秒)
	TimeWindowSize int

	// 最小请求数
	MinRequestsForAnalysis int

	// 规律性阈值 (异常检测)
	RegularityThreshold float64

	// 熵值阈值
	EntropyThreshold float64

	// 异常间隔率阈值
	AnomalousIntervalRateThreshold float64
}

// NewBehaviorAnalyzer 创建行为分析器
func NewBehaviorAnalyzer(config *BehaviorAnalysisConfig) *BehaviorAnalyzer {
	if config == nil {
		config = &BehaviorAnalysisConfig{
			TimeWindowSize:                 60,
			MinRequestsForAnalysis:         5,
			RegularityThreshold:            0.3,
			EntropyThreshold:               0.5,
			AnomalousIntervalRateThreshold: 0.2,
		}
	}

	return &BehaviorAnalyzer{
		requestHistory:      make([]RequestBehavior, 0, 100), // 预分配容量
		temporalPatterns:    make(map[string]*TemporalPattern),
		protocolProportions: make(map[string]*ProtocolProportion),
		signals:             make([]BehaviorSignal, 0, 50), // 预分配容量
		config:              config,
	}
}

// AddRequest 添加请求
func (ba *BehaviorAnalyzer) AddRequest(req RequestBehavior) {
	ba.requestHistory = append(ba.requestHistory, req)
}

// AnalyzeTemporalPattern 分析时序模式
func (ba *BehaviorAnalyzer) AnalyzeTemporalPattern(origin string) *TemporalPattern {
	// 筛选该源的请求并计算间隔（单次遍历，减少内存分配）
	intervals := make([]int64, 0, len(ba.requestHistory))
	var lastReq *RequestBehavior

	for i := range ba.requestHistory {
		req := &ba.requestHistory[i]
		// 使用 SNI 或目标地址作为源识别
		if req.SNI == origin || req.DestinationIP == origin {
			if lastReq != nil {
				interval := req.Timestamp.Sub(lastReq.Timestamp).Milliseconds()
				intervals = append(intervals, interval)
			}
			lastReq = req
		}
	}

	if len(intervals) < ba.config.MinRequestsForAnalysis-1 {
		return nil
	}

	// 计算统计特性（零分配计算）
	pattern := ba.calculateTemporalStatsZeroAlloc(intervals)

	// 计算规律性指数
	ba.calculateRegularityIndex(pattern)

	// 检测异常间隔
	ba.detectAnomalousIntervals(pattern)

	ba.temporalPatterns[origin] = pattern

	return pattern
}

// calculateTemporalStats 计算时序统计
func (ba *BehaviorAnalyzer) calculateTemporalStats(pattern *TemporalPattern) {
	if len(pattern.Intervals) == 0 {
		return
	}

	// 计算平均值
	var sum int64
	pattern.MinInterval = pattern.Intervals[0]
	pattern.MaxInterval = pattern.Intervals[0]

	for _, interval := range pattern.Intervals {
		sum += interval
		if interval < pattern.MinInterval {
			pattern.MinInterval = interval
		}
		if interval > pattern.MaxInterval {
			pattern.MaxInterval = interval
		}
	}

	pattern.MeanInterval = float64(sum) / float64(len(pattern.Intervals))

	// 计算标准差
	var variance float64
	for _, interval := range pattern.Intervals {
		diff := float64(interval) - pattern.MeanInterval
		variance += diff * diff
	}
	pattern.StdDev = math.Sqrt(variance / float64(len(pattern.Intervals)))
}

// calculateTemporalStatsZeroAlloc 计算时序统计（零分配版本）
func (ba *BehaviorAnalyzer) calculateTemporalStatsZeroAlloc(intervals []int64) *TemporalPattern {
	pattern := &TemporalPattern{
		Intervals: intervals,
	}

	if len(intervals) == 0 {
		return pattern
	}

	// 计算平均值、最小值、最大值（单次遍历）
	var sum int64
	pattern.MinInterval = intervals[0]
	pattern.MaxInterval = intervals[0]

	for _, interval := range intervals {
		sum += interval
		if interval < pattern.MinInterval {
			pattern.MinInterval = interval
		}
		if interval > pattern.MaxInterval {
			pattern.MaxInterval = interval
		}
	}

	pattern.MeanInterval = float64(sum) / float64(len(intervals))

	// 计算标准差
	var variance float64
	for _, interval := range intervals {
		diff := float64(interval) - pattern.MeanInterval
		variance += diff * diff
	}
	pattern.StdDev = math.Sqrt(variance / float64(len(intervals)))

	return pattern
}

// calculateRegularityIndex 计算规律性指数
// 规律性高 (接近 1) 说明请求间隔很规则，可能是机器人
// 规律性低 (接近 0) 说明请求间隔随机，可能是真实用户
func (ba *BehaviorAnalyzer) calculateRegularityIndex(pattern *TemporalPattern) {
	if pattern.MeanInterval == 0 || pattern.StdDev == 0 {
		pattern.RegularityIndex = 0
		return
	}

	// 变异系数 (CV) = 标准差 / 平均值
	cv := pattern.StdDev / pattern.MeanInterval

	// 规律性指数 = 1 - CV，限制在 [0, 1]
	pattern.RegularityIndex = 1.0 - cv
	if pattern.RegularityIndex < 0 {
		pattern.RegularityIndex = 0
	}
	if pattern.RegularityIndex > 1 {
		pattern.RegularityIndex = 1
	}
}

// detectAnomalousIntervals 检测异常间隔
func (ba *BehaviorAnalyzer) detectAnomalousIntervals(pattern *TemporalPattern) {
	if pattern.StdDev == 0 {
		return
	}

	// 使用 3-sigma 规则检测异常值
	threshold := pattern.MeanInterval + 3*pattern.StdDev

	for _, interval := range pattern.Intervals {
		if float64(interval) > threshold {
			pattern.AnomalousIntervals++
		}
	}
}

// AnalyzeProtocolProportion 分析协议比例
func (ba *BehaviorAnalyzer) AnalyzeProtocolProportion(origin string) *ProtocolProportion {
	var originRequests []RequestBehavior
	for _, req := range ba.requestHistory {
		if req.SNI == origin || req.DestinationIP == origin {
			originRequests = append(originRequests, req)
		}
	}

	if len(originRequests) < ba.config.MinRequestsForAnalysis {
		return nil
	}

	prop := &ProtocolProportion{
		TLSVersions:     make(map[string]float64),
		TopCipherSuites: make(map[string]float64),
		TopExtensions:   make(map[string]float64),
		HTTPVersions:    make(map[string]float64),
		ALPNProtocols:   make(map[string]float64),
	}

	// 统计 TLS 版本
	for _, req := range originRequests {
		prop.TLSVersions[req.TLSVersion]++
		prop.HTTPVersions[req.HTTPVersion]++
		prop.TopCipherSuites[req.CipherSuite]++
	}

	// 转换为比例
	for ver := range prop.TLSVersions {
		prop.TLSVersions[ver] /= float64(len(originRequests))
	}
	for ver := range prop.HTTPVersions {
		prop.HTTPVersions[ver] /= float64(len(originRequests))
	}
	for cs := range prop.TopCipherSuites {
		prop.TopCipherSuites[cs] /= float64(len(originRequests))
	}

	// 仅保留 top 5 cipher suites
	ba.keepTopN(prop.TopCipherSuites, 5)

	// 统计 extensions
	extensionCount := make(map[string]int)
	for _, req := range originRequests {
		for _, ext := range req.Extensions {
			extensionCount[ext]++
		}
	}

	// 转换为比例并保留 top 10
	for ext, count := range extensionCount {
		prop.TopExtensions[ext] = float64(count) / float64(len(originRequests))
	}
	ba.keepTopN(prop.TopExtensions, 10)

	// 计算熵值和异常检测
	ba.calculateProtocolEntropy(prop)

	ba.protocolProportions[origin] = prop

	return prop
}

// keepTopN 仅保留前 N 个最高的条目
func (ba *BehaviorAnalyzer) keepTopN(m map[string]float64, n int) {
	if len(m) <= n {
		return
	}

	// 创建排序的列表
	type kv struct {
		Key   string
		Value float64
	}
	var items []kv
	for k, v := range m {
		items = append(items, kv{k, v})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Value > items[j].Value
	})

	// 清空并仅保留前 N 个
	for k := range m {
		delete(m, k)
	}
	for i := 0; i < n && i < len(items); i++ {
		m[items[i].Key] = items[i].Value
	}
}

// calculateProtocolEntropy 计算协议分布的熵值
func (ba *BehaviorAnalyzer) calculateProtocolEntropy(prop *ProtocolProportion) {
	// 计算 TLS 版本分布的熵
	var entropy float64
	for _, prob := range prop.TLSVersions {
		if prob > 0 {
			entropy += -prob * math.Log2(prob)
		}
	}
	prop.EntropyScore = entropy

	// 检测异常：如果所有请求都使用相同的 TLS 版本或 Cipher Suite，则为异常
	if len(prop.TLSVersions) == 1 || len(prop.TopCipherSuites) == 1 {
		prop.IsAnomalous = true
	}

	// 检测熵值异常
	// 正常的真实用户会有多样化的协议版本
	if entropy < ba.config.EntropyThreshold && len(prop.TLSVersions) > 1 {
		prop.IsAnomalous = true
	}
}

// GenerateBehaviorSignals 生成行为信号
func (ba *BehaviorAnalyzer) GenerateBehaviorSignals(origin string) []BehaviorSignal {
	signals := []BehaviorSignal{}

	// 分析时序模式
	if pattern := ba.AnalyzeTemporalPattern(origin); pattern != nil {
		sig, shouldAdd := ba.evaluateTemporalPattern(pattern)
		if shouldAdd {
			signals = append(signals, sig)
		}
	}

	// 分析协议比例
	if proportion := ba.AnalyzeProtocolProportion(origin); proportion != nil {
		sig, shouldAdd := ba.evaluateProtocolProportion(proportion)
		if shouldAdd {
			signals = append(signals, sig)
		}
	}

	// 检测连接复用行为
	if sig, shouldAdd := ba.detectConnectionReuse(origin); shouldAdd {
		signals = append(signals, sig)
	}

	// 记录指标
	for _, sig := range signals {
		metrics.RecordBehaviorSignal(sig.RiskLevel)
	}

	ba.signals = append(ba.signals, signals...)
	return signals
}

// evaluateTemporalPattern 评估时序模式
func (ba *BehaviorAnalyzer) evaluateTemporalPattern(pattern *TemporalPattern) (BehaviorSignal, bool) {
	var sig BehaviorSignal
	sig.SignalType = "TEMPORAL_PATTERN"
	sig.Timestamp = time.Now()

	anomalousRate := float64(pattern.AnomalousIntervals) / float64(len(pattern.Intervals))

	if pattern.RegularityIndex > 0.8 {
		// 高规律性：可能是机器人或自动化脚本
		sig.Score = pattern.RegularityIndex
		sig.Description = fmt.Sprintf("高度规律的请求间隔: 规律性指数=%.2f, 平均间隔=%.0fms",
			pattern.RegularityIndex, pattern.MeanInterval)
		sig.RiskLevel = "high"
		return sig, true
	}

	if anomalousRate > ba.config.AnomalousIntervalRateThreshold {
		// 高比例的异常间隔：不稳定的连接或异常行为
		sig.Score = anomalousRate
		sig.Description = fmt.Sprintf("异常请求间隔率: %.1f%%, 可能的网络问题或异常行为",
			anomalousRate*100)
		sig.RiskLevel = "medium"
		return sig, true
	}

	return sig, false
}

// evaluateProtocolProportion 评估协议比例
func (ba *BehaviorAnalyzer) evaluateProtocolProportion(prop *ProtocolProportion) (BehaviorSignal, bool) {
	var sig BehaviorSignal
	sig.SignalType = "PROTOCOL_PROPORTION"
	sig.Timestamp = time.Now()

	if prop.IsAnomalous {
		sig.Score = prop.EntropyScore
		sig.Description = fmt.Sprintf("异常的协议分布: 熵值=%.2f, TLS版本数=%d, CipherSuite数=%d",
			prop.EntropyScore, len(prop.TLSVersions), len(prop.TopCipherSuites))
		sig.RiskLevel = "high"
		return sig, true
	}

	return sig, false
}

// detectConnectionReuse 检测连接复用行为
func (ba *BehaviorAnalyzer) detectConnectionReuse(origin string) (BehaviorSignal, bool) {
	var reusingCount, totalCount int

	for _, req := range ba.requestHistory {
		if req.SNI == origin || req.DestinationIP == origin {
			totalCount++
			if req.ReusingConnection {
				reusingCount++
			}
		}
	}

	if totalCount < ba.config.MinRequestsForAnalysis {
		return BehaviorSignal{}, false
	}

	reuseRate := float64(reusingCount) / float64(totalCount)

	// 如果连接复用率过高（接近 100%），可能是自动化脚本
	// 真实用户通常会有一些新连接
	if reuseRate > 0.9 {
		return BehaviorSignal{
			SignalType:  "CONNECTION_REUSE",
			Score:       reuseRate,
			Description: fmt.Sprintf("极高的连接复用率: %.1f%%", reuseRate*100),
			Timestamp:   time.Now(),
			RiskLevel:   "high",
		}, true
	}

	return BehaviorSignal{}, false
}

// GetAllSignals 获取所有信号
func (ba *BehaviorAnalyzer) GetAllSignals() []BehaviorSignal {
	return ba.signals
}

// GetRiskScore 计算综合风险分数
func (ba *BehaviorAnalyzer) GetRiskScore() float64 {
	if len(ba.signals) == 0 {
		return 0
	}

	var totalScore float64
	var weight float64

	for _, sig := range ba.signals {
		w := 1.0
		if sig.RiskLevel == "critical" {
			w = 1.0
		} else if sig.RiskLevel == "high" {
			w = 0.8
		} else if sig.RiskLevel == "medium" {
			w = 0.5
		} else if sig.RiskLevel == "low" {
			w = 0.2
		}

		totalScore += sig.Score * w
		weight += w
	}

	if weight == 0 {
		return 0
	}

	return totalScore / weight
}

// GetAnalysisSummary 获取分析总结
func (ba *BehaviorAnalyzer) GetAnalysisSummary() string {
	if len(ba.requestHistory) == 0 {
		return "未收集到请求数据"
	}

	summary := fmt.Sprintf("行为分析报告\n")
	summary += fmt.Sprintf("=== 基础信息 ===\n")
	summary += fmt.Sprintf("总请求数: %d\n", len(ba.requestHistory))
	summary += fmt.Sprintf("检测到的信号: %d\n", len(ba.signals))
	summary += fmt.Sprintf("综合风险分数: %.2f\n\n", ba.GetRiskScore())

	if len(ba.signals) == 0 {
		summary += "未检测到异常信号\n"
		return summary
	}

	summary += fmt.Sprintf("=== 检测到的信号 ===\n")
	for i, sig := range ba.signals {
		summary += fmt.Sprintf("%d. [%s] %s (风险级别: %s, 评分: %.2f)\n",
			i+1, sig.SignalType, sig.Description, sig.RiskLevel, sig.Score)
	}

	return summary
}
