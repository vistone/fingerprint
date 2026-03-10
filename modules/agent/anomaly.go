package agent

import (
	"fmt"
	"math"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
)

// AnomalyDetector knowledge-driven anomaly detector
//
// Leverages global fingerprint blueprints from KnowledgeBase to perform
// cross-layer consistency validation on each observation: TLS ↔ HTTP/2 ↔ HTTP headers ↔ TCP/IP ↔ JS features
// whether all point to the same real browser identity.
//
// Any contradictory signals are marked as Contradiction, accumulated into SuspicionScore.
type AnomalyDetector struct {
	kb *KnowledgeBase
}

// NewAnomalyDetector create knowledge-driven anomaly detector
func NewAnomalyDetector(kb *KnowledgeBase) *AnomalyDetector {
	return &AnomalyDetector{kb: kb}
}

// Analyze perform knowledge-driven anomaly analysis on an observation
//
// It will:
//  1. Extract claimed browser identity from ML classification result
//  2. Use KnowledgeBase to find known blueprint for that identity
//  3. Layer-by-layer validation: whether TLS, HTTP/2, TCP/IP features match blueprint
//  4. Aggregate contradictory signals and calculate suspicion score
func (ad *AnomalyDetector) Analyze(obs *Observation) *MatchResult {
	result := &MatchResult{
		MatchScore:     1.0,
		SuspicionScore: 0.0,
	}

	// Cannot perform knowledge validation without classification result
	if obs.Classification == nil {
		return result
	}

	family := obs.Classification.Family
	version := obs.Classification.Version
	result.ClosestFamily = family
	result.ClosestVersion = version

	bk := ad.kb.GetBrowserKnowledge(family)
	if bk == nil {
		// Unknown browser family, slightly suspicious
		result.Contradictions = append(result.Contradictions, Contradiction{
			Field: "browser_family", Expected: "known family", Actual: string(family), Severity: "low",
		})
		result.SuspicionScore = 0.15
		return result
	}

	result.Known = true

	// ---- TLS layer validation ----
	ad.checkTLS(obs, bk, result)

	// ---- HTTP/2 layer validation ----
	ad.checkHTTP2(obs, family, result)

	// ---- TCP/IP layer validation ----
	ad.checkTCPIP(obs, result)

	// ---- Classification confidence validation ----
	ad.checkClassificationConfidence(obs, result)

	// ---- Headless / automation tool marker validation ----
	ad.checkAutomationMarkers(obs, result)

	// Aggregate suspicion score
	result.SuspicionScore = ad.computeSuspicionScore(result.Contradictions)
	result.MatchScore = math.Max(0, 1.0-result.SuspicionScore)

	return result
}

// ===================================================================
// TLS layer validation
// ===================================================================

func (ad *AnomalyDetector) checkTLS(obs *Observation, bk *BrowserKnowledge, mr *MatchResult) {
	if obs.Features == nil {
		return
	}

	// Cipher suite count validation
	cipherCount := obs.Features.Get(core.FeatureCipherSuites)
	if cipherCount > 0 {
		rangeInfo, ok := ad.kb.tls.CipherCountRange[bk.Family]
		if ok {
			count := int(cipherCount)
			if count < rangeInfo[0] || count > rangeInfo[1] {
				mr.Contradictions = append(mr.Contradictions, Contradiction{
					Field:    "tls_cipher_count",
					Expected: fmt.Sprintf("%d-%d for %s", rangeInfo[0], rangeInfo[1], bk.Family),
					Actual:   fmt.Sprintf("%d", count),
					Severity: "medium",
				})
			}
		}
	}

	// Extension count validation
	extCount := obs.Features.Get(core.FeatureExtensions)
	if extCount > 0 {
		rangeInfo, ok := ad.kb.tls.ExtensionCountRange[bk.Family]
		if ok {
			count := int(extCount)
			if count < rangeInfo[0] || count > rangeInfo[1] {
				mr.Contradictions = append(mr.Contradictions, Contradiction{
					Field:    "tls_extension_count",
					Expected: fmt.Sprintf("%d-%d for %s", rangeInfo[0], rangeInfo[1], bk.Family),
					Actual:   fmt.Sprintf("%d", count),
					Severity: "medium",
				})
			}
		}
	}
}

