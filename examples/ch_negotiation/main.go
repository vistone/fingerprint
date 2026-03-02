package main

import (
	"fmt"

	fp "github.com/vistone/fingerprint"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Client Hints 协商策略演示")
	fmt.Println("========================================")

	// 演示 1: Accept-CH 协商
	demonstrateAcceptCHNegotiation()

	// 演示 2: Permissions-Policy 解析
	demonstratePermissionsPolicy()

	// 演示 3: 完整生命周期管理
	demonstrateCompleteLifecycle()

	// 演示 4: 跨域委托
	demonstrateCrossOriginDelegation()

	// 演示 5: 高风险检测
	demonstrateHighRiskDetection()
}

// 演示 1: Accept-CH 协商
func demonstrateAcceptCHNegotiation() {
	fmt.Println("=== 演示 1: Accept-CH 协商 ===")

	analyzer := fp.NewCHNegotiationAnalyzer()

	testCases := []struct {
		name        string
		acceptCH    string
		description string
	}{
		{
			"标准协商",
			"sec-ch-ua, sec-ch-ua-mobile, sec-ch-ua-platform",
			"标准浏览器提示",
		},
		{
			"扩展协商",
			"sec-ch-ua, sec-ch-ua-arch, sec-ch-ua-platform-version, sec-ch-ua-model, sec-ch-ua-bitness",
			"扩展指纹提示",
		},
		{
			"过度协商",
			"sec-ch-ua-arch, sec-ch-ua-bitness, sec-ch-ua-model, sec-ch-ua-platform-version, sec-ch-ua-full-version, sec-ch-ua-mobile, sec-ch-ua-platform, sec-ch-ua, sec-ch-ua-full-version-list, sec-ch-prefers-color-scheme",
			"疑似指纹追踪",
		},
	}

	for _, tc := range testCases {
		fmt.Printf("场景: %s\n", tc.name)
		fmt.Printf("描述: %s\n", tc.description)

		strategy := analyzer.InitializeFromAcceptCH(tc.acceptCH, "https://example.com")

		fmt.Printf("  低熵提示: %d\n", len(strategy.ServerPrefs.LowEntropyHints))
		fmt.Printf("  高熵提示: %d\n", len(strategy.ServerPrefs.HighEntropyHints))
		fmt.Printf("  风险分数: %.2f\n", strategy.RiskScore)

		if len(strategy.AnomalyFlags) > 0 {
			fmt.Println("  ⚠️ 异常标记:")
			for _, flag := range strategy.AnomalyFlags {
				fmt.Printf("    - %s\n", flag)
			}
		}

		fmt.Println()
	}
}

// 演示 2: Permissions-Policy 解析
func demonstratePermissionsPolicy() {
	fmt.Println("=== 演示 2: Permissions-Policy 解析 ===")

	analyzer := fp.NewPermissionsPolicyAnalyzer()

	testCases := []struct {
		name   string
		policy string
		desc   string
	}{
		{
			"严格策略",
			`camera=(), microphone=(), geolocation=(), payment=()`,
			"完全限制敏感特性",
		},
		{
			"中等策略",
			`camera=("self"), microphone=("self"), geolocation=()`,
			"允许自源访问，限制地理位置",
		},
		{
			"开放策略",
			`camera=(*), microphone=(*), geolocation=(*), payment=(*)`,
			"允许所有源访问敏感特性（高风险）",
		},
		{
			"跨域委托",
			`camera=("self" https://cdn.example.com), usb=("self" https://payment.example.com)`,
			"有条件地委托到特定源",
		},
	}

	for _, tc := range testCases {
		fmt.Printf("场景: %s\n", tc.name)
		fmt.Printf("描述: %s\n", tc.desc)

		policy := analyzer.ParsePermissionsPolicy(tc.policy)

		fmt.Printf("  特性数: %d\n", len(policy.Directives))

		allowAll := 0
		restricted := 0
		for _, directive := range policy.Directives {
			if directive.AllowAll {
				allowAll++
			} else if directive.HasNone {
				restricted++
			}
		}
		fmt.Printf("  允许所有: %d | 完全限制: %d\n", allowAll, restricted)
		fmt.Printf("  风险分数: %.2f\n", policy.RiskScore)

		if len(policy.AnomalyFlags) > 0 {
			fmt.Println("  ⚠️ 异常标记:")
			for _, flag := range policy.AnomalyFlags {
				fmt.Printf("    - %s\n", flag)
			}
		}

		fmt.Println()
	}
}

