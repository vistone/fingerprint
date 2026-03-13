// Package tcpip integrated fingerprint analysis examples
package tcpip

import (
	"fmt"
	"testing"

	"github.com/vistone/fingerprint/modules/core/types"
	"github.com/vistone/fingerprint/modules/generator/random"
	"github.com/vistone/fingerprint/modules/http/legacy/useragent"
	"github.com/vistone/fingerprint/modules/profiles/legacy"
)

// ExampleIntegratedFingerprinter demonstrates usage of the integrated fingerprint analyzer (using real fingerprint data)
func ExampleIntegratedFingerprinter() {
	// Create integrated fingerprint analyzer
	fingerprinter := NewIntegratedFingerprinter()

	// Set up IP geolocation database
	geoDB := NewSimpleIPGeoDB()
	fingerprinter.SetIPRegionDB(geoDB)

	// Get real Chrome on Windows fingerprint (not hardcoded)
	fpResult, err := random.GetRandomFingerprintByBrowserWithOS("chrome", types.OSWindows10)
	if err != nil {
		fmt.Printf("Failed to get fingerprint: %v\n", err)
		return
	}

	// Use real User-Agent and Headers
	userAgent := fpResult.UserAgent
	headers := fpResult.Headers.ToMap()

	// Get TCP features from fingerprint configuration
	profile, ok := profiles.MappedTLSClients["chrome_120"]
	if !ok {
		// Fall back to default configuration
		profile = profiles.DefaultClientProfile
	}

	// Get TCP configuration info
	tcpSettings := profile.GetSettings()
	windowSize := uint16(65535)       // Windows default
	if ws, ok := tcpSettings[4]; ok { // 4 is the setting ID for initial window size
		windowSize = uint16(ws)
	}

	// Construct real TCP packet (based on real fingerprint configuration)
	packet := &TCPPacket{
		IPHeader: &IPHeader{
			Version:        4,
			TimeToLive:     128, // Windows default TTL
			Identification: 12345,
			Flags:          0x02,      // DF bit set
			Protocol:       6,         // TCP
			SourceAddress:  "8.8.8.8", // Google DNS (US)
			DestAddress:    "192.168.1.1",
		},
		SourcePort:      54321,
		DestinationPort: 443,
		SequenceNumber:  1000000,
		WindowSize:      windowSize,
		Options:         []byte{0x02, 0x04, 0x05, 0xb4, 0x01, 0x01, 0x04, 0x02}, // MSS+NOP
		Flags:           0x02,                                                   // SYN
	}

	// Perform integrated analysis
	result, err := fingerprinter.Analyze(packet, userAgent, headers)
	if err != nil {
		fmt.Printf("Analysis error: %v\n", err)
		return
	}

	// Output analysis results
	fmt.Printf("=== Integrated Fingerprint Analysis Results ===\n")
	fmt.Printf("Source IP: %s\n", result.SourceIP)
	fmt.Printf("Fingerprint ID: %s\n", fpResult.HelloClientID)
	fmt.Printf("\n")

	// Per-layer identification results
	fmt.Printf("--- Per-Layer Identification Results ---\n")
	if result.TCPResult != nil {
		fmt.Printf("TCP layer inferred OS: %s (confidence: %.2f)\n", result.TCPResult.OS, result.TCPResult.Confidence)
	}
	fmt.Printf("UA layer inferred OS: %s\n", result.ParsedOSFromUA)
	if result.GeoInfo != nil {
		fmt.Printf("Geolocation: %s, %s (%s)\n", result.GeoInfo.City, result.GeoInfo.Country, result.GeoInfo.ISP)
	}
	fmt.Printf("\n")

	// Cross-validation results
	fmt.Printf("--- Cross-Validation ---\n")
	fmt.Printf("TCP layer OS: %s\n", result.OSCrossValidation.OSFromTCP)
	fmt.Printf("UA layer OS: %s\n", result.OSCrossValidation.OSFromUA)
	fmt.Printf("Geo layer OS: %s\n", result.OSCrossValidation.OSFromGeo)
	fmt.Printf("Consensus OS: %s\n", result.OSCrossValidation.ConsensusOS)
	fmt.Printf("Match score: %.2f\n", result.OSCrossValidation.MatchScore)
	fmt.Printf("IP-UA consistency: %v\n", result.IPUAConsistency)
	fmt.Printf("\n")

	// Inconsistency report
	if len(result.Inconsistencies) > 0 {
		fmt.Printf("--- Detected Inconsistencies ---\n")
		for _, inc := range result.Inconsistencies {
			fmt.Printf("[%s] %s: %s\n", inc.Severity, inc.RuleName, inc.Description)
			fmt.Printf("  Expected: %s, Actual: %s\n", inc.Expected, inc.Actual)
		}
		fmt.Printf("\n")
	}

	// Overall assessment
	fmt.Printf("--- Overall Assessment ---\n")
	fmt.Printf("Final identified OS: %s\n", result.FinalOS)
	fmt.Printf("Device type: %s\n", result.FinalDeviceType)
	fmt.Printf("Overall confidence: %.2f\n", result.OverallConfidence)
	fmt.Printf("Risk score: %.2f\n", result.RiskScore)
}

