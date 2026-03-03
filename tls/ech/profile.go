package ech

import (
	"fmt"
)

// ECHProfile ECH 配置 Profile
type ECHProfile struct {
	// 是否启用 ECH
	Enabled bool

	// ECH 版本
	Version uint16

	// 公钥名称（用于外层 ClientHello）
	PublicName string

	// 最大域名长度
	MaxNameLength uint8

	// 支持的算法套件
	CipherSuites []KEMCipherSuite

	// 配置 ID
	ConfigID uint8

	// 是否使用 GREASE
	UseGREASE bool

	// GREASE 概率 (0-1)
	GREASEProbability float64

	// 外层 ClientHello 配置
	OuterHello OuterHelloConfig

	// 内层 ClientHello 配置
	InnerHello InnerHelloConfig
}

// OuterHelloConfig 外层 ClientHello 配置
type OuterHelloConfig struct {
	// TLS 版本
	TLSVersion uint16

	// Cipher Suites（外层使用的密码套件）
	CipherSuites []uint16

	// 扩展列表
	Extensions []uint16

	// 压缩方法
	CompressionMethods []uint8
}

// InnerHelloConfig 内层 ClientHello 配置
type InnerHelloConfig struct {
	// TLS 版本（通常为 1.3）
	TLSVersion uint16

	// Cipher Suites
	CipherSuites []uint16

	// 扩展列表
	Extensions []uint16
}

// DefaultECHProfile 返回默认 ECH Profile
func DefaultECHProfile() *ECHProfile {
	return &ECHProfile{
		Enabled:           true,
		Version:           ECHVersionDraft13,
		PublicName:        "cloudflare-ech.com",
		MaxNameLength:     128,
		UseGREASE:         true,
		GREASEProbability: 0.1, // 10% 概率使用 GREASE
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

// ChromeECHProfile Chrome 浏览器的 ECH 配置
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

// FirefoxECHProfile Firefox 浏览器的 ECH 配置
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

// SafariECHProfile Safari 浏览器的 ECH 配置
// 注意：Safari 目前实验性支持 ECH
func SafariECHProfile() *ECHProfile {
	profile := DefaultECHProfile()
	profile.Enabled = false // Safari 默认未启用
	profile.UseGREASE = true
	profile.GREASEProbability = 0.05
	return profile
}

// Validate 验证 ECH Profile
func (p *ECHProfile) Validate() error {
	if !p.Enabled {
		return nil // 未启用时无需验证
	}

	// 验证版本
	if p.Version != ECHVersionDraft13 &&
		p.Version != ECHVersionDraft14 &&
		p.Version != ECHVersionDraft15 {
		return fmt.Errorf("unsupported ECH version: 0x%04x", p.Version)
	}

	// 验证公钥名称
	if len(p.PublicName) == 0 || len(p.PublicName) > 255 {
		return fmt.Errorf("invalid public name length: %d", len(p.PublicName))
	}

	// 验证算法套件
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

// GenerateECHExtension 为 Profile 生成 ECH 扩展
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
		// 内层 ClientHello 很简单，只需要类型
		return ech, nil
	}

	// 外层 ClientHello
	ech.ClientHelloType = ECHClientHelloTypeOuter

	// 选择第一个算法套件
	if len(p.CipherSuites) > 0 {
		ech.CipherSuite = p.CipherSuites[0]
	}

	// Config ID
	ech.ConfigID = p.ConfigID

	// Encoded CH 应该是加密的内层 ClientHello
	// 实际应用中需要加密
	ech.EncodedCH = []byte("encrypted_inner_hello_placeholder")
	ech.EncodedCHLength = uint16(len(ech.EncodedCH))

	return ech, nil
}

// GenerateGREASEExtension 生成 GREASE ECH 扩展
func (p *ECHProfile) GenerateGREASEExtension() (*ECHExtension, error) {
	return &ECHExtension{
		Type:    ExtensionEncryptedClientHello,
		Version: 0x0000, // GREASE 版本
	}, nil
}

// ShouldUseGREASE 判断是否应使用 GREASE
func (p *ECHProfile) ShouldUseGREASE() bool {
	if !p.UseGREASE {
		return false
	}
	// 实际应用中使用随机数判断
	return true
}

// ECHProfileFromBrowser 根据浏览器类型获取 ECH Profile
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

// GetSupportedECHVersions 获取支持的 ECH 版本列表
func GetSupportedECHVersions() []uint16 {
	return []uint16{
		ECHVersionDraft13,
		ECHVersionDraft14,
		ECHVersionDraft15,
	}
}

// ECHVersionName 获取 ECH 版本名称
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

// MergeWithBase 将 ECH 配置合并到基础 Profile
func (p *ECHProfile) MergeWithBase(baseExtensions []uint16) []uint16 {
	if !p.Enabled {
		return baseExtensions
	}

	// 添加 ECH 扩展到扩展列表
	var merged []uint16
	echAdded := false

	for _, ext := range baseExtensions {
		// 避免重复添加 ECH 扩展
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
		// 在 SNI 扩展后添加 ECH 扩展（常见位置）
		sniIndex := -1
		for i, ext := range merged {
			if ext == 0 { // server_name
				sniIndex = i
				break
			}
		}

		if sniIndex >= 0 {
			// 在 SNI 后插入
			merged = append(merged[:sniIndex+1], append([]uint16{ExtensionEncryptedClientHello}, merged[sniIndex+1:]...)...)
		} else {
			// 添加到末尾
			merged = append(merged, ExtensionEncryptedClientHello)
		}
	}

	return merged
}

// ToYAMLConfig 转换为 YAML 配置格式
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
