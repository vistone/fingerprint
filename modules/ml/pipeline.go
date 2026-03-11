// Package ml — pipeline.go 提供端到端模型推理与训练管线。
//
// ┌─────────────────────────────────────────────────────────────────────┐
// │                  ModelPipeline — 推理编排                            │
// │                                                                     │
// │  原始请求                                                           │
// │    ↓ EncodeFingerprint()                                            │
// │  30维特征向量                                                       │
// │    ↓ FingerprintEncoder.EncodeSingle()                              │
// │  32维嵌入                                                           │
// │    ├→ BrowserClassifier.ClassifySingle()  → 浏览器识别             │
// │    │                                                                │
// │    │  30维特征 + 10维跨层特征                                       │
// │    ├→ ForgeryDetector.DetectSingle()      → 伪造检测               │
// │    │                                                                │
// │    │  32维嵌入 + 伪造结果 + 8维行为特征                             │
// │    └→ ThreatAssessor.AssessSingle()       → 威胁评估 + 动作建议    │
// │                                                                     │
// │  所有模型结果聚合为 PipelineResult                                  │
// └─────────────────────────────────────────────────────────────────────┘
//
// ┌─────────────────────────────────────────────────────────────────────┐
// │                  Trainer — 训练管线                                  │
// │                                                                     │
// │  207个浏览器配置文件 → 数据增强 → 训练样本                          │
// │                                                                     │
// │  阶段1：编码器预训练（三元组损失）                                  │
// │    - 正样本对：同浏览器 + 高斯噪声                                  │
// │    - 负样本：不同浏览器家族                                         │
// │                                                                     │
// │  阶段2：分类器训练（交叉熵）                                        │
// │    - 冻结编码器，训练分类头                                         │
// │                                                                     │
// │  阶段3：伪造检测器训练（二元交叉熵）                                │
// │    - 真实样本：来自配置文件                                         │
// │    - 伪造样本：跨浏览器层混合、添加无头特征                         │
// │                                                                     │
// │  阶段4：威胁评估器训练（交叉熵）                                    │
// │    - 初始标签：规则引擎生成                                         │
// │    - 后续：在线反馈微调                                             │
// └─────────────────────────────────────────────────────────────────────┘
package ml

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"sync"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// =========================================================================
// PipelineResult — 推理管线输出
// =========================================================================

// PipelineResult 聚合所有模型的推理结果。
type PipelineResult struct {
	// 指纹嵌入 (32维)
	Embedding []float64

	// 浏览器识别结果
	Browser BrowserPrediction

	// 伪造检测结果
	Forgery ForgeryResult

	// 威胁评估结果
	Threat ThreatPrediction

	// 原始特征（供调试/解释）
	RawFeatures   []float64
	CrossFeatures []float64
}

// ModelPipeline 端到端推理管线，串联四个模型。
type ModelPipeline struct {
	encoder    *FingerprintEncoder
	classifier *BrowserClassifier
	detector   *ForgeryDetector
	assessor   *ThreatAssessor
	trained    bool
	mu         sync.RWMutex
}

// Trained 返回管线是否已训练或已加载权重。
func (p *ModelPipeline) Trained() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.trained
}

// NewModelPipeline 创建新的推理管线。
func NewModelPipeline() *ModelPipeline {
	return &ModelPipeline{
		encoder:    NewFingerprintEncoder(),
		classifier: NewBrowserClassifier(),
		detector:   NewForgeryDetector(),
		assessor:   NewThreatAssessor(),
	}
}

