package common

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"wispy-core/auth"
	"wispy-core/models"
)

// RouteEntry represents a single route mapping
type RouteEntry struct {
	Pattern    string         // URL pattern like "/blog/post/:slug"
	PageSlug   string         // The page slug that handles this route
	Page       *Page          // The actual page object
	Parameters []string       // List of parameter names like ["slug"]
	Priority   int            // Lower number = higher priority
	Regex      *regexp.Regexp // Compiled regex for matching
}

// RouteInfo represents route information for API responses
type RouteInfo struct {
	Pattern    string   `json:"pattern"`
	PageSlug   string   `json:"page_slug"`
	Parameters []string `json:"parameters"`
	Priority   int      `json:"priority"`
	PageTitle  string   `json:"page_title"`
	IsDraft    bool     `json:"is_draft"`
}

// Site represents a single site in the multisite system
type Site struct {
	Domain   string `json:"domain"`
	Name     string `json:"name"`
	BasePath string `json:"base_path"`
	IsActive bool   `json:"is_active"`
	Theme    string `json:"theme"`
}

type SiteConfig struct {
	MaxFailedLoginAttempts         int
	FailedLoginAttemptLockDuration time.Duration
	SessionCookieSameSite          http.SameSite
	SessionCookieName              string
	SectionCookieMaxAge            time.Duration
	SessionTimeout                 time.Duration
	IsCookieSecure                 bool
}

// SiteInstance handles requests & data for individual sites
type SiteInstance struct {
	Domain    string
	Site      *Site
	Manager   *SiteInstanceManager
	Templates *template.Template // Precompiled templates for this site
	Config    SiteConfig         // Site-specific configuration
	Routes    []RouteEntry       // routes for this site
	Mu        sync.RWMutex       // mutex for thread-safe route access
}

// GetDB returns a database connection for the specified site and database
func (instance *SiteInstance) GetDB(dbName string) (*sql.DB, error) {
	return instance.Manager.dbCache.GetConnection(instance.Site.Domain, dbName)
}

