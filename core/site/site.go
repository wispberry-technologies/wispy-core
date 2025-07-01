package site

import (
	"maps"
	"sync"
	"time"
	"wispy-core/core"
	"wispy-core/core/site/theme"

	"github.com/go-chi/chi/v5"
)

// Site represents a tenant instance with atomic access patterns
type site struct {
	mu                 sync.RWMutex
	ID                 string                 `toml:"id" json:"id"`
	Name               string                 `toml:"name" json:"name"`
	Domain             string                 `toml:"domain" json:"domain"`
	BaseURL            string                 `toml:"base_url" json:"base_url"`
	Theme              *theme.Root            `toml:"theme" json:"theme"`
	ContentDir         string                 `toml:"content_dir" json:"content_dir"`
	Data               map[string]interface{} `toml:"data" json:"data"`
	Router             *chi.Mux               `toml:"-" json:"-"`
	SiteTemplateEngine core.SiteTplEngine     `toml:"-" json:"-"`
	DbManager          core.DatabaseManager   `toml:"-" json:"-"` // DatabaseManager is used for database operations, if applicable
	// CreatedAt and UpdatedAt are used for tracking site creation and modification times
	CreatedAt time.Time `toml:"created_at" json:"created_at"`
	UpdatedAt time.Time `toml:"updated_at" json:"updated_at"`
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

func (s *site) GetTheme() *theme.Root {
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

func (s *site) GetStaticDir() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Assuming static directory is a subdirectory of content
	return "/static"
}

func (s *site) GetAssetsDir() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Assuming assets directory is a subdirectory of content
	return "/assets"
}

func (s *site) GetConfig() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Return a copy of the Data map to prevent external modification
	dataCopy := make(map[string]interface{})
	maps.Copy(dataCopy, map[string]interface{}{
		"FAKE_CONFIG": "This is a placeholder for actual configuration data",
		"id":          s.ID,
		"name":        s.Name,
		"domain":      s.Domain,
		"base_url":    s.BaseURL,
		"content_dir": s.ContentDir,
	})
	return dataCopy
}

func (s *site) GetDatabaseManager() core.DatabaseManager {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.DbManager
}

func (s *site) GetTemplateEngine() core.SiteTplEngine {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.SiteTemplateEngine == nil {
		// Return a default implementation or nil if not set
		return nil
	}
	return s.SiteTemplateEngine
}

func (s *site) GetData() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Return a copy to prevent external modification
	dataCopy := make(map[string]interface{})
	maps.Copy(dataCopy, s.Data)
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

func (s *site) GetRouter() *chi.Mux {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.Router == nil {
		s.Router = chi.NewRouter()
	}
	return s.Router
}
