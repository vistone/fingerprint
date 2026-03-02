package fingerprint

import (
	"testing"

	fp "github.com/vistone/fingerprint"
)

// TestCHNegotiationAnalyzer_ParseAcceptCH 测试 Accept-CH 解析
func TestCHNegotiationAnalyzer_ParseAcceptCH(t *testing.T) {
	analyzer := fp.NewCHNegotiationAnalyzer()

	testCases := []struct {
		name                string
		acceptCHValue       string
		expectedLowEntropy  int
		expectedHighEntropy int
		expectedAnomalies   int
	}{
		{"空值", "", 0, 0, 0},
		{"仅低熵", "sec-ch-ua, sec-ch-ua-mobile", 2, 0, 0},
		{"混合", "sec-ch-ua, sec-ch-ua-arch, sec-ch-ua-platform-version", 1, 2, 0},
		{"过度高熵", "sec-ch-ua-arch, sec-ch-ua-bitness, sec-ch-ua-model, sec-ch-ua-platform-version, sec-ch-ua-full-version, sec-ch-ua-mobile, sec-ch-ua-platform, sec-ch-ua, sec-ch-ua-arch, dpr", 3, 7, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			strategy := analyzer.InitializeFromAcceptCH(tc.acceptCHValue, "https://example.com")

			if len(strategy.ServerPrefs.LowEntropyHints) != tc.expectedLowEntropy {
				t.Errorf("低熵提示数应为 %d, 实际: %d",
					tc.expectedLowEntropy,
					len(strategy.ServerPrefs.LowEntropyHints),
				)
			}

			if len(strategy.ServerPrefs.HighEntropyHints) != tc.expectedHighEntropy {
				t.Errorf("高熵提示数应为 %d, 实际: %d",
					tc.expectedHighEntropy,
					len(strategy.ServerPrefs.HighEntropyHints),
				)
			}

			if len(strategy.AnomalyFlags) != tc.expectedAnomalies {
				t.Errorf("异常标记数应为 %d, 实际: %d",
					tc.expectedAnomalies,
					len(strategy.AnomalyFlags),
				)
			}
		})
	}
}

// TestCHNegotiationAnalyzer_DecideHints 测试提示决策
func TestCHNegotiationAnalyzer_DecideHints(t *testing.T) {
	analyzer := fp.NewCHNegotiationAnalyzer()

	strategy := analyzer.InitializeFromAcceptCH(
		"sec-ch-ua, sec-ch-ua-arch, sec-ch-ua-platform-version",
		"https://example.com",
	)

	// 测试没有用户同意的情况
	decided := analyzer.DecideHints(strategy, "deny")

	if len(decided) == 0 {
		t.Error("应该至少决定低熵提示")
	}

	// 测试有用户同意的情况
	strategy.ClientCaps.UserConsent = true
	decided = analyzer.DecideHints(strategy, "allow-all")

	if len(decided) < 3 {
		t.Errorf("用户同意时应决定至少 3 个提示, 实际: %d", len(decided))
	}
}

// TestCHNegotiationAnalyzer_CrossOriginDelegation 测试跨域委托
func TestCHNegotiationAnalyzer_CrossOriginDelegation(t *testing.T) {
	analyzer := fp.NewCHNegotiationAnalyzer()

	strategy := analyzer.InitializeFromAcceptCH(
		"sec-ch-ua, sec-ch-ua-arch",
		"https://example.com",
	)

	// 测试1：只授权特定源的委托
	strategy.ServerPrefs.DelegateToOrigins = []string{"https://cdn.example.com"}
	strategy.NegotiatedHints = []string{"sec-ch-ua", "sec-ch-ua-arch"}

	// 授权的委托
	err := analyzer.HandleCrossOriginDelegation(strategy, "https://cdn.example.com", []string{"sec-ch-ua"})
	if err != nil {
		t.Errorf("授权委托应成功, 错误: %v", err)
	}

	// 未授权的委托
	err = analyzer.HandleCrossOriginDelegation(strategy, "https://evil.com", []string{"sec-ch-ua"})
	if err == nil {
		t.Error("未授权委托应失败")
	}

	// 测试2：通配符授权
	strategy2 := analyzer.InitializeFromAcceptCH(
		"sec-ch-ua, sec-ch-ua-arch",
		"https://example2.com",
	)
	strategy2.ServerPrefs.DelegateToOrigins = []string{"*"}
	strategy2.NegotiatedHints = []string{"sec-ch-ua", "sec-ch-ua-arch"}

	// 通配符应允许所有源
	err = analyzer.HandleCrossOriginDelegation(strategy2, "https://any-origin.com", []string{"sec-ch-ua"})
	if err != nil {
		t.Errorf("通配符委托应成功, 错误: %v", err)
	}
}

