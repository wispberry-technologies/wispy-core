package models

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
	Items map[string]*CacheItem
	Mutex sync.RWMutex
}
