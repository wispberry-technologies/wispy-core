package models

import (
	"net/http"
	"sync"
	"time"
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
type SiteSchema struct {
	Domain         string             `json:"domain"`
	Name           string             `json:"name"`
	IsActive       bool               `json:"is_active"`
	Theme          string             `json:"theme"`
	Config         SiteConfig         `json:"config"`
	SecurityConfig SiteSecurityConfig `json:"security_config"`
}

type SiteConfig struct {
	CssProcessor   string           `json:"css_processor"` // e.g. "wispy-tail"
	OAuthProviders map[string]OAuth `json:"oauth_providers,omitempty"`
}

type OAuth struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	Enabled      bool   `json:"enabled"`
}

type SiteSecurityConfig struct {
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
	Domain         string
	Name           string
	BasePath       string
	IsActive       bool
	Theme          string
	Config         SiteConfig
	DBCache        *DBCache
	SecurityConfig *SiteSecurityConfig
	Templates      map[string]string
	Pages          map[string]*Page // routes for this site
	Mu             sync.RWMutex     // mutex for thread-safe route access
	DBManager      interface{}      // Will be *db.SiteDatabaseManager at runtime
}

// Page represents a single page
type SiteDetails struct {
	Domain string `json:"domain"`
	Name   string `json:"name"`
}

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
