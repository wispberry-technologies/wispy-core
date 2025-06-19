package core

import (
	"fmt"
	"os"
	"path/filepath"

	"wispy-core/pkg/models"

	"github.com/pelletier/go-toml"
)

// LoadSiteConfig loads the config.toml for a site and returns a SiteConfig
func LoadSiteConfig(siteBasePath string, siteInstance *models.SiteInstance) error {
	configPath := filepath.Join(siteBasePath, "config", "config.toml")

	if _, err := os.Stat(configPath); err != nil {
		return nil // config is optional, use defaults
	}
	tree, err := toml.LoadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to parse config.toml: %w", err)
	}

	// Parse [site] section for fields
	if siteTree := tree.Get("site"); siteTree != nil {
		if m, ok := siteTree.(*toml.Tree); ok {
			if v := m.Get("name"); v != nil {
				siteInstance.Name = fmt.Sprint(v)
			}
			if v := m.Get("theme"); v != nil {
				siteInstance.Theme = fmt.Sprint(v)
			}
			if v := m.Get("css_processor"); v != nil {
				siteInstance.CssProcessor = fmt.Sprint(v)
			}
			if v := m.Get("oauth_providers"); v != nil {
				if arr, ok := v.([]interface{}); ok {
					providers := make([]string, len(arr))
					for i, item := range arr {
						providers[i] = fmt.Sprint(item)
					}
					siteInstance.OAuthProviders = providers
				}
			}
		}
	}

	// Parse [route_proxies] section
	if proxies := tree.Get("route_proxies"); proxies != nil {
		if m, ok := proxies.(*toml.Tree); ok {
			for _, k := range m.Keys() {
				if v, ok := m.Get(k).(string); ok {
					siteInstance.RouteProxies[k] = v
				}
			}
		}
	}

	return nil
}
