package fingerprint

import (
	"fmt"
	"strings"
)

// NegotiationState Client Hints 协商状态
type NegotiationState int

const (
	// NEGOTIATION_INIT 初始状态（未进行协商）
	NEGOTIATION_INIT NegotiationState = iota
	// NEGOTIATION_REQUESTED 服务器请求提示
	NEGOTIATION_REQUESTED
	// NEGOTIATION_ACCEPTED 客户端已接受
	NEGOTIATION_ACCEPTED
	// NEGOTIATION_REJECTED 客户端拒绝
	NEGOTIATION_REJECTED
	// NEGOTIATION_DELEGATED 已委托到其他源
	NEGOTIATION_DELEGATED
)

// ServerPreferences 服务器对 Client Hints 的偏好
type ServerPreferences struct {
	// 请求的低熵提示
	LowEntropyHints []string

	// 请求的高熵提示
	HighEntropyHints []string

	// 优先级（0-100，越高优先级越高）
	Priority int

	// 保留时长（秒），-1 表示持久
	CacheDuration int

	// 跨域委托配置
	DelegateToOrigins []string

	// 是否允许不安全连接
	AllowInsecure bool

	// 附加说明
	Description string
}

// ClientCapabilities 客户端 Client Hints 能力声明
type ClientCapabilities struct {
	// 支持的低熵提示
	SupportedLowEntropy []string

	// 支持的高熵提示
	SupportedHighEntropy []string

	// 浏览器特定的限制
	BrowserLimitations []string

	// 隐私保护等级 (0-100)
	PrivacyLevel int

	// 用户同意状态
	UserConsent bool

	// 设备类型
	DeviceType string
}

// NegotiationStrategy Client Hints 协商策略
type NegotiationStrategy struct {
	// 协商状态
	State NegotiationState

	// 服务器偏好
	ServerPrefs *ServerPreferences

	// 客户端能力
	ClientCaps *ClientCapabilities

	// 已协商的提示集合
	NegotiatedHints []string

	// 被拒绝的提示
	RejectedHints []string

	// 下一次协商的时间（Unix 秒）
	NextNegotiationTime int64

	// 协商历史
	NegotiationHistory []NegotiationRecord

	// 风险评分
	RiskScore float64

	// 异常标记
	AnomalyFlags []string
}

// NegotiationRecord 单次协商记录
type NegotiationRecord struct {
	// 时间戳
	Timestamp int64

	// 请求的提示
	RequestedHints []string

	// 提供的提示
	ProvidedHints []string

	// 决策原因
	Decision string

	// 风险指标
	RiskIndicators []string
}

// CHNegotiationAnalyzer Client Hints 协商分析器
type CHNegotiationAnalyzer struct {
	strategies map[string]*NegotiationStrategy
}

// NewCHNegotiationAnalyzer 创建协商分析器
func NewCHNegotiationAnalyzer() *CHNegotiationAnalyzer {
	return &CHNegotiationAnalyzer{
		strategies: make(map[string]*NegotiationStrategy),
	}
}

// InitializeFromAcceptCH 从 Accept-CH 响应头初始化协商
func (a *CHNegotiationAnalyzer) InitializeFromAcceptCH(acceptCHValue string, origin string) *NegotiationStrategy {
	prefs := a.parseAcceptCH(acceptCHValue)

	strategy := &NegotiationStrategy{
		State:              NEGOTIATION_REQUESTED,
		ServerPrefs:        prefs,
		ClientCaps:         a.getDefaultCapabilities(),
		NegotiatedHints:    []string{},
		RejectedHints:      []string{},
		NegotiationHistory: []NegotiationRecord{},
		AnomalyFlags:       []string{},
	}

	a.strategies[origin] = strategy

	// 评估异常标记
	if len(prefs.HighEntropyHints) > 5 {
		strategy.AnomalyFlags = append(strategy.AnomalyFlags, "EXCESSIVE_HIGH_ENTROPY_HINTS")
	}

	a.evaluateNegotiationRisk(strategy)

	return strategy
}

