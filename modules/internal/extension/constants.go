package extension

import "errors"

var (
	// Registry errors
	ErrExtensionNotFound      = errors.New("extension not found")
	ErrExtensionNotRegistered = errors.New("extension not registered")
	ErrParserNotFound         = errors.New("parser not found")
	ErrAnalyzerNotFound       = errors.New("analyzer not found")
	ErrHandlerNotFound        = errors.New("handler not found")
	ErrPluginNotFound         = errors.New("plugin not found")

	// Validation errors
	ErrInvalidMetadata = errors.New("invalid extension metadata")
	ErrInvalidParser   = errors.New("invalid parser")
	ErrInvalidAnalyzer = errors.New("invalid analyzer")
	ErrInvalidHandler  = errors.New("invalid handler")
	ErrInvalidPlugin   = errors.New("invalid plugin")

	// Processing errors
	ErrParseFailure            = errors.New("failed to parse extension")
	ErrAnalysisFailure         = errors.New("failed to analyze extension")
	ErrHandleFailure           = errors.New("failed to handle extension")
	ErrPluginInitFailure       = errors.New("failed to initialize plugin")
	ErrPluginValidationFailure = errors.New("plugin validation failed")

	// Configuration errors
	ErrInvalidConfig = errors.New("invalid configuration")
	ErrMissingConfig = errors.New("missing required configuration")
)

// Standard extension type constants
const (
	// TLS 1.3 Extensions
	ExtensionServerName               ExtensionType = 0x0000
	ExtensionMaxFragmentLength        ExtensionType = 0x0001
	ExtensionClientCertificateType    ExtensionType = 0x0009
	ExtensionSupportedGroups          ExtensionType = 0x000a
	ExtensionECPointFormats           ExtensionType = 0x000b
	ExtensionSignatureAlgorithms      ExtensionType = 0x000d
	ExtensionUseSignatureAlgorithm    ExtensionType = 0x0010
	ExtensionApplicationLayerProtocol ExtensionType = 0x0010
	ExtensionStatus                   ExtensionType = 0x0005
	ExtensionSupportedVersions        ExtensionType = 0x002b
	ExtensionKeyShare                 ExtensionType = 0x0033
	ExtensionPSKKeyExchangeModes      ExtensionType = 0x002d
	ExtensionCertificateAuthorities   ExtensionType = 0x002f
	ExtensionOIDFilters               ExtensionType = 0x0030
	ExtensionPostHandshakeAuth        ExtensionType = 0x0031
	ExtensionSignatureAlgorithmsCert  ExtensionType = 0x0032

	// Client Hints
	ExtensionSecCHUA                ExtensionType = 0xfd01
	ExtensionSecCHUAMobile          ExtensionType = 0xfd02
	ExtensionSecCHUAPlatform        ExtensionType = 0xfd03
	ExtensionSecCHUAPlatformVersion ExtensionType = 0xfd04
	ExtensionSecCHUAModelVersion    ExtensionType = 0xfd05

	// Encrypted Client Hello (ECH)
	ExtensionEncryptedClientHello ExtensionType = 0xfe0d

	// ECH Outer Extensions
	ExtensionECHOuterExtensions ExtensionType = 0xfd00

	// GREASE Extensions (Generic unsupported extension)
	ExtensionGREASE ExtensionType = 0x0a0a

	// Padding (TLS 1.3)
	ExtensionPadding ExtensionType = 0x0015

	// Pre-shared Key
	ExtensionPreSharedKey ExtensionType = 0x0029

	// Certificate List
	ExtensionCertificateList ExtensionType = 0x0000

	// Unknown/Custom
	ExtensionCustom ExtensionType = 0xffff
)

// Standard extension categories
const (
	CategoryEncryption     = "encryption"
	CategoryNegotiation    = "negotiation"
	CategoryPreference     = "preference"
	CategoryIdentification = "identification"
	CategoryCompression    = "compression"
	CategorySecurity       = "security"
	CategoryPerformance    = "performance"
	CategoryCompatibility  = "compatibility"
	CategoryExperimental   = "experimental"
	CategoryClientHints    = "client_hints"
)

