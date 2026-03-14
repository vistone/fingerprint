package ml

import (
	"math"
	"math/rand"
)

func (t *NeuralTrainer) trainForgeryDetector(realSamples []profileSample) error {
	cfg := t.Config
	det := t.Pipeline.detector
	det.DetectorNet.SetTraining(true)
	det.TypeNet.SetTraining(true)
	defer det.DetectorNet.SetTraining(false)
	defer det.TypeNet.SetTraining(false)

	allParams := append(det.DetectorNet.Params(), det.TypeNet.Params()...)
	optimizer := NewAdamOptimizer(allParams, cfg.LearningRate)
	scheduler := NewWarmupCosineAnnealingLR(cfg.LearningRate, cfg.LearningRate*0.01, 5)

	inputDim := FingerprintFeatureDim + CrossLayerFeatureDim

	for epoch := 0; epoch < cfg.Epochs; epoch++ {
		scheduler.StepLR(optimizer, epoch, cfg.Epochs)
		totalLoss := 0.0
		count := 0

		for start := 0; start < len(realSamples); start += cfg.BatchSize {
			end := start + cfg.BatchSize
			if end > len(realSamples) {
				end = len(realSamples)
			}
			batchReal := realSamples[start:end]
			batchN := len(batchReal)

			// Compute number of forged samples
			numForgery := int(float64(batchN) * cfg.ForgeryRatio)
			if numForgery < 1 {
				numForgery = 1
			}
			totalBatch := batchN + numForgery

			inputData := make([]float64, 0, totalBatch*inputDim)
			targetData := make([]float64, totalBatch)

			// Real samples (label = 0, i.e. not forged)
			for _, s := range batchReal {
				cross := ComputeCrossLayerFeatures(s.Features)
				inputData = append(inputData, s.Features...)
				inputData = append(inputData, cross...)
			}

			// Forged samples (label = 1)
			for f := 0; f < numForgery; f++ {
				forged := t.generateForgedSample(realSamples)
				cross := ComputeCrossLayerFeatures(forged)
				inputData = append(inputData, forged...)
				inputData = append(inputData, cross...)
				targetData[batchN+f] = 1.0 // forged
			}

			input := NewTensor([]int{totalBatch, inputDim}, inputData)

			// Detection network forward pass
			det.DetectorNet.ZeroGrad()
			output := det.DetectorNet.Forward(input)

			// Binary cross-entropy loss
			lossVal, grad := BinaryCrossEntropyLoss(output, targetData)
			det.DetectorNet.Backward(grad)
			ClipGradNorm(allParams, 5.0)
			optimizer.Step()

			totalLoss += lossVal
			count++
		}

		if count > 0 {
			t.recordMetric(TrainingMetrics{Epoch: epoch, ForgeryLoss: totalLoss / float64(count)})
		}
	}
	return nil
}

