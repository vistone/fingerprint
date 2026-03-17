package ja3

import (
	"crypto/md5"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"

	tls "github.com/bogdanfinn/utls"
)

// JA3Result JA3 fingerprint result
type JA3Result struct {
	// JA3 fingerprint hash (MD5)
	Hash string
	// JA3 raw string (readable form)
	RawString string
	// TLS version
	TLSVersion uint16
	// Cipher suite list (GREASE filtered)
	CipherSuites []uint16
	// Extension list (GREASE filtered)
	Extensions []uint16
	// Elliptic curve list (GREASE filtered)
	EllipticCurves []tls.CurveID
	// Elliptic curve point format list
	EllipticCurvePointFormats []uint8
}

// ClientProfile client fingerprint configuration
type ClientProfile interface {
	GetClientHelloSpec() (tls.ClientHelloSpec, error)
}

var (
	ja3ProfileIndexOnce sync.Once
	ja3ProfileIndex     map[string][]string
	MappedTLSClients    map[string]ClientProfile
)

// InitMappedTLSClients is called by the root package to initialize client mapping
// Use interface{} to avoid type matching issues
func InitMappedTLSClients(clients interface{}) {
	if m, ok := clients.(map[string]ClientProfile); ok {
		MappedTLSClients = m
		return
	}
	// Try converting to map[string]interface{}
	if m, ok := clients.(map[string]interface{}); ok {
		MappedTLSClients = make(map[string]ClientProfile)
		for k, v := range m {
			if cp, ok := v.(ClientProfile); ok {
				MappedTLSClients[k] = cp
			}
		}
	}
}

// InitMappedTLSClientsRaw receives any ClientProfile map from root package and converts types
func InitMappedTLSClientsRaw(clients interface{}) {
	// Use reflection to export and convert underlying map
	clientsVal := reflect.ValueOf(clients)
	if clientsVal.Kind() != reflect.Map {
		return
	}

	// Create new map
	MappedTLSClients = make(map[string]ClientProfile)

	// Iterate over all key-value pairs in original map
	for _, keyVal := range clientsVal.MapKeys() {
		key := fmt.Sprintf("%v", keyVal.Interface())
		valInterface := clientsVal.MapIndex(keyVal).Interface()

		// Try converting value to ClientProfile
		// Since value implements GetClientHelloSpec, it should satisfy ClientProfile interface
		if cp, ok := valInterface.(ClientProfile); ok {
			MappedTLSClients[key] = cp
		} else {
			// If direct conversion fails, use type assertion to obtain getter method
			// Check whether the value has GetClientHelloSpec method here
			refVal := reflect.ValueOf(valInterface)
			method := refVal.MethodByName("GetClientHelloSpec")
			if method.IsValid() {
				// Create a wrapper to implement ClientProfile interface
				wrapper := &dynamicClientProfile{value: valInterface}
				MappedTLSClients[key] = wrapper
			}
		}
	}
}

// dynamicClientProfile dynamic wrapper implementing ClientProfile interface
type dynamicClientProfile struct {
	value interface{}
}

// GetClientHelloSpec implements ClientProfile interface
func (d *dynamicClientProfile) GetClientHelloSpec() (tls.ClientHelloSpec, error) {
	refVal := reflect.ValueOf(d.value)
	method := refVal.MethodByName("GetClientHelloSpec")
	if !method.IsValid() {
		return tls.ClientHelloSpec{}, fmt.Errorf("object does not implement GetClientHelloSpec method")
	}

	results := method.Call(nil)
	if len(results) >= 2 {
		if spec, ok := results[0].Interface().(tls.ClientHelloSpec); ok {
			var err error
			if !results[1].IsNil() {
				err = results[1].Interface().(error)
			}
			return spec, err
		}
	}
	return tls.ClientHelloSpec{}, fmt.Errorf("method invocation failed")
}

func buildJA3ProfileIndex() {
	index := make(map[string][]string)

	if MappedTLSClients == nil {
		ja3ProfileIndex = index
		return
	}

	for name, profile := range MappedTLSClients {
		result, err := ComputeJA3FromProfile(profile)
		if err != nil || result == nil || result.Hash == "" {
			continue
		}

		hash := strings.ToLower(result.Hash)
		index[hash] = append(index[hash], name)
	}

	for hash := range index {
		sort.Strings(index[hash])
	}

	ja3ProfileIndex = index
}

