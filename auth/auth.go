package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"wispy-core/common"
)

// DefaultAuthProvider implements AuthProvider interface
type defaultAuthProvider struct {
	config            Config
	userStore         UserStore
	sessionStore      SessionStore
	oauthProviders    map[string]OAuthProvider
	resetTokens       map[string]*PasswordResetToken // In-memory store for reset tokens
	resetTokensByUser map[string]string              // User ID to token mapping
}

// NewDefaultAuthProvider creates a new default auth provider
func NewDefaultAuthProvider(config Config) (AuthProvider, error) {
	provider := &defaultAuthProvider{
		config:            config,
		oauthProviders:    make(map[string]OAuthProvider),
		resetTokens:       make(map[string]*PasswordResetToken),
		resetTokensByUser: make(map[string]string),
	}

	if err := provider.Configure(config); err != nil {
		return nil, err
	}

	return provider, nil
}

// Configure implements AuthProvider.Configure
func (p *defaultAuthProvider) Configure(config Config) error {
	p.config = config

	// Set up the database based on config
	var db *sql.DB
	var err error

	switch strings.ToLower(config.DBType) {
	case "sqlite", "sqlite3":
		db, err = sql.Open("sqlite3", config.DBConn)
		if err != nil {
			return fmt.Errorf("failed to open SQLite database: %w", err)
		}

		if err = db.Ping(); err != nil {
			return fmt.Errorf("failed to connect to SQLite database: %w", err)
		}

		// Create user and session stores
		userStore, err := NewSQLiteUserStore(db)
		if err != nil {
			return fmt.Errorf("failed to create user store: %w", err)
		}
		p.userStore = userStore

		sessionStore, err := NewSQLiteSessionStore(db)
		if err != nil {
			return fmt.Errorf("failed to create session store: %w", err)
		}
		p.sessionStore = sessionStore

	// Add more database types here
	default:
		return fmt.Errorf("unsupported database type: %s", config.DBType)
	}

	// Initialize OAuth providers if configured
	for name, providerConfig := range config.OAuthProviders {
		var provider OAuthProvider

		switch strings.ToLower(name) {
		case "google":
			provider = NewGoogleOAuthProvider()
		case "discord":
			provider = NewDiscordOAuthProvider()
		// Add more providers here
		default:
			continue // Skip unknown providers
		}

		if err := provider.Configure(providerConfig); err != nil {
			return fmt.Errorf("failed to configure OAuth provider %s: %w", name, err)
		}

		p.oauthProviders[name] = provider
	}

	return nil
}

// GetUserStore implements AuthProvider.GetUserStore
func (p *defaultAuthProvider) GetUserStore() UserStore {
	return p.userStore
}

// GetSessionStore implements AuthProvider.GetSessionStore
func (p *defaultAuthProvider) GetSessionStore() SessionStore {
	return p.sessionStore
}

