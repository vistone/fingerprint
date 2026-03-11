// Package plugins defines the fingerprint plugin system interfaces and types.
package plugins

import (
	tls "github.com/bogdanfinn/utls"
)

// FingerprintCategory classifies fingerprint profiles.
type FingerprintCategory string

const (
	CategoryBrowser   FingerprintCategory = "browser"
	CategoryMobile    FingerprintCategory = "mobile"
	CategoryBot       FingerprintCategory = "bot"
	CategoryCustom    FingerprintCategory = "custom"
	CategoryCommunity FingerprintCategory = "community"
)

// FingerprintMetadata holds metadata for a fingerprint profile.
type FingerprintMetadata struct {
	Name           string              // Fingerprint ID
	DisplayName    string              // Display name
	Description    string              // Description
	Category       FingerprintCategory // Category
	Version        string              // Version
	Browser        string              // Browser name
	BrowserVersion string              // Browser version
	OS             string              // Operating system
	IsMobile       bool                // Whether it is a mobile device
	Author         string              // Author
	License        string              // License
	Tags           []string            // Tags
	Verified       bool                // Whether verified
}

// ClientHelloSpec describes a TLS ClientHello specification.
type ClientHelloSpec struct {
	TLSVersion                uint16   // TLS version
	CipherSuites              []uint16 // Cipher suites
	Extensions                []uint16 // Extensions
	EllipticCurves            []uint16 // Elliptic curves
	EllipticCurvePointFormats []uint8  // EC point formats
	SignatureAlgorithms       []uint16 // Signature algorithms
	SupportedVersions         []uint16 // Supported versions
	KeyShareCurves            []uint16 // Key share curves
}

// FingerprintData holds fingerprint data in the standard format.
type FingerprintData struct {
	Metadata    FingerprintMetadata    `json:"metadata"`
	ClientHello *ClientHelloSpec       `json:"client_hello"`
	UserAgent   string                 `json:"user_agent"`
	Extensions  map[string]interface{} `json:"extensions,omitempty"`
}

// Plugin is the fingerprint plugin interface.
type Plugin interface {
	Metadata() *FingerprintMetadata
	Data() *FingerprintData
	Validate() error
	ToClientHelloSpec() (*tls.ClientHelloSpec, error)
	GetUserAgent() string
	Clone() Plugin
}

// PluginSource indicates where a plugin was loaded from.
type PluginSource int

const (
	SourceBuiltin PluginSource = iota
	SourceLocal
	SourceRegistry
	SourceCommunity
)

// PluginInfo holds runtime information about a loaded plugin.
type PluginInfo struct {
	ID      string
	Version string
	Source  PluginSource
	Meta    *FingerprintMetadata
	Plugin  Plugin
	Loaded  bool
	Error   error
}

// ValidationRule is a function that validates fingerprint data.
type ValidationRule func(data *FingerprintData) error
