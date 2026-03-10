package headers

// Phase 3: This module has completed basic migration, awaiting deep optimization (see docs/5-process/modularization/PHASE_3_PLAN.md)
import (
	"fmt"

	"github.com/vistone/fingerprint/modules/kit"
	"github.com/vistone/fingerprint/modules/core/types"
)

// Global language list (sorted by usage frequency)
var Languages = []string{
	"en-US,en;q=0.9",          // English (United States)
	"zh-CN,zh;q=0.9,en;q=0.8", // Chinese (Simplified)
	"es-ES,es;q=0.9,en;q=0.8", // Spanish
	"fr-FR,fr;q=0.9,en;q=0.8", // French
	"de-DE,de;q=0.9,en;q=0.8", // German
	"ja-JP,ja;q=0.9,en;q=0.8", // Japanese
	"pt-BR,pt;q=0.9,en;q=0.8", // Portuguese (Brazil)
	"ru-RU,ru;q=0.9,en;q=0.8", // Russian
	"ar-SA,ar;q=0.9,en;q=0.8", // Arabic
	"ko-KR,ko;q=0.9,en;q=0.8", // Korean
	"it-IT,it;q=0.9,en;q=0.8", // Italian
	"tr-TR,tr;q=0.9,en;q=0.8", // Turkish
	"pl-PL,pl;q=0.9,en;q=0.8", // Polish
	"nl-NL,nl;q=0.9,en;q=0.8", // Dutch
	"sv-SE,sv;q=0.9,en;q=0.8", // Swedish
	"vi-VN,vi;q=0.9,en;q=0.8", // Vietnamese
	"th-TH,th;q=0.9,en;q=0.8", // Thai
	"id-ID,id;q=0.9,en;q=0.8", // Indonesian
	"hi-IN,hi;q=0.9,en;q=0.8", // Hindi
	"cs-CZ,cs;q=0.9,en;q=0.8", // Czech
	"ro-RO,ro;q=0.9,en;q=0.8", // Romanian
	"hu-HU,hu;q=0.9,en;q=0.8", // Hungarian
	"el-GR,el;q=0.9,en;q=0.8", // Greek
	"da-DK,da;q=0.9,en;q=0.8", // Danish
	"fi-FI,fi;q=0.9,en;q=0.8", // Finnish
	"no-NO,no;q=0.9,en;q=0.8", // Norwegian
	"he-IL,he;q=0.9,en;q=0.8", // Hebrew
	"uk-UA,uk;q=0.9,en;q=0.8", // Ukrainian
	"pt-PT,pt;q=0.9,en;q=0.8", // Portuguese (Portugal)
	"zh-TW,zh;q=0.9,en;q=0.8", // Chinese (Traditional)
}

// RandomLanguage randomly selects a language
func RandomLanguage() string {
	if len(Languages) == 0 {
		return "en-US,en;q=0.9" // Default to English
	}
	return utils.RandomChoiceString(Languages)
}

// GenerateHeaders generates standard HTTP headers based on browser type and User-Agent
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
			// Extract Chrome version from User-Agent
			chromeVersion := utils.ExtractChromeVersion(userAgent)
			headers.SecCHUA = fmt.Sprintf(`"Not A(Brand";v="8", "Chromium";v="%s", "Google Chrome";v="%s"`, chromeVersion, chromeVersion)
			headers.SecCHUAMobile = "?0"
			// Extract platform from User-Agent
			headers.SecCHUAPlatform = utils.ExtractPlatform(userAgent)
		}

	case types.BrowserFirefox:
		headers.Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"
		headers.AcceptEncoding = "gzip, deflate, br"
		// Firefox does not use Sec-Fetch-* headers (older versions)
		// Newer Firefox versions use them, but with different format
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
		// Opera uses Chromium engine, headers similar to Chrome
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
		// Edge uses Chromium engine, headers similar to Chrome
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

	// Accept-Language uses random language
	headers.AcceptLanguage = RandomLanguage()

	return headers
}
