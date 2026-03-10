package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

// WebSocketFingerprint WebSocket fingerprint
type WebSocketFingerprint struct {
	// Protocol version
	Version string

	// Handshake characteristics
	Handshake WebSocketHandshake

	// Frame characteristics
	FrameCharacteristics FrameCharacteristics

	// Extension support
	Extensions []string

	// Sub-protocols
	SubProtocols []string

	// Raw fingerprint string
	Raw string

	// Hash value
	Hash string
}

// WebSocketHandshake handshake characteristics
type WebSocketHandshake struct {
	// HTTP version
	HTTPVersion string

	// Method (should be GET)
	Method string

	// Header order
	HeaderOrder []string

	// Specific header values
	Headers map[string]string

	// User-Agent
	UserAgent string

	// Origin
	Origin string

	// Sec-WebSocket-Version
	SecWebSocketVersion string

	// Sec-WebSocket-Key characteristics
	SecWebSocketKeyCharacteristics KeyCharacteristics

	// Sec-WebSocket-Extensions
	SecWebSocketExtensions string

	// Sec-WebSocket-Protocol
	SecWebSocketProtocol string
}

// KeyCharacteristics WebSocket Key characteristics
type KeyCharacteristics struct {
	// Key length
	Length int

	// Whether standard Base64 (16-byte random data)
	IsStandardBase64 bool

	// Whether contains specific pattern
	HasPattern bool

	// Pattern type
	PatternType string

	// Entropy estimate
	Entropy float64
}

// FrameCharacteristics frame characteristics
type FrameCharacteristics struct {
	// Masking behavior
	MaskingBehavior string

	// Maximum frame size
	MaxFrameSize int

	// Supported opcodes
	SupportedOpcodes []uint8

	// Extension flag usage
	ExtensionFlags []string

	// Control frame behavior
	ControlFrameBehavior string
}

// Analyzer WebSocket fingerprint analyzer
type Analyzer struct {
	// Browser pattern database
	browserPatterns map[string]BrowserPattern
}

// BrowserPattern browser pattern signature
type BrowserPattern struct {
	Name           string
	Version        string
	HeaderOrder    []string
	KeyPattern     string
	Extensions     []string
	UserAgentMatch *regexp.Regexp
}

// NewAnalyzer creates analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		browserPatterns: initBrowserPatterns(),
	}
}

// AnalyzeRequest analyzes WebSocket handshake request
func (a *Analyzer) AnalyzeRequest(req *http.Request) (*WebSocketFingerprint, error) {
	if req.Method != "GET" {
		return nil, fmt.Errorf("invalid method: %s", req.Method)
	}

	fp := &WebSocketFingerprint{
		Version: "13", // RFC 6455
		Handshake: WebSocketHandshake{
			HTTPVersion: req.Proto,
			Method:      req.Method,
			Headers:     make(map[string]string),
		},
	}

	// Analyze headers
	a.analyzeHeaders(req, fp)

	// Analyze Sec-WebSocket-Key
	a.analyzeSecWebSocketKey(req, fp)

	// Analyze extensions
	a.analyzeExtensions(req, fp)

	// Analyze sub-protocols
	a.analyzeSubProtocols(req, fp)

	// Generate fingerprint string and hash
	fp.Raw = a.generateFingerprintString(fp)
	fp.Hash = a.generateFingerprintHash(fp.Raw)

	return fp, nil
}

// analyzeHeaders analyzes HTTP headers
func (a *Analyzer) analyzeHeaders(req *http.Request, fp *WebSocketFingerprint) {
	// Record header order
	for name := range req.Header {
		fp.Handshake.HeaderOrder = append(fp.Handshake.HeaderOrder, name)
	}

	// Record specific headers
	if ua := req.Header.Get("User-Agent"); ua != "" {
		fp.Handshake.UserAgent = ua
	}

	if origin := req.Header.Get("Origin"); origin != "" {
		fp.Handshake.Origin = origin
	}

	if version := req.Header.Get("Sec-Websocket-Version"); version != "" {
		fp.Handshake.SecWebSocketVersion = version
	}

	if ext := req.Header.Get("Sec-Websocket-Extensions"); ext != "" {
		fp.Handshake.SecWebSocketExtensions = ext
	}

	if proto := req.Header.Get("Sec-Websocket-Protocol"); proto != "" {
		fp.Handshake.SecWebSocketProtocol = proto
	}

	// Record all Sec-WebSocket-* headers
	for name, values := range req.Header {
		if strings.HasPrefix(name, "Sec-Websocket-") || strings.HasPrefix(name, "Sec-WebSocket-") {
			fp.Handshake.Headers[name] = strings.Join(values, ", ")
		}
	}
}