// Infer 从 ClientProfile 执行完整推理链。
func (p *ModelPipeline) Infer(profile *profiles.ClientProfile, behavior []float64) *PipelineResult {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 1. 特征编码
	features := EncodeFingerprint(profile)
	crossFeatures := ComputeCrossLayerFeatures(features)

	// 2. 指纹嵌入
	embedding := p.encoder.EncodeSingle(features)

	// 3. 浏览器分类
	browser := p.classifier.ClassifySingle(embedding)

	// 4. 伪造检测
	forgery := p.detector.DetectSingle(features, crossFeatures)

	// 5. 威胁评估
	threat := p.assessor.AssessSingle(embedding, &forgery, behavior)

	return &PipelineResult{
		Embedding:     embedding,
		Browser:       browser,
		Forgery:       forgery,
		Threat:        threat,
		RawFeatures:   features,
		CrossFeatures: crossFeatures,
	}
}

// InferFromFeatures 从已提取的特征向量执行推理链。
func (p *ModelPipeline) InferFromFeatures(fv *core.FeatureVector, behavior []float64) *PipelineResult {
	p.mu.RLock()
	defer p.mu.RUnlock()

	features := EncodeFingerprintFromFeatureVector(fv)
	crossFeatures := ComputeCrossLayerFeatures(features)
	embedding := p.encoder.EncodeSingle(features)
	browser := p.classifier.ClassifySingle(embedding)
	forgery := p.detector.DetectSingle(features, crossFeatures)
	threat := p.assessor.AssessSingle(embedding, &forgery, behavior)

	return &PipelineResult{
		Embedding:     embedding,
		Browser:       browser,
		Forgery:       forgery,
		Threat:        threat,
		RawFeatures:   features,
		CrossFeatures: crossFeatures,
	}
}

// InferBatch 批量推理，利用矩阵运算加速。
func (p *ModelPipeline) InferBatch(profs []*profiles.ClientProfile, behaviors [][]float64) []*PipelineResult {
	p.mu.RLock()
	defer p.mu.RUnlock()

	n := len(profs)
	results := make([]*PipelineResult, n)

	// 1. 批量特征编码
	featureBatch := make([]float64, 0, n*FingerprintFeatureDim)
	crossBatch := make([]float64, 0, n*CrossLayerFeatureDim)
	for _, prof := range profs {
		fp := EncodeFingerprint(prof)
		featureBatch = append(featureBatch, fp...)
		cross := ComputeCrossLayerFeatures(fp)
		crossBatch = append(crossBatch, cross...)
	}

	fpTensor := NewTensor([]int{n, FingerprintFeatureDim}, featureBatch)

	// 2. 批量嵌入
	embTensor := p.encoder.Encode(fpTensor)

	// 3. 批量分类
	browserPreds := p.classifier.Classify(embTensor)

	// 4. 批量伪造检测
	detInput := make([]float64, 0, n*(FingerprintFeatureDim+CrossLayerFeatureDim))
	for i := 0; i < n; i++ {
		fpStart := i * FingerprintFeatureDim
		crossStart := i * CrossLayerFeatureDim
		detInput = append(detInput, featureBatch[fpStart:fpStart+FingerprintFeatureDim]...)
		detInput = append(detInput, crossBatch[crossStart:crossStart+CrossLayerFeatureDim]...)
	}
	forgeryPreds := p.detector.Detect(NewTensor([]int{n, FingerprintFeatureDim + CrossLayerFeatureDim}, detInput))

	// 5. 批量威胁评估
	threatInputDim := EmbeddingDim + 1 + NumForgeryTypes + BehaviorFeatureDim
	threatData := make([]float64, n*threatInputDim)
	for i := 0; i < n; i++ {
		off := i * threatInputDim
		// 嵌入
		embStart := i * EmbeddingDim
		copy(threatData[off:], embTensor.Data[embStart:embStart+EmbeddingDim])
		off += EmbeddingDim
		// 伪造结果
		threatData[off] = forgeryPreds[i].ForgeryProb
		off++
		copy(threatData[off:off+NumForgeryTypes], forgeryPreds[i].TypeProbs)
		off += NumForgeryTypes
		// 行为特征
		if i < len(behaviors) && len(behaviors[i]) >= BehaviorFeatureDim {
			copy(threatData[off:off+BehaviorFeatureDim], behaviors[i][:BehaviorFeatureDim])
		}
	}
	threatPreds := p.assessor.Assess(NewTensor([]int{n, threatInputDim}, threatData))

	// 组装结果
	for i := 0; i < n; i++ {
		fpStart := i * FingerprintFeatureDim
		crossStart := i * CrossLayerFeatureDim
		embStart := i * EmbeddingDim
		results[i] = &PipelineResult{
			Embedding:     append([]float64{}, embTensor.Data[embStart:embStart+EmbeddingDim]...),
			Browser:       browserPreds[i],
			Forgery:       forgeryPreds[i],
			Threat:        threatPreds[i],
			RawFeatures:   append([]float64{}, featureBatch[fpStart:fpStart+FingerprintFeatureDim]...),
			CrossFeatures: append([]float64{}, crossBatch[crossStart:crossStart+CrossLayerFeatureDim]...),
		}
	}
	return results
}

