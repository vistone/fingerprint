package main

import (
	"fmt"
	"time"

	fp "github.com/vistone/fingerprint"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("行为信号分析演示")
	fmt.Println("========================================")

	// 演示 1: 真实浏览器行为
	demonstrateRealBrowserBehavior()

	// 演示 2: 机器人行为
	demonstrateBotBehavior()

	// 演示 3: 可疑混合行为
	demonstrateSuspiciousBehavior()

	// 演示 4: 协议分布分析
	demonstrateProtocolAnalysis()
}

// 演示 1: 真实浏览器行为（随机间隔，多样的协议）
func demonstrateRealBrowserBehavior() {
	fmt.Println("=== 演示 1: 真实浏览器行为 ===")

	analyzer := fp.NewBehaviorAnalyzer(nil)

	now := time.Now()
	// 模拟真实用户的随机请求间隔
	intervals := []int{523, 250, 1200, 180, 890, 450, 1100, 320}
	tlsVersions := []string{"1.3", "1.2", "1.3", "1.3", "1.2", "1.3", "1.3", "1.2"}
	cipherSuites := []string{
		"TLS_AES_256_GCM_SHA384",
		"TLS_CHACHA20_POLY1305_SHA256",
		"TLS_AES_128_GCM_SHA256",
		"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		"TLS_AES_256_GCM_SHA384",
		"TLS_CHACHA20_POLY1305_SHA256",
		"TLS_AES_128_GCM_SHA256",
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
	}

	currentTime := now
	for i := 0; i < len(intervals); i++ {
		req := fp.RequestBehavior{
			Timestamp:         currentTime,
			TLSVersion:        tlsVersions[i%len(tlsVersions)],
			CipherSuite:       cipherSuites[i%len(cipherSuites)],
			HTTPVersion:       "2",
			ReusingConnection: i > 0 && i%3 != 0, // 偶尔创建新连接
			SNI:               "chrome.example.com",
			Extensions: []string{
				"supported_groups", "key_share", "signature_algorithms",
				"ec_point_formats", "session_ticket",
			},
		}
		analyzer.AddRequest(req)
		currentTime = currentTime.Add(time.Duration(intervals[i]) * time.Millisecond)
	}

	signals := analyzer.GenerateBehaviorSignals("chrome.example.com")
	riskScore := analyzer.GetRiskScore()

	fmt.Printf("✅ 真实浏览器行为分析\n")
	fmt.Printf("请求数: %d\n", 8)
	fmt.Printf("检测信号数: %d\n", len(signals))
	fmt.Printf("综合风险分数: %.2f (安全: < 0.3)\n", riskScore)

	if len(signals) > 0 {
		fmt.Println("\n检测到的信号:")
		for i, sig := range signals {
			fmt.Printf("  %d. %s\n     %s\n     风险级别: %s\n",
				i+1, sig.SignalType, sig.Description, sig.RiskLevel)
		}
	}

	// 时序模式分析
	pattern := analyzer.AnalyzeTemporalPattern("chrome.example.com")
	if pattern != nil {
		fmt.Printf("\n时序模式分析:\n")
		fmt.Printf("  平均间隔: %.0f ms\n", pattern.MeanInterval)
		fmt.Printf("  标准差: %.2f\n", pattern.StdDev)
		fmt.Printf("  规律性指数: %.2f (< 0.3 表示随机)\n", pattern.RegularityIndex)
	}

	// 协议分布分析
	proportion := analyzer.AnalyzeProtocolProportion("chrome.example.com")
	if proportion != nil {
		fmt.Printf("\n协议分布分析:\n")
		fmt.Printf("  TLS版本数: %d\n", len(proportion.TLSVersions))
		fmt.Printf("  CipherSuite数: %d\n", len(proportion.TopCipherSuites))
		fmt.Printf("  熵值: %.2f\n", proportion.EntropyScore)
		fmt.Printf("  异常分布: %v\n", proportion.IsAnomalous)
	}

	fmt.Println()
}

