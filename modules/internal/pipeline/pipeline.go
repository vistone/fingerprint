// internal/pipeline/pipeline.go
// 链式责任模式的流水线框架

package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ========================================================================
// 核心接口定义
// ========================================================================

// Tracer 追踪接口（简化版，避免依赖OpenTelemetry）
type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, Span)
}

// Span 追踪Span接口
type Span interface {
	End()
	SetAttributes(attrs ...Attribute)
	RecordError(err error)
}

// Attribute 追踪属性
type Attribute struct {
	Key   string
	Value interface{}
}

// NoOpTracer 空追踪器实现
type NoOpTracer struct{}

func (t NoOpTracer) Start(ctx context.Context, name string) (context.Context, Span) {
	return ctx, NoOpSpan{}
}

// NoOpSpan 空Span实现
type NoOpSpan struct{}

func (s NoOpSpan) End()                             {}
func (s NoOpSpan) SetAttributes(attrs ...Attribute) {}
func (s NoOpSpan) RecordError(err error)            {}

// DefaultTracer 返回默认的空追踪器
func DefaultTracer() Tracer {
	return NoOpTracer{}
}

// ========================================================================
// 核心接口定义
// ========================================================================

// Stage 流水线的单个阶段
type Stage interface {
	// GetName 返回阶段名称（唯一标识符）
	GetName() string

	// GetDependencies 返回此阶段依赖的前置阶段名称
	GetDependencies() []string

	// Execute 执行阶段的核心逻辑
	Execute(ctx context.Context, data *StageData) error
}

// Middleware 中间件接口（装饰器模式）
// 用于在 Stage 执行前后进行日志、指标、追踪等操作
type Middleware interface {
	Process(ctx context.Context, stageName string, data *StageData, next ExecutionFunc) error
}

// ExecutionFunc 阶段执行函数
type ExecutionFunc func(ctx context.Context, data *StageData) error

// StageData 流动在管道中的数据结构
type StageData struct {
	// 原始输入数据
	Input interface{}

	// 当前阶段的输出（下一阶段的输入）
	Output interface{}

	// 上下文数据（各阶段可读写）
	Context map[string]interface{}

	// 追踪元数据
	ExecutedAt time.Time
	Duration   time.Duration
	Error      error
}

// ========================================================================
// Pipeline 流水线
// ========================================================================

type Pipeline struct {
	stages      []Stage
	middlewares []Middleware
	stageIndex  map[string]int // 阶段名 -> 索引，用于快速查找依赖

	tracer Tracer
}

// NewPipeline 创建新流水线
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

// AddStage 添加阶段
// 注意：阶段顺序即为执行顺序，但会根据依赖关系验证
func (p *Pipeline) AddStage(stage Stage) *Pipeline {
	p.stages = append(p.stages, stage)
	p.stageIndex[stage.GetName()] = len(p.stages) - 1
	return p
}

// AddMiddleware 添加中间件（会按添加顺序链式执行）
func (p *Pipeline) AddMiddleware(mw Middleware) *Pipeline {
	p.middlewares = append(p.middlewares, mw)
	return p
}

// Validate 验证整个流水线的有效性
// 检查：
// 1. 所有依赖的前置阶段是否存在
// 2. 是否存在循环依赖
func (p *Pipeline) Validate() error {
	for i, stage := range p.stages {
		stageName := stage.GetName()

		// 检查依赖
		for _, dep := range stage.GetDependencies() {
			depIdx, exists := p.stageIndex[dep]
			if !exists {
				return fmt.Errorf("stage %s depends on %s, but %s not found",
					stageName, dep, dep)
			}

			// 依赖必须在前面执行（索引小于当前阶段）
			if depIdx >= i {
				return fmt.Errorf("circular dependency detected: %s (index %d) depends on %s (index %d)",
					stageName, i, dep, depIdx)
			}
		}
	}

	return nil
}

// Execute 执行整个流水线
func (p *Pipeline) Execute(ctx context.Context, input interface{}) (*StageData, error) {
	// 验证流水线
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("pipeline validation failed: %w", err)
	}

	// 初始化数据
	data := &StageData{
		Input:   input,
		Output:  input, // 初始状态下，输出就是输入
		Context: make(map[string]interface{}),
	}

	// 执行各阶段
	_, pipelineSpan := p.tracer.Start(ctx, "Pipeline.Execute")
	defer pipelineSpan.End()

	for i, stage := range p.stages {
		stageName := stage.GetName()

		// 创建此阶段的 span
		stageCtx, stageSpan := p.tracer.Start(ctx, "stage."+stageName)

		startTime := time.Now()

		// 用中间件包装执行
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

// executeStageWithMiddleware 用中间件链包装执行单个阶段
func (p *Pipeline) executeStageWithMiddleware(ctx context.Context, stage Stage, data *StageData) error {
	// 构建最内层的执行函数
	var handler ExecutionFunc = func(ctx context.Context, d *StageData) error {
		return stage.Execute(ctx, d)
	}

	// 反向遍历中间件，构建链（最后添加的中间件最先执行）
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		mw := p.middlewares[i]
		nextHandler := handler
		handler = func(ctx context.Context, d *StageData) error {
			return mw.Process(ctx, stage.GetName(), d, nextHandler)
		}
	}

	// 执行中间件链
	return handler(ctx, data)
}

