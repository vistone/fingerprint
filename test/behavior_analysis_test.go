package fingerprint

import (
	"strings"
	"testing"
	"time"

	fp "github.com/vistone/fingerprint"
)

// TestBehaviorAnalyzer_NewAnalyzer 测试创建分析器
func TestBehaviorAnalyzer_NewAnalyzer(t *testing.T) {
	analyzer := fp.NewBehaviorAnalyzer(nil)

	if analyzer == nil {
		t.Error("应该成功创建分析器")
	}

	if analyzer.GetRiskScore() != 0 {
		t.Error("新分析器的风险分数应为 0")
	}

	if len(analyzer.GetAllSignals()) != 0 {
		t.Error("新分析器不应有任何信号")
	}
}

// TestBehaviorAnalyzer_AddRequest 测试添加请求
func TestBehaviorAnalyzer_AddRequest(t *testing.T) {
	analyzer := fp.NewBehaviorAnalyzer(nil)

	now := time.Now()
	req := fp.RequestBehavior{
		Timestamp:         now,
		TLSVersion:        "1.3",
		CipherSuite:       "TLS_AES_256_GCM_SHA384",
		HTTPVersion:       "2",
		RequestSize:       1024,
		ResponseSize:      2048,
		Latency:           100 * time.Millisecond,
		ReusingConnection: true,
		SourceIP:          "192.168.1.1",
		DestinationIP:     "10.0.0.1",
		DestinationPort:   443,
		SNI:               "example.com",
		Extensions:        []string{"supported_groups", "key_share"},
	}

	analyzer.AddRequest(req)

	if len(analyzer.GetAllSignals()) == 0 {
		// 单个请求不会产生信号是正常的
		t.Log("单个请求不产生信号")
	}
}

// TestBehaviorAnalyzer_TemporalPattern_Regular 测试规律的时序模式
func TestBehaviorAnalyzer_TemporalPattern_Regular(t *testing.T) {
	analyzer := fp.NewBehaviorAnalyzer(&fp.BehaviorAnalysisConfig{
		MinRequestsForAnalysis:         3,
		RegularityThreshold:            0.3,
		EntropyThreshold:               0.5,
		AnomalousIntervalRateThreshold: 0.2,
	})

	now := time.Now()
	// 创建规律的请求（每 500ms 一次）
	for i := 0; i < 10; i++ {
		req := fp.RequestBehavior{
			Timestamp:         now.Add(time.Duration(i*500) * time.Millisecond),
			TLSVersion:        "1.3",
			CipherSuite:       "TLS_AES_256_GCM_SHA384",
			HTTPVersion:       "2",
			ReusingConnection: true,
			SNI:               "example.com",
		}
		analyzer.AddRequest(req)
	}

	signals := analyzer.GenerateBehaviorSignals("example.com")

	// 应该至少检测到一个信号（时序模式或连接复用）
	if len(signals) == 0 {
		t.Error("应该检测到至少一个行为信号")
	}

	// 检查是否有规律性或高连接复用的信号
	found := false
	for _, sig := range signals {
		if (sig.SignalType == "TEMPORAL_PATTERN" && sig.Score > 0.7) ||
			sig.SignalType == "CONNECTION_REUSE" {
			found = true
			t.Logf("检测到信号: %s (分数: %.2f)", sig.Description, sig.Score)
		}
	}

	if !found {
		t.Logf("检测到的信号: %v (总数: %d)", signals, len(signals))
	}
}

