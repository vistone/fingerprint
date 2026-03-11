// Package plugin is the public API for the fingerprint plugin subsystem.
//
// It bridges three internal subsystems that serve distinct domains:
//
//   - internal/plugin: Generic plugin lifecycle (Analyzer, Transformer, Exporter,
//     Validator) with a Manager for registration and discovery.
//   - internal/extension: TLS extension analysis framework (Parser, Analyzer,
//     Handler) with a global ExtensionRegistry.
//   - internal/plugins: Fingerprint profile data model (FingerprintMetadata,
//     ClientHelloSpec) used internally by the contrib builder; not re-exported here.
//
// Callers should use this package rather than importing internal packages directly.
package plugin

import (
	ie "github.com/vistone/fingerprint/modules/internal/extension"
	ip "github.com/vistone/fingerprint/modules/internal/plugin"
)

// BasePlugin provides common plugin functionality
type BasePlugin = ip.BasePlugin

// Manager manages plugins
type Manager = ip.Manager

// PluginType represents plugin type
type PluginType = ip.PluginType

// Plugin types
const (
	TypeAnalyzer    = ip.TypeAnalyzer
	TypeTransformer = ip.TypeTransformer
	TypeExporter    = ip.TypeExporter
	TypeValidator   = ip.TypeValidator
)

// AnalysisResult contains analysis results
type AnalysisResult = ip.AnalysisResult

// ValidationResult contains validation results
type ValidationResult = ip.ValidationResult

// NewManager creates a new plugin manager
func NewManager() *Manager {
	return ip.NewManager()
}

// ExtensionType represents a TLS extension type identifier
type ExtensionType = ie.ExtensionType

// ExtensionMetadata holds metadata for a registered extension
type ExtensionMetadata = ie.ExtensionMetadata

// ExtensionData is the interface for parsed extension data
type ExtensionData = ie.ExtensionData

// Parser parses raw extension bytes into structured data
type Parser = ie.Parser

// Analyzer analyzes parsed extension data
type Analyzer = ie.Analyzer

// Handler processes extension events in a streaming/event-driven fashion
type Handler = ie.Handler

// Plugin represents a third-party extension plugin
type Plugin = ie.Plugin

// PluginInfo holds plugin metadata and status
type PluginInfo = ie.PluginInfo

// ExtensionEvent represents an event for handler processing
type ExtensionEvent = ie.ExtensionEvent

// EventResult holds the result of event processing
type EventResult = ie.EventResult

// ExtensionRegistry is the global registry for extensions
type ExtensionRegistry = ie.ExtensionRegistry

// RegisterExtension registers extension metadata in the global registry.
func RegisterExtension(metadata *ExtensionMetadata) error {
	return ie.RegisterExtension(metadata)
}

// RegisterParser registers a parser for the given extension type.
func RegisterParser(extType ExtensionType, parser Parser) error {
	return ie.RegisterParser(extType, parser)
}

// RegisterAnalyzer registers an analyzer for the given extension type.
func RegisterAnalyzer(extType ExtensionType, analyzer Analyzer) error {
	return ie.RegisterAnalyzer(extType, analyzer)
}

// RegisterHandler registers a handler for the given extension type.
func RegisterHandler(extType ExtensionType, handler Handler) error {
	return ie.RegisterHandler(extType, handler)
}

// RegisterPlugin registers a named plugin.
func RegisterPlugin(name string, plugin Plugin) error {
	return ie.RegisterPlugin(name, plugin)
}

// GetPlugin retrieves a plugin by name.
func GetPlugin(name string) (Plugin, error) {
	return ie.GetPlugin(name)
}

// LoadPlugins loads plugins from the given config path.
func LoadPlugins(configPath string) error {
	return ie.LoadPlugins(configPath)
}

// GetRegistry returns the global extension registry.
func GetRegistry() *ExtensionRegistry {
	return ie.GetRegistry()
}

// GetRegistryStats returns registry statistics.
func GetRegistryStats() map[string]interface{} {
	return ie.GetRegistryStats()
}
