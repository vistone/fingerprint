package ml

// =========================================================================
// Model 3: ForgeryDetector — forgery detection model
// =========================================================================
//
// Learn: distinguishing features of real vs forged browser fingerprints.
//        Anti-detection tools (tls-client, curl-impersonate, Puppeteer) can
//        mimic single-layer features (e.g. TLS ClientHello), but cross-layer
//        consistency is hard to forge perfectly:
//        - Chrome TLS + Firefox HTTP/2 pseudo-header order → contradiction
//        - Windows UA + Linux TTL(64) → contradiction
//        - Missing Canvas/WebGL → headless browser
//        The forgery detector learns these complex cross-layer correlation patterns.
//
// Infer: input 30-dim fingerprint + 10-dim cross-layer features = 40-dim
//        → output forgery probability (0~1) + forgery type (4 classes)
//
// Output: ForgeryResult containing:
//         - IsForgery: true/false
//         - ForgeryProb: [0, 1] forgery probability
//         - ForgeryType: Real / Headless / AntiDetect / Proxy
//         - TypeProbs: per-type probability distribution
//
// Train: Binary CE (is forged) + CE (forgery type)
//        Data: real profiles + synthetic forged samples (layer mixing/noise/missing)

// ForgeryType represents the type of fingerprint forgery.
type ForgeryType int

const (
	ForgeryReal       ForgeryType = iota // real browser
	ForgeryHeadless                      // headless browser (Puppeteer/Selenium/PhantomJS)
	ForgeryAntiDetect                    // anti-detection tool (tls-client/curl-impersonate/GoLogin)
	ForgeryProxy                         // proxy/MITM (misconfigured features)
)

var forgeryTypeNames = [NumForgeryTypes]string{
	"real", "headless", "antidetect", "proxy",
}

// ForgeryTypeName returns the string name of the forgery type.
func (ft ForgeryType) String() string {
	if int(ft) < len(forgeryTypeNames) {
		return forgeryTypeNames[ft]
	}
	return "unknown"
}

// ForgeryDetector is the forgery detection model.
type ForgeryDetector struct {
	DetectorNet *Sequential // forgery probability network
	TypeNet     *Sequential // forgery type classification network
}

// NewForgeryDetector creates a forgery detector.
// Detector: Input(40) → Dense(128) → BN → ReLU → Dropout(0.2)
//
//	→ Dense(64) → BN → ReLU → Dense(32) → ReLU → Dense(1) → Sigmoid
//
// TypeNet:  Input(40) → Dense(128) → BN → ReLU → Dropout(0.2)
//
//	→ Dense(64) → BN → ReLU → Dense(32) → ReLU → Dense(4) → Softmax
func NewForgeryDetector() *ForgeryDetector {
	inputDim := FingerprintFeatureDim + CrossLayerFeatureDim // 30 + 10 = 40
	return &ForgeryDetector{
		DetectorNet: NewSequential(
			NewDenseLayer(inputDim, 128),
			NewBatchNormLayer(128),
			NewReLULayer(),
			NewDropoutLayer(0.2),
			NewDenseLayer(128, 64),
			NewBatchNormLayer(64),
			NewReLULayer(),
			NewDenseLayer(64, 32),
			NewReLULayer(),
			NewDenseLayer(32, 1),
			NewSigmoidLayer(),
		),
		TypeNet: NewSequential(
			NewDenseLayer(inputDim, 128),
			NewBatchNormLayer(128),
			NewReLULayer(),
			NewDropoutLayer(0.2),
			NewDenseLayer(128, 64),
			NewBatchNormLayer(64),
			NewReLULayer(),
			NewDenseLayer(64, 32),
			NewReLULayer(),
			NewDenseLayer(32, NumForgeryTypes),
			NewSoftmaxLayer(),
		),
	}
}

// ForgeryResult holds forgery detection results.
type ForgeryResult struct {
	IsForgery   bool        // whether forged (ForgeryProb > 0.5)
	ForgeryProb float64     // forgery probability [0,1]
	ForgeryType ForgeryType // predicted forgery type
	TypeProbs   []float64   // per-type probability distribution
	TypeNames   []string    // type names, aligned with TypeProbs
}

