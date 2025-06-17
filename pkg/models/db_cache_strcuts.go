package models

import (
	"database/sql"
	"sync"
	"time"
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
