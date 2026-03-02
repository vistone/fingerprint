// Package plugins 定义指纹插件系统接口和类型
package plugins

import (
	tls "github.com/bogdanfinn/utls"
)

// FingerprintCategory 指纹分类
type FingerprintCategory string

const (
	CategoryBrowser   FingerprintCategory = "browser"
	CategoryMobile    FingerprintCategory = "mobile"
	CategoryBot       FingerprintCategory = "bot"
	CategoryCustom    FingerprintCategory = "custom"
	CategoryCommunity FingerprintCategory = "community"
)

// FingerprintMetadata 指纹元数据
type FingerprintMetadata struct {
	Name           string              // 指纹ID
	DisplayName    string              // 显示名称
	Description    string              // 描述
	Category       FingerprintCategory // 分类
	Version        string              // 版本
	Browser        string              // 浏览器名称
	BrowserVersion string              // 浏览器版本
	OS             string              // 操作系统
	IsMobile       bool                // 是否移动设备
	Author         string              // 作者
	License        string              // 许可证
	Tags           []string            // 标签
	Verified       bool                // 是否验证
}

// ClientHelloSpec TLS ClientHello 规范
type ClientHelloSpec struct {
	TLSVersion                uint16   // TLS版本
	CipherSuites              []uint16 // 密码套件
	Extensions                []uint16 // 扩展
	EllipticCurves            []uint16 // 椭圆曲线
	EllipticCurvePointFormats []uint8  // EC点格式
	SignatureAlgorithms       []uint16 // 签名算法
	SupportedVersions         []uint16 // 支持版本
	KeyShareCurves            []uint16 // 密钥共享曲线
}

// FingerprintData 指纹数据（标准格式）
type FingerprintData struct {
	Metadata    FingerprintMetadata    `json:"metadata"`
	ClientHello *ClientHelloSpec       `json:"client_hello"`
	UserAgent   string                 `json:"user_agent"`
	Extensions  map[string]interface{} `json:"extensions,omitempty"`
}

// Plugin 指纹插件接口
type Plugin interface {
	Metadata() *FingerprintMetadata
	Data() *FingerprintData
	Validate() error
	ToClientHelloSpec() (*tls.ClientHelloSpec, error)
	GetUserAgent() string
	Clone() Plugin
}

// PluginSource 插件来源
type PluginSource int

const (
	SourceBuiltin PluginSource = iota
	SourceLocal
	SourceRegistry
	SourceCommunity
)

// PluginInfo 插件信息
type PluginInfo struct {
	ID      string
	Version string
	Source  PluginSource
	Meta    *FingerprintMetadata
	Plugin  Plugin
	Loaded  bool
	Error   error
}

// ValidationRule 验证规则
type ValidationRule func(data *FingerprintData) error
