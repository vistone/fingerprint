package tcpip

import (
	"fmt"
	"time"
)

// NetworkBehaviorAnalyzer is a network behavior analyzer
type NetworkBehaviorAnalyzer struct {
	packets         []*TCPPacket
	rttMeasurements []time.Duration
	sequenceNumbers []uint32
	ipIDs           []uint16
	timestamps      []time.Time
	maxSamples      int // max sample count to prevent unbounded memory growth
}

// DefaultMaxSamples is the default maximum sample count
const DefaultMaxSamples = 10000

// NewNetworkBehaviorAnalyzer creates a new network behavior analyzer
func NewNetworkBehaviorAnalyzer() *NetworkBehaviorAnalyzer {
	return NewNetworkBehaviorAnalyzerWithLimit(DefaultMaxSamples)
}

// NewNetworkBehaviorAnalyzerWithLimit creates a network behavior analyzer with a sample count limit
func NewNetworkBehaviorAnalyzerWithLimit(maxSamples int) *NetworkBehaviorAnalyzer {
	if maxSamples <= 0 {
		maxSamples = DefaultMaxSamples
	}
	return &NetworkBehaviorAnalyzer{
		packets:         make([]*TCPPacket, 0, 1024),
		rttMeasurements: make([]time.Duration, 0, 1024),
		sequenceNumbers: make([]uint32, 0, 1024),
		ipIDs:           make([]uint16, 0, 1024),
		timestamps:      make([]time.Time, 0, 1024),
		maxSamples:      maxSamples,
	}
}

// RecordPacket records a packet (uses sliding window to limit sample count)
func (nba *NetworkBehaviorAnalyzer) RecordPacket(packet *TCPPacket, rtt time.Duration) {
	nba.packets = appendWithLimit(nba.packets, packet, nba.maxSamples)
	nba.rttMeasurements = appendWithLimit(nba.rttMeasurements, rtt, nba.maxSamples)
	nba.sequenceNumbers = appendWithLimit(nba.sequenceNumbers, packet.SequenceNumber, nba.maxSamples)
	if packet.IPHeader != nil {
		nba.ipIDs = appendWithLimit(nba.ipIDs, packet.IPHeader.Identification, nba.maxSamples)
	}
	nba.timestamps = appendWithLimit(nba.timestamps, time.Now(), nba.maxSamples)
}

// appendWithLimit appends to a slice with a limit (sliding window, using generics)
func appendWithLimit[T any](slice []T, item T, maxSamples int) []T {
	if len(slice) >= maxSamples {
		// Sliding window: remove the oldest 25% of data
		removeCount := maxSamples / 4
		if removeCount < 1 {
			removeCount = 1
		}
		return append(slice[removeCount:], item)
	}
	return append(slice, item)
}

// AnalyzeBehavior analyzes network behavior
func (nba *NetworkBehaviorAnalyzer) AnalyzeBehavior() *NetworkBehaviorResult {
	result := &NetworkBehaviorResult{
		TotalPackets: len(nba.packets),
		Timestamp:    time.Now(),
	}

	if len(nba.packets) == 0 {
		return result
	}

	// Analyze RTT
	result.RTTAnalysis = nba.analyzeRTT()

	// Analyze sequence numbers
	result.SequenceNumberPattern = nba.analyzeSequenceNumbers()

	// Analyze IP ID
	result.IPIDPattern = nba.analyzeIPIDs()

	// Analyze packet sizes
	result.PacketSizeVariance = nba.analyzePacketSizes()

	// Analyze protocol distribution
	result.ProtocolDistribution = nba.analyzeProtocolDistribution()

	// Analyze timing pattern
	result.TimingPattern = nba.analyzeTimingPattern()

	// Compute behavior characteristics
	result.BehaviorCharacteristics = nba.computeBehaviorCharacteristics()

	return result
}

// analyzeRTT analyzes round-trip time
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

