// simple - Simplest example of fingerprint library usage
package main

import (
	"fmt"
	"github.com/vistone/fingerprint/modules/profiles"
)

func main() {
	// Simplest usage: get a profile by ID
	profile, ok := profiles.Get("chrome_133")
	if !ok {
		fmt.Println("Profile not found")
		return
	}

	fmt.Printf("Profile: %s\n", profile.Name)
	fmt.Printf("Browser: %s %s\n", profile.BrowserType, profile.BrowserVersion)
	fmt.Printf("OS: %s %s\n", profile.OS, profile.OSVersion)
	
	// Get headers
	if profile.Headers != nil {
		fmt.Printf("\nAccept: %s\n", profile.Headers.Accept)
		fmt.Printf("Accept-Language: %s\n", profile.Headers.AcceptLanguage)
	}
}
