package main

import (
	"fmt"

	fp "github.com/vistone/fingerprint"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("综合风险评分示例")
	fmt.Println("========================================")
	fmt.Println()

	// 示例 1: 安全请求
	fmt.Println("=== 示例 1: 安全请求 ===")
	safeInput := fp.RiskInput{
		JA3Hash: "known_safe_chrome_hash",
		JA4Hash: "known_safe_ja4_hash",
		JA4SResult: &fp.JA4SResult{
			Hash:         "safe_server_hash",
			RiskScore:    0.05,
			AnomalyFlags: []string{},
		},
		HTTP2Result: &fp.HTTP2SignatureResult{
			Hash:         "chrome_http2_hash",
			RiskScore:    0.02,
			AnomalyFlags: []string{},
		},
		Context: fp.RiskContext{
			IsKnownClient:   true,
			IPReputation:    0.95,
			HistoricalScore: 0.98,
			RequestRate:     10,
		},
	}

	result1, _ := fp.CalculateRisk(safeInput)
	printRiskResult("安全请求", result1)

	// 示例 2: 低风险请求
	fmt.Println("\n=== 示例 2: 低风险请求 ===")
	lowRiskInput := fp.RiskInput{
		JA3Hash: "less_common_hash",
		JA4SResult: &fp.JA4SResult{
			Hash:         "server_hash",
			RiskScore:    0.25,
			AnomalyFlags: []string{"UNUSUAL_EXTENSION_ORDER"},
		},
		HTTP2Result: &fp.HTTP2SignatureResult{
			Hash:         "http2_hash",
			RiskScore:    0.15,
			AnomalyFlags: []string{},
		},
		Context: fp.RiskContext{
			IPReputation: 0.7,
			RequestRate:  25,
		},
	}

	result2, _ := fp.CalculateRisk(lowRiskInput)
	printRiskResult("低风险请求", result2)

	// 示例 3: 中风险请求（多个异常）
	fmt.Println("\n=== 示例 3: 中风险请求 ===")
	mediumRiskInput := fp.RiskInput{
		JA3Hash: "suspicious_hash",
		JA4SResult: &fp.JA4SResult{
			Hash:         "suspicious_server",
			RiskScore:    0.45,
			AnomalyFlags: []string{"WEAK_CIPHER", "UNUSUAL_EXTENSION_ORDER"},
		},
		HTTP2Result: &fp.HTTP2SignatureResult{
			Hash:         "unusual_http2",
			RiskScore:    0.38,
			AnomalyFlags: []string{"UNUSUAL_FRAME_ORDER"},
		},
		JA4HResult: &fp.JA4HResult{
			Hash:         "suspicious_headers",
			RiskScore:    0.42,
			AnomalyFlags: []string{"MINOR_UA_INCONSISTENCY"},
		},
		Context: fp.RiskContext{
			IPReputation: 0.5,
			RequestRate:  45,
		},
	}

	result3, _ := fp.CalculateRisk(mediumRiskInput)
	printRiskResult("中风险请求", result3)

	// 示例 4: 高风险请求（多维度异常）
	fmt.Println("\n=== 示例 4: 高风险请求 ===")
	highRiskInput := fp.RiskInput{
		JA3Hash: "unknown_hash",
		JA4Hash: "suspicious_ja4",
		JA4SResult: &fp.JA4SResult{
			Hash:         "malicious_server",
			RiskScore:    0.75,
			AnomalyFlags: []string{"WEAK_CIPHER", "UNUSUAL_EXTENSION_ORDER", "NON_STANDARD_VERSION"},
		},
		HTTP2Result: &fp.HTTP2SignatureResult{
			Hash:         "bot_http2",
			RiskScore:    0.68,
			AnomalyFlags: []string{"UNUSUAL_FRAME_ORDER", "MISSING_SETTINGS"},
		},
		JA4HResult: &fp.JA4HResult{
			Hash:         "forged_headers",
			RiskScore:    0.72,
			AnomalyFlags: []string{"UA_CH_MISMATCH", "SUSPICIOUS_HEADERS"},
		},
		QUICResult: &fp.QUICSignatureResult{
			Hash:         "unusual_quic",
			RiskScore:    0.55,
			AnomalyFlags: []string{"DRAFT_VERSION"},
		},
		Context: fp.RiskContext{
			IPReputation: 0.25,
			GeoRisk:      0.4,
			RequestRate:  85,
		},
	}

	result4, _ := fp.CalculateRisk(highRiskInput)
	printRiskResult("高风险请求", result4)

	// 示例 5: 极高风险（可能攻击）
	fmt.Println("\n=== 示例 5: 极高风险（可能攻击） ===")
	criticalInput := fp.RiskInput{
		JA3Hash: "blacklisted_hash",
		JA4Hash: "known_attack_tool",
		JA4SResult: &fp.JA4SResult{
			Hash:         "attack_signature",
			RiskScore:    0.95,
			AnomalyFlags: []string{"WEAK_CIPHER", "NON_STANDARD_VERSION", "SUSPICIOUS_CONFIG"},
		},
		HTTP2Result: &fp.HTTP2SignatureResult{
			Hash:         "automated_tool",
			RiskScore:    0.88,
			AnomalyFlags: []string{"UNUSUAL_FRAME_ORDER", "MISSING_SETTINGS", "ABNORMAL_PRIORITY"},
		},
		JA4HResult: &fp.JA4HResult{
			Hash:         "clearly_forged",
			RiskScore:    0.92,
			AnomalyFlags: []string{"UA_CH_MISMATCH", "SUSPICIOUS_HEADERS", "MISSING_COMMON_HEADERS"},
		},
		QUICResult: &fp.QUICSignatureResult{
			Hash:         "exploit_attempt",
			RiskScore:    0.85,
			AnomalyFlags: []string{"DRAFT_VERSION", "SUSPICIOUS_LIMITS"},
		},
		ECHResult: &fp.ECHAnalysisResult{
			ECHPresent:   true,
			RiskScore:    0.35,
			AnomalyFlags: []string{"ECH_WITH_VISIBLE_SNI"},
		},
		Context: fp.RiskContext{
			IPReputation: 0.05,
			GeoRisk:      0.8,
			RequestRate:  200,
		},
	}

	result5, _ := fp.CalculateRisk(criticalInput)
	printRiskResult("极高风险", result5)

	// 示例 6: 带 ECH 的请求
	fmt.Println("\n=== 示例 6: 带 ECH 的请求 ===")
	echInput := fp.RiskInput{
		JA3Hash: "modern_browser",
		ECHResult: &fp.ECHAnalysisResult{
			ECHPresent:             true,
			ECHType:                "outer",
			RiskScore:              0.15,
			VisibleFieldsSignature: "ech_signature_hash",
			Impact: fp.ECHImpact{
				ImpactLevel: "high",
				SNIVisible:  false,
			},
			AnomalyFlags: []string{},
		},
		HTTP2Result: &fp.HTTP2SignatureResult{
			RiskScore:    0.08,
			AnomalyFlags: []string{},
		},
	}

	result6, _ := fp.CalculateRisk(echInput)
	printRiskResult("ECH 请求", result6)

	// 示例 7: 自定义配置的风险评分
	fmt.Println("\n=== 示例 7: 自定义严格配置 ===")
	strictConfig := &fp.ScoringConfig{
		Weights: fp.DimensionWeights{
			TLSFingerprint:  0.25, // 提高 TLS 权重
			ServerBehavior:  0.20,
			HTTP2Signature:  0.15,
			HTTPHeaders:     0.15,
			QUICSignature:   0.10,
			ClientHints:     0.08,
			ECHImpact:       0.02,
			BehaviorAnomaly: 0.05,
		},
		Thresholds: fp.ThreatThresholds{
			Safe:     0.15, // 更严格的阈值
			Low:      0.35,
			Medium:   0.55,
			High:     0.75,
			Critical: 1.0,
		},
		StrictMode:    true,
		MinConfidence: 0.7,
	}

	scorer := fp.NewRiskScorer(strictConfig)
	strictInput := fp.RiskInput{
		JA4SResult: &fp.JA4SResult{
			RiskScore:    0.35,
			AnomalyFlags: []string{"WEAK_CIPHER"},
		},
		HTTP2Result: &fp.HTTP2SignatureResult{
			RiskScore:    0.25,
			AnomalyFlags: []string{"UNUSUAL_FRAME_ORDER"},
		},
	}

	result7, _ := scorer.CalculateRisk(strictInput)
	printRiskResult("严格模式", result7)

	// 总结
	fmt.Println("\n========================================")
	fmt.Println("风险评分总结")
	fmt.Println("========================================")
	printSummary()
}

