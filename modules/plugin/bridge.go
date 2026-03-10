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

// translated comment
type ExtensionType = ie.ExtensionType

// translated comment
type ExtensionMetadata = ie.ExtensionMetadata

// translated comment
type ExtensionData = ie.ExtensionData

// translated comment
type Parser = ie.Parser

// translated comment
type Analyzer = ie.Analyzer

// translated comment
type Handler = ie.Handler

// translated comment
type Plugin = ie.Plugin

// translated comment
type PluginInfo = ie.PluginInfo

// translated comment
type ExtensionEvent = ie.ExtensionEvent

// translated comment
type EventResult = ie.EventResult

// translated comment
type ExtensionRegistry = ie.ExtensionRegistry

// translated comment
func RegisterExtension(metadata *ExtensionMetadata) error {
	return ie.RegisterExtension(metadata)
}

// translated comment
func RegisterParser(extType ExtensionType, parser Parser) error {
	return ie.RegisterParser(extType, parser)
}

// translated comment
func RegisterAnalyzer(extType ExtensionType, analyzer Analyzer) error {
	return ie.RegisterAnalyzer(extType, analyzer)
}

// translated comment
func RegisterHandler(extType ExtensionType, handler Handler) error {
	return ie.RegisterHandler(extType, handler)
}

// translated comment
func RegisterPlugin(name string, plugin Plugin) error {
	return ie.RegisterPlugin(name, plugin)
}

// translated comment
func GetPlugin(name string) (Plugin, error) {
	return ie.GetPlugin(name)
}

// translated comment
func LoadPlugins(configPath string) error {
	return ie.LoadPlugins(configPath)
}

// translated comment
func GetRegistry() *ExtensionRegistry {
	return ie.GetRegistry()
}

// translated comment
func GetRegistryStats() map[string]interface{} {
	return ie.GetRegistryStats()
}
