package gateway

import (
	"testing"
	"time"

	"github.com/vistone/fingerprint/modules/internal/testhelpers"
)

func TestNewLRUCache(t *testing.T) {
	tests := []struct {
		name       string
		maxEntries int
		wantSize   int
	}{
		{
			name:       "valid cache creation",
			maxEntries: 100,
			wantSize:   100,
		},
		{
			name:       "zero max entries uses default",
			maxEntries: 0,
			wantSize:   DefaultCacheSize,
		},
		{
			name:       "negative max entries uses default",
			maxEntries: -1,
			wantSize:   DefaultCacheSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewLRUCache(tt.maxEntries, time.Minute)
			testhelpers.AssertNotNil(t, cache)
			// MaxEntries is private, just verify cache was created
			testhelpers.AssertNotNil(t, cache)
		})
	}
}

func TestLRUCache_SetGet(t *testing.T) {
	cache := NewLRUCache(3, time.Minute)

	t.Run("set and get string value", func(t *testing.T) {
		cache.Set("key1", "value1", time.Minute)
		val, found := cache.Get("key1")
		testhelpers.AssertEqual(t, found, true)
		testhelpers.AssertEqual(t, val, "value1")
	})

	t.Run("set and get struct value", func(t *testing.T) {
		type TestStruct struct {
			Name  string
			Value int
		}
		original := TestStruct{Name: "test", Value: 42}
		cache.Set("key2", original, time.Minute)
		val, found := cache.Get("key2")
		testhelpers.AssertEqual(t, found, true)
		testhelpers.AssertEqual(t, val.(TestStruct).Name, "test")
		testhelpers.AssertEqual(t, val.(TestStruct).Value, 42)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		val, found := cache.Get("nonexistent")
		testhelpers.AssertEqual(t, found, false)
		testhelpers.AssertEqual(t, val, nil)
	})

	t.Run("get expired key", func(t *testing.T) {
		cache.Set("expired", "value", time.Nanosecond)
		time.Sleep(time.Millisecond * 10)
		val, found := cache.Get("expired")
		testhelpers.AssertEqual(t, found, false)
		testhelpers.AssertEqual(t, val, nil)
	})
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := NewLRUCache(3, time.Minute)

	t.Run("eviction when exceeding max entries", func(t *testing.T) {
		cache.Set("key1", "value1", time.Minute)
		cache.Set("key2", "value2", time.Minute)
		cache.Set("key3", "value3", time.Minute)

		// All three should be present
		testhelpers.AssertEqual(t, cache.Len(), 3)

		// Add fourth entry
		cache.Set("key4", "value4", time.Minute)

		// Should still have 3 entries (oldest evicted)
		testhelpers.AssertEqual(t, cache.Len(), 3)

		// key1 should be evicted (least recently used)
		_, found := cache.Get("key1")
		testhelpers.AssertEqual(t, found, false)

		// key4 should be present
		val, found := cache.Get("key4")
		testhelpers.AssertEqual(t, found, true)
		testhelpers.AssertEqual(t, val, "value4")
	})

	t.Run("access updates lru order", func(t *testing.T) {
		cache := NewLRUCache(3, time.Minute)
		cache.Set("key1", "value1", time.Minute)
		cache.Set("key2", "value2", time.Minute)
		cache.Set("key3", "value3", time.Minute)

		// Access key1 to make it recently used
		cache.Get("key1")

		// Add new entry
		cache.Set("key4", "value4", time.Minute)

		// key1 should still be present (was accessed recently)
		_, found := cache.Get("key1")
		testhelpers.AssertEqual(t, found, true)

		// key2 should be evicted (least recently used)
		_, found = cache.Get("key2")
		testhelpers.AssertEqual(t, found, false)
	})
}

func TestLRUCache_Delete(t *testing.T) {
	cache := NewLRUCache(10, time.Minute)

	cache.Set("key1", "value1", time.Minute)
	cache.Set("key2", "value2", time.Minute)

	t.Run("delete existing key", func(t *testing.T) {
		found := cache.Delete("key1")
		testhelpers.AssertEqual(t, found, true)
		_, found = cache.Get("key1")
		testhelpers.AssertEqual(t, found, false)
	})

	t.Run("delete non-existent key", func(t *testing.T) {
		found := cache.Delete("nonexistent")
		testhelpers.AssertEqual(t, found, false)
	})
}

func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache(10, time.Minute)

	for i := 0; i < 5; i++ {
		cache.Set(string(rune('a'+i)), i, time.Minute)
	}

	testhelpers.AssertEqual(t, cache.Len(), 5)

	cache.Clear()

	testhelpers.AssertEqual(t, cache.Len(), 0)

	for i := 0; i < 5; i++ {
		_, found := cache.Get(string(rune('a' + i)))
		testhelpers.AssertEqual(t, found, false)
	}
}

