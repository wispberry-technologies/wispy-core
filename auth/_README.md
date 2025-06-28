# Auth Package

A flexible authentication and authorization package for Go web applications, designed with clean interfaces for easy extension.

## Features

- User management (registration, login, profile updates)
- Session management with JWT tokens
- Role-based access control
- OAuth 2.0 authentication (Google, Discord)
- Pluggable storage backends (SQLite included, extensible for others)
- HTTP middleware for securing routes
- Form and API endpoints

## Installation

```bash
go get -u github.com/yourusername/wispy-core/auth
```

## Quick Start

```go
package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"wispy-core/auth"
)

func main() {
	// Create a configuration
	config := auth.Config{
		DBType:           "sqlite3",
		DBConn:           "auth.db",
		TokenSecret:      "your-secret-key",
		TokenExpiration:  24 * time.Hour,
		PasswordMinChars: 8,
		AllowSignup:      true,
		CookieName:       "auth_token",
		CookieSecure:     true,
		CookieHTTPOnly:   true,
		OAuthProviders: map[string]map[string]string{
			"google": {
				"client_id":     "your-google-client-id",
				"client_secret": "your-google-client-secret",
			},
			"discord": {
				"client_id":     "your-discord-client-id",
				"client_secret": "your-discord-client-secret",
			},
		},
	}

	// Initialize the auth package
	authProvider, middleware, authHandlers, oauthHandlers, err := auth.InitAuth(config)
	if err != nil {
		log.Fatalf("Failed to initialize auth: %v", err)
	}

	// Create a router
	mux := http.NewServeMux()

	// Register auth routes
	auth.RegisterHandlers(mux, authHandlers, oauthHandlers, middleware)

	// Protected route example
	mux.Handle("/admin", middleware.RequireRole("admin", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := auth.UserFromContext(r.Context())
		w.Write([]byte("Welcome to admin panel, " + user.DisplayName))
	})))

	// User route example
	mux.Handle("/profile", middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := auth.UserFromContext(r.Context())
		w.Write([]byte("Hello, " + user.DisplayName))
	})))

	// Start the server
	log.Println("Starting server on :8080")
	http.ListenAndServe(":8080", mux)
}
```

## Configuration

You can configure the auth package using a JSON configuration file:

```json
{
  "database": {
    "type": "sqlite3",
    "path": "auth.db"
  },
  "security": {
    "token_secret": "change-me-in-production",
    "token_expiration_hours": 24,
    "password_min_chars": 8
  },
  "oauth": {
    "providers": {
      "google": {
        "client_id": "your-google-client-id",
        "client_secret": "your-google-client-secret"
      },
      "discord": {
        "client_id": "your-discord-client-id",
        "client_secret": "your-discord-client-secret"
      }
    }
  },
  "application": {
    "allow_signup": true,
    "require_verify_email": false
  },
  "cookie": {
    "name": "auth_token",
    "domain": "",
    "secure": true,
    "http_only": true
  }
}
```

Load it using the provided function:

```go
config, err := auth.LoadConfigFromFile("auth.json")
if err != nil {
    log.Fatalf("Failed to load auth config: %v", err)
}
```

## User Management

### Registration

```go
user, err := authProvider.Register(ctx, "user@example.com", "username", "password")
if err != nil {
    log.Printf("Registration failed: %v", err)
    return
}
log.Printf("User registered: %s", user.ID)
```

### Login

```go
session, err := authProvider.Login(ctx, "user@example.com", "password")
if err != nil {
    log.Printf("Login failed: %v", err)
    return
}
log.Printf("User logged in: %s", session.UserID)
```

## OAuth Integration

The OAuth flow is handled by the included HTTP handlers. To initiate OAuth authentication:

1. Redirect the user to `/oauth/login?provider=google` (or `provider=discord`)
2. The user will be redirected to the OAuth provider's login page
3. After login, the provider will redirect back to `/oauth/callback`
4. The handler will create or update the user account and create a session
5. The user will be redirected to the home page (or a specified redirect URL)

### Automatic Account Creation for OAuth Users

When a user authenticates via OAuth (Google or Discord) for the first time, the system automatically:

1. Checks if a user with this OAuth ID already exists
2. If not found, creates a new user account with information from their OAuth profile
3. Generates a username from their display name
4. Assigns them the default "user" role
5. Creates a session without requiring a password

You can control this behavior with the `AllowSignup` configuration parameter:
- If `true` (default): New accounts are created automatically for OAuth users
- If `false`: Users must register through your app first before using OAuth login

OAuth users can log in without a password because they're authenticated by the OAuth provider. The system includes a special `CreateSessionForOAuthUser` method for this purpose:

```go
// For OAuth users only
session, err := authProvider.(*DefaultAuthProvider).CreateSessionForOAuthUser(ctx, oauthUser)
```

## Middleware Usage

### Require Authentication

```go
// Function style
mux.Handle("/profile", middleware.RequireAuthFunc(handleProfile))

// Handler style
mux.Handle("/profile", middleware.RequireAuth(http.HandlerFunc(handleProfile)))
```

### Require Role

```go
// Function style
mux.Handle("/admin", middleware.RequireRoleFunc("admin", handleAdmin))

// Handler style
mux.Handle("/admin", middleware.RequireRole("admin", http.HandlerFunc(handleAdmin)))
```

### Optional Authentication

```go
// Function style
mux.Handle("/", middleware.OptionalAuthFunc(handleHome))

// Handler style
mux.Handle("/", middleware.OptionalAuth(http.HandlerFunc(handleHome)))
```

## Extending Storage Backends

You can implement your own storage backends by implementing the `UserStore` and `SessionStore` interfaces:

```go
// Create a custom user store
type MyUserStore struct {
    // Your fields here
}

// Implement all methods from the UserStore interface
func (s *MyUserStore) CreateUser(ctx context.Context, user *auth.User) error {
    // Your implementation
}

// ... implement all other methods

// Same for session store
type MySessionStore struct {
    // Your fields here
}

// Implement all methods from the SessionStore interface
```

## Creating a Custom OAuth Provider

You can add support for additional OAuth providers by implementing the `OAuthProvider` interface:

```go
// Create a custom OAuth provider
type MyOAuthProvider struct {
    // Your fields here
}

// Implement all methods from the OAuthProvider interface
func (p *MyOAuthProvider) Name() string {
    return "myprovider"
}

// ... implement all other methods

// Register your provider
authProvider.RegisterOAuthProviders(myProvider)
```

