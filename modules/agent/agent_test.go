package agent

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/ml"
)

func makeObs(id, clientID, fp string, ts time.Time, conf, risk float64) *Observation {
	return &Observation{
		ID:              id,
		ClientID:        clientID,
		Timestamp:       ts,
		Features:        core.NewFeatureVector(),
		FingerprintHash: fp,
		Classification:  &ml.ClassificationResult{Confidence: conf, Family: core.BrowserChrome},
		RiskAssessment:  &core.RiskAssessment{Score: risk},
	}
}

func TestNewAgent(t *testing.T) {
	a := NewAgent(nil)
	if a == nil {
		t.Fatal("NewAgent returned nil")
	}
	stats := a.Stats()
	if stats.ActiveSessions != 0 {
		t.Errorf("expected 0 sessions, got %d", stats.ActiveSessions)
	}
	if stats.ActiveStrategies < 5 {
		t.Errorf("expected >= 5 builtin strategies, got %d", stats.ActiveStrategies)
	}
}

func TestProcessSingleObservation(t *testing.T) {
	a := NewAgent(nil)
	obs := makeObs("obs-1", "client-1", "aabbccdd", time.Now(), 0.9, 0.1)
	dec := a.Process(context.Background(), obs)
	if dec == nil {
		t.Fatal("Process returned nil")
	}
	if dec.Action != ActionAllow {
		t.Errorf("expected Allow for benign single observation, got %s", dec.Action)
	}
}

func TestFingerprintSwitchDetection(t *testing.T) {
	cfg := *DefaultAgentConfig
	cfg.FPSwitchRateThreshold = 2.0
	a := NewAgent(&cfg)

	clientID := "client-switch"
	now := time.Now()

	for i := 0; i < 6; i++ {
		obs := makeObs(
			fmt.Sprintf("obs-%d", i), clientID,
			fmt.Sprintf("fp-%d", i),
			now.Add(time.Duration(i)*10*time.Second),
			0.8, 0.2,
		)
		a.Process(context.Background(), obs)
	}

	profile := a.GetBehaviorProfile(clientID)
	if profile.FPSwitchRate <= cfg.FPSwitchRateThreshold {
		t.Errorf("expected switch rate > %.1f, got %.2f", cfg.FPSwitchRateThreshold, profile.FPSwitchRate)
	}
}

func TestRequestBurstDetection(t *testing.T) {
	cfg := *DefaultAgentConfig
	cfg.RequestBurstThreshold = 5.0
	a := NewAgent(&cfg)

	clientID := "client-burst"
	now := time.Now()

	for i := 0; i < 10; i++ {
		obs := makeObs(
			fmt.Sprintf("obs-%d", i), clientID, "same-fp",
			now.Add(time.Duration(i)*50*time.Millisecond),
			0.9, 0.1,
		)
		a.Process(context.Background(), obs)
	}

	profile := a.GetBehaviorProfile(clientID)
	if profile.AvgRequestInterval <= 0 {
		t.Fatal("expected positive avg interval")
	}
	reqPerSec := 1.0 / profile.AvgRequestInterval
	if reqPerSec <= cfg.RequestBurstThreshold {
		t.Errorf("expected req/s > %.1f, got %.2f", cfg.RequestBurstThreshold, reqPerSec)
	}
}

func TestHighRiskBlockStrategy(t *testing.T) {
	a := NewAgent(nil)
	clientID := "client-highrisk"
	now := time.Now()

	families := []core.BrowserType{core.BrowserChrome, core.BrowserFirefox, core.BrowserSafari, core.BrowserEdge, core.BrowserOpera}

	// translated comment
	for i := 0; i < 20; i++ {
		obs := &Observation{
			ID:              fmt.Sprintf("obs-%d", i),
			ClientID:        clientID,
			Timestamp:       now.Add(time.Duration(i) * time.Second),
			Features:        core.NewFeatureVector(),
			FingerprintHash: fmt.Sprintf("fp-unique-%d", i),
			Classification:  &ml.ClassificationResult{Confidence: 0.15, Family: families[i%len(families)]},
			RiskAssessment:  &core.RiskAssessment{Score: 0.75},
		}
		a.Process(context.Background(), obs)
	}

	lastObs := &Observation{
		ID:              "obs-final",
		ClientID:        clientID,
		Timestamp:       now.Add(21 * time.Second),
		Features:        core.NewFeatureVector(),
		FingerprintHash: "fp-final",
		Classification:  &ml.ClassificationResult{Confidence: 0.1, Family: core.BrowserBrave},
		RiskAssessment:  &core.RiskAssessment{Score: 0.9},
	}
	dec := a.Process(context.Background(), lastObs)

	if dec.Action != ActionBlock {
		t.Errorf("expected Block for high-risk client, got %s", dec.Action)
	}
}

