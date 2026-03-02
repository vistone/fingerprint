package defense

// TODO@Phase-3: 本模块暂未迁移（参见 docs/5-process/modularization/PHASE_3_PLAN.md）
import (
	"math"
	"strings"

	"github.com/vistone/fingerprint/types"
)

// AnomalyDetector 异常指纹检测器
// 用于分析指纹数据中的可疑模式，判断是否为机器人或伪造指纹
type AnomalyDetector struct{}

// NewAnomalyDetector 创建新的异常检测器
func NewAnomalyDetector() *AnomalyDetector {
	return &AnomalyDetector{}
}

// DetectAnomalies 检测指纹数据中的异常
// data: 指纹原始字节数据
// 返回 true 表示检测到异常
func (d *AnomalyDetector) DetectAnomalies(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// 检查1：熵值过低（重复模式）
	if d.hasLowEntropy(data) {
		return true
	}

	// 检查2：熵值过高（完全随机）
	if d.hasExcessiveEntropy(data) {
		return true
	}

	// 检查3：已知自动化工具特征
	if d.containsSpoofingMarkers(data) {
		return true
	}

	return false
}

// DetectHeadlessBrowser 检测 User-Agent 是否为无头浏览器
func (d *AnomalyDetector) DetectHeadlessBrowser(userAgent string) bool {
	uaLower := strings.ToLower(userAgent)
	headlessMarkers := []string{
		"headlesschrome",
		"phantomjs",
		"selenium",
		"webdriver",
		"puppeteer",
		"playwright",
		"cypress",
		"jsdom",
		"zombie",
		"htmlunit",
	}
	for _, marker := range headlessMarkers {
		if strings.Contains(uaLower, marker) {
			return true
		}
	}
	return false
}

// hasLowEntropy 检查数据熵值是否过低（可疑的统一数据）
func (d *AnomalyDetector) hasLowEntropy(data []byte) bool {
	if len(data) < 10 {
		return false
	}
	var byteCounts [256]int
	for _, b := range data {
		byteCounts[b]++
	}
	uniqueBytes := 0
	for _, count := range byteCounts {
		if count > 0 {
			uniqueBytes++
		}
	}
	// 少于 26/256 ≈ 10% 的不同字节值，视为可疑
	return uniqueBytes < 26
}

// hasExcessiveEntropy 检查数据熵值是否过高（过于随机）
func (d *AnomalyDetector) hasExcessiveEntropy(data []byte) bool {
	if len(data) < 20 {
		return false
	}
	var byteCounts [256]int
	for _, b := range data {
		byteCounts[b]++
	}
	// 计算 Shannon 熵
	n := float64(len(data))
	entropy := 0.0
	for _, count := range byteCounts {
		if count > 0 {
			p := float64(count) / n
			entropy -= p * math.Log2(p)
		}
	}
	// 超过 7.5 bits 则视为过于随机
	return entropy > 7.5
}

// containsSpoofingMarkers 检查是否包含已知自动化工具特征
func (d *AnomalyDetector) containsSpoofingMarkers(data []byte) bool {
	patterns := [][]byte{
		[]byte("HeadlessChrome"),
		[]byte("PhantomJS"),
		[]byte("webdriver"),
		[]byte("selenium"),
		[]byte("puppeteer"),
	}
	for _, pattern := range patterns {
		if len(pattern) <= len(data) {
			for i := 0; i <= len(data)-len(pattern); i++ {
				match := true
				for j := 0; j < len(pattern); j++ {
					if data[i+j] != pattern[j] {
						match = false
						break
					}
				}
				if match {
					return true
				}
			}
		}
	}
	return false
}

// ContradictionDetector 矛盾指纹检测器
// 检测指纹属性之间的逻辑不一致性
type ContradictionDetector struct{}

// NewContradictionDetector 创建新的矛盾检测器
func NewContradictionDetector() *ContradictionDetector {
	return &ContradictionDetector{}
}

