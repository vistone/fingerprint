package random

// Phase 3: basic migration completed; deep optimization remains.
import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/core/types"
	"github.com/vistone/fingerprint/modules/http/legacy/headers"
	"github.com/vistone/fingerprint/modules/http/legacy/useragent"
	"github.com/vistone/fingerprint/modules/internal/metrics"
	"github.com/vistone/fingerprint/modules/kit"
	"github.com/vistone/fingerprint/modules/profiles/legacy"
)

// browserIndex stores precomputed browser-to-profile mappings.
var (
	browserIndex     map[string][]string
	browserIndexOnce sync.Once
)

// initBrowserIndex lazily initializes the browser index.
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

// GetRandomFingerprint selects a full fingerprint from all available profiles.
//
// The function picks a profile from profiles.MappedTLSClients and generates
// the matching User-Agent and HTTP headers. OS is selected randomly.
//
// Returns:
//   - *types.FingerprintResult: complete profile, User-Agent, and headers
//   - error: returned when profile set is empty or profile is invalid
//
// Example:
//
//	result, err := GetRandomFingerprint()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	println(result.UserAgent)
//	println(result.ClientProfile.GetClientHelloStr())
//
// Thread safety: yes.
// Performance: avg ~7.4us and ~1.8KB allocations.
func GetRandomFingerprint() (*types.FingerprintResult, error) {
	return GetRandomFingerprintWithOS(types.OperatingSystem(""))
}

// GetRandomFingerprintWithOS selects a random fingerprint with optional OS.
//
// Similar to GetRandomFingerprint, but allows a target OS.
// If os is empty, OS is chosen randomly.
//
// Args:
//   - os: target OS, e.g. "Windows NT 10.0; Win64; x64"; empty means random
//
// Returns:
//   - *types.FingerprintResult: full fingerprint result
//   - error: returned for invalid OS or empty profile set
//
// Example:
//
//	// Use explicit Windows OS
//	result, err := GetRandomFingerprintWithOS("Windows NT 10.0; Win64; x64")
//	// Choose OS randomly
//	result, err := GetRandomFingerprintWithOS("")
//
// Thread safety: yes.
func GetRandomFingerprintWithOS(os types.OperatingSystem) (*types.FingerprintResult, error) {
	start := time.Now()

	// Ensure profile map is not empty.
	if len(profiles.MappedTLSClients) == 0 {
		return nil, fmt.Errorf("no TLS client profiles available")
	}

	// Collect all profile names.
	names := make([]string, 0, len(profiles.MappedTLSClients))
	for name := range profiles.MappedTLSClients {
		names = append(names, name)
	}

	// Randomly select one profile.
	randomName := utils.RandomChoiceString(names)
	profile := profiles.MappedTLSClients[randomName]
	if profile.GetClientHelloStr() == "" {
		return nil, fmt.Errorf("profile %s is invalid (empty ClientHelloStr)", randomName)
	}

	// Resolve matching User-Agent.
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

	// Generate standard HTTP headers.
	browserTypeStr, _ := inferBrowserFromProfileName(randomName)
	isMobile := isMobileProfile(randomName)
	headers := headers.GenerateHeaders(types.BrowserType(browserTypeStr), ua, isMobile)

	// Record generation metrics.
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

// GetRandomFingerprintByBrowser selects a random fingerprint by browser type.
// browserType: "chrome", "firefox", "safari", "opera", etc.
func GetRandomFingerprintByBrowser(browserType string) (*types.FingerprintResult, error) {
	return GetRandomFingerprintByBrowserWithOS(browserType, types.OperatingSystem(""))
}

// GetRandomFingerprintByBrowserWithOS selects by browser type with optional OS.
func GetRandomFingerprintByBrowserWithOS(browserType string, os types.OperatingSystem) (*types.FingerprintResult, error) {
	if browserType == "" {
		return nil, fmt.Errorf("browser type cannot be empty")
	}
	if len(profiles.MappedTLSClients) == 0 {
		return nil, fmt.Errorf("no TLS client profiles available")
	}

	browserType = strings.ToLower(browserType)

	// Use lazily initialized precomputed index.
	initBrowserIndex()

	// Resolve candidate list from index.
	candidates, exists := browserIndex[browserType]
	if !exists || len(candidates) == 0 {
		return nil, &ErrBrowserNotFound{Browser: browserType}
	}

	// Randomly select a candidate profile.
	randomName := utils.RandomChoiceString(candidates)
	profile := profiles.MappedTLSClients[randomName]
	if profile.GetClientHelloStr() == "" {
		return nil, fmt.Errorf("profile %s is invalid (empty ClientHelloStr)", randomName)
	}

	// Resolve matching User-Agent.
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

	// Generate standard HTTP headers.
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

// ErrBrowserNotFound indicates unknown browser type.
type ErrBrowserNotFound struct {
	Browser string
}

func (e *ErrBrowserNotFound) Error() string {
	return "browser type not found: " + e.Browser
}

// isMobileProfile reports whether a profile targets mobile devices.
func isMobileProfile(profileName string) bool {
	name := strings.ToLower(profileName)
	return strings.Contains(name, "ios") ||
		strings.Contains(name, "android") ||
		strings.Contains(name, "ipad") ||
		strings.Contains(name, "mobile")
}

// inferBrowserFromProfileName infers browser family/version from profile name.
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
