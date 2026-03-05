// Package core 提供高性能内存池和对象复用
// 用于减少 GC 压力和内存分配
package core

import (
	"sync"
)

// FeatureVectorPool 特征向量内存池
type FeatureVectorPool struct {
	pool sync.Pool
}

// NewFeatureVectorPool 创建新的特征向量池
func NewFeatureVectorPool() *FeatureVectorPool {
	return &FeatureVectorPool{
		pool: sync.Pool{
			New: func() interface{} {
				return NewFeatureVector()
			},
		},
	}
}

// Get 从池中获取特征向量
func (p *FeatureVectorPool) Get() *FeatureVector {
	fv := p.pool.Get().(*FeatureVector)
	// 重置状态
	for k := range fv.Features {
		delete(fv.Features, k)
	}
	for k := range fv.Metadata {
		delete(fv.Metadata, k)
	}
	return fv
}

// Put 将特征向量放回池中
func (p *FeatureVectorPool) Put(fv *FeatureVector) {
	if fv == nil {
		return
	}
	p.pool.Put(fv)
}

// HTTPHeadersPool HTTP 头内存池
type HTTPHeadersPool struct {
	pool sync.Pool
}

// NewHTTPHeadersPool 创建新的 HTTP 头池
func NewHTTPHeadersPool() *HTTPHeadersPool {
	return &HTTPHeadersPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &HTTPHeaders{
					Custom: make(map[string]string, 8),
				}
			},
		},
	}
}

// Get 从池中获取 HTTP 头
func (p *HTTPHeadersPool) Get() *HTTPHeaders {
	h := p.pool.Get().(*HTTPHeaders)
	// 重置状态
	h.Accept = ""
	h.AcceptLanguage = ""
	h.AcceptEncoding = ""
	h.UserAgent = ""
	h.SecFetchSite = ""
	h.SecFetchMode = ""
	h.SecFetchUser = ""
	h.SecFetchDest = ""
	h.SecCHUA = ""
	h.SecCHUAMobile = ""
	h.SecCHUAPlatform = ""
	h.UpgradeInsecureRequests = ""
	for k := range h.Custom {
		delete(h.Custom, k)
	}
	return h
}

// Put 将 HTTP 头放回池中
func (p *HTTPHeadersPool) Put(h *HTTPHeaders) {
	if h == nil {
		return
	}
	p.pool.Put(h)
}

// StringPool 字符串内存池（用于频繁创建的字符串）
type StringPool struct {
	pools [32]sync.Pool // 按长度分桶，最大 32 字节
}

// NewStringPool 创建新的字符串池
func NewStringPool() *StringPool {
	sp := &StringPool{}
	for i := range sp.pools {
		size := (i + 1) * 64 // 64, 128, 192, ... 2048
		sp.pools[i] = sync.Pool{
			New: func() interface{} {
				b := make([]byte, size)
				return &b
			},
		}
	}
	return sp
}

// GetBuffer 获取缓冲区
func (p *StringPool) GetBuffer(size int) []byte {
	idx := (size-1)/64
	if idx < 0 {
		idx = 0
	}
	if idx >= len(p.pools) {
		return make([]byte, size)
	}
	
	buf := p.pools[idx].Get().(*[]byte)
	if len(*buf) < size {
		// 如果缓冲区不够大，创建新的
		return make([]byte, size)
	}
	return (*buf)[:size]
}

// PutBuffer 归还缓冲区
func (p *StringPool) PutBuffer(buf []byte) {
	idx := (len(buf)-1)/64
	if idx < 0 || idx >= len(p.pools) {
		return // 太大或太小，直接丢弃
	}
	p.pools[idx].Put(&buf)
}

// MapPool map 内存池
type MapPool struct {
	pool sync.Pool
}

// NewMapPool 创建新的 map 池
func NewMapPool() *MapPool {
	return &MapPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make(map[string]string, 16)
			},
		},
	}
}

// Get 从池中获取 map
func (p *MapPool) Get() map[string]string {
	m := p.pool.Get().(map[string]string)
	// 清空 map
	for k := range m {
		delete(m, k)
	}
	return m
}

// Put 将 map 放回池中
func (p *MapPool) Put(m map[string]string) {
	if m == nil {
		return
	}
	p.pool.Put(m)
}

// SlicePool 切片内存池
type SlicePool struct {
	pools [10]sync.Pool // 不同容量的切片池
}

