// Command validate — loads PyTorch-trained weights and validates Go inference.
//
// Usage:
//
//	go run ./cmd/validate/ [-weights path/to/weights.json]
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
)

func main() {
	weightsPath := flag.String("weights", "./models/weights.json", "path to weights.json")
	flag.Parse()

	// 1. Load the model pipeline
	pipeline := ml.NewModelPipeline()

	info, err := os.Stat(*weightsPath)
	if err != nil {
		log.Fatalf("Weights file not found: %v", err)
	}
	fmt.Printf("Loading weights: %s (%.1f KB)\n", *weightsPath, float64(info.Size())/1024)

	if err := pipeline.LoadWeights(*weightsPath); err != nil {
		log.Fatalf("Failed to load weights: %v", err)
	}
	fmt.Printf("Weights loaded, pipeline.Trained() = %v\n\n", pipeline.Trained())

	// 2. Load test profiles
	allProfiles := profiles.GetAll()
	if len(allProfiles) == 0 {
		log.Fatal("No profiles found")
	}
	fmt.Printf("Loaded %d browser profiles\n\n", len(allProfiles))

	// 3. Run inference on all profiles and collect statistics
	type stats struct {
		total     int
		correct   int // browser prediction matches actual family
		forgeries int // detected as forgery
	}
	byBrowser := make(map[core.BrowserType]*stats)

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
			s = &stats{}
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
	}

	// 4. Summary statistics
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

	// 5. Embedding quality check: same-family similarity vs cross-family
	fmt.Println()
	fmt.Println("=== Embedding Quality ===")
	type embEntry struct {
		browser   core.BrowserType
		embedding []float64
	}
	embeddings := make([]embEntry, 0, len(allProfiles))
	for _, p := range allProfiles {
		result := pipeline.Infer(&p, nil)
		embeddings = append(embeddings, embEntry{
			browser:   p.BrowserType,
			embedding: result.Embedding,
		})
	}

	// Sample some same-family and cross-family pairs
	sameFamilySim := 0.0
	sameFamilyN := 0
	crossFamilySim := 0.0
	crossFamilyN := 0

	step := max(1, len(embeddings)/50) // limit pairs for speed
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

	fmt.Println("\nValidation complete.")
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
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
