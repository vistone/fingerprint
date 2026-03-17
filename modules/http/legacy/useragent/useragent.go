package useragent

// Phase 3: This module has completed basic migration, awaiting deep optimization (see docs/5-process/modularization/PHASE_3_PLAN.md)
import (
	"fmt"
	"strings"

	"github.com/vistone/fingerprint/modules/core/types"
	utils "github.com/vistone/fingerprint/modules/kit"
)

// UserAgentGenerator User-Agent generator
type UserAgentGenerator struct {
	templates map[string]types.UserAgentTemplate
}

var (
	defaultGenerator *UserAgentGenerator
)

func init() {
	defaultGenerator = NewUserAgentGenerator()
}

// NewUserAgentGenerator creates a new User-Agent generator
func NewUserAgentGenerator() *UserAgentGenerator {
	gen := &UserAgentGenerator{
		templates: make(map[string]types.UserAgentTemplate),
	}
	gen.initTemplates()
	return gen
}

// initTemplates initializes User-Agent templates
func (g *UserAgentGenerator) initTemplates() {
	g.addDesktopBrowserTemplates("chrome_", chromeTemplates, types.BrowserChrome)
	g.addDesktopBrowserTemplates("firefox_", firefoxTemplates, types.BrowserFirefox)
	g.addSafariTemplates()
	g.addDesktopBrowserTemplates("opera_", operaTemplates, types.BrowserOpera)
	g.addDesktopBrowserTemplates("edge_", edgeTemplates, types.BrowserEdge)
	g.addMobileAppTemplates(iosAppTemplates, types.BrowserSafari, "ios")
	g.addMobileAppTemplates(androidAppTemplates, types.BrowserChrome, "android")
	g.addMobileAppTemplates(okhttpTemplates, types.BrowserChrome, "okhttp4")
	g.addCloudflareTemplate()
}

func (g *UserAgentGenerator) addDesktopBrowserTemplates(prefix string, templates map[string]string, browser types.BrowserType) {
	for version, template := range templates {
		g.templates[prefix+version] = types.UserAgentTemplate{
			Browser:    browser,
			Version:    version,
			Template:   template,
			OSRequired: true,
		}
	}
}

func (g *UserAgentGenerator) addSafariTemplates() {
	for key, template := range safariTemplates {
		mobile := strings.Contains(key, "ios") || strings.Contains(key, "ipad")
		g.templates["safari_"+key] = types.UserAgentTemplate{
			Browser:    types.BrowserSafari,
			Version:    key,
			Template:   template,
			Mobile:     mobile,
			OSRequired: !mobile,
		}
	}
}

func (g *UserAgentGenerator) addMobileAppTemplates(templates map[string]string, browser types.BrowserType, version string) {
	for key, template := range templates {
		g.templates[key] = types.UserAgentTemplate{
			Browser:    browser,
			Version:    version,
			Template:   template,
			Mobile:     true,
			OSRequired: false,
		}
	}
}

func (g *UserAgentGenerator) addCloudflareTemplate() {
	g.templates["cloudflare_custom"] = types.UserAgentTemplate{
		Browser:    types.BrowserChrome,
		Version:    "custom",
		Template:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		Mobile:     false,
		OSRequired: false,
	}
}

var chromeTemplates = map[string]string{
	"103": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36",
	"104": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
	"105": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36",
	"106": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
	"107": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
	"108": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
	"109": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
	"110": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36",
	"111": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36",
	"112": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36",
	"116": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
	"117": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
	"120": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"124": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"130": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
	"131": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"133": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
}