// parseAcceptCH 解析 Accept-CH 头
func (a *CHNegotiationAnalyzer) parseAcceptCH(acceptCHValue string) *ServerPreferences {
	prefs := &ServerPreferences{
		LowEntropyHints:   []string{},
		HighEntropyHints:  []string{},
		Priority:          50,
		CacheDuration:     -1,
		DelegateToOrigins: []string{},
	}

	if acceptCHValue == "" {
		return prefs
	}

	// 标准 Client Hints
	lowEntropyStandard := map[string]bool{
		"sec-ch-ua":          true,
		"sec-ch-ua-mobile":   true,
		"sec-ch-ua-platform": true,
		"user-agent":         true,
	}

	highEntropyStandard := map[string]bool{
		"sec-ch-ua-full-version":     true,
		"sec-ch-ua-platform-version": true,
		"sec-ch-ua-model":            true,
		"sec-ch-ua-arch":             true,
		"sec-ch-ua-bitness":          true,
		"dpr":                        true,
		"viewport-width":             true,
		"device-memory":              true,
		"downlink":                   true,
		"ect":                        true,
		"rtt":                        true,
		"save-data":                  true,
	}

	// 解析提示列表
	parts := strings.Split(acceptCHValue, ",")
	for _, part := range parts {
		hint := strings.TrimSpace(part)
		if hint == "" {
			continue
		}

		// 移除引号和额外标记
		hint = strings.Trim(hint, `"`)
		hint = strings.ToLower(hint)

		if lowEntropyStandard[hint] {
			prefs.LowEntropyHints = append(prefs.LowEntropyHints, hint)
		} else if highEntropyStandard[hint] {
			prefs.HighEntropyHints = append(prefs.HighEntropyHints, hint)
		}
	}

	// 评估优先级
	if len(prefs.HighEntropyHints) > 5 {
		prefs.Priority = 80
	}

	return prefs
}

// getDefaultCapabilities 获取默认客户端能力
func (a *CHNegotiationAnalyzer) getDefaultCapabilities() *ClientCapabilities {
	return &ClientCapabilities{
		SupportedLowEntropy: []string{
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
		},
		SupportedHighEntropy: []string{
			"sec-ch-ua-full-version",
			"sec-ch-ua-platform-version",
			"sec-ch-ua-arch",
			"sec-ch-ua-bitness",
			"sec-ch-ua-model",
		},
		BrowserLimitations: []string{},
		PrivacyLevel:       70,
		UserConsent:        false,
		DeviceType:         "desktop",
	}
}

// DecideHints 根据服务器请求和客户端能力决定要发送的提示
func (a *CHNegotiationAnalyzer) DecideHints(strategy *NegotiationStrategy, userPreference string) []string {
	decided := []string{}

	// 首先处理低熵提示（无条件允许）
	for _, hint := range strategy.ServerPrefs.LowEntropyHints {
		if a.isSupportedHint(strategy.ClientCaps, hint) {
			decided = append(decided, hint)
		} else {
			strategy.RejectedHints = append(strategy.RejectedHints, hint)
		}
	}

	// 处理高熵提示（需要用户同意）
	for _, hint := range strategy.ServerPrefs.HighEntropyHints {
		if a.isSupportedHint(strategy.ClientCaps, hint) {
			// 检查用户同意
			if strategy.ClientCaps.UserConsent || userPreference == "allow-all" {
				decided = append(decided, hint)
			} else {
				strategy.RejectedHints = append(strategy.RejectedHints, hint)
				strategy.AnomalyFlags = append(strategy.AnomalyFlags, "HIGH_ENTROPY_WITHOUT_CONSENT")
			}
		}
	}

	strategy.NegotiatedHints = decided
	strategy.State = NEGOTIATION_ACCEPTED

	return decided
}

// isSupportedHint 检查提示是否被支持
func (a *CHNegotiationAnalyzer) isSupportedHint(caps *ClientCapabilities, hint string) bool {
	for _, h := range caps.SupportedLowEntropy {
		if h == hint {
			return true
		}
	}
	for _, h := range caps.SupportedHighEntropy {
		if h == hint {
			return true
		}
	}
	return false
}

