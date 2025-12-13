package utils

import (
	"math/rand"
	"sync"
	"time"
)

// RandGenerator 统一的随机数生成器
// 提供线程安全的随机数生成功能
type RandGenerator struct {
	rng *rand.Rand
	mu  sync.Mutex
}

var (
	globalRandGen *RandGenerator
	initOnce      sync.Once
)

// GetGlobalRandGenerator 获取全局随机数生成器单例
func GetGlobalRandGenerator() *RandGenerator {
	initOnce.Do(func() {
		globalRandGen = NewRandGenerator()
	})
	return globalRandGen
}

// NewRandGenerator 创建新的随机数生成器
// 使用当前时间的纳秒作为种子，并添加一个额外的随机偏移
func NewRandGenerator() *RandGenerator {
	// 使用时间戳 + goroutine ID (间接通过多次调用) 来增加随机性
	seed := time.Now().UnixNano()
	// 添加额外的随机性：使用默认随机源生成偏移
	seed += int64(rand.Intn(10000))
	
	return &RandGenerator{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// Intn 返回 [0, n) 范围内的随机整数
func (r *RandGenerator) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Intn(n)
}

// Int63n 返回 [0, n) 范围内的随机 int64
func (r *RandGenerator) Int63n(n int64) int64 {
	if n <= 0 {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Int63n(n)
}

// Shuffle 随机打乱切片
func (r *RandGenerator) Shuffle(n int, swap func(i, j int)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rng.Shuffle(n, swap)
}

// RandomChoice 从切片中随机选择一个元素
func RandomChoice[T any](items []T) T {
	if len(items) == 0 {
		var zero T
		return zero
	}
	gen := GetGlobalRandGenerator()
	return items[gen.Intn(len(items))]
}

// RandomChoiceString 从字符串切片中随机选择一个
func RandomChoiceString(items []string) string {
	return RandomChoice(items)
}
