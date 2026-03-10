package agent

import (
	"fmt"
	"math"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
)

// AnomalyDetector 知识驱动的异常检测器
//
// 利用 KnowledgeBase 中的全球指纹蓝图，对每一次观测进行
// 跨层一致性校验：TLS ↔ HTTP/2 ↔ HTTP 头 ↔ TCP/IP ↔ JS 特征
// 是否共同指向同一个真实浏览器身份。
//
// 任何矛盾信号会被标记为 Contradiction，累积成 SuspicionScore。
type AnomalyDetector struct {
	kb *KnowledgeBase
}

// NewAnomalyDetector 创建知识驱动异常检测器
func NewAnomalyDetector(kb *KnowledgeBase) *AnomalyDetector {
	return &AnomalyDetector{kb: kb}
}

// Analyze 对一次观测进行知识驱动的异常分析
//
// 它会：
//  1. 从 ML 分类结果取出声称的浏览器身份
//  2. 用 KnowledgeBase 查找该身份的已知蓝图
//  3. 逐层校验 TLS、HTTP/2、TCP/IP 等特征是否与蓝图一致
//  4. 汇总矛盾信号，计算可疑度
func (ad *AnomalyDetector) Analyze(obs *Observation) *MatchResult {
	result := &MatchResult{
		MatchScore:     1.0,
		SuspicionScore: 0.0,
	}

	// 无分类结果时无法进行知识校验
	if obs.Classification == nil {
		return result
	}

	family := obs.Classification.Family
	version := obs.Classification.Version
	result.ClosestFamily = family
	result.ClosestVersion = version

	bk := ad.kb.GetBrowserKnowledge(family)
	if bk == nil {
		// 未知浏览器家族，轻微可疑
		result.Contradictions = append(result.Contradictions, Contradiction{
			Field: "browser_family", Expected: "known family", Actual: string(family), Severity: "low",
		})
		result.SuspicionScore = 0.15
		return result
	}

	result.Known = true

	// ---- TLS 层检验 ----
	ad.checkTLS(obs, bk, result)

	// ---- HTTP/2 层检验 ----
	ad.checkHTTP2(obs, family, result)

	// ---- TCP/IP 层检验 ----
	ad.checkTCPIP(obs, result)

	// ---- 分类置信度检验 ----
	ad.checkClassificationConfidence(obs, result)

	// ---- Headless / 自动化工具标记检验 ----
	ad.checkAutomationMarkers(obs, result)

	// 汇总可疑度
	result.SuspicionScore = ad.computeSuspicionScore(result.Contradictions)
	result.MatchScore = math.Max(0, 1.0-result.SuspicionScore)

	return result
}

// ===================================================================
// TLS 层检验
// ===================================================================

func (ad *AnomalyDetector) checkTLS(obs *Observation, bk *BrowserKnowledge, mr *MatchResult) {
	if obs.Features == nil {
		return
	}

	// 密码套件数量校验
	cipherCount := obs.Features.Get(core.FeatureCipherSuites)
	if cipherCount > 0 {
		rangeInfo, ok := ad.kb.tls.CipherCountRange[bk.Family]
		if ok {
			count := int(cipherCount)
			if count < rangeInfo[0] || count > rangeInfo[1] {
				mr.Contradictions = append(mr.Contradictions, Contradiction{
					Field:    "tls_cipher_count",
					Expected: fmt.Sprintf("%d-%d for %s", rangeInfo[0], rangeInfo[1], bk.Family),
					Actual:   fmt.Sprintf("%d", count),
					Severity: "medium",
				})
			}
		}
	}

	// 扩展数量校验
	extCount := obs.Features.Get(core.FeatureExtensions)
	if extCount > 0 {
		rangeInfo, ok := ad.kb.tls.ExtensionCountRange[bk.Family]
		if ok {
			count := int(extCount)
			if count < rangeInfo[0] || count > rangeInfo[1] {
				mr.Contradictions = append(mr.Contradictions, Contradiction{
					Field:    "tls_extension_count",
					Expected: fmt.Sprintf("%d-%d for %s", rangeInfo[0], rangeInfo[1], bk.Family),
					Actual:   fmt.Sprintf("%d", count),
					Severity: "medium",
				})
			}
		}
	}
}

