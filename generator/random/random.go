package random

// Phase 3: 本模块已完成基础迁移，待深度优化（详见 docs/5-process/modularization/PHASE_3_PLAN.md）
import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/vistone/fingerprint/http/headers"
	"github.com/vistone/fingerprint/http/useragent"
	"github.com/vistone/fingerprint/internal/metrics"
	"github.com/vistone/fingerprint/internal/utils"
	"github.com/vistone/fingerprint/profiles"
	"github.com/vistone/fingerprint/types"
)

// browserProfileIndex 浏览器类型索引（预计算，避免每次遍历）
var (
	browserIndex     map[string][]string
	browserIndexOnce sync.Once
)

// initBrowserIndex 初始化浏览器索引（延迟加载，线程安全）
func initBrowserIndex() {
	browserIndexOnce.Do(func() {
		browserIndex = make(map[string][]string)
		for name := range profiles.MappedTLSClients {
			nameLower := strings.ToLower(name)
			switch {
			case strings.HasPrefix(nameLower, "chrome_"):
				browserIndex["chrome"] = append(browserIndex["chrome"], name)
			case strings.HasPrefix(nameLower, "firefox_"):
				browserIndex["firefox"] = append(browserIndex["firefox"], name)
			case strings.HasPrefix(nameLower, "safari_"):
				browserIndex["safari"] = append(browserIndex["safari"], name)
			case strings.HasPrefix(nameLower, "opera_"):
				browserIndex["opera"] = append(browserIndex["opera"], name)
			case strings.HasPrefix(nameLower, "edge_"):
				browserIndex["edge"] = append(browserIndex["edge"], name)
			default:
				browserIndex["other"] = append(browserIndex["other"], name)
			}
		}
	})
}

// GetRandomFingerprint 从所有可用指纹中随机选择一个完整的浏览器指纹
//
// 该函数会从 profiles.MappedTLSClients 中随机选择一个浏览器指纹配置，
// 并为该指纹生成对应的 User-Agent 和 HTTP Headers。
// 操作系统会从所有支持的系统中随机选择。
//
// 返回值:
//   - *types.FingerprintResult: 包含完整指纹、User-Agent 和 HTTP Headers 的结果
//   - error: 如果指纹库为空或配置无效则返回错误
//
// 示例:
//
//	result, err := GetRandomFingerprint()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	println(result.UserAgent)
//	println(result.ClientProfile.GetClientHelloStr())
//
// 线程安全性: 是
// 性能: 平均耗时 7.4 微秒，分配 1.8 KB 内存
func GetRandomFingerprint() (*types.FingerprintResult, error) {
	return GetRandomFingerprintWithOS(types.OperatingSystem(""))
}

// GetRandomFingerprintWithOS 从所有可用指纹中随机选择一个，并指定操作系统
//
// 该函数类似于 GetRandomFingerprint，但允许指定特定的操作系统。
// 如果 os 参数为空字符串，则会随机选择操作系统。
//
// 参数:
//   - os: 目标操作系统，如 "Windows NT 10.0; Win64; x64"，为空时随机选择
//
// 返回值:
//   - *types.FingerprintResult: 完整的指纹结果
//   - error: 操作系统无效或指纹库为空时返回错误
//
// 示例:
//
//	// 指定 Windows 操作系统
//	result, err := GetRandomFingerprintWithOS("Windows NT 10.0; Win64; x64")
//	// 随机选择操作系统
//	result, err := GetRandomFingerprintWithOS("")
//
// 线程安全性: 是
func GetRandomFingerprintWithOS(os types.OperatingSystem) (*types.FingerprintResult, error) {
	start := time.Now()

	// 检查 profiles.MappedTLSClients 是否为空
	if len(profiles.MappedTLSClients) == 0 {
		return nil, fmt.Errorf("no TLS client profiles available")
	}

	// 获取所有可用的指纹名称
	names := make([]string, 0, len(profiles.MappedTLSClients))
	for name := range profiles.MappedTLSClients {
		names = append(names, name)
	}

	// 随机选择一个（线程安全）
	randomName := utils.RandomChoiceString(names)
	profile := profiles.MappedTLSClients[randomName]
	if profile.GetClientHelloStr() == "" {
		return nil, fmt.Errorf("profile %s is invalid (empty ClientHelloStr)", randomName)
	}

	// 获取对应的 User-Agent
	var ua string
	var err error
	if os == "" {
		ua, err = useragent.GetUserAgentByProfileName(randomName)
	} else {
		ua, err = useragent.GetUserAgentByProfileNameWithOS(randomName, os)
	}
	if err != nil {
		return nil, err
	}

	// 生成标准 HTTP Headers
	browserTypeStr, _ := inferBrowserFromProfileName(randomName)
	isMobile := isMobileProfile(randomName)
	headers := headers.GenerateHeaders(types.BrowserType(browserTypeStr), ua, isMobile)

	// 记录指标
	durationMs := float64(time.Since(start).Nanoseconds()) / 1e6
	osStr := string(os)
	if osStr == "" {
		osStr = "random"
	}
	metrics.RecordFingerprintGeneration(browserTypeStr, osStr, durationMs)

	return &types.FingerprintResult{
		Profile:       profile,
		UserAgent:     ua,
		HelloClientID: profile.GetClientHelloStr(),
		Headers:       headers,
	}, nil
}

