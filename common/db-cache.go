package common

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"time"

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
	connections map[string]*CachedConnection
	mutex       sync.RWMutex
	maxAge      time.Duration
}

// NewDBCache creates a new database connection cache
func NewDBCache() *DBCache {
	cache := &DBCache{
		connections: make(map[string]*CachedConnection),
		maxAge:      time.Hour, // Close connections older than 1 hour
	}

	// Start cleanup goroutine
	go cache.cleanupStaleConnections()

	return cache
}

// GetConnection retrieves or creates a database connection for a site
func (c *DBCache) GetConnection(domain, dbName string) (*sql.DB, error) {
	cacheKey := fmt.Sprintf("%s:%s", domain, dbName)

	c.mutex.RLock()
	cached, exists := c.connections[cacheKey]
	c.mutex.RUnlock()

	// Check if connection exists and is still valid
	if exists {
		// Test if connection is still alive
		if err := cached.DB.Ping(); err == nil {
			// Update last used time
			c.mutex.Lock()
			cached.LastUsed = time.Now()
			c.mutex.Unlock()
			return cached.DB, nil
		} else {
			// Connection is dead, remove from cache
			c.removeConnection(cacheKey)
		}
	}

	// Create new connection
	return c.createConnection(domain, dbName)
}

// createConnection creates a new database connection and caches it
func (c *DBCache) createConnection(domain, dbName string) (*sql.DB, error) {
	// Construct the database path
	dbPath := filepath.Join(MustGetEnv("SITES_PATH"), domain, "dbs", dbName+".db")

	// Validate path security
	safePath, err := ValidatePath(dbPath)
	if err != nil {
		return nil, fmt.Errorf("invalid database path: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite", safePath+"?_journal_mode=WAL&_timeout=5000")
	if err != nil {
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

	c.mutex.Lock()
	c.connections[cacheKey] = cached
	c.mutex.Unlock()

	return db, nil
}

// removeConnection removes a connection from cache and closes it
func (c *DBCache) removeConnection(cacheKey string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if cached, exists := c.connections[cacheKey]; exists {
		cached.DB.Close()
		delete(c.connections, cacheKey)
	}
}

// cleanupStaleConnections periodically removes stale connections
func (c *DBCache) cleanupStaleConnections() {
	ticker := time.NewTicker(15 * time.Minute) // Check every 15 minutes
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()

		for key, cached := range c.connections {
			if now.Sub(cached.LastUsed) > c.maxAge {
				cached.DB.Close()
				delete(c.connections, key)
			}
		}

		c.mutex.Unlock()
	}
}

// CloseAll closes all cached connections
func (c *DBCache) CloseAll() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, cached := range c.connections {
		cached.DB.Close()
	}

	c.connections = make(map[string]*CachedConnection)
}

// GetCacheStats returns statistics about the cache
func (c *DBCache) GetCacheStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_connections": len(c.connections),
		"max_age_hours":     c.maxAge.Hours(),
		"connections":       make([]map[string]interface{}, 0, len(c.connections)),
	}

	for key, cached := range c.connections {
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
