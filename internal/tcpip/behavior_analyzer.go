package tcpip

import (
	"fmt"
	"time"
)

// NetworkBehaviorAnalyzer 网络行为分析器
type NetworkBehaviorAnalyzer struct {
	packets         []*TCPPacket
	rttMeasurements []time.Duration
	sequenceNumbers []uint32
	ipIDs           []uint16
	timestamps      []time.Time
}

// NewNetworkBehaviorAnalyzer 创建新的网络行为分析器
func NewNetworkBehaviorAnalyzer() *NetworkBehaviorAnalyzer {
	return &NetworkBehaviorAnalyzer{
		packets:         make([]*TCPPacket, 0),
		rttMeasurements: make([]time.Duration, 0),
		sequenceNumbers: make([]uint32, 0),
		ipIDs:           make([]uint16, 0),
		timestamps:      make([]time.Time, 0),
	}
}

// RecordPacket 记录数据包
func (nba *NetworkBehaviorAnalyzer) RecordPacket(packet *TCPPacket, rtt time.Duration) {
	nba.packets = append(nba.packets, packet)
	nba.rttMeasurements = append(nba.rttMeasurements, rtt)
	nba.sequenceNumbers = append(nba.sequenceNumbers, packet.SequenceNumber)
	if packet.IPHeader != nil {
		nba.ipIDs = append(nba.ipIDs, packet.IPHeader.Identification)
	}
	nba.timestamps = append(nba.timestamps, time.Now())
}

// AnalyzeBehavior 分析网络行为
func (nba *NetworkBehaviorAnalyzer) AnalyzeBehavior() *NetworkBehaviorResult {
	result := &NetworkBehaviorResult{
		TotalPackets: len(nba.packets),
		Timestamp:    time.Now(),
	}

	if len(nba.packets) == 0 {
		return result
	}

	// 分析 RTT
	result.RTTAnalysis = nba.analyzeRTT()

	// 分析序列号
	result.SequenceNumberPattern = nba.analyzeSequenceNumbers()

	// 分析 IP ID
	result.IPIDPattern = nba.analyzeIPIDs()

	// 分析数据包大小
	result.PacketSizeVariance = nba.analyzePacketSizes()

	// 分析协议分布
	result.ProtocolDistribution = nba.analyzeProtocolDistribution()

	// 分析时间模式
	result.TimingPattern = nba.analyzeTimingPattern()

	// 计算行为特征
	result.BehaviorCharacteristics = nba.computeBehaviorCharacteristics()

	return result
}

// analyzeRTT 分析往返时间
func (nba *NetworkBehaviorAnalyzer) analyzeRTT() *RTTAnalysis {
	if len(nba.rttMeasurements) == 0 {
		return &RTTAnalysis{}
	}

	analysis := &RTTAnalysis{
		Count:   len(nba.rttMeasurements),
		Samples: nba.rttMeasurements,
	}

	var sum time.Duration
	minRTT := nba.rttMeasurements[0]
	maxRTT := nba.rttMeasurements[0]

	for _, rtt := range nba.rttMeasurements {
		sum += rtt
		if rtt < minRTT {
			minRTT = rtt
		}
		if rtt > maxRTT {
			maxRTT = rtt
		}
	}

	analysis.AverageRTT = sum / time.Duration(analysis.Count)
	analysis.MinRTT = minRTT
	analysis.MaxRTT = maxRTT
	analysis.StdDeviation = nba.calculateStdDeviation(nba.rttMeasurements, analysis.AverageRTT)
	analysis.NetworkType = nba.classifyNetworkType(analysis.AverageRTT)

	return analysis
}

// analyzeSequenceNumbers 分析序列号模式
func (nba *NetworkBehaviorAnalyzer) analyzeSequenceNumbers() string {
	if len(nba.sequenceNumbers) < 2 {
		return "insufficient_data"
	}

	// 检查序列号增长模式
	diffs := make([]int64, len(nba.sequenceNumbers)-1)
	for i := 0; i < len(diffs); i++ {
		diffs[i] = int64(nba.sequenceNumbers[i+1]) - int64(nba.sequenceNumbers[i])
	}

	// 检查是否为随机
	if nba.hasHighVariance(diffs) {
		return "random"
	}

	// 检查是否为时间相关
	if nba.isTimeRelated(diffs) {
		return "time_based"
	}

	// 检查是否为线性/顺序
	if nba.isLinear(diffs) {
		return "sequential"
	}

	return "complex_pattern"
}