// ===================================================================
// HTTP/2 层检验
// ===================================================================

func (ad *AnomalyDetector) checkHTTP2(obs *Observation, family core.BrowserType, mr *MatchResult) {
	if obs.Features == nil {
		return
	}

	h2Hash := obs.Features.Get(core.FeatureHTTP2Settings)
	if h2Hash == 0 {
		return
	}

	expected := ad.kb.GetExpectedH2(family)
	if expected == nil {
		return
	}

	// FeatureVector 中 HTTP2Settings 是 hash 值，我们检查元数据中是否有详细信息
	if obs.Features.Metadata == nil {
		return
	}

	// 检查 HTTP/2 InitialWindowSize（如果元数据中有）
	if ws, ok := obs.Features.Metadata["h2_initial_window_size"]; ok {
		if wsVal, ok := ws.(float64); ok {
			expectedWS := float64(expected.InitialWindowSize)
			// 允许 10% 偏差
			if math.Abs(wsVal-expectedWS)/expectedWS > 0.1 {
				mr.Contradictions = append(mr.Contradictions, Contradiction{
					Field:    "h2_initial_window_size",
					Expected: fmt.Sprintf("%d for %s", expected.InitialWindowSize, family),
					Actual:   fmt.Sprintf("%.0f", wsVal),
					Severity: "high",
				})
			}
		}
	}

	// 检查 MaxConcurrentStreams
	if mcs, ok := obs.Features.Metadata["h2_max_concurrent_streams"]; ok {
		if mcsVal, ok := mcs.(float64); ok {
			expectedMCS := float64(expected.MaxConcurrentStreams)
			if mcsVal != expectedMCS {
				mr.Contradictions = append(mr.Contradictions, Contradiction{
					Field:    "h2_max_concurrent_streams",
					Expected: fmt.Sprintf("%.0f for %s", expectedMCS, family),
					Actual:   fmt.Sprintf("%.0f", mcsVal),
					Severity: "high",
				})
			}
		}
	}

	// 检查 pseudo-header 顺序（Chrome vs Firefox vs Safari 各不同）
	if pho, ok := obs.Features.Metadata["pseudo_header_order"]; ok {
		if phoStr, ok := pho.(string); ok && len(expected.PseudoHeaderOrder) > 0 {
			expectedOrder := strings.Join(expected.PseudoHeaderOrder, ",")
			if phoStr != expectedOrder {
				mr.Contradictions = append(mr.Contradictions, Contradiction{
					Field:    "h2_pseudo_header_order",
					Expected: expectedOrder + " (" + string(family) + ")",
					Actual:   phoStr,
					Severity: "high",
				})
			}
		}
	}
}

// ===================================================================
// TCP/IP 层检验
// ===================================================================

