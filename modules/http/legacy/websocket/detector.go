package websocket

import (
	"fmt"
	"net/http"
	"strings"
)

// AnomalyType 异常类型
type AnomalyType string

const (
	// AnomalyInvalidMethod 无效的 HTTP 方法
	AnomalyInvalidMethod AnomalyType = "invalid_method"
	// AnomalyMissingHeaders 缺少必需头部
	AnomalyMissingHeaders AnomalyType = "missing_headers"
	// AnomalySuspiciousKey 可疑的 WebSocket Key
	AnomalySuspiciousKey AnomalyType = "suspicious_key"
	// AnomalyLowEntropyKey 低熵 Key（可能是伪随机）
	AnomalyLowEntropyKey AnomalyType = "low_entropy_key"
	// AnomalyAbnormalHeaderOrder 异常的头部顺序
	AnomalyAbnormalHeaderOrder AnomalyType = "abnormal_header_order"
	// AnomalySuspiciousExtensions 可疑的扩展
	AnomalySuspiciousExtensions AnomalyType = "suspicious_extensions"
	// AnomalyKnownBotSignature 已知的机器人特征
	AnomalyKnownBotSignature AnomalyType = "known_bot_signature"
	// AnomalyFrameAnomaly 帧异常
	AnomalyFrameAnomaly AnomalyType = "frame_anomaly"
	// AnomalyVersionMismatch 版本不匹配
	AnomalyVersionMismatch AnomalyType = "version_mismatch"
)

// Anomaly 异常检测结果的详细异常信息
type Anomaly struct {
	Type        AnomalyType
	Description string
	Severity    Severity
	Evidence    map[string]interface{}
}

// Severity 严重程度
type Severity string

const (
	// SeverityInfo 信息级别
	SeverityInfo Severity = "info"
	// SeverityLow 低级别
	SeverityLow Severity = "low"
	// SeverityMedium 中级别
	SeverityMedium Severity = "medium"
	// SeverityHigh 高级别
	SeverityHigh Severity = "high"
	// SeverityCritical 严重级别
	SeverityCritical Severity = "critical"
)

// Detector WebSocket 异常检测器
type Detector struct {
	// 已知的机器人 User-Agent 模式
	botPatterns []string
	// 已知的异常 Key 模式
	suspiciousKeyPatterns [][]byte
	// 正常的头部顺序（用于比较）
	normalHeaderOrders map[string][]string
}

// DetectionResult 检测结果
type DetectionResult struct {
	// 是否检测到异常
	HasAnomaly bool
	// 异常列表
	Anomalies []Anomaly
	// 风险评分 (0-100)
	RiskScore int
	// 检测到的浏览器类型（如果有）
	DetectedBrowser string
	// 是否为已知的自动化工具
	IsKnownBot bool
}

