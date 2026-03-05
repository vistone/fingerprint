package clienthints

// Phase 3: 本模块已完成基础迁移，待深度优化（详见 docs/5-process/modularization/PHASE_3_PLAN.md）
import (
	"fmt"
	"time"

	"github.com/vistone/fingerprint/modules/http/legacy/policy"
)

// CHPhase Client Hints 生命周期阶段
type CHPhase int

const (
	// PHASE_INITIAL_REQUEST 初始请求
	PHASE_INITIAL_REQUEST CHPhase = iota
	// PHASE_SERVER_RESPONSE 服务器响应
	PHASE_SERVER_RESPONSE
	// PHASE_SUBSEQUENT_REQUESTS 后续请求
	PHASE_SUBSEQUENT_REQUESTS
	// PHASE_CROSS_ORIGIN_SUB_REQUESTS 跨域子资源请求
	PHASE_CROSS_ORIGIN_SUB_REQUESTS
	// PHASE_TERMINATED 生命周期终止
	PHASE_TERMINATED
)

// CHLifecycleEvent 生命周期事件
type CHLifecycleEvent struct {
	// 时间戳
	Timestamp time.Time

	// 事件类型
	Type CHPhase

	// 源 URL
	OriginURL string

	// 相关的提示
	Hints []string

	// 事件详情
	Details map[string]interface{}

	// 风险指标
	RiskIndicators []string
}

// ClientHintsLifecycle 完整的 Client Hints 生命周期管理
type ClientHintsLifecycle struct {
	// 起始时间
	StartTime time.Time

	// 当前阶段
	CurrentPhase CHPhase

	// 主页面 URL
	PrimaryOriginURL string

	// 协商策略
	NegotiationStrategy *NegotiationStrategy

	// 权限策略
	PermissionsPolicy *policy.PermissionsPolicy

	// 事件日志
	EventLog []CHLifecycleEvent

	// 当前活跃的提示
	ActiveHints []string

	// 已探测的源
	DiscoveredOrigins []string

	// 生命周期完整性标记
	IntegrityFlags []string

	// 风险评分
	RiskScore float64
}

// CHLifecycleManager 生命周期管理器
type CHLifecycleManager struct {
	lifecycles map[string]*ClientHintsLifecycle

	negotiationAnalyzer *CHNegotiationAnalyzer
	policyAnalyzer      *policy.PermissionsPolicyAnalyzer
}

// NewCHLifecycleManager 创建生命周期管理器
func NewCHLifecycleManager() *CHLifecycleManager {
	return &CHLifecycleManager{
		lifecycles:          make(map[string]*ClientHintsLifecycle),
		negotiationAnalyzer: NewCHNegotiationAnalyzer(),
		policyAnalyzer:      policy.NewPermissionsPolicyAnalyzer(),
	}
}

// StartLifecycle 启动新的 Client Hints 生命周期
func (m *CHLifecycleManager) StartLifecycle(primaryOriginURL string, initialHints []string) *ClientHintsLifecycle {
	lifecycle := &ClientHintsLifecycle{
		StartTime:         time.Now(),
		CurrentPhase:      PHASE_INITIAL_REQUEST,
		PrimaryOriginURL:  primaryOriginURL,
		EventLog:          []CHLifecycleEvent{},
		ActiveHints:       initialHints,
		DiscoveredOrigins: []string{primaryOriginURL},
		IntegrityFlags:    []string{},
	}

	// 记录初始事件
	lifecycle.EventLog = append(lifecycle.EventLog, CHLifecycleEvent{
		Timestamp: lifecycle.StartTime,
		Type:      PHASE_INITIAL_REQUEST,
		OriginURL: primaryOriginURL,
		Hints:     initialHints,
		Details: map[string]interface{}{
			"hint_count": len(initialHints),
		},
	})

	m.lifecycles[primaryOriginURL] = lifecycle
	return lifecycle
}

// ProcessServerResponse 处理服务器的 Accept-CH 响应
func (m *CHLifecycleManager) ProcessServerResponse(primaryOriginURL string, acceptCHValue string, permissionsPolicyValue string) error {
	lifecycle, exists := m.lifecycles[primaryOriginURL]
	if !exists {
		return fmt.Errorf("lifecycle not found for %s", primaryOriginURL)
	}

	// 处理 Accept-CH
	strategy := m.negotiationAnalyzer.InitializeFromAcceptCH(acceptCHValue, primaryOriginURL)
	lifecycle.NegotiationStrategy = strategy

	// 处理 Permissions-Policy
	if permissionsPolicyValue != "" {
		policy := m.policyAnalyzer.ParsePermissionsPolicy(permissionsPolicyValue)
		lifecycle.PermissionsPolicy = policy
	}

	lifecycle.CurrentPhase = PHASE_SERVER_RESPONSE

	// 记录事件
	lifecycle.EventLog = append(lifecycle.EventLog, CHLifecycleEvent{
		Timestamp: time.Now(),
		Type:      PHASE_SERVER_RESPONSE,
		OriginURL: primaryOriginURL,
		Hints:     strategy.NegotiatedHints,
		Details: map[string]interface{}{
			"requested_hints":  len(strategy.ServerPrefs.LowEntropyHints) + len(strategy.ServerPrefs.HighEntropyHints),
			"negotiated_hints": len(strategy.NegotiatedHints),
			"rejected_hints":   len(strategy.RejectedHints),
		},
		RiskIndicators: strategy.AnomalyFlags,
	})

	return nil
}

