// Package ml provides lightweight neural network primitives built on Tensor.
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
	// Forward computes layer output for the input tensor.
	Forward(input *Tensor) *Tensor
	// Backward propagates output gradients and accumulates parameter gradients.
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

// DenseLayer implements a fully-connected affine layer.
// Output is y = x*W^T + b.
type DenseLayer struct {
	Weight   *Tensor // [outDim, inDim]
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
	// y = x*W^T + b
	wt := l.Weight.Transpose()
	out := DefaultDevice.MatMulAdd(input, wt, l.Bias)
	return out
}

func (l *DenseLayer) Backward(gradOutput *Tensor) *Tensor {
	batch := gradOutput.Shape[0]
	// Weight gradient dW = dY^T * X.
	gt := gradOutput.Transpose()
	dW := DefaultDevice.MatMul(gt, l.input)
	// Bias gradient dB = row-wise sum(dY).
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
	// Input gradient dX = dY * W.
	dX := DefaultDevice.MatMul(gradOutput, l.Weight)
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

// SigmoidLayer implements element-wise sigmoid activation.
type SigmoidLayer struct {
	output *Tensor // cached output, used in backward pass
}

func NewSigmoidLayer() *SigmoidLayer { return &SigmoidLayer{} }

func (l *SigmoidLayer) Forward(input *Tensor) *Tensor {
	l.output = input.SigmoidApply()
	return l.output
}

func (l *SigmoidLayer) Backward(gradOutput *Tensor) *Tensor {
	// d/dx sigmoid(x) = sigmoid(x) * (1 - sigmoid(x))
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
		dim:      dim,
		gamma:    Ones(1, dim),
		beta:     Zeros(1, dim),
		gammaG:   Zeros(1, dim),
		betaG:    Zeros(1, dim),
		runMean:  Zeros(1, dim),
		runVar:   Ones(1, dim),
		runMeanG: Zeros(1, dim),
		runVarG:  Zeros(1, dim),
		momentum: 0.1,
		eps:      1e-5,
		training: true,
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
	// Running-stat gradients are placeholders and remain zero during optimization.
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
