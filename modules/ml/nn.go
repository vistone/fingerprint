// Package ml — nn.go 提供神经网络基础组件。
//
// 包含：
//   - Layer 接口与 Dense（全连接）层实现
//   - 激活函数：ReLU、Sigmoid、Softmax、Tanh
//   - Sequential 模型：将多个层串联为一个前向/反向传播图
//   - Adam 优化器：自适应学习率
//   - 损失函数：CrossEntropy、MSE、TripletMargin
//   - 参数初始化：He、Xavier、零初始化
//
// 这些组件直接在 Tensor 之上构建，支持 mini-batch 训练。
package ml

import (
	"math"
	"math/rand/v2"
)

// ---------------------------------------------------------------------------
// Layer 接口
// ---------------------------------------------------------------------------

// Layer 定义神经网络中单个层的接口。
type Layer interface {
	// Forward 前向传播: input[batch×in] → output[batch×out]
	Forward(input *Tensor) *Tensor
	// Backward 反向传播: gradOutput[batch×out] → gradInput[batch×in]，同时累积参数梯度
	Backward(gradOutput *Tensor) *Tensor
	// Params 返回可训练参数及其梯度（用于优化器更新）
	Params() []*Param
	// SetTraining 设置训练/推理模式
	SetTraining(training bool)
}

// Param 表示一个可训练参数及其梯度。
type Param struct {
	Value *Tensor // 参数值
	Grad  *Tensor // 参数梯度
}

// ---------------------------------------------------------------------------
// Dense 全连接层
// ---------------------------------------------------------------------------

// DenseLayer 全连接层: y = x·W^T + b
// 其中 W[out×in], b[out], x[batch×in] → y[batch×out]
type DenseLayer struct {
	Weight    *Tensor // [outDim × inDim]
	Bias      *Tensor // [outDim]
	WeightG   *Tensor // 梯度
	BiasG     *Tensor // 梯度
	input     *Tensor // 缓存前向传播输入（反向传播用）
	inDim     int
	outDim    int
	useBias   bool
	training  bool
}

// NewDenseLayer 创建全连接层，使用 He 初始化。
func NewDenseLayer(inDim, outDim int) *DenseLayer {
	scale := math.Sqrt(2.0 / float64(inDim)) // He init
	w := RandNScaled(scale, outDim, inDim)
	b := Zeros(1, outDim)
	return &DenseLayer{
		Weight:  w,
		Bias:    b,
		WeightG: Zeros(outDim, inDim),
		BiasG:   Zeros(1, outDim),
		inDim:   inDim,
		outDim:  outDim,
		useBias: true,
	}
}

func (l *DenseLayer) Forward(input *Tensor) *Tensor {
	l.input = input // 缓存用于反向传播
	// y = x · W^T + b
	wt := l.Weight.Transpose() // [inDim × outDim]
	out := DefaultDevice.MatMulAdd(input, wt, l.Bias)
	return out
}

func (l *DenseLayer) Backward(gradOutput *Tensor) *Tensor {
	batch := gradOutput.Shape[0]
	// 梯度 of Weight: dW = gradOutput^T · input → [outDim × inDim]
	gt := gradOutput.Transpose() // [outDim × batch]
	dW := DefaultDevice.MatMul(gt, l.input)
	// 梯度 of Bias: dB = sum(gradOutput, axis=0) → [1 × outDim]
	dB := Zeros(1, l.outDim)
	for i := 0; i < batch; i++ {
		for j := 0; j < l.outDim; j++ {
			dB.Data[j] += gradOutput.At(i, j)
		}
	}
	// 累积梯度（支持梯度累加用于多 mini-batch）
	for i := range l.WeightG.Data {
		l.WeightG.Data[i] += dW.Data[i] / float64(batch)
	}
	for i := range l.BiasG.Data {
		l.BiasG.Data[i] += dB.Data[i] / float64(batch)
	}
	// 梯度 of input: dX = gradOutput · W → [batch × inDim]
	dX := DefaultDevice.MatMul(gradOutput, l.Weight) // [batch×out] × [out×in] → [batch×in]
	return dX
}