// findProfileByJA3NoCopy finds matching profiles by JA3 hash (internal zero-copy path)
// Return value directly references internal cache, for read-only scenarios only.
func findProfileByJA3NoCopy(ja3Hash string) []string {
	ja3ProfileIndexOnce.Do(buildJA3ProfileIndex)
	if ja3ProfileIndex == nil {
		return nil
	}
	return ja3ProfileIndex[strings.ToLower(ja3Hash)]
}

// isGREASEValue checks whether value is GREASE (RFC 8701)
// GREASE value format: 0xXAXA (hex)
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

// filterGREASECurveID filters GREASE values (CurveID slice)
func filterGREASECurveID(curves []tls.CurveID) []tls.CurveID {
	result := make([]tls.CurveID, 0, len(curves))
	for _, c := range curves {
		if !isGREASEValue(uint16(c)) {
			result = append(result, c)
		}
	}
	return result
}

// uint16SliceToString converts uint16 slice to dash-separated string
func uint16SliceToString(values []uint16) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.Itoa(int(v))
	}
	return strings.Join(parts, "-")
}

// curveIDSliceToString converts CurveID slice to dash-separated string
func curveIDSliceToString(curves []tls.CurveID) string {
	parts := make([]string, len(curves))
	for i, c := range curves {
		parts[i] = strconv.Itoa(int(c))
	}
	return strings.Join(parts, "-")
}

// uint8SliceToString converts uint8 slice to dash-separated string
func uint8SliceToString(values []uint8) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.Itoa(int(v))
	}
	return strings.Join(parts, "-")
}

// ComputeJA3FromSpec computes JA3 fingerprint from TLS ClientHello spec
// JA3 algorithm: MD5(TLSVersion,Ciphers,Extensions,EllipticCurves,EllipticCurvePointFormats)
func ComputeJA3FromSpec(spec tls.ClientHelloSpec) (*JA3Result, error) {
	if err := validateClientHelloSpec(spec); err != nil {
		return nil, err
	}

	ciphers := filterGREASEUint16(spec.CipherSuites)
	if len(ciphers) == 0 {
		return nil, fmt.Errorf("%w: no valid cipher suites", ErrInvalidClientHelloSpec)
	}
	state := extractJA3State(spec.Extensions)
	return buildJA3Result(ciphers, state), nil
}

type ja3State struct {
	tlsVersion   uint16
	extensions   []uint16
	curves       []tls.CurveID
	pointFormats []uint8
}

func extractJA3State(extensions []tls.TLSExtension) ja3State {
	state := ja3State{
		tlsVersion: tls.VersionTLS12,
		extensions: make([]uint16, 0),
	}

	for _, ext := range extensions {
		processJA3Extension(ext, &state)
	}
	state.extensions = filterGREASEUint16(state.extensions)
	return state
}

func processJA3Extension(ext tls.TLSExtension, state *ja3State) {
	if ext == nil {
		return
	}

	switch e := ext.(type) {
	case *tls.SupportedVersionsExtension:
		if e.Versions == nil {
			return
		}
		state.tlsVersion = maxTLSVersion(state.tlsVersion, e.Versions)
		state.extensions = append(state.extensions, 43)
	case *tls.SupportedCurvesExtension:
		if e.Curves == nil {
			return
		}
		state.curves = filterGREASECurveID(e.Curves)
		state.extensions = append(state.extensions, 10)
	case *tls.SupportedPointsExtension:
		if e.SupportedPoints == nil {
			return
		}
		state.pointFormats = e.SupportedPoints
		state.extensions = append(state.extensions, 11)
	case *tls.SNIExtension:
		state.extensions = append(state.extensions, 0)
	case *tls.StatusRequestExtension:
		state.extensions = append(state.extensions, 5)
	case *tls.SessionTicketExtension:
		state.extensions = append(state.extensions, 35)
	case *tls.ALPNExtension:
		state.extensions = append(state.extensions, 16)
	case *tls.SignatureAlgorithmsExtension:
		state.extensions = append(state.extensions, 13)
	case *tls.SCTExtension:
		state.extensions = append(state.extensions, 18)
	case *tls.KeyShareExtension:
		state.extensions = append(state.extensions, 51)
	case *tls.PSKKeyExchangeModesExtension:
		state.extensions = append(state.extensions, 45)
	case *tls.ExtendedMasterSecretExtension:
		state.extensions = append(state.extensions, 23)
	case *tls.RenegotiationInfoExtension:
		state.extensions = append(state.extensions, 65281)
	case *tls.UtlsCompressCertExtension:
		state.extensions = append(state.extensions, 27)
	case *tls.ApplicationSettingsExtension:
		state.extensions = append(state.extensions, 17513)
	case *tls.ApplicationSettingsExtensionNew:
		state.extensions = append(state.extensions, 17613)
	case *tls.UtlsGREASEExtension:
		return
	default:
		return
	}
}

