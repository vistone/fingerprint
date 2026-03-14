package ml

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// =========================================================================
// Trainer — train all models from browser profiles
// =========================================================================

// NeuralTrainerConfig holds neural network training configuration.
type NeuralTrainerConfig struct {
	Epochs          int     // number of training epochs
	BatchSize       int     // mini-batch size
	LearningRate    float64 // Adam learning rate
	AugmentNoise    float64 // data augmentation noise stddev
	TripletMargin   float64 // triplet loss margin
	ForgeryRatio    float64 // ratio of forged to real samples
	ValidationSplit float64 // validation set ratio
}

// DefaultNeuralTrainerConfig is the default training configuration.
var DefaultNeuralTrainerConfig = &NeuralTrainerConfig{
	Epochs:          100,
	BatchSize:       32,
	LearningRate:    0.001,
	AugmentNoise:    0.05,
	TripletMargin:   1.0,
	ForgeryRatio:    1.5,
	ValidationSplit: 0.2,
}

// TrainingMetrics holds training metrics.
type TrainingMetrics struct {
	Epoch       int
	EncoderLoss float64
	ClassLoss   float64
	ForgeryLoss float64
	ThreatLoss  float64
	ValAccuracy float64
	ForgeryAUC  float64
}

// NeuralTrainer is the neural network training pipeline.
type NeuralTrainer struct {
	Pipeline *ModelPipeline
	Config   *NeuralTrainerConfig
	Metrics  []TrainingMetrics
}

// NewNeuralTrainer creates a new neural network training pipeline.
func NewNeuralTrainer(pipeline *ModelPipeline, config *NeuralTrainerConfig) *NeuralTrainer {
	if config == nil {
		config = DefaultNeuralTrainerConfig
	}
	return &NeuralTrainer{
		Pipeline: pipeline,
		Config:   config,
	}
}

// TrainFromProfiles trains all models from browser profiles.
// Main training entry: load 207 profiles → data augmentation → multi-phase training.
func (t *NeuralTrainer) TrainFromProfiles(registry *profiles.ProfileRegistry) error {
	allProfiles := registry.GetAll()
	if len(allProfiles) == 0 {
		return fmt.Errorf("no profiles available for training")
	}

	// Build training data
	trainSet, valSet := t.buildTrainingData(allProfiles)

	// Phase 1: Encoder pre-training (triplet loss)
	if err := t.trainEncoder(trainSet); err != nil {
		return fmt.Errorf("encoder training failed: %w", err)
	}

	// Phase 2: Browser classifier training
	if err := t.trainClassifier(trainSet, valSet); err != nil {
		return fmt.Errorf("classifier training failed: %w", err)
	}

	// Phase 3: Forgery detector training
	if err := t.trainForgeryDetector(trainSet); err != nil {
		return fmt.Errorf("forgery detector training failed: %w", err)
	}

	// Phase 4: Threat assessor training
	if err := t.trainThreatAssessor(trainSet); err != nil {
		return fmt.Errorf("threat assessor training failed: %w", err)
	}

	t.Pipeline.mu.Lock()
	t.Pipeline.trained = true
	t.Pipeline.mu.Unlock()

	return nil
}

// profileSample is an internal training sample.
type profileSample struct {
	Features    []float64
	FamilyLabel int
	ProfileID   string
	BrowserType core.BrowserType
}

