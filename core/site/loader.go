package site

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type siteManager struct {
	mu             sync.RWMutex
	sites          map[string]Site
	tenantsRootDir string
	domains        DomainList
}

// DomainList represents a list of domains associated with sites
type DomainList interface {
	GetDomains() map[string]string // Maps domain to site ID
	AddDomain(domain, siteID string) error
	RemoveDomain(domain string) error
	GetSiteID(domain string) (string, bool)
}

// domainList implements DomainList
type domainList struct {
	mu      sync.RWMutex
	domains map[string]string // Maps domain to site ID
}

func (dl *domainList) GetDomains() map[string]string {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	result := make(map[string]string)
	for domain, siteID := range dl.domains {
		result[domain] = siteID
	}
	return result
}

func (dl *domainList) AddDomain(domain, siteID string) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	dl.domains[domain] = siteID
	return nil
}

func (dl *domainList) RemoveDomain(domain string) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	delete(dl.domains, domain)
	return nil
}

func (dl *domainList) GetSiteID(domain string) (string, bool) {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	siteID, ok := dl.domains[domain]
	return siteID, ok
}

// NewDomainList creates a new domain list
func NewDomainList() DomainList {
	return &domainList{
		domains: make(map[string]string),
	}
}

type SiteManager interface {
	Domains() DomainList
	LoadAllSites() (map[string]Site, error)
	LoadSite(tenantID string) (Site, error)
	GetSite(tenantDomain string) (Site, error)
	UpdateSite(tenantDomain string, site Site) error
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

			site, err := sm.LoadSite(entry.Name())
			if err != nil {
				errs <- fmt.Errorf("error loading site %s: %w", entry.Name(), err)
				return
			}

			mu.Lock()
			sites[entry.Name()] = site

			// Register domain
			if site.GetDomain() != "" {
				sm.domains.AddDomain(site.GetDomain(), site.GetID())
			}
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

// LoadSite loads a site configuration from the tenants directory
func (sm *siteManager) LoadSite(tenantID string) (Site, error) {
	configPath := filepath.Join(sm.tenantsRootDir, tenantID, "config.toml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Temporary struct for TOML parsing
	var config struct {
		Site struct {
			ID         string    `toml:"id"`
			Name       string    `toml:"name"`
			Domain     string    `toml:"domain"`
			BaseURL    string    `toml:"base_url"`
			ContentDir string    `toml:"content_dir"`
			CreatedAt  time.Time `toml:"created_at"`
			UpdatedAt  time.Time `toml:"updated_at"`
		} `toml:"site"`
		Theme struct {
			Name      string            `toml:"name"`
			Base      string            `toml:"base"`
			Variables map[string]string `toml:"variables"`
			Colors    struct {
				Primary          string `toml:"primary"`
				PrimaryContent   string `toml:"primary_content"`
				Secondary        string `toml:"secondary"`
				SecondaryContent string `toml:"secondary_content"`
				Accent           string `toml:"accent"`
				AccentContent    string `toml:"accent_content"`
				Neutral          string `toml:"neutral"`
				NeutralContent   string `toml:"neutral_content"`
				Base100          string `toml:"base100"`
				Base200          string `toml:"base200"`
				Base300          string `toml:"base300"`
				BaseContent      string `toml:"base_content"`
				Info             string `toml:"info"`
				InfoContent      string `toml:"info_content"`
				Success          string `toml:"success"`
				SuccessContent   string `toml:"success_content"`
				Warning          string `toml:"warning"`
				WarningContent   string `toml:"warning_content"`
				Error            string `toml:"error"`
				ErrorContent     string `toml:"error_content"`
			} `toml:"colors"`
			Typography struct {
				FontSans  string `toml:"font_sans"`
				FontMono  string `toml:"font_mono"`
				FontSerif string `toml:"font_serif"`
			} `toml:"typography"`
			Borders struct {
				Width          string `toml:"border_width"`
				RadiusSelector string `toml:"border_radius_selector"`
				RadiusField    string `toml:"border_radius_field"`
			} `toml:"borders"`
			Shadows struct {
				Sm    string `toml:"shadow_sm"`
				Md    string `toml:"shadow_md"`
				Lg    string `toml:"shadow_lg"`
				Xl    string `toml:"shadow_xl"`
				Inner string `toml:"shadow_inner"`
				None  string `toml:"shadow_none"`
			} `toml:"shadows"`
			Animations struct {
				// Duration  string `toml:"animation_duration"`
				// Ease      string `toml:"animation_ease"`
				// EaseIn    string `toml:"animation_ease_in"`
				// EaseOut   string `toml:"animation_ease_out"`
				// EaseInOut string `toml:"animation_ease_in_out"`
			} `toml:"animations"`
		} `toml:"theme"`
	}

	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Create Theme instance
	theme := &Theme{
		Name: config.Theme.Name,
		Base: config.Theme.Base,
		Tokens: ThemeTokens{
			Colors: ColorTokens{
				Primary:          config.Theme.Colors.Primary,
				PrimaryContent:   config.Theme.Colors.PrimaryContent,
				Secondary:        config.Theme.Colors.Secondary,
				SecondaryContent: config.Theme.Colors.SecondaryContent,
				Accent:           config.Theme.Colors.Accent,
				AccentContent:    config.Theme.Colors.AccentContent,
				Neutral:          config.Theme.Colors.Neutral,
				NeutralContent:   config.Theme.Colors.NeutralContent,
				Base100:          config.Theme.Colors.Base100,
				Base200:          config.Theme.Colors.Base200,
				Base300:          config.Theme.Colors.Base300,
				BaseContent:      config.Theme.Colors.BaseContent,
				Info:             config.Theme.Colors.Info,
				InfoContent:      config.Theme.Colors.InfoContent,
				Success:          config.Theme.Colors.Success,
				SuccessContent:   config.Theme.Colors.SuccessContent,
				Warning:          config.Theme.Colors.Warning,
				WarningContent:   config.Theme.Colors.WarningContent,
				Error:            config.Theme.Colors.Error,
				ErrorContent:     config.Theme.Colors.ErrorContent,
			},
			Spacing: SpacingTokens{
				Selector: config.Theme.Variables["spacing_selector"],
				Field:    config.Theme.Variables["spacing_field"],
				Base:     config.Theme.Variables["spacing_base"],
				Sm:       config.Theme.Variables["spacing_sm"],
				Md:       config.Theme.Variables["spacing_md"],
				Lg:       config.Theme.Variables["spacing_lg"],
				Xl:       config.Theme.Variables["spacing_xl"],
			},
			Typography: TypographyTokens{
				FontSans:  config.Theme.Variables["font_sans"],
				FontMono:  config.Theme.Variables["font_mono"],
				FontSerif: config.Theme.Variables["font_serif"],
			},
			Borders: BorderTokens{
				Width:          config.Theme.Variables["border_width"],
				RadiusSelector: config.Theme.Variables["border_radius_selector"],
				RadiusField:    config.Theme.Variables["border_radius_field"],
			},
			Shadows: ShadowTokens{
				Sm:    config.Theme.Variables["shadow_sm"],
				Md:    config.Theme.Variables["shadow_md"],
				Lg:    config.Theme.Variables["shadow_lg"],
				Xl:    config.Theme.Variables["shadow_xl"],
				Inner: config.Theme.Variables["shadow_inner"],
				None:  config.Theme.Variables["shadow_none"],
			},
			Animations: AnimationTokens{},
		},
		Variables: config.Theme.Variables,
	}

	// Create Site instance
	s := &site{
		mu:         sync.RWMutex{},
		ID:         config.Site.ID,
		Name:       config.Site.Name,
		Domain:     config.Site.Domain,
		BaseURL:    config.Site.BaseURL,
		Theme:      theme,
		ContentDir: config.Site.ContentDir,
		Data:       make(map[string]interface{}),
		CreatedAt:  config.Site.CreatedAt,
		UpdatedAt:  config.Site.UpdatedAt,
	}

	// Set defaults
	if s.ContentDir == "" {
		s.ContentDir = "content"
	}
	if s.Theme.Name == "" {
		s.Theme.Name = "default"
	}
	if s.Theme.Base == "" {
		s.Theme.Base = "light"
	}

	return s, nil
}

// GetSite returns a site by domain
func (sm *siteManager) GetSite(tenantDomain string) (Site, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	siteID, found := sm.domains.GetSiteID(tenantDomain)
	if !found {
		return nil, fmt.Errorf("site not found for domain: %s", tenantDomain)
	}

	for _, site := range sm.sites {
		if site.GetID() == siteID {
			return site, nil
		}
	}

	return nil, fmt.Errorf("site not found with ID: %s", siteID)
}

// UpdateSite updates a site by domain
func (sm *siteManager) UpdateSite(tenantDomain string, updatedSite Site) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	siteID, found := sm.domains.GetSiteID(tenantDomain)
	if !found {
		return fmt.Errorf("site not found for domain: %s", tenantDomain)
	}

	// Find the site with the matching ID
	for id, site := range sm.sites {
		if site.GetID() == siteID {
			sm.sites[id] = updatedSite
			// Update domain mapping if domain changed
			if updatedSite.GetDomain() != tenantDomain {
				sm.domains.RemoveDomain(tenantDomain)
				sm.domains.AddDomain(updatedSite.GetDomain(), updatedSite.GetID())
			}
			return nil
		}
	}

	return fmt.Errorf("site not found with ID: %s", siteID)
}
