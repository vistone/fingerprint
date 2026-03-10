package utils

import (
	"testing"

	"github.com/vistone/fingerprint/modules/core/types"
)

// translated comment
func TestHTTPHeadersPool(t *testing.T) {
	// translated comment
	h1 := GetHTTPHeaders()
	if h1 == nil {
		t.Fatal("GetHTTPHeaders() returned nil")
	}

	// translated comment
	h1.UserAgent = "test-ua"
	h1.Accept = "test-accept"
	h1.Custom["key"] = "value"

	// translated comment
	PutHTTPHeaders(h1)

	// translated comment
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

// translated comment
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

// translated comment
func TestByteSlicePool(t *testing.T) {
	// translated comment
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

	// translated comment
	smallBuf := GetSmallByteSlice()
	if cap(smallBuf) < 1024 {
		t.Errorf("Small buffer capacity too small: %d", cap(smallBuf))
	}
	PutSmallByteSlice(smallBuf)
}

// translated comment
func TestStringBuilderPool(t *testing.T) {
	sb1 := GetStringBuilder()
	sb1.WriteString("test string")

	PutStringBuilder(sb1)

	sb2 := GetStringBuilder()
	if sb2.Len() != 0 {
		t.Errorf("StringBuilder not reset: %d", sb2.Len())
	}
}

// translated comment
func TestPutHTTPHeaders_Nil(t *testing.T) {
	// translated comment
	PutHTTPHeaders(nil)
}

// translated comment
func TestPutMapStringString_Nil(t *testing.T) {
	// translated comment
	PutMapStringString(nil)
}

// translated comment
func BenchmarkHTTPHeadersPool(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h := GetHTTPHeaders()
			h.UserAgent = "test"
			PutHTTPHeaders(h)
		}
	})
}

// translated comment
func BenchmarkMapStringStringPool(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m := GetMapStringString()
			m["key"] = "value"
			PutMapStringString(m)
		}
	})
}

// translated comment
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

// translated comment
func BenchmarkMapStringStringNew(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m := make(map[string]string)
			m["key"] = "value"
			_ = m
		}
	})
}
