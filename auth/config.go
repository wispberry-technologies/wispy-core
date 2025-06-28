package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ConfigFile represents the auth configuration file structure
type ConfigFile struct {
	Database struct {
		Type     string `json:"type"`
		Path     string `json:"path"`
		Host     string `json:"host,omitempty"`
		Port     int    `json:"port,omitempty"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
		DBName   string `json:"dbname,omitempty"`
	} `json:"database"`

	Security struct {
		TokenSecret      string `json:"token_secret"`
		TokenExpiration  int    `json:"token_expiration_hours"`
		PasswordMinChars int    `json:"password_min_chars"`
	} `json:"security"`

	OAuth struct {
		Providers map[string]map[string]string `json:"providers"`
	} `json:"oauth"`

	Application struct {
		AllowSignup        bool `json:"allow_signup"`
		RequireVerifyEmail bool `json:"require_verify_email"`
	} `json:"application"`

	Cookie struct {
		Name     string `json:"name"`
		Domain   string `json:"domain"`
		Secure   bool   `json:"secure"`
		HTTPOnly bool   `json:"http_only"`
	} `json:"cookie"`
}

// LoadConfigFromFile loads the configuration from a file
func LoadConfigFromFile(path string) (Config, error) {
	// Default config
	config := DefaultConfig()

	// Read the configuration file
	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Parse the configuration file
	var configFile ConfigFile
	if err := json.Unmarshal(data, &configFile); err != nil {
		return config, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	// Database configuration
	config.DBType = configFile.Database.Type
	switch configFile.Database.Type {
	case "sqlite", "sqlite3":
		// Resolve relative path if needed
		if !filepath.IsAbs(configFile.Database.Path) {
			dir := filepath.Dir(path)
			configFile.Database.Path = filepath.Join(dir, configFile.Database.Path)
		}
		config.DBConn = configFile.Database.Path

	case "postgres":
		config.DBConn = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			configFile.Database.Host,
			configFile.Database.Port,
			configFile.Database.Username,
			configFile.Database.Password,
			configFile.Database.DBName,
		)

		// Add more database types here
	}

	// Security configuration
	if configFile.Security.TokenSecret != "" {
		config.TokenSecret = configFile.Security.TokenSecret
	}

	if configFile.Security.TokenExpiration > 0 {
		config.TokenExpiration = time.Duration(configFile.Security.TokenExpiration) * time.Hour
	}

	if configFile.Security.PasswordMinChars > 0 {
		config.PasswordMinChars = configFile.Security.PasswordMinChars
	}

	// Application configuration
	config.AllowSignup = configFile.Application.AllowSignup
	config.RequireVerifyEmail = configFile.Application.RequireVerifyEmail

	// Cookie configuration
	if configFile.Cookie.Name != "" {
		config.CookieName = configFile.Cookie.Name
	}

	if configFile.Cookie.Domain != "" {
		config.CookieDomain = configFile.Cookie.Domain
	}

	config.CookieSecure = configFile.Cookie.Secure
	config.CookieHTTPOnly = configFile.Cookie.HTTPOnly

	// OAuth configuration
	if len(configFile.OAuth.Providers) > 0 {
		config.OAuthProviders = configFile.OAuth.Providers
	}

	return config, nil
}

// CreateExampleConfig creates an example configuration file
func CreateExampleConfig(path string) error {
	configFile := ConfigFile{}

	// Set up example values
	configFile.Database.Type = "sqlite3"
	configFile.Database.Path = "auth.db"

	configFile.Security.TokenSecret = "change-me-in-production"
	configFile.Security.TokenExpiration = 24
	configFile.Security.PasswordMinChars = 8

	configFile.OAuth.Providers = map[string]map[string]string{
		"google": {
			"client_id":     "your-google-client-id",
			"client_secret": "your-google-client-secret",
			"scopes":        "https://www.googleapis.com/auth/userinfo.email,https://www.googleapis.com/auth/userinfo.profile",
		},
		"discord": {
			"client_id":     "your-discord-client-id",
			"client_secret": "your-discord-client-secret",
			"scopes":        "identify,email",
		},
	}

	configFile.Application.AllowSignup = true
	configFile.Application.RequireVerifyEmail = false

	configFile.Cookie.Name = "auth_token"
	configFile.Cookie.Domain = ""
	configFile.Cookie.Secure = true
	configFile.Cookie.HTTPOnly = true

	// Write the example configuration to file
	data, err := json.MarshalIndent(configFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}