// ===================================================================
// HTTP/2 layer validation
// ===================================================================

func (ad *AnomalyDetector) checkHTTP2(obs *Observation, family core.BrowserType, mr *MatchResult) {
	if obs.Features == nil {
		return
	}

	h2Hash := obs.Features.Get(core.FeatureHTTP2Settings)
	if h2Hash == 0 {
		return
	}

	expected := ad.kb.GetExpectedH2(family)
	if expected == nil {
		return
	}

	// FeatureVector HTTP2Settings is a hash value, check metadata for detailed info
	if obs.Features.Metadata == nil {
		return
	}

	// Check HTTP/2 InitialWindowSize (if available in metadata)
	if ws, ok := obs.Features.Metadata["h2_initial_window_size"]; ok {
		if wsVal, ok := ws.(float64); ok {
			expectedWS := float64(expected.InitialWindowSize)
			// Allow 10% deviation
			if math.Abs(wsVal-expectedWS)/expectedWS > 0.1 {
				mr.Contradictions = append(mr.Contradictions, Contradiction{
					Field:    "h2_initial_window_size",
					Expected: fmt.Sprintf("%d for %s", expected.InitialWindowSize, family),
					Actual:   fmt.Sprintf("%.0f", wsVal),
					Severity: "high",
				})
			}
		}
	}

	// check MaxConcurrentStreams
	if mcs, ok := obs.Features.Metadata["h2_max_concurrent_streams"]; ok {
		if mcsVal, ok := mcs.(float64); ok {
			expectedMCS := float64(expected.MaxConcurrentStreams)
			if mcsVal != expectedMCS {
				mr.Contradictions = append(mr.Contradictions, Contradiction{
					Field:    "h2_max_concurrent_streams",
					Expected: fmt.Sprintf("%.0f for %s", expectedMCS, family),
					Actual:   fmt.Sprintf("%.0f", mcsVal),
					Severity: "high",
				})
			}
		}
	}

	// Check pseudo-header order (Chrome vs Firefox vs Safari all different)
	if pho, ok := obs.Features.Metadata["pseudo_header_order"]; ok {
		if phoStr, ok := pho.(string); ok && len(expected.PseudoHeaderOrder) > 0 {
			expectedOrder := strings.Join(expected.PseudoHeaderOrder, ",")
			if phoStr != expectedOrder {
				mr.Contradictions = append(mr.Contradictions, Contradiction{
					Field:    "h2_pseudo_header_order",
					Expected: expectedOrder + " (" + string(family) + ")",
					Actual:   phoStr,
					Severity: "high",
				})
			}
		}
	}
}

// ===================================================================
// TCP/IP layer validation
// ===================================================================