// NewDetector 创建新的检测器
func NewDetector() *Detector {
	return &Detector{
		botPatterns: []string{
			"bot", "crawler", "spider", "scraper",
			"automation", "headless", "puppeteer",
			"selenium", "playwright", "cypress",
		},
		suspiciousKeyPatterns: [][]byte{
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // 全零
			{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, // 递增
		},
		normalHeaderOrders: map[string][]string{
			"chrome": {
				"Host", "Connection", "Pragma", "Cache-Control",
				"User-Agent", "Upgrade", "Origin",
				"Sec-WebSocket-Version", "Sec-WebSocket-Key",
				"Sec-WebSocket-Extensions", "Sec-WebSocket-Protocol",
			},
			"firefox": {
				"Host", "User-Agent", "Accept", "Accept-Language",
				"Accept-Encoding", "Sec-WebSocket-Version",
				"Origin", "Sec-WebSocket-Extensions",
				"Sec-WebSocket-Key", "Connection", "Upgrade",
			},
			"safari": {
				"Host", "Upgrade", "Connection",
				"Sec-WebSocket-Key", "Sec-WebSocket-Version",
				"Origin", "Sec-WebSocket-Extensions",
				"User-Agent", "Accept", "Accept-Language",
			},
		},
	}
}

// Detect 检测 WebSocket 请求的异常
func (d *Detector) Detect(req *http.Request, fp *WebSocketFingerprint) *DetectionResult {
	result := &DetectionResult{
		Anomalies: make([]Anomaly, 0),
	}

	// 1. 验证 HTTP 方法
	d.checkMethod(req, result)

	// 2. 检查必需头部
	d.checkRequiredHeaders(req, result)

	// 3. 分析 WebSocket Key
	d.analyzeKey(fp, result)

	// 4. 分析头部顺序
	d.analyzeHeaderOrder(fp, result)

	// 5. 检查扩展
	d.checkExtensions(fp, result)

	// 6. 检测已知机器人
	d.detectBot(req, result)

	// 7. 计算风险评分
	result.RiskScore = d.calculateRiskScore(result)
	result.HasAnomaly = len(result.Anomalies) > 0

	return result
}

// DetectFrameAnomalies 检测帧异常
func (d *Detector) DetectFrameAnomalies(frame *Frame) *DetectionResult {
	result := &DetectionResult{
		Anomalies: make([]Anomaly, 0),
	}

	// 检查 RSV 位（应为 0，除非使用扩展）
	if frame.RSV1 || frame.RSV2 || frame.RSV3 {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalyFrameAnomaly,
			Description: "RSV bits are set without proper extension negotiation",
			Severity:    SeverityMedium,
			Evidence: map[string]interface{}{
				"rsv1": frame.RSV1,
				"rsv2": frame.RSV2,
				"rsv3": frame.RSV3,
			},
		})
	}

	// 检查服务端发送的帧是否未掩码（服务端不应掩码）
	if !frame.MASK && frame.Opcode != OpCodeClose {
		// 这可能是客户端帧，需要进一步验证
		// 实际检测需要上下文
	}

	// 检查控制帧大小（控制帧载荷不应超过 125 字节）
	if frame.Opcode >= 0x8 && frame.PayloadLength > 125 {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalyFrameAnomaly,
			Description: "Control frame payload exceeds 125 bytes",
			Severity:    SeverityHigh,
			Evidence: map[string]interface{}{
				"opcode":         frame.Opcode,
				"payload_length": frame.PayloadLength,
			},
		})
	}

	// 检查异常大的 payload
	if frame.PayloadLength > 10*1024*1024 { // 10MB
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalyFrameAnomaly,
			Description: "Unusually large payload detected",
			Severity:    SeverityMedium,
			Evidence: map[string]interface{}{
				"payload_length": frame.PayloadLength,
			},
		})
	}

	result.RiskScore = d.calculateRiskScore(result)
	result.HasAnomaly = len(result.Anomalies) > 0

	return result
}

// checkMethod 检查 HTTP 方法
func (d *Detector) checkMethod(req *http.Request, result *DetectionResult) {
	if req.Method != "GET" {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalyInvalidMethod,
			Description: fmt.Sprintf("Invalid HTTP method: %s (expected GET)", req.Method),
			Severity:    SeverityHigh,
			Evidence: map[string]interface{}{
				"method": req.Method,
			},
		})
	}
}

// checkRequiredHeaders 检查必需头部
func (d *Detector) checkRequiredHeaders(req *http.Request, result *DetectionResult) {
	requiredHeaders := map[string]string{
		"Upgrade":               "websocket",
		"Connection":            "Upgrade",
		"Sec-Websocket-Key":     "",
		"Sec-Websocket-Version": "13",
	}

	missingHeaders := []string{}
	invalidHeaders := []string{}

	for header, expectedValue := range requiredHeaders {
		value := req.Header.Get(header)
		if value == "" {
			// 尝试小写版本
			value = req.Header.Get(strings.ToLower(header))
		}

		if value == "" {
			missingHeaders = append(missingHeaders, header)
		} else if expectedValue != "" && !strings.EqualFold(value, expectedValue) {
			invalidHeaders = append(invalidHeaders, fmt.Sprintf("%s=%s", header, value))
		}
	}

	if len(missingHeaders) > 0 {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalyMissingHeaders,
			Description: fmt.Sprintf("Missing required headers: %v", missingHeaders),
			Severity:    SeverityHigh,
			Evidence: map[string]interface{}{
				"missing_headers": missingHeaders,
			},
		})
	}

	if len(invalidHeaders) > 0 {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalyMissingHeaders,
			Description: fmt.Sprintf("Invalid header values: %v", invalidHeaders),
			Severity:    SeverityMedium,
			Evidence: map[string]interface{}{
				"invalid_headers": invalidHeaders,
			},
		})
	}
}

