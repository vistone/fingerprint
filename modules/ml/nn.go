// Package ml — nn.go provides neural network building blocks.
//
// Contents:
//   - Layer interface and Dense (fully-connected) layer implementation
//   - Activation functions: ReLU, Sigmoid, Softmax, Tanh
//   - Batch normalization layer for training stability
//   - Sequential model: chains multiple layers into a forward/backward graph
//   - Adam optimizer with weight decay and gradient clipping
//   - Learning rate schedulers: cosine annealing, step decay
//   - Loss functions: CrossEntropy, MSE, TripletMargin
//   - Parameter initialization: He, Xavier, zero-init
//
// These components build directly on top of Tensor and support mini-batch training.
package ml

import (
	"math"
	"math/rand/v2"
)

// ---------------------------------------------------------------------------
// Layer interface
// ---------------------------------------------------------------------------

// Layer defines the interface for a single neural network layer.
type Layer interface {
	// Forward performs forward propagation: input[batch×in] → output[batch×out]
	Forward(input *Tensor) *Tensor
	// Backward performs backpropagation: gradOutput[batch×out] → gradInput[batch×in],
	// accumulating parameter gradients.
	Backward(gradOutput *Tensor) *Tensor
	// Params returns trainable parameters and their gradients (for optimizer updates).
	Params() []*Param
	// SetTraining switches between training and inference modes.
	SetTraining(training bool)
}

// Param represents a trainable parameter and its gradient.
type Param struct {
	Value *Tensor // parameter values
	Grad  *Tensor // parameter gradients
}

// ---------------------------------------------------------------------------
// Dense fully-connected layer
// ---------------------------------------------------------------------------

// DenseLayer implements a fully-connected layer: y = x·W^T + b
// where W[out×in], b[out], x[batch×in] → y[batch×out]
type DenseLayer struct {
	Weight   *Tensor // [outDim × inDim]
	Bias     *Tensor // [outDim]
	WeightG  *Tensor // weight gradient
	BiasG    *Tensor // bias gradient
	input    *Tensor // cached forward input (used in backward pass)
	inDim    int
	outDim   int
	useBias  bool
	training bool
}

// NewDenseLayer creates a fully-connected layer with He initialization.
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
	l.input = input // cache for backward pass
	// y = x · W^T + b
	wt := l.Weight.Transpose() // [inDim × outDim]
	out := DefaultDevice.MatMulAdd(input, wt, l.Bias)
	return out
}

