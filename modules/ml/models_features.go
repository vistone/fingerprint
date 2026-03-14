package ml

import (
	"math"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

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
