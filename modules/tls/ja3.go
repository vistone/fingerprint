// Package tls provides TLS fingerprint generation
package tls

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// JA3Result represents a JA3 fingerprint result
type JA3Result struct {
	// Hash is the JA3 MD5 hash
	Hash string
	// RawString is the JA3 raw fingerprint string
	RawString string
	// TLSVersion is the TLS protocol version
	TLSVersion uint16
	// CipherSuites is the list of cipher suites
	CipherSuites []uint16
	// Extensions is the list of extension type IDs
	Extensions []uint16
	// EllipticCurves is the list of elliptic curves
	EllipticCurves []core.CurveID
	// EllipticCurvePointFormats is the list of elliptic curve point formats
	EllipticCurvePointFormats []uint8
}

// CalculateJA3 computes a JA3 fingerprint from a ClientHello specification
func CalculateJA3(spec core.ClientHelloSpec) *JA3Result {
	// Filter out GREASE values
	cipherSuites := filterGREASEUint16(spec.CipherSuites)
	extensions := filterGREASEExtensions(spec.Extensions)
	curves := filterGREASECurves(spec.SupportedCurves)

	// Build the JA3 string
	parts := []string{
		strconv.Itoa(int(spec.TLSVersion)),
		joinUint16(cipherSuites),
		joinExtensions(extensions),
		joinCurves(curves),
		joinUint8(spec.SupportedPoints),
	}

	rawString := strings.Join(parts, ",")

	// Compute the MD5 hash
	hash := md5.Sum([]byte(rawString))

	return &JA3Result{
		Hash:                      hex.EncodeToString(hash[:]),
		RawString:                 rawString,
		TLSVersion:                spec.TLSVersion,
		CipherSuites:              cipherSuites,
		Extensions:                extensionTypes(extensions),
		EllipticCurves:            curves,
		EllipticCurvePointFormats: spec.SupportedPoints,
	}
}

// CalculateJA3FromProfile computes a JA3 fingerprint from a client profile
func CalculateJA3FromProfile(profile profiles.ClientProfile) *JA3Result {
	spec := core.ClientHelloSpec{
		TLSVersion:      profile.TLSVersion,
		CipherSuites:    profile.CipherSuites,
		Extensions:      profile.Extensions,
		SupportedCurves: profile.SupportedCurves,
		SupportedPoints: profile.SupportedPoints,
	}
	return CalculateJA3(spec)
}

// IsGREASEUint16 checks whether a value is a GREASE value
func IsGREASEUint16(v uint16) bool {
	// GREASE pattern: 0x0A0A, 0x1A1A, 0x2A2A, ..., 0xFAFA
	// High and low bytes are equal, and the low byte has the form 0xXA
	return ((v >> 8) & 0xFF) == (v & 0xFF) && (v&0x0F) == 0x0A
}

// filterGREASEUint16 filters out GREASE values from a uint16 slice
func filterGREASEUint16(values []uint16) []uint16 {
	var result []uint16
	for _, v := range values {
		if !IsGREASEUint16(v) {
			result = append(result, v)
		}
	}
	return result
}

// filterGREASEExtensions filters out GREASE extensions
func filterGREASEExtensions(exts []core.TLSExtension) []core.TLSExtension {
	var result []core.TLSExtension
	for _, e := range exts {
		if !IsGREASEUint16(e.Type) {
			result = append(result, e)
		}
	}
	return result
}

// filterGREASECurves filters out GREASE curves
func filterGREASECurves(curves []core.CurveID) []core.CurveID {
	var result []core.CurveID
	for _, c := range curves {
		if !IsGREASEUint16(uint16(c)) {
			result = append(result, c)
		}
	}
	return result
}

// joinUint16 joins a uint16 slice into a dash-separated string
func joinUint16(values []uint16) string {
	var parts []string
	for _, v := range values {
		parts = append(parts, strconv.Itoa(int(v)))
	}
	return strings.Join(parts, "-")
}

