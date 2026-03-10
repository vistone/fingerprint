// translated comment
package tls

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// translated comment
type JA3Result struct {
	// translated comment
	Hash string
	// translated comment
	RawString string
	// translated comment
	TLSVersion uint16
	// translated comment
	CipherSuites []uint16
	// translated comment
	Extensions []uint16
	// translated comment
	EllipticCurves []core.CurveID
	// translated comment
	EllipticCurvePointFormats []uint8
}

// translated comment
func CalculateJA3(spec core.ClientHelloSpec) *JA3Result {
	// translated comment
	cipherSuites := filterGREASEUint16(spec.CipherSuites)
	extensions := filterGREASEExtensions(spec.Extensions)
	curves := filterGREASECurves(spec.SupportedCurves)

	// translated comment
	parts := []string{
		strconv.Itoa(int(spec.TLSVersion)),
		joinUint16(cipherSuites),
		joinExtensions(extensions),
		joinCurves(curves),
		joinUint8(spec.SupportedPoints),
	}

	rawString := strings.Join(parts, ",")

	// translated comment
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

// translated comment
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

// translated comment
func IsGREASEUint16(v uint16) bool {
	// translated comment
	// translated comment
	return ((v >> 8) & 0xFF) == (v & 0xFF) && (v&0x0F) == 0x0A
}

// translated comment
func filterGREASEUint16(values []uint16) []uint16 {
	var result []uint16
	for _, v := range values {
		if !IsGREASEUint16(v) {
			result = append(result, v)
		}
	}
	return result
}

// translated comment
func filterGREASEExtensions(exts []core.TLSExtension) []core.TLSExtension {
	var result []core.TLSExtension
	for _, e := range exts {
		if !IsGREASEUint16(e.Type) {
			result = append(result, e)
		}
	}
	return result
}

// translated comment
func filterGREASECurves(curves []core.CurveID) []core.CurveID {
	var result []core.CurveID
	for _, c := range curves {
		if !IsGREASEUint16(uint16(c)) {
			result = append(result, c)
		}
	}
	return result
}

// translated comment
func joinUint16(values []uint16) string {
	var parts []string
	for _, v := range values {
		parts = append(parts, strconv.Itoa(int(v)))
	}
	return strings.Join(parts, "-")
}

// translated comment
func joinUint8(values []uint8) string {
	var parts []string
	for _, v := range values {
		parts = append(parts, strconv.Itoa(int(v)))
	}
	return strings.Join(parts, "-")
}

// translated comment
func joinExtensions(exts []core.TLSExtension) string {
	var parts []string
	for _, e := range exts {
		parts = append(parts, strconv.Itoa(int(e.Type)))
	}
	return strings.Join(parts, "-")
}

// translated comment
func extensionTypes(exts []core.TLSExtension) []uint16 {
	var result []uint16
	for _, e := range exts {
		result = append(result, e.Type)
	}
	return result
}

// translated comment
type JA4Result struct {
	// translated comment
	Fingerprint string
	// translated comment
	TLSVersion uint16
	// translated comment
	CipherSuitesCount int
	// translated comment
	ExtensionsCount int
}

// translated comment
func CalculateJA4(spec core.ClientHelloSpec) *JA4Result {
	// translated comment
	// translated comment

	// translated comment
	version := core.TLSVersionToString(spec.TLSVersion)
	cipherCount := len(filterGREASEUint16(spec.CipherSuites))
	extCount := len(filterGREASEExtensions(spec.Extensions))

	// translated comment
	fingerprint := "t" + version + "d" +
		strconv.Itoa(cipherCount) +
		strconv.Itoa(extCount)

	return &JA4Result{
		Fingerprint:       fingerprint,
		TLSVersion:        spec.TLSVersion,
		CipherSuitesCount: cipherCount,
		ExtensionsCount:   extCount,
	}
}

// translated comment
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

// translated comment
type Analyzer struct {
	profile *profiles.ClientProfile
}

// translated comment
func NewAnalyzer(profile *profiles.ClientProfile) *Analyzer {
	return &Analyzer{profile: profile}
}

// translated comment
func (a *Analyzer) AnalyzeJA3() *JA3Result {
	if a.profile == nil {
		return nil
	}
	return CalculateJA3FromProfile(*a.profile)
}

// translated comment
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

// translated comment
func joinCurves(curves []core.CurveID) string {
	var parts []string
	for _, c := range curves {
		parts = append(parts, strconv.Itoa(int(c)))
	}
	return strings.Join(parts, "-")
}

// translated comment
func (a *Analyzer) Fingerprint() map[string]interface{} {
	return map[string]interface{}{
		"ja3": a.AnalyzeJA3(),
		"ja4": a.AnalyzeJA4(),
	}
}
