package ja4s

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// JA4SResult JA4S fingerprint result
type JA4SResult struct {
	// Complete JA4S SHA256 hash
	Hash string

	// JA4S_a: TLS version + cipher suite + extension features
	JA4Sa string

	// JA4S_r: raw-order version (unsorted extension list)
	JA4Sr string

	// Complete signature string (for debugging)
	RawString string

	// TLS version string (e.g. "1.2", "1.3")
	TLSVersion string

	// Anomaly score (0.0-1.0)
	RiskScore float64

	// Anomaly flag list
	AnomalyFlags []string

	// Matched known server fingerprints
	MatchedProfiles []string
}

// JA4SAnalyzer JA4S analyzer
type JA4SAnalyzer struct {
	// Known server fingerprint library (extensible)
	knownProfiles map[string]*ServerProfileInfo
}

// ServerProfileInfo known server profile information
type ServerProfileInfo struct {
	Name        string   // Server name
	TLSVersions []string // Supported TLS versions
	Ciphers     []string // Common cipher suites
	Extensions  []string // Common extensions
	RiskScore   float64  // Baseline risk score
}

// NewJA4SAnalyzer creates JA4S analyzer
func NewJA4SAnalyzer() *JA4SAnalyzer {
	return &JA4SAnalyzer{
		knownProfiles: initKnownServerProfiles(),
	}
}

// ServerHelloData exported ServerHello structure for computing JA4S from structured data
type ServerHelloData struct {
	TLSVersion   uint16   // TLS version (e.g. 0x0303=TLS1.2, 0x0304=TLS1.3)
	CipherSuite  uint16   // Selected cipher suite
	Extensions   []uint16 // Extension list
	Compression  uint8    // Compression method
	ServerName   string   // Server name
	SelectedALPN string   // Selected ALPN protocol
}

// AnalyzeServerHello analyzes fingerprint from structured ServerHello data
func (a *JA4SAnalyzer) AnalyzeServerHello(data ServerHelloData) (*JA4SResult, error) {
	sh := &serverHelloData{
		Version:           data.TLSVersion,
		CipherSuite:       data.CipherSuite,
		CompressionMethod: data.Compression,
		Extensions:        data.Extensions,
	}

	result := &JA4SResult{
		RawString:    sh.String(),
		TLSVersion:   tlsVersionString(sh.Version),
		AnomalyFlags: make([]string, 0, 8),
	}

	tlsVersionCode := formatTLSVersion(sh.Version)
	cipherCode := formatCipherCode(sh.CipherSuite)
	extensionCount := fmt.Sprintf("%d", len(sh.Extensions))
	compressionCode := formatCompressionCode(sh.CompressionMethod)

	result.JA4Sa = fmt.Sprintf("%s,%s,%s,%s",
		tlsVersionCode,
		cipherCode,
		extensionCount,
		compressionCode,
	)

	var extensionCodes []string
	for _, ext := range sh.Extensions {
		extensionCodes = append(extensionCodes, fmt.Sprintf("%d", ext))
	}
	result.JA4Sr = fmt.Sprintf("%s,%s",
		result.JA4Sa,
		strings.Join(extensionCodes, ","),
	)

	hash := sha256.Sum256([]byte(result.JA4Sr))
	result.Hash = hex.EncodeToString(hash[:])

	a.detectAnomalies(result, sh)

	return result, nil
}

// AnalyzeServerHelloBytes analyzes TLS ServerHello packet
// serverHelloBytes: full ServerHello byte data
func (a *JA4SAnalyzer) AnalyzeServerHelloBytes(serverHelloBytes []byte) (*JA4SResult, error) {
	if len(serverHelloBytes) < 43 {
		return nil, fmt.Errorf("ServerHello too short: %d bytes", len(serverHelloBytes))
	}

	// Parse ServerHello structure
	sh, err := parseServerHello(serverHelloBytes)
	if err != nil {
		return nil, err
	}

	// Build signature string
	result := &JA4SResult{
		RawString:    sh.String(),
		TLSVersion:   tlsVersionString(sh.Version),
		AnomalyFlags: make([]string, 0, 8), // Pre-allocate capacity
	}

	// Generate JA4S_a: TLS version, cipher suite, extension count, compression method
	tlsVersionCode := formatTLSVersion(sh.Version)
	cipherCode := formatCipherCode(sh.CipherSuite)
	extensionCount := fmt.Sprintf("%d", len(sh.Extensions))
	compressionCode := formatCompressionCode(sh.CompressionMethod)

	result.JA4Sa = fmt.Sprintf("%s,%s,%s,%s",
		tlsVersionCode,
		cipherCode,
		extensionCount,
		compressionCode,
	)

	// Generate JA4S_r: raw extension list (unsorted, keep order)
	var extensionCodes []string
	for _, ext := range sh.Extensions {
		extensionCodes = append(extensionCodes, fmt.Sprintf("%d", ext))
	}
	result.JA4Sr = fmt.Sprintf("%s,%s",
		result.JA4Sa,
		strings.Join(extensionCodes, ","),
	)

	// Calculate SHA256 hash
	hash := sha256.Sum256([]byte(result.JA4Sr))
	result.Hash = hex.EncodeToString(hash[:])

	// Anomaly detection and scoring
	a.detectAnomalies(result, sh)

	return result, nil
}

