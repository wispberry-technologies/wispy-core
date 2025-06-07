package common

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wispberry-technologies/wispy-core/auth"
	"github.com/wispberry-technologies/wispy-core/models"
)

// Site represents a single site in the multisite system
type Site struct {
	Domain        string `json:"domain"`
	Name          string `json:"name"`
	BasePath      string `json:"base_path"`
	IsActive      bool   `json:"is_active"`
	Theme         string `json:"theme"`
	ConfigPath    string `json:"config_path"`
	PublicPath    string `json:"public_path"`
	AssetsPath    string `json:"assets_path"`
	PagesPath     string `json:"pages_path"`
	LayoutPath    string `json:"layout_path"`
	SectionsPath  string `json:"sections_path"`
	TemplatesPath string `json:"templates_path"`
	BlocksPath    string `json:"blocks_path"`
	SnippetsPath  string `json:"snippets_path"`
}

// SiteManager manages multiple sites
type SiteManager struct {
	sites   map[string]*Site
	dbCache *DBCache
}

// NewSiteManager creates a new site manager
func NewSiteManager() *SiteManager {
	return &SiteManager{
		sites:   make(map[string]*Site),
		dbCache: NewDBCache(),
	}
}

// LoadSite loads a site configuration from the filesystem
func (sm *SiteManager) LoadSite(domain string) (*Site, error) {
	// Check if site is already loaded
	if site, exists := sm.sites[domain]; exists {
		return site, nil
	}

	sitePath := rootPath(GetEnv("SITES_PATH", "sites"), domain)

	// Check if site directory exists, create if it doesn't
	if !SecureExists(domain) {
		// Create site directory structure automatically
	}

	site := &Site{
		Domain:        domain,
		Name:          domain, // Default name to domain
		BasePath:      sitePath,
		IsActive:      true,
		Theme:         "pale-wisp", // Default theme
		ConfigPath:    filepath.Join(sitePath, "config"),
		PublicPath:    filepath.Join(sitePath, "public"),
		AssetsPath:    filepath.Join(sitePath, "assets"),
		PagesPath:     filepath.Join(sitePath, "pages"),
		LayoutPath:    filepath.Join(sitePath, "layout"),
		SectionsPath:  filepath.Join(sitePath, "sections"),
		TemplatesPath: filepath.Join(sitePath, "templates"),
		BlocksPath:    filepath.Join(sitePath, "blocks"),
		SnippetsPath:  filepath.Join(sitePath, "snippets"),
	}

	// Cache the site
	sm.sites[domain] = site

	return site, nil
}

// GetSite retrieves a site by domain
func (sm *SiteManager) GetSite(domain string) (*Site, error) {
	return sm.LoadSite(domain)
}

// GetSiteFromHost extracts domain from host header and gets the site
func (sm *SiteManager) GetSiteFromHost(host string) (*Site, error) {
	// Remove port if present
	domain := strings.Split(host, ":")[0]

	// Remove www. prefix if present
	if strings.HasPrefix(domain, "www.") {
		domain = strings.TrimPrefix(domain, "www.")
	}

	return sm.GetSite(domain)
}

