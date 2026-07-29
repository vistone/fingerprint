// Package profiles provides browser fingerprint profile management
package profiles

import (
	"strings"
	"sync"

	"github.com/vistone/fingerprint/modules/core"
)

// CreateTCPIP creates TCP/IP fingerprint based on operating system
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
		// iOS characteristics
		base.TTL = 64
		base.WindowSize = 65535
		base.WindowScale = 6
		base.NoOperation = 2
		base.JA4T = "t13d1814h2_8daaf6152771_b0b889a3c9b7"
	} else if strings.Contains(osStr, "Android") {
		// Android characteristics
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

// TCPIPFingerprint TCP/IP layer fingerprint profile
type TCPIPFingerprint struct {
	// IP layer profile
	IPVersion  uint8  // IP version (4 or 6)
	TTL        uint8  // TTL value
	DF         bool   // Don't Fragment flag
	Flags      uint8  // IP flags
	TotalLen   uint16 // IP total packet length
	FragOffset uint16 // fragment offset

	// TCP layer profile
	WindowSize    uint16 // TCP window size
	MSS           uint16 // maximum segment size
	WindowScale   uint8  // window scale factor
	SAckPermitted bool   // SACK permitted
	Timestamps    bool   // timestamp option
	NoOperation   int    // NOP count
	EndOfOptions  bool   // end of options list

	// TCP flag
	SYN bool
	ACK bool

	// TCP options fingerprint string (e.g. "M,W,S,T,N,N,E")
	OptionsSignature string

	// comprehensive fingerprint
	JA4T string // JA4T TCP fingerprint hash
}

// JSAntiDetection JavaScript anti-detection countermeasure profile
type JSAntiDetection struct {
	// WebGPU countermeasures
	WebGPU *WebGPUAntiDetect

	// MediaDevices countermeasures
	MediaDevices *MediaDevicesAntiDetect

	// Permissions API countermeasures
	Permissions *PermissionsAntiDetect

	// Automation countermeasures
	Automation *AutomationAntiDetect
}

// WebGPUAntiDetect WebGPU countermeasure profile
type WebGPUAntiDetect struct {
	Available    bool              // whether supported WebGPU
	AdapterName  string            // GPU adapter name
	DeviceType   string            // device type (integrated, discrete, virtual)
	VendorID     string            // GPU vendor ID
	FeatureFlags []string          // advanced features list
	LimitValues  map[string]uint64 // various limit values
	BackendType  string            // backend type (vulkan, metal, dx12, etc)
}

// MediaDevicesAntiDetect MediaDevices countermeasure profile
type MediaDevicesAntiDetect struct {
	VideoInputs          []*MediaDeviceInfo     // video input device list
	AudioInputs          []*MediaDeviceInfo     // audio input device list
	AudioOutputs         []*MediaDeviceInfo     // audio output device list
	getSources           []string               // navigator.mediaDevices.enumerateDevices return value
	UserMediaConstraints map[string]interface{} // getUserMedia constraint set
}

// MediaDeviceInfo media device information
type MediaDeviceInfo struct {
	DeviceID  string // device unique ID
	GroupID   string // device group ID (same physical device)
	Kind      string // device type (videoinput, audioinput, audiooutput)
	Label     string // device label/name
	VendorID  string // manufacturer ID (USB device)
	ProductID string // product ID (USB device)
}

// PermissionsAntiDetect Permissions API countermeasure profile
type PermissionsAntiDetect struct {
	PermissionState  map[string]string // permission name -> status (granted, denied, prompt)
	RequestsAndroid  bool              // Android: permission request behavior
	ShowNotification bool              // notifications permission
	AccessCamera     bool              // camera permission
	AccessMicrophone bool              // microphone permission
	Geolocation      bool              // geolocation permission
}

// AutomationAntiDetect Automation countermeasure profile
type AutomationAntiDetect struct {
	WebDriver        bool // navigator.webdriver marker (false = concealment)
	Headless         bool // whether hidden headless characteristics
	ChromeDebugPort  bool // Chrome DevTools port concealment
	Phantom          bool // phantomjs feature concealment
	Selenium         bool // selenium driver detection concealment
	Puppeteer        bool // puppeteer detection concealment
	Playwright       bool // playwright detection concealment
	PluginsOverride  bool // plugins array spoofing
	LanguageOverride bool // language property spoofing
	ProductOverride  bool // product property spoofing (Navigator.product)
	VendorOverride   bool // vendor property spoofing
	RuntimeOverride  bool // constructor name spoofing
}

// ClientProfile client fingerprint profile
type ClientProfile struct {
	// core identifier
	ID          string
	Name        string
	Description string

	// browser information
	BrowserType    core.BrowserType
	BrowserVersion string

	// operating system information
	OS        core.OperatingSystem
	OSVersion string
	OSArch    string
	OSBitness string

	// TLS profile
	TLSVersion      uint16
	CipherSuites    []uint16
	Extensions      []core.TLSExtension
	SupportedCurves []core.CurveID
	SupportedPoints []uint8

	// HTTP/2 profile
	HTTP2Settings     core.HTTP2Settings
	HTTP2Priorities   []core.HTTP2Priority
	PseudoHeaderOrder []string
	ConnectionFlow    uint32

	// HTTP/3 (QUIC) profile
	HTTP3Settings *core.HTTP3Settings
	QUICVersions  []uint32

	// HTTP header profile
	Headers *core.HTTPHeaders

	// TCP/IP fingerprint profile (new)
	TCPIP *TCPIPFingerprint

	// JavaScript anti-detection countermeasures (high entropy)
	JSAntiDetection *JSAntiDetection

	// metadata
	Metadata map[string]interface{}
}

// GetID returns the fingerprint unique identifier
func (p ClientProfile) GetID() string {
	return p.ID
}

// GetBrowserType returns the browser type
func (p ClientProfile) GetBrowserType() core.BrowserType {
	return p.BrowserType
}

// GetOS returns the operating system
func (p ClientProfile) GetOS() core.OperatingSystem {
	return p.OS
}

// GetUserAgent returns User-Agent
func (p ClientProfile) GetUserAgent() string {
	if p.Headers != nil {
		return p.Headers.UserAgent
	}
	return ""
}

// GetHeaders returns HTTP Headers
func (p ClientProfile) GetHeaders() *core.HTTPHeaders {
	return p.Headers
}

// GetSpec returns the fingerprint specification
func (p ClientProfile) GetSpec() core.FingerprintSpec {
	return p
}

// ProfileRegistry fingerprint registry
type ProfileRegistry struct {
	mu       sync.RWMutex
	profiles map[string]ClientProfile
}

// NewProfileRegistry creates a new fingerprint registry
func NewProfileRegistry() *ProfileRegistry {
	return &ProfileRegistry{
		profiles: make(map[string]ClientProfile),
	}
}

// Register registers a fingerprint
func (r *ProfileRegistry) Register(profile ClientProfile) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.profiles[profile.ID] = profile
}

