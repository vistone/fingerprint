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
	TLSVersion     uint16   // TLS version (e.g. 0x0303=TLS1.2, 0x0304=TLS1.3)
	CipherSuite    uint16   // Selected cipher suite
	Extensions     []uint16 // Extension list
	Compression    uint8    // Compression method
	ServerName     string   // Server name
	SelectedALPN   string   // Selected ALPN protocol
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

type serverHelloData struct {
	Version           uint16
	CipherSuite       uint16
	CompressionMethod uint8
	Extensions        []uint16
}

func (sh *serverHelloData) String() string {
	var extStr []string
	for _, e := range sh.Extensions {
		extStr = append(extStr, fmt.Sprintf("%d", e))
	}
	return fmt.Sprintf("TLS%x,Cipher%x,Comp%d,Ext[%s]",
		sh.Version,
		sh.CipherSuite,
		sh.CompressionMethod,
		strings.Join(extStr, ","),
	)
}

// parseServerHello parses ServerHello byte data
func parseServerHello(data []byte) (*serverHelloData, error) {
	if len(data) < 43 {
		return nil, fmt.Errorf("data too short")
	}

	sh := &serverHelloData{}

	// Offset: HandshakeType(1) + Length(3) = 4-byte header
	// Version(2) at offset 4-5
	sh.Version = uint16(data[4])<<8 | uint16(data[5])

	// Random(32) at offset 6-37
	// Session ID Length(1) at offset 38
	sessionIDLen := int(data[38])
	offset := 39 + sessionIDLen

	// Cipher Suite(2) follows Session ID
	if len(data) < offset+3 {
		return nil, fmt.Errorf("data too short for cipher suite: need %d, have %d", offset+3, len(data))
	}
	sh.CipherSuite = uint16(data[offset])<<8 | uint16(data[offset+1])
	offset += 2

	// Compression Method(1)
	sh.CompressionMethod = data[offset]
	offset++

	// Parse extension list
	if len(data) > offset+2 {
		extensionsLen := int(data[offset])<<8 | int(data[offset+1])
		offset += 2

		endOffset := offset + extensionsLen
		if endOffset > len(data) {
			endOffset = len(data)
		}

		for offset+4 <= endOffset {
			extType := uint16(data[offset])<<8 | uint16(data[offset+1])
			extLen := int(data[offset+2])<<8 | int(data[offset+3])
			if offset+4+extLen > endOffset {
				break // Extension data truncated, stop parsing
			}
			sh.Extensions = append(sh.Extensions, extType)
			offset += 4 + extLen
		}
	}

	return sh, nil
}

func formatTLSVersion(v uint16) string {
	switch v {
	case 0x0303:
		return "773" // TLS 1.2
	case 0x0304:
		return "774" // TLS 1.3
	default:
		return fmt.Sprintf("%d", v)
	}
}

func tlsVersionString(v uint16) string {
	switch v {
	case 0x0301:
		return "1.0"
	case 0x0302:
		return "1.1"
	case 0x0303:
		return "1.2"
	case 0x0304:
		return "1.3"
	default:
		return fmt.Sprintf("0x%04x", v)
	}
}

func formatCipherCode(cipher uint16) string {
	// Return simplified cipher suite code
	// Can be extended to a full mapping table
	switch cipher {
	case 0x002f:
		return "1" // TLS_RSA_WITH_AES_128_CBC_SHA
	case 0x007c:
		return "2" // TLS_RSA_WITH_AES_256_CBC_SHA
	case 0x1301:
		return "3" // TLS_AES_128_GCM_SHA256
	case 0x1302:
		return "4" // TLS_AES_256_GCM_SHA384
	default:
		// Return raw value as fallback
		return fmt.Sprintf("%d", cipher)
	}
}

func formatCompressionCode(c uint8) string {
	if c == 0 {
		return "0" // null compression
	}
	return fmt.Sprintf("%d", c)
}

func isSupportedTLSVersion(v uint16) bool {
	supportedVersions := map[uint16]bool{
		0x0303: true, // TLS 1.2
		0x0304: true, // TLS 1.3
	}
	return supportedVersions[v]
}

// isDeprecatedTLSVersion checks whether TLS version is deprecated (RFC 8996)
func isDeprecatedTLSVersion(v uint16) bool {
	deprecatedVersions := map[uint16]bool{
		0x0300: true, // SSL 3.0
		0x0301: true, // TLS 1.0
		0x0302: true, // TLS 1.1
	}
	return deprecatedVersions[v]
}

