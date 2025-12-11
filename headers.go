package fingerprint

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// HTTPHeaders 标准的 HTTP 请求头
type HTTPHeaders struct {
	Accept                  string // Accept 头
	AcceptLanguage          string // Accept-Language 头（支持全球语言）
	AcceptEncoding          string // Accept-Encoding 头
	UserAgent               string // User-Agent 头
	SecFetchSite            string // Sec-Fetch-Site 头
	SecFetchMode            string // Sec-Fetch-Mode 头
	SecFetchUser            string // Sec-Fetch-User 头
	SecFetchDest            string // Sec-Fetch-Dest 头
	SecCHUA                 string // Sec-CH-UA 头
	SecCHUAMobile           string // Sec-CH-UA-Mobile 头
	SecCHUAPlatform         string // Sec-CH-UA-Platform 头
	UpgradeInsecureRequests string // Upgrade-Insecure-Requests 头
}

var (
	headerRNG  *rand.Rand
	headerRNGMu sync.Mutex
)

func init() {
	headerRNG = rand.New(rand.NewSource(time.Now().UnixNano()))
}

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
	headerRNGMu.Lock()
	defer headerRNGMu.Unlock()
	return Languages[headerRNG.Intn(len(Languages))]
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
			chromeVersion := extractChromeVersion(userAgent)
			headers.SecCHUA = fmt.Sprintf(`"Not A(Brand";v="8", "Chromium";v="%s", "Google Chrome";v="%s"`, chromeVersion, chromeVersion)
			headers.SecCHUAMobile = "?0"
			// 从 User-Agent 提取平台
			headers.SecCHUAPlatform = extractPlatform(userAgent)
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
			headers.SecCHUAPlatform = extractPlatform(userAgent)
		}
	}

	// Accept-Language 使用随机语言
	headers.AcceptLanguage = RandomLanguage()

	return headers
}

// ToMap 将 HTTPHeaders 转换为 map[string]string
func (h *HTTPHeaders) ToMap() map[string]string {
	headers := make(map[string]string)

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

	return headers
}

// 辅助函数：从 User-Agent 提取 Chrome 版本
func extractChromeVersion(ua string) string {
	// 简单提取，实际应该用正则
	// Chrome/120.0.0.0 -> 120
	start := indexOf(ua, "Chrome/")
	if start == -1 {
		return "120" // 默认版本
	}
	start += 7 // "Chrome/" 长度
	end := start
	for end < len(ua) && ua[end] != '.' && ua[end] != ' ' && ua[end] != ';' {
		end++
	}
	if end > start {
		return ua[start:end]
	}
	return "120"
}

// 辅助函数：从 User-Agent 提取平台
func extractPlatform(ua string) string {
	if contains(ua, "Windows") {
		return `"Windows"`
	} else if contains(ua, "Macintosh") {
		return `"macOS"`
	} else if contains(ua, "Linux") {
		return `"Linux"`
	}
	return `"Windows"` // 默认
}

// 辅助函数
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func contains(s, substr string) bool {
	return indexOf(s, substr) != -1
}
