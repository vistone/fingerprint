package extension

import (
	"context"
	"fmt"
	"sort"

	"github.com/vistone/fingerprint/modules/internal/pipeline"
)

// ========================================================================
// ParseStage: 解析扩展数据
// ========================================================================

// ParseStage 解析扩展数据（从原始 TLS 数据解析为结构化格式）
type ParseStage struct {
	registry RegistryPort
}

// NewParseStage 创建新的 ParseStage
func NewParseStage(registry RegistryPort) *ParseStage {
	return &ParseStage{registry: registry}
}

// GetName 获取阶段名称
func (p *ParseStage) GetName() string {
	return "parse"
}

// GetDependencies 获取依赖的前置阶段
func (p *ParseStage) GetDependencies() []string {
	return []string{} // 解析是第一步，无依赖
}

// Execute 执行阶段
func (p *ParseStage) Execute(ctx context.Context, data *pipeline.StageData) error {
	// 从输入中获取请求信息
	request, ok := data.Input.(*ProcessingRequest)
	if !ok {
		return fmt.Errorf("expected *ProcessingRequest, got %T", data.Input)
	}

	// 从注册表中获取解析器
	parser, err := p.registry.GetParser(request.ExtensionType)
	if err != nil {
		return fmt.Errorf("failed to get parser: %w", err)
	}

	// 解析原始数据
	parsedData, err := parser.Parse(request.RawData, ctx)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// 将解析结果存储在 StageData 中
	data.Context["parsed_data"] = parsedData
	data.Output = map[string]interface{}{
		"parsed_data": parsedData,
	}

	return nil
}

// ========================================================================
// AnalyzeStage: 分析扩展数据
// ========================================================================

// AnalyzeStage 分析扩展数据
type AnalyzeStage struct {
	registry RegistryPort
}

// NewAnalyzeStage 创建新的 AnalyzeStage
func NewAnalyzeStage(registry RegistryPort) *AnalyzeStage {
	return &AnalyzeStage{registry: registry}
}

// GetName 获取阶段名称
func (a *AnalyzeStage) GetName() string {
	return "analyze"
}

// GetDependencies 获取依赖的前置阶段
func (a *AnalyzeStage) GetDependencies() []string {
	return []string{"parse"} // 分析依赖于解析
}

// Execute 执行阶段
func (a *AnalyzeStage) Execute(ctx context.Context, data *pipeline.StageData) error {
	// 从输入中获取请求信息
	request, ok := data.Input.(*ProcessingRequest)
	if !ok {
		return fmt.Errorf("expected *ProcessingRequest, got %T", data.Input)
	}

	// 从前一个阶段的输出中获取已解析的数据
	parsedData, ok := data.Context["parsed_data"].(ExtensionData)
	if !ok {
		return fmt.Errorf("parsed_data not found or invalid in context")
	}

	// 从注册表中获取分析器
	analyzer, err := a.registry.GetAnalyzer(request.ExtensionType)
	if err != nil {
		// 如果没有分析器，不当作错误返回（某些扩展类型可能没有分析器）
		data.Output = map[string]interface{}{
			"analysis_results": []interface{}{},
		}
		return nil
	}

	// 分析已解析的数据
	analysisResult, err := analyzer.Analyze(parsedData, request.AnalysisConfig)
	if err != nil {
		return fmt.Errorf("analyze error: %w", err)
	}

	// 存储分析结果
	data.Context["analysis_result"] = analysisResult
	data.Output = map[string]interface{}{
		"analysis_result": analysisResult,
	}

	return nil
}

// ========================================================================
// TransformStage: 转换扩展数据
// ========================================================================

// TransformStage 转换扩展数据为标准格式
type TransformStage struct {
	registry RegistryPort
}

// NewTransformStage 创建新的 TransformStage
func NewTransformStage(registry RegistryPort) *TransformStage {
	return &TransformStage{registry: registry}
}

// GetName 获取阶段名称
func (t *TransformStage) GetName() string {
	return "transform"
}

// GetDependencies 获取依赖的前置阶段
func (t *TransformStage) GetDependencies() []string {
	return []string{"parse"} // 转换依赖于解析
}

// Execute 执行阶段
func (t *TransformStage) Execute(ctx context.Context, data *pipeline.StageData) error {
	// 从输入中获取请求信息
	request, ok := data.Input.(*ProcessingRequest)
	if !ok {
		return fmt.Errorf("expected *ProcessingRequest, got %T", data.Input)
	}

	// 从前一个阶段的输出中获取已解析的数据
	parsedData, ok := data.Context["parsed_data"].(ExtensionData)
	if !ok {
		return fmt.Errorf("parsed_data not found in context")
	}

	// 如果没有分析配置，跳过转换
	if request.AnalysisConfig == nil {
		data.Output = map[string]interface{}{
			"transformed_data": parsedData,
		}
		return nil
	}

	// 从分析配置中获取转换列表
	rawTransforms, ok := request.AnalysisConfig["transforms"]
	if !ok {
		data.Output = map[string]interface{}{
			"transformed_data": parsedData,
		}
		return nil
	}

	// 解析转换名称列表
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

	// 应用转换（这里简化为存储转换名称，实际实现可能需要调用注册表的转换器）
	data.Context["transforms_applied"] = transformNames
	data.Output = map[string]interface{}{
		"transformed_data":   parsedData,
		"transforms_applied": transformNames,
	}

	return nil
}

// ========================================================================
// HandleStage: 处理扩展
// ========================================================================

// HandleStage 处理扩展的事件
type HandleStage struct {
	registry RegistryPort
}

// NewHandleStage 创建新的 HandleStage
func NewHandleStage(registry RegistryPort) *HandleStage {
	return &HandleStage{registry: registry}
}

// GetName 获取阶段名称
func (h *HandleStage) GetName() string {
	return "handle"
}

// GetDependencies 获取依赖的前置阶段
func (h *HandleStage) GetDependencies() []string {
	return []string{"parse"} // 处理依赖于解析
}

// Execute 执行阶段
func (h *HandleStage) Execute(ctx context.Context, data *pipeline.StageData) error {
	// 从输入中获取请求信息
	request, ok := data.Input.(*ProcessingRequest)
	if !ok {
		return fmt.Errorf("expected *ProcessingRequest, got %T", data.Input)
	}

	// 从前一个阶段的输出中获取已解析的数据
	parsedData, ok := data.Context["parsed_data"].(ExtensionData)
	if !ok {
		return fmt.Errorf("parsed_data not found in context")
	}

	// 从注册表中获取处理器
	handlers := h.registry.GetHandlers(request.ExtensionType)

	// 按优先级排序处理器
	sort.Slice(handlers, func(i, j int) bool {
		return handlers[i].GetPriority() > handlers[j].GetPriority()
	})

	// 执行所有处理器
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

		// 如果处理器要求停止传递，则中断
		if !handlerResult.ContinueProcessing {
			break
		}
	}

	// 存储处理事件
	data.Context["events"] = events
	data.Output = map[string]interface{}{
		"events": events,
	}

	return nil
}
