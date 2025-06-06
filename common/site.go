package common

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// Site represents a single site in the multisite system
type Site struct {
	Domain        string `json:"domain"`
	Name          string `json:"name"`
	BasePath      string `json:"base_path"`
	IsActive      bool   `json:"is_active"`
	Theme         string `json:"theme"`
	ConfigPath    string `json:"config_path"`
	PublicPath    string `json:"public_path"`
	AssetsPath    string `json:"assets_path"`
	PagesPath     string `json:"pages_path"`
	LayoutPath    string `json:"layout_path"`
	SectionsPath  string `json:"sections_path"`
	TemplatesPath string `json:"templates_path"`
	BlocksPath    string `json:"blocks_path"`
	SnippetsPath  string `json:"snippets_path"`
}

// SiteManager manages multiple sites
type SiteManager struct {
	sites         map[string]*Site
	baseSitesPath string
}

// NewSiteManager creates a new site manager
func NewSiteManager(baseSitesPath string) *SiteManager {
	return &SiteManager{
		sites:         make(map[string]*Site),
		baseSitesPath: baseSitesPath,
	}
}

// LoadSite loads a site configuration from the filesystem
func (sm *SiteManager) LoadSite(domain string) (*Site, error) {
	// Check if site is already loaded
	if site, exists := sm.sites[domain]; exists {
		return site, nil
	}

	sitePath := filepath.Join(sm.baseSitesPath, domain)

	// Check if site directory exists
	if _, err := os.Stat(sitePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("site directory does not exist: %s", domain)
	}

	site := &Site{
		Domain:        domain,
		Name:          domain, // Default name to domain
		BasePath:      sitePath,
		IsActive:      true,
		Theme:         "pale-wisp", // Default theme
		ConfigPath:    filepath.Join(sitePath, "config"),
		PublicPath:    filepath.Join(sitePath, "public"),
		AssetsPath:    filepath.Join(sitePath, "assets"),
		PagesPath:     filepath.Join(sitePath, "pages"),
		LayoutPath:    filepath.Join(sitePath, "layout"),
		SectionsPath:  filepath.Join(sitePath, "sections"),
		TemplatesPath: filepath.Join(sitePath, "templates"),
		BlocksPath:    filepath.Join(sitePath, "blocks"),
		SnippetsPath:  filepath.Join(sitePath, "snippets"),
	}

	// Cache the site
	sm.sites[domain] = site

	return site, nil
}

// GetSite retrieves a site by domain
func (sm *SiteManager) GetSite(domain string) (*Site, error) {
	return sm.LoadSite(domain)
}

// GetSiteFromHost extracts domain from host header and gets the site
func (sm *SiteManager) GetSiteFromHost(host string) (*Site, error) {
	// Remove port if present
	domain := strings.Split(host, ":")[0]

	// Remove www. prefix if present
	if strings.HasPrefix(domain, "www.") {
		domain = strings.TrimPrefix(domain, "www.")
	}

	return sm.GetSite(domain)
}

// CreateSiteDirectories creates the necessary directory structure for a new site
func (sm *SiteManager) CreateSiteDirectories(domain string) error {
	site, err := sm.LoadSite(domain)
	if err != nil {
		return err
	}

	directories := []string{
		site.ConfigPath,
		site.PublicPath,
		site.AssetsPath,
		site.PagesPath,
		site.LayoutPath,
		site.SectionsPath,
		site.TemplatesPath,
		site.BlocksPath,
		site.SnippetsPath,
		filepath.Join(site.BasePath, "dbs"),
		filepath.Join(site.BasePath, "migrations"),
		filepath.Join(site.ConfigPath, "themes"),
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetTemplate loads and parses a template for the site
func (s *Site) GetTemplate(templateName string) (*template.Template, error) {
	templatePath := filepath.Join(s.TemplatesPath, templateName+".html")

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template not found: %s", templateName)
	}

	// Parse template with layout
	tmpl := template.New(templateName + ".html")

	// Load layout templates
	layoutGlob := filepath.Join(s.LayoutPath, "*.html")
	if layouts, err := filepath.Glob(layoutGlob); err == nil && len(layouts) > 0 {
		tmpl, err = tmpl.ParseGlob(layoutGlob)
		if err != nil {
			return nil, fmt.Errorf("error parsing layout templates: %w", err)
		}
	}

	// Load snippets
	snippetsGlob := filepath.Join(s.SnippetsPath, "*.html")
	if snippets, err := filepath.Glob(snippetsGlob); err == nil && len(snippets) > 0 {
		tmpl, err = tmpl.ParseGlob(snippetsGlob)
		if err != nil {
			return nil, fmt.Errorf("error parsing snippet templates: %w", err)
		}
	}

	// Load sections
	sectionsGlob := filepath.Join(s.SectionsPath, "*.html")
	if sections, err := filepath.Glob(sectionsGlob); err == nil && len(sections) > 0 {
		tmpl, err = tmpl.ParseGlob(sectionsGlob)
		if err != nil {
			return nil, fmt.Errorf("error parsing section templates: %w", err)
		}
	}

	// Load blocks
	blocksGlob := filepath.Join(s.BlocksPath, "*.html")
	if blocks, err := filepath.Glob(blocksGlob); err == nil && len(blocks) > 0 {
		tmpl, err = tmpl.ParseGlob(blocksGlob)
		if err != nil {
			return nil, fmt.Errorf("error parsing block templates: %w", err)
		}
	}

	// Parse the main template
	tmpl, err := tmpl.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("error parsing template %s: %w", templateName, err)
	}

	return tmpl, nil
}
