package useragent

import (
	"fmt"
	"strings"

	"github.com/vistone/fingerprint/modules/profiles/legacy"
	"github.com/vistone/fingerprint/modules/core/types"
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
func GetUserAgentByProfileNameWithOS(profileName string, os types.OperatingSystem) (string, error) {
	if profileName == "" {
		return "", fmt.Errorf("profile name cannot be empty")
	}
	return GetUserAgentForProfileWithOS(profileName, os)
}

// GetUserAgentFromProfile 从 profiles.ClientProfile 对象获取 User-Agent
// 通过查找 profiles.MappedTLSClients 来匹配对应的 profile 名称
func GetUserAgentFromProfile(profile profiles.ClientProfile) (string, error) {
	// 通过 ClientHelloStr 查找对应的 profile 名称
	helloStr := profile.GetClientHelloStr()

	// 遍历 profiles.MappedTLSClients 查找匹配的 profile
	for name, p := range profiles.MappedTLSClients {
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

	return "", fmt.Errorf("unable to infer User-Agent from profiles.ClientProfile")
}

// GetUserAgentFromProfileWithOS 从 profiles.ClientProfile 对象获取 User-Agent，并指定操作系统
func GetUserAgentFromProfileWithOS(profile profiles.ClientProfile, os types.OperatingSystem) (string, error) {
	helloStr := profile.GetClientHelloStr()

	for name, p := range profiles.MappedTLSClients {
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

	return "", fmt.Errorf("unable to infer User-Agent from profiles.ClientProfile")
}

// inferBrowserFromProfileName 从 profile 名称推断浏览器类型
func inferBrowserFromProfileName(profileName string) (string, string) {
	profileName = strings.ToLower(profileName)

	if strings.HasPrefix(profileName, "chrome_") {
		version := strings.TrimPrefix(profileName, "chrome_")
		// 移除特殊后缀
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

	return string(types.BrowserChrome), "" // 默认返回 Chrome
}
