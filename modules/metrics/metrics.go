package metrics

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Collector is a Prometheus-compatible metrics collector
type Collector struct {
	mu         sync.RWMutex
	counters   map[string]*Counter
	gauges     map[string]*Gauge
	histograms map[string]*Histogram
	summaries  map[string]*Summary
}

// Counter represents a monotonically increasing counter
type Counter struct {
	Name   string
	Help   string
	Labels []string
	values map[string]float64
}

// Gauge represents a value that can go up or down
type Gauge struct {
	Name   string
	Help   string
	Labels []string
	values map[string]float64
}

// Histogram represents a distribution of values
type Histogram struct {
	Name    string
	Help    string
	Buckets []float64
	counts  map[float64]uint64
	sum     float64
	count   uint64
}

// Summary represents a sliding time window summary
type Summary struct {
	Name   string
	Help   string
	values []float64
	sum    float64
	count  uint64
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		counters:   make(map[string]*Counter),
		gauges:     make(map[string]*Gauge),
		histograms: make(map[string]*Histogram),
		summaries:  make(map[string]*Summary),
	}
}

// Counter creates or gets a counter
func (c *Collector) Counter(name, help string, labels ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.counters[name]; !exists {
		c.counters[name] = &Counter{
			Name:   name,
			Help:   help,
			Labels: labels,
			values: make(map[string]float64),
		}
	}
}

// Inc increments a counter
func (c *Collector) Inc(name string, labelValues ...string) {
	c.Add(name, 1, labelValues...)
}

// Add adds a value to a counter
func (c *Collector) Add(name string, value float64, labelValues ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	counter, exists := c.counters[name]
	if !exists {
		return
	}

	key := makeLabelKey(labelValues)
	counter.values[key] += value
}

// GetCounter gets the current value of a counter
func (c *Collector) GetCounter(name string, labelValues ...string) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	counter, exists := c.counters[name]
	if !exists {
		return 0
	}

	key := makeLabelKey(labelValues)
	return counter.values[key]
}

// Gauge creates or gets a gauge
func (c *Collector) Gauge(name, help string, labels ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.gauges[name]; !exists {
		c.gauges[name] = &Gauge{
			Name:   name,
			Help:   help,
			Labels: labels,
			values: make(map[string]float64),
		}
	}
}

// Set sets a gauge value
func (c *Collector) Set(name string, value float64, labelValues ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	gauge, exists := c.gauges[name]
	if !exists {
		return
	}

	key := makeLabelKey(labelValues)
	gauge.values[key] = value
}

// GetGauge gets the current value of a gauge
func (c *Collector) GetGauge(name string, labelValues ...string) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	gauge, exists := c.gauges[name]
	if !exists {
		return 0
	}

	key := makeLabelKey(labelValues)
	return gauge.values[key]
}

// Histogram creates or gets a histogram
func (c *Collector) Histogram(name, help string, buckets []float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.histograms[name]; !exists {
		c.histograms[name] = &Histogram{
			Name:    name,
			Help:    help,
			Buckets: buckets,
			counts:  make(map[float64]uint64),
		}
	}
}

// Observe adds an observation to a histogram
func (c *Collector) Observe(name string, value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hist, exists := c.histograms[name]
	if !exists {
		return
	}

	hist.sum += value
	hist.count++

	// Find the appropriate bucket
	for _, bucket := range hist.Buckets {
		if value <= bucket {
			hist.counts[bucket]++
		}
	}
}

// GetHistogramCount returns the count of observations
func (c *Collector) GetHistogramCount(name string) uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	hist, exists := c.histograms[name]
	if !exists {
		return 0
	}

	return hist.count
}

// GetHistogramSum returns the sum of observations
func (c *Collector) GetHistogramSum(name string) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	hist, exists := c.histograms[name]
	if !exists {
		return 0
	}

	return hist.sum
}

// Summary creates or gets a summary
func (c *Collector) Summary(name, help string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.summaries[name]; !exists {
		c.summaries[name] = &Summary{
			Name:   name,
			Help:   help,
			values: make([]float64, 0, 1000),
		}
	}
}

// ObserveSummary adds an observation to a summary
func (c *Collector) ObserveSummary(name string, value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	sum, exists := c.summaries[name]
	if !exists {
		return
	}

	sum.values = append(sum.values, value)
	sum.sum += value
	sum.count++
}

// GetSummaryCount returns the count of observations
func (c *Collector) GetSummaryCount(name string) uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sum, exists := c.summaries[name]
	if !exists {
		return 0
	}

	return sum.count
}

