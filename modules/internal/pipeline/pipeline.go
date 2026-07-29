// internal/pipeline/pipeline.go
// Chain-of-responsibility style pipeline framework

package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ========================================================================
// Core interface definitions
// ========================================================================

// Tracer defines a lightweight tracing interface (without OpenTelemetry dependency)
type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, Span)
}

// Span defines the tracing span interface
type Span interface {
	End()
	SetAttributes(attrs ...Attribute)
	RecordError(err error)
}

// Attribute defines a tracing attribute
type Attribute struct {
	Key   string
	Value interface{}
}

// NoOpTracer is a no-op tracer implementation
type NoOpTracer struct{}

func (t NoOpTracer) Start(ctx context.Context, name string) (context.Context, Span) {
	return ctx, NoOpSpan{}
}

// NoOpSpan is a no-op span implementation
type NoOpSpan struct{}

func (s NoOpSpan) End()                             {}
func (s NoOpSpan) SetAttributes(attrs ...Attribute) {}
func (s NoOpSpan) RecordError(err error)            {}

// DefaultTracer returns the default no-op tracer
func DefaultTracer() Tracer {
	return NoOpTracer{}
}

// ========================================================================
// Core interface definitions
// ========================================================================

// Stage represents one pipeline stage
type Stage interface {
	// GetName returns the unique stage name
	GetName() string

	// GetDependencies returns prerequisite stage names
	GetDependencies() []string

	// Execute runs the stage core logic
	Execute(ctx context.Context, data *StageData) error
}

// Middleware defines a decorator-style stage wrapper
// It hooks logic such as logging, metrics, and tracing around stage execution
type Middleware interface {
	Process(ctx context.Context, stageName string, data *StageData, next ExecutionFunc) error
}

// ExecutionFunc is the stage execution function signature
type ExecutionFunc func(ctx context.Context, data *StageData) error

// StageData is the data carried through the pipeline
type StageData struct {
	// Original input data
	Input interface{}

	// Current stage output (input for the next stage)
	Output interface{}

	// Shared context data (read/write across stages)
	Context map[string]interface{}

	// Tracing metadata
	ExecutedAt time.Time
	Duration   time.Duration
	Error      error
}

// ========================================================================
// Pipeline
// ========================================================================

type Pipeline struct {
	stages      []Stage
	middlewares []Middleware
	stageIndex  map[string]int // Stage name -> index for fast dependency lookup

	tracer Tracer
}

// NewPipeline creates a new pipeline
func NewPipeline(tracer Tracer) *Pipeline {
	if tracer == nil {
		tracer = DefaultTracer()
	}
	return &Pipeline{
		stages:      []Stage{},
		middlewares: []Middleware{},
		stageIndex:  make(map[string]int),
		tracer:      tracer,
	}
}

// AddStage appends a stage
// Stage order is execution order, validated against dependencies
func (p *Pipeline) AddStage(stage Stage) *Pipeline {
	p.stages = append(p.stages, stage)
	p.stageIndex[stage.GetName()] = len(p.stages) - 1
	return p
}

// AddMiddleware appends middleware in chaining order
func (p *Pipeline) AddMiddleware(mw Middleware) *Pipeline {
	p.middlewares = append(p.middlewares, mw)
	return p
}

// Validate checks full pipeline consistency
// Checks:
// 1. All declared dependencies exist
// 2. No circular dependency is present
func (p *Pipeline) Validate() error {
	for i, stage := range p.stages {
		stageName := stage.GetName()

		// Check dependencies
		for _, dep := range stage.GetDependencies() {
			depIdx, exists := p.stageIndex[dep]
			if !exists {
				return fmt.Errorf("stage %s depends on %s, but %s not found",
					stageName, dep, dep)
			}

			// Dependencies must execute earlier (smaller index)
			if depIdx >= i {
				return fmt.Errorf("circular dependency detected: %s (index %d) depends on %s (index %d)",
					stageName, i, dep, depIdx)
			}
		}
	}

	return nil
}