// generateForgedSample generates a forged fingerprint sample.
// Multiple strategies to simulate different forgery types:
//   - Cross-browser layer mixing (anti-detect tool pattern)
//   - Headless browser simulation (missing JS features)
//   - Proxy/MITM pattern (TCP anomalies)
//   - Noise injection (generic tool fingerprint)
func (t *NeuralTrainer) generateForgedSample(samples []profileSample) []float64 {
	forged := make([]float64, FingerprintFeatureDim)

	strategy := rand.Intn(4)
	s1 := samples[rand.Intn(len(samples))]
	s2 := samples[rand.Intn(len(samples))]

	switch strategy {
	case 0: // Cross-browser layer mixing (anti-detect)
		// TLS from one browser, HTTP/2 from another
		copy(forged[0:8], s1.Features[0:8])
		copy(forged[8:14], s2.Features[8:14])
		if rand.Float64() < 0.5 {
			copy(forged[14:18], s1.Features[14:18])
		} else {
			copy(forged[14:18], s2.Features[14:18])
		}
		// JS features: partial or mismatched
		forged[18] = rand.Float64() * 0.3 // low canvas (tool fingerprint)
		forged[19] = rand.Float64() * 0.3 // low webgl
		forged[25] = rand.Float64()*0.3 + 0.4
		forged[26] = s1.Features[26]          // keep original UA
		forged[27] = s2.Features[27]          // but entropy from different browser
		forged[28] = rand.Float64()*0.3 + 0.3 // moderate tool marker

	case 1: // Headless browser (Puppeteer/Selenium)
		copy(forged, s1.Features)
		// Missing or anomalous JS features
		forged[18] = 0                        // no canvas
		forged[19] = 0                        // no webgl
		forged[20] = 0                        // no audio
		forged[21] = rand.Float64() * 0.05    // very few fonts
		forged[22] = rand.Float64() * 0.2     // low storage
		forged[23] = 0                        // no webrtc
		forged[24] = 0.25                     // generic 4 cores
		forged[25] = rand.Float64()*0.3 + 0.7 // high headless score
		forged[28] = rand.Float64()*0.2 + 0.5 // tool marker present

	case 2: // Proxy/MITM (TCP anomalies)
		copy(forged, s1.Features)
		// TCP layer inconsistencies
		forged[14] = 0.5 + rand.Float64()*0.5 // anomalous TTL
		forged[15] = rand.Float64() * 0.3     // unusual window
		forged[16] = rand.Float64() * 0.4     // unusual MSS
		forged[17] = 0                        // no timestamps (proxy stripped)
		forged[29] = rand.Float64()*0.2 + 0.3 // mild behavior anomaly

	case 3: // Noise injection (generic tool)
		// Start with random mix
		for i := 0; i < FingerprintFeatureDim; i++ {
			if rand.Float64() < 0.5 {
				forged[i] = s1.Features[i]
			} else {
				forged[i] = s2.Features[i]
			}
		}
		// Add significant noise
		for i := range forged {
			forged[i] += rand.NormFloat64() * 0.1
		}
		forged[28] = rand.Float64()*0.3 + 0.2 // some tool marker
	}

	// Final noise + clamp
	for i := range forged {
		forged[i] += rand.NormFloat64() * 0.02
		forged[i] = math.Max(0, math.Min(1, forged[i]))
	}
	return forged
}

// =========================================================================
// Phase 4: Threat assessor training
// =========================================================================

func (t *NeuralTrainer) trainThreatAssessor(samples []profileSample) error {
	cfg := t.Config
	enc := t.Pipeline.encoder
	det := t.Pipeline.detector
	assessor := t.Pipeline.assessor
	assessor.ThreatNet.SetTraining(true)
	assessor.ActionNet.SetTraining(true)
	defer assessor.ThreatNet.SetTraining(false)
	defer assessor.ActionNet.SetTraining(false)

	allParams := append(assessor.ThreatNet.Params(), assessor.ActionNet.Params()...)
	optimizer := NewAdamOptimizer(allParams, cfg.LearningRate)
	scheduler := NewWarmupCosineAnnealingLR(cfg.LearningRate, cfg.LearningRate*0.01, 5)

	inputDim := EmbeddingDim + 1 + NumForgeryTypes + BehaviorFeatureDim

	for epoch := 0; epoch < cfg.Epochs; epoch++ {
		scheduler.StepLR(optimizer, epoch, cfg.Epochs)
		totalLoss := 0.0
		count := 0

		for start := 0; start < len(samples); start += cfg.BatchSize {
			end := start + cfg.BatchSize
			if end > len(samples) {
				end = len(samples)
			}
			batchSamples := samples[start:end]
			batchN := len(batchSamples)

			inputData := make([]float64, batchN*inputDim)
			threatTargets := make([]int, batchN)

			for i, s := range batchSamples {
				emb := enc.EncodeSingle(s.Features)
				cross := ComputeCrossLayerFeatures(s.Features)
				forgery := det.DetectSingle(s.Features, cross)

				off := i * inputDim
				copy(inputData[off:], emb)
				off += EmbeddingDim
				inputData[off] = forgery.ForgeryProb
				off++
				copy(inputData[off:off+NumForgeryTypes], forgery.TypeProbs)
				off += NumForgeryTypes
				// Synthetic behavioral features to prevent zero-feature bias
				behavior := t.generateSyntheticBehavior(s, &forgery)
				copy(inputData[off:off+BehaviorFeatureDim], behavior)

				threatTargets[i] = t.generateThreatLabel(s, &forgery)
			}

			input := NewTensor([]int{batchN, inputDim}, inputData)

			assessor.ThreatNet.ZeroGrad()
			output := assessor.ThreatNet.Forward(input)
			lossVal, grad := CrossEntropyLoss(output, threatTargets)
			assessor.ThreatNet.Backward(grad)
			ClipGradNorm(allParams, 5.0)
			optimizer.Step()

			totalLoss += lossVal
			count++
		}

		if count > 0 {
			t.recordMetric(TrainingMetrics{Epoch: epoch, ThreatLoss: totalLoss / float64(count)})
		}
	}
	return nil
}