// CheckContradictions 检查指纹属性是否存在矛盾
// attributes: 属性键值对列表，如 [("os", "Windows"), ("platform", "Linux")]
// 返回 true 表示发现矛盾
func (c *ContradictionDetector) CheckContradictions(attributes map[string]string) bool {
	if len(attributes) == 0 {
		return false
	}

	// 检查操作系统与平台矛盾
	if os, ok := attributes["os"]; ok {
		if platform, ok := attributes["platform"]; ok {
			if c.hasOSPlatformContradiction(os, platform) {
				return true
			}
		}
	}

	// 检查 User-Agent 与特性矛盾
	if ua, ok := attributes["user_agent"]; ok {
		if features, ok := attributes["features"]; ok {
			if c.hasUserAgentFeatureContradiction(ua, features) {
				return true
			}
		}
	}

	// 检查移动设备与屏幕分辨率矛盾
	if isMobile, ok := attributes["is_mobile"]; ok {
		if screenWidth, ok := attributes["screen_width"]; ok {
			if c.hasMobileScreenContradiction(isMobile, screenWidth) {
				return true
			}
		}
	}

	// 检查 User-Agent 与操作系统矛盾
	if ua, ok := attributes["user_agent"]; ok {
		if os, ok := attributes["os"]; ok {
			if c.hasUAOSContradiction(ua, os) {
				return true
			}
		}
	}

	return false
}

// hasOSPlatformContradiction 检查操作系统与平台的矛盾
func (c *ContradictionDetector) hasOSPlatformContradiction(os, platform string) bool {
	if strings.Contains(os, "Windows") && !strings.Contains(platform, "Win") {
		return true
	}
	if strings.Contains(os, "Mac") && !strings.Contains(platform, "Mac") {
		return true
	}
	if strings.Contains(os, "Linux") && !strings.Contains(platform, "Linux") && !strings.Contains(platform, "X11") {
		return true
	}
	return false
}

// hasUserAgentFeatureContradiction 检查 User-Agent 与特性的矛盾
func (c *ContradictionDetector) hasUserAgentFeatureContradiction(userAgent, features string) bool {
	// 旧版浏览器不应支持现代特性
	if strings.Contains(userAgent, "Chrome/60") && strings.Contains(features, "WebGL2") {
		return true
	}
	// 移动版浏览器不应声明桌面特性
	if strings.Contains(userAgent, "Mobile") && strings.Contains(features, "desktop") {
		return true
	}
	return false
}

// hasMobileScreenContradiction 检查移动设备与屏幕尺寸的矛盾
func (c *ContradictionDetector) hasMobileScreenContradiction(isMobile, screenWidth string) bool {
	width := 0
	_, err := parseIntFallback(screenWidth, &width)
	if err != nil {
		return false
	}
	// 移动设备使用桌面分辨率可疑
	if isMobile == "true" && width > 1920 {
		return true
	}
	// 桌面设备使用极小分辨率可疑
	if isMobile == "false" && width < 800 {
		return true
	}
	return false
}

// hasUAOSContradiction 检查 User-Agent 与操作系统的矛盾
func (c *ContradictionDetector) hasUAOSContradiction(userAgent, os string) bool {
	uaLower := strings.ToLower(userAgent)
	osLower := strings.ToLower(os)

	// Windows UA 声称 Mac 系统
	if strings.Contains(uaLower, "windows") && strings.Contains(osLower, "mac") {
		return true
	}
	// Mac UA 声称 Windows 系统
	if strings.Contains(uaLower, "macintosh") && strings.Contains(osLower, "windows") {
		return true
	}
	// Linux UA 声称 Windows 系统
	if strings.Contains(uaLower, "x11; linux") && strings.Contains(osLower, "windows") {
		return true
	}
	return false
}

// parseIntFallback 简单整数解析，不依赖 strconv
func parseIntFallback(s string, result *int) (int, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return 0, &parseError{s: s}
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, &parseError{s: s}
		}
		n = n*10 + int(c-'0')
	}
	*result = n
	return n, nil
}

type parseError struct{ s string }

func (e *parseError) Error() string { return "parse error: " + e.s }

// PassiveRecognizer 被动指纹识别器
// 通过分析 HTTP 请求头推断浏览器和操作系统
type PassiveRecognizer struct{}

// NewPassiveRecognizer 创建新的被动识别器
func NewPassiveRecognizer() *PassiveRecognizer {
	return &PassiveRecognizer{}
}

// RecognitionResult 被动识别结果
type RecognitionResult struct {
	// 检测到的浏览器类型
	Browser types.BrowserType
	// 检测到的操作系统
	OS types.OperatingSystem
	// 检测到的浏览器版本
	BrowserVersion string
	// 置信度（0.0-1.0）
	Confidence float64
	// 是否为移动设备
	IsMobile bool
	// 是否疑似机器人
	IsBot bool
}