// analyzeSecWebSocketKey analyzes Sec-WebSocket-Key
func (a *Analyzer) analyzeSecWebSocketKey(req *http.Request, fp *WebSocketFingerprint) {
	key := req.Header.Get("Sec-Websocket-Key")
	if key == "" {
		return
	}

	fp.Handshake.SecWebSocketKeyCharacteristics = KeyCharacteristics{
		Length: len(key),
	}

	// Decode Base64
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		fp.Handshake.SecWebSocketKeyCharacteristics.IsStandardBase64 = false
		return
	}

	// Standard should be 16-byte random data
	fp.Handshake.SecWebSocketKeyCharacteristics.IsStandardBase64 = len(decoded) == 16

	// Analyze entropy (simplified)
	fp.Handshake.SecWebSocketKeyCharacteristics.Entropy = calculateEntropy(decoded)

	// Detect pattern
	a.detectKeyPattern(decoded, fp)
}

// detectKeyPattern detects Key generation pattern
func (a *Analyzer) detectKeyPattern(key []byte, fp *WebSocketFingerprint) {
	// Detect common random number generator patterns
	if len(key) != 16 {
		return
	}

	// Detect if all zeros (test client)
	allZero := true
	for _, b := range key {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		fp.Handshake.SecWebSocketKeyCharacteristics.HasPattern = true
		fp.Handshake.SecWebSocketKeyCharacteristics.PatternType = "all_zeros"
		return
	}

	// Detect incremental pattern (simple RNG)
	incremental := true
	for i := 1; i < len(key); i++ {
		if key[i] != key[i-1]+1 {
			incremental = false
			break
		}
	}
	if incremental {
		fp.Handshake.SecWebSocketKeyCharacteristics.HasPattern = true
		fp.Handshake.SecWebSocketKeyCharacteristics.PatternType = "incremental"
		return
	}

	// Detect low entropy (some embedded devices)
	if fp.Handshake.SecWebSocketKeyCharacteristics.Entropy < 3.0 {
		fp.Handshake.SecWebSocketKeyCharacteristics.HasPattern = true
		fp.Handshake.SecWebSocketKeyCharacteristics.PatternType = "low_entropy"
	}
}

// analyzeExtensions analyzes extensions
func (a *Analyzer) analyzeExtensions(req *http.Request, fp *WebSocketFingerprint) {
	extHeader := req.Header.Get("Sec-Websocket-Extensions")
	if extHeader == "" {
		return
	}

	// Parse extension list
	// Format: extension1; param1=value1, extension2; param2=value2
	extensions := parseExtensionHeader(extHeader)
	fp.Extensions = extensions

	// Record frame characteristics
	for _, ext := range extensions {
		switch ext {
		case "permessage-deflate":
			fp.FrameCharacteristics.ExtensionFlags = append(fp.FrameCharacteristics.ExtensionFlags, "compression")
		case "client_max_window_bits":
			fp.FrameCharacteristics.ExtensionFlags = append(fp.FrameCharacteristics.ExtensionFlags, "client_max_window_bits")
		case "server_max_window_bits":
			fp.FrameCharacteristics.ExtensionFlags = append(fp.FrameCharacteristics.ExtensionFlags, "server_max_window_bits")
		}
	}
}

// analyzeSubProtocols analyzes sub-protocols
func (a *Analyzer) analyzeSubProtocols(req *http.Request, fp *WebSocketFingerprint) {
	protoHeader := req.Header.Get("Sec-Websocket-Protocol")
	if protoHeader == "" {
		return
	}

	// Parse protocol list
	protocols := splitAndTrim(protoHeader, ",")
	fp.SubProtocols = protocols
}

