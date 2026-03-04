package websocket

import (
	"net/http"
	"strings"
	"testing"
)

// TestNewDetector 测试创建检测器
func TestNewDetector(t *testing.T) {
	d := NewDetector()
	if d == nil {
		t.Fatal("NewDetector() returned nil")
	}

	if len(d.botPatterns) == 0 {
		t.Error("botPatterns should not be empty")
	}

	if len(d.suspiciousKeyPatterns) == 0 {
		t.Error("suspiciousKeyPatterns should not be empty")
	}

	if len(d.normalHeaderOrders) == 0 {
		t.Error("normalHeaderOrders should not be empty")
	}
}

// TestDetector_Detect_ValidRequest 测试有效请求
func TestDetector_Detect_ValidRequest(t *testing.T) {
	d := NewDetector()

	req := createValidWebSocketRequest()
	fp := createValidFingerprint()

	result := d.Detect(req, fp)

	if result == nil {
		t.Fatal("Detect() returned nil")
	}

	// 有效请求应该没有异常或只有低级别异常
	if result.RiskScore > 30 {
		t.Errorf("Valid request should have low risk score, got %d", result.RiskScore)
	}
}

// TestDetector_Detect_InvalidMethod 测试无效方法
func TestDetector_Detect_InvalidMethod(t *testing.T) {
	d := NewDetector()

	req := createValidWebSocketRequest()
	req.Method = "POST"
	fp := createValidFingerprint()

	result := d.Detect(req, fp)

	hasInvalidMethod := false
	for _, anomaly := range result.Anomalies {
		if anomaly.Type == AnomalyInvalidMethod {
			hasInvalidMethod = true
			if anomaly.Severity != SeverityHigh {
				t.Error("Invalid method should have HIGH severity")
			}
		}
	}

	if !hasInvalidMethod {
		t.Error("Should detect invalid method")
	}
}

// TestDetector_Detect_MissingHeaders 测试缺少头部
func TestDetector_Detect_MissingHeaders(t *testing.T) {
	d := NewDetector()

	req := createValidWebSocketRequest()
	req.Header.Del("Sec-Websocket-Key")
	fp := createValidFingerprint()

	result := d.Detect(req, fp)

	hasMissingHeaders := false
	for _, anomaly := range result.Anomalies {
		if anomaly.Type == AnomalyMissingHeaders {
			hasMissingHeaders = true
		}
	}

	if !hasMissingHeaders {
		t.Error("Should detect missing headers")
	}
}

// TestDetector_Detect_LowEntropyKey 测试低熵 Key
func TestDetector_Detect_LowEntropyKey(t *testing.T) {
	d := NewDetector()

	req := createValidWebSocketRequest()
	fp := createValidFingerprint()

	// 设置低熵 Key
	fp.Handshake.SecWebSocketKeyCharacteristics.Entropy = 2.0
	fp.Handshake.SecWebSocketKeyCharacteristics.HasPattern = true
	fp.Handshake.SecWebSocketKeyCharacteristics.PatternType = "low_entropy"

	result := d.Detect(req, fp)

	hasLowEntropy := false
	for _, anomaly := range result.Anomalies {
		if anomaly.Type == AnomalyLowEntropyKey {
			hasLowEntropy = true
		}
	}

	if !hasLowEntropy {
		t.Error("Should detect low entropy key")
	}
}

// TestDetector_Detect_KnownBot 测试已知机器人检测
func TestDetector_Detect_KnownBot(t *testing.T) {
	d := NewDetector()

	req := createValidWebSocketRequest()
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Bot/1.0)")
	fp := createValidFingerprint()
	fp.Handshake.UserAgent = "Mozilla/5.0 (compatible; Bot/1.0)"

	result := d.Detect(req, fp)

	if !result.IsKnownBot {
		t.Error("Should detect known bot")
	}

	hasBotSignature := false
	for _, anomaly := range result.Anomalies {
		if anomaly.Type == AnomalyKnownBotSignature {
			hasBotSignature = true
		}
	}

	if !hasBotSignature {
		t.Error("Should have known bot signature anomaly")
	}
}

// TestDetector_DetectFrameAnomalies 测试帧异常检测
func TestDetector_DetectFrameAnomalies(t *testing.T) {
	d := NewDetector()

	// 测试 RSV 位设置
	frame := &Frame{
		FIN:     true,
		RSV1:    true,
		Opcode:  OpCodeText,
		MASK:    true,
		Payload: []byte("test"),
	}

	result := d.DetectFrameAnomalies(frame)

	hasFrameAnomaly := false
	for _, anomaly := range result.Anomalies {
		if anomaly.Type == AnomalyFrameAnomaly && strings.Contains(anomaly.Description, "RSV") {
			hasFrameAnomaly = true
		}
	}

	if !hasFrameAnomaly {
		t.Error("Should detect RSV bits anomaly")
	}

	// 测试控制帧过大
	controlFrame := &Frame{
		FIN:           true,
		Opcode:        OpCodeClose,
		MASK:          true,
		PayloadLength: 200, // 超过 125 字节
	}

	result = d.DetectFrameAnomalies(controlFrame)

	hasControlFrameAnomaly := false
	for _, anomaly := range result.Anomalies {
		if anomaly.Type == AnomalyFrameAnomaly && strings.Contains(anomaly.Description, "Control frame") {
			hasControlFrameAnomaly = true
		}
	}

	if !hasControlFrameAnomaly {
		t.Error("Should detect control frame size anomaly")
	}
}

