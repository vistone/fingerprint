package ml

import (
	"math"
)

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