// analyzeKey 分析 WebSocket Key
func (d *Detector) analyzeKey(fp *WebSocketFingerprint, result *DetectionResult) {
	keyChar := fp.Handshake.SecWebSocketKeyCharacteristics

	// 检查 Key 长度
	if keyChar.Length != 24 {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalySuspiciousKey,
			Description: fmt.Sprintf("Invalid Sec-WebSocket-Key length: %d (expected 24)", keyChar.Length),
			Severity:    SeverityHigh,
			Evidence: map[string]interface{}{
				"key_length": keyChar.Length,
			},
		})
	}

	// 检查是否为标准 Base64
	if !keyChar.IsStandardBase64 {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalySuspiciousKey,
			Description: "Sec-WebSocket-Key is not valid Base64 encoded 16-byte value",
			Severity:    SeverityMedium,
			Evidence: map[string]interface{}{
				"is_standard_base64": keyChar.IsStandardBase64,
			},
		})
	}

	// 检查熵值
	if keyChar.Entropy < 3.0 {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalyLowEntropyKey,
			Description: fmt.Sprintf("Low entropy in Sec-WebSocket-Key: %.2f (suspicious)", keyChar.Entropy),
			Severity:    SeverityMedium,
			Evidence: map[string]interface{}{
				"entropy":       keyChar.Entropy,
				"pattern_type":  keyChar.PatternType,
				"has_pattern":   keyChar.HasPattern,
			},
		})
	}

	// 检查已知模式
	if keyChar.HasPattern && keyChar.PatternType != "" {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalySuspiciousKey,
			Description: fmt.Sprintf("Detected pattern in Sec-WebSocket-Key: %s", keyChar.PatternType),
			Severity:    SeverityHigh,
			Evidence: map[string]interface{}{
				"pattern_type": keyChar.PatternType,
			},
		})
	}
}

// analyzeHeaderOrder 分析头部顺序
func (d *Detector) analyzeHeaderOrder(fp *WebSocketFingerprint, result *DetectionResult) {
	if len(fp.Handshake.HeaderOrder) == 0 {
		return
	}

	// 检查 User-Agent 以识别浏览器
	browser := d.identifyBrowserFromUA(fp.Handshake.UserAgent)
	if browser == "" {
		return
	}

	normalOrder, exists := d.normalHeaderOrders[browser]
	if !exists {
		return
	}

	// 计算顺序匹配度
	matchScore := d.calculateHeaderOrderMatch(fp.Handshake.HeaderOrder, normalOrder)

	// 如果匹配度太低，可能是异常的客户端
	if matchScore < 0.3 {
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        AnomalyAbnormalHeaderOrder,
			Description: fmt.Sprintf("Abnormal header order for %s (match score: %.2f)", browser, matchScore),
			Severity:    SeverityLow,
			Evidence: map[string]interface{}{
				"browser":       browser,
				"match_score":   matchScore,
				"actual_order":  fp.Handshake.HeaderOrder,
				"normal_order":  normalOrder,
			},
		})
	}
}

