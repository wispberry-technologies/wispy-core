package auth

import (
	"net/http"
	"strings"

	"github.com/wispberry-technologies/wispy-core/models"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	authManager *AuthManager
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authManager *AuthManager) *AuthMiddleware {
	return &AuthMiddleware{
		authManager: authManager,
	}
}

// RequireAuth middleware that requires authentication
func (am *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := am.authManager.session.GetSessionFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, _, err := am.authManager.ValidateSession(session.Token)
		if err != nil {
			// Clear invalid session cookie
			am.authManager.session.ClearSessionCookie(w)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user and session to request context
		ctx := SetUserInContext(r.Context(), user)
		ctx = SetSessionInContext(ctx, session)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRoles middleware that requires specific roles
func (am *AuthMiddleware) RequireRoles(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return am.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
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

// OptionalAuth middleware that adds user context if authenticated
func (am *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := am.authManager.session.GetSessionFromRequest(r)
		if err == nil {
			user, _, err := am.authManager.ValidateSession(session.Token)
			if err == nil {
				// Add user and session to request context
				ctx := SetUserInContext(r.Context(), user)
				ctx = SetSessionInContext(ctx, session)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}

// AdminOnly middleware that requires admin role
func (am *AuthMiddleware) AdminOnly(next http.Handler) http.Handler {
	return am.RequireRoles("admin")(next)
}

// GetClientIP extracts the client IP address from the request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP if there are multiple
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}
	return ip
}

// GetUserAgent extracts the user agent from the request
func GetUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

// IsAuthenticated checks if the request has a valid session
func (am *AuthMiddleware) IsAuthenticated(r *http.Request) bool {
	session, err := am.authManager.session.GetSessionFromRequest(r)
	if err != nil {
		return false
	}

	_, _, err = am.authManager.ValidateSession(session.Token)
	return err == nil
}

// GetCurrentUser returns the current authenticated user from request context
func GetCurrentUser(r *http.Request) (*models.User, bool) {
	return GetUserFromContext(r.Context())
}

// GetCurrentSession returns the current session from request context
func GetCurrentSession(r *http.Request) (*models.Session, bool) {
	return GetSessionFromContext(r.Context())
}