// 演示 2: 机器人行为（规律间隔，单一协议）
func demonstrateBotBehavior() {
	fmt.Println("=== 演示 2: 机器人行为 ===")

	analyzer := fp.NewBehaviorAnalyzer(nil)

	now := time.Now()
	// 模拟机器人的规律请求
	currentTime := now
	for i := 0; i < 15; i++ {
		req := fp.RequestBehavior{
			Timestamp:         currentTime,
			TLSVersion:        "1.3",                    // 始终相同
			CipherSuite:       "TLS_AES_256_GCM_SHA384", // 始终相同
			HTTPVersion:       "2",
			ReusingConnection: true, // 总是复用连接
			SNI:               "bot.example.com",
			Extensions:        []string{"supported_groups", "key_share"},
		}
		analyzer.AddRequest(req)
		currentTime = currentTime.Add(500 * time.Millisecond) // 规律的 500ms 间隔
	}

	signals := analyzer.GenerateBehaviorSignals("bot.example.com")
	riskScore := analyzer.GetRiskScore()

	fmt.Printf("⚠️  机器人行为分析\n")
	fmt.Printf("请求数: %d\n", 15)
	fmt.Printf("检测信号数: %d\n", len(signals))
	fmt.Printf("综合风险分数: %.2f (HIGH: > 0.7)\n", riskScore)

	if len(signals) > 0 {
		fmt.Println("\n检测到的异常信号:")
		for i, sig := range signals {
			fmt.Printf("  %d. [%s] %s\n     风险级别: %s (分数: %.2f)\n",
				i+1, sig.SignalType, sig.Description, sig.RiskLevel, sig.Score)
		}
	}

	// 时序模式分析
	pattern := analyzer.AnalyzeTemporalPattern("bot.example.com")
	if pattern != nil {
		fmt.Printf("\n时序模式分析:\n")
		fmt.Printf("  平均间隔: %.0f ms\n", pattern.MeanInterval)
		fmt.Printf("  标准差: %.2f (机器人: 接近0)\n", pattern.StdDev)
		fmt.Printf("  规律性指数: %.2f (> 0.8 表示高度规律)\n", pattern.RegularityIndex)
		if pattern.RegularityIndex > 0.8 {
			fmt.Printf("  ❌ 高度规律的请求模式，可能是自动化脚本！\n")
		}
	}

	// 协议分布分析
	proportion := analyzer.AnalyzeProtocolProportion("bot.example.com")
	if proportion != nil {
		fmt.Printf("\n协议分布分析:\n")
		fmt.Printf("  TLS版本数: %d (机器人: 通常为1)\n", len(proportion.TLSVersions))
		fmt.Printf("  CipherSuite数: %d (机器人: 通常为1)\n", len(proportion.TopCipherSuites))
		fmt.Printf("  异常分布: %v\n", proportion.IsAnomalous)
		if proportion.IsAnomalous {
			fmt.Printf("  ❌ 异常的协议分布，不符合真实浏览器行为！\n")
		}
	}

	fmt.Println()
}

// 演示 3: 可疑混合行为
func demonstrateSuspiciousBehavior() {
	fmt.Println("=== 演示 3: 可疑混合行为 ===")

	analyzer := fp.NewBehaviorAnalyzer(nil)

	now := time.Now()
	// 前半段：真实用户行为，后半段：机器人行为
	// 可能表示 IP 地址在多个用户/脚本间共享
	intervals := []int{600, 150, 400, 200, 500, 200, 500, 200, 500, 200}
	behaviors := []string{"user", "user", "user", "user", "user", "bot", "bot", "bot", "bot", "bot"}

	currentTime := now
	for i := 0; i < len(intervals); i++ {
		reusingConn := false
		if behaviors[i] == "bot" {
			reusingConn = true // 机器人部分复用连接
		} else {
			reusingConn = i > 0 && i%2 == 0 // 用户部分偶尔创建新连接
		}

		req := fp.RequestBehavior{
			Timestamp:         currentTime,
			TLSVersion:        "1.3",
			CipherSuite:       "TLS_AES_256_GCM_SHA384",
			HTTPVersion:       "2",
			ReusingConnection: reusingConn,
			SNI:               "suspicious.example.com",
		}
		analyzer.AddRequest(req)
		currentTime = currentTime.Add(time.Duration(intervals[i]) * time.Millisecond)
	}

	signals := analyzer.GenerateBehaviorSignals("suspicious.example.com")
	riskScore := analyzer.GetRiskScore()

	fmt.Printf("❓ 可疑混合行为分析\n")
	fmt.Printf("请求数: %d (前5个: 真实用户, 后5个: 机器人)\n", 10)
	fmt.Printf("检测信号数: %d\n", len(signals))
	fmt.Printf("综合风险分数: %.2f (中等风险: 0.4-0.7)\n", riskScore)

	if len(signals) > 0 {
		fmt.Println("\n检测到的信号:")
		for i, sig := range signals {
			fmt.Printf("  %d. %s\n     %s\n",
				i+1, sig.SignalType, sig.Description)
		}
	}

	fmt.Println("\n⚠️  分析建议:")
	fmt.Println("  - 该 IP 显示混合行为，可能同时有真实用户和自动化脚本")
	fmt.Println("  - 建议进行更深入的行为分析并结合其他指纹识别方法")
	fmt.Println("  - 考虑使用时间窗口分析来识别行为转变点")

	fmt.Println()
}