func (ad *AnomalyDetector) checkTCPIP(obs *Observation, mr *MatchResult) {
	if obs.Metadata == nil {
		return
	}

	osFamily := normalizeOSFamily(obs.Metadata["os_family"])
	if osFamily == "" {
		return
	}

	expected := ad.kb.GetExpectedTCPIP(osFamily)
	if expected == nil {
		return
	}

	// TTL validation
	if ttlStr, ok := obs.Metadata["tcp_ttl"]; ok {
		var ttl int
		if _, err := fmt.Sscanf(ttlStr, "%d", &ttl); err == nil {
			// TTL decrements with hop count, expected value ± certain range
			expectedTTL := int(expected.TTL)
			// In hot-path only check if matches OS segment (64 segment vs 128 segment)
			if (expectedTTL > 100 && ttl <= 100) || (expectedTTL <= 100 && ttl > 100) {
				mr.Contradictions = append(mr.Contradictions, Contradiction{
					Field:    "tcp_ttl",
					Expected: fmt.Sprintf("~%d for %s", expectedTTL, osFamily),
					Actual:   fmt.Sprintf("%d", ttl),
					Severity: "high",
				})
			}
		}
	}

	// TCP Window Size validation
	if wsStr, ok := obs.Metadata["tcp_window_size"]; ok {
		var ws int
		if _, err := fmt.Sscanf(wsStr, "%d", &ws); err == nil {
			expectedWS := int(expected.WindowSize)
			// Allow 20% deviation (network conditions may affect)
			diff := math.Abs(float64(ws-expectedWS)) / float64(expectedWS)
			if diff > 0.3 {
				mr.Contradictions = append(mr.Contradictions, Contradiction{
					Field:    "tcp_window_size",
					Expected: fmt.Sprintf("%d for %s", expectedWS, osFamily),
					Actual:   fmt.Sprintf("%d", ws),
					Severity: "medium",
				})
			}
		}
	}
}

// normalizeOSFamily normalize various OS representations to knowledge base keys
func normalizeOSFamily(raw string) string {
	lower := strings.ToLower(raw)
	switch {
	case strings.Contains(lower, "windows"):
		return "windows"
	case strings.Contains(lower, "macos") || strings.Contains(lower, "mac os") || strings.Contains(lower, "macintosh"):
		return "macos"
	case strings.Contains(lower, "linux") && !strings.Contains(lower, "android"):
		return "linux"
	case strings.Contains(lower, "ios") || strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad"):
		return "ios"
	case strings.Contains(lower, "android"):
		return "android"
	default:
		return ""
	}
}

// ===================================================================
// Classification confidence validation
// ===================================================================

func (ad *AnomalyDetector) checkClassificationConfidence(obs *Observation, mr *MatchResult) {
	if obs.Classification == nil {
		return
	}
	// If ML classification confidence too low, indicates feature pattern is atypical
	if obs.Classification.Confidence < 0.5 {
		mr.Contradictions = append(mr.Contradictions, Contradiction{
			Field:    "classification_confidence",
			Expected: ">= 0.5",
			Actual:   fmt.Sprintf("%.2f", obs.Classification.Confidence),
			Severity: "medium",
		})
	}
}

// ===================================================================
// Automation tool marker validation
// ===================================================================

func (ad *AnomalyDetector) checkAutomationMarkers(obs *Observation, mr *MatchResult) {
	if obs.Features == nil {
		return
	}

	// check headless browser feature
	headless := obs.Features.Get(core.FeatureHeadlessBrowser)
	if headless > 0 {
		mr.Contradictions = append(mr.Contradictions, Contradiction{
			Field:    "headless_browser",
			Expected: "0 (no headless markers)",
			Actual:   fmt.Sprintf("%.1f", headless),
			Severity: "high",
		})
	}

	// Check automation tool markers
	toolMarker := obs.Features.Get(core.FeatureToolMarker)
	if toolMarker > 0 {
		mr.Contradictions = append(mr.Contradictions, Contradiction{
			Field:    "tool_marker",
			Expected: "0 (no tool markers)",
			Actual:   fmt.Sprintf("%.1f", toolMarker),
			Severity: "high",
		})
	}
}

// ===================================================================
// Suspicion score calculation
// ===================================================================

var severityWeights = map[string]float64{
	"low":    0.05,
	"medium": 0.15,
	"high":   0.30,
}

func (ad *AnomalyDetector) computeSuspicionScore(contradictions []Contradiction) float64 {
	if len(contradictions) == 0 {
		return 0
	}

	score := 0.0
	for _, c := range contradictions {
		w, ok := severityWeights[c.Severity]
		if !ok {
			w = 0.1
		}
		score += w
	}

	// Multiple contradictions have cumulative effect, but capped at 1.0
	if score > 1.0 {
		score = 1.0
	}
	return score
}
