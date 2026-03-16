// Command export_features — exports browser profile features as JSON for PyTorch training.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
)

type ExportedSample struct {
	ID          string    `json:"id"`
	BrowserType string    `json:"browser_type"`
	FamilyLabel int       `json:"family_label"`
	Features    []float64 `json:"features"`
}

type ExportedData struct {
	Samples          []ExportedSample `json:"samples"`
	FeatureDim       int              `json:"feature_dim"`
	EmbeddingDim     int              `json:"embedding_dim"`
	CrossLayerDim    int              `json:"cross_layer_dim"`
	BehaviorDim      int              `json:"behavior_dim"`
	NumBrowserFamily int              `json:"num_browser_family"`
	NumForgeryTypes  int              `json:"num_forgery_types"`
	NumThreatClasses int              `json:"num_threat_classes"`
	NumActions       int              `json:"num_actions"`
}

func main() {
	allProfiles := profiles.GetAll()
	if len(allProfiles) == 0 {
		fmt.Fprintln(os.Stderr, "no profiles")
		os.Exit(1)
	}

	familyLabels := map[string]int{
		"chrome": 0, "firefox": 1, "safari": 2, "edge": 3,
		"opera": 4, "brave": 5, "samsung": 6,
	}

	var samples []ExportedSample
	for _, p := range allProfiles {
		features := ml.EncodeFingerprint(&p)
		label, ok := familyLabels[string(p.BrowserType)]
		if !ok {
			label = 0
		}
		samples = append(samples, ExportedSample{
			ID:          p.ID,
			BrowserType: string(p.BrowserType),
			FamilyLabel: label,
			Features:    features,
		})
	}

	data := ExportedData{
		Samples:          samples,
		FeatureDim:       ml.FingerprintFeatureDim,
		EmbeddingDim:     ml.EmbeddingDim,
		CrossLayerDim:    ml.CrossLayerFeatureDim,
		BehaviorDim:      ml.BehaviorFeatureDim,
		NumBrowserFamily: ml.NumBrowserFamilies,
		NumForgeryTypes:  ml.NumForgeryTypes,
		NumThreatClasses: ml.NumThreatClasses,
		NumActions:       ml.NumActions,
	}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal error: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("training/profile_features.json", out, 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "write error: %v\n", err)
		os.Exit(1)
	}

	// Print summary
	familyCount := make(map[string]int)
	for _, s := range samples {
		familyCount[s.BrowserType]++
	}
	fmt.Printf("Exported %d profiles to training/profile_features.json\n", len(samples))
	for family, count := range familyCount {
		fmt.Printf("  %-12s %d\n", family, count)
	}
}
