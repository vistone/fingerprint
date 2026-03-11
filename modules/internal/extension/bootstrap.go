//go:build bootstrap
// +build bootstrap

package extension

import (
	"fmt"
)

// ApplicationBootstrapper is the application-level bootstrapper
// It coordinates dependency initialization, ensuring the single composition root principle
// Call the factory functions provided by this package at application startup, instead of calling various New* functions directly
type ApplicationBootstrapper struct {
	configProvider ConfigProvider
	rulesProvider  RulesProvider
	registryPort   RegistryPort

	// cached singletons
	appConfig *Config
	container *Container
	engine    *ProcessingEngine

	// diagnostics tool
	registryDiagnostics *RegistryDiagnostics
}

// NewApplicationBootstrapper creates an application bootstrapper (using default providers)
func NewApplicationBootstrapper() *ApplicationBootstrapper {
	return NewApplicationBootstrapperWithProviders(nil, nil, nil)
}

// NewApplicationBootstrapperWithProviders creates an application bootstrapper (with provider injection support)
func NewApplicationBootstrapperWithProviders(
	configProvider ConfigProvider,
	rulesProvider RulesProvider,
	registryPort RegistryPort,
) *ApplicationBootstrapper {
	if configProvider == nil {
		configProvider = defaultConfigProvider{}
	}
	if rulesProvider == nil {
		rulesProvider = defaultRulesProvider{}
	}
	if registryPort == nil {
		registryPort = globalRegistryPort{}
	}

	return &ApplicationBootstrapper{
		configProvider:      configProvider,
		rulesProvider:       rulesProvider,
		registryPort:        registryPort,
		registryDiagnostics: NewRegistryDiagnostics(nil),
	}
}

// BootstrapConfig initializes the application configuration
// Called once, the result is cached; returns the cached value if already initialized
func (b *ApplicationBootstrapper) BootstrapConfig() (*Config, error) {
	if b.appConfig != nil {
		return b.appConfig, nil
	}

	b.appConfig = NewUnifiedConfigFromEnvWithProviders(
		b.configProvider,
		b.rulesProvider,
	)

	if err := b.appConfig.Validate(); err != nil {
		return nil, err
	}

	return b.appConfig, nil
}

// BootstrapContainer initializes the DI container
// Called once, the result is cached; returns the cached value if already initialized
func (b *ApplicationBootstrapper) BootstrapContainer() (*Container, error) {
	if b.container != nil {
		return b.container, nil
	}

	config, err := b.BootstrapConfig()
	if err != nil {
		return nil, err
	}

	b.container = NewContainerWithProviders(
		config,
		b.configProvider,
		b.rulesProvider,
	)

	if err := b.container.Initialize(); err != nil {
		return nil, err
	}

	return b.container, nil
}

// BootstrapEngine initializes the processing engine
// Called once, the result is cached; returns the cached value if already initialized
func (b *ApplicationBootstrapper) BootstrapEngine(
	engineConfig *EngineConfig,
) (*ProcessingEngine, error) {
	if b.engine != nil {
		return b.engine, nil
	}

	if engineConfig == nil {
		engineConfig = &EngineConfig{
			ConcurrentProcessing: true,
			MaxConcurrency:       4,
			EnableCaching:        true,
			CacheSize:            256,
			TimeoutMs:            5000,
			StrictValidation:     true,
			VerboseLogging:       false,
			CustomConfig:         make(map[string]interface{}),
		}
	}

	b.engine = NewProcessingEngineWithRegistry(engineConfig, b.registryPort)
	return b.engine, nil
}

// GetConfig returns the cached application configuration (initializes if not yet initialized)
func (b *ApplicationBootstrapper) GetConfig() (*Config, error) {
	if b.appConfig == nil {
		return b.BootstrapConfig()
	}
	return b.appConfig, nil
}

// GetContainer returns the cached DI container (initializes if not yet initialized)
func (b *ApplicationBootstrapper) GetContainer() (*Container, error) {
	if b.container == nil {
		return b.BootstrapContainer()
	}
	return b.container, nil
}