// buildTrainingData converts profiles into training samples and splits into train/val sets.
func (t *NeuralTrainer) buildTrainingData(allProfiles []profiles.ClientProfile) (train, val []profileSample) {
	// Group by browser family
	familyMap := make(map[core.BrowserType][]profiles.ClientProfile)
	for _, p := range allProfiles {
		familyMap[p.BrowserType] = append(familyMap[p.BrowserType], p)
	}

	familyLabels := browserFamilyLabelMap()

	var samples []profileSample
	for _, p := range allProfiles {
		features := EncodeFingerprint(&p)
		label := familyLabels[p.BrowserType]
		samples = append(samples, profileSample{
			Features:    features,
			FamilyLabel: label,
			ProfileID:   p.ID,
			BrowserType: p.BrowserType,
		})
		// Data augmentation: generate variants with varying noise levels
		augNoises := []float64{
			t.Config.AugmentNoise * 0.5,
			t.Config.AugmentNoise * 0.75,
			t.Config.AugmentNoise,
			t.Config.AugmentNoise,
			t.Config.AugmentNoise * 1.25,
			t.Config.AugmentNoise * 1.5,
			t.Config.AugmentNoise * 2.0,
			t.Config.AugmentNoise * 2.5,
		}
		for _, noise := range augNoises {
			augFeatures := make([]float64, len(features))
			copy(augFeatures, features)
			for i := range augFeatures {
				augFeatures[i] += rand.NormFloat64() * noise
				augFeatures[i] = math.Max(0, math.Min(1, augFeatures[i]))
			}
			samples = append(samples, profileSample{
				Features:    augFeatures,
				FamilyLabel: label,
				ProfileID:   p.ID,
				BrowserType: p.BrowserType,
			})
		}
	}

	// Shuffle and split
	rand.Shuffle(len(samples), func(i, j int) { samples[i], samples[j] = samples[j], samples[i] })
	splitIdx := int(float64(len(samples)) * (1.0 - t.Config.ValidationSplit))
	return samples[:splitIdx], samples[splitIdx:]
}

// browserFamilyLabelMap returns the mapping from browser type to label index.
func browserFamilyLabelMap() map[core.BrowserType]int {
	return map[core.BrowserType]int{
		core.BrowserChrome:  0,
		core.BrowserFirefox: 1,
		core.BrowserSafari:  2,
		core.BrowserEdge:    3,
		core.BrowserOpera:   4,
		core.BrowserBrave:   5,
		core.BrowserSamsung: 6,
	}
}

// =========================================================================
// Phase 1: Encoder training — triplet loss
// =========================================================================

func (t *NeuralTrainer) trainEncoder(samples []profileSample) error {
	cfg := t.Config
	enc := t.Pipeline.encoder
	enc.Net.SetTraining(true)
	defer enc.Net.SetTraining(false)

	params := enc.Net.Params()
	optimizer := NewAdamOptimizer(params, cfg.LearningRate)
	scheduler := NewWarmupCosineAnnealingLR(cfg.LearningRate, cfg.LearningRate*0.01, 5)
	tripletMargin := cfg.TripletMargin

	// Group indices by family
	familyIdx := make(map[int][]int)
	for i, s := range samples {
		familyIdx[s.FamilyLabel] = append(familyIdx[s.FamilyLabel], i)
	}
	families := make([]int, 0, len(familyIdx))
	for f := range familyIdx {
		families = append(families, f)
	}
	sort.Ints(families)

	for epoch := 0; epoch < cfg.Epochs; epoch++ {
		scheduler.StepLR(optimizer, epoch, cfg.Epochs)
		totalLoss := 0.0
		count := 0

		// Generate triplets with semi-hard negative mining
		for _, fam := range families {
			indices := familyIdx[fam]
			if len(indices) < 2 {
				continue
			}

			batchSize := min(cfg.BatchSize, len(indices)/2)
			if batchSize < 1 {
				batchSize = 1
			}

			// Pre-compute embeddings for hard negative selection
			var candidateNegs []int
			for _, otherFam := range families {
				if otherFam != fam {
					candidateNegs = append(candidateNegs, familyIdx[otherFam]...)
				}
			}
			if len(candidateNegs) == 0 {
				continue
			}

			anchors := make([]float64, 0, batchSize*FingerprintFeatureDim)
			positives := make([]float64, 0, batchSize*FingerprintFeatureDim)
			negatives := make([]float64, 0, batchSize*FingerprintFeatureDim)

			for b := 0; b < batchSize; b++ {
				aIdx := indices[rand.Intn(len(indices))]
				pIdx := indices[rand.Intn(len(indices))]

				// Semi-hard negative mining: pick closest negative
				anchorEmb := enc.EncodeSingle(samples[aIdx].Features)
				bestNegIdx := candidateNegs[rand.Intn(len(candidateNegs))]
				bestNegDist := math.MaxFloat64

				// Sample a subset of candidates for efficiency
				numCandidates := min(16, len(candidateNegs))
				for c := 0; c < numCandidates; c++ {
					cIdx := candidateNegs[rand.Intn(len(candidateNegs))]
					negEmb := enc.EncodeSingle(samples[cIdx].Features)
					dist := 0.0
					for d := 0; d < len(anchorEmb); d++ {
						diff := anchorEmb[d] - negEmb[d]
						dist += diff * diff
					}
					if dist < bestNegDist {
						bestNegDist = dist
						bestNegIdx = cIdx
					}
				}

				anchors = append(anchors, samples[aIdx].Features...)
				positives = append(positives, samples[pIdx].Features...)
				negatives = append(negatives, samples[bestNegIdx].Features...)
			}

			anchorT := NewTensor([]int{batchSize, FingerprintFeatureDim}, anchors)
			posT := NewTensor([]int{batchSize, FingerprintFeatureDim}, positives)
			negT := NewTensor([]int{batchSize, FingerprintFeatureDim}, negatives)

			enc.Net.ZeroGrad()
			anchorEmb := enc.Encode(anchorT)
			posEmb := enc.Encode(posT)
			negEmb := enc.Encode(negT)

			loss, anchorGrad, _, _ := TripletMarginLoss(anchorEmb, posEmb, negEmb, tripletMargin)

			enc.Net.Backward(anchorGrad)
			ClipGradNorm(params, 5.0)
			optimizer.Step()

			totalLoss += loss
			count++
		}

		if count > 0 {
			t.recordMetric(epoch, totalLoss/float64(count), 0, 0, 0, 0, 0)
		}
	}
	return nil
}

