package ech

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
)

// ConfigGenerator ECH 配置生成器
type ConfigGenerator struct {
	// 配置选项
	Options ConfigOptions
}

// ConfigOptions ECH 配置选项
type ConfigOptions struct {
	// ECH 版本
	Version uint16

	// 公钥名称（用于外层 ClientHello 的 SNI）
	PublicName string

	// 最大域名长度
	MaxNameLength uint8

	// 支持的 KEM 算法
	KEMAlgorithms []uint16

	// 支持的 KDF 算法
	KDFAlgorithms []uint16

	// 支持的 AEAD 算法
	AEADAlgorithms []uint16

	// 扩展列表
	Extensions []ECHConfigExtension
}

// ECHConfigExtension ECH 配置扩展
type ECHConfigExtension struct {
	Type uint16
	Data []byte
}

// DefaultConfigOptions 默认配置选项
func DefaultConfigOptions() ConfigOptions {
	return ConfigOptions{
		Version:       ECHVersionDraft13,
		PublicName:    "cloudflare-ech.com",
		MaxNameLength: 128,
		KEMAlgorithms: []uint16{
			0x0020, // X25519
			0x0010, // P-256
		},
		KDFAlgorithms: []uint16{
			0x0001, // HKDF-SHA256
		},
		AEADAlgorithms: []uint16{
			0x0001, // AES-128-GCM
			0x0002, // AES-256-GCM
		},
	}
}

// NewConfigGenerator 创建配置生成器
func NewConfigGenerator(opts ConfigOptions) *ConfigGenerator {
	if opts.Version == 0 {
		opts.Version = ECHVersionDraft13
	}
	if opts.PublicName == "" {
		opts.PublicName = "cloudflare-ech.com"
	}
	if opts.MaxNameLength == 0 {
		opts.MaxNameLength = 128
	}
	if len(opts.KEMAlgorithms) == 0 {
		opts.KEMAlgorithms = DefaultConfigOptions().KEMAlgorithms
	}

	return &ConfigGenerator{Options: opts}
}

// GenerateECHConfig 生成 ECH 配置
func (g *ConfigGenerator) GenerateECHConfig() (*ECHConfigRecord, error) {
	// 生成密钥对（简化版本，实际应使用 HPKE）
	privateKey, publicKey, err := g.generateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("generate key pair: %w", err)
	}

	config := &ECHConfigRecord{
		Version:           g.Options.Version,
		PublicName:        g.Options.PublicName,
		MaximumNameLength: g.Options.MaxNameLength,
		PublicKey:         publicKey,
	}

	// 选择算法
	if len(g.Options.KEMAlgorithms) > 0 {
		config.KemID = g.Options.KEMAlgorithms[0]
	}
	if len(g.Options.KDFAlgorithms) > 0 {
		config.KdfID = g.Options.KDFAlgorithms[0]
	}
	if len(g.Options.AEADAlgorithms) > 0 {
		config.AeadID = g.Options.AEADAlgorithms[0]
	}

	// 序列化配置内容
	contents, err := g.serializeConfigContents(config, privateKey)
	if err != nil {
		return nil, fmt.Errorf("serialize config: %w", err)
	}

	config.Contents = contents
	config.Length = uint16(len(contents))

	return config, nil
}

// generateKeyPair 生成密钥对（简化实现）
func (g *ConfigGenerator) generateKeyPair() (privateKey, publicKey []byte, err error) {
	// 实际实现应使用 HPKE
	// 这里使用 RSA 作为占位
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// 序列化私钥
	privateKey = x509.MarshalPKCS1PrivateKey(key)

	// 序列化公钥
	publicKey, err = x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, publicKey, nil
}

