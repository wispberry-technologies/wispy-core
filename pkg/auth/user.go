package auth

import (
	"fmt"
	"strings"
	"time"
)

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
