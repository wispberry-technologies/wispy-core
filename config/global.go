package config

import (
	"os"
	"path/filepath"
	"strconv"
	"wispy-core/common"
)

type globalConfig struct {
	Env         string `json:"env" toml:"env"`
	ProjectRoot string `json:"project_root" toml:"project_root"`
	SitesPath   string `json:"sites_path" toml:"sites_path"`
	StaticPath  string `json:"static_path" toml:"static_path"`
	CacheDir    string `json:"cache_dir" toml:"cache_dir"`
	Server      serverConfig
}

type serverConfig struct {
	Host              string `json:"host" toml:"host"`
	HTTPPort          int    `json:"http_port" toml:"http_port"`
	HTTPSPort         int    `json:"https_port" toml:"https_port"`
	RequestsPerSecond int    `json:"requests_per_second" toml:"requests_per_second"`
	RequestsPerMinute int    `json:"requests_per_minute" toml:"requests_per_minute"`
}

var GlobalConf GlobalConfig // GlobalConfig is a singleton instance of GlobalConfig

func InitGlobalConf(httpPort, httpsPort int, host, env, sitesPath, staticPath, projectRoot, cacheDir string) GlobalConfig {
	GlobalConf = &globalConfig{
		Env:         env,
		ProjectRoot: projectRoot,
		SitesPath:   sitesPath,
		StaticPath:  staticPath,
		CacheDir:    cacheDir,
		Server: serverConfig{
			HTTPSPort:         httpsPort,
			HTTPPort:          httpPort,
			Host:              host,
			RequestsPerSecond: 12,  // Default value, can be overridden
			RequestsPerMinute: 240, // Default value, can be overridden
		}}
	return GlobalConf
}

type GlobalConfig interface {
	GetEnv() string
	GetProjectRoot() string
	GetSitesPath() string
	GetStaticPath() string
	GetCacheDir() string
	// Server-related methods
	GetHttpPort() int
	GetHttpsPort() int
	GetHost() string
	GetRequestsPerSecond() int
	GetRequestsPerMinute() int
}

func (c *globalConfig) GetHttpPort() int {
	return c.Server.HTTPPort
}

func (c *globalConfig) GetHttpsPort() int {
	return c.Server.HTTPSPort
}

func (c *globalConfig) GetHost() string {
	return c.Server.Host
}

func (c *globalConfig) GetEnv() string {
	return c.Env
}

func (c *globalConfig) GetSitesPath() string {
	return c.SitesPath
}

func (c *globalConfig) GetStaticPath() string {
	return c.StaticPath
}

func (c *globalConfig) GetProjectRoot() string {
	return c.ProjectRoot
}

func (c *globalConfig) GetCacheDir() string {
	return c.CacheDir
}

func (c *globalConfig) GetRequestsPerSecond() int {
	return c.Server.RequestsPerSecond
}

func (c *globalConfig) GetRequestsPerMinute() int {
	return c.Server.RequestsPerMinute
}

// ----------
// LoadGlobalConfig initializes the global configuration by reading environment variables
// ----------
func LoadGlobalConfig() GlobalConfig {
	// Load environment variables and set defaults
	common.LoadDotEnv() // Load .env file if it exists

	// Get current working directory
	var projectRoot string
	currentDir, dirErr := os.Getwd()
	if dirErr != nil {
		panic("Failed to get current working directory: " + dirErr.Error())
	}
	// If running from /server, go up two levels
	if filepath.Base(currentDir) == "server" {
		projectRoot = filepath.Dir(currentDir)
	}
	// Set WISPY_CORE_ROOT
	common.GetEnvOrSet("WISPY_CORE_ROOT", projectRoot) // Default to current directory if not set
	// Get configuration from environment
	var httpsPort, httpPort, host, env string
	if common.IsProduction() {
		env = "production"
		httpPort = common.GetEnvOrSet("HTTP_PORT", "80")
		httpsPort = common.GetEnvOrSet("HTTPS_PORT", "443")
		host = common.GetEnvOrSet("HOST", "0.0.0.0")
	} else if common.IsStaging() {
		env = "staging"
		httpPort = common.GetEnvOrSet("HTTP_PORT", "80")
		httpsPort = common.GetEnvOrSet("HTTPS_PORT", "443")
		host = common.GetEnvOrSet("HOST", "0.0.0.0")
	} else { // OR common.IsDevelopment()
		// Default to development settings
		env = "development"
		httpPort = common.GetEnvOrSet("HTTP_PORT", "8079")
		httpsPort = common.GetEnvOrSet("HTTPS_PORT", "8080")
		host = common.GetEnvOrSet("HOST", "localhost") // Default host for development
	}
	sitesPath := common.GetEnvOrSet("SITES_PATH", filepath.Join(projectRoot, "_data/tenants"))  // Default sites path
	staticPath := common.GetEnvOrSet("STATIC_PATH", filepath.Join(projectRoot, "_data/static")) // Optional, but recommended for static assets
	cacheDir := common.GetEnvOrSet("CACHE_DIR", filepath.Join(projectRoot, ".wispy/cache"))     // Default cache directory

	httpPortInt, err := strconv.Atoi(httpPort)
	if err != nil {
		common.Error("Invalid HTTP_PORT value: %v", err)
		panic("Invalid HTTP_PORT value: " + err.Error())
	}

	httpsPortInt, err := strconv.Atoi(httpsPort)
	if err != nil {
		common.Error("Invalid PORT value: %v", err)
		panic("Invalid PORT value: " + err.Error())
	}

	return InitGlobalConf(
		httpPortInt,
		httpsPortInt,
		host,
		env,
		sitesPath,
		staticPath,
		projectRoot,
		cacheDir,
	)
}
