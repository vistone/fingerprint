// Package core provides object pool implementation to reduce GC pressure
package core

import (
	"bytes"
	"strings"
	"sync"
)

// PoolConfig pool configuration
type PoolConfig struct {
	// MaxSize maximum number of objects in pool (0 = unlimited)
	MaxSize int

	// InitialSize initial number of objects
	InitialSize int

	// EnableMetrics whether to enable monitoring
	EnableMetrics bool
}

// DefaultPoolConfig default pool configuration
var DefaultPoolConfig = PoolConfig{
	MaxSize:       10000,
	InitialSize:   100,
	EnableMetrics: true,
}

// FeatureVectorPool FeatureVector object pool
type FeatureVectorPool struct {
	pool    sync.Pool
	config  PoolConfig
	metrics *poolMetrics
}

// NewFeatureVectorPool creates a new FeatureVector pool
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

	// pre-allocate initial objects
	for i := 0; i < config.InitialSize; i++ {
		p.pool.Put(NewFeatureVector())
	}

	return p
}

// Get fetches FeatureVector from pool
func (p *FeatureVectorPool) Get() *FeatureVector {
	v := p.pool.Get()
	if fv, ok := v.(*FeatureVector); ok {
		if p.metrics != nil {
			// Miss already logged by New function, only log hit here
			if len(fv.Features) > 0 {
				p.metrics.recordGet(true) // hit
			}
		}
		// reset state - rebuilding map is faster
		fv.Features = make(map[FeatureType]float64, 32)
		fv.Metadata = make(map[string]interface{}, 8)
		return fv
	}
	return nil
}

// Put returns FeatureVector back to pool
func (p *FeatureVectorPool) Put(fv *FeatureVector) {
	if fv == nil {
		return
	}

	// cleanup sensitive data - rebuilding map is faster
	fv.Features = make(map[FeatureType]float64, 32)
	fv.Metadata = make(map[string]interface{}, 8)

	if p.metrics != nil {
		p.metrics.recordPut()
	}

	p.pool.Put(fv)
}

// Stats gets pool statistics
func (p *FeatureVectorPool) Stats() PoolStats {
	if p.metrics != nil {
		return p.metrics.Stats()
	}
	return PoolStats{Name: "feature_vector"}
}

// ============================================================================
// HTTPHeadersPool HTTPHeaders object pool
// ============================================================================

// HTTPHeadersPool HTTPHeaders object pool
type HTTPHeadersPool struct {
	pool    sync.Pool
	config  PoolConfig
	metrics *poolMetrics
}

// NewHTTPHeadersPool creates a new HTTPHeaders pool
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

// Get fetches HTTPHeaders from pool
func (p *HTTPHeadersPool) Get() *HTTPHeaders {
	v := p.pool.Get()
	if h, ok := v.(*HTTPHeaders); ok {
		// reset all fields
		h.UserAgent = ""
		h.Accept = ""
		h.AcceptLanguage = ""
		h.AcceptEncoding = ""
		// Custom headers are cleared below - rebuilding map is faster
		h.Custom = make(map[string]string, 8)
		return h
	}
	return nil
}

// Put returns HTTPHeaders back to pool
func (p *HTTPHeadersPool) Put(h *HTTPHeaders) {
	if h == nil {
		return
	}

	// cleanup sensitive data - rebuilding map is faster
	h.Custom = make(map[string]string, 8)

	if p.metrics != nil {
		p.metrics.recordPut()
	}

	p.pool.Put(h)
}

// Stats gets pool statistics
func (p *HTTPHeadersPool) Stats() PoolStats {
	if p.metrics != nil {
		return p.metrics.Stats()
	}
	return PoolStats{Name: "http_headers"}
}

// ============================================================================
// StringBuilderPool strings.Builder object pool
// ============================================================================

// StringBuilderPool strings.Builder object pool
type StringBuilderPool struct {
	pool    sync.Pool
	config  PoolConfig
	metrics *poolMetrics
}

// NewStringBuilderPool creates a new StringBuilder pool
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

// Get fetches strings.Builder from pool
func (p *StringBuilderPool) Get() *strings.Builder {
	v := p.pool.Get()
	if sb, ok := v.(*strings.Builder); ok {
		sb.Reset()
		return sb
	}
	return nil
}

