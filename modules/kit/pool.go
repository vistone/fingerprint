package utils

import (
	"strings"
	"sync"

	"github.com/vistone/fingerprint/modules/core/types"
)

// StringBuilderPool is a string builder pool
var StringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// HTTPHeadersPool is an HTTPHeaders object pool
var HTTPHeadersPool = sync.Pool{
	New: func() interface{} {
		return &types.HTTPHeaders{
			Custom: make(map[string]string, 8),
		}
	},
}

// MapStringStringPool is a map[string]string object pool
var MapStringStringPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]string, 16)
	},
}

// ByteSlicePool is a byte slice pool (4KB)
var ByteSlicePool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 4096)
		return &buf
	},
}

// SmallByteSlicePool is a small byte slice pool (1KB)
var SmallByteSlicePool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 1024)
		return &buf
	},
}

// GetStringBuilder retrieves a string builder from the pool
func GetStringBuilder() *strings.Builder {
	return StringBuilderPool.Get().(*strings.Builder)
}

// PutStringBuilder returns a string builder to the pool
func PutStringBuilder(sb *strings.Builder) {
	sb.Reset()
	StringBuilderPool.Put(sb)
}

// GetHTTPHeaders retrieves an HTTPHeaders object from the pool
func GetHTTPHeaders() *types.HTTPHeaders {
	h := HTTPHeadersPool.Get().(*types.HTTPHeaders)
	// Reset state
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
	// Clear the Custom map
}

// PutHTTPHeaders returns an HTTPHeaders object to the pool
func PutHTTPHeaders(h *types.HTTPHeaders) {
	if h == nil {
		return
	}
	// Reset all fields
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
	// Clear the Custom map but retain capacity
	for k := range h.Custom {
		delete(h.Custom, k)
	}
	HTTPHeadersPool.Put(h)
}

// GetMapStringString retrieves a map[string]string from the pool
func GetMapStringString() map[string]string {
	m := MapStringStringPool.Get().(map[string]string)
	// Clear the map
}

// PutMapStringString returns a map[string]string to the pool
func PutMapStringString(m map[string]string) {
	if m == nil {
		return
	}
	// Clear the map but retain capacity
	for k := range m {
		delete(m, k)
	}
	MapStringStringPool.Put(m)
}

// GetByteSlice retrieves a byte slice (4KB) from the pool
func GetByteSlice() []byte {
	buf := *(ByteSlicePool.Get().(*[]byte))
	return buf[:0]
}

// PutByteSlice returns a byte slice to the pool
func PutByteSlice(buf []byte) {
	if cap(buf) < 4096 {
		return // Don't return small buffers
	}
	ByteSlicePool.Put(&buf)
}

// GetSmallByteSlice retrieves a small byte slice (1KB) from the pool
func GetSmallByteSlice() []byte {
	buf := *(SmallByteSlicePool.Get().(*[]byte))
	return buf[:0]
}

// PutSmallByteSlice returns a small byte slice to the pool
func PutSmallByteSlice(buf []byte) {
	if cap(buf) < 1024 {
		return
	}
	SmallByteSlicePool.Put(&buf)
}
