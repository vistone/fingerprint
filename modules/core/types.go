// Package core provides core types and fundamental definitions for the fingerprint library
// This is the base dependency for all modules, does not depend on any other internal packages
package core

// BrowserType browser type
type BrowserType string

const (
	BrowserChrome  BrowserType = "chrome"
	BrowserFirefox BrowserType = "firefox"
	BrowserSafari  BrowserType = "safari"
	BrowserOpera   BrowserType = "opera"
	BrowserEdge    BrowserType = "edge"
	BrowserBrave   BrowserType = "brave"
	BrowserSamsung BrowserType = "samsung"
)

// OperatingSystem operating system type
type OperatingSystem string

const (
	OSWindows10 OperatingSystem = "Windows NT 10.0; Win64; x64"
	// OSWindows11 UA is same as Win10 (actual browser behavior), distinguished by Sec-CH-UA-Platform-Version
	OSWindows11 OperatingSystem = "Windows NT 10.0; Win64; x64"
	OSMacOS13   OperatingSystem = "Macintosh; Intel Mac OS X 13_0_0"
	OSMacOS14   OperatingSystem = "Macintosh; Intel Mac OS X 14_0_0"
	OSMacOS15   OperatingSystem = "Macintosh; Intel Mac OS X 15_0_0"
	OSLinux     OperatingSystem = "X11; Linux x86_64"
	// Following Linux distributions have same UA (actual browser behavior), aliases kept for semantic distinction
	OSLinuxUbuntu OperatingSystem = "X11; Linux x86_64"
	OSLinuxDebian OperatingSystem = "X11; Linux x86_64"
	OSLinuxFedora OperatingSystem = "X11; Linux x86_64"
	OSiOS         OperatingSystem = "iPhone; CPU iPhone OS 17_0"
	OSiPadOS      OperatingSystem = "iPad; CPU OS 17_0"
	OSAndroid     OperatingSystem = "Linux; Android 14"
)

// OperatingSystems operating system list (for random selection, deduplicated to avoid probability bias)
var OperatingSystems = []OperatingSystem{
	OSWindows10,
	OSMacOS13,
	OSMacOS14,
	OSMacOS15,
	OSLinux,
	OSiOS,
	OSiPadOS,
	OSAndroid,
}

// HTTPHeaders standard HTTP request headers
type HTTPHeaders struct {
	Accept                  string            // Accept header
	AcceptLanguage          string            // Accept-Language header (supports global languages)
	AcceptEncoding          string            // Accept-Encoding header
	UserAgent               string            // User-Agent header
	SecFetchSite            string            // Sec-Fetch-Site header
	SecFetchMode            string            // Sec-Fetch-Mode header
	SecFetchUser            string            // Sec-Fetch-User header
	SecFetchDest            string            // Sec-Fetch-Dest header
	SecCHUA                 string            // Sec-CH-UA header
	SecCHUAMobile           string            // Sec-CH-UA-Mobile header
	SecCHUAPlatform         string            // Sec-CH-UA-Platform header
	UpgradeInsecureRequests string            // Upgrade-Insecure-Requests header
	Custom                  map[string]string // user-defined headers
}

// UserAgentTemplate User-Agent template
type UserAgentTemplate struct {
	Browser    BrowserType
	Version    string
	Template   string // template string, use %s placeholder for operating system
	Mobile     bool   // whether it is mobile
	OSRequired bool   // whether operating system info is required
}

// Clone clones HTTPHeaders object, returns a new copy
func (h *HTTPHeaders) Clone() *HTTPHeaders {
	if h == nil {
		return nil
	}
	cloned := *h

	if len(h.Custom) > 0 {
		cloned.Custom = make(map[string]string, len(h.Custom))
		for k, v := range h.Custom {
			cloned.Custom[k] = v
		}
	} else {
		cloned.Custom = nil
	}

	return &cloned
}