// SetTraining 切换所有模型到训练模式或推理模式。
func (p *ModelPipeline) SetTraining(training bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.encoder.Net.SetTraining(training)
	p.classifier.Net.SetTraining(training)
	p.detector.DetectorNet.SetTraining(training)
	p.detector.TypeNet.SetTraining(training)
	p.assessor.ThreatNet.SetTraining(training)
	p.assessor.ActionNet.SetTraining(training)
}

// =========================================================================
// Trainer — 从配置文件训练所有模型
// =========================================================================

// NeuralTrainerConfig 神经网络训练配置。
type NeuralTrainerConfig struct {
	Epochs          int     // 训练轮数
	BatchSize       int     // 小批量大小
	LearningRate    float64 // Adam 学习率
	AugmentNoise    float64 // 数据增强噪声标准差
	TripletMargin   float64 // 三元组损失边界
	ForgeryRatio    float64 // 伪造样本与真实样本的比例
	ValidationSplit float64 // 验证集比例
}

// DefaultNeuralTrainerConfig 默认训练配置。
var DefaultNeuralTrainerConfig = &NeuralTrainerConfig{
	Epochs:          50,
	BatchSize:       16,
	LearningRate:    0.001,
	AugmentNoise:    0.02,
	TripletMargin:   1.0,
	ForgeryRatio:    1.0,
	ValidationSplit: 0.2,
}

// TrainingMetrics 训练指标。
type TrainingMetrics struct {
	Epoch         int
	EncoderLoss   float64
	ClassLoss     float64
	ForgeryLoss   float64
	ThreatLoss    float64
	ValAccuracy   float64
	ForgeryAUC    float64
}

// NeuralTrainer 神经网络训练管线。
type NeuralTrainer struct {
	Pipeline *ModelPipeline
	Config   *NeuralTrainerConfig
	Metrics  []TrainingMetrics
}

// NewNeuralTrainer 创建新的神经网络训练管线。
func NewNeuralTrainer(pipeline *ModelPipeline, config *NeuralTrainerConfig) *NeuralTrainer {
	if config == nil {
		config = DefaultNeuralTrainerConfig
	}
	return &NeuralTrainer{
		Pipeline: pipeline,
		Config:   config,
	}
}

// TrainFromProfiles 从浏览器配置文件训练所有模型。
// 这是主训练入口：加载 207 个配置文件 → 数据增强 → 多阶段训练。
func (t *NeuralTrainer) TrainFromProfiles(registry *profiles.ProfileRegistry) error {
	allProfiles := registry.GetAll()
	if len(allProfiles) == 0 {
		return fmt.Errorf("no profiles available for training")
	}

	// 构建训练数据
	trainSet, valSet := t.buildTrainingData(allProfiles)

	// 阶段 1: 编码器预训练 (三元组损失)
	if err := t.trainEncoder(trainSet); err != nil {
		return fmt.Errorf("encoder training failed: %w", err)
	}

	// 阶段 2: 浏览器分类器训练
	if err := t.trainClassifier(trainSet, valSet); err != nil {
		return fmt.Errorf("classifier training failed: %w", err)
	}

	// 阶段 3: 伪造检测器训练
	if err := t.trainForgeryDetector(trainSet); err != nil {
		return fmt.Errorf("forgery detector training failed: %w", err)
	}

	// 阶段 4: 威胁评估器训练
	if err := t.trainThreatAssessor(trainSet); err != nil {
		return fmt.Errorf("threat assessor training failed: %w", err)
	}

	t.Pipeline.mu.Lock()
	t.Pipeline.trained = true
	t.Pipeline.mu.Unlock()

	return nil
}