// TestBehaviorAnalyzer_TemporalPattern_Random 测试随机的时序模式
func TestBehaviorAnalyzer_TemporalPattern_Random(t *testing.T) {
	analyzer := fp.NewBehaviorAnalyzer(&fp.BehaviorAnalysisConfig{
		MinRequestsForAnalysis:         5,
		RegularityThreshold:            0.3,
		EntropyThreshold:               0.5,
		AnomalousIntervalRateThreshold: 0.2,
	})

	now := time.Now()
	intervals := []int64{100, 500, 150, 800, 300, 600, 250} // 随机间隔

	currentTime := now
	for _, interval := range intervals {
		req := fp.RequestBehavior{
			Timestamp:   currentTime,
			TLSVersion:  "1.3",
			CipherSuite: "TLS_AES_256_GCM_SHA384",
			HTTPVersion: "2",
			SNI:         "example.com",
		}
		analyzer.AddRequest(req)
		currentTime = currentTime.Add(time.Duration(interval) * time.Millisecond)
	}

	signals := analyzer.GenerateBehaviorSignals("example.com")

	// 随机模式应该产生较少的信号或不产生信号
	for _, sig := range signals {
		if sig.SignalType == "TEMPORAL_PATTERN" {
			// 规律性应该较低
			if sig.Score > 0.7 {
				t.Errorf("随机模式的规律性应较低, 实际分数: %.2f", sig.Score)
			}
		}
	}

	t.Logf("随机模式检测信号数: %d", len(signals))
}

// TestBehaviorAnalyzer_ProtocolProportion_Anomalous 测试异常的协议分布
func TestBehaviorAnalyzer_ProtocolProportion_Anomalous(t *testing.T) {
	analyzer := fp.NewBehaviorAnalyzer(&fp.BehaviorAnalysisConfig{
		MinRequestsForAnalysis: 3,
		EntropyThreshold:       0.5,
	})

	now := time.Now()
	// 所有请求都使用相同的 TLS 版本和 Cipher Suite（异常）
	for i := 0; i < 5; i++ {
		req := fp.RequestBehavior{
			Timestamp:         now.Add(time.Duration(i*500) * time.Millisecond),
			TLSVersion:        "1.3",                    // 全部相同
			CipherSuite:       "TLS_AES_256_GCM_SHA384", // 全部相同
			HTTPVersion:       "2",
			ReusingConnection: true,
			SNI:               "bot.example.com",
			Extensions:        []string{"supported_groups", "key_share"},
		}
		analyzer.AddRequest(req)
	}

	signals := analyzer.GenerateBehaviorSignals("bot.example.com")

	found := false
	for _, sig := range signals {
		if sig.SignalType == "PROTOCOL_PROPORTION" && strings.Contains(sig.Description, "异常") {
			found = true
			if sig.RiskLevel != "high" {
				t.Errorf("异常协议分布的风险级别应为 high, 实际: %s", sig.RiskLevel)
			}
			t.Logf("检测到异常协议分布: %s", sig.Description)
		}
	}

	if !found {
		t.Error("应该检测到异常的协议分布")
	}
}

// TestBehaviorAnalyzer_ProtocolProportion_Normal 测试正常的协议分布
func TestBehaviorAnalyzer_ProtocolProportion_Normal(t *testing.T) {
	analyzer := fp.NewBehaviorAnalyzer(&fp.BehaviorAnalysisConfig{
		MinRequestsForAnalysis: 3,
		EntropyThreshold:       0.5,
	})

	now := time.Now()
	tlsVersions := []string{"1.2", "1.3", "1.3", "1.2", "1.3"}
	cipherSuites := []string{"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		"TLS_AES_256_GCM_SHA384", "TLS_CHACHA20_POLY1305_SHA256",
		"TLS_AES_128_GCM_SHA256", "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"}

	for i := 0; i < 5; i++ {
		req := fp.RequestBehavior{
			Timestamp:         now.Add(time.Duration(i*500) * time.Millisecond),
			TLSVersion:        tlsVersions[i],
			CipherSuite:       cipherSuites[i],
			HTTPVersion:       "2",
			ReusingConnection: i > 0, // 第一个请求需要新连接
			SNI:               "chrome.example.com",
			Extensions:        []string{"supported_groups", "key_share", "signature_algorithms"},
		}
		analyzer.AddRequest(req)
	}

	signals := analyzer.GenerateBehaviorSignals("chrome.example.com")

	for _, sig := range signals {
		if sig.SignalType == "PROTOCOL_PROPORTION" {
			if strings.Contains(sig.Description, "异常") {
				t.Errorf("正常分布不应标记为异常")
			}
		}
	}

	t.Logf("正常分布的信号数: %d", len(signals))
}

