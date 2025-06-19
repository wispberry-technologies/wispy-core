package auth

import (
	"context"
	"net/http"
	"wispy-core/internal/cache"
	"wispy-core/pkg/common"
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
func StrictAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get site instance from context
		siteInstance := r.Context().Value(SiteInstanceContextKey).(*models.SiteInstance)

		// Get database for the site
		dbCache := siteInstance.DBCache
		db, err := cache.GetConnection(dbCache, siteInstance.Domain, "users")
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		// Check authentication
		SqlSessionDriver := NewSessionSqlDriver(db)
		SqlUserDriver := NewUserSqlDriver(db)
		session, err := SqlSessionDriver.GetSessionFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		//  Get user from session
		user, err := SqlUserDriver.GetUserByID(session.UserID)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Add user and session to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserContextKey, user)
		ctx = context.WithValue(ctx, SessionContextKey, session)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SiteContextMiddleware adds the site instance to the request context
func SiteContextMiddleware(siteInstances map[string]*models.SiteInstance) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			host := common.NormalizeHost(r.Host)
			siteInstance, ok := siteInstances[host]
			if !ok {
				common.PlainTextError(w, http.StatusTeapot, "I am a Teapot Not ("+host+")!")
				return
			}

			ctx = context.WithValue(ctx, SiteContextKey, siteInstance.Domain)
			ctx = context.WithValue(ctx, SiteInstanceContextKey, siteInstance)

			//
			// ---
			// Get database for the site
			dbCache := siteInstance.DBCache
			db, err := cache.GetConnection(dbCache, siteInstance.Domain, "users")
			if err != nil {
				common.Debug("Failed to get db for %s: %v", siteInstance.Domain, err)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Check authentication
			SqlSessionDriver := NewSessionSqlDriver(db)
			session, err := SqlSessionDriver.GetSessionFromRequest(r)
			if err != nil {
				common.Debug("Failed to get session from request: %v", err)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Get user from session
			SqlUserDriver := NewUserSqlDriver(db)
			user, err := SqlUserDriver.GetUserByID(session.UserID)
			if err != nil {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			common.Debug("[%s] authenticated with session %s", user.DisplayName, session.ID)

			ctx = context.WithValue(ctx, UserContextKey, user)
			ctx = context.WithValue(ctx, SessionContextKey, session)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
