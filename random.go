package fingerprint

import (
	"fmt"
	"strings"

	"github.com/vistone/fingerprint/internal/utils"
)

// GetRandomFingerprint 从所有可用指纹中随机选择一个完整的浏览器指纹
//
// 该函数会从 MappedTLSClients 中随机选择一个浏览器指纹配置，
// 并为该指纹生成对应的 User-Agent 和 HTTP Headers。
// 操作系统会从所有支持的系统中随机选择。
//
// 返回值:
//   - *FingerprintResult: 包含完整指纹、User-Agent 和 HTTP Headers 的结果
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
func GetRandomFingerprint() (*FingerprintResult, error) {
	return GetRandomFingerprintWithOS(OperatingSystem(""))
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
//   - *FingerprintResult: 完整的指纹结果
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
func GetRandomFingerprintWithOS(os OperatingSystem) (*FingerprintResult, error) {
	// 检查 MappedTLSClients 是否为空
	if len(MappedTLSClients) == 0 {
		return nil, fmt.Errorf("no TLS client profiles available")
	}

	// 获取所有可用的指纹名称
	names := make([]string, 0, len(MappedTLSClients))
	for name := range MappedTLSClients {
		names = append(names, name)
	}

	// 随机选择一个（线程安全）
	randomName := utils.RandomChoiceString(names)
	profile := MappedTLSClients[randomName]
	if profile.GetClientHelloStr() == "" {
		return nil, fmt.Errorf("profile %s is invalid (empty ClientHelloStr)", randomName)
	}

	// 获取对应的 User-Agent
	var ua string
	var err error
	if os == "" {
		ua, err = GetUserAgentByProfileName(randomName)
	} else {
		ua, err = GetUserAgentByProfileNameWithOS(randomName, os)
	}
	if err != nil {
		return nil, err
	}

	// 生成标准 HTTP Headers
	browserTypeStr, _ := inferBrowserFromProfileName(randomName)
	isMobile := isMobileProfile(randomName)
	headers := GenerateHeaders(BrowserType(browserTypeStr), ua, isMobile)

	return &FingerprintResult{
		Profile:       profile,
		UserAgent:     ua,
		HelloClientID: profile.GetClientHelloStr(),
		Headers:       headers,
	}, nil
}

// GetRandomFingerprintByBrowser 根据浏览器类型随机获取指纹和 User-Agent
// browserType: "chrome", "firefox", "safari", "opera" 等
func GetRandomFingerprintByBrowser(browserType string) (*FingerprintResult, error) {
	return GetRandomFingerprintByBrowserWithOS(browserType, OperatingSystem(""))
}

// GetRandomFingerprintByBrowserWithOS 根据浏览器类型随机获取指纹和 User-Agent，并指定操作系统
func GetRandomFingerprintByBrowserWithOS(browserType string, os OperatingSystem) (*FingerprintResult, error) {
	if browserType == "" {
		return nil, fmt.Errorf("browser type cannot be empty")
	}
	if len(MappedTLSClients) == 0 {
		return nil, fmt.Errorf("no TLS client profiles available")
	}

	browserType = strings.ToLower(browserType)

	// 筛选出指定浏览器类型的指纹
	candidates := make([]string, 0)
	for name := range MappedTLSClients {
		nameLower := strings.ToLower(name)
		if strings.HasPrefix(nameLower, browserType+"_") {
			candidates = append(candidates, name)
		}
	}

	if len(candidates) == 0 {
		return nil, &ErrBrowserNotFound{Browser: browserType}
	}

	// 随机选择一个（线程安全）
	randomName := utils.RandomChoiceString(candidates)
	profile := MappedTLSClients[randomName]
	if profile.GetClientHelloStr() == "" {
		return nil, fmt.Errorf("profile %s is invalid (empty ClientHelloStr)", randomName)
	}

	// 获取对应的 User-Agent
	var ua string
	var err error
	if os == "" {
		ua, err = GetUserAgentByProfileName(randomName)
	} else {
		ua, err = GetUserAgentByProfileNameWithOS(randomName, os)
	}
	if err != nil {
		return nil, err
	}

	// 生成标准 HTTP Headers
	browserTypeStr, _ := inferBrowserFromProfileName(randomName)
	isMobile := isMobileProfile(randomName)
	headers := GenerateHeaders(BrowserType(browserTypeStr), ua, isMobile)

	return &FingerprintResult{
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
