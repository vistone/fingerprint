package utils

import (
	"strings"
	"sync"

	"github.com/vistone/fingerprint/modules/core/types"
)

// StringBuilderPool reuses string builders.
var StringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// HTTPHeadersPool reuses HTTPHeaders objects.
var HTTPHeadersPool = sync.Pool{
	New: func() interface{} {
		return &types.HTTPHeaders{
			Custom: make(map[string]string, 8),
		}
	},
}

// MapStringStringPool reuses map[string]string objects.
var MapStringStringPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]string, 16)
	},
}

// ByteSlicePool reuses byte slices with 4KB capacity.
var ByteSlicePool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 4096)
		return &buf
	},
}

// SmallByteSlicePool reuses byte slices with 1KB capacity.
var SmallByteSlicePool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 1024)
		return &buf
	},
}

// GetStringBuilder retrieves a string builder from the pool.
func GetStringBuilder() *strings.Builder {
	return StringBuilderPool.Get().(*strings.Builder)
}

// PutStringBuilder resets and returns a string builder to the pool.
func PutStringBuilder(sb *strings.Builder) {
	sb.Reset()
	StringBuilderPool.Put(sb)
}

// GetHTTPHeaders retrieves an HTTPHeaders object from the pool.
func GetHTTPHeaders() *types.HTTPHeaders {
	h := HTTPHeadersPool.Get().(*types.HTTPHeaders)
	// Reset scalar fields.
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
	// Clear custom headers.
	for k := range h.Custom {
		delete(h.Custom, k)
	}
	return h
}

// PutHTTPHeaders resets and returns an HTTPHeaders object to the pool.
func PutHTTPHeaders(h *types.HTTPHeaders) {
	if h == nil {
		return
	}
	// Reset scalar fields.
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
	// Clear custom headers while keeping map capacity.
	for k := range h.Custom {
		delete(h.Custom, k)
	}
	HTTPHeadersPool.Put(h)
}

// GetMapStringString retrieves a map from the pool.
func GetMapStringString() map[string]string {
	m := MapStringStringPool.Get().(map[string]string)
	// Clear map entries before reuse.
	for k := range m {
		delete(m, k)
	}
	return m
}

// PutMapStringString clears and returns a map to the pool.
func PutMapStringString(m map[string]string) {
	if m == nil {
		return
	}
	// Clear map entries while keeping map capacity.
	for k := range m {
		delete(m, k)
	}
	MapStringStringPool.Put(m)
}

// GetByteSlice retrieves a byte slice from the 4KB pool.
func GetByteSlice() []byte {
	buf := *(ByteSlicePool.Get().(*[]byte))
	return buf[:0]
}

// PutByteSlice returns a byte slice to the 4KB pool.
func PutByteSlice(buf []byte) {
	if cap(buf) < 4096 {
		return // Do not return smaller buffers.
	}
	ByteSlicePool.Put(&buf)
}

// GetSmallByteSlice retrieves a byte slice from the 1KB pool.
func GetSmallByteSlice() []byte {
	buf := *(SmallByteSlicePool.Get().(*[]byte))
	return buf[:0]
}

// PutSmallByteSlice returns a byte slice to the 1KB pool.
func PutSmallByteSlice(buf []byte) {
	if cap(buf) < 1024 {
		return
	}
	SmallByteSlicePool.Put(&buf)
}