func maxTLSVersion(initial uint16, versions []uint16) uint16 {
	maxVersion := initial
	for _, v := range versions {
		if !isGREASEValue(v) && v > maxVersion {
			maxVersion = v
		}
	}
	return maxVersion
}

func buildJA3Result(ciphers []uint16, state ja3State) *JA3Result {
	result := &JA3Result{
		TLSVersion:                state.tlsVersion,
		CipherSuites:              ciphers,
		Extensions:                state.extensions,
		EllipticCurves:            state.curves,
		EllipticCurvePointFormats: state.pointFormats,
	}

	result.RawString = buildJA3RawString(result)
	hash := md5.Sum([]byte(result.RawString))
	result.Hash = fmt.Sprintf("%x", hash)
	return result
}

func buildJA3RawString(result *JA3Result) string {
	rawParts := []string{
		strconv.Itoa(int(result.TLSVersion)),
		uint16SliceToString(result.CipherSuites),
		uint16SliceToString(result.Extensions),
		curveIDSliceToString(result.EllipticCurves),
		uint8SliceToString(result.EllipticCurvePointFormats),
	}
	return strings.Join(rawParts, ",")
}

// validateClientHelloSpec validates ClientHello spec validity
func validateClientHelloSpec(spec tls.ClientHelloSpec) error {
	// Check cipher suite list length
	if len(spec.CipherSuites) == 0 {
		return fmt.Errorf("%w: cipher suites list is empty", ErrInvalidClientHelloSpec)
	}
	if len(spec.CipherSuites) > 255 {
		return fmt.Errorf("%w: cipher suites list too long (%d > 255)", ErrInvalidClientHelloSpec, len(spec.CipherSuites))
	}

	// Check extension list length
	if len(spec.Extensions) == 0 {
		return fmt.Errorf("%w: extensions list is empty", ErrInvalidClientHelloSpec)
	}
	if len(spec.Extensions) > 255 {
		return fmt.Errorf("%w: extensions list too long (%d > 255)", ErrInvalidClientHelloSpec, len(spec.Extensions))
	}

	// Check validity of each cipher suite value
	for i, cipher := range spec.CipherSuites {
		if cipher == 0 {
			return fmt.Errorf("%w: cipher suite at index %d is zero", ErrInvalidClientHelloSpec, i)
		}
	}

	return nil
}

// ComputeJA3FromProfile computes JA3 fingerprint from ClientProfile
func ComputeJA3FromProfile(profile ClientProfile) (*JA3Result, error) {
	spec, err := profile.GetClientHelloSpec()
	if err != nil {
		return nil, fmt.Errorf("failed to get ClientHelloSpec: %w", err)
	}
	return ComputeJA3FromSpec(spec)
}

// ComputeJA3ByProfileName computes JA3 by profile name
func ComputeJA3ByProfileName(profileName string) (*JA3Result, error) {
	profile, ok := MappedTLSClients[profileName]
	if !ok {
		return nil, fmt.Errorf("fingerprint %s not found", profileName)
	}
	return ComputeJA3FromProfile(profile)
}

// MatchJA3 checks whether two JA3 hashes match
func MatchJA3(hash1, hash2 string) bool {
	return strings.EqualFold(hash1, hash2)
}

// FindProfileByJA3 finds matching ClientProfile names by JA3 hash
// Returns all matching profile names (can be multiple)
func FindProfileByJA3(ja3Hash string) []string {
	matches := findProfileByJA3NoCopy(ja3Hash)
	if len(matches) == 0 {
		return nil
	}

	result := make([]string, len(matches))
	copy(result, matches)
	return result
}