// TestBehaviorAnalyzer_ConnectionReuse 测试连接复用行为
func TestBehaviorAnalyzer_ConnectionReuse(t *testing.T) {
	analyzer := fp.NewBehaviorAnalyzer(&fp.BehaviorAnalysisConfig{
		MinRequestsForAnalysis: 3,
	})

	now := time.Now()
	// 极高的连接复用率（95%）
	for i := 0; i < 20; i++ {
		req := fp.RequestBehavior{
			Timestamp:         now.Add(time.Duration(i*100) * time.Millisecond),
			TLSVersion:        "1.3",
			CipherSuite:       "TLS_AES_256_GCM_SHA384",
			HTTPVersion:       "2",
			ReusingConnection: i > 0, // 仅第一个请求新建连接
			SNI:               "reuse.example.com",
		}
		analyzer.AddRequest(req)
	}

	signals := analyzer.GenerateBehaviorSignals("reuse.example.com")

	found := false
	for _, sig := range signals {
		if sig.SignalType == "CONNECTION_REUSE" && strings.Contains(sig.Description, "连接复用") {
			found = true
			if sig.RiskLevel != "high" {
				t.Errorf("高连接复用率的风险级别应为 high, 实际: %s", sig.RiskLevel)
			}
			t.Logf("检测到高连接复用: %s", sig.Description)
		}
	}

	if !found {
		t.Error("应该检测到高连接复用行为")
	}
}

// TestBehaviorAnalyzer_RiskScore 测试综合风险分数
func TestBehaviorAnalyzer_RiskScore(t *testing.T) {
	analyzer := fp.NewBehaviorAnalyzer(&fp.BehaviorAnalysisConfig{
		MinRequestsForAnalysis: 3,
	})

	now := time.Now()
	// 创建具有多种异常行为的请求序列
	for i := 0; i < 10; i++ {
		req := fp.RequestBehavior{
			Timestamp:         now.Add(time.Duration(i*500) * time.Millisecond),
			TLSVersion:        "1.3",
			CipherSuite:       "TLS_AES_256_GCM_SHA384",
			HTTPVersion:       "2",
			ReusingConnection: true, // 极高的连接复用
			SNI:               "suspicious.example.com",
		}
		analyzer.AddRequest(req)
	}

	analyzer.GenerateBehaviorSignals("suspicious.example.com")

	riskScore := analyzer.GetRiskScore()

	if riskScore < 0 || riskScore > 1 {
		t.Errorf("风险分数应在 [0, 1] 范围内, 实际: %.2f", riskScore)
	}

	if riskScore > 0.5 && len(analyzer.GetAllSignals()) == 0 {
		t.Error("有高风险分数但没有信号检测（逻辑错误）")
	}

	t.Logf("综合风险分数: %.2f", riskScore)
}

// TestBehaviorAnalyzer_GetAnalysisSummary 测试分析报告
func TestBehaviorAnalyzer_GetAnalysisSummary(t *testing.T) {
	analyzer := fp.NewBehaviorAnalyzer(nil)

	now := time.Now()
	// 添加一个异常请求序列
	for i := 0; i < 5; i++ {
		req := fp.RequestBehavior{
			Timestamp:         now.Add(time.Duration(i*500) * time.Millisecond),
			TLSVersion:        "1.3",
			CipherSuite:       "TLS_AES_256_GCM_SHA384",
			HTTPVersion:       "2",
			ReusingConnection: true,
			SNI:               "report.example.com",
		}
		analyzer.AddRequest(req)
	}

	analyzer.GenerateBehaviorSignals("report.example.com")
	summary := analyzer.GetAnalysisSummary()

	if len(summary) == 0 {
		t.Error("分析报告不应为空")
	}

	if !strings.Contains(summary, "行为分析报告") {
		t.Error("报告应包含标题")
	}

	if !strings.Contains(summary, "总请求数") {
		t.Error("报告应包含请求数统计")
	}

	t.Logf("分析报告:\n%s", summary)
}