// Get gets a fingerprint
func (r *ProfileRegistry) Get(id string) (ClientProfile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.profiles[id]
	if !ok {
		return ClientProfile{}, false
	}
	clone := p.Clone()
	if clone == nil {
		return ClientProfile{}, false
	}
	return *clone, true
}

// GetByBrowser gets all fingerprints by browser type
func (r *ProfileRegistry) GetByBrowser(browser core.BrowserType) []ClientProfile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []ClientProfile
	for _, p := range r.profiles {
		if p.BrowserType == browser {
			if clone := p.Clone(); clone != nil {
				result = append(result, *clone)
			}
		}
	}
	return result
}

// GetByOS gets all fingerprints by operating system
func (r *ProfileRegistry) GetByOS(os core.OperatingSystem) []ClientProfile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []ClientProfile
	for _, p := range r.profiles {
		if p.OS == os {
			if clone := p.Clone(); clone != nil {
				result = append(result, *clone)
			}
		}
	}
	return result
}

// GetAll gets all fingerprints
func (r *ProfileRegistry) GetAll() []ClientProfile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]ClientProfile, 0, len(r.profiles))
	for _, p := range r.profiles {
		if clone := p.Clone(); clone != nil {
			result = append(result, *clone)
		}
	}
	return result
}

// Count returns the number of fingerprints
func (r *ProfileRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.profiles)
}

// DefaultRegistry default global registry
var DefaultRegistry = NewProfileRegistry()

// Register registers a fingerprint in the default registry
func Register(profile ClientProfile) {
	DefaultRegistry.Register(profile)
}

// Get gets a fingerprint from the default registry
func Get(id string) (ClientProfile, bool) {
	return DefaultRegistry.Get(id)
}

// GetByBrowser gets fingerprints by browser type from the default registry
func GetByBrowser(browser core.BrowserType) []ClientProfile {
	return DefaultRegistry.GetByBrowser(browser)
}

// GetAll gets all fingerprints from the default registry
func GetAll() []ClientProfile {
	return DefaultRegistry.GetAll()
}

// GetRandom randomly gets a fingerprint
func GetRandom() ClientProfile {
	all := DefaultRegistry.GetAll()
	if len(all) == 0 {
		return ClientProfile{}
	}
	return core.RandomChoice(all)
}

// GetRandomByBrowser randomly gets a fingerprint by browser type
func GetRandomByBrowser(browser core.BrowserType) ClientProfile {
	profiles := DefaultRegistry.GetByBrowser(browser)
	if len(profiles) == 0 {
		return ClientProfile{}
	}
	return core.RandomChoice(profiles)
}
