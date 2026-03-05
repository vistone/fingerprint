// Package core 提供对象池实现以减少 GC 压力
package core

import (
	"bytes"
	"strings"
	"sync"
)

// PoolConfig 池配置
type PoolConfig struct {
	// MaxSize 池中最大对象数（0 = 无限制）
	MaxSize int

	// InitialSize 初始对象数
	InitialSize int

	// EnableMetrics 是否启用监控
	EnableMetrics bool
}

// DefaultPoolConfig 默认池配置
var DefaultPoolConfig = PoolConfig{
	MaxSize:       10000,
	InitialSize:   100,
	EnableMetrics: true,
}

// FeatureVectorPool FeatureVector 对象池
type FeatureVectorPool struct {
	pool    sync.Pool
	config  PoolConfig
	metrics *poolMetrics
}

// NewFeatureVectorPool 创建新的 FeatureVector 池
func NewFeatureVectorPool(config *PoolConfig) *FeatureVectorPool {
	if config == nil {
		config = &DefaultPoolConfig
	}

	p := &FeatureVectorPool{
		config: *config,
	}

	if config.EnableMetrics {
		p.metrics = GlobalMonitor.Register("feature_vector")
	}

	p.pool.New = func() interface{} {
		if p.metrics != nil {
			p.metrics.recordGet(false) // miss
		}
		return NewFeatureVector()
	}

	// 预分配初始对象
	for i := 0; i < config.InitialSize; i++ {
		p.pool.Put(NewFeatureVector())
	}

	return p
}

// Get 从池中获取 FeatureVector
func (p *FeatureVectorPool) Get() *FeatureVector {
	v := p.pool.Get()
	if fv, ok := v.(*FeatureVector); ok {
		if p.metrics != nil {
			// 通过 New 函数已经记录了 miss，这里只记录 hit
			if len(fv.Features) > 0 {
				p.metrics.recordGet(true) // hit
			}
		}
		// 重置状态 - 直接重建 map 更快
		fv.Features = make(map[FeatureType]float64, 32)
		fv.Metadata = make(map[string]interface{}, 8)
		return fv
	}
	return nil
}

// Put 将 FeatureVector 归还到池中
func (p *FeatureVectorPool) Put(fv *FeatureVector) {
	if fv == nil {
		return
	}

	// 清理敏感数据 - 直接重建 map 更快
	fv.Features = make(map[FeatureType]float64, 32)
	fv.Metadata = make(map[string]interface{}, 8)

	if p.metrics != nil {
		p.metrics.recordPut()
	}

	p.pool.Put(fv)
}

// Stats 获取池统计
func (p *FeatureVectorPool) Stats() PoolStats {
	if p.metrics != nil {
		return p.metrics.Stats()
	}
	return PoolStats{Name: "feature_vector"}
}

// ============================================================================
// HTTPHeadersPool HTTPHeaders 对象池
// ============================================================================

// HTTPHeadersPool HTTPHeaders 对象池
type HTTPHeadersPool struct {
	pool    sync.Pool
	config  PoolConfig
	metrics *poolMetrics
}

// NewHTTPHeadersPool 创建新的 HTTPHeaders 池
func NewHTTPHeadersPool(config *PoolConfig) *HTTPHeadersPool {
	if config == nil {
		config = &DefaultPoolConfig
	}

	p := &HTTPHeadersPool{
		config: *config,
	}

	if config.EnableMetrics {
		p.metrics = GlobalMonitor.Register("http_headers")
	}

	p.pool.New = func() interface{} {
		if p.metrics != nil {
			p.metrics.recordGet(false)
		}
		return &HTTPHeaders{
			Custom: make(map[string]string, 16),
		}
	}

	for i := 0; i < config.InitialSize; i++ {
		p.pool.Put(&HTTPHeaders{
			Custom: make(map[string]string, 16),
		})
	}

	return p
}