// Set sets user-defined header
func (h *HTTPHeaders) Set(key, value string) {
	if h == nil {
		return
	}
	if h.Custom == nil {
		h.Custom = make(map[string]string)
	}
	if value != "" {
		h.Custom[key] = value
	} else {
		delete(h.Custom, key)
	}
}

// SetHeaders batch sets user-defined headers
func (h *HTTPHeaders) SetHeaders(customHeaders map[string]string) {
	if h == nil {
		return
	}
	if h.Custom == nil {
		h.Custom = make(map[string]string)
	}
	for key, value := range customHeaders {
		if value != "" {
			h.Custom[key] = value
		} else {
			delete(h.Custom, key)
		}
	}
}

// Merge merges user-defined headers
func (h *HTTPHeaders) Merge(customHeaders map[string]string) *HTTPHeaders {
	if h == nil {
		return nil
	}

	merged := h.Clone()

	if len(customHeaders) == 0 {
		return merged
	}

	if merged.Custom == nil {
		merged.Custom = make(map[string]string)
	}

	for key, value := range customHeaders {
		if value == "" {
			continue
		}

		switch key {
		case "Accept":
			merged.Accept = value
		case "Accept-Language":
			merged.AcceptLanguage = value
		case "Accept-Encoding":
			merged.AcceptEncoding = value
		case "User-Agent":
			merged.UserAgent = value
		case "Sec-Fetch-Site":
			merged.SecFetchSite = value
		case "Sec-Fetch-Mode":
			merged.SecFetchMode = value
		case "Sec-Fetch-User":
			merged.SecFetchUser = value
		case "Sec-Fetch-Dest":
			merged.SecFetchDest = value
		case "Sec-CH-UA":
			merged.SecCHUA = value
		case "Sec-CH-UA-Mobile":
			merged.SecCHUAMobile = value
		case "Sec-CH-UA-Platform":
			merged.SecCHUAPlatform = value
		case "Upgrade-Insecure-Requests":
			merged.UpgradeInsecureRequests = value
		default:
			merged.Custom[key] = value
		}
	}

	return merged
}

// ToMap converts HTTPHeaders to map[string]string
func (h *HTTPHeaders) ToMap() map[string]string {
	return h.ToMapWithCustom(nil)
}

// ToMapWithCustom converts HTTPHeaders to map[string]string and merges user-defined headers
func (h *HTTPHeaders) ToMapWithCustom(customHeaders map[string]string) map[string]string {
	capacity := 12
	if h != nil && len(h.Custom) > 0 {
		capacity += len(h.Custom)
	}
	if len(customHeaders) > 0 {
		capacity += len(customHeaders)
	}
	headers := make(map[string]string, capacity)

	if h == nil {
		for key, value := range customHeaders {
			if value != "" {
				headers[key] = value
			}
		}
		return headers
	}

	fields := []struct {
		key   string
		value string
	}{
		{"Accept", h.Accept},
		{"Accept-Language", h.AcceptLanguage},
		{"Accept-Encoding", h.AcceptEncoding},
		{"User-Agent", h.UserAgent},
		{"Sec-Fetch-Site", h.SecFetchSite},
		{"Sec-Fetch-Mode", h.SecFetchMode},
		{"Sec-Fetch-User", h.SecFetchUser},
		{"Sec-Fetch-Dest", h.SecFetchDest},
		{"Sec-CH-UA", h.SecCHUA},
		{"Sec-CH-UA-Mobile", h.SecCHUAMobile},
		{"Sec-CH-UA-Platform", h.SecCHUAPlatform},
		{"Upgrade-Insecure-Requests", h.UpgradeInsecureRequests},
	}

	for _, f := range fields {
		if f.value != "" {
			headers[f.key] = f.value
		}
	}

	for key, value := range h.Custom {
		if value != "" {
			headers[key] = value
		}
	}

	for key, value := range customHeaders {
		if value != "" {
			headers[key] = value
		}
	}

	return headers
}