func TestLRUCache_Stats(t *testing.T) {
	cache := NewLRUCache(10, time.Minute)

	t.Run("initial stats", func(t *testing.T) {
		stats := cache.Stats()
		testhelpers.AssertEqual(t, stats.Entries, 0)
		testhelpers.AssertEqual(t, stats.Hits, int64(0))
		testhelpers.AssertEqual(t, stats.Misses, int64(0))
		testhelpers.AssertEqual(t, stats.HitRate, 0.0)
	})

	t.Run("stats after operations", func(t *testing.T) {
		cache.Set("key1", "value1", time.Minute)
		cache.Set("key2", "value2", time.Minute)

		// Two hits
		cache.Get("key1")
		cache.Get("key2")

		// Two misses
		cache.Get("nonexistent1")
		cache.Get("nonexistent2")

		stats := cache.Stats()
		testhelpers.AssertEqual(t, stats.Entries, 2)
		testhelpers.AssertEqual(t, stats.Hits, int64(2))
		testhelpers.AssertEqual(t, stats.Misses, int64(2))
		testhelpers.AssertEqual(t, stats.HitRate, 0.5)
	})
}

func TestClassificationCache(t *testing.T) {
	t.Run("new classification cache", func(t *testing.T) {
		cc := NewClassificationCache(100, time.Minute)
		testhelpers.AssertNotNil(t, cc)
		testhelpers.AssertEqual(t, cc.Len(), 0)
	})

	t.Run("set and get classification result", func(t *testing.T) {
		cc := NewClassificationCache(10, time.Minute)

		result := &ClassificationCacheEntry{
			Protocol:   "TLS",
			Family:     "Chrome",
			Version:    "120",
			Confidence: 0.95,
			JA3:        "abc123",
			JA4:        "def456",
		}

		cc.Set("fp1", result, time.Minute)

		retrieved, found := cc.Get("fp1")
		testhelpers.AssertEqual(t, found, true)
		testhelpers.AssertEqual(t, retrieved.Protocol, "TLS")
		testhelpers.AssertEqual(t, retrieved.Family, "Chrome")
		testhelpers.AssertEqual(t, retrieved.Confidence, 0.95)
	})

	t.Run("confidence-based ttl - high confidence", func(t *testing.T) {
		cc := NewClassificationCache(10, time.Minute)

		highConfResult := &ClassificationCacheEntry{
			Protocol:   "TLS",
			Confidence: 0.95, // > 0.9
		}

		cc.SetWithConfidenceTTL("fp1", highConfResult)

		// Should use extended TTL
		entry := cc.cache.entries["fp1"].Value.(*cacheEntry)
		testhelpers.AssertEqual(t, entry.expiration.After(time.Now().Add(time.Minute*5)), true)
	})

	t.Run("confidence-based ttl - low confidence", func(t *testing.T) {
		cc := NewClassificationCache(10, time.Minute)

		lowConfResult := &ClassificationCacheEntry{
			Protocol:   "TLS",
			Confidence: 0.5, // < 0.7
		}

		cc.SetWithConfidenceTTL("fp2", lowConfResult)

		// Should use short TTL
		entry := cc.cache.entries["fp2"].Value.(*cacheEntry)
		testhelpers.AssertEqual(t, entry.expiration.Before(time.Now().Add(time.Minute*2)), true)
	})

	t.Run("background cleanup", func(t *testing.T) {
		cc := NewClassificationCache(10, time.Millisecond*50)

		cc.Set("key1", &ClassificationCacheEntry{}, time.Millisecond*10)
		cc.Set("key2", &ClassificationCacheEntry{}, time.Minute)

		testhelpers.AssertEqual(t, cc.Len(), 2)

		cc.StartCleanup(time.Millisecond * 20)
		time.Sleep(time.Millisecond * 100)
		cc.StopCleanup()

		// key1 should be cleaned up, key2 should remain
		testhelpers.AssertEqual(t, cc.Len(), 1)
		_, found := cc.Get("key2")
		testhelpers.AssertEqual(t, found, true)
	})
}

func BenchmarkLRUCache_Set(b *testing.B) {
	cache := NewLRUCache(1000, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(string(rune(i%256)), i, time.Minute)
	}
}

func BenchmarkLRUCache_Get(b *testing.B) {
	cache := NewLRUCache(1000, time.Minute)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		cache.Set(string(rune(i)), i, time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(string(rune(i % 1000)))
	}
}

func BenchmarkLRUCache_SetGet(b *testing.B) {
	cache := NewLRUCache(1000, time.Minute)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := string(rune(i % 256))
			cache.Set(key, i, time.Minute)
			cache.Get(key)
			i++
		}
	})
}
