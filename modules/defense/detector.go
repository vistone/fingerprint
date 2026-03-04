// Package defense 提供安全防护功能
// 包括被动检测、主动防护和风险评估
package defense

import (
	"math"
	"strings"
	"sync"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/ml"
)

// Detector 安全检测器
type Detector struct {
	classifier *ml.HierarchicalClassifier
	extractor  *ml.FeatureExtractor
	rules      []DetectionRule
	mu         sync.RWMutex
}

// DetectionRule 检测规则
type DetectionRule struct {
	Name        string
	Description string
	Condition   func(*core.FeatureVector) bool
	RiskScore   float64
}

// NewDetector 创建新的安全检测器
func NewDetector() *Detector {
	d := &Detector{
		classifier: ml.NewHierarchicalClassifier(),
		extractor:  ml.NewFeatureExtractor(),
		rules:      make([]DetectionRule, 0),
	}
	d.initializeRules()
	return d
}

// initializeRules 初始化检测规则
func (d *Detector) initializeRules() {
	d.rules = []DetectionRule{
		{
			Name:        "headless_browser",
			Description: "检测无头浏览器",
			Condition: func(fv *core.FeatureVector) bool {
				return fv.Get(core.FeatureHeadlessBrowser) > 0.5
			},
			RiskScore: 0.7,
		},
		{
			Name:        "high_entropy",
			Description: "检测异常高熵值",
			Condition: func(fv *core.FeatureVector) bool {
				return fv.Get(core.FeatureEntropy) > 10.0
			},
			RiskScore: 0.5,
		},
		{
			Name:        "automation_tool",
			Description: "检测自动化工具标记",
			Condition: func(fv *core.FeatureVector) bool {
				return fv.Get(core.FeatureToolMarker) > 0.3
			},
			RiskScore: 0.8,
		},
		{
			Name:        "inconsistent_behavior",
			Description: "检测行为不一致",
			Condition: func(fv *core.FeatureVector) bool {
				return fv.Get(core.FeatureBehaviorPattern) < 0.3
			},
			RiskScore: 0.6,
		},
	}
}

// Detect 执行检测
func (d *Detector) Detect(features *core.FeatureVector) *DetectionResult {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := &DetectionResult{
		Findings:    make([]Finding, 0),
		RiskFactors: make([]core.RiskFactor, 0),
		Labels:      make(map[string]string),
	}

	// 运行所有规则
	for _, rule := range d.rules {
		if rule.Condition(features) {
			finding := Finding{
				Rule:        rule.Name,
				Description: rule.Description,
				RiskScore:   rule.RiskScore,
			}
			result.Findings = append(result.Findings, finding)
			result.RiskFactors = append(result.RiskFactors, core.RiskFactor{
				Name:        rule.Name,
				Weight:      rule.RiskScore,
				Description: rule.Description,
			})
			result.TotalRisk += rule.RiskScore
		}
	}

	// ML 分类检测
	if d.classifier != nil {
		classification := d.classifier.Classify(features)
		result.Classification = classification
		result.Labels["protocol"] = string(classification.Protocol)
		result.Labels["family"] = string(classification.Family)
		result.Labels["version"] = classification.Version
	}

	// 计算最终风险等级
	result.RiskLevel = d.calculateRiskLevel(result.TotalRisk)
	result.RiskScore = math.Min(result.TotalRisk, 1.0)

	return result
}

// calculateRiskLevel 计算风险等级
func (d *Detector) calculateRiskLevel(score float64) core.RiskLevel {
	switch {
	case score >= 0.9:
		return core.RiskLevelCritical
	case score >= 0.7:
		return core.RiskLevelHigh
	case score >= 0.4:
		return core.RiskLevelMedium
	case score >= 0.1:
		return core.RiskLevelLow
	default:
		return core.RiskLevelNone
	}
}

// DetectionResult 检测结果
type DetectionResult struct {
	Findings       []Finding
	RiskFactors    []core.RiskFactor
	Classification *ml.ClassificationResult
	RiskScore      float64
	RiskLevel      core.RiskLevel
	TotalRisk      float64
	Labels         map[string]string
}

// Finding 检测发现
type Finding struct {
	Rule        string
	Description string
	RiskScore   float64
}

// IsThreat 是否为威胁
func (r *DetectionResult) IsThreat() bool {
	return r.RiskLevel >= core.RiskLevelMedium
}

// PassiveDetector 被动检测器
type PassiveDetector struct {
	*Detector
	database *FingerprintDatabase
}

// FingerprintDatabase 指纹数据库
type FingerprintDatabase struct {
	knownFingerprints map[string]FingerprintRecord
	mu                sync.RWMutex
}

