package utils

import (
	"sync"
	"time"
)

// translated comment
type Cache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
}

type cacheItem struct {
	value      interface{}
	expiration int64
}

// translated comment
func NewCache() *Cache {
	c := &Cache{
		items: make(map[string]cacheItem),
	}
	// translated comment
	go c.cleanup()
	return c
}

// translated comment
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := time.Now().Add(ttl).UnixNano()
	c.items[key] = cacheItem{
		value:      value,
		expiration: expiration,
	}
}

// translated comment
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// translated comment
	if time.Now().UnixNano() > item.expiration {
		return nil, false
	}

	return item.value, true
}

// translated comment
func (c *Cache) GetString(key string) (string, bool) {
	val, ok := c.Get(key)
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// translated comment
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// translated comment
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]cacheItem)
}

// translated comment
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

// translated comment
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

// translated comment
func NewLRUCache(maxSize int) *LRUCache {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &LRUCache{
		items:   make(map[string]*lruItem),
		maxSize: maxSize,
	}
}

// translated comment
func (c *LRUCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := time.Now().Add(ttl).UnixNano()

	// translated comment
	if item, exists := c.items[key]; exists {
		item.value = value
		item.expiration = expiration
		c.moveToHead(item)
		return
	}

	// translated comment
	newItem := &lruItem{
		key:        key,
		value:      value,
		expiration: expiration,
	}
	c.items[key] = newItem
	c.addToHead(newItem)

	// translated comment
	if len(c.items) > c.maxSize {
		c.removeTail()
	}
}

// translated comment
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// translated comment
	if time.Now().UnixNano() > item.expiration {
		c.removeItem(item)
		return nil, false
	}

	// translated comment
	c.moveToHead(item)
	return item.value, true
}

// translated comment
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

// translated comment
func (c *LRUCache) moveToHead(item *lruItem) {
	if item == c.head {
		return
	}
	c.removeItem(item)
	c.addToHead(item)
}

// translated comment
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

// translated comment
func (c *LRUCache) removeTail() {
	if c.tail == nil {
		return
	}
	delete(c.items, c.tail.key)
	c.removeItem(c.tail)
}
