// internal/pipeline/pipeline.go
// translated comment

package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ========================================================================
// translated comment
// ========================================================================

// translated comment
type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, Span)
}

// translated comment
type Span interface {
	End()
	SetAttributes(attrs ...Attribute)
	RecordError(err error)
}

// translated comment
type Attribute struct {
	Key   string
	Value interface{}
}

// translated comment
type NoOpTracer struct{}

func (t NoOpTracer) Start(ctx context.Context, name string) (context.Context, Span) {
	return ctx, NoOpSpan{}
}

// translated comment
type NoOpSpan struct{}

func (s NoOpSpan) End()                             {}
func (s NoOpSpan) SetAttributes(attrs ...Attribute) {}
func (s NoOpSpan) RecordError(err error)            {}

// translated comment
func DefaultTracer() Tracer {
	return NoOpTracer{}
}

// ========================================================================
// translated comment
// ========================================================================

// translated comment
type Stage interface {
	// translated comment
	GetName() string

	// translated comment
	GetDependencies() []string

	// translated comment
	Execute(ctx context.Context, data *StageData) error
}

// translated comment
// translated comment
type Middleware interface {
	Process(ctx context.Context, stageName string, data *StageData, next ExecutionFunc) error
}

// translated comment
type ExecutionFunc func(ctx context.Context, data *StageData) error

// translated comment
type StageData struct {
	// translated comment
	Input interface{}

	// translated comment
	Output interface{}

	// translated comment
	Context map[string]interface{}

	// translated comment
	ExecutedAt time.Time
	Duration   time.Duration
	Error      error
}

// ========================================================================
// translated comment
// ========================================================================

type Pipeline struct {
	stages      []Stage
	middlewares []Middleware
	stageIndex  map[string]int // translated comment

	tracer Tracer
}

// translated comment
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

// translated comment
// translated comment
func (p *Pipeline) AddStage(stage Stage) *Pipeline {
	p.stages = append(p.stages, stage)
	p.stageIndex[stage.GetName()] = len(p.stages) - 1
	return p
}

// translated comment
func (p *Pipeline) AddMiddleware(mw Middleware) *Pipeline {
	p.middlewares = append(p.middlewares, mw)
	return p
}

// translated comment
// translated comment
// translated comment
// translated comment
func (p *Pipeline) Validate() error {
	for i, stage := range p.stages {
		stageName := stage.GetName()

		// translated comment
		for _, dep := range stage.GetDependencies() {
			depIdx, exists := p.stageIndex[dep]
			if !exists {
				return fmt.Errorf("stage %s depends on %s, but %s not found",
					stageName, dep, dep)
			}

			// translated comment
			if depIdx >= i {
				return fmt.Errorf("circular dependency detected: %s (index %d) depends on %s (index %d)",
					stageName, i, dep, depIdx)
			}
		}
	}

	return nil
}

// translated comment
func (p *Pipeline) Execute(ctx context.Context, input interface{}) (*StageData, error) {
	// translated comment
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("pipeline validation failed: %w", err)
	}

	// translated comment
	data := &StageData{
		Input:   input,
		Output:  input, // translated comment
		Context: make(map[string]interface{}),
	}

	// translated comment
	_, pipelineSpan := p.tracer.Start(ctx, "Pipeline.Execute")
	defer pipelineSpan.End()

	for i, stage := range p.stages {
		stageName := stage.GetName()

		// translated comment
		stageCtx, stageSpan := p.tracer.Start(ctx, "stage."+stageName)

		startTime := time.Now()

		// translated comment
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

// translated comment
func (p *Pipeline) executeStageWithMiddleware(ctx context.Context, stage Stage, data *StageData) error {
	// translated comment
	var handler ExecutionFunc = func(ctx context.Context, d *StageData) error {
		return stage.Execute(ctx, d)
	}

	// translated comment
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		mw := p.middlewares[i]
		nextHandler := handler
		handler = func(ctx context.Context, d *StageData) error {
			return mw.Process(ctx, stage.GetName(), d, nextHandler)
		}
	}

	// translated comment
	return handler(ctx, data)
}

// ========================================================================
// translated comment
// ========================================================================

// translated comment
type LoggingMiddleware struct {
	logger Logger
}

// translated comment
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

// translated comment
type MetricsRecorder interface {
	Record(stage string, duration time.Duration, success bool)
}

// translated comment
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

// translated comment
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

// translated comment
type TimeoutMiddleware struct {
	timeout time.Duration
}

func NewTimeoutMiddleware(timeout time.Duration) *TimeoutMiddleware {
	return &TimeoutMiddleware{timeout: timeout}
}

func (tm *TimeoutMiddleware) Process(ctx context.Context, stageName string, data *StageData, next ExecutionFunc) error {
	// translated comment
	timeoutCtx, cancel := context.WithTimeout(ctx, tm.timeout)
	defer cancel()

	// translated comment
	done := make(chan error, 1)
	go func() {
		done <- next(timeoutCtx, data)
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("stage %s: %w", stageName, err)
		}
		return nil
	case <-timeoutCtx.Done():
		return fmt.Errorf("stage %s timeout after %v", stageName, tm.timeout)
	}
}

// translated comment
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
	// translated comment
	cacheKey := cm.generateCacheKey(stageName, data.Input)

	cm.mu.RLock()
	if cached, exists := cm.cache[cacheKey]; exists {
		cm.mu.RUnlock()
		data.Output = cached
		return nil
	}
	cm.mu.RUnlock()

	// translated comment
	if err := next(ctx, data); err != nil {
		return fmt.Errorf("stage %s: %w", stageName, err)
	}

	// translated comment
	cm.mu.Lock()
	// translated comment
	if len(cm.cache) >= 10000 {
		cm.cache = make(map[string]interface{})
	}
	cm.cache[cacheKey] = data.Output
	cm.mu.Unlock()

	return nil
}

// translated comment
func (cm *CachingMiddleware) generateCacheKey(stageName string, input interface{}) string {
	// translated comment
	// translated comment
	switch v := input.(type) {
	case string:
		// translated comment
		if len(v) > 128 {
			return fmt.Sprintf("%s:%x", stageName, hashString(v))
		}
		return fmt.Sprintf("%s:%s", stageName, v)
	case []byte:
		// translated comment
		if len(v) > 128 {
			return fmt.Sprintf("%s:%x", stageName, hashBytes(v))
		}
		return fmt.Sprintf("%s:%x", stageName, v)
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool:
		return fmt.Sprintf("%s:%v", stageName, v)
	default:
		// translated comment
		return fmt.Sprintf("%s:%p", stageName, input)
	}
}

// translated comment
func hashString(s string) uint64 {
	// translated comment
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

// translated comment
func hashBytes(b []byte) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	hash := uint64(offset64)
	limit := len(b)
	if limit > 1024 {
		limit = 1024 // translated comment
	}
	for i := 0; i < limit; i++ {
		hash ^= uint64(b[i])
		hash *= prime64
	}
	return hash
}

// ========================================================================
// translated comment
// ========================================================================

// translated comment
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

// translated comment
func NewMockStage(name string, deps []string, fn func(ctx context.Context, data *StageData) error) *MockStage {
	return &MockStage{
		name:         name,
		dependencies: deps,
		fn:           fn,
	}
}
