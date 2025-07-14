package site

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sync"
	"time"
	"wispy-core/auth"
	"wispy-core/core/tenant/databases"
	"wispy-core/tpl"

	"github.com/go-chi/chi/v5"
)

// Site represents a tenant instance with atomic access patterns
type site struct {
	mu      sync.RWMutex
	ID      string `toml:"id" json:"id"`
	Name    string `toml:"name" json:"name"`
	Domain  string `toml:"domain" json:"domain"`
	BaseURL string `toml:"base_url" json:"base_url"`
	//
	CssThemes map[string]string // Maps theme name to CSS file path
	// ContentDir         string                 `toml:"content_dir" json:"content_dir"`
	Data   map[string]interface{} `toml:"data" json:"data"`
	Config map[string]interface{} `toml:"config" json:"config"` // Site configuration from config.toml
	//
	Router         chi.Router         `toml:"-" json:"-"`
	TemplateEngine tpl.TemplateEngine `toml:"-" json:"-"`
	// DatabaseManager is used for database operations, if applicable
	DbManager databases.Manager `toml:"-" json:"-"`
	//
	AuthManager auth.AuthProvider `toml:"-" json:"-"`
	//
	// CreatedAt and UpdatedAt are used for tracking site creation and modification times
	CreatedAt time.Time `toml:"created_at" json:"created_at"`
	UpdatedAt time.Time `toml:"updated_at" json:"updated_at"`
}

type Site interface {
	GetID() string
	GetName() string
	GetDomain() string
	GetBaseURL() string
	// GetContentDir() string
	GetStaticDir() string
	GetAssetsDir() string
	GetConfig() map[string]interface{}
	GetData() map[string]interface{}
	SetData(key string, value interface{})
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetUpdatedAt(t time.Time)
	GetRouter() chi.Router
	GetDatabaseManager() databases.Manager
	//
	GetTheme(name string) (string, error) // Get CSS theme for a domain
}

// Page represents a rendered page in the system
type Page struct {
	ID          string                 `toml:"id" json:"id"`
	Title       string                 `toml:"title" json:"title"`
	Slug        string                 `toml:"slug" json:"slug"`
	Path        string                 `toml:"path" json:"path"` // Path to the page file
	Layout      string                 `toml:"layout" json:"layout"`
	Theme       string                 `toml:"theme" json:"theme"`
	Content     string                 `toml:"content" json:"content"`
	FrontMatter map[string]interface{} `toml:"front_matter" json:"front_matter"`
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
	// Return a copy of the Config map to prevent external modification
	configCopy := make(map[string]interface{})
	maps.Copy(configCopy, s.Config)
	return configCopy
}

func (s *site) GetDatabaseManager() databases.Manager {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.DbManager
}

func (s *site) GetTemplateEngine() tpl.TemplateEngine {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.TemplateEngine == nil {
		// Return a default implementation or nil if not set
		return nil
	}
	return s.TemplateEngine
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

func (s *site) GetRouter() chi.Router {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.Router == nil {
		// Initialize the router if it hasn't been set
		s.Router = chi.NewRouter()
	}
	return s.Router
}

func (s *site) GetTheme(name string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the domain has a specific theme
	if theme, exists := s.CssThemes[name]; exists {
		return theme, nil
	}

	// Get theme from site folder
	themeBytes, err := os.ReadFile(filepath.Join("_data", "tenants", s.Domain, "themes", name+".css"))
	if err != nil {
		return "", fmt.Errorf("failed to read theme file: %w", err)
	}

	// If no specific theme is set, return a default or empty string
	return string(themeBytes), nil
}
