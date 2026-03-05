// Advanced example demonstrating fingerprint analysis features
package main

import (
	"fmt"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/fingerprint"
	"github.com/vistone/fingerprint/modules/profiles"
)

func main() {
	// Example 1: Get and analyze fingerprint
	fmt.Println("=== Example 1: Fingerprint Analysis ===")
	profile, ok := profiles.Get("chrome_133")
	if !ok {
		fmt.Println("Profile not found")
		return
	}

	fmt.Printf("Profile: %s\n", profile.Name)
	fmt.Printf("Browser: %s %s\n", profile.BrowserType, profile.BrowserVersion)
	fmt.Printf("OS: %s %s\n", profile.OS, profile.OSVersion)
	fmt.Printf("TLS Version: 0x%04x\n", profile.TLSVersion)
	fmt.Printf("Cipher Suites: %d\n", len(profile.CipherSuites))
	fmt.Printf("Extensions: %d\n", len(profile.Extensions))

	// Example 2: Use Facade API
	fmt.Println("\n=== Example 2: Facade API ===")
	randomProfile := fingerprint.GetRandom()
	fmt.Printf("Random Profile: %s\n", randomProfile.Name)

	chromeProfile := fingerprint.GetByBrowser(fingerprint.BrowserChrome)
	if chromeProfile != nil {
		fmt.Printf("Chrome Profile: %s\n", chromeProfile.Name)
	}

	// Example 3: Feature extraction
	fmt.Println("\n=== Example 3: Feature Extraction ===")
	analyzer := fingerprint.NewAnalyzer()
	features := analyzer.ExtractFeatures(randomProfile)
	fmt.Printf("Extracted %d features\n", len(features.Features))

	// Example 4: Classification
	fmt.Println("\n=== Example 4: Classification ===")
	classification := analyzer.Classify(features)
	fmt.Printf("Protocol: %s (confidence: %.2f)\n", classification.Protocol, classification.ProtocolConfidence)
	fmt.Printf("Family: %s (confidence: %.2f)\n", classification.Family, classification.FamilyConfidence)

	// Example 5: Risk assessment
	fmt.Println("\n=== Example 5: Risk Assessment ===")
	risk := analyzer.EvaluateRisk(features, classification)
	fmt.Printf("Risk Score: %.2f\n", risk.Score)
	fmt.Printf("Risk Level: %s\n", risk.Level.String())

	// Example 6: List all browsers
	fmt.Println("\n=== Example 6: Browser Statistics ===")
	allProfiles := profiles.GetAll()
	browserCount := make(map[core.BrowserType]int)
	for _, p := range allProfiles {
		browserCount[p.BrowserType]++
	}
	for browser, count := range browserCount {
		fmt.Printf("  %s: %d profiles\n", browser, count)
	}
	fmt.Printf("\nTotal Profiles: %d\n", len(allProfiles))

	fmt.Println("\n=== Advanced Examples Complete ===")
}
