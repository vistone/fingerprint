package websocket

import (
	"crypto/sha1"
	"encoding/base64"
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