// 演示 4: 协议分布分析
func demonstrateProtocolAnalysis() {
	fmt.Println("=== 演示 4: 协议分布分析 ===")

	// Firefox 风格的请求
	firefoxAnalyzer := fp.NewBehaviorAnalyzer(nil)
	firefoxTLS := []string{"1.2", "1.3", "1.3", "1.2", "1.3"}
	firefoxCS := []string{
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		"TLS_AES_256_GCM_SHA384",
		"TLS_CHACHA20_POLY1305_SHA256",
		"TLS_AES_128_GCM_SHA256",
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
	}

	// Safari 风格的请求
	safariAnalyzer := fp.NewBehaviorAnalyzer(nil)
	safariTLS := []string{"1.3", "1.3", "1.3", "1.3", "1.3"}
	safariCS := []string{
		"TLS_AES_128_GCM_SHA256",
		"TLS_CHACHA20_POLY1305_SHA256",
		"TLS_AES_256_GCM_SHA384",
		"TLS_AES_128_GCM_SHA256",
		"TLS_CHACHA20_POLY1305_SHA256",
	}

	now := time.Now()

	// 添加 Firefox 请求
	for i := 0; i < 5; i++ {
		firefoxAnalyzer.AddRequest(fp.RequestBehavior{
			Timestamp:   now.Add(time.Duration(i*500) * time.Millisecond),
			TLSVersion:  firefoxTLS[i],
			CipherSuite: firefoxCS[i],
			HTTPVersion: "2",
			SNI:         "firefox.example.com",
		})
	}

	// 添加 Safari 请求
	for i := 0; i < 5; i++ {
		safariAnalyzer.AddRequest(fp.RequestBehavior{
			Timestamp:   now.Add(time.Duration(i*500) * time.Millisecond),
			TLSVersion:  safariTLS[i],
			CipherSuite: safariCS[i],
			HTTPVersion: "2",
			SNI:         "safari.example.com",
		})
	}

	// 分析 Firefox
	firefoxProp := firefoxAnalyzer.AnalyzeProtocolProportion("firefox.example.com")
	fmt.Println("📊 Firefox 浏览器协议分布:")
	fmt.Printf("  TLS 版本数: %d\n", len(firefoxProp.TLSVersions))
	fmt.Printf("  TLS 版本分布: %v\n", firefoxProp.TLSVersions)
	fmt.Printf("  CipherSuite 数: %d\n", len(firefoxProp.TopCipherSuites))
	fmt.Printf("  熵值: %.2f\n", firefoxProp.EntropyScore)
	fmt.Printf("  异常: %v\n", firefoxProp.IsAnomalous)

	fmt.Println()

	// 分析 Safari
	safariProp := safariAnalyzer.AnalyzeProtocolProportion("safari.example.com")
	fmt.Println("📊 Safari 浏览器协议分布:")
	fmt.Printf("  TLS 版本数: %d\n", len(safariProp.TLSVersions))
	fmt.Printf("  TLS 版本分布: %v\n", safariProp.TLSVersions)
	fmt.Printf("  CipherSuite 数: %d\n", len(safariProp.TopCipherSuites))
	fmt.Printf("  熵值: %.2f\n", safariProp.EntropyScore)
	fmt.Printf("  异常: %v\n", safariProp.IsAnomalous)

	fmt.Println()
	fmt.Println("💡 分析结果:")
	fmt.Println("  - Firefox 显示更多样的 TLS 版本使用")
	fmt.Println("  - Safari 倾向于更新的 TLS 1.3")
	fmt.Println("  - 协议分布可用于补充传统的 User-Agent 识别")
}
