package utils

import (
	"strings"
	"sync"

	"github.com/vistone/fingerprint/modules/core/types"
)

// translated comment
var StringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// translated comment
var HTTPHeadersPool = sync.Pool{
	New: func() interface{} {
		return &types.HTTPHeaders{
			Custom: make(map[string]string, 8),
		}
	},
}

// translated comment
var MapStringStringPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]string, 16)
	},
}

// translated comment
var ByteSlicePool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 4096)
		return &buf
	},
}

// translated comment
var SmallByteSlicePool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 1024)
		return &buf
	},
}

// translated comment
func GetStringBuilder() *strings.Builder {
	return StringBuilderPool.Get().(*strings.Builder)
}

// translated comment
func PutStringBuilder(sb *strings.Builder) {
	sb.Reset()
	StringBuilderPool.Put(sb)
}

// translated comment
func GetHTTPHeaders() *types.HTTPHeaders {
	h := HTTPHeadersPool.Get().(*types.HTTPHeaders)
	// translated comment
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
	// translated comment
	for k := range h.Custom {
		delete(h.Custom, k)
	}
	return h
}

// translated comment
func PutHTTPHeaders(h *types.HTTPHeaders) {
	if h == nil {
		return
	}
	// translated comment
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
	// translated comment
	for k := range h.Custom {
		delete(h.Custom, k)
	}
	HTTPHeadersPool.Put(h)
}

// translated comment
func GetMapStringString() map[string]string {
	m := MapStringStringPool.Get().(map[string]string)
	// translated comment
	for k := range m {
		delete(m, k)
	}
	return m
}

// translated comment
func PutMapStringString(m map[string]string) {
	if m == nil {
		return
	}
	// translated comment
	for k := range m {
		delete(m, k)
	}
	MapStringStringPool.Put(m)
}

// translated comment
func GetByteSlice() []byte {
	buf := *(ByteSlicePool.Get().(*[]byte))
	return buf[:0]
}

// translated comment
func PutByteSlice(buf []byte) {
	if cap(buf) < 4096 {
		return // translated comment
	}
	ByteSlicePool.Put(&buf)
}

// translated comment
func GetSmallByteSlice() []byte {
	buf := *(SmallByteSlicePool.Get().(*[]byte))
	return buf[:0]
}

// translated comment
func PutSmallByteSlice(buf []byte) {
	if cap(buf) < 1024 {
		return
	}
	SmallByteSlicePool.Put(&buf)
}