// ProcessSubsequentRequest 处理后续请求中的 Client Hints
func (m *CHLifecycleManager) ProcessSubsequentRequest(primaryOriginURL string, requestOriginURL string, includedHints []string) error {
	lifecycle, exists := m.lifecycles[primaryOriginURL]
	if !exists {
		return fmt.Errorf("lifecycle not found for %s", primaryOriginURL)
	}

	// 判断是否为跨域请求
	isCrossOrigin := primaryOriginURL != requestOriginURL
	var phase CHPhase
	if isCrossOrigin {
		phase = PHASE_CROSS_ORIGIN_SUB_REQUESTS
	} else {
		phase = PHASE_SUBSEQUENT_REQUESTS
	}

	lifecycle.CurrentPhase = phase

	// 添加到已发现源列表
	found := false
	for _, origin := range lifecycle.DiscoveredOrigins {
		if origin == requestOriginURL {
			found = true
			break
		}
	}
	if !found {
		lifecycle.DiscoveredOrigins = append(lifecycle.DiscoveredOrigins, requestOriginURL)
	}

	// 检查提示完整性
	riskIndicators := []string{}
	if isCrossOrigin {
		// 验证跨域委托
		if lifecycle.NegotiationStrategy != nil {
			err := m.negotiationAnalyzer.HandleCrossOriginDelegation(
				lifecycle.NegotiationStrategy,
				requestOriginURL,
				includedHints,
			)
			if err != nil {
				riskIndicators = append(riskIndicators, fmt.Sprintf("CROSS_ORIGIN_ERROR:%s", err.Error()))
			}
		}

		// 检查是否遵守 Permissions-Policy
		if lifecycle.PermissionsPolicy != nil {
			riskIndicators = m.checkPermissionsPolicyCompliance(lifecycle.PermissionsPolicy, includedHints)
		}
	}

	// 检查提示是否与协商的一致
	if lifecycle.NegotiationStrategy != nil {
		for _, hint := range includedHints {
			found := false
			for _, negotiated := range lifecycle.NegotiationStrategy.NegotiatedHints {
				if negotiated == hint {
					found = true
					break
				}
			}
			if !found {
				riskIndicators = append(riskIndicators, fmt.Sprintf("UNAUTHORIZED_HINT_SENT:%s", hint))
			}
		}
	}

	// 记录事件
	lifecycle.EventLog = append(lifecycle.EventLog, CHLifecycleEvent{
		Timestamp: time.Now(),
		Type:      phase,
		OriginURL: requestOriginURL,
		Hints:     includedHints,
		Details: map[string]interface{}{
			"is_cross_origin": isCrossOrigin,
			"hint_count":      len(includedHints),
		},
		RiskIndicators: riskIndicators,
	})

	return nil
}

// checkPermissionsPolicyCompliance 检查是否遵守 Permissions-Policy
func (m *CHLifecycleManager) checkPermissionsPolicyCompliance(pol *policy.PermissionsPolicy, hints []string) []string {
	riskIndicators := []string{}

	// Client Hints 相关的功能
	chFeatures := map[string]bool{
		"ch-device-memory":        true,
		"ch-dpr":                  true,
		"ch-downlink":             true,
		"ch-ect":                  true,
		"ch-prefers-color-scheme": true,
		"ch-rtt":                  true,
		"ch-ua":                   true,
		"ch-ua-arch":              true,
		"ch-ua-bitness":           true,
		"ch-ua-mobile":            true,
		"ch-ua-model":             true,
		"ch-ua-platform":          true,
		"ch-ua-platform-version":  true,
	}

	for _, hint := range hints {
		// 检查提示对应的权限指令
		featureName := "ch-" + hint
		if !chFeatures[featureName] {
			featureName = hint // 可能是非标准提示
		}

		if directive, exists := pol.Directives[featureName]; exists {
			if directive.HasNone {
				riskIndicators = append(riskIndicators, fmt.Sprintf("POLICY_VIOLATION:%s", hint))
			}
		}
	}

	return riskIndicators
}

// TerminateLifecycle 终止生命周期
func (m *CHLifecycleManager) TerminateLifecycle(primaryOriginURL string) (*ClientHintsLifecycle, error) {
	lifecycle, exists := m.lifecycles[primaryOriginURL]
	if !exists {
		return nil, fmt.Errorf("lifecycle not found for %s", primaryOriginURL)
	}

	lifecycle.CurrentPhase = PHASE_TERMINATED

	// 计算最终风险分数
	m.calculateFinalRiskScore(lifecycle)

	// 记录终止事件
	lifecycle.EventLog = append(lifecycle.EventLog, CHLifecycleEvent{
		Timestamp: time.Now(),
		Type:      PHASE_TERMINATED,
		OriginURL: primaryOriginURL,
		Details: map[string]interface{}{
			"duration_seconds":   time.Since(lifecycle.StartTime).Seconds(),
			"total_events":       len(lifecycle.EventLog),
			"discovered_origins": len(lifecycle.DiscoveredOrigins),
		},
	})

	return lifecycle, nil
}