// analyzeIPIDs 分析 IP ID 模式
func (nba *NetworkBehaviorAnalyzer) analyzeIPIDs() string {
	if len(nba.ipIDs) < 2 {
		return "insufficient_data"
	}

	// 计算 IP ID 的差值
	diffs := make([]int, len(nba.ipIDs)-1)
	for i := 0; i < len(diffs); i++ {
		diffs[i] = int(nba.ipIDs[i+1]) - int(nba.ipIDs[i])
	}

	// 检查线性计数器（NATing 设备的典型行为）
	if nba.isLinearIPID(diffs) {
		return "linear_counter"
	}

	// 检查是否随机
	if nba.hasHighVariance(convertToInt64(diffs)) {
		return "random"
	}

	return "mixed_pattern"
}

// analyzePacketSizes 分析数据包大小
func (nba *NetworkBehaviorAnalyzer) analyzePacketSizes() float64 {
	if len(nba.packets) == 0 {
		return 0
	}

	sizes := make([]int64, len(nba.packets))
	var sum int64 = 0

	for i, pkt := range nba.packets {
		size := int64(len(pkt.Payload))
		sizes[i] = size
		sum += size
	}

	mean := sum / int64(len(sizes))

	var variance int64 = 0
	for _, size := range sizes {
		diff := size - mean
		variance += diff * diff
	}

	if len(sizes) > 0 {
		variance /= int64(len(sizes))
	}

	return float64(variance)
}

// analyzeProtocolDistribution 分析协议分布
func (nba *NetworkBehaviorAnalyzer) analyzeProtocolDistribution() map[string]int {
	distribution := make(map[string]int)

	for _, pkt := range nba.packets {
		if pkt.IPHeader != nil {
			switch pkt.IPHeader.Protocol {
			case 6:
				distribution["TCP"]++
			case 17:
				distribution["UDP"]++
			case 1:
				distribution["ICMP"]++
			default:
				distribution["OTHER"]++
			}
		}
	}

	return distribution
}

// analyzeTimingPattern 分析时间模式
func (nba *NetworkBehaviorAnalyzer) analyzeTimingPattern() string {
	if len(nba.timestamps) < 2 {
		return "insufficient_data"
	}

	intervals := make([]time.Duration, len(nba.timestamps)-1)
	for i := 0; i < len(intervals); i++ {
		intervals[i] = nba.timestamps[i+1].Sub(nba.timestamps[i])
	}

	// 检查是否有规律间隔
	if nba.hasRegularIntervals(intervals) {
		return "periodic"
	}

	// 检查是否为突发
	if nba.isBurstPattern(intervals) {
		return "bursty"
	}

	return "irregular"
}

// computeBehaviorCharacteristics 计算行为特征
func (nba *NetworkBehaviorAnalyzer) computeBehaviorCharacteristics() []string {
	characteristics := make([]string, 0)

	if len(nba.packets) == 0 {
		return characteristics
	}

	// 检查是否为自动化流量
	if nba.isAutomatedTraffic() {
		characteristics = append(characteristics, "automated")
	}

	// 检查是否为交互式流量
	if nba.isInteractiveTraffic() {
		characteristics = append(characteristics, "interactive")
	}

	// 检查是否为批量传输
	if nba.isBulkTransfer() {
		characteristics = append(characteristics, "bulk_transfer")
	}

	// 检查是否为扫描行为
	if nba.isScanningBehavior() {
		characteristics = append(characteristics, "scanning")
	}

	return characteristics
}

// 辅助函数
func (nba *NetworkBehaviorAnalyzer) calculateStdDeviation(measurements []time.Duration, avg time.Duration) time.Duration {
	if len(measurements) == 0 {
		return 0
	}

	var sumSq time.Duration
	for _, m := range measurements {
		diff := m - avg
		sumSq += diff * diff
	}

	variance := sumSq / time.Duration(len(measurements))
	// 返回简化的标准差估计
	return time.Duration(int64(variance.Nanoseconds()) / 2)
}

func (nba *NetworkBehaviorAnalyzer) classifyNetworkType(avgRTT time.Duration) string {
	switch {
	case avgRTT < 10*time.Millisecond:
		return "local_lan"
	case avgRTT < 50*time.Millisecond:
		return "domestic"
	case avgRTT < 150*time.Millisecond:
		return "regional"
	default:
		return "international"
	}
}

func (nba *NetworkBehaviorAnalyzer) hasHighVariance(diffs []int64) bool {
	if len(diffs) < 2 {
		return false
	}

	var sum int64
	for _, d := range diffs {
		if d < 0 {
			sum -= d
		} else {
			sum += d
		}
	}

	avg := sum / int64(len(diffs))
	return avg > 1000 // 高方差阈值
}

