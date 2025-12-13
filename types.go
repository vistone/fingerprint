package fingerprint

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
	Profile       ClientProfile // 指纹配置
	UserAgent     string        // 对应的 User-Agent
	HelloClientID string        // Client Hello ID（与 tls-client 保持一致）
	Headers       *HTTPHeaders  // 标准 HTTP 请求头（包含全球语言支持）
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
