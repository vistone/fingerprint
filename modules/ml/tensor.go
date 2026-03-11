// Package ml — tensor.go provides the tensor computation engine with CPU parallel
// execution and GPU device abstraction.
//
// This is the computational foundation for the entire fingerprint analysis model
// library. All neural network forward/backward propagation and parameter updates
// are built on top of Tensor operations.
//
// GPU support strategy:
//   - ComputeDevice interface abstracts compute backends (CPU / GPU)
//   - CPU implementation uses goroutine-parallel batch operations
//   - GPU implementation can be plugged in via build tag "gpu_cuda" + CGo
//   - Tensor data layout is row-major, compatible with CUDA/cuBLAS
package ml

import (
	"fmt"
	"math"
	"math/rand/v2"
	"runtime"
	"sync"
)

// ---------------------------------------------------------------------------
// ComputeDevice — compute device abstraction
// ---------------------------------------------------------------------------

// ComputeDevice defines the compute device interface.
// CPU uses goroutine parallelism; GPU can be plugged in via CGo + CUDA.
type ComputeDevice interface {
	Name() string
	// MatMul performs matrix multiplication: C = A × B, A[m×k], B[k×n] → C[m×n]
	MatMul(a, b *Tensor) *Tensor
	// MatMulAdd performs fused multiply-add: C = A × B + bias (broadcast)
	MatMulAdd(a, b, bias *Tensor) *Tensor
	// Apply applies an element-wise function: out[i] = fn(in[i])
	Apply(t *Tensor, fn func(float64) float64) *Tensor
	// BatchParallel executes fn in parallel over [0, batchSize)
	BatchParallel(batchSize int, fn func(i int))
}

// CPU is the default CPU compute device using goroutine parallelism.
var CPU ComputeDevice = &cpuDevice{workers: runtime.NumCPU()}

// cpuDevice implements ComputeDevice for CPU with goroutine parallelism.
type cpuDevice struct {
	workers int
}

func (d *cpuDevice) Name() string { return fmt.Sprintf("CPU(%d cores)", d.workers) }

func (d *cpuDevice) MatMul(a, b *Tensor) *Tensor {
	if len(a.Shape) != 2 || len(b.Shape) != 2 {
		panic("MatMul requires 2D tensors")
	}
	m, k1 := a.Shape[0], a.Shape[1]
	k2, n := b.Shape[0], b.Shape[1]
	if k1 != k2 {
		panic(fmt.Sprintf("MatMul shape mismatch: [%d×%d] × [%d×%d]", m, k1, k2, n))
	}
	c := Zeros(m, n)
	d.BatchParallel(m, func(i int) {
		for j := 0; j < n; j++ {
			sum := 0.0
			for p := 0; p < k1; p++ {
				sum += a.At(i, p) * b.At(p, j)
			}
			c.Set(i, j, sum)
		}
	})
	return c
}

func (d *cpuDevice) MatMulAdd(a, b, bias *Tensor) *Tensor {
	c := d.MatMul(a, b)
	if bias != nil {
		n := c.Shape[1]
		for i := 0; i < c.Shape[0]; i++ {
			for j := 0; j < n; j++ {
				c.Data[i*n+j] += bias.Data[j]
			}
		}
	}
	return c
}

func (d *cpuDevice) Apply(t *Tensor, fn func(float64) float64) *Tensor {
	out := &Tensor{Data: make([]float64, len(t.Data)), Shape: append([]int{}, t.Shape...)}
	d.BatchParallel(len(t.Data), func(i int) {
		out.Data[i] = fn(t.Data[i])
	})
	return out
}

func (d *cpuDevice) BatchParallel(n int, fn func(i int)) {
	if n <= d.workers*4 || d.workers <= 1 {
		for i := 0; i < n; i++ {
			fn(i)
		}
		return
	}
	var wg sync.WaitGroup
	chunk := (n + d.workers - 1) / d.workers
	for w := 0; w < d.workers; w++ {
		start := w * chunk
		end := start + chunk
		if end > n {
			end = n
		}
		if start >= end {
			break
		}
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			for i := s; i < e; i++ {
				fn(i)
			}
		}(start, end)
	}
	wg.Wait()
}

// DefaultDevice is the currently active compute device, defaults to CPU.
// Use SetDevice(gpu) to switch to a GPU backend.
var DefaultDevice ComputeDevice = CPU

