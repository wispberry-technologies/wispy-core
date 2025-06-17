package authHandler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"wispy-core/internal/auth"
	"wispy-core/internal/cache"
	"wispy-core/pkg/common"
	"wispy-core/pkg/models"

	_ "github.com/glebarez/go-sqlite"
	"github.com/go-chi/chi/v5"
)

// Context key types for type safety
type contextKey string

const (
	// SiteContextKey is the key used to store the SiteInstance in the request context
	SiteContextKey contextKey = "site"
	// UserContextKey is the key used to store the authenticated user in the request context
	UserContextKey contextKey = "user"
	// SessionContextKey is the key used to store the session in the request context
	SessionContextKey contextKey = "session"
)

// Handler is the structure for the auth API handlers
type Handler struct{}

// New creates a new instance of the auth handler
func NewAuthApi() *Handler {
	return &Handler{}
}

// Helper functions to get context values

// getSiteFromContext retrieves the SiteInstance from the request context
func getSiteFromContext(ctx context.Context) *models.SiteInstance {
	if site, ok := ctx.Value(SiteContextKey).(*models.SiteInstance); ok {
		return site
	}
	return nil
}

// getUserFromContext retrieves the authenticated User from the request context
func getUserFromContext(ctx context.Context) *auth.User {
	if user, ok := ctx.Value(UserContextKey).(*auth.User); ok {
		return user
	}
	return nil
}

// WithAuth is a middleware that ensures a user is authenticated
func WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get site from context
		site := getSiteFromContext(r.Context())
		if site == nil {
			sendError(w, http.StatusInternalServerError, "Site context not found", "")
			return
		}

		// Get database for the site
		db, err := cache.GetDB(site, "users")
		if err != nil {
			sendError(w, http.StatusInternalServerError, "Database connection failed", err.Error())
			return
		}

		// Validate session
		user, session, err := validateSessionFromRequest(db, r)
		if err != nil {
			sendError(w, http.StatusUnauthorized, "Authentication required", "")
			return
		}

		// Add user and session to context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, SessionContextKey, session)

		// Call the next handler
		next(w, r.WithContext(ctx))
	}
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents an API error
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Register handles user registration
func Register(w http.ResponseWriter, r *http.Request) {
	// Get site from context
	siteInstance := getSiteFromContext(r.Context())
	if siteInstance == nil {
		sendError(w, http.StatusInternalServerError, "Site context not found", "")
		return
	}

	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get database for the site using cache
	db, err := cache.GetDB(siteInstance, "users")
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Database connection failed", err.Error())
		return
	}

	// Register user
	user, err := auth.Register(db, siteInstance.Domain, req.Email, req.Password, req.FirstName, req.LastName, req.DisplayName)
	if err != nil {
		if strings.Contains(err.Error(), "email already registered") {
			sendError(w, http.StatusConflict, "Email already registered", "")
		} else if strings.Contains(err.Error(), "password does not meet requirements") {
			sendError(w, http.StatusBadRequest, "Password does not meet requirements", "")
		} else {
			sendError(w, http.StatusInternalServerError, "Registration failed", err.Error())
		}
		return
	}

	// Return success response (without sensitive data)
	sendSuccess(w, http.StatusCreated, map[string]interface{}{
		"user":    sanitizeUser(user),
		"message": "User registered successfully",
	})
}

// Login handles user login
func Login(w http.ResponseWriter, r *http.Request) {
	// Get site from context
	siteInstance := getSiteFromContext(r.Context())
	if siteInstance == nil {
		sendError(w, http.StatusInternalServerError, "Site context not found", "")
		return
	}

	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get database for the site using cache
	db, err := cache.GetDB(siteInstance, "users")
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Database connection failed", err.Error())
		return
	}

	// Get client info
	ipAddress := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Login user
	user, session, err := auth.Login(db, siteInstance.Domain, req.Email, req.Password, ipAddress, userAgent, 5, 15*time.Minute)
	if err != nil {
		if strings.Contains(err.Error(), "account is locked") {
			sendError(w, http.StatusLocked, "Account is locked", "")
		} else {
			sendError(w, http.StatusUnauthorized, "Invalid credentials", "")
		}
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     siteInstance.SecurityConfig.SessionCookieName,
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil || siteInstance.SecurityConfig.IsCookieSecure,
		SameSite: siteInstance.SecurityConfig.SessionCookieSameSite,
		Expires:  session.ExpiresAt,
	})

	// Return success response
	sendSuccess(w, http.StatusOK, map[string]interface{}{
		"user":    sanitizeUser(user),
		"session": sanitizeSession(session),
		"message": "Login successful",
	})
}

// Logout handles user logout
func Logout(w http.ResponseWriter, r *http.Request) {
	// Get site from context
	siteInstance := getSiteFromContext(r.Context())
	if siteInstance == nil {
		sendError(w, http.StatusInternalServerError, "Site context not found", "")
		return
	}

	// Get database for the site using cache
	db, err := cache.GetDB(siteInstance, "users")
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Database connection failed", err.Error())
		return
	}

	// Get session from request
	session, err := auth.GetSessionFromRequest(db, r)
	if err == nil && session != nil {
		// Delete session
		sessionDriver := auth.NewSessionSqlDriver(db)
		sessionDriver.DeleteSession(session.ID)
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     siteInstance.SecurityConfig.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	sendSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Logout successful",
	})
}