// FingerprintRecord 指纹记录
type FingerprintRecord struct {
	ID          string
	Features    *core.FeatureVector
	Labels      map[string]string
	RiskScore   float64
	FirstSeen   int64
	LastSeen    int64
	SeenCount   int
}

// NewPassiveDetector 创建新的被动检测器
func NewPassiveDetector() *PassiveDetector {
	return &PassiveDetector{
		Detector: NewDetector(),
		database: &FingerprintDatabase{
			knownFingerprints: make(map[string]FingerprintRecord),
		},
	}
}

// Lookup 查找已知指纹
func (pd *PassiveDetector) Lookup(fingerprintHash string) (FingerprintRecord, bool) {
	pd.database.mu.RLock()
	defer pd.database.mu.RUnlock()

	record, ok := pd.database.knownFingerprints[fingerprintHash]
	return record, ok
}

// Store 存储指纹
func (pd *PassiveDetector) Store(record FingerprintRecord) {
	pd.database.mu.Lock()
	defer pd.database.mu.Unlock()

	pd.database.knownFingerprints[record.ID] = record
}

// ActiveProtector 主动防护器
type ActiveProtector struct {
	noiseGenerators map[string]NoiseGenerator
	mu              sync.RWMutex
}

// NoiseGenerator 噪声生成器接口
type NoiseGenerator interface {
	Generate(seed int64) interface{}
}

// NewActiveProtector 创建新的主动防护器
func NewActiveProtector() *ActiveProtector {
	return &ActiveProtector{
		noiseGenerators: make(map[string]NoiseGenerator),
	}
}

// RegisterNoiseGenerator 注册噪声生成器
func (ap *ActiveProtector) RegisterNoiseGenerator(name string, generator NoiseGenerator) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	ap.noiseGenerators[name] = generator
}

// GenerateNoise 生成噪声
func (ap *ActiveProtector) GenerateNoise(generatorName string, seed int64) interface{} {
	ap.mu.RLock()
	defer ap.mu.RUnlock()

	if gen, ok := ap.noiseGenerators[generatorName]; ok {
		return gen.Generate(seed)
	}
	return nil
}

// ProtectionConfig 防护配置
type ProtectionConfig struct {
	EnableCanvasNoise   bool
	EnableAudioNoise    bool
	EnableWebGLNoise    bool
	EnableTimingNoise   bool
	NoiseLevel          float64
}

// DefaultProtectionConfig 默认防护配置
var DefaultProtectionConfig = &ProtectionConfig{
	EnableCanvasNoise: true,
	EnableAudioNoise:  true,
	EnableWebGLNoise:  true,
	EnableTimingNoise: true,
	NoiseLevel:        0.1,
}

// ApplyProtection 应用防护措施
func (ap *ActiveProtector) ApplyProtection(config *ProtectionConfig) *ProtectionResult {
	result := &ProtectionResult{
		AppliedMeasures: make([]string, 0),
		NoiseData:       make(map[string]interface{}),
	}

	if config.EnableCanvasNoise {
		noise := ap.GenerateNoise("canvas", 0)
		result.NoiseData["canvas"] = noise
		result.AppliedMeasures = append(result.AppliedMeasures, "canvas_noise")
	}

	if config.EnableAudioNoise {
		noise := ap.GenerateNoise("audio", 0)
		result.NoiseData["audio"] = noise
		result.AppliedMeasures = append(result.AppliedMeasures, "audio_noise")
	}

	if config.EnableWebGLNoise {
		noise := ap.GenerateNoise("webgl", 0)
		result.NoiseData["webgl"] = noise
		result.AppliedMeasures = append(result.AppliedMeasures, "webgl_noise")
	}

	if config.EnableTimingNoise {
		noise := ap.GenerateNoise("timing", 0)
		result.NoiseData["timing"] = noise
		result.AppliedMeasures = append(result.AppliedMeasures, "timing_noise")
	}

	return result
}

// ProtectionResult 防护结果
type ProtectionResult struct {
	AppliedMeasures []string
	NoiseData       map[string]interface{}
}

// RiskEngine 风险引擎
type RiskEngine struct {
	passiveDetector *PassiveDetector
	activeProtector *ActiveProtector
	weights         RiskWeights
}

// RiskWeights 风险权重
type RiskWeights struct {
	BehaviorWeight    float64
	FingerprintWeight float64
	ReputationWeight  float64
	AnomalyWeight     float64
}

// DefaultRiskWeights 默认风险权重
var DefaultRiskWeights = RiskWeights{
	BehaviorWeight:    0.3,
	FingerprintWeight: 0.3,
	ReputationWeight:  0.2,
	AnomalyWeight:     0.2,
}