// Get 从池中获取 HTTPHeaders
func (p *HTTPHeadersPool) Get() *HTTPHeaders {
	v := p.pool.Get()
	if h, ok := v.(*HTTPHeaders); ok {
		// 重置所有字段
		h.UserAgent = ""
		h.Accept = ""
		h.AcceptLanguage = ""
		h.AcceptEncoding = ""
		// Custom headers are cleared below - 直接重建 map 更快
		h.Custom = make(map[string]string, 8)
		return h
	}
	return nil
}

// Put 将 HTTPHeaders 归还到池中
func (p *HTTPHeadersPool) Put(h *HTTPHeaders) {
	if h == nil {
		return
	}

	// 清理敏感数据 - 直接重建 map 更快
	h.Custom = make(map[string]string, 8)

	if p.metrics != nil {
		p.metrics.recordPut()
	}

	p.pool.Put(h)
}

// Stats 获取池统计
func (p *HTTPHeadersPool) Stats() PoolStats {
	if p.metrics != nil {
		return p.metrics.Stats()
	}
	return PoolStats{Name: "http_headers"}
}

// ============================================================================
// StringBuilderPool strings.Builder 对象池
// ============================================================================

// StringBuilderPool strings.Builder 对象池
type StringBuilderPool struct {
	pool    sync.Pool
	config  PoolConfig
	metrics *poolMetrics
}

// NewStringBuilderPool 创建新的 StringBuilder 池
func NewStringBuilderPool(config *PoolConfig) *StringBuilderPool {
	if config == nil {
		config = &DefaultPoolConfig
	}

	p := &StringBuilderPool{
		config: *config,
	}

	if config.EnableMetrics {
		p.metrics = GlobalMonitor.Register("string_builder")
	}

	p.pool.New = func() interface{} {
		if p.metrics != nil {
			p.metrics.recordGet(false)
		}
		return &strings.Builder{}
	}

	return p
}

// Get 从池中获取 strings.Builder
func (p *StringBuilderPool) Get() *strings.Builder {
	v := p.pool.Get()
	if sb, ok := v.(*strings.Builder); ok {
		sb.Reset()
		return sb
	}
	return nil
}

// Put 将 strings.Builder 归还到池中
func (p *StringBuilderPool) Put(sb *strings.Builder) {
	if sb == nil {
		return
	}

	if p.metrics != nil {
		p.metrics.recordPut()
	}

	p.pool.Put(sb)
}

// Stats 获取池统计
func (p *StringBuilderPool) Stats() PoolStats {
	if p.metrics != nil {
		return p.metrics.Stats()
	}
	return PoolStats{Name: "string_builder"}
}

// ============================================================================
// BufferPool bytes.Buffer 对象池
// ============================================================================

// BufferPool bytes.Buffer 对象池
type BufferPool struct {
	pool    sync.Pool
	config  PoolConfig
	metrics *poolMetrics
}

// NewBufferPool 创建新的 Buffer 池
func NewBufferPool(config *PoolConfig) *BufferPool {
	if config == nil {
		config = &DefaultPoolConfig
	}

	p := &BufferPool{
		config: *config,
	}

	if config.EnableMetrics {
		p.metrics = GlobalMonitor.Register("buffer")
	}

	p.pool.New = func() interface{} {
		if p.metrics != nil {
			p.metrics.recordGet(false)
		}
		return &bytes.Buffer{}
	}

	return p
}

// Get 从池中获取 bytes.Buffer
func (p *BufferPool) Get() *bytes.Buffer {
	v := p.pool.Get()
	if b, ok := v.(*bytes.Buffer); ok {
		b.Reset()
		return b
	}
	return nil
}

// Put 将 bytes.Buffer 归还到池中
func (p *BufferPool) Put(b *bytes.Buffer) {
	if b == nil {
		return
	}

	// 限制缓冲区大小，防止内存泄漏
	if b.Cap() > 64*1024 { // 64KB
		// 不归还过大的缓冲区，让 GC 回收
		return
	}

	if p.metrics != nil {
		p.metrics.recordPut()
	}

	p.pool.Put(b)
}

