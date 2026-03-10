package utils

import (
	"testing"
	"time"
)

// translated comment
func TestCache_SetGet(t *testing.T) {
	cache := NewCache()

	// translated comment
	cache.Set("key1", "value1", 1*time.Minute)

	// translated comment
	val, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected to get value")
	}
	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}

	// translated comment
	_, ok = cache.Get("nonexistent")
	if ok {
		t.Error("Expected not to find nonexistent key")
	}
}

// translated comment
func TestCache_GetString(t *testing.T) {
	cache := NewCache()

	cache.Set("str", "test string", 1*time.Minute)

	str, ok := cache.GetString("str")
	if !ok {
		t.Error("Expected to get string")
	}
	if str != "test string" {
		t.Errorf("Expected 'test string', got %s", str)
	}

	// translated comment
	_, ok = cache.GetString("nonexistent")
	if ok {
		t.Error("Expected not to find nonexistent key")
	}

	// translated comment
	cache.Set("int", 123, 1*time.Minute)
	_, ok = cache.GetString("int")
	if ok {
		t.Error("Expected not to get non-string value as string")
	}
}

// translated comment
func TestCache_Expiration(t *testing.T) {
	cache := NewCache()

	// translated comment
	cache.Set("short", "value", 1*time.Millisecond)

	// translated comment
	_, ok := cache.Get("short")
	if !ok {
		t.Error("Expected to get value immediately")
	}

	// translated comment
	time.Sleep(50 * time.Millisecond)

	// translated comment
	_, ok = cache.Get("short")
	if ok {
		t.Error("Expected value to be expired")
	}
}

// translated comment
func TestCache_Delete(t *testing.T) {
	cache := NewCache()

	cache.Set("key", "value", 1*time.Minute)
	cache.Delete("key")

	_, ok := cache.Get("key")
	if ok {
		t.Error("Expected key to be deleted")
	}
}

// translated comment
func TestCache_Clear(t *testing.T) {
	cache := NewCache()

	cache.Set("key1", "value1", 1*time.Minute)
	cache.Set("key2", "value2", 1*time.Minute)

	cache.Clear()

	_, ok := cache.Get("key1")
	if ok {
		t.Error("Expected cache to be cleared")
	}
	_, ok = cache.Get("key2")
	if ok {
		t.Error("Expected cache to be cleared")
	}
}

// translated comment
func TestLRUCache_SetGet(t *testing.T) {
	cache := NewLRUCache(2)

	cache.Set("key1", "value1", 1*time.Minute)
	cache.Set("key2", "value2", 1*time.Minute)

	// translated comment
	val, ok := cache.Get("key1")
	if !ok || val != "value1" {
		t.Error("Expected to get value1")
	}

	// translated comment
	cache.Set("key3", "value3", 1*time.Minute)

	_, ok = cache.Get("key2")
	if ok {
		t.Error("Expected key2 to be evicted")
	}

	// translated comment
	val, ok = cache.Get("key1")
	if !ok || val != "value1" {
		t.Error("Expected key1 to still exist")
	}

	val, ok = cache.Get("key3")
	if !ok || val != "value3" {
		t.Error("Expected key3 to exist")
	}
}

// translated comment
func TestLRUCache_Expiration(t *testing.T) {
	cache := NewLRUCache(10)

	cache.Set("key", "value", 1*time.Millisecond)

	// translated comment
	_, ok := cache.Get("key")
	if !ok {
		t.Error("Expected to get value immediately")
	}

	// translated comment
	time.Sleep(50 * time.Millisecond)

	_, ok = cache.Get("key")
	if ok {
		t.Error("Expected value to be expired")
	}
}

// translated comment
func BenchmarkCache_Set(b *testing.B) {
	cache := NewCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", "value", 1*time.Minute)
	}
}

// translated comment
func BenchmarkCache_Get(b *testing.B) {
	cache := NewCache()
	cache.Set("key", "value", 1*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get("key")
	}
}

// translated comment
func BenchmarkLRUCache_Set(b *testing.B) {
	cache := NewLRUCache(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", "value", 1*time.Minute)
	}
}

// translated comment
func BenchmarkLRUCache_Get(b *testing.B) {
	cache := NewLRUCache(1000)
	cache.Set("key", "value", 1*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get("key")
	}
}
