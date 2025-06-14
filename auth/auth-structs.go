package auth

import (
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
