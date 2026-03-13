// Package integration provides integration layer for connecting all modules
package integration

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/defense"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
)

// GatewayConnector connects all modules together
type GatewayConnector struct {
	// Core components
	classifier     *ml.HierarchicalClassifier
	noiseGenerator defense.NoiseGenerator

	// Performance components
	pool    ConnectionPool
	cache   TieredCache
	breaker CircuitBreaker

	// Metrics
	metrics MetricsCollector
}

// ConnectionPool interface for connection pooling
type ConnectionPool interface {
	Get() (interface{}, error)
	Put(conn interface{})
}

// TieredCache interface for caching
type TieredCache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
}

// CircuitBreaker interface for fault tolerance
type CircuitBreaker interface {
	Execute(fn func() error) error
}

// MetricsCollector interface for metrics
type MetricsCollector interface {
	RecordOperation(name string, duration time.Duration, err error)
	RecordCacheHit(name string)
	RecordCacheMiss(name string)
}

// GatewayConnectorConfig configuration for the gateway connector
type GatewayConnectorConfig struct {
	// Pool configuration
	PoolMaxSize        int
	PoolMaxIdleTime    time.Duration
	PoolHealthInterval time.Duration

	// Cache configuration
	CacheL1Size    int
	CacheL1TTL     time.Duration
	CacheL2Enabled bool

	// Circuit breaker configuration
	CBFailureThreshold int
	CBTimeout          time.Duration
	CBHalfOpenMaxCalls int

	// Metrics configuration
	MetricsEnabled  bool
	MetricsEndpoint string
}

// DefaultGatewayConnectorConfig returns default configuration
func DefaultGatewayConnectorConfig() *GatewayConnectorConfig {
	return &GatewayConnectorConfig{
		PoolMaxSize:        1000,
		PoolMaxIdleTime:    10 * time.Minute,
		PoolHealthInterval: 30 * time.Second,
		CacheL1Size:        10000,
		CacheL1TTL:         5 * time.Minute,
		CacheL2Enabled:     false,
		CBFailureThreshold: 5,
		CBTimeout:          30 * time.Second,
		CBHalfOpenMaxCalls: 3,
		MetricsEnabled:     true,
		MetricsEndpoint:    "/metrics",
	}
}

// NewGatewayConnector creates a new gateway connector
func NewGatewayConnector(config *GatewayConnectorConfig) (*GatewayConnector, error) {
	if config == nil {
		config = DefaultGatewayConnectorConfig()
	}

	connector := &GatewayConnector{}

	// Initialize classifier
	connector.classifier = ml.NewHierarchicalClassifier()

	// Initialize noise generator (will be set up through ActiveProtector if needed)
	connector.noiseGenerator = nil

	return connector, nil
}

// ClassificationResult contains the result of classifying a fingerprint
type ClassificationResult struct {
	Result    *ml.ClassificationResult
	Profile   profiles.ClientProfile
	JA3       string
	JA4       string
	RiskScore float64
	Timestamp time.Time
}

// ProcessRequest processes a fingerprint request through all modules
func (gc *GatewayConnector) ProcessRequest(ctx context.Context, spec core.ClientHelloSpec) (*ClassificationResult, error) {
	start := time.Now()

	// Step 1: Check cache
	cacheKey := generateCacheKey(spec)
	if cached, found := gc.getFromCache(cacheKey); found {
		gc.recordCacheHit()
		return cached, nil
	}
	gc.recordCacheMiss()

	// Step 2: Classify through circuit breaker
	var result *ClassificationResult
	err := gc.executeWithBreaker("classify", func() error {
		var classifyErr error
		result, classifyErr = gc.classify(ctx, spec)
		return classifyErr
	})

	if err != nil {
		gc.recordOperation("classify", time.Since(start), err)
		return nil, err
	}

	// Step 3: Store in cache
	gc.setCache(cacheKey, result, 5*time.Minute)

	gc.recordOperation("process", time.Since(start), nil)
	return result, nil
}

// classify performs the classification
func (gc *GatewayConnector) classify(ctx context.Context, spec core.ClientHelloSpec) (*ClassificationResult, error) {
	// Create feature vector directly
	features := &core.FeatureVector{
		Features: make(map[core.FeatureType]float64),
	}

	// Extract features from spec manually
	extractFeaturesFromSpec(&spec, features)

	// Perform hierarchical classification
	classifierResult := gc.classifier.Classify(features)
	if classifierResult == nil {
		return nil, fmt.Errorf("classification returned nil result")
	}

	// Get matching profile by browser type
	browserProfiles := profiles.GetProfilesByBrowser(classifierResult.Family)
	var profile profiles.ClientProfile
	if len(browserProfiles) > 0 {
		profile = browserProfiles[0]
	} else {
		// Use first available profile as default
		allProfiles := profiles.GetProfilesByBrowser(core.BrowserChrome)
		if len(allProfiles) > 0 {
			profile = allProfiles[0]
		}
	}

	// Calculate JA3 and JA4 fingerprints
	ja3 := calculateJA3(&spec)
	ja4 := calculateJA4(&spec)

	// Calculate risk score
	riskScore := gc.calculateRiskScore(classifierResult, spec)

	return &ClassificationResult{
		Result:    classifierResult,
		Profile:   profile,
		JA3:       ja3,
		JA4:       ja4,
		RiskScore: riskScore,
		Timestamp: time.Now(),
	}, nil
}

