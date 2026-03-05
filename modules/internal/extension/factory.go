package extension

import (
	"fmt"
	"sync"
)

// ParserFactory 解析器工厂
type ParserFactory struct {
	mu       sync.RWMutex
	builders map[ExtensionType]func() (Parser, error)
}

var parserFactory = &ParserFactory{
	builders: make(map[ExtensionType]func() (Parser, error)),
}

// RegisterParserBuilder 注册解析器生成函数
func RegisterParserBuilder(extType ExtensionType, builder func() (Parser, error)) error {
	parserFactory.mu.Lock()
	defer parserFactory.mu.Unlock()

	if builder == nil {
		return ErrInvalidParser
	}

	parserFactory.builders[extType] = builder
	return nil
}

// CreateParser 创建解析器实例
func CreateParser(extType ExtensionType) (Parser, error) {
	parserFactory.mu.RLock()
	builder, exists := parserFactory.builders[extType]
	parserFactory.mu.RUnlock()

	if !exists {
		return nil, ErrParserNotFound
	}

	return builder()
}

// AnalyzerFactory 分析器工厂
type AnalyzerFactory struct {
	mu       sync.RWMutex
	builders map[ExtensionType]func() (Analyzer, error)
}

var analyzerFactory = &AnalyzerFactory{
	builders: make(map[ExtensionType]func() (Analyzer, error)),
}

// RegisterAnalyzerBuilder 注册分析器生成函数
func RegisterAnalyzerBuilder(extType ExtensionType, builder func() (Analyzer, error)) error {
	analyzerFactory.mu.Lock()
	defer analyzerFactory.mu.Unlock()

	if builder == nil {
		return ErrInvalidAnalyzer
	}

	analyzerFactory.builders[extType] = builder
	return nil
}

// CreateAnalyzer 创建分析器实例
func CreateAnalyzer(extType ExtensionType) (Analyzer, error) {
	analyzerFactory.mu.RLock()
	builder, exists := analyzerFactory.builders[extType]
	analyzerFactory.mu.RUnlock()

	if !exists {
		return nil, ErrAnalyzerNotFound
	}

	return builder()
}

// HandlerFactory 处理器工厂
type HandlerFactory struct {
	mu       sync.RWMutex
	builders map[string]func() (Handler, error)
}

var handlerFactory = &HandlerFactory{
	builders: make(map[string]func() (Handler, error)),
}

// RegisterHandlerBuilder 注册处理器生成函数
// name: 处理器唯一标识名称
func RegisterHandlerBuilder(name string, builder func() (Handler, error)) error {
	handlerFactory.mu.Lock()
	defer handlerFactory.mu.Unlock()

	if builder == nil {
		return ErrInvalidHandler
	}

	handlerFactory.builders[name] = builder
	return nil
}

// CreateHandler 创建处理器实例
func CreateHandler(name string) (Handler, error) {
	handlerFactory.mu.RLock()
	builder, exists := handlerFactory.builders[name]
	handlerFactory.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("handler not found: %s", name)
	}

	return builder()
}

// TransformFactory 转换工厂
type TransformFactory struct {
	mu       sync.RWMutex
	builders map[string]func() (Transform, error)
}

var transformFactory = &TransformFactory{
	builders: make(map[string]func() (Transform, error)),
}

// RegisterTransformBuilder 注册转换生成函数
func RegisterTransformBuilder(name string, builder func() (Transform, error)) error {
	transformFactory.mu.Lock()
	defer transformFactory.mu.Unlock()

	if builder == nil {
		return fmt.Errorf("invalid transform builder")
	}

	transformFactory.builders[name] = builder
	return nil
}

// CreateTransform 创建转换实例
func CreateTransform(name string) (Transform, error) {
	transformFactory.mu.RLock()
	builder, exists := transformFactory.builders[name]
	transformFactory.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("transform not found: %s", name)
	}

	return builder()
}

// ValidatorFactory 验证器工厂
type ValidatorFactory struct {
	mu       sync.RWMutex
	builders map[string]func() (Validator, error)
}

var validatorFactory = &ValidatorFactory{
	builders: make(map[string]func() (Validator, error)),
}

// RegisterValidatorBuilder 注册验证器生成函数
func RegisterValidatorBuilder(name string, builder func() (Validator, error)) error {
	validatorFactory.mu.Lock()
	defer validatorFactory.mu.Unlock()

	if builder == nil {
		return fmt.Errorf("invalid validator builder")
	}

	validatorFactory.builders[name] = builder
	return nil
}

// CreateValidator 创建验证器实例
func CreateValidator(name string) (Validator, error) {
	validatorFactory.mu.RLock()
	builder, exists := validatorFactory.builders[name]
	validatorFactory.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("validator not found: %s", name)
	}

	return builder()
}

