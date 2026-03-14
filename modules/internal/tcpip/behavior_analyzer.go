package tcpip

import (
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
