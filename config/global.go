package config

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"wispy-core/auth"
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
	Host               string `json:"host" toml:"host"`
	HTTPPort           int    `json:"http_port" toml:"http_port"`
	HTTPSPort          int    `json:"https_port" toml:"https_port"`
	RequestsPerSecond  int    `json:"requests_per_second" toml:"requests_per_second"`
	RequestsPerMinute  int    `json:"requests_per_minute" toml:"requests_per_minute"`
	EnableHTTPRedirect bool   `json:"enable_http_redirect" toml:"enable_http_redirect"`
	//
	AuthProvider   auth.AuthProvider `json:"-" toml:"-"`
	AuthMiddleware *auth.Middleware  `json:"-" toml:"-"`
}

var globalConf GlobalConfig // GlobalConfig is a singleton instance of GlobalConfig

func GetGlobalConfig() GlobalConfig {
	if globalConf == nil {
		common.Error("Global configuration is not initialized. Call LoadGlobalConfig() first.")
		panic("Global configuration is not initialized. Call LoadGlobalConfig() first.")
	}
	return globalConf
}

func InitGlobalConf(httpPort, httpsPort int, host, env, sitesPath, staticPath, projectRoot, cacheDir string) GlobalConfig {
	enableHTTPRedirect := true
	if env == "development" || env == "local" {
		// Default to false for local development
		// This prevents ACME challenges from redirecting to HTTPS
		// and handles it via the HTTP server redirecting to HTTPS for local development environments
		enableHTTPRedirect = common.GetEnvBool("ENABLE_HTTP_REDIRECT", false)
	}

	// Auth provider initialization
	// This is the core authentication provider used for CMS/tenant admin user authentication
	authProvider, _, _, _, err := auth.InitSQLiteAuth("./_data/system/local_dbs/wispy_auth.db", auth.DefaultConfig()) // Initialize SQLite auth provider
	if err != nil {
		common.Error("Failed to initialize core authentication provider")
		panic("Failed to initialize core authentication provider")
	}

	// Configure auth for CMS usage
	authConfig := auth.DefaultConfig()
	authConfig.LoginURL = "/wispy-cms/login" // Set login URL for CMS
	authMiddleware := auth.NewMiddleware(authProvider, authConfig)

	// Ensure default admin user exists
	if err := ensureDefaultAdminUser(authProvider); err != nil {
		common.Error("Failed to create default admin user: %v", err)
		// Don't panic here, just log the error - the system can still work
	}

	serverConfig := serverConfig{
		HTTPSPort:          httpsPort,
		HTTPPort:           httpPort,
		Host:               host,
		RequestsPerSecond:  12,
		RequestsPerMinute:  240,
		EnableHTTPRedirect: enableHTTPRedirect,
		AuthProvider:       authProvider,
		AuthMiddleware:     authMiddleware,
	}

	globalConf = &globalConfig{
		Env:         env,
		ProjectRoot: projectRoot,
		SitesPath:   sitesPath,
		StaticPath:  staticPath,
		CacheDir:    cacheDir,
		Server:      serverConfig,
	}
	return globalConf
}

type GlobalConfig interface {
	// Platform-related methods
	GetEnv() string
	GetProjectRoot() string
	GetSitesPath() string
	GetStaticPath() string
	GetCacheDir() string
	//
	// Server-related methods
	// These methods provide access to server configuration
	GetHttpPort() int
	GetHttpsPort() int
	GetHost() string
	GetRequestsPerSecond() int
	GetRequestsPerMinute() int
	GetEnableHTTPRedirect() bool
	//
	// Get the core authentication provider
	// Used for CMS/tenant admin user authentication
	// As well as for API authentication
	GetCoreAuth() auth.AuthProvider
	GetCoreAuthMiddleware() *auth.Middleware
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

func (c *globalConfig) GetEnableHTTPRedirect() bool {
	return c.Server.EnableHTTPRedirect
}

func (c *globalConfig) GetCoreAuth() auth.AuthProvider {
	return c.Server.AuthProvider
}

func (c *globalConfig) GetCoreAuthMiddleware() *auth.Middleware {
	return c.Server.AuthMiddleware
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
	} else {
		// Otherwise, use the current directory
		projectRoot = currentDir
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

// ensureDefaultAdminUser creates a default admin user if one doesn't exist
func ensureDefaultAdminUser(authProvider auth.AuthProvider) error {
	// Get default admin credentials from environment variables
	defaultEmail := common.MustGetEnv("WISPY_ADMIN_EMAIL")
	defaultUsername := common.MustGetEnv("WISPY_ADMIN_USERNAME")
	defaultPassword := common.MustGetEnv("WISPY_ADMIN_PASSWORD")

	ctx := context.Background()

	// Check if admin user already exists by trying to login
	_, err := authProvider.Login(ctx, defaultEmail, defaultPassword)
	if err == nil {
		// Admin user already exists and credentials work
		common.Info("Default admin user already exists and is accessible with email: %s", defaultEmail)
		return nil
	}

	// Try to register the default admin user
	common.Info("Creating default admin user...")
	common.Info("  Email: %s", defaultEmail)
	common.Info("  Username: %s", defaultUsername)
	if defaultPassword == "1amJustALittleDefaultPa$$w0rd!" {
		common.Warning("  Using default password '****J**********Defa*****rd****' - CHANGE THIS IN PRODUCTION!")
	}

	user, err := authProvider.Register(ctx, defaultEmail, defaultUsername, defaultPassword)
	if err != nil {
		// User might already exist but with different password
		common.Warning("Failed to register default admin user: %v", err)
		common.Warning("This might mean a user with email '%s' already exists but has a different password", defaultEmail)
		common.Info("You can set custom admin credentials using environment variables:")
		common.Info("  WISPY_ADMIN_EMAIL=%s", defaultEmail)
		common.Info("  WISPY_ADMIN_USERNAME=%s", defaultUsername)
		common.Info("  WISPY_ADMIN_PASSWORD=*******************")
		return err
	}

	common.Info("✅ Default admin user created successfully!")
	common.Info("   User ID: %s", user.ID)
	common.Info("   Email: %s", user.Email)
	common.Info("   Username: %s", user.Username)
	common.Info("🔐 Login at: /wispy-cms/login")

	if defaultPassword == "1amJustALittleDefaultPa$$w0rd!" {
		common.Warning("⚠️  SECURITY WARNING: Please change the default password after first login!")
	}

	return nil
}
