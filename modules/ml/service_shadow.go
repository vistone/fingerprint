package ml

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"
)

type shadowComparator struct {
	enabled     bool
	sampleRate  float64
	native      inferenceBackend
	onnx        inferenceBackend
	metricsPath string
	mu          sync.Mutex
	stats       ShadowCompareStats
}

// ShadowCompareStats tracks prediction parity between primary and shadow backends.
type ShadowCompareStats struct {
	SampledCount         int64   `json:"sampledCount"`
	ErrorCount           int64   `json:"errorCount"`
	BrowserTop1AgreeRate float64 `json:"browserTop1AgreeRate"`
	ActionTop1AgreeRate  float64 `json:"actionTop1AgreeRate"`
	AvgForgeryProbDelta  float64 `json:"avgForgeryProbDelta"`
	AvgThreatProbDelta   float64 `json:"avgThreatProbDelta"`
	LastError            string  `json:"lastError,omitempty"`
}

func newShadowComparator(config *ServiceConfig, pipeline *ModelPipeline) *shadowComparator {
	if config == nil || !config.ShadowCompareEnabled {
		return nil
	}

	rate := config.ShadowSampleRate
	if rate <= 0 {
		return nil
	}
	if rate > 1 {
		rate = 1
	}

	onnxBackend := newONNXInferenceBackend(config, pipeline)
	if onnxBackend == nil {
		return nil
	}

	return &shadowComparator{
		enabled:     true,
		sampleRate:  rate,
		native:      &nativeInferenceBackend{pipeline: pipeline},
		onnx:        onnxBackend,
		metricsPath: config.ShadowMetricsPath,
	}
}

func (s *MLService) runShadowCompare(call func(backend inferenceBackend) (*PipelineResult, error), primary *PipelineResult, primaryBackend inferenceBackend) {
	if s.shadow == nil || !s.shadow.shouldSample() {
		return
	}

	shadowBackend := s.shadow.shadowFor(primaryBackend)
	if shadowBackend == nil {
		return
	}

	shadowResult, err := call(shadowBackend)
	s.shadow.record(primaryBackend.Name(), shadowBackend.Name(), primary, shadowResult, err)
}

func (s *MLService) runShadowCompareBatch(call func(backend inferenceBackend) ([]*PipelineResult, error), primary []*PipelineResult, primaryBackend inferenceBackend) {
	if s.shadow == nil || !s.shadow.shouldSample() {
		return
	}

	shadowBackend := s.shadow.shadowFor(primaryBackend)
	if shadowBackend == nil {
		return
	}

	shadowResults, err := call(shadowBackend)
	if err != nil {
		s.shadow.record(primaryBackend.Name(), shadowBackend.Name(), nil, nil, err)
		return
	}

	limit := len(primary)
	if len(shadowResults) < limit {
		limit = len(shadowResults)
	}
	for i := 0; i < limit; i++ {
		s.shadow.record(primaryBackend.Name(), shadowBackend.Name(), primary[i], shadowResults[i], nil)
	}
}

func (c *shadowComparator) shouldSample() bool {
	if !c.enabled {
		return false
	}

	return rand.Float64() < c.sampleRate
}

func (c *shadowComparator) record(primaryName string, shadowName string, primary *PipelineResult, shadow *PipelineResult, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stats.SampledCount++
	entry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"primary":   primaryName,
		"shadow":    shadowName,
	}

	if err != nil {
		c.stats.ErrorCount++
		c.stats.LastError = err.Error()
		entry["error"] = err.Error()
		c.appendMetric(entry)
		return
	}
	if primary == nil || shadow == nil {
		c.stats.ErrorCount++
		c.stats.LastError = "empty shadow compare result"
		entry["error"] = c.stats.LastError
		c.appendMetric(entry)
		return
	}

	browserAgree := primary.Browser.Family == shadow.Browser.Family
	actionAgree := primary.Threat.Action == shadow.Threat.Action
	forgeryDelta := math.Abs(primary.Forgery.ForgeryProb - shadow.Forgery.ForgeryProb)
	threatDelta := math.Abs(primary.Threat.ThreatProb - shadow.Threat.ThreatProb)

	entry["browserAgree"] = browserAgree
	entry["actionAgree"] = actionAgree
	entry["forgeryDelta"] = forgeryDelta
	entry["threatDelta"] = threatDelta
	c.appendMetric(entry)

	seen := c.stats.SampledCount - c.stats.ErrorCount
	if seen <= 0 {
		return
	}

	prevCount := float64(seen - 1)
	if browserAgree {
		c.stats.BrowserTop1AgreeRate = ((c.stats.BrowserTop1AgreeRate * prevCount) + 1) / float64(seen)
	} else {
		c.stats.BrowserTop1AgreeRate = (c.stats.BrowserTop1AgreeRate * prevCount) / float64(seen)
	}
	if actionAgree {
		c.stats.ActionTop1AgreeRate = ((c.stats.ActionTop1AgreeRate * prevCount) + 1) / float64(seen)
	} else {
		c.stats.ActionTop1AgreeRate = (c.stats.ActionTop1AgreeRate * prevCount) / float64(seen)
	}

	c.stats.AvgForgeryProbDelta = ((c.stats.AvgForgeryProbDelta * prevCount) + forgeryDelta) / float64(seen)
	c.stats.AvgThreatProbDelta = ((c.stats.AvgThreatProbDelta * prevCount) + threatDelta) / float64(seen)
}

func (c *shadowComparator) shadowFor(primary inferenceBackend) inferenceBackend {
	if primary == nil {
		return nil
	}

	if primary.Name() == "onnx" {
		return c.native
	}
	return c.onnx
}

func (c *shadowComparator) appendMetric(entry map[string]interface{}) {
	if c.metricsPath == "" {
		return
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	file, err := os.OpenFile(c.metricsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	defer file.Close()

	_, _ = file.Write(append(data, '\n'))
}

func (c *shadowComparator) Snapshot() *ShadowCompareStats {
	if c == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	snapshot := c.stats
	return &snapshot
}