func (ad *AnomalyDetector) checkTCPIP(obs *Observation, mr *MatchResult) {
	if obs.Metadata == nil {
		return
	}

	osFamily := normalizeOSFamily(obs.Metadata["os_family"])
	if osFamily == "" {
		return
	}

	expected := ad.kb.GetExpectedTCPIP(osFamily)
	if expected == nil {
		return
	}

	// TTL 检验
	if ttlStr, ok := obs.Metadata["tcp_ttl"]; ok {
		var ttl int
		if _, err := fmt.Sscanf(ttlStr, "%d", &ttl); err == nil {
			// TTL 会随着 hop 数递减，期望值 ± 一定范围
			expectedTTL := int(expected.TTL)
			// 在 hot-path 中只检查是否与 OS 段匹配 (64段 vs 128段)
			if (expectedTTL > 100 && ttl <= 100) || (expectedTTL <= 100 && ttl > 100) {
				mr.Contradictions = append(mr.Contradictions, Contradiction{
					Field:    "tcp_ttl",
					Expected: fmt.Sprintf("~%d for %s", expectedTTL, osFamily),
					Actual:   fmt.Sprintf("%d", ttl),
					Severity: "high",
				})
			}
		}
	}

	// TCP Window Size 检验
	if wsStr, ok := obs.Metadata["tcp_window_size"]; ok {
		var ws int
		if _, err := fmt.Sscanf(wsStr, "%d", &ws); err == nil {
			expectedWS := int(expected.WindowSize)
			// 允许 20% 偏差（网络条件可能影响）
			diff := math.Abs(float64(ws-expectedWS)) / float64(expectedWS)
			if diff > 0.3 {
				mr.Contradictions = append(mr.Contradictions, Contradiction{
					Field:    "tcp_window_size",
					Expected: fmt.Sprintf("%d for %s", expectedWS, osFamily),
					Actual:   fmt.Sprintf("%d", ws),
					Severity: "medium",
				})
			}
		}
	}
}

// normalizeOSFamily 将各种 OS 表述规范化为知识库使用的 key
func normalizeOSFamily(raw string) string {
	lower := strings.ToLower(raw)
	switch {
	case strings.Contains(lower, "windows"):
		return "windows"
	case strings.Contains(lower, "macos") || strings.Contains(lower, "mac os") || strings.Contains(lower, "macintosh"):
		return "macos"
	case strings.Contains(lower, "linux") && !strings.Contains(lower, "android"):
		return "linux"
	case strings.Contains(lower, "ios") || strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad"):
		return "ios"
	case strings.Contains(lower, "android"):
		return "android"
	default:
		return ""
	}
}

// ===================================================================
// 分类置信度检验
// ===================================================================

func (ad *AnomalyDetector) checkClassificationConfidence(obs *Observation, mr *MatchResult) {
	if obs.Classification == nil {
		return
	}
	// 如果 ML 分类置信度过低，说明特征模式不典型
	if obs.Classification.Confidence < 0.5 {
		mr.Contradictions = append(mr.Contradictions, Contradiction{
			Field:    "classification_confidence",
			Expected: ">= 0.5",
			Actual:   fmt.Sprintf("%.2f", obs.Classification.Confidence),
			Severity: "medium",
		})
	}
}

// ===================================================================
// 自动化工具标记检验
// ===================================================================

func (ad *AnomalyDetector) checkAutomationMarkers(obs *Observation, mr *MatchResult) {
	if obs.Features == nil {
		return
	}

	// 检查 headless browser 特征
	headless := obs.Features.Get(core.FeatureHeadlessBrowser)
	if headless > 0 {
		mr.Contradictions = append(mr.Contradictions, Contradiction{
			Field:    "headless_browser",
			Expected: "0 (no headless markers)",
			Actual:   fmt.Sprintf("%.1f", headless),
			Severity: "high",
		})
	}

	// 检查自动化工具标记
	toolMarker := obs.Features.Get(core.FeatureToolMarker)
	if toolMarker > 0 {
		mr.Contradictions = append(mr.Contradictions, Contradiction{
			Field:    "tool_marker",
			Expected: "0 (no tool markers)",
			Actual:   fmt.Sprintf("%.1f", toolMarker),
			Severity: "high",
		})
	}
}

// ===================================================================
// 可疑度计算
// ===================================================================

var severityWeights = map[string]float64{
	"low":    0.05,
	"medium": 0.15,
	"high":   0.30,
}

func (ad *AnomalyDetector) computeSuspicionScore(contradictions []Contradiction) float64 {
	if len(contradictions) == 0 {
		return 0
	}

	score := 0.0
	for _, c := range contradictions {
		w, ok := severityWeights[c.Severity]
		if !ok {
			w = 0.1
		}
		score += w
	}

	// 多重矛盾有叠加效应，但上限为 1.0
	if score > 1.0 {
		score = 1.0
	}
	return score
}
