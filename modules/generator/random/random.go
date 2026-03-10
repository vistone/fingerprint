package random

// translated comment
import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/http/legacy/headers"
	"github.com/vistone/fingerprint/modules/http/legacy/useragent"
	"github.com/vistone/fingerprint/modules/internal/metrics"
	"github.com/vistone/fingerprint/modules/kit"
	"github.com/vistone/fingerprint/modules/profiles/legacy"
	"github.com/vistone/fingerprint/modules/core/types"
)

// translated comment
var (
	browserIndex     map[string][]string
	browserIndexOnce sync.Once
)

// translated comment
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

// translated comment
//
// translated comment
// translated comment
// translated comment
//
// translated comment
// translated comment
// translated comment
//
// translated comment
//
//	result, err := GetRandomFingerprint()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	println(result.UserAgent)
//	println(result.ClientProfile.GetClientHelloStr())
//
// translated comment
// translated comment
func GetRandomFingerprint() (*types.FingerprintResult, error) {
	return GetRandomFingerprintWithOS(types.OperatingSystem(""))
}

// translated comment
//
// translated comment
// translated comment
//
// translated comment
// translated comment
//
// translated comment
// translated comment
// translated comment
//
// translated comment
//
// translated comment
//	result, err := GetRandomFingerprintWithOS("Windows NT 10.0; Win64; x64")
// translated comment
//	result, err := GetRandomFingerprintWithOS("")
//
// translated comment
func GetRandomFingerprintWithOS(os types.OperatingSystem) (*types.FingerprintResult, error) {
	start := time.Now()

	// translated comment
	if len(profiles.MappedTLSClients) == 0 {
		return nil, fmt.Errorf("no TLS client profiles available")
	}

	// translated comment
	names := make([]string, 0, len(profiles.MappedTLSClients))
	for name := range profiles.MappedTLSClients {
		names = append(names, name)
	}

	// translated comment
	randomName := utils.RandomChoiceString(names)
	profile := profiles.MappedTLSClients[randomName]
	if profile.GetClientHelloStr() == "" {
		return nil, fmt.Errorf("profile %s is invalid (empty ClientHelloStr)", randomName)
	}

	// translated comment
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

	// translated comment
	browserTypeStr, _ := inferBrowserFromProfileName(randomName)
	isMobile := isMobileProfile(randomName)
	headers := headers.GenerateHeaders(types.BrowserType(browserTypeStr), ua, isMobile)

	// translated comment
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

// translated comment
// translated comment
func GetRandomFingerprintByBrowser(browserType string) (*types.FingerprintResult, error) {
	return GetRandomFingerprintByBrowserWithOS(browserType, types.OperatingSystem(""))
}

// translated comment
func GetRandomFingerprintByBrowserWithOS(browserType string, os types.OperatingSystem) (*types.FingerprintResult, error) {
	if browserType == "" {
		return nil, fmt.Errorf("browser type cannot be empty")
	}
	if len(profiles.MappedTLSClients) == 0 {
		return nil, fmt.Errorf("no TLS client profiles available")
	}

	browserType = strings.ToLower(browserType)

	// translated comment
	initBrowserIndex()

	// translated comment
	candidates, exists := browserIndex[browserType]
	if !exists || len(candidates) == 0 {
		return nil, &ErrBrowserNotFound{Browser: browserType}
	}

	// translated comment
	randomName := utils.RandomChoiceString(candidates)
	profile := profiles.MappedTLSClients[randomName]
	if profile.GetClientHelloStr() == "" {
		return nil, fmt.Errorf("profile %s is invalid (empty ClientHelloStr)", randomName)
	}

	// translated comment
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

	// translated comment
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

// translated comment
type ErrBrowserNotFound struct {
	Browser string
}

func (e *ErrBrowserNotFound) Error() string {
	return "browser type not found: " + e.Browser
}

// translated comment
func isMobileProfile(profileName string) bool {
	name := strings.ToLower(profileName)
	return strings.Contains(name, "ios") ||
		strings.Contains(name, "android") ||
		strings.Contains(name, "ipad") ||
		strings.Contains(name, "mobile")
}

// translated comment
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
