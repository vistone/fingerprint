package ja4

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	tls "github.com/bogdanfinn/utls"
)

// JA4Result JA4 fingerprint result
type JA4Result struct {
	// JA4 full fingerprint (hash form)
	Hash string
	// JA4_r full raw string form
	RawString string
	// JA4_a: protocol identifier + TLS version + SNI + cipher suite count + extension count + ALPN first/last characters
	JA4A string
	// JA4_b: cipher suites (sorted, comma-separated 4-digit hex, first 12 chars of SHA256)
	JA4B string
	// JA4_c: extensions (sorted, comma-separated 4-digit hex) + signature algorithms (first 12 chars of SHA256)
	JA4C string
}

// ClientProfile client fingerprint configuration
type ClientProfile interface {
	GetClientHelloSpec() (tls.ClientHelloSpec, error)
}

var MappedTLSClients map[string]ClientProfile

// InitMappedTLSClients called by root package to initialize client mapping table
func InitMappedTLSClients(clients interface{}) {
	if m, ok := clients.(map[string]ClientProfile); ok {
		MappedTLSClients = m
	}
}

// tlsVersionToJA4 converts TLS version to JA4 format string
func tlsVersionToJA4(version uint16) string {
	switch version {
	case tls.VersionTLS13:
		return "13"
	case tls.VersionTLS12:
		return "12"
	case tls.VersionTLS11:
		return "11"
	case tls.VersionTLS10:
		return "10"
	default:
		return "00"
	}
}

// firstLastALPN extracts first and last characters from ALPN string
// Non-ASCII characters are replaced with '9'
func firstLastALPN(s string) (byte, byte) {
	if len(s) == 0 {
		return '0', '0'
	}
	replaceNonASCII := func(b byte) byte {
		if b < 128 {
			return b
		}
		return '9'
	}
	first := replaceNonASCII(s[0])
	if len(s) == 1 {
		return first, '0'
	}
	last := replaceNonASCII(s[len(s)-1])
	return first, last
}

// sha256Hash12 computes SHA256 hash and returns the first 12 characters
func sha256Hash12(input string) string {
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)[:12]
}

// JA4Signature JA4 fingerprint signature input
type JA4Signature struct {
	// TLS version
	TLSVersion uint16
	// Cipher suites (including GREASE)
	CipherSuites []uint16
	// Extensions (including GREASE)
	Extensions []uint16
	// Signature algorithms (including GREASE)
	SignatureAlgorithms []uint16
	// SNI (optional)
	SNI string
	// ALPN first protocol (optional)
	ALPN string
}

// ComputeJA4 computes JA4 fingerprint (sorted version)
func (s *JA4Signature) ComputeJA4() *JA4Result {
	return s.computeJA4WithOrder(false)
}

// ComputeJA4Original computes JA4 fingerprint (original order version, i.e. JA4_o)
func (s *JA4Signature) ComputeJA4Original() *JA4Result {
	return s.computeJA4WithOrder(true)
}