// 演示 3: 完整生命周期管理
func demonstrateCompleteLifecycle() {
	fmt.Println("=== 演示 3: 完整生命周期管理 ===")

	manager := fp.NewCHLifecycleManager()

	// 启动生命周期
	primaryOrigin := "https://myapp.example.com"
	initialHints := []string{"sec-ch-ua", "sec-ch-ua-mobile"}

	fmt.Printf("场景: 完整的 Client Hints 生命周期\n")
	fmt.Printf("主页面: %s\n\n", primaryOrigin)

	// 步骤 1: 初始请求
	fmt.Println("步骤 1: 客户端发送初始请求")
	lifecycle := manager.StartLifecycle(primaryOrigin, initialHints)
	fmt.Printf("  启动提示: %v\n", initialHints)
	fmt.Printf("  当前阶段: 初始请求\n\n")

	// 步骤 2: 服务器响应
	fmt.Println("步骤 2: 服务器响应 Accept-CH 和 Permissions-Policy")
	acceptCH := "sec-ch-ua, sec-ch-ua-arch, sec-ch-ua-platform-version"
	permPolicy := `camera=("self"), microphone=("self"), geolocation=()`

	manager.ProcessServerResponse(primaryOrigin, acceptCH, permPolicy)
	fmt.Printf("  服务器请求提示: %s\n", acceptCH)
	fmt.Printf("  权限策略: %s\n\n", permPolicy)

	lifecycle, _ = manager.GetLifecycleReport(primaryOrigin)
	if lifecycle.NegotiationStrategy != nil {
		fmt.Printf("  协商提示: %d\n", len(lifecycle.NegotiationStrategy.NegotiatedHints))
		fmt.Printf("  被拒提示: %d\n", len(lifecycle.NegotiationStrategy.RejectedHints))
	}

	// 步骤 3: 后续请求
	fmt.Println("\n步骤 3: 客户端在后续请求中包含 Client Hints")
	subsequentHints := []string{"sec-ch-ua", "sec-ch-ua-arch"}
	manager.ProcessSubsequentRequest(primaryOrigin, primaryOrigin, subsequentHints)
	fmt.Printf("  发送提示: %v\n", subsequentHints)
	fmt.Printf("  当前阶段: 后续请求\n\n")

	// 步骤 4: 跨域请求
	fmt.Println("步骤 4: 限制的跨域子资源请求")
	cdnOrigin := "https://cdn.example.com"
	manager.ProcessSubsequentRequest(primaryOrigin, cdnOrigin, []string{"sec-ch-ua"})
	fmt.Printf("  CDN 源: %s\n", cdnOrigin)
	fmt.Printf("  发送提示: [sec-ch-ua]\n\n")

	// 步骤 5: 终止
	fmt.Println("步骤 5: 生命周期终止")
	lifecycle, _ = manager.TerminateLifecycle(primaryOrigin)
	fmt.Printf("  %s\n", manager.GetSummary(lifecycle))

	metrics := manager.GetLifecycleMetrics(lifecycle)
	fmt.Printf("\n📊 最终指标:\n")
	fmt.Printf("  发现源数: %v\n", metrics["discovered_origins"])
	fmt.Printf("  总事件数: %v\n", metrics["total_events"])
	fmt.Printf("  风险分数: %.2f\n", metrics["risk_score"])
	fmt.Printf("  完整性标记: %v\n", metrics["integrity_flags"])
}

