// Package plugins 实现全局插件管理
package plugins

import (
	"sync"
)

var (
	globalRegistry *Registry
	initOnce       sync.Once
)

// Init 初始化全局系统
func Init() error {
	var err error
	initOnce.Do(func() {
		globalRegistry = NewRegistry()
	})
	return err
}

// GetRegistry 获取全局注册表
func GetRegistry() *Registry {
	if globalRegistry == nil {
		Init()
	}
	return globalRegistry
}

// RegisterPlugin 注册插件
func RegisterPlugin(id string, plugin Plugin, source PluginSource) error {
	return GetRegistry().Register(id, plugin, source)
}

// GetPlugin 获取插件
func GetPlugin(id string) (Plugin, error) {
	return GetRegistry().Get(id)
}

// Reset 重置系统（用于测试）
func Reset() {
	globalRegistry = nil
	initOnce = sync.Once{}
}