// generateFingerprintString generates fingerprint string
func (a *Analyzer) generateFingerprintString(fp *WebSocketFingerprint) string {
	var parts []string

	// HTTP version
	parts = append(parts, fmt.Sprintf("http=%s", fp.Handshake.HTTPVersion))

	// Header order (sorted)
	sortedHeaders := make([]string, len(fp.Handshake.HeaderOrder))
	copy(sortedHeaders, fp.Handshake.HeaderOrder)
	sort.Strings(sortedHeaders)
	parts = append(parts, fmt.Sprintf("headers=%s", strings.Join(sortedHeaders, "|")))

	// WebSocket version
	parts = append(parts, fmt.Sprintf("ws_version=%s", fp.Handshake.SecWebSocketVersion))

	// Key characteristics
	keyChar := fp.Handshake.SecWebSocketKeyCharacteristics
	parts = append(parts, fmt.Sprintf("key_std=%v", keyChar.IsStandardBase64))

	// Extensions
	if len(fp.Extensions) > 0 {
		sort.Strings(fp.Extensions)
		parts = append(parts, fmt.Sprintf("ext=%s", strings.Join(fp.Extensions, ",")))
	}

	return strings.Join(parts, ";")
}

// generateFingerprintHash generates fingerprint hash
func (a *Analyzer) generateFingerprintHash(raw string) string {
	h := sha1.New()
	h.Write([]byte(raw))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// IdentifyBrowser identifies browser type
func (a *Analyzer) IdentifyBrowser(fp *WebSocketFingerprint) (string, float64) {
	if fp.Handshake.UserAgent == "" {
		return "unknown", 0.0
	}

	for name, pattern := range a.browserPatterns {
		if pattern.UserAgentMatch.MatchString(fp.Handshake.UserAgent) {
			// Further validate header order
			if a.matchHeaderOrder(fp, pattern) {
				return name, 0.8
			}
			return name, 0.5
		}
	}

	return "unknown", 0.0
}

// matchHeaderOrder validates whether header order matches
func (a *Analyzer) matchHeaderOrder(fp *WebSocketFingerprint, pattern BrowserPattern) bool {
	if len(pattern.HeaderOrder) == 0 {
		return true
	}

	// Simplified matching: check if first few key headers match
	matchCount := 0
	for i, header := range pattern.HeaderOrder {
		if i < len(fp.Handshake.HeaderOrder) &&
			equalIgnoreCase(fp.Handshake.HeaderOrder[i], header) {
			matchCount++
		}
	}

	return float64(matchCount)/float64(len(pattern.HeaderOrder)) > 0.7
}

// parseExtensionHeader parses extension header
func parseExtensionHeader(header string) []string {
	var extensions []string

	parts := splitAndTrim(header, ",")
	for _, part := range parts {
		// Extract extension name (remove parameters)
		if idx := strings.Index(part, ";"); idx != -1 {
			extensions = append(extensions, strings.TrimSpace(part[:idx]))
		} else {
			extensions = append(extensions, strings.TrimSpace(part))
		}
	}

	return extensions
}

// splitAndTrim splits string and trims whitespace
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// equalIgnoreCase compares ignoring case
func equalIgnoreCase(a, b string) bool {
	return strings.EqualFold(a, b)
}

// calculateEntropy calculates entropy of byte array
func calculateEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0.0
	}

	// Count byte frequency
	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}

	// Calculate entropy
	var entropy float64
	length := float64(len(data))
	for _, count := range freq {
		p := float64(count) / length
		if p > 0 {
			entropy -= p * mathLog2(p)
		}
	}

	return entropy
}

// mathLog2 calculates log2
func mathLog2(x float64) float64 {
	// Simplified implementation, use natural logarithm conversion
	const ln2 = 0.6931471805599453
	if x <= 0 {
		return 0
	}
	// Use math package
	return log2(x)
}

// log2 calculates log2 (simplified implementation)
func log2(x float64) float64 {
	// Use change of base formula: log2(x) = log(x) / log(2)
	// Use approximation algorithm here
	result := 0.0
	for x >= 2.0 {
		x /= 2.0
		result++
	}
	if x > 1.0 {
		result += x - 1.0 // Approximation
	}
	return result
}