// serializeConfigContents 序列化配置内容
func (g *ConfigGenerator) serializeConfigContents(config *ECHConfigRecord, privateKey []byte) ([]byte, error) {
	// ECH 配置内容格式（Draft 13）：
	// - KemID (2 bytes)
	// - PublicKey (length-prefixed)
	// - CipherSuites (length-prefixed list)
	// - MaximumNameLength (1 byte)
	// - PublicName (length-prefixed)
	// - Extensions (length-prefixed)

	var contents []byte

	// KemID
	kemIDBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(kemIDBytes, config.KemID)
	contents = append(contents, kemIDBytes...)

	// PublicKey (长度前缀 2 字节)
	pubKeyLen := make([]byte, 2)
	binary.BigEndian.PutUint16(pubKeyLen, uint16(len(config.PublicKey)))
	contents = append(contents, pubKeyLen...)
	contents = append(contents, config.PublicKey...)

	// CipherSuites (KDF + AEAD 对)
	cipherSuites := g.serializeCipherSuites()
	cipherSuitesLen := make([]byte, 2)
	binary.BigEndian.PutUint16(cipherSuitesLen, uint16(len(cipherSuites)))
	contents = append(contents, cipherSuitesLen...)
	contents = append(contents, cipherSuites...)

	// MaximumNameLength
	contents = append(contents, config.MaximumNameLength)

	// PublicName
	publicNameBytes := []byte(config.PublicName)
	contents = append(contents, byte(len(publicNameBytes)))
	contents = append(contents, publicNameBytes...)

	// Extensions（空列表）
	contents = append(contents, 0, 0) // 长度 = 0

	return contents, nil
}

// serializeCipherSuites 序列化密码套件列表
func (g *ConfigGenerator) serializeCipherSuites() []byte {
	var suites []byte

	// 为每对 KDF/AEAD 创建套件
	for _, kdf := range g.Options.KDFAlgorithms {
		for _, aead := range g.Options.AEADAlgorithms {
			kdfBytes := make([]byte, 2)
			binary.BigEndian.PutUint16(kdfBytes, kdf)
			suites = append(suites, kdfBytes...)

			aeadBytes := make([]byte, 2)
			binary.BigEndian.PutUint16(aeadBytes, aead)
			suites = append(suites, aeadBytes...)
		}
	}

	return suites
}

// GenerateECHConfigList 生成 ECH 配置列表
func (g *ConfigGenerator) GenerateECHConfigList(configCount int) (*ECHConfigList, error) {
	list := &ECHConfigList{
		Configs: make([]ECHConfigRecord, 0, configCount),
	}

	for i := 0; i < configCount; i++ {
		config, err := g.GenerateECHConfig()
		if err != nil {
			return nil, fmt.Errorf("generate config %d: %w", i, err)
		}
		list.Configs = append(list.Configs, *config)
	}

	return list, nil
}

// SerializeECHConfigList 序列化 ECH 配置列表
func SerializeECHConfigList(list *ECHConfigList) ([]byte, error) {
	var data []byte

	for _, config := range list.Configs {
		// 版本
		versionBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(versionBytes, config.Version)
		data = append(data, versionBytes...)

		// 长度
		lengthBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(lengthBytes, config.Length)
		data = append(data, lengthBytes...)

		// 内容
		data = append(data, config.Contents...)
	}

	// 添加总长度前缀
	result := make([]byte, 2)
	binary.BigEndian.PutUint16(result, uint16(len(data)))
	result = append(result, data...)

	return result, nil
}

