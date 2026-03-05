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

// ExtensionType 扩展类型。
type ExtensionType = ie.ExtensionType

// ExtensionMetadata 扩展元数据。
type ExtensionMetadata = ie.ExtensionMetadata

// ExtensionData 扩展数据接口。
type ExtensionData = ie.ExtensionData

// Parser 解析器接口。
type Parser = ie.Parser

// Analyzer 分析器接口。
type Analyzer = ie.Analyzer

// Handler 处理器接口。
type Handler = ie.Handler

// Plugin 插件接口。
type Plugin = ie.Plugin

// PluginInfo 插件信息。
type PluginInfo = ie.PluginInfo

// ExtensionEvent 扩展事件。
type ExtensionEvent = ie.ExtensionEvent

// EventResult 事件处理结果。
type EventResult = ie.EventResult

// ExtensionRegistry 扩展注册表。
type ExtensionRegistry = ie.ExtensionRegistry

// RegisterExtension 注册扩展元数据。
func RegisterExtension(metadata *ExtensionMetadata) error {
	return ie.RegisterExtension(metadata)
}

// RegisterParser 注册解析器。
func RegisterParser(extType ExtensionType, parser Parser) error {
	return ie.RegisterParser(extType, parser)
}

// RegisterAnalyzer 注册分析器。
func RegisterAnalyzer(extType ExtensionType, analyzer Analyzer) error {
	return ie.RegisterAnalyzer(extType, analyzer)
}

// RegisterHandler 注册处理器。
func RegisterHandler(extType ExtensionType, handler Handler) error {
	return ie.RegisterHandler(extType, handler)
}

// RegisterPlugin 注册插件。
func RegisterPlugin(name string, plugin Plugin) error {
	return ie.RegisterPlugin(name, plugin)
}

// GetPlugin 获取插件。
func GetPlugin(name string) (Plugin, error) {
	return ie.GetPlugin(name)
}

// LoadPlugins 从配置加载插件。
func LoadPlugins(configPath string) error {
	return ie.LoadPlugins(configPath)
}

// GetRegistry 获取全局扩展注册表。
func GetRegistry() *ExtensionRegistry {
	return ie.GetRegistry()
}

// GetRegistryStats 获取注册表统计。
func GetRegistryStats() map[string]interface{} {
	return ie.GetRegistryStats()
}
