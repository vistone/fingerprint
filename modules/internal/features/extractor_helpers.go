package features

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

func (b *BaseFeatureExtractor) extractMobileScreenContradictionFeature(data interface{}, cfg *FeatureConfig) (float64, bool) {
	attrs := toStringMap(data)
	if attrs == nil {
		return 0.0, false
	}

	isMobile := attrs["is_mobile"]
	screenWidth := attrs["screen_width"]

	if isMobile == "" || screenWidth == "" {
		return 0.0, false
	}

	// 解析宽度
	var width int
	_, err := fmt.Sscanf(screenWidth, "%d", &width)
	if err != nil {
		return 0.0, false
	}

	// 移动设备不应使用超大分辨率
	if isMobile == "true" && width > cfg.MobileScreenWidthMax {
		return 0.85, true
	}
	// 桌面设备不应使用超小分辨率
	if isMobile == "false" && width < cfg.DesktopScreenWidthMin {
		return 0.85, true
	}

	return 0.0, false
}

// extractUAFeatureContradictionFeature User-Agent 特性矛盾
func (b *BaseFeatureExtractor) extractUAFeatureContradictionFeature(data interface{}) (float64, bool) {
	attrs := toStringMap(data)
	if attrs == nil {
		return 0.0, false
	}

	ua := attrs["user_agent"]
	features := attrs["features"]

	if ua == "" || features == "" {
		return 0.0, false
	}

	uaLower := strings.ToLower(ua)
	featuresLower := strings.ToLower(features)

	// Chrome 60 不应支持 WebGL2
	if strings.Contains(uaLower, "chrome/60") && strings.Contains(featuresLower, "webgl2") {
		return 0.8, true
	}
	// 移动版 UA 不应声称桌面特性
	if strings.Contains(uaLower, "mobile") && strings.Contains(featuresLower, "desktop") {
		return 0.8, true
	}

	return 0.0, false
}

// extractHeadlessBrowserFeature 无头浏览器检测
func (b *BaseFeatureExtractor) extractHeadlessBrowserFeature(data interface{}, cfg *FeatureConfig) (float64, bool) {
	var ua string

	switch v := data.(type) {
	case string:
		ua = v
	case map[string]string:
		ua = v["user_agent"]
	default:
		return 0.0, false
	}

	if ua == "" {
		return 0.0, false
	}

	uaLower := strings.ToLower(ua)
	for _, marker := range cfg.HeadlessMarkers {
		if strings.Contains(uaLower, strings.ToLower(marker)) {
			return 0.95, true
		}
	}

	return 0.0, false
}

// memEquals 比较两个字节切片是否相等
func memEquals(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// toStringMap 将 interface{} 转换为 map[string]string
// 支持 map[string]string 和 map[string]interface{} 类型
func toStringMap(data interface{}) map[string]string {
	switch v := data.(type) {
	case map[string]string:
		return v
	case map[string]interface{}:
		result := make(map[string]string, len(v))
		for key, val := range v {
			if s, ok := val.(string); ok {
				result[key] = s
			}
		}
		return result
	default:
		return nil
	}
}

// FeatureVector 特征向量
type FeatureVector struct {
	// 各特征的得分（0.0-1.0）
	Scores map[FeatureType]float64
	// MD5 哈希（便于去重）
	Hash string
	// 检测到的异常类型
	Anomalies []FeatureType
	// 总体风险评分（0.0-1.0）
	RiskScore float64
}

// ExtractFeatureVector 从多个数据源提取完整的特征向量
// 注意：此方法并发安全，但 config 参数仅在本次调用中生效
func (b *BaseFeatureExtractor) ExtractFeatureVector(data map[string]interface{}, config *FeatureConfig) *FeatureVector {
	// 使用传入的 config 或默认 config，不修改提取器的状态
	cfg := b.config
	if config != nil {
		cfg = config
	}

	vector := &FeatureVector{
		Scores:    make(map[FeatureType]float64),
		Anomalies: []FeatureType{},
	}

	// 定义要提取的特征列表
	featuresToExtract := []FeatureType{
		FeatureEntropy,
		FeatureToolMarker,
		FeatureHeadlessBrowser,
		FeatureOSPlatformContradiction,
		FeatureUAOSContradiction,
		FeatureMobileScreenContradiction,
		FeatureUAFeatureContradiction,
	}

	// 提取所有特征
	for _, fType := range featuresToExtract {
		score, isAnomaly := b.ExtractFeature(fType, data, cfg)
		vector.Scores[fType] = score
		if isAnomaly {
			vector.Anomalies = append(vector.Anomalies, fType)
		}
	}

	// 计算总体风险评分
	vector.RiskScore = calculateRiskScore(vector.Scores)

	// 计算特征向量的 MD5 哈希（用于去重）
	var hashInput strings.Builder
	hashInput.Grow(len(featuresToExtract) * 20)
	for _, fType := range featuresToExtract {
		fmt.Fprintf(&hashInput, "%s:%.2f;", fType, vector.Scores[fType])
	}
	h := md5.Sum([]byte(hashInput.String()))
	vector.Hash = hex.EncodeToString(h[:])

	return vector
}

// calculateRiskScore 计算综合风险评分
func calculateRiskScore(scores map[FeatureType]float64) float64 {
	if len(scores) == 0 {
		return 0.0
	}

	// 最大值法：取最高风险特征分数
	maxScore := 0.0
	for _, score := range scores {
		if score > maxScore {
			maxScore = score
		}
	}

	// 结合异常数量加权
	anomalyCount := 0
	for _, score := range scores {
		if score > 0.5 {
			anomalyCount++
		}
	}

	// 多重异常会增加风险
	weightedScore := maxScore * (1.0 + float64(anomalyCount-1)*0.1)
	if weightedScore > 1.0 {
		weightedScore = 1.0
	}

	return weightedScore
}