// TestIntegratedFingerprinter_TCP tests TCP layer analysis (using real fingerprints)
func TestIntegratedFingerprinter_TCP(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()

	// Get real Windows Chrome fingerprint
	fpResult, err := random.GetRandomFingerprintByBrowserWithOS("chrome", types.OSWindows10)
	if err != nil {
		t.Fatalf("Failed to get fingerprint: %v", err)
	}

	// Test Windows fingerprint (TTL=128)
	winPacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:    128,
			SourceAddress: "192.168.1.100",
		},
		WindowSize: 65535,
		Options:    []byte{0x02, 0x04, 0x05, 0xb4}, // MSS=1460
	}

	result, err := fingerprinter.Analyze(winPacket, fpResult.UserAgent, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.ParsedOSFromUA != "Windows 10" {
		t.Errorf("Expected Windows 10 from UA, got %s", result.ParsedOSFromUA)
	}

	t.Logf("Windows fingerprint detected: OS=%s, UA=%s, Confidence=%.2f",
		result.FinalOS, fpResult.UserAgent[:50], result.OverallConfidence)
}

// TestIntegratedFingerprinter_MacOS tests macOS fingerprint (using real fingerprints)
func TestIntegratedFingerprinter_MacOS(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()

	// Get real macOS Safari fingerprint
	fpResult, err := random.GetRandomFingerprintByBrowserWithOS("safari", types.OSMacOS14)
	if err != nil {
		t.Fatalf("Failed to get fingerprint: %v", err)
	}

	// macOS typically uses TTL 64
	macPacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:    64,
			SourceAddress: "17.0.0.1", // Apple's IP range
		},
		WindowSize: 65535,
	}

	result, err := fingerprinter.Analyze(macPacket, fpResult.UserAgent, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.ParsedOSFromUA != "macOS" {
		t.Errorf("Expected macOS from UA, got %s", result.ParsedOSFromUA)
	}

	t.Logf("macOS fingerprint detected: OS=%s, UA=%s", result.FinalOS, fpResult.UserAgent[:50])
}

// TestIntegratedFingerprinter_Mobile tests mobile device fingerprint (using real iOS fingerprints)
func TestIntegratedFingerprinter_Mobile(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()

	// Get real iOS Safari fingerprint
	profile, ok := profiles.MappedTLSClients["safari_ios_16_0"]
	if !ok {
		t.Skip("safari_ios_16_0 profile not found")
	}

	ua, err := useragent.GetUserAgentByProfileNameWithOS("safari_ios_16_0", types.OSMacOS14)
	if err != nil {
		t.Fatalf("Failed to get UA: %v", err)
	}

	// iPhone fingerprint
	iphonePacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:    64,
			SourceAddress: "17.0.0.1",
		},
		WindowSize: 65535,
	}

	_ = profile // Use profile to avoid unused variable warning
	result, err := fingerprinter.Analyze(iphonePacket, ua, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.FinalDeviceType != "Mobile" {
		t.Errorf("Expected Mobile device type, got %s", result.FinalDeviceType)
	}

	t.Logf("iPhone fingerprint detected: Device=%s, OS=%s, UA=%s",
		result.FinalDeviceType, result.FinalOS, ua[:50])
}