// analyzeSequenceNumbers analyzes sequence number patterns
func (nba *NetworkBehaviorAnalyzer) analyzeSequenceNumbers() string {
	if len(nba.sequenceNumbers) < 2 {
		return "insufficient_data"
	}

	// Check sequence number increment pattern
	diffs := make([]int64, len(nba.sequenceNumbers)-1)
	for i := 0; i < len(diffs); i++ {
		diffs[i] = int64(nba.sequenceNumbers[i+1]) - int64(nba.sequenceNumbers[i])
	}

	// Check if random
	if nba.hasHighVariance(diffs) {
		return "random"
	}

	// Check if time-related
	if nba.isTimeRelated(diffs) {
		return "time_based"
	}

	// Check if linear/sequential
	if nba.isLinear(diffs) {
		return "sequential"
	}

	return "complex_pattern"
}

// analyzeIPIDs analyzes IP ID patterns
func (nba *NetworkBehaviorAnalyzer) analyzeIPIDs() string {
	if len(nba.ipIDs) < 2 {
		return "insufficient_data"
	}

	// Calculate IP ID differences
	diffs := make([]int, len(nba.ipIDs)-1)
	for i := 0; i < len(diffs); i++ {
		diffs[i] = int(nba.ipIDs[i+1]) - int(nba.ipIDs[i])
	}

	// Check for linear counter (typical behavior of NATing devices)
	if nba.isLinearIPID(diffs) {
		return "linear_counter"
	}

	// Check if random
	if nba.hasHighVariance(convertToInt64(diffs)) {
		return "random"
	}

	return "mixed_pattern"
}

// analyzePacketSizes analyzes packet sizes
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

// analyzeProtocolDistribution analyzes protocol distribution
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

// analyzeTimingPattern analyzes timing patterns
func (nba *NetworkBehaviorAnalyzer) analyzeTimingPattern() string {
	if len(nba.timestamps) < 2 {
		return "insufficient_data"
	}

	intervals := make([]time.Duration, len(nba.timestamps)-1)
	for i := 0; i < len(intervals); i++ {
		intervals[i] = nba.timestamps[i+1].Sub(nba.timestamps[i])
	}

	// Check for regular intervals
	if nba.hasRegularIntervals(intervals) {
		return "periodic"
	}

	// Check for burst pattern
	if nba.isBurstPattern(intervals) {
		return "bursty"
	}

	return "irregular"
}

// computeBehaviorCharacteristics computes behavior characteristics
func (nba *NetworkBehaviorAnalyzer) computeBehaviorCharacteristics() []string {
	characteristics := make([]string, 0)

	if len(nba.packets) == 0 {
		return characteristics
	}

	// Check for automated traffic
	if nba.isAutomatedTraffic() {
		characteristics = append(characteristics, "automated")
	}

	// Check for interactive traffic
	if nba.isInteractiveTraffic() {
		characteristics = append(characteristics, "interactive")
	}

	// Check for bulk transfer
	if nba.isBulkTransfer() {
		characteristics = append(characteristics, "bulk_transfer")
	}

	// Check for scanning behavior
	if nba.isScanningBehavior() {
		characteristics = append(characteristics, "scanning")
	}

	return characteristics
}

// Helper functions
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
	// Return simplified standard deviation estimate
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
	return avg > 1000 // high variance threshold
}

func (nba *NetworkBehaviorAnalyzer) isTimeRelated(diffs []int64) bool {
	if len(diffs) < 2 {
		return false
	}

	// Time-related sequence numbers have a certain incremental pattern
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

	// Scanning behavior characteristic: connections fail quickly
	resetCount := 0
	for _, pkt := range nba.packets {
		if pkt.Flags&0x04 != 0 { // RST flag
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

// NetworkBehaviorResult represents network behavior analysis results
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

// RTTAnalysis represents RTT analysis results
type RTTAnalysis struct {
	Count        int
	Samples      []time.Duration
	AverageRTT   time.Duration
	MinRTT       time.Duration
	MaxRTT       time.Duration
	StdDeviation time.Duration
	NetworkType  string
}

// String returns a human-readable analysis result
func (nbr *NetworkBehaviorResult) String() string {
	return fmt.Sprintf("NetworkBehavior[packets=%d, avgRTT=%v, seqPattern=%s, timing=%s, characteristics=%v]",
		nbr.TotalPackets,
		nbr.RTTAnalysis.AverageRTT,
		nbr.SequenceNumberPattern,
		nbr.TimingPattern,
		nbr.BehaviorCharacteristics,
	)
}
