// random - Example of using random fingerprint generation
package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Example 1: Get random profile from all available
	fmt.Println("=== Example 1: Random Profile ===")
	allProfiles := profiles.GetAll()
	if len(allProfiles) > 0 {
		selected := allProfiles[rand.Intn(len(allProfiles))]
		fmt.Printf("Randomly selected: %s\n", selected.Name)
		fmt.Printf("Browser: %s %s\n", selected.BrowserType, selected.BrowserVersion)
		fmt.Printf("OS: %s %s\n", selected.OS, selected.OSVersion)
	}

	// Example 2: Get random profile by browser type
	fmt.Println("\n=== Example 2: Random Chrome Profile ===")
	chromeProfiles := profiles.GetByBrowser(core.BrowserChrome)
	if len(chromeProfiles) > 0 {
		selected := chromeProfiles[rand.Intn(len(chromeProfiles))]
		fmt.Printf("Random Chrome: %s\n", selected.Name)
		fmt.Printf("Cipher Suites: %d\n", len(selected.CipherSuites))
		fmt.Printf("Extensions: %d\n", len(selected.Extensions))
	}

	// Example 3: Get random profile by specific IDs
	fmt.Println("\n=== Example 3: Random from Specific List ===")
	profileIDs := []string{"chrome_133", "firefox_135", "safari_17_0", "edge_130"}
	selectedID := profileIDs[rand.Intn(len(profileIDs))]
	if profile, ok := profiles.Get(selectedID); ok {
		fmt.Printf("Selected from list: %s\n", profile.Name)
	}

	// Example 4: Statistics
	fmt.Println("\n=== Example 4: Profile Statistics ===")
	fmt.Printf("Total available profiles: %d\n", len(allProfiles))
	fmt.Printf("Chrome profiles: %d\n", len(profiles.GetByBrowser(core.BrowserChrome)))
	fmt.Printf("Firefox profiles: %d\n", len(profiles.GetByBrowser(core.BrowserFirefox)))
	fmt.Printf("Safari profiles: %d\n", len(profiles.GetByBrowser(core.BrowserSafari)))

	fmt.Println("\n=== Random Examples Complete ===")
}