// GenerateBase64ECHConfig 生成 Base64 编码的 ECH 配置
func GenerateBase64ECHConfig(opts ConfigOptions) (string, error) {
	generator := NewConfigGenerator(opts)
	config, err := generator.GenerateECHConfig()
	if err != nil {
		return "", err
	}

	list := &ECHConfigList{Configs: []ECHConfigRecord{*config}}
	data, err := SerializeECHConfigList(list)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// ECHKeySet ECH 密钥集（包含私钥和公钥配置）
type ECHKeySet struct {
	// 配置 ID
	ConfigID uint8

	// 公钥配置（用于分发给客户端）
	PublicConfig *ECHConfigRecord

	// 私钥（用于解密，服务器端保密）
	PrivateKey []byte

	// 创建时间
	CreatedAt int64

	// 过期时间
	ExpiresAt int64
}

// ECHKeyManager ECH 密钥管理器
type ECHKeyManager struct {
	// 当前活跃的密钥
	activeKeys map[uint8]*ECHKeySet

	// 历史密钥（用于解密旧请求）
	expiredKeys map[uint8]*ECHKeySet
}

// NewECHKeyManager 创建密钥管理器
func NewECHKeyManager() *ECHKeyManager {
	return &ECHKeyManager{
		activeKeys:  make(map[uint8]*ECHKeySet),
		expiredKeys: make(map[uint8]*ECHKeySet),
	}
}

// GenerateNewKey 生成新密钥
func (m *ECHKeyManager) GenerateNewKey(configID uint8, opts ConfigOptions) (*ECHKeySet, error) {
	generator := NewConfigGenerator(opts)
	config, err := generator.GenerateECHConfig()
	if err != nil {
		return nil, err
	}

	// 生成私钥（简化实现）
	_, privateKey, err := generator.generateKeyPair()
	if err != nil {
		return nil, err
	}

	keySet := &ECHKeySet{
		ConfigID:     configID,
		PublicConfig: config,
		PrivateKey:   privateKey,
	}

	m.activeKeys[configID] = keySet
	return keySet, nil
}

// GetPublicConfig 获取公钥配置（用于 ECH 扩展）
func (m *ECHKeyManager) GetPublicConfig(configID uint8) (*ECHConfigRecord, error) {
	keySet, ok := m.activeKeys[configID]
	if !ok {
		return nil, fmt.Errorf("config ID %d not found", configID)
	}
	return keySet.PublicConfig, nil
}

// DecryptECH 解密 ECH 加密数据（简化实现）
func (m *ECHKeyManager) DecryptECH(configID uint8, encryptedData []byte) ([]byte, error) {
	keySet, ok := m.activeKeys[configID]
	if !ok {
		// 尝试历史密钥
		keySet, ok = m.expiredKeys[configID]
		if !ok {
			return nil, fmt.Errorf("config ID %d not found", configID)
		}
	}

	// 实际实现需要使用 HPKE 解密
	// 这里仅作为占位
	_ = keySet
	return nil, fmt.Errorf("ECH decryption not implemented")
}

// RotateKeys 轮换密钥
func (m *ECHKeyManager) RotateKeys() {
	// 将当前活跃密钥移到历史密钥
	for id, keySet := range m.activeKeys {
		m.expiredKeys[id] = keySet
		delete(m.activeKeys, id)
	}
}

// GenerateECHPEM 生成 ECH 配置的 PEM 格式
func GenerateECHPEM(config *ECHConfigRecord, privateKey []byte) ([]byte, error) {
	// 序列化配置列表
	list := &ECHConfigList{Configs: []ECHConfigRecord{*config}}
	configData, err := SerializeECHConfigList(list)
	if err != nil {
		return nil, err
	}

	// 创建 PEM 块
	var pemBlocks []byte

	// 公钥配置块
	pubBlock := &pem.Block{
		Type:  "ECH CONFIG LIST",
		Bytes: configData,
	}
	pemBlocks = append(pemBlocks, pem.EncodeToMemory(pubBlock)...)

	// 私钥块（如果提供）
	if privateKey != nil {
		privBlock := &pem.Block{
			Type:  "ECH PRIVATE KEY",
			Bytes: privateKey,
		}
		pemBlocks = append(pemBlocks, pem.EncodeToMemory(privBlock)...)
	}

	return pemBlocks, nil
}

// GenerateRandomConfigID 生成随机配置 ID
func GenerateRandomConfigID() (uint8, error) {
	var id [1]byte
	if _, err := io.ReadFull(rand.Reader, id[:]); err != nil {
		return 0, err
	}
	return id[0], nil
}

// SupportedKEMs 返回支持的 KEM 算法列表
func SupportedKEMs() []KEMInfo {
	return []KEMInfo{
		{ID: 0x0020, Name: "X25519"},
		{ID: 0x0010, Name: "P-256"},
		{ID: 0x0011, Name: "P-384"},
		{ID: 0x0012, Name: "P-521"},
	}
}

// SupportedKDFs 返回支持的 KDF 算法列表
func SupportedKDFs() []KDFInfo {
	return []KDFInfo{
		{ID: 0x0001, Name: "HKDF-SHA256"},
		{ID: 0x0002, Name: "HKDF-SHA384"},
		{ID: 0x0003, Name: "HKDF-SHA512"},
	}
}

// SupportedAEADs 返回支持的 AEAD 算法列表
func SupportedAEADs() []AEADInfo {
	return []AEADInfo{
		{ID: 0x0001, Name: "AES-128-GCM"},
		{ID: 0x0002, Name: "AES-256-GCM"},
		{ID: 0x0003, Name: "ChaCha20Poly1305"},
	}
}

// KEMInfo KEM 算法信息
type KEMInfo struct {
	ID   uint16
	Name string
}

// KDFInfo KDF 算法信息
type KDFInfo struct {
	ID   uint16
	Name string
}

// AEADInfo AEAD 算法信息
type AEADInfo struct {
	ID   uint16
	Name string
}
