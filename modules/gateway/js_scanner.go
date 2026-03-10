package gateway

import (
	"context"
	"fmt"
	htmlpkg "html"
	"regexp"
	"strings"
)

// translated comment
type JSDetectionResult struct {
	Detected             []DetectedAPI `json:"detected"`
	NotDetected          []string      `json:"notDetected"`
	TotalDetected        int           `json:"totalDetected"`
	TotalNotDetected     int           `json:"totalNotDetected"`
	ExecutionDetails     string        `json:"executionDetails"`
	DynamicFeaturesFound bool          `json:"dynamicFeaturesFound"`
}

// translated comment
type DetectedAPI struct {
	Name        string   `json:"name"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
	Count       int      `json:"count"`
	Samples     []string `json:"samples"`
}

// translated comment
type detectionPattern struct {
	Name        string
	Severity    string
	Description string
	Patterns    []*regexp.Regexp
}

var detectionPatterns = []detectionPattern{
	{
		Name:        "WebGPU Detection",
		Severity:    "high",
		Description: "检测 WebGPU 支持",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)navigator\.gpu\b`),
			regexp.MustCompile(`(?i)GPUAdapter|requestAdapter|createRenderPipeline`),
			regexp.MustCompile(`(?i)GPUDevice|adapter\.features`),
		},
	},
	{
		Name:        "Automation Detection",
		Severity:    "high",
		Description: "检测自动化工具特征",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)navigator\.webdriver`),
			regexp.MustCompile(`(?i)__nightmare__|__phantom__|__selenium__|_selenium`),
			regexp.MustCompile(`(?i)HeadlessChrome|\bheadless\b`),
			regexp.MustCompile(`(?i)cdc_[a-zA-Z0-9_]+`),
			regexp.MustCompile(`(?i)puppeteer|playwright|selenium`),
		},
	},
	{
		Name:        "MediaDevices Detection",
		Severity:    "medium",
		Description: "检测摄像头/麦克风设备",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)navigator\.mediaDevices`),
			regexp.MustCompile(`(?i)enumerateDevices\s*\(`),
			regexp.MustCompile(`(?i)getUserMedia\s*\(`),
			regexp.MustCompile(`(?i)getDisplayMedia\s*\(`),
		},
	},
	{
		Name:        "Permissions Detection",
		Severity:    "medium",
		Description: "检测权限状态",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)navigator\.permissions`),
			regexp.MustCompile(`(?i)\.query\s*\(\s*\{?\s*name`),
			regexp.MustCompile(`(?i)Permissions\.query`),
			regexp.MustCompile(`(?i)camera|microphone|geolocation|clipboard-read|notifications`),
		},
	},
	{
		Name:        "Canvas Fingerprinting",
		Severity:    "medium",
		Description: "可能的 Canvas 指纹识别",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)createElement\(['"]canvas['"]\)`),
			regexp.MustCompile(`(?i)getContext\s*\(\s*['"]2d['"]\s*\)`),
			regexp.MustCompile(`(?i)fillText|measureText|strokeText`),
			regexp.MustCompile(`(?i)toDataURL|toBlob|getImageData`),
		},
	},
	{
		Name:        "WebGL Fingerprinting",
		Severity:    "medium",
		Description: "可能的 WebGL 指纹识别",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)getContext\s*\(\s*['"]webgl2?['"]\s*\)`),
			regexp.MustCompile(`(?i)getParameter\s*\(`),
			regexp.MustCompile(`(?i)UNMASKED_VENDOR_WEBGL|UNMASKED_RENDERER_WEBGL`),
			regexp.MustCompile(`(?i)getSupportedExtensions|MAX_TEXTURE_SIZE`),
		},
	},
	{
		Name:        "Plugin Detection",
		Severity:    "low",
		Description: "检测浏览器插件",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)navigator\.plugins`),
			regexp.MustCompile(`(?i)navigator\.mimeTypes`),
			regexp.MustCompile(`(?i)ActiveXObject`),
		},
	},
	{
		Name:        "User Agent Detection",
		Severity:    "low",
		Description: "检测 User-Agent 特征",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)navigator\.userAgent\b`),
			regexp.MustCompile(`(?i)userAgentData|sec-ch-ua`),
			regexp.MustCompile(`(?i)indexOf\s*\(\s*['"]Headless['"]|includes\s*\(\s*['"]Headless['"]`),
		},
	},
	{
		Name:        "Font Detection",
		Severity:    "medium",
		Description: "检测字体枚举与字体指纹",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)document\.fonts|FontFace|check\s*\(\s*['"].*?['"]\s*\)`),
			regexp.MustCompile(`(?i)measureText\s*\(`),
			regexp.MustCompile(`(?i)getComputedStyle\s*\(`),
		},
	},
	{
		Name:        "Timezone Locale Detection",
		Severity:    "medium",
		Description: "检测时区/语言环境指纹",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)Intl\.DateTimeFormat\s*\(\s*\)\.resolvedOptions\s*\(\s*\)\.timeZone`),
			regexp.MustCompile(`(?i)getTimezoneOffset\s*\(`),
			regexp.MustCompile(`(?i)navigator\.language|navigator\.languages`),
		},
	},
	{
		Name:        "Hardware Detection",
		Severity:    "medium",
		Description: "检测硬件并发和设备内存指纹",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)navigator\.hardwareConcurrency`),
			regexp.MustCompile(`(?i)navigator\.deviceMemory`),
			regexp.MustCompile(`(?i)maxTouchPoints|platform`),
		},
	},
	{
		Name:        "Storage Quota Detection",
		Severity:    "low",
		Description: "检测存储配额指纹",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)navigator\.storage\.estimate\s*\(`),
			regexp.MustCompile(`(?i)localStorage|sessionStorage|indexedDB`),
		},
	},
	{
		Name:        "AntiBot Vendor Script",
		Severity:    "high",
		Description: "检测第三方反爬/指纹供应商脚本",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)fingerprintjs|openfpcdn|botd`),
			regexp.MustCompile(`(?i)perimeterx|datadome|arkoselabs|hcaptcha|recaptcha`),
			regexp.MustCompile(`(?i)cloudflare|turnstile|cf-chl|challenge-platform`),
			regexp.MustCompile(`(?i)humansecurity|kasada|shape-security|akamai`),
		},
	},
	{
		Name:        "Facebook Pixel Tracking",
		Severity:    "medium",
		Description: "检测Facebook Pixel追踪和事件采集",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)fbq\s*\(|_fbq\.push|fbevents`),
			regexp.MustCompile(`(?i)facebook\.com/tr\?|fbevents\.js`),
			regexp.MustCompile(`(?i)_fbp|_fbc|fb:pixel_id`),
		},
	},
	{
		Name:        "Google Analytics Tracking",
		Severity:    "medium",
		Description: "检测Google Analytics和广告追踪",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)gtag\s*\(|google-analytics|ga\s*\(`),
			regexp.MustCompile(`(?i)dataLayer\.push|gtm\.js`),
			regexp.MustCompile(`(?i)_gid|_ga|_gat|googletagmanager`),
		},
	},
	{
		Name:        "HubSpot Marketing Automation",
		Severity:    "medium",
		Description: "检测HubSpot营销自动化和行为追踪",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)_hsq\.push|hbspt\.forms|hs-script-loader`),
			regexp.MustCompile(`(?i)hubspotutk|hs-analytics\.js|hsadspixel`),
			regexp.MustCompile(`(?i)js\.hs-|hubspotapi|hubspot\.identify`),
		},
	},
	{
		Name:        "PostHog Analytics",
		Severity:    "medium",
		Description: "检测PostHog产品分析和会话回放",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)posthog\.capture|posthog\.identify|posthog\.setPersonProperties`),
			regexp.MustCompile(`(?i)posthog\(|PostHog\(|array\.js`),
			regexp.MustCompile(`(?i)ph-lib|posthog-js|posthog-tracking`),
		},
	},
	{
		Name:        "LinkedIn Insight Tag",
		Severity:    "medium",
		Description: "检测LinkedIn Insight Tag和行为追踪",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)_linkedin_data_partner_id|_linkedin_partner_id`),
			regexp.MustCompile(`(?i)linkedin\.com/li\.js|px\.ads\.linkedin`),
			regexp.MustCompile(`(?i)linkedInLeadGen|linkedinInsight`),
		},
	},
	{
		Name:        "Reddit Pixel Tracking",
		Severity:    "low",
		Description: "检测Reddit Pixel转化和再营销追踪",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)reddit\.com/pixels|rdt\.js`),
			regexp.MustCompile(`(?i)reddit_pixel_id|redditPixelId`),
			regexp.MustCompile(`(?i)redditsb|reddit-conversion`),
		},
	},
	{
		Name:        "Dreamdata Analytics",
		Severity:    "low",
		Description: "检测Dreamdata关键账户客户数据采集",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)dreamdata|drda\.io|dr_tracking`),
			regexp.MustCompile(`(?i)dreamdata\.com/.*?js|dreamdata-tracker`),
			regexp.MustCompile(`(?i)dream\.data|dreamdata\.push`),
		},
	},
}

// translated comment
func ScanJavaScriptWithV8(ctx context.Context, htmlContent string) (*JSDetectionResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	// translated comment
	normalizedCode := normalizeJavaScript(htmlContent)

	// translated comment
	result := &JSDetectionResult{
		Detected:         []DetectedAPI{},
		NotDetected:      []string{},
		ExecutionDetails: "使用高级静态代码分析\n",
	}

	// translated comment
	detectedSet := make(map[string]*DetectedAPI)

	// translated comment
	for _, pattern := range detectionPatterns {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		uniqueMatches := map[string]bool{}
		samples := []string{}

		for _, regexPattern := range pattern.Patterns {
			foundMatches := regexPattern.FindAllString(normalizedCode, -1)
			if len(foundMatches) > 0 {
				// translated comment
				for _, match := range foundMatches {
					uniqueMatches[match] = true
					if len(samples) < 3 {
						samples = append(samples, match)
					}
				}
			}
		}

		matches := len(uniqueMatches)

		if matches > 0 {
			// translated comment
			detected := DetectedAPI{
				Name:        pattern.Name,
				Severity:    pattern.Severity,
				Description: pattern.Description,
				Count:       matches,
				Samples:     samples,
			}
			detectedSet[pattern.Name] = &detected
			result.ExecutionDetails += fmt.Sprintf("✓ %s: 检测到 %d 个匹配\n", pattern.Name, matches)
		} else {
			// translated comment
			result.NotDetected = append(result.NotDetected, pattern.Name)
			result.ExecutionDetails += fmt.Sprintf("✗ %s: 未检测到\n", pattern.Name)
		}
	}

	// translated comment
	for _, pattern := range detectionPatterns {
		if detected, found := detectedSet[pattern.Name]; found {
			result.Detected = append(result.Detected, *detected)
		}
	}

	result.TotalDetected = len(result.Detected)
	result.TotalNotDetected = len(result.NotDetected)
	result.DynamicFeaturesFound = result.TotalDetected > 0

	return result, nil
}

// translated comment
func normalizeJavaScript(htmlContent string) string {
	// translated comment
	scripts := extractScriptTags(htmlContent)
	scriptSrcs := extractScriptSrcs(htmlContent)
	capturedScriptSrcs := extractCapturedScriptSrcs(htmlContent)
	inlineCode := strings.Join(scripts, "\n")

	// translated comment
	// translated comment
	re := regexp.MustCompile(`(?m)//.*$`)
	inlineCode = re.ReplaceAllString(inlineCode, "")

	// translated comment
	re = regexp.MustCompile(`(?s)/\*.*?\*/`)
	inlineCode = re.ReplaceAllString(inlineCode, "")

	combinedCode := inlineCode + "\n" + strings.Join(scriptSrcs, "\n") + "\n" + strings.Join(capturedScriptSrcs, "\n")

	// translated comment
	re = regexp.MustCompile(`\s+`)
	combinedCode = re.ReplaceAllString(combinedCode, " ")

	return combinedCode
}

