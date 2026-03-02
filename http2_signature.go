package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// HTTP2SignatureResult HTTP/2 签名结果
type HTTP2SignatureResult struct {
	// 完整 HTTP/2 签名（SHA256）
	Hash string

	// 原始签名字符串（用于调试/验证）
	RawSignature string

	// 分解的签名部分
	SettingsSignature     string // Settings 帧签名
	PrioritySignature     string // Priority 帧签名
	HeadersSignature      string // Headers 帧签名
	WindowUpdateSignature string // WindowUpdate 帧签名

	// 综合特征
	FrameSequence string // 帧发送顺序特征

	// 异常判定分数
	RiskScore float64

	// 异常标记列表
	AnomalyFlags []string

	// 匹配的已知客户端指纹
	MatchedClients []string
}

// HTTP2FrameData HTTP/2 帧数据
type HTTP2FrameData struct {
	Type       string                 // SETTINGS, PRIORITY, HEADERS, WINDOW_UPDATE
	FrameID    uint32                 // 流 ID
	Priority   *PriorityData          // 优先级信息
	Settings   map[string]interface{} // SETTINGS 帧参数
	Headers    []string               // 请求头顺序
	WindowSize uint32                 // WINDOW_UPDATE 大小
	Metadata   map[string]string      // 其他元数据
}

// PriorityData 优先级信息
type PriorityData struct {
	DependsOn  uint32
	StreamDep  uint32
	Exclusive  bool
	Weight     uint8
	DefaultExp bool
}

// HTTP2SignatureAnalyzer HTTP/2 签名分析器
type HTTP2SignatureAnalyzer struct {
	knownClientProfiles map[string]*HTTP2ClientProfile
}

// HTTP2ClientProfile 已知的客户端配置
type HTTP2ClientProfile struct {
	Name               string
	BrowserName        string
	BrowserVersion     string
	SettingsOrder      []string
	PriorityStrategy   string
	HeaderOrder        []string
	WindowUpdateTiming string
	RiskScore          float64
}

// NewHTTP2SignatureAnalyzer 创建分析器
func NewHTTP2SignatureAnalyzer() *HTTP2SignatureAnalyzer {
	return &HTTP2SignatureAnalyzer{
		knownClientProfiles: initKnownHTTP2Profiles(),
	}
}

// AnalyzeHTTP2Stream 分析 HTTP/2 流特征
func (a *HTTP2SignatureAnalyzer) AnalyzeHTTP2Stream(frames []HTTP2FrameData) (*HTTP2SignatureResult, error) {
	if len(frames) == 0 {
		return nil, fmt.Errorf("no frames provided")
	}

	result := &HTTP2SignatureResult{
		AnomalyFlags:      make([]string, 0, 8), // 预分配容量
		SettingsSignature: "",
		PrioritySignature: "",
		HeadersSignature:  "",
	}

	// 分类处理帧
	var settingsFrames, priorityFrames, headersFrames, windowUpdateFrames []HTTP2FrameData

	for _, frame := range frames {
		switch frame.Type {
		case "SETTINGS":
			settingsFrames = append(settingsFrames, frame)
		case "PRIORITY":
			priorityFrames = append(priorityFrames, frame)
		case "HEADERS":
			headersFrames = append(headersFrames, frame)
		case "WINDOW_UPDATE":
			windowUpdateFrames = append(windowUpdateFrames, frame)
		}
	}

	// 1. SETTINGS 帧签名
	if len(settingsFrames) > 0 {
		result.SettingsSignature = generateSettingsSignature(settingsFrames[0])
	}

	// 2. PRIORITY 帧签名
	if len(priorityFrames) > 0 {
		result.PrioritySignature = generatePrioritySignature(priorityFrames)
	}

	// 3. HEADERS 帧签名（请求头顺序）
	if len(headersFrames) > 0 {
		result.HeadersSignature = generateHeadersSignature(headersFrames[0])
	}

	// 4. WINDOW_UPDATE 帧签名
	if len(windowUpdateFrames) > 0 {
		result.WindowUpdateSignature = generateWindowUpdateSignature(windowUpdateFrames)
	}

	// 5. 帧序列特征
	var frameTypes []string
	for _, frame := range frames {
		frameTypes = append(frameTypes, strings.ToLower(frame.Type[:3])) // SET, PRI, HEA, WIN
	}
	result.FrameSequence = strings.Join(frameTypes, "-")

	// 构建完整签名字符串
	result.RawSignature = fmt.Sprintf(
		"http2|%s|%s|%s|%s|%s",
		result.SettingsSignature,
		result.PrioritySignature,
		result.HeadersSignature,
		result.WindowUpdateSignature,
		result.FrameSequence,
	)

	// 计算 SHA256 哈希
	hash := sha256.Sum256([]byte(result.RawSignature))
	result.Hash = hex.EncodeToString(hash[:])

	// 异常检测
	a.detectHTTP2Anomalies(result, settingsFrames, priorityFrames, headersFrames)

	return result, nil
}