// 演示 4: 跨域委托
func demonstrateCrossOriginDelegation() {
	fmt.Println("\n=== 演示 4: 跨域委托 ===")

	analyzer := fp.NewCHNegotiationAnalyzer()

	testCases := []struct {
		name           string
		delegateOrigin string
		isAuthorized   bool
	}{
		{
			"授权委托",
			"https://cdn.example.com",
			true,
		},
		{
			"未授权委托",
			"https://evil.com",
			false,
		},
		{
			"通配符授权",
			"https://anything.example.com",
			true,
		},
	}

	for _, tc := range testCases {
		fmt.Printf("场景: %s\n", tc.name)

		strategy := analyzer.InitializeFromAcceptCH("sec-ch-ua, sec-ch-ua-arch", "https://example.com")
		strategy.ServerPrefs.DelegateToOrigins = []string{"https://cdn.example.com", "*.example.com"}
		strategy.NegotiatedHints = []string{"sec-ch-ua", "sec-ch-ua-arch"}

		err := analyzer.HandleCrossOriginDelegation(strategy, tc.delegateOrigin, []string{"sec-ch-ua"})

		if tc.isAuthorized {
			if err != nil {
				fmt.Printf("  ❌ 错误: %v\n", err)
			} else {
				fmt.Printf("  ✅ 委托已接受\n")
			}
		} else {
			if err != nil {
				fmt.Printf("  ✅ 正确拒绝: %v\n", err)
			} else {
				fmt.Printf("  ❌ 应该拒绝但接受了\n")
			}
		}

		if len(strategy.AnomalyFlags) > 0 {
			fmt.Println("  异常标记:")
			for _, flag := range strategy.AnomalyFlags {
				fmt.Printf("    - %s\n", flag)
			}
		}

		fmt.Println()
	}
}

// 演示 5: 高风险检测
func demonstrateHighRiskDetection() {
	fmt.Println("=== 演示 5: 高风险检测 ===")

	// 场景 A: 过度指纹追踪
	fmt.Println("📍 场景 A: 疑似指纹追踪企图")
	analyzer := fp.NewCHNegotiationAnalyzer()

	excessiveHints := "sec-ch-ua, sec-ch-ua-arch, sec-ch-ua-bitness, sec-ch-ua-model, sec-ch-ua-platform-version, sec-ch-ua-full-version, dpr, viewport-width, device-memory, downlink, ect, rtt, save-data"
	strategy := analyzer.InitializeFromAcceptCH(excessiveHints, "https://example.com")

	fmt.Printf("  请求的高熵提示: %d\n", len(strategy.ServerPrefs.HighEntropyHints))
	fmt.Printf("  风险分数: %.2f\n", strategy.RiskScore)

	if len(strategy.AnomalyFlags) > 0 {
		fmt.Println("  ⚠️ 风险指标:")
		for _, flag := range strategy.AnomalyFlags {
			fmt.Printf("    • %s\n", flag)
		}
	}

	// 场景 B: 危险的权限组合
	fmt.Println("\n📍 场景 B: 危险的权限组合")
	policyAnalyzer := fp.NewPermissionsPolicyAnalyzer()

	dangerousPolicy := `camera=(*), microphone=(*), geolocation=(*), payment=(*), usb=(*)`
	policy := policyAnalyzer.ParsePermissionsPolicy(dangerousPolicy)

	fmt.Printf("  允许所有源访问的特性: 5\n")
	fmt.Printf("  风险分数: %.2f\n", policy.RiskScore)

	if len(policy.AnomalyFlags) > 0 {
		fmt.Println("  ⚠️ 风险指标:")
		for _, flag := range policy.AnomalyFlags {
			fmt.Printf("    • %s\n", flag)
		}
	}

	// 场景 C: 异常生命周期
	fmt.Println("\n📍 场景 C: 异常生命周期")
	manager := fp.NewCHLifecycleManager()

	lifecycle := manager.StartLifecycle("https://example.com", []string{"sec-ch-ua"})

	// 添加多个跨域请求
	for i := 0; i < 12; i++ {
		manager.ProcessSubsequentRequest(
			"https://example.com",
			fmt.Sprintf("https://tracker%d.com", i),
			[]string{"sec-ch-ua"},
		)
	}

	manager.TerminateLifecycle("https://example.com")
	lifecycle, _ = manager.GetLifecycleReport("https://example.com")

	fmt.Printf("  发现的跨域源: %d\n", len(lifecycle.DiscoveredOrigins))
	fmt.Printf("  总事件数: %d\n", len(lifecycle.EventLog))
	fmt.Printf("  风险分数: %.2f\n", lifecycle.RiskScore)

	if len(lifecycle.IntegrityFlags) > 0 {
		fmt.Println("  ⚠️ 完整性标记:")
		for _, flag := range lifecycle.IntegrityFlags {
			fmt.Printf("    • %s\n", flag)
		}
	}
}
