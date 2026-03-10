package extension

import (
	"fmt"
	"os"
)

// translated comment
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvTesting     Environment = "testing"
	EnvProduction  Environment = "production"
)

// translated comment
//
// translated comment
//
//	config := NewConfig(EnvProduction)
//	config.Defense.MaxInputSize = 4096
//	config.Logger.Level = 2
//
// translated comment
//
// translated comment
// translated comment
// translated comment
type Config struct {
	// translated comment
	Environment Environment

	// translated comment
	Defense *DefensePolicy

	// translated comment
	Logger *LoggerConfig

	// translated comment
	Validator *ValidatorConfig

	// translated comment
	Audit *AuditConfig

	// translated comment
	Rules *RulesConfig

	// translated comment
	RulesSource string
}

// translated comment
type LoggerConfig struct {
	// translated comment
	Level int

	// translated comment
	// "stdout", "stderr", "file"
	Output string

	// translated comment
	FilePath string

	// translated comment
	Structured bool

	// translated comment
	// "text", "json"
	Format string
}

// translated comment
type ValidatorConfig struct {
	// translated comment
	MaxDataSize int

	// translated comment
	StrictMode bool

	// translated comment
	ValidateMetadata bool

	// translated comment
	ValidateConfig bool
}

// translated comment
type AuditConfig struct {
	// translated comment
	Enable bool

	// translated comment
	MaxEvents int

	// translated comment
	AlertThreshold int

	// translated comment
	BlockThreshold int

	// translated comment
	LogOutput string

	// translated comment
	LogFilePath string
}

// translated comment
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

	// translated comment
	config.loadFromEnv()

	return config
}

// translated comment
func NewConfigFromEnv() *Config {
	env := Environment(getEnv("FINGERPRINT_ENV", string(EnvDevelopment)))
	return NewConfig(env)
}

// translated comment
// translated comment
// translated comment
// translated comment
//
// translated comment
// translated comment
// translated comment
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

// translated comment
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

// translated comment
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

// translated comment
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

// translated comment
func (c *Config) loadFromEnv() {
	// translated comment
	if logLevel := getEnv("FINGERPRINT_LOG_LEVEL", ""); logLevel != "" {
		var level int
		if _, err := fmt.Sscanf(logLevel, "%d", &level); err == nil {
			c.Logger.Level = level
		}
	}

	// translated comment
	if maxSize := getEnv("FINGERPRINT_MAX_INPUT_SIZE", ""); maxSize != "" {
		var size int
		if _, err := fmt.Sscanf(maxSize, "%d", &size); err == nil {
			c.Defense.MaxInputSize = size
			c.Validator.MaxDataSize = size
		}
	}

	// translated comment
	if maxMem := getEnv("FINGERPRINT_MAX_MEMORY_MB", ""); maxMem != "" {
		var mem int
		if _, err := fmt.Sscanf(maxMem, "%d", &mem); err == nil {
			c.Defense.MaxMemoryMB = mem
		}
	}

	// translated comment
	if timeout := getEnv("FINGERPRINT_TIMEOUT_SEC", ""); timeout != "" {
		var sec int
		if _, err := fmt.Sscanf(timeout, "%d", &sec); err == nil {
			c.Defense.TimeoutSec = sec
		}
	}

	// translated comment
	if rateLimit := getEnv("FINGERPRINT_RATE_LIMIT", ""); rateLimit != "" {
		var limit int
		if _, err := fmt.Sscanf(rateLimit, "%d", &limit); err == nil {
			c.Defense.RateLimit = limit
		}
	}
}

// translated comment
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

// translated comment
func newValidatorConfig(env Environment) *ValidatorConfig {
	config := &ValidatorConfig{
		ValidateMetadata: true,
		ValidateConfig:   true,
	}

	switch env {
	case EnvProduction:
		config.MaxDataSize = 4096 // translated comment
		config.StrictMode = true
	case EnvTesting:
		config.MaxDataSize = 65536 // 64KB
		config.StrictMode = true
	case EnvDevelopment:
		config.MaxDataSize = 1024 * 1024 // translated comment
		config.StrictMode = false
	}

	return config
}

// translated comment
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
		config.Enable = false // translated comment
	}

	return config
}

// translated comment
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// translated comment
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
