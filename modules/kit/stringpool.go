package utils

import (
	"strings"
	"sync"
)

// translated comment
type StringPool struct {
	mu    sync.RWMutex
	cache map[string]string
}

// translated comment
func NewStringPool() *StringPool {
	return &StringPool{
		cache: make(map[string]string, 1024),
	}
}

// translated comment
func (p *StringPool) ToLower(s string) string {
	// translated comment
	if isAllLower(s) {
		return s
	}

	// translated comment
	p.mu.RLock()
	cached, exists := p.cache[s]
	p.mu.RUnlock()
	if exists {
		return cached
	}

	// translated comment
	lower := strings.ToLower(s)
	p.mu.Lock()
	p.cache[s] = lower
	p.mu.Unlock()
	return lower
}

// translated comment
func isAllLower(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			return false
		}
	}
	return true
}

// translated comment
var GlobalStringPool = NewStringPool()

// translated comment
func ToLowerGlobal(s string) string {
	return GlobalStringPool.ToLower(s)
}

// translated comment
// translated comment
func CaseInsensitiveContains(text, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(text) < len(substr) {
		return false
	}

	// translated comment
	if len(text) < 1024 {
		return strings.Contains(strings.ToLower(text), strings.ToLower(substr))
	}

	// translated comment
	textLower := ToLowerGlobal(text)
	substrLower := ToLowerGlobal(substr)
	return strings.Contains(textLower, substrLower)
}

// translated comment
func FastContains(textLower, substrLower string) bool {
	return strings.Contains(textLower, substrLower)
}
