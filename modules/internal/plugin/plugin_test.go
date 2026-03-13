package plugin

import (
	"context"
	"testing"

	"github.com/vistone/fingerprint/modules/internal/testhelpers"
)

// MockPlugin is a mock implementation for testing
type MockPlugin struct {
	BasePlugin
	initCalled  bool
	closeCalled bool
}

func NewMockPlugin(name string, pluginType PluginType) *MockPlugin {
	return &MockPlugin{
		BasePlugin: BasePlugin{
			name:       name,
			version:    "1.0.0",
			pluginType: pluginType,
		},
	}
}

func (m *MockPlugin) Init(config map[string]interface{}) error {
	m.initCalled = true
	return nil
}

func (m *MockPlugin) Close() error {
	m.closeCalled = true
	return nil
}

// MockAnalyzer is a mock analyzer plugin
type MockAnalyzer struct {
	MockPlugin
	analyzeFunc func(ctx context.Context, data interface{}) (*AnalysisResult, error)
}

func NewMockAnalyzer(name string) *MockAnalyzer {
	return &MockAnalyzer{
		MockPlugin: *NewMockPlugin(name, TypeAnalyzer),
	}
}

func (m *MockAnalyzer) Analyze(ctx context.Context, data interface{}) (*AnalysisResult, error) {
	if m.analyzeFunc != nil {
		return m.analyzeFunc(ctx, data)
	}
	return &AnalysisResult{
		Score:      0.5,
		Confidence: 0.8,
		Labels:     map[string]string{"test": "true"},
	}, nil
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	testhelpers.AssertNotNil(t, registry)
	testhelpers.AssertNotNil(t, registry.plugins)
	testhelpers.AssertNotNil(t, registry.enabled)
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	t.Run("register new plugin", func(t *testing.T) {
		plugin := NewMockPlugin("test-plugin", TypeAnalyzer)
		err := registry.Register(plugin)
		testhelpers.AssertNoError(t, err)

		_, found := registry.Get("test-plugin")
		testhelpers.AssertEqual(t, found, true)
	})

	t.Run("register duplicate plugin", func(t *testing.T) {
		plugin := NewMockPlugin("test-plugin-dup", TypeAnalyzer)
		err := registry.Register(plugin)
		testhelpers.AssertNoError(t, err)

		// Try to register again
		err = registry.Register(plugin)
		testhelpers.AssertError(t, err)
	})
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()
	plugin := NewMockPlugin("test-unregister", TypeAnalyzer)

	err := registry.Register(plugin)
	testhelpers.AssertNoError(t, err)

	t.Run("unregister existing plugin", func(t *testing.T) {
		err := registry.Unregister("test-unregister")
		testhelpers.AssertNoError(t, err)

		_, found := registry.Get("test-unregister")
		testhelpers.AssertEqual(t, found, false)
		testhelpers.AssertEqual(t, plugin.closeCalled, true)
	})

	t.Run("unregister non-existing plugin", func(t *testing.T) {
		err := registry.Unregister("non-existing")
		testhelpers.AssertError(t, err)
	})
}

func TestRegistry_EnableDisable(t *testing.T) {
	registry := NewRegistry()
	plugin := NewMockPlugin("test-enable", TypeAnalyzer)

	err := registry.Register(plugin)
	testhelpers.AssertNoError(t, err)

	t.Run("disable plugin", func(t *testing.T) {
		err := registry.Disable("test-enable")
		testhelpers.AssertNoError(t, err)
		testhelpers.AssertEqual(t, registry.IsEnabled("test-enable"), false)
	})

	t.Run("enable plugin", func(t *testing.T) {
		err := registry.Enable("test-enable")
		testhelpers.AssertNoError(t, err)
		testhelpers.AssertEqual(t, registry.IsEnabled("test-enable"), true)
	})

	t.Run("enable non-existing plugin", func(t *testing.T) {
		err := registry.Enable("non-existing")
		testhelpers.AssertError(t, err)
	})
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	plugin1 := NewMockPlugin("plugin-1", TypeAnalyzer)
	plugin2 := NewMockPlugin("plugin-2", TypeTransformer)

	registry.Register(plugin1)
	registry.Register(plugin2)

	t.Run("list all plugins", func(t *testing.T) {
		list := registry.List()
		testhelpers.AssertEqual(t, len(list), 2)
	})
}

