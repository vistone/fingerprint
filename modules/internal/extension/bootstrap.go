//go:build bootstrap
// +build bootstrap

package extension

import (
	"fmt"
)

// translated comment
// translated comment
// translated comment
type ApplicationBootstrapper struct {
	configProvider ConfigProvider
	rulesProvider  RulesProvider
	registryPort   RegistryPort

	// translated comment
	appConfig *Config
	container *Container
	engine    *ProcessingEngine

	// translated comment
	registryDiagnostics *RegistryDiagnostics
}

// translated comment
func NewApplicationBootstrapper() *ApplicationBootstrapper {
	return NewApplicationBootstrapperWithProviders(nil, nil, nil)
}

// translated comment
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

// translated comment
// translated comment
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

// translated comment
// translated comment
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

// translated comment
// translated comment
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

// translated comment
func (b *ApplicationBootstrapper) GetConfig() (*Config, error) {
	if b.appConfig == nil {
		return b.BootstrapConfig()
	}
	return b.appConfig, nil
}

// translated comment
func (b *ApplicationBootstrapper) GetContainer() (*Container, error) {
	if b.container == nil {
		return b.BootstrapContainer()
	}
	return b.container, nil
}

// translated comment
func (b *ApplicationBootstrapper) GetEngine() (*ProcessingEngine, error) {
	if b.engine == nil {
		return b.BootstrapEngine(nil)
	}
	return b.engine, nil
}

// translated comment
func (b *ApplicationBootstrapper) Reset() {
	b.appConfig = nil
	b.container = nil
	b.engine = nil
	if b.container != nil {
		b.container.Reset()
	}
}

// translated comment
func (b *ApplicationBootstrapper) SetConfigProvider(provider ConfigProvider) {
	if b.appConfig == nil && b.container == nil && b.engine == nil {
		if provider != nil {
			b.configProvider = provider
		}
	}
}

// translated comment
func (b *ApplicationBootstrapper) SetRulesProvider(provider RulesProvider) {
	if b.appConfig == nil && b.container == nil && b.engine == nil {
		if provider != nil {
			b.rulesProvider = provider
		}
	}
}

// translated comment
func (b *ApplicationBootstrapper) SetRegistryPort(port RegistryPort) {
	if b.engine == nil {
		if port != nil {
			b.registryPort = port
		}
	}
}

// translated comment
func (b *ApplicationBootstrapper) GetRegistryDiagnostics() *RegistryDiagnostics {
	return b.registryDiagnostics
}

// translated comment
// translated comment
func (b *ApplicationBootstrapper) ValidateStartup(requiredExtensions []ExtensionType) (healthy bool, report string, issues []string) {
	if b.registryDiagnostics == nil {
		b.registryDiagnostics = NewRegistryDiagnostics(nil)
	}

	// translated comment
	healthy, diagIssues := b.registryDiagnostics.HealthCheck()

	// translated comment
	if len(requiredExtensions) > 0 {
		if valid, missing := b.registryDiagnostics.ValidateRequiredExtensions(requiredExtensions); !valid {
			healthy = false
			for _, ext := range missing {
				diagIssues = append(diagIssues, fmt.Sprintf("required extension not fully loaded: type=%d", ext))
			}
		}
	}

	// translated comment
	report = b.registryDiagnostics.GetDiagnosticReport()

	return healthy, report, diagIssues
}

// translated comment
func (b *ApplicationBootstrapper) GetDiagnosticReport() string {
	if b.registryDiagnostics == nil {
		b.registryDiagnostics = NewRegistryDiagnostics(nil)
	}
	return b.registryDiagnostics.GetDiagnosticReport()
}

// translated comment
// translated comment
func (b *ApplicationBootstrapper) BootstrapWithValidation(requiredExtensions []ExtensionType) error {
	// translated comment
	if _, err := b.BootstrapConfig(); err != nil {
		return err
	}
	if _, err := b.BootstrapContainer(); err != nil {
		return err
	}

	// translated comment
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
