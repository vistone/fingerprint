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
// translated comment
// ========================================================================

// translated comment
// translated comment
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

	// translated comment
	ctx := request.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// translated comment
	if err := e.executeInterceptors("pre", request, result); err != nil {
		result.Error = fmt.Sprintf("pre-interceptor error: %v", err)
		result.Success = false
		return result
	}

	// translated comment
	tracer := otel.Tracer("processing-engine")
	pipeline := proc_pipeline.NewPipeline(tracer)

	// translated comment
	steps := request.Steps
	if len(steps) == 0 {
		steps = []string{"parse", "analyze"} // translated comment
	}

	// translated comment
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

	// translated comment
	startTime := time.Now()
	stageData, err := pipeline.Execute(ctx, request)
	duration := time.Since(startTime)
	result.ElapsedMs = duration.Milliseconds()

	if err != nil {
		result.Error = fmt.Sprintf("pipeline error: %v", err)
		result.Success = false
		return result
	}

	// translated comment
	if err := e.extractPipelineResults(stageData, result, steps); err != nil {
		result.Error = fmt.Sprintf("result extraction error: %v", err)
		result.Success = false
		return result
	}

	// translated comment
	if err := e.executeInterceptors("post", request, result); err != nil {
		result.Error = fmt.Sprintf("post-interceptor error: %v", err)
		result.Success = false
		return result
	}

	return result
}

// translated comment
func (e *ProcessingEngine) extractPipelineResults(
	stageData *proc_pipeline.StageData,
	result *ProcessingResult,
	steps []string,
) error {
	// translated comment
	if parsedData, ok := stageData.Context["parsed_data"].(ExtensionData); ok {
		result.ParsedData = parsedData
	}

	// translated comment
	if contains(steps, "analyze") {
		if analysisResult, ok := stageData.Context["analysis_result"].(AnalysisResult); ok {
			result.AnalysisResults = append(result.AnalysisResults, analysisResult)
		}
	}

	// translated comment
	if contains(steps, "handle") {
		if events, ok := stageData.Context["events"].([]*ExtensionEvent); ok {
			result.Events = events
		}
	}

	return nil
}

// translated comment
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// ========================================================================
// translated comment
// ========================================================================

// translated comment
// translated comment
type ProcessingEngineWithPipeline struct {
	engine      *ProcessingEngine
	tracer      trace.Tracer
	usePipeline bool // translated comment
}

// translated comment
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

// translated comment
func (pwp *ProcessingEngineWithPipeline) Process(request *ProcessingRequest) *ProcessingResult {
	if pwp.usePipeline {
		return pwp.engine.ProcessWithPipeline(request)
	}
	return pwp.engine.Process(request)
}

// translated comment
func (pwp *ProcessingEngineWithPipeline) SwitchPipelineMode(enable bool) {
	pwp.usePipeline = enable
}

// translated comment
func (pwp *ProcessingEngineWithPipeline) GetPipelineMode() bool {
	return pwp.usePipeline
}

// translated comment
func (pwp *ProcessingEngineWithPipeline) GetUnderlyingEngine() *ProcessingEngine {
	return pwp.engine
}
