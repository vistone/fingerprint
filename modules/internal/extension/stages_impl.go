package extension

import (
	"context"
	"fmt"
	"sort"

	"github.com/vistone/fingerprint/modules/internal/pipeline"
)

// ========================================================================
// ParseStage: parses extension data
// ========================================================================

// ParseStage parses extension data (from raw TLS data into structured format)
type ParseStage struct {
	registry RegistryPort
}

// NewParseStage creates a new ParseStage
func NewParseStage(registry RegistryPort) *ParseStage {
	return &ParseStage{registry: registry}
}

// GetName returns the stage name
func (p *ParseStage) GetName() string {
	return "parse"
}

// GetDependencies returns the required preceding stages
func (p *ParseStage) GetDependencies() []string {
	return []string{} // parsing is the first step, no dependencies
}

// Execute executes the stage
func (p *ParseStage) Execute(ctx context.Context, data *pipeline.StageData) error {
	// Get request information from input
	request, ok := data.Input.(*ProcessingRequest)
	if !ok {
		return fmt.Errorf("expected *ProcessingRequest, got %T", data.Input)
	}

	// Get parser from registry
	parser, err := p.registry.GetParser(request.ExtensionType)
	if err != nil {
		return fmt.Errorf("failed to get parser: %w", err)
	}

	// Parse raw data
	parsedData, err := parser.Parse(request.RawData, ctx)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Store parse results in StageData
	data.Context["parsed_data"] = parsedData
	data.Output = map[string]interface{}{
		"parsed_data": parsedData,
	}

	return nil
}

// ========================================================================
// AnalyzeStage: analyzes extension data
// ========================================================================

// AnalyzeStage analyzes extension data
type AnalyzeStage struct {
	registry RegistryPort
}

// NewAnalyzeStage creates a new AnalyzeStage
func NewAnalyzeStage(registry RegistryPort) *AnalyzeStage {
	return &AnalyzeStage{registry: registry}
}

// GetName returns the stage name
func (a *AnalyzeStage) GetName() string {
	return "analyze"
}

// GetDependencies returns the required preceding stages
func (a *AnalyzeStage) GetDependencies() []string {
	return []string{"parse"} // analysis depends on parsing
}

// Execute executes the stage
func (a *AnalyzeStage) Execute(ctx context.Context, data *pipeline.StageData) error {
	// Get request information from input
	request, ok := data.Input.(*ProcessingRequest)
	if !ok {
		return fmt.Errorf("expected *ProcessingRequest, got %T", data.Input)
	}

	// Get parsed data from previous stage output
	parsedData, ok := data.Context["parsed_data"].(ExtensionData)
	if !ok {
		return fmt.Errorf("parsed_data not found or invalid in context")
	}

	// Get analyzer from registry
	analyzer, err := a.registry.GetAnalyzer(request.ExtensionType)
	if err != nil {
		// If no analyzer exists, do not treat as an error (some extension types may not have analyzers)
		data.Output = map[string]interface{}{
			"analysis_results": []interface{}{},
		}
		return nil
	}

	// Analyze parsed data
	analysisResult, err := analyzer.Analyze(parsedData, request.AnalysisConfig)
	if err != nil {
		return fmt.Errorf("analyze error: %w", err)
	}

	// Store analysis results
	data.Context["analysis_result"] = analysisResult
	data.Output = map[string]interface{}{
		"analysis_result": analysisResult,
	}

	return nil
}

// ========================================================================
// TransformStage: transforms extension data
// ========================================================================

// TransformStage transforms extension data into standard format
type TransformStage struct {
	registry RegistryPort
}

// NewTransformStage creates a new TransformStage
func NewTransformStage(registry RegistryPort) *TransformStage {
	return &TransformStage{registry: registry}
}

// GetName returns the stage name
func (t *TransformStage) GetName() string {
	return "transform"
}

// GetDependencies returns the required preceding stages
func (t *TransformStage) GetDependencies() []string {
	return []string{"parse"} // transformation depends on parsing
}

// Execute executes the stage
func (t *TransformStage) Execute(ctx context.Context, data *pipeline.StageData) error {
	// Get request information from input
	request, ok := data.Input.(*ProcessingRequest)
	if !ok {
		return fmt.Errorf("expected *ProcessingRequest, got %T", data.Input)
	}

	// Get parsed data from previous stage output
	parsedData, ok := data.Context["parsed_data"].(ExtensionData)
	if !ok {
		return fmt.Errorf("parsed_data not found in context")
	}

	// If no analysis configuration, skip transformation
	if request.AnalysisConfig == nil {
		data.Output = map[string]interface{}{
			"transformed_data": parsedData,
		}
		return nil
	}

	// Get transform list from analysis configuration
	rawTransforms, ok := request.AnalysisConfig["transforms"]
	if !ok {
		data.Output = map[string]interface{}{
			"transformed_data": parsedData,
		}
		return nil
	}

	// Parse transform name list
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

	// Apply transforms (simplified here to storing transform names; actual implementation may call registry transformers)
	data.Context["transforms_applied"] = transformNames
	data.Output = map[string]interface{}{
		"transformed_data":   parsedData,
		"transforms_applied": transformNames,
	}

	return nil
}

// ========================================================================
// HandleStage: handles extensions
// ========================================================================

// HandleStage handles extension events
type HandleStage struct {
	registry RegistryPort
}

// NewHandleStage creates a new HandleStage
func NewHandleStage(registry RegistryPort) *HandleStage {
	return &HandleStage{registry: registry}
}

// GetName returns the stage name
func (h *HandleStage) GetName() string {
	return "handle"
}

// GetDependencies returns the required preceding stages
func (h *HandleStage) GetDependencies() []string {
	return []string{"parse"} // handling depends on parsing
}

// Execute executes the stage
func (h *HandleStage) Execute(ctx context.Context, data *pipeline.StageData) error {
	// Get request information from input
	request, ok := data.Input.(*ProcessingRequest)
	if !ok {
		return fmt.Errorf("expected *ProcessingRequest, got %T", data.Input)
	}

	// Get parsed data from previous stage output
	parsedData, ok := data.Context["parsed_data"].(ExtensionData)
	if !ok {
		return fmt.Errorf("parsed_data not found in context")
	}

	// Get handlers from registry
	handlers := h.registry.GetHandlers(request.ExtensionType)

	// Sort handlers by priority
	sort.Slice(handlers, func(i, j int) bool {
		return handlers[i].GetPriority() > handlers[j].GetPriority()
	})

	// Execute all handlers
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

		// If the handler requests to stop propagation, break
		if !handlerResult.ContinueProcessing {
			break
		}
	}

	// Store handling events
	data.Context["events"] = events
	data.Output = map[string]interface{}{
		"events": events,
	}

	return nil
}
