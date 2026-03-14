package waf

import (
	"sync"
	"time"
)

// BlockList manages blocked IPs with expiration
type BlockList struct {
	blocks   map[string]*BlockEntry
	duration time.Duration
	mu       sync.RWMutex
}

// BlockEntry represents a block entry
type BlockEntry struct {
	IP        string
	BlockedAt time.Time
	ExpiresAt time.Time
	Reason    string
}

// NewBlockList creates a new block list
func NewBlockList(duration time.Duration) *BlockList {
	bl := &BlockList{
		blocks:   make(map[string]*BlockEntry),
		duration: duration,
	}

	// Start cleanup goroutine
	go bl.cleanup()

	return bl
}

// Block adds an IP to the block list
func (bl *BlockList) Block(ip, reason string) {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	now := time.Now()
	bl.blocks[ip] = &BlockEntry{
		IP:        ip,
		BlockedAt: now,
		ExpiresAt: now.Add(bl.duration),
		Reason:    reason,
	}
}

// Unblock removes an IP from the block list
func (bl *BlockList) Unblock(ip string) {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	delete(bl.blocks, ip)
}

// IsBlocked checks if an IP is blocked
func (bl *BlockList) IsBlocked(ip string) bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	entry, exists := bl.blocks[ip]
	if !exists {
		return false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		return false
	}

	return true
}

// RemainingTime returns the remaining block duration for an IP
func (bl *BlockList) RemainingTime(ip string) time.Duration {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	entry, exists := bl.blocks[ip]
	if !exists {
		return 0
	}

	remaining := entry.ExpiresAt.Sub(time.Now())
	if remaining < 0 {
		return 0
	}

	return remaining
}

// GetAll returns all active blocks
func (bl *BlockList) GetAll() []*BlockEntry {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	result := make([]*BlockEntry, 0)
	now := time.Now()

	for _, entry := range bl.blocks {
		if now.Before(entry.ExpiresAt) {
			result = append(result, entry)
		}
	}

	return result
}

// cleanup removes expired entries periodically
func (bl *BlockList) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		bl.mu.Lock()
		now := time.Now()
		for ip, entry := range bl.blocks {
			if now.After(entry.ExpiresAt) {
				delete(bl.blocks, ip)
			}
		}
		bl.mu.Unlock()
	}
}

// Stop stops the cleanup goroutine
func (bl *BlockList) Stop() {
	// Signal cleanup to stop if needed
	// For simplicity, we let it run until program termination
}