func TestRegistry_ListByType(t *testing.T) {
	registry := NewRegistry()

	analyzer := NewMockPlugin("analyzer-1", TypeAnalyzer)
	transformer := NewMockPlugin("transformer-1", TypeTransformer)

	registry.Register(analyzer)
	registry.Register(transformer)

	t.Run("list by type analyzer", func(t *testing.T) {
		list := registry.ListByType(TypeAnalyzer)
		testhelpers.AssertEqual(t, len(list), 1)
		testhelpers.AssertEqual(t, list[0].Name(), "analyzer-1")
	})

	t.Run("list by type transformer", func(t *testing.T) {
		list := registry.ListByType(TypeTransformer)
		testhelpers.AssertEqual(t, len(list), 1)
		testhelpers.AssertEqual(t, list[0].Name(), "transformer-1")
	})

	t.Run("list by type with disabled", func(t *testing.T) {
		registry.Disable("analyzer-1")
		list := registry.ListByType(TypeAnalyzer)
		testhelpers.AssertEqual(t, len(list), 0)
	})
}

func TestNewManager(t *testing.T) {
	manager := NewManager()
	testhelpers.AssertNotNil(t, manager)
	testhelpers.AssertNotNil(t, manager.registry)
}

func TestManager_ExecuteAnalyzers(t *testing.T) {
	manager := NewManager()

	analyzer := NewMockAnalyzer("test-analyzer")
	manager.Registry().Register(analyzer)

	t.Run("execute analyzers", func(t *testing.T) {
		results, err := manager.ExecuteAnalyzers(context.Background(), "test-data")
		testhelpers.AssertNoError(t, err)
		testhelpers.AssertEqual(t, len(results), 1)
		testhelpers.AssertEqual(t, results[0].Score, 0.5)
	})

	t.Run("execute with disabled analyzer", func(t *testing.T) {
		manager.Registry().Disable("test-analyzer")
		results, err := manager.ExecuteAnalyzers(context.Background(), "test-data")
		testhelpers.AssertNoError(t, err)
		testhelpers.AssertEqual(t, len(results), 0)
	})
}

func TestManager_ExecuteValidators(t *testing.T) {
	manager := NewManager()

	t.Run("execute validators - no validators", func(t *testing.T) {
		result, err := manager.ExecuteValidators(context.Background(), "test-data")
		testhelpers.AssertNoError(t, err)
		testhelpers.AssertEqual(t, result.Valid, true)
		testhelpers.AssertEqual(t, len(result.Errors), 0)
	})
}

func TestBasePlugin(t *testing.T) {
	plugin := &BasePlugin{
		name:       "base-test",
		version:    "1.0.0",
		pluginType: TypeAnalyzer,
	}

	t.Run("name", func(t *testing.T) {
		testhelpers.AssertEqual(t, plugin.Name(), "base-test")
	})

	t.Run("version", func(t *testing.T) {
		testhelpers.AssertEqual(t, plugin.Version(), "1.0.0")
	})

	t.Run("type", func(t *testing.T) {
		testhelpers.AssertEqual(t, plugin.Type(), TypeAnalyzer)
	})

	t.Run("init", func(t *testing.T) {
		err := plugin.Init(nil)
		testhelpers.AssertNoError(t, err)
	})

	t.Run("close", func(t *testing.T) {
		err := plugin.Close()
		testhelpers.AssertNoError(t, err)
	})
}

func TestAnalysisResult(t *testing.T) {
	result := &AnalysisResult{
		Score:       0.9,
		Confidence:  0.85,
		Labels:      map[string]string{"browser": "chrome"},
		Annotations: map[string]interface{}{"version": 120},
	}

	testhelpers.AssertEqual(t, result.Score, 0.9)
	testhelpers.AssertEqual(t, result.Confidence, 0.85)
	testhelpers.AssertEqual(t, result.Labels["browser"], "chrome")
}

func TestValidationResult(t *testing.T) {
	result := &ValidationResult{
		Valid:    false,
		Errors:   []string{"error1", "error2"},
		Warnings: []string{"warning1"},
	}

	testhelpers.AssertEqual(t, result.Valid, false)
	testhelpers.AssertEqual(t, len(result.Errors), 2)
	testhelpers.AssertEqual(t, len(result.Warnings), 1)
}

func TestPluginInfo(t *testing.T) {
	info := PluginInfo{
		Name:    "test-plugin",
		Type:    "analyzer",
		Version: "1.0.0",
		Enabled: true,
	}

	testhelpers.AssertEqual(t, info.Name, "test-plugin")
	testhelpers.AssertEqual(t, info.Type, "analyzer")
	testhelpers.AssertEqual(t, info.Version, "1.0.0")
	testhelpers.AssertEqual(t, info.Enabled, true)
}
