package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// JA4HResult JA4H 指纹结果（HTTP 请求头指纹）
type JA4HResult struct {
	// 完整 JA4H SHA256 哈希
	Hash string

	// JA4H_a: 基本字符串（请求方法+路径特征+协议+头计数等）
	JA4Ha string

	// JA4H_r: 原始版本（请求头顺序）
	JA4Hr string

	// 分解的签名部分
	HeaderOrderSignature    string // 请求头顺序哈希
	HeaderValueSignature    string // 特定头值哈希
	QueryParameterSignature string // 查询参数顺序哈希

	// 完整原始签名字符串
	RawSignature string

	// 异常判定分数
	RiskScore float64

	// 异常标记列表
	AnomalyFlags []string

	// 匹配的已知客户端/浏览器
	MatchedBrowsers []string
}

// HTTP2RequestData HTTP 请求信息
type HTTP2RequestData struct {
	// 请求行
	Method   string // GET, POST, etc.
	Path     string // /path/to/resource
	Protocol string // HTTP/1.1, HTTP/2, HTTP/3

	// 请求头（序列化的，保持原始顺序）
	Headers []struct {
		Name  string
		Value string
	}

	// 查询参数
	QueryParams map[string]string

	// 元数据
	Metadata map[string]string
}

// JA4HAnalyzer JA4H 分析器
type JA4HAnalyzer struct {
	knownBrowserProfiles map[string]*HTTP2BrowserProfile
}

// HTTP2BrowserProfile 已知的浏览器配置
type HTTP2BrowserProfile struct {
	Name           string
	BrowserName    string
	BrowserVersion string
	HeaderOrder    []string // 标准请求头顺序
	TypicalHeaders map[string]string
	HeaderStrategy string // "standard", "randomized", "optimized"
	RiskScore      float64
}

// NewJA4HAnalyzer 创建分析器
func NewJA4HAnalyzer() *JA4HAnalyzer {
	return &JA4HAnalyzer{
		knownBrowserProfiles: initKnownBrowserProfiles(),
	}
}

// AnalyzeHTTPRequest 分析 HTTP 请求头
func (a *JA4HAnalyzer) AnalyzeHTTPRequest(req HTTP2RequestData) (*JA4HResult, error) {
	if req.Method == "" {
		return nil, fmt.Errorf("method required")
	}

	result := &JA4HResult{
		AnomalyFlags: make([]string, 0, 8), // 预分配容量
	}

	// 1. 构建基本签名字符串：方法,协议,路径特征,头计数
	pathSignature := generatePathSignature(req.Path)
	headerCount := len(req.Headers)
	acceptedEncodingPresent := hasHeader(req.Headers, "accept-encoding")

	result.JA4Ha = fmt.Sprintf("%s,%s,%s,%d,%v",
		req.Method,
		req.Protocol,
		pathSignature,
		headerCount,
		acceptedEncodingPresent,
	)

	// 2. 提取请求头顺序
	var headerNames []string
	for _, h := range req.Headers {
		headerNames = append(headerNames, strings.ToLower(h.Name))
	}
	headerOrderStr := strings.Join(headerNames, ",")
	result.HeaderOrderSignature = generateHeaderOrderSignature(headerNames)

	// 3. 生成特定请求头值的签名（用于指纹识别）
	result.HeaderValueSignature = generateHeaderValueSignature(req.Headers)

	// 4. 查询参数签名
	if len(req.QueryParams) > 0 {
		result.QueryParameterSignature = generateQueryParamSignature(req.QueryParams)
	}

	// 5. 完整签名字符串
	result.RawSignature = fmt.Sprintf(
		"ja4h|%s|%s|%s|%s|%s",
		result.JA4Ha,
		headerOrderStr,
		result.HeaderOrderSignature,
		result.HeaderValueSignature,
		result.QueryParameterSignature,
	)

	// 计算 SHA256 哈希
	hash := sha256.Sum256([]byte(result.RawSignature))
	result.Hash = hex.EncodeToString(hash[:])

	// 异常检测
	a.detectJA4HAnomalies(result, req)

	return result, nil
}

