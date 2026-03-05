// Package plugin provides a plugin system for the fingerprint gateway
package plugin

import (
	"context"
	"fmt"
	"sync"
)

// PluginType represents the type of plugin
type PluginType string

const (
	// TypeAnalyzer plugins analyze fingerprint data
	TypeAnalyzer PluginType = "analyzer"
	// TypeTransformer plugins transform fingerprint data
	TypeTransformer PluginType = "transformer"
	// TypeExporter plugins export metrics/data
	TypeExporter PluginType = "exporter"
	// TypeValidator plugins validate fingerprint data
	TypeValidator PluginType = "validator"
)

// Plugin interface that all plugins must implement
type Plugin interface {
	// Name returns the unique name of the plugin
	Name() string
	// Type returns the type of the plugin
	Type() PluginType
	// Version returns the plugin version
	Version() string
	// Init initializes the plugin with configuration
	Init(config map[string]interface{}) error
	// Close cleans up plugin resources
	Close() error
}

// AnalyzerPlugin interface for analysis plugins
type AnalyzerPlugin interface {
	Plugin
	// Analyze performs analysis on fingerprint data
	Analyze(ctx context.Context, data interface{}) (*AnalysisResult, error)
}

// AnalysisResult contains analysis results
type AnalysisResult struct {
	Score       float64
	Confidence  float64
	Labels      map[string]string
	Annotations map[string]interface{}
}

// TransformerPlugin interface for transformation plugins
type TransformerPlugin interface {
	Plugin
	// Transform transforms fingerprint data
	Transform(ctx context.Context, data interface{}) (interface{}, error)
}

// ExporterPlugin interface for export plugins
type ExporterPlugin interface {
	Plugin
	// Export exports data
	Export(ctx context.Context, data interface{}) error
}

// ValidatorPlugin interface for validation plugins
type ValidatorPlugin interface {
	Plugin
	// Validate validates fingerprint data
	Validate(ctx context.Context, data interface{}) (*ValidationResult, error)
}

// ValidationResult contains validation results
type ValidationResult struct {
	Valid   bool
	Errors  []string
	Warnings []string
}

// Registry manages plugins
type Registry struct {
	mu       sync.RWMutex
	plugins  map[string]Plugin
	enabled  map[string]bool
	order    []string // Execution order
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
		enabled: make(map[string]bool),
		order:   make([]string, 0),
	}
}

// Register registers a plugin
func (r *Registry) Register(plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := plugin.Name()
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	r.plugins[name] = plugin
	r.enabled[name] = true
	r.order = append(r.order, name)

	return nil
}

// Unregister removes a plugin
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if err := plugin.Close(); err != nil {
		return fmt.Errorf("failed to close plugin %s: %w", name, err)
	}

	delete(r.plugins, name)
	delete(r.enabled, name)

	// Remove from order
	for i, n := range r.order {
		if n == name {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}

	return nil
}

// Get retrieves a plugin by name
func (r *Registry) Get(name string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	return plugin, exists
}

// Enable enables a plugin
func (r *Registry) Enable(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	r.enabled[name] = true
	return nil
}

// Disable disables a plugin
func (r *Registry) Disable(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	r.enabled[name] = false
	return nil
}

// IsEnabled checks if a plugin is enabled
func (r *Registry) IsEnabled(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.enabled[name]
}

// List returns all registered plugins
func (r *Registry) List() []PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]PluginInfo, 0, len(r.plugins))
	for name, plugin := range r.plugins {
		result = append(result, PluginInfo{
			Name:    name,
			Type:    string(plugin.Type()),
			Version: plugin.Version(),
			Enabled: r.enabled[name],
		})
	}

	return result
}

// ListByType returns plugins of a specific type
func (r *Registry) ListByType(pluginType PluginType) []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Plugin, 0)
	for _, name := range r.order {
		plugin := r.plugins[name]
		if plugin.Type() == pluginType && r.enabled[name] {
			result = append(result, plugin)
		}
	}

	return result
}

// PluginInfo contains plugin metadata
type PluginInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Enabled bool   `json:"enabled"`
}

// Manager manages plugin lifecycle and execution
type Manager struct {
	registry *Registry
}

// NewManager creates a new plugin manager
func NewManager() *Manager {
	return &Manager{
		registry: NewRegistry(),
	}
}

// Registry returns the plugin registry
func (m *Manager) Registry() *Registry {
	return m.registry
}

// ExecuteAnalyzers runs all enabled analyzer plugins
func (m *Manager) ExecuteAnalyzers(ctx context.Context, data interface{}) ([]*AnalysisResult, error) {
	plugins := m.registry.ListByType(TypeAnalyzer)
	results := make([]*AnalysisResult, 0, len(plugins))

	for _, plugin := range plugins {
		analyzer, ok := plugin.(AnalyzerPlugin)
		if !ok {
			continue
		}

		result, err := analyzer.Analyze(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("analyzer %s failed: %w", plugin.Name(), err)
		}

		results = append(results, result)
	}

	return results, nil
}

// ExecuteTransformers runs all enabled transformer plugins
func (m *Manager) ExecuteTransformers(ctx context.Context, data interface{}) (interface{}, error) {
	plugins := m.registry.ListByType(TypeTransformer)
	result := data

	for _, plugin := range plugins {
		transformer, ok := plugin.(TransformerPlugin)
		if !ok {
			continue
		}

		var err error
		result, err = transformer.Transform(ctx, result)
		if err != nil {
			return nil, fmt.Errorf("transformer %s failed: %w", plugin.Name(), err)
		}
	}

	return result, nil
}

// ExecuteValidators runs all enabled validator plugins
func (m *Manager) ExecuteValidators(ctx context.Context, data interface{}) (*ValidationResult, error) {
	plugins := m.registry.ListByType(TypeValidator)
	
	allErrors := make([]string, 0)
	allWarnings := make([]string, 0)
	valid := true

	for _, plugin := range plugins {
		validator, ok := plugin.(ValidatorPlugin)
		if !ok {
			continue
		}

		result, err := validator.Validate(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("validator %s failed: %w", plugin.Name(), err)
		}

		if !result.Valid {
			valid = false
		}
		allErrors = append(allErrors, result.Errors...)
		allWarnings = append(allWarnings, result.Warnings...)
	}

	return &ValidationResult{
		Valid:    valid,
		Errors:   allErrors,
		Warnings: allWarnings,
	}, nil
}

// ExecuteExporters runs all enabled exporter plugins
func (m *Manager) ExecuteExporters(ctx context.Context, data interface{}) error {
	plugins := m.registry.ListByType(TypeExporter)

	for _, plugin := range plugins {
		exporter, ok := plugin.(ExporterPlugin)
		if !ok {
			continue
		}

		if err := exporter.Export(ctx, data); err != nil {
			return fmt.Errorf("exporter %s failed: %w", plugin.Name(), err)
		}
	}

	return nil
}

// BasePlugin provides a base implementation of the Plugin interface
type BasePlugin struct {
	name    string
	version string
	pluginType PluginType
}

// Name returns the plugin name
func (p *BasePlugin) Name() string {
	return p.name
}

// Type returns the plugin type
func (p *BasePlugin) Type() PluginType {
	return p.pluginType
}

// Version returns the plugin version
func (p *BasePlugin) Version() string {
	return p.version
}

// Init initializes the plugin (should be overridden)
func (p *BasePlugin) Init(config map[string]interface{}) error {
	return nil
}

// Close cleans up plugin resources (should be overridden)
func (p *BasePlugin) Close() error {
	return nil
}
