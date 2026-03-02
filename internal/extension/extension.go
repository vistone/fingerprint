package extension

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// ExtensionType 定义扩展类型
type ExtensionType uint16

// ExtensionMetadata 扩展元数据
type ExtensionMetadata struct {
	// 扩展类型ID
	Type ExtensionType

	// 扩展名称（如 "Encrypted Client Hello"）
	Name string

	// 扩展描述
	Description string

	// RFC 文档引用
	RFC string

	// IANA 注册编号
	IANANumber uint16

	// 最后更新时间
	LastUpdated string

	// 扩展类别（如 "encryption", "negotiation", "preference"）
	Category string

	// 是否为实验性扩展
	IsExperimental bool

	// 兼容的 TLS 版本
	CompatibleTLSVersions []uint16
}

// ExtensionData 扩展数据接口
type ExtensionData interface {
	// 获取扩展类型
	GetType() ExtensionType

	// 获取原始字节数据
	GetRawData() []byte

	// 获取扩展名称
	GetName() string

	// 转换为 map 用于序列化
	ToMap() map[string]interface{}
}

// AnalysisResult 通用分析结果接口
type AnalysisResult interface {
	// 获取分析的扩展类型
	GetExtensionType() ExtensionType

	// 是否存在异常
	HasAnomalies() bool

	// 获取异常信息
	GetAnomalies() []string

	// 获取风险评分（0.0-1.0）
	GetRiskScore() float64

	// 转换为 map 用于序列化
	ToMap() map[string]interface{}
}

// ExtensionRegistry 全局扩展注册表
type ExtensionRegistry struct {
	mu            sync.RWMutex
	metadata      map[ExtensionType]*ExtensionMetadata
	parsers       map[ExtensionType]Parser
	analyzers     map[ExtensionType]Analyzer
	handlers      map[ExtensionType][]Handler
	typeNames     map[string]ExtensionType // 反向查找
	customPlugins map[string]Plugin        // 第三方插件
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

	// 初始化标准扩展
	initStandardExtensions()
}

// RegisterExtension 注册扩展类型及其元数据
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

// RegisterParser 注册扩展解析器
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

// RegisterAnalyzer 注册扩展分析器
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

// RegisterHandler 注册扩展处理器（事件驱动）
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

// GetMetadata 获取扩展元数据
func GetMetadata(extType ExtensionType) (*ExtensionMetadata, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	metadata, exists := globalRegistry.metadata[extType]
	if !exists {
		return nil, ErrExtensionNotFound
	}

	return metadata, nil
}

// GetParser 获取扩展解析器
func GetParser(extType ExtensionType) (Parser, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	parser, exists := globalRegistry.parsers[extType]
	if !exists {
		return nil, ErrParserNotFound
	}

	return parser, nil
}

// GetAnalyzer 获取扩展分析器
func GetAnalyzer(extType ExtensionType) (Analyzer, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	analyzer, exists := globalRegistry.analyzers[extType]
	if !exists {
		return nil, ErrAnalyzerNotFound
	}

	return analyzer, nil
}

// GetHandlers 获取扩展处理器
func GetHandlers(extType ExtensionType) []Handler {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	handlers := globalRegistry.handlers[extType]
	// 返回副本以避免并发修改
	result := make([]Handler, len(handlers))
	copy(result, handlers)
	return result
}

// ListAllExtensions 列举所有注册的扩展
func ListAllExtensions() []*ExtensionMetadata {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	extensions := make([]*ExtensionMetadata, 0, len(globalRegistry.metadata))
	for _, metadata := range globalRegistry.metadata {
		extensions = append(extensions, metadata)
	}

	return extensions
}

// FindExtensionByName 按名称查找扩展
func FindExtensionByName(name string) (ExtensionType, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	extType, exists := globalRegistry.typeNames[name]
	if !exists {
		return 0, ErrExtensionNotFound
	}

	return extType, nil
}

// RegisterPlugin 注册第三方插件
func RegisterPlugin(name string, plugin Plugin) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	if plugin == nil {
		return ErrInvalidPlugin
	}

	// 验证插件
	if err := plugin.Validate(); err != nil {
		return err
	}

	globalRegistry.customPlugins[name] = plugin
	return nil
}

// GetPlugin 获取第三方插件
func GetPlugin(name string) (Plugin, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	plugin, exists := globalRegistry.customPlugins[name]
	if !exists {
		return nil, ErrPluginNotFound
	}

	return plugin, nil
}

// LoadPlugins 加载配置文件中的所有插件
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

// GetRegistry 获取全局注册表实例
func GetRegistry() *ExtensionRegistry {
	return globalRegistry
}

// GetRegistryStats 获取注册表统计信息
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