// profileSample 内部训练样本。
type profileSample struct {
	Features    []float64
	FamilyLabel int
	ProfileID   string
	BrowserType core.BrowserType
}

// buildTrainingData 将配置文件转换为训练样本，并按比例划分训练/验证集。
func (t *NeuralTrainer) buildTrainingData(allProfiles []profiles.ClientProfile) (train, val []profileSample) {
	// 按浏览器家族分组
	familyMap := make(map[core.BrowserType][]profiles.ClientProfile)
	for _, p := range allProfiles {
		familyMap[p.BrowserType] = append(familyMap[p.BrowserType], p)
	}

	familyLabels := browserFamilyLabelMap()

	var samples []profileSample
	for _, p := range allProfiles {
		features := EncodeFingerprint(&p)
		label := familyLabels[p.BrowserType]
		samples = append(samples, profileSample{
			Features:    features,
			FamilyLabel: label,
			ProfileID:   p.ID,
			BrowserType: p.BrowserType,
		})
		// 数据增强: 添加高斯噪声生成变体
		for aug := 0; aug < 3; aug++ {
			augFeatures := make([]float64, len(features))
			copy(augFeatures, features)
			for i := range augFeatures {
				augFeatures[i] += rand.NormFloat64() * t.Config.AugmentNoise
				augFeatures[i] = math.Max(0, math.Min(1, augFeatures[i]))
			}
			samples = append(samples, profileSample{
				Features:    augFeatures,
				FamilyLabel: label,
				ProfileID:   p.ID,
				BrowserType: p.BrowserType,
			})
		}
	}

	// 打乱并划分
	rand.Shuffle(len(samples), func(i, j int) { samples[i], samples[j] = samples[j], samples[i] })
	splitIdx := int(float64(len(samples)) * (1.0 - t.Config.ValidationSplit))
	return samples[:splitIdx], samples[splitIdx:]
}

// browserFamilyLabelMap 返回浏览器类型到标签索引的映射。
func browserFamilyLabelMap() map[core.BrowserType]int {
	return map[core.BrowserType]int{
		core.BrowserChrome:  0,
		core.BrowserFirefox: 1,
		core.BrowserSafari:  2,
		core.BrowserEdge:    3,
		core.BrowserOpera:   4,
		core.BrowserBrave:   5,
		core.BrowserSamsung: 6,
	}
}

// =========================================================================
// 阶段 1: 编码器训练 — 三元组损失
// =========================================================================

