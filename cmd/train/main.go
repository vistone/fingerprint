// Command train — trains the ML model library from browser profile data.
//
// Usage:
//
//	go run ./cmd/train/ [flags]
//
// Flags:
//
//	-output    Model output directory (default: ./models)
//	-epochs    Training epochs (default: 100)
//	-batch     Batch size (default: 32)
//	-lr        Learning rate (default: 0.001)
//	-verbose   Print training progress (default: true)
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
)

func main() {
	outputDir := flag.String("output", "./models", "model output directory")
	epochs := flag.Int("epochs", 100, "training epochs")
	batchSize := flag.Int("batch", 32, "batch size")
	lr := flag.Float64("lr", 0.001, "learning rate")
	verbose := flag.Bool("verbose", true, "print training progress")
	flag.Parse()

	if *verbose {
		fmt.Println("=== Fingerprint ML Model Training ===")
		fmt.Println()
	}

	// 1. Load profiles from default registry
	allProfiles := profiles.GetAll()
	if len(allProfiles) == 0 {
		log.Fatal("no browser profiles found — ensure profiles package is imported")
	}

	// Count by family
	familyCount := make(map[string]int)
	for _, p := range allProfiles {
		familyCount[string(p.BrowserType)]++
	}
	if *verbose {
		fmt.Printf("Loaded %d browser profiles:\n", len(allProfiles))
		for family, count := range familyCount {
			fmt.Printf("  %-12s %d profiles\n", family, count)
		}
		fmt.Println()
	}

	// 2. Create training pipeline
	pipeline := ml.NewModelPipeline()

	config := &ml.NeuralTrainerConfig{
		Epochs:          *epochs,
		BatchSize:       *batchSize,
		LearningRate:    *lr,
		AugmentNoise:    0.05,
		TripletMargin:   1.0,
		ForgeryRatio:    1.5,
		ValidationSplit: 0.2,
	}

	trainer := ml.NewNeuralTrainer(pipeline, config)

	if *verbose {
		fmt.Printf("Training config:\n")
		fmt.Printf("  Epochs:       %d\n", config.Epochs)
		fmt.Printf("  Batch size:   %d\n", config.BatchSize)
		fmt.Printf("  Learning rate: %.4f\n", config.LearningRate)
		fmt.Printf("  Augment noise: %.3f\n", config.AugmentNoise)
		fmt.Printf("  Triplet margin: %.1f\n", config.TripletMargin)
		fmt.Printf("  Forgery ratio: %.1f\n", config.ForgeryRatio)
		fmt.Printf("  Val split:    %.1f%%\n", config.ValidationSplit*100)
		fmt.Println()
	}

	// 3. Train
	startTime := time.Now()
	if *verbose {
		fmt.Println("Starting 4-phase training...")
		fmt.Println("  Phase 1: Encoder pre-training (triplet loss + hard negative mining)")
		fmt.Println("  Phase 2: Browser classifier (cross-entropy)")
		fmt.Println("  Phase 3: Forgery detector (binary cross-entropy)")
		fmt.Println("  Phase 4: Threat assessor (cross-entropy + synthetic behavior)")
		fmt.Println()
	}

	registry := profiles.DefaultRegistry
	if err := trainer.TrainFromProfiles(registry); err != nil {
		log.Fatalf("Training failed: %v", err)
	}

	trainDuration := time.Since(startTime)
	if *verbose {
		fmt.Printf("Training completed in %s\n\n", trainDuration.Round(time.Millisecond))
	}

	// 4. Print metrics summary
	if *verbose && len(trainer.Metrics) > 0 {
		fmt.Println("Training metrics (last epoch for each phase):")
		last := trainer.Metrics[len(trainer.Metrics)-1]
		fmt.Printf("  Encoder loss:    %.4f\n", last.EncoderLoss)
		fmt.Printf("  Classifier loss: %.4f\n", last.ClassLoss)
		fmt.Printf("  Forgery loss:    %.4f\n", last.ForgeryLoss)
		fmt.Printf("  Threat loss:     %.4f\n", last.ThreatLoss)
		fmt.Printf("  Val accuracy:    %.2f%%\n", last.ValAccuracy*100)
		fmt.Println()
	}

	// 5. Save to model store
	storeConfig := ml.DefaultStoreConfig(*outputDir)
	store, err := ml.NewModelStore(storeConfig)
	if err != nil {
		log.Fatalf("Failed to create model store: %v", err)
	}

	var lastMetrics *ml.TrainingMetrics
	if len(trainer.Metrics) > 0 {
		m := trainer.Metrics[len(trainer.Metrics)-1]
		lastMetrics = &m
	}

	if err := store.Save(pipeline, fmt.Sprintf("full training: %d profiles, %d epochs", len(allProfiles), config.Epochs), lastMetrics); err != nil {
		log.Fatalf("Failed to save model: %v", err)
	}

	ver := store.Latest()
	if *verbose && ver != nil {
		fmt.Printf("Model saved to %s (version %d)\n", *outputDir, ver.Version)
	}

	// 6. Also save weights as standalone file
	weightsPath := *outputDir + "/weights.json"
	if err := pipeline.SaveWeights(weightsPath); err != nil {
		log.Fatalf("Failed to save weights: %v", err)
	}
	if *verbose {
		info, _ := os.Stat(weightsPath)
		if info != nil {
			fmt.Printf("Weights file: %s (%.1f KB)\n", weightsPath, float64(info.Size())/1024)
		}
	}

	// 7. Validation: test inference on a sample profile
	if *verbose {
		fmt.Println("\n=== Validation ===")
		testProfile := allProfiles[0]
		result := pipeline.Infer(&testProfile, nil)
		fmt.Printf("Test profile: %s (%s %s)\n", testProfile.ID, testProfile.BrowserType, testProfile.BrowserVersion)
		fmt.Printf("  Browser:    %s (confidence: %.1f%%)\n", result.Browser.Family, result.Browser.Confidence*100)
		fmt.Printf("  Forgery:    %s (prob: %.1f%%)\n", result.Forgery.ForgeryType.String(), result.Forgery.ForgeryProb*100)
		fmt.Printf("  Threat:     %s (prob: %.1f%%)\n", result.Threat.ThreatClass.String(), result.Threat.ThreatProb*100)
		fmt.Printf("  Action:     %s (confidence: %.1f%%)\n", result.Threat.Action.String(), result.Threat.ActionConfidence*100)
	}

	fmt.Println("\nDone.")
}
