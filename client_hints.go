package fingerprint

import (
	"fmt"
	"strings"
)

// ClientHintsPolicy Client Hints 策略配置
type ClientHintsPolicy struct {
	// 低熵提示（默认发送）
	SendLowEntropyHints bool

	// 高熵提示（需服务器 Accept-CH 请求）
	HighEntropyHints []string

	// 是否支持委托跨源
	SupportsCrossOriginDelegation bool

	// Permissions-Policy 配置
	PermissionsPolicy map[string]string
}

// ClientHintsData 完整的 Client Hints 数据
type ClientHintsData struct {
	// 低熵提示（总是可用）
	SecCHUA         string // "Google Chrome";v="120", "Chromium";v="120"
	SecCHUAMobile   string // ?0 或 ?1
	SecCHUAPlatform string // "Windows", "macOS", "Linux", "Android", "iOS"

	// 高熵提示（需 Accept-CH 授权）
	SecCHUAArch               string // "x86", "arm"
	SecCHUABitness            string // "64", "32"
	SecCHUAFullVersionList    string // 完整版本列表
	SecCHUAModel              string // 设备型号（移动端）
	SecCHUAPlatformVersion    string // 平台版本号
	SecCHUAWoW64              string // Windows 64 位仿真标记
	SecCHUAFormFactor         string // "Desktop", "Mobile", "Tablet", "VR"
	SecCHPreferredColorScheme string // "dark", "light"
	SecCHPrefersReducedMotion string // "reduce", "no-preference"

	// 视口和网络提示
	ViewportWidth string // 视口宽度
	DeviceMemory  string // 设备内存（GB）
	DPR           string // 设备像素比
	DownlinkSpeed string // 下行带宽（Mbps）
	ECT           string // 有效连接类型："slow-2g", "2g", "3g", "4g"
	RTT           string // 往返时间（ms）
	SaveData      string // "on"/"off"
}

// NewClientHintsPolicy 创建默认策略
func NewClientHintsPolicy(browserType BrowserType) *ClientHintsPolicy {
	policy := &ClientHintsPolicy{
		SendLowEntropyHints:           true,
		SupportsCrossOriginDelegation: false,
		PermissionsPolicy:             make(map[string]string),
	}

	// 根据浏览器配置高熵提示
	switch browserType {
	case BrowserChrome:
		policy.HighEntropyHints = []string{
			"Sec-CH-UA-Arch",
			"Sec-CH-UA-Bitness",
			"Sec-CH-UA-Full-Version-List",
			"Sec-CH-UA-Platform-Version",
			"Sec-CH-UA-Model",
			"Sec-CH-UA-WoW64",
		}
		policy.SupportsCrossOriginDelegation = true
		policy.PermissionsPolicy["ch-ua"] = "self"
		policy.PermissionsPolicy["ch-ua-arch"] = "()"
		policy.PermissionsPolicy["ch-ua-model"] = "()"
		policy.PermissionsPolicy["ch-ua-platform"] = "self"

	case BrowserEdge:
		policy.HighEntropyHints = []string{
			"Sec-CH-UA-Arch",
			"Sec-CH-UA-Bitness",
			"Sec-CH-UA-Full-Version-List",
			"Sec-CH-UA-Platform-Version",
			"Sec-CH-UA-WoW64",
		}
		policy.SupportsCrossOriginDelegation = true
		policy.PermissionsPolicy["ch-ua"] = "self"

	case BrowserFirefox, BrowserSafari:
		// Firefox 和 Safari 目前不支持 Client Hints
		policy.HighEntropyHints = []string{}
	}

	return policy
}

// GenerateClientHintsFromProfile 从 profile 生成 Client Hints
func GenerateClientHintsFromProfile(profile *ClientProfile, policy *ClientHintsPolicy) *ClientHintsData {
	hints := &ClientHintsData{}

	if !policy.SendLowEntropyHints {
		return hints
	}

	// 低熵提示（总是发送）
	hints.SecCHUA = generateSecCHUA(profile)
	hints.SecCHUAMobile = generateSecCHUAMobile(profile)
	hints.SecCHUAPlatform = generateSecCHUAPlatform(profile)

	// 高熵提示（仅当策略允许时）
	if contains(policy.HighEntropyHints, "Sec-CH-UA-Arch") {
		hints.SecCHUAArch = `"` + profile.OSArch + `"`
	}
	if contains(policy.HighEntropyHints, "Sec-CH-UA-Bitness") {
		hints.SecCHUABitness = `"` + profile.OSBitness + `"`
	}
	if contains(policy.HighEntropyHints, "Sec-CH-UA-Full-Version-List") {
		hints.SecCHUAFullVersionList = generateFullVersionList(profile)
	}
	if contains(policy.HighEntropyHints, "Sec-CH-UA-Platform-Version") {
		hints.SecCHUAPlatformVersion = `"` + profile.OSVersion + `"`
	}
	if contains(policy.HighEntropyHints, "Sec-CH-UA-Model") && profile.IsMobile {
		hints.SecCHUAModel = `"` + profile.DeviceModel + `"`
	}
	if contains(policy.HighEntropyHints, "Sec-CH-UA-WoW64") {
		hints.SecCHUAWoW64 = "?0"
	}

	return hints
}

