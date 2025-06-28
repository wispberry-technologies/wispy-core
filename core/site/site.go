package site

import (
	"sync"
	"time"
)

// Site represents a tenant instance with atomic access patterns
type site struct {
	mu         sync.RWMutex
	ID         string                 `toml:"id" json:"id"`
	Name       string                 `toml:"name" json:"name"`
	Domain     string                 `toml:"domain" json:"domain"`
	BaseURL    string                 `toml:"base_url" json:"base_url"`
	Theme      *Theme                 `toml:"theme" json:"theme"`
	ContentDir string                 `toml:"content_dir" json:"content_dir"`
	Data       map[string]interface{} `toml:"data" json:"data"`
	CreatedAt  time.Time              `toml:"created_at" json:"created_at"`
	UpdatedAt  time.Time              `toml:"updated_at" json:"updated_at"`
}

type Site interface {
	GetID() string
	GetName() string
	GetDomain() string
	GetBaseURL() string
	GetTheme() *Theme
	GetContentDir() string
	GetData() map[string]interface{}
	SetData(key string, value interface{})
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetUpdatedAt(t time.Time)
}

func (s *site) GetID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ID
}

func (s *site) GetName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Name
}

func (s *site) GetDomain() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Domain
}

func (s *site) GetBaseURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.BaseURL
}

func (s *site) GetTheme() *Theme {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Return a copy of the theme to prevent external modification
	themeCopy := *s.Theme
	return &themeCopy
}

func (s *site) GetContentDir() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ContentDir
}

func (s *site) GetData() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Return a copy to prevent external modification
	dataCopy := make(map[string]interface{})
	for k, v := range s.Data {
		dataCopy[k] = v
	}
	return dataCopy
}

func (s *site) SetData(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Data[key] = value
}

func (s *site) GetCreatedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CreatedAt
}

func (s *site) GetUpdatedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.UpdatedAt
}

func (s *site) SetUpdatedAt(t time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.UpdatedAt = t
}