// GetEngine returns the cached processing engine (initializes with default config if not yet initialized)
func (b *ApplicationBootstrapper) GetEngine() (*ProcessingEngine, error) {
	if b.engine == nil {
		return b.BootstrapEngine(nil)
	}
	return b.engine, nil
}

// Reset clears all cached instances (only for testing or special scenarios)
func (b *ApplicationBootstrapper) Reset() {
	b.appConfig = nil
	b.container = nil
	b.engine = nil
	if b.container != nil {
		b.container.Reset()
	}
}

// SetConfigProvider sets the configuration provider (must be called before any bootstrap call)
func (b *ApplicationBootstrapper) SetConfigProvider(provider ConfigProvider) {
	if b.appConfig == nil && b.container == nil && b.engine == nil {
		if provider != nil {
			b.configProvider = provider
		}
	}
}

// SetRulesProvider sets the rules provider (must be called before any bootstrap call)
func (b *ApplicationBootstrapper) SetRulesProvider(provider RulesProvider) {
	if b.appConfig == nil && b.container == nil && b.engine == nil {
		if provider != nil {
			b.rulesProvider = provider
		}
	}
}

// SetRegistryPort sets the registry port (must be called before any bootstrap call)
func (b *ApplicationBootstrapper) SetRegistryPort(port RegistryPort) {
	if b.engine == nil {
		if port != nil {
			b.registryPort = port
		}
	}
}

// GetRegistryDiagnostics returns the registry diagnostics tool
func (b *ApplicationBootstrapper) GetRegistryDiagnostics() *RegistryDiagnostics {
	return b.registryDiagnostics
}

// ValidateStartup performs startup validation, checking whether required extensions are loaded
// Returns diagnostic results and any detected issues
func (b *ApplicationBootstrapper) ValidateStartup(requiredExtensions []ExtensionType) (healthy bool, report string, issues []string) {
	if b.registryDiagnostics == nil {
		b.registryDiagnostics = NewRegistryDiagnostics(nil)
	}

	// Perform health check
	healthy, diagIssues := b.registryDiagnostics.HealthCheck()

	// Validate required extensions
	if len(requiredExtensions) > 0 {
		if valid, missing := b.registryDiagnostics.ValidateRequiredExtensions(requiredExtensions); !valid {
			healthy = false
			for _, ext := range missing {
				diagIssues = append(diagIssues, fmt.Sprintf("required extension not fully loaded: type=%d", ext))
			}
		}
	}

	// Generate diagnostic report
	report = b.registryDiagnostics.GetDiagnosticReport()

	return healthy, report, diagIssues
}

// GetDiagnosticReport returns the complete diagnostic report
func (b *ApplicationBootstrapper) GetDiagnosticReport() string {
	if b.registryDiagnostics == nil {
		b.registryDiagnostics = NewRegistryDiagnostics(nil)
	}
	return b.registryDiagnostics.GetDiagnosticReport()
}

// BootstrapWithValidation bootstraps the application with validation
// Checks whether required extensions are loaded before startup
func (b *ApplicationBootstrapper) BootstrapWithValidation(requiredExtensions []ExtensionType) error {
	// Bootstrap first
	if _, err := b.BootstrapConfig(); err != nil {
		return err
	}
	if _, err := b.BootstrapContainer(); err != nil {
		return err
	}

	// Then validate
	if len(requiredExtensions) > 0 {
		if valid, missing := b.registryDiagnostics.ValidateRequiredExtensions(requiredExtensions); !valid {
			missingStr := make([]string, 0, len(missing))
			for _, ext := range missing {
				missingStr = append(missingStr, fmt.Sprintf("type=%d", ext))
			}
			return NewErrorWithCause(ErrCodeMissingConfig,
				fmt.Sprintf("required extensions not fully loaded: %v", missingStr), nil)
		}
	}

	return nil
}
