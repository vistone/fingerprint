package gateway

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// mockAnalyzeResponse 创建模拟的 AnalyzeResponse
func mockAnalyzeResponse(id string) *AnalyzeResponse {
	return &AnalyzeResponse{
		FingerprintHash:  id,
		Classification:   nil,
		RiskAssessment:   nil,
		Findings:         nil,
		JA3:              &JA3Info{Hash: "mock_ja3_hash"},
		JA4:              &JA4Info{Fingerprint: "mock_ja4_fp"},
		JA4H:             &JA4HInfo{Fingerprint: "mock_ja4h_fp"},
		DefenseHints:     []string{"hint1", "hint2"},
		Cached:           false,
		CacheTime:        time.Now(),
		ProcessingTimeMs: rand.Int63n(100),
	}
}

// BenchmarkLRUCacheGet 测试缓存读取性能
func BenchmarkLRUCacheGet(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			cache := NewLRUCache(size, 5*time.Minute)

			// 预填充缓存
			for i := 0; i < size; i++ {
				key := fmt.Sprintf("key_%d", i)
				cache.Set(key, mockAnalyzeResponse(key), 0)
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					key := fmt.Sprintf("key_%d", i%size)
					cache.Get(key)
					i++
				}
			})
		})
	}
}

// BenchmarkLRUCacheSet 测试缓存写入性能
func BenchmarkLRUCacheSet(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			cache := NewLRUCache(size, 5*time.Minute)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("key_%d", i)
				cache.Set(key, mockAnalyzeResponse(key), 0)
			}
		})
	}
}

// BenchmarkLRUCacheMixed 测试混合读写性能
func BenchmarkLRUCacheMixed(b *testing.B) {
	cache := NewLRUCache(10000, 5*time.Minute)

	// 预填充 50% 缓存
	for i := 0; i < 5000; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, mockAnalyzeResponse(key), 0)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// 80% 读，20% 写
			if i%5 == 0 {
				// 写操作
				key := fmt.Sprintf("key_%d", rand.Intn(10000))
				cache.Set(key, mockAnalyzeResponse(key), 0)
			} else {
				// 读操作
				key := fmt.Sprintf("key_%d", rand.Intn(10000))
				cache.Get(key)
			}
			i++
		}
	})
}

// BenchmarkFingerprintCache 对比新旧缓存实现
func BenchmarkFingerprintCache(b *testing.B) {
	b.Run("LRUCache", func(b *testing.B) {
		cache := NewFingerprintCache(10000, 5*time.Minute)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("fp_%d", i%10000)
			if i%5 == 0 {
				cache.Set(key, mockAnalyzeResponse(key))
			} else {
				cache.Get(key)
			}
		}
	})
}

// TestCacheHitRate 测试缓存命中率
func TestCacheHitRate(t *testing.T) {
	cache := NewLRUCache(1000, 5*time.Minute)

	// 填充缓存
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, mockAnalyzeResponse(key), 0)
	}

	// 测试 1000 次访问，期望命中率接近 100%
	hits := 0
	misses := 0
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		if _, found := cache.Get(key); found {
			hits++
		} else {
			misses++
		}
	}

	hitRate := float64(hits) / float64(hits+misses)
	if hitRate < 0.99 {
		t.Errorf("Cache hit rate = %.2f%%, want >= 99%%", hitRate*100)
	}

	t.Logf("Cache hit rate: %.2f%% (hits: %d, misses: %d)", hitRate*100, hits, misses)
}

// TestCacheEviction 测试缓存淘汰
func TestCacheEviction(t *testing.T) {
	cache := NewLRUCache(100, 5*time.Minute)

	// 填充超过容量
	for i := 0; i < 200; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, mockAnalyzeResponse(key), 0)
	}

	// 验证缓存大小不超过限制
	if cache.Len() > 100 {
		t.Errorf("Cache size = %d, want <= 100", cache.Len())
	}

	// 验证最旧的项被淘汰（前面插入的应该已经不存在了）
	_, found := cache.Get("key_0")
	if found {
		t.Error("Oldest entry should have been evicted")
	}

	// 验证最新的项存在
	_, found = cache.Get("key_199")
	if !found {
		t.Error("Newest entry should still exist")
	}
}

