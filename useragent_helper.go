package fingerprint

import (
	"fmt"
	"strings"
)

// GetUserAgentByProfileName 根据 profile 名称获取 User-Agent
// 这是最推荐的方式，因为可以直接匹配指纹名称
func GetUserAgentByProfileName(profileName string) (string, error) {
	if profileName == "" {
		return "", fmt.Errorf("profile name cannot be empty")
	}
	return GetUserAgentForProfile(profileName)
}

// GetUserAgentByProfileNameWithOS 根据 profile 名称和指定操作系统获取 User-Agent
func GetUserAgentByProfileNameWithOS(profileName string, os OperatingSystem) (string, error) {
	if profileName == "" {
		return "", fmt.Errorf("profile name cannot be empty")
	}
	return GetUserAgentForProfileWithOS(profileName, os)
}

// GetUserAgentFromProfile 从 ClientProfile 对象获取 User-Agent
// 通过查找 MappedTLSClients 来匹配对应的 profile 名称
func GetUserAgentFromProfile(profile ClientProfile) (string, error) {
	// 通过 ClientHelloStr 查找对应的 profile 名称
	helloStr := profile.GetClientHelloStr()
	
	// 遍历 MappedTLSClients 查找匹配的 profile
	for name, p := range MappedTLSClients {
		if p.GetClientHelloStr() == helloStr {
			return GetUserAgentForProfile(name)
		}
	}
	
	// 如果找不到，尝试从 helloStr 中推断浏览器类型
	helloStrLower := strings.ToLower(helloStr)
	if strings.Contains(helloStrLower, "chrome") {
		return GetUserAgentForProfile("chrome_133")
	} else if strings.Contains(helloStrLower, "firefox") {
		return GetUserAgentForProfile("firefox_135")
	} else if strings.Contains(helloStrLower, "safari") {
		return GetUserAgentForProfile("safari_16_0")
	} else if strings.Contains(helloStrLower, "opera") {
		return GetUserAgentForProfile("opera_91")
	}
	
	return "", fmt.Errorf("unable to infer User-Agent from ClientProfile")
}

// GetUserAgentFromProfileWithOS 从 ClientProfile 对象获取 User-Agent，并指定操作系统
func GetUserAgentFromProfileWithOS(profile ClientProfile, os OperatingSystem) (string, error) {
	helloStr := profile.GetClientHelloStr()
	
	for name, p := range MappedTLSClients {
		if p.GetClientHelloStr() == helloStr {
			return GetUserAgentForProfileWithOS(name, os)
		}
	}
	
	helloStrLower := strings.ToLower(helloStr)
	if strings.Contains(helloStrLower, "chrome") {
		return GetUserAgentForProfileWithOS("chrome_133", os)
	} else if strings.Contains(helloStrLower, "firefox") {
		return GetUserAgentForProfileWithOS("firefox_135", os)
	} else if strings.Contains(helloStrLower, "safari") {
		return GetUserAgentForProfileWithOS("safari_16_0", os)
	} else if strings.Contains(helloStrLower, "opera") {
		return GetUserAgentForProfileWithOS("opera_91", os)
	}
	
	return "", fmt.Errorf("unable to infer User-Agent from ClientProfile")
}

// GetUserAgentForMappedProfile 从 MappedTLSClients 中获取指定名称的 profile 的 User-Agent
func GetUserAgentForMappedProfile(profileName string) (string, error) {
	profile, ok := MappedTLSClients[profileName]
	if !ok {
		return "", fmt.Errorf("profile %s not found", profileName)
	}
	return GetUserAgentFromProfile(profile)
}

// GetUserAgentForMappedProfileWithOS 从 MappedTLSClients 中获取指定名称的 profile 的 User-Agent，并指定操作系统
func GetUserAgentForMappedProfileWithOS(profileName string, os OperatingSystem) (string, error) {
	profile, ok := MappedTLSClients[profileName]
	if !ok {
		return "", fmt.Errorf("profile %s not found", profileName)
	}
	return GetUserAgentFromProfileWithOS(profile, os)
}

// 辅助函数：从 profile 名称推断浏览器类型
func inferBrowserFromProfileName(profileName string) (string, string) {
	profileName = strings.ToLower(profileName)
	
	if strings.HasPrefix(profileName, "chrome_") {
		version := strings.TrimPrefix(profileName, "chrome_")
		// 移除特殊后缀
		version = strings.Split(version, "_")[0]
		return string(BrowserChrome), version
	} else if strings.HasPrefix(profileName, "firefox_") {
		version := strings.TrimPrefix(profileName, "firefox_")
		return string(BrowserFirefox), version
	} else if strings.HasPrefix(profileName, "safari_") {
		version := strings.TrimPrefix(profileName, "safari_")
		return string(BrowserSafari), version
	} else if strings.HasPrefix(profileName, "opera_") {
		version := strings.TrimPrefix(profileName, "opera_")
		return string(BrowserOpera), version
	}
	
	return string(BrowserChrome), "" // 默认返回 Chrome
}

