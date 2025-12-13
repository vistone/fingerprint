package fingerprint

import (
	"fmt"

	"github.com/vistone/fingerprint/internal/utils"
)

// 全球语言列表（按使用频率排序）
var Languages = []string{
	"en-US,en;q=0.9",          // 英语（美国）
	"zh-CN,zh;q=0.9,en;q=0.8", // 中文（简体）
	"es-ES,es;q=0.9,en;q=0.8", // 西班牙语
	"fr-FR,fr;q=0.9,en;q=0.8", // 法语
	"de-DE,de;q=0.9,en;q=0.8", // 德语
	"ja-JP,ja;q=0.9,en;q=0.8", // 日语
	"pt-BR,pt;q=0.9,en;q=0.8", // 葡萄牙语（巴西）
	"ru-RU,ru;q=0.9,en;q=0.8", // 俄语
	"ar-SA,ar;q=0.9,en;q=0.8", // 阿拉伯语
	"ko-KR,ko;q=0.9,en;q=0.8", // 韩语
	"it-IT,it;q=0.9,en;q=0.8", // 意大利语
	"tr-TR,tr;q=0.9,en;q=0.8", // 土耳其语
	"pl-PL,pl;q=0.9,en;q=0.8", // 波兰语
	"nl-NL,nl;q=0.9,en;q=0.8", // 荷兰语
	"sv-SE,sv;q=0.9,en;q=0.8", // 瑞典语
	"vi-VN,vi;q=0.9,en;q=0.8", // 越南语
	"th-TH,th;q=0.9,en;q=0.8", // 泰语
	"id-ID,id;q=0.9,en;q=0.8", // 印尼语
	"hi-IN,hi;q=0.9,en;q=0.8", // 印地语
	"cs-CZ,cs;q=0.9,en;q=0.8", // 捷克语
	"ro-RO,ro;q=0.9,en;q=0.8", // 罗马尼亚语
	"hu-HU,hu;q=0.9,en;q=0.8", // 匈牙利语
	"el-GR,el;q=0.9,en;q=0.8", // 希腊语
	"da-DK,da;q=0.9,en;q=0.8", // 丹麦语
	"fi-FI,fi;q=0.9,en;q=0.8", // 芬兰语
	"no-NO,no;q=0.9,en;q=0.8", // 挪威语
	"he-IL,he;q=0.9,en;q=0.8", // 希伯来语
	"uk-UA,uk;q=0.9,en;q=0.8", // 乌克兰语
	"pt-PT,pt;q=0.9,en;q=0.8", // 葡萄牙语（葡萄牙）
	"zh-TW,zh;q=0.9,en;q=0.8", // 中文（繁体）
}

// RandomLanguage 随机选择一个语言
func RandomLanguage() string {
	if len(Languages) == 0 {
		return "en-US,en;q=0.9" // 默认返回英语
	}
	return utils.RandomChoiceString(Languages)
}

// GenerateHeaders 根据浏览器类型和 User-Agent 生成标准 HTTP headers
func GenerateHeaders(browserType BrowserType, userAgent string, isMobile bool) *HTTPHeaders {
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	}
	headers := &HTTPHeaders{
		UserAgent: userAgent,
	}

	switch browserType {
	case BrowserChrome:
		headers.Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"
		headers.AcceptEncoding = "gzip, deflate, br, zstd"
		headers.SecFetchSite = "none"
		headers.SecFetchMode = "navigate"
		headers.SecFetchUser = "?1"
		headers.SecFetchDest = "document"
		headers.UpgradeInsecureRequests = "1"

		if isMobile {
			headers.SecCHUA = `"Not A(Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`
			headers.SecCHUAMobile = "?1"
			headers.SecCHUAPlatform = `"Android"`
		} else {
			// 从 User-Agent 提取 Chrome 版本
			chromeVersion := utils.ExtractChromeVersion(userAgent)
			headers.SecCHUA = fmt.Sprintf(`"Not A(Brand";v="8", "Chromium";v="%s", "Google Chrome";v="%s"`, chromeVersion, chromeVersion)
			headers.SecCHUAMobile = "?0"
			// 从 User-Agent 提取平台
			headers.SecCHUAPlatform = utils.ExtractPlatform(userAgent)
		}

	case BrowserFirefox:
		headers.Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"
		headers.AcceptEncoding = "gzip, deflate, br"
		// Firefox 不使用 Sec-Fetch-* headers（旧版本）
		// 新版本 Firefox 使用，但格式不同
		if isMobile {
			headers.SecFetchSite = "none"
			headers.SecFetchMode = "navigate"
			headers.SecFetchUser = "?1"
			headers.SecFetchDest = "document"
		}

	case BrowserSafari:
		headers.Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
		headers.AcceptEncoding = "gzip, deflate, br"
		if !isMobile {
			headers.SecFetchSite = "none"
			headers.SecFetchMode = "navigate"
			headers.SecFetchUser = "?1"
			headers.SecFetchDest = "document"
		}

	case BrowserOpera:
		// Opera 使用 Chrome 内核，headers 类似 Chrome
		headers.Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"
		headers.AcceptEncoding = "gzip, deflate, br, zstd"
		headers.SecFetchSite = "none"
		headers.SecFetchMode = "navigate"
		headers.SecFetchUser = "?1"
		headers.SecFetchDest = "document"
		headers.UpgradeInsecureRequests = "1"

		if isMobile {
			headers.SecCHUA = `"Opera";v="91", "Chromium";v="105", "Not A(Brand";v="8"`
			headers.SecCHUAMobile = "?1"
			headers.SecCHUAPlatform = `"Android"`
		} else {
			headers.SecCHUA = `"Opera";v="91", "Chromium";v="105", "Not A(Brand";v="8"`
			headers.SecCHUAMobile = "?0"
			headers.SecCHUAPlatform = utils.ExtractPlatform(userAgent)
		}
	}

	// Accept-Language 使用随机语言
	headers.AcceptLanguage = RandomLanguage()

	return headers
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
