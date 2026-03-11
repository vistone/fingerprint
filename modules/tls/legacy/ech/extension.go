package ech

import (
	"encoding/binary"
	"fmt"
)

// ECHExtensionType ECH extension type constants
const (
	// ExtensionEncryptedClientHello ECH extension type (draft-ietf-tls-esni)
	ExtensionEncryptedClientHello uint16 = 0xfe0d

	// ExtensionECHOuterExtensions ECH Outer Extensions
	ExtensionECHOuterExtensions uint16 = 0xfd00

	// ECHVersionDraft13 ECH Draft 13 version
	ECHVersionDraft13 uint16 = 0xfe0d

	// ECHVersionDraft14 ECH Draft 14 version
	ECHVersionDraft14 uint16 = 0xfe0e

	// ECHVersionDraft15 ECH Draft 15 version
	ECHVersionDraft15 uint16 = 0xfe0f
)

// ECHClientHelloType ECH ClientHello type
type ECHClientHelloType uint8

const (
	// ECHClientHelloTypeOuter Outer ClientHello (RFC draft-ietf-tls-esni: outer=0)
	ECHClientHelloTypeOuter ECHClientHelloType = 0

	// ECHClientHelloTypeInner Inner ClientHello (RFC draft-ietf-tls-esni: inner=1)
	ECHClientHelloTypeInner ECHClientHelloType = 1
)

// ECHExtension ECH extension structure
type ECHExtension struct {
	// Extension type
	Type uint16

	// ECH version
	Version uint16

	// ClientHello type (inner/outer)
	ClientHelloType ECHClientHelloType

	// Cipher Suite
	CipherSuite KEMCipherSuite

	// Encoded CH length
	EncodedCHLength uint16

	// Encoded CH content (encrypted or encoded ClientHello)
	EncodedCH []byte

	// Config ID (used by outer ClientHello)
	ConfigID uint8

	// KEM ID
	KEMID uint16

	// KDF ID
	KDFID uint16

	// AEAD ID
	AEADID uint16

	// Raw data
	Raw []byte
}

// KEMCipherSuite KEM algorithm suite
type KEMCipherSuite struct {
	KDFID  uint16
	AEADID uint16
}

// ParseECHExtension parses ECH extension data
func ParseECHExtension(extType uint16, data []byte) (*ECHExtension, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("ECH extension too short: %d bytes", len(data))
	}

	ech := &ECHExtension{
		Type: extType,
		Raw:  data,
	}

	// Parse version
	ech.Version = binary.BigEndian.Uint16(data[0:2])

	// Parse based on version
	switch ech.Version {
	case ECHVersionDraft13, ECHVersionDraft14, ECHVersionDraft15:
		return parseECHDraft13(ech, data)
	default:
		// Unknown version, attempt generic parsing
		return parseECHGeneric(ech, data)
	}
}

// parseECHDraft13 parses Draft 13+ format
func parseECHDraft13(ech *ECHExtension, data []byte) (*ECHExtension, error) {
	if len(data) < 3 {
		return nil, fmt.Errorf("ECH draft 13 extension too short: %d bytes", len(data))
	}

	offset := 2 // Skip version

	// For inner ClientHello
	if ech.Type == ExtensionEncryptedClientHello {
		// Read ClientHello type
		if offset >= len(data) {
			return nil, fmt.Errorf("ECH extension truncated at type")
		}
		ech.ClientHelloType = ECHClientHelloType(data[offset])
		offset++

		// Inner ClientHello only contains type
		if ech.ClientHelloType == ECHClientHelloTypeInner {
			return ech, nil
		}

		// Outer ClientHello contains more information
		if len(data) < offset+7 {
			return nil, fmt.Errorf("ECH outer hello truncated")
		}

		// Cipher Suite (KDF + AEAD)
		ech.CipherSuite.KDFID = binary.BigEndian.Uint16(data[offset : offset+2])
		ech.CipherSuite.AEADID = binary.BigEndian.Uint16(data[offset+2 : offset+4])
		offset += 4

		// Config ID
		if offset >= len(data) {
			return nil, fmt.Errorf("ECH extension truncated at config ID")
		}
		ech.ConfigID = data[offset]
		offset++

		// Encoded CH length and content
		if len(data) < offset+2 {
			return nil, fmt.Errorf("ECH extension truncated at length")
		}
		ech.EncodedCHLength = binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2

		if len(data) < offset+int(ech.EncodedCHLength) {
			return nil, fmt.Errorf("ECH extension truncated at encoded CH")
		}
		ech.EncodedCH = data[offset : offset+int(ech.EncodedCHLength)]
	}

	return ech, nil
}

// parseECHGeneric generic parsing (unknown version)
func parseECHGeneric(ech *ECHExtension, data []byte) (*ECHExtension, error) {
	// For unknown versions, only save raw data
	// In practice, version-specific logic may be needed
	ech.EncodedCH = data[2:] // Skip version field
	return ech, nil
}

// IsOuterHello returns whether this is an outer ClientHello
func (e *ECHExtension) IsOuterHello() bool {
	return e.ClientHelloType == ECHClientHelloTypeOuter
}

// IsInnerHello returns whether this is an inner ClientHello
func (e *ECHExtension) IsInnerHello() bool {
	return e.ClientHelloType == ECHClientHelloTypeInner
}

// IsGREASE returns whether this is a GREASE ECH
func (e *ECHExtension) IsGREASE() bool {
	// GREASE ECH typically has version 0 or specific test values
	return e.Version == 0x0000 || e.Version == 0x0a0a || e.Version == 0x1a1a
}

