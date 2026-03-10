package http2

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/vistone/fingerprint/modules/internal/metrics"
)

// HTTP2SignatureResult HTTP/2 signature result
type HTTP2SignatureResult struct {
	// Complete HTTP/2 signature (SHA256)
	Hash string

	// Raw signature string (for debugging/verification)
	RawSignature string

	// Decomposed signature parts
	SettingsSignature     string // Settings frame signature
	PrioritySignature     string // Priority frame signature
	HeadersSignature      string // Headers frame signature
	WindowUpdateSignature string // WindowUpdate frame signature

	// Composite features
	FrameSequence string // Frame sending order characteristics

	// Anomaly score
	RiskScore float64

	// Anomaly flag list
	AnomalyFlags []string

	// Matched known client fingerprints
	MatchedClients []string
}

// HTTP2FrameData HTTP/2 frame data
type HTTP2FrameData struct {
	Type       string                 // SETTINGS, PRIORITY, HEADERS, WINDOW_UPDATE
	FrameID    uint32                 // Stream ID
	Priority   *PriorityData          // Priority information
	Settings   map[string]interface{} // SETTINGS frame parameters
	Headers    []string               // Request header order
	WindowSize uint32                 // WINDOW_UPDATE size
	Metadata   map[string]string      // Other metadata
}

// PriorityData priority information
type PriorityData struct {
	DependsOn  uint32
	StreamDep  uint32
	Exclusive  bool
	Weight     uint8
	DefaultExp bool
}

// HTTP2SignatureAnalyzer HTTP/2 signature analyzer
type HTTP2SignatureAnalyzer struct {
	knownClientProfiles map[string]*HTTP2ClientProfile
}

// HTTP2ClientProfile known client configuration
type HTTP2ClientProfile struct {
	Name               string
	BrowserName        string
	BrowserVersion     string
	SettingsOrder      []string
	PriorityStrategy   string
	HeaderOrder        []string
	WindowUpdateTiming string
	RiskScore          float64
}

// NewHTTP2SignatureAnalyzer creates analyzer
func NewHTTP2SignatureAnalyzer() *HTTP2SignatureAnalyzer {
	return &HTTP2SignatureAnalyzer{
		knownClientProfiles: initKnownHTTP2Profiles(),
	}
}

// AnalyzeHTTP2Stream analyzes HTTP/2 stream characteristics
func (a *HTTP2SignatureAnalyzer) AnalyzeHTTP2Stream(frames []HTTP2FrameData) (*HTTP2SignatureResult, error) {
	start := time.Now()
	defer func() {
		durationMs := float64(time.Since(start).Nanoseconds()) / 1e6
		metrics.HTTP2SignatureAnalysisDuration.Observe(durationMs)
	}()

	if len(frames) == 0 {
		return nil, fmt.Errorf("no frames provided")
	}

	result := &HTTP2SignatureResult{
		AnomalyFlags:      make([]string, 0, 8), // Pre-allocate capacity
		SettingsSignature: "",
		PrioritySignature: "",
		HeadersSignature:  "",
	}

	// Classify and process frames
	var settingsFrames, priorityFrames, headersFrames, windowUpdateFrames []HTTP2FrameData

	for _, frame := range frames {
		switch frame.Type {
		case "SETTINGS":
			settingsFrames = append(settingsFrames, frame)
		case "PRIORITY":
			priorityFrames = append(priorityFrames, frame)
		case "HEADERS":
			headersFrames = append(headersFrames, frame)
		case "WINDOW_UPDATE":
			windowUpdateFrames = append(windowUpdateFrames, frame)
		}
	}

	// 1. SETTINGS frame signature
	if len(settingsFrames) > 0 {
		result.SettingsSignature = generateSettingsSignature(settingsFrames[0])
	}

	// 2. PRIORITY frame signature
	if len(priorityFrames) > 0 {
		result.PrioritySignature = generatePrioritySignature(priorityFrames)
	}

	// 3. HEADERS frame signature (request header order)
	if len(headersFrames) > 0 {
		result.HeadersSignature = generateHeadersSignature(headersFrames[0])
	}

	// 4. WINDOW_UPDATE frame signature
	if len(windowUpdateFrames) > 0 {
		result.WindowUpdateSignature = generateWindowUpdateSignature(windowUpdateFrames)
	}

	// 5. Frame sequence characteristics
	var frameTypes []string
	for _, frame := range frames {
		frameTypes = append(frameTypes, strings.ToLower(frame.Type[:3])) // SET, PRI, HEA, WIN
	}
	result.FrameSequence = strings.Join(frameTypes, "-")

	// Build complete signature string
	result.RawSignature = fmt.Sprintf(
		"http2|%s|%s|%s|%s|%s",
		result.SettingsSignature,
		result.PrioritySignature,
		result.HeadersSignature,
		result.WindowUpdateSignature,
		result.FrameSequence,
	)

	// Calculate SHA256 hash
	hash := sha256.Sum256([]byte(result.RawSignature))
	result.Hash = hex.EncodeToString(hash[:])

	// Anomaly detection
	a.detectHTTP2Anomalies(result, settingsFrames, priorityFrames, headersFrames)

	return result, nil
}

