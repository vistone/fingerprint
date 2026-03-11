package utils

import (
	"testing"
	"time"
)

// TestCache_SetGet tests cache set and get
func TestCache_SetGet(t *testing.T) {
	cache := NewCache()

	// Set value
	cache.Set("key1", "value1", 1*time.Minute)

	// Get value
	val, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected to get value")
	}
	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}

	// Get non-existent key
	_, ok = cache.Get("nonexistent")
	if ok {
		t.Error("Expected not to find nonexistent key")
	}
}

// TestCache_GetString tests getting a string
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

	// Get non-existent string
	_, ok = cache.GetString("nonexistent")
	if ok {
		t.Error("Expected not to find nonexistent key")
	}

	// Get non-string value
	cache.Set("int", 123, 1*time.Minute)
	_, ok = cache.GetString("int")
	if ok {
		t.Error("Expected not to get non-string value as string")
	}
}

// TestCache_Expiration tests cache expiration
func TestCache_Expiration(t *testing.T) {
	cache := NewCache()

	// Set a value with short expiration
	cache.Set("short", "value", 1*time.Millisecond)

	// Getting immediately should succeed
	_, ok := cache.Get("short")
	if !ok {
		t.Error("Expected to get value immediately")
	}

	// Wait for expiration
	time.Sleep(50 * time.Millisecond)

	// Getting after expiration should fail
	_, ok = cache.Get("short")
	if ok {
		t.Error("Expected value to be expired")
	}
}

// TestCache_Delete tests deletion
func TestCache_Delete(t *testing.T) {
	cache := NewCache()

	cache.Set("key", "value", 1*time.Minute)
	cache.Delete("key")

	_, ok := cache.Get("key")
	if ok {
		t.Error("Expected key to be deleted")
	}
}

// TestCache_Clear tests clearing
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

// TestLRUCache_SetGet tests LRU cache
func TestLRUCache_SetGet(t *testing.T) {
	cache := NewLRUCache(2)

	cache.Set("key1", "value1", 1*time.Minute)
	cache.Set("key2", "value2", 1*time.Minute)

	// Get value
	val, ok := cache.Get("key1")
	if !ok || val != "value1" {
		t.Error("Expected to get value1")
	}

	// Add a third value, should evict the least recently used key2
	cache.Set("key3", "value3", 1*time.Minute)

	_, ok = cache.Get("key2")
	if ok {
		t.Error("Expected key2 to be evicted")
	}

	// key1 and key3 should still exist
	val, ok = cache.Get("key1")
	if !ok || val != "value1" {
		t.Error("Expected key1 to still exist")
	}

	val, ok = cache.Get("key3")
	if !ok || val != "value3" {
		t.Error("Expected key3 to exist")
	}
}

// TestLRUCache_Expiration tests LRU cache expiration
func TestLRUCache_Expiration(t *testing.T) {
	cache := NewLRUCache(10)

	cache.Set("key", "value", 1*time.Millisecond)

	// Get immediately
	_, ok := cache.Get("key")
	if !ok {
		t.Error("Expected to get value immediately")
	}

	// Wait for expiration
	time.Sleep(50 * time.Millisecond)

	_, ok = cache.Get("key")
	if ok {
		t.Error("Expected value to be expired")
	}
}

// BenchmarkCache_Set benchmarks cache set
func BenchmarkCache_Set(b *testing.B) {
	cache := NewCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", "value", 1*time.Minute)
	}
}

// BenchmarkCache_Get benchmarks cache get
func BenchmarkCache_Get(b *testing.B) {
	cache := NewCache()
	cache.Set("key", "value", 1*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get("key")
	}
}

// BenchmarkLRUCache_Set benchmarks LRU cache set
func BenchmarkLRUCache_Set(b *testing.B) {
	cache := NewLRUCache(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", "value", 1*time.Minute)
	}
}

// BenchmarkLRUCache_Get benchmarks LRU cache get
func BenchmarkLRUCache_Get(b *testing.B) {
	cache := NewLRUCache(1000)
	cache.Set("key", "value", 1*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get("key")
	}
}