// TestBehaviorAnalyzer_MinimalData 测试最少数据场景
func TestBehaviorAnalyzer_MinimalData(t *testing.T) {
	config := &fp.BehaviorAnalysisConfig{
		MinRequestsForAnalysis: 10, // 要求至少 10 个请求
	}
	analyzer := fp.NewBehaviorAnalyzer(config)

	now := time.Now()
	// 仅添加 5 个请求
	for i := 0; i < 5; i++ {
		req := fp.RequestBehavior{
			Timestamp:   now.Add(time.Duration(i*500) * time.Millisecond),
			TLSVersion:  "1.3",
			CipherSuite: "TLS_AES_256_GCM_SHA384",
			HTTPVersion: "2",
			SNI:         "minimal.example.com",
		}
		analyzer.AddRequest(req)
	}

	signals := analyzer.GenerateBehaviorSignals("minimal.example.com")

	// 数据不足，不应产生信号
	if len(signals) > 0 {
		t.Error("数据不足时不应产生信号")
	}
}

// BenchmarkBehaviorAnalyzer_AddRequest 基准测试：添加请求
func BenchmarkBehaviorAnalyzer_AddRequest(b *testing.B) {
	analyzer := fp.NewBehaviorAnalyzer(nil)
	req := fp.RequestBehavior{
		Timestamp:   time.Now(),
		TLSVersion:  "1.3",
		CipherSuite: "TLS_AES_256_GCM_SHA384",
		HTTPVersion: "2",
	}

	for i := 0; i < b.N; i++ {
		analyzer.AddRequest(req)
	}
}

// BenchmarkBehaviorAnalyzer_AnalyzeTemporalPattern 基准测试：时序分析
func BenchmarkBehaviorAnalyzer_AnalyzeTemporalPattern(b *testing.B) {
	analyzer := fp.NewBehaviorAnalyzer(nil)

	now := time.Now()
	for i := 0; i < 100; i++ {
		req := fp.RequestBehavior{
			Timestamp:   now.Add(time.Duration(i*500) * time.Millisecond),
			TLSVersion:  "1.3",
			CipherSuite: "TLS_AES_256_GCM_SHA384",
			HTTPVersion: "2",
			SNI:         "example.com",
		}
		analyzer.AddRequest(req)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeTemporalPattern("example.com")
	}
}

// BenchmarkBehaviorAnalyzer_AnalyzeProtocolProportion 基准测试：协议分析
func BenchmarkBehaviorAnalyzer_AnalyzeProtocolProportion(b *testing.B) {
	analyzer := fp.NewBehaviorAnalyzer(nil)

	now := time.Now()
	tlsVersions := []string{"1.2", "1.3"}
	cipherSuites := []string{"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		"TLS_AES_256_GCM_SHA384", "TLS_CHACHA20_POLY1305_SHA256"}

	for i := 0; i < 100; i++ {
		req := fp.RequestBehavior{
			Timestamp:         now.Add(time.Duration(i*500) * time.Millisecond),
			TLSVersion:        tlsVersions[i%len(tlsVersions)],
			CipherSuite:       cipherSuites[i%len(cipherSuites)],
			HTTPVersion:       "2",
			ReusingConnection: i > 0,
			SNI:               "example.com",
			Extensions:        []string{"supported_groups", "key_share"},
		}
		analyzer.AddRequest(req)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeProtocolProportion("example.com")
	}
}

// BenchmarkBehaviorAnalyzer_GenerateBehaviorSignals 基准测试：信号生成
func BenchmarkBehaviorAnalyzer_GenerateBehaviorSignals(b *testing.B) {
	analyzer := fp.NewBehaviorAnalyzer(nil)

	now := time.Now()
	for i := 0; i < 100; i++ {
		req := fp.RequestBehavior{
			Timestamp:         now.Add(time.Duration(i*500) * time.Millisecond),
			TLSVersion:        "1.3",
			CipherSuite:       "TLS_AES_256_GCM_SHA384",
			HTTPVersion:       "2",
			ReusingConnection: true,
			SNI:               "example.com",
		}
		analyzer.AddRequest(req)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.GenerateBehaviorSignals("example.com")
	}
}
