package extension

import (
	"fmt"
	"os"
)

// Environment 运行环境枚举
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvTesting     Environment = "testing"
	EnvProduction  Environment = "production"
)

// Config 统一配置对象，管理所有模块配置
//
// 使用示例：
//
//	config := NewConfig(EnvProduction)
//	config.Defense.MaxInputSize = 4096
//	config.Logger.Level = 2
//
// 环境变量支持：
//
//	FINGERPRINT_ENV - 运行环境 (development/testing/production)
//	FINGERPRINT_LOG_LEVEL - 日志级别 (0-4)
//	FINGERPRINT_MAX_INPUT_SIZE - 最大输入大小 (字节)
type Config struct {
	// 运行环境
	Environment Environment

	// 防御配置
	Defense *DefensePolicy

	// 日志配置
	Logger *LoggerConfig

	// 验证配置
	Validator *ValidatorConfig

	// 审计配置
	Audit *AuditConfig

	// 规则配置
	Rules *RulesConfig

	// 规则配置来源（文件路径或文件名）
	RulesSource string
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	// 日志级别 (0=debug, 1=info, 2=warn, 3=error, 4=fatal)
	Level int

	// 日志输出目标
	// "stdout", "stderr", "file"
	Output string

	// 日志文件路径（仅当 Output="file" 时有效）
	FilePath string

	// 是否启用结构化日志
	Structured bool

	// 日志格式
	// "text", "json"
	Format string
}

// ValidatorConfig 验证器配置
type ValidatorConfig struct {
	// 最大数据大小（字节）
	MaxDataSize int

	// 是否启用严格模式
	StrictMode bool

	// 是否验证元数据
	ValidateMetadata bool

	// 是否验证配置
	ValidateConfig bool
}

// AuditConfig 审计配置
type AuditConfig struct {
	// 是否启用审计
	Enable bool

	// 最大审计事件数（内存中保留）
	MaxEvents int

	// 警告阈值
	AlertThreshold int

	// 阻止阈值
	BlockThreshold int

	// 审计日志输出目标
	LogOutput string

	// 审计日志文件路径
	LogFilePath string
}

// NewConfig 创建配置对象，根据环境应用预设
func NewConfig(env Environment) *Config {
	config := &Config{
		Environment: env,
		Defense:     DefaultDefensePolicy(),
		Logger:      newLoggerConfig(env),
		Validator:   newValidatorConfig(env),
		Audit:       newAuditConfig(env),
		Rules:       DefaultRulesConfig(),
		RulesSource: "rules.json",
	}

	// 从环境变量覆盖配置
	config.loadFromEnv()

	return config
}

// NewConfigFromEnv 从环境变量创建配置
func NewConfigFromEnv() *Config {
	env := Environment(getEnv("FINGERPRINT_ENV", string(EnvDevelopment)))
	return NewConfig(env)
}

// NewUnifiedConfigFromEnv 从环境变量创建统一配置
// 该入口会同时加载：
//   - extension 运行配置（Defense/Logger/Validator/Audit）
//   - 规则配置（internal/config/rules.json）
//
// 环境变量：
//   - FINGERPRINT_RULES_PATH: 规则配置绝对/相对路径
//   - FINGERPRINT_RULES_FILE: 规则配置文件名（用于按候选路径搜索）
func NewUnifiedConfigFromEnv() *Config {
	config := NewConfigFromEnv()

	rulesPath := getEnv("FINGERPRINT_RULES_PATH", "")
	rulesFile := getEnv("FINGERPRINT_RULES_FILE", "rules.json")

	if rulesPath != "" {
		rules, err := LoadRulesConfig(rulesPath)
		if err == nil && rules != nil {
			config.Rules = rules
			config.RulesSource = rulesPath
			config.applyRulesOverlay()
			return config
		}
	}

	rules, err := LoadRulesConfigByFilename(rulesFile)
	if err == nil && rules != nil {
		config.Rules = rules
		config.RulesSource = rulesFile
		config.applyRulesOverlay()
	}

	return config
}

// LoadRulesFromPath 从指定路径加载规则配置并应用覆盖
func (c *Config) LoadRulesFromPath(path string) error {
	rules, err := LoadRulesConfig(path)
	if err != nil {
		return NewErrorWithCause(ErrCodeInvalidConfig, "failed to load rules config", err).
			WithContext("path", path)
	}

	c.Rules = rules
	c.RulesSource = path
	c.applyRulesOverlay()
	return nil
}

// LoadRulesByFilename 按文件名加载规则配置并应用覆盖
func (c *Config) LoadRulesByFilename(filename string) error {
	rules, err := LoadRulesConfigByFilename(filename)
	if err != nil {
		return NewErrorWithCause(ErrCodeInvalidConfig, "failed to load rules config by filename", err).
			WithContext("filename", filename)
	}

	c.Rules = rules
	c.RulesSource = filename
	c.applyRulesOverlay()
	return nil
}

