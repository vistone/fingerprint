package extension

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// ProcessingEngine 扩展处理引擎
// 协调解析、分析、转换和处理流程
type ProcessingEngine struct {
	mu           sync.RWMutex
	config       *EngineConfig
	registry     *ExtensionRegistry
	interceptors map[string][]Interceptor
}

// EngineConfig 引擎配置
type EngineConfig struct {
	// 是否启用并发处理
	ConcurrentProcessing bool

	// 最大并发数
	MaxConcurrency int

	// 是否启用缓存
	EnableCaching bool

	// 缓存大小
	CacheSize int

	// 处理超时（毫秒）
	TimeoutMs int

	// 是否严格验证
	StrictValidation bool

	// 是否记录详细日志
	VerboseLogging bool

	// 自定义配置
	CustomConfig map[string]interface{}
}

// ProcessingRequest 处理请求
type ProcessingRequest struct {
	// 扩展类型
	ExtensionType ExtensionType

	// 原始扩展数据
	RawData []byte

	// 处理步骤（parse, analyze, transform）
	Steps []string

	// 分析配置
	AnalysisConfig map[string]interface{}

	// 上下文信息
	Context context.Context

	// 请求元数据
	Metadata map[string]interface{}
}

// ProcessingResult 处理结果
type ProcessingResult struct {
	// 请求ID
	RequestID string

	// 是否成功
	Success bool

	// 错误消息
	Error string

	// 解析结果
	ParsedData ExtensionData

	// 分析结果列表
	AnalysisResults []AnalysisResult

	// 处理事件列表
	Events []*ExtensionEvent

	// 处理耗时（毫秒）
	ElapsedMs int64

	// 结果元数据
	Metadata map[string]interface{}
}

// Interceptor 拦截器接口
// 用于在处理流程中插入自定义逻辑
type Interceptor interface {
	// Intercept 拦截处理
	// phase: 处理阶段（pre, post）
	// request: 处理请求
	// result: 处理结果（post 阶段有效）
	Intercept(phase string, request *ProcessingRequest, result *ProcessingResult) error
}

// NewProcessingEngine 创建处理引擎
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

// RegisterInterceptor 注册拦截器
// phase: pre 或 post
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

// Process 处理扩展请求
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

	// 决定处理步骤
	steps := request.Steps
	if len(steps) == 0 {
		steps = []string{"parse", "analyze"} // 默认步骤
	}

	// 执行处理步骤
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

	// 执行 Post 拦截器
	if err := e.executeInterceptors("post", request, result); err != nil {
		result.Error = fmt.Sprintf("post-interceptor error: %v", err)
		result.Success = false
		return result
	}

	return result
}

// parseExtension 解析扩展
func (e *ProcessingEngine) parseExtension(ctx context.Context, request *ProcessingRequest, result *ProcessingResult) error {
	parser, err := GetParser(request.ExtensionType)
	if err != nil {
		return err
	}

	parsedData, err := parser.Parse(request.RawData, ctx)
	if err != nil {
		return err
	}

	result.ParsedData = parsedData
	return nil
}

// analyzeExtension 分析扩展
func (e *ProcessingEngine) analyzeExtension(ctx context.Context, request *ProcessingRequest, result *ProcessingResult) error {
	analyzer, err := GetAnalyzer(request.ExtensionType)
	if err != nil {
		// 如果没有分析器，不当作错误返回
		return nil
	}

	analysisResult, err := analyzer.Analyze(result.ParsedData, request.AnalysisConfig)
	if err != nil {
		return err
	}

	result.AnalysisResults = append(result.AnalysisResults, analysisResult)
	return nil
}

// handleExtension 处理扩展
func (e *ProcessingEngine) handleExtension(ctx context.Context, request *ProcessingRequest, result *ProcessingResult) error {
	handlers := GetHandlers(request.ExtensionType)

	// 按优先级排序处理器
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
			return err
		}

		result.Events = append(result.Events, event)

		// 如果处理器要求停止传递，则中断
		if !handlerResult.ContinueProcessing {
			break
		}
	}

	return nil
}

// transformExtension 转换扩展
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

// executeInterceptors 执行拦截器
func (e *ProcessingEngine) executeInterceptors(phase string, request *ProcessingRequest, result *ProcessingResult) error {
	e.mu.RLock()
	interceptors := e.interceptors[phase]
	e.mu.RUnlock()

	for _, interceptor := range interceptors {
		if err := interceptor.Intercept(phase, request, result); err != nil {
			return err
		}
	}

	return nil
}

// GetConfig 获取引擎配置
func (e *ProcessingEngine) GetConfig() *EngineConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// 返回配置副本
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

// SetConfig 设置引擎配置
func (e *ProcessingEngine) SetConfig(config *EngineConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.config = config
	return nil
}
