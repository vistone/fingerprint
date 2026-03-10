package extension

import (
	"fmt"
	"sync"
)

// translated comment
type ParserFactory struct {
	mu       sync.RWMutex
	builders map[ExtensionType]func() (Parser, error)
}

var parserFactory = &ParserFactory{
	builders: make(map[ExtensionType]func() (Parser, error)),
}

// translated comment
func RegisterParserBuilder(extType ExtensionType, builder func() (Parser, error)) error {
	parserFactory.mu.Lock()
	defer parserFactory.mu.Unlock()

	if builder == nil {
		return ErrInvalidParser
	}

	parserFactory.builders[extType] = builder
	return nil
}

// translated comment
func CreateParser(extType ExtensionType) (Parser, error) {
	parserFactory.mu.RLock()
	builder, exists := parserFactory.builders[extType]
	parserFactory.mu.RUnlock()

	if !exists {
		return nil, ErrParserNotFound
	}

	return builder()
}

// translated comment
type AnalyzerFactory struct {
	mu       sync.RWMutex
	builders map[ExtensionType]func() (Analyzer, error)
}

var analyzerFactory = &AnalyzerFactory{
	builders: make(map[ExtensionType]func() (Analyzer, error)),
}

// translated comment
func RegisterAnalyzerBuilder(extType ExtensionType, builder func() (Analyzer, error)) error {
	analyzerFactory.mu.Lock()
	defer analyzerFactory.mu.Unlock()

	if builder == nil {
		return ErrInvalidAnalyzer
	}

	analyzerFactory.builders[extType] = builder
	return nil
}

// translated comment
func CreateAnalyzer(extType ExtensionType) (Analyzer, error) {
	analyzerFactory.mu.RLock()
	builder, exists := analyzerFactory.builders[extType]
	analyzerFactory.mu.RUnlock()

	if !exists {
		return nil, ErrAnalyzerNotFound
	}

	return builder()
}

// translated comment
type HandlerFactory struct {
	mu       sync.RWMutex
	builders map[string]func() (Handler, error)
}

var handlerFactory = &HandlerFactory{
	builders: make(map[string]func() (Handler, error)),
}

// translated comment
// translated comment
func RegisterHandlerBuilder(name string, builder func() (Handler, error)) error {
	handlerFactory.mu.Lock()
	defer handlerFactory.mu.Unlock()

	if builder == nil {
		return ErrInvalidHandler
	}

	handlerFactory.builders[name] = builder
	return nil
}

// translated comment
func CreateHandler(name string) (Handler, error) {
	handlerFactory.mu.RLock()
	builder, exists := handlerFactory.builders[name]
	handlerFactory.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("handler not found: %s", name)
	}

	return builder()
}

// translated comment
type TransformFactory struct {
	mu       sync.RWMutex
	builders map[string]func() (Transform, error)
}

var transformFactory = &TransformFactory{
	builders: make(map[string]func() (Transform, error)),
}

// translated comment
func RegisterTransformBuilder(name string, builder func() (Transform, error)) error {
	transformFactory.mu.Lock()
	defer transformFactory.mu.Unlock()

	if builder == nil {
		return fmt.Errorf("invalid transform builder")
	}

	transformFactory.builders[name] = builder
	return nil
}

// translated comment
func CreateTransform(name string) (Transform, error) {
	transformFactory.mu.RLock()
	builder, exists := transformFactory.builders[name]
	transformFactory.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("transform not found: %s", name)
	}

	return builder()
}

// translated comment
type ValidatorFactory struct {
	mu       sync.RWMutex
	builders map[string]func() (Validator, error)
}

var validatorFactory = &ValidatorFactory{
	builders: make(map[string]func() (Validator, error)),
}

// translated comment
func RegisterValidatorBuilder(name string, builder func() (Validator, error)) error {
	validatorFactory.mu.Lock()
	defer validatorFactory.mu.Unlock()

	if builder == nil {
		return fmt.Errorf("invalid validator builder")
	}

	validatorFactory.builders[name] = builder
	return nil
}

// translated comment
func CreateValidator(name string) (Validator, error) {
	validatorFactory.mu.RLock()
	builder, exists := validatorFactory.builders[name]
	validatorFactory.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("validator not found: %s", name)
	}

	return builder()
}

