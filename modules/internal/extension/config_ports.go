package extension

// ConfigProvider 运行配置提供端口
type ConfigProvider interface {
	NewConfigFromEnv() *Config
}

// RulesProvider 规则配置提供端口
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
