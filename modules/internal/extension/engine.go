package extension

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// translated comment
// translated comment
type ProcessingEngine struct {
	mu           sync.RWMutex
	config       *EngineConfig
	registry     *ExtensionRegistry
	interceptors map[string][]Interceptor
}

// translated comment
type EngineConfig struct {
	// translated comment
	ConcurrentProcessing bool

	// translated comment
	MaxConcurrency int

	// translated comment
	EnableCaching bool

	// translated comment
	CacheSize int

	// translated comment
	TimeoutMs int

	// translated comment
	StrictValidation bool

	// translated comment
	VerboseLogging bool

	// translated comment
	CustomConfig map[string]interface{}
}

// translated comment
type ProcessingRequest struct {
	// translated comment
	ExtensionType ExtensionType

	// translated comment
	RawData []byte

	// translated comment
	Steps []string

	// translated comment
	AnalysisConfig map[string]interface{}

	// translated comment
	Context context.Context

	// translated comment
	Metadata map[string]interface{}
}

// translated comment
type ProcessingResult struct {
	// translated comment
	RequestID string

	// translated comment
	Success bool

	// translated comment
	Error string

	// translated comment
	ParsedData ExtensionData

	// translated comment
	AnalysisResults []AnalysisResult

	// translated comment
	Events []*ExtensionEvent

	// translated comment
	ElapsedMs int64

	// translated comment
	Metadata map[string]interface{}
}

// translated comment
// translated comment
type Interceptor interface {
	// translated comment
	// translated comment
	// translated comment
	// translated comment
	Intercept(phase string, request *ProcessingRequest, result *ProcessingResult) error
}

// translated comment
func NewProcessingEngine(config *EngineConfig) *ProcessingEngine {
	if config == nil {
		config = &EngineConfig{
			ConcurrentProcessing: true,
			MaxConcurrency:       16,
			EnableCaching:        true,
			CacheSize:            1000,
			TimeoutMs:            5000,
			StrictValidation:     true,
			VerboseLogging:       false,
		}
	}

	return &ProcessingEngine{
		config:       config,
		registry:     GetRegistry(),
		interceptors: make(map[string][]Interceptor),
	}
}

// translated comment
// translated comment
func (e *ProcessingEngine) RegisterInterceptor(phase string, interceptor Interceptor) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if phase != "pre" && phase != "post" {
		return fmt.Errorf("invalid phase: %s", phase)
	}

	if interceptor == nil {
		return fmt.Errorf("interceptor cannot be nil")
	}

	key := phase
	e.interceptors[key] = append(e.interceptors[key], interceptor)
	return nil
}

