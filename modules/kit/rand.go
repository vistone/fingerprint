package utils

import (
	"math/rand"
	"sync"
	"time"
)

// RandGenerator is a unified random number generator.
// It provides thread-safe random number generation.
type RandGenerator struct {
	rng *rand.Rand
	mu  sync.Mutex
}

var (
	globalRandGen *RandGenerator
	initOnce      sync.Once
)

// GetGlobalRandGenerator returns the global random number generator singleton
func GetGlobalRandGenerator() *RandGenerator {
	initOnce.Do(func() {
		globalRandGen = NewRandGenerator()
	})
	return globalRandGen
}

// NewRandGenerator creates a new random number generator.
// Uses current time in nanoseconds as seed, with an additional random offset.
func NewRandGenerator() *RandGenerator {
	// Use timestamp + goroutine ID (indirectly via multiple calls) to increase randomness
	seed := time.Now().UnixNano()
	// Add extra randomness: use default random source to generate an offset
	seed += int64(rand.Intn(10000))

	return &RandGenerator{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// Intn returns a random integer in the range [0, n)
func (r *RandGenerator) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Intn(n)
}

// Int63n returns a random int64 in the range [0, n)
func (r *RandGenerator) Int63n(n int64) int64 {
	if n <= 0 {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Int63n(n)
}

// Shuffle randomly shuffles a slice
func (r *RandGenerator) Shuffle(n int, swap func(i, j int)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rng.Shuffle(n, swap)
}

// RandomChoice randomly selects an element from a slice
func RandomChoice[T any](items []T) T {
	if len(items) == 0 {
		var zero T
		return zero
	}
	gen := GetGlobalRandGenerator()
	return items[gen.Intn(len(items))]
}

// RandomChoiceString randomly selects a string from a string slice
func RandomChoiceString(items []string) string {
	return RandomChoice(items)
}