func (l *DenseLayer) Backward(gradOutput *Tensor) *Tensor {
	batch := gradOutput.Shape[0]
	// Weight gradient: dW = gradOutput^T · input → [outDim × inDim]
	gt := gradOutput.Transpose() // [outDim × batch]
	dW := DefaultDevice.MatMul(gt, l.input)
	// Bias gradient: dB = sum(gradOutput, axis=0) → [1 × outDim]
	dB := Zeros(1, l.outDim)
	for i := 0; i < batch; i++ {
		for j := 0; j < l.outDim; j++ {
			dB.Data[j] += gradOutput.At(i, j)
		}
	}
	// Accumulate gradients (supports gradient accumulation across mini-batches)
	for i := range l.WeightG.Data {
		l.WeightG.Data[i] += dW.Data[i] / float64(batch)
	}
	for i := range l.BiasG.Data {
		l.BiasG.Data[i] += dB.Data[i] / float64(batch)
	}
	// Input gradient: dX = gradOutput · W → [batch × inDim]
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
// Activation layers
// ---------------------------------------------------------------------------

// ReLULayer implements ReLU activation: max(0, x)
type ReLULayer struct {
	mask *Tensor // forward mask, used in backward pass
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

func (l *ReLULayer) Params() []*Param          { return nil }
func (l *ReLULayer) SetTraining(training bool) {}

// SigmoidLayer implements sigmoid activation: σ(x) = 1/(1+e^(-x))
type SigmoidLayer struct {
	output *Tensor // cached output, used in backward pass
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

func (l *SigmoidLayer) Params() []*Param          { return nil }
func (l *SigmoidLayer) SetTraining(training bool) {}

// SoftmaxLayer implements row-wise softmax activation.
type SoftmaxLayer struct {
	output *Tensor
}

func NewSoftmaxLayer() *SoftmaxLayer { return &SoftmaxLayer{} }

func (l *SoftmaxLayer) Forward(input *Tensor) *Tensor {
	l.output = input.SoftmaxRow()
	return l.output
}

func (l *SoftmaxLayer) Backward(gradOutput *Tensor) *Tensor {
	// For CrossEntropy + Softmax combined losses, gradient passes directly to logits.
	// Here we provide the full Jacobian backward pass for standalone usage.
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

func (l *SoftmaxLayer) Params() []*Param          { return nil }
func (l *SoftmaxLayer) SetTraining(training bool) {}

// DropoutLayer implements dropout regularization.
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

func (l *DropoutLayer) Params() []*Param          { return nil }
func (l *DropoutLayer) SetTraining(training bool) { l.training = training }

// ---------------------------------------------------------------------------
// Batch normalization layer
// ---------------------------------------------------------------------------

// BatchNormLayer implements batch normalization: normalizes activations across the batch,
// then applies learned scale (gamma) and shift (beta).
// During inference, uses running mean/variance estimated during training.
type BatchNormLayer struct {
	dim      int
	gamma    *Tensor // learned scale [dim]
	beta     *Tensor // learned shift [dim]
	gammaG   *Tensor // gamma gradient
	betaG    *Tensor // beta gradient
	runMean  *Tensor // running mean for inference
	runVar   *Tensor // running variance for inference
	runMeanG *Tensor // zero gradient placeholder for serialization
	runVarG  *Tensor // zero gradient placeholder for serialization
	momentum float64 // running stat momentum (default 0.1)
	eps      float64 // numerical stability
	training bool

	// cached for backward pass
	xNorm *Tensor
	xMu   *Tensor
	std   []float64
	batch int
}

// NewBatchNormLayer creates a batch normalization layer.
func NewBatchNormLayer(dim int) *BatchNormLayer {
	return &BatchNormLayer{
		dim:        dim,
		gamma:      Ones(1, dim),
		beta:       Zeros(1, dim),
		gammaG:     Zeros(1, dim),
		betaG:      Zeros(1, dim),
		runMean:    Zeros(1, dim),
		runVar:     Ones(1, dim),
		runMeanG:   Zeros(1, dim),
		runVarG:    Zeros(1, dim),
		momentum:   0.1,
		eps:        1e-5,
		training:   true,
	}
}

func (l *BatchNormLayer) Forward(input *Tensor) *Tensor {
	rows, cols := input.Shape[0], input.Shape[1]
	l.batch = rows
	out := Zeros(rows, cols)

	if l.training && rows > 1 {
		// Compute batch mean and variance
		mean := make([]float64, cols)
		variance := make([]float64, cols)
		for j := 0; j < cols; j++ {
			sum := 0.0
			for i := 0; i < rows; i++ {
				sum += input.At(i, j)
			}
			mean[j] = sum / float64(rows)
		}
		for j := 0; j < cols; j++ {
			sum := 0.0
			for i := 0; i < rows; i++ {
				d := input.At(i, j) - mean[j]
				sum += d * d
			}
			variance[j] = sum / float64(rows)
		}

		// Normalize
		l.std = make([]float64, cols)
		l.xMu = Zeros(rows, cols)
		l.xNorm = Zeros(rows, cols)
		for j := 0; j < cols; j++ {
			l.std[j] = math.Sqrt(variance[j] + l.eps)
		}
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				d := input.At(i, j) - mean[j]
				l.xMu.Set(i, j, d)
				norm := d / l.std[j]
				l.xNorm.Set(i, j, norm)
				out.Set(i, j, norm*l.gamma.Data[j]+l.beta.Data[j])
			}
		}

		// Update running statistics
		for j := 0; j < cols; j++ {
			l.runMean.Data[j] = (1-l.momentum)*l.runMean.Data[j] + l.momentum*mean[j]
			l.runVar.Data[j] = (1-l.momentum)*l.runVar.Data[j] + l.momentum*variance[j]
		}
	} else {
		// Inference: use running statistics
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				norm := (input.At(i, j) - l.runMean.Data[j]) / math.Sqrt(l.runVar.Data[j]+l.eps)
				out.Set(i, j, norm*l.gamma.Data[j]+l.beta.Data[j])
			}
		}
	}
	return out
}