// joinUint8 joins a uint8 slice into a dash-separated string
func joinUint8(values []uint8) string {
	var parts []string
	for _, v := range values {
		parts = append(parts, strconv.Itoa(int(v)))
	}
	return strings.Join(parts, "-")
}

// joinExtensions joins extensions into a dash-separated string of type IDs
func joinExtensions(exts []core.TLSExtension) string {
	var parts []string
	for _, e := range exts {
		parts = append(parts, strconv.Itoa(int(e.Type)))
	}
	return strings.Join(parts, "-")
}

// extensionTypes extracts extension type IDs from a list of extensions
func extensionTypes(exts []core.TLSExtension) []uint16 {
	var result []uint16
	for _, e := range exts {
		result = append(result, e.Type)
	}
	return result
}

// JA4Result represents a JA4 fingerprint result per the JA4+ specification
type JA4Result struct {
	// Fingerprint is the full JA4 fingerprint (JA4_a + "_" + JA4_b + "_" + JA4_c)
	Fingerprint string
	// JA4A is the JA4_a section: protocol type, TLS version, SNI, cipher count, extension count, ALPN
	JA4A string
	// JA4B is the JA4_b section: truncated SHA-256 of sorted cipher suites
	JA4B string
	// JA4C is the JA4_c section: truncated SHA-256 of sorted extensions (excluding SNI and ALPN)
	JA4C string
	// TLSVersion is the TLS protocol version
	TLSVersion uint16
	// CipherSuitesCount is the number of cipher suites (excluding GREASE)
	CipherSuitesCount int
	// ExtensionsCount is the number of extensions (excluding GREASE)
	ExtensionsCount int
	// SNI indicates whether a Server Name Indication extension is present
	SNI bool
	// ALPN is the first ALPN protocol value, if present
	ALPN string
}

const (
	// extTypeSNI is the TLS extension type for Server Name Indication
	extTypeSNI uint16 = 0x0000
	// extTypeALPN is the TLS extension type for Application-Layer Protocol Negotiation
	extTypeALPN uint16 = 0x0010
)

// CalculateJA4 computes a JA4 fingerprint from a ClientHello specification
// following the JA4+ specification with JA4_a, JA4_b, and JA4_c sections.
func CalculateJA4(spec core.ClientHelloSpec) *JA4Result {
	ciphers := filterGREASEUint16(spec.CipherSuites)
	exts := filterGREASEExtensions(spec.Extensions)

	cipherCount := len(ciphers)
	extCount := len(exts)

	// Detect SNI and ALPN from extensions
	hasSNI := false
	alpnValue := ""
	for _, e := range exts {
		switch e.Type {
		case extTypeSNI:
			hasSNI = true
		case extTypeALPN:
			alpnValue = parseALPN(e.Data)
		}
	}

	// JA4_a: {proto}{version}{sni}{cipher_count:2d}{ext_count:2d}_{alpn}
	version := TLSVersionToString(spec.TLSVersion)
	sniChar := "i"
	if hasSNI {
		sniChar = "d"
	}
	alpnChars := alpnToJA4(alpnValue)
	ja4a := fmt.Sprintf("t%s%s%02d%02d_%s", version, sniChar, cipherCount, extCount, alpnChars)

	// JA4_b: first 12 hex chars of SHA-256 of sorted cipher suites (as 4-digit hex, comma-separated)
	ja4b := truncatedSHA256Hex(sortedHexList(ciphers), 12)

	// JA4_c: first 12 hex chars of SHA-256 of sorted extensions (excluding SNI and ALPN)
	ja4cExtTypes := make([]uint16, 0, len(exts))
	for _, e := range exts {
		if e.Type != extTypeSNI && e.Type != extTypeALPN {
			ja4cExtTypes = append(ja4cExtTypes, e.Type)
		}
	}
	ja4c := truncatedSHA256Hex(sortedHexList(ja4cExtTypes), 12)

	fingerprint := ja4a + "_" + ja4b + "_" + ja4c

	return &JA4Result{
		Fingerprint:       fingerprint,
		JA4A:              ja4a,
		JA4B:              ja4b,
		JA4C:              ja4c,
		TLSVersion:        spec.TLSVersion,
		CipherSuitesCount: cipherCount,
		ExtensionsCount:   extCount,
		SNI:               hasSNI,
		ALPN:              alpnValue,
	}
}

