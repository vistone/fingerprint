package gateway

import (
	"container/list"
	"sync"
	"time"
)

// DefaultCacheSize is the default maximum number of cache entries
const DefaultCacheSize = 10000

// CacheEntry represents a single cache entry
type cacheEntry struct {
	key        string
	value      interface{}
	expiration time.Time
}

// isExpired returns true if the entry has expired
func (e *cacheEntry) isExpired() bool {
	return time.Now().After(e.expiration)
}

// LRUCache implements a thread-safe LRU cache with TTL
type LRUCache struct {
	mu         sync.RWMutex
	entries    map[string]*list.Element
	order      *list.List
	maxEntries int
	defaultTTL time.Duration
	hits       int64
	misses     int64
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(maxEntries int, defaultTTL time.Duration) *LRUCache {
	if maxEntries <= 0 {
		maxEntries = DefaultCacheSize
	}
	if defaultTTL <= 0 {
		defaultTTL = time.Minute * 5
	}

	return &LRUCache{
		entries:    make(map[string]*list.Element, maxEntries),
		order:      list.New(),
		maxEntries: maxEntries,
		defaultTTL: defaultTTL,
	}
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, found := c.entries[key]
	if !found {
		c.misses++
		return nil, false
	}

	entry := elem.Value.(*cacheEntry)
	if entry.isExpired() {
		c.removeElement(elem)
		c.misses++
		return nil, false
	}

	// Move to front (most recently used)
	c.order.MoveToFront(elem)
	c.hits++
	return entry.value, true
}

// Set adds or updates a cache entry
func (c *LRUCache) Set(key string, value interface{}, ttl time.Duration) {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing entry
	if elem, found := c.entries[key]; found {
		c.order.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		entry.value = value
		entry.expiration = time.Now().Add(ttl)
		return
	}

	// Add new entry
	entry := &cacheEntry{
		key:        key,
		value:      value,
		expiration: time.Now().Add(ttl),
	}
	elem := c.order.PushFront(entry)
	c.entries[key] = elem

	// Evict oldest entries if over capacity
	for len(c.entries) > c.maxEntries {
		c.evictOldest()
	}
}

// Delete removes an entry from the cache
func (c *LRUCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, found := c.entries[key]; found {
		c.removeElement(elem)
		return true
	}
	return false
}

// Clear removes all entries from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*list.Element, c.maxEntries)
	c.order = list.New()
}

// Reconfigure updates capacity and default TTL while preserving existing entries.
func (c *LRUCache) Reconfigure(maxEntries int, defaultTTL time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if maxEntries <= 0 {
		maxEntries = DefaultCacheSize
	}
	if defaultTTL <= 0 {
		defaultTTL = time.Minute * 5
	}

	c.maxEntries = maxEntries
	c.defaultTTL = defaultTTL

	for len(c.entries) > c.maxEntries {
		c.evictOldest()
	}
}

// Len returns the current number of entries
func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// evictOldest removes the oldest entry
func (c *LRUCache) evictOldest() {
	elem := c.order.Back()
	if elem != nil {
		c.removeElement(elem)
	}
}

// removeElement removes an element from the cache
func (c *LRUCache) removeElement(elem *list.Element) {
	c.order.Remove(elem)
	entry := elem.Value.(*cacheEntry)
	delete(c.entries, entry.key)
}

// CacheStats holds cache statistics
type CacheStats struct {
	Entries int
	Hits    int64
	Misses  int64
	HitRate float64
}

// Stats returns cache statistics
func (c *LRUCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Entries: len(c.entries),
		Hits:    c.hits,
		Misses:  c.misses,
		HitRate: hitRate,
	}
}

// ClassificationCacheEntry holds classification results
type ClassificationCacheEntry struct {
	Protocol   string
	Family     string
	Version    string
	Confidence float64
	JA3        string
	JA4        string
	Timestamp  time.Time
}

// ClassificationCache is a specialized cache for classification results
type ClassificationCache struct {
	cache          *LRUCache
	stopCleanup    chan struct{}
	cleanupRunning bool
}

// NewClassificationCache creates a new classification cache
func NewClassificationCache(maxEntries int, defaultTTL time.Duration) *ClassificationCache {
	return &ClassificationCache{
		cache:       NewLRUCache(maxEntries, defaultTTL),
		stopCleanup: make(chan struct{}),
	}
}

// Get retrieves a classification result
func (cc *ClassificationCache) Get(key string) (*ClassificationCacheEntry, bool) {
	val, found := cc.cache.Get(key)
	if !found {
		return nil, false
	}
	return val.(*ClassificationCacheEntry), true
}

// Set stores a classification result
func (cc *ClassificationCache) Set(key string, entry *ClassificationCacheEntry, ttl time.Duration) {
	cc.cache.Set(key, entry, ttl)
}

// SetWithConfidenceTTL stores with TTL based on confidence level
func (cc *ClassificationCache) SetWithConfidenceTTL(key string, entry *ClassificationCacheEntry) {
	var ttl time.Duration

	switch {
	case entry.Confidence >= 0.9:
		ttl = time.Hour // High confidence: cache longer
	case entry.Confidence >= 0.7:
		ttl = time.Minute * 10 // Medium confidence
	default:
		ttl = time.Minute // Low confidence: cache briefly
	}

	cc.Set(key, entry, ttl)
}

// Len returns the number of cached entries
func (cc *ClassificationCache) Len() int {
	return cc.cache.Len()
}

// StartCleanup starts the background cleanup goroutine
func (cc *ClassificationCache) StartCleanup(interval time.Duration) {
	cc.cache.mu.Lock()
	if cc.cleanupRunning {
		cc.cache.mu.Unlock()
		return
	}
	cc.cleanupRunning = true
	cc.stopCleanup = make(chan struct{})
	cc.cache.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				cc.performCleanup()
			case <-cc.stopCleanup:
				return
			}
		}
	}()
}

// StopCleanup stops the background cleanup
func (cc *ClassificationCache) StopCleanup() {
	cc.cache.mu.Lock()
	defer cc.cache.mu.Unlock()

	if cc.cleanupRunning {
		cc.cleanupRunning = false
		close(cc.stopCleanup)
	}
}

// performCleanup removes expired entries
func (cc *ClassificationCache) performCleanup() {
	cc.cache.mu.Lock()
	defer cc.cache.mu.Unlock()

	var toRemove []*list.Element

	for elem := cc.cache.order.Back(); elem != nil; elem = elem.Prev() {
		entry := elem.Value.(*cacheEntry)
		if entry.isExpired() {
			toRemove = append(toRemove, elem)
		}
	}

	for _, elem := range toRemove {
		cc.cache.removeElement(elem)
	}
}
