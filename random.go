package fingerprint

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// FingerprintResult 指纹结果，包含指纹、User-Agent 和标准 HTTP Headers
type FingerprintResult struct {
	Profile       ClientProfile // 指纹配置
	UserAgent     string        // 对应的 User-Agent
	HelloClientID string        // Client Hello ID（与 tls-client 保持一致）
	Headers       *HTTPHeaders  // 标准 HTTP 请求头（包含全球语言支持）
}

var (
	rng   *rand.Rand
	rngMu sync.Mutex
)

func init() {
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// GetRandomFingerprint 随机获取一个指纹和对应的 User-Agent
// 操作系统会随机选择
func GetRandomFingerprint() (*FingerprintResult, error) {
	return GetRandomFingerprintWithOS(OperatingSystem(""))
}

// GetRandomFingerprintWithOS 随机获取一个指纹和对应的 User-Agent，并指定操作系统
// 如果 os 为空字符串，则随机选择操作系统
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
	rngMu.Lock()
	randomName := names[rng.Intn(len(names))]
	rngMu.Unlock()
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

	browserType = toLower(browserType)

	// 筛选出指定浏览器类型的指纹
	candidates := make([]string, 0)
	for name := range MappedTLSClients {
		nameLower := toLower(name)
		if hasPrefix(nameLower, browserType+"_") {
			candidates = append(candidates, name)
		}
	}

	if len(candidates) == 0 {
		return nil, &ErrBrowserNotFound{Browser: browserType}
	}

	// 随机选择一个（线程安全）
	rngMu.Lock()
	randomName := candidates[rng.Intn(len(candidates))]
	rngMu.Unlock()
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

// 辅助函数
func toLower(s string) string {
	// 使用简单的实现，避免导入 strings 包
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + ('a' - 'A')
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func hasPrefix(s, prefix string) bool {
	if len(prefix) > len(s) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if s[i] != prefix[i] {
			return false
		}
	}
	return true
}

// containsString 检查字符串是否包含子字符串
func containsString(s, substr string) bool {
	return indexOfString(s, substr) != -1
}

func indexOfString(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// isMobileProfile 判断是否为移动端 profile
func isMobileProfile(profileName string) bool {
	name := toLower(profileName)
	return containsString(name, "ios") || containsString(name, "android") || containsString(name, "ipad") || containsString(name, "mobile")
}
