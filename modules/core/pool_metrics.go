// Package core 提供内存池监控指标
package core

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// PoolStats 内存池统计信息
type PoolStats struct {
	Name          string  `json:"name"`
	HitRate       float64 `json:"hit_rate"`        // 缓存命中率
	ObjectsInUse  int64   `json:"objects_in_use"`  // 正在使用的对象数
	TotalGets     int64   `json:"total_gets"`      // 总获取次数
	TotalPuts     int64   `json:"total_puts"`      // 总归还次数
	CacheHits     int64   `json:"cache_hits"`      // 缓存命中次数
	CacheMisses   int64   `json:"cache_misses"`    // 缓存未命中次数
	ActiveObjects int64   `json:"active_objects"`  // 当前活跃对象数（估算）
}

// GlobalPoolMetrics 全局池监控（原子操作）
var GlobalPoolMetrics = struct {
	// 各池统计
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

	// 总计
	TotalAllocations int64
	TotalReleases    int64
}{}

// poolMetrics 内部监控结构
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

// newPoolMetrics 创建池监控
func newPoolMetrics(name string) *poolMetrics {
	return &poolMetrics{name: name}
}

// recordGet 记录 Get 操作
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

// recordPut 记录 Put 操作
func (m *poolMetrics) recordPut() {
	atomic.AddInt64(&m.puts, 1)
	atomic.AddInt64(&m.inUse, -1)
}

// Stats 获取当前统计
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

// Reset 重置统计
func (m *poolMetrics) Reset() {
	atomic.StoreInt64(&m.gets, 0)
	atomic.StoreInt64(&m.puts, 0)
	atomic.StoreInt64(&m.hits, 0)
	atomic.StoreInt64(&m.misses, 0)
	atomic.StoreInt64(&m.inUse, 0)
}

// ============================================================================
// 全局池监控管理器
// ============================================================================

// PoolMonitor 池监控管理器
type PoolMonitor struct {
	metrics map[string]*poolMetrics
	mu      sync.RWMutex
}

// GlobalMonitor 全局监控实例
var GlobalMonitor = NewPoolMonitor()

// NewPoolMonitor 创建新的池监控器
func NewPoolMonitor() *PoolMonitor {
	return &PoolMonitor{
		metrics: make(map[string]*poolMetrics),
	}
}

// Register 注册池监控
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

// Get 获取池监控
func (pm *PoolMonitor) Get(name string) (*poolMetrics, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	m, ok := pm.metrics[name]
	return m, ok
}

// AllStats 获取所有池统计
func (pm *PoolMonitor) AllStats() map[string]PoolStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := make(map[string]PoolStats, len(pm.metrics))
	for name, m := range pm.metrics {
		stats[name] = m.Stats()
	}
	return stats
}

// Summary 获取汇总统计
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

// ResetAll 重置所有统计
func (pm *PoolMonitor) ResetAll() {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, m := range pm.metrics {
		m.Reset()
	}
}

// PoolSummary 池汇总统计
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
// 实用函数
// ============================================================================

// GetPoolStats 获取指定池的统计
func GetPoolStats(name string) (PoolStats, bool) {
	if m, ok := GlobalMonitor.Get(name); ok {
		return m.Stats(), true
	}
	return PoolStats{}, false
}

// GetAllPoolStats 获取所有池统计
func GetAllPoolStats() map[string]PoolStats {
	return GlobalMonitor.AllStats()
}

// GetPoolSummary 获取池汇总统计
func GetPoolSummary() PoolSummary {
	return GlobalMonitor.Summary()
}

// ResetPoolMetrics 重置所有池监控
func ResetPoolMetrics() {
	GlobalMonitor.ResetAll()
}

// ForceGC 强制垃圾回收并返回统计
func ForceGC() runtime.MemStats {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	return m
}

// GetMemoryStats 获取内存统计
func GetMemoryStats() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}
