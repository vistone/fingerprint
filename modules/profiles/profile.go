// translated comment
package profiles

import (
	"strings"
	"sync"

	"github.com/vistone/fingerprint/modules/core"
)

// translated comment
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
		// translated comment
		base.TTL = 64
		base.WindowSize = 65535
		base.WindowScale = 6
		base.NoOperation = 2
		base.JA4T = "t13d1814h2_8daaf6152771_b0b889a3c9b7"
	} else if strings.Contains(osStr, "Android") {
		// translated comment
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

// translated comment
type TCPIPFingerprint struct {
	// translated comment
	IPVersion  uint8  // translated comment
	TTL        uint8  // translated comment
	DF         bool   // translated comment
	Flags      uint8  // translated comment
	TotalLen   uint16 // translated comment
	FragOffset uint16 // translated comment

	// translated comment
	WindowSize    uint16 // translated comment
	MSS           uint16 // translated comment
	WindowScale   uint8  // translated comment
	SAckPermitted bool   // translated comment
	Timestamps    bool   // translated comment
	NoOperation   int    // translated comment
	EndOfOptions  bool   // translated comment

	// translated comment
	SYN bool
	ACK bool

	// translated comment
	OptionsSignature string

	// translated comment
	JA4T string // translated comment
}

// translated comment
type JSAntiDetection struct {
	// translated comment
	WebGPU *WebGPUAntiDetect

	// translated comment
	MediaDevices *MediaDevicesAntiDetect

	// translated comment
	Permissions *PermissionsAntiDetect

	// translated comment
	Automation *AutomationAntiDetect
}

// translated comment
type WebGPUAntiDetect struct {
	Available    bool              // translated comment
	AdapterName  string            // translated comment
	DeviceType   string            // translated comment
	VendorID     string            // translated comment
	FeatureFlags []string          // translated comment
	LimitValues  map[string]uint64 // translated comment
	BackendType  string            // translated comment
}

// translated comment
type MediaDevicesAntiDetect struct {
	VideoInputs          []*MediaDeviceInfo     // translated comment
	AudioInputs          []*MediaDeviceInfo     // translated comment
	AudioOutputs         []*MediaDeviceInfo     // translated comment
	getSources           []string               // translated comment
	UserMediaConstraints map[string]interface{} // translated comment
}

// translated comment
type MediaDeviceInfo struct {
	DeviceID  string // translated comment
	GroupID   string // translated comment
	Kind      string // translated comment
	Label     string // translated comment
	VendorID  string // translated comment
	ProductID string // translated comment
}

// translated comment
type PermissionsAntiDetect struct {
	PermissionState  map[string]string // translated comment
	RequestsAndroid  bool              // translated comment
	ShowNotification bool              // translated comment
	AccessCamera     bool              // translated comment
	AccessMicrophone bool              // translated comment
	Geolocation      bool              // translated comment
}

// translated comment
type AutomationAntiDetect struct {
	WebDriver        bool // translated comment
	Headless         bool // translated comment
	ChromeDebugPort  bool // translated comment
	Phantom          bool // translated comment
	Selenium         bool // translated comment
	Puppeteer        bool // translated comment
	Playwright       bool // translated comment
	PluginsOverride  bool // translated comment
	LanguageOverride bool // translated comment
	ProductOverride  bool // translated comment
	VendorOverride   bool // translated comment
	RuntimeOverride  bool // translated comment
}

// translated comment
type ClientProfile struct {
	// translated comment
	ID          string
	Name        string
	Description string

	// translated comment
	BrowserType    core.BrowserType
	BrowserVersion string

	// translated comment
	OS        core.OperatingSystem
	OSVersion string
	OSArch    string
	OSBitness string

	// translated comment
	TLSVersion      uint16
	CipherSuites    []uint16
	Extensions      []core.TLSExtension
	SupportedCurves []core.CurveID
	SupportedPoints []uint8

	// translated comment
	HTTP2Settings     core.HTTP2Settings
	HTTP2Priorities   []core.HTTP2Priority
	PseudoHeaderOrder []string
	ConnectionFlow    uint32

	// translated comment
	HTTP3Settings *core.HTTP3Settings
	QUICVersions  []uint32

	// translated comment
	Headers *core.HTTPHeaders

	// translated comment
	TCPIP *TCPIPFingerprint

	// translated comment
	JSAntiDetection *JSAntiDetection

	// translated comment
	Metadata map[string]interface{}
}

// translated comment
func (p ClientProfile) GetID() string {
	return p.ID
}

// translated comment
func (p ClientProfile) GetBrowserType() core.BrowserType {
	return p.BrowserType
}

// translated comment
func (p ClientProfile) GetOS() core.OperatingSystem {
	return p.OS
}

// translated comment
func (p ClientProfile) GetUserAgent() string {
	if p.Headers != nil {
		return p.Headers.UserAgent
	}
	return ""
}

// translated comment
func (p ClientProfile) GetHeaders() *core.HTTPHeaders {
	return p.Headers
}

// translated comment
func (p ClientProfile) GetSpec() core.FingerprintSpec {
	return p
}

// translated comment
type ProfileRegistry struct {
	mu       sync.RWMutex
	profiles map[string]ClientProfile
}

// translated comment
func NewProfileRegistry() *ProfileRegistry {
	return &ProfileRegistry{
		profiles: make(map[string]ClientProfile),
	}
}

// translated comment
func (r *ProfileRegistry) Register(profile ClientProfile) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.profiles[profile.ID] = profile
}

// translated comment
func (r *ProfileRegistry) Get(id string) (ClientProfile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.profiles[id]
	return p, ok
}

// translated comment
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

// translated comment
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

// translated comment
func (r *ProfileRegistry) GetAll() []ClientProfile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]ClientProfile, 0, len(r.profiles))
	for _, p := range r.profiles {
		result = append(result, p)
	}
	return result
}

// translated comment
func (r *ProfileRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.profiles)
}

// translated comment
var DefaultRegistry = NewProfileRegistry()

// translated comment
func Register(profile ClientProfile) {
	DefaultRegistry.Register(profile)
}

// translated comment
func Get(id string) (ClientProfile, bool) {
	return DefaultRegistry.Get(id)
}

// translated comment
func GetByBrowser(browser core.BrowserType) []ClientProfile {
	return DefaultRegistry.GetByBrowser(browser)
}

// translated comment
func GetAll() []ClientProfile {
	return DefaultRegistry.GetAll()
}

// translated comment
func GetRandom() ClientProfile {
	all := DefaultRegistry.GetAll()
	if len(all) == 0 {
		return ClientProfile{}
	}
	return core.RandomChoice(all)
}

// translated comment
func GetRandomByBrowser(browser core.BrowserType) ClientProfile {
	profiles := DefaultRegistry.GetByBrowser(browser)
	if len(profiles) == 0 {
		return ClientProfile{}
	}
	return core.RandomChoice(profiles)
}