// applyRulesOverlay 将规则配置中的关键阈值映射到运行配置
func (c *Config) applyRulesOverlay() {
	if c.Rules == nil {
		return
	}

	if c.Rules.Entropy != nil && c.Rules.Entropy.Enabled {
		if c.Rules.Entropy.LowThreshold > 0 {
			low := c.Rules.Entropy.LowThreshold
			if low < 1024 {
				low = 1024
			}
			if low > c.Validator.MaxDataSize {
				c.Validator.MaxDataSize = low
			}
		}
	}

	if c.Rules.Scoring != nil && len(c.Rules.Scoring.AnomalyWeights) > 0 {
		if !c.Audit.Enable {
			c.Audit.Enable = true
		}
	}
}

// loadFromEnv 从环境变量加载配置
func (c *Config) loadFromEnv() {
	// 日志级别
	if logLevel := getEnv("FINGERPRINT_LOG_LEVEL", ""); logLevel != "" {
		var level int
		if _, err := fmt.Sscanf(logLevel, "%d", &level); err == nil {
			c.Logger.Level = level
		}
	}

	// 最大输入大小
	if maxSize := getEnv("FINGERPRINT_MAX_INPUT_SIZE", ""); maxSize != "" {
		var size int
		if _, err := fmt.Sscanf(maxSize, "%d", &size); err == nil {
			c.Defense.MaxInputSize = size
			c.Validator.MaxDataSize = size
		}
	}

	// 最大内存
	if maxMem := getEnv("FINGERPRINT_MAX_MEMORY_MB", ""); maxMem != "" {
		var mem int
		if _, err := fmt.Sscanf(maxMem, "%d", &mem); err == nil {
			c.Defense.MaxMemoryMB = mem
		}
	}

	// 超时设置
	if timeout := getEnv("FINGERPRINT_TIMEOUT_SEC", ""); timeout != "" {
		var sec int
		if _, err := fmt.Sscanf(timeout, "%d", &sec); err == nil {
			c.Defense.TimeoutSec = sec
		}
	}

	// 速率限制
	if rateLimit := getEnv("FINGERPRINT_RATE_LIMIT", ""); rateLimit != "" {
		var limit int
		if _, err := fmt.Sscanf(rateLimit, "%d", &limit); err == nil {
			c.Defense.RateLimit = limit
		}
	}
}

// newLoggerConfig 创建日志配置预设
func newLoggerConfig(env Environment) *LoggerConfig {
	config := &LoggerConfig{
		Output:     "stdout",
		Structured: env == EnvProduction,
		Format:     "text",
	}

	switch env {
	case EnvProduction:
		config.Level = 1 // INFO
		config.Format = "json"
	case EnvTesting:
		config.Level = 2 // WARN
	case EnvDevelopment:
		config.Level = 0 // DEBUG
	}

	return config
}

// newValidatorConfig 创建验证器配置预设
func newValidatorConfig(env Environment) *ValidatorConfig {
	config := &ValidatorConfig{
		ValidateMetadata: true,
		ValidateConfig:   true,
	}

	switch env {
	case EnvProduction:
		config.MaxDataSize = 4096 // 4KB，严格限制
		config.StrictMode = true
	case EnvTesting:
		config.MaxDataSize = 65536 // 64KB
		config.StrictMode = true
	case EnvDevelopment:
		config.MaxDataSize = 1024 * 1024 // 1MB，宽松限制
		config.StrictMode = false
	}

	return config
}

// newAuditConfig 创建审计配置预设
func newAuditConfig(env Environment) *AuditConfig {
	config := &AuditConfig{
		Enable:         true,
		MaxEvents:      10000,
		AlertThreshold: 10,
		BlockThreshold: 20,
		LogOutput:      "stdout",
	}

	switch env {
	case EnvProduction:
		config.Enable = true
		config.MaxEvents = 100000
		config.LogOutput = "file"
		config.LogFilePath = "/var/log/fingerprint-audit.log"
	case EnvTesting:
		config.Enable = true
		config.MaxEvents = 10000
	case EnvDevelopment:
		config.Enable = false // 开发环境默认禁用审计
	}

	return config
}

// getEnv 获取环境变量，带默认值
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if c.Rules == nil {
		return NewError(ErrCodeInvalidConfig, "Rules config cannot be nil")
	}

	if c.Logger.Level < 0 || c.Logger.Level > 4 {
		return NewError(ErrCodeInvalidConfig,
			"Logger.Level must be 0-4").
			WithContext("value", c.Logger.Level)
	}

	if c.Defense.MaxInputSize <= 0 {
		return NewError(ErrCodeInvalidConfig,
			"Defense.MaxInputSize must be positive").
			WithContext("value", c.Defense.MaxInputSize)
	}

	if c.Defense.MaxMemoryMB <= 0 {
		return NewError(ErrCodeInvalidConfig,
			"Defense.MaxMemoryMB must be positive").
			WithContext("value", c.Defense.MaxMemoryMB)
	}

	if c.Defense.TimeoutSec <= 0 {
		return NewError(ErrCodeInvalidConfig,
			"Defense.TimeoutSec must be positive").
			WithContext("value", c.Defense.TimeoutSec)
	}

	if c.Validator.MaxDataSize <= 0 {
		return NewError(ErrCodeInvalidConfig,
			"Validator.MaxDataSize must be positive").
			WithContext("value", c.Validator.MaxDataSize)
	}

	return nil
}