func (t *NeuralTrainer) trainEncoder(samples []profileSample) error {
	cfg := t.Config
	enc := t.Pipeline.encoder
	enc.Net.SetTraining(true)
	defer enc.Net.SetTraining(false)

	optimizer := NewAdamOptimizer(enc.Net.Params(), cfg.LearningRate)
	tripletMargin := cfg.TripletMargin

	// 按家族分组索引
	familyIdx := make(map[int][]int)
	for i, s := range samples {
		familyIdx[s.FamilyLabel] = append(familyIdx[s.FamilyLabel], i)
	}
	families := make([]int, 0, len(familyIdx))
	for f := range familyIdx {
		families = append(families, f)
	}
	sort.Ints(families)

	for epoch := 0; epoch < cfg.Epochs; epoch++ {
		totalLoss := 0.0
		count := 0

		// 生成三元组: anchor, positive (同家族), negative (不同家族)
		for _, fam := range families {
			indices := familyIdx[fam]
			if len(indices) < 2 {
				continue
			}
			// 选择不同家族作为负样本源
			var negFam int
			for {
				negFam = families[rand.Intn(len(families))]
				if negFam != fam && len(familyIdx[negFam]) > 0 {
					break
				}
				if len(families) <= 1 {
					negFam = fam
					break
				}
			}
			negIndices := familyIdx[negFam]

			batchSize := min(cfg.BatchSize, len(indices)/2)
			if batchSize < 1 {
				batchSize = 1
			}

			// 为当前家族生成 batchSize 个三元组
			anchors := make([]float64, 0, batchSize*FingerprintFeatureDim)
			positives := make([]float64, 0, batchSize*FingerprintFeatureDim)
			negatives := make([]float64, 0, batchSize*FingerprintFeatureDim)

			for b := 0; b < batchSize; b++ {
				aIdx := indices[rand.Intn(len(indices))]
				pIdx := indices[rand.Intn(len(indices))]
				nIdx := negIndices[rand.Intn(len(negIndices))]
				anchors = append(anchors, samples[aIdx].Features...)
				positives = append(positives, samples[pIdx].Features...)
				negatives = append(negatives, samples[nIdx].Features...)
			}

			anchorT := NewTensor([]int{batchSize, FingerprintFeatureDim}, anchors)
			posT := NewTensor([]int{batchSize, FingerprintFeatureDim}, positives)
			negT := NewTensor([]int{batchSize, FingerprintFeatureDim}, negatives)

			enc.Net.ZeroGrad()
			anchorEmb := enc.Encode(anchorT)
			posEmb := enc.Encode(posT)
			negEmb := enc.Encode(negT)

			loss, anchorGrad, _, _ := TripletMarginLoss(anchorEmb, posEmb, negEmb, tripletMargin)

			// 反向传播
			enc.Net.Backward(anchorGrad)
			optimizer.Step()

			totalLoss += loss
			count++
		}

		if count > 0 {
			t.recordMetric(epoch, totalLoss/float64(count), 0, 0, 0, 0, 0)
		}
	}
	return nil
}

// =========================================================================
// 阶段 2: 分类器训练 — 交叉熵
// =========================================================================

func (t *NeuralTrainer) trainClassifier(trainSet, valSet []profileSample) error {
	cfg := t.Config
	enc := t.Pipeline.encoder
	cls := t.Pipeline.classifier
	cls.Net.SetTraining(true)
	defer cls.Net.SetTraining(false)

	optimizer := NewAdamOptimizer(cls.Net.Params(), cfg.LearningRate)

	for epoch := 0; epoch < cfg.Epochs; epoch++ {
		totalLoss := 0.0
		count := 0

		// Mini-batch 训练
		for start := 0; start < len(trainSet); start += cfg.BatchSize {
			end := start + cfg.BatchSize
			if end > len(trainSet) {
				end = len(trainSet)
			}
			batchSamples := trainSet[start:end]
			batchN := len(batchSamples)

			// 编码 → 嵌入 (编码器已冻结)
			fpData := make([]float64, 0, batchN*FingerprintFeatureDim)
			targets := make([]int, batchN)
			for i, s := range batchSamples {
				fpData = append(fpData, s.Features...)
				targets[i] = s.FamilyLabel
			}

			embeddings := enc.Encode(NewTensor([]int{batchN, FingerprintFeatureDim}, fpData))

			// 前向传播
			cls.Net.ZeroGrad()
			output := cls.Net.Forward(embeddings)

			// 计算损失
			lossVal, grad := CrossEntropyLoss(output, targets)

			// 反向传播
			cls.Net.Backward(grad)
			optimizer.Step()

			totalLoss += lossVal
			count++
		}

		// 验证精度
		valAcc := t.evaluateClassifier(enc, cls, valSet)

		if count > 0 {
			t.recordMetric(epoch, 0, totalLoss/float64(count), 0, 0, valAcc, 0)
		}
	}
	return nil
}