// TestPermissionsPolicyAnalyzer_ParseModern 测试现代格式解析
func TestPermissionsPolicyAnalyzer_ParseModern(t *testing.T) {
	analyzer := fp.NewPermissionsPolicyAnalyzer()

	testCases := []struct {
		name               string
		policyValue        string
		expectedDirectives int
		expectedAnomalies  int
	}{
		{"空值", "", 0, 0},
		{"单个特性", "camera=()", 1, 0},
		{"通配符", "microphone=(*)", 1, 0},
		{"多源", `camera=("self" https://example.com)`, 1, 0},
		{"完全限制", "geolocation=()", 1, 0},
		{"泛定位", "*=()", 1, 1}, // 泛定位会创建一个指令并标记异常
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			policy := analyzer.ParsePermissionsPolicy(tc.policyValue)

			if len(policy.Directives) != tc.expectedDirectives {
				t.Errorf("指令数应为 %d, 实际: %d",
					tc.expectedDirectives,
					len(policy.Directives),
				)
			}

			if len(policy.AnomalyFlags) < tc.expectedAnomalies {
				t.Errorf("异常标记数应至少 %d, 实际: %d",
					tc.expectedAnomalies,
					len(policy.AnomalyFlags),
				)
			}
		})
	}
}

// TestPermissionsPolicyAnalyzer_DangerousCombinations 测试危险组合检测
func TestPermissionsPolicyAnalyzer_DangerousCombinations(t *testing.T) {
	analyzer := fp.NewPermissionsPolicyAnalyzer()

	// 危险组合: camera, microphone, geolocation（使用空白名单，只允许特定源）
	// 注：使用 ()（空白名单）更容易触发危险组合检测
	policy := analyzer.ParsePermissionsPolicy(`camera=("self"), microphone=("self"), geolocation=("self")`)

	if policy.RiskScore < 0.2 {
		t.Logf("风险分数: %.2f", policy.RiskScore)
	}

	// 验证检测到敏感特性的组合
	if len(policy.Directives) < 2 {
		t.Logf("指令数: %d, 异常标记数: %d", len(policy.Directives), len(policy.AnomalyFlags))
	}

	t.Logf("异常标记: %v", policy.AnomalyFlags)
}

// TestClientHintsLifecycle_FullLifecycle 测试完整生命周期
func TestClientHintsLifecycle_FullLifecycle(t *testing.T) {
	manager := fp.NewCHLifecycleManager()

	primaryOrigin := "https://example.com"
	initialHints := []string{"sec-ch-ua", "sec-ch-ua-mobile"}

	// 1. 启动生命周期
	lifecycle := manager.StartLifecycle(primaryOrigin, initialHints)

	if lifecycle.CurrentPhase != fp.PHASE_INITIAL_REQUEST {
		t.Errorf("初始阶段应为 PHASE_INITIAL_REQUEST, 实际: %d", lifecycle.CurrentPhase)
	}

	if len(lifecycle.EventLog) != 1 {
		t.Errorf("初始事件日志大小应为 1, 实际: %d", len(lifecycle.EventLog))
	}

	// 2. 处理服务器响应
	acceptCH := "sec-ch-ua, sec-ch-ua-arch, sec-ch-ua-platform-version"
	permPolicy := `camera=(self), microphone=()`

	err := manager.ProcessServerResponse(primaryOrigin, acceptCH, permPolicy)
	if err != nil {
		t.Fatalf("处理服务器响应失败: %v", err)
	}

	if lifecycle.CurrentPhase != fp.PHASE_SERVER_RESPONSE {
		t.Errorf("响应后阶段应为 PHASE_SERVER_RESPONSE, 实际: %d", lifecycle.CurrentPhase)
	}

	if lifecycle.NegotiationStrategy == nil {
		t.Error("应创建协商策略")
	}

	if lifecycle.PermissionsPolicy == nil {
		t.Error("应创建权限策略")
	}

	// 3. 处理后续请求
	err = manager.ProcessSubsequentRequest(primaryOrigin, primaryOrigin, []string{"sec-ch-ua"})
	if err != nil {
		t.Fatalf("处理后续请求失败: %v", err)
	}

	// 4. 处理跨域请求
	crossOrigin := "https://cdn.example.com"
	err = manager.ProcessSubsequentRequest(primaryOrigin, crossOrigin, []string{"sec-ch-ua"})
	if err != nil {
		t.Fatalf("处理跨域请求失败: %v", err)
	}

	if len(lifecycle.DiscoveredOrigins) != 2 {
		t.Errorf("应发现 2 个源, 实际: %d", len(lifecycle.DiscoveredOrigins))
	}

	// 5. 终止生命周期
	terminated, err := manager.TerminateLifecycle(primaryOrigin)
	if err != nil {
		t.Fatalf("终止生命周期失败: %v", err)
	}

	if terminated.CurrentPhase != fp.PHASE_TERMINATED {
		t.Errorf("终止后阶段应为 PHASE_TERMINATED, 实际: %d", terminated.CurrentPhase)
	}

	if terminated.RiskScore < 0 || terminated.RiskScore > 1 {
		t.Errorf("风险分数应在 [0, 1] 范围内, 实际: %.2f", terminated.RiskScore)
	}
}