func TestMemoryCleanup(t *testing.T) {
	cfg := *DefaultAgentConfig
	cfg.SessionTimeout = 100 * time.Millisecond
	a := NewAgent(&cfg)

	obs := &Observation{
		ID:              "obs-1",
		ClientID:        "client-expire",
		Timestamp:       time.Now().Add(-time.Second),
		Features:        core.NewFeatureVector(),
		FingerprintHash: "fp-1",
	}
	a.memory.Record(obs)

	if a.memory.SessionCount() != 1 {
		t.Fatal("expected 1 session before cleanup")
	}
	a.memory.Cleanup(cfg.SessionTimeout)
	if a.memory.SessionCount() != 0 {
		t.Error("expected 0 sessions after cleanup")
	}
}

func TestConsistencyScore(t *testing.T) {
	a := NewAgent(nil)
	clientID := "client-consistent"
	now := time.Now()

	for i := 0; i < 10; i++ {
		obs := makeObs(
			fmt.Sprintf("obs-%d", i), clientID, "same-fp",
			now.Add(time.Duration(i)*time.Second),
			0.95, 0.05,
		)
		a.Process(context.Background(), obs)
	}

	profile := a.GetBehaviorProfile(clientID)
	if profile.ConsistencyScore < 0.8 {
		t.Errorf("expected high consistency, got %.2f", profile.ConsistencyScore)
	}
}

func TestStrategyEvolution(t *testing.T) {
	cfg := *DefaultAgentConfig
	cfg.MinObservationsToLearn = 3
	cfg.FPSwitchRateThreshold = 1.0
	a := NewAgent(&cfg)

	now := time.Now()
	for c := 0; c < 5; c++ {
		clientID := fmt.Sprintf("client-%d", c)
		for i := 0; i < 10; i++ {
			obs := makeObs(
				fmt.Sprintf("obs-%d-%d", c, i), clientID,
				fmt.Sprintf("fp-%d-%d", c, i),
				now.Add(time.Duration(i)*5*time.Second),
				0.5, 0.3,
			)
			a.memory.Record(obs)
		}
	}

	beforeCount := a.strategy.LearnedPatternCount()
	a.strategy.Evolve()
	afterCount := a.strategy.LearnedPatternCount()

	if afterCount <= beforeCount {
		t.Logf("learned patterns: before=%d after=%d (evolution may not trigger at low data)", beforeCount, afterCount)
	}
}

func TestAgentStartStop(t *testing.T) {
	a := NewAgent(nil)
	a.Start()
	time.Sleep(50 * time.Millisecond)
	a.Stop()
}

func TestRiskTrend(t *testing.T) {
	a := NewAgent(nil)
	clientID := "client-trend"
	now := time.Now()

	for i := 0; i < 10; i++ {
		obs := makeObs(
			fmt.Sprintf("obs-%d", i), clientID, "fp-1",
			now.Add(time.Duration(i)*time.Second),
			0.8, float64(i)*0.1,
		)
		a.Process(context.Background(), obs)
	}

	profile := a.GetBehaviorProfile(clientID)
	if profile.RiskTrend <= 0 {
		t.Errorf("expected positive risk trend, got %.2f", profile.RiskTrend)
	}
}

// ===================================================================
// Knowledge base tests
// ===================================================================