// translated comment
func (e *ProcessingEngine) Process(request *ProcessingRequest) *ProcessingResult {
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
	steps := request.Steps
	if len(steps) == 0 {
		steps = []string{"parse", "analyze"} // translated comment
	}

	// translated comment
	for _, step := range steps {
		switch step {
		case "parse":
			if err := e.parseExtension(ctx, request, result); err != nil {
				result.Error = fmt.Sprintf("parse error: %v", err)
				result.Success = false
				return result
			}

		case "analyze":
			if result.ParsedData == nil {
				if err := e.parseExtension(ctx, request, result); err != nil {
					result.Error = fmt.Sprintf("parse error: %v", err)
					result.Success = false
					return result
				}
			}

			if err := e.analyzeExtension(ctx, request, result); err != nil {
				result.Error = fmt.Sprintf("analyze error: %v", err)
				result.Success = false
				return result
			}

		case "handle":
			if err := e.handleExtension(ctx, request, result); err != nil {
				result.Error = fmt.Sprintf("handle error: %v", err)
				result.Success = false
				return result
			}

		case "transform":
			if result.ParsedData == nil {
				if err := e.parseExtension(ctx, request, result); err != nil {
					result.Error = fmt.Sprintf("parse error: %v", err)
					result.Success = false
					return result
				}
			}

			if err := e.transformExtension(ctx, request, result); err != nil {
				result.Error = fmt.Sprintf("transform error: %v", err)
				result.Success = false
				return result
			}

		default:
			result.Error = fmt.Sprintf("unknown step: %s", step)
			result.Success = false
			return result
		}
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
func (e *ProcessingEngine) parseExtension(ctx context.Context, request *ProcessingRequest, result *ProcessingResult) error {
	parser, err := GetParser(request.ExtensionType)
	if err != nil {
		return fmt.Errorf("get parser for %v: %w", request.ExtensionType, err)
	}

	parsedData, err := parser.Parse(request.RawData, ctx)
	if err != nil {
		return fmt.Errorf("parse extension data: %w", err)
	}

	result.ParsedData = parsedData
	return nil
}

// translated comment
func (e *ProcessingEngine) analyzeExtension(ctx context.Context, request *ProcessingRequest, result *ProcessingResult) error {
	analyzer, err := GetAnalyzer(request.ExtensionType)
	if err != nil {
		// translated comment
		return nil
	}

	analysisResult, err := analyzer.Analyze(result.ParsedData, request.AnalysisConfig)
	if err != nil {
		return fmt.Errorf("analyze extension data: %w", err)
	}

	result.AnalysisResults = append(result.AnalysisResults, analysisResult)
	return nil
}

// translated comment
func (e *ProcessingEngine) handleExtension(ctx context.Context, request *ProcessingRequest, result *ProcessingResult) error {
	handlers := GetHandlers(request.ExtensionType)

	// translated comment
	sort.Slice(handlers, func(i, j int) bool {
		return handlers[i].GetPriority() > handlers[j].GetPriority()
	})

	for _, handler := range handlers {
		event := &ExtensionEvent{
			Type:          "handle",
			ExtensionType: request.ExtensionType,
			Data:          result.ParsedData,
			Context:       ctx,
			Metadata:      request.Metadata,
		}

		handlerResult, err := handler.Handle(event)
		if err != nil {
			return fmt.Errorf("handle extension event: %w", err)
		}

		result.Events = append(result.Events, event)

		// translated comment
		if !handlerResult.ContinueProcessing {
			break
		}
	}

	return nil
}

// translated comment
func (e *ProcessingEngine) transformExtension(ctx context.Context, request *ProcessingRequest, result *ProcessingResult) error {
	if result.ParsedData == nil {
		return fmt.Errorf("parsed data is nil")
	}

	if request.AnalysisConfig == nil {
		return nil
	}

	rawTransforms, ok := request.AnalysisConfig["transforms"]
	if !ok {
		return nil
	}

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
		return fmt.Errorf("invalid transforms config type: %T", rawTransforms)
	}

	if len(transformNames) == 0 {
		return nil
	}

	current := interface{}(result.ParsedData)
	for _, name := range transformNames {
		transformer, err := CreateTransform(name)
		if err != nil {
			return fmt.Errorf("create transform %s failed: %w", name, err)
		}

		output, err := transformer.Transform(current)
		if err != nil {
			return fmt.Errorf("transform %s failed: %w", name, err)
		}

		current = output
	}

	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["transform_chain"] = transformNames
	result.Metadata["transformed_data"] = current

	result.Events = append(result.Events, &ExtensionEvent{
		Type:          "transform",
		ExtensionType: request.ExtensionType,
		Data:          current,
		Context:       ctx,
		Metadata: map[string]interface{}{
			"transform_chain": transformNames,
		},
	})

	return nil
}

// translated comment
func (e *ProcessingEngine) executeInterceptors(phase string, request *ProcessingRequest, result *ProcessingResult) error {
	e.mu.RLock()
	interceptors := e.interceptors[phase]
	e.mu.RUnlock()

	for _, interceptor := range interceptors {
		if err := interceptor.Intercept(phase, request, result); err != nil {
			return fmt.Errorf("intercept phase %s: %w", phase, err)
		}
	}

	return nil
}

// translated comment
func (e *ProcessingEngine) GetConfig() *EngineConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// translated comment
	if e.config == nil {
		return nil
	}

	customConfig := make(map[string]interface{})
	for k, v := range e.config.CustomConfig {
		customConfig[k] = v
	}

	return &EngineConfig{
		ConcurrentProcessing: e.config.ConcurrentProcessing,
		MaxConcurrency:       e.config.MaxConcurrency,
		EnableCaching:        e.config.EnableCaching,
		CacheSize:            e.config.CacheSize,
		TimeoutMs:            e.config.TimeoutMs,
		StrictValidation:     e.config.StrictValidation,
		VerboseLogging:       e.config.VerboseLogging,
		CustomConfig:         customConfig,
	}
}

// translated comment
func (e *ProcessingEngine) SetConfig(config *EngineConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.config = config
	return nil
}