func (nba *NetworkBehaviorAnalyzer) isTimeRelated(diffs []int64) bool {
	if len(diffs) < 2 {
		return false
	}

	// 时间相关的序列号会有一定的递增模式
	increasing := 0
	for _, d := range diffs {
		if d > 0 && d < 100000 {
			increasing++
		}
	}

	return float64(increasing)/float64(len(diffs)) > 0.7
}

func (nba *NetworkBehaviorAnalyzer) isLinear(diffs []int64) bool {
	if len(diffs) < 2 {
		return false
	}

	firstDiff := diffs[0]
	for i := 1; i < len(diffs); i++ {
		if diffs[i] != firstDiff {
			return false
		}
	}

	return true
}

func (nba *NetworkBehaviorAnalyzer) isLinearIPID(diffs []int) bool {
	if len(diffs) < 2 {
		return false
	}

	for i := 1; i < len(diffs); i++ {
		if diffs[i] == 0 || diffs[i] == 1 {
			continue
		}
		return false
	}

	return true
}

func (nba *NetworkBehaviorAnalyzer) hasRegularIntervals(intervals []time.Duration) bool {
	if len(intervals) < 3 {
		return false
	}

	baseInterval := intervals[0]
	tolerance := baseInterval / 4

	matches := 0
	for _, interval := range intervals {
		if interval >= baseInterval-tolerance && interval <= baseInterval+tolerance {
			matches++
		}
	}

	return float64(matches)/float64(len(intervals)) > 0.7
}

func (nba *NetworkBehaviorAnalyzer) isBurstPattern(intervals []time.Duration) bool {
	if len(intervals) < 3 {
		return false
	}

	small := 0
	large := 0

	for _, interval := range intervals {
		if interval < 100*time.Millisecond {
			small++
		} else {
			large++
		}
	}

	return small > 0 && large > 0 && (small > len(intervals)/2 || large > len(intervals)/2)
}

func (nba *NetworkBehaviorAnalyzer) isAutomatedTraffic() bool {
	return len(nba.packets) > 100 && len(nba.rttMeasurements) > 0
}

func (nba *NetworkBehaviorAnalyzer) isInteractiveTraffic() bool {
	if len(nba.rttMeasurements) < 2 {
		return false
	}

	var sum time.Duration
	for _, rtt := range nba.rttMeasurements {
		sum += rtt
	}

	avg := sum / time.Duration(len(nba.rttMeasurements))
	return avg > 50*time.Millisecond && avg < 500*time.Millisecond
}

func (nba *NetworkBehaviorAnalyzer) isBulkTransfer() bool {
	if len(nba.packets) == 0 {
		return false
	}

	return len(nba.packets) > 50
}

func (nba *NetworkBehaviorAnalyzer) isScanningBehavior() bool {
	if len(nba.packets) < 10 {
		return false
	}

	// 扫描行为特征：连接快速失败
	resetCount := 0
	for _, pkt := range nba.packets {
		if pkt.Flags&0x04 != 0 { // RST 标志
			resetCount++
		}
	}

	return float64(resetCount)/float64(len(nba.packets)) > 0.3
}

func convertToInt64(diffs []int) []int64 {
	result := make([]int64, len(diffs))
	for i, d := range diffs {
		result[i] = int64(d)
	}
	return result
}

// NetworkBehaviorResult 网络行为分析结果
type NetworkBehaviorResult struct {
	TotalPackets            int
	RTTAnalysis             *RTTAnalysis
	SequenceNumberPattern   string
	IPIDPattern             string
	PacketSizeVariance      float64
	ProtocolDistribution    map[string]int
	TimingPattern           string
	BehaviorCharacteristics []string
	Timestamp               time.Time
}

// RTTAnalysis RTT 分析结果
type RTTAnalysis struct {
	Count        int
	Samples      []time.Duration
	AverageRTT   time.Duration
	MinRTT       time.Duration
	MaxRTT       time.Duration
	StdDeviation time.Duration
	NetworkType  string
}

// String 返回易读的分析结果
func (nbr *NetworkBehaviorResult) String() string {
	return fmt.Sprintf("NetworkBehavior[packets=%d, avgRTT=%v, seqPattern=%s, timing=%s, characteristics=%v]",
		nbr.TotalPackets,
		nbr.RTTAnalysis.AverageRTT,
		nbr.SequenceNumberPattern,
		nbr.TimingPattern,
		nbr.BehaviorCharacteristics,
	)
}
