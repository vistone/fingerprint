package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

// WebSocketFingerprint WebSocket 指纹
type WebSocketFingerprint struct {
	// 协议版本
	Version string

	// 握手特征
	Handshake WebSocketHandshake

	// 帧特征
	FrameCharacteristics FrameCharacteristics

	// 扩展支持
	Extensions []string

	// 子协议
	SubProtocols []string

	// 原始指纹字符串
	Raw string

	// 哈希值
	Hash string
}

// WebSocketHandshake 握手特征
type WebSocketHandshake struct {
	// HTTP 版本
	HTTPVersion string

	// 方法（应为 GET）
	Method string

	// 头部顺序
	HeaderOrder []string

	// 特定头部值
	Headers map[string]string

	// User-Agent
	UserAgent string

	//  Origin
	Origin string

	// Sec-WebSocket-Version
	SecWebSocketVersion string

	// Sec-WebSocket-Key 特征
	SecWebSocketKeyCharacteristics KeyCharacteristics

	// Sec-WebSocket-Extensions
	SecWebSocketExtensions string

	// Sec-WebSocket-Protocol
	SecWebSocketProtocol string
}

// KeyCharacteristics WebSocket Key 特征
type KeyCharacteristics struct {
	// Key 长度
	Length int

	// 是否为标准 Base64（16字节随机数据）
	IsStandardBase64 bool

	// 是否包含特定模式
	HasPattern bool

	// 模式类型
	PatternType string

	// 熵值估计
	Entropy float64
}

// FrameCharacteristics 帧特征
type FrameCharacteristics struct {
	// 掩码行为
	MaskingBehavior string

	// 最大帧大小
	MaxFrameSize int

	// 支持的 opcode
	SupportedOpcodes []uint8

	// 扩展标志使用
	ExtensionFlags []string

	// 控制帧行为
	ControlFrameBehavior string
}

// Analyzer WebSocket 指纹分析器
type Analyzer struct {
	// 浏览器特征数据库
	browserPatterns map[string]BrowserPattern
}

// BrowserPattern 浏览器特征模式
type BrowserPattern struct {
	Name           string
	Version        string
	HeaderOrder    []string
	KeyPattern     string
	Extensions     []string
	UserAgentMatch *regexp.Regexp
}

// NewAnalyzer 创建分析器
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		browserPatterns: initBrowserPatterns(),
	}
}

// AnalyzeRequest 分析 WebSocket 握手请求
func (a *Analyzer) AnalyzeRequest(req *http.Request) (*WebSocketFingerprint, error) {
	if req.Method != "GET" {
		return nil, fmt.Errorf("invalid method: %s", req.Method)
	}

	fp := &WebSocketFingerprint{
		Version: "13", // RFC 6455
		Handshake: WebSocketHandshake{
			HTTPVersion: req.Proto,
			Method:      req.Method,
			Headers:     make(map[string]string),
		},
	}

	// 分析头部
	a.analyzeHeaders(req, fp)

	// 分析 Sec-WebSocket-Key
	a.analyzeSecWebSocketKey(req, fp)

	// 分析扩展
	a.analyzeExtensions(req, fp)

	// 分析子协议
	a.analyzeSubProtocols(req, fp)

	// 生成指纹字符串和哈希
	fp.Raw = a.generateFingerprintString(fp)
	fp.Hash = a.generateFingerprintHash(fp.Raw)

	return fp, nil
}

// analyzeHeaders 分析 HTTP 头部
func (a *Analyzer) analyzeHeaders(req *http.Request, fp *WebSocketFingerprint) {
	// 记录头部顺序
	for name := range req.Header {
		fp.Handshake.HeaderOrder = append(fp.Handshake.HeaderOrder, name)
	}

	// 记录特定头部
	if ua := req.Header.Get("User-Agent"); ua != "" {
		fp.Handshake.UserAgent = ua
	}

	if origin := req.Header.Get("Origin"); origin != "" {
		fp.Handshake.Origin = origin
	}

	if version := req.Header.Get("Sec-Websocket-Version"); version != "" {
		fp.Handshake.SecWebSocketVersion = version
	}

	if ext := req.Header.Get("Sec-Websocket-Extensions"); ext != "" {
		fp.Handshake.SecWebSocketExtensions = ext
	}

	if proto := req.Header.Get("Sec-Websocket-Protocol"); proto != "" {
		fp.Handshake.SecWebSocketProtocol = proto
	}

	// 记录所有 Sec-WebSocket-* 头部
	for name, values := range req.Header {
		if strings.HasPrefix(name, "Sec-Websocket-") || strings.HasPrefix(name, "Sec-WebSocket-") {
			fp.Handshake.Headers[name] = strings.Join(values, ", ")
		}
	}
}

