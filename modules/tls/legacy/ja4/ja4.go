package ja4

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	tls "github.com/bogdanfinn/utls"
)

// translated comment
type JA4Result struct {
	// translated comment
	Hash string
	// translated comment
	RawString string
	// translated comment
	JA4A string
	// translated comment
	JA4B string
	// translated comment
	JA4C string
}

// translated comment
type ClientProfile interface {
	GetClientHelloSpec() (tls.ClientHelloSpec, error)
}

var MappedTLSClients map[string]ClientProfile

// translated comment
func InitMappedTLSClients(clients interface{}) {
	if m, ok := clients.(map[string]ClientProfile); ok {
		MappedTLSClients = m
	}
}

// translated comment
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

// translated comment
// translated comment
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

// translated comment
func sha256Hash12(input string) string {
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)[:12]
}

// translated comment
type JA4Signature struct {
	// translated comment
	TLSVersion uint16
	// translated comment
	CipherSuites []uint16
	// translated comment
	Extensions []uint16
	// translated comment
	SignatureAlgorithms []uint16
	// translated comment
	SNI string
	// translated comment
	ALPN string
}

// translated comment
func (s *JA4Signature) ComputeJA4() *JA4Result {
	return s.computeJA4WithOrder(false)
}

// translated comment
func (s *JA4Signature) ComputeJA4Original() *JA4Result {
	return s.computeJA4WithOrder(true)
}

// translated comment
func (s *JA4Signature) computeJA4WithOrder(originalOrder bool) *JA4Result {
	// translated comment
	filteredCiphers := filterGREASEUint16(s.CipherSuites)
	filteredExtensions := filterGREASEUint16(s.Extensions)
	filteredSigAlgs := filterGREASEUint16(s.SignatureAlgorithms)

	// translated comment
	protocol := "t"

	// translated comment
	tlsVersionStr := tlsVersionToJA4(s.TLSVersion)

	// translated comment
	sniIndicator := "i"
	if s.SNI != "" {
		sniIndicator = "d"
	}

	// translated comment
	cipherCount := fmt.Sprintf("%02d", min99(len(filteredCiphers)))

	// translated comment
	extensionCount := fmt.Sprintf("%02d", min99(len(filteredExtensions)))

	// translated comment
	alpnFirst, alpnLast := byte('0'), byte('0')
	if s.ALPN != "" {
		alpnFirst, alpnLast = firstLastALPN(s.ALPN)
	}

	// translated comment
	ja4a := fmt.Sprintf("%s%s%s%s%s%c%c", protocol, tlsVersionStr, sniIndicator, cipherCount, extensionCount, alpnFirst, alpnLast)

	// translated comment
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

	// translated comment
	extensionsForC := make([]uint16, len(filteredExtensions))
	copy(extensionsForC, filteredExtensions)

	// translated comment
	// translated comment
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

	// translated comment
	sigAlgParts := make([]string, len(filteredSigAlgs))
	for i, s := range filteredSigAlgs {
		sigAlgParts[i] = fmt.Sprintf("%04x", s)
	}
	sigAlgsStr := strings.Join(sigAlgParts, ",")

	// translated comment
	var ja4cRaw string
	if sigAlgsStr == "" {
		ja4cRaw = extensionsStr
	} else if extensionsStr == "" {
		ja4cRaw = sigAlgsStr
	} else {
		ja4cRaw = extensionsStr + "_" + sigAlgsStr
	}

	// translated comment
	ja4bHash := sha256Hash12(ja4bRaw)
	ja4cHash := sha256Hash12(ja4cRaw)

	// translated comment
	ja4Hash := fmt.Sprintf("%s_%s_%s", ja4a, ja4bHash, ja4cHash)

	// translated comment
	ja4Raw := fmt.Sprintf("%s_%s_%s", ja4a, ja4bRaw, ja4cRaw)

	return &JA4Result{
		Hash:      ja4Hash,
		RawString: ja4Raw,
		JA4A:      ja4a,
		JA4B:      ja4bRaw,
		JA4C:      ja4cRaw,
	}
}

// translated comment
func min99(n int) int {
	if n > 99 {
		return 99
	}
	return n
}

// translated comment
func isGREASEValue(v uint16) bool {
	return v&0x0f0f == 0x0a0a && (v>>8) == (v&0x00ff)
}

// translated comment
func filterGREASEUint16(values []uint16) []uint16 {
	result := make([]uint16, 0, len(values))
	for _, v := range values {
		if !isGREASEValue(v) {
			result = append(result, v)
		}
	}
	return result
}

// translated comment
func ComputeJA4FromSpec(spec tls.ClientHelloSpec) (*JA4Result, error) {
	sig := &JA4Signature{
		TLSVersion: tls.VersionTLS12,
	}

	// translated comment
	sig.CipherSuites = make([]uint16, len(spec.CipherSuites))
	copy(sig.CipherSuites, spec.CipherSuites)

	// translated comment
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
			// translated comment
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
			// translated comment

		default:
			_ = e
		}
	}

	sig.Extensions = extensions
	sig.SignatureAlgorithms = sigAlgs

	return sig.ComputeJA4(), nil
}

// translated comment
func ComputeJA4FromProfile(profile ClientProfile) (*JA4Result, error) {
	spec, err := profile.GetClientHelloSpec()
	if err != nil {
		return nil, fmt.Errorf("failed to get ClientHelloSpec: %w", err)
	}
	return ComputeJA4FromSpec(spec)
}

// translated comment
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
