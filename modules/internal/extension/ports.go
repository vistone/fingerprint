package extension

// RegistryPort defines the extension registry access port
// Used to decouple the engine from the global registry, supporting implementation replacement and test injection.
type RegistryPort interface {
	GetParser(extType ExtensionType) (Parser, error)
	GetAnalyzer(extType ExtensionType) (Analyzer, error)
	GetHandlers(extType ExtensionType) []Handler
}

// globalRegistryPort is the default adapter: bridges to global registry functions, preserving existing behavior.
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
