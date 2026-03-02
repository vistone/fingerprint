// Package plugins 实现插件注册表
package plugins

import (
	"fmt"
	"sync"
)

// Registry 插件注册表
type Registry struct {
	plugins map[string]*PluginInfo
	mu      sync.RWMutex
}

// NewRegistry 创建注册表
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]*PluginInfo),
	}
}

// Register 注册插件
func (r *Registry) Register(id string, plugin Plugin, source PluginSource) error {
	if id == "" || plugin == nil {
		return fmt.Errorf("invalid parameters")
	}

	if err := plugin.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	meta := plugin.Metadata()
	r.plugins[id] = &PluginInfo{
		ID:      id,
		Version: meta.Version,
		Source:  source,
		Loaded:  true,
		Meta:    meta,
		Plugin:  plugin,
	}

	return nil
}

// Get 获取插件
func (r *Registry) Get(id string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.plugins[id]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", id)
	}

	return info.Plugin, nil
}

// List 列出所有插件
func (r *Registry) List() map[string]*PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*PluginInfo)
	for id, info := range r.plugins {
		result[id] = info
	}
	return result
}

// ListByCategory 按分类列出
func (r *Registry) ListByCategory(category FingerprintCategory) map[string]*PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*PluginInfo)
	for id, info := range r.plugins {
		if info.Meta != nil && info.Meta.Category == category {
			result[id] = info
		}
	}
	return result
}

// Count 获取数量
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.plugins)
}

// Exists 检查是否存在
func (r *Registry) Exists(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.plugins[id]
	return exists
}

// Unregister 注销插件
func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[id]; !exists {
		return fmt.Errorf("plugin %s not found", id)
	}

	delete(r.plugins, id)
	return nil
}