// analyzeSecWebSocketKey 分析 Sec-WebSocket-Key
func (a *Analyzer) analyzeSecWebSocketKey(req *http.Request, fp *WebSocketFingerprint) {
	key := req.Header.Get("Sec-Websocket-Key")
	if key == "" {
		return
	}

	fp.Handshake.SecWebSocketKeyCharacteristics = KeyCharacteristics{
		Length: len(key),
	}

	// 解码 Base64
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		fp.Handshake.SecWebSocketKeyCharacteristics.IsStandardBase64 = false
		return
	}

	// 标准应该是 16 字节随机数据
	fp.Handshake.SecWebSocketKeyCharacteristics.IsStandardBase64 = len(decoded) == 16

	// 分析熵值（简化）
	fp.Handshake.SecWebSocketKeyCharacteristics.Entropy = calculateEntropy(decoded)

	// 检测模式
	a.detectKeyPattern(decoded, fp)
}

// detectKeyPattern 检测 Key 生成模式
func (a *Analyzer) detectKeyPattern(key []byte, fp *WebSocketFingerprint) {
	// 检测常见的随机数生成器模式
	if len(key) != 16 {
		return
	}

	// 检测是否全为零（测试客户端）
	allZero := true
	for _, b := range key {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		fp.Handshake.SecWebSocketKeyCharacteristics.HasPattern = true
		fp.Handshake.SecWebSocketKeyCharacteristics.PatternType = "all_zeros"
		return
	}

	// 检测递增模式（简单 RNG）
	incremental := true
	for i := 1; i < len(key); i++ {
		if key[i] != key[i-1]+1 {
			incremental = false
			break
		}
	}
	if incremental {
		fp.Handshake.SecWebSocketKeyCharacteristics.HasPattern = true
		fp.Handshake.SecWebSocketKeyCharacteristics.PatternType = "incremental"
		return
	}

	// 检测低熵（某些嵌入式设备）
	if fp.Handshake.SecWebSocketKeyCharacteristics.Entropy < 3.0 {
		fp.Handshake.SecWebSocketKeyCharacteristics.HasPattern = true
		fp.Handshake.SecWebSocketKeyCharacteristics.PatternType = "low_entropy"
	}
}

// analyzeExtensions 分析扩展
func (a *Analyzer) analyzeExtensions(req *http.Request, fp *WebSocketFingerprint) {
	extHeader := req.Header.Get("Sec-Websocket-Extensions")
	if extHeader == "" {
		return
	}

	// 解析扩展列表
	// 格式: extension1; param1=value1, extension2; param2=value2
	extensions := parseExtensionHeader(extHeader)
	fp.Extensions = extensions

	// 记录帧特征
	for _, ext := range extensions {
		switch ext {
		case "permessage-deflate":
			fp.FrameCharacteristics.ExtensionFlags = append(fp.FrameCharacteristics.ExtensionFlags, "compression")
		case "client_max_window_bits":
			fp.FrameCharacteristics.ExtensionFlags = append(fp.FrameCharacteristics.ExtensionFlags, "client_max_window_bits")
		case "server_max_window_bits":
			fp.FrameCharacteristics.ExtensionFlags = append(fp.FrameCharacteristics.ExtensionFlags, "server_max_window_bits")
		}
	}
}

// analyzeSubProtocols 分析子协议
func (a *Analyzer) analyzeSubProtocols(req *http.Request, fp *WebSocketFingerprint) {
	protoHeader := req.Header.Get("Sec-Websocket-Protocol")
	if protoHeader == "" {
		return
	}

	// 解析协议列表
	protocols := splitAndTrim(protoHeader, ",")
	fp.SubProtocols = protocols
}

