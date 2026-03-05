// Package ml 提供机器学习指纹分类功能
// 实现三层分层分类器架构，移植自 Rust 版本的 ML 设计
package ml

import (
	"math"
	"sort"
	"sync"

	"github.com/vistone/fingerprint/modules/core"
)

// Classifier 分类器接口
type Classifier interface {
	// Train 训练分类器
	Train(features [][]float64, labels []string) error
	// Predict 预测标签
	Predict(features []float64) (string, float64)
	// PredictTopK 返回 TopK 预测结果
	PredictTopK(features []float64, k int) []Prediction
}

// Prediction 预测结果
type Prediction struct {
	Label      string
	Confidence float64
}

// SimpleClassifier 简单分类器（基于距离）
type SimpleClassifier struct {
	classes    map[string][]float64 // 类别中心点
	weights    []float64            // 特征权重
	classCount int
	mu         sync.RWMutex
}

// NewSimpleClassifier 创建新的简单分类器
func NewSimpleClassifier(featureCount int) *SimpleClassifier {
	weights := make([]float64, featureCount)
	for i := range weights {
		weights[i] = 1.0
	}
	return &SimpleClassifier{
		classes:    make(map[string][]float64),
		weights:    weights,
		classCount: 0,
	}
}

// Train 训练分类器
func (c *SimpleClassifier) Train(features [][]float64, labels []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 计算每个类别的中心点
	classSums := make(map[string][]float64)
	classCounts := make(map[string]int)

	for i, feat := range features {
		label := labels[i]
		if _, ok := classSums[label]; !ok {
			classSums[label] = make([]float64, len(feat))
		}
		for j, v := range feat {
			classSums[label][j] += v
		}
		classCounts[label]++
	}

	// 计算平均值
	for label, sum := range classSums {
		count := classCounts[label]
		center := make([]float64, len(sum))
		for i, v := range sum {
			center[i] = v / float64(count)
		}
		c.classes[label] = center
	}

	c.classCount = len(c.classes)
	return nil
}

// Predict 预测标签
func (c *SimpleClassifier) Predict(features []float64) (string, float64) {
	predictions := c.PredictTopK(features, 1)
	if len(predictions) == 0 {
		return "", 0
	}
	return predictions[0].Label, predictions[0].Confidence
}

// PredictTopK 返回 TopK 预测结果
func (c *SimpleClassifier) PredictTopK(features []float64, k int) []Prediction {
	c.mu.RLock()
	defer c.mu.RUnlock()

	type distanceResult struct {
		label    string
		distance float64
	}

	// 计算到每个类别的加权距离
	results := make([]distanceResult, 0, len(c.classes))
	for label, center := range c.classes {
		dist := c.weightedDistance(features, center)
		results = append(results, distanceResult{label: label, distance: dist})
	}

	// 按距离排序（距离越小越相似）
	sort.Slice(results, func(i, j int) bool {
		return results[i].distance < results[j].distance
	})

	// 转换为置信度
	predictions := make([]Prediction, 0, min(k, len(results)))
	for i := 0; i < k && i < len(results); i++ {
		// 将距离转换为置信度（距离越小置信度越高）
		confidence := 1.0 / (1.0 + results[i].distance)
		predictions = append(predictions, Prediction{
			Label:      results[i].label,
			Confidence: confidence,
		})
	}

	return predictions
}

// weightedDistance 计算加权欧氏距离
func (c *SimpleClassifier) weightedDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.MaxFloat64
	}

	sum := 0.0
	for i := 0; i < len(a) && i < len(c.weights); i++ {
		diff := a[i] - b[i]
		sum += c.weights[i] * diff * diff
	}

	return math.Sqrt(sum)
}

// ProtocolClassifier 协议类型分类器（第一层）
type ProtocolClassifier struct {
	classifier *SimpleClassifier
}

// NewProtocolClassifier 创建新的协议分类器
func NewProtocolClassifier() *ProtocolClassifier {
	return &ProtocolClassifier{
		classifier: NewSimpleClassifier(10), // 10 个协议相关特征
	}
}

// Train 训练协议分类器
func (pc *ProtocolClassifier) Train(features [][]float64, labels []core.ProtocolType) error {
	strLabels := make([]string, len(labels))
	for i, l := range labels {
		strLabels[i] = string(l)
	}
	return pc.classifier.Train(features, strLabels)
}

// Predict 预测协议类型
func (pc *ProtocolClassifier) Predict(features []float64) (core.ProtocolType, float64) {
	label, conf := pc.classifier.Predict(features)
	return core.ProtocolType(label), conf
}

// FamilyClassifier 浏览器家族分类器（第二层）
type FamilyClassifier struct {
	classifier *SimpleClassifier
	protocol   core.ProtocolType
}

// NewFamilyClassifier 创建新的浏览器家族分类器
func NewFamilyClassifier(protocol core.ProtocolType) *FamilyClassifier {
	return &FamilyClassifier{
		classifier: NewSimpleClassifier(15), // 15 个浏览器相关特征
		protocol:   protocol,
	}
}

// Train 训练浏览器家族分类器
func (fc *FamilyClassifier) Train(features [][]float64, labels []core.BrowserType) error {
	strLabels := make([]string, len(labels))
	for i, l := range labels {
		strLabels[i] = string(l)
	}
	return fc.classifier.Train(features, strLabels)
}

// Predict 预测浏览器家族
func (fc *FamilyClassifier) Predict(features []float64) (core.BrowserType, float64) {
	label, conf := fc.classifier.Predict(features)
	return core.BrowserType(label), conf
}

// VersionClassifier 版本识别分类器（第三层）
type VersionClassifier struct {
	classifier    *SimpleClassifier
	browserFamily core.BrowserType
}

// NewVersionClassifier 创建新的版本分类器
func NewVersionClassifier(family core.BrowserType) *VersionClassifier {
	return &VersionClassifier{
		classifier:    NewSimpleClassifier(20), // 20 个版本相关特征
		browserFamily: family,
	}
}

// Train 训练版本分类器
func (vc *VersionClassifier) Train(features [][]float64, labels []string) error {
	return vc.classifier.Train(features, labels)
}

// Predict 预测版本
func (vc *VersionClassifier) Predict(features []float64) (string, float64) {
	return vc.classifier.Predict(features)
}

// min returns the smaller of a and b
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