func (l *DenseLayer) Params() []*Param {
	params := []*Param{
		{Value: l.Weight, Grad: l.WeightG},
	}
	if l.useBias {
		params = append(params, &Param{Value: l.Bias, Grad: l.BiasG})
	}
	return params
}

func (l *DenseLayer) SetTraining(training bool) { l.training = training }

// ---------------------------------------------------------------------------
// 激活函数层
// ---------------------------------------------------------------------------

// ReLULayer ReLU 激活: max(0, x)
type ReLULayer struct {
	mask *Tensor // 前向传播的 mask，反向传播用
}

func NewReLULayer() *ReLULayer { return &ReLULayer{} }

func (l *ReLULayer) Forward(input *Tensor) *Tensor {
	out := input.Clone()
	l.mask = Zeros(input.Shape...)
	for i, v := range out.Data {
		if v > 0 {
			l.mask.Data[i] = 1
		} else {
			out.Data[i] = 0
		}
	}
	return out
}

func (l *ReLULayer) Backward(gradOutput *Tensor) *Tensor {
	return gradOutput.MulElem(l.mask)
}

func (l *ReLULayer) Params() []*Param           { return nil }
func (l *ReLULayer) SetTraining(training bool)   {}

// SigmoidLayer Sigmoid 激活: σ(x) = 1/(1+e^(-x))
type SigmoidLayer struct {
	output *Tensor // 缓存输出，反向传播用
}

func NewSigmoidLayer() *SigmoidLayer { return &SigmoidLayer{} }

func (l *SigmoidLayer) Forward(input *Tensor) *Tensor {
	l.output = input.SigmoidApply()
	return l.output
}

func (l *SigmoidLayer) Backward(gradOutput *Tensor) *Tensor {
	// dσ/dx = σ(x) × (1 - σ(x))
	ones := Ones(l.output.Shape...)
	dsig := l.output.MulElem(ones.Sub(l.output))
	return gradOutput.MulElem(dsig)
}

func (l *SigmoidLayer) Params() []*Param         { return nil }
func (l *SigmoidLayer) SetTraining(training bool) {}

// SoftmaxLayer Softmax 激活（按行）
type SoftmaxLayer struct {
	output *Tensor
}

func NewSoftmaxLayer() *SoftmaxLayer { return &SoftmaxLayer{} }

func (l *SoftmaxLayer) Forward(input *Tensor) *Tensor {
	l.output = input.SoftmaxRow()
	return l.output
}

func (l *SoftmaxLayer) Backward(gradOutput *Tensor) *Tensor {
	// 对于 CrossEntropy + Softmax 的组合，梯度直接传递到 logits
	// 这里提供完整的 Jacobian 反向传播以支持独立使用
	rows, cols := l.output.Shape[0], l.output.Shape[1]
	grad := Zeros(rows, cols)
	for i := 0; i < rows; i++ {
		s := l.output.Row(i)
		g := gradOutput.Row(i)
		for j := 0; j < cols; j++ {
			sum := 0.0
			for k := 0; k < cols; k++ {
				kronecker := 0.0
				if j == k {
					kronecker = 1.0
				}
				sum += g[k] * s[k] * (kronecker - s[j])
			}
			grad.Set(i, j, sum)
		}
	}
	return grad
}

func (l *SoftmaxLayer) Params() []*Param         { return nil }
func (l *SoftmaxLayer) SetTraining(training bool) {}

// DropoutLayer Dropout 正则化层
type DropoutLayer struct {
	rate     float64
	mask     *Tensor
	training bool
}

func NewDropoutLayer(rate float64) *DropoutLayer {
	return &DropoutLayer{rate: rate, training: true}
}

func (l *DropoutLayer) Forward(input *Tensor) *Tensor {
	if !l.training || l.rate <= 0 {
		return input.Clone()
	}
	out := input.Clone()
	l.mask = Ones(input.Shape...)
	scale := 1.0 / (1.0 - l.rate)
	for i := range out.Data {
		if rand.Float64() < l.rate {
			out.Data[i] = 0
			l.mask.Data[i] = 0
		} else {
			out.Data[i] *= scale
			l.mask.Data[i] = scale
		}
	}
	return out
}