// generateFingerprintString 生成指纹字符串
func (a *Analyzer) generateFingerprintString(fp *WebSocketFingerprint) string {
	var parts []string

	// HTTP 版本
	parts = append(parts, fmt.Sprintf("http=%s", fp.Handshake.HTTPVersion))

	// 头部顺序（排序后）
	sortedHeaders := make([]string, len(fp.Handshake.HeaderOrder))
	copy(sortedHeaders, fp.Handshake.HeaderOrder)
	sort.Strings(sortedHeaders)
	parts = append(parts, fmt.Sprintf("headers=%s", strings.Join(sortedHeaders, "|")))

	// WebSocket 版本
	parts = append(parts, fmt.Sprintf("ws_version=%s", fp.Handshake.SecWebSocketVersion))

	// Key 特征
	keyChar := fp.Handshake.SecWebSocketKeyCharacteristics
	parts = append(parts, fmt.Sprintf("key_std=%v", keyChar.IsStandardBase64))

	// 扩展
	if len(fp.Extensions) > 0 {
		sort.Strings(fp.Extensions)
		parts = append(parts, fmt.Sprintf("ext=%s", strings.Join(fp.Extensions, ",")))
	}

	return strings.Join(parts, ";")
}

// generateFingerprintHash 生成指纹哈希
func (a *Analyzer) generateFingerprintHash(raw string) string {
	h := sha1.New()
	h.Write([]byte(raw))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// IdentifyBrowser 识别浏览器类型
func (a *Analyzer) IdentifyBrowser(fp *WebSocketFingerprint) (string, float64) {
	if fp.Handshake.UserAgent == "" {
		return "unknown", 0.0
	}

	for name, pattern := range a.browserPatterns {
		if pattern.UserAgentMatch.MatchString(fp.Handshake.UserAgent) {
			// 进一步验证头部顺序
			if a.matchHeaderOrder(fp, pattern) {
				return name, 0.8
			}
			return name, 0.5
		}
	}

	return "unknown", 0.0
}

// matchHeaderOrder 验证头部顺序是否匹配
func (a *Analyzer) matchHeaderOrder(fp *WebSocketFingerprint, pattern BrowserPattern) bool {
	if len(pattern.HeaderOrder) == 0 {
		return true
	}

	// 简化匹配：检查前几个关键头部是否匹配
	matchCount := 0
	for i, header := range pattern.HeaderOrder {
		if i < len(fp.Handshake.HeaderOrder) &&
			equalIgnoreCase(fp.Handshake.HeaderOrder[i], header) {
			matchCount++
		}
	}

	return float64(matchCount)/float64(len(pattern.HeaderOrder)) > 0.7
}

// parseExtensionHeader 解析扩展头部
func parseExtensionHeader(header string) []string {
	var extensions []string

	parts := splitAndTrim(header, ",")
	for _, part := range parts {
		// 提取扩展名（去掉参数）
		if idx := strings.Index(part, ";"); idx != -1 {
			extensions = append(extensions, strings.TrimSpace(part[:idx]))
		} else {
			extensions = append(extensions, strings.TrimSpace(part))
		}
	}

	return extensions
}

// splitAndTrim 分割字符串并修剪空白
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// equalIgnoreCase 忽略大小写比较
func equalIgnoreCase(a, b string) bool {
	return strings.EqualFold(a, b)
}

// calculateEntropy 计算字节数组的熵值
func calculateEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0.0
	}

	// 统计字节频率
	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}

	// 计算熵
	var entropy float64
	length := float64(len(data))
	for _, count := range freq {
		p := float64(count) / length
		if p > 0 {
			entropy -= p * mathLog2(p)
		}
	}

	return entropy
}

// mathLog2 计算 log2
func mathLog2(x float64) float64 {
	// 简化实现，使用自然对数转换
	const ln2 = 0.6931471805599453
	if x <= 0 {
		return 0
	}
	// 使用 math 包
	return log2(x)
}

// log2 计算 log2 (简化实现)
func log2(x float64) float64 {
	// 使用换底公式: log2(x) = log(x) / log(2)
	// 这里使用近似算法
	result := 0.0
	for x >= 2.0 {
		x /= 2.0
		result++
	}
	if x > 1.0 {
		result += x - 1.0 // 近似
	}
	return result
}

