package extension

// RegistryPort 定义扩展注册表访问端口（Port）
// 用于将引擎与全局注册表解耦，支持后续替换实现与测试注入。
type RegistryPort interface {
	GetParser(extType ExtensionType) (Parser, error)
	GetAnalyzer(extType ExtensionType) (Analyzer, error)
	GetHandlers(extType ExtensionType) []Handler
}

// globalRegistryPort 默认适配器：桥接到全局注册表函数，保持现有行为不变。
type globalRegistryPort struct{}

func (globalRegistryPort) GetParser(extType ExtensionType) (Parser, error) {
	return GetParser(extType)
}

func (globalRegistryPort) GetAnalyzer(extType ExtensionType) (Analyzer, error) {
	return GetAnalyzer(extType)
}

func (globalRegistryPort) GetHandlers(extType ExtensionType) []Handler {
	return GetHandlers(extType)
}