// initBrowserPatterns initializes browser patterns
func initBrowserPatterns() map[string]BrowserPattern {
	return map[string]BrowserPattern{
		"chrome": {
			Name:    "Chrome",
			Version: "*",
			HeaderOrder: []string{
				"Host",
				"Connection",
				"Upgrade",
				"Origin",
				"Sec-WebSocket-Version",
				"Sec-WebSocket-Key",
				"Sec-WebSocket-Extensions",
			},
			UserAgentMatch: regexp.MustCompile(`Chrome/[\d.]+`),
		},
		"firefox": {
			Name:    "Firefox",
			Version: "*",
			HeaderOrder: []string{
				"Host",
				"User-Agent",
				"Accept",
				"Accept-Language",
				"Accept-Encoding",
				"Sec-WebSocket-Version",
				"Origin",
				"Sec-WebSocket-Extensions",
				"Sec-WebSocket-Key",
			},
			UserAgentMatch: regexp.MustCompile(`Firefox/[\d.]+`),
		},
		"safari": {
			Name:    "Safari",
			Version: "*",
			HeaderOrder: []string{
				"Host",
				"Upgrade",
				"Connection",
				"Sec-WebSocket-Key",
				"Sec-WebSocket-Version",
				"Origin",
			},
			UserAgentMatch: regexp.MustCompile(`Safari/[\d.]+`),
		},
	}
}

// CompareFingerprints compares similarity of two fingerprints
func CompareFingerprints(fp1, fp2 *WebSocketFingerprint) float64 {
	if fp1 == nil || fp2 == nil {
		return 0.0
	}

	score := 0.0
	weight := 0.0

	// Compare HTTP version
	weight++
	if fp1.Handshake.HTTPVersion == fp2.Handshake.HTTPVersion {
		score++
	}

	// Compare header order (Jaccard similarity)
	weight++
	score += jaccardSimilarity(fp1.Handshake.HeaderOrder, fp2.Handshake.HeaderOrder)

	// Compare extensions
	weight++
	score += jaccardSimilarity(fp1.Extensions, fp2.Extensions)

	// Compare Key standardness
	weight++
	if fp1.Handshake.SecWebSocketKeyCharacteristics.IsStandardBase64 ==
		fp2.Handshake.SecWebSocketKeyCharacteristics.IsStandardBase64 {
		score++
	}

	return score / weight
}