// Execute runs the full pipeline
func (p *Pipeline) Execute(ctx context.Context, input interface{}) (*StageData, error) {
	// Validate pipeline
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("pipeline validation failed: %w", err)
	}

	// Initialize stage data
	data := &StageData{
		Input:   input,
		Output:  input, // Initial output equals input
		Context: make(map[string]interface{}),
	}

	// Execute each stage
	_, pipelineSpan := p.tracer.Start(ctx, "Pipeline.Execute")
	defer pipelineSpan.End()

	for i, stage := range p.stages {
		stageName := stage.GetName()

		// Create span for this stage
		stageCtx, stageSpan := p.tracer.Start(ctx, "stage."+stageName)

		startTime := time.Now()

		// Wrap execution with middleware
		if err := p.executeStageWithMiddleware(stageCtx, stage, data); err != nil {
			duration := time.Since(startTime)
			data.Duration = duration
			data.Error = err

			stageSpan.RecordError(err)
			stageSpan.SetAttributes(
				Attribute{Key: "stage.name", Value: stageName},
				Attribute{Key: "stage.index", Value: i},
				Attribute{Key: "duration_ms", Value: duration.Milliseconds()},
				Attribute{Key: "error", Value: true},
			)
			stageSpan.End()

			return data, fmt.Errorf("stage %s failed: %w", stageName, err)
		}

		duration := time.Since(startTime)
		data.ExecutedAt = time.Now()
		data.Duration = duration

		stageSpan.SetAttributes(
			Attribute{Key: "stage.name", Value: stageName},
			Attribute{Key: "stage.index", Value: i},
			Attribute{Key: "duration_ms", Value: duration.Milliseconds()},
			Attribute{Key: "error", Value: false},
		)
		stageSpan.End()
	}

	pipelineSpan.SetAttributes(
		Attribute{Key: "stages.count", Value: len(p.stages)},
		Attribute{Key: "total_duration_ms", Value: data.Duration.Milliseconds()},
	)

	return data, nil
}

// executeStageWithMiddleware wraps one stage with middleware chain
func (p *Pipeline) executeStageWithMiddleware(ctx context.Context, stage Stage, data *StageData) error {
	// Build the innermost execution handler
	var handler ExecutionFunc = func(ctx context.Context, d *StageData) error {
		return stage.Execute(ctx, d)
	}

	// Build the chain in reverse (last added middleware runs first)
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		mw := p.middlewares[i]
		nextHandler := handler
		handler = func(ctx context.Context, d *StageData) error {
			return mw.Process(ctx, stage.GetName(), d, nextHandler)
		}
	}

	// Execute middleware chain
	return handler(ctx, data)
}

// ========================================================================
// Built-in middleware
// ========================================================================

// LoggingMiddleware logs stage lifecycle events
type LoggingMiddleware struct {
	logger Logger
}

// Logger is a local logging interface alias to avoid import cycles
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

func NewLoggingMiddleware(logger Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

func (lm *LoggingMiddleware) Process(ctx context.Context, stageName string, data *StageData, next ExecutionFunc) error {
	lm.logger.Info("stage started", "stage", stageName)

	startTime := time.Now()
	err := next(ctx, data)
	duration := time.Since(startTime)

	if err != nil {
		lm.logger.Error("stage failed",
			"stage", stageName,
			"error", err,
			"duration_ms", duration.Milliseconds(),
		)
		return fmt.Errorf("stage %s: %w", stageName, err)
	}
	lm.logger.Info("stage completed",
		"stage", stageName,
		"duration_ms", duration.Milliseconds(),
	)
	return nil
}

// MetricsRecorder defines the metrics sink interface
type MetricsRecorder interface {
	Record(stage string, duration time.Duration, success bool)
}

// MetricsMiddleware records stage latency and status
type MetricsMiddleware struct {
	metrics MetricsRecorder
}

func NewMetricsMiddleware(metrics MetricsRecorder) *MetricsMiddleware {
	return &MetricsMiddleware{metrics: metrics}
}

func (mm *MetricsMiddleware) Process(ctx context.Context, stageName string, data *StageData, next ExecutionFunc) error {
	startTime := time.Now()
	err := next(ctx, data)
	duration := time.Since(startTime)

	success := err == nil
	mm.metrics.Record(stageName, duration, success)

	if err != nil {
		return fmt.Errorf("stage %s: %w", stageName, err)
	}
	return nil
}

// RecoveryMiddleware catches panics
type RecoveryMiddleware struct {
	handler func(stageName string, recovered interface{})
}

func NewRecoveryMiddleware(handler func(stageName string, recovered interface{})) *RecoveryMiddleware {
	return &RecoveryMiddleware{handler: handler}
}

func (rm *RecoveryMiddleware) Process(ctx context.Context, stageName string, data *StageData, next ExecutionFunc) (err error) {
	defer func() {
		if r := recover(); r != nil {
			rm.handler(stageName, r)
			err = fmt.Errorf("stage %s panicked: %v", stageName, r)
		}
	}()

	return next(ctx, data)
}

// TimeoutMiddleware enforces stage timeout
type TimeoutMiddleware struct {
	timeout time.Duration
}

func NewTimeoutMiddleware(timeout time.Duration) *TimeoutMiddleware {
	return &TimeoutMiddleware{timeout: timeout}
}

func (tm *TimeoutMiddleware) Process(ctx context.Context, stageName string, data *StageData, next ExecutionFunc) error {
	// Create timeout-bound context
	timeoutCtx, cancel := context.WithTimeout(ctx, tm.timeout)
	defer cancel()

	// Execute with timeout-bound context
	done := make(chan error, 1)
	workingData := cloneStageData(data)
	go func() {
		done <- next(timeoutCtx, workingData)
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("stage %s: %w", stageName, err)
		}
		commitStageData(data, workingData)
		return nil
	case <-timeoutCtx.Done():
		return fmt.Errorf("stage %s timeout after %v", stageName, tm.timeout)
	}
}