// calculateFinalRiskScore 计算最终风险分数
func (m *CHLifecycleManager) calculateFinalRiskScore(lifecycle *ClientHintsLifecycle) {
	risk := 0.0

	// 协商策略风险
	if lifecycle.NegotiationStrategy != nil {
		risk += lifecycle.NegotiationStrategy.RiskScore * 0.3
	}

	// 权限策略风险
	if lifecycle.PermissionsPolicy != nil {
		risk += lifecycle.PermissionsPolicy.RiskScore * 0.3
	}

	// 事件中的风险指标
	anomalyCount := 0
	for _, event := range lifecycle.EventLog {
		anomalyCount += len(event.RiskIndicators)
	}
	if anomalyCount > 5 {
		risk += 0.2
		lifecycle.IntegrityFlags = append(lifecycle.IntegrityFlags, "EXCESSIVE_ANOMALIES_DETECTED")
	}

	// 跨域发现过多
	if len(lifecycle.DiscoveredOrigins) > 10 {
		risk += 0.15
		lifecycle.IntegrityFlags = append(lifecycle.IntegrityFlags, "EXCESSIVE_CROSS_ORIGIN_DISCOVERY")
	}

	// 生命周期过长
	duration := time.Since(lifecycle.StartTime)
	if duration < time.Second {
		// 过于快速可能表示自动化
		risk += 0.1
		lifecycle.IntegrityFlags = append(lifecycle.IntegrityFlags, "UNUSUALLY_SHORT_LIFECYCLE")
	} else if duration > 24*time.Hour {
		// 超过 24 小时的会话可能有问题
		risk += 0.05
		lifecycle.IntegrityFlags = append(lifecycle.IntegrityFlags, "UNUSUALLY_LONG_LIFECYCLE")
	}

	lifecycle.RiskScore = risk
}

// GetLifecycleReport 获取生命周期报告
func (m *CHLifecycleManager) GetLifecycleReport(primaryOriginURL string) (*ClientHintsLifecycle, error) {
	lifecycle, exists := m.lifecycles[primaryOriginURL]
	if !exists {
		return nil, fmt.Errorf("lifecycle not found for %s", primaryOriginURL)
	}

	return lifecycle, nil
}

// GetLifecycleMetrics 获取生命周期指标
func (m *CHLifecycleManager) GetLifecycleMetrics(lifecycle *ClientHintsLifecycle) map[string]interface{} {
	metrics := make(map[string]interface{})

	metrics["primary_origin"] = lifecycle.PrimaryOriginURL
	metrics["current_phase"] = lifecycle.CurrentPhase
	metrics["total_events"] = len(lifecycle.EventLog)
	metrics["discovered_origins"] = len(lifecycle.DiscoveredOrigins)
	metrics["active_hints"] = len(lifecycle.ActiveHints)
	metrics["risk_score"] = lifecycle.RiskScore
	metrics["duration_seconds"] = time.Since(lifecycle.StartTime).Seconds()
	metrics["integrity_flags"] = lifecycle.IntegrityFlags

	if lifecycle.NegotiationStrategy != nil {
		metrics["negotiated_hints"] = len(lifecycle.NegotiationStrategy.NegotiatedHints)
		metrics["rejected_hints"] = len(lifecycle.NegotiationStrategy.RejectedHints)
	}

	if lifecycle.PermissionsPolicy != nil {
		metrics["policy_features"] = len(lifecycle.PermissionsPolicy.Directives)
		metrics["policy_anomalies"] = len(lifecycle.PermissionsPolicy.AnomalyFlags)
	}

	return metrics
}

// GetSummary 获取摘要
func (m *CHLifecycleManager) GetSummary(lifecycle *ClientHintsLifecycle) string {
	phaseStr := ""
	switch lifecycle.CurrentPhase {
	case PHASE_INITIAL_REQUEST:
		phaseStr = "初始请求"
	case PHASE_SERVER_RESPONSE:
		phaseStr = "服务器响应"
	case PHASE_SUBSEQUENT_REQUESTS:
		phaseStr = "后续请求"
	case PHASE_CROSS_ORIGIN_SUB_REQUESTS:
		phaseStr = "跨域子资源"
	case PHASE_TERMINATED:
		phaseStr = "已终止"
	}

	return fmt.Sprintf(
		"源: %s | 阶段: %s | 发现源: %d | 事件: %d | 风险分数: %.2f",
		lifecycle.PrimaryOriginURL,
		phaseStr,
		len(lifecycle.DiscoveredOrigins),
		len(lifecycle.EventLog),
		lifecycle.RiskScore,
	)
}
