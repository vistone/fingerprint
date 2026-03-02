// Package plugins 实现基础插件类型
package plugins

import (
	"fmt"

	tls "github.com/bogdanfinn/utls"
)

// BasicPlugin 基础插件实现
type BasicPlugin struct {
	data *FingerprintData
}

// NewBasicPlugin 创建新指纹
func NewBasicPlugin(data *FingerprintData) *BasicPlugin {
	return &BasicPlugin{data: data}
}

// Metadata 返回元数据
func (bp *BasicPlugin) Metadata() *FingerprintMetadata {
	if bp.data == nil {
		return nil
	}
	return &bp.data.Metadata
}

// Data 返回指纹数据
func (bp *BasicPlugin) Data() *FingerprintData {
	return bp.data
}

// Validate 验证指纹
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

// ToClientHelloSpec 转换为 utls spec
func (bp *BasicPlugin) ToClientHelloSpec() (*tls.ClientHelloSpec, error) {
	if bp.data == nil || bp.data.ClientHello == nil {
		return nil, fmt.Errorf("no client hello data")
	}

	// 返回一个基本的 ClientHelloSpec
	// 注意：这是一个简化版本，实际用途需要根据 utls 库的要求调整
	spec := &tls.ClientHelloSpec{
		TLSVersMin:         bp.data.ClientHello.TLSVersion,
		TLSVersMax:         bp.data.ClientHello.TLSVersion,
		CipherSuites:       bp.data.ClientHello.CipherSuites,
		CompressionMethods: []uint8{0},
	}

	return spec, nil
}

// GetUserAgent 获取 User-Agent
func (bp *BasicPlugin) GetUserAgent() string {
	if bp.data == nil {
		return ""
	}
	return bp.data.UserAgent
}

// Clone 克隆插件
func (bp *BasicPlugin) Clone() Plugin {
	if bp.data == nil {
		return &BasicPlugin{}
	}

	dataCopy := *bp.data
	return &BasicPlugin{data: &dataCopy}
}
