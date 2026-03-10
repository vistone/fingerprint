// Package profiles 提供浏览器指纹配置管理
package profiles

import (
	"strings"
	"sync"

	"github.com/vistone/fingerprint/modules/core"
)

// CreateTCPIP 根据操作系统创建 TCP/IP 指纹
func CreateTCPIP(osType core.OperatingSystem) *TCPIPFingerprint {
	base := &TCPIPFingerprint{
		IPVersion:        4,
		DF:               true,
		SYN:              true,
		ACK:              false,
		MSS:              1460,
		SAckPermitted:    true,
		Timestamps:       true,
		EndOfOptions:     true,
		OptionsSignature: "M,N,W,N,N,S,T,E",
	}

	osStr := string(osType)

	if strings.Contains(osStr, "Windows") {
		base.TTL = 128
		base.WindowSize = 64240
		base.WindowScale = 8
		base.NoOperation = 2
		base.JA4T = "t13d1715h2_8daaf6152771_9e7c7c2f41aa"
	} else if strings.Contains(osStr, "Macintosh") || strings.Contains(osStr, "Mac OS") {
		base.TTL = 64
		base.WindowSize = 65535
		base.WindowScale = 6
		base.NoOperation = 2
		base.JA4T = "t13d1814h2_8daaf6152771_b0b889a3c9b7"
	} else if strings.Contains(osStr, "Linux") || strings.Contains(osStr, "X11") {
		base.TTL = 64
		base.WindowSize = 64240
		base.WindowScale = 7
		base.NoOperation = 2
		base.JA4T = "t13d1714h2_8daaf6152771_02713a6ec338"
	} else if strings.Contains(osStr, "iPhone") || strings.Contains(osStr, "iPad") {
		// iOS 特征
		base.TTL = 64
		base.WindowSize = 65535
		base.WindowScale = 6
		base.NoOperation = 2
		base.JA4T = "t13d1814h2_8daaf6152771_b0b889a3c9b7"
	} else if strings.Contains(osStr, "Android") {
		// Android 特征
		base.TTL = 64
		base.WindowSize = 65535
		base.WindowScale = 6
		base.NoOperation = 2
		base.JA4T = "t13d1814h2_8daaf6152771_b0b889a3c9b7"
	} else {
		base.TTL = 128
		base.WindowSize = 64240
		base.WindowScale = 8
		base.NoOperation = 2
		base.JA4T = "t13d1715h2_8daaf6152771_9e7c7c2f41aa"
	}

	return base
}

// TCPIPFingerprint TCP/IP 层指纹配置
type TCPIPFingerprint struct {
	// IP 层配置
	IPVersion  uint8  // IP 版本 (4 或 6)
	TTL        uint8  // TTL 值
	DF         bool   // Don't Fragment 标志
	Flags      uint8  // IP 标志位
	TotalLen   uint16 // IP 包总长度
	FragOffset uint16 // 分片偏移

	// TCP 层配置
	WindowSize    uint16 // TCP 窗口大小
	MSS           uint16 // 最大段大小
	WindowScale   uint8  // 窗口缩放因子
	SAckPermitted bool   // SACK 许可
	Timestamps    bool   // 时间戳选项
	NoOperation   int    // NOP 数量
	EndOfOptions  bool   // 选项列表结束

	// TCP 标志
	SYN bool
	ACK bool

	// TCP 选项指纹字符串 (如 "M,W,S,T,N,N,E")
	OptionsSignature string

	// 综合指纹
	JA4T string // JA4T TCP 指纹哈希
}

// JSAntiDetection JavaScript 反检测对抗点配置
type JSAntiDetection struct {
	// WebGPU 对抗点
	WebGPU *WebGPUAntiDetect

	// MediaDevices 对抗点
	MediaDevices *MediaDevicesAntiDetect

	// Permissions API 对抗点
	Permissions *PermissionsAntiDetect

	// Automation 对抗点
	Automation *AutomationAntiDetect
}

