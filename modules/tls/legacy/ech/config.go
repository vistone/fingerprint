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

// ConfigGenerator ECH config generator
type ConfigGenerator struct {
	// Config options
	Options ConfigOptions
}

// ConfigOptions ECH config options
type ConfigOptions struct {
	// ECH version
	Version uint16

	// Public name (used for the outer ClientHello SNI)
	PublicName string

	// Maximum domain name length
	MaxNameLength uint8

	// Supported KEM algorithms
	KEMAlgorithms []uint16

	// Supported KDF algorithms
	KDFAlgorithms []uint16

	// Supported AEAD algorithms
	AEADAlgorithms []uint16

	// Extension list
	Extensions []ECHConfigExtension
}

// ECHConfigExtension ECH config extension
type ECHConfigExtension struct {
	Type uint16
	Data []byte
}

// DefaultConfigOptions returns default config options
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

// NewConfigGenerator creates a config generator
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

// GenerateECHConfig generates an ECH config
func (g *ConfigGenerator) GenerateECHConfig() (*ECHConfigRecord, error) {
	// Generate key pair (simplified version, should use HPKE in practice)
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

	// Select algorithms
	if len(g.Options.KEMAlgorithms) > 0 {
		config.KemID = g.Options.KEMAlgorithms[0]
	}
	if len(g.Options.KDFAlgorithms) > 0 {
		config.KdfID = g.Options.KDFAlgorithms[0]
	}
	if len(g.Options.AEADAlgorithms) > 0 {
		config.AeadID = g.Options.AEADAlgorithms[0]
	}

	// Serialize config contents
	contents, err := g.serializeConfigContents(config, privateKey)
	if err != nil {
		return nil, fmt.Errorf("serialize config: %w", err)
	}

	config.Contents = contents
	config.Length = uint16(len(contents))

	return config, nil
}

// generateKeyPair generates a key pair (simplified implementation)
func (g *ConfigGenerator) generateKeyPair() (privateKey, publicKey []byte, err error) {
	// Actual implementation should use HPKE
	// Using RSA as a placeholder here
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Serialize private key
	privateKey = x509.MarshalPKCS1PrivateKey(key)

	// Serialize public key
	publicKey, err = x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, publicKey, nil
}

// serializeConfigContents serializes config contents
func (g *ConfigGenerator) serializeConfigContents(config *ECHConfigRecord, privateKey []byte) ([]byte, error) {
	// ECH config contents format (Draft 13):
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

	// PublicKey (2-byte length prefix)
	pubKeyLen := make([]byte, 2)
	binary.BigEndian.PutUint16(pubKeyLen, uint16(len(config.PublicKey)))
	contents = append(contents, pubKeyLen...)
	contents = append(contents, config.PublicKey...)

	// CipherSuites (KDF + AEAD pairs)
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

	// Extensions (empty list)
	contents = append(contents, 0, 0) // length = 0

	return contents, nil
}

