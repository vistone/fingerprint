// Package core provides memory pool monitoring metrics
package core

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// PoolStats memory pool statistics info
type PoolStats struct {
	Name          string  `json:"name"`
	HitRate       float64 `json:"hit_rate"`        // cache hit rate
	ObjectsInUse  int64   `json:"objects_in_use"`  // number of objects in use
	TotalGets     int64   `json:"total_gets"`      // total get count
	TotalPuts     int64   `json:"total_puts"`      // total put count
	CacheHits     int64   `json:"cache_hits"`      // cache hit count
	CacheMisses   int64   `json:"cache_misses"`    // cache miss count
	ActiveObjects int64   `json:"active_objects"`  // current active object count (estimated)
}

// GlobalPoolMetrics global pool metrics (atomic operations)
var GlobalPoolMetrics = struct {
	// individual pool statistics
	FeatureVectorGets     int64
	FeatureVectorPuts     int64
	FeatureVectorHits     int64
	HTTPHeadersGets       int64
	HTTPHeadersPuts       int64
	HTTPHeadersHits       int64
	StringBuilderGets     int64
	StringBuilderPuts     int64
	MapPoolGets           int64
	MapPoolPuts           int64
	SlicePoolGets         int64
	SlicePoolPuts         int64

	// summaries
	TotalAllocations int64
	TotalReleases    int64
}{}

// poolMetrics internal metrics structure
type poolMetrics struct {
	name        string
	gets        int64
	puts        int64
	hits        int64
	misses      int64
	inUse       int64
	mu          sync.RWMutex
	onGet       func()
	onPut       func()
	onHit       func()
	onMiss      func()
}

// newPoolMetrics creates pool metrics
func newPoolMetrics(name string) *poolMetrics {
	return &poolMetrics{name: name}
}

// recordGet records Get operation
func (m *poolMetrics) recordGet(hit bool) {
	atomic.AddInt64(&m.gets, 1)
	if hit {
		atomic.AddInt64(&m.hits, 1)
		atomic.AddInt64(&m.inUse, 1)
	} else {
		atomic.AddInt64(&m.misses, 1)
		atomic.AddInt64(&m.inUse, 1)
	}
}

// recordPut records Put operation
func (m *poolMetrics) recordPut() {
	atomic.AddInt64(&m.puts, 1)
	atomic.AddInt64(&m.inUse, -1)
}

// Stats gets current statistics
func (m *poolMetrics) Stats() PoolStats {
	gets := atomic.LoadInt64(&m.gets)
	hits := atomic.LoadInt64(&m.hits)
	misses := atomic.LoadInt64(&m.misses)
	puts := atomic.LoadInt64(&m.puts)
	inUse := atomic.LoadInt64(&m.inUse)

	hitRate := float64(0)
	if gets > 0 {
		hitRate = float64(hits) / float64(gets)
	}

	return PoolStats{
		Name:          m.name,
		HitRate:       hitRate,
		ObjectsInUse:  inUse,
		TotalGets:     gets,
		TotalPuts:     puts,
		CacheHits:     hits,
		CacheMisses:   misses,
		ActiveObjects: inUse,
	}
}

// Reset resets statistics
func (m *poolMetrics) Reset() {
	atomic.StoreInt64(&m.gets, 0)
	atomic.StoreInt64(&m.puts, 0)
	atomic.StoreInt64(&m.hits, 0)
	atomic.StoreInt64(&m.misses, 0)
	atomic.StoreInt64(&m.inUse, 0)
}

// ============================================================================
// global pool monitoring manager
// ============================================================================

// PoolMonitor pool monitoring manager
type PoolMonitor struct {
	metrics map[string]*poolMetrics
	mu      sync.RWMutex
}

// GlobalMonitor global monitor instance
var GlobalMonitor = NewPoolMonitor()

// NewPoolMonitor creates new pool monitor
func NewPoolMonitor() *PoolMonitor {
	return &PoolMonitor{
		metrics: make(map[string]*poolMetrics),
	}
}

// Register registers pool metrics
func (pm *PoolMonitor) Register(name string) *poolMetrics {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if m, exists := pm.metrics[name]; exists {
		return m
	}

	m := newPoolMetrics(name)
	pm.metrics[name] = m
	return m
}

// Get gets pool metrics
func (pm *PoolMonitor) Get(name string) (*poolMetrics, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	m, ok := pm.metrics[name]
	return m, ok
}

// AllStats gets all pool statistics
func (pm *PoolMonitor) AllStats() map[string]PoolStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := make(map[string]PoolStats, len(pm.metrics))
	for name, m := range pm.metrics {
		stats[name] = m.Stats()
	}
	return stats
}

// Summary gets summary statistics
func (pm *PoolMonitor) Summary() PoolSummary {
	allStats := pm.AllStats()

	var totalGets, totalPuts, totalHits, totalMisses, totalInUse int64
	for _, s := range allStats {
		totalGets += s.TotalGets
		totalPuts += s.TotalPuts
		totalHits += s.CacheHits
		totalMisses += s.CacheMisses
		totalInUse += s.ObjectsInUse
	}

	totalRate := float64(0)
	if totalGets > 0 {
		totalRate = float64(totalHits) / float64(totalGets)
	}

	return PoolSummary{
		PoolCount:     len(allStats),
		TotalGets:     totalGets,
		TotalPuts:     totalPuts,
		TotalHits:     totalHits,
		TotalMisses:   totalMisses,
		TotalInUse:    totalInUse,
		OverallHitRate: totalRate,
		Pools:         allStats,
	}
}

// ResetAll resets all statistics
func (pm *PoolMonitor) ResetAll() {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, m := range pm.metrics {
		m.Reset()
	}
}

// PoolSummary pool summary statistics
type PoolSummary struct {
	PoolCount      int                 `json:"pool_count"`
	TotalGets      int64               `json:"total_gets"`
	TotalPuts      int64               `json:"total_puts"`
	TotalHits      int64               `json:"total_hits"`
	TotalMisses    int64               `json:"total_misses"`
	TotalInUse     int64               `json:"total_in_use"`
	OverallHitRate float64             `json:"overall_hit_rate"`
	Pools          map[string]PoolStats `json:"pools"`
}

// ============================================================================
// utility functions
// ============================================================================

// GetPoolStats gets statistics for specified pool
func GetPoolStats(name string) (PoolStats, bool) {
	if m, ok := GlobalMonitor.Get(name); ok {
		return m.Stats(), true
	}
	return PoolStats{}, false
}

// GetAllPoolStats gets all pool statistics
func GetAllPoolStats() map[string]PoolStats {
	return GlobalMonitor.AllStats()
}

// GetPoolSummary gets pool summary statistics
func GetPoolSummary() PoolSummary {
	return GlobalMonitor.Summary()
}

// ResetPoolMetrics resets all pool metrics
func ResetPoolMetrics() {
	GlobalMonitor.ResetAll()
}

// ForceGC forces garbage collection and returns statistics
func ForceGC() runtime.MemStats {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	return m
}

// GetMemoryStats gets memory statistics
func GetMemoryStats() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}