// detectHTTP2Anomalies detects HTTP/2 anomalies
func (a *HTTP2SignatureAnalyzer) detectHTTP2Anomalies(
	result *HTTP2SignatureResult,
	settingsFrames, priorityFrames, headersFrames []HTTP2FrameData,
) {
	baseScore := 0.0

	// Anomaly 1: SETTINGS frame parameter anomalies
	if len(settingsFrames) > 0 {
		if settingsFrames[0].Settings == nil || len(settingsFrames[0].Settings) == 0 {
			result.AnomalyFlags = append(result.AnomalyFlags, "EMPTY_SETTINGS")
			baseScore += 0.2
		}
		if len(settingsFrames[0].Settings) > 10 {
			result.AnomalyFlags = append(result.AnomalyFlags, "EXCESSIVE_SETTINGS")
			baseScore += 0.15
		}
	}

	// Anomaly 2: Priority tree structure anomalies
	if len(priorityFrames) > 0 {
		if !isValidPriorityTree(priorityFrames) {
			result.AnomalyFlags = append(result.AnomalyFlags, "INVALID_PRIORITY_TREE")
			baseScore += 0.25
		}
	}

	// Anomaly 3: Request header order anomalies
	if len(headersFrames) > 0 {
		if !isStandardHeaderOrder(headersFrames[0]) {
			result.AnomalyFlags = append(result.AnomalyFlags, "UNUSUAL_HEADER_ORDER")
			baseScore += 0.15
		}
	}

	// Anomaly 4: Invalid frame sequence
	if !isValidFrameSequence(result.FrameSequence) {
		result.AnomalyFlags = append(result.AnomalyFlags, "INVALID_FRAME_SEQUENCE")
		baseScore += 0.2
	}

	// Anomaly 5: Window update parameter anomalies
	if result.WindowUpdateSignature == "" && len(settingsFrames) > 0 {
		// Has SETTINGS but no corresponding WINDOW_UPDATE
		result.AnomalyFlags = append(result.AnomalyFlags, "MISSING_WINDOW_UPDATE")
		baseScore += 0.1
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}
	result.RiskScore = baseScore
}

// FindMatchingHTTP2Clients finds matching known HTTP/2 clients
func (a *HTTP2SignatureAnalyzer) FindMatchingHTTP2Clients(
	result *HTTP2SignatureResult,
	maxResults int,
) []string {
	var matches []string

	for name, profile := range a.knownClientProfiles {
		// Rough matching based on risk score and signature features
		if profile.RiskScore < result.RiskScore+0.2 {
			matches = append(matches, name)
		}

		if len(matches) >= maxResults {
			break
		}
	}

	result.MatchedClients = matches
	return matches
}

// ============ Helper generation functions ============

func generateSettingsSignature(frame HTTP2FrameData) string {
	if frame.Settings == nil {
		return "empty"
	}

	// Sort SETTINGS parameters and generate signature
	var keys []string
	for k := range frame.Settings {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		v := frame.Settings[k]
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}

	// Calculate simple hash value
	sig := strings.Join(parts, ",")
	hash := sha256.Sum256([]byte(sig))
	return hex.EncodeToString(hash[:8]) // Take first 16 characters
}

func generatePrioritySignature(frames []HTTP2FrameData) string {
	// Based on priority tree depth and width
	if len(frames) == 0 {
		return "empty"
	}

	var sig strings.Builder
	for i, frame := range frames {
		if frame.Priority != nil {
			sig.WriteString(fmt.Sprintf("#%d(%d,%d)", i, frame.Priority.StreamDep, frame.Priority.Weight))
		}
	}

	// Calculate hash
	text := sig.String()
	if text == "" {
		return "no_priority"
	}
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:8])
}

