// Package profiles 提供安全的指纹配置操作
package profiles

import (
	"github.com/vistone/fingerprint/modules/core"
)

// SafeGetUserAgent 安全获取 User-Agent（带 nil 检查）
func (p *ClientProfile) SafeGetUserAgent() string {
	if p == nil || p.Headers == nil {
		return ""
	}
	return p.Headers.UserAgent
}

// SafeGetHeader 安全获取指定 Header
func (p *ClientProfile) SafeGetHeader(key string) string {
	if p == nil || p.Headers == nil {
		return ""
	}
	
	switch key {
	case "Accept":
		return p.Headers.Accept
	case "Accept-Language":
		return p.Headers.AcceptLanguage
	case "Accept-Encoding":
		return p.Headers.AcceptEncoding
	case "User-Agent":
		return p.Headers.UserAgent
	default:
		if p.Headers.Custom != nil {
			return p.Headers.Custom[key]
		}
		return ""
	}
}

// Validate 验证指纹配置有效性
func (p *ClientProfile) Validate() error {
	if p == nil {
		return core.NewCodedError(core.ErrCodeNilPointer, "ClientProfile.Validate", nil)
	}
	
	validator := core.NewValidator()
	
	// 基本字段验证
	validator.NotEmpty(p.ID, "profile.ID").
		NotEmpty(p.Name, "profile.Name").
		ValidBrowserType(p.BrowserType, "profile.BrowserType").
		NotEmpty(p.BrowserVersion, "profile.BrowserVersion").
		ValidOS(p.OS, "profile.OS").
		NotEmpty(p.OSVersion, "profile.OSVersion")
	
	// 验证 TLS 版本
	if p.TLSVersion != 0x0301 && p.TLSVersion != 0x0302 && p.TLSVersion != 0x0303 && p.TLSVersion != 0x0304 {
		validator.AddErrorf("profile.TLSVersion is invalid: 0x%04x", p.TLSVersion)
	}
	
	// 验证 CipherSuites 非空
	if len(p.CipherSuites) == 0 {
		validator.AddErrorf("profile.CipherSuites cannot be empty")
	}
	
	return validator.Error()
}

// IsValid 检查指纹是否有效
func (p *ClientProfile) IsValid() bool {
	return p.Validate() == nil
}

// Clone 安全克隆指纹配置
func (p *ClientProfile) Clone() *ClientProfile {
	if p == nil {
		return nil
	}
	
	clone := &ClientProfile{
		ID:                 p.ID,
		Name:               p.Name,
		Description:        p.Description,
		BrowserType:        p.BrowserType,
		BrowserVersion:     p.BrowserVersion,
		OS:                 p.OS,
		OSVersion:          p.OSVersion,
		OSArch:             p.OSArch,
		OSBitness:          p.OSBitness,
		TLSVersion:         p.TLSVersion,
		CipherSuites:       make([]uint16, len(p.CipherSuites)),
		Extensions:         make([]core.TLSExtension, len(p.Extensions)),
		SupportedCurves:    make([]core.CurveID, len(p.SupportedCurves)),
		SupportedPoints:    make([]uint8, len(p.SupportedPoints)),
		HTTP2Settings:      p.HTTP2Settings,
		HTTP2Priorities:    make([]core.HTTP2Priority, len(p.HTTP2Priorities)),
		PseudoHeaderOrder:  make([]string, len(p.PseudoHeaderOrder)),
		ConnectionFlow:     p.ConnectionFlow,
		Metadata:           make(map[string]interface{}, len(p.Metadata)),
	}
	
	// 复制切片
	copy(clone.CipherSuites, p.CipherSuites)
	copy(clone.Extensions, p.Extensions)
	copy(clone.SupportedCurves, p.SupportedCurves)
	copy(clone.SupportedPoints, p.SupportedPoints)
	copy(clone.HTTP2Priorities, p.HTTP2Priorities)
	copy(clone.PseudoHeaderOrder, p.PseudoHeaderOrder)
	
	// 复制 Headers
	if p.Headers != nil {
		clone.Headers = p.Headers.Clone()
	}
	
	// 复制 Metadata
	for k, v := range p.Metadata {
		clone.Metadata[k] = v
	}
	
	return clone
}

// GetRegistrySafe 安全获取注册表（带 nil 检查）
func GetRegistrySafe() *ProfileRegistry {
	if DefaultRegistry == nil {
		DefaultRegistry = NewProfileRegistry()
	}
	return DefaultRegistry
}

// RegisterSafe 安全注册指纹
func RegisterSafe(profile ClientProfile) error {
	if err := profile.Validate(); err != nil {
		return err
	}
	GetRegistrySafe().Register(profile)
	return nil
}

// GetSafe 安全获取指纹
func GetSafe(id string) (ClientProfile, error) {
	if id == "" {
		return ClientProfile{}, core.NewCodedError(core.ErrCodeInvalidInput, "GetSafe", 
			core.ErrProfileNotFound)
	}
	
	p, ok := GetRegistrySafe().Get(id)
	if !ok {
		return ClientProfile{}, core.NewCodedErrorf(core.ErrCodeProfileNotFound, "GetSafe",
			"profile not found: %s", id)
	}
	
	return p, nil
}
