package behavior

import (
	"math"
	"time"
)

// BehaviorSignal represents behavior signal analysis
type BehaviorSignal struct {
	// Signal type
	SignalType string

	// Signal value (between 0-1)
	Score float64

	// Signal description
	Description string

	// Detection time
	Timestamp time.Time

	// Risk level
	RiskLevel string // "safe", "low", "medium", "high", "critical"
}

// TemporalPattern represents temporal pattern
type TemporalPattern struct {
	// Request intervals (milliseconds)
	Intervals []int64

	// Average interval
	MeanInterval float64

	// Standard deviation
	StdDev float64

	// Minimum/maximum interval
	MinInterval int64
	MaxInterval int64

	// Regularity index (0-1, higher means more regular)
	RegularityIndex float64

	// Anomalous interval count
	AnomalousIntervals int
}

// ProtocolProportion represents protocol proportions
type ProtocolProportion struct {
	// TLS version distribution
	TLSVersions map[string]float64 // "1.2" -> 0.6, "1.3" -> 0.4

	// Cipher Suite distribution (top 5)
	TopCipherSuites map[string]float64

	// Extension distribution (top 10)
	TopExtensions map[string]float64

	// HTTP version distribution
	HTTPVersions map[string]float64 // "1.1" -> 0.3, "2" -> 0.7

	// ALPN protocol distribution
	ALPNProtocols map[string]float64

	// Entropy value (used to detect unnatural distributions)
	EntropyScore float64

	// Whether anomalous distribution exists
	IsAnomalous bool
}

// RequestBehavior represents single request behavior
type RequestBehavior struct {
	// Timestamp
	Timestamp time.Time

	// TLS version
	TLSVersion string

	// Cipher Suite
	CipherSuite string

	// Extensions used
	Extensions []string

	// HTTP version
	HTTPVersion string

	// requestsize
	RequestSize int

	// responsesize
	ResponseSize int

	// delay
	Latency time.Duration

	// Whether connection reuse was used
	ReusingConnection bool

	// Source address
	SourceIP string

	// Destination address
	DestinationIP string

	// Destination port
	DestinationPort int

	// SNI
	SNI string
}

// BehaviorAnalyzer is behavior signal analyzer
type BehaviorAnalyzer struct {
	// Request history
	requestHistory []RequestBehavior

	// Temporal pattern
	temporalPatterns map[string]*TemporalPattern

	// Protocol proportion
	protocolProportions map[string]*ProtocolProportion

	// Detected signals
	signals []BehaviorSignal

	// analyzeparameter
	config *BehaviorAnalysisConfig
}

// BehaviorAnalysisConfig analyzeconfiguration
type BehaviorAnalysisConfig struct {
	// Time window size (seconds)
	TimeWindowSize int

	// Minimum request count
	MinRequestsForAnalysis int

	// Regularity threshold (anomaly detection)
	RegularityThreshold float64

	// Entropy threshold
	EntropyThreshold float64

	// Anomalous interval rate threshold
	AnomalousIntervalRateThreshold float64
}

// NewBehaviorAnalyzer creates a behavior analyzer
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
		requestHistory:      make([]RequestBehavior, 0, 100), // Pre-allocated capacity
		temporalPatterns:    make(map[string]*TemporalPattern),
		protocolProportions: make(map[string]*ProtocolProportion),
		signals:             make([]BehaviorSignal, 0, 50), // Pre-allocated capacity
		config:              config,
	}
}

// AddRequest adds a request
func (ba *BehaviorAnalyzer) AddRequest(req RequestBehavior) {
	ba.requestHistory = append(ba.requestHistory, req)
}

// AnalyzeTemporalPattern analyzes temporal pattern
func (ba *BehaviorAnalyzer) AnalyzeTemporalPattern(origin string) *TemporalPattern {
	// Filter requests from this source and calculate intervals (single pass, reduce memory allocation)
	intervals := make([]int64, 0, len(ba.requestHistory))
	var lastReq *RequestBehavior

	for i := range ba.requestHistory {
		req := &ba.requestHistory[i]
		// Use SNI or destination address as source identifier
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

	// Calculate statistical features (zero-allocation calculation)
	pattern := ba.calculateTemporalStatsZeroAlloc(intervals)

	// Calculate regularity index
	ba.calculateRegularityIndex(pattern)

	// Detect anomalous intervals
	ba.detectAnomalousIntervals(pattern)

	ba.temporalPatterns[origin] = pattern

	return pattern
}

// calculateTemporalStats calculates temporal statistics
func (ba *BehaviorAnalyzer) calculateTemporalStats(pattern *TemporalPattern) {
	if len(pattern.Intervals) == 0 {
		return
	}

	// Calculate average value
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

	// Calculate standard deviation
	var variance float64
	for _, interval := range pattern.Intervals {
		diff := float64(interval) - pattern.MeanInterval
		variance += diff * diff
	}
	pattern.StdDev = math.Sqrt(variance / float64(len(pattern.Intervals)))
}

// calculateTemporalStatsZeroAlloc calculates temporal statistics (zero-allocation version)
func (ba *BehaviorAnalyzer) calculateTemporalStatsZeroAlloc(intervals []int64) *TemporalPattern {
	pattern := &TemporalPattern{
		Intervals: intervals,
	}

	if len(intervals) == 0 {
		return pattern
	}

	// Calculate average value, minimum value, maximum value (single pass)
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

	// Calculate standard deviation
	var variance float64
	for _, interval := range intervals {
		diff := float64(interval) - pattern.MeanInterval
		variance += diff * diff
	}
	pattern.StdDev = math.Sqrt(variance / float64(len(intervals)))

	return pattern
}

// calculateRegularityIndex calculates regularity index
// High regularity (close to 1) indicates very regular request intervals, possibly a bot
// Low regularity (close to 0) indicates random request intervals, possibly a real user
func (ba *BehaviorAnalyzer) calculateRegularityIndex(pattern *TemporalPattern) {
	if pattern.MeanInterval == 0 || pattern.StdDev == 0 {
		pattern.RegularityIndex = 0
		return
	}

	// Coefficient of variation (CV) = standard deviation / average value
	cv := pattern.StdDev / pattern.MeanInterval

	// Regularity index = 1 - CV, limited to [0, 1]
	pattern.RegularityIndex = 1.0 - cv
	if pattern.RegularityIndex < 0 {
		pattern.RegularityIndex = 0
	}
	if pattern.RegularityIndex > 1 {
		pattern.RegularityIndex = 1
	}
}

// detectAnomalousIntervals detects anomalous intervals
func (ba *BehaviorAnalyzer) detectAnomalousIntervals(pattern *TemporalPattern) {
	if pattern.StdDev == 0 {
		return
	}

	// Use 3-sigma rule to detect anomalous values
	threshold := pattern.MeanInterval + 3*pattern.StdDev

	for _, interval := range pattern.Intervals {
		if float64(interval) > threshold {
			pattern.AnomalousIntervals++
		}
	}
}

// AnalyzeProtocolProportion analyzes protocol proportions