// computeJA4WithOrder computes JA4 fingerprint (specified order)
func (s *JA4Signature) computeJA4WithOrder(originalOrder bool) *JA4Result {
	// Filter GREASE values
	filteredCiphers := filterGREASEUint16(s.CipherSuites)
	filteredExtensions := filterGREASEUint16(s.Extensions)
	filteredSigAlgs := filterGREASEUint16(s.SignatureAlgorithms)

	// Protocol identifier ('t' for TLS, 'q' for QUIC)
	protocol := "t"

	// TLS version string
	tlsVersionStr := tlsVersionToJA4(s.TLSVersion)

	// SNI indicator: 'd' if SNI present, 'i' if absent
	sniIndicator := "i"
	if s.SNI != "" {
		sniIndicator = "d"
	}

	// Cipher suite count (2-digit decimal, max 99) - uses count after filtering GREASE (per JA4 spec)
	cipherCount := fmt.Sprintf("%02d", min99(len(filteredCiphers)))

	// Extension count (2-digit decimal, max 99) - uses count after filtering GREASE (per JA4 spec)
	extensionCount := fmt.Sprintf("%02d", min99(len(filteredExtensions)))

	// ALPN first and last characters
	alpnFirst, alpnLast := byte('0'), byte('0')
	if s.ALPN != "" {
		alpnFirst, alpnLast = firstLastALPN(s.ALPN)
	}

	// JA4_a format: protocol + version + sni + cipher_count + extension_count + alpn_first + alpn_last
	ja4a := fmt.Sprintf("%s%s%s%s%s%c%c", protocol, tlsVersionStr, sniIndicator, cipherCount, extensionCount, alpnFirst, alpnLast)

	// JA4_b: cipher suites (sorted or original order, comma-separated, 4-digit hex) - GREASE filtered
	ciphersForB := make([]uint16, len(filteredCiphers))
	copy(ciphersForB, filteredCiphers)
	if !originalOrder {
		sort.Slice(ciphersForB, func(i, j int) bool { return ciphersForB[i] < ciphersForB[j] })
	}
	ja4bParts := make([]string, len(ciphersForB))
	for i, c := range ciphersForB {
		ja4bParts[i] = fmt.Sprintf("%04x", c)
	}
	ja4bRaw := strings.Join(ja4bParts, ",")

	// JA4_c: extensions (sorted or original order) + signature algorithms
	extensionsForC := make([]uint16, len(filteredExtensions))
	copy(extensionsForC, filteredExtensions)

	// Sorted version: remove SNI (0x0000) and ALPN (0x0010) then sort
	// Original order version: keep SNI/ALPN and maintain original order
	if !originalOrder {
		filtered := extensionsForC[:0]
		for _, ext := range extensionsForC {
			if ext != 0x0000 && ext != 0x0010 {
				filtered = append(filtered, ext)
			}
		}
		extensionsForC = filtered
		sort.Slice(extensionsForC, func(i, j int) bool { return extensionsForC[i] < extensionsForC[j] })
	}

	extParts := make([]string, len(extensionsForC))
	for i, e := range extensionsForC {
		extParts[i] = fmt.Sprintf("%04x", e)
	}
	extensionsStr := strings.Join(extParts, ",")

	// Signature algorithms are not sorted, but GREASE is filtered
	sigAlgParts := make([]string, len(filteredSigAlgs))
	for i, s := range filteredSigAlgs {
		sigAlgParts[i] = fmt.Sprintf("%04x", s)
	}
	sigAlgsStr := strings.Join(sigAlgParts, ",")

	// Per spec, if there are no signature algorithms, the string does not end with underscore
	var ja4cRaw string
	if sigAlgsStr == "" {
		ja4cRaw = extensionsStr
	} else if extensionsStr == "" {
		ja4cRaw = sigAlgsStr
	} else {
		ja4cRaw = extensionsStr + "_" + sigAlgsStr
	}

	// Generate JA4_b and JA4_c hashes (first 12 characters of SHA256)
	ja4bHash := sha256Hash12(ja4bRaw)
	ja4cHash := sha256Hash12(ja4cRaw)

	// JA4 full hash: ja4_a + "_" + ja4_b_hash + "_" + ja4_c_hash
	ja4Hash := fmt.Sprintf("%s_%s_%s", ja4a, ja4bHash, ja4cHash)

	// JA4_r raw: ja4_a + "_" + ja4_b_raw + "_" + ja4_c_raw
	ja4Raw := fmt.Sprintf("%s_%s_%s", ja4a, ja4bRaw, ja4cRaw)

	return &JA4Result{
		Hash:      ja4Hash,
		RawString: ja4Raw,
		JA4A:      ja4a,
		JA4B:      ja4bRaw,
		JA4C:      ja4cRaw,
	}
}

// min99 returns the smaller of n and 99
func min99(n int) int {
	if n > 99 {
		return 99
	}
	return n
}

// isGREASEValue checks if the value is a GREASE value (RFC 8701)
func isGREASEValue(v uint16) bool {
	return v&0x0f0f == 0x0a0a && (v>>8) == (v&0x00ff)
}

