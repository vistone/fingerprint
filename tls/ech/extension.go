package ech

import (
	"encoding/binary"
	"fmt"
)

// ECHExtensionType ECH 扩展类型常量
const (
	// ExtensionEncryptedClientHello ECH 扩展类型 (draft-ietf-tls-esni)
	ExtensionEncryptedClientHello uint16 = 0xfe0d

	// ExtensionECHOuterExtensions ECH Outer Extensions
	ExtensionECHOuterExtensions uint16 = 0xfd00

	// ECHVersionDraft13 ECH Draft 13 版本
	ECHVersionDraft13 uint16 = 0xfe0d

	// ECHVersionDraft14 ECH Draft 14 版本
	ECHVersionDraft14 uint16 = 0xfe0e

	// ECHVersionDraft15 ECH Draft 15 版本
	ECHVersionDraft15 uint16 = 0xfe0f
)

// ECHClientHelloType ECH ClientHello 类型
type ECHClientHelloType uint8

const (
	// ECHClientHelloTypeOuter 外层 ClientHello（RFC draft-ietf-tls-esni: outer=0）
	ECHClientHelloTypeOuter ECHClientHelloType = 0

	// ECHClientHelloTypeInner 内层 ClientHello（RFC draft-ietf-tls-esni: inner=1）
	ECHClientHelloTypeInner ECHClientHelloType = 1
)

// ECHExtension ECH 扩展结构
type ECHExtension struct {
	// 扩展类型
	Type uint16

	// ECH 版本
	Version uint16

	// ClientHello 类型（内层/外层）
	ClientHelloType ECHClientHelloType

	// Cipher Suite
	CipherSuite KEMCipherSuite

	// Encoded CH 长度
	EncodedCHLength uint16

	// Encoded CH 内容（加密或编码的 ClientHello）
	EncodedCH []byte

	// Config ID（外层 ClientHello 使用）
	ConfigID uint8

	// KEM ID
	KEMID uint16

	// KDF ID
	KDFID uint16

	// AEAD ID
	AEADID uint16

	// 原始数据
	Raw []byte
}

// KEMCipherSuite KEM 算法套件
type KEMCipherSuite struct {
	KDFID  uint16
	AEADID uint16
}

// ParseECHExtension 解析 ECH 扩展数据
func ParseECHExtension(extType uint16, data []byte) (*ECHExtension, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("ECH extension too short: %d bytes", len(data))
	}

	ech := &ECHExtension{
		Type: extType,
		Raw:  data,
	}

	// 解析版本
	ech.Version = binary.BigEndian.Uint16(data[0:2])

	// 根据版本解析
	switch ech.Version {
	case ECHVersionDraft13, ECHVersionDraft14, ECHVersionDraft15:
		return parseECHDraft13(ech, data)
	default:
		// 未知版本，尝试通用解析
		return parseECHGeneric(ech, data)
	}
}

