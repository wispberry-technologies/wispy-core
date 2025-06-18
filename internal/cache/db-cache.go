package cache

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
	"wispy-core/pkg/common"

	_ "github.com/glebarez/go-sqlite"
)

// CachedConnection represents a cached database connection with metadata
type CachedConnection struct {
	DB       *sql.DB
	LastUsed time.Time
	Domain   string
	DBPath   string
	DBName   string
}

// DBCache manages database connections with automatic cleanup of stale connections
type DBCache struct {
	Connections map[string]*CachedConnection
	Mutex       sync.RWMutex
	MaxAge      time.Duration
}

// NewDBCache creates a new database connection cache
func NewDBCache() *DBCache {
	cache := &DBCache{
		Connections: make(map[string]*CachedConnection),
		MaxAge:      time.Hour, // Close connections older than 1 hour
	}

	// Start cleanup goroutine
	go CleanupStaleConnections(cache)

	return cache
}

// GetConnection retrieves or creates a database connection for a site
func GetConnection(c *DBCache, domain, dbName string) (*sql.DB, error) {
	cacheKey := fmt.Sprintf("%s:%s", domain, dbName)

	c.Mutex.RLock()
	cached, exists := c.Connections[cacheKey]
	c.Mutex.RUnlock()

	// Check if connection exists and is still valid
	if exists {
		// Test if connection is still alive
		if err := cached.DB.Ping(); err == nil {
			// Update last used time
			c.Mutex.Lock()
			cached.LastUsed = time.Now()
			c.Mutex.Unlock()
			return cached.DB, nil
		} else {
			// Connection is dead, remove from cache
			RemoveConnection(c, cacheKey)
		}
	}

	// Create new connection
	return CreateConnection(c, domain, dbName)
}

// CreateConnection creates a new database connection and caches it
func CreateConnection(c *DBCache, domain, dbName string) (*sql.DB, error) {
	// Construct the database path
	safePath := common.RootSitesPath(domain, "dbs", dbName+".db")

	// Open database connection
	db, err := sql.Open("sqlite", safePath+"?_journal_mode=WAL&_timeout=5000")
	if err != nil {
		common.Error("Failed to open database", "domain", domain, "dbName", dbName, "error", err)
		return nil, fmt.Errorf("failed to open database %s for site %s: %w", dbName, domain, err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database %s for site %s: %w", dbName, domain, err)
	}

	// Cache the connection
	cacheKey := fmt.Sprintf("%s:%s", domain, dbName)
	cached := &CachedConnection{
		DB:       db,
		LastUsed: time.Now(),
		Domain:   domain,
		DBPath:   safePath,
		DBName:   dbName,
	}

	c.Mutex.Lock()
	c.Connections[cacheKey] = cached
	c.Mutex.Unlock()

	return db, nil
}

// RemoveConnection removes a connection from cache and closes it
func RemoveConnection(c *DBCache, cacheKey string) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if cached, exists := c.Connections[cacheKey]; exists {
		cached.DB.Close()
		delete(c.Connections, cacheKey)
	}
}

// CleanupStaleConnections periodically removes stale connections
func CleanupStaleConnections(c *DBCache) {
	ticker := time.NewTicker(15 * time.Minute) // Check every 15 minutes
	defer ticker.Stop()

	for range ticker.C {
		c.Mutex.Lock()
		now := time.Now()

		for key, cached := range c.Connections {
			if now.Sub(cached.LastUsed) > c.MaxAge {
				cached.DB.Close()
				delete(c.Connections, key)
			}
		}

		c.Mutex.Unlock()
	}
}

// CloseAll closes all cached connections
func CloseAll(c *DBCache) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	for _, cached := range c.Connections {
		cached.DB.Close()
	}

	c.Connections = make(map[string]*CachedConnection)
}

// GetCacheStats returns statistics about the cache
func GetCacheStats(c *DBCache) map[string]interface{} {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	stats := map[string]interface{}{
		"total_connections": len(c.Connections),
		"max_age_hours":     c.MaxAge.Hours(),
		"connections":       make([]map[string]interface{}, 0, len(c.Connections)),
	}

	for key, cached := range c.Connections {
		connInfo := map[string]interface{}{
			"key":         key,
			"domain":      cached.Domain,
			"db_name":     cached.DBName,
			"last_used":   cached.LastUsed,
			"age_minutes": time.Since(cached.LastUsed).Minutes(),
		}
		stats["connections"] = append(stats["connections"].([]map[string]interface{}), connInfo)
	}

	return stats
}
