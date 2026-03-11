// Package ml — learner.go provides an online learning system that
// continuously improves ML models from real-world feedback.
//
// The OnlineLearner collects feedback samples, detects accuracy drift,
// and triggers model evolution when quality degrades.
//
// Architecture:
//
//	Feedback samples → ring buffer → drift detector → evolve trigger
//	                                                ↓
//	                                         ProfileRegistry update
//	                                                ↓
//	                                         NeuralTrainer.Evolve()
package ml

import (
	"math"
	"sync"
	"time"
)

// =========================================================================
// OnlineLearnerConfig
// =========================================================================

// OnlineLearnerConfig configures the online learning system.
type OnlineLearnerConfig struct {
	// BufferSize is the max number of feedback samples retained.
	BufferSize int

	// DriftThreshold is the accuracy drop that triggers evolution.
	// E.g., 0.05 means 5% drop from peak accuracy triggers retraining.
	DriftThreshold float64

	// DriftWindowSize is how many recent samples to use for drift detection.
	DriftWindowSize int

	// MinSamplesForDrift is the minimum samples before drift can be detected.
	MinSamplesForDrift int
}

// DefaultOnlineLearnerConfig provides sensible defaults.
var DefaultOnlineLearnerConfig = &OnlineLearnerConfig{
	BufferSize:         1024,
	DriftThreshold:     0.05,
	DriftWindowSize:    100,
	MinSamplesForDrift: 50,
}

// =========================================================================
// OnlineLearner
// =========================================================================

// OnlineLearner manages online learning from real-world feedback.
type OnlineLearner struct {
	config  *OnlineLearnerConfig
	mu      sync.Mutex
	samples []FeedbackSample
	head    int // ring buffer write position
	count   int // total samples added (including overwrites)

	// Drift tracking
	peakAccuracy    float64
	recentAccuracy  float64
	driftDetected   bool
	lastDriftTime   time.Time
	driftEventCount int
}

// NewOnlineLearner creates a new online learning system.
func NewOnlineLearner(config *OnlineLearnerConfig) *OnlineLearner {
	if config == nil {
		config = DefaultOnlineLearnerConfig
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 1024
	}
	if config.DriftWindowSize <= 0 {
		config.DriftWindowSize = 100
	}
	if config.MinSamplesForDrift <= 0 {
		config.MinSamplesForDrift = 50
	}

	return &OnlineLearner{
		config:  config,
		samples: make([]FeedbackSample, config.BufferSize),
	}
}

// AddSample adds a feedback sample to the learning buffer.
func (l *OnlineLearner) AddSample(sample *FeedbackSample) {
	if sample == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if sample.Timestamp.IsZero() {
		sample.Timestamp = time.Now()
	}

	l.samples[l.head] = *sample
	l.head = (l.head + 1) % l.config.BufferSize
	l.count++

	l.updateDriftDetection()
}

// updateDriftDetection re-evaluates whether accuracy has drifted.
func (l *OnlineLearner) updateDriftDetection() {
	filled := l.count
	if filled > l.config.BufferSize {
		filled = l.config.BufferSize
	}
	if filled < l.config.MinSamplesForDrift {
		return
	}

	// Compute recent accuracy from rewards in the drift window.
	windowSize := l.config.DriftWindowSize
	if windowSize > filled {
		windowSize = filled
	}

	sum := 0.0
	for i := 0; i < windowSize; i++ {
		idx := (l.head - 1 - i + l.config.BufferSize) % l.config.BufferSize
		sum += l.samples[idx].Reward
	}
	l.recentAccuracy = sum / float64(windowSize)

	if l.recentAccuracy > l.peakAccuracy {
		l.peakAccuracy = l.recentAccuracy
	}

	// Detect drift: significant drop from peak.
	drop := l.peakAccuracy - l.recentAccuracy
	if drop > l.config.DriftThreshold {
		if !l.driftDetected {
			l.driftDetected = true
			l.lastDriftTime = time.Now()
			l.driftEventCount++
		}
	} else {
		l.driftDetected = false
	}
}

// DriftDetected returns whether accuracy drift has been detected.
func (l *OnlineLearner) DriftDetected() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.driftDetected
}

// RecentSamples returns the most recent n samples from the buffer.
func (l *OnlineLearner) RecentSamples(n int) []FeedbackSample {
	l.mu.Lock()
	defer l.mu.Unlock()

	filled := l.count
	if filled > l.config.BufferSize {
		filled = l.config.BufferSize
	}
	if n > filled {
		n = filled
	}

	result := make([]FeedbackSample, n)
	for i := 0; i < n; i++ {
		idx := (l.head - 1 - i + l.config.BufferSize) % l.config.BufferSize
		result[i] = l.samples[idx]
	}
	return result
}

// OnlineLearnerStats holds online learner statistics.
type OnlineLearnerStats struct {
	TotalSamples    int
	BufferFilled    int
	PeakAccuracy    float64
	RecentAccuracy  float64
	DriftDetected   bool
	DriftEventCount int
	LastDriftTime   time.Time
}

// Stats returns current learner statistics.
func (l *OnlineLearner) Stats() *OnlineLearnerStats {
	l.mu.Lock()
	defer l.mu.Unlock()

	filled := l.count
	if filled > l.config.BufferSize {
		filled = l.config.BufferSize
	}

	return &OnlineLearnerStats{
		TotalSamples:    l.count,
		BufferFilled:    filled,
		PeakAccuracy:    l.peakAccuracy,
		RecentAccuracy:  l.recentAccuracy,
		DriftDetected:   l.driftDetected,
		DriftEventCount: l.driftEventCount,
		LastDriftTime:   l.lastDriftTime,
	}
}

// =========================================================================
// Browser Distribution Tracker
// =========================================================================

// BrowserDistribution tracks the observed real-world browser distribution.
type BrowserDistribution struct {
	mu     sync.Mutex
	counts map[string]int64
	total  int64
}

// NewBrowserDistribution creates a new distribution tracker.
func NewBrowserDistribution() *BrowserDistribution {
	return &BrowserDistribution{
		counts: make(map[string]int64),
	}
}

// Record records an observation of a browser family.
func (d *BrowserDistribution) Record(family string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.counts[family]++
	d.total++
}

// Distribution returns the current probability distribution.
func (d *BrowserDistribution) Distribution() map[string]float64 {
	d.mu.Lock()
	defer d.mu.Unlock()

	dist := make(map[string]float64, len(d.counts))
	if d.total == 0 {
		return dist
	}
	for k, v := range d.counts {
		dist[k] = float64(v) / float64(d.total)
	}
	return dist
}

// KLDivergence computes the KL divergence between the observed distribution
// and a reference distribution. Higher values indicate more drift.
func (d *BrowserDistribution) KLDivergence(reference map[string]float64) float64 {
	observed := d.Distribution()
	if len(observed) == 0 || len(reference) == 0 {
		return 0
	}

	kl := 0.0
	epsilon := 1e-10 // avoid log(0)
	for k, p := range observed {
		q := reference[k]
		if q < epsilon {
			q = epsilon
		}
		if p < epsilon {
			continue
		}
		kl += p * math.Log(p/q)
	}
	return kl
}

// Total returns the total number of observations.
func (d *BrowserDistribution) Total() int64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.total
}