// PrometheusExport exports metrics in Prometheus format
func (c *Collector) PrometheusExport() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var sb strings.Builder

	// Export counters
	for name, counter := range c.counters {
		fmt.Fprintf(&sb, "# HELP %s %s\n", name, counter.Help)
		fmt.Fprintf(&sb, "# TYPE %s counter\n", name)
		for labels, value := range counter.values {
			if labels == "" {
				fmt.Fprintf(&sb, "%s %v\n", name, value)
			} else {
				fmt.Fprintf(&sb, "%s{%s} %v\n", name, labels, value)
			}
		}
	}

	// Export gauges
	for name, gauge := range c.gauges {
		fmt.Fprintf(&sb, "# HELP %s %s\n", name, gauge.Help)
		fmt.Fprintf(&sb, "# TYPE %s gauge\n", name)
		for labels, value := range gauge.values {
			if labels == "" {
				fmt.Fprintf(&sb, "%s %v\n", name, value)
			} else {
				fmt.Fprintf(&sb, "%s{%s} %v\n", name, labels, value)
			}
		}
	}

	// Export histograms
	for name, hist := range c.histograms {
		fmt.Fprintf(&sb, "# HELP %s %s\n", name, hist.Help)
		fmt.Fprintf(&sb, "# TYPE %s histogram\n", name)
		for _, bucket := range hist.Buckets {
			count := hist.counts[bucket]
			fmt.Fprintf(&sb, "%s_bucket{le=\"%v\"} %d\n", name, bucket, count)
		}
		fmt.Fprintf(&sb, "%s_sum %v\n", name, hist.sum)
		fmt.Fprintf(&sb, "%s_count %d\n", name, hist.count)
	}

	return sb.String()
}

// Reset clears all metrics
func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.counters = make(map[string]*Counter)
	c.gauges = make(map[string]*Gauge)
	c.histograms = make(map[string]*Histogram)
	c.summaries = make(map[string]*Summary)
}

// Timer is a helper for timing operations
type Timer struct {
	collector *Collector
	histogram string
	start     time.Time
}

// StartTimer starts a new timer
func (c *Collector) StartTimer(histogram string) *Timer {
	return &Timer{
		collector: c,
		histogram: histogram,
		start:     time.Now(),
	}
}

// ObserveDuration records the elapsed time
func (t *Timer) ObserveDuration() {
	elapsed := time.Since(t.start).Seconds()
	t.collector.Observe(t.histogram, elapsed)
}

// Time is a helper that returns a function to defer
func (c *Collector) Time(histogram string) func() {
	start := time.Now()
	return func() {
		elapsed := time.Since(start).Seconds()
		c.Observe(histogram, elapsed)
	}
}

// makeLabelKey creates a key from label values
func makeLabelKey(labelValues []string) string {
	if len(labelValues) == 0 {
		return ""
	}
	return strings.Join(labelValues, ",")
}

// FingerprintMetrics provides specialized metrics for fingerprint operations
type FingerprintMetrics struct {
	mu                   sync.RWMutex
	totalClassifications int64
	cacheHits            int64
	cacheMisses          int64
	totalConfidence      float64
}

// NewFingerprintMetrics creates new fingerprint metrics
func NewFingerprintMetrics() *FingerprintMetrics {
	return &FingerprintMetrics{}
}

// RecordClassification records a classification
func (fm *FingerprintMetrics) RecordClassification(family string, confidence float64, duration time.Duration) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.totalClassifications++
	fm.totalConfidence += confidence
}

// RecordCacheHit records a cache hit
func (fm *FingerprintMetrics) RecordCacheHit() {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.cacheHits++
}

// RecordCacheMiss records a cache miss
func (fm *FingerprintMetrics) RecordCacheMiss() {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.cacheMisses++
}

// FingerprintStats holds fingerprint metrics statistics
type FingerprintStats struct {
	TotalClassifications int64
	AverageConfidence    float64
	CacheHits            int64
	CacheMisses          int64
	CacheHitRate         float64
}

// GetStats returns fingerprint statistics
func (fm *FingerprintMetrics) GetStats() FingerprintStats {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	totalCache := fm.cacheHits + fm.cacheMisses
	hitRate := 0.0
	if totalCache > 0 {
		hitRate = float64(fm.cacheHits) / float64(totalCache)
	}

	avgConfidence := 0.0
	if fm.totalClassifications > 0 {
		avgConfidence = fm.totalConfidence / float64(fm.totalClassifications)
	}

	return FingerprintStats{
		TotalClassifications: fm.totalClassifications,
		AverageConfidence:    avgConfidence,
		CacheHits:            fm.cacheHits,
		CacheMisses:          fm.cacheMisses,
		CacheHitRate:         hitRate,
	}
}

// GetCacheHitRate returns the cache hit rate
func (fm *FingerprintMetrics) GetCacheHitRate() float64 {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	total := fm.cacheHits + fm.cacheMisses
	if total == 0 {
		return 0
	}
	return float64(fm.cacheHits) / float64(total)
}
