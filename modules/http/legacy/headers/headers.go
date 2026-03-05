package headers

// Phase 3: 本模块已完成基础迁移，待深度优化（详见 docs/5-process/modularization/PHASE_3_PLAN.md）
import (
	"fmt"

	"github.com/vistone/fingerprint/modules/kit"
	"github.com/vistone/fingerprint/modules/core/types"
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
func GenerateHeaders(browserType types.BrowserType, userAgent string, isMobile bool) *types.HTTPHeaders {
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	}
	headers := &types.HTTPHeaders{
		UserAgent: userAgent,
	}

	switch browserType {
	case types.BrowserChrome:
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

	case types.BrowserFirefox:
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

	case types.BrowserSafari:
		headers.Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
		headers.AcceptEncoding = "gzip, deflate, br"
		if !isMobile {
			headers.SecFetchSite = "none"
			headers.SecFetchMode = "navigate"
			headers.SecFetchUser = "?1"
			headers.SecFetchDest = "document"
		}

	case types.BrowserOpera:
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

	case types.BrowserEdge:
		// Edge 使用 Chromium 内核，headers 类似 Chrome
		headers.Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"
		headers.AcceptEncoding = "gzip, deflate, br, zstd"
		headers.SecFetchSite = "none"
		headers.SecFetchMode = "navigate"
		headers.SecFetchUser = "?1"
		headers.SecFetchDest = "document"
		headers.UpgradeInsecureRequests = "1"

		if isMobile {
			headers.SecCHUA = `"Not A(Brand";v="8", "Chromium";v="120", "Microsoft Edge";v="120"`
			headers.SecCHUAMobile = "?1"
			headers.SecCHUAPlatform = `"Android"`
		} else {
			edgeVersion := utils.ExtractChromeVersion(userAgent)
			headers.SecCHUA = fmt.Sprintf(`"Not A(Brand";v="8", "Chromium";v="%s", "Microsoft Edge";v="%s"`, edgeVersion, edgeVersion)
			headers.SecCHUAMobile = "?0"
			headers.SecCHUAPlatform = utils.ExtractPlatform(userAgent)
		}

	}

	// Accept-Language 使用随机语言
	headers.AcceptLanguage = RandomLanguage()

	return headers
}