// Login implements AuthProvider.Login
func (p *defaultAuthProvider) Login(ctx context.Context, email, password string) (*Session, error) {
	// Find user by email
	user, err := p.userStore.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Special case for OAuth users (they might not have a password)
	if user.OAuthProvider != "" && password == "" {
		// OAuth users can log in without password if they're using the special OAuth login flow
		// This allows sessions to be created for OAuth users
	} else {
		// For regular password-based auth, verify the password
		valid, err := p.userStore.VerifyPassword(ctx, user.ID, password)
		if err != nil || !valid {
			return nil, fmt.Errorf("invalid email or password")
		}
	}

	// Update last login time
	user.LastLogin = time.Now()
	if err := p.userStore.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user login time: %w", err)
	}

	// Create a new session
	session, err := p.createSession(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// Register implements AuthProvider.Register
func (p *defaultAuthProvider) Register(ctx context.Context, email, username, password string) (*User, error) {
	// Check if registration is allowed
	if !p.config.AllowSignup {
		return nil, errors.New("registration is disabled")
	}

	// Validate email, username, and password
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	if err := validateUsername(username); err != nil {
		return nil, err
	}

	if err := validatePassword(password, p.config.PasswordMinChars); err != nil {
		return nil, err
	}

	// Check if email already exists
	_, err := p.userStore.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, errors.New("email is already registered")
	}

	// Check if username already exists
	_, err = p.userStore.GetUserByUsername(ctx, username)
	if err == nil {
		return nil, errors.New("username is already taken")
	}

	// Create the user
	user := &User{
		Email:    email,
		Username: username,
		Password: password,         // Will be hashed by the store
		Roles:    []string{"user"}, // Default role
	}

	if err := p.userStore.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Logout implements AuthProvider.Logout
func (p *defaultAuthProvider) Logout(ctx context.Context, sessionToken string) error {
	session, err := p.sessionStore.GetSessionByToken(ctx, sessionToken)
	if err != nil {
		return nil // If session doesn't exist, consider it already logged out
	}

	return p.sessionStore.DeleteSession(ctx, session.ID)
}

// GetOAuthProvider implements AuthProvider.GetOAuthProvider
func (p *defaultAuthProvider) GetOAuthProvider(name string) (OAuthProvider, error) {
	provider, exists := p.oauthProviders[name]
	if !exists {
		return nil, fmt.Errorf("OAuth provider not found: %s", name)
	}
	return provider, nil
}

// RegisterOAuthProviders implements AuthProvider.RegisterOAuthProviders
func (p *defaultAuthProvider) RegisterOAuthProviders(providers ...OAuthProvider) {
	for _, provider := range providers {
		p.oauthProviders[provider.Name()] = provider
	}
}

// ValidateSession implements AuthProvider.ValidateSession
func (p *defaultAuthProvider) ValidateSession(ctx context.Context, token string) (*Session, *User, error) {
	session, err := p.sessionStore.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid session: %w", err)
	}

	// Check if session has expired
	if session.ExpiresAt.Before(time.Now()) {
		_ = p.sessionStore.DeleteSession(ctx, session.ID)
		return nil, nil, errors.New("session has expired")
	}

	// Get the user associated with this session
	user, err := p.userStore.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found: %w", err)
	}

	return session, user, nil
}

// RefreshSession implements AuthProvider.RefreshSession
func (p *defaultAuthProvider) RefreshSession(ctx context.Context, token string) (*Session, error) {
	// Get the existing session
	oldSession, err := p.sessionStore.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	// Check if session has expired
	if oldSession.ExpiresAt.Before(time.Now()) {
		_ = p.sessionStore.DeleteSession(ctx, oldSession.ID)
		return nil, errors.New("session has expired")
	}

	// Store the old token for logging
	oldToken := oldSession.Token

	// Generate a new token
	newToken, err := p.generateToken(oldSession.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Log the token change
	common.Debug("Session refresh: Old token=%s..., New token=%s...",
		oldToken[:5]+"...",
		newToken[:5]+"...")

	// Generate a random session ID
	randomBytes := make([]byte, 16)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}
	sessionID := hex.EncodeToString(randomBytes)

	// Create new session with new token but same user
	newSession := &Session{
		ID:        sessionID,
		UserID:    oldSession.UserID,
		Token:     newToken,
		ExpiresAt: time.Now().Add(p.config.TokenExpiration),
		IP:        oldSession.IP,
		UserAgent: oldSession.UserAgent,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Data:      oldSession.Data,
	}

	common.Debug("Session refresh: Old token=%s, New token=%s",
		oldSession.Token[:5]+"...", newToken[:5]+"...")

	// Delete the old session
	if err := p.sessionStore.DeleteSession(ctx, oldSession.ID); err != nil {
		common.Error("Failed to delete old session during refresh: %v", err)
		// Continue anyway, we'll create a new one
	}

	// Create the new session
	if err := p.sessionStore.CreateSession(ctx, newSession); err != nil {
		return nil, fmt.Errorf("failed to create new session: %w", err)
	}

	// Verify the new session was created
	verifySession, err := p.sessionStore.GetSessionByToken(ctx, newToken)
	if err != nil {
		common.Error("Failed to verify new session creation: %v", err)
	} else {
		common.Debug("New session verified with ID=%s and token=%s",
			verifySession.ID, verifySession.Token[:5]+"...")
	}

	return newSession, nil
}