func (l *DropoutLayer) Backward(gradOutput *Tensor) *Tensor {
	if !l.training || l.mask == nil {
		return gradOutput.Clone()
	}
	return gradOutput.MulElem(l.mask)
}

func (l *DropoutLayer) Params() []*Param         { return nil }
func (l *DropoutLayer) SetTraining(training bool) { l.training = training }

// ---------------------------------------------------------------------------
// Sequential 模型
// ---------------------------------------------------------------------------

// Sequential 将多个 Layer 串联成一个端到端模型。
type Sequential struct {
	Layers []Layer
}

// NewSequential 创建序列模型。
func NewSequential(layers ...Layer) *Sequential {
	return &Sequential{Layers: layers}
}

// Forward 依次执行所有层的前向传播。
func (m *Sequential) Forward(input *Tensor) *Tensor {
	x := input
	for _, l := range m.Layers {
		x = l.Forward(x)
	}
	return x
}

// Backward 逆序执行所有层的反向传播。
func (m *Sequential) Backward(gradOutput *Tensor) *Tensor {
	g := gradOutput
	for i := len(m.Layers) - 1; i >= 0; i-- {
		g = m.Layers[i].Backward(g)
	}
	return g
}

// Params 收集所有层的可训练参数。
func (m *Sequential) Params() []*Param {
	var all []*Param
	for _, l := range m.Layers {
		if ps := l.Params(); ps != nil {
			all = append(all, ps...)
		}
	}
	return all
}

// SetTraining 设置所有层的训练/推理模式。
func (m *Sequential) SetTraining(training bool) {
	for _, l := range m.Layers {
		l.SetTraining(training)
	}
}

// ZeroGrad 清零所有参数梯度。
func (m *Sequential) ZeroGrad() {
	for _, p := range m.Params() {
		for i := range p.Grad.Data {
			p.Grad.Data[i] = 0
		}
	}
}

// ---------------------------------------------------------------------------
// Adam 优化器
// ---------------------------------------------------------------------------

// AdamOptimizer 自适应矩估计优化器（Adam）。
type AdamOptimizer struct {
	LR      float64 // 学习率
	Beta1   float64 // 一阶矩衰减率 (default 0.9)
	Beta2   float64 // 二阶矩衰减率 (default 0.999)
	Epsilon float64 // 数值稳定性 (default 1e-8)
	step    int
	m       []*Tensor // 一阶矩
	v       []*Tensor // 二阶矩
	params  []*Param
}

// NewAdamOptimizer 创建 Adam 优化器。
func NewAdamOptimizer(params []*Param, lr float64) *AdamOptimizer {
	m := make([]*Tensor, len(params))
	v := make([]*Tensor, len(params))
	for i, p := range params {
		m[i] = Zeros(p.Value.Shape...)
		v[i] = Zeros(p.Value.Shape...)
	}
	return &AdamOptimizer{
		LR:      lr,
		Beta1:   0.9,
		Beta2:   0.999,
		Epsilon: 1e-8,
		params:  params,
		m:       m,
		v:       v,
	}
}

// Step 执行一步参数更新。
func (opt *AdamOptimizer) Step() {
	opt.step++
	bc1 := 1.0 - math.Pow(opt.Beta1, float64(opt.step))
	bc2 := 1.0 - math.Pow(opt.Beta2, float64(opt.step))

	for i, p := range opt.params {
		for j := range p.Value.Data {
			g := p.Grad.Data[j]
			opt.m[i].Data[j] = opt.Beta1*opt.m[i].Data[j] + (1-opt.Beta1)*g
			opt.v[i].Data[j] = opt.Beta2*opt.v[i].Data[j] + (1-opt.Beta2)*g*g
			mHat := opt.m[i].Data[j] / bc1
			vHat := opt.v[i].Data[j] / bc2
			p.Value.Data[j] -= opt.LR * mHat / (math.Sqrt(vHat) + opt.Epsilon)
		}
	}
}

