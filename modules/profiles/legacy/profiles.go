// Package profiles 包含浏览器 TLS 指纹配置。
// 注意：本包使用了来自 utls 库的无键字段结构体初始化，
// 这是为了与库的设计兼容，由此产生的 go vet 警告应被忽略。
//
//nolint:composites
package profiles

import (
	"sync"

	"github.com/bogdanfinn/fhttp/http2"
	tls "github.com/bogdanfinn/utls"
)

var DefaultClientProfile = Chrome_133

// MappedTLSClients 存储所有 TLS 客户端指纹配置（并发安全）
var MappedTLSClients = map[string]ClientProfile{
	"chrome_103":        Chrome_103,
	"chrome_104":        Chrome_104,
	"chrome_105":        Chrome_105,
	"chrome_106":        Chrome_106,
	"chrome_107":        Chrome_107,
	"chrome_108":        Chrome_108,
	"chrome_109":        Chrome_109,
	"chrome_110":        Chrome_110,
	"chrome_111":        Chrome_111,
	"chrome_112":        Chrome_112,
	"chrome_116_PSK":    Chrome_116_PSK,
	"chrome_116_PSK_PQ": Chrome_116_PSK_PQ,
	"chrome_117":        Chrome_117,
	"chrome_120":        Chrome_120,
	"chrome_124":        Chrome_124,
	"chrome_130_PSK":    Chrome_130_PSK,
	"chrome_131":        Chrome_131,
	"chrome_131_PSK":    Chrome_131_PSK,
	"chrome_133":        Chrome_133,
	"chrome_133_PSK":    Chrome_133_PSK,
	"safari_15_6_1":     Safari_15_6_1,
	"safari_16_0":       Safari_16_0,
	"safari_ipad_15_6":  Safari_Ipad_15_6,
	"safari_ios_15_5":   Safari_IOS_15_5,
	"safari_ios_15_6":   Safari_IOS_15_6,
	"safari_ios_16_0":   Safari_IOS_16_0,
	"safari_ios_17_0":   Safari_IOS_17_0,
	"safari_ios_18_0":   Safari_IOS_18_0,
	"safari_ios_18_5":   Safari_IOS_18_5,
	"firefox_102":       Firefox_102,
	"firefox_104":       Firefox_104,
	"firefox_105":       Firefox_105,
	"firefox_106":       Firefox_106,
	"firefox_108":       Firefox_108,
	"firefox_110":       Firefox_110,
	"firefox_117":       Firefox_117,
	"firefox_120":       Firefox_120,
	"firefox_123":       Firefox_123,
	"firefox_132":       Firefox_132,
	"firefox_133":       Firefox_133,
	"firefox_135":       Firefox_135,
	"opera_89":          Opera_89,
	"opera_90":          Opera_90,
	"opera_91":          Opera_91,
	// 移动端和自定义指纹
	"zalando_android_mobile": ZalandoAndroidMobile,
	"zalando_ios_mobile":     ZalandoIosMobile,
	"nike_ios_mobile":        NikeIosMobile,
	"nike_android_mobile":    NikeAndroidMobile,
	"mms_ios":                MMSIos,
	"mms_ios_2":              MMSIos2,
	"mms_ios_3":              MMSIos3,
	"mesh_ios":               MeshIos,
	"mesh_android":           MeshAndroid,
	"mesh_ios_2":             MeshIos2,
	"mesh_android_2":         MeshAndroid2,
	"confirmed_ios":          ConfirmedIos,
	"confirmed_android":      ConfirmedAndroid,
	"confirmed_android_2":    ConfirmedAndroid2,
	"okhttp4_android_7":      Okhttp4Android7,
	"okhttp4_android_8":      Okhttp4Android8,
	"okhttp4_android_9":      Okhttp4Android9,
	"okhttp4_android_10":     Okhttp4Android10,
	"okhttp4_android_11":     Okhttp4Android11,
	"okhttp4_android_12":     Okhttp4Android12,
	"okhttp4_android_13":     Okhttp4Android13,
	"cloudflare_custom":      CloudflareCustom,
	// Edge 系列
	"edge_99":  Edge_99,
	"edge_101": Edge_101,
	"edge_120": Edge_120,
	"edge_131": Edge_131,
	"edge_133": Edge_133,
}

// clientsMu 保护 MappedTLSClients 的并发访问
var clientsMu sync.RWMutex

// GetClientProfile 并发安全地获取客户端配置
func GetClientProfile(name string) (ClientProfile, bool) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()
	profile, ok := MappedTLSClients[name]
	return profile, ok
}

// GetAllProfiles 并发安全地获取所有配置名称列表
func GetAllProfiles() []string {
	clientsMu.RLock()
	defer clientsMu.RUnlock()
	names := make([]string, 0, len(MappedTLSClients))
	for name := range MappedTLSClients {
		names = append(names, name)
	}
	return names
}

