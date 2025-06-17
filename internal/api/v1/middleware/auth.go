package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"wispy-core/internal/auth"
	"wispy-core/internal/cache"
	"wispy-core/pkg/models"
)

// Context keys
type contextKey string

const (
	UserContextKey         contextKey = "userKey"
	SessionContextKey      contextKey = "sessionKey"
	SiteContextKey         contextKey = "siteKey"
	SiteInstanceContextKey contextKey = "siteInstanceKey"
)

// authMiddleware validates user authentication for API requests
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get site instance from context
		siteInstance := r.Context().Value(SiteInstanceContextKey).(*models.SiteInstance)

		// Get database for the site
		db, err := cache.GetDB(siteInstance, "users")
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		// Check authentication
		user, session, err := validateSessionFromRequest(db, r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user and session to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserContextKey, user)
		ctx = context.WithValue(ctx, SessionContextKey, session)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateSessionFromRequest validates a session from the request
func validateSessionFromRequest(db *sql.DB, r *http.Request) (*auth.User, *auth.Session, error) {
	sessionToken := getSessionToken(r)
	if sessionToken == "" {
		return nil, nil, fmt.Errorf("no session token found")
	}

	return auth.ValidateSession(db, sessionToken)
}

// getSessionToken extracts session token from cookie or Authorization header
func getSessionToken(r *http.Request) string {
	// Try to get session token from cookie
	if cookie, err := r.Cookie("session"); err == nil {
		return cookie.Value
	}

	// Try to get from Authorization header
	if auth := r.Header.Get("Authorization"); auth != "" && strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	return ""
}

// siteContextMiddleware adds the site instance to the request context
func siteContextMiddleware(siteInstance *models.SiteInstance) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, SiteContextKey, siteInstance.Domain)
			ctx = context.WithValue(ctx, SiteInstanceContextKey, siteInstance)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