func printRiskResult(name string, result *fp.RiskScore) {
	fmt.Printf("场景: %s\n", name)
	fmt.Printf("总分: %.2f\n", result.TotalScore)
	fmt.Printf("威胁等级: %s\n", result.ThreatLevel)
	fmt.Printf("置信度: %.2f\n", result.Confidence)
	fmt.Printf("异常数量: %d\n", result.AnomalyCount)

	fmt.Println("\n各维度评分:")
	fmt.Printf("  TLS 指纹: %.2f\n", result.Dimensions.TLSFingerprint)
	fmt.Printf("  服务端行为: %.2f\n", result.Dimensions.ServerBehavior)
	fmt.Printf("  HTTP/2: %.2f\n", result.Dimensions.HTTP2Signature)
	fmt.Printf("  HTTP 请求头: %.2f\n", result.Dimensions.HTTPHeaders)
	fmt.Printf("  QUIC: %.2f\n", result.Dimensions.QUICSignature)
	fmt.Printf("  Client Hints: %.2f\n", result.Dimensions.ClientHints)
	fmt.Printf("  ECH 影响: %.2f\n", result.Dimensions.ECHImpact)
	fmt.Printf("  行为异常: %.2f\n", result.Dimensions.BehaviorAnomaly)

	if len(result.RiskFactors) > 0 {
		fmt.Println("\n⚠️ 风险因素:")
		for i, factor := range result.RiskFactors {
			if i >= 5 {
				fmt.Printf("  ... 还有 %d 个因素\n", len(result.RiskFactors)-5)
				break
			}
			fmt.Printf("  - [%.2f] %s: %s\n",
				factor.Severity,
				factor.Type,
				factor.Description,
			)
		}
	} else {
		fmt.Println("\n✓ 无风险因素")
	}

	if len(result.Recommendations) > 0 {
		fmt.Println("\n建议措施:")
		for i, rec := range result.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec)
		}
	}
}

