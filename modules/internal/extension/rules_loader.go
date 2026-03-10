package extension

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// translated comment
// translated comment
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

// translated comment
type EntropyConfig struct {
	Enabled       bool    `json:"enabled"`
	HighThreshold float64 `json:"high_threshold"`
	LowThreshold  int     `json:"low_threshold"`
	Description   string  `json:"description"`
	Sensitivity   string  `json:"sensitivity"`
}

// translated comment
type ToolMarkersConfig struct {
	Enabled     bool         `json:"enabled"`
	Patterns    []MarkerInfo `json:"patterns"`
	Description string       `json:"description"`
}

// translated comment
type MarkerInfo struct {
	Marker   string `json:"marker"`
	Type     string `json:"type"`
	Severity string `json:"severity"`
}

// translated comment
type HeadlessBrowserConfig struct {
	Enabled     bool     `json:"enabled"`
	Markers     []string `json:"markers"`
	Description string   `json:"description"`
}

// translated comment
type OSPlatformConfig struct {
	Enabled     bool     `json:"enabled"`
	Rules       []OSRule `json:"rules"`
	Description string   `json:"description"`
}

// translated comment
type OSRule struct {
	OS                  string      `json:"os"`
	PlatformMustContain interface{} `json:"platform_must_contain"`
	Severity            string      `json:"severity"`
}

// translated comment
type UAOSConfig struct {
	Enabled     bool       `json:"enabled"`
	Rules       []UARuleOS `json:"rules"`
	Description string     `json:"description"`
}

// translated comment
type UARuleOS struct {
	UAContains      string `json:"ua_contains"`
	OSCannotContain string `json:"os_cannot_contain"`
	Severity        string `json:"severity"`
}

// translated comment
type MobileScreenConfig struct {
	Enabled               bool               `json:"enabled"`
	MobileScreenWidthMax  int                `json:"mobile_screen_width_max"`
	DesktopScreenWidthMin int                `json:"desktop_screen_width_min"`
	Description           string             `json:"description"`
	Rules                 []MobileScreenRule `json:"rules"`
}

// translated comment
type MobileScreenRule struct {
	Device               string `json:"device"`
	ScreenWidthThreshold int    `json:"screen_width_threshold"`
	Rule                 string `json:"rule"`
}

// translated comment
type UAFeatureConfig struct {
	Enabled     bool            `json:"enabled"`
	Rules       []UAFeatureRule `json:"rules"`
	Description string          `json:"description"`
}

// translated comment
type UAFeatureRule struct {
	UAContains           string `json:"ua_contains"`
	FeatureCannotSupport string `json:"feature_cannot_support"`
	Severity             string `json:"severity"`
}

// translated comment
// translated comment
type RulesScoringConfig struct {
	AnomalyWeights map[string]float64   `json:"anomaly_weights"`
	RiskLevels     map[string]RiskLevel `json:"risk_levels"`
}

// translated comment
type RiskLevel struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// translated comment
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

// translated comment
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

// translated comment
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

// translated comment
func (c *RulesConfig) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}