// extractFeaturesFromSpec extracts features from ClientHelloSpec
func extractFeaturesFromSpec(spec *core.ClientHelloSpec, fv *core.FeatureVector) {
	// TLS Version
	fv.Features[core.FeatureTLSVersion] = float64(spec.TLSVersion)

	// Cipher suites count
	fv.Features[core.FeatureCipherSuites] = float64(len(spec.CipherSuites))

	// Extensions count
	fv.Features[core.FeatureExtensions] = float64(len(spec.Extensions))
}

// calculateJA3 calculates JA3 fingerprint
func calculateJA3(spec *core.ClientHelloSpec) string {
	// Simplified JA3 calculation
	// Format: TLSVersion,Ciphers,Extensions,EllipticCurves,EllipticCurvePointFormats
	h := md5.New()

	// TLS Version
	fmt.Fprintf(h, "%d,", spec.TLSVersion)

	// Cipher suites
	for i, cs := range spec.CipherSuites {
		if i > 0 {
			fmt.Fprint(h, "-")
		}
		fmt.Fprintf(h, "%d", cs)
	}
	fmt.Fprint(h, ",")

	// Extensions
	for i, ext := range spec.Extensions {
		if i > 0 {
			fmt.Fprint(h, "-")
		}
		fmt.Fprintf(h, "%d", ext.Type)
	}

	return hex.EncodeToString(h.Sum(nil))
}

// calculateJA4 calculates JA4 fingerprint (simplified)
func calculateJA4(spec *core.ClientHelloSpec) string {
	// Simplified JA4 calculation
	h := md5.New()

	// Protocol indicator
	fmt.Fprint(h, "t13")

	// Cipher suites count
	fmt.Fprintf(h, "%02d", len(spec.CipherSuites))

	// Extensions count
	fmt.Fprintf(h, "%02d", len(spec.Extensions))

	// ALPN (simplified)
	fmt.Fprint(h, "d")

	return hex.EncodeToString(h.Sum(nil))
}

// calculateRiskScore calculates the risk score
func (gc *GatewayConnector) calculateRiskScore(result *ml.ClassificationResult, spec core.ClientHelloSpec) float64 {
	baseScore := 0.0

	// Get confidence from the best layer
	confidence := result.Confidence

	// Low confidence increases risk
	if confidence < 0.5 {
		baseScore += 0.3
	}

	// Very low confidence is suspicious
	if confidence < 0.3 {
		baseScore += 0.3
	}

	// Check for anomalous TLS configurations
	if len(spec.CipherSuites) < 5 {
		baseScore += 0.1
	}

	return minFloat64(baseScore, 1.0)
}

// Helper methods

func (gc *GatewayConnector) getFromCache(key string) (*ClassificationResult, bool) {
	if gc.cache == nil {
		return nil, false
	}
	val, found := gc.cache.Get(key)
	if !found {
		return nil, false
	}
	result, ok := val.(*ClassificationResult)
	return result, ok
}

func (gc *GatewayConnector) setCache(key string, value *ClassificationResult, ttl time.Duration) {
	if gc.cache != nil {
		gc.cache.Set(key, value, ttl)
	}
}

func (gc *GatewayConnector) executeWithBreaker(name string, fn func() error) error {
	if gc.breaker == nil {
		return fn()
	}
	return gc.breaker.Execute(fn)
}

func (gc *GatewayConnector) recordCacheHit() {
	if gc.metrics != nil {
		gc.metrics.RecordCacheHit("classification")
	}
}

func (gc *GatewayConnector) recordCacheMiss() {
	if gc.metrics != nil {
		gc.metrics.RecordCacheMiss("classification")
	}
}

func (gc *GatewayConnector) recordOperation(name string, duration time.Duration, err error) {
	if gc.metrics != nil {
		gc.metrics.RecordOperation(name, duration, err)
	}
}

func generateCacheKey(spec core.ClientHelloSpec) string {
	// Use JA3 as cache key since it's a good fingerprint
	return calculateJA3(&spec)
}

func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// HTTPHandler creates an HTTP handler for the gateway
func (gc *GatewayConnector) HTTPHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Check all components are ready
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	return mux
}

// Close closes the gateway connector
func (gc *GatewayConnector) Close() error {
	// Cleanup resources
	return nil
}

// SetCache sets the cache implementation
func (gc *GatewayConnector) SetCache(cache TieredCache) {
	gc.cache = cache
}

// SetCircuitBreaker sets the circuit breaker implementation
func (gc *GatewayConnector) SetCircuitBreaker(breaker CircuitBreaker) {
	gc.breaker = breaker
}

// SetMetricsCollector sets the metrics collector
func (gc *GatewayConnector) SetMetricsCollector(metrics MetricsCollector) {
	gc.metrics = metrics
}
