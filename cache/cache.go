package cache

import (
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
}

// Cache represents an in-memory cache with expiration
type Cache struct {
	items map[string]*CacheItem
	mutex sync.RWMutex
}

// NewCache creates a new cache instance
func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]*CacheItem),
	}

	// Start cleanup routine
	go cache.cleanup()

	return cache
}

// Set stores a value in the cache with expiration
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(duration),
	}
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if item has expired
	if time.Now().After(item.ExpiresAt) {
		// Clean up expired item
		go func() {
			c.mutex.Lock()
			delete(c.items, key)
			c.mutex.Unlock()
		}()
		return nil, false
	}

	return item.Value, true
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*CacheItem)
}

// cleanup removes expired items from the cache
func (c *Cache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mutex.Lock()
			now := time.Now()
			for key, item := range c.items {
				if now.After(item.ExpiresAt) {
					delete(c.items, key)
				}
			}
			c.mutex.Unlock()
		}
	}
}

// Size returns the number of items in the cache
func (c *Cache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.items)
}

// Keys returns all keys in the cache
func (c *Cache) Keys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}

	return keys
}
