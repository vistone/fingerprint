package ech

import (
	"fmt"
)

// translated comment
type ECHProfile struct {
	// translated comment
	Enabled bool

	// translated comment
	Version uint16

	// translated comment
	PublicName string

	// translated comment
	MaxNameLength uint8

	// translated comment
	CipherSuites []KEMCipherSuite

	// translated comment
	ConfigID uint8

	// translated comment
	UseGREASE bool

	// translated comment
	GREASEProbability float64

	// translated comment
	OuterHello OuterHelloConfig

	// translated comment
	InnerHello InnerHelloConfig
}

// translated comment
type OuterHelloConfig struct {
	// translated comment
	TLSVersion uint16

	// translated comment
	CipherSuites []uint16

	// translated comment
	Extensions []uint16

	// translated comment
	CompressionMethods []uint8
}

// translated comment
type InnerHelloConfig struct {
	// translated comment
	TLSVersion uint16

	// Cipher Suites
	CipherSuites []uint16

	// translated comment
	Extensions []uint16
}

// translated comment
func DefaultECHProfile() *ECHProfile {
	return &ECHProfile{
		Enabled:           true,
		Version:           ECHVersionDraft13,
		PublicName:        "cloudflare-ech.com",
		MaxNameLength:     128,
		UseGREASE:         true,
		GREASEProbability: 0.1, // translated comment
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

// translated comment
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

// translated comment
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

// translated comment
// translated comment
func SafariECHProfile() *ECHProfile {
	profile := DefaultECHProfile()
	profile.Enabled = false // translated comment
	profile.UseGREASE = true
	profile.GREASEProbability = 0.05
	return profile
}

// translated comment
func (p *ECHProfile) Validate() error {
	if !p.Enabled {
		return nil // translated comment
	}

	// translated comment
	if p.Version != ECHVersionDraft13 &&
		p.Version != ECHVersionDraft14 &&
		p.Version != ECHVersionDraft15 {
		return fmt.Errorf("unsupported ECH version: 0x%04x", p.Version)
	}

	// translated comment
	if len(p.PublicName) == 0 || len(p.PublicName) > 255 {
		return fmt.Errorf("invalid public name length: %d", len(p.PublicName))
	}

	// translated comment
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

// translated comment
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
		// translated comment
		return ech, nil
	}

	// translated comment
	ech.ClientHelloType = ECHClientHelloTypeOuter

	// translated comment
	if len(p.CipherSuites) > 0 {
		ech.CipherSuite = p.CipherSuites[0]
	}

	// Config ID
	ech.ConfigID = p.ConfigID

	// translated comment
	// translated comment
	ech.EncodedCH = []byte("encrypted_inner_hello_placeholder")
	ech.EncodedCHLength = uint16(len(ech.EncodedCH))

	return ech, nil
}

// translated comment
func (p *ECHProfile) GenerateGREASEExtension() (*ECHExtension, error) {
	return &ECHExtension{
		Type:    ExtensionEncryptedClientHello,
		Version: 0x0000, // translated comment
	}, nil
}

// translated comment
func (p *ECHProfile) ShouldUseGREASE() bool {
	if !p.UseGREASE {
		return false
	}
	// translated comment
	return true
}

// translated comment
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

// translated comment
func GetSupportedECHVersions() []uint16 {
	return []uint16{
		ECHVersionDraft13,
		ECHVersionDraft14,
		ECHVersionDraft15,
	}
}

// translated comment
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

// translated comment
func (p *ECHProfile) MergeWithBase(baseExtensions []uint16) []uint16 {
	if !p.Enabled {
		return baseExtensions
	}

	// translated comment
	var merged []uint16
	echAdded := false

	for _, ext := range baseExtensions {
		// translated comment
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
		// translated comment
		sniIndex := -1
		for i, ext := range merged {
			if ext == 0 { // server_name
				sniIndex = i
				break
			}
		}

		if sniIndex >= 0 {
			// translated comment
			merged = append(merged[:sniIndex+1], append([]uint16{ExtensionEncryptedClientHello}, merged[sniIndex+1:]...)...)
		} else {
			// translated comment
			merged = append(merged, ExtensionEncryptedClientHello)
		}
	}

	return merged
}

// translated comment
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