// GetRandomFingerprintByBrowser 根据浏览器类型随机获取指纹和 User-Agent
// browserType: "chrome", "firefox", "safari", "opera" 等
func GetRandomFingerprintByBrowser(browserType string) (*types.FingerprintResult, error) {
	return GetRandomFingerprintByBrowserWithOS(browserType, types.OperatingSystem(""))
}

// GetRandomFingerprintByBrowserWithOS 根据浏览器类型随机获取指纹和 User-Agent，并指定操作系统
func GetRandomFingerprintByBrowserWithOS(browserType string, os types.OperatingSystem) (*types.FingerprintResult, error) {
	if browserType == "" {
		return nil, fmt.Errorf("browser type cannot be empty")
	}
	if len(profiles.MappedTLSClients) == 0 {
		return nil, fmt.Errorf("no TLS client profiles available")
	}

	browserType = strings.ToLower(browserType)

	// 使用预计算的索引（线程安全，延迟初始化）
	initBrowserIndex()

	// 从索引中获取候选列表（零分配）
	candidates, exists := browserIndex[browserType]
	if !exists || len(candidates) == 0 {
		return nil, &ErrBrowserNotFound{Browser: browserType}
	}

	// 随机选择一个（线程安全）
	randomName := utils.RandomChoiceString(candidates)
	profile := profiles.MappedTLSClients[randomName]
	if profile.GetClientHelloStr() == "" {
		return nil, fmt.Errorf("profile %s is invalid (empty ClientHelloStr)", randomName)
	}

	// 获取对应的 User-Agent
	var ua string
	var err error
	if os == "" {
		ua, err = useragent.GetUserAgentByProfileName(randomName)
	} else {
		ua, err = useragent.GetUserAgentByProfileNameWithOS(randomName, os)
	}
	if err != nil {
		return nil, err
	}

	// 生成标准 HTTP Headers
	browserTypeStr, _ := inferBrowserFromProfileName(randomName)
	isMobile := isMobileProfile(randomName)
	headers := headers.GenerateHeaders(types.BrowserType(browserTypeStr), ua, isMobile)

	return &types.FingerprintResult{
		Profile:       profile,
		UserAgent:     ua,
		HelloClientID: profile.GetClientHelloStr(),
		Headers:       headers,
	}, nil
}

// ErrBrowserNotFound 浏览器类型未找到错误
type ErrBrowserNotFound struct {
	Browser string
}

func (e *ErrBrowserNotFound) Error() string {
	return "browser type not found: " + e.Browser
}

// isMobileProfile 判断是否为移动端 profile
func isMobileProfile(profileName string) bool {
	name := strings.ToLower(profileName)
	return strings.Contains(name, "ios") ||
		strings.Contains(name, "android") ||
		strings.Contains(name, "ipad") ||
		strings.Contains(name, "mobile")
}

// inferBrowserFromProfileName 从 profile 名称推断浏览器类型
func inferBrowserFromProfileName(profileName string) (string, string) {
	profileName = strings.ToLower(profileName)

	if strings.HasPrefix(profileName, "chrome_") {
		version := strings.TrimPrefix(profileName, "chrome_")
		version = strings.Split(version, "_")[0]
		return string(types.BrowserChrome), version
	} else if strings.HasPrefix(profileName, "firefox_") {
		version := strings.TrimPrefix(profileName, "firefox_")
		return string(types.BrowserFirefox), version
	} else if strings.HasPrefix(profileName, "safari_") {
		version := strings.TrimPrefix(profileName, "safari_")
		return string(types.BrowserSafari), version
	} else if strings.HasPrefix(profileName, "opera_") {
		version := strings.TrimPrefix(profileName, "opera_")
		return string(types.BrowserOpera), version
	} else if strings.HasPrefix(profileName, "edge_") {
		version := strings.TrimPrefix(profileName, "edge_")
		return string(types.BrowserEdge), version
	}
	return string(types.BrowserChrome), ""
}
