package ml

import (
	"math/rand"
	"sync/atomic"
)

type canaryRouter struct {
	enabled bool
	rate    float64
	backend inferenceBackend
	stats   CanaryStats
}

// CanaryStats tracks gradual rollout routing and fallback behavior.
type CanaryStats struct {
	Enabled       bool    `json:"enabled"`
	CanaryBackend string  `json:"canaryBackend,omitempty"`
	CanaryRate    float64 `json:"canaryRate"`
	TotalRequests int64   `json:"totalRequests"`
	CanaryRouted  int64   `json:"canaryRouted"`
	FallbackCount int64   `json:"fallbackCount"`
}

func newCanaryRouter(config *ServiceConfig, primary inferenceBackend, pipeline *ModelPipeline) *canaryRouter {
	if config == nil || !config.CanaryEnabled {
		return nil
	}

	rate := config.CanaryRate
	if rate <= 0 || rate > 1 {
		return nil
	}

	canaryBackendName := config.CanaryBackend
	if canaryBackendName == "" {
		canaryBackendName = "onnx"
	}
	if canaryBackendName == primary.Name() {
		return nil
	}

	var backend inferenceBackend
	switch canaryBackendName {
	case "onnx":
		onnxBackend := newONNXInferenceBackend(config, pipeline)
		if onnxBackend == nil {
			return nil
		}
		backend = onnxBackend
	case "native":
		backend = &nativeInferenceBackend{pipeline: pipeline}
	default:
		return nil
	}
	if backend == nil {
		return nil
	}

	router := &canaryRouter{
		enabled: true,
		rate:    rate,
		backend: backend,
	}
	router.stats.Enabled = true
	router.stats.CanaryBackend = backend.Name()
	router.stats.CanaryRate = rate

	return router
}

func (c *canaryRouter) pick(primary inferenceBackend) inferenceBackend {
	if c == nil || !c.enabled {
		return primary
	}

	atomic.AddInt64(&c.stats.TotalRequests, 1)
	if rand.Float64() < c.rate {
		atomic.AddInt64(&c.stats.CanaryRouted, 1)
		return c.backend
	}

	return primary
}

func (c *canaryRouter) recordFallback() {
	if c == nil {
		return
	}
	atomic.AddInt64(&c.stats.FallbackCount, 1)
}

func (c *canaryRouter) Snapshot() *CanaryStats {
	if c == nil {
		return nil
	}

	snapshot := CanaryStats{
		Enabled:       c.stats.Enabled,
		CanaryBackend: c.stats.CanaryBackend,
		CanaryRate:    c.stats.CanaryRate,
		TotalRequests: atomic.LoadInt64(&c.stats.TotalRequests),
		CanaryRouted:  atomic.LoadInt64(&c.stats.CanaryRouted),
		FallbackCount: atomic.LoadInt64(&c.stats.FallbackCount),
	}
	return &snapshot
}
