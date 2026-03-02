package fingerprint

import (
	"fmt"
	"strings"
)

// PermissionDirective 权限指令
type PermissionDirective struct {
	// 功能名称
	FeatureName string

	// 获准的源列表
	AllowedOrigins []string

	// 包含通配符
	HasWildcard bool

	// 包含 self
	HasSelf bool

	// 包含明确的 none
	HasNone bool

	// 是否允许所有
	AllowAll bool

	// 是否是默认允许行为
	IsDefault bool
}

// PermissionsPolicy 权限策略（Permissions-Policy 头）
type PermissionsPolicy struct {
	// 权限指令映射
	Directives map[string]*PermissionDirective

	// 是否为传统格式（Feature-Policy）
	IsLegacy bool

	// 原始头值
	RawValue string

	// 异常标记
	AnomalyFlags []string

	// 风险分数
	RiskScore float64

	// 签名哈希
	Hash string
}

// PermissionsPolicyAnalyzer 权限策略分析器
type PermissionsPolicyAnalyzer struct {
	// 标准功能列表
	standardFeatures map[string]bool

	// 已知危险组合
	dangerousCombinations [][]string
}

// NewPermissionsPolicyAnalyzer 创建权限策略分析器
func NewPermissionsPolicyAnalyzer() *PermissionsPolicyAnalyzer {
	return &PermissionsPolicyAnalyzer{
		standardFeatures: map[string]bool{
			"accelerometer":                   true,
			"ambient-light-sensor":            true,
			"autoplay":                        true,
			"bluetooth":                       true,
			"browsing-topics":                 true,
			"camera":                          true,
			"can-pause-videos-autoplay":       true,
			"ch-device-memory":                true,
			"ch-dpr":                          true,
			"ch-downlink":                     true,
			"ch-ect":                          true,
			"ch-prefers-color-scheme":         true,
			"ch-rtt":                          true,
			"ch-ua":                           true,
			"ch-ua-arch":                      true,
			"ch-ua-bitness":                   true,
			"ch-ua-mobile":                    true,
			"ch-ua-model":                     true,
			"ch-ua-platform":                  true,
			"ch-ua-platform-version":          true,
			"clipboard-read":                  true,
			"clipboard-write":                 true,
			"cross-origin-isolated":           true,
			"document-domain":                 true,
			"document-picture-in-picture":     true,
			"encrypted-media":                 true,
			"execution-while-not-rendered":    true,
			"execution-while-out-of-viewport": true,
			"fullscreen":                      true,
			"geolocation":                     true,
			"gyroscope":                       true,
			"magnetometer":                    true,
			"microphone":                      true,
			"midi":                            true,
			"otp-credentials":                 true,
			"payment":                         true,
			"picture-in-picture":              true,
			"publickey-credentials-get":       true,
			"speaker-selection":               true,
			"sync-xhr":                        true,
			"usb":                             true,
			"vr":                              true,
			"window-placement":                true,
			"xr-spatial-tracking":             true,
		},
		dangerousCombinations: [][]string{
			{"camera", "microphone", "geolocation"}, // 全面追踪
			{"clipboard-read", "clipboard-write"},   // 剪贴板访问
			{"payment", "usb", "serial"},            // 支付敏感操作
		},
	}
}

// ParsePermissionsPolicy 解析 Permissions-Policy 头
func (a *PermissionsPolicyAnalyzer) ParsePermissionsPolicy(policyValue string) *PermissionsPolicy {
	policy := &PermissionsPolicy{
		Directives:   make(map[string]*PermissionDirective),
		RawValue:     policyValue,
		AnomalyFlags: []string{},
	}

	if policyValue == "" {
		return policy
	}

	// 检测是否为传统格式
	if strings.Contains(policyValue, ":") && !strings.Contains(policyValue, "=(") {
		policy.IsLegacy = true
	}

	if policy.IsLegacy {
		a.parseFeaturePolicy(policy)
	} else {
		a.parseModernPolicy(policy)
	}

	a.evaluatePolicyRisk(policy)
	a.calculateHash(policy)

	return policy
}

// parseModernPolicy 解析现代 Permissions-Policy 格式
func (a *PermissionsPolicyAnalyzer) parseModernPolicy(policy *PermissionsPolicy) {
	// 格式: feature=(allowlist), feature2=(allowlist2)
	directives := strings.Split(policy.RawValue, ",")

	for _, directive := range directives {
		directive = strings.TrimSpace(directive)
		if directive == "" {
			continue
		}

		// 分离特性名和源列表
		parts := strings.SplitN(directive, "=(", 2)
		if len(parts) != 2 {
			policy.AnomalyFlags = append(policy.AnomalyFlags, fmt.Sprintf("MALFORMED_DIRECTIVE:%s", directive))
			continue
		}

		featureName := strings.TrimSpace(parts[0])
		sourceList := strings.TrimSuffix(strings.TrimSpace(parts[1]), ")")

		// 检查未知特性
		if !a.standardFeatures[featureName] {
			policy.AnomalyFlags = append(policy.AnomalyFlags, fmt.Sprintf("UNKNOWN_FEATURE:%s", featureName))
		}

		permDirective := a.parseSourceList(sourceList)
		permDirective.FeatureName = featureName
		policy.Directives[featureName] = permDirective
	}
}

