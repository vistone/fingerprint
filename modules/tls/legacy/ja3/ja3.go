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
	// Input validation: check whether spec is empty
	if err := validateClientHelloSpec(spec); err != nil {
		return nil, err
	}

	result := &JA3Result{}

	// Extract TLS version (defaults to TLS 1.2)
	result.TLSVersion = tls.VersionTLS12

	// Extract cipher suites (GREASE filtered)
	ciphers := filterGREASEUint16(spec.CipherSuites)
	
	// Validate cipher suite list
	if len(ciphers) == 0 {
		return nil, fmt.Errorf("%w: no valid cipher suites", ErrInvalidClientHelloSpec)
	}
	result.CipherSuites = ciphers

	// Extract extension information
	extensions := make([]uint16, 0)
	var curves []tls.CurveID
	var pointFormats []uint8

	for _, ext := range spec.Extensions {
		if ext == nil {
			continue // Skip nil extension
		}
		
		switch e := ext.(type) {
		case *tls.SupportedVersionsExtension:
			// Validate extension data
			if e.Versions == nil {
				continue
			}
			// Extract highest TLS version
			for _, v := range e.Versions {
				if !isGREASEValue(v) && v > result.TLSVersion {
					result.TLSVersion = v
				}
			}
			extensions = append(extensions, 43) // extension_type_supported_versions

		case *tls.SupportedCurvesExtension:
			if e.Curves == nil {
				continue
			}
			curves = filterGREASECurveID(e.Curves)
			extensions = append(extensions, 10) // extension_type_supported_groups

		case *tls.SupportedPointsExtension:
			if e.SupportedPoints == nil {
				continue
			}
			pointFormats = e.SupportedPoints
			extensions = append(extensions, 11) // extension_type_ec_point_formats

		case *tls.SNIExtension:
			extensions = append(extensions, 0) // extension_type_server_name

		case *tls.StatusRequestExtension:
			extensions = append(extensions, 5) // extension_type_status_request

		case *tls.SessionTicketExtension:
			extensions = append(extensions, 35) // extension_type_session_ticket

		case *tls.ALPNExtension:
			extensions = append(extensions, 16) // extension_type_alpn

		case *tls.SignatureAlgorithmsExtension:
			extensions = append(extensions, 13) // extension_type_signature_algorithms

		case *tls.SCTExtension:
			extensions = append(extensions, 18) // extension_type_signed_certificate_timestamp

		case *tls.KeyShareExtension:
			extensions = append(extensions, 51) // extension_type_key_share

		case *tls.PSKKeyExchangeModesExtension:
			extensions = append(extensions, 45) // extension_type_psk_key_exchange_modes

		case *tls.ExtendedMasterSecretExtension:
			extensions = append(extensions, 23) // extension_type_extended_master_secret

		case *tls.RenegotiationInfoExtension:
			extensions = append(extensions, 65281) // extension_type_renegotiation_info (0xff01)

		case *tls.UtlsCompressCertExtension:
			extensions = append(extensions, 27) // extension_type_compress_certificate

		case *tls.ApplicationSettingsExtension:
			extensions = append(extensions, 17513) // extension_type_application_settings

		case *tls.ApplicationSettingsExtensionNew:
			extensions = append(extensions, 17613) // extension_type_application_settings_new

		case *tls.UtlsGREASEExtension:
			// Skip GREASE extensions

		default:
			// Ignore unknown extensions
			_ = e
		}
	}

	// Filter GREASE values in extensions
	result.Extensions = filterGREASEUint16(extensions)
	result.EllipticCurves = curves
	result.EllipticCurvePointFormats = pointFormats

	// Build JA3 raw string
	// Format: TLSVersion,CipherSuites,Extensions,EllipticCurves,EllipticCurvePointFormats
	rawParts := []string{
		strconv.Itoa(int(result.TLSVersion)),
		uint16SliceToString(result.CipherSuites),
		uint16SliceToString(result.Extensions),
		curveIDSliceToString(result.EllipticCurves),
		uint8SliceToString(result.EllipticCurvePointFormats),
	}
	result.RawString = strings.Join(rawParts, ",")

	// Calculate MD5 hash
	hash := md5.Sum([]byte(result.RawString))
	result.Hash = fmt.Sprintf("%x", hash)

	return result, nil
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
