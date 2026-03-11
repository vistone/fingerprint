// Package plugins implements basic plugin types.
package plugins

import (
	"fmt"

	tls "github.com/bogdanfinn/utls"
)

// BasicPlugin is the default Plugin implementation backed by FingerprintData.
type BasicPlugin struct {
	data *FingerprintData
}

// NewBasicPlugin creates a new BasicPlugin from the given fingerprint data.
func NewBasicPlugin(data *FingerprintData) *BasicPlugin {
	return &BasicPlugin{data: data}
}

// Metadata returns the fingerprint metadata.
func (bp *BasicPlugin) Metadata() *FingerprintMetadata {
	if bp.data == nil {
		return nil
	}
	return &bp.data.Metadata
}

// Data returns the full fingerprint data.
func (bp *BasicPlugin) Data() *FingerprintData {
	return bp.data
}

// Validate checks that all required fields are present in the fingerprint data.
func (bp *BasicPlugin) Validate() error {
	if bp.data == nil {
		return fmt.Errorf("fingerprint data is nil")
	}

	if bp.data.Metadata.Name == "" {
		return fmt.Errorf("name is required")
	}

	if bp.data.Metadata.Category == "" {
		return fmt.Errorf("category is required")
	}

	if bp.data.ClientHello == nil {
		return fmt.Errorf("client_hello is required")
	}

	if bp.data.ClientHello.TLSVersion == 0 {
		return fmt.Errorf("tls_version is required")
	}

	if len(bp.data.ClientHello.CipherSuites) == 0 {
		return fmt.Errorf("cipher_suites cannot be empty")
	}

	if bp.data.UserAgent == "" {
		return fmt.Errorf("user_agent is required")
	}

	return nil
}

// ToClientHelloSpec converts the fingerprint data to a utls ClientHelloSpec.
func (bp *BasicPlugin) ToClientHelloSpec() (*tls.ClientHelloSpec, error) {
	if bp.data == nil || bp.data.ClientHello == nil {
		return nil, fmt.Errorf("no client hello data")
	}

	// Return a basic ClientHelloSpec.
	// Note: this is a simplified version; production use should adapt to utls requirements.
	spec := &tls.ClientHelloSpec{
		TLSVersMin:         bp.data.ClientHello.TLSVersion,
		TLSVersMax:         bp.data.ClientHello.TLSVersion,
		CipherSuites:       bp.data.ClientHello.CipherSuites,
		CompressionMethods: []uint8{0},
	}

	return spec, nil
}

// GetUserAgent returns the user-agent string.
func (bp *BasicPlugin) GetUserAgent() string {
	if bp.data == nil {
		return ""
	}
	return bp.data.UserAgent
}

// Clone returns a shallow copy of the plugin.
func (bp *BasicPlugin) Clone() Plugin {
	if bp.data == nil {
		return &BasicPlugin{}
	}

	dataCopy := *bp.data
	return &BasicPlugin{data: &dataCopy}
}