// translated comment
type ComparerFactory struct {
	mu       sync.RWMutex
	builders map[string]func() (Comparer, error)
}

var comparerFactory = &ComparerFactory{
	builders: make(map[string]func() (Comparer, error)),
}

// translated comment
type PluginFactory struct {
	mu       sync.RWMutex
	builders map[string]func() (Plugin, error)
}

var pluginFactory = &PluginFactory{
	builders: make(map[string]func() (Plugin, error)),
}

// translated comment
func RegisterComparerBuilder(name string, builder func() (Comparer, error)) error {
	comparerFactory.mu.Lock()
	defer comparerFactory.mu.Unlock()

	if builder == nil {
		return fmt.Errorf("invalid comparer builder")
	}

	comparerFactory.builders[name] = builder
	return nil
}

// translated comment
func CreateComparer(name string) (Comparer, error) {
	comparerFactory.mu.RLock()
	builder, exists := comparerFactory.builders[name]
	comparerFactory.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("comparer not found: %s", name)
	}

	return builder()
}

// translated comment
func RegisterPluginBuilder(name string, builder func() (Plugin, error)) error {
	pluginFactory.mu.Lock()
	defer pluginFactory.mu.Unlock()

	if builder == nil {
		return ErrInvalidPlugin
	}

	pluginFactory.builders[name] = builder
	return nil
}

// translated comment
func CreatePlugin(name string) (Plugin, error) {
	pluginFactory.mu.RLock()
	builder, exists := pluginFactory.builders[name]
	pluginFactory.mu.RUnlock()

	if !exists {
		return nil, ErrPluginNotFound
	}

	return builder()
}

// translated comment
// translated comment
type ExtensionBuilder struct {
	extType    ExtensionType
	parser     Parser
	analyzers  []Analyzer
	handlers   []Handler
	transforms []Transform
	validators []Validator
}

// translated comment
func NewExtensionBuilder(extType ExtensionType) *ExtensionBuilder {
	return &ExtensionBuilder{
		extType:    extType,
		analyzers:  []Analyzer{},
		handlers:   []Handler{},
		transforms: []Transform{},
		validators: []Validator{},
	}
}

// translated comment
func (b *ExtensionBuilder) WithParser(parser Parser) *ExtensionBuilder {
	b.parser = parser
	return b
}

// translated comment
func (b *ExtensionBuilder) WithAnalyzer(analyzer Analyzer) *ExtensionBuilder {
	if analyzer != nil {
		b.analyzers = append(b.analyzers, analyzer)
	}
	return b
}

// translated comment
func (b *ExtensionBuilder) WithHandler(handler Handler) *ExtensionBuilder {
	if handler != nil {
		b.handlers = append(b.handlers, handler)
	}
	return b
}

// translated comment
func (b *ExtensionBuilder) WithTransform(transform Transform) *ExtensionBuilder {
	if transform != nil {
		b.transforms = append(b.transforms, transform)
	}
	return b
}

// translated comment
func (b *ExtensionBuilder) WithValidator(validator Validator) *ExtensionBuilder {
	if validator != nil {
		b.validators = append(b.validators, validator)
	}
	return b
}

// translated comment
func (b *ExtensionBuilder) GetExtensionType() ExtensionType {
	return b.extType
}

// translated comment
func (b *ExtensionBuilder) GetParser() Parser {
	return b.parser
}

// translated comment
func (b *ExtensionBuilder) GetAnalyzers() []Analyzer {
	result := make([]Analyzer, len(b.analyzers))
	copy(result, b.analyzers)
	return result
}

// translated comment
func (b *ExtensionBuilder) GetHandlers() []Handler {
	result := make([]Handler, len(b.handlers))
	copy(result, b.handlers)
	return result
}

// translated comment
func (b *ExtensionBuilder) GetTransforms() []Transform {
	result := make([]Transform, len(b.transforms))
	copy(result, b.transforms)
	return result
}

// translated comment
func (b *ExtensionBuilder) GetValidators() []Validator {
	result := make([]Validator, len(b.validators))
	copy(result, b.validators)
	return result
}