// HandleCrossOriginDelegation 处理跨域委托
func (a *CHNegotiationAnalyzer) HandleCrossOriginDelegation(strategy *NegotiationStrategy, delegateOrigin string, delegateHints []string) error {
	// 验证委托源是否被授权
	isAuthorized := false
	for _, allowed := range strategy.ServerPrefs.DelegateToOrigins {
		if allowed == delegateOrigin || allowed == "*" {
			isAuthorized = true
			break
		}
	}

	if !isAuthorized {
		strategy.AnomalyFlags = append(strategy.AnomalyFlags, fmt.Sprintf("UNAUTHORIZED_DELEGATION_ATTEMPT:%s", delegateOrigin))
		return fmt.Errorf("delegation to %s not authorized", delegateOrigin)
	}

	strategy.State = NEGOTIATION_DELEGATED

	// 验证委托的提示是否在服务器允许的范围内
	for _, delegatedHint := range delegateHints {
		found := false
		for _, negotiated := range strategy.NegotiatedHints {
			if negotiated == delegatedHint {
				found = true
				break
			}
		}
		if !found {
			strategy.AnomalyFlags = append(strategy.AnomalyFlags, fmt.Sprintf("UNAUTHORIZED_HINT_DELEGATION:%s", delegatedHint))
		}
	}

	return nil
}

// evaluateNegotiationRisk 评估协商风险
func (a *CHNegotiationAnalyzer) evaluateNegotiationRisk(strategy *NegotiationStrategy) {
	risk := 0.0

	// 检查异常数量
	if len(strategy.AnomalyFlags) > 2 {
		risk += 0.2
	}

	// 检查高熵提示过多（可能是指纹追踪企图）
	if len(strategy.ServerPrefs.HighEntropyHints) > 8 {
		risk += 0.3
		strategy.AnomalyFlags = append(strategy.AnomalyFlags, "EXCESSIVE_FINGERPRINTING_ATTEMPT")
	}

	// 检查委托到过多源
	if len(strategy.ServerPrefs.DelegateToOrigins) > 5 {
		risk += 0.2
		strategy.AnomalyFlags = append(strategy.AnomalyFlags, "EXCESSIVE_DOMAIN_DELEGATION")
	}

	// 检查不安全连接上的高熵提示
	if strategy.ServerPrefs.AllowInsecure && len(strategy.ServerPrefs.HighEntropyHints) > 0 {
		risk += 0.4
		strategy.AnomalyFlags = append(strategy.AnomalyFlags, "HIGH_ENTROPY_OVER_INSECURE")
	}

	// 缓存时间过长
	if strategy.ServerPrefs.CacheDuration > 31536000 { // 超过 1 年
		risk += 0.15
	}

	strategy.RiskScore = risk
}

// GetNegotiationSummary 获取协商总结
func (a *CHNegotiationAnalyzer) GetNegotiationSummary(strategy *NegotiationStrategy) string {
	stateStr := ""
	switch strategy.State {
	case NEGOTIATION_INIT:
		stateStr = "初始"
	case NEGOTIATION_REQUESTED:
		stateStr = "已请求"
	case NEGOTIATION_ACCEPTED:
		stateStr = "已接受"
	case NEGOTIATION_REJECTED:
		stateStr = "已拒绝"
	case NEGOTIATION_DELEGATED:
		stateStr = "已委托"
	}

	return fmt.Sprintf(
		"状态: %s | 协商提示: %d | 被拒提示: %d | 异常标记: %d | 风险分数: %.2f",
		stateStr,
		len(strategy.NegotiatedHints),
		len(strategy.RejectedHints),
		len(strategy.AnomalyFlags),
		strategy.RiskScore,
	)
}

// GetAllStrategies 获取所有源的协商策略
func (a *CHNegotiationAnalyzer) GetAllStrategies() map[string]*NegotiationStrategy {
	return a.strategies
}
