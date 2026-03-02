package extension

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// RulesConfig 规则配置顶级结构
// 统一配置入口后，该结构由 extension 包维护。
type RulesConfig struct {
	Metadata                  map[string]interface{} `json:"_metadata"`
	Entropy                   *EntropyConfig         `json:"entropy"`
	ToolMarkers               *ToolMarkersConfig     `json:"tool_markers"`
	HeadlessBrowserUA         *HeadlessBrowserConfig `json:"headless_browser_ua"`
	OSPlatformContradiction   *OSPlatformConfig      `json:"os_platform_contradiction"`
	UAOSContradiction         *UAOSConfig            `json:"ua_os_contradiction"`
	MobileScreenContradiction *MobileScreenConfig    `json:"mobile_screen_contradiction"`
	UAFeatureContradiction    *UAFeatureConfig       `json:"ua_feature_contradiction"`
	Scoring                   *RulesScoringConfig    `json:"scoring"`
}

// EntropyConfig 熵配置
type EntropyConfig struct {
	Enabled       bool    `json:"enabled"`
	HighThreshold float64 `json:"high_threshold"`
	LowThreshold  int     `json:"low_threshold"`
	Description   string  `json:"description"`
	Sensitivity   string  `json:"sensitivity"`
}

// ToolMarkersConfig 工具特征配置
type ToolMarkersConfig struct {
	Enabled     bool         `json:"enabled"`
	Patterns    []MarkerInfo `json:"patterns"`
	Description string       `json:"description"`
}

// MarkerInfo 单个标记信息
type MarkerInfo struct {
	Marker   string `json:"marker"`
	Type     string `json:"type"`
	Severity string `json:"severity"`
}

// HeadlessBrowserConfig 无头浏览器配置
type HeadlessBrowserConfig struct {
	Enabled     bool     `json:"enabled"`
	Markers     []string `json:"markers"`
	Description string   `json:"description"`
}

// OSPlatformConfig OS/Platform 矛盾配置
type OSPlatformConfig struct {
	Enabled     bool     `json:"enabled"`
	Rules       []OSRule `json:"rules"`
	Description string   `json:"description"`
}

// OSRule OS 规则
type OSRule struct {
	OS                  string      `json:"os"`
	PlatformMustContain interface{} `json:"platform_must_contain"`
	Severity            string      `json:"severity"`
}

// UAOSConfig UA/OS 矛盾配置
type UAOSConfig struct {
	Enabled     bool       `json:"enabled"`
	Rules       []UARuleOS `json:"rules"`
	Description string     `json:"description"`
}

// UARuleOS UA/OS 规则
type UARuleOS struct {
	UAContains      string `json:"ua_contains"`
	OSCannotContain string `json:"os_cannot_contain"`
	Severity        string `json:"severity"`
}

// MobileScreenConfig 移动屏幕配置
type MobileScreenConfig struct {
	Enabled               bool               `json:"enabled"`
	MobileScreenWidthMax  int                `json:"mobile_screen_width_max"`
	DesktopScreenWidthMin int                `json:"desktop_screen_width_min"`
	Description           string             `json:"description"`
	Rules                 []MobileScreenRule `json:"rules"`
}

// MobileScreenRule 移动屏幕规则
type MobileScreenRule struct {
	Device               string `json:"device"`
	ScreenWidthThreshold int    `json:"screen_width_threshold"`
	Rule                 string `json:"rule"`
}

// UAFeatureConfig UA/Feature 矛盾配置
type UAFeatureConfig struct {
	Enabled     bool            `json:"enabled"`
	Rules       []UAFeatureRule `json:"rules"`
	Description string          `json:"description"`
}

// UAFeatureRule UA/Feature 规则
type UAFeatureRule struct {
	UAContains           string `json:"ua_contains"`
	FeatureCannotSupport string `json:"feature_cannot_support"`
	Severity             string `json:"severity"`
}

// RulesScoringConfig 评分配置
// 为避免与其他模块同名配置结构冲突，规则配置侧使用 RulesScoringConfig。
type RulesScoringConfig struct {
	AnomalyWeights map[string]float64   `json:"anomaly_weights"`
	RiskLevels     map[string]RiskLevel `json:"risk_levels"`
}

// RiskLevel 风险等级
type RiskLevel struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// LoadRulesConfig 从 JSON 文件加载规则配置
func LoadRulesConfig(path string) (*RulesConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config RulesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadRulesConfigByFilename 根据文件名在项目内查找规则配置
func LoadRulesConfigByFilename(filename string) (*RulesConfig, error) {
	candidates := []string{
		filepath.Join("internal", "config", filename),
		filepath.Join("..", "internal", "config", filename),
		filepath.Join("...", "internal", "config", filename),
		filepath.Join(os.Getenv("HOME"), ".fingerprint", "config", filename),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return LoadRulesConfig(candidate)
		}
	}

	return DefaultRulesConfig(), nil
}

// DefaultRulesConfig 返回默认规则配置
func DefaultRulesConfig() *RulesConfig {
	return &RulesConfig{
		Metadata: map[string]interface{}{
			"version":     "1.0",
			"description": "Default fingerprint detection rules configuration",
		},
		Entropy: &EntropyConfig{
			Enabled:       true,
			HighThreshold: 7.5,
			LowThreshold:  26,
			Sensitivity:   "high",
		},
		ToolMarkers: &ToolMarkersConfig{
			Enabled: true,
			Patterns: []MarkerInfo{
				{Marker: "HeadlessChrome", Type: "browser_automation", Severity: "critical"},
				{Marker: "PhantomJS", Type: "headless_browser", Severity: "critical"},
				{Marker: "webdriver", Type: "selenium_indicator", Severity: "critical"},
				{Marker: "selenium", Type: "selenium_driver", Severity: "critical"},
				{Marker: "puppeteer", Type: "puppeteer_client", Severity: "critical"},
			},
		},
		HeadlessBrowserUA: &HeadlessBrowserConfig{
			Enabled: true,
			Markers: []string{"headlesschrome", "phantomjs", "selenium", "webdriver", "puppeteer", "playwright", "cypress", "jsdom", "zombie", "htmlunit"},
		},
		OSPlatformContradiction: &OSPlatformConfig{
			Enabled: true,
			Rules: []OSRule{
				{OS: "Windows", PlatformMustContain: "Win", Severity: "high"},
				{OS: "Mac", PlatformMustContain: "Mac", Severity: "high"},
				{OS: "Linux", PlatformMustContain: []string{"Linux", "X11"}, Severity: "high"},
			},
		},
		MobileScreenContradiction: &MobileScreenConfig{
			Enabled:               true,
			MobileScreenWidthMax:  1920,
			DesktopScreenWidthMin: 800,
		},
		Scoring: &RulesScoringConfig{
			AnomalyWeights: map[string]float64{
				"entropy":                     0.15,
				"tool_marker":                 0.25,
				"headless_browser":            0.20,
				"os_platform_contradiction":   0.15,
				"ua_os_contradiction":         0.20,
				"mobile_screen_contradiction": 0.10,
				"ua_feature_contradiction":    0.10,
			},
		},
	}
}

// String 返回配置的 JSON 字符串表示
func (c *RulesConfig) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}