// Register creates a new user account for a site
func Register(site *SiteInstance, email, password, firstName, lastName, displayName string) (*models.User, error) {
	db, err := site.GetDB("users")
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection \"%s\": %w", "users", err)
	}

	repository := auth.NewUserSqlDriver(db)

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
func (instance *SiteInstance) Login(email, password, ipAddress, userAgent string) (*models.User, *models.Session, error) {
	db, err := instance.GetDB("users")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	repository := auth.NewUserSqlDriver(db)

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
		if user.FailedLoginCount >= instance.Config.MaxFailedLoginAttempts {
			user.IsLocked = true
			lockUntil := time.Now().Add(instance.Config.FailedLoginAttemptLockDuration)
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
	sessionDriver := auth.NewSessionSqlDriver(db)
	session, err := sessionDriver.CreateSession(user.ID, ipAddress, userAgent)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	return user, session, nil
}

// ValidateSession checks if a session is valid and returns the user for a site
func ValidateSession(instance *SiteInstance, sessionToken string) (*models.User, *models.Session, error) {
	db, err := instance.GetDB("users")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	sessionDriver := auth.NewSessionSqlDriver(db)
	repository := auth.NewUserSqlDriver(db)

	session, err := sessionDriver.GetSession(sessionToken)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid session: %w", err)
	}

	if session.IsExpired() {
		sessionDriver.DeleteSession(session.ID)
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
func GetSessionFromRequest(instance *SiteInstance, r *http.Request) (*models.Session, error) {
	db, err := instance.GetDB("users")
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	sessionDriver := auth.NewSessionSqlDriver(db)

	return sessionDriver.GetSessionFromRequest(r)
}

// Authentication middleware methods

// RequireAuth middleware that requires authentication for a site
func (instance *SiteInstance) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := GetSessionFromRequest(instance, r)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			db, err := instance.GetDB("users")
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			user, _, err := auth.ValidateSession(db, session.Token)
			if err != nil { // Clear invalid session cookie
				db, _ := instance.GetDB("users")
				sessionDriver := auth.NewSessionSqlDriver(db)
				sessionDriver.ClearSessionCookie(w, r)
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
func (instance *SiteInstance) RequireRoles(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return instance.RequireAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

// Routing methods for SiteInstance

// parseRoutePattern converts a URL pattern like "/blog/post/:slug" into a regex
func (instance *SiteInstance) parseRoutePattern(pattern string) (*regexp.Regexp, []string, error) {
	if pattern == "" {
		pattern = "/"
	}

	// Ensure pattern starts with /
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}

	// Track parameter names
	var parameters []string

	// Escape special regex characters except for our parameter syntax
	regexPattern := regexp.QuoteMeta(pattern)

	// Replace escaped parameter patterns with regex groups
	paramRegex := regexp.MustCompile(`\\:([a-zA-Z][a-zA-Z0-9_]*)\\:`)
	regexPattern = paramRegex.ReplaceAllStringFunc(regexPattern, func(match string) string {
		// Extract parameter name (remove the escaped colons)
		paramName := strings.Trim(match, "\\:")
		parameters = append(parameters, paramName)
		// Return regex pattern for the parameter
		return "([^/]+)"
	})

	// Anchor the pattern to match the full path
	regexPattern = "^" + regexPattern + "$"

	compiled, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to compile route pattern '%s': %w", pattern, err)
	}

	return compiled, parameters, nil
}

// calculatePriority determines the priority of a route pattern
// Lower numbers have higher priority
func (instance *SiteInstance) calculatePriority(pattern string) int {
	if pattern == "/" {
		return 1000 // Home page gets lower priority to allow other exact matches
	}

	priority := 0
	segments := strings.Split(strings.Trim(pattern, "/"), "/")

	for _, segment := range segments {
		if strings.Contains(segment, ":") {
			// Parameter segments get lower priority
			priority += 100
		} else {
			// Static segments get higher priority
			priority += 1
		}
	}

	return priority
}

// AddRoute adds a new route for this site
func (instance *SiteInstance) AddRoute(page *Page) error {
	instance.Mu.Lock()
	defer instance.Mu.Unlock()

	if page.Meta.URL == "" {
		return fmt.Errorf("page URL pattern is required")
	}

	// Remove existing route for this page first
	instance.RemoveRoute(page.Slug)

	// Parse the URL pattern
	regex, parameters, err := instance.parseRoutePattern(page.Meta.URL)
	if err != nil {
		return fmt.Errorf("failed to parse route pattern '%s': %w", page.Meta.URL, err)
	}

	// Calculate priority
	priority := instance.calculatePriority(page.Meta.URL)

	// Create route entry
	route := RouteEntry{
		Pattern:    page.Meta.URL,
		PageSlug:   page.Slug,
		Page:       page,
		Parameters: parameters,
		Priority:   priority,
		Regex:      regex,
	}

	// Add route to site's routes
	instance.Routes = append(instance.Routes, route)

	// Sort routes by priority (lower number = higher priority)
	sort.Slice(instance.Routes, func(i, j int) bool {
		return instance.Routes[i].Priority < instance.Routes[j].Priority
	})

	return nil
}

// RemoveRoute removes a route for a page
func (instance *SiteInstance) RemoveRoute(pageSlug string) {
	instance.Mu.Lock()
	defer instance.Mu.Unlock()

	for i := len(instance.Routes) - 1; i >= 0; i-- {
		if instance.Routes[i].PageSlug == pageSlug {
			instance.Routes = append(instance.Routes[:i], instance.Routes[i+1:]...)
			break
		}
	}
}

// UpdateRoute updates the route for a page
func (instance *SiteInstance) UpdateRoute(page *Page) error {
	return instance.AddRoute(page) // AddRoute already handles removing existing routes
}

// FindRoute finds the matching route for a given URL path in this site
func (instance *SiteInstance) FindRoute(urlPath string) (*RouteEntry, map[string]string, error) {
	instance.Mu.RLock()
	defer instance.Mu.RUnlock()

	// Clean the URL path
	if urlPath == "" {
		urlPath = "/"
	}

	// Try to match against all routes in priority order
	for _, route := range instance.Routes {
		matches := route.Regex.FindStringSubmatch(urlPath)
		if matches != nil {
			// Extract parameters
			params := make(map[string]string)
			for i, paramName := range route.Parameters {
				if i+1 < len(matches) {
					params[paramName] = matches[i+1]
				}
			}

			return &route, params, nil
		}
	}

	return nil, nil, fmt.Errorf("no route found for path: %s", urlPath)
}