// serializeCipherSuites serializes the cipher suite list
func (g *ConfigGenerator) serializeCipherSuites() []byte {
	var suites []byte

	// Create a suite for each KDF/AEAD pair
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

// GenerateECHConfigList generates an ECH config list
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

// SerializeECHConfigList serializes an ECH config list
func SerializeECHConfigList(list *ECHConfigList) ([]byte, error) {
	var data []byte

	for _, config := range list.Configs {
		// Version
		versionBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(versionBytes, config.Version)
		data = append(data, versionBytes...)

		// Length
		lengthBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(lengthBytes, config.Length)
		data = append(data, lengthBytes...)

		// Contents
		data = append(data, config.Contents...)
	}

	// Add total length prefix
	result := make([]byte, 2)
	binary.BigEndian.PutUint16(result, uint16(len(data)))
	result = append(result, data...)

	return result, nil
}

// GenerateBase64ECHConfig generates a Base64-encoded ECH config
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

// ECHKeySet ECH key set (contains private key and public key config)
type ECHKeySet struct {
	// Config ID
	ConfigID uint8

	// Public key config (for distribution to clients)
	PublicConfig *ECHConfigRecord

	// Private key (for decryption, kept secret on the server side)
	PrivateKey []byte

	// Creation time
	CreatedAt int64

	// Expiration time
	ExpiresAt int64
}

// ECHKeyManager ECH key manager
type ECHKeyManager struct {
	// Currently active keys
	activeKeys map[uint8]*ECHKeySet

	// Historical keys (for decrypting old requests)
	expiredKeys map[uint8]*ECHKeySet
}

// NewECHKeyManager creates a key manager
func NewECHKeyManager() *ECHKeyManager {
	return &ECHKeyManager{
		activeKeys:  make(map[uint8]*ECHKeySet),
		expiredKeys: make(map[uint8]*ECHKeySet),
	}
}

// GenerateNewKey generates a new key
func (m *ECHKeyManager) GenerateNewKey(configID uint8, opts ConfigOptions) (*ECHKeySet, error) {
	generator := NewConfigGenerator(opts)
	config, err := generator.GenerateECHConfig()
	if err != nil {
		return nil, err
	}

	// Generate private key (simplified implementation)
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

// GetPublicConfig returns the public key config (for ECH extension)
func (m *ECHKeyManager) GetPublicConfig(configID uint8) (*ECHConfigRecord, error) {
	keySet, ok := m.activeKeys[configID]
	if !ok {
		return nil, fmt.Errorf("config ID %d not found", configID)
	}
	return keySet.PublicConfig, nil
}

// DecryptECH decrypts ECH encrypted data (simplified implementation)
func (m *ECHKeyManager) DecryptECH(configID uint8, encryptedData []byte) ([]byte, error) {
	keySet, ok := m.activeKeys[configID]
	if !ok {
		// Try historical keys
		keySet, ok = m.expiredKeys[configID]
		if !ok {
			return nil, fmt.Errorf("config ID %d not found", configID)
		}
	}

	// Actual implementation needs to use HPKE decryption
	// This is only a placeholder
	_ = keySet
	return nil, fmt.Errorf("ECH decryption not implemented")
}

// RotateKeys rotates keys
func (m *ECHKeyManager) RotateKeys() {
	// Move currently active keys to historical keys
	for id, keySet := range m.activeKeys {
		m.expiredKeys[id] = keySet
		delete(m.activeKeys, id)
	}
}

// GenerateECHPEM generates ECH config in PEM format
func GenerateECHPEM(config *ECHConfigRecord, privateKey []byte) ([]byte, error) {
	// Serialize config list
	list := &ECHConfigList{Configs: []ECHConfigRecord{*config}}
	configData, err := SerializeECHConfigList(list)
	if err != nil {
		return nil, err
	}

	// Create PEM blocks
	var pemBlocks []byte

	// Public key config block
	pubBlock := &pem.Block{
		Type:  "ECH CONFIG LIST",
		Bytes: configData,
	}
	pemBlocks = append(pemBlocks, pem.EncodeToMemory(pubBlock)...)

	// Private key block (if provided)
	if privateKey != nil {
		privBlock := &pem.Block{
			Type:  "ECH PRIVATE KEY",
			Bytes: privateKey,
		}
		pemBlocks = append(pemBlocks, pem.EncodeToMemory(privBlock)...)
	}

	return pemBlocks, nil
}

// GenerateRandomConfigID generates a random config ID
func GenerateRandomConfigID() (uint8, error) {
	var id [1]byte
	if _, err := io.ReadFull(rand.Reader, id[:]); err != nil {
		return 0, err
	}
	return id[0], nil
}

// SupportedKEMs returns the list of supported KEM algorithms
func SupportedKEMs() []KEMInfo {
	return []KEMInfo{
		{ID: 0x0020, Name: "X25519"},
		{ID: 0x0010, Name: "P-256"},
		{ID: 0x0011, Name: "P-384"},
		{ID: 0x0012, Name: "P-521"},
	}
}

// SupportedKDFs returns the list of supported KDF algorithms
func SupportedKDFs() []KDFInfo {
	return []KDFInfo{
		{ID: 0x0001, Name: "HKDF-SHA256"},
		{ID: 0x0002, Name: "HKDF-SHA384"},
		{ID: 0x0003, Name: "HKDF-SHA512"},
	}
}

// SupportedAEADs returns the list of supported AEAD algorithms
func SupportedAEADs() []AEADInfo {
	return []AEADInfo{
		{ID: 0x0001, Name: "AES-128-GCM"},
		{ID: 0x0002, Name: "AES-256-GCM"},
		{ID: 0x0003, Name: "ChaCha20Poly1305"},
	}
}

// KEMInfo KEM algorithm info
type KEMInfo struct {
	ID   uint16
	Name string
}

// KDFInfo KDF algorithm info
type KDFInfo struct {
	ID   uint16
	Name string
}

// AEADInfo AEAD algorithm info
type AEADInfo struct {
	ID   uint16
	Name string
}
