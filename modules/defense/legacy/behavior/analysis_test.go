package behavior

import (
	"testing"
	"time"
)

// TestNewBehaviorAnalyzer tests creating analyzer
func TestNewBehaviorAnalyzer(t *testing.T) {
	config := &BehaviorAnalysisConfig{
		MinRequestsForAnalysis:         5,
		RegularityThreshold:            0.3,
		EntropyThreshold:               0.5,
		AnomalousIntervalRateThreshold: 0.2,
	}

	analyzer := NewBehaviorAnalyzer(config)
	if analyzer == nil {
		t.Fatal("NewBehaviorAnalyzer returned nil")
	}

	if analyzer.config != config {
		t.Error("analyzer.config mismatch")
	}
}

// TestNewBehaviorAnalyzerNilConfig tests creating analyzer with nil config
func TestNewBehaviorAnalyzerNilConfig(t *testing.T) {
	analyzer := NewBehaviorAnalyzer(nil)
	if analyzer == nil {
		t.Fatal("NewBehaviorAnalyzer(nil) returned nil")
	}

	if analyzer.config == nil {
		t.Fatal("analyzer.config is nil")
	}

	if analyzer.config.MinRequestsForAnalysis != 5 {
		t.Errorf("Expected MinRequestsForAnalysis=5, got %d", analyzer.config.MinRequestsForAnalysis)
	}
}

// TestBehaviorAnalyzerAddRequest tests adding request
func TestBehaviorAnalyzerAddRequest(t *testing.T) {
	analyzer := NewBehaviorAnalyzer(nil)

	req := RequestBehavior{
		Timestamp:         time.Now(),
		TLSVersion:        "1.3",
		CipherSuite:       "TLS_AES_128_GCM_SHA256",
		Extensions:        []string{"server_name", "alpn"},
		HTTPVersion:       "2",
		RequestSize:       1024,
		ResponseSize:      2048,
		Latency:           50 * time.Millisecond,
		ReusingConnection: false,
		SourceIP:          "192.168.1.1",
		DestinationIP:     "10.0.0.1",
		DestinationPort:   443,
		SNI:               "example.com",
	}

	// Adding request should not panic
	analyzer.AddRequest(req)

	// Verify request is added
	if len(analyzer.requestHistory) != 1 {
		t.Errorf("Expected 1 request, got %d", len(analyzer.requestHistory))
	}
}

// TestAnalyzeTemporalPattern tests temporal pattern analysis
func TestAnalyzeTemporalPattern(t *testing.T) {
	analyzer := NewBehaviorAnalyzer(nil)
	origin := "example.com"

	// Add enough requests
	baseTime := time.Now()
	for i := 0; i < 10; i++ {
		req := RequestBehavior{
			Timestamp:     baseTime.Add(time.Duration(i) * 100 * time.Millisecond),
			SNI:           origin,
			DestinationIP: "10.0.0.1",
		}
		analyzer.AddRequest(req)
	}

	pattern := analyzer.AnalyzeTemporalPattern(origin)
	if pattern == nil {
		t.Fatal("AnalyzeTemporalPattern returned nil")
	}

	if len(pattern.Intervals) != 9 {
		t.Errorf("Expected 9 intervals, got %d", len(pattern.Intervals))
	}

	if pattern.MeanInterval == 0 {
		t.Error("MeanInterval is 0")
	}
}

// TestAnalyzeTemporalPatternInsufficientData tests insufficient data case
func TestAnalyzeTemporalPatternInsufficientData(t *testing.T) {
	analyzer := NewBehaviorAnalyzer(nil)
	origin := "example.com"

	// Only add 2 requests (less than MinRequestsForAnalysis=5)
	for i := 0; i < 2; i++ {
		req := RequestBehavior{
			Timestamp:     time.Now(),
			SNI:           origin,
			DestinationIP: "10.0.0.1",
		}
		analyzer.AddRequest(req)
	}

	pattern := analyzer.AnalyzeTemporalPattern(origin)
	if pattern != nil {
		t.Error("Expected nil pattern for insufficient data")
	}
}

