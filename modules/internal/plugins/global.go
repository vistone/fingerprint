// Package plugins implements global plugin management.
package plugins

import (
	"sync"
)

var (
	globalRegistry *Registry
	initOnce       sync.Once
)

// Init initializes the global system.
func Init() error {
	var err error
	initOnce.Do(func() {
		globalRegistry = NewRegistry()
	})
	return err
}

// GetRegistry gets the global registry.
func GetRegistry() *Registry {
	if globalRegistry == nil {
		Init()
	}
	return globalRegistry
}

// RegisterPlugin registers a plugin.
func RegisterPlugin(id string, plugin Plugin, source PluginSource) error {
	return GetRegistry().Register(id, plugin, source)
}

// GetPlugin gets a plugin.
func GetPlugin(id string) (Plugin, error) {
	return GetRegistry().Get(id)
}

// Reset resets the system (for testing).
func Reset() {
	globalRegistry = nil
	initOnce = sync.Once{}
}
