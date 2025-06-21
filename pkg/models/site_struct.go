package models

import (
	"sync"
	"time"
	"wispy-core/internal/cache"

	"github.com/go-chi/chi/v5"
)

type HtmlDocumentTags struct {
	TagType     string
	TagName     string
	Contents    string // Inner HTML content
	Attributes  map[string]string
	SelfClosing bool   // If true, tag is self-closing (e.g. <img />, <br />)
	Location    string // e.g. "head", "pre-footer"
	Priority    int    // Lower numbers are higher priority
}
type HtmlMetaTag struct {
	Name       string
	Content    string
	Property   string
	HttpEquiv  string
	Charset    string
	CustomAttr map[string]string
}
type ConstructHTMLDocument struct {
	Body         string
	Lang         string
	Title        string
	MetaTags     []HtmlMetaTag
	DocumentTags []HtmlDocumentTags
}

// Site represents a single site in the multisite system
type OAuth struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	Enabled      bool   `json:"enabled"`
}

// SiteAuthConfig holds all authentication and security configuration for a site
type SiteAuthConfig struct {
	// Security settings
	MaxFailedLoginAttempts         int           `json:"max_failed_login_attempts" toml:"max_failed_login_attempts"`
	FailedLoginAttemptLockDuration time.Duration `json:"failed_login_attempt_lock_duration" toml:"failed_login_attempt_lock_duration"`
	// Session settings (commented out until implemented)
	// SessionCookieSameSite          http.SameSite  `json:"session_cookie_same_site" toml:"session_cookie_same_site"`
	// SessionCookieName              string         `json:"session_cookie_name" toml:"session_cookie_name"`
	// SectionCookieMaxAge            time.Duration  `json:"section_cookie_max_age" toml:"section_cookie_max_age"`
	// SessionTimeout                 time.Duration  `json:"session_timeout" toml:"session_timeout"`
	// IsCookieSecure                 bool           `json:"is_cookie_secure" toml:"is_cookie_secure"`

	// Registration settings
	RegistrationEnabled  bool     `json:"registration_enabled" toml:"RegistrationEnabled"`
	RequiredFields       []string `json:"required_fields" toml:"required_fields"`
	EnabledFields        []string `json:"enabled_fields" toml:"enabled_fields"`
	DefaultRoles         []string `json:"default_roles" toml:"default_roles"`
	AllowedPasswordReset bool     `json:"allowed_password_reset" toml:"allowed_password_reset"`
}

// SiteInstance handles requests & data for individual sites
type SiteInstance struct {
	Domain   string
	Name     string
	BasePath string
	IsActive bool
	Theme    string
	//
	CssProcessor   string   `json:"css_processor"`
	OAuthProviders []string `json:"oauth_providers,omitempty"`
	// RouteProxies maps route prefixes to proxy targets (e.g. "/api": "http://localhost:3000")
	RouteProxies map[string]string `json:"route_proxies" toml:"route_proxies"`
	//
	DBCache    *cache.DBCache
	Router     *chi.Mux
	AuthConfig *SiteAuthConfig
	Templates  map[string]string
	Pages      map[string]*Page // routes for this site
	Mu         sync.RWMutex     // mutex for thread-safe route access
}
type SiteSchema struct {
	Domain     string         `json:"domain"`
	Name       string         `json:"name"`
	IsActive   bool           `json:"is_active"`
	Theme      string         `json:"theme"`
	AuthConfig SiteAuthConfig `json:"auth_config"`
}

// Page represents a single page

type Page struct {
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Lang          string            `json:"lang"`
	Slug          string            `json:"slug"`
	Keywords      []string          `json:"keywords"`
	Author        string            `json:"author"`
	LayoutName    string            `json:"layout"`
	IsDraft       bool              `json:"is_draft"`
	IsStatic      bool              `json:"is_static"`
	RequireAuth   bool              `json:"require_auth"`
	RequiredRoles []string          `json:"required_roles"`
	FilePath      string            `json:"file_path"`
	Protected     string            `json:"protected"`
	PageData      map[string]string `json:"custom_data"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	PublishedAt   *time.Time        `json:"published_at,omitempty"`
	MetaTags      []HtmlMetaTag     `json:"meta_tags"`
	// SiteDetails contains information about the site this page belongs to
	SiteDetails SiteDetails `json:"site_details"`
}

// PageContent represents language-specific and content data for a page
// content_json is stored as a string (JSON) in SQLite
// meta_tags and keywords are comma-separated or JSON
// lang is required
type PageContent struct {
	ID        string            `json:"id"`
	PageID    string            `json:"page_id"`
	Lang      string            `json:"lang"`
	Keywords  []string          `json:"keywords"`
	MetaTags  []HtmlMetaTag     `json:"meta_tags"`
	Content   map[string]string `json:"content_json"` // JSON fields for template population
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type SiteDetails struct {
	Domain string `json:"domain"`
	Name   string `json:"name"`
}

// Blog represents a blog post
type Blog struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Author      string     `json:"author"`
	Body        string     `json:"body"`
	Tags        []string   `json:"tags"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

// Media represents a media asset
type Media struct {
	ID         string    `json:"id"`
	FileName   string    `json:"file_name"`
	FileType   string    `json:"file_type"`
	URL        string    `json:"url"`
	Size       int64     `json:"size"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// Statistic represents a statistics record
type Statistic struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	CreatedAt time.Time `json:"created_at"`
}