var firefoxTemplates = map[string]string{
	"102": "Mozilla/5.0 (%s; rv:102.0) Gecko/20100101 Firefox/102.0",
	"104": "Mozilla/5.0 (%s; rv:104.0) Gecko/20100101 Firefox/104.0",
	"105": "Mozilla/5.0 (%s; rv:105.0) Gecko/20100101 Firefox/105.0",
	"106": "Mozilla/5.0 (%s; rv:106.0) Gecko/20100101 Firefox/106.0",
	"108": "Mozilla/5.0 (%s; rv:108.0) Gecko/20100101 Firefox/108.0",
	"110": "Mozilla/5.0 (%s; rv:110.0) Gecko/20100101 Firefox/110.0",
	"117": "Mozilla/5.0 (%s; rv:117.0) Gecko/20100101 Firefox/117.0",
	"120": "Mozilla/5.0 (%s; rv:120.0) Gecko/20100101 Firefox/120.0",
	"123": "Mozilla/5.0 (%s; rv:123.0) Gecko/20100101 Firefox/123.0",
	"132": "Mozilla/5.0 (%s; rv:132.0) Gecko/20100101 Firefox/132.0",
	"133": "Mozilla/5.0 (%s; rv:133.0) Gecko/20100101 Firefox/133.0",
	"135": "Mozilla/5.0 (%s; rv:135.0) Gecko/20100101 Firefox/135.0",
}

var safariTemplates = map[string]string{
	"15_6_1":    "Mozilla/5.0 (%s) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.6.1 Safari/605.1.15",
	"16_0":      "Mozilla/5.0 (%s) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Safari/605.1.15",
	"ipad_15_6": "Mozilla/5.0 (iPad; CPU OS 15_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.6 Mobile/15E148 Safari/604.1",
	"ios_15_5":  "Mozilla/5.0 (iPhone; CPU iPhone OS 15_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.5 Mobile/15E148 Safari/604.1",
	"ios_15_6":  "Mozilla/5.0 (iPhone; CPU iPhone OS 15_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.6 Mobile/15E148 Safari/604.1",
	"ios_16_0":  "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
	"ios_17_0":  "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	"ios_18_0":  "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Mobile/15E148 Safari/604.1",
	"ios_18_5":  "Mozilla/5.0 (iPhone; CPU iPhone OS 18_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.5 Mobile/15E148 Safari/604.1",
}

var operaTemplates = map[string]string{
	"89": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36 OPR/89.0.0.0",
	"90": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36 OPR/90.0.0.0",
	"91": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36 OPR/91.0.0.0",
}

var edgeTemplates = map[string]string{
	"99":  "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.51 Safari/537.36 Edg/99.0.1150.36",
	"101": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.64 Safari/537.36 Edg/101.0.1210.53",
	"120": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
	"131": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0",
	"133": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/133.0.0.0",
}

var iosAppTemplates = map[string]string{
	"zalando_ios_mobile": "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	"nike_ios_mobile":    "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	"mms_ios":            "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
	"mms_ios_2":          "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
	"mms_ios_3":          "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	"mesh_ios":           "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
	"mesh_ios_2":         "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	"confirmed_ios":      "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
}