// Put returns strings.Builder back to pool
func (p *StringBuilderPool) Put(sb *strings.Builder) {
	if sb == nil {
		return
	}

	if p.metrics != nil {
		p.metrics.recordPut()
	}

	p.pool.Put(sb)
}

// Stats gets pool statistics
func (p *StringBuilderPool) Stats() PoolStats {
	if p.metrics != nil {
		return p.metrics.Stats()
	}
	return PoolStats{Name: "string_builder"}
}

// ============================================================================
// BufferPool bytes.Buffer object pool
// ============================================================================

// BufferPool bytes.Buffer object pool
type BufferPool struct {
	pool    sync.Pool
	config  PoolConfig
	metrics *poolMetrics
}

// NewBufferPool creates a new Buffer pool
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

// Get fetches bytes.Buffer from pool
func (p *BufferPool) Get() *bytes.Buffer {
	v := p.pool.Get()
	if b, ok := v.(*bytes.Buffer); ok {
		b.Reset()
		return b
	}
	return nil
}

// Put returns bytes.Buffer back to pool
func (p *BufferPool) Put(b *bytes.Buffer) {
	if b == nil {
		return
	}

	// limit buffer size to prevent memory leak
	if b.Cap() > 64*1024 { // 64KB
		// don't return oversized buffer, let GC collect it
		return
	}

	if p.metrics != nil {
		p.metrics.recordPut()
	}

	p.pool.Put(b)
}

// Stats gets pool statistics
func (p *BufferPool) Stats() PoolStats {
	if p.metrics != nil {
		return p.metrics.Stats()
	}
	return PoolStats{Name: "buffer"}
}

// ============================================================================
// MapPool map[string]interface{} object pool
// ============================================================================

// MapPool map object pool
type MapPool struct {
	pool    sync.Pool
	config  PoolConfig
	metrics *poolMetrics
}

// NewMapPool creates a new Map pool
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

// Get fetches map from pool
func (p *MapPool) Get() map[string]interface{} {
	v := p.pool.Get()
	if _, ok := v.(map[string]interface{}); ok {
		// clear map - rebuilding is faster
		return make(map[string]interface{}, 16)
	}
	return nil
}

// Put returns map back to pool
func (p *MapPool) Put(m map[string]interface{}) {
	if m == nil {
		return
	}

	// cleanup map - rebuilding is faster, let GC collect old map
	m = make(map[string]interface{}, 16)

	if p.metrics != nil {
		p.metrics.recordPut()
	}

	p.pool.Put(m)
}

// Stats gets pool statistics
func (p *MapPool) Stats() PoolStats {
	if p.metrics != nil {
		return p.metrics.Stats()
	}
	return PoolStats{Name: "map"}
}

// ============================================================================
// SlicePool []byte object pool
// ============================================================================

// SlicePool byte slice object pool
type SlicePool struct {
	pool    sync.Pool
	config  PoolConfig
	metrics *poolMetrics
	size    int
}