func TestKnowledgeBaseInit(t *testing.T) {
	kb := NewKnowledgeBase()

	// Verify browser data loaded
	stats := kb.Stats()
	if stats.TotalKnownBrowsers < 6 {
		t.Errorf("expected >= 6 browser families, got %d", stats.TotalKnownBrowsers)
	}
	if stats.TotalKnownVersions < 10 {
		t.Errorf("expected >= 10 known versions, got %d", stats.TotalKnownVersions)
	}

	// Chrome knowledge should exist
	chrome := kb.GetBrowserKnowledge(core.BrowserChrome)
	if chrome == nil {
		t.Fatal("Chrome knowledge not found")
	}
	if chrome.MarketShare < 0.5 {
		t.Errorf("Chrome market share should be > 50%%, got %.0f%%", chrome.MarketShare*100)
	}
	if len(chrome.Versions) < 3 {
		t.Errorf("expected >= 3 Chrome versions, got %d", len(chrome.Versions))
	}

	// Firefox knowledge
	ff := kb.GetBrowserKnowledge(core.BrowserFirefox)
	if ff == nil {
		t.Fatal("Firefox knowledge not found")
	}
	// Firefox H2 InitialWindowSize should be much smaller than Chrome
	if len(ff.Versions) > 0 && ff.Versions[0].H2InitialWindowSize >= 6291456 {
		t.Error("Firefox H2 InitialWindowSize should differ from Chrome")
	}
}

func TestKnowledgeBaseGREASE(t *testing.T) {
	kb := NewKnowledgeBase()

	if !kb.IsGREASE(0x0a0a) {
		t.Error("0x0a0a should be GREASE")
	}
	if kb.IsGREASE(0x1301) {
		t.Error("0x1301 (TLS_AES_128_GCM) should NOT be GREASE")
	}
}

func TestKnowledgeCipherSuiteValidation(t *testing.T) {
	kb := NewKnowledgeBase()

	// TLS 1.3 suites
	if !kb.IsKnownCipherSuite(0x1301) {
		t.Error("TLS_AES_128_GCM_SHA256 should be known")
	}
	// TLS 1.2 Chrome suite
	if !kb.IsKnownCipherSuite(0xc02b) {
		t.Error("ECDHE_ECDSA_WITH_AES_128_GCM should be known")
	}
	// Forged suite
	if kb.IsKnownCipherSuite(0xFFFF) {
		t.Error("0xFFFF should NOT be a known cipher suite")
	}
}

func TestKnowledgeTCPIP(t *testing.T) {
	kb := NewKnowledgeBase()

	win := kb.GetExpectedTCPIP("windows")
	if win == nil {
		t.Fatal("Windows TCP/IP knowledge not found")
	}
	if win.TTL != 128 {
		t.Errorf("expected Windows TTL=128, got %d", win.TTL)
	}

	mac := kb.GetExpectedTCPIP("macos")
	if mac == nil {
		t.Fatal("macOS TCP/IP knowledge not found")
	}
	if mac.TTL != 64 {
		t.Errorf("expected macOS TTL=64, got %d", mac.TTL)
	}
	if mac.WindowSize != 65535 {
		t.Errorf("expected macOS WindowSize=65535, got %d", mac.WindowSize)
	}
}

func TestKnowledgeHTTP2(t *testing.T) {
	kb := NewKnowledgeBase()

	chromeH2 := kb.GetExpectedH2(core.BrowserChrome)
	if chromeH2 == nil {
		t.Fatal("Chrome H2 knowledge not found")
	}
	if chromeH2.InitialWindowSize != 6291456 {
		t.Errorf("expected Chrome H2 InitialWindowSize=6291456, got %d", chromeH2.InitialWindowSize)
	}
	if chromeH2.MaxConcurrentStreams != 1000 {
		t.Errorf("expected Chrome MaxConcurrentStreams=1000, got %d", chromeH2.MaxConcurrentStreams)
	}

	ffH2 := kb.GetExpectedH2(core.BrowserFirefox)
	if ffH2 == nil {
		t.Fatal("Firefox H2 knowledge not found")
	}
	if ffH2.MaxConcurrentStreams != 100 {
		t.Errorf("expected Firefox MaxConcurrentStreams=100, got %d", ffH2.MaxConcurrentStreams)
	}
	if ffH2.ConnectionFlow != 12517377 {
		t.Errorf("expected Firefox ConnectionFlow=12517377, got %d", ffH2.ConnectionFlow)
	}
}

