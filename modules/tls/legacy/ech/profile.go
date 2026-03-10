package ech

import (
	"fmt"
)

// ECHProfile ECH configuration profile
type ECHProfile struct {
	// Whether ECH is enabled
	Enabled bool

	// ECH version
	Version uint16

	// Public key name (used for outer ClientHello)
	PublicName string

	// Maximum domain name length
	MaxNameLength uint8

	// Supported algorithm suites
	CipherSuites []KEMCipherSuite

	// Config ID
	ConfigID uint8

	// Whether to use GREASE
	UseGREASE bool

	// GREASE probability (0-1)
	GREASEProbability float64

	// Outer ClientHello configuration
	OuterHello OuterHelloConfig

	// Inner ClientHello configuration
	InnerHello InnerHelloConfig
}

// OuterHelloConfig outer ClientHello configuration
type OuterHelloConfig struct {
	// TLS version
	TLSVersion uint16

	// Cipher Suites (cipher suites used by outer)
	CipherSuites []uint16

	// Extension list
	Extensions []uint16

	// Compression methods
	CompressionMethods []uint8
}

// InnerHelloConfig inner ClientHello configuration
type InnerHelloConfig struct {
	// TLS version (typically 1.3)
	TLSVersion uint16

	// Cipher Suites
	CipherSuites []uint16

	// Extension list
	Extensions []uint16
}

// DefaultECHProfile returns the default ECH profile
func DefaultECHProfile() *ECHProfile {
	return &ECHProfile{
		Enabled:           true,
		Version:           ECHVersionDraft13,
		PublicName:        "cloudflare-ech.com",
		MaxNameLength:     128,
		UseGREASE:         true,
		GREASEProbability: 0.1, // 10% probability of using GREASE
		CipherSuites: []KEMCipherSuite{
			{KDFID: 0x0001, AEADID: 0x0001}, // HKDF-SHA256 + AES-128-GCM
			{KDFID: 0x0001, AEADID: 0x0002}, // HKDF-SHA256 + AES-256-GCM
		},
		OuterHello: OuterHelloConfig{
			TLSVersion:         0x0303,                           // TLS 1.2
			CipherSuites:       []uint16{0x1301, 0x1302, 0x1303}, // TLS 1.3 suites
			Extensions:         []uint16{0, 10, 11, 13, 16, 23, 43, 45, 51, 0xfe0d},
			CompressionMethods: []uint8{0},
		},
		InnerHello: InnerHelloConfig{
			TLSVersion: 0x0304, // TLS 1.3
			CipherSuites: []uint16{
				0x1301, // TLS_AES_128_GCM_SHA256
				0x1302, // TLS_AES_256_GCM_SHA384
				0x1303, // TLS_CHACHA20_POLY1305_SHA256
			},
			Extensions: []uint16{0, 5, 10, 13, 16, 18, 23, 27, 35, 43, 45, 51},
		},
	}
}

// ChromeECHProfile Chrome browser ECH configuration
func ChromeECHProfile() *ECHProfile {
	profile := DefaultECHProfile()
	profile.PublicName = "google-ech.cloudflareresearch.com"
	profile.MaxNameLength = 128
	profile.CipherSuites = []KEMCipherSuite{
		{KDFID: 0x0001, AEADID: 0x0001},
		{KDFID: 0x0001, AEADID: 0x0002},
	}
	profile.OuterHello = OuterHelloConfig{
		TLSVersion:         0x0303,
		CipherSuites:       []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f},
		Extensions:         []uint16{0, 5, 10, 11, 13, 16, 18, 21, 23, 27, 35, 43, 45, 51, 0xfe0d, 0x0a0a},
		CompressionMethods: []uint8{0},
	}
	return profile
}

// FirefoxECHProfile Firefox browser ECH configuration
func FirefoxECHProfile() *ECHProfile {
	profile := DefaultECHProfile()
	profile.PublicName = "firefox-ech.cloudflareresearch.com"
	profile.MaxNameLength = 128
	profile.CipherSuites = []KEMCipherSuite{
		{KDFID: 0x0001, AEADID: 0x0001},
	}
	profile.OuterHello = OuterHelloConfig{
		TLSVersion:         0x0303,
		CipherSuites:       []uint16{0x1301, 0x1302, 0x1303},
		Extensions:         []uint16{0, 5, 10, 11, 13, 16, 23, 34, 43, 45, 51, 0xfe0d},
		CompressionMethods: []uint8{0},
	}
	return profile
}

// SafariECHProfile Safari browser ECH configuration
// Note: Safari currently has experimental ECH support
func SafariECHProfile() *ECHProfile {
	profile := DefaultECHProfile()
	profile.Enabled = false // Safari is not enabled by default
	profile.UseGREASE = true
	profile.GREASEProbability = 0.05
	return profile
}

// Validate validates the ECH profile
func (p *ECHProfile) Validate() error {
	if !p.Enabled {
		return nil // No validation needed when disabled
	}

	// Validate version
	if p.Version != ECHVersionDraft13 &&
		p.Version != ECHVersionDraft14 &&
		p.Version != ECHVersionDraft15 {
		return fmt.Errorf("unsupported ECH version: 0x%04x", p.Version)
	}

	// Validate public key name
	if len(p.PublicName) == 0 || len(p.PublicName) > 255 {
		return fmt.Errorf("invalid public name length: %d", len(p.PublicName))
	}

	// Validate algorithm suites
	if len(p.CipherSuites) == 0 {
		return fmt.Errorf("at least one cipher suite required")
	}

	for _, cs := range p.CipherSuites {
		if cs.KDFID == 0 || cs.AEADID == 0 {
			return fmt.Errorf("invalid cipher suite: KDF=%d, AEAD=%d", cs.KDFID, cs.AEADID)
		}
	}

	return nil
}