// =========================================================================
// Phase 2: Classifier training — cross-entropy
// =========================================================================

func (t *NeuralTrainer) trainClassifier(trainSet, valSet []profileSample) error {
	cfg := t.Config
	enc := t.Pipeline.encoder
	cls := t.Pipeline.classifier
	cls.Net.SetTraining(true)
	defer cls.Net.SetTraining(false)

	params := cls.Net.Params()
	optimizer := NewAdamOptimizer(params, cfg.LearningRate)
	scheduler := NewWarmupCosineAnnealingLR(cfg.LearningRate, cfg.LearningRate*0.01, 5)

	for epoch := 0; epoch < cfg.Epochs; epoch++ {
		scheduler.StepLR(optimizer, epoch, cfg.Epochs)
		totalLoss := 0.0
		count := 0

		// Mini-batch training
		for start := 0; start < len(trainSet); start += cfg.BatchSize {
			end := start + cfg.BatchSize
			if end > len(trainSet) {
				end = len(trainSet)
			}
			batchSamples := trainSet[start:end]
			batchN := len(batchSamples)

			// Encode → embedding (encoder is frozen)
			fpData := make([]float64, 0, batchN*FingerprintFeatureDim)
			targets := make([]int, batchN)
			for i, s := range batchSamples {
				fpData = append(fpData, s.Features...)
				targets[i] = s.FamilyLabel
			}

			embeddings := enc.Encode(NewTensor([]int{batchN, FingerprintFeatureDim}, fpData))

			// Forward pass
			cls.Net.ZeroGrad()
			output := cls.Net.Forward(embeddings)

			// Compute loss
			lossVal, grad := CrossEntropyLoss(output, targets)

			// Backpropagation
			cls.Net.Backward(grad)
			ClipGradNorm(params, 5.0)
			optimizer.Step()

			totalLoss += lossVal
			count++
		}

		// Validation accuracy
		valAcc := t.evaluateClassifier(enc, cls, valSet)

		if count > 0 {
			t.recordMetric(epoch, 0, totalLoss/float64(count), 0, 0, valAcc, 0)
		}
	}
	return nil
}

func (t *NeuralTrainer) evaluateClassifier(enc *FingerprintEncoder, cls *BrowserClassifier, valSet []profileSample) float64 {
	if len(valSet) == 0 {
		return 0
	}
	correct := 0
	familyLabels := browserFamilyLabelMap()
	for _, s := range valSet {
		emb := enc.EncodeSingle(s.Features)
		pred := cls.ClassifySingle(emb)
		if familyLabels[pred.Family] == s.FamilyLabel {
			correct++
		}
	}
	return float64(correct) / float64(len(valSet))
}

// =========================================================================
// Phase 3: Forgery detector training
// =========================================================================
