//go:build bootstrap
// +build bootstrap

package extension

import (
	"fmt"
)

// ApplicationBootstrapper 应用级别启动引导程序
// 负责协调依赖初始化，确保单一组合根原则
// 在应用启动时调用此包提供的工厂函数，而非直接调用各种 New* 函数
type ApplicationBootstrapper struct {
	configProvider ConfigProvider
	rulesProvider  RulesProvider
	registryPort   RegistryPort

	// 缓存的单例
	appConfig *Config
	container *Container
	engine    *ProcessingEngine

	// 诊断工具
	registryDiagnostics *RegistryDiagnostics
}

// NewApplicationBootstrapper 创建应用引导程序（使用默认 provider）
func NewApplicationBootstrapper() *ApplicationBootstrapper {
	return NewApplicationBootstrapperWithProviders(nil, nil, nil)
}

// NewApplicationBootstrapperWithProviders 创建应用引导程序（支持 provider 注入）
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

// BootstrapConfig 初始化应用配置
// 调用一次，结果被缓存；若已初始化则返回缓存值
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

// BootstrapContainer 初始化DI容器
// 调用一次，结果被缓存；若已初始化则返回缓存值
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

// BootstrapEngine 初始化处理引擎
// 调用一次，结果被缓存；若已初始化则返回缓存值
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

// GetConfig 获取缓存的应用配置（若未初始化则初始化）
func (b *ApplicationBootstrapper) GetConfig() (*Config, error) {
	if b.appConfig == nil {
		return b.BootstrapConfig()
	}
	return b.appConfig, nil
}

// GetContainer 获取缓存的DI容器（若未初始化则初始化）
func (b *ApplicationBootstrapper) GetContainer() (*Container, error) {
	if b.container == nil {
		return b.BootstrapContainer()
	}
	return b.container, nil
}

// GetEngine 获取缓存的处理引擎（若未初始化则用默认配置初始化）
func (b *ApplicationBootstrapper) GetEngine() (*ProcessingEngine, error) {
	if b.engine == nil {
		return b.BootstrapEngine(nil)
	}
	return b.engine, nil
}

// Reset 重置所有缓存（仅用于测试或特殊场景）
func (b *ApplicationBootstrapper) Reset() {
	b.appConfig = nil
	b.container = nil
	b.engine = nil
	if b.container != nil {
		b.container.Reset()
	}
}

// SetConfigProvider 设置配置提供器（必须在任何 bootstrap 调用之前）
func (b *ApplicationBootstrapper) SetConfigProvider(provider ConfigProvider) {
	if b.appConfig == nil && b.container == nil && b.engine == nil {
		if provider != nil {
			b.configProvider = provider
		}
	}
}

// SetRulesProvider 设置规则提供器（必须在任何 bootstrap 调用之前）
func (b *ApplicationBootstrapper) SetRulesProvider(provider RulesProvider) {
	if b.appConfig == nil && b.container == nil && b.engine == nil {
		if provider != nil {
			b.rulesProvider = provider
		}
	}
}

// SetRegistryPort 设置注册表端口（必须在任何 bootstrap 调用之前）
func (b *ApplicationBootstrapper) SetRegistryPort(port RegistryPort) {
	if b.engine == nil {
		if port != nil {
			b.registryPort = port
		}
	}
}

// GetRegistryDiagnostics 获取注册表诊断工具
func (b *ApplicationBootstrapper) GetRegistryDiagnostics() *RegistryDiagnostics {
	return b.registryDiagnostics
}

// ValidateStartup 执行启动验证，检查必需的扩展是否已加载
// 返回诊断结果和任何检测到的问题
func (b *ApplicationBootstrapper) ValidateStartup(requiredExtensions []ExtensionType) (healthy bool, report string, issues []string) {
	if b.registryDiagnostics == nil {
		b.registryDiagnostics = NewRegistryDiagnostics(nil)
	}

	// 执行健康检查
	healthy, diagIssues := b.registryDiagnostics.HealthCheck()

	// 验证必需的扩展
	if len(requiredExtensions) > 0 {
		if valid, missing := b.registryDiagnostics.ValidateRequiredExtensions(requiredExtensions); !valid {
			healthy = false
			for _, ext := range missing {
				diagIssues = append(diagIssues, fmt.Sprintf("required extension not fully loaded: type=%d", ext))
			}
		}
	}

	// 生成诊断报告
	report = b.registryDiagnostics.GetDiagnosticReport()

	return healthy, report, diagIssues
}

// GetDiagnosticReport 获取完整的诊断报告
func (b *ApplicationBootstrapper) GetDiagnosticReport() string {
	if b.registryDiagnostics == nil {
		b.registryDiagnostics = NewRegistryDiagnostics(nil)
	}
	return b.registryDiagnostics.GetDiagnosticReport()
}

// BootstrapWithValidation 启动应用并进行验证
// 在启动前检查必需的扩展是否已加载
func (b *ApplicationBootstrapper) BootstrapWithValidation(requiredExtensions []ExtensionType) error {
	// 先进行启动
	if _, err := b.BootstrapConfig(); err != nil {
		return err
	}
	if _, err := b.BootstrapContainer(); err != nil {
		return err
	}

	// 再进行验证
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