// ProcessAcceptCH 处理服务器的 Accept-CH 响应头
func (p *ClientHintsPolicy) ProcessAcceptCH(acceptCHValue string) {
	if acceptCHValue == "" {
		return
	}

	// 解析 Accept-CH 头
	hints := strings.Split(acceptCHValue, ",")
	for _, hint := range hints {
		hint = strings.TrimSpace(hint)
		if hint == "" {
			continue
		}

		// 只添加我们支持的高熵提示
		if isSupportedHighEntropyHint(hint) && !contains(p.HighEntropyHints, hint) {
			p.HighEntropyHints = append(p.HighEntropyHints, hint)
		}
	}
}

// ApplyToHeaders 将 Client Hints 应用到 HTTP 头
func (hints *ClientHintsData) ApplyToHeaders(headers map[string]string) {
	if hints.SecCHUA != "" {
		headers["Sec-CH-UA"] = hints.SecCHUA
	}
	if hints.SecCHUAMobile != "" {
		headers["Sec-CH-UA-Mobile"] = hints.SecCHUAMobile
	}
	if hints.SecCHUAPlatform != "" {
		headers["Sec-CH-UA-Platform"] = hints.SecCHUAPlatform
	}
	if hints.SecCHUAArch != "" {
		headers["Sec-CH-UA-Arch"] = hints.SecCHUAArch
	}
	if hints.SecCHUABitness != "" {
		headers["Sec-CH-UA-Bitness"] = hints.SecCHUABitness
	}
	if hints.SecCHUAFullVersionList != "" {
		headers["Sec-CH-UA-Full-Version-List"] = hints.SecCHUAFullVersionList
	}
	if hints.SecCHUAPlatformVersion != "" {
		headers["Sec-CH-UA-Platform-Version"] = hints.SecCHUAPlatformVersion
	}
	if hints.SecCHUAModel != "" {
		headers["Sec-CH-UA-Model"] = hints.SecCHUAModel
	}
	if hints.SecCHUAWoW64 != "" {
		headers["Sec-CH-UA-WoW64"] = hints.SecCHUAWoW64
	}
	if hints.DeviceMemory != "" {
		headers["Device-Memory"] = hints.DeviceMemory
	}
	if hints.DPR != "" {
		headers["DPR"] = hints.DPR
	}
	if hints.ViewportWidth != "" {
		headers["Viewport-Width"] = hints.ViewportWidth
	}
	if hints.DownlinkSpeed != "" {
		headers["Downlink"] = hints.DownlinkSpeed
	}
	if hints.ECT != "" {
		headers["ECT"] = hints.ECT
	}
	if hints.RTT != "" {
		headers["RTT"] = hints.RTT
	}
	if hints.SaveData != "" {
		headers["Save-Data"] = hints.SaveData
	}
}

// ============ 辅助函数 ============

func generateSecCHUA(profile *ClientProfile) string {
	// 格式: "Brand";v="major", "Brand";v="major"
	// Chrome 示例: "Not A(Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"

	version := profile.BrowserVersion
	majorVersion := strings.Split(version, ".")[0]

	browserType := strings.ToLower(profile.BrowserType)

	switch browserType {
	case "chrome":
		return fmt.Sprintf(`"Not A(Brand";v="8", "Chromium";v="%s", "Google Chrome";v="%s"`,
			majorVersion, majorVersion)
	case "edge":
		return fmt.Sprintf(`"Not A(Brand";v="8", "Chromium";v="%s", "Microsoft Edge";v="%s"`,
			majorVersion, majorVersion)
	default:
		return ""
	}
}

func generateSecCHUAMobile(profile *ClientProfile) string {
	if profile.IsMobile {
		return "?1"
	}
	return "?0"
}

func generateSecCHUAPlatform(profile *ClientProfile) string {
	// 根据操作系统返回标准平台名
	os := strings.ToLower(profile.OS)

	switch {
	case strings.Contains(os, "windows"):
		return `"Windows"`
	case strings.Contains(os, "mac"):
		return `"macOS"`
	case strings.Contains(os, "linux"):
		return `"Linux"`
	case strings.Contains(os, "android"):
		return `"Android"`
	case strings.Contains(os, "ios"):
		return `"iOS"`
	default:
		return `"Unknown"`
	}
}

func generateFullVersionList(profile *ClientProfile) string {
	// 格式: "Brand";v="full.version", "Brand";v="full.version"
	version := profile.BrowserVersion

	browserType := strings.ToLower(profile.BrowserType)

	switch browserType {
	case "chrome":
		return fmt.Sprintf(`"Not A(Brand";v="8.0.0.0", "Chromium";v="%s", "Google Chrome";v="%s"`,
			version, version)
	case "edge":
		return fmt.Sprintf(`"Not A(Brand";v="8.0.0.0", "Chromium";v="%s", "Microsoft Edge";v="%s"`,
			version, version)
	default:
		return ""
	}
}

func isSupportedHighEntropyHint(hint string) bool {
	supportedHints := []string{
		"Sec-CH-UA-Arch",
		"Sec-CH-UA-Bitness",
		"Sec-CH-UA-Full-Version-List",
		"Sec-CH-UA-Platform-Version",
		"Sec-CH-UA-Model",
		"Sec-CH-UA-WoW64",
		"Device-Memory",
		"DPR",
		"Viewport-Width",
		"Downlink",
		"ECT",
		"RTT",
		"Save-Data",
	}

	hint = strings.TrimSpace(hint)
	for _, supported := range supportedHints {
		if strings.EqualFold(hint, supported) {
			return true
		}
	}
	return false
}