// ---------------------------------------------------------------------------
// 损失函数
// ---------------------------------------------------------------------------

// CrossEntropyLoss 计算交叉熵损失与梯度。
// probs: Softmax 输出 [batch × classes]
// targets: 目标类别索引 [batch]（int 数组）
// 返回: 损失值, 对 logits 的梯度 [batch × classes]
func CrossEntropyLoss(probs *Tensor, targets []int) (float64, *Tensor) {
	batch, classes := probs.Shape[0], probs.Shape[1]
	loss := 0.0
	grad := probs.Clone()
	for i := 0; i < batch; i++ {
		t := targets[i]
		p := probs.At(i, t)
		if p < 1e-12 {
			p = 1e-12
		}
		loss -= math.Log(p)
		for j := 0; j < classes; j++ {
			if j == t {
				grad.Set(i, j, probs.At(i, j)-1.0)
			} else {
				grad.Set(i, j, probs.At(i, j))
			}
		}
	}
	return loss / float64(batch), grad.MulScalar(1.0 / float64(batch))
}

// BinaryCrossEntropyLoss 计算二元交叉熵损失与梯度。
// preds: Sigmoid 输出 [batch × 1]
// targets: 目标值 [batch]（0.0 或 1.0）
func BinaryCrossEntropyLoss(preds *Tensor, targets []float64) (float64, *Tensor) {
	batch := len(targets)
	loss := 0.0
	grad := Zeros(preds.Shape...)
	for i := 0; i < batch; i++ {
		p := preds.Data[i]
		t := targets[i]
		p = math.Max(1e-12, math.Min(1-1e-12, p))
		loss -= t*math.Log(p) + (1-t)*math.Log(1-p)
		grad.Data[i] = (p - t) / (p * (1 - p) + 1e-12)
	}
	return loss / float64(batch), grad.MulScalar(1.0 / float64(batch))
}

// MSELoss 计算均方误差损失与梯度。
func MSELoss(preds, targets *Tensor) (float64, *Tensor) {
	diff := preds.Sub(targets)
	n := float64(len(diff.Data))
	loss := 0.0
	for _, v := range diff.Data {
		loss += v * v
	}
	grad := diff.MulScalar(2.0 / n)
	return loss / n, grad
}

// TripletMarginLoss 计算三元组损失。
// anchor, positive, negative 各为 [batch × dim] 的嵌入向量。
// margin: 最小间距。
func TripletMarginLoss(anchor, positive, negative *Tensor, margin float64) (float64, *Tensor, *Tensor, *Tensor) {
	batch := anchor.Shape[0]
	dim := anchor.Shape[1]
	totalLoss := 0.0
	gradA := Zeros(batch, dim)
	gradP := Zeros(batch, dim)
	gradN := Zeros(batch, dim)

	for i := 0; i < batch; i++ {
		// 计算距离
		dPos, dNeg := 0.0, 0.0
		for j := 0; j < dim; j++ {
			dp := anchor.At(i, j) - positive.At(i, j)
			dn := anchor.At(i, j) - negative.At(i, j)
			dPos += dp * dp
			dNeg += dn * dn
		}
		loss := dPos - dNeg + margin
		if loss > 0 {
			totalLoss += loss
			// 梯度: d(loss)/d(anchor) = 2(anchor-positive) - 2(anchor-negative)
			// d(loss)/d(positive) = -2(anchor-positive)
			// d(loss)/d(negative) = 2(anchor-negative)
			for j := 0; j < dim; j++ {
				ap := anchor.At(i, j) - positive.At(i, j)
				an := anchor.At(i, j) - negative.At(i, j)
				gradA.Set(i, j, 2*(ap-an)/float64(batch))
				gradP.Set(i, j, -2*ap/float64(batch))
				gradN.Set(i, j, 2*an/float64(batch))
			}
		}
	}
	return totalLoss / float64(batch), gradA, gradP, gradN
}
