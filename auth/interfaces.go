package auth

import (
	"context"
	"time"
)

// User represents the basic user identity
type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Password    string    `json:"-"` // Hashed, never sent to client
	Roles       []string  `json:"roles"`
	Metadata    []byte    `json:"metadata,omitempty"` // Additional user data as JSON
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastLogin   time.Time `json:"last_login,omitempty"`

	// OAuth related fields
	OAuthProvider string `json:"oauth_provider,omitempty"`
	OAuthID       string `json:"oauth_id,omitempty"`
}

// PasswordResetToken represents a token for resetting a password
type PasswordResetToken struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"` // JWT or other token format
	ExpiresAt time.Time `json:"expires_at"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Data      []byte    `json:"data,omitempty"` // Session data as JSON
}

// UserStore defines the interface for user persistence
type UserStore interface {
	// Core operations
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) error

	// Additional operations
	ListUsers(ctx context.Context, offset, limit int) ([]*User, error)
	CountUsers(ctx context.Context) (int, error)

	// Authentication operations
	VerifyPassword(ctx context.Context, userID, password string) (bool, error)
	UpdatePassword(ctx context.Context, userID, password string) error

	// Role management
	AddUserToRole(ctx context.Context, userID, role string) error
	RemoveUserFromRole(ctx context.Context, userID, role string) error
	GetUserRoles(ctx context.Context, userID string) ([]string, error)

	// OAuth operations
	GetUserByOAuthID(ctx context.Context, provider, oauthID string) (*User, error)
	LinkOAuthToUser(ctx context.Context, userID, provider, oauthID string) error
	UnlinkOAuthFromUser(ctx context.Context, userID, provider string) error
}

// SessionStore defines the interface for session persistence
type SessionStore interface {
	// Core operations
	CreateSession(ctx context.Context, session *Session) error
	GetSessionByID(ctx context.Context, id string) (*Session, error)
	GetSessionByToken(ctx context.Context, token string) (*Session, error)
	UpdateSession(ctx context.Context, session *Session) error
	DeleteSession(ctx context.Context, id string) error

	// Session management
	GetUserSessions(ctx context.Context, userID string) ([]*Session, error)
	DeleteUserSessions(ctx context.Context, userID string) error
	DeleteExpiredSessions(ctx context.Context) (int, error)

	// Data management
	StoreSessionData(ctx context.Context, sessionID string, key string, value interface{}) error
	GetSessionData(ctx context.Context, sessionID string, key string, valuePtr interface{}) error
	RemoveSessionData(ctx context.Context, sessionID string, key string) error
}

// AuthProvider is the main interface for authentication operations
type AuthProvider interface {
	// User management
	GetUserStore() UserStore

	// Session management
	GetSessionStore() SessionStore

	// Token management
	CreatePasswordResetToken(ctx context.Context, userID string) (*PasswordResetToken, error)
	ValidatePasswordResetToken(ctx context.Context, token string) (*PasswordResetToken, error)
	DeletePasswordResetToken(ctx context.Context, token string) error

	// Authentication
	Login(ctx context.Context, email, password string) (*Session, error)
	Register(ctx context.Context, email, username, password string) (*User, error)
	Logout(ctx context.Context, sessionToken string) error

	// OAuth functionality
	GetOAuthProvider(name string) (OAuthProvider, error)
	RegisterOAuthProviders(providers ...OAuthProvider)

	// Session validation
	ValidateSession(ctx context.Context, token string) (*Session, *User, error)
	RefreshSession(ctx context.Context, token string) (*Session, error)

	// Utility functions
	GeneratePasswordResetToken(ctx context.Context, email string) (string, error)
	ResetPassword(ctx context.Context, token, newPassword string) error

	// Configuration
	Configure(config Config) error
}

// OAuthProvider defines the interface for OAuth authentication providers
type OAuthProvider interface {
	// Provider information
	Name() string
	DisplayName() string

	// OAuth flow
	GetAuthURL(state string, redirectURI string) string
	ExchangeCode(ctx context.Context, code string, redirectURI string) (*OAuthToken, error)
	GetUserInfo(ctx context.Context, token *OAuthToken) (*OAuthUserInfo, error)

	// Configuration
	Configure(config map[string]string) error
}

// OAuthToken holds the OAuth access and refresh tokens
type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`
}

// OAuthUserInfo contains user information returned from the OAuth provider
type OAuthUserInfo struct {
	ID            string                 `json:"id"`
	Email         string                 `json:"email"`
	VerifiedEmail bool                   `json:"verified_email"`
	Name          string                 `json:"name"`
	GivenName     string                 `json:"given_name,omitempty"`
	FamilyName    string                 `json:"family_name,omitempty"`
	Picture       string                 `json:"picture,omitempty"`
	Locale        string                 `json:"locale,omitempty"`
	RawData       map[string]interface{} `json:"raw_data,omitempty"`
}

// Config defines configuration options for the auth package
type Config struct {
	// Database configuration
	DBType string `json:"db_type"`       // "sqlite", "postgres", etc.
	DBConn string `json:"db_connection"` // Connection string or file path

	// Security settings
	TokenSecret      string        `json:"token_secret"`       // Secret key for signing tokens
	TokenExpiration  time.Duration `json:"token_expiration"`   // Duration for token validity
	PasswordMinChars int           `json:"password_min_chars"` // Minimum password length

	// OAuth providers configuration
	OAuthProviders map[string]map[string]string `json:"oauth_providers"`

	// Application settings
	AllowSignup        bool `json:"allow_signup"`         // Allow new user registration
	RequireVerifyEmail bool `json:"require_verify_email"` // Require email verification

	// Cookie settings
	CookieName     string `json:"cookie_name"`     // Name of the authentication cookie
	CookieDomain   string `json:"cookie_domain"`   // Domain for the cookie
	CookieSecure   bool   `json:"cookie_secure"`   // Use secure cookies
	CookieHTTPOnly bool   `json:"cookie_httponly"` // HTTP only flag
}
