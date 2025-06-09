package cache

import (
	"database/sql"
	"time"
	"wispy-core/models"
)

// GetDB returns a database connection for the specified site and database
func GetDB(instance *models.SiteInstance, dbName string) (*sql.DB, error) {
	return GetConnection(instance.DBCache, instance.Site.Domain, dbName)
}

// NewCache creates a new cache instance
func NewCache() *models.Cache {
	cache := &models.Cache{
		Items: make(map[string]*models.CacheItem),
	}

	// Start cleanup routine
	Cleanup(cache)

	return cache
}

// Set stores a value in the cache with expiration
func Set(c *models.Cache, key string, value interface{}, duration time.Duration) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	c.Items[key] = &models.CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(duration),
	}
}

// Get retrieves a value from the cache
func Get(c *models.Cache, key string) (interface{}, bool) {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	item, exists := c.Items[key]
	if !exists {
		return nil, false
	}

	// Check if item has expired
	if time.Now().After(item.ExpiresAt) {
		// Clean up expired item
		go func() {
			c.Mutex.Lock()
			delete(c.Items, key)
			c.Mutex.Unlock()
		}()
		return nil, false
	}

	return item.Value, true
}

// Delete removes a value from the cache
func Delete(c *models.Cache, key string) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	delete(c.Items, key)
}

// Clear removes all items from the cache
func Clear(c *models.Cache) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	c.Items = make(map[string]*models.CacheItem)
}

// cleanup removes expired items from the cache
func Cleanup(c *models.Cache) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.Mutex.Lock()
			now := time.Now()
			for key, item := range c.Items {
				if now.After(item.ExpiresAt) {
					delete(c.Items, key)
				}
			}
			c.Mutex.Unlock()
		}
	}
}

// Size returns the number of items in the cache
func Size(c *models.Cache) int {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	return len(c.Items)
}

// Keys returns all keys in the cache
func Keys(c *models.Cache) []string {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	keys := make([]string, 0, len(c.Items))
	for key := range c.Items {
		keys = append(keys, key)
	}

	return keys
}
