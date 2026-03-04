// Package ml 提供三层分层分类器实现
package ml

import (
	"sync"

	"github.com/vistone/fingerprint/modules/core"
)

// HierarchicalClassifier 三层分层分类器
// 移植自 Rust 版本的 ML 架构
// Layer 1: 协议类型分类器 (TLS/HTTP/QUIC)
// Layer 2: 浏览器家族分类器 (Chrome/Firefox/Safari)
// Layer 3: 版本识别分类器 (Chrome 120/121/122)
type HierarchicalClassifier struct {
	// Layer 1: 协议类型分类器
	protocolClassifier *ProtocolClassifier

	// Layer 2: 浏览器家族分类器，按协议类型索引
	familyClassifiers map[core.ProtocolType]*FamilyClassifier

	// Layer 3: 版本识别分类器，按浏览器家族索引
	versionClassifiers map[core.BrowserType]*VersionClassifier

	// 训练状态
	trained bool
	mu      sync.RWMutex
}

// NewHierarchicalClassifier 创建新的三层分层分类器
func NewHierarchicalClassifier() *HierarchicalClassifier {
	return &HierarchicalClassifier{
		protocolClassifier: NewProtocolClassifier(),
		familyClassifiers:  make(map[core.ProtocolType]*FamilyClassifier),
		versionClassifiers: make(map[core.BrowserType]*VersionClassifier),
		trained:            false,
	}
}

// Initialize 初始化分类器层次结构
func (hc *HierarchicalClassifier) Initialize() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	// 初始化 Layer 2: 为每种协议类型创建浏览器家族分类器
	protocols := []core.ProtocolType{
		core.ProtocolTLS,
		core.ProtocolHTTP,
		core.ProtocolHTTP2,
		core.ProtocolQUIC,
		core.ProtocolHTTP3,
	}

	for _, proto := range protocols {
		hc.familyClassifiers[proto] = NewFamilyClassifier(proto)
	}

	// 初始化 Layer 3: 为每种浏览器家族创建版本分类器
	families := []core.BrowserType{
		core.BrowserChrome,
		core.BrowserFirefox,
		core.BrowserSafari,
		core.BrowserOpera,
		core.BrowserEdge,
	}

	for _, family := range families {
		hc.versionClassifiers[family] = NewVersionClassifier(family)
	}
}

// TrainingData 训练数据结构
type TrainingData struct {
	ProtocolFeatures [][]float64
	ProtocolLabels   []core.ProtocolType

	FamilyFeatures map[core.ProtocolType][][]float64
	FamilyLabels   map[core.ProtocolType][]core.BrowserType

	VersionFeatures map[core.BrowserType][][]float64
	VersionLabels   map[core.BrowserType][]string
}

// Train 训练三层分层分类器
func (hc *HierarchicalClassifier) Train(data *TrainingData) error {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	// 训练 Layer 1: 协议分类器
	if err := hc.protocolClassifier.Train(data.ProtocolFeatures, data.ProtocolLabels); err != nil {
		return err
	}

	// 训练 Layer 2: 浏览器家族分类器
	for proto, features := range data.FamilyFeatures {
		if classifier, ok := hc.familyClassifiers[proto]; ok {
			labels := data.FamilyLabels[proto]
			if err := classifier.Train(features, labels); err != nil {
				return err
			}
		}
	}

	// 训练 Layer 3: 版本分类器
	for family, features := range data.VersionFeatures {
		if classifier, ok := hc.versionClassifiers[family]; ok {
			labels := data.VersionLabels[family]
			if err := classifier.Train(features, labels); err != nil {
				return err
			}
		}
	}

	hc.trained = true
	return nil
}

// ClassificationResult 分类结果
type ClassificationResult struct {
	Protocol   core.ProtocolType
	Family     core.BrowserType
	Version    string
	Confidence float64
	Labels     map[string]string
	LayerScores
}

// LayerScores 各层分数
type LayerScores struct {
	ProtocolConfidence float64
	FamilyConfidence   float64
	VersionConfidence  float64
}

// Classify 执行分层分类
func (hc *HierarchicalClassifier) Classify(features *core.FeatureVector) *ClassificationResult {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	if !hc.trained {
		return &ClassificationResult{
			Labels: map[string]string{"error": "classifier not trained"},
		}
	}

	result := &ClassificationResult{
		Labels: make(map[string]string),
	}

	// Layer 1: 协议类型分类
	protocolFeatures := hc.extractProtocolFeatures(features)
	protocol, protocolConf := hc.protocolClassifier.Predict(protocolFeatures)
	result.Protocol = protocol
	result.ProtocolConfidence = protocolConf
	result.Labels["protocol"] = string(protocol)

	// Layer 2: 浏览器家族分类
	familyClassifier, ok := hc.familyClassifiers[protocol]
	if !ok {
		result.Confidence = protocolConf
		return result
	}

	familyFeatures := hc.extractFamilyFeatures(features)
	family, familyConf := familyClassifier.Predict(familyFeatures)
	result.Family = family
	result.FamilyConfidence = familyConf
	result.Labels["family"] = string(family)

	// Layer 3: 版本识别
	versionClassifier, ok := hc.versionClassifiers[family]
	if !ok {
		result.Confidence = familyConf * protocolConf
		return result
	}

	versionFeatures := hc.extractVersionFeatures(features)
	version, versionConf := versionClassifier.Predict(versionFeatures)
	result.Version = version
	result.VersionConfidence = versionConf
	result.Labels["version"] = version

	// 综合置信度
	result.Confidence = protocolConf * familyConf * versionConf

	return result
}