// Detect performs forgery detection.
// input: [batch × 40] (30-dim fingerprint + 10-dim cross-layer features)
func (fd *ForgeryDetector) Detect(input *Tensor) []ForgeryResult {
	probOut := fd.DetectorNet.Forward(input)
	typeOut := fd.TypeNet.Forward(input)
	batch := input.Shape[0]
	results := make([]ForgeryResult, batch)
	for i := 0; i < batch; i++ {
		prob := probOut.Data[i]
		typeRow := typeOut.Row(i)
		maxIdx := 0
		maxVal := typeRow[0]
		typeProbs := make([]float64, NumForgeryTypes)
		copy(typeProbs, typeRow)
		for j := 1; j < NumForgeryTypes; j++ {
			if typeRow[j] > maxVal {
				maxVal = typeRow[j]
				maxIdx = j
			}
		}
		names := make([]string, NumForgeryTypes)
		copy(names, forgeryTypeNames[:])
		results[i] = ForgeryResult{
			IsForgery:   prob > 0.5,
			ForgeryProb: prob,
			ForgeryType: ForgeryType(maxIdx),
			TypeProbs:   typeProbs,
			TypeNames:   names,
		}
	}
	return results
}

// DetectSingle detects a single sample (convenience method).
func (fd *ForgeryDetector) DetectSingle(fpFeatures, crossLayerFeatures []float64) ForgeryResult {
	combined := make([]float64, FingerprintFeatureDim+CrossLayerFeatureDim)
	copy(combined, fpFeatures)
	copy(combined[FingerprintFeatureDim:], crossLayerFeatures)
	input := FromSlice(combined)
	return fd.Detect(input)[0]
}

// =========================================================================
// Model 4: ThreatAssessor — threat assessment and action recommendation
// =========================================================================
//
// Learn: optimal security response after combining all model outputs and
//        behavioral features. This is the "decision brain" of the system,
//        integrating:
//        - Fingerprint embedding (32-dim) → client identity representation
//        - Forgery detection (5-dim) → authenticity assessment
//        - Behavioral features (8-dim) → temporal behavior patterns
//        Learning the optimal security action for different threat scenarios.
//
// Infer: input 45-dim composite vector → output threat class (6) + action (5)
//
// Output: ThreatPrediction containing:
//         - ThreatClass: None/Bot/FingerprintSpoof/SessionAnomaly/BehavioralAnomaly/Evasion
//         - ThreatProb: probability of threat presence
//         - Action: Allow/Monitor/Challenge/Throttle/Block
//         - ActionConfidence: action confidence
//         - ClassProbs: per-threat-class probabilities
//         - ActionProbs: per-action probabilities
//
// Train: CE (threat class) + CE (recommended action)
//        Initial: learn from rule engine labels (policy distillation)
//        Ongoing: fine-tune from ReportReward feedback (online adaptation)

// ThreatClass represents the threat classification.
type ThreatClass int

const (
	ThreatNone              ThreatClass = iota // no threat
	ThreatBot                                  // bot / crawler
	ThreatFingerprintSpoof                     // fingerprint forgery
	ThreatSessionAnomaly                       // session anomaly
	ThreatBehavioralAnomaly                    // behavioral anomaly
	ThreatEvasion                              // evasion behavior
)

var threatClassNames = [NumThreatClasses]string{
	"none", "bot", "fingerprint_spoof", "session_anomaly", "behavioral_anomaly", "evasion",
}

// String returns the string name of the threat class.
func (tc ThreatClass) String() string {
	if int(tc) < len(threatClassNames) {
		return threatClassNames[tc]
	}
	return "unknown"
}

// ActionClass represents a security action.
type ActionClass int

const (
	ActAllow     ActionClass = iota // allow
	ActMonitor                      // monitor
	ActChallenge                    // challenge verification
	ActThrottle                     // throttle
	ActBlock                        // block
)

var actionClassNames = [NumActions]string{
	"allow", "monitor", "challenge", "throttle", "block",
}

// String returns the string name of the action.
func (a ActionClass) String() string {
	if int(a) < len(actionClassNames) {
		return actionClassNames[a]
	}
	return "unknown"
}

// ThreatAssessor is the threat assessment model.
type ThreatAssessor struct {
	ThreatNet *Sequential // threat classification network
	ActionNet *Sequential // action recommendation network
}