// CreateSiteDirectories creates the necessary directory structure for a new site
func (sm *SiteManager) CreateSiteDirectories(domain string) error {
	sitePath, err := ValidateSitePath(domain)
	if err != nil {
		return fmt.Errorf("failed to validate site path %s: %w", sitePath, err)
	}

	// Create the base site directory first
	if err := os.MkdirAll(sitePath, 0755); err != nil {
		return fmt.Errorf("failed to create site directory %s: %w", sitePath, err)
	}

	// Define all required directories
	directories := []string{
		filepath.Join(sitePath, "config"),
		filepath.Join(sitePath, "config", "themes"),
		filepath.Join(sitePath, "public"),
		filepath.Join(sitePath, "assets"),
		filepath.Join(sitePath, "pages"),
		filepath.Join(sitePath, "pages", "(main)"),
		filepath.Join(sitePath, "layout"),
		filepath.Join(sitePath, "sections"),
		filepath.Join(sitePath, "templates"),
		filepath.Join(sitePath, "blocks"),
		filepath.Join(sitePath, "snippets"),
		filepath.Join(sitePath, "dbs"),
		filepath.Join(sitePath, "migrations"),
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create default config file if it doesn't exist
	configPath := filepath.Join(sitePath, "config", "config.toml")
	relativeConfigPath := filepath.Join("config", "config.toml")
	if !SecureExists(relativeConfigPath) {
		defaultConfig := fmt.Sprintf(`[site]
name = "%s"
domain = "%s"
description = "A new site powered by Wispy Core"
language = "en"
timezone = "UTC"
theme = "pale-wisp"
`, domain, domain)

		if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
	}

	// Create default theme files if they don't exist
	themes := map[string]string{
		"pale-wisp.css": `/* Pale Wisp Theme */
@plugin "daisyui" {
  themes: light --default;
}`,
		"midnight-wisp.css": `/* Midnight Wisp Theme */
@plugin "daisyui" {
  themes: dark --default;
}`,
	}

	for themeName, themeContent := range themes {
		themePath := filepath.Join(sitePath, "config", "themes", themeName)
		relativeThemePath := filepath.Join("config", "themes", themeName)
		if !SecureExists(relativeThemePath) {
			if err := os.WriteFile(themePath, []byte(themeContent), 0644); err != nil {
				return fmt.Errorf("failed to create default theme %s: %w", themeName, err)
			}
		}
	}

	// Create default layout if it doesn't exist
	layoutPath := filepath.Join(sitePath, "layout", "default.html")
	relativeLayoutPath := filepath.Join("layout", "default.html")
	if !SecureExists(relativeLayoutPath) {
		defaultLayout := `<!DOCTYPE html>
<html lang="en" class="h-full">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Page.Meta.Title}} - {{.Site.Name}}</title>
    <meta name="description" content="{{.Page.Meta.Description}}">
    
    <!-- daisyUI CSS Framework -->
    <link href="https://cdn.jsdelivr.net/npm/daisyui@5" rel="stylesheet" type="text/css" />
    <script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>
    
    {{import "css" .Site "/config/themes/pale-wisp.css"}}
</head>
<body class="min-h-screen bg-base-100 text-base-content">
    <div class="container mx-auto px-4 py-8">
        <!-- Main Content -->
        <main>
            {{block "page-content" .}}
            <div class="prose max-w-none">
                <h1>Welcome to {{.Site.Name}}</h1>
                <p>This is the default layout template.</p>
            </div>
            {{end}}
        </main>

        <!-- Footer -->
        <footer class="footer footer-center p-4 text-base-content mt-8">
            <div>
                <p>&copy; 2025 {{.Site.Name}}. Powered by Wispy Core.</p>
            </div>
        </footer>
    </div>
</body>
</html>`

		if err := os.WriteFile(layoutPath, []byte(defaultLayout), 0644); err != nil {
			return fmt.Errorf("failed to create default layout: %w", err)
		}
	}

	// Create default home page if it doesn't exist
	homePage := filepath.Join(sitePath, "pages", "(main)", "home.html")
	relativeHomePage := filepath.Join("pages", "(main)", "home.html")
	if !SecureExists(relativeHomePage) {
		defaultHomePage := `<!--
@name home.html
@url /
@author Wispy Core Team
@layout default
@is_draft false
@require_auth false
@required_roles []
-->
{{define "page-content"}}
<div class="hero min-h-screen bg-base-200">
    <div class="hero-content text-center">
        <div class="max-w-4xl">
            <h1 class="text-5xl font-bold mb-8">Welcome to {{.Site.Name}}</h1>
            <p class="text-xl mb-8">A new site powered by Wispy Core CMS</p>
            <div class="flex gap-4 justify-center">
                <a href="/admin" class="btn btn-primary btn-lg">Admin Panel</a>
                <a href="/about" class="btn btn-outline btn-lg">Learn More</a>
            </div>
        </div>
    </div>
</div>
{{end}}`

		if err := os.WriteFile(homePage, []byte(defaultHomePage), 0644); err != nil {
			return fmt.Errorf("failed to create default home page: %w", err)
		}
	}

	return nil
}

// GetTemplate loads and parses a template for the site
func (s *Site) GetTemplate(templateName string) (*template.Template, error) {
	templatePath := filepath.Join(s.TemplatesPath, templateName+".html")

	// Ensure templates directory exists
	if err := os.MkdirAll(s.TemplatesPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create templates directory: %w", err)
	}

	// Check if template exists
	relativeTemplatePath, pathErr := filepath.Rel(filepath.Join(rootPath(), "sites", s.Domain), templatePath)
	if pathErr != nil {
		return nil, fmt.Errorf("error getting relative template path: %w", pathErr)
	}
	if !SecureExists(relativeTemplatePath) {
		return nil, fmt.Errorf("template not found: %s", templateName)
	}

	// Parse template with layout
	tmpl := template.New(templateName + ".html")

	// Load layout templates
	layoutGlob := "layout/*.html"
	layouts, err := SecureGlob(layoutGlob)
	if err == nil && len(layouts) > 0 {
		tmpl, err = tmpl.ParseFiles(layouts...)
		if err != nil {
			return nil, fmt.Errorf("error parsing layout templates: %w", err)
		}
	}

	// Load snippets
	snippets, err := SecureGlob("snippets/*.html")
	if err == nil && len(snippets) > 0 {
		tmpl, err = tmpl.ParseFiles(snippets...)
		if err != nil {
			return nil, fmt.Errorf("error parsing snippet templates: %w", err)
		}
	}

	// Load sections
	sections, err := SecureGlob("sections/*.html")
	if err == nil && len(sections) > 0 {
		tmpl, err = tmpl.ParseFiles(sections...)
		if err != nil {
			return nil, fmt.Errorf("error parsing section templates: %w", err)
		}
	}

	// Load blocks
	blocks, err := SecureGlob("blocks/*.html")
	if err == nil && len(blocks) > 0 {
		tmpl, err = tmpl.ParseFiles(blocks...)
		if err != nil {
			return nil, fmt.Errorf("error parsing block templates: %w", err)
		}
	}

	// Parse the main template
	tmpl, err = tmpl.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("error parsing template %s: %w", templateName, err)
	}

	return tmpl, nil
}

