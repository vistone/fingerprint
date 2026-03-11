// Verify that PyTorch-exported weights load correctly in Go pipeline.
package main

import (
	"fmt"
	"os"

	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
)

func main() {
	// Make sure profiles are registered.
	_ = profiles.AllProfiles()

	pipeline := ml.NewModelPipeline(nil)

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
	allProfiles := profiles.AllProfiles()
	testCases := []string{}
	families := map[string]bool{}
	for id, p := range allProfiles {
		fam := p.BrowserType
		if !families[fam] && len(testCases) < 7 {
			families[fam] = true
			testCases = append(testCases, id)
		}
	}

	for _, id := range testCases {
		p := allProfiles[id]

		result, err := pipeline.Analyze(p)
		if err != nil {
			fmt.Printf("  %s (%s): ERROR %v\n", id, p.BrowserType, err)
			continue
		}

		fmt.Printf("  %-35s (%s)  predicted=%-8s conf=%.1f%%  forgery=%.1f%%  threat=%s\n",
			id, p.BrowserType,
			result.Classification.PredictedFamily,
			result.Classification.Confidence*100,
			result.ForgeryAnalysis.ForgeryProbability*100,
			result.ThreatAssessment.ThreatLevel,
		)
	}

	fmt.Println("\nAll OK.")
}
