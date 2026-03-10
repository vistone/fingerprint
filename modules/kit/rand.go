package utils

import (
	"math/rand"
	"sync"
	"time"
)

// translated comment
// translated comment
type RandGenerator struct {
	rng *rand.Rand
	mu  sync.Mutex
}

var (
	globalRandGen *RandGenerator
	initOnce      sync.Once
)

// translated comment
func GetGlobalRandGenerator() *RandGenerator {
	initOnce.Do(func() {
		globalRandGen = NewRandGenerator()
	})
	return globalRandGen
}

// translated comment
// translated comment
func NewRandGenerator() *RandGenerator {
	// translated comment
	seed := time.Now().UnixNano()
	// translated comment
	seed += int64(rand.Intn(10000))

	return &RandGenerator{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// translated comment
func (r *RandGenerator) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Intn(n)
}

// translated comment
func (r *RandGenerator) Int63n(n int64) int64 {
	if n <= 0 {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Int63n(n)
}

// translated comment
func (r *RandGenerator) Shuffle(n int, swap func(i, j int)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rng.Shuffle(n, swap)
}

// translated comment
func RandomChoice[T any](items []T) T {
	if len(items) == 0 {
		var zero T
		return zero
	}
	gen := GetGlobalRandGenerator()
	return items[gen.Intn(len(items))]
}

// translated comment
func RandomChoiceString(items []string) string {
	return RandomChoice(items)
}