// NewSlicePool 创建新的切片池
func NewSlicePool() *SlicePool {
	sp := &SlicePool{}
	capacities := []int{8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096}
	for i, cap := range capacities {
		size := cap
		sp.pools[i] = sync.Pool{
			New: func() interface{} {
				s := make([]uint16, 0, size)
				return &s
			},
		}
	}
	return sp
}

// GetUint16Slice 获取 uint16 切片
func (p *SlicePool) GetUint16Slice(capacity int) []uint16 {
	idx := p.selectPool(capacity)
	if idx < 0 {
		return make([]uint16, 0, capacity)
	}
	
	s := p.pools[idx].Get().(*[]uint16)
	return (*s)[:0]
}

// PutUint16Slice 归还 uint16 切片
func (p *SlicePool) PutUint16Slice(s []uint16) {
	idx := p.selectPool(cap(s))
	if idx < 0 {
		return
	}
	p.pools[idx].Put(&s)
}

// selectPool 选择合适的池
func (p *SlicePool) selectPool(capacity int) int {
	capacities := []int{8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096}
	for i, cap := range capacities {
		if capacity <= cap {
			return i
		}
	}
	return -1 // 太大，不归还给池
}

// GlobalPools 全局内存池实例
var GlobalPools = &GlobalPoolSet{
	FeatureVectors: NewFeatureVectorPool(),
	HTTPHeaders:    NewHTTPHeadersPool(),
	Strings:        NewStringPool(),
	Maps:           NewMapPool(),
	Slices:         NewSlicePool(),
}

// GlobalPoolSet 全局内存池集合
type GlobalPoolSet struct {
	FeatureVectors *FeatureVectorPool
	HTTPHeaders    *HTTPHeadersPool
	Strings        *StringPool
	Maps           *MapPool
	Slices         *SlicePool
}

// Reset 清空所有池（主要用于测试）
func (gps *GlobalPoolSet) Reset() {
	// sync.Pool 没有清空方法，只能通过重新创建
	GlobalPools = &GlobalPoolSet{
		FeatureVectors: NewFeatureVectorPool(),
		HTTPHeaders:    NewHTTPHeadersPool(),
		Strings:        NewStringPool(),
		Maps:           NewMapPool(),
		Slices:         NewSlicePool(),
	}
}

// PoolStats 内存池统计
type PoolStats struct {
	Name         string
	HitRate      float64
	ObjectsInUse int
	TotalGets    int64
	TotalPuts    int64
}

// GetStats 获取内存池统计
func (gps *GlobalPoolSet) GetStats() []PoolStats {
	// 注意：sync.Pool 不提供直接的统计信息
	// 这里返回占位符，实际可以通过包装器实现
	return []PoolStats{
		{Name: "FeatureVectors", HitRate: 0},
		{Name: "HTTPHeaders", HitRate: 0},
		{Name: "Maps", HitRate: 0},
	}
}

// PooledFeatureVector 池化的特征向量操作
type PooledFeatureVector struct {
	*FeatureVector
	pool *FeatureVectorPool
}

// NewPooledFeatureVector 从池创建特征向量
func NewPooledFeatureVector() *PooledFeatureVector {
	return &PooledFeatureVector{
		FeatureVector: GlobalPools.FeatureVectors.Get(),
		pool:          GlobalPools.FeatureVectors,
	}
}

// Release 释放回池
func (p *PooledFeatureVector) Release() {
	if p.pool != nil && p.FeatureVector != nil {
		p.pool.Put(p.FeatureVector)
		p.FeatureVector = nil
	}
}

// PooledHTTPHeaders 池化的 HTTP 头操作
type PooledHTTPHeaders struct {
	*HTTPHeaders
	pool *HTTPHeadersPool
}

// NewPooledHTTPHeaders 从池创建 HTTP 头
func NewPooledHTTPHeaders() *PooledHTTPHeaders {
	return &PooledHTTPHeaders{
		HTTPHeaders: GlobalPools.HTTPHeaders.Get(),
		pool:        GlobalPools.HTTPHeaders,
	}
}

// Release 释放回池
func (p *PooledHTTPHeaders) Release() {
	if p.pool != nil && p.HTTPHeaders != nil {
		p.pool.Put(p.HTTPHeaders)
		p.HTTPHeaders = nil
	}
}