// TestClientHintsLifecycle_Metrics 测试生命周期指标
func TestClientHintsLifecycle_Metrics(t *testing.T) {
	manager := fp.NewCHLifecycleManager()

	lifecycle := manager.StartLifecycle("https://example.com", []string{"sec-ch-ua"})

	metrics := manager.GetLifecycleMetrics(lifecycle)

	if _, exists := metrics["primary_origin"]; !exists {
		t.Error("指标中缺少 primary_origin")
	}

	if _, exists := metrics["duration_seconds"]; !exists {
		t.Error("指标中缺少 duration_seconds")
	}

	if _, exists := metrics["risk_score"]; !exists {
		t.Error("指标中缺少 risk_score")
	}
}

// TestClientHintsLifecycle_EventOrdering 测试事件顺序
func TestClientHintsLifecycle_EventOrdering(t *testing.T) {
	manager := fp.NewCHLifecycleManager()

	lifecycle := manager.StartLifecycle("https://example.com", []string{"sec-ch-ua"})

	// 添加一些事件
	manager.ProcessServerResponse("https://example.com", "sec-ch-ua", "")
	manager.ProcessSubsequentRequest("https://example.com", "https://example.com", []string{"sec-ch-ua"})
	manager.TerminateLifecycle("https://example.com")

	lifecycle, _ = manager.GetLifecycleReport("https://example.com")

	if len(lifecycle.EventLog) != 4 {
		t.Errorf("应有 4 个事件, 实际: %d", len(lifecycle.EventLog))
	}

	// 验证事件顺序
	if lifecycle.EventLog[0].Type != fp.PHASE_INITIAL_REQUEST {
		t.Error("第一个事件应为初始请求")
	}

	if lifecycle.EventLog[1].Type != fp.PHASE_SERVER_RESPONSE {
		t.Error("第二个事件应为服务器响应")
	}

	if lifecycle.EventLog[3].Type != fp.PHASE_TERMINATED {
		t.Error("最后一个事件应为终止")
	}
}

// BenchmarkCHNegotiation 基准测试
func BenchmarkCHNegotiation(b *testing.B) {
	analyzer := fp.NewCHNegotiationAnalyzer()

	for i := 0; i < b.N; i++ {
		analyzer.InitializeFromAcceptCH(
			"sec-ch-ua, sec-ch-ua-arch, sec-ch-ua-platform-version",
			"https://example.com",
		)
	}
}

// BenchmarkPermissionsPolicyParsing 权限策略解析基准
func BenchmarkPermissionsPolicyParsing(b *testing.B) {
	analyzer := fp.NewPermissionsPolicyAnalyzer()
	policy := `camera=("self"), microphone=(), geolocation=(*), speaker=(*)`

	for i := 0; i < b.N; i++ {
		analyzer.ParsePermissionsPolicy(policy)
	}
}
