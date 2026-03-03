package extension

import (
	"context"
	"fmt"
	"time"

	proc_pipeline "github.com/vistone/fingerprint/internal/pipeline"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// ========================================================================
// Pipeline 适配器：ProcessingEngine 改造
// ========================================================================

// ProcessWithPipeline 使用 Pipeline 框架处理扩展请求（新方式）
// 与 Process() 方法功能相同，但使用 Pipeline 框架
func (e *ProcessingEngine) ProcessWithPipeline(request *ProcessingRequest) *ProcessingResult {
	result := &ProcessingResult{
		Success:         true,
		AnalysisResults: []AnalysisResult{},
		Events:          []*ExtensionEvent{},
		Metadata:        make(map[string]interface{}),
	}

	if request == nil {
		result.Error = "request is nil"
		result.Success = false
		return result
	}

	// 创建上下文
	ctx := request.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// 执行 Pre 拦截器
	if err := e.executeInterceptors("pre", request, result); err != nil {
		result.Error = fmt.Sprintf("pre-interceptor error: %v", err)
		result.Success = false
		return result
	}

	// 创建 Pipeline
	tracer := otel.Tracer("processing-engine")
	pipeline := proc_pipeline.NewPipeline(tracer)

	// 决定处理步骤
	steps := request.Steps
	if len(steps) == 0 {
		steps = []string{"parse", "analyze"} // 默认步骤
	}

	// 根据请求的步骤添加相应的 Stage
	stageMap := map[string]proc_pipeline.Stage{
		"parse":     NewParseStage(e.registry),
		"analyze":   NewAnalyzeStage(e.registry),
		"transform": NewTransformStage(e.registry),
		"handle":    NewHandleStage(e.registry),
	}

	for _, stepName := range steps {
		stage, ok := stageMap[stepName]
		if !ok {
			result.Error = fmt.Sprintf("unknown step: %s", stepName)
			result.Success = false
			return result
		}
		pipeline.AddStage(stage)
	}

	// 执行 Pipeline
	startTime := time.Now()
	stageData, err := pipeline.Execute(ctx, request)
	duration := time.Since(startTime)
	result.ElapsedMs = duration.Milliseconds()

	if err != nil {
		result.Error = fmt.Sprintf("pipeline error: %v", err)
		result.Success = false
		return result
	}

	// 从 Pipeline 的 StageData 提取结果
	if err := e.extractPipelineResults(stageData, result, steps); err != nil {
		result.Error = fmt.Sprintf("result extraction error: %v", err)
		result.Success = false
		return result
	}

	// 执行 Post 拦截器
	if err := e.executeInterceptors("post", request, result); err != nil {
		result.Error = fmt.Sprintf("post-interceptor error: %v", err)
		result.Success = false
		return result
	}

	return result
}

// extractPipelineResults 从 Pipeline 的 StageData 中提取结果到 ProcessingResult
func (e *ProcessingEngine) extractPipelineResults(
	stageData *proc_pipeline.StageData,
	result *ProcessingResult,
	steps []string,
) error {
	// 提取已解析的数据
	if parsedData, ok := stageData.Context["parsed_data"].(ExtensionData); ok {
		result.ParsedData = parsedData
	}

	// 提取分析结果（如果请求了分析步骤）
	if contains(steps, "analyze") {
		if analysisResult, ok := stageData.Context["analysis_result"].(AnalysisResult); ok {
			result.AnalysisResults = append(result.AnalysisResults, analysisResult)
		}
	}

	// 提取处理事件（如果请求了处理步骤）
	if contains(steps, "handle") {
		if events, ok := stageData.Context["events"].([]*ExtensionEvent); ok {
			result.Events = events
		}
	}

	return nil
}

// contains 检查字符串切片是否包含指定的元素
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// ========================================================================
// 混合模式：ProcessingEngineWithPipeline
// ========================================================================

// ProcessingEngineWithPipeline 混合模式处理引擎：同时支持旧的 switch-case 和新的 Pipeline 方式
// 用于从旧的 Process() 逐步迁移到新的 ProcessWithPipeline()
type ProcessingEngineWithPipeline struct {
	engine      *ProcessingEngine
	tracer      trace.Tracer
	usePipeline bool // 是否使用 Pipeline 框架
}

// NewProcessingEngineWithPipeline 创建混合模式处理引擎
func NewProcessingEngineWithPipeline(
	engine *ProcessingEngine,
	tracer trace.Tracer,
	usePipeline bool,
) *ProcessingEngineWithPipeline {
	if tracer == nil {
		tracer = otel.Tracer("processing-engine-hybrid")
	}
	return &ProcessingEngineWithPipeline{
		engine:      engine,
		tracer:      tracer,
		usePipeline: usePipeline,
	}
}

// Process 处理请求，根据 usePipeline 标志决定使用哪种方式
func (pwp *ProcessingEngineWithPipeline) Process(request *ProcessingRequest) *ProcessingResult {
	if pwp.usePipeline {
		return pwp.engine.ProcessWithPipeline(request)
	}
	return pwp.engine.Process(request)
}

// SwitchPipelineMode 切换到 Pipeline 模式
func (pwp *ProcessingEngineWithPipeline) SwitchPipelineMode(enable bool) {
	pwp.usePipeline = enable
}

// GetPipelineMode 获取当前是否使用 Pipeline 模式
func (pwp *ProcessingEngineWithPipeline) GetPipelineMode() bool {
	return pwp.usePipeline
}

// GetUnderlyingEngine 获取底层的 ProcessingEngine
func (pwp *ProcessingEngineWithPipeline) GetUnderlyingEngine() *ProcessingEngine {
	return pwp.engine
}
