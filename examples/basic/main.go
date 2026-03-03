// basic - Basic example of using the fingerprint library
package main

import (
	"fmt"
	"log"

	"github.com/vistone/fingerprint/profiles"
	"github.com/vistone/fingerprint/tls/ja3"
)

func main() {
	// Example: Get Chrome 120 profile
	profile := profiles.Chrome_120

	// Get ClientHello spec
	spec, err := profile.GetClientHelloSpec()
	if err != nil {
		log.Fatalf("Failed to get ClientHello spec: %v", err)
	}

	fmt.Printf("Profile: Chrome 120\n")
	fmt.Printf("TLS Version: %d\n", spec.TLSVersMax)
	fmt.Printf("Cipher Suites: %d\n", len(spec.CipherSuites))
	fmt.Printf("Extensions: %d\n", len(spec.Extensions))

	// Example: Compute JA3 from spec
	result, err := ja3.ComputeJA3FromSpec(spec)
	if err != nil {
		log.Printf("Failed to compute JA3: %v", err)
	} else {
		fmt.Printf("\nJA3 Fingerprint:\n")
		fmt.Printf("  Hash: %s\n", result.Hash)
		fmt.Printf("  Raw: %s\n", result.RawString)
	}

	// Example: Get HTTP/2 settings
	settings := profile.GetSettings()
	fmt.Printf("\nHTTP/2 Settings: %d\n", len(settings))
}
