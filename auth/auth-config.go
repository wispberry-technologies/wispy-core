package auth

import (
	"time"
)

// AuthConfig holds configuration for authentication
type AuthConfig struct {
	SecretKey              string
	SessionTimeout         time.Duration
	BCryptCost             int
	MaxFailedLoginAttempts int
	AccountLockoutDuration time.Duration
	SessionCookieName      string
	SessionCookieHTTPOnly  bool
	SessionCookieSecure    bool
	SessionCookieSameSite  string
	OAuth                  map[string]OAuthConfig
}

// OAuthConfig holds OAuth provider configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// NewAuthConfig creates a new auth configuration with default values
func NewAuthConfig() *AuthConfig {
	return &AuthConfig{
		SessionTimeout:         24 * time.Hour,
		BCryptCost:             12,
		MaxFailedLoginAttempts: 5,
		AccountLockoutDuration: 15 * time.Minute,
		SessionCookieName:      "wispy_session",
		SessionCookieHTTPOnly:  true,
		SessionCookieSecure:    true,
		SessionCookieSameSite:  "Strict",
		OAuth:                  make(map[string]OAuthConfig),
	}
}

// SessionTimeout returns the session timeout duration
func (c *AuthConfig) GetSessionTimeout() time.Duration {
	if c.SessionTimeout == 0 {
		return 24 * time.Hour
	}
	return c.SessionTimeout
}

// GetBCryptCost returns the bcrypt cost
func (c *AuthConfig) GetBCryptCost() int {
	if c.BCryptCost == 0 {
		return 12
	}
	return c.BCryptCost
}

// GetMaxFailedLoginAttempts returns the maximum failed login attempts
func (c *AuthConfig) GetMaxFailedLoginAttempts() int {
	if c.MaxFailedLoginAttempts == 0 {
		return 5
	}
	return c.MaxFailedLoginAttempts
}

// GetAccountLockoutDuration returns the account lockout duration
func (c *AuthConfig) GetAccountLockoutDuration() time.Duration {
	if c.AccountLockoutDuration == 0 {
		return 15 * time.Minute
	}
	return c.AccountLockoutDuration
}
