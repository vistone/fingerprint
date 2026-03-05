// Basic example of using the fingerprint library
// Demonstrates core module usage
package main

import (
	"fmt"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
	"github.com/vistone/fingerprint/modules/tls"
)

func main() {
	// Example 1: Calculate JA3 fingerprint
	fmt.Println("=== Example 1: JA3 Calculation ===")
	spec := core.ClientHelloSpec{
		TLSVersion:      0x0303,
		CipherSuites:    []uint16{0x1301, 0x1302, 0x1303},
		Extensions:      []core.TLSExtension{{Type: 0}, {Type: 5}, {Type: 10}},
		SupportedCurves: []core.CurveID{23, 24, 25},
		SupportedPoints: []uint8{0},
	}

	ja3Result := tls.CalculateJA3(spec)
	fmt.Printf("JA3 Hash: %s\n\n", ja3Result.Hash)

	// Example 2: Get profiles by browser type
	fmt.Println("=== Example 2: Chrome Profiles ===")
	chromeProfiles := profiles.GetProfilesByBrowser(core.BrowserChrome)
	fmt.Printf("Found %d Chrome profiles\n", len(chromeProfiles))
	if len(chromeProfiles) > 0 {
		fmt.Printf("First Chrome Profile: %s\n\n", chromeProfiles[0].Name)
	}

	// Example 3: Get specific profile
	fmt.Println("=== Example 3: Get Specific Profile ===")
	if profile, ok := profiles.Get("chrome_133"); ok {
		fmt.Printf("Profile: %s\n", profile.Name)
		fmt.Printf("Browser: %s %s\n", profile.BrowserType, profile.BrowserVersion)
		fmt.Printf("OS: %s %s\n", profile.OS, profile.OSVersion)
	}

	fmt.Println("\nBasic examples completed!")
}