// GenerateECHExtension generates an ECH extension for the profile
func (p *ECHProfile) GenerateECHExtension(isInner bool) (*ECHExtension, error) {
	if !p.Enabled {
		return nil, fmt.Errorf("ECH not enabled in profile")
	}

	ech := &ECHExtension{
		Type:    ExtensionEncryptedClientHello,
		Version: p.Version,
	}

	if isInner {
		ech.ClientHelloType = ECHClientHelloTypeInner
		// Inner ClientHello is simple, only needs type
		return ech, nil
	}

	// Outer ClientHello
	ech.ClientHelloType = ECHClientHelloTypeOuter

	// Select the first algorithm suite
	if len(p.CipherSuites) > 0 {
		ech.CipherSuite = p.CipherSuites[0]
	}

	// Config ID
	ech.ConfigID = p.ConfigID

	// Encoded CH should be the encrypted inner ClientHello
	// Encryption is needed in practice
	ech.EncodedCH = []byte("encrypted_inner_hello_placeholder")
	ech.EncodedCHLength = uint16(len(ech.EncodedCH))

	return ech, nil
}

// GenerateGREASEExtension generates a GREASE ECH extension
func (p *ECHProfile) GenerateGREASEExtension() (*ECHExtension, error) {
	return &ECHExtension{
		Type:    ExtensionEncryptedClientHello,
		Version: 0x0000, // GREASE version
	}, nil
}

// ShouldUseGREASE determines whether GREASE should be used
func (p *ECHProfile) ShouldUseGREASE() bool {
	if !p.UseGREASE {
		return false
	}
	// In practice, use random numbers to determine
	return true
}

// ECHProfileFromBrowser returns an ECH profile based on browser type
func ECHProfileFromBrowser(browser string) *ECHProfile {
	switch browser {
	case "chrome", "Chrome":
		return ChromeECHProfile()
	case "firefox", "Firefox":
		return FirefoxECHProfile()
	case "safari", "Safari":
		return SafariECHProfile()
	default:
		return DefaultECHProfile()
	}
}

// GetSupportedECHVersions returns the list of supported ECH versions
func GetSupportedECHVersions() []uint16 {
	return []uint16{
		ECHVersionDraft13,
		ECHVersionDraft14,
		ECHVersionDraft15,
	}
}

// ECHVersionName returns the ECH version name
func ECHVersionName(version uint16) string {
	switch version {
	case ECHVersionDraft13:
		return "Draft 13"
	case ECHVersionDraft14:
		return "Draft 14"
	case ECHVersionDraft15:
		return "Draft 15"
	default:
		return fmt.Sprintf("Unknown(0x%04x)", version)
	}
}

// MergeWithBase merges ECH configuration into the base profile
func (p *ECHProfile) MergeWithBase(baseExtensions []uint16) []uint16 {
	if !p.Enabled {
		return baseExtensions
	}

	// Add ECH extension to the extension list
	var merged []uint16
	echAdded := false

	for _, ext := range baseExtensions {
		// Avoid adding duplicate ECH extensions
		if ext == ExtensionEncryptedClientHello || ext == ExtensionECHOuterExtensions {
			if !echAdded {
				merged = append(merged, ExtensionEncryptedClientHello)
				echAdded = true
			}
			continue
		}
		merged = append(merged, ext)
	}

	if !echAdded {
		// Add ECH extension after SNI extension (common position)
		sniIndex := -1
		for i, ext := range merged {
			if ext == 0 { // server_name
				sniIndex = i
				break
			}
		}

		if sniIndex >= 0 {
			// Insert after SNI
			merged = append(merged[:sniIndex+1], append([]uint16{ExtensionEncryptedClientHello}, merged[sniIndex+1:]...)...)
		} else {
			// Append to end
			merged = append(merged, ExtensionEncryptedClientHello)
		}
	}

	return merged
}

// ToYAMLConfig converts to YAML configuration format
func (p *ECHProfile) ToYAMLConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled":            p.Enabled,
		"version":            fmt.Sprintf("0x%04x", p.Version),
		"public_name":        p.PublicName,
		"max_name_length":    p.MaxNameLength,
		"use_grease":         p.UseGREASE,
		"grease_probability": p.GREASEProbability,
		"outer_hello": map[string]interface{}{
			"tls_version":   fmt.Sprintf("0x%04x", p.OuterHello.TLSVersion),
			"cipher_suites": p.OuterHello.CipherSuites,
			"extensions":    p.OuterHello.Extensions,
			"compression":   p.OuterHello.CompressionMethods,
		},
		"inner_hello": map[string]interface{}{
			"tls_version":   fmt.Sprintf("0x%04x", p.InnerHello.TLSVersion),
			"cipher_suites": p.InnerHello.CipherSuites,
			"extensions":    p.InnerHello.Extensions,
		},
	}
}