var androidAppTemplates = map[string]string{
	"zalando_android_mobile": "Mozilla/5.0 (Linux; Android 13; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	"nike_android_mobile":    "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	"mesh_android":           "Mozilla/5.0 (Linux; Android 12; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	"mesh_android_2":         "Mozilla/5.0 (Linux; Android 13; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	"confirmed_android":      "Mozilla/5.0 (Linux; Android 12; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	"confirmed_android_2":    "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
}

var okhttpTemplates = map[string]string{
	"okhttp4_android_7":  "Mozilla/5.0 (Linux; Android 7.0; SM-G930F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	"okhttp4_android_8":  "Mozilla/5.0 (Linux; Android 8.0; SM-G950F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	"okhttp4_android_9":  "Mozilla/5.0 (Linux; Android 9; SM-G960F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	"okhttp4_android_10": "Mozilla/5.0 (Linux; Android 10; SM-G970F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	"okhttp4_android_11": "Mozilla/5.0 (Linux; Android 11; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	"okhttp4_android_12": "Mozilla/5.0 (Linux; Android 12; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	"okhttp4_android_13": "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
}

// GetUserAgent gets User-Agent based on fingerprint name
// If fingerprint requires OS information, randomly selects an OS
func (g *UserAgentGenerator) GetUserAgent(profileName string) (string, error) {
	return g.GetUserAgentWithOS(profileName, types.OperatingSystem(""))
}

// GetUserAgentWithOS gets User-Agent based on fingerprint name and specified OS
// If os is empty and OS information is required, randomly selects an OS
func (g *UserAgentGenerator) GetUserAgentWithOS(profileName string, os types.OperatingSystem) (string, error) {
	if profileName == "" {
		return "", fmt.Errorf("profile name cannot be empty")
	}
	template, ok := g.templates[profileName]
	if !ok {
		// Try to extract browser type and version from profileName
		return g.generateFromProfileName(profileName, os)
	}

	// If OS information is not required, return template directly
	if !template.OSRequired {
		return template.Template, nil
	}

	// If OS information is required
	if os == "" {
		// Randomly select an OS
		os = RandomOS()
	}

	return fmt.Sprintf(template.Template, string(os)), nil
}

// generateFromProfileName generates User-Agent from profile name
func (g *UserAgentGenerator) generateFromProfileName(profileName string, os types.OperatingSystem) (string, error) {
	profileName = strings.ToLower(profileName)

	// Parse browser type and version
	var browser types.BrowserType
	var version string

	if strings.HasPrefix(profileName, "chrome_") {
		browser = types.BrowserChrome
		version = strings.TrimPrefix(profileName, "chrome_")
		// Handle special versions
		if strings.Contains(version, "_psk") {
			version = strings.Split(version, "_psk")[0]
		}
		if strings.Contains(version, "_pq") {
			version = strings.Split(version, "_pq")[0]
		}
	} else if strings.HasPrefix(profileName, "firefox_") {
		browser = types.BrowserFirefox
		version = strings.TrimPrefix(profileName, "firefox_")
	} else if strings.HasPrefix(profileName, "safari_") {
		browser = types.BrowserSafari
		version = strings.TrimPrefix(profileName, "safari_")
	} else if strings.HasPrefix(profileName, "opera_") {
		browser = types.BrowserOpera
		version = strings.TrimPrefix(profileName, "opera_")
	} else if strings.HasPrefix(profileName, "edge_") {
		browser = types.BrowserEdge
		version = strings.TrimPrefix(profileName, "edge_")
	} else {
		// Default to Chrome 133
		return g.GetUserAgentWithOS("chrome_133", os)
	}

	// Generate User-Agent
	if os == "" {
		os = RandomOS()
	}

	switch browser {
	case types.BrowserChrome:
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36", string(os), version), nil
	case types.BrowserFirefox:
		return fmt.Sprintf("Mozilla/5.0 (%s; rv:%s.0) Gecko/20100101 Firefox/%s.0", string(os), version, version), nil
	case types.BrowserSafari:
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/%s Safari/605.1.15", string(os), version), nil
	case types.BrowserOpera:
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36 OPR/%s.0.0.0", string(os), version, version), nil
	case types.BrowserEdge:
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36 Edg/%s.0.0.0", string(os), version, version), nil
	default:
		return "", fmt.Errorf("unsupported browser type: %s", browser)
	}
}

// RandomOS randomly selects an operating system
func RandomOS() types.OperatingSystem {
	if len(types.OperatingSystems) == 0 {
		return types.OSWindows10 // Default to Windows 10
	}
	return utils.RandomChoice(types.OperatingSystems)
}

// GetUserAgentForProfile gets User-Agent for specified profiles.ClientProfile
func GetUserAgentForProfile(profileName string) (string, error) {
	return defaultGenerator.GetUserAgent(profileName)
}

// GetUserAgentForProfileWithOS gets User-Agent for specified profiles.ClientProfile and OS
func GetUserAgentForProfileWithOS(profileName string, os types.OperatingSystem) (string, error) {
	return defaultGenerator.GetUserAgentWithOS(profileName, os)
}
