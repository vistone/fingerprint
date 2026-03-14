package waf

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
)

// DeviceEngine analyzes device fingerprints
type DeviceEngine struct {
	fingerprints map[string]*DeviceFingerprint
}

// DeviceFingerprint stores device characteristics
type DeviceFingerprint struct {
	ID               string
	JA3              string
	JA4              string
	UserAgent        string
	AcceptLanguage   string
	ScreenResolution string
	Timezone         string
	Fonts            []string
	WebGLRenderer    string
	SuspiciousScore  float64
}

// DeviceResult contains device analysis results
type DeviceResult struct {
	Score           float64
	Factors         []core.RiskFactor
	DeviceID        string
	IsKnownBad      bool
	Inconsistencies []string
}

// NewDeviceEngine creates a new device analysis engine
func NewDeviceEngine() *DeviceEngine {
	return &DeviceEngine{
		fingerprints: make(map[string]*DeviceFingerprint),
	}
}

// Analyze performs device fingerprinting analysis
func (e *DeviceEngine) Analyze(req *http.Request) *DeviceResult {
	result := &DeviceResult{
		Score:           0,
		Factors:         make([]core.RiskFactor, 0),
		Inconsistencies: make([]string, 0),
	}

	// Generate device ID from stable characteristics
	deviceID := e.generateDeviceID(req)
	result.DeviceID = deviceID

	// Check for consistency between layers
	ja3 := req.Header.Get("X-JA3-Fingerprint")
	ua := req.UserAgent()

	// Check TLS fingerprint vs User-Agent consistency
	if ja3 != "" && ua != "" {
		if !e.isConsistent(ja3, ua) {
			result.Score += 0.4
			result.Inconsistencies = append(result.Inconsistencies, "ja3_ua_mismatch")
			result.Factors = append(result.Factors, core.RiskFactor{
				Name:        "fingerprint_inconsistency",
				Weight:      0.4,
				Description: "JA3 fingerprint does not match User-Agent",
			})
		}
	}

	// Check for headless browser indicators
	if e.isHeadlessBrowser(req) {
		result.Score += 0.6
		result.Factors = append(result.Factors, core.RiskFactor{
			Name:        "headless_browser",
			Weight:      0.6,
			Description: "Headless browser detected",
		})
	}

	// Check for virtualization/emulation
	if e.isVirtualized(req) {
		result.Score += 0.5
		result.Factors = append(result.Factors, core.RiskFactor{
			Name:        "virtualized_environment",
			Weight:      0.5,
			Description: "Virtualized or emulated environment detected",
		})
	}

	return result
}

func (e *DeviceEngine) generateDeviceID(req *http.Request) string {
	// Combine stable characteristics
	data := fmt.Sprintf("%s|%s|%s",
		req.Header.Get("X-JA3-Fingerprint"),
		req.Header.Get("X-JA4-Fingerprint"),
		req.UserAgent(),
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

func (e *DeviceEngine) isConsistent(ja3, ua string) bool {
	// Check if JA3 fingerprint matches expected for the User-Agent
	ua = strings.ToLower(ua)

	// Chrome typically has specific JA3 characteristics
	if strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg") {
		// Chrome JA3 should have specific extensions
		return true // Simplified - real implementation would check actual JA3
	}

	// Firefox has different characteristics
	if strings.Contains(ua, "firefox") {
		return true
	}

	// Safari
	if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		return true
	}

	return true // Default to consistent if unable to determine
}

func (e *DeviceEngine) isHeadlessBrowser(req *http.Request) bool {
	ua := strings.ToLower(req.UserAgent())

	// Check for headless indicators in User-Agent
	headlessIndicators := []string{
		"headlesschrome",
		"headlessfirefox",
		"phantomjs",
		"slimerjs",
	}

	for _, indicator := range headlessIndicators {
		if strings.Contains(ua, indicator) {
			return true
		}
	}

	// Check for headless-specific headers
	suspiciousHeaders := []string{
		"X-Headless-Chrome",
		"X-PhantomJS",
	}

	for _, header := range suspiciousHeaders {
		if req.Header.Get(header) != "" {
			return true
		}
	}

	return false
}

func (e *DeviceEngine) isVirtualized(req *http.Request) bool {
	// Check for indicators of virtualized environments
	// This is a simplified check

	// Check for common VM signatures in headers
	vmIndicators := []string{
		"VMware",
		"VirtualBox",
		"QEMU",
	}

	for _, indicator := range vmIndicators {
		if strings.Contains(req.UserAgent(), indicator) {
			return true
		}
	}

	return false
}
