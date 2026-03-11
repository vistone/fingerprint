// Package ml — tensor.go 提供张量计算引擎，支持 CPU 并行计算和 GPU 设备抽象。
//
// 这是整个指纹智能分析模型库的计算基础。所有神经网络的前向传播、反向传播、
// 参数更新都建立在 Tensor 运算之上。
//
// GPU 支持策略：
//   - ComputeDevice 接口抽象计算设备（CPU / GPU）
//   - CPU 实现使用 goroutine 并行化批量运算
//   - GPU 实现可通过 build tag "gpu_cuda" 接入 CUDA 后端
//   - 张量数据布局采用行优先（row-major），与 CUDA/cuBLAS 兼容
package ml

import (
	"fmt"
	"math"
	"math/rand/v2"
	"runtime"
	"sync"
)

// ---------------------------------------------------------------------------
// ComputeDevice — 计算设备抽象
// ---------------------------------------------------------------------------

// ComputeDevice 定义计算设备接口。
// CPU 实现使用 goroutine 并行化；GPU 实现可通过 CGo + CUDA 接入。
type ComputeDevice interface {
	Name() string
	// MatMul 矩阵乘法: C = A × B, A[m×k], B[k×n] → C[m×n]
	MatMul(a, b *Tensor) *Tensor
	// MatMulAdd 矩阵乘加: C = A × B + bias (broadcast)
	MatMulAdd(a, b, bias *Tensor) *Tensor
	// Apply 逐元素函数: out[i] = fn(in[i])
	Apply(t *Tensor, fn func(float64) float64) *Tensor
	// BatchParallel 并行处理批量数据
	BatchParallel(batchSize int, fn func(i int))
}

// CPU 是默认的 CPU 计算设备，使用 goroutine 并行化。
var CPU ComputeDevice = &cpuDevice{workers: runtime.NumCPU()}

// cpuDevice CPU 计算设备实现
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

// DefaultDevice 当前活跃的计算设备，默认 CPU。
// 通过 SetDevice(gpu) 可切换到 GPU。
var DefaultDevice ComputeDevice = CPU

// SetDevice 设置全局默认计算设备。
func SetDevice(dev ComputeDevice) { DefaultDevice = dev }

// ---------------------------------------------------------------------------
// Tensor — 多维张量
// ---------------------------------------------------------------------------

// Tensor 表示一个多维浮点张量，行优先存储。
// Shape 说明各维度大小，Data 是展平后的一维数组。
type Tensor struct {
	Data  []float64
	Shape []int
}

// NewTensor 从给定的形状和数据创建张量。
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

// Zeros 创建全零张量。
func Zeros(dims ...int) *Tensor {
	size := 1
	for _, d := range dims {
		size *= d
	}
	return &Tensor{Data: make([]float64, size), Shape: append([]int{}, dims...)}
}

// Ones 创建全 1 张量。
func Ones(dims ...int) *Tensor {
	t := Zeros(dims...)
	for i := range t.Data {
		t.Data[i] = 1.0
	}
	return t
}

// RandN 创建标准正态分布随机张量。
func RandN(dims ...int) *Tensor {
	t := Zeros(dims...)
	for i := range t.Data {
		t.Data[i] = rand.NormFloat64()
	}
	return t
}

// RandNScaled 创建缩放正态分布随机张量（用于 He/Xavier 初始化）。
func RandNScaled(scale float64, dims ...int) *Tensor {
	t := Zeros(dims...)
	for i := range t.Data {
		t.Data[i] = rand.NormFloat64() * scale
	}
	return t
}

// Clone 深拷贝张量。
func (t *Tensor) Clone() *Tensor {
	d := make([]float64, len(t.Data))
	copy(d, t.Data)
	return &Tensor{Data: d, Shape: append([]int{}, t.Shape...)}
}

// Size 返回元素总数。
func (t *Tensor) Size() int { return len(t.Data) }

// Rows 返回第 0 维大小（行数），要求至少 2D。
func (t *Tensor) Rows() int {
	if len(t.Shape) < 2 {
		return t.Shape[0]
	}
	return t.Shape[0]
}

