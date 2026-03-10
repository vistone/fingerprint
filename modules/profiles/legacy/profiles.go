// translated comment
// translated comment
// translated comment
//
//nolint:composites
package profiles

import (
	"sync"

	"github.com/bogdanfinn/fhttp/http2"
	tls "github.com/bogdanfinn/utls"
)

var DefaultClientProfile = Chrome_133

// translated comment
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
	// translated comment
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
	// translated comment
	"edge_99":  Edge_99,
	"edge_101": Edge_101,
	"edge_120": Edge_120,
	"edge_131": Edge_131,
	"edge_133": Edge_133,
}

// translated comment
var clientsMu sync.RWMutex

// translated comment
func GetClientProfile(name string) (ClientProfile, bool) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()
	profile, ok := MappedTLSClients[name]
	return profile, ok
}

// translated comment
func GetAllProfiles() []string {
	clientsMu.RLock()
	defer clientsMu.RUnlock()
	names := make([]string, 0, len(MappedTLSClients))
	for name := range MappedTLSClients {
		names = append(names, name)
	}
	return names
}

// translated comment
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

	// translated comment
	BrowserType    string // "chrome", "firefox", "safari", "edge"
	BrowserVersion string // "120.0.6099.109"
	OS             string // "Windows NT 10.0; Win64; x64"
	OSVersion      string // "10.0.19045"
	OSArch         string // "x86", "arm"
	OSBitness      string // "64", "32"
	IsMobile       bool   // translated comment
	DeviceModel    string // translated comment
}

func NewClientProfile(clientHelloId tls.ClientHelloID, settings map[http2.SettingID]uint32, settingsOrder []http2.SettingID, pseudoHeaderOrder []string, connectionFlow uint32, priorities []http2.Priority, headerPriority *http2.PriorityParam) ClientProfile {
	return ClientProfile{
		clientHelloId:     clientHelloId,
		settings:          settings,
		settingsOrder:     settingsOrder,
		pseudoHeaderOrder: pseudoHeaderOrder,
		connectionFlow:    connectionFlow,
		priorities:        priorities,
		headerPriority:    headerPriority,
		// translated comment
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

// translated comment
func (c ClientProfile) GetClientHelloID() tls.ClientHelloID {
	return c.clientHelloId
}

// translated comment
func (c ClientProfile) GetMetadata() (browserType, browserVersion, os, osVersion string, isMobile bool) {
	return c.BrowserType, c.BrowserVersion, c.OS, c.OSVersion, c.IsMobile
}

// translated comment
func (c ClientProfile) Clone() ClientProfile {
	// translated comment
	newSettings := make(map[http2.SettingID]uint32, len(c.settings))
	for k, v := range c.settings {
		newSettings[k] = v
	}

	// translated comment
	newPriorities := make([]http2.Priority, len(c.priorities))
	copy(newPriorities, c.priorities)

	// translated comment
	newPseudoHeaderOrder := make([]string, len(c.pseudoHeaderOrder))
	copy(newPseudoHeaderOrder, c.pseudoHeaderOrder)

	// translated comment
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
