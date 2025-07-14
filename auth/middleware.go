package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ContextKey is the type used for context keys
type ContextKey string

const (
	// ContextKeyUser is the context key for user data
	ContextKeyUser ContextKey = "auth_user"

	// ContextKeySession is the context key for session data
	ContextKeySession ContextKey = "auth_session"
)

// Middleware contains authentication middleware functions
type Middleware struct {
	authProvider AuthProvider
	config       Config
}

// NewMiddleware creates a new authentication middleware
func NewMiddleware(authProvider AuthProvider, config Config) *Middleware {
	return &Middleware{
		authProvider: authProvider,
		config:       config,
	}
}

// RequireAuth is a middleware that requires authentication
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, session, err := m.getUserFromRequest(r)
		if err != nil {
			// Redirect to login page - use configured login URL or default to /login
			loginURL := m.config.LoginURL
			if loginURL == "" {
				loginURL = "/login"
			}
			http.Redirect(w, r, loginURL, http.StatusFound)
			return
		}

		// Store user and session in context
		ctx := context.WithValue(r.Context(), ContextKeyUser, user)
		ctx = context.WithValue(ctx, ContextKeySession, session)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuthFunc is a middleware function that requires authentication
func (m *Middleware) RequireAuthFunc(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return m.RequireAuth(http.HandlerFunc(next))
}

// RequireRole is a middleware that requires the user to have a specific role
func (m *Middleware) RequireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, session, err := m.getUserFromRequest(r)
		if err != nil {
			// Redirect to login page - use configured login URL or default to /login
			loginURL := m.config.LoginURL
			if loginURL == "" {
				loginURL = "/login"
			}
			http.Redirect(w, r, loginURL, http.StatusFound)
			return
		}

		// Check if the user has the required role
		hasRole := false
		for _, userRole := range user.Roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Store user and session in context
		ctx := context.WithValue(r.Context(), ContextKeyUser, user)
		ctx = context.WithValue(ctx, ContextKeySession, session)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRoleFunc is a middleware function that requires the user to have a specific role
func (m *Middleware) RequireRoleFunc(role string, next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return m.RequireRole(role, http.HandlerFunc(next))
}

// OptionalAuth is a middleware that adds user to context if authenticated but doesn't require auth
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, session, err := m.getUserFromRequest(r)
		if err == nil {
			// Store user and session in context
			ctx := context.WithValue(r.Context(), ContextKeyUser, user)
			ctx = context.WithValue(ctx, ContextKeySession, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			// Proceed without authentication
			next.ServeHTTP(w, r)
		}
	})
}

// OptionalAuthFunc is a middleware function that adds user to context if authenticated
func (m *Middleware) OptionalAuthFunc(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return m.OptionalAuth(http.HandlerFunc(next))
}

// UserFromContext extracts the user from the request context
func UserFromContext(ctx context.Context) (*User, error) {
	user, ok := ctx.Value(ContextKeyUser).(*User)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}

// SessionFromContext extracts the session from the request context
func SessionFromContext(ctx context.Context) (*Session, error) {
	session, ok := ctx.Value(ContextKeySession).(*Session)
	if !ok {
		return nil, errors.New("session not found in context")
	}
	return session, nil
}

// getUserFromRequest extracts and validates the user from the request
func (m *Middleware) getUserFromRequest(r *http.Request) (*User, *Session, error) {
	// Look for the token in the cookie
	cookie, err := r.Cookie(m.config.CookieName)
	if err != nil {
		return nil, nil, fmt.Errorf("no auth cookie found: %w", err)
	}

	token := cookie.Value

	// Also check for Authorization header as a fallback (useful for API calls)
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// If no token is found, fail
	if token == "" {
		return nil, nil, errors.New("no auth token found")
	}

	// Validate the session and get the user
	session, user, err := m.authProvider.ValidateSession(r.Context(), token)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid session: %w", err)
	}

	// Update the session IP and user agent if different
	if session.IP != r.RemoteAddr || session.UserAgent != r.UserAgent() {
		session.IP = r.RemoteAddr
		session.UserAgent = r.UserAgent()

		err = m.authProvider.GetSessionStore().UpdateSession(r.Context(), session)
		if err != nil {
			// Log the error but don't fail the request
			fmt.Printf("Failed to update session: %v\n", err)
		}
	}

	return user, session, nil
}

// SetAuthCookie sets the authentication cookie
func (m *Middleware) SetAuthCookie(w http.ResponseWriter, token string) {
	cookie := http.Cookie{
		Name:     m.config.CookieName,
		Value:    token,
		Path:     "/",
		Domain:   m.config.CookieDomain,
		Secure:   m.config.CookieSecure,
		HttpOnly: m.config.CookieHTTPOnly,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(m.config.TokenExpiration.Seconds()),
	}

	http.SetCookie(w, &cookie)
}

// ClearAuthCookie clears the authentication cookie
func (m *Middleware) ClearAuthCookie(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     m.config.CookieName,
		Value:    "",
		Path:     "/",
		Domain:   m.config.CookieDomain,
		Secure:   m.config.CookieSecure,
		HttpOnly: m.config.CookieHTTPOnly,
		MaxAge:   -1, // Delete the cookie
	}

	http.SetCookie(w, &cookie)
}