// HasProfile 并发安全地检查配置是否存在
func HasProfile(name string) bool {
	clientsMu.RLock()
	defer clientsMu.RUnlock()
	_, ok := MappedTLSClients[name]
	return ok
}

type ClientProfile struct {
	clientHelloId     tls.ClientHelloID
	headerPriority    *http2.PriorityParam
	settings          map[http2.SettingID]uint32
	priorities        []http2.Priority
	pseudoHeaderOrder []string
	settingsOrder     []http2.SettingID
	connectionFlow    uint32

	// 元数据字段（用于 Client Hints 和 User-Agent 生成）
	BrowserType    string // "chrome", "firefox", "safari", "edge"
	BrowserVersion string // "120.0.6099.109"
	OS             string // "Windows NT 10.0; Win64; x64"
	OSVersion      string // "10.0.19045"
	OSArch         string // "x86", "arm"
	OSBitness      string // "64", "32"
	IsMobile       bool   // 是否为移动设备
	DeviceModel    string // 移动设备型号（如 "Pixel 7"）
}

// ClientProfileParams groups parameters for NewClientProfile.
type ClientProfileParams struct {
	ClientHelloID     tls.ClientHelloID
	Settings          map[http2.SettingID]uint32
	SettingsOrder     []http2.SettingID
	PseudoHeaderOrder []string
	ConnectionFlow    uint32
	Priorities        []http2.Priority
	HeaderPriority    *http2.PriorityParam
}

func NewClientProfile(p ClientProfileParams) ClientProfile {
	return ClientProfile{
		clientHelloId:     p.ClientHelloID,
		settings:          p.Settings,
		settingsOrder:     p.SettingsOrder,
		pseudoHeaderOrder: p.PseudoHeaderOrder,
		connectionFlow:    p.ConnectionFlow,
		priorities:        p.Priorities,
		headerPriority:    p.HeaderPriority,
		// 元数据字段保持默认零值
		BrowserType:    "",
		BrowserVersion: "",
		OS:             "",
		OSVersion:      "",
		OSArch:         "x86",
		OSBitness:      "64",
		IsMobile:       false,
		DeviceModel:    "",
	}
}

func (c ClientProfile) GetClientHelloSpec() (tls.ClientHelloSpec, error) {
	return c.clientHelloId.ToSpec()
}

func (c ClientProfile) GetClientHelloStr() string {
	return c.clientHelloId.Str()
}

func (c ClientProfile) GetSettings() map[http2.SettingID]uint32 {
	return c.settings
}

func (c ClientProfile) GetSettingsOrder() []http2.SettingID {
	return c.settingsOrder
}

func (c ClientProfile) GetConnectionFlow() uint32 {
	return c.connectionFlow
}

func (c ClientProfile) GetPseudoHeaderOrder() []string {
	return c.pseudoHeaderOrder
}

func (c ClientProfile) GetHeaderPriority() *http2.PriorityParam {
	return c.headerPriority
}

func (c ClientProfile) GetClientHelloId() tls.ClientHelloID {
	return c.clientHelloId
}

func (c ClientProfile) GetPriorities() []http2.Priority {
	return c.priorities
}

// GetClientHelloID 获取 ClientHelloID
func (c ClientProfile) GetClientHelloID() tls.ClientHelloID {
	return c.clientHelloId
}

// GetMetadata 获取元数据信息
func (c ClientProfile) GetMetadata() (browserType, browserVersion, os, osVersion string, isMobile bool) {
	return c.BrowserType, c.BrowserVersion, c.OS, c.OSVersion, c.IsMobile
}

// Clone 创建 ClientProfile 的深拷贝
func (c ClientProfile) Clone() ClientProfile {
	// 拷贝 settings
	newSettings := make(map[http2.SettingID]uint32, len(c.settings))
	for k, v := range c.settings {
		newSettings[k] = v
	}

	// 拷贝 priorities
	newPriorities := make([]http2.Priority, len(c.priorities))
	copy(newPriorities, c.priorities)

	// 拷贝 pseudoHeaderOrder
	newPseudoHeaderOrder := make([]string, len(c.pseudoHeaderOrder))
	copy(newPseudoHeaderOrder, c.pseudoHeaderOrder)

	// 拷贝 settingsOrder
	newSettingsOrder := make([]http2.SettingID, len(c.settingsOrder))
	copy(newSettingsOrder, c.settingsOrder)

	return ClientProfile{
		clientHelloId:     c.clientHelloId,
		headerPriority:    c.headerPriority,
		settings:          newSettings,
		priorities:        newPriorities,
		pseudoHeaderOrder: newPseudoHeaderOrder,
		settingsOrder:     newSettingsOrder,
		connectionFlow:    c.connectionFlow,
		BrowserType:       c.BrowserType,
		BrowserVersion:    c.BrowserVersion,
		OS:                c.OS,
		OSVersion:         c.OSVersion,
		OSArch:            c.OSArch,
		OSBitness:         c.OSBitness,
		IsMobile:          c.IsMobile,
		DeviceModel:       c.DeviceModel,
	}
}
