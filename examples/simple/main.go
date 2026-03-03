// simple - Simplest example of fingerprint library usage
package main

import (
	"fmt"
	"github.com/vistone/fingerprint/profiles"
)

func main() {
	// Simplest usage: just get a profile
	profile := profiles.Chrome_120
	
	// Get HTTP/2 settings
	settings := profile.GetSettings()
	fmt.Printf("HTTP/2 Settings (%d total):\n", len(settings))
	
	for id, value := range settings {
		fmt.Printf("  Setting %d: %d\n", id, value)
	}
	
	// Get pseudo header order
	order := profile.GetPseudoHeaderOrder()
	fmt.Printf("\nPseudo Header Order:\n")
	for i, header := range order {
		fmt.Printf("  %d. %s\n", i+1, header)
	}
}