// initBrowserPatterns 初始化浏览器模式
func initBrowserPatterns() map[string]BrowserPattern {
	return map[string]BrowserPattern{
		"chrome": {
			Name:    "Chrome",
			Version: "*",
			HeaderOrder: []string{
				"Host",
				"Connection",
				"Upgrade",
				"Origin",
				"Sec-WebSocket-Version",
				"Sec-WebSocket-Key",
				"Sec-WebSocket-Extensions",
			},
			UserAgentMatch: regexp.MustCompile(`Chrome/[\d.]+`),
		},
		"firefox": {
			Name:    "Firefox",
			Version: "*",
			HeaderOrder: []string{
				"Host",
				"User-Agent",
				"Accept",
				"Accept-Language",
				"Accept-Encoding",
				"Sec-WebSocket-Version",
				"Origin",
				"Sec-WebSocket-Extensions",
				"Sec-WebSocket-Key",
			},
			UserAgentMatch: regexp.MustCompile(`Firefox/[\d.]+`),
		},
		"safari": {
			Name:    "Safari",
			Version: "*",
			HeaderOrder: []string{
				"Host",
				"Upgrade",
				"Connection",
				"Sec-WebSocket-Key",
				"Sec-WebSocket-Version",
				"Origin",
			},
			UserAgentMatch: regexp.MustCompile(`Safari/[\d.]+`),
		},
	}
}

// CompareFingerprints 比较两个指纹的相似度
func CompareFingerprints(fp1, fp2 *WebSocketFingerprint) float64 {
	if fp1 == nil || fp2 == nil {
		return 0.0
	}

	score := 0.0
	weight := 0.0

	// 比较 HTTP 版本
	weight++
	if fp1.Handshake.HTTPVersion == fp2.Handshake.HTTPVersion {
		score++
	}

	// 比较头部顺序（Jaccard 相似度）
	weight++
	score += jaccardSimilarity(fp1.Handshake.HeaderOrder, fp2.Handshake.HeaderOrder)

	// 比较扩展
	weight++
	score += jaccardSimilarity(fp1.Extensions, fp2.Extensions)

	// 比较 Key 标准性
	weight++
	if fp1.Handshake.SecWebSocketKeyCharacteristics.IsStandardBase64 ==
		fp2.Handshake.SecWebSocketKeyCharacteristics.IsStandardBase64 {
		score++
	}

	return score / weight
}