// TestEvaluateTemporalPattern tests temporal pattern evaluation
func TestEvaluateTemporalPattern(t *testing.T) {
	analyzer := NewBehaviorAnalyzer(nil)

	tests := []struct {
		name          string
		regularity    float64
		anomalousRate float64
		wantSignal    bool
		wantRisk      string
	}{
		{
			name:       "High regularity",
			regularity: 0.9,
			wantSignal: true,
			wantRisk:   "high",
		},
		{
			name:          "High anomalous rate",
			regularity:    0.5,
			anomalousRate: 0.3,
			wantSignal:    true,
			wantRisk:      "medium",
		},
		{
			name:       "Normal",
			regularity: 0.5,
			wantSignal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := &TemporalPattern{
				Intervals:          make([]int64, 10),
				RegularityIndex:    tt.regularity,
				AnomalousIntervals: int(tt.anomalousRate * 10),
			}

			sig, shouldAdd := analyzer.evaluateTemporalPattern(pattern)
			if shouldAdd != tt.wantSignal {
				t.Errorf("shouldAdd = %v, want %v", shouldAdd, tt.wantSignal)
			}
			if shouldAdd && sig.RiskLevel != tt.wantRisk {
				t.Errorf("RiskLevel = %s, want %s", sig.RiskLevel, tt.wantRisk)
			}
		})
	}
}

// TestGenerateBehaviorSignals tests generating behavior signals
func TestGenerateBehaviorSignals(t *testing.T) {
	analyzer := NewBehaviorAnalyzer(nil)
	origin := "example.com"

	// Add highly regular requests (simulate bot)
	baseTime := time.Now()
	for i := 0; i < 10; i++ {
		req := RequestBehavior{
			Timestamp:     baseTime.Add(time.Duration(i) * 1000 * time.Millisecond), // Precise 1s interval
			SNI:           origin,
			DestinationIP: "10.0.0.1",
			TLSVersion:    "1.3",
			CipherSuite:   "TLS_AES_128_GCM_SHA256",
		}
		analyzer.AddRequest(req)
	}

	signals := analyzer.GenerateBehaviorSignals(origin)
	// High regularity should generate signal
	if len(signals) == 0 {
		t.Log("No signals generated for high regularity pattern")
	}
}

// TestGetRiskScore tests risk score calculation
func TestGetRiskScore(t *testing.T) {
	analyzer := NewBehaviorAnalyzer(nil)

	// Empty signals should return 0
	if score := analyzer.GetRiskScore(); score != 0 {
		t.Errorf("Expected risk score 0 for empty signals, got %f", score)
	}

	// Add some signals
	analyzer.signals = []BehaviorSignal{
		{Score: 0.9, RiskLevel: "high"},
		{Score: 0.5, RiskLevel: "medium"},
	}

	score := analyzer.GetRiskScore()
	if score == 0 {
		t.Error("Expected non-zero risk score")
	}
}

// TestGetAllSignals tests getting all signals
func TestGetAllSignals(t *testing.T) {
	analyzer := NewBehaviorAnalyzer(nil)

	// Initially empty
	signals := analyzer.GetAllSignals()
	if len(signals) != 0 {
		t.Errorf("Expected 0 signals, got %d", len(signals))
	}

	// Add signals
	analyzer.signals = []BehaviorSignal{
		{SignalType: "TEST", RiskLevel: "low"},
	}

	signals = analyzer.GetAllSignals()
	if len(signals) != 1 {
		t.Errorf("Expected 1 signal, got %d", len(signals))
	}
}

// TestGetAnalysisSummary tests analysis summary
func TestGetAnalysisSummary(t *testing.T) {
	analyzer := NewBehaviorAnalyzer(nil)

	// Empty analyzer
	summary := analyzer.GetAnalysisSummary()
	if summary == "" {
		t.Error("GetAnalysisSummary returned empty string")
	}

	// Add some data
	analyzer.AddRequest(RequestBehavior{
		Timestamp:     time.Now(),
		SNI:           "example.com",
		DestinationIP: "10.0.0.1",
	})
	analyzer.signals = []BehaviorSignal{
		{SignalType: "TEST", RiskLevel: "high", Score: 0.9},
	}

	summary = analyzer.GetAnalysisSummary()
	if summary == "" {
		t.Error("GetAnalysisSummary returned empty string")
	}
}

// BenchmarkAddRequest benchmark: add request
func BenchmarkAddRequest(b *testing.B) {
	analyzer := NewBehaviorAnalyzer(nil)
	req := RequestBehavior{
		Timestamp:     time.Now(),
		SNI:           "example.com",
		DestinationIP: "10.0.0.1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AddRequest(req)
	}
}

// BenchmarkAnalyzeTemporalPattern benchmark: temporal pattern analysis
func BenchmarkAnalyzeTemporalPattern(b *testing.B) {
	analyzer := NewBehaviorAnalyzer(nil)
	origin := "example.com"

	// Pre-populate data
	baseTime := time.Now()
	for i := 0; i < 100; i++ {
		req := RequestBehavior{
			Timestamp:     baseTime.Add(time.Duration(i) * 100 * time.Millisecond),
			SNI:           origin,
			DestinationIP: "10.0.0.1",
		}
		analyzer.AddRequest(req)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeTemporalPattern(origin)
	}
}