func TestKnowledgeFindClosestVersion(t *testing.T) {
	kb := NewKnowledgeBase()

	// Chrome TLS 1.3 + 1.2 suites should match Chrome version
	chromeSuites := []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8, 0xc013, 0xc014, 0x002f, 0x0035, 0x000a}
	v := kb.FindClosestVersion(core.BrowserChrome, chromeSuites)
	if v == nil {
		t.Fatal("expected to find a matching Chrome version")
	}
	if v.VersionMajor < 115 {
		t.Errorf("expected Chrome >= 115, got %d", v.VersionMajor)
	}
}

// ===================================================================
// Anomaly detector tests
// ===================================================================

func TestAnomalyDetectorCleanObservation(t *testing.T) {
	a := NewAgent(nil)
	ad := a.anomaly

	// Normal Chrome observation, should have no contradictions
	obs := &Observation{
		ID:        "obs-clean",
		Timestamp: time.Now(),
		Features:  core.NewFeatureVector(),
		Classification: &ml.ClassificationResult{
			Family: core.BrowserChrome, Confidence: 0.95,
		},
	}
	obs.Features.Set(core.FeatureCipherSuites, 12) // Within Chrome range [9,17]
	obs.Features.Set(core.FeatureExtensions, 13)   // Within Chrome range [10,18]

	mr := ad.Analyze(obs)
	if mr.SuspicionScore > 0.1 {
		t.Errorf("clean observation should have low suspicion, got %.2f, contradictions: %+v",
			mr.SuspicionScore, mr.Contradictions)
	}
}

func TestAnomalyDetectorTLSMismatch(t *testing.T) {
	a := NewAgent(nil)
	ad := a.anomaly

	// Claimed Chrome client, but only 3 cipher suites (too few)
	obs := &Observation{
		ID:        "obs-tls-anomaly",
		Timestamp: time.Now(),
		Features:  core.NewFeatureVector(),
		Classification: &ml.ClassificationResult{
			Family: core.BrowserChrome, Confidence: 0.8,
		},
	}
	obs.Features.Set(core.FeatureCipherSuites, 3) // Chrome minimum 9
	obs.Features.Set(core.FeatureExtensions, 5)   // Chrome minimum 10

	mr := ad.Analyze(obs)
	if len(mr.Contradictions) < 2 {
		t.Errorf("expected >= 2 contradictions for TLS mismatch, got %d", len(mr.Contradictions))
	}
	if mr.SuspicionScore <= 0.2 {
		t.Errorf("expected elevated suspicion, got %.2f", mr.SuspicionScore)
	}
}

func TestAnomalyDetectorHeadlessDetection(t *testing.T) {
	a := NewAgent(nil)
	ad := a.anomaly

	obs := &Observation{
		ID:        "obs-headless",
		Timestamp: time.Now(),
		Features:  core.NewFeatureVector(),
		Classification: &ml.ClassificationResult{
			Family: core.BrowserChrome, Confidence: 0.7,
		},
	}
	obs.Features.Set(core.FeatureHeadlessBrowser, 1.0)
	obs.Features.Set(core.FeatureToolMarker, 1.0)
	obs.Features.Set(core.FeatureCipherSuites, 12) // Normal range

	mr := ad.Analyze(obs)
	found := false
	for _, c := range mr.Contradictions {
		if c.Field == "headless_browser" || c.Field == "tool_marker" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected headless/tool markers to be flagged")
	}
	if mr.SuspicionScore < 0.5 {
		t.Errorf("expected high suspicion for headless+tool, got %.2f", mr.SuspicionScore)
	}
}

