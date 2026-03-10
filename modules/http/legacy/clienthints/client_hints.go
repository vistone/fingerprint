package clienthints

// Phase 3: This module has completed basic migration, awaiting deep optimization (see docs/5-process/modularization/PHASE_3_PLAN.md)
import (
	"fmt"
	"strings"

	"github.com/vistone/fingerprint/modules/profiles/legacy"
	"github.com/vistone/fingerprint/modules/core/types"
)

// ClientHintsPolicy Client Hints policy configuration
type ClientHintsPolicy struct {
	// Low entropy hints (sent by default)
	SendLowEntropyHints bool

	// High entropy hints (require Accept-CH request from server)
	HighEntropyHints []string

	// Cross-origin delegation support
	SupportsCrossOriginDelegation bool

	// Permissions-Policy configuration
	PermissionsPolicy map[string]string
}

// ClientHintsData complete Client Hints data
type ClientHintsData struct {
	// Low entropy hints (always available)
	SecCHUA         string // "Google Chrome";v="120", "Chromium";v="120"
	SecCHUAMobile   string // ?0 or ?1
	SecCHUAPlatform string // "Windows", "macOS", "Linux", "Android", "iOS"

	// High entropy hints (require Accept-CH authorization)
	SecCHUAArch               string // "x86", "arm"
	SecCHUABitness            string // "64", "32"
	SecCHUAFullVersionList    string // Full version list
	SecCHUAModel              string // Device model (mobile)
	SecCHUAPlatformVersion    string // Platform version number
	SecCHUAWoW64              string // Windows 64-bit emulation marker
	SecCHUAFormFactor         string // "Desktop", "Mobile", "Tablet", "VR"
	SecCHPreferredColorScheme string // "dark", "light"
	SecCHPrefersReducedMotion string // "reduce", "no-preference"

	// Viewport and network hints
	ViewportWidth string // Viewport width
	DeviceMemory  string // Device memory (GB)
	DPR           string // Device pixel ratio
	DownlinkSpeed string // Downlink bandwidth (Mbps)
	ECT           string // Effective connection type: "slow-2g", "2g", "3g", "4g"
	RTT           string // Round-trip time (ms)
	SaveData      string // "on"/"off"
}

// NewClientHintsPolicy creates default policy
func NewClientHintsPolicy(browserType types.BrowserType) *ClientHintsPolicy {
	policy := &ClientHintsPolicy{
		SendLowEntropyHints:           true,
		SupportsCrossOriginDelegation: false,
		PermissionsPolicy:             make(map[string]string),
	}

	// Configure high entropy hints based on browser type
	switch browserType {
	case types.BrowserChrome:
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

	case types.BrowserEdge:
		policy.HighEntropyHints = []string{
			"Sec-CH-UA-Arch",
			"Sec-CH-UA-Bitness",
			"Sec-CH-UA-Full-Version-List",
			"Sec-CH-UA-Platform-Version",
			"Sec-CH-UA-WoW64",
		}
		policy.SupportsCrossOriginDelegation = true
		policy.PermissionsPolicy["ch-ua"] = "self"

	case types.BrowserFirefox, types.BrowserSafari:
		// Firefox and Safari currently do not support Client Hints
		policy.HighEntropyHints = []string{}
	}

	return policy
}

// GenerateClientHintsFromProfile generates Client Hints from profile
func GenerateClientHintsFromProfile(profile *profiles.ClientProfile, policy *ClientHintsPolicy) *ClientHintsData {
	hints := &ClientHintsData{}

	if !policy.SendLowEntropyHints {
		return hints
	}

	// Low entropy hints (always sent)
	hints.SecCHUA = generateSecCHUA(profile)
	hints.SecCHUAMobile = generateSecCHUAMobile(profile)
	hints.SecCHUAPlatform = generateSecCHUAPlatform(profile)

	// High entropy hints (only when policy allows)
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

// ProcessAcceptCH processes the server's Accept-CH response header
func (p *ClientHintsPolicy) ProcessAcceptCH(acceptCHValue string) {
	if acceptCHValue == "" {
		return
	}

	// Parse Accept-CH header
	hints := strings.Split(acceptCHValue, ",")
	for _, hint := range hints {
		hint = strings.TrimSpace(hint)
		if hint == "" {
			continue
		}

		// Only add high entropy hints that we support
		if isSupportedHighEntropyHint(hint) && !contains(p.HighEntropyHints, hint) {
			p.HighEntropyHints = append(p.HighEntropyHints, hint)
		}
	}
}

// ApplyToHeaders applies Client Hints to HTTP headers
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

// ============ Helper functions ============

func generateSecCHUA(profile *profiles.ClientProfile) string {
	// Format: "Brand";v="major", "Brand";v="major"
	// Chrome example: "Not A(Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"

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

func generateSecCHUAMobile(profile *profiles.ClientProfile) string {
	if profile.IsMobile {
		return "?1"
	}
	return "?0"
}

func generateSecCHUAPlatform(profile *profiles.ClientProfile) string {
	// Return standard platform name based on operating system
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

func generateFullVersionList(profile *profiles.ClientProfile) string {
	// Format: "Brand";v="full.version", "Brand";v="full.version"
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

// contains checks if a string slice contains a specific value
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if strings.EqualFold(v, item) {
			return true
		}
	}
	return false
}
