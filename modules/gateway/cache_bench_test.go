package gateway

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// Translated comment
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

// Translated comment
func BenchmarkLRUCacheGet(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			cache := NewLRUCache(size, 5*time.Minute)

			// Translated comment
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

// Translated comment
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

// Translated comment
func BenchmarkLRUCacheMixed(b *testing.B) {
	cache := NewLRUCache(10000, 5*time.Minute)

	// Translated comment
	for i := 0; i < 5000; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, mockAnalyzeResponse(key), 0)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Translated comment
			if i%5 == 0 {
				// Translated comment
				key := fmt.Sprintf("key_%d", rand.Intn(10000))
				cache.Set(key, mockAnalyzeResponse(key), 0)
			} else {
				// Translated comment
				key := fmt.Sprintf("key_%d", rand.Intn(10000))
				cache.Get(key)
			}
			i++
		}
	})
}

// Translated comment
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

// Translated comment
func TestCacheHitRate(t *testing.T) {
	cache := NewLRUCache(1000, 5*time.Minute)

	// Translated comment
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, mockAnalyzeResponse(key), 0)
	}

	// Translated comment
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

// Translated comment
func TestCacheEviction(t *testing.T) {
	cache := NewLRUCache(100, 5*time.Minute)

	// Translated comment
	for i := 0; i < 200; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, mockAnalyzeResponse(key), 0)
	}

	// Translated comment
	if cache.Len() > 100 {
		t.Errorf("Cache size = %d, want <= 100", cache.Len())
	}

	// Translated comment
	_, found := cache.Get("key_0")
	if found {
		t.Error("Oldest entry should have been evicted")
	}

	// Translated comment
	_, found = cache.Get("key_199")
	if !found {
		t.Error("Newest entry should still exist")
	}
}

// Translated comment
func TestCacheExpiration(t *testing.T) {
	cache := NewLRUCache(100, 100*time.Millisecond)

	// Translated comment
	cache.Set("key1", mockAnalyzeResponse("key1"), 0)

	// Translated comment
	if _, found := cache.Get("key1"); !found {
		t.Error("Entry should exist immediately after set")
	}

	// Translated comment
	time.Sleep(150 * time.Millisecond)

	// Translated comment
	if _, found := cache.Get("key1"); found {
		t.Error("Entry should have expired")
	}
}

// Translated comment
func TestCacheThreadSafety(t *testing.T) {
	cache := NewLRUCache(1000, 5*time.Minute)
	done := make(chan bool)

	// Translated comment
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

	// Translated comment
	for i := 0; i < 10; i++ {
		<-done
	}

	// Translated comment
	if cache.Len() == 0 {
		t.Error("Cache should have some entries")
	}
}

// Translated comment
func TestCacheStats(t *testing.T) {
	cache := NewLRUCache(100, 5*time.Minute)

	// Translated comment
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Error("Initial stats should be zero")
	}

	// Translated comment
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, mockAnalyzeResponse(key), 0)
	}

	// Translated comment
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Get(key)
	}

	// Translated comment
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

// Translated comment
func ExampleLRUCache() {
	cache := NewLRUCache(1000, 5*time.Minute)

	// Translated comment
	cache.Set("user:123", "user data", 0)

	// Translated comment
	if data, found := cache.Get("user:123"); found {
		fmt.Printf("Found: %v\n", data)
	}

	// Translated comment
}

// Translated comment
func TestCacheComparisonWithOldImplementation(t *testing.T) {
	// Translated comment
	// Translated comment
	// Translated comment

	t.Run("MemoryEfficiency", func(t *testing.T) {
		// Translated comment
		// Translated comment
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

// Translated comment
func BenchmarkCacheComparison(b *testing.B) {
	scenarios := []struct {
		name      string
		size      int
		readRatio float64 // Translated comment
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

			// Translated comment
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