func (l *BatchNormLayer) Backward(gradOutput *Tensor) *Tensor {
	rows, cols := gradOutput.Shape[0], gradOutput.Shape[1]
	n := float64(rows)
	gradInput := Zeros(rows, cols)

	for j := 0; j < cols; j++ {
		// dGamma = sum(gradOutput * xNorm)
		// dBeta = sum(gradOutput)
		dGamma := 0.0
		dBeta := 0.0
		for i := 0; i < rows; i++ {
			dGamma += gradOutput.At(i, j) * l.xNorm.At(i, j)
			dBeta += gradOutput.At(i, j)
		}
		l.gammaG.Data[j] += dGamma / n
		l.betaG.Data[j] += dBeta / n

		// gradInput: standard batch norm backward
		invStd := 1.0 / l.std[j]
		dxNorm := 0.0
		dxNormXmu := 0.0
		for i := 0; i < rows; i++ {
			dx := gradOutput.At(i, j) * l.gamma.Data[j]
			dxNorm += dx
			dxNormXmu += dx * l.xMu.At(i, j)
		}
		for i := 0; i < rows; i++ {
			dx := gradOutput.At(i, j) * l.gamma.Data[j]
			gradInput.Set(i, j, invStd*(dx-dxNorm/n-l.xMu.At(i, j)*dxNormXmu/(n*l.std[j]*l.std[j])))
		}
	}
	return gradInput
}

func (l *BatchNormLayer) Params() []*Param {
	// Include running stats so they are serialized/deserialized along with learned params.
	// Order: gamma, beta, running_mean, running_var
	// Running stats have permanent zero gradients — optimizer won't modify them.
	return []*Param{
		{Value: l.gamma, Grad: l.gammaG},
		{Value: l.beta, Grad: l.betaG},
		{Value: l.runMean, Grad: l.runMeanG},
		{Value: l.runVar, Grad: l.runVarG},
	}
}

func (l *BatchNormLayer) SetTraining(training bool) { l.training = training }

// ---------------------------------------------------------------------------
// Sequential model
// ---------------------------------------------------------------------------

// Sequential chains multiple layers into an end-to-end model.
type Sequential struct {
	Layers []Layer
}

// NewSequential creates a sequential model.
func NewSequential(layers ...Layer) *Sequential {
	return &Sequential{Layers: layers}
}

// Forward executes all layers in order.
func (m *Sequential) Forward(input *Tensor) *Tensor {
	x := input
	for _, l := range m.Layers {
		x = l.Forward(x)
	}
	return x
}

// Backward executes all layers in reverse order.
func (m *Sequential) Backward(gradOutput *Tensor) *Tensor {
	g := gradOutput
	for i := len(m.Layers) - 1; i >= 0; i-- {
		g = m.Layers[i].Backward(g)
	}
	return g
}

// Params collects trainable parameters from all layers.
func (m *Sequential) Params() []*Param {
	var all []*Param
	for _, l := range m.Layers {
		if ps := l.Params(); ps != nil {
			all = append(all, ps...)
		}
	}
	return all
}

// SetTraining sets training/inference mode for all layers.
func (m *Sequential) SetTraining(training bool) {
	for _, l := range m.Layers {
		l.SetTraining(training)
	}
}

// ZeroGrad zeroes all parameter gradients.
func (m *Sequential) ZeroGrad() {
	for _, p := range m.Params() {
		for i := range p.Grad.Data {
			p.Grad.Data[i] = 0
		}
	}
}

// ---------------------------------------------------------------------------
// Adam optimizer
// ---------------------------------------------------------------------------

// AdamOptimizer implements the Adam (Adaptive Moment Estimation) optimizer.
type AdamOptimizer struct {
	LR      float64 // learning rate
	Beta1   float64 // first moment decay rate (default 0.9)
	Beta2   float64 // second moment decay rate (default 0.999)
	Epsilon float64 // numerical stability (default 1e-8)
	step    int
	m       []*Tensor // first moment estimates
	v       []*Tensor // second moment estimates
	params  []*Param
}

// NewAdamOptimizer creates an Adam optimizer.
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

// Step performs a single parameter update.
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