// Me returns current user info
func Me(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "User not found", "")
		return
	}

	sendSuccess(w, http.StatusOK, map[string]interface{}{
		"user": sanitizeUser(user),
	})
}

// Refresh refreshes a session
func Refresh(w http.ResponseWriter, r *http.Request) {
	// Get site instance from context
	siteInstance := getSiteFromContext(r.Context())
	if siteInstance == nil {
		sendError(w, http.StatusInternalServerError, "Site context not found", "")
		return
	}

	// Get database for the site using cache
	db, err := cache.GetDB(siteInstance, "users")
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Database connection failed", err.Error())
		return
	}

	// Get current session
	session, err := auth.GetSessionFromRequest(db, r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Invalid session", "")
		return
	}

	// Refresh session
	sessionDriver := auth.NewSessionSqlDriver(db)
	err = sessionDriver.RefreshSession(session.ID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to refresh session", err.Error())
		return
	}

	// Set new session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     siteInstance.SecurityConfig.SessionCookieName,
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil || siteInstance.SecurityConfig.IsCookieSecure,
		SameSite: siteInstance.SecurityConfig.SessionCookieSameSite,
		Expires:  session.ExpiresAt,
	})

	sendSuccess(w, http.StatusOK, map[string]interface{}{
		"session": sanitizeSession(session),
		"message": "Session refreshed",
	})
}

// ChangePassword handles password changes
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req auth.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user from context
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "User not found", "")
		return
	}

	// Get site instance from context
	siteInstance := getSiteFromContext(r.Context())
	if siteInstance == nil {
		sendError(w, http.StatusInternalServerError, "Site instance not found in context", "")
		return
	}

	// Get database using cache
	db, err := cache.GetDB(siteInstance, "users")
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Database connection failed", err.Error())
		return
	}

	// Verify current password
	if err := auth.VerifyPassword(req.CurrentPassword, user.PasswordHash); err != nil {
		sendError(w, http.StatusUnauthorized, "Invalid current password", "")
		return
	}

	// Validate new password
	if !auth.IsValidPassword(req.NewPassword) {
		sendError(w, http.StatusBadRequest, "New password does not meet requirements", "")
		return
	}

	// Hash new password
	newPasswordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to hash password", err.Error())
		return
	}

	// Update password
	userDriver := auth.NewUserSqlDriver(db)
	if err := userDriver.UpdatePassword(user.ID, newPasswordHash); err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to update password", err.Error())
		return
	}

	sendSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Password changed successfully",
	})
}

// authMiddleware validates session and adds user to context
func authMiddleware(siteInstance *models.SiteInstance, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get database for the site using cache
		db, err := cache.GetDB(siteInstance, "users")
		if err != nil {
			sendError(w, http.StatusInternalServerError, "Database connection failed", err.Error())
			return
		}

		// Validate session
		user, session, err := validateSessionFromRequest(db, r)
		if err != nil {
			sendError(w, http.StatusUnauthorized, "Authentication required", "")
			return
		}

		// Add user and session to context (site is already in the context)
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserContextKey, user)
		ctx = context.WithValue(ctx, SessionContextKey, session)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

// OAuth handlers
// OAuthRedirect handles redirection to OAuth provider for authentication
func OAuthRedirect(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	if provider != "discord" {
		sendError(w, http.StatusBadRequest, "Unsupported OAuth provider", provider)
		return
	}

	// Get site from context
	siteInstance := getSiteFromContext(r.Context())
	if siteInstance == nil {
		sendError(w, http.StatusInternalServerError, "Site context not found", "")
		return
	}

	// Generate state parameter to prevent CSRF
	state := generateRandomString(32)

	// Store state in cookie for verification in callback
	stateCookie := http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil || siteInstance.SecurityConfig.IsCookieSecure,
		MaxAge:   600, // 10 minutes
		SameSite: siteInstance.SecurityConfig.SessionCookieSameSite,
	}
	http.SetCookie(w, &stateCookie)

	// Get Discord OAuth configuration from SiteInstance
	providerConfig, ok := siteInstance.Config.OAuthProviders[provider]

	// If not found in site config, try to get from environment variables
	if !ok || providerConfig.ClientID == "" {
		// Load from environment variables as fallback
		clientID := common.GetEnv("DISCORD_CLIENT_ID", "")
		redirectURI := common.GetEnv("DISCORD_REDIRECT_URI", "")

		if clientID != "" {
			// Initialize if needed
			if siteInstance.Config.OAuthProviders == nil {
				siteInstance.Config.OAuthProviders = make(map[string]models.OAuth)
			}

			// Store in SiteInstance for future use
			siteInstance.Config.OAuthProviders[provider] = models.OAuth{
				ClientID:     clientID,
				RedirectURI:  redirectURI,
				ClientSecret: common.GetEnv("DISCORD_CLIENT_SECRET", ""),
				Enabled:      true,
			}

			providerConfig = siteInstance.Config.OAuthProviders[provider]
		} else {
			sendError(w, http.StatusInternalServerError, "Discord OAuth client ID not configured", "")
			return
		}
	}

	// Build Discord authorization URL
	authURL := fmt.Sprintf(
		"https://discord.com/api/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=identify%%20email&state=%s",
		providerConfig.ClientID,
		providerConfig.RedirectURI,
		state,
	)

	// Redirect user to Discord
	http.Redirect(w, r, authURL, http.StatusFound)
}

