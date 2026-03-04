package utils

import (
	"testing"
	"time"
)

// TestCache_SetGet 测试缓存设置和获取
func TestCache_SetGet(t *testing.T) {
	cache := NewCache()

	// 设置值
	cache.Set("key1", "value1", 1*time.Minute)

	// 获取值
	val, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected to get value")
	}
	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}

	// 获取不存在的键
	_, ok = cache.Get("nonexistent")
	if ok {
		t.Error("Expected not to find nonexistent key")
	}
}

// TestCache_GetString 测试获取字符串
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

	// 获取不存在的字符串
	_, ok = cache.GetString("nonexistent")
	if ok {
		t.Error("Expected not to find nonexistent key")
	}

	// 获取非字符串值
	cache.Set("int", 123, 1*time.Minute)
	_, ok = cache.GetString("int")
	if ok {
		t.Error("Expected not to get non-string value as string")
	}
}

// TestCache_Expiration 测试缓存过期
func TestCache_Expiration(t *testing.T) {
	cache := NewCache()

	// 设置短过期时间的值
	cache.Set("short", "value", 1*time.Millisecond)

	// 立即获取应该成功
	_, ok := cache.Get("short")
	if !ok {
		t.Error("Expected to get value immediately")
	}

	// 等待过期
	time.Sleep(50 * time.Millisecond)

	// 过期后获取应该失败
	_, ok = cache.Get("short")
	if ok {
		t.Error("Expected value to be expired")
	}
}

// TestCache_Delete 测试删除
func TestCache_Delete(t *testing.T) {
	cache := NewCache()

	cache.Set("key", "value", 1*time.Minute)
	cache.Delete("key")

	_, ok := cache.Get("key")
	if ok {
		t.Error("Expected key to be deleted")
	}
}

// TestCache_Clear 测试清空
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

// TestLRUCache_SetGet 测试 LRU 缓存
func TestLRUCache_SetGet(t *testing.T) {
	cache := NewLRUCache(2)

	cache.Set("key1", "value1", 1*time.Minute)
	cache.Set("key2", "value2", 1*time.Minute)

	// 获取值
	val, ok := cache.Get("key1")
	if !ok || val != "value1" {
		t.Error("Expected to get value1")
	}

	// 添加第三个值，应该淘汰最久未使用的 key2
	cache.Set("key3", "value3", 1*time.Minute)

	_, ok = cache.Get("key2")
	if ok {
		t.Error("Expected key2 to be evicted")
	}

	// key1 和 key3 应该还在
	val, ok = cache.Get("key1")
	if !ok || val != "value1" {
		t.Error("Expected key1 to still exist")
	}

	val, ok = cache.Get("key3")
	if !ok || val != "value3" {
		t.Error("Expected key3 to exist")
	}
}

// TestLRUCache_Expiration 测试 LRU 缓存过期
func TestLRUCache_Expiration(t *testing.T) {
	cache := NewLRUCache(10)

	cache.Set("key", "value", 1*time.Millisecond)

	// 立即获取
	_, ok := cache.Get("key")
	if !ok {
		t.Error("Expected to get value immediately")
	}

	// 等待过期
	time.Sleep(50 * time.Millisecond)

	_, ok = cache.Get("key")
	if ok {
		t.Error("Expected value to be expired")
	}
}

// BenchmarkCache_Set 基准测试缓存设置
func BenchmarkCache_Set(b *testing.B) {
	cache := NewCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", "value", 1*time.Minute)
	}
}

// BenchmarkCache_Get 基准测试缓存获取
func BenchmarkCache_Get(b *testing.B) {
	cache := NewCache()
	cache.Set("key", "value", 1*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get("key")
	}
}

// BenchmarkLRUCache_Set 基准测试 LRU 缓存设置
func BenchmarkLRUCache_Set(b *testing.B) {
	cache := NewLRUCache(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", "value", 1*time.Minute)
	}
}

// BenchmarkLRUCache_Get 基准测试 LRU 缓存获取
func BenchmarkLRUCache_Get(b *testing.B) {
	cache := NewLRUCache(1000)
	cache.Set("key", "value", 1*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get("key")
	}
}
