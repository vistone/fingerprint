package extension

import (
	"context"
	"fmt"
	"sort"

	"github.com/vistone/fingerprint/modules/internal/pipeline"
)

// ========================================================================
// translated comment
// ========================================================================

// translated comment
type ParseStage struct {
	registry RegistryPort
}

// translated comment
func NewParseStage(registry RegistryPort) *ParseStage {
	return &ParseStage{registry: registry}
}

// translated comment
func (p *ParseStage) GetName() string {
	return "parse"
}

// translated comment
func (p *ParseStage) GetDependencies() []string {
	return []string{} // translated comment
}

// translated comment
func (p *ParseStage) Execute(ctx context.Context, data *pipeline.StageData) error {
	// translated comment
	request, ok := data.Input.(*ProcessingRequest)
	if !ok {
		return fmt.Errorf("expected *ProcessingRequest, got %T", data.Input)
	}

	// translated comment
	parser, err := p.registry.GetParser(request.ExtensionType)
	if err != nil {
		return fmt.Errorf("failed to get parser: %w", err)
	}

	// translated comment
	parsedData, err := parser.Parse(request.RawData, ctx)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// translated comment
	data.Context["parsed_data"] = parsedData
	data.Output = map[string]interface{}{
		"parsed_data": parsedData,
	}

	return nil
}

// ========================================================================
// translated comment
// ========================================================================

// translated comment
type AnalyzeStage struct {
	registry RegistryPort
}

// translated comment
func NewAnalyzeStage(registry RegistryPort) *AnalyzeStage {
	return &AnalyzeStage{registry: registry}
}

// translated comment
func (a *AnalyzeStage) GetName() string {
	return "analyze"
}

// translated comment
func (a *AnalyzeStage) GetDependencies() []string {
	return []string{"parse"} // translated comment
}

// translated comment
func (a *AnalyzeStage) Execute(ctx context.Context, data *pipeline.StageData) error {
	// translated comment
	request, ok := data.Input.(*ProcessingRequest)
	if !ok {
		return fmt.Errorf("expected *ProcessingRequest, got %T", data.Input)
	}

	// translated comment
	parsedData, ok := data.Context["parsed_data"].(ExtensionData)
	if !ok {
		return fmt.Errorf("parsed_data not found or invalid in context")
	}

	// translated comment
	analyzer, err := a.registry.GetAnalyzer(request.ExtensionType)
	if err != nil {
		// translated comment
		data.Output = map[string]interface{}{
			"analysis_results": []interface{}{},
		}
		return nil
	}

	// translated comment
	analysisResult, err := analyzer.Analyze(parsedData, request.AnalysisConfig)
	if err != nil {
		return fmt.Errorf("analyze error: %w", err)
	}

	// translated comment
	data.Context["analysis_result"] = analysisResult
	data.Output = map[string]interface{}{
		"analysis_result": analysisResult,
	}

	return nil
}

// ========================================================================
// translated comment
// ========================================================================

// translated comment
type TransformStage struct {
	registry RegistryPort
}

// translated comment
func NewTransformStage(registry RegistryPort) *TransformStage {
	return &TransformStage{registry: registry}
}

// translated comment
func (t *TransformStage) GetName() string {
	return "transform"
}

// translated comment
func (t *TransformStage) GetDependencies() []string {
	return []string{"parse"} // translated comment
}

// translated comment
func (t *TransformStage) Execute(ctx context.Context, data *pipeline.StageData) error {
	// translated comment
	request, ok := data.Input.(*ProcessingRequest)
	if !ok {
		return fmt.Errorf("expected *ProcessingRequest, got %T", data.Input)
	}

	// translated comment
	parsedData, ok := data.Context["parsed_data"].(ExtensionData)
	if !ok {
		return fmt.Errorf("parsed_data not found in context")
	}

	// translated comment
	if request.AnalysisConfig == nil {
		data.Output = map[string]interface{}{
			"transformed_data": parsedData,
		}
		return nil
	}

	// translated comment
	rawTransforms, ok := request.AnalysisConfig["transforms"]
	if !ok {
		data.Output = map[string]interface{}{
			"transformed_data": parsedData,
		}
		return nil
	}

	// translated comment
	transformNames := make([]string, 0)
	switch values := rawTransforms.(type) {
	case []string:
		transformNames = append(transformNames, values...)
	case []interface{}:
		for _, value := range values {
			name, ok := value.(string)
			if !ok {
				return fmt.Errorf("invalid transform name type: %T", value)
			}
			transformNames = append(transformNames, name)
		}
	default:
		return fmt.Errorf("unexpected transforms type: %T", rawTransforms)
	}

	// translated comment
	data.Context["transforms_applied"] = transformNames
	data.Output = map[string]interface{}{
		"transformed_data":   parsedData,
		"transforms_applied": transformNames,
	}

	return nil
}

// ========================================================================
// translated comment
// ========================================================================

// translated comment
type HandleStage struct {
	registry RegistryPort
}

// translated comment
func NewHandleStage(registry RegistryPort) *HandleStage {
	return &HandleStage{registry: registry}
}

// translated comment
func (h *HandleStage) GetName() string {
	return "handle"
}

// translated comment
func (h *HandleStage) GetDependencies() []string {
	return []string{"parse"} // translated comment
}

// translated comment
func (h *HandleStage) Execute(ctx context.Context, data *pipeline.StageData) error {
	// translated comment
	request, ok := data.Input.(*ProcessingRequest)
	if !ok {
		return fmt.Errorf("expected *ProcessingRequest, got %T", data.Input)
	}

	// translated comment
	parsedData, ok := data.Context["parsed_data"].(ExtensionData)
	if !ok {
		return fmt.Errorf("parsed_data not found in context")
	}

	// translated comment
	handlers := h.registry.GetHandlers(request.ExtensionType)

	// translated comment
	sort.Slice(handlers, func(i, j int) bool {
		return handlers[i].GetPriority() > handlers[j].GetPriority()
	})

	// translated comment
	events := make([]*ExtensionEvent, 0, len(handlers))
	for _, handler := range handlers {
		event := &ExtensionEvent{
			Type:          "handle",
			ExtensionType: request.ExtensionType,
			Data:          parsedData,
			Context:       ctx,
			Metadata:      request.Metadata,
		}

		handlerResult, err := handler.Handle(event)
		if err != nil {
			return fmt.Errorf("handler error: %w", err)
		}

		events = append(events, event)

		// translated comment
		if !handlerResult.ContinueProcessing {
			break
		}
	}

	// translated comment
	data.Context["events"] = events
	data.Output = map[string]interface{}{
		"events": events,
	}

	return nil
}
