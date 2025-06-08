package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// User represents a user in the system
type User struct {
	ID               string     `json:"id" db:"id"`
	Email            string     `json:"email" db:"email"`
	EmailVerified    bool       `json:"email_verified" db:"email_verified"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at" db:"email_verified_at"`
	PasswordHash     string     `json:"-" db:"password_hash"`
	FirstName        string     `json:"first_name" db:"first_name"`
	LastName         string     `json:"last_name" db:"last_name"`
	DisplayName      string     `json:"display_name" db:"display_name"`
	Avatar           string     `json:"avatar" db:"avatar"`
	Roles            string     `json:"roles" db:"roles"` // JSON array stored as string
	IsActive         bool       `json:"is_active" db:"is_active"`
	IsLocked         bool       `json:"is_locked" db:"is_locked"`
	LastLoginAt      *time.Time `json:"last_login_at" db:"last_login_at"`
	FailedLoginCount int        `json:"failed_login_count" db:"failed_login_count"`
	LockedUntil      *time.Time `json:"locked_until" db:"locked_until"`
	TwoFactorEnabled bool       `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret  string     `json:"-" db:"two_factor_secret"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// OAuthAccount represents an OAuth account linked to a user
type OAuthAccount struct {
	ID           string     `json:"id" db:"id"`
	UserID       string     `json:"user_id" db:"user_id"`
	Provider     string     `json:"provider" db:"provider"` // 'discord', 'google'
	ProviderID   string     `json:"provider_id" db:"provider_id"`
	Email        string     `json:"email" db:"email"`
	DisplayName  string     `json:"display_name" db:"display_name"`
	Avatar       string     `json:"avatar" db:"avatar"`
	AccessToken  string     `json:"-" db:"access_token"`
	RefreshToken string     `json:"-" db:"refresh_token"`
	ExpiresAt    *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
}

// OAuthConfig holds OAuth provider configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	DisplayName string `json:"display_name"`
}

// UserInfo represents user information for API responses (without sensitive data)
type UserInfo struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	User        *UserInfo `json:"user,omitempty"`
	RedirectURL string    `json:"redirect_url,omitempty"`
}

// NewUser creates a new user with generated ID and timestamps
func NewUser(email, firstName, lastName, displayName string) *User {
	now := time.Now()
	id, _ := generateID()

	if displayName == "" {
		displayName = fmt.Sprintf("%s %s", firstName, lastName)
	}

	return &User{
		ID:               id,
		Email:            email,
		EmailVerified:    false,
		FirstName:        firstName,
		LastName:         lastName,
		DisplayName:      displayName,
		Roles:            "[]", // Empty JSON array
		IsActive:         true,
		IsLocked:         false,
		FailedLoginCount: 0,
		TwoFactorEnabled: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// NewSession creates a new session with generated ID and token
func NewSession(userID string, duration time.Duration, ipAddress, userAgent string) *Session {
	now := time.Now()
	id, _ := generateID()
	token, _ := generateToken()

	return &Session{
		ID:        id,
		UserID:    userID,
		Token:     token,
		ExpiresAt: now.Add(duration),
		CreatedAt: now,
		UpdatedAt: now,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
}

// NewOAuthAccount creates a new OAuth account
func NewOAuthAccount(userID, provider, providerID, email, displayName, avatar string) *OAuthAccount {
	now := time.Now()
	id, _ := generateID()

	return &OAuthAccount{
		ID:          id,
		UserID:      userID,
		Provider:    provider,
		ProviderID:  providerID,
		Email:       email,
		DisplayName: displayName,
		Avatar:      avatar,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// GetRoles returns the user's roles as a slice of strings
func (u *User) GetRoles() []string {
	if u.Roles == "" || u.Roles == "[]" {
		return []string{}
	}

	// Simple JSON array parsing - removes brackets and splits by comma
	roles := strings.Trim(u.Roles, "[]")
	if roles == "" {
		return []string{}
	}

	parts := strings.Split(roles, ",")
	result := make([]string, len(parts))
	for i, part := range parts {
		result[i] = strings.Trim(strings.TrimSpace(part), `"`)
	}
	return result
}

// SetRoles sets the user's roles from a slice of strings
func (u *User) SetRoles(roles []string) {
	if len(roles) == 0 {
		u.Roles = "[]"
		return
	}

	quotedRoles := make([]string, len(roles))
	for i, role := range roles {
		quotedRoles[i] = fmt.Sprintf(`"%s"`, role)
	}
	u.Roles = fmt.Sprintf("[%s]", strings.Join(quotedRoles, ","))
}

// HasRole checks if the user has a specific role
func (u *User) HasRole(role string) bool {
	roles := u.GetRoles()
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// ToUserInfo converts a User to UserInfo for API responses (removes sensitive data)
func (u *User) ToUserInfo() *UserInfo {
	return &UserInfo{
		ID:          u.ID,
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the session is valid (not expired and not empty)
func (s *Session) IsValid() bool {
	return s.Token != "" && !s.IsExpired()
}

// generateID generates a random ID
func generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateToken generates a random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