// SetDevice sets the global default compute device.
func SetDevice(dev ComputeDevice) { DefaultDevice = dev }

// ---------------------------------------------------------------------------
// Tensor — multi-dimensional tensor
// ---------------------------------------------------------------------------

// Tensor represents a multi-dimensional float64 tensor in row-major layout.
// Shape holds the size of each dimension; Data is the flattened 1D array.
type Tensor struct {
	Data  []float64
	Shape []int
}

// NewTensor creates a tensor from the given shape and data.
func NewTensor(shape []int, data []float64) *Tensor {
	size := 1
	for _, s := range shape {
		size *= s
	}
	if len(data) != size {
		panic(fmt.Sprintf("tensor: data length %d != shape product %d", len(data), size))
	}
	return &Tensor{Data: data, Shape: append([]int{}, shape...)}
}

// Zeros creates a zero-filled tensor.
func Zeros(dims ...int) *Tensor {
	size := 1
	for _, d := range dims {
		size *= d
	}
	return &Tensor{Data: make([]float64, size), Shape: append([]int{}, dims...)}
}

// Ones creates a tensor filled with ones.
func Ones(dims ...int) *Tensor {
	t := Zeros(dims...)
	for i := range t.Data {
		t.Data[i] = 1.0
	}
	return t
}

// RandN creates a tensor with standard normal random values.
func RandN(dims ...int) *Tensor {
	t := Zeros(dims...)
	for i := range t.Data {
		t.Data[i] = rand.NormFloat64()
	}
	return t
}

// RandNScaled creates a scaled normal random tensor (for He/Xavier initialization).
func RandNScaled(scale float64, dims ...int) *Tensor {
	t := Zeros(dims...)
	for i := range t.Data {
		t.Data[i] = rand.NormFloat64() * scale
	}
	return t
}

// Clone returns a deep copy of the tensor.
func (t *Tensor) Clone() *Tensor {
	d := make([]float64, len(t.Data))
	copy(d, t.Data)
	return &Tensor{Data: d, Shape: append([]int{}, t.Shape...)}
}

// Size returns the total number of elements.
func (t *Tensor) Size() int { return len(t.Data) }

// Rows returns the size of dimension 0 (row count).
func (t *Tensor) Rows() int {
	if len(t.Shape) < 2 {
		return t.Shape[0]
	}
	return t.Shape[0]
}

// Cols returns the size of dimension 1 (column count) for 2D tensors.
func (t *Tensor) Cols() int {
	if len(t.Shape) < 2 {
		return 1
	}
	return t.Shape[1]
}

// At returns the value at [i,j] in a 2D tensor.
func (t *Tensor) At(i, j int) float64 {
	return t.Data[i*t.Shape[1]+j]
}

// Set sets the value at [i,j] in a 2D tensor.
func (t *Tensor) Set(i, j int, v float64) {
	t.Data[i*t.Shape[1]+j] = v
}

// Row returns a slice view of row i (shares underlying data).
func (t *Tensor) Row(i int) []float64 {
	cols := t.Shape[1]
	return t.Data[i*cols : (i+1)*cols]
}

// FromSlice creates a [1×n] row vector tensor from a float64 slice.
func FromSlice(data []float64) *Tensor {
	d := make([]float64, len(data))
	copy(d, data)
	return &Tensor{Data: d, Shape: []int{1, len(data)}}
}

// ToSlice returns a flattened copy of the tensor data.
func (t *Tensor) ToSlice() []float64 {
	d := make([]float64, len(t.Data))
	copy(d, t.Data)
	return d
}

// ---------------------------------------------------------------------------
// Tensor operations
// ---------------------------------------------------------------------------

// Add performs element-wise addition: C = A + B (same shape or B broadcastable to A).
func (t *Tensor) Add(other *Tensor) *Tensor {
	if len(t.Data) == len(other.Data) {
		out := t.Clone()
		for i := range out.Data {
			out.Data[i] += other.Data[i]
		}
		return out
	}
	// Broadcast: other is single-row, t is multi-row
	if len(t.Shape) == 2 && len(other.Shape) >= 1 && other.Size() == t.Shape[1] {
		out := t.Clone()
		cols := t.Shape[1]
		for i := 0; i < t.Shape[0]; i++ {
			for j := 0; j < cols; j++ {
				out.Data[i*cols+j] += other.Data[j]
			}
		}
		return out
	}
	panic(fmt.Sprintf("Add shape mismatch: %v vs %v", t.Shape, other.Shape))
}