// Stats 获取池统计
func (p *BufferPool) Stats() PoolStats {
	if p.metrics != nil {
		return p.metrics.Stats()
	}
	return PoolStats{Name: "buffer"}
}

// ============================================================================
// MapPool map[string]interface{} 对象池
// ============================================================================

// MapPool map 对象池
type MapPool struct {
	pool    sync.Pool
	config  PoolConfig
	metrics *poolMetrics
}

// NewMapPool 创建新的 Map 池
func NewMapPool(config *PoolConfig) *MapPool {
	if config == nil {
		config = &DefaultPoolConfig
	}

	p := &MapPool{
		config: *config,
	}

	if config.EnableMetrics {
		p.metrics = GlobalMonitor.Register("map")
	}

	p.pool.New = func() interface{} {
		if p.metrics != nil {
			p.metrics.recordGet(false)
		}
		return make(map[string]interface{}, 16)
	}

	return p
}

// Get 从池中获取 map
func (p *MapPool) Get() map[string]interface{} {
	v := p.pool.Get()
	if _, ok := v.(map[string]interface{}); ok {
		// 清空 map - 直接重建更快
		return make(map[string]interface{}, 16)
	}
	return nil
}

// Put 将 map 归还到池中
func (p *MapPool) Put(m map[string]interface{}) {
	if m == nil {
		return
	}

	// 清理 map - 直接重建更快，让 GC 回收旧 map
	m = make(map[string]interface{}, 16)

	if p.metrics != nil {
		p.metrics.recordPut()
	}

	p.pool.Put(m)
}

// Stats 获取池统计
func (p *MapPool) Stats() PoolStats {
	if p.metrics != nil {
		return p.metrics.Stats()
	}
	return PoolStats{Name: "map"}
}

// ============================================================================
// SlicePool []byte 对象池
// ============================================================================

// SlicePool byte slice 对象池
type SlicePool struct {
	pool    sync.Pool
	config  PoolConfig
	metrics *poolMetrics
	size    int
}

// NewSlicePool 创建新的 Slice 池
func NewSlicePool(size int, config *PoolConfig) *SlicePool {
	if config == nil {
		config = &DefaultPoolConfig
	}
	if size <= 0 {
		size = 1024 // 默认 1KB
	}

	p := &SlicePool{
		config: *config,
		size:   size,
	}

	if config.EnableMetrics {
		p.metrics = GlobalMonitor.Register("slice")
	}

	p.pool.New = func() interface{} {
		if p.metrics != nil {
			p.metrics.recordGet(false)
		}
		return make([]byte, 0, size)
	}

	return p
}

// Get 从池中获取 byte slice
func (p *SlicePool) Get() []byte {
	v := p.pool.Get()
	if s, ok := v.([]byte); ok {
		return s[:0] // 重置长度，保留容量
	}
	return nil
}

// Put 将 byte slice 归还到池中
func (p *SlicePool) Put(s []byte) {
	if s == nil {
		return
	}

	// 限制容量，防止内存泄漏
	if cap(s) > p.size*4 {
		return // 不归还过大的切片
	}

	if p.metrics != nil {
		p.metrics.recordPut()
	}

	p.pool.Put(s[:0]) // 重置长度
}

// Stats 获取池统计
func (p *SlicePool) Stats() PoolStats {
	if p.metrics != nil {
		return p.metrics.Stats()
	}
	return PoolStats{Name: "slice"}
}

// ============================================================================
// 全局池实例
// ============================================================================

// GlobalPools 全局对象池集合
type GlobalPools struct {
	FeatureVectors *FeatureVectorPool
	HTTPHeaders    *HTTPHeadersPool
	Strings        *StringBuilderPool
	Buffers        *BufferPool
	Maps           *MapPool
	Slices         *SlicePool
}