// checkExtensions 检查扩展
func (d *Detector) checkExtensions(fp *WebSocketFingerprint, result *DetectionResult) {
	// 检查是否有已知的可疑扩展
	suspiciousExts := []string{"x-unknown-extension", "malicious-ext"}

	for _, ext := range fp.Extensions {
		for _, suspicious := range suspiciousExts {
			if strings.EqualFold(ext, suspicious) {
				result.Anomalies = append(result.Anomalies, Anomaly{
					Type:        AnomalySuspiciousExtensions,
					Description: fmt.Sprintf("Suspicious extension detected: %s", ext),
					Severity:    SeverityHigh,
					Evidence: map[string]interface{}{
						"extension": ext,
					},
				})
			}
		}
	}
}

// detectBot 检测机器人
func (d *Detector) detectBot(req *http.Request, result *DetectionResult) {
	ua := req.Header.Get("User-Agent")
	if ua == "" {
		return
	}

	uaLower := strings.ToLower(ua)

	for _, pattern := range d.botPatterns {
		if strings.Contains(uaLower, pattern) {
			result.IsKnownBot = true
			result.Anomalies = append(result.Anomalies, Anomaly{
				Type:        AnomalyKnownBotSignature,
				Description: fmt.Sprintf("Detected known bot/automation pattern: %s", pattern),
				Severity:    SeverityHigh,
				Evidence: map[string]interface{}{
					"pattern":     pattern,
					"user_agent":  ua,
				},
			})
			break
		}
	}
}

// calculateRiskScore 计算风险评分
func (d *Detector) calculateRiskScore(result *DetectionResult) int {
	if len(result.Anomalies) == 0 {
		return 0
	}

	score := 0
	for _, anomaly := range result.Anomalies {
		switch anomaly.Severity {
		case SeverityInfo:
			score += 5
		case SeverityLow:
			score += 15
		case SeverityMedium:
			score += 30
		case SeverityHigh:
			score += 50
		case SeverityCritical:
			score += 100
		}
	}

	// 如果是已知机器人，增加风险分
	if result.IsKnownBot {
		score += 50
	}

	// 限制在 0-100
	if score > 100 {
		score = 100
	}

	return score
}

// identifyBrowserFromUA 从 User-Agent 识别浏览器
func (d *Detector) identifyBrowserFromUA(ua string) string {
	uaLower := strings.ToLower(ua)

	if strings.Contains(uaLower, "chrome") && !strings.Contains(uaLower, "edg") {
		return "chrome"
	}
	if strings.Contains(uaLower, "firefox") {
		return "firefox"
	}
	if strings.Contains(uaLower, "safari") && !strings.Contains(uaLower, "chrome") {
		return "safari"
	}
	if strings.Contains(uaLower, "edge") || strings.Contains(uaLower, "edg") {
		return "edge"
	}

	return ""
}

// calculateHeaderOrderMatch 计算头部顺序匹配度
func (d *Detector) calculateHeaderOrderMatch(actual, normal []string) float64 {
	if len(actual) == 0 || len(normal) == 0 {
		return 0.0
	}

	// 创建位置映射，使用 -1 表示不存在
	normalPos := make(map[string]int)
	for i, h := range normal {
		normalPos[strings.ToLower(h)] = i
	}

	// 计算匹配的前几个头部
	matchCount := 0
	checkCount := min(len(actual), len(normal), 5) // 检查前5个

	for i := 0; i < checkCount; i++ {
		if i < len(actual) {
			actualLower := strings.ToLower(actual[i])
			// 只有当头部存在于 normal 中且位置匹配时才计数
			if pos, exists := normalPos[actualLower]; exists && pos == i {
				matchCount++
			}
		}
	}

	return float64(matchCount) / float64(checkCount)
}

// GetSeverityWeight 获取严重程度权重
func GetSeverityWeight(s Severity) int {
	switch s {
	case SeverityInfo:
		return 1
	case SeverityLow:
		return 2
	case SeverityMedium:
		return 3
	case SeverityHigh:
		return 4
	case SeverityCritical:
		return 5
	default:
		return 0
	}
}

// min 返回最小值
func min(values ...int) int {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values[1:] {
		if v < m {
			m = v
		}
	}
	return m
}