func (t *NeuralTrainer) evaluateClassifier(enc *FingerprintEncoder, cls *BrowserClassifier, valSet []profileSample) float64 {
	if len(valSet) == 0 {
		return 0
	}
	correct := 0
	familyLabels := browserFamilyLabelMap()
	for _, s := range valSet {
		emb := enc.EncodeSingle(s.Features)
		pred := cls.ClassifySingle(emb)
		if familyLabels[pred.Family] == s.FamilyLabel {
			correct++
		}
	}
	return float64(correct) / float64(len(valSet))
}

// =========================================================================
// 阶段 3: 伪造检测器训练
// =========================================================================

func (t *NeuralTrainer) trainForgeryDetector(realSamples []profileSample) error {
	cfg := t.Config
	det := t.Pipeline.detector
	det.DetectorNet.SetTraining(true)
	det.TypeNet.SetTraining(true)
	defer det.DetectorNet.SetTraining(false)
	defer det.TypeNet.SetTraining(false)

	allParams := append(det.DetectorNet.Params(), det.TypeNet.Params()...)
	optimizer := NewAdamOptimizer(allParams, cfg.LearningRate)

	inputDim := FingerprintFeatureDim + CrossLayerFeatureDim

	for epoch := 0; epoch < cfg.Epochs; epoch++ {
		totalLoss := 0.0
		count := 0

		for start := 0; start < len(realSamples); start += cfg.BatchSize {
			end := start + cfg.BatchSize
			if end > len(realSamples) {
				end = len(realSamples)
			}
			batchReal := realSamples[start:end]
			batchN := len(batchReal)

			// 生成伪造样本数量
			numForgery := int(float64(batchN) * cfg.ForgeryRatio)
			if numForgery < 1 {
				numForgery = 1
			}
			totalBatch := batchN + numForgery

			inputData := make([]float64, 0, totalBatch*inputDim)
			targetData := make([]float64, totalBatch)

			// 真实样本 (标签 = 0, 即非伪造)
			for _, s := range batchReal {
				cross := ComputeCrossLayerFeatures(s.Features)
				inputData = append(inputData, s.Features...)
				inputData = append(inputData, cross...)
			}

			// 伪造样本 (标签 = 1)
			for f := 0; f < numForgery; f++ {
				forged := t.generateForgedSample(realSamples)
				cross := ComputeCrossLayerFeatures(forged)
				inputData = append(inputData, forged...)
				inputData = append(inputData, cross...)
				targetData[batchN+f] = 1.0 // 伪造
			}

			input := NewTensor([]int{totalBatch, inputDim}, inputData)

			// 检测网络前向
			det.DetectorNet.ZeroGrad()
			output := det.DetectorNet.Forward(input)

			// 二元交叉熵损失
			lossVal, grad := BinaryCrossEntropyLoss(output, targetData)
			det.DetectorNet.Backward(grad)
			optimizer.Step()

			totalLoss += lossVal
			count++
		}

		if count > 0 {
			t.recordMetric(epoch, 0, 0, totalLoss/float64(count), 0, 0, 0)
		}
	}
	return nil
}

// generateForgedSample 生成伪造指纹样本。
// 策略: 随机混合不同浏览器家族的层特征，制造跨层不一致。
func (t *NeuralTrainer) generateForgedSample(samples []profileSample) []float64 {
	forged := make([]float64, FingerprintFeatureDim)

	// 从两个不同样本混合特征层
	s1 := samples[rand.Intn(len(samples))]
	s2 := samples[rand.Intn(len(samples))]

	// TLS 层来自 s1 (索引 0-7)
	copy(forged[0:8], s1.Features[0:8])
	// HTTP/2 层来自 s2 (索引 8-13，制造 TLS↔HTTP/2 不一致)
	copy(forged[8:14], s2.Features[8:14])
	// TCP/IP 层随机选择
	if rand.Float64() < 0.5 {
		copy(forged[14:18], s1.Features[14:18])
	} else {
		copy(forged[14:18], s2.Features[14:18])
	}
	// JS 层: 添加无头浏览器特征
	forged[25] = rand.Float64()*0.3 + 0.5 // headless_score 偏高

	// 添加噪声
	for i := range forged {
		forged[i] += rand.NormFloat64() * 0.05
		forged[i] = math.Max(0, math.Min(1, forged[i]))
	}
	return forged
}