// GetVersionName returns the version name
func (e *ECHExtension) GetVersionName() string {
	switch e.Version {
	case ECHVersionDraft13:
		return "Draft 13"
	case ECHVersionDraft14:
		return "Draft 14"
	case ECHVersionDraft15:
		return "Draft 15"
	case 0x0000:
		return "GREASE"
	default:
		return fmt.Sprintf("Unknown(0x%04x)", e.Version)
	}
}

// Serialize serializes the ECH extension
func (e *ECHExtension) Serialize() ([]byte, error) {
	if e.IsInnerHello() {
		// Inner ClientHello: version + type
		data := make([]byte, 3)
		binary.BigEndian.PutUint16(data[0:2], e.Version)
		data[2] = byte(ECHClientHelloTypeInner)
		return data, nil
	}

	// Outer ClientHello
	if e.EncodedCH == nil {
		return nil, fmt.Errorf("encoded CH is required for outer hello")
	}

	// Calculate total length
	length := 2 + // version
		1 + // client_hello_type
		4 + // cipher_suite (kdf + aead)
		1 + // config_id
		2 + // encoded_ch length
		len(e.EncodedCH) // encoded_ch

	data := make([]byte, 0, length)

	// Version
	versionBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(versionBytes, e.Version)
	data = append(data, versionBytes...)

	// ClientHello Type
	data = append(data, byte(ECHClientHelloTypeOuter))

	// Cipher Suite
	kdfBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(kdfBytes, e.CipherSuite.KDFID)
	data = append(data, kdfBytes...)

	aeadBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(aeadBytes, e.CipherSuite.AEADID)
	data = append(data, aeadBytes...)

	// Config ID
	data = append(data, e.ConfigID)

	// Encoded CH Length
	lengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBytes, uint16(len(e.EncodedCH)))
	data = append(data, lengthBytes...)

	// Encoded CH
	data = append(data, e.EncodedCH...)

	return data, nil
}

// Validate validates the ECH extension
func (e *ECHExtension) Validate() error {
	// Validate version
	if e.Version == 0 {
		return fmt.Errorf("ECH version cannot be 0 (except GREASE)")
	}

	// Validate ClientHello type
	if e.ClientHelloType != ECHClientHelloTypeInner &&
		e.ClientHelloType != ECHClientHelloTypeOuter {
		return fmt.Errorf("invalid ECH client hello type: %d", e.ClientHelloType)
	}

	// Outer ClientHello requires encoded CH
	if e.IsOuterHello() && len(e.EncodedCH) == 0 {
		return fmt.Errorf("outer hello requires encoded CH")
	}

	return nil
}

// String returns a description of the ECH extension
func (e *ECHExtension) String() string {
	return fmt.Sprintf("ECH{type=%s, version=%s, hello_type=%s, encoded_ch_len=%d}",
		hexType(e.Type),
		e.GetVersionName(),
		clientHelloTypeName(e.ClientHelloType),
		len(e.EncodedCH),
	)
}

func hexType(t uint16) string {
	return fmt.Sprintf("0x%04x", t)
}

func clientHelloTypeName(t ECHClientHelloType) string {
	switch t {
	case ECHClientHelloTypeInner:
		return "inner"
	case ECHClientHelloTypeOuter:
		return "outer"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

// ECHConfigList ECH config list (used for ECH config extension)
type ECHConfigList struct {
	Configs []ECHConfigRecord
}

// ECHConfigRecord single ECH config record
type ECHConfigRecord struct {
	Version           uint16
	Length            uint16
	Contents          []byte
	PublicName        string
	PublicKey         []byte
	KemID             uint16
	KdfID             uint16
	AeadID            uint16
	MaximumNameLength uint8
}

// ParseECHConfigList parses the ECH config list
func ParseECHConfigList(data []byte) (*ECHConfigList, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("ECH config list too short")
	}

	list := &ECHConfigList{}

	// Read total length
	totalLength := binary.BigEndian.Uint16(data[0:2])

	// If total length is 0, return empty list
	if totalLength == 0 {
		return list, nil
	}

	// Validate data length
	if len(data) < 2+int(totalLength) {
		return nil, fmt.Errorf("ECH config list truncated: expected %d, got %d", totalLength, len(data)-2)
	}

	// Actual config data (skip length prefix)
	configData := data[2:]
	offset := 0

	for offset < len(configData) {
		if len(configData) < offset+4 {
			return nil, fmt.Errorf("ECH config truncated at offset %d", offset)
		}

		config := ECHConfigRecord{}

		// Version
		config.Version = binary.BigEndian.Uint16(configData[offset : offset+2])
		offset += 2

		// Length
		config.Length = binary.BigEndian.Uint16(configData[offset : offset+2])
		offset += 2

		// Content
		if len(configData) < offset+int(config.Length) {
			return nil, fmt.Errorf("ECH config content truncated")
		}
		config.Contents = configData[offset : offset+int(config.Length)]
		offset += int(config.Length)

		// Parse content (simplified version)
		if err := parseECHConfigContents(&config); err != nil {
			// Continue even if parsing fails, record config but don't parse content
			// In practice, stricter handling may be needed
			_ = err
		}

		list.Configs = append(list.Configs, config)
	}

	return list, nil
}

// parseECHConfigContents parses ECH config content
func parseECHConfigContents(config *ECHConfigRecord) error {
	if len(config.Contents) < 8 {
		return fmt.Errorf("ECH config contents too short")
	}

	data := config.Contents
	offset := 0

	// Public Name length
	publicNameLen := data[offset]
	offset++

	// Public Name
	if len(data) < offset+int(publicNameLen) {
		return fmt.Errorf("public name truncated")
	}
	config.PublicName = string(data[offset : offset+int(publicNameLen)])
	offset += int(publicNameLen)

	// Public Key (simplified parsing)
	// Full parsing requires complete HPKE public key parsing

	return nil
}