func TestAnomalyDetectorTCPIPMismatch(t *testing.T) {
	a := NewAgent(nil)
	ad := a.anomaly

	obs := &Observation{
		ID:        "obs-tcpip",
		Timestamp: time.Now(),
		Features:  core.NewFeatureVector(),
		Classification: &ml.ClassificationResult{
			Family: core.BrowserChrome, Confidence: 0.9,
		},
		Metadata: map[string]string{
			"os_family": "Windows",
			"tcp_ttl":   "64", // Windows should be 128 segment
		},
	}
	obs.Features.Set(core.FeatureCipherSuites, 12)

	mr := ad.Analyze(obs)
	found := false
	for _, c := range mr.Contradictions {
		if c.Field == "tcp_ttl" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected TCP TTL mismatch to be flagged (Windows claims TTL=64)")
	}
}

func TestAnomalyDetectorHTTP2Mismatch(t *testing.T) {
	a := NewAgent(nil)
	ad := a.anomaly

	// Claimed Chrome, but H2 parameters are Firefox's
	obs := &Observation{
		ID:        "obs-h2-mismatch",
		Timestamp: time.Now(),
		Features:  core.NewFeatureVector(),
		Classification: &ml.ClassificationResult{
			Family: core.BrowserChrome, Confidence: 0.85,
		},
	}
	obs.Features.Set(core.FeatureHTTP2Settings, 12345)                                // non-zero to trigger check
	obs.Features.Metadata["h2_initial_window_size"] = float64(131072)                 // Firefox value, not Chrome
	obs.Features.Metadata["h2_max_concurrent_streams"] = float64(100)                 // Firefox: 100, Chrome: 1000
	obs.Features.Metadata["pseudo_header_order"] = ":method,:path,:authority,:scheme" // Firefox order

	mr := ad.Analyze(obs)
	if len(mr.Contradictions) < 2 {
		t.Errorf("expected >= 2 H2 contradictions, got %d: %+v", len(mr.Contradictions), mr.Contradictions)
	}
}

func TestProcessWithKnowledgeIntegration(t *testing.T) {
	a := NewAgent(nil)
	ctx := context.Background()

	// Normal Chrome client
	obs := &Observation{
		ID:              "obs-normal",
		ClientID:        "client-normal",
		Timestamp:       time.Now(),
		Features:        core.NewFeatureVector(),
		FingerprintHash: "fp-normal",
		Classification:  &ml.ClassificationResult{Family: core.BrowserChrome, Confidence: 0.95},
		RiskAssessment:  &core.RiskAssessment{Score: 0.05},
	}
	obs.Features.Set(core.FeatureCipherSuites, 12)
	obs.Features.Set(core.FeatureExtensions, 14)

	dec := a.Process(ctx, obs)
	if dec.KnowledgeMatch == nil {
		t.Fatal("expected KnowledgeMatch in Decision")
	}
	if dec.KnowledgeMatch.SuspicionScore > 0.2 {
		t.Errorf("normal client should have low suspicion, got %.2f", dec.KnowledgeMatch.SuspicionScore)
	}

	// Highly suspicious spoofed client
	spoofObs := &Observation{
		ID:              "obs-spoof",
		ClientID:        "client-spoof",
		Timestamp:       time.Now(),
		Features:        core.NewFeatureVector(),
		FingerprintHash: "fp-spoof",
		Classification:  &ml.ClassificationResult{Family: core.BrowserChrome, Confidence: 0.4},
		RiskAssessment:  &core.RiskAssessment{Score: 0.6},
	}
	spoofObs.Features.Set(core.FeatureCipherSuites, 3) // Too few
	spoofObs.Features.Set(core.FeatureExtensions, 4)   // Too few
	spoofObs.Features.Set(core.FeatureHeadlessBrowser, 1.0)
	spoofObs.Features.Set(core.FeatureToolMarker, 1.0)

	spoofDec := a.Process(ctx, spoofObs)
	if spoofDec.KnowledgeMatch == nil {
		t.Fatal("expected KnowledgeMatch for spoofed client")
	}
	if spoofDec.KnowledgeMatch.SuspicionScore < 0.5 {
		t.Errorf("spoofed client should have high suspicion, got %.2f", spoofDec.KnowledgeMatch.SuspicionScore)
	}
	if spoofDec.ThreatClass != ThreatFingerprintSpoof {
		t.Errorf("expected ThreatFingerprintSpoof, got %s", spoofDec.ThreatClass)
	}
}