// NewRiskEngine 创建新的风险引擎
func NewRiskEngine() *RiskEngine {
	return &RiskEngine{
		passiveDetector: NewPassiveDetector(),
		activeProtector: NewActiveProtector(),
		weights:         DefaultRiskWeights,
	}
}

// Evaluate 评估风险
func (re *RiskEngine) Evaluate(features *core.FeatureVector, classification *ml.ClassificationResult) *core.RiskAssessment {
	// 被动检测
	detectionResult := re.passiveDetector.Detect(features)

	// 计算综合风险分数
	behaviorScore := detectionResult.RiskScore * re.weights.BehaviorWeight
	fingerprintScore := (1.0 - classification.Confidence) * re.weights.FingerprintWeight

	// 总风险分数
	totalScore := behaviorScore + fingerprintScore

	// 生成建议
	suggestions := re.generateSuggestions(detectionResult, classification)

	return &core.RiskAssessment{
		Score:       totalScore,
		Level:       re.calculateRiskLevel(totalScore),
		Factors:     detectionResult.RiskFactors,
		Suggestions: suggestions,
	}
}

// generateSuggestions 生成防护建议
func (re *RiskEngine) generateSuggestions(detection *DetectionResult, classification *ml.ClassificationResult) []string {
	var suggestions []string

	if detection.RiskScore > 0.7 {
		suggestions = append(suggestions, "启用增强防护模式")
		suggestions = append(suggestions, "增加噪声注入级别")
	}

	if classification.Confidence < 0.5 {
		suggestions = append(suggestions, "请求额外验证")
	}

	for _, finding := range detection.Findings {
		if strings.Contains(finding.Rule, "headless") {
			suggestions = append(suggestions, "检测无头浏览器特征")
		}
		if strings.Contains(finding.Rule, "automation") {
			suggestions = append(suggestions, "验证非人类行为")
		}
	}

	return suggestions
}

// calculateRiskLevel 计算风险等级
func (re *RiskEngine) calculateRiskLevel(score float64) core.RiskLevel {
	switch {
	case score >= 0.9:
		return core.RiskLevelCritical
	case score >= 0.7:
		return core.RiskLevelHigh
	case score >= 0.4:
		return core.RiskLevelMedium
	case score >= 0.1:
		return core.RiskLevelLow
	default:
		return core.RiskLevelNone
	}
}

// DefenseSystem 综合防护系统
type DefenseSystem struct {
	Detector   *Detector
	Protector  *ActiveProtector
	RiskEngine *RiskEngine
}

// NewDefenseSystem 创建新的防护系统
func NewDefenseSystem() *DefenseSystem {
	return &DefenseSystem{
		Detector:   NewDetector(),
		Protector:  NewActiveProtector(),
		RiskEngine: NewRiskEngine(),
	}
}

// Analyze 分析并返回防护建议
func (ds *DefenseSystem) Analyze(features *core.FeatureVector, classification *ml.ClassificationResult) *DefenseAdvice {
	// 检测
	detection := ds.Detector.Detect(features)

	// 风险评估
	risk := ds.RiskEngine.Evaluate(features, classification)

	// 生成防护配置
	protectionConfig := ds.generateProtectionConfig(risk)

	// 应用防护
	protection := ds.Protector.ApplyProtection(protectionConfig)

	return &DefenseAdvice{
		Detection:   detection,
		Risk:        risk,
		Protection:  protection,
		Recommended: protectionConfig,
	}
}

// generateProtectionConfig 根据风险生成防护配置
func (ds *DefenseSystem) generateProtectionConfig(risk *core.RiskAssessment) *ProtectionConfig {
	config := &ProtectionConfig{
		EnableCanvasNoise:   true,
		EnableAudioNoise:    true,
		EnableWebGLNoise:    true,
		EnableTimingNoise:   true,
		NoiseLevel:          0.1,
	}

	// 根据风险等级调整配置
	switch risk.Level {
	case core.RiskLevelCritical:
		config.NoiseLevel = 0.5
		config.EnableTimingNoise = true
	case core.RiskLevelHigh:
		config.NoiseLevel = 0.3
	case core.RiskLevelMedium:
		config.NoiseLevel = 0.2
	case core.RiskLevelLow:
		config.NoiseLevel = 0.1
		config.EnableTimingNoise = false
	default:
		config.EnableCanvasNoise = false
		config.EnableAudioNoise = false
		config.EnableWebGLNoise = false
	}

	return config
}

// DefenseAdvice 防护建议
type DefenseAdvice struct {
	Detection   *DetectionResult
	Risk        *core.RiskAssessment
	Protection  *ProtectionResult
	Recommended *ProtectionConfig
}
