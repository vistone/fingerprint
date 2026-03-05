package metrics

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vistone/fingerprint/modules/internal/testhelpers"
)

func TestNewCollector(t *testing.T) {
	collector := NewCollector()
	testhelpers.AssertNotNil(t, collector)
}

func TestCollector_Counter(t *testing.T) {
	collector := NewCollector()

	t.Run("create and increment counter", func(t *testing.T) {
		collector.Counter("requests", "Total requests")
		collector.Inc("requests")

		val := collector.GetCounter("requests")
		testhelpers.AssertEqual(t, val, float64(1))
	})

	t.Run("increment by value", func(t *testing.T) {
		collector.Counter("bytes", "Total bytes")
		collector.Add("bytes", 1024)

		val := collector.GetCounter("bytes")
		testhelpers.AssertEqual(t, val, float64(1024))
	})

	t.Run("multiple increments", func(t *testing.T) {
		collector.Counter("ops", "Operations")
		collector.Inc("ops")
		collector.Inc("ops")
		collector.Inc("ops")

		val := collector.GetCounter("ops")
		testhelpers.AssertEqual(t, val, float64(3))
	})

	t.Run("counter with labels", func(t *testing.T) {
		collector.Counter("requests_by_method", "Requests by HTTP method", "method")
		collector.Inc("requests_by_method", "GET")
		collector.Inc("requests_by_method", "GET")
		collector.Inc("requests_by_method", "POST")

		getVal := collector.GetCounter("requests_by_method", "GET")
		postVal := collector.GetCounter("requests_by_method", "POST")

		testhelpers.AssertEqual(t, getVal, float64(2))
		testhelpers.AssertEqual(t, postVal, float64(1))
	})
}

func TestCollector_Gauge(t *testing.T) {
	collector := NewCollector()

	t.Run("create and set gauge", func(t *testing.T) {
		collector.Gauge("connections", "Active connections")
		collector.Set("connections", 42)

		val := collector.GetGauge("connections")
		testhelpers.AssertEqual(t, val, float64(42))
	})

	t.Run("gauge can decrease", func(t *testing.T) {
		collector.Gauge("queue_size", "Queue size")
		collector.Set("queue_size", 100)
		collector.Set("queue_size", 50)

		val := collector.GetGauge("queue_size")
		testhelpers.AssertEqual(t, val, float64(50))
	})
}

func TestCollector_Histogram(t *testing.T) {
	collector := NewCollector()

	t.Run("create and observe histogram", func(t *testing.T) {
		buckets := []float64{0.1, 0.5, 1.0, 2.0, 5.0}
		collector.Histogram("latency", "Request latency", buckets)

		collector.Observe("latency", 0.05)
		collector.Observe("latency", 0.3)
		collector.Observe("latency", 1.5)
		collector.Observe("latency", 3.0)

		// Verify observations were recorded
		count := collector.GetHistogramCount("latency")
		sum := collector.GetHistogramSum("latency")

		testhelpers.AssertEqual(t, count, uint64(4))
		testhelpers.AssertEqual(t, sum > 0, true)
	})
}

func TestCollector_Summary(t *testing.T) {
	collector := NewCollector()

	t.Run("create and observe summary", func(t *testing.T) {
		collector.Summary("response_time", "Response time summary")

		collector.ObserveSummary("response_time", 0.1)
		collector.ObserveSummary("response_time", 0.2)
		collector.ObserveSummary("response_time", 0.3)

		count := collector.GetSummaryCount("response_time")
		testhelpers.AssertEqual(t, count, uint64(3))
	})
}

func TestCollector_PrometheusExport(t *testing.T) {
	collector := NewCollector()

	collector.Counter("requests", "Total requests")
	collector.Inc("requests")
	collector.Inc("requests")

	collector.Gauge("active", "Active connections")
	collector.Set("active", 5)

	export := collector.PrometheusExport()

	t.Run("export contains counter", func(t *testing.T) {
		testhelpers.AssertEqual(t, strings.Contains(export, "requests"), true)
		testhelpers.AssertEqual(t, strings.Contains(export, "2"), true)
	})

	t.Run("export contains gauge", func(t *testing.T) {
		testhelpers.AssertEqual(t, strings.Contains(export, "active"), true)
		testhelpers.AssertEqual(t, strings.Contains(export, "5"), true)
	})
}