// jaccardSimilarity calculates Jaccard similarity
func jaccardSimilarity(a, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Create sets
	setA := make(map[string]bool)
	for _, s := range a {
		setA[strings.ToLower(s)] = true
	}

	setB := make(map[string]bool)
	for _, s := range b {
		setB[strings.ToLower(s)] = true
	}

	// Calculate intersection and union
	intersection := 0
	for k := range setA {
		if setB[k] {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection

	return float64(intersection) / float64(union)
}

// GenerateAcceptKey generates Sec-WebSocket-Accept response
func GenerateAcceptKey(key string) (string, error) {
	const magicGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

	h := sha1.New()
	h.Write([]byte(key))
	h.Write([]byte(magicGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// IsValidWebSocketRequest validates whether valid WebSocket request
func IsValidWebSocketRequest(req *http.Request) bool {
	// Check required headers
	if req.Method != "GET" {
		return false
	}

	if req.Header.Get("Upgrade") != "websocket" {
		return false
	}

	if !strings.Contains(req.Header.Get("Connection"), "Upgrade") {
		return false
	}

	if req.Header.Get("Sec-Websocket-Key") == "" {
		return false
	}

	if req.Header.Get("Sec-Websocket-Version") != "13" {
		return false
	}

	return true
}

// Frame WebSocket frame structure
type Frame struct {
	// FIN flag
	FIN bool

	// RSV flags
	RSV1, RSV2, RSV3 bool

	// Opcode
	Opcode uint8

	// MASK flag
	MASK bool

	// Payload length
	PayloadLength uint64

	// Masking key (client-sent frame)
	MaskingKey [4]byte

	// Payload data
	Payload []byte
}

// ParseFrame parses WebSocket frame
func ParseFrame(data []byte) (*Frame, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("frame too short")
	}

	frame := &Frame{}

	// First byte: FIN, RSV, Opcode
	frame.FIN = data[0]&0x80 != 0
	frame.RSV1 = data[0]&0x40 != 0
	frame.RSV2 = data[0]&0x20 != 0
	frame.RSV3 = data[0]&0x10 != 0
	frame.Opcode = data[0] & 0x0f

	// Second byte: MASK, Payload length
	frame.MASK = data[1]&0x80 != 0
	length := uint64(data[1] & 0x7f)

	offset := 2

	// Extensions length
	if length == 126 {
		if len(data) < 4 {
			return nil, fmt.Errorf("frame truncated at extended length")
		}
		length = uint64(binary.BigEndian.Uint16(data[2:4]))
		offset = 4
	} else if length == 127 {
		if len(data) < 10 {
			return nil, fmt.Errorf("frame truncated at extended length")
		}
		length = binary.BigEndian.Uint64(data[2:10])
		offset = 10
	}

	frame.PayloadLength = length

	// Masking key
	if frame.MASK {
		if len(data) < offset+4 {
			return nil, fmt.Errorf("frame truncated at masking key")
		}
		copy(frame.MaskingKey[:], data[offset:offset+4])
		offset += 4
	}

	// Payload
	if uint64(len(data)-offset) < length {
		return nil, fmt.Errorf("frame truncated at payload")
	}

	frame.Payload = data[offset : offset+int(length)]

	// Unmask
	if frame.MASK {
		frame.Payload = unmaskPayload(frame.Payload, frame.MaskingKey)
	}

	return frame, nil
}

// unmaskPayload unmasks payload
func unmaskPayload(payload []byte, key [4]byte) []byte {
	result := make([]byte, len(payload))
	for i, b := range payload {
		result[i] = b ^ key[i%4]
	}
	return result
}

// FrameOpCode frame opcode constants
const (
	OpCodeContinuation uint8 = 0x0
	OpCodeText         uint8 = 0x1
	OpCodeBinary       uint8 = 0x2
	OpCodeClose        uint8 = 0x8
	OpCodePing         uint8 = 0x9
	OpCodePong         uint8 = 0xa
)

// AnalyzeFrame analyzes single frame characteristics
func AnalyzeFrame(frame *Frame) map[string]interface{} {
	features := make(map[string]interface{})

	features["opcode"] = frame.Opcode
	features["fin"] = frame.FIN
	features["mask"] = frame.MASK
	features["payload_length"] = frame.PayloadLength
	features["has_rsv"] = frame.RSV1 || frame.RSV2 || frame.RSV3

	// Frame type
	switch frame.Opcode {
	case OpCodeText:
		features["frame_type"] = "text"
	case OpCodeBinary:
		features["frame_type"] = "binary"
	case OpCodeClose:
		features["frame_type"] = "close"
	case OpCodePing:
		features["frame_type"] = "ping"
	case OpCodePong:
		features["frame_type"] = "pong"
	case OpCodeContinuation:
		features["frame_type"] = "continuation"
	default:
		features["frame_type"] = "unknown"
	}

	return features
}

// AnalyzeFrameStream analyzes frame stream characteristics
func AnalyzeFrameStream(frames []*Frame) map[string]interface{} {
	features := make(map[string]interface{})

	if len(frames) == 0 {
		return features
	}

	// Statistics
	var (
		maskedCount   int
		unmaskedCount int
		textCount     int
		binaryCount   int
		controlCount  int
		maxPayload    uint64
		totalPayload  uint64
	)

	for _, frame := range frames {
		if frame.MASK {
			maskedCount++
		} else {
			unmaskedCount++
		}

		switch frame.Opcode {
		case OpCodeText:
			textCount++
		case OpCodeBinary:
			binaryCount++
		case OpCodeClose, OpCodePing, OpCodePong:
			controlCount++
		}

		if frame.PayloadLength > maxPayload {
			maxPayload = frame.PayloadLength
		}
		totalPayload += frame.PayloadLength
	}

	features["total_frames"] = len(frames)
	features["masked_ratio"] = float64(maskedCount) / float64(len(frames))
	features["text_ratio"] = float64(textCount) / float64(len(frames))
	features["binary_ratio"] = float64(binaryCount) / float64(len(frames))
	features["control_ratio"] = float64(controlCount) / float64(len(frames))
	features["max_payload"] = maxPayload
	features["avg_payload"] = float64(totalPayload) / float64(len(frames))

	return features
}