// detectJA4HAnomalies 检测 JA4H 异常
func (a *JA4HAnalyzer) detectJA4HAnomalies(result *JA4HResult, req HTTP2RequestData) {
	baseScore := 0.0

	// 异常 1: 请求头顺序异常
	headerNames := make([]string, len(req.Headers))
	for i, h := range req.Headers {
		headerNames[i] = strings.ToLower(h.Name)
	}

	if !isStandardHeaderOrderJA4H(headerNames) {
		result.AnomalyFlags = append(result.AnomalyFlags, "UNUSUAL_HEADER_ORDER")
		baseScore += 0.2
	}

	// 异常 2: 缺少常见请求头
	missingHeaders := checkMissingCommonHeaders(headerNames)
	if len(missingHeaders) > 2 {
		result.AnomalyFlags = append(result.AnomalyFlags, fmt.Sprintf("MISSING_HEADERS_%d", len(missingHeaders)))
		baseScore += 0.15
	}

	// 异常 3: 异常的请求头值
	for _, h := range req.Headers {
		if isSuspiciousHeaderValue(h.Name, h.Value) {
			result.AnomalyFlags = append(result.AnomalyFlags, fmt.Sprintf("SUSPICIOUS_%s", h.Name))
			baseScore += 0.1
		}
	}

	// 异常 4: User-Agent 不匹配
	if ua := getHeaderValue(req.Headers, "user-agent"); ua != "" {
		if !isValidUserAgent(ua, req.Method, req.Path) {
			result.AnomalyFlags = append(result.AnomalyFlags, "INVALID_USER_AGENT")
			baseScore += 0.2
		}
	}

	// 异常 5: 请求与路径不匹配
	if !isMethodPathConsistent(req.Method, req.Path) {
		result.AnomalyFlags = append(result.AnomalyFlags, "METHOD_PATH_MISMATCH")
		baseScore += 0.15
	}

	// 异常 6: 查询参数异常（SQL 注入等）
	if len(req.QueryParams) > 0 && isSuspiciousQueryParams(req.QueryParams) {
		result.AnomalyFlags = append(result.AnomalyFlags, "SUSPICIOUS_QUERY_PARAMS")
		baseScore += 0.25
	}

	// 异常 7: Client Hints 不一致
	chInconsistencies := detectClientHintsInconsistencies(req.Headers)
	for _, inconsistency := range chInconsistencies {
		result.AnomalyFlags = append(result.AnomalyFlags, fmt.Sprintf("CH_%s", inconsistency))
		baseScore += 0.1
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}
	result.RiskScore = baseScore
}

// FindMatchingBrowsers 查找匹配的已知浏览器
func (a *JA4HAnalyzer) FindMatchingBrowsers(
	result *JA4HResult,
	maxResults int,
) []string {
	var matches []string

	for name, profile := range a.knownBrowserProfiles {
		// 基于风险分数和特征相似度
		if profile.RiskScore < result.RiskScore+0.15 {
			matches = append(matches, name)
		}

		if len(matches) >= maxResults {
			break
		}
	}

	result.MatchedBrowsers = matches
	return matches
}

// ============ 辅助函数 ============

