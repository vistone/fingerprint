package utils

import (
	"strings"
	"sync"

	"github.com/vistone/fingerprint/modules/core/types"
)

// StringBuilderPool 字符串构建器池
var StringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// HTTPHeadersPool HTTPHeaders 对象池
var HTTPHeadersPool = sync.Pool{
	New: func() interface{} {
		return &types.HTTPHeaders{
			Custom: make(map[string]string, 8),
		}
	},
}

// MapStringStringPool map[string]string 对象池
var MapStringStringPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]string, 16)
	},
}

// ByteSlicePool 字节切片池 (4KB)
var ByteSlicePool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 4096)
		return &buf
	},
}

// SmallByteSlicePool 小字节切片池 (1KB)
var SmallByteSlicePool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 1024)
		return &buf
	},
}

// GetStringBuilder 获取字符串构建器
func GetStringBuilder() *strings.Builder {
	return StringBuilderPool.Get().(*strings.Builder)
}

// PutStringBuilder 归还字符串构建器
func PutStringBuilder(sb *strings.Builder) {
	sb.Reset()
	StringBuilderPool.Put(sb)
}

// GetHTTPHeaders 获取 HTTPHeaders
func GetHTTPHeaders() *types.HTTPHeaders {
	h := HTTPHeadersPool.Get().(*types.HTTPHeaders)
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
	// 清空 Custom map
	for k := range h.Custom {
		delete(h.Custom, k)
	}
	return h
}

// PutHTTPHeaders 归还 HTTPHeaders
func PutHTTPHeaders(h *types.HTTPHeaders) {
	if h == nil {
		return
	}
	// 重置所有字段
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
	// 清空 Custom map 但保留容量
	for k := range h.Custom {
		delete(h.Custom, k)
	}
	HTTPHeadersPool.Put(h)
}

// GetMapStringString 获取 map[string]string
func GetMapStringString() map[string]string {
	m := MapStringStringPool.Get().(map[string]string)
	// 清空 map
	for k := range m {
		delete(m, k)
	}
	return m
}

// PutMapStringString 归还 map[string]string
func PutMapStringString(m map[string]string) {
	if m == nil {
		return
	}
	// 清空 map 但保留容量
	for k := range m {
		delete(m, k)
	}
	MapStringStringPool.Put(m)
}

// GetByteSlice 获取字节切片 (4KB)
func GetByteSlice() []byte {
	buf := *(ByteSlicePool.Get().(*[]byte))
	return buf[:0]
}

// PutByteSlice 归还字节切片
func PutByteSlice(buf []byte) {
	if cap(buf) < 4096 {
		return // 不归还小缓冲区
	}
	ByteSlicePool.Put(&buf)
}

// GetSmallByteSlice 获取小字节切片 (1KB)
func GetSmallByteSlice() []byte {
	buf := *(SmallByteSlicePool.Get().(*[]byte))
	return buf[:0]
}

// PutSmallByteSlice 归还小字节切片
func PutSmallByteSlice(buf []byte) {
	if cap(buf) < 1024 {
		return
	}
	SmallByteSlicePool.Put(&buf)
}