// parseFeaturePolicy 解析传统 Feature-Policy 格式
func (a *PermissionsPolicyAnalyzer) parseFeaturePolicy(policy *PermissionsPolicy) {
	// 格式: Feature-Policy: feature 'self' https://example.com
	directives := strings.Split(policy.RawValue, ";")

	for _, directive := range directives {
		directive = strings.TrimSpace(directive)
		if directive == "" {
			continue
		}

		parts := strings.Fields(directive)
		if len(parts) == 0 {
			continue
		}

		featureName := parts[0]
		if !a.standardFeatures[featureName] {
			policy.AnomalyFlags = append(policy.AnomalyFlags, fmt.Sprintf("UNKNOWN_FEATURE:%s", featureName))
		}

		sourceList := strings.Join(parts[1:], " ")
		permDirective := a.parseSourceListLegacy(sourceList)
		permDirective.FeatureName = featureName
		policy.Directives[featureName] = permDirective
	}
}

// parseSourceList 解析源列表（现代格式）
func (a *PermissionsPolicyAnalyzer) parseSourceList(sourceList string) *PermissionDirective {
	directive := &PermissionDirective{
		AllowedOrigins: []string{},
		HasWildcard:    false,
		HasSelf:        false,
		HasNone:        false,
		AllowAll:       false,
		IsDefault:      false,
	}

	sourceList = strings.TrimSpace(sourceList)
	if sourceList == "" {
		directive.HasNone = true
		return directive
	}

	// 分离各源
	sources := strings.FieldsFunc(sourceList, func(r rune) bool {
		return r == ' ' || r == '\t'
	})

	for _, source := range sources {
		source = strings.TrimSpace(source)
		if source == "" {
			continue
		}

		if source == "*" {
			directive.HasWildcard = true
			directive.AllowAll = true
		} else if source == "'self'" || source == "self" {
			directive.HasSelf = true
		} else if source == "'none'" || source == "none" {
			directive.HasNone = true
		} else {
			directive.AllowedOrigins = append(directive.AllowedOrigins, source)
		}
	}

	return directive
}

// parseSourceListLegacy 解析源列表（传统格式）
func (a *PermissionsPolicyAnalyzer) parseSourceListLegacy(sourceList string) *PermissionDirective {
	directive := &PermissionDirective{
		AllowedOrigins: []string{},
		HasWildcard:    false,
		HasSelf:        false,
		HasNone:        false,
		AllowAll:       false,
		IsDefault:      false,
	}

	sourceList = strings.TrimSpace(sourceList)
	if sourceList == "" {
		directive.IsDefault = true
		return directive
	}

	sources := strings.Fields(sourceList)
	for _, source := range sources {
		source = strings.TrimSpace(source)
		if source == "" {
			continue
		}

		if source == "*" {
			directive.HasWildcard = true
			directive.AllowAll = true
		} else if source == "'self'" {
			directive.HasSelf = true
		} else if source == "'none'" {
			directive.HasNone = true
		} else {
			directive.AllowedOrigins = append(directive.AllowedOrigins, source)
		}
	}

	return directive
}

// evaluatePolicyRisk 评估策略风险
func (a *PermissionsPolicyAnalyzer) evaluatePolicyRisk(policy *PermissionsPolicy) {
	risk := 0.0

	// 检查异常
	if len(policy.AnomalyFlags) > 0 {
		risk += 0.1 * float64(len(policy.AnomalyFlags))
	}

	// 检查是否所有特性都允许 (*)
	var allowAllFeatures []string
	for feature, directive := range policy.Directives {
		if directive.AllowAll {
			allowAllFeatures = append(allowAllFeatures, feature)
			risk += 0.15
		}
	}
	if len(allowAllFeatures) > 5 {
		policy.AnomalyFlags = append(policy.AnomalyFlags, "EXCESSIVE_WILDCARD_PERMISSIONS")
		risk += 0.2
	}

	// 检查已知危险组合
	for _, combination := range a.dangerousCombinations {
		count := 0
		for _, danger := range combination {
			if directive, exists := policy.Directives[danger]; exists && !directive.HasNone {
				count++
			}
		}
		if count == len(combination) {
			policy.AnomalyFlags = append(policy.AnomalyFlags, fmt.Sprintf("DANGEROUS_COMBINATION_DETECTED:%v", combination))
			risk += 0.3
		}
	}

	// 检查缺少安全限制的敏感特性
	sensitiveFeaturesWithoutRestriction := []string{
		"camera",
		"microphone",
		"geolocation",
		"payment",
		"usb",
	}
	for _, feature := range sensitiveFeaturesWithoutRestriction {
		if directive, exists := policy.Directives[feature]; exists && directive.AllowAll {
			policy.AnomalyFlags = append(policy.AnomalyFlags, fmt.Sprintf("UNRESTRICTED_SENSITIVE_FEATURE:%s", feature))
			risk += 0.2
		}
	}

	// 检查传统格式（可能已过时或配置不当）
	if policy.IsLegacy {
		policy.AnomalyFlags = append(policy.AnomalyFlags, "LEGACY_FEATURE_POLICY_FORMAT")
		risk += 0.1
	}

	policy.RiskScore = risk
}

// calculateHash 计算策略哈希
func (a *PermissionsPolicyAnalyzer) calculateHash(policy *PermissionsPolicy) {
	// 简单哈希：特性数量和异常数
	policy.Hash = fmt.Sprintf(
		"%d-%d-%d",
		len(policy.Directives),
		len(policy.AnomalyFlags),
		int(policy.RiskScore*100),
	)
}

// GetPolicySummary 获取策略总结
func (a *PermissionsPolicyAnalyzer) GetPolicySummary(policy *PermissionsPolicy) string {
	allowAll := 0
	restricted := 0

	for _, directive := range policy.Directives {
		if directive.AllowAll {
			allowAll++
		} else if directive.HasNone {
			restricted++
		}
	}

	return fmt.Sprintf(
		"特性数: %d | 允许所有: %d | 受限: %d | 异常标记: %d | 风险分数: %.2f",
		len(policy.Directives),
		allowAll,
		restricted,
		len(policy.AnomalyFlags),
		policy.RiskScore,
	)
}
