package tcpip

import (
	"fmt"
	"time"
)

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