// detectHTTP2Anomalies 检测 HTTP/2 异常
func (a *HTTP2SignatureAnalyzer) detectHTTP2Anomalies(
	result *HTTP2SignatureResult,
	settingsFrames, priorityFrames, headersFrames []HTTP2FrameData,
) {
	baseScore := 0.0

	// 异常 1: SETTINGS 帧参数异常
	if len(settingsFrames) > 0 {
		if settingsFrames[0].Settings == nil || len(settingsFrames[0].Settings) == 0 {
			result.AnomalyFlags = append(result.AnomalyFlags, "EMPTY_SETTINGS")
			baseScore += 0.2
		}
		if len(settingsFrames[0].Settings) > 10 {
			result.AnomalyFlags = append(result.AnomalyFlags, "EXCESSIVE_SETTINGS")
			baseScore += 0.15
		}
	}

	// 异常 2: 优先级树结构异常
	if len(priorityFrames) > 0 {
		if !isValidPriorityTree(priorityFrames) {
			result.AnomalyFlags = append(result.AnomalyFlags, "INVALID_PRIORITY_TREE")
			baseScore += 0.25
		}
	}

	// 异常 3: 请求头顺序异常
	if len(headersFrames) > 0 {
		if !isStandardHeaderOrder(headersFrames[0]) {
			result.AnomalyFlags = append(result.AnomalyFlags, "UNUSUAL_HEADER_ORDER")
			baseScore += 0.15
		}
	}

	// 异常 4: 帧序列不合理
	if !isValidFrameSequence(result.FrameSequence) {
		result.AnomalyFlags = append(result.AnomalyFlags, "INVALID_FRAME_SEQUENCE")
		baseScore += 0.2
	}

	// 异常 5: 窗口更新参数异常
	if result.WindowUpdateSignature == "" && len(settingsFrames) > 0 {
		// 有 SETTINGS 但没有对应的 WINDOW_UPDATE
		result.AnomalyFlags = append(result.AnomalyFlags, "MISSING_WINDOW_UPDATE")
		baseScore += 0.1
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}
	result.RiskScore = baseScore
}

// FindMatchingHTTP2Clients 查找匹配的已知 HTTP/2 客户端
func (a *HTTP2SignatureAnalyzer) FindMatchingHTTP2Clients(
	result *HTTP2SignatureResult,
	maxResults int,
) []string {
	var matches []string

	for name, profile := range a.knownClientProfiles {
		// 基于风险分数和签名特征的粗略匹配
		if profile.RiskScore < result.RiskScore+0.2 {
			matches = append(matches, name)
		}

		if len(matches) >= maxResults {
			break
		}
	}

	result.MatchedClients = matches
	return matches
}

// ============ 辅助生成函数 ============

func generateSettingsSignature(frame HTTP2FrameData) string {
	if frame.Settings == nil {
		return "empty"
	}

	// 对 SETTINGS 参数排序并生成签名
	var keys []string
	for k := range frame.Settings {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		v := frame.Settings[k]
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}

	// 计算简单的哈希值
	sig := strings.Join(parts, ",")
	hash := sha256.Sum256([]byte(sig))
	return hex.EncodeToString(hash[:8]) // 取前 16 个字符
}

func generatePrioritySignature(frames []HTTP2FrameData) string {
	// 基于优先级树的深度和宽度
	if len(frames) == 0 {
		return "empty"
	}

	var sig strings.Builder
	for i, frame := range frames {
		if frame.Priority != nil {
			sig.WriteString(fmt.Sprintf("#%d(%d,%d)", i, frame.Priority.StreamDep, frame.Priority.Weight))
		}
	}

	// 计算哈希
	text := sig.String()
	if text == "" {
		return "no_priority"
	}
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:8])
}

