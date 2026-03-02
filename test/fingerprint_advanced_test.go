package fingerprint_test

import (
	"testing"

	"github.com/vistone/fingerprint"
	"github.com/vistone/fingerprint/internal/features"
)

// TestJA4S_BasicFunctionality 测试 JA4S 基本功能
func TestJA4S_BasicFunctionality(t *testing.T) {
	analyzer := fingerprint.NewJA4SAnalyzer()

	// 测试标准 TLS 1.3 配置
	tlsVersion := uint16(0x0304)  // TLS 1.3
	cipherSuite := uint16(0x1302) // TLS_AES_256_GCM_SHA384
	extensions := []uint16{0, 10, 11, 16, 23, 35, 43}

	result, err := analyzer.GenerateServerHelloSignature(tlsVersion, cipherSuite, extensions, 0)
	if err != nil {
		t.Fatalf("JA4S generation failed: %v", err)
	}

	// 验证哈希生成
	if len(result.Hash) != 64 { // SHA256 = 64 hex chars
		t.Errorf("Expected hash length 64, got %d", len(result.Hash))
	}

	// 验证 JA4S_a 格式：TLS版本,密码套件,扩展数,压缩
	if result.JA4Sa == "" {
		t.Error("JA4Sa should not be empty")
	}

	// 标准 TLS 1.3 配置不应有异常
	if result.RiskScore > 0.1 {
		t.Errorf("Standard TLS 1.3 should have low risk, got %.2f", result.RiskScore)
	}
}

// TestJA4S_WeakCipher 测试弱密码检测
func TestJA4S_WeakCipher(t *testing.T) {
	analyzer := fingerprint.NewJA4SAnalyzer()

	// 使用已知的弱密码套件
	weakCipher := uint16(0x0004) // TLS_RSA_WITH_RC4_128_MD5
	result, err := analyzer.GenerateServerHelloSignature(0x0303, weakCipher, []uint16{0, 10}, 0)

	if err != nil {
		t.Fatalf("Should handle weak cipher: %v", err)
	}

	// 应该检测到弱密码异常
	hasWeakCipherFlag := false
	for _, flag := range result.AnomalyFlags {
		if flag == "WEAK_CIPHER_SUITE" {
			hasWeakCipherFlag = true
			break
		}
	}

	if !hasWeakCipherFlag {
		t.Error("Should detect weak cipher suite")
	}

	if result.RiskScore < 0.2 {
		t.Errorf("Weak cipher should have higher risk, got %.2f", result.RiskScore)
	}
}

// TestHTTP2Signature_BasicFrames 测试 HTTP/2 帧签名基本功能
func TestHTTP2Signature_BasicFrames(t *testing.T) {
	analyzer := fingerprint.NewHTTP2SignatureAnalyzer()

	frames := []fingerprint.HTTP2FrameData{
		{
			Type: "SETTINGS",
			Settings: map[string]interface{}{
				"HEADER_TABLE_SIZE":   4096,
				"INITIAL_WINDOW_SIZE": 65535,
			},
		},
		{
			Type:    "HEADERS",
			FrameID: 1,
			Headers: []string{":method", ":scheme", ":authority", ":path"},
		},
	}

	result, err := analyzer.AnalyzeHTTP2Stream(frames)
	if err != nil {
		t.Fatalf("HTTP/2 analysis failed: %v", err)
	}

	// 验证签名生成
	if len(result.Hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(result.Hash))
	}

	// 验证帧序列
	if result.FrameSequence == "" {
		t.Error("Frame sequence should not be empty")
	}

	// 标准帧序列不应有高风险
	if result.RiskScore > 0.3 {
		t.Errorf("Standard frames should have low risk, got %.2f", result.RiskScore)
	}
}

// TestHTTP2Signature_InvalidFrameSequence 测试无效帧序列检测
func TestHTTP2Signature_InvalidFrameSequence(t *testing.T) {
	analyzer := fingerprint.NewHTTP2SignatureAnalyzer()

	// 错误：不以 SETTINGS 开头
	frames := []fingerprint.HTTP2FrameData{
		{
			Type:    "HEADERS",
			FrameID: 1,
			Headers: []string{":method"},
		},
	}

	result, err := analyzer.AnalyzeHTTP2Stream(frames)
	if err != nil {
		t.Fatalf("Should handle invalid sequence: %v", err)
	}

	// 应该检测到无效序列
	hasInvalidSeq := false
	for _, flag := range result.AnomalyFlags {
		if flag == "INVALID_FRAME_SEQUENCE" {
			hasInvalidSeq = true
			break
		}
	}

	if !hasInvalidSeq {
		t.Error("Should detect invalid frame sequence")
	}
}