func cloneStageData(data *StageData) *StageData {
	if data == nil {
		return &StageData{}
	}

	cloned := *data
	if data.Context != nil {
		cloned.Context = make(map[string]interface{}, len(data.Context))
		for key, value := range data.Context {
			cloned.Context[key] = value
		}
	}

	return &cloned
}

func commitStageData(dst, src *StageData) {
	if dst == nil || src == nil {
		return
	}

	dst.Input = src.Input
	dst.Output = src.Output
	dst.ExecutedAt = src.ExecutedAt
	dst.Duration = src.Duration
	dst.Error = src.Error

	if src.Context == nil {
		dst.Context = nil
		return
	}

	dst.Context = make(map[string]interface{}, len(src.Context))
	for key, value := range src.Context {
		dst.Context[key] = value
	}
}

// CachingMiddleware caches stage output
type CachingMiddleware struct {
	cache map[string]interface{}
	mu    sync.RWMutex
}

func NewCachingMiddleware() *CachingMiddleware {
	return &CachingMiddleware{
		cache: make(map[string]interface{}),
	}
}

func (cm *CachingMiddleware) Process(ctx context.Context, stageName string, data *StageData, next ExecutionFunc) error {
	// Generate cache key (hashing avoids large-object keys)
	cacheKey := cm.generateCacheKey(stageName, data.Input)

	cm.mu.RLock()
	if cached, exists := cm.cache[cacheKey]; exists {
		cm.mu.RUnlock()
		data.Output = cached
		return nil
	}
	cm.mu.RUnlock()

	// Execute stage
	if err := next(ctx, data); err != nil {
		return fmt.Errorf("stage %s: %w", stageName, err)
	}

	// Store cache entry (bounded cache size)
	cm.mu.Lock()
	// Simple eviction strategy: clear cache when capacity is exceeded
	if len(cm.cache) >= 10000 {
		cm.cache = make(map[string]interface{})
	}
	cm.cache[cacheKey] = data.Output
	cm.mu.Unlock()

	return nil
}

// generateCacheKey builds a cache key (hash for large inputs)
func (cm *CachingMiddleware) generateCacheKey(stageName string, input interface{}) string {
	// Use pointer address as key basis for object-like inputs
	// For strings and primitive values, use direct value-based keys
	switch v := input.(type) {
	case string:
		// Hash long strings
		if len(v) > 128 {
			return fmt.Sprintf("%s:%x", stageName, hashString(v))
		}
		return fmt.Sprintf("%s:%s", stageName, v)
	case []byte:
		// Hash byte slices
		if len(v) > 128 {
			return fmt.Sprintf("%s:%x", stageName, hashBytes(v))
		}
		return fmt.Sprintf("%s:%x", stageName, v)
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool:
		return fmt.Sprintf("%s:%v", stageName, v)
	default:
		// Use pointer address for other types
		return fmt.Sprintf("%s:%p", stageName, input)
	}
}

// hashString computes a simplified string hash
func hashString(s string) uint64 {
	// Use FNV-1a hash algorithm
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	hash := uint64(offset64)
	for i := 0; i < len(s) && i < 1024; i++ {
		hash ^= uint64(s[i])
		hash *= prime64
	}
	return hash
}

// hashBytes computes a simplified byte-slice hash
func hashBytes(b []byte) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	hash := uint64(offset64)
	limit := len(b)
	if limit > 1024 {
		limit = 1024 // Cap bytes used for hashing
	}
	for i := 0; i < limit; i++ {
		hash ^= uint64(b[i])
		hash *= prime64
	}
	return hash
}

// ========================================================================
// Unit-test helper types
// ========================================================================

// MockStage is a mock stage used in tests
type MockStage struct {
	name         string
	dependencies []string
	fn           func(ctx context.Context, data *StageData) error
}

func (ms *MockStage) GetName() string {
	return ms.name
}

func (ms *MockStage) GetDependencies() []string {
	return ms.dependencies
}

func (ms *MockStage) Execute(ctx context.Context, data *StageData) error {
	return ms.fn(ctx, data)
}

// NewMockStage creates a mock stage
func NewMockStage(name string, deps []string, fn func(ctx context.Context, data *StageData) error) *MockStage {
	return &MockStage{
		name:         name,
		dependencies: deps,
		fn:           fn,
	}
}
