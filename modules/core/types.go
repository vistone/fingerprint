// Package core 提供指纹库的核心类型和基础定义
// 这是所有模块的基础依赖，不依赖任何其他内部包
package core

// BrowserType 浏览器类型
type BrowserType string

const (
	BrowserChrome   BrowserType = "chrome"
	BrowserFirefox  BrowserType = "firefox"
	BrowserSafari   BrowserType = "safari"
	BrowserOpera    BrowserType = "opera"
	BrowserEdge     BrowserType = "edge"
	BrowserBrave    BrowserType = "brave"
	BrowserSamsung  BrowserType = "samsung"
)

// OperatingSystem 操作系统类型
type OperatingSystem string

const (
	OSWindows10   OperatingSystem = "Windows NT 10.0; Win64; x64"
	OSWindows11   OperatingSystem = "Windows NT 10.0; Win64; x64"
	OSMacOS13     OperatingSystem = "Macintosh; Intel Mac OS X 13_0_0"
	OSMacOS14     OperatingSystem = "Macintosh; Intel Mac OS X 14_0_0"
	OSMacOS15     OperatingSystem = "Macintosh; Intel Mac OS X 15_0_0"
	OSLinux       OperatingSystem = "X11; Linux x86_64"
	OSLinuxUbuntu OperatingSystem = "X11; Linux x86_64"
	OSLinuxDebian OperatingSystem = "X11; Linux x86_64"
	OSLinuxFedora OperatingSystem = "X11; Linux x86_64"
	OSiOS         OperatingSystem = "iPhone; CPU iPhone OS 17_0"
	OSiPadOS      OperatingSystem = "iPad; CPU OS 17_0"
	OSAndroid     OperatingSystem = "Linux; Android 14"
)

// OperatingSystems 操作系统列表（用于随机选择）
var OperatingSystems = []OperatingSystem{
	OSWindows10,
	OSWindows11,
	OSMacOS13,
	OSMacOS14,
	OSMacOS15,
	OSLinux,
	OSLinuxUbuntu,
	OSLinuxDebian,
	OSLinuxFedora,
	OSiOS,
	OSiPadOS,
	OSAndroid,
}

// HTTPHeaders 标准的 HTTP 请求头
type HTTPHeaders struct {
	Accept                  string            // Accept 头
	AcceptLanguage          string            // Accept-Language 头（支持全球语言）
	AcceptEncoding          string            // Accept-Encoding 头
	UserAgent               string            // User-Agent 头
	SecFetchSite            string            // Sec-Fetch-Site 头
	SecFetchMode            string            // Sec-Fetch-Mode 头
	SecFetchUser            string            // Sec-Fetch-User 头
	SecFetchDest            string            // Sec-Fetch-Dest 头
	SecCHUA                 string            // Sec-CH-UA 头
	SecCHUAMobile           string            // Sec-CH-UA-Mobile 头
	SecCHUAPlatform         string            // Sec-CH-UA-Platform 头
	UpgradeInsecureRequests string            // Upgrade-Insecure-Requests 头
	Custom                  map[string]string // 用户自定义的 headers
}

// UserAgentTemplate User-Agent 模板
type UserAgentTemplate struct {
	Browser    BrowserType
	Version    string
	Template   string // 模板字符串，使用 %s 占位符表示操作系统
	Mobile     bool   // 是否为移动端
	OSRequired bool   // 是否需要操作系统信息
}

// Clone 克隆 HTTPHeaders 对象，返回一个新的副本
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

// Set 设置用户自定义的 header
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

// SetHeaders 批量设置用户自定义的 headers
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

// Merge 合并用户自定义的 headers
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

// ToMap 将 HTTPHeaders 转换为 map[string]string
func (h *HTTPHeaders) ToMap() map[string]string {
	return h.ToMapWithCustom(nil)
}

// ToMapWithCustom 将 HTTPHeaders 转换为 map[string]string，并合并用户自定义的 headers
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