// ComparerFactory 比较器工厂
type ComparerFactory struct {
	mu       sync.RWMutex
	builders map[string]func() (Comparer, error)
}

var comparerFactory = &ComparerFactory{
	builders: make(map[string]func() (Comparer, error)),
}

// PluginFactory 插件工厂
type PluginFactory struct {
	mu       sync.RWMutex
	builders map[string]func() (Plugin, error)
}

var pluginFactory = &PluginFactory{
	builders: make(map[string]func() (Plugin, error)),
}

// RegisterComparerBuilder 注册比较器生成函数
func RegisterComparerBuilder(name string, builder func() (Comparer, error)) error {
	comparerFactory.mu.Lock()
	defer comparerFactory.mu.Unlock()

	if builder == nil {
		return fmt.Errorf("invalid comparer builder")
	}

	comparerFactory.builders[name] = builder
	return nil
}

// CreateComparer 创建比较器实例
func CreateComparer(name string) (Comparer, error) {
	comparerFactory.mu.RLock()
	builder, exists := comparerFactory.builders[name]
	comparerFactory.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("comparer not found: %s", name)
	}

	return builder()
}

// RegisterPluginBuilder 注册插件生成函数
func RegisterPluginBuilder(name string, builder func() (Plugin, error)) error {
	pluginFactory.mu.Lock()
	defer pluginFactory.mu.Unlock()

	if builder == nil {
		return ErrInvalidPlugin
	}

	pluginFactory.builders[name] = builder
	return nil
}

// CreatePlugin 创建插件实例
func CreatePlugin(name string) (Plugin, error) {
	pluginFactory.mu.RLock()
	builder, exists := pluginFactory.builders[name]
	pluginFactory.mu.RUnlock()

	if !exists {
		return nil, ErrPluginNotFound
	}

	return builder()
}

// ExtensionBuilder 通用扩展构建器
// 组合使用所有工厂来构建完整的扩展处理流程
type ExtensionBuilder struct {
	extType    ExtensionType
	parser     Parser
	analyzers  []Analyzer
	handlers   []Handler
	transforms []Transform
	validators []Validator
}

// NewExtensionBuilder 创建扩展构建器
func NewExtensionBuilder(extType ExtensionType) *ExtensionBuilder {
	return &ExtensionBuilder{
		extType:    extType,
		analyzers:  []Analyzer{},
		handlers:   []Handler{},
		transforms: []Transform{},
		validators: []Validator{},
	}
}

// WithParser 设置解析器
func (b *ExtensionBuilder) WithParser(parser Parser) *ExtensionBuilder {
	b.parser = parser
	return b
}

// WithAnalyzer 添加分析器
func (b *ExtensionBuilder) WithAnalyzer(analyzer Analyzer) *ExtensionBuilder {
	if analyzer != nil {
		b.analyzers = append(b.analyzers, analyzer)
	}
	return b
}

// WithHandler 添加处理器
func (b *ExtensionBuilder) WithHandler(handler Handler) *ExtensionBuilder {
	if handler != nil {
		b.handlers = append(b.handlers, handler)
	}
	return b
}

// WithTransform 添加转换
func (b *ExtensionBuilder) WithTransform(transform Transform) *ExtensionBuilder {
	if transform != nil {
		b.transforms = append(b.transforms, transform)
	}
	return b
}

// WithValidator 添加验证器
func (b *ExtensionBuilder) WithValidator(validator Validator) *ExtensionBuilder {
	if validator != nil {
		b.validators = append(b.validators, validator)
	}
	return b
}

// GetExtensionType 获取扩展类型
func (b *ExtensionBuilder) GetExtensionType() ExtensionType {
	return b.extType
}

// GetParser 获取解析器
func (b *ExtensionBuilder) GetParser() Parser {
	return b.parser
}

// GetAnalyzers 获取所有分析器
func (b *ExtensionBuilder) GetAnalyzers() []Analyzer {
	result := make([]Analyzer, len(b.analyzers))
	copy(result, b.analyzers)
	return result
}

// GetHandlers 获取所有处理器
func (b *ExtensionBuilder) GetHandlers() []Handler {
	result := make([]Handler, len(b.handlers))
	copy(result, b.handlers)
	return result
}

// GetTransforms 获取所有转换
func (b *ExtensionBuilder) GetTransforms() []Transform {
	result := make([]Transform, len(b.transforms))
	copy(result, b.transforms)
	return result
}

// GetValidators 获取所有验证器
func (b *ExtensionBuilder) GetValidators() []Validator {
	result := make([]Validator, len(b.validators))
	copy(result, b.validators)
	return result
}
