package ech

import (
	"encoding/binary"
	"fmt"
)

// translated comment
const (
	// translated comment
	ExtensionEncryptedClientHello uint16 = 0xfe0d

	// ExtensionECHOuterExtensions ECH Outer Extensions
	ExtensionECHOuterExtensions uint16 = 0xfd00

	// translated comment
	ECHVersionDraft13 uint16 = 0xfe0d

	// translated comment
	ECHVersionDraft14 uint16 = 0xfe0e

	// translated comment
	ECHVersionDraft15 uint16 = 0xfe0f
)

// translated comment
type ECHClientHelloType uint8

const (
	// translated comment
	ECHClientHelloTypeOuter ECHClientHelloType = 0

	// translated comment
	ECHClientHelloTypeInner ECHClientHelloType = 1
)

// translated comment
type ECHExtension struct {
	// translated comment
	Type uint16

	// translated comment
	Version uint16

	// translated comment
	ClientHelloType ECHClientHelloType

	// Cipher Suite
	CipherSuite KEMCipherSuite

	// translated comment
	EncodedCHLength uint16

	// translated comment
	EncodedCH []byte

	// translated comment
	ConfigID uint8

	// KEM ID
	KEMID uint16

	// KDF ID
	KDFID uint16

	// AEAD ID
	AEADID uint16

	// translated comment
	Raw []byte
}

// translated comment
type KEMCipherSuite struct {
	KDFID  uint16
	AEADID uint16
}

// translated comment
func ParseECHExtension(extType uint16, data []byte) (*ECHExtension, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("ECH extension too short: %d bytes", len(data))
	}

	ech := &ECHExtension{
		Type: extType,
		Raw:  data,
	}

	// translated comment
	ech.Version = binary.BigEndian.Uint16(data[0:2])

	// translated comment
	switch ech.Version {
	case ECHVersionDraft13, ECHVersionDraft14, ECHVersionDraft15:
		return parseECHDraft13(ech, data)
	default:
		// translated comment
		return parseECHGeneric(ech, data)
	}
}

// translated comment
func parseECHDraft13(ech *ECHExtension, data []byte) (*ECHExtension, error) {
	if len(data) < 3 {
		return nil, fmt.Errorf("ECH draft 13 extension too short: %d bytes", len(data))
	}

	offset := 2 // translated comment

	// translated comment
	if ech.Type == ExtensionEncryptedClientHello {
		// translated comment
		if offset >= len(data) {
			return nil, fmt.Errorf("ECH extension truncated at type")
		}
		ech.ClientHelloType = ECHClientHelloType(data[offset])
		offset++

		// translated comment
		if ech.ClientHelloType == ECHClientHelloTypeInner {
			return ech, nil
		}

		// translated comment
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

		// translated comment
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

// translated comment
func parseECHGeneric(ech *ECHExtension, data []byte) (*ECHExtension, error) {
	// translated comment
	// translated comment
	ech.EncodedCH = data[2:] // translated comment
	return ech, nil
}

// translated comment
func (e *ECHExtension) IsOuterHello() bool {
	return e.ClientHelloType == ECHClientHelloTypeOuter
}

// translated comment
func (e *ECHExtension) IsInnerHello() bool {
	return e.ClientHelloType == ECHClientHelloTypeInner
}

// translated comment
func (e *ECHExtension) IsGREASE() bool {
	// translated comment
	return e.Version == 0x0000 || e.Version == 0x0a0a || e.Version == 0x1a1a
}

// translated comment
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

// translated comment
func (e *ECHExtension) Serialize() ([]byte, error) {
	if e.IsInnerHello() {
		// translated comment
		data := make([]byte, 3)
		binary.BigEndian.PutUint16(data[0:2], e.Version)
		data[2] = byte(ECHClientHelloTypeInner)
		return data, nil
	}

	// translated comment
	if e.EncodedCH == nil {
		return nil, fmt.Errorf("encoded CH is required for outer hello")
	}

	// translated comment
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

// translated comment
func (e *ECHExtension) Validate() error {
	// translated comment
	if e.Version == 0 {
		return fmt.Errorf("ECH version cannot be 0 (except GREASE)")
	}

	// translated comment
	if e.ClientHelloType != ECHClientHelloTypeInner &&
		e.ClientHelloType != ECHClientHelloTypeOuter {
		return fmt.Errorf("invalid ECH client hello type: %d", e.ClientHelloType)
	}

	// translated comment
	if e.IsOuterHello() && len(e.EncodedCH) == 0 {
		return fmt.Errorf("outer hello requires encoded CH")
	}

	return nil
}

// translated comment
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

// translated comment
type ECHConfigList struct {
	Configs []ECHConfigRecord
}

// translated comment
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

// translated comment
func ParseECHConfigList(data []byte) (*ECHConfigList, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("ECH config list too short")
	}

	list := &ECHConfigList{}

	// translated comment
	totalLength := binary.BigEndian.Uint16(data[0:2])

	// translated comment
	if totalLength == 0 {
		return list, nil
	}

	// translated comment
	if len(data) < 2+int(totalLength) {
		return nil, fmt.Errorf("ECH config list truncated: expected %d, got %d", totalLength, len(data)-2)
	}

	// translated comment
	configData := data[2:]
	offset := 0

	for offset < len(configData) {
		if len(configData) < offset+4 {
			return nil, fmt.Errorf("ECH config truncated at offset %d", offset)
		}

		config := ECHConfigRecord{}

		// translated comment
		config.Version = binary.BigEndian.Uint16(configData[offset : offset+2])
		offset += 2

		// translated comment
		config.Length = binary.BigEndian.Uint16(configData[offset : offset+2])
		offset += 2

		// translated comment
		if len(configData) < offset+int(config.Length) {
			return nil, fmt.Errorf("ECH config content truncated")
		}
		config.Contents = configData[offset : offset+int(config.Length)]
		offset += int(config.Length)

		// translated comment
		if err := parseECHConfigContents(&config); err != nil {
			// translated comment
			// translated comment
			_ = err
		}

		list.Configs = append(list.Configs, config)
	}

	return list, nil
}

// translated comment
func parseECHConfigContents(config *ECHConfigRecord) error {
	if len(config.Contents) < 8 {
		return fmt.Errorf("ECH config contents too short")
	}

	data := config.Contents
	offset := 0

	// translated comment
	publicNameLen := data[offset]
	offset++

	// Public Name
	if len(data) < offset+int(publicNameLen) {
		return fmt.Errorf("public name truncated")
	}
	config.PublicName = string(data[offset : offset+int(publicNameLen)])
	offset += int(publicNameLen)

	// translated comment
	// translated comment

	return nil
}