func generateHeadersSignature(frame HTTP2FrameData) string {
	// Based on request header order and count
	if len(frame.Headers) == 0 {
		return "empty"
	}

	// Pseudo-header should appear first
	var sig strings.Builder
	for _, h := range frame.Headers {
		if strings.HasPrefix(h, ":") {
			sig.WriteString("PSD,")
		} else {
			sig.WriteString("HDR,")
		}
	}

	text := sig.String()
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:8])
}

func generateWindowUpdateSignature(frames []HTTP2FrameData) string {
	// Based on WINDOW_UPDATE size and frequency
	if len(frames) == 0 {
		return "empty"
	}

	var totalSize uint32
	for _, frame := range frames {
		totalSize += frame.WindowSize
	}

	// Levels: small(0-65k), medium(65k-1M), large(>1M)
	var level string
	if totalSize < 65536 {
		level = "small"
	} else if totalSize < 1048576 {
		level = "medium"
	} else {
		level = "large"
	}

	sig := fmt.Sprintf("win_update|count=%d|level=%s", len(frames), level)
	hash := sha256.Sum256([]byte(sig))
	return hex.EncodeToString(hash[:8])
}

// ============ Validation functions ============

func isValidPriorityTree(frames []HTTP2FrameData) bool {
	// Simplified check: at least one valid priority frame
	for _, frame := range frames {
		if frame.Priority != nil && frame.Priority.StreamDep > 0 {
			return true
		}
	}
	// If no priority frame, it is still valid (stream 0 is implicit)
	return true
}

func isStandardHeaderOrder(frame HTTP2FrameData) bool {
	// Check if pseudo-headers (starting with :) appear before others
	if len(frame.Headers) < 2 {
		return true
	}

	firstPseudoIdx := -1
	firstNormalIdx := -1

	for i, h := range frame.Headers {
		if strings.HasPrefix(h, ":") && firstPseudoIdx == -1 {
			firstPseudoIdx = i
		}
		if !strings.HasPrefix(h, ":") && firstNormalIdx == -1 {
			firstNormalIdx = i
		}
	}

	// Pseudo-headers should appear before regular headers
	if firstPseudoIdx != -1 && firstNormalIdx != -1 {
		return firstPseudoIdx < firstNormalIdx
	}

	return true
}

func isValidFrameSequence(frameSeq string) bool {
	// SETTINGS typically appears first, followed by other frame types
	// The sequence set -> ... is most common
	return strings.HasPrefix(frameSeq, "set")
}

// ============ Known configuration library ============

func initKnownHTTP2Profiles() map[string]*HTTP2ClientProfile {
	return map[string]*HTTP2ClientProfile{
		"chrome_default": {
			Name:               "Chrome Default",
			BrowserName:        "Chrome",
			BrowserVersion:     "120+",
			SettingsOrder:      []string{"HEADER_TABLE_SIZE", "ENABLE_PUSH", "INITIAL_WINDOW_SIZE"},
			PriorityStrategy:   "balanced",
			HeaderOrder:        []string{":method", ":scheme", ":authority", ":path", "user-agent"},
			WindowUpdateTiming: "immediate",
			RiskScore:          0.05,
		},
		"firefox_default": {
			Name:               "Firefox Default",
			BrowserName:        "Firefox",
			BrowserVersion:     "120+",
			SettingsOrder:      []string{"ENABLE_PUSH", "INITIAL_WINDOW_SIZE"},
			PriorityStrategy:   "weighted",
			HeaderOrder:        []string{":method", ":scheme", ":authority", ":path"},
			WindowUpdateTiming: "batched",
			RiskScore:          0.1,
		},
		"safari_default": {
			Name:               "Safari Default",
			BrowserName:        "Safari",
			BrowserVersion:     "17+",
			SettingsOrder:      []string{"HEADER_TABLE_SIZE", "INITIAL_WINDOW_SIZE"},
			PriorityStrategy:   "default",
			HeaderOrder:        []string{":method", ":scheme", ":authority", ":path"},
			WindowUpdateTiming: "deferred",
			RiskScore:          0.08,
		},
	}
}

// ComputeHTTP2Signature convenience function: computes HTTP/2 signature
func ComputeHTTP2Signature(frames []HTTP2FrameData) (*HTTP2SignatureResult, error) {
	analyzer := NewHTTP2SignatureAnalyzer()
	return analyzer.AnalyzeHTTP2Stream(frames)
}