func generatePathSignature(path string) string {
	// 基于路径深度和特征
	if path == "" || path == "/" {
		return "root"
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	depth := len(parts)

	// 检测常见路径模式
	if contains(parts, "api") {
		return fmt.Sprintf("api_d%d", depth)
	}
	if contains(parts, "static") {
		return fmt.Sprintf("static_d%d", depth)
	}
	if contains(parts, "v1") || contains(parts, "v2") {
		return fmt.Sprintf("versioned_d%d", depth)
	}

	return fmt.Sprintf("custom_d%d", depth)
}

func generateHeaderOrderSignature(headers []string) string {
	// 规范化头顺序并计算哈希
	orderStr := strings.Join(headers, ",")
	hash := sha256.Sum256([]byte(orderStr))
	return hex.EncodeToString(hash[:8])
}

func generateHeaderValueSignature(headers []struct {
	Name  string
	Value string
}) string {
	// 基于关键头的值生成签名
	var keyValues []string

	keyHeaders := []string{"user-agent", "accept", "accept-language", "accept-encoding"}

	for _, h := range headers {
		if contains(keyHeaders, strings.ToLower(h.Name)) {
			// 计算值的简单哈希
			hash := sha256.Sum256([]byte(h.Value))
			keyValues = append(keyValues, fmt.Sprintf("%s:%s", h.Name, hex.EncodeToString(hash[:4])))
		}
	}

	if len(keyValues) == 0 {
		return "no_key_headers"
	}

	combinedSig := strings.Join(keyValues, "|")
	hash := sha256.Sum256([]byte(combinedSig))
	return hex.EncodeToString(hash[:8])
}

func generateQueryParamSignature(params map[string]string) string {
	// 参数顺序和名称的哈希
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	paramStr := strings.Join(keys, ",")
	hash := sha256.Sum256([]byte(paramStr))
	return hex.EncodeToString(hash[:8])
}

// ============ 验证函数 ============

func isStandardHeaderOrderJA4H(headers []string) bool {
	// 标准浏览器的请求头顺序通常如下：
	// Host, User-Agent, Accept, Accept-Language, Accept-Encoding, ...
	if len(headers) == 0 {
		return false
	}

	// 检查是否以 Host 开头（非常常见）
	if headers[0] != "host" && headers[0] != "connection" {
		// 不一定异常，某些客户端可能不同
		return true
	}

	// 检查 User-Agent 相对位置（通常在前 3 个）
	uaIdx := -1
	for i, h := range headers {
		if h == "user-agent" {
			uaIdx = i
			break
		}
	}

	if uaIdx > 5 {
		return false // 异常：UA 在很后面
	}

	return true
}

func checkMissingCommonHeaders(headers []string) []string {
	commonHeaders := []string{"host", "user-agent", "accept", "accept-language"}
	var missing []string

	for _, common := range commonHeaders {
		if !contains(headers, common) {
			missing = append(missing, common)
		}
	}

	return missing
}

func isSuspiciousHeaderValue(name, value string) bool {
	name = strings.ToLower(name)

	// 检查已知的异常值
	switch name {
	case "user-agent":
		// 检查已知的机器人关键词
		botKeywords := []string{"bot", "crawler", "spider", "scraper", "headless"}
		for _, kw := range botKeywords {
			if strings.Contains(strings.ToLower(value), kw) {
				return false // 这可能不一定是异常，取决于上下文
			}
		}

	case "accept":
		// 检查是否缺少合理的 MIME 类型
		if !strings.Contains(value, "text/html") && !strings.Contains(value, "*/*") {
			return true
		}

	case "accept-encoding":
		// 检查是否为空或包含异常编码
		if value == "" || strings.Contains(value, "identity") {
			return true
		}
	}

	return false
}

func isValidUserAgent(ua string, method string, path string) bool {
	// 简单的检查：User-Agent 不应该为空或太短
	if ua == "" || len(ua) < 10 {
		return false
	}

	// 如果是 GET 请求到某些路径，User-Agent 应该像浏览器
	if method == "GET" && !strings.HasPrefix(path, "/api") {
		if !strings.Contains(strings.ToLower(ua), "mozilla") &&
			!strings.Contains(strings.ToLower(ua), "opera") &&
			!strings.Contains(strings.ToLower(ua), "curl") {
			return false
		}
	}

	return true
}

func isMethodPathConsistent(method string, path string) bool {
	// GET 不应该用于 /api/post 或 /api/delete 等
	if method == "GET" && strings.Contains(strings.ToLower(path), "/delete") {
		return false
	}

	// POST/PUT 通常用于提交数据
	if (method == "POST" || method == "PUT") && strings.HasSuffix(path, ".ico") {
		return false
	}

	return true
}

func isSuspiciousQueryParams(params map[string]string) bool {
	// 检查 SQL 注入迹象
	sqlKeywords := []string{"union", "select", "insert", "delete", "drop", "--", "/*", "*/"}

	for _, v := range params {
		for _, kw := range sqlKeywords {
			if strings.Contains(strings.ToLower(v), kw) {
				return true
			}
		}
	}

	return false
}

// detectClientHintsInconsistencies 检测 Client Hints 的不一致性
func detectClientHintsInconsistencies(headers []struct {
	Name  string
	Value string
}) []string {
	var inconsistencies []string

	// 提取 Client Hints
	secCHUA := getHeaderValue(headers, "sec-ch-ua")
	secCHUAMobile := getHeaderValue(headers, "sec-ch-ua-mobile")
	secCHUAPlatform := getHeaderValue(headers, "sec-ch-ua-platform")
	secCHUAArch := getHeaderValue(headers, "sec-ch-ua-arch")
	secCHUABitness := getHeaderValue(headers, "sec-ch-ua-bitness")
	secCHUAPlatformVersion := getHeaderValue(headers, "sec-ch-ua-platform-version")
	userAgent := getHeaderValue(headers, "user-agent")

	// 1. Sec-CH-UA 存在但 User-Agent 不像 Chromium
	if secCHUA != "" && userAgent != "" {
		if !strings.Contains(userAgent, "Chrome") && !strings.Contains(userAgent, "Chromium") && !strings.Contains(userAgent, "Edge") {
			inconsistencies = append(inconsistencies, "UA_CH_MISMATCH")
		}
	}

	// 2. 声称移动设备但平台不是移动平台
	if secCHUAMobile == "?1" {
		platform := strings.ToLower(secCHUAPlatform)
		if platform != "" && platform != `"android"` && platform != `"ios"` {
			inconsistencies = append(inconsistencies, "MOBILE_PLATFORM_MISMATCH")
		}
	}

	// 3. 高熵提示存在但低熵提示缺失（异常）
	hasHighEntropyHints := secCHUAArch != "" || secCHUABitness != "" || secCHUAPlatformVersion != ""
	hasLowEntropyHints := secCHUA != "" || secCHUAMobile != "" || secCHUAPlatform != ""
	if hasHighEntropyHints && !hasLowEntropyHints {
		inconsistencies = append(inconsistencies, "MISSING_LOW_ENTROPY_HINTS")
	}

	// 4. 平台架构与 User-Agent 不一致
	if secCHUAPlatform != "" && userAgent != "" {
		platform := strings.ToLower(secCHUAPlatform)
		ua := strings.ToLower(userAgent)

		if strings.Contains(platform, "windows") && !strings.Contains(ua, "windows") {
			inconsistencies = append(inconsistencies, "PLATFORM_UA_MISMATCH")
		}
		if strings.Contains(platform, "macos") && !strings.Contains(ua, "macintosh") {
			inconsistencies = append(inconsistencies, "PLATFORM_UA_MISMATCH")
		}
		if strings.Contains(platform, "linux") && !strings.Contains(ua, "linux") {
			inconsistencies = append(inconsistencies, "PLATFORM_UA_MISMATCH")
		}
	}

	// 5. 架构位宽不一致
	if secCHUABitness != "" && userAgent != "" {
		bitness := strings.Trim(secCHUABitness, `"`)
		ua := strings.ToLower(userAgent)

		if bitness == "64" && !strings.Contains(ua, "64") && !strings.Contains(ua, "x86_64") && !strings.Contains(ua, "win64") {
			inconsistencies = append(inconsistencies, "BITNESS_MISMATCH")
		}
	}

	return inconsistencies
}

// AnalyzeClientHints 分析请求中的 Client Hints
func (a *JA4HAnalyzer) AnalyzeClientHints(req HTTP2RequestData) *ClientHintsAnalysis {
	analysis := &ClientHintsAnalysis{
		LowEntropyHints:  make(map[string]string),
		HighEntropyHints: make(map[string]string),
		Inconsistencies:  []string{},
	}

	// 提取低熵提示
	for _, h := range req.Headers {
		name := strings.ToLower(h.Name)
		switch name {
		case "sec-ch-ua":
			analysis.LowEntropyHints["Sec-CH-UA"] = h.Value
			analysis.HasLowEntropyHints = true
		case "sec-ch-ua-mobile":
			analysis.LowEntropyHints["Sec-CH-UA-Mobile"] = h.Value
			analysis.HasLowEntropyHints = true
		case "sec-ch-ua-platform":
			analysis.LowEntropyHints["Sec-CH-UA-Platform"] = h.Value
			analysis.HasLowEntropyHints = true
		}
	}

	// 提取高熵提示
	for _, h := range req.Headers {
		name := strings.ToLower(h.Name)
		switch name {
		case "sec-ch-ua-arch":
			analysis.HighEntropyHints["Sec-CH-UA-Arch"] = h.Value
			analysis.HasHighEntropyHints = true
		case "sec-ch-ua-bitness":
			analysis.HighEntropyHints["Sec-CH-UA-Bitness"] = h.Value
			analysis.HasHighEntropyHints = true
		case "sec-ch-ua-full-version-list":
			analysis.HighEntropyHints["Sec-CH-UA-Full-Version-List"] = h.Value
			analysis.HasHighEntropyHints = true
		case "sec-ch-ua-platform-version":
			analysis.HighEntropyHints["Sec-CH-UA-Platform-Version"] = h.Value
			analysis.HasHighEntropyHints = true
		case "sec-ch-ua-model":
			analysis.HighEntropyHints["Sec-CH-UA-Model"] = h.Value
			analysis.HasHighEntropyHints = true
		case "sec-ch-ua-wow64":
			analysis.HighEntropyHints["Sec-CH-UA-WoW64"] = h.Value
			analysis.HasHighEntropyHints = true
		}
	}

	// 检测不一致性
	analysis.Inconsistencies = detectClientHintsInconsistencies(req.Headers)

	// 识别浏览器类型
	if secCHUA, ok := analysis.LowEntropyHints["Sec-CH-UA"]; ok {
		ua := strings.ToLower(secCHUA)
		if strings.Contains(ua, "chrome") {
			analysis.BrowserType = "Chrome"
		} else if strings.Contains(ua, "edge") {
			analysis.BrowserType = "Edge"
		} else if strings.Contains(ua, "chromium") {
			analysis.BrowserType = "Chromium"
		}
	}

	return analysis
}

// ClientHintsAnalysis Client Hints 分析结果
type ClientHintsAnalysis struct {
	HasLowEntropyHints  bool
	HasHighEntropyHints bool
	LowEntropyHints     map[string]string
	HighEntropyHints    map[string]string
	Inconsistencies     []string
	BrowserType         string
}

// ============ 辅助函数 ============

func hasHeader(headers []struct {
	Name  string
	Value string
}, name string) bool {
	name = strings.ToLower(name)
	for _, h := range headers {
		if strings.ToLower(h.Name) == name {
			return true
		}
	}
	return false
}

func getHeaderValue(headers []struct {
	Name  string
	Value string
}, name string) string {
	name = strings.ToLower(name)
	for _, h := range headers {
		if strings.ToLower(h.Name) == name {
			return h.Value
		}
	}
	return ""
}

func contains(slice []string, item string) bool {
	item = strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == item {
			return true
		}
	}
	return false
}