func generateHeadersSignature(frame HTTP2FrameData) string {
	// 基于请求头顺序和数量
	if len(frame.Headers) == 0 {
		return "empty"
	}

	// 伪代码头应该首先出现
	var sig strings.Builder
	for _, h := range frame.Headers {
		if strings.HasPrefix(h, ":") {
			sig.WriteString("PSD,")
		} else {
			sig.WriteString("HDR,")
		}
	}

	text := sig.String()
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:8])
}

func generateWindowUpdateSignature(frames []HTTP2FrameData) string {
	// 基于 WINDOW_UPDATE 的规模和频率
	if len(frames) == 0 {
		return "empty"
	}

	var totalSize uint32
	for _, frame := range frames {
		totalSize += frame.WindowSize
	}

	// 分级：小(0-65k), 中(65k-1M), 大(>1M)
	var level string
	if totalSize < 65536 {
		level = "small"
	} else if totalSize < 1048576 {
		level = "medium"
	} else {
		level = "large"
	}

	sig := fmt.Sprintf("win_update|count=%d|level=%s", len(frames), level)
	hash := sha256.Sum256([]byte(sig))
	return hex.EncodeToString(hash[:8])
}

// ============ 验证函数 ============

func isValidPriorityTree(frames []HTTP2FrameData) bool {
	// 简化检查：至少有一个有效的优先级帧
	for _, frame := range frames {
		if frame.Priority != nil && frame.Priority.StreamDep > 0 {
			return true
		}
	}
	// 如果没有优先级帧也是有效的（流 0 是隐含的）
	return true
}

func isStandardHeaderOrder(frame HTTP2FrameData) bool {
	// 检查伪代码头（以 : 开头）是否排在前面
	if len(frame.Headers) < 2 {
		return true
	}

	firstPseudoIdx := -1
	firstNormalIdx := -1

	for i, h := range frame.Headers {
		if strings.HasPrefix(h, ":") && firstPseudoIdx == -1 {
			firstPseudoIdx = i
		}
		if !strings.HasPrefix(h, ":") && firstNormalIdx == -1 {
			firstNormalIdx = i
		}
	}

	// 伪代码头应该在普通头之前
	if firstPseudoIdx != -1 && firstNormalIdx != -1 {
		return firstPseudoIdx < firstNormalIdx
	}

	return true
}

func isValidFrameSequence(frameSeq string) bool {
	// SETTINGS 通常首先出现，然后是其他帧类型
	// set -> ... 的序列是最常见的
	return strings.HasPrefix(frameSeq, "set")
}

// ============ 已知配置库 ============

func initKnownHTTP2Profiles() map[string]*HTTP2ClientProfile {
	return map[string]*HTTP2ClientProfile{
		"chrome_default": {
			Name:               "Chrome Default",
			BrowserName:        "Chrome",
			BrowserVersion:     "120+",
			SettingsOrder:      []string{"HEADER_TABLE_SIZE", "ENABLE_PUSH", "INITIAL_WINDOW_SIZE"},
			PriorityStrategy:   "balanced",
			HeaderOrder:        []string{":method", ":scheme", ":authority", ":path", "user-agent"},
			WindowUpdateTiming: "immediate",
			RiskScore:          0.05,
		},
		"firefox_default": {
			Name:               "Firefox Default",
			BrowserName:        "Firefox",
			BrowserVersion:     "120+",
			SettingsOrder:      []string{"ENABLE_PUSH", "INITIAL_WINDOW_SIZE"},
			PriorityStrategy:   "weighted",
			HeaderOrder:        []string{":method", ":scheme", ":authority", ":path"},
			WindowUpdateTiming: "batched",
			RiskScore:          0.1,
		},
		"safari_default": {
			Name:               "Safari Default",
			BrowserName:        "Safari",
			BrowserVersion:     "17+",
			SettingsOrder:      []string{"HEADER_TABLE_SIZE", "INITIAL_WINDOW_SIZE"},
			PriorityStrategy:   "default",
			HeaderOrder:        []string{":method", ":scheme", ":authority", ":path"},
			WindowUpdateTiming: "deferred",
			RiskScore:          0.08,
		},
	}
}

// ComputeHTTP2Signature 便捷函数：计算 HTTP/2 签名
func ComputeHTTP2Signature(frames []HTTP2FrameData) (*HTTP2SignatureResult, error) {
	analyzer := NewHTTP2SignatureAnalyzer()
	return analyzer.AnalyzeHTTP2Stream(frames)
}