// NewSlicePool creates a new Slice pool
func NewSlicePool(size int, config *PoolConfig) *SlicePool {
	if config == nil {
		config = &DefaultPoolConfig
	}
	if size <= 0 {
		size = 1024 // default 1KB
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

// Get fetches byte slice from pool
func (p *SlicePool) Get() []byte {
	v := p.pool.Get()
	if s, ok := v.([]byte); ok {
		return s[:0] // reset length, preserve capacity
	}
	return nil
}

// Put returns byte slice back to pool
func (p *SlicePool) Put(s []byte) {
	if s == nil {
		return
	}

	// limit capacity to prevent memory leak
	if cap(s) > p.size*4 {
		return // don't return oversized slice
	}

	if p.metrics != nil {
		p.metrics.recordPut()
	}

	p.pool.Put(s[:0]) // reset length
}

// Stats gets pool statistics
func (p *SlicePool) Stats() PoolStats {
	if p.metrics != nil {
		return p.metrics.Stats()
	}
	return PoolStats{Name: "slice"}
}

// ============================================================================
// global pool instances
// ============================================================================

// GlobalPools global object pool collection
type GlobalPools struct {
	FeatureVectors *FeatureVectorPool
	HTTPHeaders    *HTTPHeadersPool
	Strings        *StringBuilderPool
	Buffers        *BufferPool
	Maps           *MapPool
	Slices         *SlicePool
}

// Pools global pool instances
var Pools = &GlobalPools{}

// InitPools initializes global object pools
func InitPools() {
	config := &DefaultPoolConfig

	Pools.FeatureVectors = NewFeatureVectorPool(config)
	Pools.HTTPHeaders = NewHTTPHeadersPool(config)
	Pools.Strings = NewStringBuilderPool(config)
	Pools.Buffers = NewBufferPool(config)
	Pools.Maps = NewMapPool(config)
	Pools.Slices = NewSlicePool(4096, config) // 4KB
}

// InitPoolsWithConfig initializes global object pools with custom configuration
func InitPoolsWithConfig(config *PoolConfig) {
	Pools.FeatureVectors = NewFeatureVectorPool(config)
	Pools.HTTPHeaders = NewHTTPHeadersPool(config)
	Pools.Strings = NewStringBuilderPool(config)
	Pools.Buffers = NewBufferPool(config)
	Pools.Maps = NewMapPool(config)
	Pools.Slices = NewSlicePool(4096, config)
}

func init() {
	// lazy initialization, avoid overhead at startup
	// users must explicitly call InitPools()
}

// ============================================================================
// convenience functions
// ============================================================================

// AcquireFeatureVector get FeatureVector
func AcquireFeatureVector() *FeatureVector {
	if Pools.FeatureVectors == nil {
		return NewFeatureVector()
	}
	return Pools.FeatureVectors.Get()
}

// ReleaseFeatureVector releases FeatureVector back to pool
func ReleaseFeatureVector(fv *FeatureVector) {
	if Pools.FeatureVectors != nil && fv != nil {
		Pools.FeatureVectors.Put(fv)
	}
}

// AcquireHTTPHeaders get HTTPHeaders
func AcquireHTTPHeaders() *HTTPHeaders {
	if Pools.HTTPHeaders == nil {
		return &HTTPHeaders{Custom: make(map[string]string, 8)}
	}
	return Pools.HTTPHeaders.Get()
}

// ReleaseHTTPHeaders releases HTTPHeaders back to pool
func ReleaseHTTPHeaders(h *HTTPHeaders) {
	if Pools.HTTPHeaders != nil && h != nil {
		Pools.HTTPHeaders.Put(h)
	}
}

// AcquireStringBuilder get strings.Builder
func AcquireStringBuilder() *strings.Builder {
	if Pools.Strings == nil {
		return &strings.Builder{}
	}
	return Pools.Strings.Get()
}

// ReleaseStringBuilder releases strings.Builder back to pool
func ReleaseStringBuilder(sb *strings.Builder) {
	if Pools.Strings != nil && sb != nil {
		Pools.Strings.Put(sb)
	}
}

// AcquireBuffer get bytes.Buffer
func AcquireBuffer() *bytes.Buffer {
	if Pools.Buffers == nil {
		return &bytes.Buffer{}
	}
	return Pools.Buffers.Get()
}

// ReleaseBuffer releases bytes.Buffer back to pool
func ReleaseBuffer(b *bytes.Buffer) {
	if Pools.Buffers != nil && b != nil {
		Pools.Buffers.Put(b)
	}
}

// AcquireMap get map
func AcquireMap() map[string]interface{} {
	if Pools.Maps == nil {
		return make(map[string]interface{})
	}
	return Pools.Maps.Get()
}

// ReleaseMap releases map back to pool
func ReleaseMap(m map[string]interface{}) {
	if Pools.Maps != nil && m != nil {
		Pools.Maps.Put(m)
	}
}

// AcquireSlice get byte slice
func AcquireSlice() []byte {
	if Pools.Slices == nil {
		return make([]byte, 0, 4096)
	}
	return Pools.Slices.Get()
}

// ReleaseSlice releases byte slice back to pool
func ReleaseSlice(s []byte) {
	if Pools.Slices != nil && s != nil {
		Pools.Slices.Put(s)
	}
}