// ============ 已知配置库 ============

func initKnownBrowserProfiles() map[string]*HTTP2BrowserProfile {
	return map[string]*HTTP2BrowserProfile{
		"chrome_windows": {
			Name:           "Chrome on Windows",
			BrowserName:    "Chrome",
			BrowserVersion: "120+",
			HeaderOrder: []string{
				"host", "user-agent", "accept", "accept-language",
				"accept-encoding", "connection", "sec-fetch-dest",
				"sec-fetch-mode", "sec-fetch-site", "upgrade-insecure-requests",
			},
			TypicalHeaders: map[string]string{
				"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				"accept-language": "en-US,en;q=0.5",
				"accept-encoding": "gzip, deflate",
			},
			HeaderStrategy: "standard",
			RiskScore:      0.05,
		},
		"firefox_windows": {
			Name:           "Firefox on Windows",
			BrowserName:    "Firefox",
			BrowserVersion: "121+",
			HeaderOrder: []string{
				"host", "user-agent", "accept", "accept-language",
				"accept-encoding", "connection",
			},
			TypicalHeaders: map[string]string{
				"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				"accept-language": "en-US,en;q=0.5",
				"accept-encoding": "gzip, deflate",
			},
			HeaderStrategy: "standard",
			RiskScore:      0.08,
		},
		"safari_macos": {
			Name:           "Safari on macOS",
			BrowserName:    "Safari",
			BrowserVersion: "17+",
			HeaderOrder: []string{
				"host", "user-agent", "accept", "accept-language",
				"accept-encoding", "connection",
			},
			TypicalHeaders: map[string]string{
				"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				"accept-language": "en-US,en;q=0.5",
				"accept-encoding": "gzip, deflate, br",
			},
			HeaderStrategy: "standard",
			RiskScore:      0.06,
		},
	}
}

// ComputeJA4H 便捷函数：计算 JA4H
func ComputeJA4H(req HTTP2RequestData) (*JA4HResult, error) {
	analyzer := NewJA4HAnalyzer()
	return analyzer.AnalyzeHTTPRequest(req)
}