// parseECHDraft13 解析 Draft 13+ 格式
func parseECHDraft13(ech *ECHExtension, data []byte) (*ECHExtension, error) {
	if len(data) < 3 {
		return nil, fmt.Errorf("ECH draft 13 extension too short: %d bytes", len(data))
	}

	offset := 2 // 跳过版本

	// 对于内层 ClientHello
	if ech.Type == ExtensionEncryptedClientHello {
		// 读取 ClientHello 类型
		if offset >= len(data) {
			return nil, fmt.Errorf("ECH extension truncated at type")
		}
		ech.ClientHelloType = ECHClientHelloType(data[offset])
		offset++

		// 内层 ClientHello 只包含类型
		if ech.ClientHelloType == ECHClientHelloTypeInner {
			return ech, nil
		}

		// 外层 ClientHello 包含更多信息
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

		// Encoded CH 长度和内容
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

// parseECHGeneric 通用解析（未知版本）
func parseECHGeneric(ech *ECHExtension, data []byte) (*ECHExtension, error) {
	// 对于未知版本，仅保存原始数据
	// 实际应用中可能需要根据版本特定的逻辑处理
	ech.EncodedCH = data[2:] // 跳过版本字段
	return ech, nil
}

// IsOuterHello 是否为外层 ClientHello
func (e *ECHExtension) IsOuterHello() bool {
	return e.ClientHelloType == ECHClientHelloTypeOuter
}

// IsInnerHello 是否为内层 ClientHello
func (e *ECHExtension) IsInnerHello() bool {
	return e.ClientHelloType == ECHClientHelloTypeInner
}

// IsGREASE 是否为 GREASE ECH
func (e *ECHExtension) IsGREASE() bool {
	// GREASE ECH 通常版本号为 0 或特定测试值
	return e.Version == 0x0000 || e.Version == 0x0a0a || e.Version == 0x1a1a
}

// GetVersionName 获取版本名称
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

// Serialize 序列化 ECH 扩展
func (e *ECHExtension) Serialize() ([]byte, error) {
	if e.IsInnerHello() {
		// 内层 ClientHello: 版本 + 类型
		data := make([]byte, 3)
		binary.BigEndian.PutUint16(data[0:2], e.Version)
		data[2] = byte(ECHClientHelloTypeInner)
		return data, nil
	}

	// 外层 ClientHello
	if e.EncodedCH == nil {
		return nil, fmt.Errorf("encoded CH is required for outer hello")
	}

	// 计算总长度
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

// Validate 验证 ECH 扩展
func (e *ECHExtension) Validate() error {
	// 验证版本
	if e.Version == 0 {
		return fmt.Errorf("ECH version cannot be 0 (except GREASE)")
	}

	// 验证 ClientHello 类型
	if e.ClientHelloType != ECHClientHelloTypeInner &&
		e.ClientHelloType != ECHClientHelloTypeOuter {
		return fmt.Errorf("invalid ECH client hello type: %d", e.ClientHelloType)
	}

	// 外层 ClientHello 需要 encoded CH
	if e.IsOuterHello() && len(e.EncodedCH) == 0 {
		return fmt.Errorf("outer hello requires encoded CH")
	}

	return nil
}

// String 返回 ECH 扩展的描述
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

// ECHConfigList ECH 配置列表（用于 ECH 配置扩展）
type ECHConfigList struct {
	Configs []ECHConfigRecord
}

// ECHConfigRecord 单个 ECH 配置记录
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

// ParseECHConfigList 解析 ECH 配置列表
func ParseECHConfigList(data []byte) (*ECHConfigList, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("ECH config list too short")
	}

	list := &ECHConfigList{}

	// 读取总长度
	totalLength := binary.BigEndian.Uint16(data[0:2])

	// 如果总长度为 0，返回空列表
	if totalLength == 0 {
		return list, nil
	}

	// 验证数据长度
	if len(data) < 2+int(totalLength) {
		return nil, fmt.Errorf("ECH config list truncated: expected %d, got %d", totalLength, len(data)-2)
	}

	// 实际配置数据（跳过长度前缀）
	configData := data[2:]
	offset := 0

	for offset < len(configData) {
		if len(configData) < offset+4 {
			return nil, fmt.Errorf("ECH config truncated at offset %d", offset)
		}

		config := ECHConfigRecord{}

		// 版本
		config.Version = binary.BigEndian.Uint16(configData[offset : offset+2])
		offset += 2

		// 长度
		config.Length = binary.BigEndian.Uint16(configData[offset : offset+2])
		offset += 2

		// 内容
		if len(configData) < offset+int(config.Length) {
			return nil, fmt.Errorf("ECH config content truncated")
		}
		config.Contents = configData[offset : offset+int(config.Length)]
		offset += int(config.Length)

		// 解析内容（简化版本）
		if err := parseECHConfigContents(&config); err != nil {
			// 解析失败也继续，记录配置但不解析内容
			// 实际应用可能需要更严格的处理
			_ = err
		}

		list.Configs = append(list.Configs, config)
	}

	return list, nil
}

// parseECHConfigContents 解析 ECH 配置内容
func parseECHConfigContents(config *ECHConfigRecord) error {
	if len(config.Contents) < 8 {
		return fmt.Errorf("ECH config contents too short")
	}

	data := config.Contents
	offset := 0

	// Public Name 长度
	publicNameLen := data[offset]
	offset++

	// Public Name
	if len(data) < offset+int(publicNameLen) {
		return fmt.Errorf("public name truncated")
	}
	config.PublicName = string(data[offset : offset+int(publicNameLen)])
	offset += int(publicNameLen)

	// Public Key（简化解析）
	// 实际解析需要完整的 HPKE 公钥解析

	return nil
}