// =========================================================================
// 阶段 4: 威胁评估器训练
// =========================================================================

func (t *NeuralTrainer) trainThreatAssessor(samples []profileSample) error {
	cfg := t.Config
	enc := t.Pipeline.encoder
	det := t.Pipeline.detector
	assessor := t.Pipeline.assessor
	assessor.ThreatNet.SetTraining(true)
	assessor.ActionNet.SetTraining(true)
	defer assessor.ThreatNet.SetTraining(false)
	defer assessor.ActionNet.SetTraining(false)

	allParams := append(assessor.ThreatNet.Params(), assessor.ActionNet.Params()...)
	optimizer := NewAdamOptimizer(allParams, cfg.LearningRate)

	inputDim := EmbeddingDim + 1 + NumForgeryTypes + BehaviorFeatureDim

	for epoch := 0; epoch < cfg.Epochs; epoch++ {
		totalLoss := 0.0
		count := 0

		for start := 0; start < len(samples); start += cfg.BatchSize {
			end := start + cfg.BatchSize
			if end > len(samples) {
				end = len(samples)
			}
			batchSamples := samples[start:end]
			batchN := len(batchSamples)

			inputData := make([]float64, batchN*inputDim)
			threatTargets := make([]int, batchN)

			for i, s := range batchSamples {
				emb := enc.EncodeSingle(s.Features)
				cross := ComputeCrossLayerFeatures(s.Features)
				forgery := det.DetectSingle(s.Features, cross)

				off := i * inputDim
				copy(inputData[off:], emb)
				off += EmbeddingDim
				inputData[off] = forgery.ForgeryProb
				off++
				copy(inputData[off:off+NumForgeryTypes], forgery.TypeProbs)
				// 行为特征: 训练时使用全零 (推理时由 Agent 填充)

				threatTargets[i] = t.generateThreatLabel(s, &forgery)
			}

			input := NewTensor([]int{batchN, inputDim}, inputData)

			assessor.ThreatNet.ZeroGrad()
			output := assessor.ThreatNet.Forward(input)
			lossVal, grad := CrossEntropyLoss(output, threatTargets)
			assessor.ThreatNet.Backward(grad)
			optimizer.Step()

			totalLoss += lossVal
			count++
		}

		if count > 0 {
			t.recordMetric(epoch, 0, 0, 0, totalLoss/float64(count), 0, 0)
		}
	}
	return nil
}

// generateThreatLabel 根据样本特征和伪造检测结果生成规则标签。
func (t *NeuralTrainer) generateThreatLabel(s profileSample, forgery *ForgeryResult) int {
	if forgery.ForgeryProb > 0.7 {
		switch forgery.ForgeryType {
		case ForgeryHeadless:
			return int(ThreatBot)
		case ForgeryAntiDetect:
			return int(ThreatFingerprintSpoof)
		case ForgeryProxy:
			return int(ThreatEvasion)
		}
		return int(ThreatFingerprintSpoof)
	}
	return int(ThreatNone)
}