// Authentication methods for SiteManager

// GetDB returns a database connection for the specified site and database
func (sm *SiteManager) GetDB(domain, dbName string) (*sql.DB, error) {
	return sm.dbCache.GetConnection(domain, dbName)
}

// createAuthTables creates all necessary authentication tables for a site
func (sm *SiteManager) createAuthTables(domain string) error {
	db, err := sm.GetDB(domain, "users")
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Create users table
	if _, err := db.Exec(models.CreateUserTableSQL); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create sessions table
	if _, err := db.Exec(models.CreateSessionTableSQL); err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	// Create OAuth accounts table
	if _, err := db.Exec(models.CreateOAuthAccountTableSQL); err != nil {
		return fmt.Errorf("failed to create oauth accounts table: %w", err)
	}

	return nil
}

// Register creates a new user account for a site
func (sm *SiteManager) Register(domain, email, password, firstName, lastName, displayName string) (*models.User, error) {
	db, err := sm.GetDB(domain, "users")
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	repository := auth.NewUserRepository(db)

	// Check if email already exists
	exists, err := repository.EmailExists(email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already registered")
	}

	// Validate password
	if !auth.IsValidPassword(password) {
		return nil, fmt.Errorf("password does not meet requirements")
	}

	// Hash password
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := models.NewUser(email, firstName, lastName, displayName)
	if displayName != "" {
		user.DisplayName = displayName
	}
	user.PasswordHash = passwordHash

	if err := repository.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user and creates a session for a site
func (sm *SiteManager) Login(domain, email, password, ipAddress, userAgent string) (*models.User, *models.Session, error) {
	db, err := sm.GetDB(domain, "users")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	repository := auth.NewUserRepository(db)
	config := auth.NewAuthConfig()
	sessionManager := auth.NewSessionManager(db, config)

	// Get user by email
	user, err := repository.GetUserByEmail(email)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Check if account is locked
	if repository.IsUserLocked(user) {
		return nil, nil, fmt.Errorf("account is locked")
	}

	// Verify password
	if err := auth.VerifyPassword(password, user.PasswordHash); err != nil {
		// Increment failed login count
		user.FailedLoginCount++

		// Lock account if too many failed attempts
		if user.FailedLoginCount >= config.GetMaxFailedLoginAttempts() {
			user.IsLocked = true
			lockUntil := time.Now().Add(config.GetAccountLockoutDuration())
			user.LockedUntil = &lockUntil
		}

		repository.UpdateUserLoginAttempt(user.ID, user.FailedLoginCount, user.LockedUntil, nil)
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Reset failed login count on successful login
	now := time.Now()
	user.FailedLoginCount = 0
	user.LastLoginAt = &now
	user.IsLocked = false
	user.LockedUntil = nil

	if err := repository.UpdateUserLoginAttempt(user.ID, 0, nil, &now); err != nil {
		return nil, nil, fmt.Errorf("failed to update login info: %w", err)
	}

	// Create session
	session, err := sessionManager.CreateSession(user.ID, ipAddress, userAgent)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	return user, session, nil
}

// ValidateSession checks if a session is valid and returns the user for a site
func (sm *SiteManager) ValidateSession(domain, sessionToken string) (*models.User, *models.Session, error) {
	db, err := sm.GetDB(domain, "users")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	config := auth.NewAuthConfig()
	sessionManager := auth.NewSessionManager(db, config)
	repository := auth.NewUserRepository(db)

	session, err := sessionManager.GetSession(sessionToken)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid session: %w", err)
	}

	if session.IsExpired() {
		sessionManager.DeleteSession(session.ID)
		return nil, nil, fmt.Errorf("session expired")
	}

	user, err := repository.GetUserByID(session.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found: %w", err)
	}

	if user.IsLocked {
		return nil, nil, fmt.Errorf("account is locked")
	}

	return user, session, nil
}

// GetSessionFromRequest extracts session from request for a site
func (sm *SiteManager) GetSessionFromRequest(domain string, r *http.Request) (*models.Session, error) {
	db, err := sm.GetDB(domain, "users")
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	config := auth.NewAuthConfig()
	sessionManager := auth.NewSessionManager(db, config)

	return sessionManager.GetSessionFromRequest(r)
}

// Authentication middleware methods

// RequireAuth middleware that requires authentication for a site
func (sm *SiteManager) RequireAuth(domain string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := sm.GetSessionFromRequest(domain, r)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			user, _, err := sm.ValidateSession(domain, session.Token)
			if err != nil { // Clear invalid session cookie
				db, _ := sm.GetDB(domain, "users")
				config := auth.NewAuthConfig()
				sessionManager := auth.NewSessionManager(db, config)
				sessionManager.ClearSessionCookie(w)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add user and session to request context
			ctx := auth.SetUserInContext(r.Context(), user)
			ctx = auth.SetSessionInContext(ctx, session)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRoles middleware that requires specific roles for a site
func (sm *SiteManager) RequireRoles(domain string, roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return sm.RequireAuth(domain)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := auth.GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if user has any of the required roles
			userRoles := user.GetRoles()
			hasRole := false
			for _, requiredRole := range roles {
				for _, userRole := range userRoles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}))
	}
}

// Close closes the database cache
func (sm *SiteManager) Close() error {
	sm.dbCache.CloseAll()
	return nil
}