var standardExtensionEntries = []*ExtensionMetadata{
	{
		Type:                  ExtensionEncryptedClientHello,
		Name:                  "Encrypted Client Hello",
		Description:           "Encodes ClientHello to reduce cleartext information",
		RFC:                   "RFC 9180",
		IANANumber:            0xfe0d,
		Category:              CategoryEncryption,
		IsExperimental:        false,
		CompatibleTLSVersions: []uint16{0x0304},
	},
	{
		Type:                  ExtensionServerName,
		Name:                  "server_name",
		Description:           "Indicates server name in SNI",
		RFC:                   "RFC 6066",
		IANANumber:            0x0000,
		Category:              CategoryNegotiation,
		IsExperimental:        false,
		CompatibleTLSVersions: []uint16{0x0303, 0x0304},
	},
	{
		Type:                  ExtensionSupportedGroups,
		Name:                  "supported_groups",
		Description:           "Indicates supported ECDH groups/curves",
		RFC:                   "RFC 8446",
		IANANumber:            0x000a,
		Category:              CategoryNegotiation,
		IsExperimental:        false,
		CompatibleTLSVersions: []uint16{0x0303, 0x0304},
	},
	{
		Type:                  ExtensionECPointFormats,
		Name:                  "ec_point_formats",
		Description:           "Indicates supported EC point formats",
		RFC:                   "RFC 8446",
		IANANumber:            0x000b,
		Category:              CategoryNegotiation,
		IsExperimental:        false,
		CompatibleTLSVersions: []uint16{0x0303},
	},
	{
		Type:                  ExtensionSignatureAlgorithms,
		Name:                  "signature_algorithms",
		Description:           "Indicates supported signature algorithms",
		RFC:                   "RFC 8446",
		IANANumber:            0x000d,
		Category:              CategoryNegotiation,
		IsExperimental:        false,
		CompatibleTLSVersions: []uint16{0x0303, 0x0304},
	},
	{
		Type:                  ExtensionSupportedVersions,
		Name:                  "supported_versions",
		Description:           "Indicates supported TLS versions",
		RFC:                   "RFC 8446",
		IANANumber:            0x002b,
		Category:              CategoryNegotiation,
		IsExperimental:        false,
		CompatibleTLSVersions: []uint16{0x0303, 0x0304},
	},
	{
		Type:                  ExtensionKeyShare,
		Name:                  "key_share",
		Description:           "Provides key share information",
		RFC:                   "RFC 8446",
		IANANumber:            0x0033,
		Category:              CategoryNegotiation,
		IsExperimental:        false,
		CompatibleTLSVersions: []uint16{0x0304},
	},
	{
		Type:                  ExtensionPreSharedKey,
		Name:                  "pre_shared_key",
		Description:           "Provides pre-shared key information",
		RFC:                   "RFC 8446",
		IANANumber:            0x0029,
		Category:              CategoryNegotiation,
		IsExperimental:        false,
		CompatibleTLSVersions: []uint16{0x0304},
	},
	{
		Type:                  ExtensionStatus,
		Name:                  "status_request",
		Description:           "OCSP stapling",
		RFC:                   "RFC 6066",
		IANANumber:            0x0005,
		Category:              CategorySecurity,
		IsExperimental:        false,
		CompatibleTLSVersions: []uint16{0x0303, 0x0304},
	},
	{
		Type:                  ExtensionApplicationLayerProtocol,
		Name:                  "application_layer_protocol_negotiation",
		Description:           "Indicates application layer protocols",
		RFC:                   "RFC 7301",
		IANANumber:            0x0010,
		Category:              CategoryNegotiation,
		IsExperimental:        false,
		CompatibleTLSVersions: []uint16{0x0303, 0x0304},
	},
	{
		Type:                  ExtensionPadding,
		Name:                  "padding",
		Description:           "Adds padding to ClientHello",
		RFC:                   "RFC 7685",
		IANANumber:            0x0015,
		Category:              CategoryPerformance,
		IsExperimental:        false,
		CompatibleTLSVersions: []uint16{0x0303, 0x0304},
	},
	{
		Type:                  ExtensionECHOuterExtensions,
		Name:                  "ech_outer_extensions",
		Description:           "Lists extensions in the outer ClientHello",
		RFC:                   "RFC 9180",
		IANANumber:            0xfd00,
		Category:              CategoryEncryption,
		IsExperimental:        false,
		CompatibleTLSVersions: []uint16{0x0304},
	},
	{
		Type:                  ExtensionGREASE,
		Name:                  "GREASE",
		Description:           "Generic unsupported extension for compatibility",
		RFC:                   "RFC 8701",
		IANANumber:            0x0a0a,
		Category:              CategoryCompatibility,
		IsExperimental:        true,
		CompatibleTLSVersions: []uint16{0x0303, 0x0304},
	},
}

// initStandardExtensions initializes standard extensions
func initStandardExtensions() {
	registerExtensionBatch(standardExtensionEntries)
}

func registerExtensionBatch(entries []*ExtensionMetadata) {
	for _, entry := range entries {
		RegisterExtension(entry)
	}
}
