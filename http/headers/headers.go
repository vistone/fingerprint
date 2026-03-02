package headers

import fp "github.com/vistone/fingerprint"

// HTTPHeaders HTTP 请求头结构。
type HTTPHeaders = fp.HTTPHeaders

// BrowserType 浏览器类型。
type BrowserType = fp.BrowserType

// RandomLanguage 随机选择一个语言。
func RandomLanguage() string {
	return fp.RandomLanguage()
}

// Generate 根据浏览器类型和 User-Agent 生成标准 HTTP 头。
func Generate(browserType BrowserType, userAgent string, isMobile bool) *HTTPHeaders {
	return fp.GenerateHeaders(browserType, userAgent, isMobile)
}
