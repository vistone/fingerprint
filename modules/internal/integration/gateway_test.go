package integration

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/internal/testhelpers"
	"github.com/vistone/fingerprint/modules/ml"
)

// Mock implementations for testing
type MockTieredCache struct {
	data map[string]interface{}
}

func NewMockTieredCache() *MockTieredCache {
	return &MockTieredCache{
		data: make(map[string]interface{}),
	}
}

func (m *MockTieredCache) Get(key string) (interface{}, bool) {
	val, found := m.data[key]
	return val, found
}

func (m *MockTieredCache) Set(key string, value interface{}, ttl time.Duration) {
	m.data[key] = value
}

type MockCircuitBreaker struct {
	shouldFail bool
}

func (m *MockCircuitBreaker) Execute(fn func() error) error {
	if m.shouldFail {
		return errors.New("circuit breaker open")
	}
	return fn()
}

type MockMetricsCollector struct {
	hits   int
	misses int
	ops    int
}

func (m *MockMetricsCollector) RecordOperation(name string, duration time.Duration, err error) {
	m.ops++
}

func (m *MockMetricsCollector) RecordCacheHit(name string) {
	m.hits++
}

func (m *MockMetricsCollector) RecordCacheMiss(name string) {
	m.misses++
}

func TestNewGatewayConnector(t *testing.T) {
	t.Run("create with nil config", func(t *testing.T) {
		gc, err := NewGatewayConnector(nil)
		testhelpers.AssertNoError(t, err)
		testhelpers.AssertNotNil(t, gc)
	})

	t.Run("create with custom config", func(t *testing.T) {
		config := &GatewayConnectorConfig{
			CacheL1Size:    5000,
			CacheL1TTL:     time.Minute * 5,
			MetricsEnabled: true,
		}
		gc, err := NewGatewayConnector(config)
		testhelpers.AssertNoError(t, err)
		testhelpers.AssertNotNil(t, gc)
	})
}

func TestGatewayConnector_ProcessRequest(t *testing.T) {
	gc, _ := NewGatewayConnector(nil)

	t.Run("process valid request", func(t *testing.T) {
		spec := core.ClientHelloSpec{
			TLSVersion:      0x0303,
			CipherSuites:    []uint16{0x1301, 0x1302, 0x1303},
			Extensions:      []core.TLSExtension{{Type: 0, Data: []byte{}}},
			SupportedCurves: []core.CurveID{23, 24},
			SupportedPoints: []uint8{0},
		}

		result, err := gc.ProcessRequest(context.Background(), spec)
		testhelpers.AssertNoError(t, err)
		testhelpers.AssertNotNil(t, result)
		testhelpers.AssertNotNil(t, result.Result)
		testhelpers.AssertNotNil(t, result.JA3)
		testhelpers.AssertNotNil(t, result.JA4)
	})

	t.Run("cache hit on second request", func(t *testing.T) {
		cache := NewMockTieredCache()
		gc.SetCache(cache)

		metrics := &MockMetricsCollector{}
		gc.SetMetricsCollector(metrics)

		spec := core.ClientHelloSpec{
			TLSVersion:   0x0303,
			CipherSuites: []uint16{0x1301, 0x1302},
		}

		// First request - cache miss
		_, err := gc.ProcessRequest(context.Background(), spec)
		testhelpers.AssertNoError(t, err)
		testhelpers.AssertEqual(t, metrics.misses, 1)
		testhelpers.AssertEqual(t, metrics.hits, 0)

		// Second request - cache hit
		_, err = gc.ProcessRequest(context.Background(), spec)
		testhelpers.AssertNoError(t, err)
		testhelpers.AssertEqual(t, metrics.misses, 1)
		testhelpers.AssertEqual(t, metrics.hits, 1)
	})

	t.Run("circuit breaker blocks request", func(t *testing.T) {
		gc2, _ := NewGatewayConnector(nil)
		breaker := &MockCircuitBreaker{shouldFail: true}
		gc2.SetCircuitBreaker(breaker)

		spec := core.ClientHelloSpec{
			TLSVersion:   0x0303,
			CipherSuites: []uint16{0x1301},
		}

		_, err := gc2.ProcessRequest(context.Background(), spec)
		testhelpers.AssertError(t, err)
	})
}