func isWeakCipherSuite(cipher uint16) bool {
	// Known weak ciphers list (RFC 7457, CRIME, BEAST, SWEET32 attack vectors, etc.)
	weakCiphers := map[uint16]bool{
		0x0000: true, // TLS_NULL_WITH_NULL_NULL
		0x0001: true, // TLS_RSA_WITH_NULL_MD5
		0x0002: true, // TLS_RSA_WITH_NULL_SHA
		0x0003: true, // TLS_RSA_EXPORT_WITH_RC4_40_MD5
		0x0004: true, // TLS_RSA_WITH_RC4_128_MD5
		0x0005: true, // TLS_RSA_WITH_RC4_128_SHA
		0x0006: true, // TLS_RSA_EXPORT_WITH_RC2_CBC_40_MD5
		0x0008: true, // TLS_RSA_EXPORT_WITH_DES40_CBC_SHA
		0x0009: true, // TLS_RSA_WITH_DES_CBC_SHA
		0x000A: true, // TLS_RSA_WITH_3DES_EDE_CBC_SHA (SWEET32)
		0x0011: true, // TLS_DHE_DSS_EXPORT_WITH_DES40_CBC_SHA
		0x0012: true, // TLS_DHE_DSS_WITH_DES_CBC_SHA
		0x0014: true, // TLS_DHE_RSA_EXPORT_WITH_DES40_CBC_SHA
		0x0015: true, // TLS_DHE_RSA_WITH_DES_CBC_SHA
		0x0017: true, // TLS_DH_anon_EXPORT_WITH_RC4_40_MD5
		0x0018: true, // TLS_DH_anon_WITH_RC4_128_MD5
		0x0019: true, // TLS_DH_anon_EXPORT_WITH_DES40_CBC_SHA
		0x001A: true, // TLS_DH_anon_WITH_DES_CBC_SHA
		0x003B: true, // TLS_RSA_WITH_NULL_SHA256
		0xC00A: true, // TLS_ECDHE_ECDSA_WITH_3DES_EDE_CBC_SHA (SWEET32)
		0xC014: true, // TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA (SWEET32)
		0xC011: true, // TLS_ECDHE_RSA_WITH_RC4_128_SHA
		0xC007: true, // TLS_ECDHE_ECDSA_WITH_RC4_128_SHA
	}
	return weakCiphers[cipher]
}

func hasValidExtensionOrder(extensions []uint16) bool {
	// ServerHello extensions have no spec-mandated ordering,
	// so anomalies should not be judged by ordering.
	// Only detect duplicate extensions (duplicates are invalid).
	if len(extensions) < 2 {
		return true
	}

	seen := make(map[uint16]bool, len(extensions))
	for _, ext := range extensions {
		if seen[ext] {
			return false // Duplicate extensions are anomalous
		}
		seen[ext] = true
	}
	return true
}

// initKnownServerProfiles initializes known server profile library
func initKnownServerProfiles() map[string]*ServerProfileInfo {
	return map[string]*ServerProfileInfo{
		"nginx_default": {
			Name:        "Nginx (Default)",
			TLSVersions: []string{"TLS 1.2", "TLS 1.3"},
			Ciphers:     []string{"AES_128_GCM", "AES_256_GCM", "CHACHA20"},
			Extensions:  []string{"0", "10", "11", "16", "23", "35"},
			RiskScore:   0.1,
		},
		"apache_default": {
			Name:        "Apache (Default)",
			TLSVersions: []string{"TLS 1.2", "TLS 1.3"},
			Ciphers:     []string{"AES_128_GCM", "AES_256_GCM"},
			Extensions:  []string{"0", "10", "11", "16", "23", "35"},
			RiskScore:   0.15,
		},
		"cloudflare": {
			Name:        "Cloudflare",
			TLSVersions: []string{"TLS 1.3"},
			Ciphers:     []string{"AES_256_GCM", "CHACHA20"},
			Extensions:  []string{"0", "10", "11", "16", "23", "35", "43"},
			RiskScore:   0.05,
		},
	}
}

// ComputeJA4S convenience function: compute JA4S from ServerHelloData structure
func ComputeJA4S(data ServerHelloData) (*JA4SResult, error) {
	analyzer := NewJA4SAnalyzer()
	return analyzer.AnalyzeServerHello(data)
}

// ComputeJA4SFromBytes convenience function: compute JA4S directly from byte data
func ComputeJA4SFromBytes(serverHelloBytes []byte) (*JA4SResult, error) {
	analyzer := NewJA4SAnalyzer()
	return analyzer.AnalyzeServerHelloBytes(serverHelloBytes)
}

// MatchJA4S compares whether two JA4S hashes match
func MatchJA4S(hash1, hash2 string) bool {
	return len(hash1) == 64 && len(hash2) == 64 && hash1 == hash2
}

// ComputeJA4SFromProfileData computes JA4S from profile data (for client simulation)
func ComputeJA4SFromProfileData(
	tlsVersion uint16,
	cipherSuite uint16,
	extensions []uint16,
) (*JA4SResult, error) {
	analyzer := NewJA4SAnalyzer()
	return analyzer.GenerateServerHelloSignature(tlsVersion, cipherSuite, extensions, 0)
}