// OAuthCallback handles the callback from the OAuth provider
func OAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Get site from context
	siteInstance := getSiteFromContext(r.Context())
	if siteInstance == nil {
		sendError(w, http.StatusInternalServerError, "Site context not found", "")
		return
	}
	provider := chi.URLParam(r, "provider")

	if provider != "discord" {
		sendError(w, http.StatusBadRequest, "Unsupported OAuth provider", provider)
		return
	}

	// Verify state parameter to prevent CSRF
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		sendError(w, http.StatusBadRequest, "Missing OAuth state", "")
		return
	}

	queryState := r.URL.Query().Get("state")
	if queryState == "" || queryState != stateCookie.Value {
		sendError(w, http.StatusBadRequest, "Invalid OAuth state", "")
		return
	}

	// Clear the state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		errorParam := r.URL.Query().Get("error")
		errorDescription := r.URL.Query().Get("error_description")
		sendError(w, http.StatusBadRequest, "OAuth authorization failed", fmt.Sprintf("%s: %s", errorParam, errorDescription))
		return
	}

	// Get provider configuration
	providerConfig, ok := siteInstance.Config.OAuthProviders[provider]
	if !ok || providerConfig.ClientID == "" || providerConfig.ClientSecret == "" {
		// Try environment variables as fallback
		clientID := common.GetEnv("DISCORD_CLIENT_ID", "")
		clientSecret := common.GetEnv("DISCORD_CLIENT_SECRET", "")

		if clientID != "" && clientSecret != "" {
			// Initialize if needed
			if siteInstance.Config.OAuthProviders == nil {
				siteInstance.Config.OAuthProviders = make(map[string]models.OAuth)
			}

			// Store in SiteInstance for future use
			siteInstance.Config.OAuthProviders[provider] = models.OAuth{
				ClientID:     clientID,
				ClientSecret: clientSecret,
				RedirectURI:  common.GetEnv("DISCORD_REDIRECT_URI", ""),
				Enabled:      true,
			}

			providerConfig = siteInstance.Config.OAuthProviders[provider]
		} else {
			sendError(w, http.StatusInternalServerError, "Discord OAuth configuration incomplete", "")
			return
		}
	}

	// This is where the actual OAuth token exchange would happen using providerConfig values
	// For now we'll display a message indicating Discord login is coming soon
	http.Redirect(w, r, "/?message=Discord+login+coming+soon", http.StatusFound)
}

// Helper functions

func sendSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	response := &APIResponse{
		Success: true,
		Data:    data,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func sendError(w http.ResponseWriter, statusCode int, message, details string) {
	response := &APIResponse{
		Success: false,
		Error: &APIError{
			Code:    statusCode,
			Message: message,
			Details: details,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func sanitizeUser(user *auth.User) map[string]interface{} {
	return map[string]interface{}{
		"id":           user.ID,
		"email":        user.Email,
		"first_name":   user.FirstName,
		"last_name":    user.LastName,
		"display_name": user.DisplayName,
		"avatar":       user.Avatar,
		"roles":        user.Roles,
		"is_active":    user.IsActive,
		"created_at":   user.CreatedAt,
		"updated_at":   user.UpdatedAt,
	}
}

func sanitizeSession(session *auth.Session) map[string]interface{} {
	return map[string]interface{}{
		"id":         session.ID,
		"expires_at": session.ExpiresAt,
		"created_at": session.CreatedAt,
	}
}

func getClientIP(r *http.Request) string {
	// Try to get real IP from headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func validateSessionFromRequest(db *sql.DB, r *http.Request) (*auth.User, *auth.Session, error) {
	// Try to get session token from cookie or Authorization header
	var sessionToken string

	// Get site from context to get the correct cookie name
	site := getSiteFromContext(r.Context())
	if site == nil {
		return nil, nil, fmt.Errorf("site context not found")
	}

	// Try cookie first
	if cookie, err := r.Cookie(site.SecurityConfig.SessionCookieName); err == nil {
		sessionToken = cookie.Value
	} else if auth := r.Header.Get("Authorization"); auth != "" {
		// Try Bearer token
		if strings.HasPrefix(auth, "Bearer ") {
			sessionToken = strings.TrimPrefix(auth, "Bearer ")
		}
	}

	if sessionToken == "" {
		return nil, nil, fmt.Errorf("no session token found")
	}

	return auth.ValidateSession(db, sessionToken)
}

// Helper functions for handling API responses

// generateRandomString creates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
