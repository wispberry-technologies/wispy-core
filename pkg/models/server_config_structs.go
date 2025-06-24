package models

// ServerConfig holds the configuration for the HTTP server
type ServerConfig struct {
	Port      int
	Host      string
	Env       string
	SitesPath string
	QuietMode bool
}