// parseALPN extracts the first ALPN protocol name from raw extension data.
// The ALPN extension data format is: 2-byte list length, then for each
// protocol: 1-byte name length followed by the protocol name bytes.
func parseALPN(data []byte) string {
	if len(data) < 4 {
		return ""
	}
	protoLen := int(data[2])
	if len(data) < 3+protoLen {
		return ""
	}
	return string(data[3 : 3+protoLen])
}

// alpnToJA4 converts an ALPN protocol name to its 2-character JA4 representation
// using the first and last characters of the protocol name.
func alpnToJA4(alpn string) string {
	if alpn == "" {
		return "00"
	}
	return alpn[0:1] + alpn[len(alpn)-1:]
}

// sortedHexList converts a uint16 slice to sorted 4-character lowercase hex
// strings and returns them as a comma-separated string.
func sortedHexList(values []uint16) string {
	hexValues := make([]string, len(values))
	for i, v := range values {
		hexValues[i] = fmt.Sprintf("%04x", v)
	}
	sort.Strings(hexValues)
	return strings.Join(hexValues, ",")
}

// truncatedSHA256Hex computes the SHA-256 hash of the input string and returns
// the first n characters of the hex-encoded digest.
func truncatedSHA256Hex(input string, n int) string {
	hash := sha256.Sum256([]byte(input))
	fullHex := hex.EncodeToString(hash[:])
	if n > len(fullHex) {
		return fullHex
	}
	return fullHex[:n]
}

// TLSVersionToString converts a TLS version to the 2-character JA4 format
func TLSVersionToString(version uint16) string {
	switch version {
	case 0x0301:
		return "10"
	case 0x0302:
		return "11"
	case 0x0303:
		return "12"
	case 0x0304:
		return "13"
	default:
		return "00"
	}
}

// Analyzer is a TLS fingerprint analyzer
type Analyzer struct {
	profile *profiles.ClientProfile
}

// NewAnalyzer creates a new TLS fingerprint analyzer
func NewAnalyzer(profile *profiles.ClientProfile) *Analyzer {
	return &Analyzer{profile: profile}
}

// AnalyzeJA3 computes and returns the JA3 fingerprint
func (a *Analyzer) AnalyzeJA3() *JA3Result {
	if a.profile == nil {
		return nil
	}
	return CalculateJA3FromProfile(*a.profile)
}

// AnalyzeJA4 computes and returns the JA4 fingerprint
func (a *Analyzer) AnalyzeJA4() *JA4Result {
	if a.profile == nil {
		return nil
	}
	spec := core.ClientHelloSpec{
		TLSVersion:   a.profile.TLSVersion,
		CipherSuites: a.profile.CipherSuites,
		Extensions:   a.profile.Extensions,
	}
	return CalculateJA4(spec)
}

// joinCurves joins a CurveID slice into a dash-separated string
func joinCurves(curves []core.CurveID) string {
	var parts []string
	for _, c := range curves {
		parts = append(parts, strconv.Itoa(int(c)))
	}
	return strings.Join(parts, "-")
}

// Fingerprint generates a complete TLS fingerprint map
func (a *Analyzer) Fingerprint() map[string]interface{} {
	return map[string]interface{}{
		"ja3": a.AnalyzeJA3(),
		"ja4": a.AnalyzeJA4(),
	}
}
