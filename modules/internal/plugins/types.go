// translated comment
package plugins

import (
	tls "github.com/bogdanfinn/utls"
)

// translated comment
type FingerprintCategory string

const (
	CategoryBrowser   FingerprintCategory = "browser"
	CategoryMobile    FingerprintCategory = "mobile"
	CategoryBot       FingerprintCategory = "bot"
	CategoryCustom    FingerprintCategory = "custom"
	CategoryCommunity FingerprintCategory = "community"
)

// translated comment
type FingerprintMetadata struct {
	Name           string              // translated comment
	DisplayName    string              // translated comment
	Description    string              // translated comment
	Category       FingerprintCategory // translated comment
	Version        string              // translated comment
	Browser        string              // translated comment
	BrowserVersion string              // translated comment
	OS             string              // translated comment
	IsMobile       bool                // translated comment
	Author         string              // translated comment
	License        string              // translated comment
	Tags           []string            // translated comment
	Verified       bool                // translated comment
}

// translated comment
type ClientHelloSpec struct {
	TLSVersion                uint16   // translated comment
	CipherSuites              []uint16 // translated comment
	Extensions                []uint16 // translated comment
	EllipticCurves            []uint16 // translated comment
	EllipticCurvePointFormats []uint8  // translated comment
	SignatureAlgorithms       []uint16 // translated comment
	SupportedVersions         []uint16 // translated comment
	KeyShareCurves            []uint16 // translated comment
}

// translated comment
type FingerprintData struct {
	Metadata    FingerprintMetadata    `json:"metadata"`
	ClientHello *ClientHelloSpec       `json:"client_hello"`
	UserAgent   string                 `json:"user_agent"`
	Extensions  map[string]interface{} `json:"extensions,omitempty"`
}

// translated comment
type Plugin interface {
	Metadata() *FingerprintMetadata
	Data() *FingerprintData
	Validate() error
	ToClientHelloSpec() (*tls.ClientHelloSpec, error)
	GetUserAgent() string
	Clone() Plugin
}

// translated comment
type PluginSource int

const (
	SourceBuiltin PluginSource = iota
	SourceLocal
	SourceRegistry
	SourceCommunity
)

// translated comment
type PluginInfo struct {
	ID      string
	Version string
	Source  PluginSource
	Meta    *FingerprintMetadata
	Plugin  Plugin
	Loaded  bool
	Error   error
}

// translated comment
type ValidationRule func(data *FingerprintData) error