// TestDetector_CalculateRiskScore 测试风险评分计算
func TestDetector_CalculateRiskScore(t *testing.T) {
	d := NewDetector()

	// 空结果
	emptyResult := &DetectionResult{Anomalies: []Anomaly{}}
	score := d.calculateRiskScore(emptyResult)
	if score != 0 {
		t.Errorf("Empty result should have score 0, got %d", score)
	}

	// 多个异常
	result := &DetectionResult{
		Anomalies: []Anomaly{
			{Severity: SeverityLow},    // 15
			{Severity: SeverityMedium}, // 30
			{Severity: SeverityHigh},   // 50
		},
	}
	score = d.calculateRiskScore(result)
	expectedScore := 15 + 30 + 50 // 95
	if score != expectedScore {
		t.Errorf("Expected score %d, got %d", expectedScore, score)
	}

	// 已知机器人
	result.IsKnownBot = true
	score = d.calculateRiskScore(result)
	expectedScore = 95 + 50 // 145 -> capped at 100
	if score != 100 {
		t.Errorf("Score should be capped at 100, got %d", score)
	}
}

// TestGetSeverityWeight 测试严重程度权重
func TestGetSeverityWeight(t *testing.T) {
	tests := []struct {
		severity Severity
		expected int
	}{
		{SeverityInfo, 1},
		{SeverityLow, 2},
		{SeverityMedium, 3},
		{SeverityHigh, 4},
		{SeverityCritical, 5},
		{"unknown", 0},
	}

	for _, tt := range tests {
		weight := GetSeverityWeight(tt.severity)
		if weight != tt.expected {
			t.Errorf("GetSeverityWeight(%s) = %d, want %d", tt.severity, weight, tt.expected)
		}
	}
}

// TestIdentifyBrowserFromUA 测试浏览器识别
func TestIdentifyBrowserFromUA(t *testing.T) {
	d := NewDetector()

	tests := []struct {
		ua       string
		expected string
	}{
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/91.0", "chrome"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0", "firefox"},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Safari/605.1", "safari"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Edg/91.0", "edge"},
		{"", ""},
		{"UnknownBot/1.0", ""},
	}

	for _, tt := range tests {
		result := d.identifyBrowserFromUA(tt.ua)
		if result != tt.expected {
			t.Errorf("identifyBrowserFromUA(%s) = %s, want %s", tt.ua, result, tt.expected)
		}
	}
}

// TestCalculateHeaderOrderMatch 测试头部顺序匹配计算
func TestCalculateHeaderOrderMatch(t *testing.T) {
	d := NewDetector()

	tests := []struct {
		name     string
		actual   []string
		normal   []string
		expected float64
	}{
		{
			name:     "完美匹配",
			actual:   []string{"Host", "Connection", "Upgrade"},
			normal:   []string{"Host", "Connection", "Upgrade"},
			expected: 1.0,
		},
		{
			name:     "部分匹配",
			actual:   []string{"Host", "Upgrade", "Connection"},
			normal:   []string{"Host", "Connection", "Upgrade"},
			expected: 0.33, // 1/3
		},
		{
			name:     "完全不匹配",
			actual:   []string{"A", "B", "C"},
			normal:   []string{"X", "Y", "Z"},
			expected: 0.0, // 修正: 实际为 0.0，因为没有任何匹配
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := d.calculateHeaderOrderMatch(tt.actual, tt.normal)
			// 允许小误差
			if score < tt.expected-0.1 || score > tt.expected+0.1 {
				t.Errorf("Expected %.2f, got %.2f", tt.expected, score)
			}
		})
	}
}

// TestMin 测试 min 函数
func TestMin(t *testing.T) {
	if min() != 0 {
		t.Error("min() with no args should return 0")
	}

	if min(5) != 5 {
		t.Error("min(5) should return 5")
	}

	if min(5, 3, 7) != 3 {
		t.Error("min(5, 3, 7) should return 3")
	}
}

// BenchmarkDetector_Detect 基准测试检测
func BenchmarkDetector_Detect(b *testing.B) {
	d := NewDetector()
	req := createValidWebSocketRequest()
	fp := createValidFingerprint()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Detect(req, fp)
	}
}

// BenchmarkIdentifyBrowserFromUA 基准测试浏览器识别
func BenchmarkIdentifyBrowserFromUA(b *testing.B) {
	d := NewDetector()
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/91.0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.identifyBrowserFromUA(ua)
	}
}

// 辅助函数

func createValidWebSocketRequest() *http.Request {
	req := &http.Request{
		Method: "GET",
		Header: make(http.Header),
	}
	req.Header.Set("Host", "example.com")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-Websocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Sec-Websocket-Version", "13")
	req.Header.Set("User-Agent", "Mozilla/5.0 Chrome/91.0")
	return req
}

func createValidFingerprint() *WebSocketFingerprint {
	return &WebSocketFingerprint{
		Version: "13",
		Handshake: WebSocketHandshake{
			HTTPVersion:         "HTTP/1.1",
			Method:              "GET",
			HeaderOrder:         []string{"Host", "Upgrade", "Connection", "Sec-Websocket-Key", "Sec-Websocket-Version"},
			UserAgent:           "Mozilla/5.0 Chrome/91.0",
			SecWebSocketVersion: "13",
			SecWebSocketKeyCharacteristics: KeyCharacteristics{
				Length:           24,
				IsStandardBase64: true,
				Entropy:          5.5,
				HasPattern:       false,
			},
		},
		Extensions:   []string{"permessage-deflate"},
		SubProtocols: []string{},
	}
}
