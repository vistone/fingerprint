// random - Example of using random fingerprint generation
package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/vistone/fingerprint/generator/random"
	"github.com/vistone/fingerprint/profiles"
)

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
	
	// Example: Get a random Chrome profile
	chromeProfiles := []profiles.ClientProfile{
		profiles.Chrome_120,
		profiles.Chrome_117,
		profiles.Chrome_110,
	}
	
	selected := chromeProfiles[rand.Intn(len(chromeProfiles))]
	spec, err := selected.GetClientHelloSpec()
	if err != nil {
		log.Fatalf("Failed to get spec: %v", err)
	}
	
	fmt.Printf("Selected Chrome profile with %d cipher suites\n", len(spec.CipherSuites))
	
	// Example: Get random fingerprint
	fp, err := random.GetRandomFingerprint()
	if err != nil {
		log.Fatalf("Failed to get random fingerprint: %v", err)
	}
	
	fmt.Printf("Generated fingerprint profile: %s\n", fp.HelloClientID)
	fmt.Printf("  User-Agent: %s\n", fp.UserAgent)
	
	if fp.Headers != nil {
		fmt.Printf("  Accept: %s\n", fp.Headers.Accept)
	}
}
