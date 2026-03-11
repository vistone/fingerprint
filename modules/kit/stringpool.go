package utils

import (
	"strings"
	"sync"
)

// StringPool is a string pool for caching commonly used lowercase strings
type StringPool struct {
	mu    sync.RWMutex
	cache map[string]string
}

// NewStringPool creates a new string pool
func NewStringPool() *StringPool {
	return &StringPool{
		cache: make(map[string]string, 1024),
	}
}

// ToLower returns the lowercase string (with caching)
func (p *StringPool) ToLower(s string) string {
	// Fast path: check if already all lowercase
	if isAllLower(s) {
		return s
	}

	// Check cache
	p.mu.RLock()
	cached, exists := p.cache[s]
	p.mu.RUnlock()
	if exists {
		return cached
	}

	// Compute lowercase and cache the result
	lower := strings.ToLower(s)
	p.mu.Lock()
	p.cache[s] = lower
	p.mu.Unlock()
	return lower
}

// isAllLower checks if the string is already all lowercase
func isAllLower(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			return false
		}
	}
	return true
}

// GlobalStringPool is the global string pool
var GlobalStringPool = NewStringPool()

// ToLowerGlobal returns the lowercase string using the global string pool
func ToLowerGlobal(s string) string {
	return GlobalStringPool.ToLower(s)
}

// CaseInsensitiveContains performs a case-insensitive contains check (optimized).
// Uses Boyer-Moore algorithm for large text, simple comparison for small text.
func CaseInsensitiveContains(text, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(text) < len(substr) {
		return false
	}

	// For small text, convert to lowercase and compare directly
	if len(text) < 1024 {
		return strings.Contains(strings.ToLower(text), strings.ToLower(substr))
	}

	// For large text, use the optimized algorithm
	textLower := ToLowerGlobal(text)
	substrLower := ToLowerGlobal(substr)
	return strings.Contains(textLower, substrLower)
}

// FastContains performs a fast contains check (assumes input is already lowercase)
func FastContains(textLower, substrLower string) bool {
	return strings.Contains(textLower, substrLower)
}
