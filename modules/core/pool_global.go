package core

import (
	"bytes"
	"strings"
)

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