// ========================================================================
// 内置中间件
// ========================================================================

// LoggingMiddleware 日志中间件
type LoggingMiddleware struct {
	logger interface { // 接收任意实现了 Info/Error 的日志器
		Info(msg string, fields ...interface{})
		Error(msg string, fields ...interface{})
	}
}

func NewLoggingMiddleware(logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}) *LoggingMiddleware {
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
	} else {
		lm.logger.Info("stage completed",
			"stage", stageName,
			"duration_ms", duration.Milliseconds(),
		)
	}

	return err
}

// MetricsMiddleware 指标中间件
type MetricsMiddleware struct {
	metrics interface { // 接收实现了 Record 的指标收集器
		Record(stage string, duration time.Duration, success bool)
	}
}

func NewMetricsMiddleware(metrics interface {
	Record(stage string, duration time.Duration, success bool)
}) *MetricsMiddleware {
	return &MetricsMiddleware{metrics: metrics}
}

func (mm *MetricsMiddleware) Process(ctx context.Context, stageName string, data *StageData, next ExecutionFunc) error {
	startTime := time.Now()
	err := next(ctx, data)
	duration := time.Since(startTime)

	success := err == nil
	mm.metrics.Record(stageName, duration, success)

	return err
}

// RecoveryMiddleware 恢复中间件（捕获 panic）
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

// TimeoutMiddleware 超时中间件
type TimeoutMiddleware struct {
	timeout time.Duration
}

func NewTimeoutMiddleware(timeout time.Duration) *TimeoutMiddleware {
	return &TimeoutMiddleware{timeout: timeout}
}

func (tm *TimeoutMiddleware) Process(ctx context.Context, stageName string, data *StageData, next ExecutionFunc) error {
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, tm.timeout)
	defer cancel()

	// 使用带超时的上下文执行
	done := make(chan error, 1)
	go func() {
		done <- next(timeoutCtx, data)
	}()

	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		return fmt.Errorf("stage %s timeout after %v", stageName, tm.timeout)
	}
}

// CachingMiddleware 缓存中间件
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
	// 生成缓存键（使用哈希避免大对象问题）
	cacheKey := cm.generateCacheKey(stageName, data.Input)

	cm.mu.RLock()
	if cached, exists := cm.cache[cacheKey]; exists {
		cm.mu.RUnlock()
		data.Output = cached
		return nil
	}
	cm.mu.RUnlock()

	// 执行阶段
	if err := next(ctx, data); err != nil {
		return err
	}

	// 缓存结果（限制缓存大小）
	cm.mu.Lock()
	// 简单的缓存淘汰策略：当缓存过大时清空
	if len(cm.cache) >= 10000 {
		cm.cache = make(map[string]interface{})
	}
	cm.cache[cacheKey] = data.Output
	cm.mu.Unlock()

	return nil
}

// generateCacheKey 生成缓存键（使用哈希避免大对象问题）
func (cm *CachingMiddleware) generateCacheKey(stageName string, input interface{}) string {
	// 使用指针地址作为键的基础（对对象有效）
	// 对于字符串或基本类型，直接使用值
	switch v := input.(type) {
	case string:
		// 对长字符串使用哈希
		if len(v) > 128 {
			return fmt.Sprintf("%s:%x", stageName, hashString(v))
		}
		return fmt.Sprintf("%s:%s", stageName, v)
	case []byte:
		// 对字节切片使用哈希
		if len(v) > 128 {
			return fmt.Sprintf("%s:%x", stageName, hashBytes(v))
		}
		return fmt.Sprintf("%s:%x", stageName, v)
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool:
		return fmt.Sprintf("%s:%v", stageName, v)
	default:
		// 对其他类型使用指针地址
		return fmt.Sprintf("%s:%p", stageName, input)
	}
}

// hashString 计算字符串的简化哈希
func hashString(s string) uint64 {
	// 使用 FNV-1a 哈希算法
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

// hashBytes 计算字节切片的简化哈希
func hashBytes(b []byte) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	hash := uint64(offset64)
	limit := len(b)
	if limit > 1024 {
		limit = 1024 // 限制哈希计算的数据量
	}
	for i := 0; i < limit; i++ {
		hash ^= uint64(b[i])
		hash *= prime64
	}
	return hash
}

// ========================================================================
// 单元测试辅助函数
// ========================================================================

// MockStage 用于测试的模拟阶段
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

// NewMockStage 创建模拟阶段
func NewMockStage(name string, deps []string, fn func(ctx context.Context, data *StageData) error) *MockStage {
	return &MockStage{
		name:         name,
		dependencies: deps,
		fn:           fn,
	}
}
