// Command validate — loads PyTorch-trained weights and validates Go inference.
//
// Usage:
//
//	go run ./cmd/validate/ [-weights path/to/weights.json]
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
)

type browserStats struct {
	total     int
	correct   int // browser prediction matches actual family
	forgeries int // detected as forgery
}

type embEntry struct {
	browser   core.BrowserType
	embedding []float64
}

var errNoProfilesFound = errors.New("no profiles found")

func main() {
	weightsPath := flag.String("weights", "./models/weights.json", "path to weights.json")
	flag.Parse()

	if err := runValidation(*weightsPath); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}
}

func runValidation(weightsPath string) error {
	pipeline, err := loadPipeline(weightsPath)
	if err != nil {
		return err
	}

	allProfiles := profiles.GetAll()
	if len(allProfiles) == 0 {
		return errNoProfilesFound
	}
	fmt.Printf("Loaded %d browser profiles\n\n", len(allProfiles))

	byBrowser, embeddings := runProfileInference(pipeline, allProfiles)
	printClassificationSummary(byBrowser)
	printEmbeddingQuality(embeddings)
	fmt.Println("\nValidation complete.")

	return nil
}

func loadPipeline(weightsPath string) (*ml.ModelPipeline, error) {
	pipeline := ml.NewModelPipeline()

	info, err := os.Stat(weightsPath)
	if err != nil {
		return nil, fmt.Errorf("weights file not found: %w", err)
	}
	fmt.Printf("Loading weights: %s (%.1f KB)\n", weightsPath, float64(info.Size())/1024)

	if err := pipeline.LoadWeights(weightsPath); err != nil {
		return nil, fmt.Errorf("failed to load weights: %w", err)
	}
	fmt.Printf("Weights loaded, pipeline.Trained() = %v\n\n", pipeline.Trained())

	return pipeline, nil
}

func runProfileInference(pipeline *ml.ModelPipeline, allProfiles []profiles.ClientProfile) (map[core.BrowserType]*browserStats, []embEntry) {
	byBrowser := make(map[core.BrowserType]*browserStats)
	embeddings := make([]embEntry, 0, len(allProfiles))

	fmt.Println("=== Per-Profile Inference Results (first 20) ===")
	fmt.Printf("%-40s %-12s %-12s %6s  %-10s %6s  %-10s\n",
		"Profile", "Actual", "Predicted", "Conf%", "Forgery", "Prob%", "Threat")
	fmt.Println(repeat("-", 110))

	for i, p := range allProfiles {
		result := pipeline.Infer(&p, nil)

		actual := p.BrowserType
		predicted := result.Browser.Family
		conf := result.Browser.Confidence

		s, ok := byBrowser[actual]
		if !ok {
			s = &browserStats{}
			byBrowser[actual] = s
		}
		s.total++
		if predicted == actual {
			s.correct++
		}
		if result.Forgery.IsForgery {
			s.forgeries++
		}

		if i < 20 {
			forgeryStr := "genuine"
			if result.Forgery.IsForgery {
				forgeryStr = result.Forgery.ForgeryType.String()
			}
			fmt.Printf("%-40s %-12s %-12s %5.1f%%  %-10s %5.1f%%  %-10s\n",
				truncate(p.ID, 40),
				actual,
				predicted,
				conf*100,
				forgeryStr,
				result.Forgery.ForgeryProb*100,
				result.Threat.ThreatClass.String(),
			)
		}

		embeddings = append(embeddings, embEntry{
			browser:   p.BrowserType,
			embedding: result.Embedding,
		})
	}

	return byBrowser, embeddings
}

func printClassificationSummary(byBrowser map[core.BrowserType]*browserStats) {
	fmt.Println()
	fmt.Println("=== Browser Classification Accuracy ===")
	fmt.Printf("%-15s %6s %6s %7s  %7s\n", "Browser", "Total", "Correct", "Acc%", "Forgery%")
	fmt.Println(repeat("-", 50))

	totalCorrect, totalAll := 0, 0
	for browser, s := range byBrowser {
		acc := 0.0
		if s.total > 0 {
			acc = float64(s.correct) / float64(s.total) * 100
		}
		forgeryPct := 0.0
		if s.total > 0 {
			forgeryPct = float64(s.forgeries) / float64(s.total) * 100
		}
		fmt.Printf("%-15s %6d %6d %6.1f%%  %6.1f%%\n",
			browser, s.total, s.correct, acc, forgeryPct)
		totalCorrect += s.correct
		totalAll += s.total
	}

	overallAcc := 0.0
	if totalAll > 0 {
		overallAcc = float64(totalCorrect) / float64(totalAll) * 100
	}
	fmt.Println(repeat("-", 50))
	fmt.Printf("%-15s %6d %6d %6.1f%%\n", "OVERALL", totalAll, totalCorrect, overallAcc)
}

func printEmbeddingQuality(embeddings []embEntry) {
	fmt.Println()
	fmt.Println("=== Embedding Quality ===")

	sameFamilySim := 0.0
	sameFamilyN := 0
	crossFamilySim := 0.0
	crossFamilyN := 0

	step := maxInt(1, len(embeddings)/50) // Limit pairs for speed.
	for i := 0; i < len(embeddings); i += step {
		for j := i + 1; j < len(embeddings); j += step {
			sim := cosineSim(embeddings[i].embedding, embeddings[j].embedding)
			if embeddings[i].browser == embeddings[j].browser {
				sameFamilySim += sim
				sameFamilyN++
			} else {
				crossFamilySim += sim
				crossFamilyN++
			}
		}
	}

	if sameFamilyN > 0 {
		fmt.Printf("Same-family avg cosine similarity:  %.4f (%d pairs)\n", sameFamilySim/float64(sameFamilyN), sameFamilyN)
	}
	if crossFamilyN > 0 {
		fmt.Printf("Cross-family avg cosine similarity: %.4f (%d pairs)\n", crossFamilySim/float64(crossFamilyN), crossFamilyN)
	}
	if sameFamilyN > 0 && crossFamilyN > 0 {
		gap := sameFamilySim/float64(sameFamilyN) - crossFamilySim/float64(crossFamilyN)
		fmt.Printf("Separation gap:                     %.4f\n", gap)
	}
}

func cosineSim(a, b []float64) float64 {
	dot, normA, normB := 0.0, 0.0, 0.0
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom < 1e-12 {
		return 0
	}
	return dot / denom
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func repeat(s string, n int) string {
	return strings.Repeat(s, n)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
