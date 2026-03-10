package useragent

import (
	"fmt"
	"strings"

	"github.com/vistone/fingerprint/modules/core/types"
	profiles "github.com/vistone/fingerprint/modules/profiles/legacy"
)

// GetUserAgentByProfileName get User-Agent by profile name
// This is the recommended approach as it directly matches the fingerprint name
func GetUserAgentByProfileName(profileName string) (string, error) {
	if profileName == "" {
		return "", fmt.Errorf("profile name cannot be empty")
	}
	return GetUserAgentForProfile(profileName)
}

// GetUserAgentByProfileNameWithOS get User-Agent by profile name with specified operating system
func GetUserAgentByProfileNameWithOS(profileName string, os types.OperatingSystem) (string, error) {
	if profileName == "" {
		return "", fmt.Errorf("profile name cannot be empty")
	}
	return GetUserAgentForProfileWithOS(profileName, os)
}

// GetUserAgentFromProfile get User-Agent from profiles.ClientProfile object
// Match corresponding profile name by searching profiles.MappedTLSClients
func GetUserAgentFromProfile(profile profiles.ClientProfile) (string, error) {
	// Find corresponding profile name through ClientHelloStr
	helloStr := profile.GetClientHelloStr()

	// Iterate through profiles.MappedTLSClients to find matching profile
	for name, p := range profiles.MappedTLSClients {
		if p.GetClientHelloStr() == helloStr {
			return GetUserAgentForProfile(name)
		}
	}

	// If not found, try to infer browser type from helloStr
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

// GetUserAgentFromProfileWithOS get User-Agent from profiles.ClientProfile object with specified operating system
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

// inferBrowserFromProfileName infer browser type from profile name
func inferBrowserFromProfileName(profileName string) (string, string) {
	profileName = strings.ToLower(profileName)

	if strings.HasPrefix(profileName, "chrome_") {
		version := strings.TrimPrefix(profileName, "chrome_")
		// Remove special suffix
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

	return string(types.BrowserChrome), "" // Default to Chrome
}