// GeneratePasswordResetToken implements AuthProvider.GeneratePasswordResetToken
func (p *defaultAuthProvider) GeneratePasswordResetToken(ctx context.Context, email string) (string, error) {
	// Find user by email
	user, err := p.userStore.GetUserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	// Create a new reset token
	resetToken, err := p.CreatePasswordResetToken(ctx, user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to create reset token: %w", err)
	}

	return resetToken.Token, nil
}

// CreatePasswordResetToken creates a new password reset token for a user
func (p *defaultAuthProvider) CreatePasswordResetToken(ctx context.Context, userID string) (*PasswordResetToken, error) {
	// Generate a secure random token
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate reset token: %w", err)
	}
	tokenString := hex.EncodeToString(tokenBytes)

	// Create the token record
	token := &PasswordResetToken{
		Token:     tokenString,
		UserID:    userID,
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour expiration
		CreatedAt: time.Now(),
	}

	// Store the token
	p.resetTokens[tokenString] = token

	// If there's an existing token for this user, remove it
	if oldToken, exists := p.resetTokensByUser[userID]; exists {
		delete(p.resetTokens, oldToken)
	}

	// Associate the new token with the user
	p.resetTokensByUser[userID] = tokenString

	return token, nil
}

// ResetPassword implements AuthProvider.ResetPassword
func (p *defaultAuthProvider) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Validate the token
	resetToken, err := p.ValidatePasswordResetToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid reset token: %w", err)
	}

	// Get the user ID from the token
	userID := resetToken.UserID

	// Validate the new password
	if err := validatePassword(newPassword, p.config.PasswordMinChars); err != nil {
		return err
	}

	// Update the user's password
	if err := p.userStore.UpdatePassword(ctx, userID, newPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	} // Delete the token so it can't be used again
	if err := p.DeletePasswordResetToken(ctx, token); err != nil {
		common.Error("Failed to delete used reset token: %v", err)
		// Continue anyway since the password was reset successfully
	}

	return nil
}

// ValidatePasswordResetToken validates a password reset token
func (p *defaultAuthProvider) ValidatePasswordResetToken(ctx context.Context, token string) (*PasswordResetToken, error) {
	// Look up the token
	resetToken, exists := p.resetTokens[token]
	if !exists {
		return nil, errors.New("token not found")
	}

	// Check if token has expired
	if resetToken.ExpiresAt.Before(time.Now()) {
		// Remove expired token
		delete(p.resetTokens, token)
		delete(p.resetTokensByUser, resetToken.UserID)
		return nil, errors.New("token has expired")
	}

	return resetToken, nil
}

// DeletePasswordResetToken deletes a password reset token
func (p *defaultAuthProvider) DeletePasswordResetToken(ctx context.Context, token string) error {
	resetToken, exists := p.resetTokens[token]
	if !exists {
		return nil // Token already deleted or doesn't exist
	}

	// Remove from maps
	delete(p.resetTokens, token)
	delete(p.resetTokensByUser, resetToken.UserID)

	return nil
}

// Helper method to create a new session
func (p *defaultAuthProvider) createSession(ctx context.Context, userID string) (*Session, error) {
	// Generate a token
	token, err := p.generateToken(userID)
	if err != nil {
		return nil, err
	}

	// Create a session
	session := &Session{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(p.config.TokenExpiration),
		// IP and UserAgent should be set by middleware
	}

	// Store the session
	if err := p.sessionStore.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// Helper method to generate a secure random token
func (p *defaultAuthProvider) generateToken(userID string) (string, error) {
	// Generate random bytes for the token
	const tokenLength = 32
	randomBytes := make([]byte, tokenLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Convert to hex string for usability
	tokenString := hex.EncodeToString(randomBytes)

	return tokenString, nil
}

// Helper function to validate email format
func validateEmail(email string) error {
	// This is a simple validation. In a real implementation, you'd want a more robust solution.
	if email == "" || !strings.Contains(email, "@") {
		return errors.New("invalid email address")
	}
	return nil
}

// Helper function to validate username
func validateUsername(username string) error {
	if len(username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	return nil
}

// Helper function to validate password
func validatePassword(password string, minLength int) error {
	if len(password) < minLength {
		return fmt.Errorf("password must be at least %d characters", minLength)
	}
	return nil
}