// TestJA4H_StandardBrowserRequest 测试标准浏览器请求
func TestJA4H_StandardBrowserRequest(t *testing.T) {
	analyzer := fingerprint.NewJA4HAnalyzer()

	request := fingerprint.HTTP2RequestData{
		Method:   "GET",
		Path:     "/api/v1/users",
		Protocol: "HTTP/2",
		Headers: []struct {
			Name  string
			Value string
		}{
			{Name: "Host", Value: "api.example.com"},
			{Name: "User-Agent", Value: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"},
			{Name: "Accept", Value: "text/html,application/xhtml+xml"},
			{Name: "Accept-Language", Value: "en-US,en;q=0.9"},
			{Name: "Accept-Encoding", Value: "gzip, deflate"},
		},
	}

	result, err := analyzer.AnalyzeHTTPRequest(request)
	if err != nil {
		t.Fatalf("JA4H analysis failed: %v", err)
	}

	// 验证哈希
	if len(result.Hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(result.Hash))
	}

	// 标准请求不应有高风险
	if result.RiskScore > 0.2 {
		t.Errorf("Standard request should have low risk, got %.2f", result.RiskScore)
	}
}

// TestJA4H_SuspiciousQueryParams 测试可疑查询参数检测
func TestJA4H_SuspiciousQueryParams(t *testing.T) {
	analyzer := fingerprint.NewJA4HAnalyzer()

	request := fingerprint.HTTP2RequestData{
		Method:   "GET",
		Path:     "/search",
		Protocol: "HTTP/2",
		Headers: []struct {
			Name  string
			Value string
		}{
			{Name: "Host", Value: "example.com"},
			{Name: "User-Agent", Value: "Mozilla/5.0"},
		},
		QueryParams: map[string]string{
			"q": "test' UNION SELECT * FROM users--",
		},
	}

	result, err := analyzer.AnalyzeHTTPRequest(request)
	if err != nil {
		t.Fatalf("Should handle suspicious params: %v", err)
	}

	// 应该检测到可疑查询参数
	hasSuspiciousParams := false
	for _, flag := range result.AnomalyFlags {
		if flag == "SUSPICIOUS_QUERY_PARAMS" {
			hasSuspiciousParams = true
			break
		}
	}

	if !hasSuspiciousParams {
		t.Error("Should detect SQL injection patterns")
	}

	if result.RiskScore < 0.2 {
		t.Errorf("SQL injection should have higher risk, got %.2f", result.RiskScore)
	}
}

// TestIntegration_CombinedFingerprints 集成测试：组合多个指纹
func TestIntegration_CombinedFingerprints(t *testing.T) {
	// 1. 创建分析器
	ja4sAnalyzer := fingerprint.NewJA4SAnalyzer()
	http2Analyzer := fingerprint.NewHTTP2SignatureAnalyzer()
	ja4hAnalyzer := fingerprint.NewJA4HAnalyzer()
	featureExtractor := features.NewBaseFeatureExtractor(nil)

	// 2. 模拟真实请求数据
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

	// 3. 特征提取 (P0)
	featureVector := featureExtractor.ExtractFeatureVector(map[string]interface{}{
		"user_agent": userAgent,
	}, nil)

	if featureVector == nil {
		t.Fatal("Feature extraction failed")
	}

	// 4. JA4S 分析
	ja4sResult, err := ja4sAnalyzer.GenerateServerHelloSignature(
		0x0304, 0x1302, []uint16{0, 10, 11, 16, 23, 35}, 0)
	if err != nil {
		t.Fatalf("JA4S failed: %v", err)
	}

	// 5. HTTP/2 分析
	http2Result, err := http2Analyzer.AnalyzeHTTP2Stream([]fingerprint.HTTP2FrameData{
		{Type: "SETTINGS", Settings: map[string]interface{}{"INITIAL_WINDOW_SIZE": 65535}},
	})
	if err != nil {
		t.Fatalf("HTTP/2 failed: %v", err)
	}

	// 6. JA4H 分析
	ja4hResult, err := ja4hAnalyzer.AnalyzeHTTPRequest(fingerprint.HTTP2RequestData{
		Method:   "GET",
		Path:     "/",
		Protocol: "HTTP/2",
		Headers: []struct {
			Name  string
			Value string
		}{
			{Name: "User-Agent", Value: userAgent},
		},
	})
	if err != nil {
		t.Fatalf("JA4H failed: %v", err)
	}

	// 7. 计算综合风险评分（简单加权）
	featureRisk := featureVector.RiskScore
	fingerprintRisk := (ja4sResult.RiskScore + http2Result.RiskScore + ja4hResult.RiskScore) / 3.0
	overallRisk := 0.3*featureRisk + 0.4*fingerprintRisk

	// 8. 验证结果
	t.Logf("Feature Risk: %.2f", featureRisk)
	t.Logf("Fingerprint Risk: %.2f", fingerprintRisk)
	t.Logf("Overall Risk: %.2f", overallRisk)

	if overallRisk > 1.0 {
		t.Errorf("Risk score should be <= 1.0, got %.2f", overallRisk)
	}

	// 9. 收集所有异常
	totalAnomalies := len(featureVector.Anomalies) +
		len(ja4sResult.AnomalyFlags) +
		len(http2Result.AnomalyFlags) +
		len(ja4hResult.AnomalyFlags)

	t.Logf("Total anomalies detected: %d", totalAnomalies)

	// 标准请求不应有太多异常
	if totalAnomalies > 5 {
		t.Errorf("Standard request should have few anomalies, got %d", totalAnomalies)
	}
}