// Sub performs element-wise subtraction.
func (t *Tensor) Sub(other *Tensor) *Tensor {
	out := t.Clone()
	for i := range out.Data {
		out.Data[i] -= other.Data[i]
	}
	return out
}

// MulScalar performs scalar multiplication.
func (t *Tensor) MulScalar(s float64) *Tensor {
	out := t.Clone()
	for i := range out.Data {
		out.Data[i] *= s
	}
	return out
}

// MulElem performs element-wise (Hadamard) multiplication.
func (t *Tensor) MulElem(other *Tensor) *Tensor {
	out := t.Clone()
	for i := range out.Data {
		out.Data[i] *= other.Data[i]
	}
	return out
}

// Transpose returns the transpose of a 2D tensor.
func (t *Tensor) Transpose() *Tensor {
	if len(t.Shape) != 2 {
		panic("Transpose requires 2D tensor")
	}
	r, c := t.Shape[0], t.Shape[1]
	out := Zeros(c, r)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			out.Set(j, i, t.At(i, j))
		}
	}
	return out
}

// MatMul performs matrix multiplication, delegated to DefaultDevice.
func (t *Tensor) MatMul(other *Tensor) *Tensor {
	return DefaultDevice.MatMul(t, other)
}

// Sum returns the sum of all elements.
func (t *Tensor) Sum() float64 {
	s := 0.0
	for _, v := range t.Data {
		s += v
	}
	return s
}

// Mean returns the mean of all elements.
func (t *Tensor) Mean() float64 {
	return t.Sum() / float64(len(t.Data))
}

// L2Norm returns the L2 norm.
func (t *Tensor) L2Norm() float64 {
	s := 0.0
	for _, v := range t.Data {
		s += v * v
	}
	return math.Sqrt(s)
}

// Normalize returns an L2-normalized copy (unit vector).
func (t *Tensor) Normalize() *Tensor {
	norm := t.L2Norm()
	if norm < 1e-12 {
		return t.Clone()
	}
	return t.MulScalar(1.0 / norm)
}

// Clamp restricts all elements to the range [lo, hi].
func (t *Tensor) Clamp(lo, hi float64) *Tensor {
	out := t.Clone()
	for i, v := range out.Data {
		if v < lo {
			out.Data[i] = lo
		} else if v > hi {
			out.Data[i] = hi
		}
	}
	return out
}

// Argmax returns the index of the maximum element.
func (t *Tensor) Argmax() int {
	maxIdx := 0
	maxVal := t.Data[0]
	for i := 1; i < len(t.Data); i++ {
		if t.Data[i] > maxVal {
			maxVal = t.Data[i]
			maxIdx = i
		}
	}
	return maxIdx
}

// Max returns the maximum element value.
func (t *Tensor) Max() float64 {
	m := t.Data[0]
	for _, v := range t.Data[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

// SoftmaxRow applies softmax to each row of a 2D tensor.
func (t *Tensor) SoftmaxRow() *Tensor {
	if len(t.Shape) != 2 {
		panic("SoftmaxRow requires 2D tensor")
	}
	out := t.Clone()
	rows, cols := t.Shape[0], t.Shape[1]
	for i := 0; i < rows; i++ {
		row := out.Data[i*cols : (i+1)*cols]
		maxV := row[0]
		for _, v := range row[1:] {
			if v > maxV {
				maxV = v
			}
		}
		sumExp := 0.0
		for j := range row {
			row[j] = math.Exp(row[j] - maxV)
			sumExp += row[j]
		}
		for j := range row {
			row[j] /= sumExp
		}
	}
	return out
}

// SigmoidApply applies element-wise sigmoid.
func (t *Tensor) SigmoidApply() *Tensor {
	out := t.Clone()
	for i, v := range out.Data {
		out.Data[i] = 1.0 / (1.0 + math.Exp(-v))
	}
	return out
}

// ReluApply applies element-wise ReLU.
func (t *Tensor) ReluApply() *Tensor {
	out := t.Clone()
	for i, v := range out.Data {
		if v < 0 {
			out.Data[i] = 0
		}
	}
	return out
}
