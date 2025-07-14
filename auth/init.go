package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"wispy-core/common"
)

// InitAuth initializes the auth package with the given configuration
func InitAuth(config Config) (AuthProvider, *Middleware, *AuthHandlers, *OAuthHandlers, error) {
	// Create the auth provider
	authProvider, err := NewDefaultAuthProvider(config)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create auth provider: %w", err)
	}

	// Create middleware
	middleware := NewMiddleware(authProvider, config)

	// Create HTTP handlers
	authHandlers := NewAuthHandlers(authProvider, middleware, config)
	oauthHandlers := NewOAuthHandlers(authProvider, middleware, config)

	return authProvider, middleware, authHandlers, oauthHandlers, nil
}

// DefaultConfig returns a default configuration for the auth package
func DefaultConfig() Config {
	return Config{
		DBType:           "sqlite3",
		DBConn:           ":memory:", // In-memory SQLite for quick tests
		TokenSecret:      "change-me-in-production",
		TokenExpiration:  24 * time.Hour,
		PasswordMinChars: 8,
		AllowSignup:      true,
		CookieName:       "auth_token",
		CookieSecure:     common.IsProduction(),
		CookieHTTPOnly:   true,
	}
}

// RegisterHandlers registers all HTTP handlers with the given mux
func RegisterHandlers(mux *http.ServeMux, authHandlers *AuthHandlers, oauthHandlers *OAuthHandlers, middleware *Middleware) {
	// Auth API endpoints (JSON)
	mux.Handle("/api/auth/login", http.HandlerFunc(authHandlers.HandleLogin))
	mux.Handle("/api/auth/register", http.HandlerFunc(authHandlers.HandleRegister))
	mux.Handle("/api/auth/logout", http.HandlerFunc(authHandlers.HandleLogout))
	mux.Handle("/api/auth/me", middleware.RequireAuth(http.HandlerFunc(authHandlers.HandleMe)))
	mux.Handle("/api/auth/refresh", middleware.RequireAuth(http.HandlerFunc(authHandlers.HandleRefreshToken)))

	// OAuth endpoints
	mux.Handle("/oauth/login", http.HandlerFunc(oauthHandlers.HandleOAuthLogin))
	mux.Handle("/oauth/callback", http.HandlerFunc(oauthHandlers.HandleOAuthCallback))

	// Web form endpoints
	mux.Handle("/login", middleware.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandlers.HandleLoginForm(w, r)
			return
		}

		// Check if already logged in
		_, err := UserFromContext(r.Context())
		if err == nil {
			// Already logged in, redirect to home
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// Render login page (not implemented here)
		// You'd render your login template here
		w.Write([]byte("Login page - implementation left to the caller"))
	})))

	mux.Handle("/register", middleware.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandlers.HandleRegisterForm(w, r)
			return
		}

		// Check if already logged in
		_, err := UserFromContext(r.Context())
		if err == nil {
			// Already logged in, redirect to home
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// Render register page (not implemented here)
		// You'd render your registration template here
		w.Write([]byte("Register page - implementation left to the caller"))
	})))

	mux.Handle("/logout", http.HandlerFunc(authHandlers.HandleLogout))
}

// InitSQLiteAuth initializes auth with a SQLite database
func InitSQLiteAuth(dbPath string, config Config) (AuthProvider, *Middleware, *AuthHandlers, *OAuthHandlers, error) {
	// Override configuration with SQLite settings
	config.DBType = "sqlite3"
	config.DBConn = dbPath

	return InitAuth(config)
}

// InitWithDB initializes auth with an existing database connection
func InitWithDB(db *sql.DB, dbType string, config Config) (AuthProvider, *Middleware, *AuthHandlers, *OAuthHandlers, error) {
	// Create the auth provider directly
	authProvider := &defaultAuthProvider{
		config:         config,
		oauthProviders: make(map[string]OAuthProvider),
	}

	// Create user and session stores based on DB type
	var userStore UserStore
	var sessionStore SessionStore
	var err error

	switch dbType {
	case "sqlite", "sqlite3":
		userStore, err = NewSQLiteUserStore(db)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to create SQLite user store: %w", err)
		}

		sessionStore, err = NewSQLiteSessionStore(db)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to create SQLite session store: %w", err)
		}

	default:
		return nil, nil, nil, nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	// Set up the stores
	authProvider.userStore = userStore
	authProvider.sessionStore = sessionStore

	// Initialize OAuth providers
	for name, providerConfig := range config.OAuthProviders {
		var provider OAuthProvider

		switch name {
		case "google":
			provider = NewGoogleOAuthProvider()
		case "discord":
			provider = NewDiscordOAuthProvider()
		default:
			continue
		}

		if err := provider.Configure(providerConfig); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to configure OAuth provider %s: %w", name, err)
		}

		authProvider.oauthProviders[name] = provider
	}

	// Create middleware and handlers
	middleware := NewMiddleware(authProvider, config)
	authHandlers := NewAuthHandlers(authProvider, middleware, config)
	oauthHandlers := NewOAuthHandlers(authProvider, middleware, config)

	return authProvider, middleware, authHandlers, oauthHandlers, nil
}