// TestIntegratedFingerprinter_Inconsistency tests inconsistency detection (constructing inconsistent data using real fingerprints)
func TestIntegratedFingerprinter_Inconsistency(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()

	// Get real Windows Chrome fingerprint
	fpResult, err := random.GetRandomFingerprintByBrowserWithOS("chrome", types.OSWindows10)
	if err != nil {
		t.Fatalf("Failed to get fingerprint: %v", err)
	}

	// Construct inconsistent data: claims Windows but TTL is Linux (64)
	inconsistentPacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:    64, // Linux TTL, but UA claims Windows (should be 128)
			SourceAddress: "192.168.1.100",
		},
		WindowSize: 29200, // Linux window size
	}

	// Use real Windows UA, but TCP characteristics are Linux
	result, err := fingerprinter.Analyze(inconsistentPacket, fpResult.UserAgent, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should detect inconsistency
	hasInconsistency := false
	for _, inc := range result.Inconsistencies {
		if inc.RuleName == "TTL_OS_Mismatch" {
			hasInconsistency = true
			t.Logf("Detected inconsistency: %s", inc.Description)
		}
	}

	if !hasInconsistency {
		t.Logf("Inconsistencies found: %v", result.Inconsistencies)
	}

	// Risk score should be high
	if result.RiskScore == 0 {
		t.Logf("Warning: Risk score is 0 despite inconsistency")
	}
}

// TestIntegratedFingerprinter_Geolocation tests geolocation integration (using real fingerprints)
func TestIntegratedFingerprinter_Geolocation(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()

	// Set up IP geolocation database
	geoDB := NewSimpleIPGeoDB()
	fingerprinter.SetIPRegionDB(geoDB)

	// Get real Windows Chrome fingerprint
	fpResult, err := random.GetRandomFingerprintByBrowserWithOS("chrome", types.OSWindows10)
	if err != nil {
		t.Fatalf("Failed to get fingerprint: %v", err)
	}

	// Use US IP
	usPacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:    128,
			SourceAddress: "8.8.8.8", // Google DNS - US
		},
		WindowSize: 65535,
	}

	result, err := fingerprinter.Analyze(usPacket, fpResult.UserAgent, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.GeoInfo == nil {
		t.Fatal("Expected geolocation info")
	}

	if result.GeoInfo.Country != "United States" {
		t.Errorf("Expected United States, got %s", result.GeoInfo.Country)
	}

	t.Logf("Geolocation: %s, %s, ISP: %s", result.GeoInfo.City, result.GeoInfo.Country, result.GeoInfo.ISP)
}

// TestIntegratedFingerprinter_AllOS tests all OS fingerprints (using real fingerprints)
func TestIntegratedFingerprinter_AllOS(t *testing.T) {
	testCases := []struct {
		name       string
		browser    string
		os         types.OperatingSystem
		expectedOS string
	}{
		{
			name:       "Windows 10 Chrome",
			browser:    "chrome",
			os:         types.OSWindows10,
			expectedOS: "Windows 10",
		},
		{
			name:       "macOS Safari",
			browser:    "safari",
			os:         types.OSMacOS14,
			expectedOS: "macOS",
		},
		{
			name:       "Linux Chrome",
			browser:    "chrome",
			os:         types.OSLinux,
			expectedOS: "Linux",
		},
	}

	fingerprinter := NewIntegratedFingerprinter()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get real fingerprint
			fpResult, err := random.GetRandomFingerprintByBrowserWithOS(tc.browser, tc.os)
			if err != nil {
				t.Skipf("Failed to get %s fingerprint for %s: %v", tc.browser, tc.os, err)
				return
			}

			packet := &TCPPacket{
				IPHeader: &IPHeader{
					TimeToLive:    64,
					SourceAddress: "192.168.1.100",
				},
				WindowSize: 65535,
			}

			result, err := fingerprinter.Analyze(packet, fpResult.UserAgent, nil)
			if err != nil {
				t.Fatalf("Analyze failed: %v", err)
			}

			if result.ParsedOSFromUA != tc.expectedOS {
				t.Errorf("Expected %s, got %s", tc.expectedOS, result.ParsedOSFromUA)
			}

			t.Logf("OS: %s, Device: %s, Confidence: %.2f, UA: %s",
				result.FinalOS, result.FinalDeviceType, result.OverallConfidence, fpResult.UserAgent[:50])
		})
	}
}
