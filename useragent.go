package fingerprint

import (
	"fmt"
	"strings"

	"github.com/vistone/fingerprint/internal/utils"
)

// UserAgentGenerator User-Agent 生成器
type UserAgentGenerator struct {
	templates map[string]UserAgentTemplate
}

var (
	defaultGenerator *UserAgentGenerator
)

func init() {
	defaultGenerator = NewUserAgentGenerator()
}

// NewUserAgentGenerator 创建新的 User-Agent 生成器
func NewUserAgentGenerator() *UserAgentGenerator {
	gen := &UserAgentGenerator{
		templates: make(map[string]UserAgentTemplate),
	}
	gen.initTemplates()
	return gen
}

// initTemplates 初始化 User-Agent 模板
func (g *UserAgentGenerator) initTemplates() {
	// Chrome User-Agent 模板
	chromeTemplates := map[string]string{
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

	for version, template := range chromeTemplates {
		g.templates["chrome_"+version] = UserAgentTemplate{
			Browser:    BrowserChrome,
			Version:    version,
			Template:   template,
			OSRequired: true,
		}
	}

	// Firefox User-Agent 模板
	firefoxTemplates := map[string]string{
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

	for version, template := range firefoxTemplates {
		g.templates["firefox_"+version] = UserAgentTemplate{
			Browser:    BrowserFirefox,
			Version:    version,
			Template:   template,
			OSRequired: true,
		}
	}

	// Safari User-Agent 模板
	safariTemplates := map[string]string{
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

	for key, template := range safariTemplates {
		mobile := strings.Contains(key, "ios") || strings.Contains(key, "ipad")
		g.templates["safari_"+key] = UserAgentTemplate{
			Browser:    BrowserSafari,
			Version:    key,
			Template:   template,
			Mobile:     mobile,
			OSRequired: !mobile, // 移动端不需要操作系统信息
		}
	}

	// Opera User-Agent 模板
	operaTemplates := map[string]string{
		"89": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36 OPR/89.0.0.0",
		"90": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36 OPR/90.0.0.0",
		"91": "Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36 OPR/91.0.0.0",
	}

	for version, template := range operaTemplates {
		g.templates["opera_"+version] = UserAgentTemplate{
			Browser:    BrowserOpera,
			Version:    version,
			Template:   template,
			OSRequired: true,
		}
	}

	// 移动端和自定义指纹的 User-Agent 模板
	// iOS 应用指纹 - 使用 iOS Safari User-Agent
	iosAppTemplates := map[string]string{
		"zalando_ios_mobile": "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		"nike_ios_mobile":    "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		"mms_ios":            "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
		"mms_ios_2":          "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
		"mms_ios_3":          "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		"mesh_ios":           "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
		"mesh_ios_2":         "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		"confirmed_ios":      "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
	}

	for key, template := range iosAppTemplates {
		g.templates[key] = UserAgentTemplate{
			Browser:    BrowserSafari,
			Version:    "ios",
			Template:   template,
			Mobile:     true,
			OSRequired: false, // iOS 移动端不需要操作系统占位符
		}
	}

	// Android 应用指纹 - 使用 Android Chrome User-Agent
	androidAppTemplates := map[string]string{
		"zalando_android_mobile": "Mozilla/5.0 (Linux; Android 13; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"nike_android_mobile":    "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"mesh_android":           "Mozilla/5.0 (Linux; Android 12; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"mesh_android_2":         "Mozilla/5.0 (Linux; Android 13; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"confirmed_android":      "Mozilla/5.0 (Linux; Android 12; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"confirmed_android_2":    "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	}

	for key, template := range androidAppTemplates {
		g.templates[key] = UserAgentTemplate{
			Browser:    BrowserChrome,
			Version:    "android",
			Template:   template,
			Mobile:     true,
			OSRequired: false, // Android 移动端不需要操作系统占位符
		}
	}

	// OkHttp4 Android 指纹 - 使用 Android Chrome User-Agent（不同 Android 版本）
	okhttpTemplates := map[string]string{
		"okhttp4_android_7":  "Mozilla/5.0 (Linux; Android 7.0; SM-G930F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"okhttp4_android_8":  "Mozilla/5.0 (Linux; Android 8.0; SM-G950F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"okhttp4_android_9":  "Mozilla/5.0 (Linux; Android 9; SM-G960F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"okhttp4_android_10": "Mozilla/5.0 (Linux; Android 10; SM-G970F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"okhttp4_android_11": "Mozilla/5.0 (Linux; Android 11; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"okhttp4_android_12": "Mozilla/5.0 (Linux; Android 12; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"okhttp4_android_13": "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	}

	for key, template := range okhttpTemplates {
		g.templates[key] = UserAgentTemplate{
			Browser:    BrowserChrome,
			Version:    "okhttp4",
			Template:   template,
			Mobile:     true,
			OSRequired: false, // Android 移动端不需要操作系统占位符
		}
	}

	// Cloudflare Custom - 使用 Chrome User-Agent（通常用于 cloudscraper）
	g.templates["cloudflare_custom"] = UserAgentTemplate{
		Browser:    BrowserChrome,
		Version:    "custom",
		Template:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		Mobile:     false,
		OSRequired: false, // 固定 User-Agent，不需要操作系统占位符
	}
}

// GetUserAgent 根据指纹名称获取 User-Agent
// 如果指纹需要操作系统信息，会随机选择一个操作系统
func (g *UserAgentGenerator) GetUserAgent(profileName string) (string, error) {
	return g.GetUserAgentWithOS(profileName, OperatingSystem(""))
}

// GetUserAgentWithOS 根据指纹名称和指定操作系统获取 User-Agent
// 如果 os 为空，且需要操作系统信息，会随机选择一个操作系统
func (g *UserAgentGenerator) GetUserAgentWithOS(profileName string, os OperatingSystem) (string, error) {
	if profileName == "" {
		return "", fmt.Errorf("profile name cannot be empty")
	}
	template, ok := g.templates[profileName]
	if !ok {
		// 尝试从 profileName 中提取浏览器类型和版本
		return g.generateFromProfileName(profileName, os)
	}

	// 如果不需要操作系统信息，直接返回模板
	if !template.OSRequired {
		return template.Template, nil
	}

	// 如果需要操作系统信息
	if os == "" {
		// 随机选择操作系统
		os = RandomOS()
	}

	return fmt.Sprintf(template.Template, string(os)), nil
}

// generateFromProfileName 从 profile 名称生成 User-Agent
func (g *UserAgentGenerator) generateFromProfileName(profileName string, os OperatingSystem) (string, error) {
	profileName = strings.ToLower(profileName)

	// 解析浏览器类型和版本
	var browser BrowserType
	var version string

	if strings.HasPrefix(profileName, "chrome_") {
		browser = BrowserChrome
		version = strings.TrimPrefix(profileName, "chrome_")
		// 处理特殊版本
		if strings.Contains(version, "_psk") {
			version = strings.Split(version, "_psk")[0]
		}
		if strings.Contains(version, "_pq") {
			version = strings.Split(version, "_pq")[0]
		}
	} else if strings.HasPrefix(profileName, "firefox_") {
		browser = BrowserFirefox
		version = strings.TrimPrefix(profileName, "firefox_")
	} else if strings.HasPrefix(profileName, "safari_") {
		browser = BrowserSafari
		version = strings.TrimPrefix(profileName, "safari_")
	} else if strings.HasPrefix(profileName, "opera_") {
		browser = BrowserOpera
		version = strings.TrimPrefix(profileName, "opera_")
	} else {
		// 默认使用 Chrome 133
		return g.GetUserAgentWithOS("chrome_133", os)
	}

	// 生成 User-Agent
	if os == "" {
		os = RandomOS()
	}

	switch browser {
	case BrowserChrome:
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36", string(os), version), nil
	case BrowserFirefox:
		return fmt.Sprintf("Mozilla/5.0 (%s; rv:%s.0) Gecko/20100101 Firefox/%s.0", string(os), version, version), nil
	case BrowserSafari:
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/%s Safari/605.1.15", string(os), version), nil
	case BrowserOpera:
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36 OPR/%s.0.0.0", string(os), version, version), nil
	default:
		return "", fmt.Errorf("unsupported browser type: %s", browser)
	}
}

// RandomOS 随机选择一个操作系统
func RandomOS() OperatingSystem {
	if len(OperatingSystems) == 0 {
		return OSWindows10 // 默认返回 Windows 10
	}
	return utils.RandomChoice(OperatingSystems)
}

// GetUserAgentForProfile 为指定的 ClientProfile 获取 User-Agent
func GetUserAgentForProfile(profileName string) (string, error) {
	return defaultGenerator.GetUserAgent(profileName)
}

// GetUserAgentForProfileWithOS 为指定的 ClientProfile 和操作系统获取 User-Agent
func GetUserAgentForProfileWithOS(profileName string, os OperatingSystem) (string, error) {
	return defaultGenerator.GetUserAgentWithOS(profileName, os)
}