// Cols 返回第 1 维大小（列数），要求 2D。
func (t *Tensor) Cols() int {
	if len(t.Shape) < 2 {
		return 1
	}
	return t.Shape[1]
}

// At 获取 2D 张量 [i,j] 处的值。
func (t *Tensor) At(i, j int) float64 {
	return t.Data[i*t.Shape[1]+j]
}

// Set 设置 2D 张量 [i,j] 处的值。
func (t *Tensor) Set(i, j int, v float64) {
	t.Data[i*t.Shape[1]+j] = v
}

// Row 返回第 i 行的切片视图（共享底层数据）。
func (t *Tensor) Row(i int) []float64 {
	cols := t.Shape[1]
	return t.Data[i*cols : (i+1)*cols]
}

// FromSlice 从一维 float64 切片创建 [1×n] 行向量张量。
func FromSlice(data []float64) *Tensor {
	d := make([]float64, len(data))
	copy(d, data)
	return &Tensor{Data: d, Shape: []int{1, len(data)}}
}

// ToSlice 将张量展平为一维切片（拷贝）。
func (t *Tensor) ToSlice() []float64 {
	d := make([]float64, len(t.Data))
	copy(d, t.Data)
	return d
}

// ---------------------------------------------------------------------------
// 张量运算
// ---------------------------------------------------------------------------

// Add 逐元素加法: C = A + B (要求同形状或 B 可广播到 A)。
func (t *Tensor) Add(other *Tensor) *Tensor {
	if len(t.Data) == len(other.Data) {
		out := t.Clone()
		for i := range out.Data {
			out.Data[i] += other.Data[i]
		}
		return out
	}
	// 广播: other 是单行，t 是多行
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

// Sub 逐元素减法。
func (t *Tensor) Sub(other *Tensor) *Tensor {
	out := t.Clone()
	for i := range out.Data {
		out.Data[i] -= other.Data[i]
	}
	return out
}

// MulScalar 标量乘法。
func (t *Tensor) MulScalar(s float64) *Tensor {
	out := t.Clone()
	for i := range out.Data {
		out.Data[i] *= s
	}
	return out
}

// MulElem 逐元素乘法（Hadamard 积）。
func (t *Tensor) MulElem(other *Tensor) *Tensor {
	out := t.Clone()
	for i := range out.Data {
		out.Data[i] *= other.Data[i]
	}
	return out
}

// Transpose 返回 2D 张量的转置。
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

// MatMul 矩阵乘法，委托给 DefaultDevice。
func (t *Tensor) MatMul(other *Tensor) *Tensor {
	return DefaultDevice.MatMul(t, other)
}

// Sum 返回所有元素之和。
func (t *Tensor) Sum() float64 {
	s := 0.0
	for _, v := range t.Data {
		s += v
	}
	return s
}

// Mean 返回所有元素的均值。
func (t *Tensor) Mean() float64 {
	return t.Sum() / float64(len(t.Data))
}

// L2Norm 返回 L2 范数。
func (t *Tensor) L2Norm() float64 {
	s := 0.0
	for _, v := range t.Data {
		s += v * v
	}
	return math.Sqrt(s)
}

// Normalize 原地 L2 归一化（使向量模为 1）。
func (t *Tensor) Normalize() *Tensor {
	norm := t.L2Norm()
	if norm < 1e-12 {
		return t.Clone()
	}
	return t.MulScalar(1.0 / norm)
}

// Clamp 将所有元素限制在 [lo, hi] 范围内。
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

// Argmax 返回最大元素的索引。
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

// Max 返回最大元素值。
func (t *Tensor) Max() float64 {
	m := t.Data[0]
	for _, v := range t.Data[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

// SoftmaxRow 对 2D 张量的每一行做 Softmax 变换。
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

// SigmoidApply 逐元素 Sigmoid 变换。
func (t *Tensor) SigmoidApply() *Tensor {
	out := t.Clone()
	for i, v := range out.Data {
		out.Data[i] = 1.0 / (1.0 + math.Exp(-v))
	}
	return out
}

// ReluApply 逐元素 ReLU 变换。
func (t *Tensor) ReluApply() *Tensor {
	out := t.Clone()
	for i, v := range out.Data {
		if v < 0 {
			out.Data[i] = 0
		}
	}
	return out
}
