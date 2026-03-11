// Package ml — models.go provides domain-specific neural network models for
// multi-layer browser fingerprint analysis.
//
// ┌─────────────────────────────────────────────────────────────────────┐
// │       Fingerprint Analysis Model Library — Domain-Driven Design       │
// ├─────────────────────────────────────────────────────────────────────┤
// │                                                                     │
// │  Core mission:                                                       │
// │    Identify real browsers vs automation tools / forged clients        │
// │    through multi-layer fingerprint analysis                           │
// │    (TLS + HTTP/2 + TCP/IP + JS + Behavioral)                         │
// │                                                                     │
// │  Four specialized models, each with clear purpose:                    │
// │                                                                     │
// │  ┌──────────────────┐                                               │
// │  │ FingerprintEncoder│ Learn: intrinsic structure & browser uniqueness │
// │  │  (encoder)        │ Infer: map raw features to 32-dim embeddings   │
// │  │                  │ Output: embeddings — same browser close,        │
// │  │                  │         different browsers far apart             │
// │  └────────┬─────────┘                                               │
// │           │ 32-dim embedding                                         │
// │           ▼                                                          │
// │  ┌──────────────────┐                                               │
// │  │ BrowserClassifier│ Learn: embedding-to-identity mapping             │
// │  │  (classifier)     │ Infer: identify browser family from embedding   │
// │  │                  │ Output: family probability + confidence         │
// │  └────────┬─────────┘                                               │
// │           │ classification result                                    │
// │  ┌──────────────────┐                                               │
// │  │ ForgeryDetector  │ Learn: real vs forged fingerprint patterns      │
// │  │  (detector)       │ Infer: cross-layer consistency analysis         │
// │  │                  │ Output: forgery prob + type (real/headless/      │
// │  │                  │         antidetect/proxy)                        │
// │  └────────┬─────────┘                                               │
// │           │ detection result                                         │
// │  ┌──────────────────┐                                               │
// │  │ ThreatAssessor   │ Learn: optimal security response from all       │
// │  │  (assessor)       │       signals combined                          │
// │  │                  │ Infer: threat class + recommended action         │
// │  │                  │ Output: threat probabilities + action probs      │
// │  └──────────────────┘                                               │
// │                                                                     │
// │  Training strategies:                                                 │
// │    FingerprintEncoder: Triplet Margin Loss over 207 browser profiles  │
// │    BrowserClassifier:  Cross-entropy from profile labels              │
// │    ForgeryDetector:    Binary CE from real + synthetic forged data    │
// │    ThreatAssessor:     CE from rule labels, then feedback fine-tuning │
// │                                                                     │
// │  GPU acceleration: all forward/backward operations use Tensor ops,    │
// │  switchable to GPU backend via SetDevice(gpu)                        │
// └─────────────────────────────────────────────────────────────────────┘
package ml