// ClassifyBatch 批量分类
func (hc *HierarchicalClassifier) ClassifyBatch(features []*core.FeatureVector) []*ClassificationResult {
	results := make([]*ClassificationResult, len(features))
	for i, fv := range features {
		results[i] = hc.Classify(fv)
	}
	return results
}

// extractProtocolFeatures 提取协议相关特征
func (hc *HierarchicalClassifier) extractProtocolFeatures(fv *core.FeatureVector) []float64 {
	// 提取协议识别相关的特征
	features := []float64{
		fv.Get(core.FeatureTLSVersion),
		fv.Get(core.FeatureCipherSuites),
		fv.Get(core.FeatureExtensions),
		fv.Get(core.FeatureHTTP2Settings),
		fv.Get(core.FeatureHTTPHeaders),
		fv.Get(core.FeatureEntropy),
	}
	return features
}

// extractFamilyFeatures 提取浏览器家族相关特征
func (hc *HierarchicalClassifier) extractFamilyFeatures(fv *core.FeatureVector) []float64 {
	// 提取浏览器识别相关的特征
	features := []float64{
		fv.Get(core.FeatureUserAgent),
		fv.Get(core.FeatureHTTPHeaders),
		fv.Get(core.FeatureCipherSuites),
		fv.Get(core.FeatureExtensions),
		fv.Get(core.FeatureHTTP2Settings),
		fv.Get(core.FeatureCanvas),
		fv.Get(core.FeatureWebGL),
		fv.Get(core.FeatureBehaviorPattern),
	}
	return features
}

// extractVersionFeatures 提取版本识别相关特征
func (hc *HierarchicalClassifier) extractVersionFeatures(fv *core.FeatureVector) []float64 {
	// 提取版本识别相关的特征
	features := []float64{
		fv.Get(core.FeatureTLSVersion),
		fv.Get(core.FeatureCipherSuites),
		fv.Get(core.FeatureExtensions),
		fv.Get(core.FeatureHTTP2Settings),
		fv.Get(core.FeatureHTTPHeaders),
		fv.Get(core.FeatureUserAgent),
		fv.Get(core.FeatureCanvas),
		fv.Get(core.FeatureWebGL),
		fv.Get(core.FeatureFonts),
		fv.Get(core.FeatureAudio),
	}
	return features
}

// GetConfidenceThresholds 获取各层置信度阈值
func (hc *HierarchicalClassifier) GetConfidenceThresholds() (protocol, family, version float64) {
	return 0.7, 0.8, 0.6
}

// IsHighConfidence 判断是否高置信度分类
func (r *ClassificationResult) IsHighConfidence() bool {
	return r.Confidence >= 0.8 &&
		r.ProtocolConfidence >= 0.7 &&
		r.FamilyConfidence >= 0.8 &&
		r.VersionConfidence >= 0.6
}

// FingerprintMatcher 指纹匹配器
type FingerprintMatcher struct {
	classifier *HierarchicalClassifier
}

// NewFingerprintMatcher 创建新的指纹匹配器
func NewFingerprintMatcher() *FingerprintMatcher {
	return &FingerprintMatcher{
		classifier: NewHierarchicalClassifier(),
	}
}

// Initialize 初始化匹配器
func (fm *FingerprintMatcher) Initialize() {
	fm.classifier.Initialize()
}

// Train 训练匹配器
func (fm *FingerprintMatcher) Train(data *TrainingData) error {
	return fm.classifier.Train(data)
}

// Match 匹配指纹
func (fm *FingerprintMatcher) Match(features *core.FeatureVector) *ClassificationResult {
	return fm.classifier.Classify(features)
}

// MatchWithProfile 与已知配置进行匹配
func (fm *FingerprintMatcher) MatchWithProfile(features *core.FeatureVector, profiles []core.FingerprintSpec) *ClassificationResult {
	result := fm.classifier.Classify(features)

	// 在已知配置中查找最接近的匹配
	bestMatch := ""
	bestScore := 0.0

	for _, profile := range profiles {
		score := fm.calculateMatchScore(result, profile)
		if score > bestScore {
			bestScore = score
			bestMatch = profile.GetID()
		}
	}

	result.Labels["best_match"] = bestMatch
	result.Labels["match_score"] = ""

	return result
}

// calculateMatchScore 计算匹配分数
func (fm *FingerprintMatcher) calculateMatchScore(result *ClassificationResult, profile core.FingerprintSpec) float64 {
	score := 0.0
	if result.Family == profile.GetBrowserType() {
		score += 0.5
	}
	if result.Labels["best_match"] == profile.GetID() {
		score += 0.5
	}
	return score
}