func printSummary() {
	fmt.Println(`
综合风险评分系统整合了多个指纹分析维度：

1. 评分维度 (8个):
   - TLS 指纹 (JA3/JA4)
   - 服务端行为 (JA4S)
   - HTTP/2 签名
   - HTTP 请求头 (JA4H)
   - QUIC 签名
   - Client Hints 一致性
   - ECH 影响评估
   - 行为异常

2. 威胁等级 (5级):
   - safe: 0.0-0.2 (安全)
   - low: 0.2-0.4 (低风险)
   - medium: 0.4-0.6 (中风险)
   - high: 0.6-0.8 (高风险)
   - critical: 0.8-1.0 (极高风险)

3. 上下文调整:
   - IP 信誉评分
   - 地理位置风险
   - 历史行为记录
   - 请求频率分析

4. 风险因素跟踪:
   - 类型分类
   - 严重程度量化
   - 证据记录
   - 置信度评估

5. 建议措施:
   - 基于威胁等级的分级响应
   - 异常模式的针对性建议
   - 数据不足时的改进方向

6. 可配置性:
   - 自定义维度权重
   - 可调整威胁阈值
   - 严格/宽松模式切换
   - 最小置信度要求

使用场景:
- API 网关访问控制
- WAF 威胁情报输入
- 反爬虫系统
- 欺诈检测
- 行为分析`)
}