import (
	"math"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// =========================================================================
// Feature encoding constants — define input dimensions
// =========================================================================

const (
	// FingerprintFeatureDim is the raw fingerprint feature dimension (30-dim).
	//
	// TLS layer (8-dim):
	//   [0] tls_version:       TLS version normalized (1.0=TLS1.0, 1.1, 1.2→0.75, 1.3→1.0)
	//   [1] cipher_count:      cipher suite count / 20.0
	//   [2] tls13_ratio:       TLS 1.3 cipher suite ratio
	//   [3] extension_count:   extension count / 20.0
	//   [4] has_sni:           SNI present (0/1)
	//   [5] has_alpn:          ALPN present (0/1)
	//   [6] curve_count:       elliptic curve count / 6.0
	//   [7] grease_ratio:      GREASE value ratio
	//
	// HTTP/2 layer (6-dim):
	//   [8]  h2_window:        initial window size / 10M
	//   [9]  h2_streams:       max concurrent streams / 1000
	//   [10] h2_header_table:  header table size / 100K
	//   [11] h2_pseudo_order:  pseudo-header order encoding / 24 (24 permutations)
	//   [12] h2_priority:      priority frame (0/1)
	//   [13] h2_settings_cnt:  settings count / 10
	//
	// TCP/IP layer (4-dim):
	//   [14] tcp_ttl:          TTL normalized (32→0.25, 64→0.5, 128→1.0)
	//   [15] tcp_window:       TCP window size / 128K
	//   [16] tcp_mss:          MSS / 2000
	//   [17] tcp_timestamps:   TCP timestamps (0/1)
	//
	// JS frontend layer (8-dim):
	//   [18] canvas_entropy:   Canvas fingerprint entropy
	//   [19] webgl_score:      WebGL score
	//   [20] audio_entropy:    Audio fingerprint entropy
	//   [21] font_count:       font count / 200
	//   [22] storage_score:    storage score
	//   [23] webrtc_active:    WebRTC active (0/1)
	//   [24] hardware_cores:   CPU core count / 16
	//   [25] headless_score:   headless browser score
	//
	// Meta-feature layer (4-dim):
	//   [26] ua_entropy:       User-Agent entropy
	//   [27] config_entropy:   overall config entropy
	//   [28] tool_marker:      automation tool marker
	//   [29] behavior_pattern: behavioral pattern feature
	FingerprintFeatureDim = 30

	// EmbeddingDim is the embedding vector dimension.
	EmbeddingDim = 32

	// CrossLayerFeatureDim is the cross-layer consistency feature dimension (10-dim).
	//   [0] tls_h2_window_match:   TLS↔HTTP/2 window match score
	//   [1] tls_h2_pseudo_match:   TLS↔HTTP/2 pseudo-header order match
	//   [2] tls_tcp_ttl_match:     TLS↔TCP/IP TTL match score
	//   [3] ua_tls_version_match:  UA↔TLS version match
	//   [4] ua_h2_settings_match:  UA↔HTTP/2 settings match
	//   [5] js_headless_indicator: JS headless browser indicator
	//   [6] canvas_webgl_consist:  Canvas↔WebGL consistency
	//   [7] cipher_order_anomaly:  cipher suite order anomaly score
	//   [8] ext_pattern_anomaly:   extension pattern anomaly score
	//   [9] contradiction_count:   cross-layer contradiction count (normalized)
	CrossLayerFeatureDim = 10

	// BehaviorFeatureDim is the behavioral feature dimension (8-dim).
	//   [0] fp_switch_rate:      fingerprint switch rate / 10
	//   [1] request_rate:        request rate / 20
	//   [2] consistency_score:   consistency score [0,1]
	//   [3] risk_trend:          risk trend [-1,1] → [0,1]
	//   [4] observations_norm:   observation count normalized
	//   [5] unique_fp_ratio:     unique fingerprint ratio
	//   [6] session_duration:    session duration normalized
	//   [7] burst_indicator:     burst request indicator
	BehaviorFeatureDim = 8

	// NumBrowserFamilies is the number of browser families.
	NumBrowserFamilies = 7 // chrome, firefox, safari, edge, opera, brave, samsung

	// NumForgeryTypes is the number of forgery types.
	NumForgeryTypes = 4 // real, headless, antidetect, proxy

	// NumThreatClasses is the number of threat classes.
	NumThreatClasses = 6 // none, bot, fingerprint_spoof, session_anomaly, behavioral_anomaly, evasion

	// NumActions is the number of security actions.
	NumActions = 5 // allow, monitor, challenge, throttle, block
)

// =========================================================================
// Browser family encoding
// =========================================================================

var familyIndex = map[core.BrowserType]int{
	core.BrowserChrome:  0,
	core.BrowserFirefox: 1,
	core.BrowserSafari:  2,
	core.BrowserEdge:    3,
	core.BrowserOpera:   4,
	core.BrowserBrave:   5,
	core.BrowserSamsung: 6,
}

var indexFamily = [NumBrowserFamilies]core.BrowserType{
	core.BrowserChrome,
	core.BrowserFirefox,
	core.BrowserSafari,
	core.BrowserEdge,
	core.BrowserOpera,
	core.BrowserBrave,
	core.BrowserSamsung,
}

// =========================================================================
// Model 1: FingerprintEncoder — fingerprint embedding encoder
// =========================================================================
//
// Learn: intrinsic structure and browser uniqueness from multi-layer fingerprints.
//        Fingerprints from the same browser (different sessions, different networks)
//        should map to nearby positions in embedding space; different browsers
//        should map to distant positions.
//
// Infer: input 30-dim raw fingerprint vector → output 32-dim L2-normalized embedding
//
// Output: dense embedding vectors satisfying:
//         - Same browser family: cosine similarity > 0.8
//         - Different browser families: cosine similarity < 0.3
//         - Forged fingerprints: not near any known browser cluster center
//
// Train: Triplet Margin Loss
//        anchor=known browser, positive=same browser variant, negative=other browser
//        Data: 207 real browser profiles + augmentation

// FingerprintEncoder is the fingerprint embedding model.
type FingerprintEncoder struct {
	Net *Sequential // internal neural network
}

// NewFingerprintEncoder creates a fingerprint encoder.
// Architecture: Input(30) → Dense(256) → BN → ReLU → Dropout(0.2)
//
//	→ Dense(128) → BN → ReLU → Dropout(0.1)
//	→ Dense(64) → BN → ReLU → Dense(32)
func NewFingerprintEncoder() *FingerprintEncoder {
	return &FingerprintEncoder{
		Net: NewSequential(
			NewDenseLayer(FingerprintFeatureDim, 256),
			NewBatchNormLayer(256),
			NewReLULayer(),
			NewDropoutLayer(0.2),
			NewDenseLayer(256, 128),
			NewBatchNormLayer(128),
			NewReLULayer(),
			NewDropoutLayer(0.1),
			NewDenseLayer(128, 64),
			NewBatchNormLayer(64),
			NewReLULayer(),
			NewDenseLayer(64, EmbeddingDim),
		),
	}
}

// Encode encodes raw fingerprint features into embedding vectors.
// features: [batch × 30] → embedding: [batch × 32] (L2-normalized)
func (enc *FingerprintEncoder) Encode(features *Tensor) *Tensor {
	raw := enc.Net.Forward(features)
	// Row-wise L2 normalization
	rows := raw.Shape[0]
	cols := raw.Shape[1]
	for i := 0; i < rows; i++ {
		norm := 0.0
		for j := 0; j < cols; j++ {
			v := raw.At(i, j)
			norm += v * v
		}
		norm = math.Sqrt(norm)
		if norm < 1e-12 {
			norm = 1e-12
		}
		for j := 0; j < cols; j++ {
			raw.Set(i, j, raw.At(i, j)/norm)
		}
	}
	return raw
}

// EncodeSingle encodes a single fingerprint vector (convenience method).
func (enc *FingerprintEncoder) EncodeSingle(features []float64) []float64 {
	input := FromSlice(features)
	return enc.Encode(input).ToSlice()
}

// =========================================================================
// Model 2: BrowserClassifier — browser family classifier
// =========================================================================
//
// Learn: mapping from embedding space to browser identity.
//        Different browser families have different TLS configs, HTTP/2 settings,
//        TCP behavior — these differences are compressed by the encoder into
//        embeddings; the classifier learns to interpret them.
//
// Infer: input 32-dim embedding → output browser family probability (7 classes)
//
// Output: BrowserPrediction containing:
//         - Family probability distribution (Chrome 95%, Firefox 3%, Safari 2%, ...)
//         - Predicted family (chrome)
//         - Confidence (0.95)
//
// Train: Cross-entropy loss from 207 known browser profiles

// BrowserClassifier is the browser family classifier model.
type BrowserClassifier struct {
	Net *Sequential // internal neural network
}

// NewBrowserClassifier creates a browser classifier.
// Architecture: Embedding(32) → Dense(128) → BN → ReLU → Dropout(0.2)
//
//	→ Dense(64) → BN → ReLU → Dropout(0.1)
//	→ Dense(7) → Softmax
func NewBrowserClassifier() *BrowserClassifier {
	return &BrowserClassifier{
		Net: NewSequential(
			NewDenseLayer(EmbeddingDim, 128),
			NewBatchNormLayer(128),
			NewReLULayer(),
			NewDropoutLayer(0.2),
			NewDenseLayer(128, 64),
			NewBatchNormLayer(64),
			NewReLULayer(),
			NewDropoutLayer(0.1),
			NewDenseLayer(64, NumBrowserFamilies),
			NewSoftmaxLayer(),
		),
	}
}

// BrowserPrediction holds browser classification results.
type BrowserPrediction struct {
	Family      core.BrowserType   // predicted browser family
	Confidence  float64            // confidence [0,1]
	Probs       []float64          // family probability distribution
	FamilyNames []core.BrowserType // family names, aligned with Probs
}

// Classify predicts browser family.
// embedding: [batch × 32] → returns prediction for each row
func (bc *BrowserClassifier) Classify(embedding *Tensor) []BrowserPrediction {
	probs := bc.Net.Forward(embedding)
	batch := probs.Shape[0]
	results := make([]BrowserPrediction, batch)
	for i := 0; i < batch; i++ {
		row := probs.Row(i)
		maxIdx := 0
		maxVal := row[0]
		probsCopy := make([]float64, NumBrowserFamilies)
		copy(probsCopy, row)
		for j := 1; j < NumBrowserFamilies; j++ {
			if row[j] > maxVal {
				maxVal = row[j]
				maxIdx = j
			}
		}
		names := make([]core.BrowserType, NumBrowserFamilies)
		copy(names, indexFamily[:])
		results[i] = BrowserPrediction{
			Family:      indexFamily[maxIdx],
			Confidence:  maxVal,
			Probs:       probsCopy,
			FamilyNames: names,
		}
	}
	return results
}

// ClassifySingle classifies a single embedding vector (convenience method).
func (bc *BrowserClassifier) ClassifySingle(embedding []float64) BrowserPrediction {
	input := FromSlice(embedding)
	return bc.Classify(input)[0]
}

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

// =========================================================================
// Feature encoding functions — build model inputs from raw data
// =========================================================================

// EncodeFingerprint extracts a 30-dim fingerprint feature vector from a ClientProfile.
// This bridges raw profile data and neural network models.
func EncodeFingerprint(profile *profiles.ClientProfile) []float64 {
	vec := make([]float64, FingerprintFeatureDim)

	// TLS layer (index 0-7)
	vec[0] = normalizeTLSVersion(profile.TLSVersion)
	vec[1] = float64(len(profile.CipherSuites)) / 20.0
	vec[2] = tls13Ratio(profile.CipherSuites)
	vec[3] = float64(len(profile.Extensions)) / 20.0
	vec[4] = boolToFloat(hasSNI(profile.Extensions))
	vec[5] = boolToFloat(hasALPN(profile.Extensions))
	vec[6] = float64(len(profile.SupportedCurves)) / 6.0
	vec[7] = greaseRatio(profile.CipherSuites)

	// HTTP/2 layer (index 8-13)
	h2 := profile.HTTP2Settings
	vec[8] = float64(h2.InitialWindowSize) / 10_000_000.0
	vec[9] = float64(h2.MaxConcurrentStreams) / 1000.0
	vec[10] = float64(h2.HeaderTableSize) / 100_000.0
	vec[11] = encodePseudoHeaderOrder(profile.PseudoHeaderOrder) / 24.0
	if h2.EnablePush > 0 {
		vec[12] = 1.0
	}
	vec[13] = float64(countH2Settings(h2)) / 10.0

	// TCP/IP layer (index 14-17)
	if tcp := profile.TCPIP; tcp != nil {
		vec[14] = float64(tcp.TTL) / 128.0
		vec[15] = float64(tcp.WindowSize) / 131072.0
		vec[16] = float64(tcp.MSS) / 2000.0
		vec[17] = boolToFloat(tcp.Timestamps)
	}

	// JS frontend layer: unavailable in profile (zero values), populated at runtime by Frontend SDK
	// Index 18-25 remain 0

	// Meta-feature layer (index 26-29)
	if profile.Headers != nil {
		vec[26] = stringEntropy(profile.Headers.UserAgent) / 5.0
	}
	vec[27] = profileEntropy(profile) / 5.0
	// Index 28-29 populated at runtime

	// Clip to [0, 1] range
	for i := range vec {
		vec[i] = math.Max(0, math.Min(1, vec[i]))
	}
	return vec
}

// EncodeFingerprintFromFeatureVector builds a 30-dim vector from an already-extracted FeatureVector.
// This allows integration with the existing FeatureExtractor pipeline.
func EncodeFingerprintFromFeatureVector(fv *core.FeatureVector) []float64 {
	vec := make([]float64, FingerprintFeatureDim)
	if fv == nil {
		return vec
	}
	// Map existing features to new dimensions
	vec[0] = normalizeFeatureValue(fv.Get(core.FeatureTLSVersion), 0x0304)
	vec[1] = fv.Get(core.FeatureCipherSuites) / 20.0
	vec[3] = fv.Get(core.FeatureExtensions) / 20.0
	vec[8] = fv.Get(core.FeatureHTTP2Settings) / 10_000_000.0
	vec[10] = fv.Get(core.FeatureHTTPHeaders) / 100.0
	vec[18] = fv.Get(core.FeatureCanvas) / 100.0
	vec[19] = fv.Get(core.FeatureWebGL) / 100.0
	vec[20] = fv.Get(core.FeatureAudio) / 100.0
	vec[21] = fv.Get(core.FeatureFonts) / 200.0
	vec[22] = fv.Get(core.FeatureStorage) / 100.0
	vec[23] = fv.Get(core.FeatureWebRTC) // already 0/1
	vec[24] = fv.Get(core.FeatureHardware) / 16.0
	vec[25] = fv.Get(core.FeatureHeadlessBrowser)
	vec[26] = fv.Get(core.FeatureUserAgent) / 200.0
	vec[27] = fv.Get(core.FeatureEntropy) / 5.0
	vec[28] = fv.Get(core.FeatureToolMarker)
	vec[29] = fv.Get(core.FeatureBehaviorPattern)

	for i := range vec {
		vec[i] = math.Max(0, math.Min(1, vec[i]))
	}
	return vec
}

// ComputeCrossLayerFeatures computes cross-layer consistency features (10-dim).
// These features capture contradiction levels between layers — key signals for the forgery detector.
func ComputeCrossLayerFeatures(fp []float64) []float64 {
	cross := make([]float64, CrossLayerFeatureDim)
	if len(fp) < FingerprintFeatureDim {
		return cross
	}

	// [0] TLS<>HTTP/2 window match: Chrome 6.2M, Firefox 131K, Safari 2M
	// Window size and cipher count should correspond to the same browser
	cross[0] = 1.0 - math.Abs(fp[1]-fp[8])*2.0 // cipher_count vs h2_window coordination

	// [1] TLS<>HTTP/2 pseudo-header order match
	cross[1] = 1.0 - math.Abs(fp[2]-fp[11]) // tls13_ratio vs pseudo_order coordination

	// [2] TLS<>TCP/IP TTL match
	// Chrome/Firefox/Safari on different OSes should have corresponding TTL values
	cross[2] = 1.0 - math.Abs(fp[0]-fp[14]) // tls_version vs ttl coordination

	// [3] UA<>TLS version match
	cross[3] = 1.0 - math.Abs(fp[26]-fp[0]) // ua_entropy vs tls_version

	// [4] UA<>HTTP/2 settings match
	cross[4] = 1.0 - math.Abs(fp[26]-fp[8]) // ua_entropy vs h2_window

	// [5] JS headless browser indicator (used directly)
	cross[5] = fp[25] // headless_score

	// [6] Canvas<>WebGL consistency (both should be present or both absent)
	if fp[18] > 0.1 && fp[19] > 0.1 {
		cross[6] = 1.0 // both present
	} else if fp[18] < 0.1 && fp[19] < 0.1 {
		cross[6] = 0.8 // both absent (possible privacy mode or headless)
	} else {
		cross[6] = 0.2 // inconsistent — suspicious
	}

	// [7] Cipher suite order anomaly score
	// TLS 1.3 browsers should have a high tls13_ratio
	if fp[0] > 0.8 { // TLS 1.3
		cross[7] = fp[2] // tls13_ratio should be high
	} else {
		cross[7] = 1.0 - fp[2] // TLS 1.2 should not have too many 1.3 suites
	}

	// [8] Extension pattern anomaly score
	// Extension count and cipher count should be in a reasonable ratio
	if fp[1] > 0 {
		ratio := fp[3] / fp[1]               // ext_count / cipher_count
		cross[8] = 1.0 - math.Abs(ratio-1.0) // ideal ratio approx 1:1
	}

	// [9] Normalized cross-layer contradiction count (computed from consistency scores)
	contradictions := 0.0
	for i := 0; i < 9; i++ {
		if cross[i] < 0.3 {
			contradictions++
		}
	}
	cross[9] = contradictions / 9.0

	for i := range cross {
		cross[i] = math.Max(0, math.Min(1, cross[i]))
	}
	return cross
}

// =========================================================================
// Helper functions
// =========================================================================

func normalizeTLSVersion(v uint16) float64 {
	switch v {
	case 0x0304:
		return 1.0 // TLS 1.3
	case 0x0303:
		return 0.75 // TLS 1.2
	case 0x0302:
		return 0.5 // TLS 1.1
	case 0x0301:
		return 0.25 // TLS 1.0
	default:
		return 0.0
	}
}

func normalizeFeatureValue(val, maxVal float64) float64 {
	if maxVal == 0 {
		return 0
	}
	return math.Min(1.0, val/maxVal)
}

func tls13Ratio(suites []uint16) float64 {
	if len(suites) == 0 {
		return 0
	}
	count := 0
	for _, s := range suites {
		if s == 0x1301 || s == 0x1302 || s == 0x1303 {
			count++
		}
	}
	return float64(count) / float64(len(suites))
}

func hasSNI(exts []core.TLSExtension) bool {
	for _, e := range exts {
		if e.Type == 0x0000 { // SNI extension
			return true
		}
	}
	return false
}

func hasALPN(exts []core.TLSExtension) bool {
	for _, e := range exts {
		if e.Type == 0x0010 { // ALPN extension
			return true
		}
	}
	return false
}

func greaseRatio(suites []uint16) float64 {
	if len(suites) == 0 {
		return 0
	}
	count := 0
	for _, s := range suites {
		if isGREASE(s) {
			count++
		}
	}
	return float64(count) / float64(len(suites))
}

func isGREASE(v uint16) bool {
	return (v & 0x0f0f) == 0x0a0a
}

func encodePseudoHeaderOrder(order []string) float64 {
	// Encode pseudo-header order as a unique integer
	// Chrome: :method, :authority, :scheme, :path → 0
	// Firefox: :method, :path, :authority, :scheme → 1
	// Safari: :method, :scheme, :path, :authority → 2
	if len(order) < 4 {
		return 0
	}
	hash := 0.0
	for i, h := range order {
		switch h {
		case ":method":
			hash += float64(i) * 1
		case ":authority":
			hash += float64(i) * 4
		case ":scheme":
			hash += float64(i) * 16
		case ":path":
			hash += float64(i) * 64
		}
	}
	return hash
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func countH2Settings(h2 core.HTTP2Settings) int {
	count := 0
	if h2.HeaderTableSize != 0 {
		count++
	}
	if h2.EnablePush != 0 {
		count++
	}
	if h2.MaxConcurrentStreams != 0 {
		count++
	}
	if h2.InitialWindowSize != 0 {
		count++
	}
	if h2.MaxFrameSize != 0 {
		count++
	}
	if h2.MaxHeaderListSize != 0 {
		count++
	}
	return count
}

func stringEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	freq := make(map[rune]int)
	for _, c := range s {
		freq[c]++
	}
	entropy := 0.0
	n := float64(len([]rune(s)))
	for _, count := range freq {
		p := float64(count) / n
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}

func profileEntropy(p *profiles.ClientProfile) float64 {
	// Measure entropy from diversity of profile parameters
	e := 0.0
	e += float64(len(p.CipherSuites)) * 0.1
	e += float64(len(p.Extensions)) * 0.1
	e += float64(len(p.SupportedCurves)) * 0.2
	return e
}