func TestGatewayConnector_CalculateRiskScore(t *testing.T) {
	gc, _ := NewGatewayConnector(nil)

	tests := []struct {
		name       string
		confidence float64
		cipherLen  int
		wantRisk   float64
	}{
		{
			name:       "high confidence, normal config",
			confidence: 0.95,
			cipherLen:  10,
			wantRisk:   0.0,
		},
		{
			name:       "low confidence",
			confidence: 0.4,
			cipherLen:  10,
			wantRisk:   0.3,
		},
		{
			name:       "very low confidence",
			confidence: 0.2,
			cipherLen:  10,
			wantRisk:   0.6,
		},
		{
			name:       "few cipher suites",
			confidence: 0.9,
			cipherLen:  3,
			wantRisk:   0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ml.ClassificationResult{
				Confidence: tt.confidence,
			}
			spec := core.ClientHelloSpec{
				CipherSuites: make([]uint16, tt.cipherLen),
			}

			risk := gc.calculateRiskScore(result, spec)
			testhelpers.AssertEqual(t, risk >= tt.wantRisk, true)
			testhelpers.AssertEqual(t, risk <= 1.0, true)
		})
	}
}

func TestGatewayConnector_HTTPHandler(t *testing.T) {
	gc, _ := NewGatewayConnector(nil)
	handler := gc.HTTPHandler()

	t.Run("health endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		testhelpers.AssertEqual(t, rr.Code, http.StatusOK)
		testhelpers.AssertEqual(t, rr.Body.String(), `{"status":"healthy"}`)
		testhelpers.AssertEqual(t, rr.Header().Get("Content-Type"), "application/json")
	})

	t.Run("ready endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ready", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		testhelpers.AssertEqual(t, rr.Code, http.StatusOK)
		testhelpers.AssertEqual(t, rr.Body.String(), `{"status":"ready"}`)
	})

	t.Run("unknown endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/unknown", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		testhelpers.AssertEqual(t, rr.Code, http.StatusNotFound)
	})
}

func TestGatewayConnector_Setters(t *testing.T) {
	gc, _ := NewGatewayConnector(nil)

	t.Run("set cache", func(t *testing.T) {
		cache := NewMockTieredCache()
		gc.SetCache(cache)
		testhelpers.AssertNotNil(t, gc.cache)
	})

	t.Run("set circuit breaker", func(t *testing.T) {
		breaker := &MockCircuitBreaker{}
		gc.SetCircuitBreaker(breaker)
		testhelpers.AssertNotNil(t, gc.breaker)
	})

	t.Run("set metrics collector", func(t *testing.T) {
		metrics := &MockMetricsCollector{}
		gc.SetMetricsCollector(metrics)
		testhelpers.AssertNotNil(t, gc.metrics)
	})
}

func TestGatewayConnector_Close(t *testing.T) {
	gc, _ := NewGatewayConnector(nil)

	err := gc.Close()
	testhelpers.AssertNoError(t, err)
}

func TestGenerateCacheKey(t *testing.T) {
	t.Run("same spec generates same key", func(t *testing.T) {
		spec1 := core.ClientHelloSpec{
			TLSVersion:   0x0303,
			CipherSuites: []uint16{0x1301, 0x1302},
		}
		spec2 := core.ClientHelloSpec{
			TLSVersion:   0x0303,
			CipherSuites: []uint16{0x1301, 0x1302},
		}

		key1 := generateCacheKey(spec1)
		key2 := generateCacheKey(spec2)

		testhelpers.AssertEqual(t, key1, key2)
	})

	t.Run("different spec generates different key", func(t *testing.T) {
		spec1 := core.ClientHelloSpec{
			TLSVersion:   0x0303,
			CipherSuites: []uint16{0x1301},
		}
		spec2 := core.ClientHelloSpec{
			TLSVersion:   0x0303,
			CipherSuites: []uint16{0x1302},
		}

		key1 := generateCacheKey(spec1)
		key2 := generateCacheKey(spec2)

		testhelpers.AssertEqual(t, key1 != key2, true)
	})
}

func TestMinFloat64(t *testing.T) {
	tests := []struct {
		a, b, want float64
	}{
		{1.0, 2.0, 1.0},
		{2.0, 1.0, 1.0},
		{1.0, 1.0, 1.0},
		{-1.0, 0.0, -1.0},
	}

	for _, tt := range tests {
		got := minFloat64(tt.a, tt.b)
		testhelpers.AssertEqual(t, got, tt.want)
	}
}

func BenchmarkGatewayConnector_ProcessRequest(b *testing.B) {
	gc, _ := NewGatewayConnector(nil)
	cache := NewMockTieredCache()
	gc.SetCache(cache)

	spec := core.ClientHelloSpec{
		TLSVersion:   0x0303,
		CipherSuites: []uint16{0x1301, 0x1302, 0x1303},
		Extensions:   []core.TLSExtension{{Type: 0}},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gc.ProcessRequest(ctx, spec)
	}
}