// translated comment
func extractScriptTags(htmlContent string) []string {
	var scripts []string

	// translated comment
	re := regexp.MustCompile(`(?i)<script[^>]*>([\s\S]*?)</script>`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	for _, match := range matches {
		if len(match) > 1 {
			scriptContent := strings.TrimSpace(match[1])
			if scriptContent != "" {
				// translated comment
				if !strings.Contains(scriptContent, "src=") {
					scripts = append(scripts, scriptContent)
				}
			}
		}
	}

	return scripts
}

// translated comment
func extractScriptSrcs(htmlContent string) []string {
	var srcs []string

	// translated comment
	re := regexp.MustCompile(`(?is)<script[^>]*\bsrc\s*=\s*['"]([^'"]+)['"][^>]*>`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	for _, match := range matches {
		if len(match) > 1 {
			src := strings.TrimSpace(htmlpkg.UnescapeString(match[1]))
			src = strings.ReplaceAll(src, "&amp;", "&")
			if src != "" {
				srcs = append(srcs, src)
			}
		}
	}

	return srcs
}

// translated comment
func extractCapturedScriptSrcs(htmlContent string) []string {
	var srcs []string

	re := regexp.MustCompile(`(?is)<script[^>]*\bdata-captured-src\s*=\s*['"]([^'"]+)['"][^>]*>`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	for _, match := range matches {
		if len(match) > 1 {
			src := strings.TrimSpace(htmlpkg.UnescapeString(match[1]))
			src = strings.ReplaceAll(src, "&amp;", "&")
			if src != "" {
				srcs = append(srcs, src)
			}
		}
	}

	return srcs
}
