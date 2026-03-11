package extension

// ConfigProvider is the runtime configuration provider port
type ConfigProvider interface {
	NewConfigFromEnv() *Config
}

// RulesProvider is the rules configuration provider port
type RulesProvider interface {
	LoadRules(path string) (*RulesConfig, error)
	LoadRulesByFilename(filename string) (*RulesConfig, error)
	DefaultRules() *RulesConfig
}

type defaultConfigProvider struct{}

func (defaultConfigProvider) NewConfigFromEnv() *Config {
	return NewConfigFromEnv()
}

type defaultRulesProvider struct{}

func (defaultRulesProvider) LoadRules(path string) (*RulesConfig, error) {
	return LoadRulesConfig(path)
}

func (defaultRulesProvider) LoadRulesByFilename(filename string) (*RulesConfig, error) {
	return LoadRulesConfigByFilename(filename)
}

func (defaultRulesProvider) DefaultRules() *RulesConfig {
	return DefaultRulesConfig()
}
