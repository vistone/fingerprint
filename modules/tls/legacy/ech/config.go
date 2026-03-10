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

// translated comment
type ConfigGenerator struct {
	// translated comment
	Options ConfigOptions
}

// translated comment
type ConfigOptions struct {
	// translated comment
	Version uint16

	// translated comment
	PublicName string

	// translated comment
	MaxNameLength uint8

	// translated comment
	KEMAlgorithms []uint16

	// translated comment
	KDFAlgorithms []uint16

	// translated comment
	AEADAlgorithms []uint16

	// translated comment
	Extensions []ECHConfigExtension
}

// translated comment
type ECHConfigExtension struct {
	Type uint16
	Data []byte
}

// translated comment
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

// translated comment
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

// translated comment
func (g *ConfigGenerator) GenerateECHConfig() (*ECHConfigRecord, error) {
	// translated comment
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

	// translated comment
	if len(g.Options.KEMAlgorithms) > 0 {
		config.KemID = g.Options.KEMAlgorithms[0]
	}
	if len(g.Options.KDFAlgorithms) > 0 {
		config.KdfID = g.Options.KDFAlgorithms[0]
	}
	if len(g.Options.AEADAlgorithms) > 0 {
		config.AeadID = g.Options.AEADAlgorithms[0]
	}

	// translated comment
	contents, err := g.serializeConfigContents(config, privateKey)
	if err != nil {
		return nil, fmt.Errorf("serialize config: %w", err)
	}

	config.Contents = contents
	config.Length = uint16(len(contents))

	return config, nil
}

// translated comment
func (g *ConfigGenerator) generateKeyPair() (privateKey, publicKey []byte, err error) {
	// translated comment
	// translated comment
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// translated comment
	privateKey = x509.MarshalPKCS1PrivateKey(key)

	// translated comment
	publicKey, err = x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, publicKey, nil
}

// translated comment
func (g *ConfigGenerator) serializeConfigContents(config *ECHConfigRecord, privateKey []byte) ([]byte, error) {
	// translated comment
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

	// translated comment
	pubKeyLen := make([]byte, 2)
	binary.BigEndian.PutUint16(pubKeyLen, uint16(len(config.PublicKey)))
	contents = append(contents, pubKeyLen...)
	contents = append(contents, config.PublicKey...)

	// translated comment
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

	// translated comment
	contents = append(contents, 0, 0) // translated comment

	return contents, nil
}

// translated comment
func (g *ConfigGenerator) serializeCipherSuites() []byte {
	var suites []byte

	// translated comment
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

// translated comment
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

// translated comment
func SerializeECHConfigList(list *ECHConfigList) ([]byte, error) {
	var data []byte

	for _, config := range list.Configs {
		// translated comment
		versionBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(versionBytes, config.Version)
		data = append(data, versionBytes...)

		// translated comment
		lengthBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(lengthBytes, config.Length)
		data = append(data, lengthBytes...)

		// translated comment
		data = append(data, config.Contents...)
	}

	// translated comment
	result := make([]byte, 2)
	binary.BigEndian.PutUint16(result, uint16(len(data)))
	result = append(result, data...)

	return result, nil
}

// translated comment
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

// translated comment
type ECHKeySet struct {
	// translated comment
	ConfigID uint8

	// translated comment
	PublicConfig *ECHConfigRecord

	// translated comment
	PrivateKey []byte

	// translated comment
	CreatedAt int64

	// translated comment
	ExpiresAt int64
}

// translated comment
type ECHKeyManager struct {
	// translated comment
	activeKeys map[uint8]*ECHKeySet

	// translated comment
	expiredKeys map[uint8]*ECHKeySet
}

// translated comment
func NewECHKeyManager() *ECHKeyManager {
	return &ECHKeyManager{
		activeKeys:  make(map[uint8]*ECHKeySet),
		expiredKeys: make(map[uint8]*ECHKeySet),
	}
}

// translated comment
func (m *ECHKeyManager) GenerateNewKey(configID uint8, opts ConfigOptions) (*ECHKeySet, error) {
	generator := NewConfigGenerator(opts)
	config, err := generator.GenerateECHConfig()
	if err != nil {
		return nil, err
	}

	// translated comment
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

// translated comment
func (m *ECHKeyManager) GetPublicConfig(configID uint8) (*ECHConfigRecord, error) {
	keySet, ok := m.activeKeys[configID]
	if !ok {
		return nil, fmt.Errorf("config ID %d not found", configID)
	}
	return keySet.PublicConfig, nil
}

// translated comment
func (m *ECHKeyManager) DecryptECH(configID uint8, encryptedData []byte) ([]byte, error) {
	keySet, ok := m.activeKeys[configID]
	if !ok {
		// translated comment
		keySet, ok = m.expiredKeys[configID]
		if !ok {
			return nil, fmt.Errorf("config ID %d not found", configID)
		}
	}

	// translated comment
	// translated comment
	_ = keySet
	return nil, fmt.Errorf("ECH decryption not implemented")
}

// translated comment
func (m *ECHKeyManager) RotateKeys() {
	// translated comment
	for id, keySet := range m.activeKeys {
		m.expiredKeys[id] = keySet
		delete(m.activeKeys, id)
	}
}

// translated comment
func GenerateECHPEM(config *ECHConfigRecord, privateKey []byte) ([]byte, error) {
	// translated comment
	list := &ECHConfigList{Configs: []ECHConfigRecord{*config}}
	configData, err := SerializeECHConfigList(list)
	if err != nil {
		return nil, err
	}

	// translated comment
	var pemBlocks []byte

	// translated comment
	pubBlock := &pem.Block{
		Type:  "ECH CONFIG LIST",
		Bytes: configData,
	}
	pemBlocks = append(pemBlocks, pem.EncodeToMemory(pubBlock)...)

	// translated comment
	if privateKey != nil {
		privBlock := &pem.Block{
			Type:  "ECH PRIVATE KEY",
			Bytes: privateKey,
		}
		pemBlocks = append(pemBlocks, pem.EncodeToMemory(privBlock)...)
	}

	return pemBlocks, nil
}

// translated comment
func GenerateRandomConfigID() (uint8, error) {
	var id [1]byte
	if _, err := io.ReadFull(rand.Reader, id[:]); err != nil {
		return 0, err
	}
	return id[0], nil
}

// translated comment
func SupportedKEMs() []KEMInfo {
	return []KEMInfo{
		{ID: 0x0020, Name: "X25519"},
		{ID: 0x0010, Name: "P-256"},
		{ID: 0x0011, Name: "P-384"},
		{ID: 0x0012, Name: "P-521"},
	}
}

// translated comment
func SupportedKDFs() []KDFInfo {
	return []KDFInfo{
		{ID: 0x0001, Name: "HKDF-SHA256"},
		{ID: 0x0002, Name: "HKDF-SHA384"},
		{ID: 0x0003, Name: "HKDF-SHA512"},
	}
}

// translated comment
func SupportedAEADs() []AEADInfo {
	return []AEADInfo{
		{ID: 0x0001, Name: "AES-128-GCM"},
		{ID: 0x0002, Name: "AES-256-GCM"},
		{ID: 0x0003, Name: "ChaCha20Poly1305"},
	}
}

// translated comment
type KEMInfo struct {
	ID   uint16
	Name string
}

// translated comment
type KDFInfo struct {
	ID   uint16
	Name string
}

// translated comment
type AEADInfo struct {
	ID   uint16
	Name string
}
