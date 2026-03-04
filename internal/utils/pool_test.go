package utils

import (
	"testing"

	"github.com/vistone/fingerprint/types"
)

// TestHTTPHeadersPool 测试 HTTPHeaders 对象池
func TestHTTPHeadersPool(t *testing.T) {
	// 获取对象
	h1 := GetHTTPHeaders()
	if h1 == nil {
		t.Fatal("GetHTTPHeaders() returned nil")
	}

	// 设置一些值
	h1.UserAgent = "test-ua"
	h1.Accept = "test-accept"
	h1.Custom["key"] = "value"

	// 归还对象
	PutHTTPHeaders(h1)

	// 再次获取，应该被重置
	h2 := GetHTTPHeaders()
	if h2.UserAgent != "" {
		t.Errorf("UserAgent not reset: %s", h2.UserAgent)
	}
	if h2.Accept != "" {
		t.Errorf("Accept not reset: %s", h2.Accept)
	}
	if len(h2.Custom) != 0 {
		t.Errorf("Custom not cleared: %v", h2.Custom)
	}
}

// TestMapStringStringPool 测试 map[string]string 对象池
func TestMapStringStringPool(t *testing.T) {
	m1 := GetMapStringString()
	if m1 == nil {
		t.Fatal("GetMapStringString() returned nil")
	}

	m1["key"] = "value"
	m1["key2"] = "value2"

	PutMapStringString(m1)

	m2 := GetMapStringString()
	if len(m2) != 0 {
		t.Errorf("Map not cleared: %v", m2)
	}
}

// TestByteSlicePool 测试字节切片池
func TestByteSlicePool(t *testing.T) {
	// 测试大缓冲区
	buf1 := GetByteSlice()
	if cap(buf1) < 4096 {
		t.Errorf("Buffer capacity too small: %d", cap(buf1))
	}

	buf1 = append(buf1, []byte("test data")...)
	PutByteSlice(buf1)

	buf2 := GetByteSlice()
	if len(buf2) != 0 {
		t.Errorf("Buffer not reset: %d", len(buf2))
	}

	// 测试小缓冲区
	smallBuf := GetSmallByteSlice()
	if cap(smallBuf) < 1024 {
		t.Errorf("Small buffer capacity too small: %d", cap(smallBuf))
	}
	PutSmallByteSlice(smallBuf)
}

// TestStringBuilderPool 测试字符串构建器池
func TestStringBuilderPool(t *testing.T) {
	sb1 := GetStringBuilder()
	sb1.WriteString("test string")

	PutStringBuilder(sb1)

	sb2 := GetStringBuilder()
	if sb2.Len() != 0 {
		t.Errorf("StringBuilder not reset: %d", sb2.Len())
	}
}

// TestPutHTTPHeaders_Nil 测试归还 nil
func TestPutHTTPHeaders_Nil(t *testing.T) {
	// 不应 panic
	PutHTTPHeaders(nil)
}

// TestPutMapStringString_Nil 测试归还 nil map
func TestPutMapStringString_Nil(t *testing.T) {
	// 不应 panic
	PutMapStringString(nil)
}

// BenchmarkHTTPHeadersPool 基准测试 HTTPHeaders 池
func BenchmarkHTTPHeadersPool(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h := GetHTTPHeaders()
			h.UserAgent = "test"
			PutHTTPHeaders(h)
		}
	})
}

// BenchmarkMapStringStringPool 基准测试 map 池
func BenchmarkMapStringStringPool(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m := GetMapStringString()
			m["key"] = "value"
			PutMapStringString(m)
		}
	})
}

// BenchmarkHTTPHeadersNew 基准测试直接创建 HTTPHeaders
func BenchmarkHTTPHeadersNew(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h := &types.HTTPHeaders{
				Custom: make(map[string]string),
			}
			h.UserAgent = "test"
			_ = h
		}
	})
}

// BenchmarkMapStringStringNew 基准测试直接创建 map
func BenchmarkMapStringStringNew(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m := make(map[string]string)
			m["key"] = "value"
			_ = m
		}
	})
}