// AnalyzeServerHelloProfile analyzes from fingerprint profile (for client simulation)
// Used to generate virtual ServerHello highly consistent with real servers
func (a *JA4SAnalyzer) GenerateServerHelloSignature(
	tlsVersion uint16,
	cipherSuite uint16,
	extensions []uint16,
	compressionMethod uint8,
) (*JA4SResult, error) {

	// Build virtual ServerHello structure
	sh := &serverHelloData{
		Version:           tlsVersion,
		CipherSuite:       cipherSuite,
		CompressionMethod: compressionMethod,
		Extensions:        extensions,
	}

	result := &JA4SResult{
		RawString:    sh.String(),
		TLSVersion:   tlsVersionString(sh.Version),
		AnomalyFlags: make([]string, 0, 8), // Pre-allocate capacity
	}
	tlsVersionCode := formatTLSVersion(sh.Version)
	cipherCode := formatCipherCode(sh.CipherSuite)
	extensionCount := fmt.Sprintf("%d", len(sh.Extensions))
	compressionCode := formatCompressionCode(sh.CompressionMethod)

	result.JA4Sa = fmt.Sprintf("%s,%s,%s,%s",
		tlsVersionCode,
		cipherCode,
		extensionCount,
		compressionCode,
	)

	var extensionCodes []string
	for _, ext := range sh.Extensions {
		extensionCodes = append(extensionCodes, fmt.Sprintf("%d", ext))
	}
	result.JA4Sr = fmt.Sprintf("%s,%s",
		result.JA4Sa,
		strings.Join(extensionCodes, ","),
	)

	hash := sha256.Sum256([]byte(result.JA4Sr))
	result.Hash = hex.EncodeToString(hash[:])

	a.detectAnomalies(result, sh)

	return result, nil
}

// detectAnomalies detects anomaly characteristics
func (a *JA4SAnalyzer) detectAnomalies(result *JA4SResult, sh *serverHelloData) {
	baseScore := 0.0

	// Anomaly check 1: TLS version check
	if isDeprecatedTLSVersion(sh.Version) {
		// TLS 1.0/1.1/SSL 3.0 are deprecated (RFC 8996) and pose security risks
		result.AnomalyFlags = append(result.AnomalyFlags, "DEPRECATED_TLS_VERSION")
		baseScore += 0.2
	} else if !isSupportedTLSVersion(sh.Version) {
		result.AnomalyFlags = append(result.AnomalyFlags, "UNSUPPORTED_TLS_VERSION")
		baseScore += 0.3
	}

	// Anomaly check 2: known weak cipher suites
	if isWeakCipherSuite(sh.CipherSuite) {
		result.AnomalyFlags = append(result.AnomalyFlags, "WEAK_CIPHER_SUITE")
		baseScore += 0.25
	}

	// Anomaly check 3: abnormal extension combinations
	if len(sh.Extensions) < 3 {
		result.AnomalyFlags = append(result.AnomalyFlags, "MINIMAL_EXTENSIONS")
		baseScore += 0.2
	}
	if len(sh.Extensions) > 30 {
		result.AnomalyFlags = append(result.AnomalyFlags, "EXCESSIVE_EXTENSIONS")
		baseScore += 0.15
	}

	// Anomaly check 4: extension list anomalies (detect duplicate extensions)
	if !hasValidExtensionOrder(sh.Extensions) {
		result.AnomalyFlags = append(result.AnomalyFlags, "DUPLICATE_EXTENSIONS")
		baseScore += 0.2
	}

	// Anomaly check 5: compression method anomaly (TLS compression has CRIME risk; only null=0 is safe)
	if sh.CompressionMethod != 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "UNSAFE_COMPRESSION")
		baseScore += 0.2
	}

	// Normalize score
	if baseScore > 1.0 {
		baseScore = 1.0
	}
	result.RiskScore = baseScore
}

// FindMatchingProfiles finds matching known server profiles
func (a *JA4SAnalyzer) FindMatchingProfiles(result *JA4SResult, maxResults int) []string {
	// Simple hash matching (can be extended to similarity matching)
	var matches []string

	for name, profile := range a.knownProfiles {
		// Calculate similarity (currently simple matching, can be improved)
		if profile.RiskScore < result.RiskScore-0.1 {
			matches = append(matches, name)
		}

		if len(matches) >= maxResults {
			break
		}
	}

	result.MatchedProfiles = matches
	return matches
}

// ============ Helper functions ============
