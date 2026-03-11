package extension

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// ExtensionType defines the extension type
type ExtensionType uint16

// ExtensionMetadata holds extension metadata
type ExtensionMetadata struct {
	// Extension type ID
	Type ExtensionType

	// Extension name (e.g. "Encrypted Client Hello")
	Name string

	// Extension description
	Description string

	// RFC document reference
	RFC string

	// IANA registration number
	IANANumber uint16

	// Last update time
	LastUpdated string

	// Extension category (e.g. "encryption", "negotiation", "preference")
	Category string

	// Whether this is an experimental extension
	IsExperimental bool

	// Compatible TLS versions
	CompatibleTLSVersions []uint16
}

// ExtensionData is the extension data interface
type ExtensionData interface {
	// Get extension type
	GetType() ExtensionType

	// Get raw byte data
	GetRawData() []byte

	// Get extension name
	GetName() string

	// Convert to map for serialization
	ToMap() map[string]interface{}
}

// AnalysisResult is the general analysis result interface
type AnalysisResult interface {
	// Get the analyzed extension type
	GetExtensionType() ExtensionType

	// Whether anomalies exist
	HasAnomalies() bool

	// Get anomaly details
	GetAnomalies() []string

	// Get risk score (0.0-1.0)
	GetRiskScore() float64

	// Convert to map for serialization
	ToMap() map[string]interface{}
}

// ExtensionRegistry is the global extension registry
type ExtensionRegistry struct {
	mu            sync.RWMutex
	metadata      map[ExtensionType]*ExtensionMetadata
	parsers       map[ExtensionType]Parser
	analyzers     map[ExtensionType]Analyzer
	handlers      map[ExtensionType][]Handler
	typeNames     map[string]ExtensionType // reverse lookup
	customPlugins map[string]Plugin        // third-party plugins
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

	// Initialize standard extensions
	initStandardExtensions()
}

// RegisterExtension registers an extension type and its metadata
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

// RegisterParser registers an extension parser
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

// RegisterAnalyzer registers an extension analyzer
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

// RegisterHandler registers an extension handler (event-driven)
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

// GetMetadata returns extension metadata
func GetMetadata(extType ExtensionType) (*ExtensionMetadata, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	metadata, exists := globalRegistry.metadata[extType]
	if !exists {
		return nil, ErrExtensionNotFound
	}

	return metadata, nil
}

// GetParser returns the extension parser
func GetParser(extType ExtensionType) (Parser, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	parser, exists := globalRegistry.parsers[extType]
	if !exists {
		return nil, ErrParserNotFound
	}

	return parser, nil
}

// GetAnalyzer returns the extension analyzer
func GetAnalyzer(extType ExtensionType) (Analyzer, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	analyzer, exists := globalRegistry.analyzers[extType]
	if !exists {
		return nil, ErrAnalyzerNotFound
	}

	return analyzer, nil
}

// GetHandlers returns extension handlers
func GetHandlers(extType ExtensionType) []Handler {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	handlers := globalRegistry.handlers[extType]
	// Return a copy to avoid concurrent modification
	result := make([]Handler, len(handlers))
	copy(result, handlers)
	return result
}

// ListAllExtensions lists all registered extensions
func ListAllExtensions() []*ExtensionMetadata {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	extensions := make([]*ExtensionMetadata, 0, len(globalRegistry.metadata))
	for _, metadata := range globalRegistry.metadata {
		extensions = append(extensions, metadata)
	}

	return extensions
}

// FindExtensionByName finds an extension by name
func FindExtensionByName(name string) (ExtensionType, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	extType, exists := globalRegistry.typeNames[name]
	if !exists {
		return 0, ErrExtensionNotFound
	}

	return extType, nil
}

// RegisterPlugin registers a third-party plugin
func RegisterPlugin(name string, plugin Plugin) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	if plugin == nil {
		return ErrInvalidPlugin
	}

	// Validate the plugin
	if err := plugin.Validate(); err != nil {
		return err
	}

	globalRegistry.customPlugins[name] = plugin
	return nil
}

// GetPlugin returns a third-party plugin
func GetPlugin(name string) (Plugin, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	plugin, exists := globalRegistry.customPlugins[name]
	if !exists {
		return nil, ErrPluginNotFound
	}

	return plugin, nil
}

// LoadPlugins loads all plugins from the configuration file
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

// GetRegistry returns the global registry instance
func GetRegistry() *ExtensionRegistry {
	return globalRegistry
}

// GetRegistryStats returns registry statistics
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
