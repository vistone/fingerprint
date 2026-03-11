// Package plugins implements the fingerprint plugin registry.
package plugins

import (
	"fmt"
	"sync"
)

// Registry is a thread-safe registry for fingerprint plugins.
type Registry struct {
	plugins map[string]*PluginInfo
	mu      sync.RWMutex
}

// NewRegistry creates a new empty plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]*PluginInfo),
	}
}

// Register adds a plugin to the registry with the given ID and source.
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

// Get retrieves a plugin by ID.
func (r *Registry) Get(id string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.plugins[id]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", id)
	}

	return info.Plugin, nil
}

// List returns all registered plugins.
func (r *Registry) List() map[string]*PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*PluginInfo)
	for id, info := range r.plugins {
		result[id] = info
	}
	return result
}

// ListByCategory returns plugins matching the given category.
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

// Count returns the number of registered plugins.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.plugins)
}

// Exists reports whether a plugin with the given ID is registered.
func (r *Registry) Exists(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.plugins[id]
	return exists
}

// Unregister removes a plugin from the registry.
func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[id]; !exists {
		return fmt.Errorf("plugin %s not found", id)
	}

	delete(r.plugins, id)
	return nil
}
