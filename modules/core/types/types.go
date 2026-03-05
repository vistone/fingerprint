package types

import (
	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles/legacy"
)

// BrowserType browser type (type alias)
type BrowserType = core.BrowserType

const (
	BrowserChrome  = core.BrowserChrome
	BrowserFirefox = core.BrowserFirefox
	BrowserSafari  = core.BrowserSafari
	BrowserOpera   = core.BrowserOpera
	BrowserEdge    = core.BrowserEdge
)

// OperatingSystem operating system type (type alias)
type OperatingSystem = core.OperatingSystem

const (
	OSWindows10   = core.OSWindows10
	OSWindows11   = core.OSWindows11
	OSMacOS13     = core.OSMacOS13
	OSMacOS14     = core.OSMacOS14
	OSMacOS15     = core.OSMacOS15
	OSLinux       = core.OSLinux
	OSLinuxUbuntu = core.OSLinuxUbuntu
	OSLinuxDebian = core.OSLinuxDebian
	OSLinuxFedora = core.OSLinuxFedora
	OSiOS         = core.OSiOS
	OSiPadOS      = core.OSiPadOS
	OSAndroid     = core.OSAndroid
)

// OperatingSystems operating system list (for random selection)
var OperatingSystems = core.OperatingSystems

// FingerprintResult fingerprint result containing profile, User-Agent and standard HTTP Headers
type FingerprintResult struct {
	Profile       profiles.ClientProfile // Fingerprint profile
	UserAgent     string                 // Corresponding User-Agent
	HelloClientID string                 // Client Hello ID (consistent with tls-client)
	Headers       *HTTPHeaders           // Standard HTTP headers (with global language support)
}

// HTTPHeaders standard HTTP request headers (type alias)
type HTTPHeaders = core.HTTPHeaders

// UserAgentTemplate User-Agent template (type alias)
type UserAgentTemplate = core.UserAgentTemplate
