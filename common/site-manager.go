package common

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SiteInstanceManager manages multiple sites
type SiteInstanceManager struct {
	sites         map[string]*SiteInstance
	apiDispatcher *APIDispatcher
	renderEngine  *RenderEngine
	dbCache       *DBCache
}

// NewSiteInstanceManager creates a new site instance manager
func NewSiteInstanceManager() *SiteInstanceManager {
	return &SiteInstanceManager{
		sites:   make(map[string]*SiteInstance),
		dbCache: NewDBCache(),
	}
}

// SetRenderEngine sets the render engine for the site manager
func (manager *SiteInstanceManager) SetRenderEngine(renderEngine *RenderEngine) {
	manager.renderEngine = renderEngine
}

// LoadSite loads a site configuration from the filesystem
func (manager *SiteInstanceManager) LoadSite(domain string) (*Site, error) {
	// Check if site is already loaded
	if siteInstance, exists := manager.sites[domain]; exists {
		return siteInstance.Site, nil
	}

	sitePath := rootPath(MustGetEnv("SITES_PATH"), domain)

	// Check if site directory exists, create if it doesn't
	if !SecureExists(domain) {
		// Create site directory structure automatically
	}

	site := &Site{
		Domain:   domain,
		Name:     domain, // Default name to domain
		BasePath: sitePath,
		IsActive: true,
		Theme:    "pale-wisp", // Default theme
		// ConfigPath:    filepath.Join(sitePath, "config"),
		// PublicPath:    filepath.Join(sitePath, "public"),
		// AssetsPath:    filepath.Join(sitePath, "assets"),
		// PagesPath:     filepath.Join(sitePath, "pages"),
		// LayoutPath:    filepath.Join(sitePath, "layout"),
		// SectionsPath:  filepath.Join(sitePath, "sections"),
		// TemplatesPath: filepath.Join(sitePath, "templates"),
		// BlocksPath:    filepath.Join(sitePath, "blocks"),
		// SnippetsPath:  filepath.Join(sitePath, "snippets"),
	}

	// Create site instance
	siteInstance := &SiteInstance{
		Domain:    domain,
		Site:      site,
		Manager:   manager,
		Templates: manager.renderEngine.template, // Clone the global templates
		Routes:    make([]RouteEntry, 0),         // Initialize empty routes slice
		Config: SiteConfig{
			MaxFailedLoginAttempts:         5,
			FailedLoginAttemptLockDuration: 60 * time.Minute,
			SessionCookieSameSite:          http.SameSiteLaxMode,
			SessionCookieName:              fmt.Sprintf("%s_session", domain),
			SectionCookieMaxAge:            24 * 7 * time.Hour,
			SessionTimeout:                 24 * 7 * time.Hour,
		},
	}

	// Load pages and create routes for this site instance
	if err := manager.loadSitePages(siteInstance); err != nil {
		panic(fmt.Errorf("failed to load pages for site %s: %w", domain, err))
	}

	// Cache the site instance
	manager.sites[domain] = siteInstance

	return site, nil
}

// GetSite retrieves a site by domain
func (manager *SiteInstanceManager) GetSite(domain string) (*Site, error) {
	return manager.LoadSite(domain)
}

// GetSiteInstance retrieves a site instance by domain
func (manager *SiteInstanceManager) GetSiteInstance(domain string) (*SiteInstance, error) {
	// Check if site instance is already loaded
	if siteInstance, exists := manager.sites[domain]; exists {
		return siteInstance, nil
	}

	// Load the site first (this will create the site instance)
	_, err := manager.LoadSite(domain)
	if err != nil {
		return nil, err
	}

	// Return the cached site instance
	if siteInstance, exists := manager.sites[domain]; exists {
		return siteInstance, nil
	}

	return nil, fmt.Errorf("failed to create site instance for domain: %s", domain)
}

// GetSiteFromHost extracts domain from host header and gets the site
func (manager *SiteInstanceManager) GetSiteFromHost(host string) (*Site, error) {
	// Remove port if present
	domain := strings.Split(host, ":")[0]

	// Remove www. prefix if present
	domain = strings.TrimPrefix(domain, "www.")

	return manager.GetSite(domain)
}

// loadSitePages loads all pages for a site instance and creates routes
func (manager *SiteInstanceManager) loadSitePages(instance *SiteInstance) error {
	sitePath := filepath.Join(MustGetEnv("SITES_PATH"), instance.Domain)
	pagesPath := filepath.Join(sitePath, "pages")

	// Check if pages directory exists
	if !SecureExists(pagesPath) {
		return fmt.Errorf("pages directory not found: %s", pagesPath)
	}

	var routes []RouteEntry

	// Walk through all HTML files in pages directory
	err := filepath.WalkDir(pagesPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-HTML files
		if d.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", path, err)
		}

		// Parse page from HTML content
		page, err := ParsePageHTML(instance, string(content))
		if err != nil {
			return fmt.Errorf("error parsing page %s: %w", path, err)
		}

		// Calculate relative path from pages directory
		relPath, err := filepath.Rel(pagesPath, path)
		if err != nil {
			return fmt.Errorf("error calculating relative path for %s: %w", path, err)
		}

		// Set page path and slug
		page.Path = path
		page.Slug = strings.TrimSuffix(relPath, ".html")

		// Create route entry
		routeEntry := RouteEntry{
			Pattern:  page.Meta.URL,
			PageSlug: page.Slug,
			Page:     page,
			Priority: 0, // Default priority
		}

		routes = append(routes, routeEntry)

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking pages directory: %w", err)
	}

	// Sort routes by priority (lower number = higher priority)
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Priority < routes[j].Priority
	})

	// Update instance routes
	instance.Mu.Lock()
	instance.Routes = routes
	instance.Mu.Unlock()

	return nil
}