// TestIntegration_HighRiskScenario 集成测试：高风险场景
func TestIntegration_HighRiskScenario(t *testing.T) {
	featureExtractor := features.NewBaseFeatureExtractor(nil)
	ja4hAnalyzer := fingerprint.NewJA4HAnalyzer()

	// 模拟高风险请求：HeadlessChrome + SQL 注入
	userAgent := "HeadlessChrome/120.0"

	featureVector := featureExtractor.ExtractFeatureVector(map[string]interface{}{
		"user_agent": userAgent,
	}, nil)

	ja4hResult, err := ja4hAnalyzer.AnalyzeHTTPRequest(fingerprint.HTTP2RequestData{
		Method:   "GET",
		Path:     "/admin",
		Protocol: "HTTP/1.1",
		Headers: []struct {
			Name  string
			Value string
		}{
			{Name: "User-Agent", Value: userAgent},
		},
		QueryParams: map[string]string{
			"id": "1' OR '1'='1",
		},
	})

	if err != nil {
		t.Fatalf("High risk analysis failed: %v", err)
	}

	// 应该检测到多个异常
	totalAnomalies := len(featureVector.Anomalies) + len(ja4hResult.AnomalyFlags)
	t.Logf("High risk anomalies: %d", totalAnomalies)

	if totalAnomalies == 0 {
		t.Error("Should detect anomalies in high risk scenario")
	}

	// 综合评分应该高于正常请求
	combinedRisk := 0.3*featureVector.RiskScore + 0.4*ja4hResult.RiskScore
	t.Logf("Combined risk score: %.2f", combinedRisk)

	if combinedRisk < 0.1 {
		t.Errorf("High risk scenario should have elevated score, got %.2f", combinedRisk)
	}
}

// BenchmarkJA4S_Generation JA4S 性能基准测试
func BenchmarkJA4S_Generation(b *testing.B) {
	analyzer := fingerprint.NewJA4SAnalyzer()
	tlsVersion := uint16(0x0304)
	cipherSuite := uint16(0x1302)
	extensions := []uint16{0, 10, 11, 16, 23, 35}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = analyzer.GenerateServerHelloSignature(tlsVersion, cipherSuite, extensions, 0)
	}
}

// BenchmarkHTTP2Signature_Analysis HTTP/2 签名性能基准测试
func BenchmarkHTTP2Signature_Analysis(b *testing.B) {
	analyzer := fingerprint.NewHTTP2SignatureAnalyzer()
	frames := []fingerprint.HTTP2FrameData{
		{Type: "SETTINGS", Settings: map[string]interface{}{"INITIAL_WINDOW_SIZE": 65535}},
		{Type: "HEADERS", FrameID: 1, Headers: []string{":method", ":scheme"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = analyzer.AnalyzeHTTP2Stream(frames)
	}
}

// BenchmarkJA4H_Analysis JA4H 性能基准测试
func BenchmarkJA4H_Analysis(b *testing.B) {
	analyzer := fingerprint.NewJA4HAnalyzer()
	request := fingerprint.HTTP2RequestData{
		Method:   "GET",
		Path:     "/api/test",
		Protocol: "HTTP/2",
		Headers: []struct {
			Name  string
			Value string
		}{
			{Name: "Host", Value: "example.com"},
			{Name: "User-Agent", Value: "Mozilla/5.0"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = analyzer.AnalyzeHTTPRequest(request)
	}
}