// jaccardSimilarity 计算 Jaccard 相似度
func jaccardSimilarity(a, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// 创建集合
	setA := make(map[string]bool)
	for _, s := range a {
		setA[strings.ToLower(s)] = true
	}

	setB := make(map[string]bool)
	for _, s := range b {
		setB[strings.ToLower(s)] = true
	}

	// 计算交集和并集
	intersection := 0
	for k := range setA {
		if setB[k] {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection

	return float64(intersection) / float64(union)
}

// GenerateAcceptKey 生成 Sec-WebSocket-Accept 响应
func GenerateAcceptKey(key string) (string, error) {
	const magicGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

	h := sha1.New()
	h.Write([]byte(key))
	h.Write([]byte(magicGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// IsValidWebSocketRequest 验证是否为有效的 WebSocket 请求
func IsValidWebSocketRequest(req *http.Request) bool {
	// 检查必需头部
	if req.Method != "GET" {
		return false
	}

	if req.Header.Get("Upgrade") != "websocket" {
		return false
	}

	if !strings.Contains(req.Header.Get("Connection"), "Upgrade") {
		return false
	}

	if req.Header.Get("Sec-Websocket-Key") == "" {
		return false
	}

	if req.Header.Get("Sec-Websocket-Version") != "13" {
		return false
	}

	return true
}

// Frame WebSocket 帧结构
type Frame struct {
	// FIN 标志
	FIN bool

	// RSV 标志
	RSV1, RSV2, RSV3 bool

	// Opcode
	Opcode uint8

	// MASK 标志
	MASK bool

	// Payload 长度
	PayloadLength uint64

	// 掩码密钥（客户端发送的帧）
	MaskingKey [4]byte

	// Payload 数据
	Payload []byte
}

// ParseFrame 解析 WebSocket 帧
func ParseFrame(data []byte) (*Frame, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("frame too short")
	}

	frame := &Frame{}

	// 第一个字节：FIN, RSV, Opcode
	frame.FIN = data[0]&0x80 != 0
	frame.RSV1 = data[0]&0x40 != 0
	frame.RSV2 = data[0]&0x20 != 0
	frame.RSV3 = data[0]&0x10 != 0
	frame.Opcode = data[0] & 0x0f

	// 第二个字节：MASK, Payload length
	frame.MASK = data[1]&0x80 != 0
	length := uint64(data[1] & 0x7f)

	offset := 2

	// 扩展长度
	if length == 126 {
		if len(data) < 4 {
			return nil, fmt.Errorf("frame truncated at extended length")
		}
		length = uint64(binary.BigEndian.Uint16(data[2:4]))
		offset = 4
	} else if length == 127 {
		if len(data) < 10 {
			return nil, fmt.Errorf("frame truncated at extended length")
		}
		length = binary.BigEndian.Uint64(data[2:10])
		offset = 10
	}

	frame.PayloadLength = length

	// 掩码密钥
	if frame.MASK {
		if len(data) < offset+4 {
			return nil, fmt.Errorf("frame truncated at masking key")
		}
		copy(frame.MaskingKey[:], data[offset:offset+4])
		offset += 4
	}

	// Payload
	if uint64(len(data)-offset) < length {
		return nil, fmt.Errorf("frame truncated at payload")
	}

	frame.Payload = data[offset : offset+int(length)]

	// 解掩码
	if frame.MASK {
		frame.Payload = unmaskPayload(frame.Payload, frame.MaskingKey)
	}

	return frame, nil
}

// unmaskPayload 解掩码 payload
func unmaskPayload(payload []byte, key [4]byte) []byte {
	result := make([]byte, len(payload))
	for i, b := range payload {
		result[i] = b ^ key[i%4]
	}
	return result
}

// FrameOpCode 帧操作码常量
const (
	OpCodeContinuation uint8 = 0x0
	OpCodeText         uint8 = 0x1
	OpCodeBinary       uint8 = 0x2
	OpCodeClose        uint8 = 0x8
	OpCodePing         uint8 = 0x9
	OpCodePong         uint8 = 0xa
)

// AnalyzeFrame 分析单个帧的特征
func AnalyzeFrame(frame *Frame) map[string]interface{} {
	features := make(map[string]interface{})

	features["opcode"] = frame.Opcode
	features["fin"] = frame.FIN
	features["mask"] = frame.MASK
	features["payload_length"] = frame.PayloadLength
	features["has_rsv"] = frame.RSV1 || frame.RSV2 || frame.RSV3

	// 帧类型
	switch frame.Opcode {
	case OpCodeText:
		features["frame_type"] = "text"
	case OpCodeBinary:
		features["frame_type"] = "binary"
	case OpCodeClose:
		features["frame_type"] = "close"
	case OpCodePing:
		features["frame_type"] = "ping"
	case OpCodePong:
		features["frame_type"] = "pong"
	case OpCodeContinuation:
		features["frame_type"] = "continuation"
	default:
		features["frame_type"] = "unknown"
	}

	return features
}

// AnalyzeFrameStream 分析帧流特征
func AnalyzeFrameStream(frames []*Frame) map[string]interface{} {
	features := make(map[string]interface{})

	if len(frames) == 0 {
		return features
	}

	// 统计
	var (
		maskedCount   int
		unmaskedCount int
		textCount     int
		binaryCount   int
		controlCount  int
		maxPayload    uint64
		totalPayload  uint64
	)

	for _, frame := range frames {
		if frame.MASK {
			maskedCount++
		} else {
			unmaskedCount++
		}

		switch frame.Opcode {
		case OpCodeText:
			textCount++
		case OpCodeBinary:
			binaryCount++
		case OpCodeClose, OpCodePing, OpCodePong:
			controlCount++
		}

		if frame.PayloadLength > maxPayload {
			maxPayload = frame.PayloadLength
		}
		totalPayload += frame.PayloadLength
	}

	features["total_frames"] = len(frames)
	features["masked_ratio"] = float64(maskedCount) / float64(len(frames))
	features["text_ratio"] = float64(textCount) / float64(len(frames))
	features["binary_ratio"] = float64(binaryCount) / float64(len(frames))
	features["control_ratio"] = float64(controlCount) / float64(len(frames))
	features["max_payload"] = maxPayload
	features["avg_payload"] = float64(totalPayload) / float64(len(frames))

	return features
}