// RecognizeFromHeaders 从 HTTP 请求头识别浏览器指纹
func (r *PassiveRecognizer) RecognizeFromHeaders(headers map[string]string) *RecognitionResult {
	result := &RecognitionResult{}

	ua := headers["User-Agent"]
	if ua == "" {
		result.Confidence = 0.0
		return result
	}

	// 检测机器人
	anomalyDetector := NewAnomalyDetector()
	if anomalyDetector.DetectHeadlessBrowser(ua) {
		result.IsBot = true
		result.Confidence = 0.9
		return result
	}

	uaLower := strings.ToLower(ua)

	// 检测移动设备
	result.IsMobile = strings.Contains(uaLower, "mobile") ||
		strings.Contains(uaLower, "android") ||
		strings.Contains(uaLower, "iphone") ||
		strings.Contains(uaLower, "ipad")

	// 识别浏览器
	result.Browser, result.BrowserVersion = detectBrowserFromUA(ua)

	// 识别操作系统
	result.OS = detectOSFromUA(ua)

	// 计算置信度
	result.Confidence = calculateConfidence(headers, result)

	return result
}

// detectBrowserFromUA 从 User-Agent 检测浏览器类型和版本
func detectBrowserFromUA(ua string) (types.BrowserType, string) {
	uaLower := strings.ToLower(ua)

	// Edge 必须在 Chrome 之前检测（Edge UA 也包含 Chrome）
	if strings.Contains(uaLower, "edg/") || strings.Contains(uaLower, "edge/") {
		version := extractVersionFromUA(ua, "Edg/")
		if version == "" {
			version = extractVersionFromUA(ua, "Edge/")
		}
		return types.BrowserEdge, version
	}

	// Opera 必须在 Chrome 之前检测
	if strings.Contains(uaLower, "opr/") {
		version := extractVersionFromUA(ua, "OPR/")
		return types.BrowserOpera, version
	}

	// Chrome
	if strings.Contains(uaLower, "chrome/") && !strings.Contains(uaLower, "chromium") {
		version := extractVersionFromUA(ua, "Chrome/")
		return types.BrowserChrome, version
	}

	// Firefox
	if strings.Contains(uaLower, "firefox/") {
		version := extractVersionFromUA(ua, "Firefox/")
		return types.BrowserFirefox, version
	}

	// Safari
	if strings.Contains(uaLower, "safari/") {
		version := extractVersionFromUA(ua, "Version/")
		return types.BrowserSafari, version
	}

	return types.BrowserChrome, ""
}

// extractVersionFromUA 从 User-Agent 提取特定标识符后的版本号
func extractVersionFromUA(ua, prefix string) string {
	idx := strings.Index(ua, prefix)
	if idx == -1 {
		return ""
	}
	start := idx + len(prefix)
	end := start
	for end < len(ua) && ua[end] != '.' && ua[end] != ' ' && ua[end] != ';' {
		end++
	}
	if end > start {
		return ua[start:end]
	}
	return ""
}

// detectOSFromUA 从 User-Agent 检测操作系统
func detectOSFromUA(ua string) types.OperatingSystem {
	if strings.Contains(ua, "Windows NT 10.0") {
		return types.OSWindows10
	}
	if strings.Contains(ua, "Macintosh; Intel Mac OS X 15") {
		return types.OSMacOS15
	}
	if strings.Contains(ua, "Macintosh; Intel Mac OS X 14") {
		return types.OSMacOS14
	}
	if strings.Contains(ua, "Macintosh; Intel Mac OS X 13") {
		return types.OSMacOS13
	}
	if strings.Contains(ua, "X11; Linux") {
		return types.OSLinux
	}
	return types.OSWindows10 // 默认
}

// calculateConfidence 计算识别置信度
func calculateConfidence(headers map[string]string, result *RecognitionResult) float64 {
	score := 0.5

	// 有 User-Agent 加分
	if headers["User-Agent"] != "" {
		score += 0.2
	}

	// 有 Accept 头加分
	if headers["Accept"] != "" {
		score += 0.1
	}

	// 有 Accept-Language 加分
	if headers["Accept-Language"] != "" {
		score += 0.1
	}

	// Chrome 类浏览器有 Sec-CH-UA 加分
	if headers["Sec-CH-UA"] != "" {
		score += 0.1
	}

	return math.Min(score, 1.0)
}

// RecognizeFromUserAgent 仅从 User-Agent 字符串识别浏览器指纹
func RecognizeFromUserAgent(userAgent string) *RecognitionResult {
	recognizer := NewPassiveRecognizer()
	return recognizer.RecognizeFromHeaders(map[string]string{
		"User-Agent": userAgent,
	})
}