// WebGPUAntiDetect WebGPU 对抗配置
type WebGPUAntiDetect struct {
	Available    bool              // 是否支持 WebGPU
	AdapterName  string            // GPU 适配器名称
	DeviceType   string            // 设备类型 (integrated, discrete, virtual)
	VendorID     string            // GPU 厂商 ID
	FeatureFlags []string          // 高级特性列表
	LimitValues  map[string]uint64 // 各项限制值
	BackendType  string            // 后端类型 (vulkan, metal, dx12, etc)
}

// MediaDevicesAntiDetect MediaDevices 对抗配置
type MediaDevicesAntiDetect struct {
	VideoInputs          []*MediaDeviceInfo     // 视频输入设备列表
	AudioInputs          []*MediaDeviceInfo     // 音频输入设备列表
	AudioOutputs         []*MediaDeviceInfo     // 音频输出设备列表
	getSources           []string               // navigator.mediaDevices.enumerateDevices 返回值
	UserMediaConstraints map[string]interface{} // getUserMedia 约束集
}

// MediaDeviceInfo 媒体设备信息
type MediaDeviceInfo struct {
	DeviceID  string // 设备唯一 ID
	GroupID   string // 设备组 ID (同物理设备)
	Kind      string // 设备类型 (videoinput, audioinput, audiooutput)
	Label     string // 设备标签/名称
	VendorID  string // 制造商 ID (USB 设备)
	ProductID string // 产品 ID (USB 设备)
}

// PermissionsAntiDetect Permissions API 对抗配置
type PermissionsAntiDetect struct {
	PermissionState  map[string]string // 权限名称 -> 状态 (granted, denied, prompt)
	RequestsAndroid  bool              // Android: 权限请求行为
	ShowNotification bool              // notifications 权限
	AccessCamera     bool              // camera 权限
	AccessMicrophone bool              // microphone 权限
	Geolocation      bool              // geolocation 权限
}

// AutomationAntiDetect Automation 对抗配置
type AutomationAntiDetect struct {
	WebDriver        bool // navigator.webdriver 标记 (false = 隐藏)
	Headless         bool // 是否隐藏 headless 特征
	ChromeDebugPort  bool // Chrome DevTools 端口隐藏
	Phantom          bool // phantomjs 特征隐藏
	Selenium         bool // selenium 驱动检测隐藏
	Puppeteer        bool // puppeteer 检测隐藏
	Playwright       bool // playwright 检测隐藏
	PluginsOverride  bool // plugins 数组欺骗
	LanguageOverride bool // language 属性欺骗
	ProductOverride  bool // product 属性欺骗 (Navigator.product)
	VendorOverride   bool // vendor 属性欺骗
	RuntimeOverride  bool // constructor name 欺骗
}

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
	HTTP2Settings     core.HTTP2Settings
	HTTP2Priorities   []core.HTTP2Priority
	PseudoHeaderOrder []string
	ConnectionFlow    uint32

	// HTTP/3 (QUIC) 配置
	HTTP3Settings *core.HTTP3Settings
	QUICVersions  []uint32

	// HTTP 头配置
	Headers *core.HTTPHeaders

	// TCP/IP 指纹配置 (新增)
	TCPIP *TCPIPFingerprint

	// JavaScript 反检测对抗点 (P3 高熵)
	JSAntiDetection *JSAntiDetection

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
	mu       sync.RWMutex
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
	r.mu.Lock()
	defer r.mu.Unlock()
	r.profiles[profile.ID] = profile
}

// Get 获取指纹
func (r *ProfileRegistry) Get(id string) (ClientProfile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.profiles[id]
	return p, ok
}

// GetByBrowser 按浏览器类型获取所有指纹
func (r *ProfileRegistry) GetByBrowser(browser core.BrowserType) []ClientProfile {
	r.mu.RLock()
	defer r.mu.RUnlock()
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
	r.mu.RLock()
	defer r.mu.RUnlock()
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
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]ClientProfile, 0, len(r.profiles))
	for _, p := range r.profiles {
		result = append(result, p)
	}
	return result
}

// Count 返回指纹数量
func (r *ProfileRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
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
