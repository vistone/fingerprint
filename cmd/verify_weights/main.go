// Verify that PyTorch-exported weights load correctly in Go pipeline.
package main

import (
	"fmt"
	"os"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
)

func main() {
	// Make sure profiles are registered.
	_ = profiles.GetAll()

	pipeline := ml.NewModelPipeline()

	weightsPath := "models/weights.json"
	if len(os.Args) > 1 {
		weightsPath = os.Args[1]
	}

	fmt.Printf("Loading weights from %s ...\n", weightsPath)
	if err := pipeline.LoadWeights(weightsPath); err != nil {
		fmt.Fprintf(os.Stderr, "FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Weights loaded successfully.")

	// Run inference on a few profiles.
	allProfiles := profiles.GetAll()
	families := map[core.BrowserType]bool{}
	var testProfiles []profiles.ClientProfile
	for _, p := range allProfiles {
		fam := p.BrowserType
		if !families[fam] && len(testProfiles) < 7 {
			families[fam] = true
			testProfiles = append(testProfiles, p)
		}
	}

	for _, p := range testProfiles {
		result := pipeline.Infer(&p, nil)

		fmt.Printf("  %-35s (%s)  predicted=%-8s conf=%.1f%%  forgery=%.1f%%  threat=%s\n",
			p.ID, p.BrowserType,
			result.Browser.Family,
			result.Browser.Confidence*100,
			result.Forgery.ForgeryProb*100,
			result.Threat.ThreatClass,
		)
	}

	fmt.Println("\nAll OK.")
}
