package types

import "github.com/vistone/fingerprint/profiles"

// BrowserType 浏览器类型
type BrowserType string

const (
	BrowserChrome  BrowserType = "chrome"
	BrowserFirefox BrowserType = "firefox"
	BrowserSafari  BrowserType = "safari"
	BrowserOpera   BrowserType = "opera"
	BrowserEdge    BrowserType = "edge"
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
}

// FingerprintResult 指纹结果，包含指纹、User-Agent 和标准 HTTP Headers
type FingerprintResult struct {
	Profile       profiles.ClientProfile // 指纹配置
	UserAgent     string                 // 对应的 User-Agent
	HelloClientID string                 // Client Hello ID（与 tls-client 保持一致）
	Headers       *HTTPHeaders           // 标准 HTTP 请求头（包含全球语言支持）
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
	Custom                  map[string]string // 用户自定义的 headers（如 Cookie、Authorization、X-API-Key 等）
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
	cloned := &HTTPHeaders{
		Accept:                  h.Accept,
		AcceptLanguage:          h.AcceptLanguage,
		AcceptEncoding:          h.AcceptEncoding,
		UserAgent:               h.UserAgent,
		SecFetchSite:            h.SecFetchSite,
		SecFetchMode:            h.SecFetchMode,
		SecFetchUser:            h.SecFetchUser,
		SecFetchDest:            h.SecFetchDest,
		SecCHUA:                 h.SecCHUA,
		SecCHUAMobile:           h.SecCHUAMobile,
		SecCHUAPlatform:         h.SecCHUAPlatform,
		UpgradeInsecureRequests: h.UpgradeInsecureRequests,
	}

	// 克隆 Custom map
	if h.Custom != nil {
		cloned.Custom = make(map[string]string)
		for k, v := range h.Custom {
			cloned.Custom[k] = v
		}
	}

	return cloned
}

// Set 设置用户自定义的 header（系统会自动合并到 ToMap() 中）
// 这是推荐的方式，设置后调用 ToMap() 即可自动包含自定义 headers
// 示例：result.Headers.Set("Cookie", "session_id=abc123")
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
		// 如果值为空，删除该 header
		delete(h.Custom, key)
	}
}

// SetHeaders 批量设置用户自定义的 headers（系统会自动合并到 ToMap() 中）
// 示例：result.Headers.SetHeaders(map[string]string{"Cookie": "session_id=abc123", "X-API-Key": "key"})
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

// Merge 合并用户自定义的 headers，用户自定义的优先级更高
// customHeaders: 用户自定义的 headers（如 session、cookie、apikey 等）
// 返回合并后的新 HTTPHeaders 对象，不会修改原始对象
// 注意：推荐使用 Set 或 SetHeaders，然后直接调用 ToMap()
func (h *HTTPHeaders) Merge(customHeaders map[string]string) *HTTPHeaders {
	if h == nil {
		return nil
	}

	// 克隆当前 headers
	merged := h.Clone()

	if len(customHeaders) == 0 {
		return merged
	}

	// 初始化 Custom map（如果还没有）
	if merged.Custom == nil {
		merged.Custom = make(map[string]string)
	}

	// 将用户自定义的 headers 合并到标准 headers 中
	// 用户自定义的优先级更高，会覆盖系统生成的 headers
	for key, value := range customHeaders {
		if value == "" {
			continue // 跳过空值
		}

		// 根据 header 名称更新对应的字段
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
			// 其他自定义 headers（如 Cookie、Authorization、X-API-Key 等）存储在 Custom map 中
			merged.Custom[key] = value
		}
	}

	return merged
}

// ToMap 将 HTTPHeaders 转换为 map[string]string
// 系统会自动合并 Custom 中的用户自定义 headers（如 Cookie、Authorization、X-API-Key 等）
// 用户只需使用 Set 或 SetHeaders 设置自定义 headers，然后调用 ToMap() 即可
// 用户自定义的 headers 优先级更高，会覆盖系统生成的 headers
// 无需手动调用 Merge 或 ToMapWithCustom，系统会自动完成合并
func (h *HTTPHeaders) ToMap() map[string]string {
	return h.ToMapWithCustom(nil)
}

// ToMapWithCustom 将 HTTPHeaders 转换为 map[string]string，并合并用户自定义的 headers
// customHeaders: 用户自定义的 headers（如 session、cookie、apikey 等）
// 用户自定义的 headers 优先级更高，会覆盖系统生成的 headers
func (h *HTTPHeaders) ToMapWithCustom(customHeaders map[string]string) map[string]string {
	headers := make(map[string]string)

	// 先添加系统生成的标准 headers
	if h.Accept != "" {
		headers["Accept"] = h.Accept
	}
	if h.AcceptLanguage != "" {
		headers["Accept-Language"] = h.AcceptLanguage
	}
	if h.AcceptEncoding != "" {
		headers["Accept-Encoding"] = h.AcceptEncoding
	}
	if h.UserAgent != "" {
		headers["User-Agent"] = h.UserAgent
	}
	if h.SecFetchSite != "" {
		headers["Sec-Fetch-Site"] = h.SecFetchSite
	}
	if h.SecFetchMode != "" {
		headers["Sec-Fetch-Mode"] = h.SecFetchMode
	}
	if h.SecFetchUser != "" {
		headers["Sec-Fetch-User"] = h.SecFetchUser
	}
	if h.SecFetchDest != "" {
		headers["Sec-Fetch-Dest"] = h.SecFetchDest
	}
	if h.SecCHUA != "" {
		headers["Sec-CH-UA"] = h.SecCHUA
	}
	if h.SecCHUAMobile != "" {
		headers["Sec-CH-UA-Mobile"] = h.SecCHUAMobile
	}
	if h.SecCHUAPlatform != "" {
		headers["Sec-CH-UA-Platform"] = h.SecCHUAPlatform
	}
	if h.UpgradeInsecureRequests != "" {
		headers["Upgrade-Insecure-Requests"] = h.UpgradeInsecureRequests
	}

	// 合并 HTTPHeaders 中的 Custom headers
	if h.Custom != nil {
		for key, value := range h.Custom {
			if value != "" {
				headers[key] = value
			}
		}
	}

	// 合并传入的 customHeaders（优先级最高，会覆盖所有已有的 headers）
	if len(customHeaders) > 0 {
		for key, value := range customHeaders {
			if value != "" {
				headers[key] = value
			}
		}
	}

	return headers
}