// Pools 全局池实例
var Pools = &GlobalPools{}

// InitPools 初始化全局对象池
func InitPools() {
	config := &DefaultPoolConfig

	Pools.FeatureVectors = NewFeatureVectorPool(config)
	Pools.HTTPHeaders = NewHTTPHeadersPool(config)
	Pools.Strings = NewStringBuilderPool(config)
	Pools.Buffers = NewBufferPool(config)
	Pools.Maps = NewMapPool(config)
	Pools.Slices = NewSlicePool(4096, config) // 4KB
}

// InitPoolsWithConfig 使用自定义配置初始化全局对象池
func InitPoolsWithConfig(config *PoolConfig) {
	Pools.FeatureVectors = NewFeatureVectorPool(config)
	Pools.HTTPHeaders = NewHTTPHeadersPool(config)
	Pools.Strings = NewStringBuilderPool(config)
	Pools.Buffers = NewBufferPool(config)
	Pools.Maps = NewMapPool(config)
	Pools.Slices = NewSlicePool(4096, config)
}

func init() {
	// 延迟初始化，避免启动时开销
	// 用户需要显式调用 InitPools()
}

// ============================================================================
// 便捷函数
// ============================================================================

// AcquireFeatureVector 获取 FeatureVector
func AcquireFeatureVector() *FeatureVector {
	if Pools.FeatureVectors == nil {
		return NewFeatureVector()
	}
	return Pools.FeatureVectors.Get()
}

// ReleaseFeatureVector 释放 FeatureVector
func ReleaseFeatureVector(fv *FeatureVector) {
	if Pools.FeatureVectors != nil && fv != nil {
		Pools.FeatureVectors.Put(fv)
	}
}

// AcquireHTTPHeaders 获取 HTTPHeaders
func AcquireHTTPHeaders() *HTTPHeaders {
	if Pools.HTTPHeaders == nil {
		return &HTTPHeaders{Custom: make(map[string]string, 8)}
	}
	return Pools.HTTPHeaders.Get()
}

// ReleaseHTTPHeaders 释放 HTTPHeaders
func ReleaseHTTPHeaders(h *HTTPHeaders) {
	if Pools.HTTPHeaders != nil && h != nil {
		Pools.HTTPHeaders.Put(h)
	}
}

// AcquireStringBuilder 获取 strings.Builder
func AcquireStringBuilder() *strings.Builder {
	if Pools.Strings == nil {
		return &strings.Builder{}
	}
	return Pools.Strings.Get()
}

// ReleaseStringBuilder 释放 strings.Builder
func ReleaseStringBuilder(sb *strings.Builder) {
	if Pools.Strings != nil && sb != nil {
		Pools.Strings.Put(sb)
	}
}

// AcquireBuffer 获取 bytes.Buffer
func AcquireBuffer() *bytes.Buffer {
	if Pools.Buffers == nil {
		return &bytes.Buffer{}
	}
	return Pools.Buffers.Get()
}

// ReleaseBuffer 释放 bytes.Buffer
func ReleaseBuffer(b *bytes.Buffer) {
	if Pools.Buffers != nil && b != nil {
		Pools.Buffers.Put(b)
	}
}

// AcquireMap 获取 map
func AcquireMap() map[string]interface{} {
	if Pools.Maps == nil {
		return make(map[string]interface{})
	}
	return Pools.Maps.Get()
}

// ReleaseMap 释放 map
func ReleaseMap(m map[string]interface{}) {
	if Pools.Maps != nil && m != nil {
		Pools.Maps.Put(m)
	}
}

// AcquireSlice 获取 byte slice
func AcquireSlice() []byte {
	if Pools.Slices == nil {
		return make([]byte, 0, 4096)
	}
	return Pools.Slices.Get()
}

// ReleaseSlice 释放 byte slice
func ReleaseSlice(s []byte) {
	if Pools.Slices != nil && s != nil {
		Pools.Slices.Put(s)
	}
}
