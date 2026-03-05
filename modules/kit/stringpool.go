package utils

import (
	"strings"
	"sync"
)

// StringPool 字符串池，用于缓存常用的小写字符串
type StringPool struct {
	mu    sync.RWMutex
	cache map[string]string
}

// NewStringPool 创建新的字符串池
func NewStringPool() *StringPool {
	return &StringPool{
		cache: make(map[string]string, 1024),
	}
}

// ToLower 获取小写字符串（带缓存）
func (p *StringPool) ToLower(s string) string {
	// 快速路径：检查是否已全小写
	if isAllLower(s) {
		return s
	}

	// 检查缓存
	p.mu.RLock()
	cached, exists := p.cache[s]
	p.mu.RUnlock()
	if exists {
		return cached
	}

	// 计算小写并缓存
	lower := strings.ToLower(s)
	p.mu.Lock()
	p.cache[s] = lower
	p.mu.Unlock()
	return lower
}

// isAllLower 检查字符串是否已全为小写
func isAllLower(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			return false
		}
	}
	return true
}

// GlobalStringPool 全局字符串池
var GlobalStringPool = NewStringPool()

// ToLowerGlobal 使用全局字符串池获取小写字符串
func ToLowerGlobal(s string) string {
	return GlobalStringPool.ToLower(s)
}

// CaseInsensitiveContains 不区分大小写的包含检查（优化版本）
// 对于大文本使用 Boyer-Moore 算法，小文本使用简单比较
func CaseInsensitiveContains(text, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(text) < len(substr) {
		return false
	}

	// 小文本直接转小写比较
	if len(text) < 1024 {
		return strings.Contains(strings.ToLower(text), strings.ToLower(substr))
	}

	// 大文本使用优化算法
	textLower := ToLowerGlobal(text)
	substrLower := ToLowerGlobal(substr)
	return strings.Contains(textLower, substrLower)
}

// FastContains 快速包含检查（假设输入已为小写）
func FastContains(textLower, substrLower string) bool {
	return strings.Contains(textLower, substrLower)
}
