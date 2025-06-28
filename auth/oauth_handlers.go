package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// OAuthHandlers contains HTTP handlers for OAuth authentication flows
type OAuthHandlers struct {
	authProvider AuthProvider
	middleware   *Middleware
	config       Config
	stateStore   map[string]string // In a real app, use a secure storage for state tokens
}

// NewOAuthHandlers creates new OAuth handlers
func NewOAuthHandlers(authProvider AuthProvider, middleware *Middleware, config Config) *OAuthHandlers {
	return &OAuthHandlers{
		authProvider: authProvider,
		middleware:   middleware,
		config:       config,
		stateStore:   make(map[string]string),
	}
}

// HandleOAuthLogin initiates OAuth login flow
func (h *OAuthHandlers) HandleOAuthLogin(w http.ResponseWriter, r *http.Request) {
	// Extract provider name from query or path
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		http.Error(w, "Provider not specified", http.StatusBadRequest)
		return
	}

	// Get the OAuth provider
	oauthProvider, err := h.authProvider.GetOAuthProvider(provider)
	if err != nil {
		http.Error(w, "Provider not supported", http.StatusBadRequest)
		return
	}

	// Generate a secure state parameter to prevent CSRF
	state, err := generateSecureState()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store state with provider information
	h.stateStore[state] = provider

	// Build the redirect URI for the callback
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	redirectURI := fmt.Sprintf("%s://%s/oauth/callback", scheme, r.Host)

	// Redirect to provider's auth URL
	authURL := oauthProvider.GetAuthURL(state, redirectURI)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// HandleOAuthCallback handles OAuth callback
func (h *OAuthHandlers) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Get state and code from query parameters
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	// Validate state to prevent CSRF
	provider, exists := h.stateStore[state]
	if !exists {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Clean up state
	delete(h.stateStore, state)

	// Get the OAuth provider
	oauthProvider, err := h.authProvider.GetOAuthProvider(provider)
	if err != nil {
		http.Error(w, "Provider not supported", http.StatusBadRequest)
		return
	}

	// Build the redirect URI
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	redirectURI := fmt.Sprintf("%s://%s/oauth/callback", scheme, r.Host)

	// Exchange the authorization code for a token
	token, err := oauthProvider.ExchangeCode(r.Context(), code, redirectURI)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange code: %v", err), http.StatusInternalServerError)
		return
	}

	// Get user info from the provider
	userInfo, err := oauthProvider.GetUserInfo(r.Context(), token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if a user with this OAuth ID already exists
	user, err := h.authProvider.GetUserStore().GetUserByOAuthID(r.Context(), provider, userInfo.ID)
	if err != nil {
		// User doesn't exist, create a new one if allowed
		if !h.config.AllowSignup {
			http.Error(w, "Registration is disabled", http.StatusForbidden)
			return
		}

		// Generate a random username if not available
		username := strings.ToLower(strings.ReplaceAll(userInfo.Name, " ", "."))
		if username == "" {
			username = fmt.Sprintf("user%s", userInfo.ID[:8])
		}

		// Create metadata with additional info
		metadata, _ := json.Marshal(map[string]interface{}{
			"oauth_provider":  provider,
			"oauth_user_info": userInfo,
		})

		// Create the user
		user = &User{
			Email:         userInfo.Email,
			Username:      username,
			DisplayName:   userInfo.Name,
			OAuthProvider: provider,
			OAuthID:       userInfo.ID,
			Roles:         []string{"user"},
			Metadata:      metadata,
		}

		err = h.authProvider.GetUserStore().CreateUser(r.Context(), user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Create a session for the OAuth user using our special method
	session, err := h.authProvider.(*DefaultAuthProvider).CreateSessionForOAuthUser(r.Context(), user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create session: %v", err), http.StatusInternalServerError)
		return
	}

	// Set auth cookie
	h.middleware.SetAuthCookie(w, session.Token)

	// Redirect to home page or a specified redirect URL
	redirectTo := r.URL.Query().Get("redirect_to")
	if redirectTo == "" {
		redirectTo = "/"
	}

	http.Redirect(w, r, redirectTo, http.StatusFound)
}

// GenerateSecureState generates a secure random state string
func generateSecureState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
