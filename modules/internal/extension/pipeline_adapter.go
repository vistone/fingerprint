//go:build pipelineadapter
// +build pipelineadapter

package extension

import (
	"context"
	"fmt"
	"time"

	proc_pipeline "github.com/vistone/fingerprint/modules/internal/pipeline"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// ========================================================================
// Pipeline Adapter: ProcessingEngine Refactoring
// ========================================================================

// ProcessWithPipeline processes extension requests using the Pipeline framework (new approach).
// Functionally equivalent to Process(), but uses the Pipeline framework.
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

	// Create context
	ctx := request.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// Execute pre-interceptors
	if err := e.executeInterceptors("pre", request, result); err != nil {
		result.Error = fmt.Sprintf("pre-interceptor error: %v", err)
		result.Success = false
		return result
	}

	// Create pipeline
	tracer := otel.Tracer("processing-engine")
	pipeline := proc_pipeline.NewPipeline(tracer)

	// Determine processing steps
	steps := request.Steps
	if len(steps) == 0 {
		steps = []string{"parse", "analyze"} // default steps
	}

	// Add the corresponding Stage for each requested step
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

	// Execute pipeline
	startTime := time.Now()
	stageData, err := pipeline.Execute(ctx, request)
	duration := time.Since(startTime)
	result.ElapsedMs = duration.Milliseconds()

	if err != nil {
		result.Error = fmt.Sprintf("pipeline error: %v", err)
		result.Success = false
		return result
	}

	// Extract results from Pipeline StageData
	if err := e.extractPipelineResults(stageData, result, steps); err != nil {
		result.Error = fmt.Sprintf("result extraction error: %v", err)
		result.Success = false
		return result
	}

	// Execute post-interceptors
	if err := e.executeInterceptors("post", request, result); err != nil {
		result.Error = fmt.Sprintf("post-interceptor error: %v", err)
		result.Success = false
		return result
	}

	return result
}

// extractPipelineResults extracts results from Pipeline StageData into ProcessingResult
func (e *ProcessingEngine) extractPipelineResults(
	stageData *proc_pipeline.StageData,
	result *ProcessingResult,
	steps []string,
) error {
	// Extract parsed data
	if parsedData, ok := stageData.Context["parsed_data"].(ExtensionData); ok {
		result.ParsedData = parsedData
	}

	// Extract analysis results (if analysis step was requested)
	if contains(steps, "analyze") {
		if analysisResult, ok := stageData.Context["analysis_result"].(AnalysisResult); ok {
			result.AnalysisResults = append(result.AnalysisResults, analysisResult)
		}
	}

	// Extract handling events (if handle step was requested)
	if contains(steps, "handle") {
		if events, ok := stageData.Context["events"].([]*ExtensionEvent); ok {
			result.Events = events
		}
	}

	return nil
}

// contains checks whether a string slice contains the specified element
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// ========================================================================
// Hybrid mode: ProcessingEngineWithPipeline
// ========================================================================

// ProcessingEngineWithPipeline is a hybrid processing engine: supports both old switch-case and new Pipeline approaches
// Used for gradual migration from old Process() to new ProcessWithPipeline()
type ProcessingEngineWithPipeline struct {
	engine      *ProcessingEngine
	tracer      trace.Tracer
	usePipeline bool // whether to use the Pipeline framework
}

// NewProcessingEngineWithPipeline creates a hybrid processing engine
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

// Process processes a request, choosing the approach based on the usePipeline flag
func (pwp *ProcessingEngineWithPipeline) Process(request *ProcessingRequest) *ProcessingResult {
	if pwp.usePipeline {
		return pwp.engine.ProcessWithPipeline(request)
	}
	return pwp.engine.Process(request)
}

// SwitchPipelineMode switches to Pipeline mode
func (pwp *ProcessingEngineWithPipeline) SwitchPipelineMode(enable bool) {
	pwp.usePipeline = enable
}

// GetPipelineMode returns whether Pipeline mode is currently active
func (pwp *ProcessingEngineWithPipeline) GetPipelineMode() bool {
	return pwp.usePipeline
}

// GetUnderlyingEngine returns the underlying ProcessingEngine
func (pwp *ProcessingEngineWithPipeline) GetUnderlyingEngine() *ProcessingEngine {
	return pwp.engine
}