// NewThreatAssessor creates a threat assessor.
// Input: embedding(32) + forgery output(5: prob+4 type probs) + behavior(8) = 45
// Threat: Input(45) → Dense(128) → BN → ReLU → Dropout(0.2)
//
//	→ Dense(64) → BN → ReLU → Dense(32) → ReLU → Dense(6) → Softmax
//
// Action: Input(45) → Dense(128) → BN → ReLU → Dropout(0.2)
//
//	→ Dense(64) → BN → ReLU → Dense(32) → ReLU → Dense(5) → Softmax
func NewThreatAssessor() *ThreatAssessor {
	inputDim := EmbeddingDim + 1 + NumForgeryTypes + BehaviorFeatureDim // 32+1+4+8 = 45
	return &ThreatAssessor{
		ThreatNet: NewSequential(
			NewDenseLayer(inputDim, 128),
			NewBatchNormLayer(128),
			NewReLULayer(),
			NewDropoutLayer(0.2),
			NewDenseLayer(128, 64),
			NewBatchNormLayer(64),
			NewReLULayer(),
			NewDenseLayer(64, 32),
			NewReLULayer(),
			NewDenseLayer(32, NumThreatClasses),
			NewSoftmaxLayer(),
		),
		ActionNet: NewSequential(
			NewDenseLayer(inputDim, 128),
			NewBatchNormLayer(128),
			NewReLULayer(),
			NewDropoutLayer(0.2),
			NewDenseLayer(128, 64),
			NewBatchNormLayer(64),
			NewReLULayer(),
			NewDenseLayer(64, 32),
			NewReLULayer(),
			NewDenseLayer(32, NumActions),
			NewSoftmaxLayer(),
		),
	}
}

// ThreatPrediction holds threat assessment results.
type ThreatPrediction struct {
	ThreatClass      ThreatClass // predicted threat class
	ThreatProb       float64     // probability of threat (1 - P(none))
	Action           ActionClass // recommended security action
	ActionConfidence float64     // action confidence
	ClassProbs       []float64   // per-threat-class probability distribution
	ActionProbs      []float64   // per-action probability distribution
}

// Assess evaluates threat level and recommends actions.
// input: [batch × 45] (embedding + forgery output + behavior features)
func (ta *ThreatAssessor) Assess(input *Tensor) []ThreatPrediction {
	threatOut := ta.ThreatNet.Forward(input)
	actionOut := ta.ActionNet.Forward(input)
	batch := input.Shape[0]
	results := make([]ThreatPrediction, batch)
	for i := 0; i < batch; i++ {
		tRow := threatOut.Row(i)
		aRow := actionOut.Row(i)
		// Find highest-probability threat class
		tMaxIdx, tMaxVal := 0, tRow[0]
		classProbs := make([]float64, NumThreatClasses)
		copy(classProbs, tRow)
		for j := 1; j < NumThreatClasses; j++ {
			if tRow[j] > tMaxVal {
				tMaxVal = tRow[j]
				tMaxIdx = j
			}
		}
		// Find highest-probability action
		aMaxIdx, aMaxVal := 0, aRow[0]
		actionProbs := make([]float64, NumActions)
		copy(actionProbs, aRow)
		for j := 1; j < NumActions; j++ {
			if aRow[j] > aMaxVal {
				aMaxVal = aRow[j]
				aMaxIdx = j
			}
		}
		results[i] = ThreatPrediction{
			ThreatClass:      ThreatClass(tMaxIdx),
			ThreatProb:       1.0 - tRow[0], // P(threat) = 1 - P(none)
			Action:           ActionClass(aMaxIdx),
			ActionConfidence: aMaxVal,
			ClassProbs:       classProbs,
			ActionProbs:      actionProbs,
		}
	}
	return results
}

// AssessSingle assesses a single sample (convenience method).
func (ta *ThreatAssessor) AssessSingle(embedding []float64, forgeryResult *ForgeryResult, behavior []float64) ThreatPrediction {
	input := buildThreatInput(embedding, forgeryResult, behavior)
	return ta.Assess(FromSlice(input))[0]
}

// buildThreatInput constructs the threat assessor input vector.
func buildThreatInput(embedding []float64, fr *ForgeryResult, behavior []float64) []float64 {
	dim := EmbeddingDim + 1 + NumForgeryTypes + BehaviorFeatureDim
	vec := make([]float64, dim)
	// Embedding (32-dim)
	copy(vec[:EmbeddingDim], embedding)
	// Forgery probability (1-dim) + forgery type probabilities (4-dim)
	off := EmbeddingDim
	if fr != nil {
		vec[off] = fr.ForgeryProb
		if len(fr.TypeProbs) == NumForgeryTypes {
			copy(vec[off+1:off+1+NumForgeryTypes], fr.TypeProbs)
		}
	}
	// Behavioral features (8-dim)
	off += 1 + NumForgeryTypes
	if len(behavior) >= BehaviorFeatureDim {
		copy(vec[off:off+BehaviorFeatureDim], behavior[:BehaviorFeatureDim])
	}
	return vec
}
