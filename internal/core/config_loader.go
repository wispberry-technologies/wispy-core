package core

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

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
	// Parse [auth] section for auth configuration
	if authTree := tree.Get("auth"); authTree != nil {
		if m, ok := authTree.(*toml.Tree); ok {
			// Initialize AuthConfig if not already set
			if siteInstance.AuthConfig == nil {
				siteInstance.AuthConfig = &models.SiteAuthConfig{
					// Security settings
					MaxFailedLoginAttempts:         5,
					FailedLoginAttemptLockDuration: 30 * time.Minute,
					// Registration settings
					RegistrationEnabled:  true,
					RequiredFields:       []string{"email", "password", "first_name", "last_name"},
					DefaultRoles:         []string{},
					AllowedPasswordReset: true,
				}
			}

			// Security settings
			if v := m.Get("max_failed_login_attempts"); v != nil {
				if attempts, ok := v.(int64); ok {
					siteInstance.AuthConfig.MaxFailedLoginAttempts = int(attempts)
				}
			}

			if v := m.Get("failed_login_attempt_lock_duration"); v != nil {
				if duration, ok := v.(string); ok {
					if d, err := time.ParseDuration(duration); err == nil {
						siteInstance.AuthConfig.FailedLoginAttemptLockDuration = d
					}
				}
			}

			// Registration enabled
			if v := m.Get("registration_enabled"); v != nil {
				if enabled, ok := v.(bool); ok {
					siteInstance.AuthConfig.RegistrationEnabled = enabled
				}
			}

			// Required fields
			if v := m.Get("required_fields"); v != nil {
				if arr, ok := v.([]interface{}); ok {
					fields := make([]string, len(arr))
					for i, item := range arr {
						fields[i] = fmt.Sprint(item)
					}
					siteInstance.AuthConfig.RequiredFields = fields
				}
			}

			// Default roles
			if v := m.Get("default_roles"); v != nil {
				if arr, ok := v.([]interface{}); ok {
					roles := make([]string, len(arr))
					for i, item := range arr {
						roles[i] = fmt.Sprint(item)
					}
					siteInstance.AuthConfig.DefaultRoles = roles
				}
			}

			// Password reset
			if v := m.Get("allowed_password_reset"); v != nil {
				if allowed, ok := v.(bool); ok {
					siteInstance.AuthConfig.AllowedPasswordReset = allowed
				}
			}
		}
	} else {
		// Set default auth config if not specified
		siteInstance.AuthConfig = &models.SiteAuthConfig{
			// Security settings
			MaxFailedLoginAttempts:         5,
			FailedLoginAttemptLockDuration: 30 * time.Minute,
			// Registration settings
			RegistrationEnabled:  true,
			RequiredFields:       []string{"email", "password", "first_name", "last_name"},
			DefaultRoles:         []string{},
			AllowedPasswordReset: true,
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
