package site

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"wispy-core/auth"
	"wispy-core/common"

	"github.com/pelletier/go-toml/v2"
)

type siteManager struct {
	mu             sync.RWMutex
	sites          map[string]Site // Maps domain to Site
	tenantsRootDir string
	domains        DomainList
}

// DomainList represents a list of domains associated with sites
type DomainList interface {
	GetDomains() []string // Returns list of all domains
	AddDomain(domain string) error
	RemoveDomain(domain string) error
	HasDomain(domain string) bool
}

// domainList implements DomainList
type domainList struct {
	mu      sync.RWMutex
	domains map[string]bool // Set of domains
}

func (dl *domainList) GetDomains() []string {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	result := make([]string, 0, len(dl.domains))
	for domain := range dl.domains {
		result = append(result, domain)
	}
	return result
}

func (dl *domainList) AddDomain(domain string) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	dl.domains[domain] = true
	return nil
}

func (dl *domainList) RemoveDomain(domain string) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	delete(dl.domains, domain)
	return nil
}

func (dl *domainList) HasDomain(domain string) bool {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	return dl.domains[domain]
}

// NewDomainList creates a new domain list
func NewDomainList() DomainList {
	return &domainList{
		domains: make(map[string]bool),
	}
}

type SiteManager interface {
	Domains() DomainList
	LoadAllSites() (map[string]Site, error)
	LoadSiteByDomain(domain string) (Site, error)
	GetSite(domain string) (Site, error)
	UpdateSite(domain string, site Site) error
}

// NewSiteManager creates a new site manager
func NewSiteManager(tenantsRootDir string) SiteManager {
	return &siteManager{
		sites:          make(map[string]Site),
		tenantsRootDir: tenantsRootDir,
		domains:        NewDomainList(),
	}
}

// Domains returns the domain list managed by this site manager
func (sm *siteManager) Domains() DomainList {
	return sm.domains
}

// LoadAllSites loads all tenant sites from the directory
func (sm *siteManager) LoadAllSites() (map[string]Site, error) {
	entries, err := os.ReadDir(sm.tenantsRootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read tenants directory: %w", err)
	}

	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		sites = make(map[string]Site)
		errs  = make(chan error, len(entries))
	)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		wg.Add(1)
		go func(entry os.DirEntry) {
			defer wg.Done()

			site, err := sm.LoadSiteByDomain(entry.Name())
			if err != nil {
				errs <- fmt.Errorf("error loading site %s: %w", entry.Name(), err)
				return
			}

			mu.Lock()
			domain := site.GetDomain()
			if domain == "" {
				domain = entry.Name() // Use directory name as fallback
			}
			sites[domain] = site

			// Store in the site manager's sites map
			sm.mu.Lock()
			sm.sites[domain] = site
			sm.mu.Unlock()

			// Register domain
			sm.domains.AddDomain(domain)
			mu.Unlock()
		}(entry)
	}

	wg.Wait()
	close(errs)

	// Return first error if any occurred
	for err := range errs {
		return sites, err
	}

	return sites, nil
}

// LoadSiteByDomain loads a site configuration by domain name
func (sm *siteManager) LoadSiteByDomain(domain string) (Site, error) {
	// Normalize domain for config file lookup (removes port and .localhost)
	normalizedDomain := common.NormalizeHost(domain)
	configPath := filepath.Join(sm.tenantsRootDir, normalizedDomain, "config.toml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration: %w", err)
	}

	// Parse the entire configuration as a map
	var fullConfig map[string]interface{}
	if err := toml.Unmarshal(data, &fullConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Also parse the site-specific configuration for structured access
	var siteConfig struct {
		Site struct {
			Name      string    `toml:"name"`
			Domain    string    `toml:"domain"`
			BaseURL   string    `toml:"base_url"`
			CreatedAt time.Time `toml:"created_at"`
			UpdatedAt time.Time `toml:"updated_at"`
		} `toml:"site"`
	}

	if err := toml.Unmarshal(data, &siteConfig); err != nil {
		return nil, fmt.Errorf("failed to parse site config: %w", err)
	}

	// Use domain from config or fallback to directory name
	siteDomain := siteConfig.Site.Domain
	if siteDomain == "" {
		siteDomain = domain
	}

	// Create Site instance
	s := &site{
		mu:        sync.RWMutex{},
		Name:      siteConfig.Site.Name,
		Domain:    siteDomain,
		BaseURL:   siteConfig.Site.BaseURL,
		CssThemes: make(map[string]string),
		Data:      make(map[string]interface{}),
		Config:    fullConfig, // Store the full configuration
		CreatedAt: siteConfig.Site.CreatedAt,
		UpdatedAt: siteConfig.Site.UpdatedAt,
	}

	// Setup Database manager
	s.DbManager = NewDatabaseManager(s.Domain)

	// Setup authentication manager
	// Auth
	// Setup authentication and authorization
	authConfig := auth.DefaultConfig()
	authProvider, aProviderErr := auth.NewDefaultAuthProvider(authConfig)
	if aProviderErr != nil {
		common.Fatal("Failed to create auth provider: %v", aProviderErr)
	}
	s.AuthManager = authProvider
	// authMiddleware := auth.NewMiddleware(authProvider, authConfig)

	return s, nil
}

// GetSite returns a site by domain
func (sm *siteManager) GetSite(domain string) (Site, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// First try exact match
	site, found := sm.sites[domain]
	if found {
		return site, nil
	}

	// Use common.NormalizeHost to handle domain normalization
	normalizedDomain := common.NormalizeHost(domain)

	site, found = sm.sites[normalizedDomain]
	if found {
		return site, nil
	}

	return nil, fmt.Errorf("site not found for domain: %s (tried: %s)", domain, normalizedDomain)
}

// UpdateSite updates a site by domain
func (sm *siteManager) UpdateSite(domain string, updatedSite Site) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	_, found := sm.sites[domain]
	if !found {
		return fmt.Errorf("site not found for domain: %s", domain)
	}

	sm.sites[domain] = updatedSite

	// Update domain registration if domain changed
	newDomain := updatedSite.GetDomain()
	if newDomain != domain {
		delete(sm.sites, domain)
		sm.sites[newDomain] = updatedSite
		sm.domains.RemoveDomain(domain)
		sm.domains.AddDomain(newDomain)
	}

	return nil
}
