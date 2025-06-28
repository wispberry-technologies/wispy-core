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
	Port              int    `json:"port" toml:"port"`
	RequestsPerSecond int    `json:"requests_per_second" toml:"requests_per_second"`
	RequestsPerMinute int    `json:"requests_per_minute" toml:"requests_per_minute"`
}

var GlobalConf GlobalConfig // GlobalConfig is a singleton instance of GlobalConfig

func InitGlobalConf(port int, host, env, sitesPath, staticPath, projectRoot, cacheDir string) GlobalConfig {
	GlobalConf = &globalConfig{
		Env:         env,
		ProjectRoot: projectRoot,
		SitesPath:   sitesPath,
		StaticPath:  staticPath,
		CacheDir:    cacheDir,
		Server: serverConfig{
			Port:              port,
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
	GetPort() int
	GetHost() string
	GetRequestsPerSecond() int
	GetRequestsPerMinute() int
}

func (c *globalConfig) GetPort() int {
	return c.Server.Port
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
	var port, host string
	if common.IsProduction() {
		port = common.GetEnvOrSet("PORT", "443")
		host = common.GetEnvOrSet("HOST", "0.0.0.0")
	} else {
		port = common.GetEnvOrSet("PORT", "8080")      // Default port for development
		host = common.GetEnvOrSet("HOST", "localhost") // Default host for development
	}
	env := common.GetEnvOrSet("ENV", "development")
	sitesPath := common.GetEnvOrSet("SITES_PATH", filepath.Join(projectRoot, "_data/sites"))    // Default sites path
	staticPath := common.GetEnvOrSet("STATIC_PATH", filepath.Join(projectRoot, "_data/static")) // Optional, but recommended for static assets
	cacheDir := common.GetEnvOrSet("CACHE_DIR", filepath.Join(projectRoot, ".wispy-cache"))     // Default cache directory

	portInt, err := strconv.Atoi(port)
	if err != nil {
		common.Error("Invalid PORT value: %v", err)
		panic("Invalid PORT value: " + err.Error())
	}

	return InitGlobalConf(
		portInt,
		host,
		env,
		sitesPath,
		staticPath,
		projectRoot,
		cacheDir,
	)
}
