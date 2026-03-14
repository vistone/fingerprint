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
