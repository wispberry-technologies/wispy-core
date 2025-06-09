package models

import (
	"net/http"
	"sync"
	"time"
)

// Site represents a single site in the multisite system
type Site struct {
	Domain   string `json:"domain"`
	Name     string `json:"name"`
	BasePath string `json:"base_path"`
	IsActive bool   `json:"is_active"`
	Theme    string `json:"theme"`
}

type SiteConfig struct {
	MaxFailedLoginAttempts         int
	FailedLoginAttemptLockDuration time.Duration
	SessionCookieSameSite          http.SameSite
	SessionCookieName              string
	SectionCookieMaxAge            time.Duration
	SessionTimeout                 time.Duration
	IsCookieSecure                 bool
}

// SiteInstance handles requests & data for individual sites
type SiteInstance struct {
	Domain    string
	Site      *Site
	DBCache   *DBCache
	Templates map[string]string
	Config    SiteConfig      // Site-specific configuration
	Pages     map[string]Page // routes for this site
	Mu        sync.RWMutex    // mutex for thread-safe route access
}

// Page represents a single page
type Page struct {
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Keywords      []string          `json:"keywords"`
	Author        string            `json:"author"`
	Layout        string            `json:"layout"`
	IsDraft       bool              `json:"is_draft"`
	IsStatic      bool              `json:"is_static"`
	RequireAuth   bool              `json:"require_auth"`
	RequiredRoles []string          `json:"required_roles"`
	TemplatePath  string            `json:"template_path"`
	URL           string            `json:"url"`
	Fetch         string            `json:"fetch"`
	Protected     string            `json:"protected"`
	CustomData    map[string]string `json:"custom_data"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	PublishedAt   *time.Time        `json:"published_at,omitempty"`
	Content       string            `json:"content"`
	Site          Site              `json:"-"`
}