// filterGREASEUint16 filters GREASE values (uint16 slice)
func filterGREASEUint16(values []uint16) []uint16 {
	result := make([]uint16, 0, len(values))
	for _, v := range values {
		if !isGREASEValue(v) {
			result = append(result, v)
		}
	}
	return result
}

// ComputeJA4FromSpec computes JA4 fingerprint from TLS ClientHello spec
func ComputeJA4FromSpec(spec tls.ClientHelloSpec) (*JA4Result, error) {
	sig := &JA4Signature{
		TLSVersion: tls.VersionTLS12,
	}

	// Extract cipher suites
	sig.CipherSuites = make([]uint16, len(spec.CipherSuites))
	copy(sig.CipherSuites, spec.CipherSuites)

	// Extract extension information
	extensions := make([]uint16, 0)
	var sigAlgs []uint16

	for _, ext := range spec.Extensions {
		switch e := ext.(type) {
		case *tls.SupportedVersionsExtension:
			for _, v := range e.Versions {
				if !isGREASEValue(v) && v > sig.TLSVersion {
					sig.TLSVersion = v
				}
			}
			extensions = append(extensions, 43)

		case *tls.SNIExtension:
			// Extract SNI (first name)
			if e.ServerName != "" {
				sig.SNI = e.ServerName
			}
			extensions = append(extensions, 0)

		case *tls.ALPNExtension:
			if len(e.AlpnProtocols) > 0 {
				sig.ALPN = e.AlpnProtocols[0]
			}
			extensions = append(extensions, 16)

		case *tls.SignatureAlgorithmsExtension:
			for _, sa := range e.SupportedSignatureAlgorithms {
				sigAlgs = append(sigAlgs, uint16(sa))
			}
			extensions = append(extensions, 13)

		case *tls.SupportedCurvesExtension:
			extensions = append(extensions, 10)

		case *tls.SupportedPointsExtension:
			extensions = append(extensions, 11)

		case *tls.StatusRequestExtension:
			extensions = append(extensions, 5)

		case *tls.SessionTicketExtension:
			extensions = append(extensions, 35)

		case *tls.SCTExtension:
			extensions = append(extensions, 18)

		case *tls.KeyShareExtension:
			extensions = append(extensions, 51)

		case *tls.PSKKeyExchangeModesExtension:
			extensions = append(extensions, 45)

		case *tls.ExtendedMasterSecretExtension:
			extensions = append(extensions, 23)

		case *tls.RenegotiationInfoExtension:
			extensions = append(extensions, 65281)

		case *tls.UtlsCompressCertExtension:
			extensions = append(extensions, 27)

		case *tls.ApplicationSettingsExtension:
			extensions = append(extensions, 17513)

		case *tls.ApplicationSettingsExtensionNew:
			extensions = append(extensions, 17613)

		case *tls.UtlsGREASEExtension:
			// Skip GREASE extensions

		default:
			_ = e
		}
	}

	sig.Extensions = extensions
	sig.SignatureAlgorithms = sigAlgs

	return sig.ComputeJA4(), nil
}

// ComputeJA4FromProfile computes JA4 fingerprint from ClientProfile
func ComputeJA4FromProfile(profile ClientProfile) (*JA4Result, error) {
	spec, err := profile.GetClientHelloSpec()
	if err != nil {
		return nil, fmt.Errorf("failed to get ClientHelloSpec: %w", err)
	}
	return ComputeJA4FromSpec(spec)
}

// ComputeJA4ByProfileName computes JA4 fingerprint by profile name
func ComputeJA4ByProfileName(profileName string) (*JA4Result, error) {
	if MappedTLSClients == nil {
		return nil, fmt.Errorf("JA4 client mapping not initialized")
	}
	profile, ok := MappedTLSClients[profileName]
	if !ok {
		return nil, fmt.Errorf("fingerprint %s not found", profileName)
	}
	return ComputeJA4FromProfile(profile)
}