// ClipGradNorm clips parameter gradients by global L2 norm.
// maxNorm: maximum allowed gradient norm (typical: 1.0-5.0).
func ClipGradNorm(params []*Param, maxNorm float64) {
	totalNorm := 0.0
	for _, p := range params {
		for _, g := range p.Grad.Data {
			totalNorm += g * g
		}
	}
	totalNorm = math.Sqrt(totalNorm)
	if totalNorm > maxNorm {
		scale := maxNorm / totalNorm
		for _, p := range params {
			for i := range p.Grad.Data {
				p.Grad.Data[i] *= scale
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Learning rate schedulers
// ---------------------------------------------------------------------------

// LRScheduler adjusts the optimizer learning rate over training.
type LRScheduler interface {
	// StepLR updates the learning rate. Call once per epoch.
	StepLR(optimizer *AdamOptimizer, epoch, totalEpochs int)
}

// CosineAnnealingLR implements cosine annealing: LR decays from initial to minLR following a cosine curve.
type CosineAnnealingLR struct {
	InitialLR float64
	MinLR     float64
}

func NewCosineAnnealingLR(initialLR, minLR float64) *CosineAnnealingLR {
	return &CosineAnnealingLR{InitialLR: initialLR, MinLR: minLR}
}

func (s *CosineAnnealingLR) StepLR(opt *AdamOptimizer, epoch, totalEpochs int) {
	progress := float64(epoch) / float64(totalEpochs)
	opt.LR = s.MinLR + 0.5*(s.InitialLR-s.MinLR)*(1.0+math.Cos(math.Pi*progress))
}

// WarmupCosineAnnealingLR adds a linear warmup phase before cosine decay.
type WarmupCosineAnnealingLR struct {
	InitialLR    float64
	MinLR        float64
	WarmupEpochs int
}

func NewWarmupCosineAnnealingLR(initialLR, minLR float64, warmupEpochs int) *WarmupCosineAnnealingLR {
	return &WarmupCosineAnnealingLR{InitialLR: initialLR, MinLR: minLR, WarmupEpochs: warmupEpochs}
}

func (s *WarmupCosineAnnealingLR) StepLR(opt *AdamOptimizer, epoch, totalEpochs int) {
	if epoch < s.WarmupEpochs {
		// Linear warmup
		opt.LR = s.InitialLR * float64(epoch+1) / float64(s.WarmupEpochs)
	} else {
		// Cosine decay after warmup
		progress := float64(epoch-s.WarmupEpochs) / float64(totalEpochs-s.WarmupEpochs)
		opt.LR = s.MinLR + 0.5*(s.InitialLR-s.MinLR)*(1.0+math.Cos(math.Pi*progress))
	}
}

// ---------------------------------------------------------------------------
// Loss functions
// ---------------------------------------------------------------------------

// CrossEntropyLoss computes cross-entropy loss and gradient.
// probs: softmax output [batch × classes]
// targets: target class indices [batch] (int array)
// Returns: loss value, gradient w.r.t. logits [batch × classes]
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

// BinaryCrossEntropyLoss computes binary cross-entropy loss and gradient.
// preds: sigmoid output [batch × 1]
// targets: target values [batch] (0.0 or 1.0)
func BinaryCrossEntropyLoss(preds *Tensor, targets []float64) (float64, *Tensor) {
	batch := len(targets)
	loss := 0.0
	grad := Zeros(preds.Shape...)
	for i := 0; i < batch; i++ {
		p := preds.Data[i]
		t := targets[i]
		p = math.Max(1e-12, math.Min(1-1e-12, p))
		loss -= t*math.Log(p) + (1-t)*math.Log(1-p)
		grad.Data[i] = (p - t) / (p*(1-p) + 1e-12)
	}
	return loss / float64(batch), grad.MulScalar(1.0 / float64(batch))
}

// MSELoss computes mean squared error loss and gradient.
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

// TripletMarginLoss computes triplet margin loss.
// anchor, positive, negative are each [batch × dim] embedding tensors.
// margin: minimum margin between positive and negative distances.
func TripletMarginLoss(anchor, positive, negative *Tensor, margin float64) (float64, *Tensor, *Tensor, *Tensor) {
	batch := anchor.Shape[0]
	dim := anchor.Shape[1]
	totalLoss := 0.0
	gradA := Zeros(batch, dim)
	gradP := Zeros(batch, dim)
	gradN := Zeros(batch, dim)

	for i := 0; i < batch; i++ {
		// Compute positive and negative distances
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
			// Gradient: d(loss)/d(anchor) = 2(anchor-positive) - 2(anchor-negative)
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