func TestCollector_Reset(t *testing.T) {
	collector := NewCollector()

	collector.Counter("requests", "Total requests")
	collector.Inc("requests")
	collector.Inc("requests")

	collector.Gauge("active", "Active connections")
	collector.Set("active", 5)

	testhelpers.AssertEqual(t, collector.GetCounter("requests"), float64(2))
	testhelpers.AssertEqual(t, collector.GetGauge("active"), float64(5))

	collector.Reset()

	testhelpers.AssertEqual(t, collector.GetCounter("requests"), float64(0))
	// Gauge may or may not be reset depending on implementation
}

func TestOperationTimer(t *testing.T) {
	collector := NewCollector()
	buckets := []float64{0.001, 0.01, 0.1, 1.0}
	collector.Histogram("operation_duration", "Operation duration", buckets)

	t.Run("time operation", func(t *testing.T) {
		timer := collector.StartTimer("operation_duration")
		time.Sleep(time.Millisecond * 10)
		timer.ObserveDuration()

		count := collector.GetHistogramCount("operation_duration")
		testhelpers.AssertEqual(t, count, uint64(1))
	})

	t.Run("time operation with defer", func(t *testing.T) {
		func() {
			defer collector.Time("operation_duration")()
			time.Sleep(time.Millisecond * 5)
		}()

		count := collector.GetHistogramCount("operation_duration")
		testhelpers.AssertEqual(t, count, uint64(2))
	})
}

func TestCollector_ConcurrentAccess(t *testing.T) {
	collector := NewCollector()

	collector.Counter("counter", "Test counter")
	collector.Gauge("gauge", "Test gauge")

	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrent increments
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				collector.Inc("counter")
				collector.Set("gauge", float64(j))
			}
		}()
	}

	wg.Wait()

	testhelpers.AssertEqual(t, collector.GetCounter("counter"), float64(numGoroutines*100))
}

func TestFingerprintMetrics(t *testing.T) {
	fm := NewFingerprintMetrics()
	testhelpers.AssertNotNil(t, fm)

	t.Run("record classification", func(t *testing.T) {
		fm.RecordClassification("Chrome", 0.95, time.Millisecond*10)
		fm.RecordClassification("Firefox", 0.87, time.Millisecond*15)

		stats := fm.GetStats()
		testhelpers.AssertEqual(t, stats.TotalClassifications, int64(2))
		testhelpers.AssertEqual(t, stats.AverageConfidence > 0, true)
	})

	t.Run("record cache hit", func(t *testing.T) {
		fm.RecordCacheHit()
		fm.RecordCacheHit()

		stats := fm.GetStats()
		testhelpers.AssertEqual(t, stats.CacheHits, int64(2))
	})

	t.Run("record cache miss", func(t *testing.T) {
		fm.RecordCacheMiss()

		stats := fm.GetStats()
		testhelpers.AssertEqual(t, stats.CacheMisses, int64(1))
		testhelpers.AssertEqual(t, stats.CacheHitRate < 1.0, true)
	})

	t.Run("get cache hit rate", func(t *testing.T) {
		fm2 := NewFingerprintMetrics()
		fm2.RecordCacheHit()
		fm2.RecordCacheHit()
		fm2.RecordCacheMiss()

		rate := fm2.GetCacheHitRate()
		testhelpers.AssertEqual(t, rate, 2.0/3.0)
	})
}

func BenchmarkCollector_Inc(b *testing.B) {
	collector := NewCollector()
	collector.Counter("bench", "Benchmark counter")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			collector.Inc("bench")
		}
	})
}

func BenchmarkCollector_Observe(b *testing.B) {
	collector := NewCollector()
	buckets := []float64{0.1, 0.5, 1.0, 2.0, 5.0}
	collector.Histogram("bench", "Benchmark histogram", buckets)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0.0
		for pb.Next() {
			collector.Observe("bench", i)
			i += 0.01
		}
	})
}
