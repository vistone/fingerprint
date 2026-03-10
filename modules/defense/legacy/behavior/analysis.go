package behavior

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/vistone/fingerprint/modules/internal/metrics"
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

	// Count TLS versions
	for _, req := range originRequests {
		prop.TLSVersions[req.TLSVersion]++
		prop.HTTPVersions[req.HTTPVersion]++
		prop.TopCipherSuites[req.CipherSuite]++
	}

	// Convert to proportions
	for ver := range prop.TLSVersions {
		prop.TLSVersions[ver] /= float64(len(originRequests))
	}
	for ver := range prop.HTTPVersions {
		prop.HTTPVersions[ver] /= float64(len(originRequests))
	}
	for cs := range prop.TopCipherSuites {
		prop.TopCipherSuites[cs] /= float64(len(originRequests))
	}

	// Keep only top 5 cipher suites
	ba.keepTopN(prop.TopCipherSuites, 5)

	// statistics extensions
	extensionCount := make(map[string]int)
	for _, req := range originRequests {
		for _, ext := range req.Extensions {
			extensionCount[ext]++
		}
	}

	// Convert to proportions and keep top 10
	for ext, count := range extensionCount {
		prop.TopExtensions[ext] = float64(count) / float64(len(originRequests))
	}
	ba.keepTopN(prop.TopExtensions, 10)

	// Calculate entropy and anomaly detection
	ba.calculateProtocolEntropy(prop)

	ba.protocolProportions[origin] = prop

	return prop
}

// keepTopN keeps only the top N highest entries
func (ba *BehaviorAnalyzer) keepTopN(m map[string]float64, n int) {
	if len(m) <= n {
		return
	}

	// Create sorted list
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

	// Clear and keep only top N
	for k := range m {
		delete(m, k)
	}
	for i := 0; i < n && i < len(items); i++ {
		m[items[i].Key] = items[i].Value
	}
}

// calculateProtocolEntropy calculates entropy of protocol distribution
func (ba *BehaviorAnalyzer) calculateProtocolEntropy(prop *ProtocolProportion) {
	// Calculate entropy of TLS version distribution
	var entropy float64
	for _, prob := range prop.TLSVersions {
		if prob > 0 {
			entropy += -prob * math.Log2(prob)
		}
	}
	prop.EntropyScore = entropy

	// Detect anomaly: if all requests use the same TLS version or Cipher Suite, it's anomalous
	if len(prop.TLSVersions) == 1 || len(prop.TopCipherSuites) == 1 {
		prop.IsAnomalous = true
	}

	// Detect entropy anomaly
	// Normal real users will have diversified protocol versions
	if entropy < ba.config.EntropyThreshold && len(prop.TLSVersions) > 1 {
		prop.IsAnomalous = true
	}
}

// GenerateBehaviorSignals generates behavior signals
func (ba *BehaviorAnalyzer) GenerateBehaviorSignals(origin string) []BehaviorSignal {
	signals := []BehaviorSignal{}

	// Analyze temporal pattern
	if pattern := ba.AnalyzeTemporalPattern(origin); pattern != nil {
		sig, shouldAdd := ba.evaluateTemporalPattern(pattern)
		if shouldAdd {
			signals = append(signals, sig)
		}
	}

	// Analyze protocol proportions
	if proportion := ba.AnalyzeProtocolProportion(origin); proportion != nil {
		sig, shouldAdd := ba.evaluateProtocolProportion(proportion)
		if shouldAdd {
			signals = append(signals, sig)
		}
	}

	// Detect connection reuse behavior
	if sig, shouldAdd := ba.detectConnectionReuse(origin); shouldAdd {
		signals = append(signals, sig)
	}

	// Log metrics
	for _, sig := range signals {
		metrics.RecordBehaviorSignal(sig.RiskLevel)
	}

	ba.signals = append(ba.signals, signals...)
	return signals
}

// evaluateTemporalPattern evaluates temporal pattern
func (ba *BehaviorAnalyzer) evaluateTemporalPattern(pattern *TemporalPattern) (BehaviorSignal, bool) {
	var sig BehaviorSignal
	sig.SignalType = "TEMPORAL_PATTERN"
	sig.Timestamp = time.Now()

	anomalousRate := float64(pattern.AnomalousIntervals) / float64(len(pattern.Intervals))

	if pattern.RegularityIndex > 0.8 {
		// High regularity: possibly a bot or automated script
		sig.Score = pattern.RegularityIndex
		sig.Description = fmt.Sprintf("Highly regular request intervals: regularity index=%.2f, average interval=%.0fms",
			pattern.RegularityIndex, pattern.MeanInterval)
		sig.RiskLevel = "high"
		return sig, true
	}

	if anomalousRate > ba.config.AnomalousIntervalRateThreshold {
		// High proportion of anomalous intervals: unstable connection or anomalous behavior
		sig.Score = anomalousRate
		sig.Description = fmt.Sprintf("Anomalous request interval rate: %.1f%%, possible network issues or anomalous behavior",
			anomalousRate*100)
		sig.RiskLevel = "medium"
		return sig, true
	}

	return sig, false
}

// evaluateProtocolProportion evaluates protocol proportions
func (ba *BehaviorAnalyzer) evaluateProtocolProportion(prop *ProtocolProportion) (BehaviorSignal, bool) {
	var sig BehaviorSignal
	sig.SignalType = "PROTOCOL_PROPORTION"
	sig.Timestamp = time.Now()

	if prop.IsAnomalous {
		sig.Score = prop.EntropyScore
		sig.Description = fmt.Sprintf("Anomalous protocol distribution: entropy=%.2f, TLS versions=%d, CipherSuites=%d",
			prop.EntropyScore, len(prop.TLSVersions), len(prop.TopCipherSuites))
		sig.RiskLevel = "high"
		return sig, true
	}

	return sig, false
}

// detectConnectionReuse detects connection reuse behavior
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

	// If connection reuse rate is too high (close to 100%), possibly an automated script
	// Real users typically have some new connections
	if reuseRate > 0.9 {
		return BehaviorSignal{
			SignalType:  "CONNECTION_REUSE",
			Score:       reuseRate,
			Description: fmt.Sprintf("Extremely high connection reuse rate: %.1f%%", reuseRate*100),
			Timestamp:   time.Now(),
			RiskLevel:   "high",
		}, true
	}

	return BehaviorSignal{}, false
}

// GetAllSignals gets all signals
func (ba *BehaviorAnalyzer) GetAllSignals() []BehaviorSignal {
	return ba.signals
}

// GetRiskScore calculates overall risk score
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

// GetAnalysisSummary gets analysis summary
func (ba *BehaviorAnalyzer) GetAnalysisSummary() string {
	if len(ba.requestHistory) == 0 {
		return "No request data collected"
	}

	summary := fmt.Sprintf("Behavior analysis report\n")
	summary += fmt.Sprintf("=== Basic Info ===\n")
	summary += fmt.Sprintf("Total requests: %d\n", len(ba.requestHistory))
	summary += fmt.Sprintf("Detected signals: %d\n", len(ba.signals))
	summary += fmt.Sprintf("Overall risk score: %.2f\n\n", ba.GetRiskScore())

	if len(ba.signals) == 0 {
		summary += "No anomalous signals detected\n"
		return summary
	}

	summary += fmt.Sprintf("=== Detected Signals ===\n")
	for i, sig := range ba.signals {
		summary += fmt.Sprintf("%d. [%s] %s (Risk level: %s, Score: %.2f)\n",
			i+1, sig.SignalType, sig.Description, sig.RiskLevel, sig.Score)
	}

	return summary
}
