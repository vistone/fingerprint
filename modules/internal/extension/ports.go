package extension

// translated comment
// translated comment
type RegistryPort interface {
	GetParser(extType ExtensionType) (Parser, error)
	GetAnalyzer(extType ExtensionType) (Analyzer, error)
	GetHandlers(extType ExtensionType) []Handler
}

// translated comment
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
