// Package profiles 提供浏览器指纹配置管理
package profiles

import (
	"github.com/vistone/fingerprint/modules/core"
)

// ClientProfile 客户端指纹配置
type ClientProfile struct {
	// 核心标识
	ID          string
	Name        string
	Description string
	
	// 浏览器信息
	BrowserType    core.BrowserType
	BrowserVersion string
	
	// 操作系统信息
	OS        core.OperatingSystem
	OSVersion string
	OSArch    string
	OSBitness string
	
	// TLS 配置
	TLSVersion      uint16
	CipherSuites    []uint16
	Extensions      []core.TLSExtension
	SupportedCurves []core.CurveID
	SupportedPoints []uint8
	
	// HTTP/2 配置
	HTTP2Settings      core.HTTP2Settings
	HTTP2Priorities    []core.HTTP2Priority
	PseudoHeaderOrder  []string
	ConnectionFlow     uint32
	
	// HTTP 头配置
	Headers *core.HTTPHeaders
	
	// 元数据
	Metadata map[string]interface{}
}

// GetID 返回指纹唯一标识
func (p ClientProfile) GetID() string {
	return p.ID
}

// GetBrowserType 返回浏览器类型
func (p ClientProfile) GetBrowserType() core.BrowserType {
	return p.BrowserType
}

// GetOS 返回操作系统
func (p ClientProfile) GetOS() core.OperatingSystem {
	return p.OS
}

// GetUserAgent 返回 User-Agent
func (p ClientProfile) GetUserAgent() string {
	if p.Headers != nil {
		return p.Headers.UserAgent
	}
	return ""
}

// GetHeaders 返回 HTTP Headers
func (p ClientProfile) GetHeaders() *core.HTTPHeaders {
	return p.Headers
}

// GetSpec 返回指纹规范
func (p ClientProfile) GetSpec() core.FingerprintSpec {
	return p
}

// ProfileRegistry 指纹注册表
type ProfileRegistry struct {
	profiles map[string]ClientProfile
}

// NewProfileRegistry 创建新的指纹注册表
func NewProfileRegistry() *ProfileRegistry {
	return &ProfileRegistry{
		profiles: make(map[string]ClientProfile),
	}
}

// Register 注册指纹
func (r *ProfileRegistry) Register(profile ClientProfile) {
	r.profiles[profile.ID] = profile
}

// Get 获取指纹
func (r *ProfileRegistry) Get(id string) (ClientProfile, bool) {
	p, ok := r.profiles[id]
	return p, ok
}

// GetByBrowser 按浏览器类型获取所有指纹
func (r *ProfileRegistry) GetByBrowser(browser core.BrowserType) []ClientProfile {
	var result []ClientProfile
	for _, p := range r.profiles {
		if p.BrowserType == browser {
			result = append(result, p)
		}
	}
	return result
}

// GetByOS 按操作系统获取所有指纹
func (r *ProfileRegistry) GetByOS(os core.OperatingSystem) []ClientProfile {
	var result []ClientProfile
	for _, p := range r.profiles {
		if p.OS == os {
			result = append(result, p)
		}
	}
	return result
}

// GetAll 获取所有指纹
func (r *ProfileRegistry) GetAll() []ClientProfile {
	result := make([]ClientProfile, 0, len(r.profiles))
	for _, p := range r.profiles {
		result = append(result, p)
	}
	return result
}

// Count 返回指纹数量
func (r *ProfileRegistry) Count() int {
	return len(r.profiles)
}

// DefaultRegistry 默认全局注册表
var DefaultRegistry = NewProfileRegistry()

// Register 向默认注册表注册指纹
func Register(profile ClientProfile) {
	DefaultRegistry.Register(profile)
}

// Get 从默认注册表获取指纹
func Get(id string) (ClientProfile, bool) {
	return DefaultRegistry.Get(id)
}

// GetByBrowser 从默认注册表按浏览器类型获取指纹
func GetByBrowser(browser core.BrowserType) []ClientProfile {
	return DefaultRegistry.GetByBrowser(browser)
}

// GetAll 从默认注册表获取所有指纹
func GetAll() []ClientProfile {
	return DefaultRegistry.GetAll()
}

// GetRandom 随机获取一个指纹
func GetRandom() ClientProfile {
	all := DefaultRegistry.GetAll()
	if len(all) == 0 {
		return ClientProfile{}
	}
	return core.RandomChoice(all)
}

// GetRandomByBrowser 按浏览器类型随机获取指纹
func GetRandomByBrowser(browser core.BrowserType) ClientProfile {
	profiles := DefaultRegistry.GetByBrowser(browser)
	if len(profiles) == 0 {
		return ClientProfile{}
	}
	return core.RandomChoice(profiles)
}