// TestCacheExpiration 测试缓存过期
func TestCacheExpiration(t *testing.T) {
	cache := NewLRUCache(100, 100*time.Millisecond)

	// 添加项
	cache.Set("key1", mockAnalyzeResponse("key1"), 0)

	// 立即获取应该存在
	if _, found := cache.Get("key1"); !found {
		t.Error("Entry should exist immediately after set")
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 应该过期
	if _, found := cache.Get("key1"); found {
		t.Error("Entry should have expired")
	}
}

// TestCacheThreadSafety 测试并发安全
func TestCacheThreadSafety(t *testing.T) {
	cache := NewLRUCache(1000, 5*time.Minute)
	done := make(chan bool)

	// 启动多个 goroutine 同时读写
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("goroutine_%d_key_%d", id, j)
				cache.Set(key, mockAnalyzeResponse(key), 0)
				cache.Get(key)
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证没有 panic，缓存状态正常
	if cache.Len() == 0 {
		t.Error("Cache should have some entries")
	}
}

// TestCacheStats 测试缓存统计
func TestCacheStats(t *testing.T) {
	cache := NewLRUCache(100, 5*time.Minute)

	// 初始统计应为零
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Error("Initial stats should be zero")
	}

	// 添加并获取
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, mockAnalyzeResponse(key), 0)
	}

	// 获取存在的项（命中）
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Get(key)
	}

	// 获取不存在的项（未命中）
	for i := 0; i < 5; i++ {
		cache.Get("nonexistent_key")
	}

	stats = cache.Stats()
	if stats.Hits != 10 {
		t.Errorf("Hits = %d, want 10", stats.Hits)
	}
	if stats.Misses != 5 {
		t.Errorf("Misses = %d, want 5", stats.Misses)
	}
	if stats.Entries != 10 {
		t.Errorf("Entries = %d, want 10", stats.Entries)
	}

	expectedHitRate := 10.0 / 15.0
	if stats.HitRate != expectedHitRate {
		t.Errorf("HitRate = %v, want %v", stats.HitRate, expectedHitRate)
	}
}

// ExampleLRUCache 使用示例
func ExampleLRUCache() {
	cache := NewLRUCache(1000, 5*time.Minute)

	// 设置值
	cache.Set("user:123", "user data", 0)

	// 获取值
	if data, found := cache.Get("user:123"); found {
		fmt.Printf("Found: %v\n", data)
	}

	// 输出: Found: user data
}

// TestCacheComparisonWithOldImplementation 对比新旧实现
func TestCacheComparisonWithOldImplementation(t *testing.T) {
	// 这个测试用于记录新旧实现的对比
	// 旧实现：FingerprintCache（简单 map + 线性扫描淘汰）
	// 新实现：基于 LRUCache（container/list + map）

	t.Run("MemoryEfficiency", func(t *testing.T) {
		// LRU 实现使用更多内存但提供更好的性能特征
		// 此测试主要作为文档说明
		t.Log("LRU cache uses container/list which has O(1) eviction")
		t.Log("Old implementation used linear scan O(n) for eviction")
	})

	t.Run("TimeComplexity", func(t *testing.T) {
		// Get: O(1)
		// Set: O(1)
		// Eviction: O(1)
		t.Log("All operations are O(1) in LRU cache")
		t.Log("Old implementation eviction was O(n)")
	})
}

// BenchmarkCacheComparison 详细的性能对比
func BenchmarkCacheComparison(b *testing.B) {
	scenarios := []struct {
		name      string
		size      int
		readRatio float64 // 读操作比例
	}{
		{"small_read_heavy", 100, 0.9},
		{"small_write_heavy", 100, 0.1},
		{"large_read_heavy", 10000, 0.9},
		{"large_write_heavy", 10000, 0.1},
		{"balanced", 1000, 0.5},
	}

	for _, sc := range scenarios {
		b.Run(sc.name, func(b *testing.B) {
			cache := NewLRUCache(sc.size, 5*time.Minute)

			// 预填充 50%
			for i := 0; i < sc.size/2; i++ {
				key := fmt.Sprintf("key_%d", i)
				cache.Set(key, mockAnalyzeResponse(key), 0)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("key_%d", rand.Intn(sc.size))
				if rand.Float64() < sc.readRatio {
					cache.Get(key)
				} else {
					cache.Set(key, mockAnalyzeResponse(key), 0)
				}
			}
		})
	}
}
