package extension

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// RulesConfig is the top-level rules configuration structure
// After the unified configuration entry, this structure is maintained by the extension package.
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

// EntropyConfig holds entropy configuration
type EntropyConfig struct {
	Enabled       bool    `json:"enabled"`
	HighThreshold float64 `json:"high_threshold"`
	LowThreshold  int     `json:"low_threshold"`
	Description   string  `json:"description"`
	Sensitivity   string  `json:"sensitivity"`
}

// ToolMarkersConfig holds tool marker configuration
type ToolMarkersConfig struct {
	Enabled     bool         `json:"enabled"`
	Patterns    []MarkerInfo `json:"patterns"`
	Description string       `json:"description"`
}

// MarkerInfo holds information for a single marker
type MarkerInfo struct {
	Marker   string `json:"marker"`
	Type     string `json:"type"`
	Severity string `json:"severity"`
}

// HeadlessBrowserConfig holds headless browser configuration
type HeadlessBrowserConfig struct {
	Enabled     bool     `json:"enabled"`
	Markers     []string `json:"markers"`
	Description string   `json:"description"`
}

// OSPlatformConfig holds OS/Platform contradiction configuration
type OSPlatformConfig struct {
	Enabled     bool     `json:"enabled"`
	Rules       []OSRule `json:"rules"`
	Description string   `json:"description"`
}

// OSRule represents an OS rule
type OSRule struct {
	OS                  string      `json:"os"`
	PlatformMustContain interface{} `json:"platform_must_contain"`
	Severity            string      `json:"severity"`
}

// UAOSConfig holds UA/OS contradiction configuration
type UAOSConfig struct {
	Enabled     bool       `json:"enabled"`
	Rules       []UARuleOS `json:"rules"`
	Description string     `json:"description"`
}

// UARuleOS represents a UA/OS rule
type UARuleOS struct {
	UAContains      string `json:"ua_contains"`
	OSCannotContain string `json:"os_cannot_contain"`
	Severity        string `json:"severity"`
}

// MobileScreenConfig holds mobile screen configuration
type MobileScreenConfig struct {
	Enabled               bool               `json:"enabled"`
	MobileScreenWidthMax  int                `json:"mobile_screen_width_max"`
	DesktopScreenWidthMin int                `json:"desktop_screen_width_min"`
	Description           string             `json:"description"`
	Rules                 []MobileScreenRule `json:"rules"`
}

// MobileScreenRule represents a mobile screen rule
type MobileScreenRule struct {
	Device               string `json:"device"`
	ScreenWidthThreshold int    `json:"screen_width_threshold"`
	Rule                 string `json:"rule"`
}

// UAFeatureConfig holds UA/Feature contradiction configuration
type UAFeatureConfig struct {
	Enabled     bool            `json:"enabled"`
	Rules       []UAFeatureRule `json:"rules"`
	Description string          `json:"description"`
}

// UAFeatureRule represents a UA/Feature rule
type UAFeatureRule struct {
	UAContains           string `json:"ua_contains"`
	FeatureCannotSupport string `json:"feature_cannot_support"`
	Severity             string `json:"severity"`
}

// RulesScoringConfig holds scoring configuration
// Uses RulesScoringConfig to avoid conflicts with identically named configuration structures in other modules.
type RulesScoringConfig struct {
	AnomalyWeights map[string]float64   `json:"anomaly_weights"`
	RiskLevels     map[string]RiskLevel `json:"risk_levels"`
}

// RiskLevel represents a risk level
type RiskLevel struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// LoadRulesConfig loads rules configuration from a JSON file
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

// LoadRulesConfigByFilename finds and loads rules configuration by filename within the project
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

// DefaultRulesConfig returns the default rules configuration
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

// String returns the JSON string representation of the configuration
func (c *RulesConfig) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}
