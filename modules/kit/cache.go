package utils

import (
	"sync"
	"time"
)

// Cache is a simple concurrent-safe cache
type Cache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
}

type cacheItem struct {
	value      interface{}
	expiration int64
}

// NewCache creates a new cache
func NewCache() *Cache {
	c := &Cache{
		items: make(map[string]cacheItem),
	}
	// Start cleanup goroutine
	go c.cleanup()
	return c
}

// Set sets a cache item
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := time.Now().Add(ttl).UnixNano()
	c.items[key] = cacheItem{
		value:      value,
		expiration: expiration,
	}
}

// Get retrieves a cache item
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().UnixNano() > item.expiration {
		return nil, false
	}

	return item.value, true
}

// GetString retrieves a string cache item
func (c *Cache) GetString(key string) (string, bool) {
	val, ok := c.Get(key)
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// Delete removes a cache item
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear empties the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]cacheItem)
}

// cleanup periodically removes expired items
func (c *Cache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now().UnixNano()
		for key, item := range c.items {
			if now > item.expiration {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// LRUCache is a simple LRU cache
type LRUCache struct {
	mu       sync.RWMutex
	items    map[string]*lruItem
	maxSize  int
	head     *lruItem
	tail     *lruItem
}

type lruItem struct {
	key        string
	value      interface{}
	expiration int64
	prev       *lruItem
	next       *lruItem
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(maxSize int) *LRUCache {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &LRUCache{
		items:   make(map[string]*lruItem),
		maxSize: maxSize,
	}
}

// Set sets a cache item
func (c *LRUCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := time.Now().Add(ttl).UnixNano()

	// 如果已存在，更新值并移动到头部
	if item, exists := c.items[key]; exists {
		item.value = value
		item.expiration = expiration
		c.moveToHead(item)
		return
	}

	// 创建新项
	newItem := &lruItem{
		key:        key,
		value:      value,
		expiration: expiration,
	}
	c.items[key] = newItem
	c.addToHead(newItem)

	// 如果超过最大大小，移除尾部
	if len(c.items) > c.maxSize {
		c.removeTail()
	}
}

// Get retrieves a cache item
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().UnixNano() > item.expiration {
		c.removeItem(item)
		return nil, false
	}

	// 移动到头部（最近使用）
	c.moveToHead(item)
	return item.value, true
}

// addToHead adds an item to the head
func (c *LRUCache) addToHead(item *lruItem) {
	item.prev = nil
	item.next = c.head
	if c.head != nil {
		c.head.prev = item
	}
	c.head = item
	if c.tail == nil {
		c.tail = item
	}
}

// moveToHead moves an item to the head
func (c *LRUCache) moveToHead(item *lruItem) {
	if item == c.head {
		return
	}
	c.removeItem(item)
	c.addToHead(item)
}

// removeItem removes an item
func (c *LRUCache) removeItem(item *lruItem) {
	if item.prev != nil {
		item.prev.next = item.next
	} else {
		c.head = item.next
	}
	if item.next != nil {
		item.next.prev = item.prev
	} else {
		c.tail = item.prev
	}
}

// removeTail removes the tail item (least recently used)
func (c *LRUCache) removeTail() {
	if c.tail == nil {
		return
	}
	delete(c.items, c.tail.key)
	c.removeItem(c.tail)
}
