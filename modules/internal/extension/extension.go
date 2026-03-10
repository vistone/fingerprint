package extension

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// translated comment
type ExtensionType uint16

// translated comment
type ExtensionMetadata struct {
	// translated comment
	Type ExtensionType

	// translated comment
	Name string

	// translated comment
	Description string

	// translated comment
	RFC string

	// translated comment
	IANANumber uint16

	// translated comment
	LastUpdated string

	// translated comment
	Category string

	// translated comment
	IsExperimental bool

	// translated comment
	CompatibleTLSVersions []uint16
}

// translated comment
type ExtensionData interface {
	// translated comment
	GetType() ExtensionType

	// translated comment
	GetRawData() []byte

	// translated comment
	GetName() string

	// translated comment
	ToMap() map[string]interface{}
}

// translated comment
type AnalysisResult interface {
	// translated comment
	GetExtensionType() ExtensionType

	// translated comment
	HasAnomalies() bool

	// translated comment
	GetAnomalies() []string

	// translated comment
	GetRiskScore() float64

	// translated comment
	ToMap() map[string]interface{}
}

// translated comment
type ExtensionRegistry struct {
	mu            sync.RWMutex
	metadata      map[ExtensionType]*ExtensionMetadata
	parsers       map[ExtensionType]Parser
	analyzers     map[ExtensionType]Analyzer
	handlers      map[ExtensionType][]Handler
	typeNames     map[string]ExtensionType // translated comment
	customPlugins map[string]Plugin        // translated comment
}

// Global registry instance
var globalRegistry *ExtensionRegistry

func init() {
	globalRegistry = &ExtensionRegistry{
		metadata:      make(map[ExtensionType]*ExtensionMetadata),
		parsers:       make(map[ExtensionType]Parser),
		analyzers:     make(map[ExtensionType]Analyzer),
		handlers:      make(map[ExtensionType][]Handler),
		typeNames:     make(map[string]ExtensionType),
		customPlugins: make(map[string]Plugin),
	}

	// translated comment
	initStandardExtensions()
}

// translated comment
func RegisterExtension(metadata *ExtensionMetadata) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	if metadata == nil {
		return ErrInvalidMetadata
	}

	globalRegistry.metadata[metadata.Type] = metadata
	globalRegistry.typeNames[metadata.Name] = metadata.Type

	return nil
}

// translated comment
func RegisterParser(extType ExtensionType, parser Parser) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	if parser == nil {
		return ErrInvalidParser
	}

	if _, exists := globalRegistry.metadata[extType]; !exists {
		return ErrExtensionNotRegistered
	}

	globalRegistry.parsers[extType] = parser
	return nil
}

// translated comment
func RegisterAnalyzer(extType ExtensionType, analyzer Analyzer) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	if analyzer == nil {
		return ErrInvalidAnalyzer
	}

	if _, exists := globalRegistry.metadata[extType]; !exists {
		return ErrExtensionNotRegistered
	}

	globalRegistry.analyzers[extType] = analyzer
	return nil
}

// translated comment
func RegisterHandler(extType ExtensionType, handler Handler) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	if handler == nil {
		return ErrInvalidHandler
	}

	if _, exists := globalRegistry.metadata[extType]; !exists {
		return ErrExtensionNotRegistered
	}

	globalRegistry.handlers[extType] = append(globalRegistry.handlers[extType], handler)
	return nil
}

// translated comment
func GetMetadata(extType ExtensionType) (*ExtensionMetadata, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	metadata, exists := globalRegistry.metadata[extType]
	if !exists {
		return nil, ErrExtensionNotFound
	}

	return metadata, nil
}

// translated comment
func GetParser(extType ExtensionType) (Parser, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	parser, exists := globalRegistry.parsers[extType]
	if !exists {
		return nil, ErrParserNotFound
	}

	return parser, nil
}

// translated comment
func GetAnalyzer(extType ExtensionType) (Analyzer, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	analyzer, exists := globalRegistry.analyzers[extType]
	if !exists {
		return nil, ErrAnalyzerNotFound
	}

	return analyzer, nil
}

// translated comment
func GetHandlers(extType ExtensionType) []Handler {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	handlers := globalRegistry.handlers[extType]
	// translated comment
	result := make([]Handler, len(handlers))
	copy(result, handlers)
	return result
}

// translated comment
func ListAllExtensions() []*ExtensionMetadata {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	extensions := make([]*ExtensionMetadata, 0, len(globalRegistry.metadata))
	for _, metadata := range globalRegistry.metadata {
		extensions = append(extensions, metadata)
	}

	return extensions
}

// translated comment
func FindExtensionByName(name string) (ExtensionType, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	extType, exists := globalRegistry.typeNames[name]
	if !exists {
		return 0, ErrExtensionNotFound
	}

	return extType, nil
}

// translated comment
func RegisterPlugin(name string, plugin Plugin) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	if plugin == nil {
		return ErrInvalidPlugin
	}

	// translated comment
	if err := plugin.Validate(); err != nil {
		return err
	}

	globalRegistry.customPlugins[name] = plugin
	return nil
}

// translated comment
func GetPlugin(name string) (Plugin, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	plugin, exists := globalRegistry.customPlugins[name]
	if !exists {
		return nil, ErrPluginNotFound
	}

	return plugin, nil
}

// translated comment
func LoadPlugins(configPath string) error {
	type pluginConfigItem struct {
		Name    string                 `json:"name"`
		Enabled *bool                  `json:"enabled,omitempty"`
		Config  map[string]interface{} `json:"config,omitempty"`
	}

	type pluginLoadConfig struct {
		Plugins []pluginConfigItem `json:"plugins"`
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return NewErrorWithCause(ErrCodeMissingConfig,
			fmt.Sprintf("failed to read plugin config: %s", configPath), err)
	}

	var cfg pluginLoadConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return NewErrorWithCause(ErrCodeInvalidConfig,
			"failed to parse plugin config json", err)
	}

	for _, item := range cfg.Plugins {
		if item.Name == "" {
			return NewError(ErrCodeInvalidConfig, "plugin name cannot be empty")
		}

		enabled := true
		if item.Enabled != nil {
			enabled = *item.Enabled
		}
		if !enabled {
			continue
		}

		plugin, err := CreatePlugin(item.Name)
		if err != nil {
			return NewErrorWithCause(ErrCodePluginNotFound,
				fmt.Sprintf("failed to create plugin: %s", item.Name), err)
		}

		if item.Config == nil {
			item.Config = map[string]interface{}{}
		}

		if err := plugin.Init(item.Config); err != nil {
			return NewErrorWithCause(ErrCodePluginInitFailed,
				fmt.Sprintf("failed to init plugin: %s", item.Name), err)
		}

		if err := plugin.Register(); err != nil {
			return NewErrorWithCause(ErrCodePluginLoadFailed,
				fmt.Sprintf("failed to register plugin: %s", item.Name), err)
		}

		if err := RegisterPlugin(item.Name, plugin); err != nil {
			return NewErrorWithCause(ErrCodePluginLoadFailed,
				fmt.Sprintf("failed to add plugin to registry: %s", item.Name), err)
		}
	}

	return nil
}

// translated comment
func GetRegistry() *ExtensionRegistry {
	return globalRegistry
}

// translated comment
func GetRegistryStats() map[string]interface{} {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	return map[string]interface{}{
		"total_extensions":     len(globalRegistry.metadata),
		"registered_parsers":   len(globalRegistry.parsers),
		"registered_analyzers": len(globalRegistry.analyzers),
		"registered_handlers":  len(globalRegistry.handlers),
		"custom_plugins":       len(globalRegistry.customPlugins),
	}
}