func (t *NeuralTrainer) recordMetric(epoch int, encLoss, clsLoss, forLoss, thrLoss, valAcc, forAUC float64) {
	// 如果同一 epoch 已有记录，合并
	for i := range t.Metrics {
		if t.Metrics[i].Epoch == epoch {
			if encLoss > 0 {
				t.Metrics[i].EncoderLoss = encLoss
			}
			if clsLoss > 0 {
				t.Metrics[i].ClassLoss = clsLoss
			}
			if forLoss > 0 {
				t.Metrics[i].ForgeryLoss = forLoss
			}
			if thrLoss > 0 {
				t.Metrics[i].ThreatLoss = thrLoss
			}
			if valAcc > 0 {
				t.Metrics[i].ValAccuracy = valAcc
			}
			if forAUC > 0 {
				t.Metrics[i].ForgeryAUC = forAUC
			}
			return
		}
	}
	t.Metrics = append(t.Metrics, TrainingMetrics{
		Epoch:       epoch,
		EncoderLoss: encLoss,
		ClassLoss:   clsLoss,
		ForgeryLoss: forLoss,
		ThreatLoss:  thrLoss,
		ValAccuracy: valAcc,
		ForgeryAUC:  forAUC,
	})
}

// =========================================================================
// 模型序列化 — 保存/加载已训练的模型权重
// =========================================================================

// ModelWeights 序列化的模型权重。
type ModelWeights struct {
	Version     string                `json:"version"`
	Encoder     []SerializedParam     `json:"encoder"`
	Classifier  []SerializedParam     `json:"classifier"`
	DetectorNet []SerializedParam     `json:"detector_net"`
	TypeNet     []SerializedParam     `json:"type_net"`
	ThreatNet   []SerializedParam     `json:"threat_net"`
	ActionNet   []SerializedParam     `json:"action_net"`
	Metrics     []TrainingMetrics     `json:"metrics,omitempty"`
}

// SerializedParam 序列化的参数。
type SerializedParam struct {
	Shape []int     `json:"shape"`
	Data  []float64 `json:"data"`
}

// SaveWeights 将模型权重保存到文件。
func (p *ModelPipeline) SaveWeights(path string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	weights := &ModelWeights{
		Version:     "1.0.14",
		Encoder:     serializeParams(p.encoder.Net.Params()),
		Classifier:  serializeParams(p.classifier.Net.Params()),
		DetectorNet: serializeParams(p.detector.DetectorNet.Params()),
		TypeNet:     serializeParams(p.detector.TypeNet.Params()),
		ThreatNet:   serializeParams(p.assessor.ThreatNet.Params()),
		ActionNet:   serializeParams(p.assessor.ActionNet.Params()),
	}

	data, err := json.Marshal(weights)
	if err != nil {
		return fmt.Errorf("marshal weights: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

// LoadWeights 从文件加载模型权重。
func (p *ModelPipeline) LoadWeights(path string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read weights: %w", err)
	}

	var weights ModelWeights
	if err := json.Unmarshal(data, &weights); err != nil {
		return fmt.Errorf("unmarshal weights: %w", err)
	}

	deserializeParams(p.encoder.Net.Params(), weights.Encoder)
	deserializeParams(p.classifier.Net.Params(), weights.Classifier)
	deserializeParams(p.detector.DetectorNet.Params(), weights.DetectorNet)
	deserializeParams(p.detector.TypeNet.Params(), weights.TypeNet)
	deserializeParams(p.assessor.ThreatNet.Params(), weights.ThreatNet)
	deserializeParams(p.assessor.ActionNet.Params(), weights.ActionNet)

	p.trained = true
	return nil
}

func serializeParams(params []*Param) []SerializedParam {
	out := make([]SerializedParam, len(params))
	for i, p := range params {
		out[i] = SerializedParam{
			Shape: append([]int{}, p.Value.Shape...),
			Data:  append([]float64{}, p.Value.Data...),
		}
	}
	return out
}

func deserializeParams(params []*Param, serialized []SerializedParam) {
	for i := range params {
		if i >= len(serialized) {
			break
		}
		s := serialized[i]
		if len(s.Data) == len(params[i].Value.Data) {
			copy(params[i].Value.Data, s.Data)
			params[i].Value.Shape = append([]int{}, s.Shape...)
		}
	}
}

// =========================================================================
// 辅助函数
// =========================================================================


