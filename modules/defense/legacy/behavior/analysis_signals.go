package behavior

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/vistone/fingerprint/modules/internal/metrics"
)

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