// generateThreatLabel generates a rule-based label from sample features and forgery detection results.
func (t *NeuralTrainer) generateThreatLabel(s profileSample, forgery *ForgeryResult) int {
	if forgery.ForgeryProb > 0.7 {
		switch forgery.ForgeryType {
		case ForgeryHeadless:
			return int(ThreatBot)
		case ForgeryAntiDetect:
			return int(ThreatFingerprintSpoof)
		case ForgeryProxy:
			return int(ThreatEvasion)
		}
		return int(ThreatFingerprintSpoof)
	}
	// Check for behavioral anomaly signals from features
	if s.Features[28] > 0.3 { // tool marker
		return int(ThreatBot)
	}
	if s.Features[29] > 0.3 { // behavior pattern anomaly
		return int(ThreatBehavioralAnomaly)
	}
	return int(ThreatNone)
}

// generateSyntheticBehavior generates synthetic behavioral features for training.
// This prevents the model from learning to ignore the behavior input dimensions.
func (t *NeuralTrainer) generateSyntheticBehavior(s profileSample, forgery *ForgeryResult) []float64 {
	behavior := make([]float64, BehaviorFeatureDim)

	if forgery.ForgeryProb > 0.5 {
		// Forged clients tend to have anomalous behavior
		behavior[0] = rand.Float64()*0.3 + 0.5 // high fingerprint switch rate
		behavior[1] = rand.Float64()*0.4 + 0.4 // high request rate
		behavior[2] = rand.Float64() * 0.4     // low consistency
		behavior[3] = rand.Float64()*0.3 + 0.5 // rising risk trend
		behavior[4] = rand.Float64() * 0.3     // few observations
		behavior[5] = rand.Float64()*0.3 + 0.5 // high unique FP ratio
		behavior[6] = rand.Float64() * 0.3     // short sessions
		behavior[7] = rand.Float64()*0.3 + 0.3 // burst indicator
	} else {
		// Normal clients have stable behavior
		behavior[0] = rand.Float64() * 0.2     // low switch rate
		behavior[1] = rand.Float64() * 0.3     // moderate request rate
		behavior[2] = rand.Float64()*0.3 + 0.6 // high consistency
		behavior[3] = rand.Float64() * 0.3     // low risk trend
		behavior[4] = rand.Float64()*0.3 + 0.4 // many observations
		behavior[5] = rand.Float64() * 0.3     // low unique FP ratio
		behavior[6] = rand.Float64()*0.3 + 0.5 // longer sessions
		behavior[7] = rand.Float64() * 0.2     // no burst
	}

	// Add noise
	for i := range behavior {
		behavior[i] += rand.NormFloat64() * 0.05
		behavior[i] = math.Max(0, math.Min(1, behavior[i]))
	}
	return behavior
}

func (t *NeuralTrainer) recordMetric(m TrainingMetrics) {
	// If same epoch already has a record, merge
	for i := range t.Metrics {
		if t.Metrics[i].Epoch == m.Epoch {
			if m.EncoderLoss > 0 {
				t.Metrics[i].EncoderLoss = m.EncoderLoss
			}
			if m.ClassLoss > 0 {
				t.Metrics[i].ClassLoss = m.ClassLoss
			}
			if m.ForgeryLoss > 0 {
				t.Metrics[i].ForgeryLoss = m.ForgeryLoss
			}
			if m.ThreatLoss > 0 {
				t.Metrics[i].ThreatLoss = m.ThreatLoss
			}
			if m.ValAccuracy > 0 {
				t.Metrics[i].ValAccuracy = m.ValAccuracy
			}
			if m.ForgeryAUC > 0 {
				t.Metrics[i].ForgeryAUC = m.ForgeryAUC
			}
			return
		}
	}
	t.Metrics = append(t.Metrics, m)
}
